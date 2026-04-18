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

type PaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(ctx context.Context, tx *sql.Tx, p *model.Payment) error {
	query := `
		INSERT INTO payments (order_id, amount, method, status, notes, paid_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	err := tx.QueryRowContext(ctx, query,
		p.OrderID, p.Amount, p.Method, p.Status, p.Notes, p.PaidAt,
	).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		return fmt.Errorf("PaymentRepository.Create: %w", err)
	}
	return nil
}

func (r *PaymentRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) (*model.Payment, error) {
	query := `
		SELECT id, order_id, amount, method, status, notes, paid_at, created_at
		FROM payments
		WHERE order_id = $1
		LIMIT 1
	`
	p := &model.Payment{}
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&p.ID, &p.OrderID, &p.Amount, &p.Method, &p.Status,
		&p.Notes, &p.PaidAt, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("PaymentRepository.GetByOrderID: %w", err)
	}
	return p, nil
}

func (r *PaymentRepository) Update(ctx context.Context, p *model.Payment) error {
	query := `
		UPDATE payments
		SET amount = $1, method = $2, status = $3, notes = $4, paid_at = $5
		WHERE id = $6
		RETURNING created_at
	`
	err := r.db.QueryRowContext(ctx, query,
		p.Amount, p.Method, p.Status, p.Notes, p.PaidAt, p.ID,
	).Scan(&p.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("PaymentRepository.Update: %w", err)
	}
	return nil
}
