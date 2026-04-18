package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stefanuspet/solekita/backend/internal/middleware"
	"github.com/stefanuspet/solekita/backend/internal/service"
)

type AttendanceHandler struct {
	attendanceService *service.AttendanceService
}

func NewAttendanceHandler(attendanceService *service.AttendanceService) *AttendanceHandler {
	return &AttendanceHandler{attendanceService: attendanceService}
}

// HandleAttendance POST /attendances — catat absensi masuk atau keluar.
// Field "type" di form body: "masuk" | "keluar".
func (h *AttendanceHandler) HandleAttendance(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	attendanceType := c.PostForm("type")
	if attendanceType != "masuk" && attendanceType != "keluar" {
		respondError(c, http.StatusBadRequest, `field "type" wajib diisi dengan "masuk" atau "keluar"`, nil)
		return
	}

	file, _, _ := c.Request.FormFile("selfie")
	if file != nil {
		defer file.Close()
	}

	var (
		res *service.AttendanceResponse
		err error
	)
	if attendanceType == "masuk" {
		res, err = h.attendanceService.CheckIn(c.Request.Context(), user, file)
	} else {
		res, err = h.attendanceService.CheckOut(c.Request.Context(), user, file)
	}
	if err != nil {
		handleServiceError(c, err)
		return
	}

	msg := "Absen masuk berhasil dicatat"
	if attendanceType == "keluar" {
		msg = "Absen keluar berhasil dicatat"
	}
	respondSuccess(c, http.StatusCreated, msg, res)
}

func (h *AttendanceHandler) HandleGetToday(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	res, err := h.attendanceService.GetToday(c.Request.Context(), user)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Status absensi hari ini", res)
}

func (h *AttendanceHandler) HandleList(c *gin.Context) {
	user := middleware.GetUserFromContext(c)

	filters := service.ListAttendanceFilters{}

	if s := c.Query("user_id"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			filters.UserID = &id
		}
	}
	if s := c.Query("date_from"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			filters.DateFrom = &t
		}
	}
	if s := c.Query("date_to"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			end := t.Add(24*time.Hour - time.Second)
			filters.DateTo = &end
		}
	}

	page, limit := service.ParseAttendancePage(
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("limit", "20"),
	)

	res, err := h.attendanceService.List(c.Request.Context(), user.OutletID, filters, page, limit)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, "Daftar absensi berhasil diambil", res)
}
