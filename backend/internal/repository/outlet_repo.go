package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	apperrors "github.com/stefanuspet/solekita/backend/internal/errors"
	"github.com/stefanuspet/solekita/backend/internal/model"
)

// dbtx memungkinkan method repo dipanggil dengan *sql.DB maupun *sql.Tx.
type dbtx interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

type OutletRepository struct {
	db *sql.DB
}

func NewOutletRepository(db *sql.DB) *OutletRepository {
	return &OutletRepository{db: db}
}

// Create menyimpan outlet baru. Gunakan tx untuk operasi dalam transaction.
func (r *OutletRepository) Create(ctx context.Context, q dbtx, outlet *model.Outlet) error {
	query := `
		INSERT INTO outlets (name, code, address, phone, owner_id, subscription_status, overdue_threshold_days, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	err := q.QueryRowContext(ctx, query,
		outlet.Name,
		outlet.Code,
		outlet.Address,
		outlet.Phone,
		outlet.OwnerID,
		outlet.SubscriptionStatus,
		outlet.OverdueThresholdDays,
		outlet.IsActive,
	).Scan(&outlet.ID, &outlet.CreatedAt, &outlet.UpdatedAt)

	if err != nil {
		return fmt.Errorf("OutletRepository.Create: %w", err)
	}
	return nil
}

// Update memperbarui field outlet yang bisa diedit oleh owner.
func (r *OutletRepository) Update(ctx context.Context, outlet *model.Outlet) error {
	query := `
		UPDATE outlets
		SET name = $1, address = $2, phone = $3, overdue_threshold_days = $4
		WHERE id = $5
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		outlet.Name,
		outlet.Address,
		outlet.Phone,
		outlet.OverdueThresholdDays,
		outlet.ID,
	).Scan(&outlet.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("OutletRepository.Update: %w", err)
	}
	return nil
}

// UpdateSubscriptionStatus memperbarui subscription_status outlet.
// Digunakan oleh webhook payment dan scheduler.
func (r *OutletRepository) UpdateSubscriptionStatus(ctx context.Context, outletID uuid.UUID, status model.SubscriptionStatus) error {
	err := r.db.QueryRowContext(ctx, `
		UPDATE outlets SET subscription_status = $2, updated_at = now()
		WHERE id = $1
		RETURNING id
	`, outletID, status).Scan(new(uuid.UUID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("OutletRepository.UpdateSubscriptionStatus: %w", err)
	}
	return nil
}

// CodeExists cek apakah outlet code sudah dipakai.
func (r *OutletRepository) CodeExists(ctx context.Context, code string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM outlets WHERE code = $1)`, code,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("OutletRepository.CodeExists: %w", err)
	}
	return exists, nil
}

// GetByID mengambil outlet berdasarkan ID.
func (r *OutletRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Outlet, error) {
	query := `
		SELECT id, name, code, address, phone, owner_id, subscription_status,
		       overdue_threshold_days, is_active, created_at, updated_at
		FROM outlets
		WHERE id = $1
	`
	outlet := &model.Outlet{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&outlet.ID,
		&outlet.Name,
		&outlet.Code,
		&outlet.Address,
		&outlet.Phone,
		&outlet.OwnerID,
		&outlet.SubscriptionStatus,
		&outlet.OverdueThresholdDays,
		&outlet.IsActive,
		&outlet.CreatedAt,
		&outlet.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("OutletRepository.GetByID: %w", err)
	}
	return outlet, nil
}
