package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/stefanuspet/solekita/backend/internal/errors"
	"github.com/stefanuspet/solekita/backend/internal/model"
)

type DeliveryRepository struct {
	db *sql.DB
}

func NewDeliveryRepository(db *sql.DB) *DeliveryRepository {
	return &DeliveryRepository{db: db}
}

type DeliveryFilters struct {
	Type   string     // "pickup" | "delivery"
	Status *string    // filter by pickup_status atau delivery_status
	Date   *time.Time // filter by tanggal order dibuat
}

// ── helpers ───────────────────────────────────────────────────────────────────

func scanDelivery(row interface{ Scan(...any) error }) (*model.Delivery, error) {
	d := &model.Delivery{}
	err := row.Scan(
		&d.ID, &d.OrderID, &d.CourierID,
		&d.PickupAddress, &d.DeliveryAddress,
		&d.PickupStatus, &d.DeliveryStatus,
		&d.PickupNotes, &d.DeliveryNotes,
		&d.PickedUpAt, &d.DeliveredAt,
		&d.CreatedAt, &d.UpdatedAt,
	)
	return d, err
}

const deliverySelectCols = `
	d.id, d.order_id, d.courier_id,
	d.pickup_address, d.delivery_address,
	d.pickup_status, d.delivery_status,
	d.pickup_notes, d.delivery_notes,
	d.picked_up_at, d.delivered_at,
	d.created_at, d.updated_at`

// ── Create ────────────────────────────────────────────────────────────────────

func (r *DeliveryRepository) Create(ctx context.Context, tx *sql.Tx, d *model.Delivery) error {
	err := tx.QueryRowContext(ctx, `
		INSERT INTO deliveries (order_id, pickup_address, delivery_address, pickup_status, delivery_status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`, d.OrderID, d.PickupAddress, d.DeliveryAddress, d.PickupStatus, d.DeliveryStatus,
	).Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return fmt.Errorf("DeliveryRepository.Create: %w", err)
	}
	return nil
}

// ── GetByID ───────────────────────────────────────────────────────────────────

func (r *DeliveryRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Delivery, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT `+deliverySelectCols+`
		FROM deliveries d
		WHERE d.id = $1
	`, id)
	d, err := scanDelivery(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("DeliveryRepository.GetByID: %w", err)
	}
	return d, nil
}

// ── GetByOrderID ──────────────────────────────────────────────────────────────

// GetByOrderID mengambil delivery berdasarkan order_id.
// outletID dipakai untuk memastikan order milik outlet yang sama (multi-tenant isolation).
func (r *DeliveryRepository) GetByOrderID(ctx context.Context, orderID, outletID uuid.UUID) (*model.Delivery, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT `+deliverySelectCols+`
		FROM deliveries d
		JOIN orders o ON o.id = d.order_id
		WHERE d.order_id = $1
		  AND o.outlet_id = $2
	`, orderID, outletID)
	d, err := scanDelivery(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("DeliveryRepository.GetByOrderID: %w", err)
	}
	return d, nil
}

// ── GetByCourierID ────────────────────────────────────────────────────────────

// GetByCourierID mengambil daftar delivery yang di-assign ke kurir tertentu.
// Filter type: "pickup" → filter o.is_pickup=true; "delivery" → o.is_delivery=true.
// Filter status diterapkan ke pickup_status atau delivery_status sesuai type.
func (r *DeliveryRepository) GetByCourierID(ctx context.Context, courierID uuid.UUID, f DeliveryFilters) ([]*model.Delivery, error) {
	args := []any{courierID}
	cond := []string{"d.courier_id = $1"}

	if f.Type == "pickup" {
		cond = append(cond, "o.is_pickup = true")
		if f.Status != nil {
			args = append(args, *f.Status)
			cond = append(cond, fmt.Sprintf("d.pickup_status = $%d", len(args)))
		}
	} else if f.Type == "delivery" {
		cond = append(cond, "o.is_delivery = true")
		if f.Status != nil {
			args = append(args, *f.Status)
			cond = append(cond, fmt.Sprintf("d.delivery_status = $%d", len(args)))
		}
	}

	if f.Date != nil {
		args = append(args, f.Date.Format("2006-01-02"))
		cond = append(cond, fmt.Sprintf("o.created_at::date = $%d::date", len(args)))
	}

	query := `
		SELECT ` + deliverySelectCols + `
		FROM deliveries d
		JOIN orders o ON o.id = d.order_id
		WHERE ` + strings.Join(cond, " AND ") + `
		ORDER BY d.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("DeliveryRepository.GetByCourierID: %w", err)
	}
	defer rows.Close()

	var deliveries []*model.Delivery
	for rows.Next() {
		d, err := scanDelivery(rows)
		if err != nil {
			return nil, fmt.Errorf("DeliveryRepository.GetByCourierID scan: %w", err)
		}
		deliveries = append(deliveries, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("DeliveryRepository.GetByCourierID rows: %w", err)
	}
	return deliveries, nil
}

// ── Update ────────────────────────────────────────────────────────────────────

func (r *DeliveryRepository) Update(ctx context.Context, d *model.Delivery) error {
	err := r.db.QueryRowContext(ctx, `
		UPDATE deliveries
		SET courier_id       = $2,
		    pickup_status    = $3,
		    delivery_status  = $4,
		    pickup_notes     = $5,
		    delivery_notes   = $6,
		    picked_up_at     = $7,
		    delivered_at     = $8,
		    updated_at       = now()
		WHERE id = $1
		RETURNING updated_at
	`, d.ID,
		d.CourierID,
		d.PickupStatus, d.DeliveryStatus,
		d.PickupNotes, d.DeliveryNotes,
		d.PickedUpAt, d.DeliveredAt,
	).Scan(&d.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("DeliveryRepository.Update: %w", err)
	}
	return nil
}
