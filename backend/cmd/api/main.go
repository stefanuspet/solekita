package main

import (
	"fmt"
	"log/slog"

	"github.com/stefanuspet/solekita/backend/internal/config"
	"github.com/stefanuspet/solekita/backend/internal/database"
	"github.com/stefanuspet/solekita/backend/internal/handler"
	"github.com/stefanuspet/solekita/backend/internal/notification"
	"github.com/stefanuspet/solekita/backend/internal/payment"
	"github.com/stefanuspet/solekita/backend/internal/repository"
	"github.com/stefanuspet/solekita/backend/internal/scheduler"
	"github.com/stefanuspet/solekita/backend/internal/service"
	"github.com/stefanuspet/solekita/backend/internal/storage"
)

func main() {
	cfg := config.Load()

	db := database.Connect(cfg)
	defer db.Close()

	database.RunMigrations(db, "migrations")

	// Repositories
	outletRepo := repository.NewOutletRepository(db)
	userRepo := repository.NewUserRepository(db)
	permRepo := repository.NewUserPermissionRepository(db)
	subRepo := repository.NewSubscriptionRepository(db)
	rtRepo := repository.NewRefreshTokenRepository(db)
	treatmentRepo := repository.NewTreatmentRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	photoRepo := repository.NewPhotoRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	deliveryRepo := repository.NewDeliveryRepository(db)
	logRepo := repository.NewOrderLogRepository(db)
	attendanceRepo := repository.NewAttendanceRepository(db)
	dashboardRepo := repository.NewDashboardRepository(db)
	invoiceRepo := repository.NewSubscriptionInvoiceRepository(db)
	adminRepo := repository.NewAdminRepository(db)

	// Notifications
	var fonnte *notification.FonnteClient
	if cfg.FonnteToken != "" {
		fonnte = notification.NewFonnte(cfg)
	}

	// Services
	authService := service.NewAuthService(db, outletRepo, userRepo, subRepo, rtRepo, cfg, fonnte)
	outletService := service.NewOutletService(outletRepo)
	userService := service.NewUserService(db, userRepo, permRepo)
	treatmentService := service.NewTreatmentService(treatmentRepo)
	customerService := service.NewCustomerService(customerRepo)

	var r2 *storage.R2Storage
	if cfg.R2AccountID != "" {
		if r2Initialized, err := storage.NewR2(cfg); err != nil {
			slog.Warn("R2 storage tidak bisa diinisialisasi", "error", err)
		} else {
			r2 = r2Initialized
		}
	}
	orderService := service.NewOrderService(db, orderRepo, photoRepo, paymentRepo, deliveryRepo, logRepo, treatmentRepo, customerRepo, r2)
	attendanceService := service.NewAttendanceService(attendanceRepo, r2)
	dashboardService := service.NewDashboardService(dashboardRepo, outletRepo, fonnte)
	deliveryService := service.NewDeliveryService(deliveryRepo, orderRepo, customerRepo, userRepo, permRepo, logRepo)

	var tripay service.TripayClient
	if cfg.TripayAPIKey != "" {
		tripay = payment.NewTripay(cfg)
	}
	subscriptionService := service.NewSubscriptionService(subRepo, invoiceRepo, outletRepo, userRepo, tripay)
	webhookService := service.NewWebhookService(db, invoiceRepo, subRepo, outletRepo, userRepo, fonnte, tripay)
	adminService := service.NewAdminService(adminRepo, outletRepo, subRepo, userRepo, fonnte)
	syncService := service.NewSyncService(db, orderRepo, photoRepo, paymentRepo, deliveryRepo, logRepo, treatmentRepo, customerRepo, r2)

	// Handlers
	handlers := &Handlers{
		Health:    handler.NewHealthHandler(db),
		Auth:      handler.NewAuthHandler(authService),
		Outlet:    handler.NewOutletHandler(outletService),
		User:      handler.NewUserHandler(userService),
		Treatment: handler.NewTreatmentHandler(treatmentService),
		Customer:  handler.NewCustomerHandler(customerService),
		Order:        handler.NewOrderHandler(orderService),
		Attendance:   handler.NewAttendanceHandler(attendanceService),
		Dashboard:    handler.NewDashboardHandler(dashboardService),
		Delivery:     handler.NewDeliveryHandler(deliveryService),
		Subscription: handler.NewSubscriptionHandler(subscriptionService),
		Webhook:      handler.NewWebhookHandler(webhookService),
		Admin:        handler.NewAdminHandler(adminService),
		Sync:         handler.NewSyncHandler(syncService),
	}

	r := setupRouter(cfg, handlers)

	// Scheduler
	sched := scheduler.New(subRepo, invoiceRepo, outletRepo, userRepo, subscriptionService, fonnte,
		photoRepo, attendanceRepo, r2, cfg)
	sched.Start()
	defer sched.Stop()

	addr := fmt.Sprintf(":%s", cfg.AppPort)
	slog.Info("Server running", "port", cfg.AppPort)
	if err := r.Run(addr); err != nil {
		slog.Error("Server gagal dijalankan", "error", err)
		panic(err)
	}
}
