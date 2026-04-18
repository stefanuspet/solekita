package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID  `json:"id"`
	OutletID     uuid.UUID  `json:"outlet_id"`
	Name         string     `json:"name"`
	Phone        string     `json:"phone"`
	PasswordHash string     `json:"-"`
	IsOwner      bool       `json:"is_owner"`
	IsActive     bool       `json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// UserClaims adalah payload yang disimpan di dalam JWT access token.
type UserClaims struct {
	ID          uuid.UUID  `json:"id"`
	OutletID    uuid.UUID  `json:"outlet_id"`
	OutletCode  string     `json:"outlet_code"`
	Name        string     `json:"name"`
	Phone       string     `json:"phone"`
	IsOwner     bool       `json:"is_owner"`
	Permissions []Permission `json:"permissions"`
}

// HasPermission memeriksa apakah user memiliki permission tertentu.
// Owner selalu dianggap memiliki semua permission — pengecekan ini
// hanya relevan untuk karyawan (non-owner).
func (u *UserClaims) HasPermission(p Permission) bool {
	for _, perm := range u.Permissions {
		if perm == p {
			return true
		}
	}
	return false
}
