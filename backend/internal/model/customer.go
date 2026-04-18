package model

import (
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID              uuid.UUID  `json:"id"`
	OutletID        uuid.UUID  `json:"outlet_id"`
	Name            string     `json:"name"`
	Phone           string     `json:"phone"`
	TotalOrders     int        `json:"total_orders"`
	LastOrderAt     *time.Time `json:"last_order_at"`
	Notes           *string    `json:"notes"`
	IsBlacklisted   bool       `json:"is_blacklisted"`
	BlacklistReason *string    `json:"blacklist_reason"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
