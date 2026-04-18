package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stefanuspet/solekita/backend/internal/middleware"
	"github.com/stefanuspet/solekita/backend/internal/model"
	"github.com/stefanuspet/solekita/backend/internal/service"
)

type DeliveryHandler struct {
	deliveryService *service.DeliveryService
}

func NewDeliveryHandler(deliveryService *service.DeliveryService) *DeliveryHandler {
	return &DeliveryHandler{deliveryService: deliveryService}
}

// HandleListDeliveries GET /v1/deliveries
func (h *DeliveryHandler) HandleListDeliveries(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	deliveryType := c.Query("type")
	status := c.Query("status")

	var date *time.Time
	if s := c.Query("date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			date = &t
		}
	}

	res, err := h.deliveryService.ListMyDeliveries(c.Request.Context(), user, deliveryType, status, date)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Daftar delivery berhasil diambil", res)
}

// HandleAssignCourier PATCH /v1/deliveries/:id/assign
func (h *DeliveryHandler) HandleAssignCourier(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	deliveryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "delivery ID tidak valid", nil)
		return
	}

	var req struct {
		CourierID uuid.UUID `json:"courier_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "courier_id wajib diisi", nil)
		return
	}

	res, err := h.deliveryService.AssignCourier(c.Request.Context(), deliveryID, req.CourierID, user.OutletID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Kurir berhasil di-assign", res)
}

// HandleUpdatePickupStatus PATCH /v1/deliveries/:id/pickup-status
func (h *DeliveryHandler) HandleUpdatePickupStatus(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	deliveryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "delivery ID tidak valid", nil)
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
		Notes  string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "status wajib diisi", nil)
		return
	}

	res, err := h.deliveryService.UpdatePickupStatus(
		c.Request.Context(), user, deliveryID,
		model.PickupStatus(req.Status), req.Notes,
	)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Status pickup berhasil diperbarui", res)
}

// HandleUpdateDeliveryStatus PATCH /v1/deliveries/:id/delivery-status
func (h *DeliveryHandler) HandleUpdateDeliveryStatus(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	deliveryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "delivery ID tidak valid", nil)
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
		Notes  string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "status wajib diisi", nil)
		return
	}

	res, err := h.deliveryService.UpdateDeliveryStatus(
		c.Request.Context(), user, deliveryID,
		model.DeliveryStatus(req.Status), req.Notes,
	)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Status delivery berhasil diperbarui", res)
}
