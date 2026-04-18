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

type AdminRepository struct {
	db *sql.DB
}

func NewAdminRepository(db *sql.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

// ── Response types ────────────────────────────────────────────────────────────

// AdminOutletRow adalah satu baris di daftar outlet untuk admin panel.
type AdminOutletRow struct {
	ID                    uuid.UUID                `json:"id"`
	Name                  string                   `json:"name"`
	Code                  string                   `json:"code"`
	OwnerName             string                   `json:"owner_name"`
	OwnerPhone            string                   `json:"owner_phone"`
	SubscriptionStatus    model.SubscriptionStatus `json:"subscription_status"`
	Plan                  model.SubscriptionPlan   `json:"plan"`
	TrialEndsAt           *time.Time               `json:"trial_ends_at"`
	SubscriptionStartedAt *time.Time               `json:"subscription_started_at"`
	NextDueDate           *time.Time               `json:"next_due_date"`
	TotalOrders           int                      `json:"total_orders"`
	OrdersThisMonth       int                      `json:"orders_this_month"`
	CreatedAt             time.Time                `json:"created_at"`
}

// AdminOutletFilters adalah filter untuk ListOutlets.
type AdminOutletFilters struct {
	Status        *model.SubscriptionStatus // filter berdasarkan subscription_status
	TrialInactive bool                      // outlet trial yang belum pernah buat order
	Search        string                    // cari berdasarkan nama outlet atau nomor HP owner
}

// AdminOutletDetail adalah detail lengkap satu outlet untuk admin.
type AdminOutletDetail struct {
	ID                    uuid.UUID                `json:"id"`
	Name                  string                   `json:"name"`
	Code                  string                   `json:"code"`
	Address               *string                  `json:"address"`
	Phone                 *string                  `json:"phone"`
	OwnerName             string                   `json:"owner_name"`
	OwnerPhone            string                   `json:"owner_phone"`
	SubscriptionStatus    model.SubscriptionStatus `json:"subscription_status"`
	Plan                  model.SubscriptionPlan   `json:"plan"`
	PricePerMonth         int                      `json:"price_per_month"`
	TrialStartedAt        *time.Time               `json:"trial_started_at"`
	TrialEndsAt           *time.Time               `json:"trial_ends_at"`
	SubscriptionStartedAt *time.Time               `json:"subscription_started_at"`
	NextDueDate           *time.Time               `json:"next_due_date"`
	SuspendedAt           *time.Time               `json:"suspended_at"`
	TotalOrders           int                      `json:"total_orders"`
	OrdersThisMonth       int                      `json:"orders_this_month"`
	StaffCount            int                      `json:"staff_count"`
	CreatedAt             time.Time                `json:"created_at"`
}

// AdminSummaryStats adalah hitungan outlet per status untuk header dashboard admin.
type AdminSummaryStats struct {
	Total     int `json:"total"`
	Trial     int `json:"trial"`
	Active    int `json:"active"`
	Suspended int `json:"suspended"`
	Inactive  int `json:"inactive"`
	Cancelled int `json:"cancelled"`
}

// ── ListOutlets ───────────────────────────────────────────────────────────────

// ListOutlets mengembalikan daftar outlet dengan filter, pencarian, dan pagination.
// Filter status, trial_inactive (trial tapi nol order), dan search (nama/HP owner).
func (r *AdminRepository) ListOutlets(ctx context.Context, f AdminOutletFilters, page, limit int) ([]*AdminOutletRow, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	args := []any{}
	conditions := []string{"u.is_owner = true"}

	if f.Status != nil {
		args = append(args, string(*f.Status))
		conditions = append(conditions, fmt.Sprintf("o.subscription_status = $%d", len(args)))
	}

	if f.TrialInactive {
		conditions = append(conditions, "o.subscription_status = 'trial'")
		conditions = append(conditions, "NOT EXISTS (SELECT 1 FROM orders WHERE outlet_id = o.id)")
	}

	if f.Search != "" {
		args = append(args, "%"+f.Search+"%")
		idx := len(args)
		conditions = append(conditions, fmt.Sprintf("(o.name ILIKE $%d OR u.phone ILIKE $%d)", idx, idx))
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	// Count total
	countQuery := `
		SELECT COUNT(DISTINCT o.id)
		FROM outlets o
		JOIN users u ON u.id = o.owner_id
		LEFT JOIN subscriptions s ON s.outlet_id = o.id
		` + where

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("AdminRepository.ListOutlets count: %w", err)
	}

	// Paginated list
	args = append(args, limit, offset)
	listQuery := `
		SELECT
			o.id, o.name, o.code,
			u.name  AS owner_name,
			u.phone AS owner_phone,
			o.subscription_status,
			COALESCE(s.plan, 'monthly')             AS plan,
			s.trial_ends_at,
			s.subscription_started_at,
			s.next_due_date,
			COUNT(DISTINCT ord.id)                  AS total_orders,
			COUNT(DISTINCT ord.id) FILTER (
				WHERE ord.created_at >= date_trunc('month', now())
			)                                       AS orders_this_month,
			o.created_at
		FROM outlets o
		JOIN users u ON u.id = o.owner_id
		LEFT JOIN subscriptions s ON s.outlet_id = o.id
		LEFT JOIN orders ord ON ord.outlet_id = o.id
			AND ord.status != 'dibatalkan'
		` + where + `
		GROUP BY o.id, u.name, u.phone,
		         s.plan, s.trial_ends_at, s.subscription_started_at, s.next_due_date
		ORDER BY o.created_at DESC
		LIMIT $` + fmt.Sprintf("%d", len(args)-1) + ` OFFSET $` + fmt.Sprintf("%d", len(args))

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("AdminRepository.ListOutlets: %w", err)
	}
	defer rows.Close()

	var outlets []*AdminOutletRow
	for rows.Next() {
		row := &AdminOutletRow{}
		if err := rows.Scan(
			&row.ID, &row.Name, &row.Code,
			&row.OwnerName, &row.OwnerPhone,
			&row.SubscriptionStatus, &row.Plan,
			&row.TrialEndsAt, &row.SubscriptionStartedAt, &row.NextDueDate,
			&row.TotalOrders, &row.OrdersThisMonth,
			&row.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("AdminRepository.ListOutlets scan: %w", err)
		}
		outlets = append(outlets, row)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("AdminRepository.ListOutlets rows: %w", err)
	}
	return outlets, total, nil
}

