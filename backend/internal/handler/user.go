package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stefanuspet/solekita/backend/internal/middleware"
	"github.com/stefanuspet/solekita/backend/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) HandleListUsers(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	var isActive *bool
	if val, ok := c.GetQuery("is_active"); ok {
		b := val == "true"
		isActive = &b
	}

	res, err := h.userService.ListUsers(c.Request.Context(), user.OutletID, isActive)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Daftar user berhasil diambil", res)
}

func (h *UserHandler) HandleCreateUser(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	var req service.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Validasi gagal", err)
		return
	}

	res, err := h.userService.CreateUser(c.Request.Context(), user.OutletID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusCreated, "User berhasil dibuat", res)
}

func (h *UserHandler) HandleGetUser(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	res, err := h.userService.GetUser(c.Request.Context(), targetID, user.OutletID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Data user berhasil diambil", res)
}

func (h *UserHandler) HandleUpdateUser(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	var req service.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Validasi gagal", err)
		return
	}

	res, err := h.userService.UpdateUser(c.Request.Context(), targetID, user.OutletID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "User berhasil diperbarui", res)
}

func (h *UserHandler) HandleResetPassword(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	res, err := h.userService.ResetPassword(c.Request.Context(), targetID, user.OutletID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Password berhasil direset", res)
}
