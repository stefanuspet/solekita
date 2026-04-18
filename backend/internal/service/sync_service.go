package service

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/stefanuspet/solekita/backend/internal/errors"
	"github.com/stefanuspet/solekita/backend/internal/model"
	"github.com/stefanuspet/solekita/backend/internal/repository"
	"github.com/stefanuspet/solekita/backend/internal/storage"
)

type SyncService struct {
	db            *sql.DB
	orderRepo     *repository.OrderRepository
	photoRepo     *repository.PhotoRepository
	paymentRepo   *repository.PaymentRepository
	deliveryRepo  *repository.DeliveryRepository
	logRepo       *repository.OrderLogRepository
	treatmentRepo *repository.TreatmentRepository
	customerRepo  *repository.CustomerRepository
	r2            *storage.R2Storage
}

func NewSyncService(
	db *sql.DB,
	orderRepo *repository.OrderRepository,
	photoRepo *repository.PhotoRepository,
	paymentRepo *repository.PaymentRepository,
	deliveryRepo *repository.DeliveryRepository,
	logRepo *repository.OrderLogRepository,
	treatmentRepo *repository.TreatmentRepository,
	customerRepo *repository.CustomerRepository,
	r2 *storage.R2Storage,
) *SyncService {
	return &SyncService{
		db:            db,
		orderRepo:     orderRepo,
		photoRepo:     photoRepo,
		paymentRepo:   paymentRepo,
		deliveryRepo:  deliveryRepo,
		logRepo:       logRepo,
		treatmentRepo: treatmentRepo,
		customerRepo:  customerRepo,
		r2:            r2,
	}
}

// ── Request / Response ────────────────────────────────────────────────────────

// SyncOrderItem mewakili satu order dari offline queue Flutter.
type SyncOrderItem struct {
	LocalID         string     `json:"local_id"`         // ID unik dari sisi Flutter
	CustomerPhone   string     `json:"customer_phone"`   // untuk find-or-create
	CustomerName    string     `json:"customer_name"`
	TreatmentID     uuid.UUID  `json:"treatment_id"`
	ConditionNotes  *string    `json:"condition_notes"`
	IsPickup        bool       `json:"is_pickup"`
	IsDelivery      bool       `json:"is_delivery"`
	PickupAddress   *string    `json:"pickup_address"`
	DeliveryAddress *string    `json:"delivery_address"`
	DeliveryFee     int        `json:"delivery_fee"`
	EstimatedDoneAt *time.Time `json:"estimated_done_at"`
	PaymentMethod   string     `json:"payment_method"`
	PaymentAmount   int        `json:"payment_amount"`
	CreatedAtLocal  time.Time  `json:"created_at_local"` // timestamp asli dari perangkat
}

// SyncedOrder adalah informasi order yang berhasil disinkronkan.
type SyncedOrder struct {
	LocalID     string    `json:"local_id"`
	ServerID    uuid.UUID `json:"server_id"`
	OrderNumber string    `json:"order_number"`
}

// FailedOrder adalah informasi order yang gagal disinkronkan beserta alasannya.
type FailedOrder struct {
	LocalID string `json:"local_id"`
	Reason  string `json:"reason"`
}

// SyncResult adalah hasil keseluruhan proses sinkronisasi.
type SyncResult struct {
	Synced []SyncedOrder `json:"synced"`
	Failed []FailedOrder `json:"failed"`
}

// SyncOrders memproses offline queue dari Flutter.
// Setiap order diproses dalam transaction terpisah sehingga gagalnya satu
// tidak membatalkan order lain dalam batch yang sama.
// photos adalah map dari local_id → bytes foto before (opsional per order).
func (s *SyncService) SyncOrders(ctx context.Context, user *model.UserClaims, orders []SyncOrderItem, photos map[string][]byte) (*SyncResult, error) {
	result := &SyncResult{
		Synced: make([]SyncedOrder, 0, len(orders)),
		Failed: make([]FailedOrder, 0),
	}

	for _, item := range orders {
		var photoBytes []byte
		if photos != nil {
			photoBytes = photos[item.LocalID]
		}

		serverID, orderNumber, err := s.syncOneOrder(ctx, user, item, photoBytes)
		if err != nil {
			result.Failed = append(result.Failed, FailedOrder{
				LocalID: item.LocalID,
				Reason:  err.Error(),
			})
			continue
		}
		result.Synced = append(result.Synced, SyncedOrder{
			LocalID:     item.LocalID,
			ServerID:    serverID,
			OrderNumber: orderNumber,
		})
	}

	return result, nil
}

