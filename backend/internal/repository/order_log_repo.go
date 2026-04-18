package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/stefanuspet/solekita/backend/internal/model"
)

type OrderLogRepository struct {
	db *sql.DB
}

func NewOrderLogRepository(db *sql.DB) *OrderLogRepository {
	return &OrderLogRepository{db: db}
}

// Create mencatat aksi pada order. Append-only — tidak ada update atau delete.
func (r *OrderLogRepository) Create(ctx context.Context, log *model.OrderLog) error {
	query := `
		INSERT INTO order_logs (order_id, user_id, action, old_value, new_value, notes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	err := r.db.QueryRowContext(ctx, query,
		log.OrderID, log.UserID, log.Action, log.OldValue, log.NewValue, log.Notes,
	).Scan(&log.ID, &log.CreatedAt)
	if err != nil {
		return fmt.Errorf("OrderLogRepository.Create: %w", err)
	}
	return nil
}

// GetByOrderID mengambil semua log untuk satu order, urut dari terlama ke terbaru.
func (r *OrderLogRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*model.OrderLog, error) {
	query := `
		SELECT id, order_id, user_id, action, old_value, new_value, notes, created_at
		FROM order_logs
		WHERE order_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("OrderLogRepository.GetByOrderID: %w", err)
	}
	defer rows.Close()

	var logs []*model.OrderLog
	for rows.Next() {
		l := &model.OrderLog{}
		if err := rows.Scan(
			&l.ID, &l.OrderID, &l.UserID, &l.Action,
			&l.OldValue, &l.NewValue, &l.Notes, &l.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("OrderLogRepository.GetByOrderID scan: %w", err)
		}
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("OrderLogRepository.GetByOrderID rows: %w", err)
	}
	return logs, nil
}
