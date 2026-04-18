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

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// OrderFilters dipakai di ListByOutletID.
type OrderFilters struct {
	Status      *model.OrderStatus
	KasirID     *uuid.UUID
	TreatmentID *uuid.UUID
	DateFrom    *time.Time
	DateTo      *time.Time
	Search      string // order_number
}

const orderSelectCols = `
	id, order_number, outlet_id, customer_id, kasir_id, treatment_id,
	treatment_name, material, status, base_price, delivery_fee, total_price,
	is_price_edited, original_price, condition_notes, is_pickup, is_delivery,
	estimated_done_at, cancel_reason, cancelled_by, cancelled_at,
	created_at, updated_at
`

func scanOrder(row interface{ Scan(...any) error }) (*model.Order, error) {
	o := &model.Order{}
	err := row.Scan(
		&o.ID, &o.OrderNumber, &o.OutletID, &o.CustomerID, &o.KasirID, &o.TreatmentID,
		&o.TreatmentName, &o.Material, &o.Status, &o.BasePrice, &o.DeliveryFee, &o.TotalPrice,
		&o.IsPriceEdited, &o.OriginalPrice, &o.ConditionNotes, &o.IsPickup, &o.IsDelivery,
		&o.EstimatedDoneAt, &o.CancelReason, &o.CancelledBy, &o.CancelledAt,
		&o.CreatedAt, &o.UpdatedAt,
	)
	return o, err
}

// Create menyimpan order baru dalam transaction.
func (r *OrderRepository) Create(ctx context.Context, tx *sql.Tx, o *model.Order) error {
	query := `
		INSERT INTO orders (
			order_number, outlet_id, customer_id, kasir_id, treatment_id,
			treatment_name, material, status, base_price, delivery_fee, total_price,
			condition_notes, is_pickup, is_delivery, estimated_done_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		RETURNING id, is_price_edited, original_price, created_at, updated_at
	`
	err := tx.QueryRowContext(ctx, query,
		o.OrderNumber, o.OutletID, o.CustomerID, o.KasirID, o.TreatmentID,
		o.TreatmentName, o.Material, o.Status, o.BasePrice, o.DeliveryFee, o.TotalPrice,
		o.ConditionNotes, o.IsPickup, o.IsDelivery, o.EstimatedDoneAt,
	).Scan(&o.ID, &o.IsPriceEdited, &o.OriginalPrice, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return fmt.Errorf("OrderRepository.Create: %w", err)
	}
	return nil
}

