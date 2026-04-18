package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stefanuspet/solekita/backend/internal/model"
)

type PhotoRepository struct {
	db *sql.DB
}

func NewPhotoRepository(db *sql.DB) *PhotoRepository {
	return &PhotoRepository{db: db}
}

func (r *PhotoRepository) Create(ctx context.Context, tx *sql.Tx, p *model.Photo) error {
	return r.create(ctx, tx, p)
}

// CreateDirect menyimpan foto tanpa transaction — dipakai di luar alur CreateOrder.
func (r *PhotoRepository) CreateDirect(ctx context.Context, p *model.Photo) error {
	return r.create(ctx, r.db, p)
}

func (r *PhotoRepository) create(ctx context.Context, q dbtx, p *model.Photo) error {
	query := `
		INSERT INTO photos (order_id, type, r2_key, file_size_kb)
		VALUES ($1, $2, $3, $4)
		RETURNING id, is_deleted, uploaded_at
	`
	err := q.QueryRowContext(ctx, query,
		p.OrderID, p.Type, p.R2Key, p.FileSizeKB,
	).Scan(&p.ID, &p.IsDeleted, &p.UploadedAt)
	if err != nil {
		return fmt.Errorf("PhotoRepository.Create: %w", err)
	}
	return nil
}

func (r *PhotoRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*model.Photo, error) {
	query := `
		SELECT id, order_id, type, r2_key, file_size_kb, is_deleted, uploaded_at, deleted_at
		FROM photos
		WHERE order_id = $1 AND is_deleted = false
		ORDER BY uploaded_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("PhotoRepository.GetByOrderID: %w", err)
	}
	defer rows.Close()

	var photos []*model.Photo
	for rows.Next() {
		p := &model.Photo{}
		if err := rows.Scan(
			&p.ID, &p.OrderID, &p.Type, &p.R2Key, &p.FileSizeKB,
			&p.IsDeleted, &p.UploadedAt, &p.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("PhotoRepository.GetByOrderID scan: %w", err)
		}
		photos = append(photos, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("PhotoRepository.GetByOrderID rows: %w", err)
	}
	return photos, nil
}

func (r *PhotoRepository) GetByOrderIDAndType(ctx context.Context, orderID uuid.UUID, photoType model.PhotoType) (*model.Photo, error) {
	query := `
		SELECT id, order_id, type, r2_key, file_size_kb, is_deleted, uploaded_at, deleted_at
		FROM photos
		WHERE order_id = $1 AND type = $2 AND is_deleted = false
		LIMIT 1
	`
	p := &model.Photo{}
	err := r.db.QueryRowContext(ctx, query, orderID, photoType).Scan(
		&p.ID, &p.OrderID, &p.Type, &p.R2Key, &p.FileSizeKB,
		&p.IsDeleted, &p.UploadedAt, &p.DeletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("PhotoRepository.GetByOrderIDAndType: %w", err)
	}
	return p, nil
}

// ExpiredPhotoRow adalah baris hasil query untuk cleanup — hanya butuh ID dan R2Key.
type ExpiredPhotoRow struct {
	ID    uuid.UUID
	R2Key string
}

// ListExpired mencari foto yang sudah melewati batas retensi dan belum dihapus.
// Digunakan oleh CleanupExpiredPhotos scheduler.
func (r *PhotoRepository) ListExpired(ctx context.Context, photoType model.PhotoType, olderThan time.Time) ([]*ExpiredPhotoRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, r2_key FROM photos
		WHERE type = $1
		  AND uploaded_at < $2
		  AND is_deleted = false
		ORDER BY uploaded_at ASC
	`, photoType, olderThan)
	if err != nil {
		return nil, fmt.Errorf("PhotoRepository.ListExpired: %w", err)
	}
	defer rows.Close()

	var result []*ExpiredPhotoRow
	for rows.Next() {
		row := &ExpiredPhotoRow{}
		if err := rows.Scan(&row.ID, &row.R2Key); err != nil {
			return nil, fmt.Errorf("PhotoRepository.ListExpired scan: %w", err)
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// MarkDeleted menandai foto sebagai terhapus setelah file-nya dihapus dari R2.
func (r *PhotoRepository) MarkDeleted(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE photos SET is_deleted = true, deleted_at = now()
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("PhotoRepository.MarkDeleted: %w", err)
	}
	return nil
}
