package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/stefanuspet/solekita/backend/internal/errors"
	"github.com/stefanuspet/solekita/backend/internal/model"
)

type SubscriptionInvoiceRepository struct {
	db *sql.DB
}

func NewSubscriptionInvoiceRepository(db *sql.DB) *SubscriptionInvoiceRepository {
	return &SubscriptionInvoiceRepository{db: db}
}

func (r *SubscriptionInvoiceRepository) Create(ctx context.Context, inv *model.SubscriptionInvoice) error {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO subscription_invoices
		    (subscription_id, outlet_id, amount, due_date, tripay_reference, tripay_payment_url, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`, inv.SubscriptionID, inv.OutletID, inv.Amount, inv.DueDate,
		inv.TripayReference, inv.TripayPaymentURL, inv.Status,
	).Scan(&inv.ID, &inv.CreatedAt, &inv.UpdatedAt)
	if err != nil {
		return fmt.Errorf("SubscriptionInvoiceRepository.Create: %w", err)
	}
	return nil
}

func (r *SubscriptionInvoiceRepository) GetByMerchantRef(ctx context.Context, ref string) (*model.SubscriptionInvoice, error) {
	inv := &model.SubscriptionInvoice{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, subscription_id, outlet_id, amount, due_date,
		       tripay_reference, tripay_payment_url, status, paid_at, created_at, updated_at
		FROM subscription_invoices
		WHERE tripay_reference = $1
	`, ref).Scan(
		&inv.ID, &inv.SubscriptionID, &inv.OutletID, &inv.Amount, &inv.DueDate,
		&inv.TripayReference, &inv.TripayPaymentURL, &inv.Status, &inv.PaidAt,
		&inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("SubscriptionInvoiceRepository.GetByMerchantRef: %w", err)
	}
	return inv, nil
}

func (r *SubscriptionInvoiceRepository) Update(ctx context.Context, inv *model.SubscriptionInvoice) error {
	err := r.db.QueryRowContext(ctx, `
		UPDATE subscription_invoices
		SET tripay_reference   = $2,
		    tripay_payment_url = $3,
		    status             = $4,
		    paid_at            = $5,
		    updated_at         = now()
		WHERE id = $1
		RETURNING updated_at
	`, inv.ID, inv.TripayReference, inv.TripayPaymentURL, inv.Status, inv.PaidAt,
	).Scan(&inv.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("SubscriptionInvoiceRepository.Update: %w", err)
	}
	return nil
}

func (r *SubscriptionInvoiceRepository) ListByOutletID(ctx context.Context, outletID uuid.UUID, page, limit int) ([]*model.SubscriptionInvoice, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM subscription_invoices WHERE outlet_id = $1`, outletID,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("SubscriptionInvoiceRepository.ListByOutletID count: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, subscription_id, outlet_id, amount, due_date,
		       tripay_reference, tripay_payment_url, status, paid_at, created_at, updated_at
		FROM subscription_invoices
		WHERE outlet_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, outletID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("SubscriptionInvoiceRepository.ListByOutletID: %w", err)
	}
	defer rows.Close()

	var invoices []*model.SubscriptionInvoice
	for rows.Next() {
		inv := &model.SubscriptionInvoice{}
		if err := rows.Scan(
			&inv.ID, &inv.SubscriptionID, &inv.OutletID, &inv.Amount, &inv.DueDate,
			&inv.TripayReference, &inv.TripayPaymentURL, &inv.Status, &inv.PaidAt,
			&inv.CreatedAt, &inv.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("SubscriptionInvoiceRepository.ListByOutletID scan: %w", err)
		}
		invoices = append(invoices, inv)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("SubscriptionInvoiceRepository.ListByOutletID rows: %w", err)
	}
	return invoices, total, nil
}

func (r *SubscriptionInvoiceRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.SubscriptionInvoice, error) {
	inv := &model.SubscriptionInvoice{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, subscription_id, outlet_id, amount, due_date,
		       tripay_reference, tripay_payment_url, status, paid_at, created_at, updated_at
		FROM subscription_invoices
		WHERE id = $1
	`, id).Scan(
		&inv.ID, &inv.SubscriptionID, &inv.OutletID, &inv.Amount, &inv.DueDate,
		&inv.TripayReference, &inv.TripayPaymentURL, &inv.Status, &inv.PaidAt,
		&inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("SubscriptionInvoiceRepository.GetByID: %w", err)
	}
	return inv, nil
}

// GetPendingByOutletID mengambil invoice pending terbaru untuk outlet.
// Digunakan oleh scheduler untuk cek apakah tagihan sudah ada.
func (r *SubscriptionInvoiceRepository) GetPendingByOutletID(ctx context.Context, outletID uuid.UUID) (*model.SubscriptionInvoice, error) {
	inv := &model.SubscriptionInvoice{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, subscription_id, outlet_id, amount, due_date,
		       tripay_reference, tripay_payment_url, status, paid_at, created_at, updated_at
		FROM subscription_invoices
		WHERE outlet_id = $1 AND status = 'pending'
		ORDER BY created_at DESC
		LIMIT 1
	`, outletID).Scan(
		&inv.ID, &inv.SubscriptionID, &inv.OutletID, &inv.Amount, &inv.DueDate,
		&inv.TripayReference, &inv.TripayPaymentURL, &inv.Status, &inv.PaidAt,
		&inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("SubscriptionInvoiceRepository.GetPendingByOutletID: %w", err)
	}
	return inv, nil
}

// ListOutletsWithDueDateOn mengambil daftar outlet yang next_due_date jatuh pada tanggal tertentu.
// Digunakan oleh billing scheduler untuk generate tagihan otomatis.
func (r *SubscriptionInvoiceRepository) ListOutletsWithDueDateOn(ctx context.Context, date time.Time) ([]uuid.UUID, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT outlet_id FROM subscriptions
		WHERE next_due_date::date = $1::date
	`, date.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("SubscriptionInvoiceRepository.ListOutletsWithDueDateOn: %w", err)
	}
	defer rows.Close()

	var outletIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("SubscriptionInvoiceRepository.ListOutletsWithDueDateOn scan: %w", err)
		}
		outletIDs = append(outletIDs, id)
	}
	return outletIDs, rows.Err()
}
