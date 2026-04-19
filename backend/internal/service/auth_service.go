package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/stefanuspet/solekita/backend/internal/config"
	apperrors "github.com/stefanuspet/solekita/backend/internal/errors"
	"github.com/stefanuspet/solekita/backend/internal/middleware"
	"github.com/stefanuspet/solekita/backend/internal/model"
	"github.com/stefanuspet/solekita/backend/internal/notification"
	"github.com/stefanuspet/solekita/backend/internal/repository"
)

type AuthService struct {
	db         *sql.DB
	outletRepo *repository.OutletRepository
	userRepo   *repository.UserRepository
	subRepo    *repository.SubscriptionRepository
	rtRepo     *repository.RefreshTokenRepository
	cfg        *config.Config
	fonnte     *notification.FonnteClient
}

func NewAuthService(
	db *sql.DB,
	outletRepo *repository.OutletRepository,
	userRepo *repository.UserRepository,
	subRepo *repository.SubscriptionRepository,
	rtRepo *repository.RefreshTokenRepository,
	cfg *config.Config,
	fonnte *notification.FonnteClient,
) *AuthService {
	return &AuthService{
		db:         db,
		outletRepo: outletRepo,
		userRepo:   userRepo,
		subRepo:    subRepo,
		rtRepo:     rtRepo,
		cfg:        cfg,
		fonnte:     fonnte,
	}
}

// ── Request / Response ────────────────────────────────────────────────────────

type RegisterRequest struct {
	BusinessName string `json:"business_name" binding:"required,min=3"`
	Phone        string `json:"phone" binding:"required"`
	Password     string `json:"password" binding:"required,min=8"`
}

type TrialData struct {
	StartedAt     time.Time `json:"started_at"`
	EndsAt        time.Time `json:"ends_at"`
	DaysRemaining int       `json:"days_remaining"`
}

type AuthResponse struct {
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	ExpiresIn    int        `json:"expires_in"`
	User         UserData   `json:"user"`
	Trial        *TrialData `json:"trial,omitempty"`
}

type UserData struct {
	ID          uuid.UUID    `json:"id"`
	Name        string       `json:"name"`
	Phone       string       `json:"phone"`
	IsOwner     bool         `json:"is_owner"`
	Outlet      OutletData   `json:"outlet"`
	Permissions []string     `json:"permissions"`
}

type OutletData struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Code string    `json:"code"`
}

