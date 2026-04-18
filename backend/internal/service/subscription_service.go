package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/stefanuspet/solekita/backend/internal/errors"
	"github.com/stefanuspet/solekita/backend/internal/model"
	"github.com/stefanuspet/solekita/backend/internal/repository"
)

// TripayClient adalah interface untuk payment gateway Tripay.
// Implementasi konkret ada di internal/payment/tripay.go (T192).
type TripayClient interface {
	CreateTransaction(ctx context.Context, req TripayTransactionRequest) (*TripayTransactionResult, error)
	ValidateSignature(body []byte, signature string) bool
}

type TripayTransactionRequest struct {
	MerchantRef  string // ID unik invoice kita
	Amount       int
	CustomerName string
	CustomerPhone string
	CustomerEmail string
	OrderItems   []TripayOrderItem
	ExpiredTime  int64 // unix timestamp
	ReturnURL    string
}

type TripayOrderItem struct {
	Name     string
	Price    int
	Quantity int
}

type TripayTransactionResult struct {
	Reference  string
	PaymentURL string
}

// ── Pricing constants ─────────────────────────────────────────────────────────

const (
	priceMonthly  = 29000  // Rp29.000/bulan
	priceBiannual = 156600 // Rp156.600/6 bulan (diskon 10%)
)

// ── Service ───────────────────────────────────────────────────────────────────

type SubscriptionService struct {
	subRepo     *repository.SubscriptionRepository
	invoiceRepo *repository.SubscriptionInvoiceRepository
	outletRepo  *repository.OutletRepository
	userRepo    *repository.UserRepository
	tripay      TripayClient
}

func NewSubscriptionService(
	subRepo *repository.SubscriptionRepository,
	invoiceRepo *repository.SubscriptionInvoiceRepository,
	outletRepo *repository.OutletRepository,
	userRepo *repository.UserRepository,
	tripay TripayClient,
) *SubscriptionService {
	return &SubscriptionService{
		subRepo:     subRepo,
		invoiceRepo: invoiceRepo,
		outletRepo:  outletRepo,
		userRepo:    userRepo,
		tripay:      tripay,
	}
}

// ── Response types ────────────────────────────────────────────────────────────

type SubscriptionResponse struct {
	ID                    uuid.UUID              `json:"id"`
	Plan                  model.SubscriptionPlan `json:"plan"`
	Status                string                 `json:"status"`
	PricePerMonth         int                    `json:"price_per_month"`
	TrialStartedAt        *time.Time             `json:"trial_started_at"`
	TrialEndsAt           *time.Time             `json:"trial_ends_at"`
	TrialDaysRemaining    int                    `json:"trial_days_remaining"`
	SubscriptionStartedAt *time.Time             `json:"subscription_started_at"`
	NextDueDate           *time.Time             `json:"next_due_date"`
}

type ConvertTrialResponse struct {
	InvoiceID  uuid.UUID `json:"invoice_id"`
	Amount     int       `json:"amount"`
	DueDate    time.Time `json:"due_date"`
	PaymentURL string    `json:"payment_url"`
	Reference  string    `json:"reference"`
}

// ── GetMySubscription ─────────────────────────────────────────────────────────

// GetMySubscription mengembalikan detail subscription dengan trial_days_remaining.
func (s *SubscriptionService) GetMySubscription(ctx context.Context, outletID uuid.UUID) (*SubscriptionResponse, error) {
	sub, err := s.subRepo.GetByOutletID(ctx, outletID)
	if err != nil {
		return nil, fmt.Errorf("GetMySubscription: %w", err)
	}

	outlet, err := s.outletRepo.GetByID(ctx, outletID)
	if err != nil {
		return nil, fmt.Errorf("GetMySubscription: ambil outlet: %w", err)
	}

	trialDaysRemaining := 0
	if sub.TrialEndsAt != nil && outlet.SubscriptionStatus == model.SubscriptionStatusTrial {
		remaining := time.Until(*sub.TrialEndsAt)
		if remaining > 0 {
			trialDaysRemaining = int(remaining.Hours()/24) + 1
		}
	}

	return &SubscriptionResponse{
		ID:                    sub.ID,
		Plan:                  sub.Plan,
		Status:                string(outlet.SubscriptionStatus),
		PricePerMonth:         sub.PricePerMonth,
		TrialStartedAt:        sub.TrialStartedAt,
		TrialEndsAt:           sub.TrialEndsAt,
		TrialDaysRemaining:    trialDaysRemaining,
		SubscriptionStartedAt: sub.SubscriptionStartedAt,
		NextDueDate:           sub.NextDueDate,
	}, nil
}

