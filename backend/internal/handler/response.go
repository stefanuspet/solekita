package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	apperrors "github.com/stefanuspet/solekita/backend/internal/errors"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type APIErrorResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors,omitempty"`
}

func respondSuccess(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func respondError(c *gin.Context, status int, message string, err error) {
	resp := APIErrorResponse{
		Success: false,
		Message: message,
	}

	var validationErr *apperrors.ValidationError
	if errors.As(err, &validationErr) {
		resp.Errors = validationErr.Fields
	}

	c.JSON(status, resp)
}

func parsePageLimit(c *gin.Context) (page, limit int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ = strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return
}

func handleServiceError(c *gin.Context, err error) {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		respondError(c, appErr.Code, appErr.Message, err)
		return
	}

	slog.Error("unexpected error", "error", err, "path", c.FullPath())
	respondError(c, http.StatusInternalServerError, "Terjadi kesalahan pada server", nil)
}
