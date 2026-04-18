package errors

import "fmt"

// AppError adalah error type standar untuk semua business logic error.
type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// New membuat AppError baru dengan konteks error tambahan.
func (e *AppError) New(err error) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: e.Message,
		Err:     err,
	}
}

// Sentinel errors — gunakan langsung atau wrap dengan .New(err).
var (
	ErrNotFound      = &AppError{Code: 404, Message: "Data tidak ditemukan"}
	ErrUnauthorized  = &AppError{Code: 401, Message: "Tidak terautentikasi"}
	ErrForbidden     = &AppError{Code: 403, Message: "Tidak punya akses"}
	ErrConflict      = &AppError{Code: 409, Message: "Data sudah ada"}
	ErrUnprocessable = &AppError{Code: 422, Message: "Proses gagal"}
	ErrBadRequest    = &AppError{Code: 400, Message: "Request tidak valid"}
)

// ValidationError digunakan saat validasi field gagal.
// Fields berisi map field_name → pesan error.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %v", e.Fields)
}
