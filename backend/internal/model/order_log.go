package model

import (
	"time"

	"github.com/google/uuid"
)

type OrderAction string

const (
	OrderActionCreated         OrderAction = "created"
	OrderActionStatusChanged   OrderAction = "status_changed"
	OrderActionPriceEdited     OrderAction = "price_edited"
	OrderActionPhotoAdded      OrderAction = "photo_added"
	OrderActionCancelled       OrderAction = "cancelled"
	OrderActionDeliveryUpdated OrderAction = "delivery_updated"
)

type OrderLog struct {
	ID        uuid.UUID   `json:"id"`
	OrderID   uuid.UUID   `json:"order_id"`
	UserID    uuid.UUID   `json:"user_id"`
	Action    OrderAction `json:"action"`
	OldValue  *string     `json:"old_value"`
	NewValue  *string     `json:"new_value"`
	Notes     *string     `json:"notes"`
	CreatedAt time.Time   `json:"created_at"`
}