// ── Register ──────────────────────────────────────────────────────────────────

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (res *AuthResponse, err error) {
	// 1. Validasi nomor HP tidak duplikat
	_, lookupErr := s.userRepo.GetByPhone(ctx, req.Phone)
	if lookupErr == nil {
		return nil, apperrors.ErrConflict.New(fmt.Errorf("nomor HP sudah terdaftar. Silakan login"))
	}
	if !errors.Is(lookupErr, apperrors.ErrNotFound) {
		return nil, fmt.Errorf("Register: cek duplikasi phone: %w", lookupErr)
	}

	// 2. Auto-generate outlet code unik
	outletCode, err := s.generateUniqueOutletCode(ctx, req.BusinessName)
	if err != nil {
		return nil, fmt.Errorf("Register: generate outlet code: %w", err)
	}

	// 3. Hash password bcrypt cost 12
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("Register: hash password: %w", err)
	}

	// 4. Buat outlet + user + subscription dalam 1 transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("Register: begin tx: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// Pre-generate user UUID agar bisa di-set sebagai owner_id outlet
	// sebelum user diinsert. FK fk_outlets_owner bersifat DEFERRABLE INITIALLY DEFERRED,
	// sehingga constraint baru dicek saat COMMIT — tidak error di tengah transaction.
	userID := uuid.New()

	outlet := &model.Outlet{
		Name:                 req.BusinessName,
		Code:                 outletCode,
		OwnerID:              userID,
		SubscriptionStatus:   model.SubscriptionStatusTrial,
		OverdueThresholdDays: 7,
		IsActive:             true,
	}
	if err = s.outletRepo.Create(ctx, tx, outlet); err != nil {
		return nil, fmt.Errorf("Register: create outlet: %w", err)
	}

	// Nama user default = nomor HP sampai owner update profil
	user := &model.User{
		ID:           userID,
		OutletID:     outlet.ID,
		Name:         req.Phone,
		Phone:        req.Phone,
		PasswordHash: string(hash),
		IsOwner:      true,
		IsActive:     true,
	}
	if err = s.userRepo.Create(ctx, tx, user); err != nil {
		return nil, fmt.Errorf("Register: create user: %w", err)
	}

	now := time.Now()
	trialEnds := now.Add(14 * 24 * time.Hour)
	sub := &model.Subscription{
		OutletID:       outlet.ID,
		Plan:           model.SubscriptionPlanMonthly,
		PricePerMonth:  29000,
		TrialStartedAt: &now,
		TrialEndsAt:    &trialEnds,
	}
	if err = s.subRepo.Create(ctx, tx, sub); err != nil {
		return nil, fmt.Errorf("Register: create subscription: %w", err)
	}

	// 5. Generate access token + refresh token
	accessToken, expiresIn, err := s.generateAccessToken(user, outlet, []model.Permission{})
	if err != nil {
		return nil, fmt.Errorf("Register: generate access token: %w", err)
	}

	rawRefreshToken, tokenHash, refreshExpiry, err := s.generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("Register: generate refresh token: %w", err)
	}

	rt := &model.RefreshToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiredAt: refreshExpiry,
	}
	if err = s.rtRepo.Create(ctx, tx, rt); err != nil {
		return nil, fmt.Errorf("Register: save refresh token: %w", err)
	}

	// tx.Commit() dipanggil oleh defer

	// Kirim WA sambutan async
	if s.fonnte != nil {
		outletName := req.BusinessName
		phone := req.Phone
		go func() {
			msg := fmt.Sprintf(
				"Halo! Selamat datang di Solekita 👟\n\nOutlet *%s* berhasil terdaftar. Nikmati 14 hari masa trial gratis!\n\nSilakan login dan mulai kelola laundry sepatu Anda.",
				outletName,
			)
			if err := s.fonnte.Send(context.Background(), phone, msg); err != nil {
				_ = err
			}
		}()
	}

	daysRemaining := int(time.Until(trialEnds).Hours() / 24)

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		ExpiresIn:    expiresIn,
		User: UserData{
			ID:      user.ID,
			Name:    user.Name,
			Phone:   user.Phone,
			IsOwner: user.IsOwner,
			Outlet: OutletData{
				ID:   outlet.ID,
				Name: outlet.Name,
				Code: outlet.Code,
			},
			Permissions: []string{},
		},
		Trial: &TrialData{
			StartedAt:     now,
			EndsAt:        trialEnds,
			DaysRemaining: daysRemaining,
		},
	}, nil
}

