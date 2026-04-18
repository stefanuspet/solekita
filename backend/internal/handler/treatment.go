package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stefanuspet/solekita/backend/internal/middleware"
	"github.com/stefanuspet/solekita/backend/internal/service"
)

type TreatmentHandler struct {
	treatmentService *service.TreatmentService
}

func NewTreatmentHandler(treatmentService *service.TreatmentService) *TreatmentHandler {
	return &TreatmentHandler{treatmentService: treatmentService}
}

func (h *TreatmentHandler) HandleListTreatments(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	var isActive *bool
	if val, ok := c.GetQuery("is_active"); ok {
		b := val == "true"
		isActive = &b
	}

	var material *string
	if val, ok := c.GetQuery("material"); ok && val != "" {
		material = &val
	}

	res, err := h.treatmentService.ListTreatments(c.Request.Context(), user.OutletID, isActive, material)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Daftar treatment berhasil diambil", res)
}

func (h *TreatmentHandler) HandleCreateTreatment(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	var req service.CreateTreatmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Validasi gagal", err)
		return
	}

	res, err := h.treatmentService.CreateTreatment(c.Request.Context(), user.OutletID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusCreated, "Treatment berhasil dibuat", res)
}

func (h *TreatmentHandler) HandleUpdateTreatment(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	treatmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	var req service.UpdateTreatmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Validasi gagal", err)
		return
	}

	res, err := h.treatmentService.UpdateTreatment(c.Request.Context(), treatmentID, user.OutletID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Treatment berhasil diperbarui", res)
}

func (h *TreatmentHandler) HandleDeleteTreatment(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	treatmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	if err := h.treatmentService.DeleteTreatment(c.Request.Context(), treatmentID, user.OutletID); err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Treatment berhasil dihapus", nil)
}
