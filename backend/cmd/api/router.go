package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stefanuspet/solekita/backend/internal/config"
	"github.com/stefanuspet/solekita/backend/internal/handler"
	"github.com/stefanuspet/solekita/backend/internal/middleware"
	"github.com/stefanuspet/solekita/backend/internal/model"
)

type Handlers struct {
	Health       *handler.HealthHandler
	Auth         *handler.AuthHandler
	Outlet       *handler.OutletHandler
	User         *handler.UserHandler
	Treatment    *handler.TreatmentHandler
	Customer     *handler.CustomerHandler
	Order        *handler.OrderHandler
	Attendance   *handler.AttendanceHandler
	Dashboard    *handler.DashboardHandler
	Delivery     *handler.DeliveryHandler
	Subscription *handler.SubscriptionHandler
	Webhook      *handler.WebhookHandler
	Admin        *handler.AdminHandler
	Sync         *handler.SyncHandler
}

func setupRouter(cfg *config.Config, h *Handlers) *gin.Engine {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	// notImplemented adalah placeholder untuk endpoint yang belum diimplementasi
	notImplemented := func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{
			"success": false,
			"message": "Belum diimplementasi",
		})
	}

	v1 := r.Group("/v1")

	// ── Public ────────────────────────────────────────────────────────────────
	v1.GET("/health", h.Health.HandleHealthCheck)
	v1.POST("/webhooks/tripay", h.Webhook.HandleTripayWebhook)
	v1.POST("/auth/register", middleware.RateLimit(3, time.Minute), h.Auth.HandleRegister)
	v1.POST("/auth/login", middleware.RateLimit(5, time.Minute), h.Auth.HandleLogin)
	v1.POST("/auth/refresh", h.Auth.HandleRefresh)

	// ── Authenticated ─────────────────────────────────────────────────────────
	authed := v1.Group("/")
	authed.Use(middleware.Auth(cfg.JWTSecret))
	{
		authed.POST("/auth/logout", h.Auth.HandleLogout)

		// Outlet
		authed.GET("/outlets/me", h.Outlet.HandleGetMyOutlet)
		authed.PATCH("/outlets/me", middleware.RequirePermission(model.PermissionManageOutlet), h.Outlet.HandleUpdateMyOutlet)

		// Users (karyawan)
		authed.GET("/users", middleware.RequireOwner(), h.User.HandleListUsers)
		authed.POST("/users", middleware.RequirePermission(model.PermissionManageOutlet), h.User.HandleCreateUser)
		authed.GET("/users/:id", middleware.RequirePermission(model.PermissionManageOutlet), h.User.HandleGetUser)
		authed.PATCH("/users/:id", middleware.RequirePermission(model.PermissionManageOutlet), h.User.HandleUpdateUser)
		authed.POST("/users/:id/reset-password", middleware.RequirePermission(model.PermissionManageOutlet), h.User.HandleResetPassword)

		// Treatments
		authed.GET("/treatments", h.Treatment.HandleListTreatments)
		authed.POST("/treatments", middleware.RequirePermission(model.PermissionManageOutlet), h.Treatment.HandleCreateTreatment)
		authed.PATCH("/treatments/:id", middleware.RequirePermission(model.PermissionManageOutlet), h.Treatment.HandleUpdateTreatment)
		authed.DELETE("/treatments/:id", middleware.RequirePermission(model.PermissionManageOutlet), h.Treatment.HandleDeleteTreatment)

		// Customers
		authed.GET("/customers", middleware.RequirePermission(model.PermissionManageOrder, model.PermissionViewReport), h.Customer.HandleListCustomers)
		authed.POST("/customers", middleware.RequirePermission(model.PermissionManageOrder), h.Customer.HandleFindOrCreate)
		authed.GET("/customers/:id", middleware.RequirePermission(model.PermissionManageOrder, model.PermissionViewReport), h.Customer.HandleGetCustomer)
		authed.PATCH("/customers/:id", middleware.RequirePermission(model.PermissionManageOrder), h.Customer.HandleUpdateCustomer)

		// Orders
		authed.GET("/orders", middleware.RequirePermission(model.PermissionManageOrder, model.PermissionViewReport), h.Order.HandleListOrders)
		authed.POST("/orders", middleware.RequirePermission(model.PermissionManageOrder), h.Order.HandleCreateOrder)
		authed.GET("/orders/:id", middleware.RequirePermission(model.PermissionManageOrder, model.PermissionViewReport), h.Order.HandleGetOrder)
		authed.PATCH("/orders/:id/status", middleware.RequirePermission(model.PermissionUpdateStatus), h.Order.HandleUpdateOrderStatus)
		authed.POST("/orders/:id/cancel", middleware.RequirePermission(model.PermissionManageOrder), h.Order.HandleCancelOrder)
		authed.PATCH("/orders/:id/price", middleware.RequirePermission(model.PermissionManageOrder), h.Order.HandleEditOrderPrice)

		// Photos
		authed.GET("/orders/:id/photos", middleware.RequirePermission(model.PermissionManageOrder, model.PermissionViewReport), h.Order.HandleGetOrderPhotos)
		authed.POST("/orders/:id/photos", middleware.RequirePermission(model.PermissionManageOrder), h.Order.HandleUploadAfterPhoto)

		// Payments
		authed.GET("/orders/:id/payment", middleware.RequirePermission(model.PermissionManageOrder, model.PermissionViewReport), notImplemented)
		authed.PATCH("/orders/:id/payment", middleware.RequirePermission(model.PermissionManageOrder), notImplemented)

		// Deliveries
		authed.GET("/deliveries", middleware.RequirePermission(model.PermissionManageDelivery), h.Delivery.HandleListDeliveries)
		authed.PATCH("/deliveries/:id/pickup-status", middleware.RequirePermission(model.PermissionManageDelivery), h.Delivery.HandleUpdatePickupStatus)
		authed.PATCH("/deliveries/:id/delivery-status", middleware.RequirePermission(model.PermissionManageDelivery), h.Delivery.HandleUpdateDeliveryStatus)
		authed.PATCH("/deliveries/:id/assign", middleware.RequirePermission(model.PermissionManageOutlet), h.Delivery.HandleAssignCourier)

		// Attendances
		authed.POST("/attendances", h.Attendance.HandleAttendance)
		authed.GET("/attendances/today", h.Attendance.HandleGetToday)
		authed.GET("/attendances", middleware.RequirePermission(model.PermissionManageOutlet, model.PermissionViewReport), h.Attendance.HandleList)

		// Dashboard
		authed.GET("/outlets/me/summary", middleware.RequirePermission(model.PermissionViewReport), h.Dashboard.HandleGetSummary)
		authed.GET("/outlets/me/stats", middleware.RequirePermission(model.PermissionViewReport), h.Dashboard.HandleGetStats)
		authed.POST("/outlets/me/reminders/overdue", middleware.RequirePermission(model.PermissionViewReport), h.Dashboard.HandleSendOverdueReminders)

		// Subscription
		authed.GET("/subscriptions/me", middleware.RequireOwner(), h.Subscription.HandleGetMySubscription)
		authed.POST("/subscriptions/me/convert", middleware.RequireOwner(), h.Subscription.HandleConvertTrial)
		authed.GET("/subscriptions/me/invoices", middleware.RequireOwner(), h.Subscription.HandleListInvoices)
	}

	// ── Admin Panel ───────────────────────────────────────────────────────────
	admin := v1.Group("/admin")
	admin.Use(middleware.AdminAuth(cfg.AdminSecretKey))
	{
		admin.GET("/outlets", h.Admin.HandleListOutlets)
		admin.GET("/outlets/:id", h.Admin.HandleGetOutletDetail)
		admin.PATCH("/outlets/:id", h.Admin.HandleUpdateOutlet)
		admin.POST("/outlets/:id/suspend", h.Admin.HandleSuspendOutlet)
		admin.POST("/outlets/:id/activate", h.Admin.HandleActivateOutlet)
		admin.POST("/outlets/:id/follow-up", h.Admin.HandleSendFollowUpWA)
		admin.GET("/stats", h.Admin.HandleGetStats)
	}

	return r
}
