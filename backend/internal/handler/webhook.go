package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stefanuspet/solekita/backend/internal/service"
)

type WebhookHandler struct {
	webhookService *service.WebhookService
}

func NewWebhookHandler(webhookService *service.WebhookService) *WebhookHandler {
	return &WebhookHandler{webhookService: webhookService}
}

// HandleTripayWebhook POST /v1/webhooks/tripay
func (h *WebhookHandler) HandleTripayWebhook(c *gin.Context) {
	// Baca raw body dulu — diperlukan untuk validasi signature sebelum di-parse
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		respondError(c, http.StatusBadRequest, "gagal membaca request body", nil)
		return
	}

	signature := c.GetHeader("X-Callback-Signature")

	if err := h.webhookService.HandleTripayPayment(c.Request.Context(), rawBody, signature); err != nil {
		handleServiceError(c, err)
		return
	}

	// Tripay mengharapkan response 200 dengan body tertentu
	c.JSON(http.StatusOK, gin.H{"success": true})
}
