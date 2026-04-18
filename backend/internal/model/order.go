package model

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusDijemput   OrderStatus = "dijemput"
	OrderStatusBaru       OrderStatus = "baru"
	OrderStatusProses     OrderStatus = "proses"
	OrderStatusSelesai    OrderStatus = "selesai"
	OrderStatusDiambil    OrderStatus = "diambil"
	OrderStatusDiantar    OrderStatus = "diantar"
	OrderStatusDibatalkan OrderStatus = "dibatalkan"
)

type Order struct {
	ID               uuid.UUID   `json:"id"`
	OrderNumber      string      `json:"order_number"`
	OutletID         uuid.UUID   `json:"outlet_id"`
	CustomerID       uuid.UUID   `json:"customer_id"`
	KasirID          uuid.UUID   `json:"kasir_id"`
	TreatmentID      uuid.UUID   `json:"treatment_id"`
	TreatmentName    string      `json:"treatment_name"`
	Material         string      `json:"material"`
	Status           OrderStatus `json:"status"`
	BasePrice        int         `json:"base_price"`
	DeliveryFee      int         `json:"delivery_fee"`
	TotalPrice       int         `json:"total_price"`
	IsPriceEdited    bool        `json:"is_price_edited"`
	OriginalPrice    *int        `json:"original_price"`
	ConditionNotes   *string     `json:"condition_notes"`
	IsPickup         bool        `json:"is_pickup"`
	IsDelivery       bool        `json:"is_delivery"`
	EstimatedDoneAt  *time.Time  `json:"estimated_done_at"`
	CancelReason     *string     `json:"cancel_reason"`
	CancelledBy      *uuid.UUID  `json:"cancelled_by"`
	CancelledAt      *time.Time  `json:"cancelled_at"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}
