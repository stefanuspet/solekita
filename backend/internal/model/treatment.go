package model

import (
	"time"

	"github.com/google/uuid"
)

type Treatment struct {
	ID        uuid.UUID `json:"id"`
	OutletID  uuid.UUID `json:"outlet_id"`
	Name      string    `json:"name"`
	Material  *string   `json:"material"`
	Price     int       `json:"price"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