// syncOneOrder memproses satu order dalam transaction tersendiri.
func (s *SyncService) syncOneOrder(ctx context.Context, user *model.UserClaims, item SyncOrderItem, photoBefore []byte) (uuid.UUID, string, error) {
	// 1. Validasi treatment ada dan aktif
	treatment, err := s.treatmentRepo.GetByID(ctx, item.TreatmentID, user.OutletID)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("treatment tidak ditemukan")
	}
	if !treatment.IsActive {
		return uuid.Nil, "", fmt.Errorf("treatment tidak aktif")
	}

	// 2. Find-or-create customer berdasarkan nomor HP
	customer, err := s.findOrCreateCustomer(ctx, user, item.CustomerPhone, item.CustomerName)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("customer: %w", err)
	}

	// 3. Upload foto before ke R2 sebelum membuka transaction
	// (upload di luar tx agar tidak memperpanjang durasi lock DB)
	var photoKey string
	if len(photoBefore) > 0 && s.r2 != nil {
		tempID := uuid.New()
		photoKey, err = s.r2.UploadPhoto(ctx, bytes.NewReader(photoBefore), tempID, user.OutletID, model.PhotoTypeBefore)
		if err != nil {
			// non-fatal: order tetap dibuat tanpa foto
			photoKey = ""
		}
	}

	// 4. Begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		if photoKey != "" {
			_ = s.r2.Delete(ctx, photoKey)
		}
		return uuid.Nil, "", fmt.Errorf("begin tx: %w", err)
	}

	var txErr error
	defer func() {
		if txErr != nil {
			tx.Rollback()
			if photoKey != "" {
				_ = s.r2.Delete(ctx, photoKey)
			}
		}
	}()

	// 5. Generate order number menggunakan tanggal lokal perangkat
	orderNumber, txErr := s.orderRepo.GenerateOrderNumberForDate(ctx, tx, user.OutletID, user.OutletCode, item.CreatedAtLocal)
	if txErr != nil {
		return uuid.Nil, "", fmt.Errorf("generate order number: %w", txErr)
	}

	// 6. Tentukan status awal
	status := model.OrderStatusBaru
	if item.IsPickup {
		status = model.OrderStatusDijemput
	}

	material := ""
	if treatment.Material != nil {
		material = *treatment.Material
	}

	order := &model.Order{
		OrderNumber:     orderNumber,
		OutletID:        user.OutletID,
		CustomerID:      customer.ID,
		KasirID:         user.ID,
		TreatmentID:     treatment.ID,
		TreatmentName:   treatment.Name,
		Material:        material,
		Status:          status,
		BasePrice:       treatment.Price,
		DeliveryFee:     item.DeliveryFee,
		TotalPrice:      treatment.Price + item.DeliveryFee,
		ConditionNotes:  item.ConditionNotes,
		IsPickup:        item.IsPickup,
		IsDelivery:      item.IsDelivery,
		EstimatedDoneAt: item.EstimatedDoneAt,
	}

	// 7. Simpan order dengan created_at dari perangkat
	if txErr = s.orderRepo.CreateWithTimestamp(ctx, tx, order, item.CreatedAtLocal); txErr != nil {
		return uuid.Nil, "", fmt.Errorf("simpan order: %w", txErr)
	}

	// 8. Simpan foto before jika berhasil diupload
	if photoKey != "" {
		photo := &model.Photo{
			OrderID: order.ID,
			Type:    model.PhotoTypeBefore,
			R2Key:   photoKey,
		}
		if txErr = s.photoRepo.Create(ctx, tx, photo); txErr != nil {
			return uuid.Nil, "", fmt.Errorf("simpan foto: %w", txErr)
		}
	}

	// 9. Simpan payment — gunakan created_at_local sebagai waktu pembayaran
	paidAt := item.CreatedAtLocal
	payment := &model.Payment{
		OrderID: order.ID,
		Amount:  item.PaymentAmount,
		Method:  model.PaymentMethod(item.PaymentMethod),
		Status:  model.PaymentStatusPaid,
		PaidAt:  &paidAt,
	}
	if txErr = s.paymentRepo.Create(ctx, tx, payment); txErr != nil {
		return uuid.Nil, "", fmt.Errorf("simpan payment: %w", txErr)
	}

	// 10. Buat record delivery jika is_pickup
	if item.IsPickup {
		delivery := &model.Delivery{
			OrderID:         order.ID,
			PickupAddress:   item.PickupAddress,
			DeliveryAddress: item.DeliveryAddress,
			PickupStatus:    model.PickupStatusPending,
			DeliveryStatus:  model.DeliveryStatusPending,
		}
		if txErr = s.deliveryRepo.Create(ctx, tx, delivery); txErr != nil {
			return uuid.Nil, "", fmt.Errorf("buat delivery: %w", txErr)
		}
	}

	// 11. Commit
	if txErr = tx.Commit(); txErr != nil {
		return uuid.Nil, "", fmt.Errorf("commit: %w", txErr)
	}

	// 12. Catat log (best-effort, setelah commit)
	logNote := fmt.Sprintf("order %s dibuat via sync offline", orderNumber)
	_ = s.logRepo.Create(ctx, &model.OrderLog{
		OrderID:  order.ID,
		UserID:   user.ID,
		Action:   model.OrderActionCreated,
		NewValue: &logNote,
	})

	return order.ID, orderNumber, nil
}

// findOrCreateCustomer mencari customer berdasarkan nomor HP atau membuat baru
// jika belum ada dalam outlet yang sama.
func (s *SyncService) findOrCreateCustomer(ctx context.Context, user *model.UserClaims, phone, name string) (*model.Customer, error) {
	c, err := s.customerRepo.GetByPhone(ctx, user.OutletID, phone)
	if err == nil {
		return c, nil
	}
	if !errors.Is(err, apperrors.ErrNotFound) {
		return nil, fmt.Errorf("cari customer: %w", err)
	}

	c = &model.Customer{
		OutletID: user.OutletID,
		Name:     name,
		Phone:    phone,
	}
	if err := s.customerRepo.Create(ctx, c); err != nil {
		return nil, fmt.Errorf("buat customer: %w", err)
	}
	return c, nil
}
