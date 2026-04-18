package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stefanuspet/solekita/backend/internal/middleware"
	"github.com/stefanuspet/solekita/backend/internal/model"
	"github.com/stefanuspet/solekita/backend/internal/service"
)

type SubscriptionHandler struct {
	subscriptionService *service.SubscriptionService
}

func NewSubscriptionHandler(subscriptionService *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{subscriptionService: subscriptionService}
}

// HandleGetMySubscription GET /v1/subscriptions/me
func (h *SubscriptionHandler) HandleGetMySubscription(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	res, err := h.subscriptionService.GetMySubscription(c.Request.Context(), user.OutletID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Detail subscription berhasil diambil", res)
}

// HandleListInvoices GET /v1/subscriptions/me/invoices
func (h *SubscriptionHandler) HandleListInvoices(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	page, limit := parsePageLimit(c)

	res, err := h.subscriptionService.ListInvoices(c.Request.Context(), user.OutletID, page, limit)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Riwayat tagihan berhasil diambil", res)
}

// HandleConvertTrial POST /v1/subscriptions/me/convert
func (h *SubscriptionHandler) HandleConvertTrial(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	var req struct {
		Plan string `json:"plan" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "plan wajib diisi (monthly / biannual)", nil)
		return
	}

	plan := model.SubscriptionPlan(req.Plan)
	if plan != model.SubscriptionPlanMonthly && plan != model.SubscriptionPlanBiannual {
		respondError(c, http.StatusBadRequest, `plan tidak valid — gunakan "monthly" atau "biannual"`, nil)
		return
	}

	res, err := h.subscriptionService.ConvertTrial(c.Request.Context(), user.OutletID, plan)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusCreated, "Tagihan berhasil dibuat", res)
}
