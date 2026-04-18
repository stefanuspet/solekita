package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	apperrors "github.com/stefanuspet/solekita/backend/internal/errors"
	"github.com/stefanuspet/solekita/backend/internal/model"
)

type RefreshTokenRepository struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create menyimpan refresh token baru. Gunakan tx untuk operasi dalam transaction.
func (r *RefreshTokenRepository) Create(ctx context.Context, q dbtx, rt *model.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expired_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	err := q.QueryRowContext(ctx, query,
		rt.UserID,
		rt.TokenHash,
		rt.ExpiredAt,
	).Scan(&rt.ID, &rt.CreatedAt)

	if err != nil {
		return fmt.Errorf("RefreshTokenRepository.Create: %w", err)
	}
	return nil
}

// GetByTokenHash mengambil refresh token berdasarkan hash-nya.
func (r *RefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expired_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`
	rt := &model.RefreshToken{}
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&rt.ID,
		&rt.UserID,
		&rt.TokenHash,
		&rt.ExpiredAt,
		&rt.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("RefreshTokenRepository.GetByTokenHash: %w", err)
	}
	return rt, nil
}

// DeleteByTokenHash menghapus refresh token berdasarkan hash-nya (dipakai saat logout).
func (r *RefreshTokenRepository) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	query := `DELETE FROM refresh_tokens WHERE token_hash = $1`
	_, err := r.db.ExecContext(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("RefreshTokenRepository.DeleteByTokenHash: %w", err)
	}
	return nil
}
