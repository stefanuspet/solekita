package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stefanuspet/solekita/backend/internal/service"
)

type AdminHandler struct {
	adminService *service.AdminService
}

func NewAdminHandler(adminService *service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

// HandleListOutlets GET /v1/admin/outlets
// Query: status, trial_inactive (bool), search, page, limit
func (h *AdminHandler) HandleListOutlets(c *gin.Context) {
	filters := service.AdminListFilters{
		Search:        c.Query("search"),
		TrialInactive: c.Query("trial_inactive") == "true" || c.Query("trial_inactive") == "1",
	}
	if s := c.Query("status"); s != "" {
		filters.Status = &s
	}

	page, limit := parsePageLimit(c)

	res, err := h.adminService.ListOutlets(c.Request.Context(), filters, page, limit)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Daftar outlet berhasil diambil", res)
}

// HandleGetOutletDetail GET /v1/admin/outlets/:id
func (h *AdminHandler) HandleGetOutletDetail(c *gin.Context) {
	outletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "ID outlet tidak valid", nil)
		return
	}

	detail, err := h.adminService.GetOutletDetail(c.Request.Context(), outletID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Detail outlet berhasil diambil", detail)
}

// HandleUpdateOutlet PATCH /v1/admin/outlets/:id
// Body: {"subscription_status": "active", "owner_phone": "08123..."}
func (h *AdminHandler) HandleUpdateOutlet(c *gin.Context) {
	outletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "ID outlet tidak valid", nil)
		return
	}

	var req service.AdminUpdateOutletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Request tidak valid", nil)
		return
	}

	if req.SubscriptionStatus == nil && req.OwnerPhone == nil {
		respondError(c, http.StatusBadRequest, "Tidak ada field yang diperbarui", nil)
		return
	}

	if err := h.adminService.UpdateOutlet(c.Request.Context(), outletID, req); err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Outlet berhasil diperbarui", nil)
}

// HandleSuspendOutlet POST /v1/admin/outlets/:id/suspend
// Body: {"reason": "..."}
func (h *AdminHandler) HandleSuspendOutlet(c *gin.Context) {
	outletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "ID outlet tidak valid", nil)
		return
	}

	var body struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&body) // reason opsional

	if err := h.adminService.SuspendOutlet(c.Request.Context(), outletID, body.Reason); err != nil {
		handleServiceError(c, err)
		return
	}

	slog.Info("admin: outlet disuspend", "outlet_id", outletID, "reason", body.Reason)
	respondSuccess(c, http.StatusOK, "Outlet berhasil ditangguhkan", nil)
}

// HandleActivateOutlet POST /v1/admin/outlets/:id/activate
func (h *AdminHandler) HandleActivateOutlet(c *gin.Context) {
	outletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "ID outlet tidak valid", nil)
		return
	}

	if err := h.adminService.ActivateOutlet(c.Request.Context(), outletID); err != nil {
		handleServiceError(c, err)
		return
	}

	slog.Info("admin: outlet diaktifkan", "outlet_id", outletID)
	respondSuccess(c, http.StatusOK, "Outlet berhasil diaktifkan", nil)
}

// HandleGetStats GET /v1/admin/stats
func (h *AdminHandler) HandleGetStats(c *gin.Context) {
	stats, err := h.adminService.GetSummaryStats(c.Request.Context())
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Statistik berhasil diambil", stats)
}

// HandleSendFollowUpWA POST /v1/admin/outlets/:id/follow-up
func (h *AdminHandler) HandleSendFollowUpWA(c *gin.Context) {
	outletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "ID outlet tidak valid", nil)
		return
	}

	if err := h.adminService.SendFollowUpWA(c.Request.Context(), outletID); err != nil {
		handleServiceError(c, err)
		return
	}

	slog.Info("admin: follow-up WA terkirim", "outlet_id", outletID)
	respondSuccess(c, http.StatusOK, "WA follow-up berhasil dikirim", nil)
}
