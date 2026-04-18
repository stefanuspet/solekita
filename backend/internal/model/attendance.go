package model

import (
	"time"

	"github.com/google/uuid"
)

type AttendanceType string

const (
	AttendanceTypeMasuk  AttendanceType = "masuk"
	AttendanceTypeKeluar AttendanceType = "keluar"
)

type Attendance struct {
	ID              uuid.UUID      `json:"id"`
	UserID          uuid.UUID      `json:"user_id"`
	OutletID        uuid.UUID      `json:"outlet_id"`
	Type            AttendanceType `json:"type"`
	SelfieR2Key     *string        `json:"selfie_r2_key"`
	IsSelfieDeleted bool           `json:"is_selfie_deleted"`
	CreatedAt       time.Time      `json:"created_at"`
}
