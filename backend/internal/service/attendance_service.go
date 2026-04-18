package service

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/stefanuspet/solekita/backend/internal/errors"
	"github.com/stefanuspet/solekita/backend/internal/model"
	"github.com/stefanuspet/solekita/backend/internal/repository"
	"github.com/stefanuspet/solekita/backend/internal/storage"
)

type AttendanceService struct {
	attendanceRepo *repository.AttendanceRepository
	r2             *storage.R2Storage
}

func NewAttendanceService(attendanceRepo *repository.AttendanceRepository, r2 *storage.R2Storage) *AttendanceService {
	return &AttendanceService{attendanceRepo: attendanceRepo, r2: r2}
}

// ── Request / Response ────────────────────────────────────────────────────────

type AttendanceResponse struct {
	ID         uuid.UUID              `json:"id"`
	UserID     uuid.UUID              `json:"user_id"`
	Type       model.AttendanceType   `json:"type"`
	SelfieURL  *string                `json:"selfie_url"`
	CreatedAt  time.Time              `json:"created_at"`
}

type TodayAttendanceResponse struct {
	CheckIn  *AttendanceResponse `json:"check_in"`
	CheckOut *AttendanceResponse `json:"check_out"`
}

type ListAttendanceFilters struct {
	UserID   *uuid.UUID
	DateFrom *time.Time
	DateTo   *time.Time
}

type ListAttendanceResponse struct {
	Attendances []*AttendanceResponse `json:"attendances"`
	Total       int                   `json:"total"`
	Page        int                   `json:"page"`
	Limit       int                   `json:"limit"`
}

// ── CheckIn ───────────────────────────────────────────────────────────────────

func (s *AttendanceService) CheckIn(ctx context.Context, user *model.UserClaims, selfieFile io.Reader) (*AttendanceResponse, error) {
	today, err := s.attendanceRepo.GetTodayByUserID(ctx, user.ID, user.OutletID)
	if err != nil {
		return nil, fmt.Errorf("CheckIn: ambil absensi hari ini: %w", err)
	}

	for _, a := range today {
		if a.Type == model.AttendanceTypeMasuk {
			return nil, apperrors.ErrConflict.New(fmt.Errorf("sudah absen masuk hari ini"))
		}
	}

	var selfieKey *string
	if selfieFile != nil && s.r2 != nil {
		key, err := s.r2.UploadSelfie(ctx, selfieFile, user.ID, user.OutletID)
		if err != nil {
			return nil, fmt.Errorf("CheckIn: upload selfie: %w", err)
		}
		selfieKey = &key
	}

	a := &model.Attendance{
		UserID:      user.ID,
		OutletID:    user.OutletID,
		Type:        model.AttendanceTypeMasuk,
		SelfieR2Key: selfieKey,
	}
	if err := s.attendanceRepo.Create(ctx, a); err != nil {
		if selfieKey != nil {
			_ = s.r2.Delete(ctx, *selfieKey)
		}
		return nil, fmt.Errorf("CheckIn: simpan: %w", err)
	}

	return s.toResponse(ctx, a), nil
}

// ── CheckOut ──────────────────────────────────────────────────────────────────

func (s *AttendanceService) CheckOut(ctx context.Context, user *model.UserClaims, selfieFile io.Reader) (*AttendanceResponse, error) {
	today, err := s.attendanceRepo.GetTodayByUserID(ctx, user.ID, user.OutletID)
	if err != nil {
		return nil, fmt.Errorf("CheckOut: ambil absensi hari ini: %w", err)
	}

	hasCheckIn := false
	for _, a := range today {
		if a.Type == model.AttendanceTypeMasuk {
			hasCheckIn = true
		}
		if a.Type == model.AttendanceTypeKeluar {
			return nil, apperrors.ErrConflict.New(fmt.Errorf("sudah absen keluar hari ini"))
		}
	}
	if !hasCheckIn {
		return nil, apperrors.ErrUnprocessable.New(fmt.Errorf("belum absen masuk hari ini"))
	}

	var selfieKey *string
	if selfieFile != nil && s.r2 != nil {
		key, err := s.r2.UploadSelfie(ctx, selfieFile, user.ID, user.OutletID)
		if err != nil {
			return nil, fmt.Errorf("CheckOut: upload selfie: %w", err)
		}
		selfieKey = &key
	}

	a := &model.Attendance{
		UserID:      user.ID,
		OutletID:    user.OutletID,
		Type:        model.AttendanceTypeKeluar,
		SelfieR2Key: selfieKey,
	}
	if err := s.attendanceRepo.Create(ctx, a); err != nil {
		if selfieKey != nil {
			_ = s.r2.Delete(ctx, *selfieKey)
		}
		return nil, fmt.Errorf("CheckOut: simpan: %w", err)
	}

	return s.toResponse(ctx, a), nil
}

// ── GetToday ──────────────────────────────────────────────────────────────────

func (s *AttendanceService) GetToday(ctx context.Context, user *model.UserClaims) (*TodayAttendanceResponse, error) {
	today, err := s.attendanceRepo.GetTodayByUserID(ctx, user.ID, user.OutletID)
	if err != nil {
		return nil, fmt.Errorf("GetToday: %w", err)
	}

	res := &TodayAttendanceResponse{}
	for _, a := range today {
		r := s.toResponse(ctx, a)
		switch a.Type {
		case model.AttendanceTypeMasuk:
			res.CheckIn = r
		case model.AttendanceTypeKeluar:
			res.CheckOut = r
		}
	}
	return res, nil
}

// ── List ──────────────────────────────────────────────────────────────────────

func (s *AttendanceService) List(ctx context.Context, outletID uuid.UUID, f ListAttendanceFilters, page, limit int) (*ListAttendanceResponse, error) {
	repoFilters := repository.AttendanceFilters{
		UserID:   f.UserID,
		DateFrom: f.DateFrom,
		DateTo:   f.DateTo,
	}

	attendances, total, err := s.attendanceRepo.ListByOutletID(ctx, outletID, repoFilters, page, limit)
	if err != nil {
		return nil, fmt.Errorf("List: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	res := make([]*AttendanceResponse, 0, len(attendances))
	for _, a := range attendances {
		res = append(res, s.toResponse(ctx, a))
	}

	return &ListAttendanceResponse{
		Attendances: res,
		Total:       total,
		Page:        page,
		Limit:       limit,
	}, nil
}

// ── Helper ────────────────────────────────────────────────────────────────────

func (s *AttendanceService) toResponse(ctx context.Context, a *model.Attendance) *AttendanceResponse {
	res := &AttendanceResponse{
		ID:        a.ID,
		UserID:    a.UserID,
		Type:      a.Type,
		CreatedAt: a.CreatedAt,
	}
	if a.SelfieR2Key != nil && !a.IsSelfieDeleted && s.r2 != nil {
		url, err := s.r2.GetSignedURL(ctx, *a.SelfieR2Key, time.Hour)
		if err == nil {
			res.SelfieURL = &url
		}
	}
	return res
}

// ParseAttendancePage adalah helper untuk handler agar tidak duplikasi logika.
func ParseAttendancePage(pageStr, limitStr string) (int, int) {
	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	return page, limit
}
