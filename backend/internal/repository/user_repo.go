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

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create menyimpan user baru. user.ID harus di-set sebelum dipanggil (pre-generated UUID).
// Gunakan tx untuk operasi dalam transaction.
func (r *UserRepository) Create(ctx context.Context, q dbtx, user *model.User) error {
	query := `
		INSERT INTO users (id, outlet_id, name, phone, password_hash, is_owner, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at
	`
	err := q.QueryRowContext(ctx, query,
		user.ID,
		user.OutletID,
		user.Name,
		user.Phone,
		user.PasswordHash,
		user.IsOwner,
		user.IsActive,
	).Scan(&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("UserRepository.Create: %w", err)
	}
	return nil
}

// GetByPhone mengambil user berdasarkan nomor HP.
// Digunakan saat login — tidak perlu filter outlet_id karena phone bersifat unique global.
func (r *UserRepository) GetByPhone(ctx context.Context, phone string) (*model.User, error) {
	query := `
		SELECT id, outlet_id, name, phone, password_hash, is_owner, is_active,
		       last_login_at, created_at, updated_at
		FROM users
		WHERE phone = $1
	`
	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, phone).Scan(
		&user.ID,
		&user.OutletID,
		&user.Name,
		&user.Phone,
		&user.PasswordHash,
		&user.IsOwner,
		&user.IsActive,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("UserRepository.GetByPhone: %w", err)
	}
	return user, nil
}

// GetByIDInternal mengambil user berdasarkan ID tanpa filter outlet_id.
// Hanya dipakai di konteks tepercaya (misal: refresh token flow) di mana user_id sudah divalidasi lewat DB.
func (r *UserRepository) GetByIDInternal(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, outlet_id, name, phone, password_hash, is_owner, is_active,
		       last_login_at, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.OutletID,
		&user.Name,
		&user.Phone,
		&user.PasswordHash,
		&user.IsOwner,
		&user.IsActive,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("UserRepository.GetByIDInternal: %w", err)
	}
	return user, nil
}

// GetByID mengambil user berdasarkan ID dengan validasi outlet_id untuk cegah akses lintas outlet.
func (r *UserRepository) GetByID(ctx context.Context, id, outletID uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, outlet_id, name, phone, password_hash, is_owner, is_active,
		       last_login_at, created_at, updated_at
		FROM users
		WHERE id = $1 AND outlet_id = $2
	`
	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, id, outletID).Scan(
		&user.ID,
		&user.OutletID,
		&user.Name,
		&user.Phone,
		&user.PasswordHash,
		&user.IsOwner,
		&user.IsActive,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("UserRepository.GetByID: %w", err)
	}
	return user, nil
}

// UpdatePhone memperbarui nomor HP user. Digunakan admin untuk koreksi nomor owner.
func (r *UserRepository) UpdatePhone(ctx context.Context, id uuid.UUID, phone string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE users SET phone = $1, updated_at = now() WHERE id = $2`,
		phone, id,
	)
	if err != nil {
		return fmt.Errorf("UserRepository.UpdatePhone: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// UpdateLastLogin memperbarui timestamp login terakhir user.
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET last_login_at = now() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("UserRepository.UpdateLastLogin: %w", err)
	}
	return nil
}

// UpdatePassword memperbarui password_hash user.
func (r *UserRepository) UpdatePassword(ctx context.Context, id, outletID uuid.UUID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1 WHERE id = $2 AND outlet_id = $3`
	res, err := r.db.ExecContext(ctx, query, passwordHash, id, outletID)
	if err != nil {
		return fmt.Errorf("UserRepository.UpdatePassword: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// ListByOutletID mengambil semua user dalam satu outlet.
// Jika isActive != nil, filter berdasarkan nilai is_active-nya.
func (r *UserRepository) ListByOutletID(ctx context.Context, outletID uuid.UUID, isActive *bool) ([]*model.User, error) {
	query := `
		SELECT id, outlet_id, name, phone, password_hash, is_owner, is_active,
		       last_login_at, created_at, updated_at
		FROM users
		WHERE outlet_id = $1
	`
	args := []any{outletID}
	if isActive != nil {
		query += fmt.Sprintf(" AND is_active = $%d", len(args)+1)
		args = append(args, *isActive)
	}
	query += " ORDER BY created_at ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("UserRepository.ListByOutletID: %w", err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		u := &model.User{}
		if err := rows.Scan(
			&u.ID, &u.OutletID, &u.Name, &u.Phone, &u.PasswordHash,
			&u.IsOwner, &u.IsActive, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("UserRepository.ListByOutletID scan: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("UserRepository.ListByOutletID rows: %w", err)
	}
	return users, nil
}

// Update memperbarui nama dan status aktif user.
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users
		SET name = $1, is_active = $2
		WHERE id = $3 AND outlet_id = $4
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		user.Name,
		user.IsActive,
		user.ID,
		user.OutletID,
	).Scan(&user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("UserRepository.Update: %w", err)
	}
	return nil
}

// GetPermissions mengambil daftar permission user dari tabel user_permissions.
func (r *UserRepository) GetPermissions(ctx context.Context, userID uuid.UUID) ([]model.Permission, error) {
	query := `SELECT permission FROM user_permissions WHERE user_id = $1`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("UserRepository.GetPermissions: %w", err)
	}
	defer rows.Close()

	var perms []model.Permission
	for rows.Next() {
		var p model.Permission
		if err := rows.Scan(&p); err != nil {
			return nil, fmt.Errorf("UserRepository.GetPermissions scan: %w", err)
		}
		perms = append(perms, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("UserRepository.GetPermissions rows: %w", err)
	}
	return perms, nil
}
