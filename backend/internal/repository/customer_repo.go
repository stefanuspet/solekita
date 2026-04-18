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

type CustomerRepository struct {
	db *sql.DB
}

func NewCustomerRepository(db *sql.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

const customerSelectCols = `
	id, outlet_id, name, phone, total_orders, last_order_at,
	notes, is_blacklisted, blacklist_reason, created_at, updated_at
`

func scanCustomer(row interface{ Scan(...any) error }) (*model.Customer, error) {
	c := &model.Customer{}
	err := row.Scan(
		&c.ID, &c.OutletID, &c.Name, &c.Phone, &c.TotalOrders, &c.LastOrderAt,
		&c.Notes, &c.IsBlacklisted, &c.BlacklistReason, &c.CreatedAt, &c.UpdatedAt,
	)
	return c, err
}

func (r *CustomerRepository) Create(ctx context.Context, c *model.Customer) error {
	query := `
		INSERT INTO customers (outlet_id, name, phone, notes)
		VALUES ($1, $2, $3, $4)
		RETURNING id, total_orders, last_order_at, is_blacklisted, blacklist_reason, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		c.OutletID, c.Name, c.Phone, c.Notes,
	).Scan(
		&c.ID, &c.TotalOrders, &c.LastOrderAt,
		&c.IsBlacklisted, &c.BlacklistReason, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("CustomerRepository.Create: %w", err)
	}
	return nil
}

func (r *CustomerRepository) GetByID(ctx context.Context, id, outletID uuid.UUID) (*model.Customer, error) {
	query := `SELECT` + customerSelectCols + `FROM customers WHERE id = $1 AND outlet_id = $2`
	c, err := scanCustomer(r.db.QueryRowContext(ctx, query, id, outletID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("CustomerRepository.GetByID: %w", err)
	}
	return c, nil
}

// GetByPhone mencari customer berdasarkan nomor HP dalam satu outlet.
// Phone bersifat unik per outlet (bukan global).
func (r *CustomerRepository) GetByPhone(ctx context.Context, outletID uuid.UUID, phone string) (*model.Customer, error) {
	query := `SELECT` + customerSelectCols + `FROM customers WHERE outlet_id = $1 AND phone = $2`
	c, err := scanCustomer(r.db.QueryRowContext(ctx, query, outletID, phone))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("CustomerRepository.GetByPhone: %w", err)
	}
	return c, nil
}

// ListByOutletID mengambil daftar customer dengan pencarian nama/phone dan pagination.
func (r *CustomerRepository) ListByOutletID(ctx context.Context, outletID uuid.UUID, search string, page, limit int) ([]*model.Customer, int, error) {
	args := []any{outletID}
	where := "WHERE outlet_id = $1"

	if search != "" {
		args = append(args, "%"+search+"%")
		where += fmt.Sprintf(" AND (name ILIKE $%d OR phone ILIKE $%d)", len(args), len(args))
	}

	// total count
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM customers `+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("CustomerRepository.ListByOutletID count: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit
	args = append(args, limit, offset)

	query := `SELECT` + customerSelectCols + `FROM customers ` +
		where + fmt.Sprintf(` ORDER BY name ASC LIMIT $%d OFFSET $%d`, len(args)-1, len(args))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("CustomerRepository.ListByOutletID: %w", err)
	}
	defer rows.Close()

	var customers []*model.Customer
	for rows.Next() {
		c, err := scanCustomer(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("CustomerRepository.ListByOutletID scan: %w", err)
		}
		customers = append(customers, c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("CustomerRepository.ListByOutletID rows: %w", err)
	}
	return customers, total, nil
}

func (r *CustomerRepository) Update(ctx context.Context, c *model.Customer) error {
	query := `
		UPDATE customers
		SET name = $1, phone = $2, notes = $3, is_blacklisted = $4, blacklist_reason = $5
		WHERE id = $6 AND outlet_id = $7
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		c.Name, c.Phone, c.Notes, c.IsBlacklisted, c.BlacklistReason, c.ID, c.OutletID,
	).Scan(&c.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("CustomerRepository.Update: %w", err)
	}
	return nil
}
