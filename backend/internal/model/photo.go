package model

import (
	"time"

	"github.com/google/uuid"
)

type PhotoType string

const (
	PhotoTypeBefore PhotoType = "before"
	PhotoTypeAfter  PhotoType = "after"
)

type Photo struct {
	ID         uuid.UUID  `json:"id"`
	OrderID    uuid.UUID  `json:"order_id"`
	Type       PhotoType  `json:"type"`
	R2Key      string     `json:"r2_key"`
	FileSizeKB *int       `json:"file_size_kb"`
	IsDeleted  bool       `json:"is_deleted"`
	UploadedAt time.Time  `json:"uploaded_at"`
	DeletedAt  *time.Time `json:"deleted_at"`
}