// generateUniqueOutletCode membuat kode outlet dari 3 huruf pertama nama bisnis + 2 digit random.
// Contoh: "Solekita Jogja" → "SOL" + "47" = "SOL47"
// Coba maksimal 10 kali jika ada collision.
func (s *AuthService) generateUniqueOutletCode(ctx context.Context, businessName string) (string, error) {
	// Ambil 3 huruf pertama (hanya A-Z)
	re := regexp.MustCompile(`[A-Za-z]`)
	letters := re.FindAllString(businessName, -1)
	prefix := ""
	for i, l := range letters {
		if i >= 3 {
			break
		}
		prefix += strings.ToUpper(l)
	}
	for len(prefix) < 3 {
		prefix += "X"
	}

	for range 10 {
		n1, _ := rand.Int(rand.Reader, big.NewInt(10))
		n2, _ := rand.Int(rand.Reader, big.NewInt(10))
		code := fmt.Sprintf("%s%s%s", prefix, n1.String(), n2.String())

		exists, err := s.outletRepo.CodeExists(ctx, code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}
	return "", fmt.Errorf("tidak bisa generate kode outlet unik")
}

// ── Login ─────────────────────────────────────────────────────────────────────

type LoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	// 1. Cek nomor HP ada
	user, err := s.userRepo.GetByPhone(ctx, req.Phone)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.ErrUnauthorized.New(fmt.Errorf("nomor HP atau password salah"))
		}
		return nil, fmt.Errorf("Login: cek user: %w", err)
	}

	// 2. Verifikasi password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apperrors.ErrUnauthorized.New(fmt.Errorf("nomor HP atau password salah"))
	}

	// 3. Cek user aktif
	if !user.IsActive {
		return nil, apperrors.ErrForbidden.New(fmt.Errorf("akun tidak aktif"))
	}

	// 4. Cek status langganan outlet — kasir diblokir jika suspended/inactive, owner tetap bisa masuk
	outlet, err := s.outletRepo.GetByID(ctx, user.OutletID)
	if err != nil {
		return nil, fmt.Errorf("Login: ambil outlet: %w", err)
	}
	if !user.IsOwner {
		switch outlet.SubscriptionStatus {
		case model.SubscriptionStatusSuspended:
			return nil, apperrors.ErrForbidden.New(fmt.Errorf("langganan outlet ditangguhkan"))
		case model.SubscriptionStatusInactive:
			return nil, apperrors.ErrForbidden.New(fmt.Errorf("langganan outlet tidak aktif"))
		}
	}

	// 5. Ambil permissions
	perms, err := s.userRepo.GetPermissions(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("Login: ambil permissions: %w", err)
	}

	// 6. Generate token
	accessToken, expiresIn, err := s.generateAccessToken(user, outlet, perms)
	if err != nil {
		return nil, fmt.Errorf("Login: generate access token: %w", err)
	}

	rawRefreshToken, tokenHash, refreshExpiry, err := s.generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("Login: generate refresh token: %w", err)
	}

	rt := &model.RefreshToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiredAt: refreshExpiry,
	}
	if err := s.rtRepo.Create(ctx, s.db, rt); err != nil {
		return nil, fmt.Errorf("Login: simpan refresh token: %w", err)
	}

	// 7. Update last_login_at (best-effort, tidak gagalkan login)
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	permStrings := make([]string, len(perms))
	for i, p := range perms {
		permStrings[i] = string(p)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		ExpiresIn:    expiresIn,
		User: UserData{
			ID:      user.ID,
			Name:    user.Name,
			Phone:   user.Phone,
			IsOwner: user.IsOwner,
			Outlet: OutletData{
				ID:   outlet.ID,
				Name: outlet.Name,
				Code: outlet.Code,
			},
			Permissions: permStrings,
		},
	}, nil
}

// ── Logout ────────────────────────────────────────────────────────────────────

func (s *AuthService) Logout(ctx context.Context, rawToken string) error {
	tokenHash := HashToken(rawToken)
	if err := s.rtRepo.DeleteByTokenHash(ctx, tokenHash); err != nil {
		return fmt.Errorf("Logout: hapus token: %w", err)
	}
	return nil
}

// ── RefreshToken ──────────────────────────────────────────────────────────────

