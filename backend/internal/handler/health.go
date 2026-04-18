package handler

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db *sql.DB
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) HandleHealthCheck(c *gin.Context) {
	deps := map[string]string{
		"database":   h.checkDB(),
		"r2_storage": "unchecked",
		"fonnte":     "unchecked",
		"tripay":     "unchecked",
	}

	allHealthy := true
	for _, status := range deps {
		if status == "unhealthy" {
			allHealthy = false
			break
		}
	}

	if allHealthy {
		respondSuccess(c, http.StatusOK, "OK", gin.H{
			"status":       "healthy",
			"version":      "1.0.0",
			"timestamp":    time.Now().UTC().Format(time.RFC3339),
			"dependencies": deps,
		})
		return
	}

	c.JSON(http.StatusServiceUnavailable, APIErrorResponse{
		Success: false,
		Message: "Service degraded",
	})
}

func (h *HealthHandler) checkDB() string {
	if err := h.db.Ping(); err != nil {
		return "unhealthy"
	}
	return "healthy"
}
