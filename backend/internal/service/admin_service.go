package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	apperrors "github.com/stefanuspet/solekita/backend/internal/errors"
	"github.com/stefanuspet/solekita/backend/internal/model"
	"github.com/stefanuspet/solekita/backend/internal/notification"
	"github.com/stefanuspet/solekita/backend/internal/repository"
)

type AdminService struct {
	adminRepo  *repository.AdminRepository
	outletRepo *repository.OutletRepository
	subRepo    *repository.SubscriptionRepository
	userRepo   *repository.UserRepository
	fonnte     *notification.FonnteClient
}

func NewAdminService(
	adminRepo *repository.AdminRepository,
	outletRepo *repository.OutletRepository,
	subRepo *repository.SubscriptionRepository,
	userRepo *repository.UserRepository,
	fonnte *notification.FonnteClient,
) *AdminService {
	return &AdminService{
		adminRepo:  adminRepo,
		outletRepo: outletRepo,
		subRepo:    subRepo,
		userRepo:   userRepo,
		fonnte:     fonnte,
	}
}

// ── Request types ─────────────────────────────────────────────────────────────

type AdminListFilters struct {
	Status        *string `form:"status"`         // filter by subscription_status
	TrialInactive bool    `form:"trial_inactive"` // trial outlets yang belum buat order
	Search        string  `form:"search"`         // cari by nama outlet / HP owner
}

type AdminUpdateOutletRequest struct {
	SubscriptionStatus *string `json:"subscription_status"` // override status
	OwnerPhone         *string `json:"owner_phone"`         // koreksi nomor HP owner
}

// ── Response types ────────────────────────────────────────────────────────────

type AdminListResponse struct {
	Outlets []*repository.AdminOutletRow `json:"outlets"`
	Total   int                          `json:"total"`
	Page    int                          `json:"page"`
	Limit   int                          `json:"limit"`
}

// ── ListOutlets ───────────────────────────────────────────────────────────────

func (s *AdminService) ListOutlets(ctx context.Context, f AdminListFilters, page, limit int) (*AdminListResponse, error) {
	repoFilters := repository.AdminOutletFilters{
		TrialInactive: f.TrialInactive,
		Search:        f.Search,
	}
	if f.Status != nil {
		status := model.SubscriptionStatus(*f.Status)
		repoFilters.Status = &status
	}

	outlets, total, err := s.adminRepo.ListOutlets(ctx, repoFilters, page, limit)
	if err != nil {
		return nil, fmt.Errorf("ListOutlets: %w", err)
	}

	return &AdminListResponse{
		Outlets: outlets,
		Total:   total,
		Page:    page,
		Limit:   limit,
	}, nil
}

// ── GetOutletDetail ───────────────────────────────────────────────────────────

func (s *AdminService) GetOutletDetail(ctx context.Context, outletID uuid.UUID) (*repository.AdminOutletDetail, error) {
	detail, err := s.adminRepo.GetOutletDetail(ctx, outletID)
	if err != nil {
		return nil, fmt.Errorf("GetOutletDetail: %w", err)
	}
	return detail, nil
}

// ── GetSummaryStats ───────────────────────────────────────────────────────────

func (s *AdminService) GetSummaryStats(ctx context.Context) (*repository.AdminSummaryStats, error) {
	stats, err := s.adminRepo.GetOutletSummaryStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetSummaryStats: %w", err)
	}
	return stats, nil
}

// ── UpdateOutlet ──────────────────────────────────────────────────────────────

// UpdateOutlet memperbarui subscription_status dan/atau owner_phone outlet.
// Gunakan untuk koreksi manual oleh admin — tidak melalui alur pembayaran normal.
func (s *AdminService) UpdateOutlet(ctx context.Context, outletID uuid.UUID, req AdminUpdateOutletRequest) error {
	outlet, err := s.outletRepo.GetByID(ctx, outletID)
	if err != nil {
		return fmt.Errorf("UpdateOutlet: ambil outlet: %w", err)
	}

	if req.SubscriptionStatus != nil {
		newStatus := model.SubscriptionStatus(*req.SubscriptionStatus)
		switch newStatus {
		case model.SubscriptionStatusTrial,
			model.SubscriptionStatusActive,
			model.SubscriptionStatusSuspended,
			model.SubscriptionStatusInactive,
			model.SubscriptionStatusCancelled:
			// valid
		default:
			return apperrors.ErrBadRequest.New(fmt.Errorf("status tidak valid: %s", newStatus))
		}
		if err := s.outletRepo.UpdateSubscriptionStatus(ctx, outletID, newStatus); err != nil {
			return fmt.Errorf("UpdateOutlet: update status: %w", err)
		}
	}

	if req.OwnerPhone != nil {
		if err := s.userRepo.UpdatePhone(ctx, outlet.OwnerID, *req.OwnerPhone); err != nil {
			return fmt.Errorf("UpdateOutlet: update phone owner: %w", err)
		}
	}

	return nil
}

// ── SuspendOutlet ─────────────────────────────────────────────────────────────

