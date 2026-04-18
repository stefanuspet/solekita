package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stefanuspet/solekita/backend/internal/middleware"
	"github.com/stefanuspet/solekita/backend/internal/service"
)

type DashboardHandler struct {
	dashboardService *service.DashboardService
}

func NewDashboardHandler(dashboardService *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

// HandleGetSummary GET /v1/dashboard/summary?date=2006-01-02
func (h *DashboardHandler) HandleGetSummary(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	date := time.Now()
	if s := c.Query("date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			date = t
		}
	}

	res, err := h.dashboardService.GetSummary(c.Request.Context(), user.OutletID, date)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Summary harian berhasil diambil", res)
}

// HandleSendOverdueReminders POST /v1/dashboard/reminders/overdue
func (h *DashboardHandler) HandleSendOverdueReminders(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	res, err := h.dashboardService.SendOverdueReminders(c.Request.Context(), user.OutletID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Reminder berhasil dikirim", res)
}

// HandleGetStats GET /v1/dashboard/stats?period=7
func (h *DashboardHandler) HandleGetStats(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	period := 7
	if s := c.Query("period"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			period = n
		}
	}

	res, err := h.dashboardService.GetStats(c.Request.Context(), user.OutletID, period)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Statistik order berhasil diambil", res)
}
