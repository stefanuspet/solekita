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

type SubscriptionRepository struct {
	db *sql.DB
}

func NewSubscriptionRepository(db *sql.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// Create menyimpan subscription baru. Gunakan tx untuk operasi dalam transaction.
func (r *SubscriptionRepository) Create(ctx context.Context, q dbtx, sub *model.Subscription) error {
	query := `
		INSERT INTO subscriptions (outlet_id, plan, price_per_month, trial_started_at, trial_ends_at,
		                           subscription_started_at, next_due_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`
	err := q.QueryRowContext(ctx, query,
		sub.OutletID,
		sub.Plan,
		sub.PricePerMonth,
		sub.TrialStartedAt,
		sub.TrialEndsAt,
		sub.SubscriptionStartedAt,
		sub.NextDueDate,
	).Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)

	if err != nil {
		return fmt.Errorf("SubscriptionRepository.Create: %w", err)
	}
	return nil
}

// GetByOutletID mengambil subscription berdasarkan outlet_id.
func (r *SubscriptionRepository) GetByOutletID(ctx context.Context, outletID uuid.UUID) (*model.Subscription, error) {
	query := `
		SELECT id, outlet_id, plan, price_per_month, trial_started_at, trial_ends_at,
		       subscription_started_at, next_due_date, suspended_at, inactive_at,
		       created_at, updated_at
		FROM subscriptions
		WHERE outlet_id = $1
	`
	sub := &model.Subscription{}
	err := r.db.QueryRowContext(ctx, query, outletID).Scan(
		&sub.ID,
		&sub.OutletID,
		&sub.Plan,
		&sub.PricePerMonth,
		&sub.TrialStartedAt,
		&sub.TrialEndsAt,
		&sub.SubscriptionStartedAt,
		&sub.NextDueDate,
		&sub.SuspendedAt,
		&sub.InactiveAt,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("SubscriptionRepository.GetByOutletID: %w", err)
	}
	return sub, nil
}

// UpdatePlan memperbarui plan, next_due_date, dan subscription_started_at dalam transaction.
// Digunakan saat webhook Tripay konfirmasi pembayaran berhasil.
func (r *SubscriptionRepository) UpdatePlan(ctx context.Context, q dbtx, subscriptionID uuid.UUID, plan model.SubscriptionPlan, nextDueDate, subscriptionStartedAt *time.Time) error {
	err := q.QueryRowContext(ctx, `
		UPDATE subscriptions
		SET plan                    = $2,
		    next_due_date           = $3,
		    subscription_started_at = COALESCE($4, subscription_started_at),
		    updated_at              = now()
		WHERE id = $1
		RETURNING updated_at
	`, subscriptionID, plan, nextDueDate, subscriptionStartedAt).Scan(new(interface{}))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("SubscriptionRepository.UpdatePlan: %w", err)
	}
	return nil
}

// TrialOutletRow adalah baris hasil query trial yang membutuhkan owner phone untuk WA.
type TrialOutletRow struct {
	OutletID    uuid.UUID
	OutletName  string
	OwnerPhone  string
	TrialEndsAt time.Time
}

