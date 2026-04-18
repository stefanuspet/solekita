package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/stefanuspet/solekita/backend/internal/model"
)

type UserPermissionRepository struct {
	db *sql.DB
}

func NewUserPermissionRepository(db *sql.DB) *UserPermissionRepository {
	return &UserPermissionRepository{db: db}
}

// GetByUserID mengambil semua permission milik user.
func (r *UserPermissionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.Permission, error) {
	query := `SELECT permission FROM user_permissions WHERE user_id = $1`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("UserPermissionRepository.GetByUserID: %w", err)
	}
	defer rows.Close()

	var perms []model.Permission
	for rows.Next() {
		var p model.Permission
		if err := rows.Scan(&p); err != nil {
			return nil, fmt.Errorf("UserPermissionRepository.GetByUserID scan: %w", err)
		}
		perms = append(perms, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("UserPermissionRepository.GetByUserID rows: %w", err)
	}
	return perms, nil
}

// ReplaceAll mengganti seluruh permission user dalam satu transaction.
// Hapus semua permission lama lalu insert yang baru — atomic.
func (r *UserPermissionRepository) ReplaceAll(ctx context.Context, q dbtx, userID uuid.UUID, permissions []model.Permission) error {
	if _, err := q.ExecContext(ctx, `DELETE FROM user_permissions WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("UserPermissionRepository.ReplaceAll delete: %w", err)
	}

	for _, p := range permissions {
		_, err := q.ExecContext(ctx,
			`INSERT INTO user_permissions (user_id, permission) VALUES ($1, $2)`,
			userID, p,
		)
		if err != nil {
			return fmt.Errorf("UserPermissionRepository.ReplaceAll insert %s: %w", p, err)
		}
	}
	return nil
}
