package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	apperrors "github.com/stefanuspet/solekita/backend/internal/errors"
	"github.com/stefanuspet/solekita/backend/internal/model"
	"github.com/stefanuspet/solekita/backend/internal/repository"
)

type UserService struct {
	db         *sql.DB
	userRepo   *repository.UserRepository
	permRepo   *repository.UserPermissionRepository
}

func NewUserService(
	db *sql.DB,
	userRepo *repository.UserRepository,
	permRepo *repository.UserPermissionRepository,
) *UserService {
	return &UserService{db: db, userRepo: userRepo, permRepo: permRepo}
}

// ── Request / Response ────────────────────────────────────────────────────────

type CreateUserRequest struct {
	Name        string   `json:"name" binding:"required"`
	Phone       string   `json:"phone" binding:"required"`
	Permissions []string `json:"permissions"`
}

type UpdateUserRequest struct {
	Name        string   `json:"name" binding:"required"`
	IsActive    *bool    `json:"is_active"`
	Permissions []string `json:"permissions"`
}

type UserResponse struct {
	ID          uuid.UUID   `json:"id"`
	Name        string      `json:"name"`
	Phone       string      `json:"phone"`
	IsOwner     bool        `json:"is_owner"`
	IsActive    bool        `json:"is_active"`
	Permissions []string    `json:"permissions"`
}

type CreateUserResponse struct {
	User            UserResponse `json:"user"`
	TempPassword    string       `json:"temp_password"`
}

type ResetPasswordResponse struct {
	TempPassword string `json:"temp_password"`
}

// ── ListUsers ─────────────────────────────────────────────────────────────────

func (s *UserService) ListUsers(ctx context.Context, outletID uuid.UUID, isActive *bool) ([]*UserResponse, error) {
	users, err := s.userRepo.ListByOutletID(ctx, outletID, isActive)
	if err != nil {
		return nil, fmt.Errorf("ListUsers: %w", err)
	}

	res := make([]*UserResponse, 0, len(users))
	for _, u := range users {
		perms, err := s.permRepo.GetByUserID(ctx, u.ID)
		if err != nil {
			return nil, fmt.Errorf("ListUsers: ambil permissions user %s: %w", u.ID, err)
		}
		res = append(res, userToResponse(u, perms))
	}
	return res, nil
}

// ── GetUser ───────────────────────────────────────────────────────────────────

func (s *UserService) GetUser(ctx context.Context, userID, outletID uuid.UUID) (*UserResponse, error) {
	u, err := s.userRepo.GetByID(ctx, userID, outletID)
	if err != nil {
		return nil, fmt.Errorf("GetUser: %w", err)
	}

	perms, err := s.permRepo.GetByUserID(ctx, u.ID)
	if err != nil {
		return nil, fmt.Errorf("GetUser: ambil permissions: %w", err)
	}
	return userToResponse(u, perms), nil
}

// ── CreateUser ────────────────────────────────────────────────────────────────

func (s *UserService) CreateUser(ctx context.Context, outletID uuid.UUID, req CreateUserRequest) (*CreateUserResponse, error) {
	// Cek nomor HP tidak duplikat dalam outlet
	existing, err := s.userRepo.GetByPhone(ctx, req.Phone)
	if err == nil && existing.OutletID == outletID {
		return nil, apperrors.ErrConflict.New(fmt.Errorf("nomor HP sudah terdaftar di outlet ini"))
	}

	tempPass := generateTempPassword(8)
	hash, err := bcrypt.GenerateFromPassword([]byte(tempPass), 12)
	if err != nil {
		return nil, fmt.Errorf("CreateUser: hash password: %w", err)
	}

	perms := parsePermissions(req.Permissions)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("CreateUser: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	user := &model.User{
		ID:           uuid.New(),
		OutletID:     outletID,
		Name:         req.Name,
		Phone:        req.Phone,
		PasswordHash: string(hash),
		IsOwner:      false,
		IsActive:     true,
	}
	if err = s.userRepo.Create(ctx, tx, user); err != nil {
		return nil, fmt.Errorf("CreateUser: simpan user: %w", err)
	}

	if err = s.permRepo.ReplaceAll(ctx, tx, user.ID, perms); err != nil {
		return nil, fmt.Errorf("CreateUser: simpan permissions: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("CreateUser: commit: %w", err)
	}

	return &CreateUserResponse{
		User:         *userToResponse(user, perms),
		TempPassword: tempPass,
	}, nil
}

// ── UpdateUser ────────────────────────────────────────────────────────────────

func (s *UserService) UpdateUser(ctx context.Context, userID, outletID uuid.UUID, req UpdateUserRequest) (*UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID, outletID)
	if err != nil {
		return nil, fmt.Errorf("UpdateUser: ambil user: %w", err)
	}
	if user.IsOwner {
		return nil, apperrors.ErrForbidden.New(fmt.Errorf("tidak bisa mengubah data owner"))
	}

	user.Name = req.Name
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	perms := parsePermissions(req.Permissions)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("UpdateUser: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if err = s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("UpdateUser: simpan user: %w", err)
	}

	if err = s.permRepo.ReplaceAll(ctx, tx, user.ID, perms); err != nil {
		return nil, fmt.Errorf("UpdateUser: simpan permissions: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("UpdateUser: commit: %w", err)
	}

	return userToResponse(user, perms), nil
}

// ── ResetPassword ─────────────────────────────────────────────────────────────

func (s *UserService) ResetPassword(ctx context.Context, userID, outletID uuid.UUID) (*ResetPasswordResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID, outletID)
	if err != nil {
		return nil, fmt.Errorf("ResetPassword: ambil user: %w", err)
	}
	if user.IsOwner {
		return nil, apperrors.ErrForbidden.New(fmt.Errorf("tidak bisa reset password owner"))
	}

	tempPass := generateTempPassword(8)
	hash, err := bcrypt.GenerateFromPassword([]byte(tempPass), 12)
	if err != nil {
		return nil, fmt.Errorf("ResetPassword: hash password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, userID, outletID, string(hash)); err != nil {
		return nil, fmt.Errorf("ResetPassword: simpan password: %w", err)
	}

	return &ResetPasswordResponse{TempPassword: tempPass}, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

const tempPasswordChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateTempPassword(length int) string {
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(tempPasswordChars))))
		b[i] = tempPasswordChars[n.Int64()]
	}
	return string(b)
}

func parsePermissions(raw []string) []model.Permission {
	perms := make([]model.Permission, 0, len(raw))
	valid := map[string]model.Permission{
		string(model.PermissionManageOrder):    model.PermissionManageOrder,
		string(model.PermissionUpdateStatus):   model.PermissionUpdateStatus,
		string(model.PermissionManageDelivery): model.PermissionManageDelivery,
		string(model.PermissionViewReport):     model.PermissionViewReport,
		string(model.PermissionManageOutlet):   model.PermissionManageOutlet,
	}
	for _, r := range raw {
		if p, ok := valid[r]; ok {
			perms = append(perms, p)
		}
	}
	return perms
}

func userToResponse(u *model.User, perms []model.Permission) *UserResponse {
	permStrings := make([]string, len(perms))
	for i, p := range perms {
		permStrings[i] = string(p)
	}
	return &UserResponse{
		ID:          u.ID,
		Name:        u.Name,
		Phone:       u.Phone,
		IsOwner:     u.IsOwner,
		IsActive:    u.IsActive,
		Permissions: permStrings,
	}
}
