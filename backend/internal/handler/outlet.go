package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stefanuspet/solekita/backend/internal/middleware"
	"github.com/stefanuspet/solekita/backend/internal/service"
)

type OutletHandler struct {
	outletService *service.OutletService
}

func NewOutletHandler(outletService *service.OutletService) *OutletHandler {
	return &OutletHandler{outletService: outletService}
}

func (h *OutletHandler) HandleGetMyOutlet(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	res, err := h.outletService.GetMyOutlet(c.Request.Context(), user.OutletID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Data outlet berhasil diambil", res)
}

func (h *OutletHandler) HandleUpdateMyOutlet(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	var req service.UpdateOutletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Validasi gagal", err)
		return
	}

	res, err := h.outletService.UpdateOutlet(c.Request.Context(), user.OutletID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Data outlet berhasil diperbarui", res)
}