// ── GetOutletDetail ───────────────────────────────────────────────────────────

// GetOutletDetail mengembalikan detail lengkap satu outlet: data outlet, owner,
// subscription, plus stats (total order, order bulan ini, jumlah karyawan aktif).
func (r *AdminRepository) GetOutletDetail(ctx context.Context, outletID uuid.UUID) (*AdminOutletDetail, error) {
	query := `
		SELECT
			o.id, o.name, o.code, o.address, o.phone,
			u.name  AS owner_name,
			u.phone AS owner_phone,
			o.subscription_status,
			COALESCE(s.plan, 'monthly')        AS plan,
			COALESCE(s.price_per_month, 0)     AS price_per_month,
			s.trial_started_at,
			s.trial_ends_at,
			s.subscription_started_at,
			s.next_due_date,
			s.suspended_at,
			COUNT(DISTINCT ord.id) FILTER (
				WHERE ord.status != 'dibatalkan'
			)                                  AS total_orders,
			COUNT(DISTINCT ord.id) FILTER (
				WHERE ord.status != 'dibatalkan'
				  AND ord.created_at >= date_trunc('month', now())
			)                                  AS orders_this_month,
			COUNT(DISTINCT staff.id) FILTER (
				WHERE staff.is_owner = false
				  AND staff.is_active = true
			)                                  AS staff_count,
			o.created_at
		FROM outlets o
		JOIN users u ON u.id = o.owner_id AND u.is_owner = true
		LEFT JOIN subscriptions s ON s.outlet_id = o.id
		LEFT JOIN orders ord ON ord.outlet_id = o.id
		LEFT JOIN users staff ON staff.outlet_id = o.id
		WHERE o.id = $1
		GROUP BY
			o.id, o.name, o.code, o.address, o.phone,
			u.name, u.phone,
			s.plan, s.price_per_month, s.trial_started_at, s.trial_ends_at,
			s.subscription_started_at, s.next_due_date, s.suspended_at
	`

	d := &AdminOutletDetail{}
	err := r.db.QueryRowContext(ctx, query, outletID).Scan(
		&d.ID, &d.Name, &d.Code, &d.Address, &d.Phone,
		&d.OwnerName, &d.OwnerPhone,
		&d.SubscriptionStatus, &d.Plan, &d.PricePerMonth,
		&d.TrialStartedAt, &d.TrialEndsAt,
		&d.SubscriptionStartedAt, &d.NextDueDate, &d.SuspendedAt,
		&d.TotalOrders, &d.OrdersThisMonth, &d.StaffCount,
		&d.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("AdminRepository.GetOutletDetail: %w", err)
	}
	return d, nil
}

// ── GetOutletSummaryStats ─────────────────────────────────────────────────────

// GetOutletSummaryStats mengembalikan hitungan outlet per subscription_status.
// Digunakan untuk header/card ringkasan di dashboard admin.
func (r *AdminRepository) GetOutletSummaryStats(ctx context.Context) (*AdminSummaryStats, error) {
	query := `
		SELECT
			COUNT(*)                                                    AS total,
			COUNT(*) FILTER (WHERE subscription_status = 'trial')      AS trial,
			COUNT(*) FILTER (WHERE subscription_status = 'active')     AS active,
			COUNT(*) FILTER (WHERE subscription_status = 'suspended')  AS suspended,
			COUNT(*) FILTER (WHERE subscription_status = 'inactive')   AS inactive,
			COUNT(*) FILTER (WHERE subscription_status = 'cancelled')  AS cancelled
		FROM outlets
	`
	s := &AdminSummaryStats{}
	err := r.db.QueryRowContext(ctx, query).Scan(
		&s.Total, &s.Trial, &s.Active, &s.Suspended, &s.Inactive, &s.Cancelled,
	)
	if err != nil {
		return nil, fmt.Errorf("AdminRepository.GetOutletSummaryStats: %w", err)
	}
	return s, nil
}
