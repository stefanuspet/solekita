package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stefanuspet/solekita/backend/internal/middleware"
	"github.com/stefanuspet/solekita/backend/internal/service"
)

type CustomerHandler struct {
	customerService *service.CustomerService
}

func NewCustomerHandler(customerService *service.CustomerService) *CustomerHandler {
	return &CustomerHandler{customerService: customerService}
}

func (h *CustomerHandler) HandleFindOrCreate(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	var req service.FindOrCreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Validasi gagal", err)
		return
	}

	res, err := h.customerService.FindOrCreate(c.Request.Context(), user.OutletID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	status := http.StatusOK
	message := "Pelanggan ditemukan"
	if res.IsNew {
		status = http.StatusCreated
		message = "Pelanggan baru berhasil dibuat"
	}

	respondSuccess(c, status, message, res)
}

func (h *CustomerHandler) HandleListCustomers(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	search := c.Query("search")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	res, err := h.customerService.ListCustomers(c.Request.Context(), user.OutletID, search, page, limit)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Daftar pelanggan berhasil diambil", res)
}

func (h *CustomerHandler) HandleGetCustomer(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	customerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	res, err := h.customerService.GetCustomer(c.Request.Context(), customerID, user.OutletID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Data pelanggan berhasil diambil", res)
}

func (h *CustomerHandler) HandleUpdateCustomer(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	customerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	var req service.UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Validasi gagal", err)
		return
	}

	res, err := h.customerService.UpdateCustomer(c.Request.Context(), customerID, user.OutletID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Data pelanggan berhasil diperbarui", res)
}