// SuspendOutlet menangguhkan akses outlet secara manual oleh admin.
// reason dicatat di log untuk audit trail — tidak tersimpan di DB karena belum ada kolom.
func (s *AdminService) SuspendOutlet(ctx context.Context, outletID uuid.UUID, reason string) error {
	outlet, err := s.outletRepo.GetByID(ctx, outletID)
	if err != nil {
		return fmt.Errorf("SuspendOutlet: ambil outlet: %w", err)
	}

	if outlet.SubscriptionStatus == model.SubscriptionStatusSuspended {
		return apperrors.ErrUnprocessable.New(fmt.Errorf("outlet sudah dalam status suspended"))
	}

	if err := s.outletRepo.UpdateSubscriptionStatus(ctx, outletID, model.SubscriptionStatusSuspended); err != nil {
		return fmt.Errorf("SuspendOutlet: update status: %w", err)
	}

	if err := s.subRepo.UpdateSuspendedAt(ctx, outletID); err != nil {
		// non-fatal — status utama sudah di-update
		_ = err
	}

	// reason dicatat di caller (handler) via slog — tidak ada kolom DB untuk ini
	_ = reason

	return nil
}

// ── ActivateOutlet ────────────────────────────────────────────────────────────

// ActivateOutlet mengaktifkan kembali outlet yang suspended atau inactive.
// Digunakan admin setelah masalah diselesaikan (misal: pembayaran dikonfirmasi manual).
func (s *AdminService) ActivateOutlet(ctx context.Context, outletID uuid.UUID) error {
	outlet, err := s.outletRepo.GetByID(ctx, outletID)
	if err != nil {
		return fmt.Errorf("ActivateOutlet: ambil outlet: %w", err)
	}

	switch outlet.SubscriptionStatus {
	case model.SubscriptionStatusSuspended, model.SubscriptionStatusInactive:
		// valid untuk diaktifkan
	case model.SubscriptionStatusActive:
		return apperrors.ErrUnprocessable.New(fmt.Errorf("outlet sudah aktif"))
	default:
		return apperrors.ErrUnprocessable.New(fmt.Errorf("outlet berstatus %s, tidak bisa langsung diaktifkan", outlet.SubscriptionStatus))
	}

	if err := s.outletRepo.UpdateSubscriptionStatus(ctx, outletID, model.SubscriptionStatusActive); err != nil {
		return fmt.Errorf("ActivateOutlet: update status: %w", err)
	}

	if err := s.subRepo.ClearSuspendedAt(ctx, outletID); err != nil {
		_ = err // non-fatal
	}

	return nil
}

// ── SendFollowUpWA ────────────────────────────────────────────────────────────

// SendFollowUpWA mengirim WA follow-up manual dari admin ke owner outlet.
// Pesan disesuaikan berdasarkan status outlet: trial yang belum aktif vs sudah aktif tapi belum bayar.
func (s *AdminService) SendFollowUpWA(ctx context.Context, outletID uuid.UUID) error {
	if s.fonnte == nil {
		return apperrors.ErrUnprocessable.New(fmt.Errorf("notifikasi WhatsApp belum dikonfigurasi"))
	}

	outlet, err := s.outletRepo.GetByID(ctx, outletID)
	if err != nil {
		return fmt.Errorf("SendFollowUpWA: ambil outlet: %w", err)
	}

	owner, err := s.userRepo.GetByIDInternal(ctx, outlet.OwnerID)
	if err != nil {
		return fmt.Errorf("SendFollowUpWA: ambil owner: %w", err)
	}

	msg := followUpMessage(outlet.Name, outlet.SubscriptionStatus)

	if err := s.fonnte.Send(ctx, owner.Phone, msg); err != nil {
		return fmt.Errorf("SendFollowUpWA: kirim WA: %w", err)
	}

	return nil
}

// followUpMessage memilih template WA berdasarkan status outlet.
// Sesuai onboarding-flow.md Section 3 — Founder Follow-Up Playbook.
func followUpMessage(outletName string, status model.SubscriptionStatus) string {
	switch status {
	case model.SubscriptionStatusTrial:
		// Outlet trial yang mungkin belum sempat mencoba
		return fmt.Sprintf(
			"Halo %s 👋\n\n"+
				"Saya dari tim Solekita.\n\n"+
				"Saya lihat akun Anda sudah terdaftar tapi belum sempat mencoba membuat order. "+
				"Ada kendala saat setup?\n\n"+
				"Saya siap bantu via WA jika ada yang membingungkan. "+
				"Biasanya butuh 10 menit aja untuk mulai 🙏",
			outletName,
		)
	case model.SubscriptionStatusSuspended, model.SubscriptionStatusInactive:
		// Outlet yang sudah pernah pakai tapi belum perpanjang
		return fmt.Sprintf(
			"Halo %s 👋\n\n"+
				"Saya dari tim Solekita.\n\n"+
				"Saya lihat Anda sudah aktif pakai selama trial — terima kasih sudah coba! 🙏\n\n"+
				"Ada kendala saat mau lanjut berlangganan? "+
				"Kalau ada pertanyaan soal pembayaran, saya siap bantu.\n\n"+
				"Buka aplikasi > Langganan untuk lanjutkan.",
			outletName,
		)
	default:
		// Fallback umum
		return fmt.Sprintf(
			"Halo %s 👋\n\n"+
				"Saya dari tim Solekita. Ada yang bisa dibantu?\n\n"+
				"Balas WA ini jika ada pertanyaan 🙏",
			outletName,
		)
	}
}
