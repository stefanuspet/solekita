package repository

import (
	"context"
	"fmt"
	"time"

	"database/sql"

	"github.com/google/uuid"
)

type DashboardRepository struct {
	db *sql.DB
}

func NewDashboardRepository(db *sql.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

// ── Response types ────────────────────────────────────────────────────────────

type KasirSummary struct {
	KasirID    uuid.UUID `json:"kasir_id"`
	KasirName  string    `json:"kasir_name"`
	OrderCount int       `json:"order_count"`
	Revenue    int       `json:"revenue"`
}

type PaymentMethodSummary struct {
	Method     string `json:"method"`
	OrderCount int    `json:"order_count"`
	Amount     int    `json:"amount"`
}

type DailySummary struct {
	Date           time.Time              `json:"date"`
	Revenue        int                    `json:"revenue"`
	OrderCount     int                    `json:"order_count"`
	StaffSummary   []*KasirSummary        `json:"staff_summary"`
	PaymentSummary []*PaymentMethodSummary `json:"payment_summary"`
}

type OverdueOrder struct {
	ID             uuid.UUID `json:"id"`
	OrderNumber    string    `json:"order_number"`
	CustomerID     uuid.UUID `json:"customer_id"`
	CustomerName   string    `json:"customer_name"`
	CustomerPhone  string    `json:"customer_phone"`
	DaysOverdue    int       `json:"days_overdue"`
	TotalPrice     int       `json:"total_price"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type DailyStatPoint struct {
	Date       time.Time `json:"date"`
	Revenue    int       `json:"revenue"`
	OrderCount int       `json:"order_count"`
}

// ── GetDailySummary ───────────────────────────────────────────────────────────

func (r *DashboardRepository) GetDailySummary(ctx context.Context, outletID uuid.UUID, date time.Time) (*DailySummary, error) {
	dateStr := date.Format("2006-01-02")

	// Total revenue + order count (exclude dibatalkan)
	var revenue, orderCount int
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(p.amount), 0), COUNT(o.id)
		FROM orders o
		LEFT JOIN payments p ON p.order_id = o.id AND p.status = 'paid'
		WHERE o.outlet_id = $1
		  AND o.created_at::date = $2::date
		  AND o.status != 'dibatalkan'
	`, outletID, dateStr).Scan(&revenue, &orderCount)
	if err != nil {
		return nil, fmt.Errorf("DashboardRepository.GetDailySummary revenue: %w", err)
	}

	// Breakdown per kasir
	staffRows, err := r.db.QueryContext(ctx, `
		SELECT o.kasir_id, u.name,
		       COUNT(o.id) AS order_count,
		       COALESCE(SUM(p.amount), 0) AS revenue
		FROM orders o
		JOIN users u ON u.id = o.kasir_id
		LEFT JOIN payments p ON p.order_id = o.id AND p.status = 'paid'
		WHERE o.outlet_id = $1
		  AND o.created_at::date = $2::date
		  AND o.status != 'dibatalkan'
		GROUP BY o.kasir_id, u.name
		ORDER BY revenue DESC
	`, outletID, dateStr)
	if err != nil {
		return nil, fmt.Errorf("DashboardRepository.GetDailySummary staff: %w", err)
	}
	defer staffRows.Close()

	var staffSummary []*KasirSummary
	for staffRows.Next() {
		s := &KasirSummary{}
		if err := staffRows.Scan(&s.KasirID, &s.KasirName, &s.OrderCount, &s.Revenue); err != nil {
			return nil, fmt.Errorf("DashboardRepository.GetDailySummary staff scan: %w", err)
		}
		staffSummary = append(staffSummary, s)
	}
	if err := staffRows.Err(); err != nil {
		return nil, fmt.Errorf("DashboardRepository.GetDailySummary staff rows: %w", err)
	}

	// Breakdown per metode bayar
	payRows, err := r.db.QueryContext(ctx, `
		SELECT p.method, COUNT(p.id) AS order_count, COALESCE(SUM(p.amount), 0) AS amount
		FROM payments p
		JOIN orders o ON o.id = p.order_id
		WHERE o.outlet_id = $1
		  AND o.created_at::date = $2::date
		  AND o.status != 'dibatalkan'
		  AND p.status = 'paid'
		GROUP BY p.method
		ORDER BY amount DESC
	`, outletID, dateStr)
	if err != nil {
		return nil, fmt.Errorf("DashboardRepository.GetDailySummary payment: %w", err)
	}
	defer payRows.Close()

	var paymentSummary []*PaymentMethodSummary
	for payRows.Next() {
		p := &PaymentMethodSummary{}
		if err := payRows.Scan(&p.Method, &p.OrderCount, &p.Amount); err != nil {
			return nil, fmt.Errorf("DashboardRepository.GetDailySummary payment scan: %w", err)
		}
		paymentSummary = append(paymentSummary, p)
	}
	if err := payRows.Err(); err != nil {
		return nil, fmt.Errorf("DashboardRepository.GetDailySummary payment rows: %w", err)
	}

	return &DailySummary{
		Date:           date,
		Revenue:        revenue,
		OrderCount:     orderCount,
		StaffSummary:   staffSummary,
		PaymentSummary: paymentSummary,
	}, nil
}

