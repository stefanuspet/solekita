package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/stefanuspet/solekita/backend/internal/middleware"
	"github.com/stefanuspet/solekita/backend/internal/service"
)

type SyncHandler struct {
	syncService *service.SyncService
}

func NewSyncHandler(syncService *service.SyncService) *SyncHandler {
	return &SyncHandler{syncService: syncService}
}

// HandleSync POST /v1/sync
//
// Multipart form:
//   - orders  (text field) — JSON array of SyncOrderItem
//   - photo_before_{local_id} (file field, opsional) — foto before per order
//
// Response:
//   - 200 jika semua order berhasil
//   - 207 jika sebagian berhasil, sebagian gagal
//   - 400 jika tidak ada order sama sekali / JSON tidak valid
func (h *SyncHandler) HandleSync(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	// 1. Parse multipart (maks 32 MB total)
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		respondError(c, http.StatusBadRequest, "Gagal membaca multipart form", nil)
		return
	}

	// 2. Bind field "orders" (JSON array)
	ordersJSON := c.Request.FormValue("orders")
	if ordersJSON == "" {
		respondError(c, http.StatusBadRequest, "Field 'orders' wajib diisi", nil)
		return
	}

	var items []service.SyncOrderItem
	if err := json.Unmarshal([]byte(ordersJSON), &items); err != nil {
		respondError(c, http.StatusBadRequest, "Format orders tidak valid (harus JSON array)", nil)
		return
	}

	if len(items) == 0 {
		respondError(c, http.StatusBadRequest, "Tidak ada order untuk disinkronkan", nil)
		return
	}

	// 3. Kumpulkan foto berdasarkan local_id
	// Flutter mengirim file dengan key: photo_before_{local_id}
	photos := make(map[string][]byte)
	if mf := c.Request.MultipartForm; mf != nil {
		for key, fileHeaders := range mf.File {
			if !strings.HasPrefix(key, "photo_before_") {
				continue
			}
			localID := strings.TrimPrefix(key, "photo_before_")
			if len(fileHeaders) == 0 {
				continue
			}
			f, err := fileHeaders[0].Open()
			if err != nil {
				continue
			}
			data, err := io.ReadAll(f)
			f.Close()
			if err != nil || len(data) == 0 {
				continue
			}
			photos[localID] = data
		}
	}

	// 4. Proses sinkronisasi
	result, err := h.syncService.SyncOrders(c.Request.Context(), user, items, photos)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// 5. Tentukan status kode respons
	//    200 — semua berhasil
	//    207 — sebagian berhasil (partial success)
	status := http.StatusOK
	if len(result.Failed) > 0 {
		status = http.StatusMultiStatus
	}

	message := fmt.Sprintf("%d order berhasil disinkronkan", len(result.Synced))
	if len(result.Failed) > 0 {
		message = fmt.Sprintf("%d berhasil, %d gagal", len(result.Synced), len(result.Failed))
	}

	c.JSON(status, APIResponse{
		Success: len(result.Failed) == 0,
		Message: message,
		Data:    result,
	})
}
