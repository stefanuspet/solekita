package scheduler

import (
	"log/slog"

	"github.com/robfig/cron/v3"
	appcfg "github.com/stefanuspet/solekita/backend/internal/config"
	"github.com/stefanuspet/solekita/backend/internal/notification"
	"github.com/stefanuspet/solekita/backend/internal/repository"
	"github.com/stefanuspet/solekita/backend/internal/service"
	"github.com/stefanuspet/solekita/backend/internal/storage"
)

type Scheduler struct {
	c                   *cron.Cron
	subRepo             *repository.SubscriptionRepository
	invoiceRepo         *repository.SubscriptionInvoiceRepository
	outletRepo          *repository.OutletRepository
	userRepo            *repository.UserRepository
	subscriptionService *service.SubscriptionService
	fonnte              *notification.FonnteClient
	photoRepo           *repository.PhotoRepository
	attendanceRepo      *repository.AttendanceRepository
	r2                  *storage.R2Storage
	cfg                 *appcfg.Config
}

func New(
	subRepo *repository.SubscriptionRepository,
	invoiceRepo *repository.SubscriptionInvoiceRepository,
	outletRepo *repository.OutletRepository,
	userRepo *repository.UserRepository,
	subscriptionService *service.SubscriptionService,
	fonnte *notification.FonnteClient,
	photoRepo *repository.PhotoRepository,
	attendanceRepo *repository.AttendanceRepository,
	r2 *storage.R2Storage,
	cfg *appcfg.Config,
) *Scheduler {
	return &Scheduler{
		c:                   cron.New(),
		subRepo:             subRepo,
		invoiceRepo:         invoiceRepo,
		outletRepo:          outletRepo,
		userRepo:            userRepo,
		subscriptionService: subscriptionService,
		fonnte:              fonnte,
		photoRepo:           photoRepo,
		attendanceRepo:      attendanceRepo,
		r2:                  r2,
		cfg:                 cfg,
	}
}

// Start mendaftarkan semua cron job dan menjalankan scheduler.
func (s *Scheduler) Start() {
	// ── Trial reminders ───────────────────────────────────────────────────────
	s.c.AddFunc("0 9 * * *", s.SendTrialReminderH3)   // WA #2 — H-3, 09:00
	s.c.AddFunc("0 9 * * *", s.SendTrialReminderH1)   // WA #3 — H-1, 09:00
	s.c.AddFunc("0 9 * * *", s.SendTrialExpiredNotif) // WA #4 — H-0, 09:00
	s.c.AddFunc("0 0 * * *", s.SuspendExpiredTrials)  // suspend trial habis, 00:00

	// ── Billing ───────────────────────────────────────────────────────────────
	s.c.AddFunc("0 8 * * *", s.GenerateMonthlyInvoices) // WA #8 — H-0, generate + kirim, 08:00
	s.c.AddFunc("0 9 * * *", s.SendBillingReminderH3)   // WA #6 — H-3, 09:00
	s.c.AddFunc("0 9 * * *", s.SendBillingReminderH1)   // WA #7 — H-1, 09:00
	s.c.AddFunc("0 0 * * *", s.SuspendUnpaidOutlets)    // WA #9 — suspend grace habis, 00:00
	s.c.AddFunc("0 0 * * *", s.MarkInactiveOutlets)     // tandai inactive > 30 hari, 00:00

	// ── Maintenance ───────────────────────────────────────────────────────────
	s.c.AddFunc("0 2 * * *", s.CleanupExpiredPhotos)  // 02:00 setiap hari
	s.c.AddFunc("0 2 * * *", s.CleanupExpiredSelfies) // 02:00 setiap hari
	s.c.AddFunc("0 4 * * *", s.BackupDatabase)        // 04:00 setiap hari

	s.c.Start()
	slog.Info("Scheduler started")
}

// Stop menghentikan scheduler dengan graceful shutdown.
func (s *Scheduler) Stop() {
	ctx := s.c.Stop()
	<-ctx.Done()
	slog.Info("Scheduler stopped")
}
