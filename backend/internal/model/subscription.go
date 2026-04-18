package model

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionStatus string

const (
	SubscriptionStatusTrial     SubscriptionStatus = "trial"
	SubscriptionStatusActive    SubscriptionStatus = "active"
	SubscriptionStatusSuspended SubscriptionStatus = "suspended"
	SubscriptionStatusInactive  SubscriptionStatus = "inactive"
	SubscriptionStatusCancelled SubscriptionStatus = "cancelled"
)

type SubscriptionPlan string

const (
	SubscriptionPlanMonthly  SubscriptionPlan = "monthly"
	SubscriptionPlanBiannual SubscriptionPlan = "biannual"
)

type Subscription struct {
	ID                    uuid.UUID          `json:"id"`
	OutletID              uuid.UUID          `json:"outlet_id"`
	Plan                  SubscriptionPlan   `json:"plan"`
	PricePerMonth         int                `json:"price_per_month"`
	TrialStartedAt        *time.Time         `json:"trial_started_at"`
	TrialEndsAt           *time.Time         `json:"trial_ends_at"`
	SubscriptionStartedAt *time.Time         `json:"subscription_started_at"`
	NextDueDate           *time.Time         `json:"next_due_date"`
	SuspendedAt           *time.Time         `json:"suspended_at"`
	InactiveAt            *time.Time         `json:"inactive_at"`
	CreatedAt             time.Time          `json:"created_at"`
	UpdatedAt             time.Time          `json:"updated_at"`
}
