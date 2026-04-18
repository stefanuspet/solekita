package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	apperrors "github.com/stefanuspet/solekita/backend/internal/errors"
	"github.com/stefanuspet/solekita/backend/internal/middleware"
	"github.com/stefanuspet/solekita/backend/internal/model"
	"github.com/stefanuspet/solekita/backend/internal/service"
)

type OrderHandler struct {
	orderService *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

func (h *OrderHandler) HandleListOrders(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	filters := service.ListOrdersFilters{}

	if s := c.Query("status"); s != "" {
		status := model.OrderStatus(s)
		filters.Status = &status
	}
	if s := c.Query("kasir_id"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			filters.KasirID = &id
		}
	}
	if s := c.Query("treatment_id"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			filters.TreatmentID = &id
		}
	}
	if s := c.Query("date_from"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			filters.DateFrom = &t
		}
	}
	if s := c.Query("date_to"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			// Inklusif sampai akhir hari
			end := t.Add(24*time.Hour - time.Second)
			filters.DateTo = &end
		}
	}
	filters.Search = c.Query("search")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	res, err := h.orderService.ListOrders(c.Request.Context(), user, filters, page, limit)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Daftar order berhasil diambil", res)
}

func (h *OrderHandler) HandleGetOrder(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	res, err := h.orderService.GetOrder(c.Request.Context(), orderID, user.OutletID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Detail order berhasil diambil", res)
}

func (h *OrderHandler) HandleUpdateOrderStatus(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Validasi gagal", err)
		return
	}

	res, err := h.orderService.UpdateOrderStatus(c.Request.Context(), user, orderID, user.OutletID, model.OrderStatus(req.Status))
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Status order berhasil diperbarui", res)
}

func (h *OrderHandler) HandleCancelOrder(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Validasi gagal", err)
		return
	}

	res, err := h.orderService.CancelOrder(c.Request.Context(), user, orderID, user.OutletID, req.Reason)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Order berhasil dibatalkan", res)
}

func (h *OrderHandler) HandleEditOrderPrice(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	var req service.EditOrderPriceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Validasi gagal", err)
		return
	}

	res, err := h.orderService.EditOrderPrice(c.Request.Context(), user, orderID, user.OutletID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Harga order berhasil diperbarui", res)
}

func (h *OrderHandler) HandleGetOrderPhotos(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	res, err := h.orderService.GetOrderPhotos(c.Request.Context(), orderID, user.OutletID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Foto order berhasil diambil", res)
}

func (h *OrderHandler) HandleUploadAfterPhoto(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	file, _, err := c.Request.FormFile("photo_after")
	if err != nil {
		respondError(c, http.StatusBadRequest, "Validasi gagal",
			apperrors.ErrBadRequest.New(err))
		return
	}
	defer file.Close()

	res, err := h.orderService.UploadAfterPhoto(c.Request.Context(), user, orderID, user.OutletID, file)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusCreated, "Foto after berhasil diupload", res)
}

func (h *OrderHandler) HandleCreateOrder(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	var req service.CreateOrderRequest
	if err := c.ShouldBind(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Validasi gagal", err)
		return
	}

	// Ambil foto before — wajib
	beforeFile, _, err := c.Request.FormFile("photo_before")
	if err != nil {
		respondError(c, http.StatusBadRequest, "Validasi gagal",
			apperrors.ErrBadRequest.New(err))
		return
	}
	defer beforeFile.Close()

	res, err := h.orderService.CreateOrder(c.Request.Context(), user, req, beforeFile)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusCreated, "Order berhasil dibuat", res)
}
