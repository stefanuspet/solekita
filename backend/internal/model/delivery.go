package model

import (
	"time"

	"github.com/google/uuid"
)

type PickupStatus string

const (
	PickupStatusPending   PickupStatus = "pending"
	PickupStatusOnTheWay  PickupStatus = "on_the_way"
	PickupStatusPickedUp  PickupStatus = "picked_up"
	PickupStatusFailed    PickupStatus = "failed"
)

type DeliveryStatus string

const (
	DeliveryStatusPending   DeliveryStatus = "pending"
	DeliveryStatusOnTheWay  DeliveryStatus = "on_the_way"
	DeliveryStatusDelivered DeliveryStatus = "delivered"
	DeliveryStatusFailed    DeliveryStatus = "failed"
)

type Delivery struct {
	ID              uuid.UUID      `json:"id"`
	OrderID         uuid.UUID      `json:"order_id"`
	CourierID       *uuid.UUID     `json:"courier_id"`
	PickupAddress   *string        `json:"pickup_address"`
	DeliveryAddress *string        `json:"delivery_address"`
	PickupStatus    PickupStatus   `json:"pickup_status"`
	DeliveryStatus  DeliveryStatus `json:"delivery_status"`
	PickupNotes     *string        `json:"pickup_notes"`
	DeliveryNotes   *string        `json:"delivery_notes"`
	PickedUpAt      *time.Time     `json:"picked_up_at"`
	DeliveredAt     *time.Time     `json:"delivered_at"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}