// ── GetOverdueOrders ──────────────────────────────────────────────────────────

func (r *DashboardRepository) GetOverdueOrders(ctx context.Context, outletID uuid.UUID, thresholdDays int) ([]*OverdueOrder, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT o.id, o.order_number, o.customer_id,
		       c.name AS customer_name, c.phone AS customer_phone,
		       EXTRACT(DAY FROM now() - o.updated_at)::int AS days_overdue,
		       o.total_price, o.updated_at
		FROM orders o
		JOIN customers c ON c.id = o.customer_id
		WHERE o.outlet_id = $1
		  AND o.status = 'selesai'
		  AND o.updated_at < now() - ($2 || ' days')::interval
		ORDER BY o.updated_at ASC
	`, outletID, thresholdDays)
	if err != nil {
		return nil, fmt.Errorf("DashboardRepository.GetOverdueOrders: %w", err)
	}
	defer rows.Close()

	var orders []*OverdueOrder
	for rows.Next() {
		o := &OverdueOrder{}
		if err := rows.Scan(&o.ID, &o.OrderNumber, &o.CustomerID, &o.CustomerName, &o.CustomerPhone, &o.DaysOverdue, &o.TotalPrice, &o.UpdatedAt); err != nil {
			return nil, fmt.Errorf("DashboardRepository.GetOverdueOrders scan: %w", err)
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("DashboardRepository.GetOverdueOrders rows: %w", err)
	}
	return orders, nil
}

// ── GetOrderStats ─────────────────────────────────────────────────────────────

// GetOrderStats mengambil data statistik harian untuk grafik.
// period: jumlah hari ke belakang (7 atau 30).
// Menghasilkan satu data point per hari, termasuk hari tanpa order (revenue=0).
func (r *DashboardRepository) GetOrderStats(ctx context.Context, outletID uuid.UUID, period int) ([]*DailyStatPoint, error) {
	rows, err := r.db.QueryContext(ctx, `
		WITH date_series AS (
			SELECT generate_series(
				(CURRENT_DATE - ($2 - 1) * INTERVAL '1 day')::date,
				CURRENT_DATE,
				INTERVAL '1 day'
			)::date AS day
		)
		SELECT
			ds.day,
			COALESCE(SUM(p.amount), 0) AS revenue,
			COUNT(o.id) AS order_count
		FROM date_series ds
		LEFT JOIN orders o
			ON o.outlet_id = $1
			AND o.created_at::date = ds.day
			AND o.status != 'dibatalkan'
		LEFT JOIN payments p
			ON p.order_id = o.id
			AND p.status = 'paid'
		GROUP BY ds.day
		ORDER BY ds.day ASC
	`, outletID, period)
	if err != nil {
		return nil, fmt.Errorf("DashboardRepository.GetOrderStats: %w", err)
	}
	defer rows.Close()

	var stats []*DailyStatPoint
	for rows.Next() {
		s := &DailyStatPoint{}
		if err := rows.Scan(&s.Date, &s.Revenue, &s.OrderCount); err != nil {
			return nil, fmt.Errorf("DashboardRepository.GetOrderStats scan: %w", err)
		}
		stats = append(stats, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("DashboardRepository.GetOrderStats rows: %w", err)
	}
	return stats, nil
}
