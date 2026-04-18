package model

import (
	"time"

	"github.com/google/uuid"
)

type InvoiceStatus string

const (
	InvoiceStatusPending InvoiceStatus = "pending"
	InvoiceStatusPaid    InvoiceStatus = "paid"
	InvoiceStatusExpired InvoiceStatus = "expired"
	InvoiceStatusFailed  InvoiceStatus = "failed"
)

type SubscriptionInvoice struct {
	ID              uuid.UUID     `json:"id"`
	SubscriptionID  uuid.UUID     `json:"subscription_id"`
	OutletID        uuid.UUID     `json:"outlet_id"`
	Amount          int           `json:"amount"`
	DueDate         time.Time     `json:"due_date"`
	TripayReference *string       `json:"tripay_reference"`
	TripayPaymentURL *string      `json:"tripay_payment_url"`
	Status          InvoiceStatus `json:"status"`
	PaidAt          *time.Time    `json:"paid_at"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}
