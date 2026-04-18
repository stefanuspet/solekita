package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/stefanuspet/solekita/backend/internal/model"
)

type AttendanceRepository struct {
	db *sql.DB
}

func NewAttendanceRepository(db *sql.DB) *AttendanceRepository {
	return &AttendanceRepository{db: db}
}

func (r *AttendanceRepository) Create(ctx context.Context, a *model.Attendance) error {
	query := `
		INSERT INTO attendances (user_id, outlet_id, type, selfie_r2_key)
		VALUES ($1, $2, $3, $4)
		RETURNING id, is_selfie_deleted, created_at
	`
	err := r.db.QueryRowContext(ctx, query,
		a.UserID, a.OutletID, a.Type, a.SelfieR2Key,
	).Scan(&a.ID, &a.IsSelfieDeleted, &a.CreatedAt)
	if err != nil {
		return fmt.Errorf("AttendanceRepository.Create: %w", err)
	}
	return nil
}

// AttendanceFilters dipakai di ListByOutletID.
type AttendanceFilters struct {
	UserID   *uuid.UUID
	DateFrom *time.Time
	DateTo   *time.Time
}

// ListByOutletID mengambil daftar absensi dalam outlet dengan filter opsional.
func (r *AttendanceRepository) ListByOutletID(ctx context.Context, outletID uuid.UUID, f AttendanceFilters, page, limit int) ([]*model.Attendance, int, error) {
	args := []any{outletID}
	conditions := []string{"outlet_id = $1"}

	if f.UserID != nil {
		args = append(args, *f.UserID)
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", len(args)))
	}
	if f.DateFrom != nil {
		args = append(args, *f.DateFrom)
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", len(args)))
	}
	if f.DateTo != nil {
		args = append(args, *f.DateTo)
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", len(args)))
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM attendances `+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("AttendanceRepository.ListByOutletID count: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit
	args = append(args, limit, offset)

	query := `
		SELECT id, user_id, outlet_id, type, selfie_r2_key, is_selfie_deleted, created_at
		FROM attendances ` + where +
		fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, len(args)-1, len(args))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("AttendanceRepository.ListByOutletID: %w", err)
	}
	defer rows.Close()

	var attendances []*model.Attendance
	for rows.Next() {
		a := &model.Attendance{}
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.OutletID, &a.Type,
			&a.SelfieR2Key, &a.IsSelfieDeleted, &a.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("AttendanceRepository.ListByOutletID scan: %w", err)
		}
		attendances = append(attendances, a)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("AttendanceRepository.ListByOutletID rows: %w", err)
	}
	return attendances, total, nil
}

// ExpiredSelfieRow adalah baris hasil query untuk cleanup selfie attendance.
type ExpiredSelfieRow struct {
	ID    uuid.UUID
	R2Key string
}

// ListExpiredSelfies mencari selfie attendance yang sudah melewati batas retensi.
// Digunakan oleh CleanupExpiredSelfies scheduler.
func (r *AttendanceRepository) ListExpiredSelfies(ctx context.Context, olderThan time.Time) ([]*ExpiredSelfieRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, selfie_r2_key FROM attendances
		WHERE created_at < $1
		  AND is_selfie_deleted = false
		  AND selfie_r2_key IS NOT NULL
		ORDER BY created_at ASC
	`, olderThan)
	if err != nil {
		return nil, fmt.Errorf("AttendanceRepository.ListExpiredSelfies: %w", err)
	}
	defer rows.Close()

	var result []*ExpiredSelfieRow
	for rows.Next() {
		row := &ExpiredSelfieRow{}
		if err := rows.Scan(&row.ID, &row.R2Key); err != nil {
			return nil, fmt.Errorf("AttendanceRepository.ListExpiredSelfies scan: %w", err)
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// MarkSelfieDeleted menandai selfie sudah dihapus dari R2.
func (r *AttendanceRepository) MarkSelfieDeleted(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE attendances SET is_selfie_deleted = true WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("AttendanceRepository.MarkSelfieDeleted: %w", err)
	}
	return nil
}

// GetTodayByUserID mengambil semua absensi user pada hari ini.
func (r *AttendanceRepository) GetTodayByUserID(ctx context.Context, userID, outletID uuid.UUID) ([]*model.Attendance, error) {
	query := `
		SELECT id, user_id, outlet_id, type, selfie_r2_key, is_selfie_deleted, created_at
		FROM attendances
		WHERE user_id = $1 AND outlet_id = $2 AND created_at::date = CURRENT_DATE
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, userID, outletID)
	if err != nil {
		return nil, fmt.Errorf("AttendanceRepository.GetTodayByUserID: %w", err)
	}
	defer rows.Close()

	var attendances []*model.Attendance
	for rows.Next() {
		a := &model.Attendance{}
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.OutletID, &a.Type,
			&a.SelfieR2Key, &a.IsSelfieDeleted, &a.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("AttendanceRepository.GetTodayByUserID scan: %w", err)
		}
		attendances = append(attendances, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("AttendanceRepository.GetTodayByUserID rows: %w", err)
	}
	return attendances, nil
}
