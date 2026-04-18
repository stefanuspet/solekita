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

type TreatmentRepository struct {
	db *sql.DB
}

func NewTreatmentRepository(db *sql.DB) *TreatmentRepository {
	return &TreatmentRepository{db: db}
}

func (r *TreatmentRepository) Create(ctx context.Context, t *model.Treatment) error {
	query := `
		INSERT INTO treatments (outlet_id, name, material, price, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		t.OutletID, t.Name, t.Material, t.Price, t.IsActive,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return fmt.Errorf("TreatmentRepository.Create: %w", err)
	}
	return nil
}

func (r *TreatmentRepository) GetByID(ctx context.Context, id, outletID uuid.UUID) (*model.Treatment, error) {
	query := `
		SELECT id, outlet_id, name, material, price, is_active, created_at, updated_at
		FROM treatments
		WHERE id = $1 AND outlet_id = $2
	`
	t := &model.Treatment{}
	err := r.db.QueryRowContext(ctx, query, id, outletID).Scan(
		&t.ID, &t.OutletID, &t.Name, &t.Material, &t.Price,
		&t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("TreatmentRepository.GetByID: %w", err)
	}
	return t, nil
}

// ListByOutletID mengambil semua treatment dalam outlet.
// isActive dan material bersifat opsional (nil = tidak difilter).
func (r *TreatmentRepository) ListByOutletID(ctx context.Context, outletID uuid.UUID, isActive *bool, material *string) ([]*model.Treatment, error) {
	query := `
		SELECT id, outlet_id, name, material, price, is_active, created_at, updated_at
		FROM treatments
		WHERE outlet_id = $1
	`
	args := []any{outletID}

	if isActive != nil {
		args = append(args, *isActive)
		query += fmt.Sprintf(" AND is_active = $%d", len(args))
	}
	if material != nil {
		args = append(args, *material)
		query += fmt.Sprintf(" AND material = $%d", len(args))
	}
	query += " ORDER BY name ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("TreatmentRepository.ListByOutletID: %w", err)
	}
	defer rows.Close()

	var treatments []*model.Treatment
	for rows.Next() {
		t := &model.Treatment{}
		if err := rows.Scan(
			&t.ID, &t.OutletID, &t.Name, &t.Material, &t.Price,
			&t.IsActive, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("TreatmentRepository.ListByOutletID scan: %w", err)
		}
		treatments = append(treatments, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("TreatmentRepository.ListByOutletID rows: %w", err)
	}
	return treatments, nil
}

func (r *TreatmentRepository) Update(ctx context.Context, t *model.Treatment) error {
	query := `
		UPDATE treatments
		SET name = $1, material = $2, price = $3, is_active = $4
		WHERE id = $5 AND outlet_id = $6
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		t.Name, t.Material, t.Price, t.IsActive, t.ID, t.OutletID,
	).Scan(&t.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("TreatmentRepository.Update: %w", err)
	}
	return nil
}

// Delete menghapus treatment secara permanen.
// Caller harus memastikan treatment belum pernah dipakai di order sebelum memanggil ini.
func (r *TreatmentRepository) Delete(ctx context.Context, id, outletID uuid.UUID) error {
	query := `DELETE FROM treatments WHERE id = $1 AND outlet_id = $2`
	res, err := r.db.ExecContext(ctx, query, id, outletID)
	if err != nil {
		return fmt.Errorf("TreatmentRepository.Delete: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// IsUsedInOrder memeriksa apakah treatment pernah dipakai di order manapun.
func (r *TreatmentRepository) IsUsedInOrder(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM orders WHERE treatment_id = $1)`, id,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("TreatmentRepository.IsUsedInOrder: %w", err)
	}
	return exists, nil
}
