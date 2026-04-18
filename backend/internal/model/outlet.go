package model

import (
	"time"

	"github.com/google/uuid"
)

type Outlet struct {
	ID                   uuid.UUID          `json:"id"`
	Name                 string             `json:"name"`
	Code                 string             `json:"code"`
	Address              *string            `json:"address"`
	Phone                *string            `json:"phone"`
	OwnerID              uuid.UUID          `json:"owner_id"`
	SubscriptionStatus   SubscriptionStatus `json:"subscription_status"`
	OverdueThresholdDays int                `json:"overdue_threshold_days"`
	IsActive             bool               `json:"is_active"`
	CreatedAt            time.Time          `json:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at"`
}
