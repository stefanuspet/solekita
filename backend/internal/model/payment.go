package model

import (
	"time"

	"github.com/google/uuid"
)

type PaymentMethod string

const (
	PaymentMethodCash     PaymentMethod = "cash"
	PaymentMethodTransfer PaymentMethod = "transfer"
	PaymentMethodQRIS     PaymentMethod = "qris"
	PaymentMethodLainnya  PaymentMethod = "lainnya"
)

type PaymentStatus string

const (
	PaymentStatusPaid     PaymentStatus = "paid"
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusRefunded PaymentStatus = "refunded"
)

type Payment struct {
	ID        uuid.UUID     `json:"id"`
	OrderID   uuid.UUID     `json:"order_id"`
	Amount    int           `json:"amount"`
	Method    PaymentMethod `json:"method"`
	Status    PaymentStatus `json:"status"`
	Notes     *string       `json:"notes"`
	PaidAt    *time.Time    `json:"paid_at"`
	CreatedAt time.Time     `json:"created_at"`
}