func (s *AuthService) RefreshToken(ctx context.Context, rawToken string) (*AuthResponse, error) {
	tokenHash := HashToken(rawToken)

	// 1. Cari token di DB
	rt, err := s.rtRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.ErrUnauthorized.New(fmt.Errorf("refresh token tidak valid"))
		}
		return nil, fmt.Errorf("RefreshToken: cari token: %w", err)
	}

	// 2. Cek expired
	if time.Now().After(rt.ExpiredAt) {
		_ = s.rtRepo.DeleteByTokenHash(ctx, tokenHash)
		return nil, apperrors.ErrUnauthorized.New(fmt.Errorf("refresh token sudah kadaluarsa"))
	}

	// 3. Ambil user (trusted — token sudah diverifikasi dari DB)
	user, err := s.userRepo.GetByIDInternal(ctx, rt.UserID)
	if err != nil {
		return nil, fmt.Errorf("RefreshToken: ambil user: %w", err)
	}

	// 4. Cek user aktif
	if !user.IsActive {
		return nil, apperrors.ErrForbidden.New(fmt.Errorf("akun tidak aktif"))
	}

	// 5. Ambil outlet + cek status langganan (kasir diblokir jika suspended/inactive)
	outlet, err := s.outletRepo.GetByID(ctx, user.OutletID)
	if err != nil {
		return nil, fmt.Errorf("RefreshToken: ambil outlet: %w", err)
	}
	if !user.IsOwner {
		switch outlet.SubscriptionStatus {
		case model.SubscriptionStatusSuspended:
			return nil, apperrors.ErrForbidden.New(fmt.Errorf("langganan outlet ditangguhkan"))
		case model.SubscriptionStatusInactive:
			return nil, apperrors.ErrForbidden.New(fmt.Errorf("langganan outlet tidak aktif"))
		}
	}

	// 6. Ambil permissions
	perms, err := s.userRepo.GetPermissions(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("RefreshToken: ambil permissions: %w", err)
	}

	// 7. Token rotation: hapus lama, buat baru dalam transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("RefreshToken: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if err = s.rtRepo.DeleteByTokenHash(ctx, tokenHash); err != nil {
		return nil, fmt.Errorf("RefreshToken: hapus token lama: %w", err)
	}

	rawNew, hashNew, expiryNew, err := s.generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("RefreshToken: generate token baru: %w", err)
	}

	newRT := &model.RefreshToken{
		UserID:    user.ID,
		TokenHash: hashNew,
		ExpiredAt: expiryNew,
	}
	if err = s.rtRepo.Create(ctx, tx, newRT); err != nil {
		return nil, fmt.Errorf("RefreshToken: simpan token baru: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("RefreshToken: commit: %w", err)
	}

	// 8. Generate access token baru
	accessToken, expiresIn, err := s.generateAccessToken(user, outlet, perms)
	if err != nil {
		return nil, fmt.Errorf("RefreshToken: generate access token: %w", err)
	}

	permStrings := make([]string, len(perms))
	for i, p := range perms {
		permStrings[i] = string(p)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: rawNew,
		ExpiresIn:    expiresIn,
		User: UserData{
			ID:      user.ID,
			Name:    user.Name,
			Phone:   user.Phone,
			IsOwner: user.IsOwner,
			Outlet: OutletData{
				ID:   outlet.ID,
				Name: outlet.Name,
				Code: outlet.Code,
			},
			Permissions: permStrings,
		},
	}, nil
}

// ── Token helpers ─────────────────────────────────────────────────────────────

func (s *AuthService) generateAccessToken(user *model.User, outlet *model.Outlet, permissions []model.Permission) (string, int, error) {
	expiry, err := time.ParseDuration(s.cfg.JWTAccessExpiry)
	if err != nil {
		expiry = 15 * time.Minute
	}

	claims := middleware.JWTClaims{
		UserID:      user.ID,
		OutletID:    outlet.ID,
		OutletCode:  outlet.Code,
		Name:        user.Name,
		Phone:       user.Phone,
		IsOwner:     user.IsOwner,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", 0, fmt.Errorf("generateAccessToken: %w", err)
	}

	return signed, int(expiry.Seconds()), nil
}

func (s *AuthService) generateRefreshToken() (rawToken, tokenHash string, expiry time.Time, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", time.Time{}, fmt.Errorf("generateRefreshToken: %w", err)
	}
	rawToken = hex.EncodeToString(b)

	sum := sha256.Sum256([]byte(rawToken))
	tokenHash = hex.EncodeToString(sum[:])

	refreshDuration, parseErr := time.ParseDuration(s.cfg.JWTRefreshExpiry)
	if parseErr != nil {
		refreshDuration = 720 * time.Hour
	}
	expiry = time.Now().Add(refreshDuration)

	return rawToken, tokenHash, expiry, nil
}

// HashToken menghasilkan SHA256 hex dari token string.
// Digunakan untuk mencari token di DB dari raw token yang diterima dari client.
func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