// ListTrialEndingOn mencari outlet yang trial_ends_at jatuh tepat pada tanggal tertentu
// dan masih dalam status trial. Digunakan oleh scheduler H-3, H-1, dan H-0.
func (r *SubscriptionRepository) ListTrialEndingOn(ctx context.Context, date time.Time) ([]*TrialOutletRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT s.outlet_id, o.name, u.phone, s.trial_ends_at
		FROM subscriptions s
		JOIN outlets o ON o.id = s.outlet_id
		JOIN users u ON u.id = o.owner_id
		WHERE s.trial_ends_at::date = $1::date
		  AND o.subscription_status = 'trial'
	`, date.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("SubscriptionRepository.ListTrialEndingOn: %w", err)
	}
	defer rows.Close()

	var result []*TrialOutletRow
	for rows.Next() {
		row := &TrialOutletRow{}
		if err := rows.Scan(&row.OutletID, &row.OutletName, &row.OwnerPhone, &row.TrialEndsAt); err != nil {
			return nil, fmt.Errorf("SubscriptionRepository.ListTrialEndingOn scan: %w", err)
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// ListExpiredTrials mencari outlet yang trial_ends_at sudah lewat dan masih berstatus trial.
// Digunakan oleh scheduler untuk suspend outlet yang tidak membayar setelah trial.
func (r *SubscriptionRepository) ListExpiredTrials(ctx context.Context) ([]*TrialOutletRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT s.outlet_id, o.name, u.phone, s.trial_ends_at
		FROM subscriptions s
		JOIN outlets o ON o.id = s.outlet_id
		JOIN users u ON u.id = o.owner_id
		WHERE s.trial_ends_at < now()
		  AND o.subscription_status = 'trial'
	`)
	if err != nil {
		return nil, fmt.Errorf("SubscriptionRepository.ListExpiredTrials: %w", err)
	}
	defer rows.Close()

	var result []*TrialOutletRow
	for rows.Next() {
		row := &TrialOutletRow{}
		if err := rows.Scan(&row.OutletID, &row.OutletName, &row.OwnerPhone, &row.TrialEndsAt); err != nil {
			return nil, fmt.Errorf("SubscriptionRepository.ListExpiredTrials scan: %w", err)
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// BillingOutletRow adalah baris hasil query billing yang membutuhkan data owner untuk WA & Tripay.
type BillingOutletRow struct {
	OutletID       uuid.UUID
	OutletName     string
	OwnerPhone     string
	OwnerName      string
	Plan           model.SubscriptionPlan
	PricePerMonth  int
	SubscriptionID uuid.UUID
	NextDueDate    *time.Time
}

const billingJoinSelect = `
	SELECT s.outlet_id, o.name, u.phone, u.name, s.plan, s.price_per_month, s.id, s.next_due_date
	FROM subscriptions s
	JOIN outlets o ON o.id = s.outlet_id
	JOIN users u ON u.id = o.owner_id
`

func scanBillingRows(rows *sql.Rows) ([]*BillingOutletRow, error) {
	var result []*BillingOutletRow
	for rows.Next() {
		row := &BillingOutletRow{}
		if err := rows.Scan(
			&row.OutletID, &row.OutletName, &row.OwnerPhone, &row.OwnerName,
			&row.Plan, &row.PricePerMonth, &row.SubscriptionID, &row.NextDueDate,
		); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// ListActiveWithDueDateOn mencari outlet aktif yang next_due_date jatuh pada tanggal tertentu.
// Digunakan oleh billing scheduler: GenerateMonthlyInvoices (H-0), SendBillingReminderH3, H1.
func (r *SubscriptionRepository) ListActiveWithDueDateOn(ctx context.Context, date time.Time) ([]*BillingOutletRow, error) {
	rows, err := r.db.QueryContext(ctx,
		billingJoinSelect+`
		WHERE o.subscription_status = 'active'
		  AND s.next_due_date::date = $1::date
	`, date.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("SubscriptionRepository.ListActiveWithDueDateOn: %w", err)
	}
	defer rows.Close()
	result, err := scanBillingRows(rows)
	if err != nil {
		return nil, fmt.Errorf("SubscriptionRepository.ListActiveWithDueDateOn scan: %w", err)
	}
	return result, nil
}

// ListActiveOverdue mencari outlet aktif yang next_due_date + 3 hari sudah lewat.
// Digunakan oleh SuspendUnpaidOutlets (grace period 3 hari setelah jatuh tempo).
func (r *SubscriptionRepository) ListActiveOverdue(ctx context.Context) ([]*BillingOutletRow, error) {
	rows, err := r.db.QueryContext(ctx,
		billingJoinSelect+`
		WHERE o.subscription_status = 'active'
		  AND s.next_due_date < now() - INTERVAL '3 days'
	`)
	if err != nil {
		return nil, fmt.Errorf("SubscriptionRepository.ListActiveOverdue: %w", err)
	}
	defer rows.Close()
	result, err := scanBillingRows(rows)
	if err != nil {
		return nil, fmt.Errorf("SubscriptionRepository.ListActiveOverdue scan: %w", err)
	}
	return result, nil
}

// ListSuspendedOlderThan mencari outlet yang sudah tersuspend > 30 hari.
// Digunakan oleh MarkInactiveOutlets.
func (r *SubscriptionRepository) ListSuspendedOlderThan(ctx context.Context, days int) ([]*BillingOutletRow, error) {
	rows, err := r.db.QueryContext(ctx,
		billingJoinSelect+`
		WHERE o.subscription_status = 'suspended'
		  AND s.suspended_at IS NOT NULL
		  AND s.suspended_at < now() - ($1 * INTERVAL '1 day')
	`, days)
	if err != nil {
		return nil, fmt.Errorf("SubscriptionRepository.ListSuspendedOlderThan: %w", err)
	}
	defer rows.Close()
	result, err := scanBillingRows(rows)
	if err != nil {
		return nil, fmt.Errorf("SubscriptionRepository.ListSuspendedOlderThan scan: %w", err)
	}
	return result, nil
}

// ClearSuspendedAt menghapus suspended_at saat outlet diaktifkan kembali oleh admin.
func (r *SubscriptionRepository) ClearSuspendedAt(ctx context.Context, outletID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE subscriptions SET suspended_at = NULL, updated_at = now()
		WHERE outlet_id = $1
	`, outletID)
	if err != nil {
		return fmt.Errorf("SubscriptionRepository.ClearSuspendedAt: %w", err)
	}
	return nil
}

// UpdateSuspendedAt mencatat waktu suspend pada tabel subscriptions.
// Dipanggil bersama OutletRepository.UpdateSubscriptionStatus saat menyuspend outlet.
func (r *SubscriptionRepository) UpdateSuspendedAt(ctx context.Context, outletID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE subscriptions SET suspended_at = now(), updated_at = now()
		WHERE outlet_id = $1
	`, outletID)
	if err != nil {
		return fmt.Errorf("SubscriptionRepository.UpdateSuspendedAt: %w", err)
	}
	return nil
}

// Update memperbarui data subscription.
func (r *SubscriptionRepository) Update(ctx context.Context, sub *model.Subscription) error {
	query := `
		UPDATE subscriptions
		SET plan                    = $1,
		    price_per_month         = $2,
		    trial_started_at        = $3,
		    trial_ends_at           = $4,
		    subscription_started_at = $5,
		    next_due_date           = $6,
		    suspended_at            = $7,
		    inactive_at             = $8
		WHERE id = $9
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		sub.Plan,
		sub.PricePerMonth,
		sub.TrialStartedAt,
		sub.TrialEndsAt,
		sub.SubscriptionStartedAt,
		sub.NextDueDate,
		sub.SuspendedAt,
		sub.InactiveAt,
		sub.ID,
	).Scan(&sub.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("SubscriptionRepository.Update: %w", err)
	}
	return nil
}