// ── ListInvoices ──────────────────────────────────────────────────────────────

type InvoiceListResponse struct {
	Invoices []*model.SubscriptionInvoice `json:"invoices"`
	Total    int                          `json:"total"`
	Page     int                          `json:"page"`
	Limit    int                          `json:"limit"`
}

func (s *SubscriptionService) ListInvoices(ctx context.Context, outletID uuid.UUID, page, limit int) (*InvoiceListResponse, error) {
	invoices, total, err := s.invoiceRepo.ListByOutletID(ctx, outletID, page, limit)
	if err != nil {
		return nil, fmt.Errorf("ListInvoices: %w", err)
	}
	return &InvoiceListResponse{
		Invoices: invoices,
		Total:    total,
		Page:     page,
		Limit:    limit,
	}, nil
}

// ── ConvertTrial ──────────────────────────────────────────────────────────────

// ConvertTrial mengkonversi outlet dari trial/inactive ke subscription berbayar.
// Membuat tagihan Tripay dan menyimpan invoice ke DB.
// Status outlet diubah ke active oleh webhook setelah pembayaran berhasil.
func (s *SubscriptionService) ConvertTrial(ctx context.Context, outletID uuid.UUID, plan model.SubscriptionPlan) (*ConvertTrialResponse, error) {
	if s.tripay == nil {
		return nil, apperrors.ErrUnprocessable.New(fmt.Errorf("payment gateway belum dikonfigurasi"))
	}

	outlet, err := s.outletRepo.GetByID(ctx, outletID)
	if err != nil {
		return nil, fmt.Errorf("ConvertTrial: ambil outlet: %w", err)
	}

	// Validasi status outlet
	switch outlet.SubscriptionStatus {
	case model.SubscriptionStatusTrial, model.SubscriptionStatusInactive:
		// OK
	case model.SubscriptionStatusActive:
		return nil, apperrors.ErrUnprocessable.New(fmt.Errorf("outlet sudah aktif berlangganan"))
	default:
		return nil, apperrors.ErrUnprocessable.New(fmt.Errorf("status outlet tidak mendukung konversi"))
	}

	sub, err := s.subRepo.GetByOutletID(ctx, outletID)
	if err != nil {
		return nil, fmt.Errorf("ConvertTrial: ambil subscription: %w", err)
	}

	// Hitung amount dan due date berdasarkan plan
	amount, dueDate := planDetails(plan)

	// Ambil data owner untuk customer info Tripay
	owner, err := s.userRepo.GetByIDInternal(ctx, outlet.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("ConvertTrial: ambil owner: %w", err)
	}

	// Simpan invoice ke DB dulu (status pending, belum ada reference Tripay)
	inv := &model.SubscriptionInvoice{
		SubscriptionID: sub.ID,
		OutletID:       outletID,
		Amount:         amount,
		DueDate:        dueDate,
		Status:         model.InvoiceStatusPending,
	}
	if err := s.invoiceRepo.Create(ctx, inv); err != nil {
		return nil, fmt.Errorf("ConvertTrial: simpan invoice: %w", err)
	}

	// Buat transaksi Tripay
	expiredAt := time.Now().Add(24 * time.Hour) // expired 24 jam
	tripayReq := TripayTransactionRequest{
		MerchantRef:   inv.ID.String(),
		Amount:        amount,
		CustomerName:  owner.Name,
		CustomerPhone: owner.Phone,
		OrderItems: []TripayOrderItem{{
			Name:     planLabel(plan),
			Price:    amount,
			Quantity: 1,
		}},
		ExpiredTime: expiredAt.Unix(),
	}

	tripayRes, err := s.tripay.CreateTransaction(ctx, tripayReq)
	if err != nil {
		return nil, fmt.Errorf("ConvertTrial: buat transaksi Tripay: %w", err)
	}

	// Update invoice dengan reference dan payment URL dari Tripay
	inv.TripayReference = &tripayRes.Reference
	inv.TripayPaymentURL = &tripayRes.PaymentURL
	if err := s.invoiceRepo.Update(ctx, inv); err != nil {
		return nil, fmt.Errorf("ConvertTrial: update invoice: %w", err)
	}

	return &ConvertTrialResponse{
		InvoiceID:  inv.ID,
		Amount:     amount,
		DueDate:    dueDate,
		PaymentURL: tripayRes.PaymentURL,
		Reference:  tripayRes.Reference,
	}, nil
}