// GetByID mengambil order berdasarkan ID dengan validasi outlet_id.
func (r *OrderRepository) GetByID(ctx context.Context, orderID, outletID uuid.UUID) (*model.Order, error) {
	query := `SELECT` + orderSelectCols + `FROM orders WHERE id = $1 AND outlet_id = $2`
	o, err := scanOrder(r.db.QueryRowContext(ctx, query, orderID, outletID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("OrderRepository.GetByID: %w", err)
	}
	return o, nil
}

// ListByOutletID mengambil list order dengan filter dan pagination.
func (r *OrderRepository) ListByOutletID(ctx context.Context, outletID uuid.UUID, f OrderFilters, page, limit int) ([]*model.Order, int, error) {
	args := []any{outletID}
	conditions := []string{"outlet_id = $1"}

	if f.Status != nil {
		args = append(args, *f.Status)
		conditions = append(conditions, fmt.Sprintf("status = $%d", len(args)))
	}
	if f.KasirID != nil {
		args = append(args, *f.KasirID)
		conditions = append(conditions, fmt.Sprintf("kasir_id = $%d", len(args)))
	}
	if f.TreatmentID != nil {
		args = append(args, *f.TreatmentID)
		conditions = append(conditions, fmt.Sprintf("treatment_id = $%d", len(args)))
	}
	if f.DateFrom != nil {
		args = append(args, *f.DateFrom)
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", len(args)))
	}
	if f.DateTo != nil {
		args = append(args, *f.DateTo)
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", len(args)))
	}
	if f.Search != "" {
		args = append(args, "%"+f.Search+"%")
		conditions = append(conditions, fmt.Sprintf("order_number ILIKE $%d", len(args)))
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM orders `+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("OrderRepository.ListByOutletID count: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit
	args = append(args, limit, offset)

	query := `SELECT` + orderSelectCols + `FROM orders ` +
		where + fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, len(args)-1, len(args))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("OrderRepository.ListByOutletID: %w", err)
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("OrderRepository.ListByOutletID scan: %w", err)
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("OrderRepository.ListByOutletID rows: %w", err)
	}
	return orders, total, nil
}

// Update memperbarui status dan field yang bisa berubah pada order.
func (r *OrderRepository) Update(ctx context.Context, o *model.Order) error {
	query := `
		UPDATE orders SET
			status = $1, total_price = $2, is_price_edited = $3, original_price = $4,
			cancel_reason = $5, cancelled_by = $6, cancelled_at = $7
		WHERE id = $8 AND outlet_id = $9
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		o.Status, o.TotalPrice, o.IsPriceEdited, o.OriginalPrice,
		o.CancelReason, o.CancelledBy, o.CancelledAt,
		o.ID, o.OutletID,
	).Scan(&o.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("OrderRepository.Update: %w", err)
	}
	return nil
}

// CreateWithTimestamp menyimpan order baru dengan created_at eksplisit.
// Digunakan oleh SyncService untuk menjaga timestamp asli dari perangkat offline.
func (r *OrderRepository) CreateWithTimestamp(ctx context.Context, tx *sql.Tx, o *model.Order, createdAt time.Time) error {
	query := `
		INSERT INTO orders (
			order_number, outlet_id, customer_id, kasir_id, treatment_id,
			treatment_name, material, status, base_price, delivery_fee, total_price,
			condition_notes, is_pickup, is_delivery, estimated_done_at, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		RETURNING id, is_price_edited, original_price, created_at, updated_at
	`
	err := tx.QueryRowContext(ctx, query,
		o.OrderNumber, o.OutletID, o.CustomerID, o.KasirID, o.TreatmentID,
		o.TreatmentName, o.Material, o.Status, o.BasePrice, o.DeliveryFee, o.TotalPrice,
		o.ConditionNotes, o.IsPickup, o.IsDelivery, o.EstimatedDoneAt, createdAt,
	).Scan(&o.ID, &o.IsPriceEdited, &o.OriginalPrice, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return fmt.Errorf("OrderRepository.CreateWithTimestamp: %w", err)
	}
	return nil
}

// GenerateOrderNumber menghasilkan nomor order yang unik secara atomic menggunakan
// INSERT ... ON CONFLICT DO UPDATE dalam transaction yang sama dengan pembuatan order.
// Format: ORD-{outletCode}-{YYYYMMDD}-{seq padded 3 digit}
// Sesuai rules-backend.md section 6.
func (r *OrderRepository) GenerateOrderNumber(ctx context.Context, tx *sql.Tx, outletID uuid.UUID, outletCode string) (string, error) {
	query := `
		INSERT INTO order_sequences (outlet_id, date, last_seq)
		VALUES ($1, CURRENT_DATE, 1)
		ON CONFLICT (outlet_id, date)
		DO UPDATE SET last_seq = order_sequences.last_seq + 1
		RETURNING last_seq
	`
	var seq int
	err := tx.QueryRowContext(ctx, query, outletID).Scan(&seq)
	if err != nil {
		return "", fmt.Errorf("GenerateOrderNumber: %w", err)
	}
	date := time.Now().Format("20060102")
	return fmt.Sprintf("ORD-%s-%s-%03d", outletCode, date, seq), nil
}

// GenerateOrderNumberForDate sama dengan GenerateOrderNumber tetapi menggunakan
// tanggal eksplisit. Digunakan oleh SyncService agar nomor order mencerminkan
// tanggal asli dari perangkat offline (created_at_local).
func (r *OrderRepository) GenerateOrderNumberForDate(ctx context.Context, tx *sql.Tx, outletID uuid.UUID, outletCode string, date time.Time) (string, error) {
	query := `
		INSERT INTO order_sequences (outlet_id, date, last_seq)
		VALUES ($1, $2::date, 1)
		ON CONFLICT (outlet_id, date)
		DO UPDATE SET last_seq = order_sequences.last_seq + 1
		RETURNING last_seq
	`
	var seq int
	err := tx.QueryRowContext(ctx, query, outletID, date.Format("2006-01-02")).Scan(&seq)
	if err != nil {
		return "", fmt.Errorf("GenerateOrderNumberForDate: %w", err)
	}
	return fmt.Sprintf("ORD-%s-%s-%03d", outletCode, date.Format("20060102"), seq), nil
}
