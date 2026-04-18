package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/stefanuspet/solekita/backend/internal/errors"
	"github.com/stefanuspet/solekita/backend/internal/model"
	"github.com/stefanuspet/solekita/backend/internal/notification"
	"github.com/stefanuspet/solekita/backend/internal/repository"
)

type WebhookService struct {
	db          *sql.DB
	invoiceRepo *repository.SubscriptionInvoiceRepository
	subRepo     *repository.SubscriptionRepository
	outletRepo  *repository.OutletRepository
	userRepo    *repository.UserRepository
	fonnte      *notification.FonnteClient
	tripay      TripayClient
}

func NewWebhookService(
	db *sql.DB,
	invoiceRepo *repository.SubscriptionInvoiceRepository,
	subRepo *repository.SubscriptionRepository,
	outletRepo *repository.OutletRepository,
	userRepo *repository.UserRepository,
	fonnte *notification.FonnteClient,
	tripay TripayClient,
) *WebhookService {
	return &WebhookService{
		db:          db,
		invoiceRepo: invoiceRepo,
		subRepo:     subRepo,
		outletRepo:  outletRepo,
		userRepo:    userRepo,
		fonnte:      fonnte,
		tripay:      tripay,
	}
}

// ── Tripay webhook body ───────────────────────────────────────────────────────

type TripayWebhookPayload struct {
	Reference   string `json:"reference"`
	MerchantRef string `json:"merchant_ref"`
	Status      string `json:"status"`
	PaidAt      *int64 `json:"paid_at"`
	TotalAmount int    `json:"total_amount"`
}

// ── HandleTripayPayment ───────────────────────────────────────────────────────

// HandleTripayPayment memproses webhook dari Tripay.
// rawBody digunakan untuk validasi signature sebelum di-parse.
func (s *WebhookService) HandleTripayPayment(ctx context.Context, rawBody []byte, signature string) error {
	// 1. Validasi signature
	if s.tripay == nil || !s.tripay.ValidateSignature(rawBody, signature) {
		return apperrors.ErrUnauthorized.New(fmt.Errorf("signature tidak valid"))
	}

	// 2. Parse body
	var payload TripayWebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return apperrors.ErrBadRequest.New(fmt.Errorf("body tidak valid: %w", err))
	}

	// Hanya proses status PAID
	if payload.Status != "PAID" {
		return nil
	}

	// 3. Ambil invoice by merchant_ref (= invoice UUID yang kita kirim ke Tripay)
	invoiceID, err := uuid.Parse(payload.MerchantRef)
	if err != nil {
		return apperrors.ErrBadRequest.New(fmt.Errorf("merchant_ref bukan UUID valid"))
	}

	inv, err := s.invoiceRepo.GetByID(ctx, invoiceID)
	if err != nil {
		return fmt.Errorf("HandleTripayPayment: ambil invoice: %w", err)
	}

	// 4. Idempotent — sudah paid, abaikan
	if inv.Status == model.InvoiceStatusPaid {
		return nil
	}

	// 5. Update invoice → paid
	now := time.Now()
	if payload.PaidAt != nil {
		t := time.Unix(*payload.PaidAt, 0)
		now = t
	}
	inv.Status = model.InvoiceStatusPaid
	inv.PaidAt = &now
	ref := payload.Reference
	inv.TripayReference = &ref
	if err := s.invoiceRepo.Update(ctx, inv); err != nil {
		return fmt.Errorf("HandleTripayPayment: update invoice: %w", err)
	}

	// 6. Ambil subscription
	sub, err := s.subRepo.GetByOutletID(ctx, inv.OutletID)
	if err != nil {
		return fmt.Errorf("HandleTripayPayment: ambil subscription: %w", err)
	}

	// 7. Tentukan plan dan next_due_date dari amount invoice
	plan, nextDueDate := planFromAmount(inv.Amount, now)

	// 8. Set subscription_started_at hanya jika belum pernah diset (konversi trial pertama kali)
	var subscriptionStartedAt *time.Time
	if sub.SubscriptionStartedAt == nil {
		subscriptionStartedAt = &now
	}

	if err := s.subRepo.UpdatePlan(ctx, s.db, sub.ID, plan, &nextDueDate, subscriptionStartedAt); err != nil {
		return fmt.Errorf("HandleTripayPayment: update subscription: %w", err)
	}

	// 9. Update outlet subscription_status = active
	if err := s.outletRepo.UpdateSubscriptionStatus(ctx, inv.OutletID, model.SubscriptionStatusActive); err != nil {
		return fmt.Errorf("HandleTripayPayment: update outlet status: %w", err)
	}

	// 10. Kirim WA konfirmasi ke owner (async, best-effort)
	if s.fonnte != nil {
		go s.sendPaymentConfirmationWA(inv.OutletID, plan, nextDueDate)
	}

	return nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func planFromAmount(amount int, paidAt time.Time) (model.SubscriptionPlan, time.Time) {
	if amount >= priceBiannual {
		return model.SubscriptionPlanBiannual, paidAt.AddDate(0, 6, 0)
	}
	return model.SubscriptionPlanMonthly, paidAt.AddDate(0, 1, 0)
}

func (s *WebhookService) sendPaymentConfirmationWA(outletID uuid.UUID, plan model.SubscriptionPlan, nextDue time.Time) {
	ctx := context.Background()

	outlet, err := s.outletRepo.GetByID(ctx, outletID)
	if err != nil {
		slog.Warn("sendPaymentConfirmationWA: ambil outlet gagal", "error", err)
		return
	}

	owner, err := s.userRepo.GetByIDInternal(ctx, outlet.OwnerID)
	if err != nil {
		slog.Warn("sendPaymentConfirmationWA: ambil owner gagal", "error", err)
		return
	}

	planStr := "Bulanan"
	if plan == model.SubscriptionPlanBiannual {
		planStr = "6 Bulan"
	}

	msg := fmt.Sprintf(
		"✅ Pembayaran berhasil!\n\nHalo %s, langganan Solekita outlet *%s* paket *%s* sudah aktif.\n\nAktif hingga: *%s*\n\nTerima kasih! 🙏",
		owner.Name, outlet.Name, planStr, nextDue.Format("02 Jan 2006"),
	)

	if err := s.fonnte.Send(ctx, owner.Phone, msg); err != nil {
		slog.Warn("sendPaymentConfirmationWA: kirim WA gagal", "error", err)
	}
}