// ── GenerateBillingInvoice ────────────────────────────────────────────────────

// GenerateBillingInvoice membuat tagihan perpanjangan untuk outlet aktif.
// Idempotent: jika sudah ada invoice pending untuk outlet ini, kembalikan yang ada.
// Dipanggil oleh billing scheduler saat next_due_date jatuh hari ini.
func (s *SubscriptionService) GenerateBillingInvoice(ctx context.Context, outletID uuid.UUID) (*ConvertTrialResponse, error) {
	if s.tripay == nil {
		return nil, apperrors.ErrUnprocessable.New(fmt.Errorf("payment gateway belum dikonfigurasi"))
	}

	// Idempotency: kembalikan invoice pending yang sudah ada
	if existing, err := s.invoiceRepo.GetPendingByOutletID(ctx, outletID); err == nil {
		paymentURL := ""
		if existing.TripayPaymentURL != nil {
			paymentURL = *existing.TripayPaymentURL
		}
		ref := ""
		if existing.TripayReference != nil {
			ref = *existing.TripayReference
		}
		return &ConvertTrialResponse{
			InvoiceID:  existing.ID,
			Amount:     existing.Amount,
			DueDate:    existing.DueDate,
			PaymentURL: paymentURL,
			Reference:  ref,
		}, nil
	}

	outlet, err := s.outletRepo.GetByID(ctx, outletID)
	if err != nil {
		return nil, fmt.Errorf("GenerateBillingInvoice: ambil outlet: %w", err)
	}

	sub, err := s.subRepo.GetByOutletID(ctx, outletID)
	if err != nil {
		return nil, fmt.Errorf("GenerateBillingInvoice: ambil subscription: %w", err)
	}

	owner, err := s.userRepo.GetByIDInternal(ctx, outlet.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("GenerateBillingInvoice: ambil owner: %w", err)
	}

	amount, dueDate := planDetails(sub.Plan)

	inv := &model.SubscriptionInvoice{
		SubscriptionID: sub.ID,
		OutletID:       outletID,
		Amount:         amount,
		DueDate:        dueDate,
		Status:         model.InvoiceStatusPending,
	}
	if err := s.invoiceRepo.Create(ctx, inv); err != nil {
		return nil, fmt.Errorf("GenerateBillingInvoice: simpan invoice: %w", err)
	}

	expiredAt := time.Now().Add(72 * time.Hour) // batas bayar 3 hari
	tripayReq := TripayTransactionRequest{
		MerchantRef:   inv.ID.String(),
		Amount:        amount,
		CustomerName:  owner.Name,
		CustomerPhone: owner.Phone,
		OrderItems: []TripayOrderItem{{
			Name:     planLabel(sub.Plan),
			Price:    amount,
			Quantity: 1,
		}},
		ExpiredTime: expiredAt.Unix(),
	}

	tripayRes, err := s.tripay.CreateTransaction(ctx, tripayReq)
	if err != nil {
		return nil, fmt.Errorf("GenerateBillingInvoice: buat transaksi Tripay: %w", err)
	}

	inv.TripayReference = &tripayRes.Reference
	inv.TripayPaymentURL = &tripayRes.PaymentURL
	if err := s.invoiceRepo.Update(ctx, inv); err != nil {
		return nil, fmt.Errorf("GenerateBillingInvoice: update invoice: %w", err)
	}

	return &ConvertTrialResponse{
		InvoiceID:  inv.ID,
		Amount:     amount,
		DueDate:    dueDate,
		PaymentURL: tripayRes.PaymentURL,
		Reference:  tripayRes.Reference,
	}, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func planDetails(plan model.SubscriptionPlan) (amount int, nextDueDate time.Time) {
	now := time.Now()
	switch plan {
	case model.SubscriptionPlanBiannual:
		return priceBiannual, now.AddDate(0, 6, 0)
	default: // monthly
		return priceMonthly, now.AddDate(0, 1, 0)
	}
}

func planLabel(plan model.SubscriptionPlan) string {
	switch plan {
	case model.SubscriptionPlanBiannual:
		return "Solekita 6 Bulan"
	default:
		return "Solekita Bulanan"
	}
}
