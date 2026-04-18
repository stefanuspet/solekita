package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/stefanuspet/solekita/backend/internal/errors"
	"github.com/stefanuspet/solekita/backend/internal/model"
	"github.com/stefanuspet/solekita/backend/internal/repository"
)

type DeliveryService struct {
	deliveryRepo *repository.DeliveryRepository
	orderRepo    *repository.OrderRepository
	customerRepo *repository.CustomerRepository
	userRepo     *repository.UserRepository
	permRepo     *repository.UserPermissionRepository
	logRepo      *repository.OrderLogRepository
}

func NewDeliveryService(
	deliveryRepo *repository.DeliveryRepository,
	orderRepo *repository.OrderRepository,
	customerRepo *repository.CustomerRepository,
	userRepo *repository.UserRepository,
	permRepo *repository.UserPermissionRepository,
	logRepo *repository.OrderLogRepository,
) *DeliveryService {
	return &DeliveryService{
		deliveryRepo: deliveryRepo,
		orderRepo:    orderRepo,
		customerRepo: customerRepo,
		userRepo:     userRepo,
		permRepo:     permRepo,
		logRepo:      logRepo,
	}
}

// ── Response types ────────────────────────────────────────────────────────────

type DeliveryDetailResponse struct {
	ID              uuid.UUID             `json:"id"`
	OrderID         uuid.UUID             `json:"order_id"`
	OrderNumber     string                `json:"order_number"`
	CourierID       *uuid.UUID            `json:"courier_id"`
	PickupAddress   *string               `json:"pickup_address"`
	DeliveryAddress *string               `json:"delivery_address"`
	PickupStatus    model.PickupStatus    `json:"pickup_status"`
	DeliveryStatus  model.DeliveryStatus  `json:"delivery_status"`
	PickupNotes     *string               `json:"pickup_notes"`
	DeliveryNotes   *string               `json:"delivery_notes"`
	PickedUpAt      *time.Time            `json:"picked_up_at"`
	DeliveredAt     *time.Time            `json:"delivered_at"`
	Customer        *DeliveryCustomerInfo `json:"customer"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
}

type DeliveryCustomerInfo struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

// ── ListMyDeliveries ──────────────────────────────────────────────────────────

// ListMyDeliveries mengembalikan daftar delivery yang di-assign ke kurir yang sedang login.
func (s *DeliveryService) ListMyDeliveries(ctx context.Context, user *model.UserClaims, deliveryType, status string, date *time.Time) ([]*DeliveryDetailResponse, error) {
	f := repository.DeliveryFilters{
		Type: deliveryType,
		Date: date,
	}
	if status != "" {
		f.Status = &status
	}

	deliveries, err := s.deliveryRepo.GetByCourierID(ctx, user.ID, f)
	if err != nil {
		return nil, fmt.Errorf("ListMyDeliveries: %w", err)
	}

	res := make([]*DeliveryDetailResponse, 0, len(deliveries))
	for _, d := range deliveries {
		order, _ := s.orderRepo.GetByID(ctx, d.OrderID, user.OutletID)
		res = append(res, s.toResponse(ctx, d, order))
	}
	return res, nil
}

// ── AssignCourier ─────────────────────────────────────────────────────────────

// AssignCourier meng-assign atau re-assign kurir ke delivery.
// Validasi: courier harus punya permission manage_delivery dan milik outlet yang sama.
func (s *DeliveryService) AssignCourier(ctx context.Context, deliveryID, courierID, outletID uuid.UUID) (*DeliveryDetailResponse, error) {
	// Validasi courier ada dan milik outlet yang sama
	courier, err := s.userRepo.GetByIDInternal(ctx, courierID)
	if err != nil {
		return nil, apperrors.ErrNotFound.New(fmt.Errorf("kurir tidak ditemukan"))
	}
	if courier.OutletID != outletID {
		return nil, apperrors.ErrForbidden.New(fmt.Errorf("kurir tidak ditemukan di outlet ini"))
	}

	// Validasi courier punya permission manage_delivery (owner bypass)
	if !courier.IsOwner {
		perms, err := s.permRepo.GetByUserID(ctx, courierID)
		if err != nil {
			return nil, fmt.Errorf("AssignCourier: ambil permission kurir: %w", err)
		}
		hasDeliveryPerm := false
		for _, p := range perms {
			if p == model.PermissionManageDelivery {
				hasDeliveryPerm = true
				break
			}
		}
		if !hasDeliveryPerm {
			return nil, apperrors.ErrUnprocessable.New(fmt.Errorf("user tersebut tidak memiliki permission manage_delivery"))
		}
	}

	// Ambil delivery dan verifikasi milik outlet ini
	delivery, err := s.deliveryRepo.GetByID(ctx, deliveryID)
	if err != nil {
		return nil, err
	}
	order, err := s.orderRepo.GetByID(ctx, delivery.OrderID, outletID)
	if err != nil {
		return nil, apperrors.ErrNotFound.New(fmt.Errorf("delivery tidak ditemukan"))
	}

	delivery.CourierID = &courierID
	if err := s.deliveryRepo.Update(ctx, delivery); err != nil {
		return nil, fmt.Errorf("AssignCourier: %w", err)
	}

	return s.toResponse(ctx, delivery, order), nil
}

// ── UpdatePickupStatus ────────────────────────────────────────────────────────

// UpdatePickupStatus memperbarui status pickup oleh kurir.
// Transisi valid: pending → on_the_way; on_the_way → picked_up | failed.
// picked_up → set picked_up_at + update order status ke 'baru'.
// failed    → notes wajib diisi.
func (s *DeliveryService) UpdatePickupStatus(ctx context.Context, user *model.UserClaims, deliveryID uuid.UUID, newStatus model.PickupStatus, notes string) (*DeliveryDetailResponse, error) {
	delivery, err := s.deliveryRepo.GetByID(ctx, deliveryID)
	if err != nil {
		return nil, err
	}
	order, err := s.orderRepo.GetByID(ctx, delivery.OrderID, user.OutletID)
	if err != nil {
		return nil, apperrors.ErrNotFound.New(fmt.Errorf("delivery tidak ditemukan"))
	}

	if delivery.CourierID == nil || *delivery.CourierID != user.ID {
		return nil, apperrors.ErrForbidden.New(fmt.Errorf("Anda tidak di-assign ke delivery ini"))
	}

	if err := validatePickupTransition(delivery.PickupStatus, newStatus); err != nil {
		return nil, err
	}
	if newStatus == model.PickupStatusFailed && notes == "" {
		return nil, apperrors.ErrBadRequest.New(fmt.Errorf("notes wajib diisi jika status = failed"))
	}

	delivery.PickupStatus = newStatus
	if notes != "" {
		delivery.PickupNotes = &notes
	}
	if newStatus == model.PickupStatusPickedUp {
		now := time.Now()
		delivery.PickedUpAt = &now

		oldStatus := order.Status
		order.Status = model.OrderStatusBaru
		if err := s.orderRepo.Update(ctx, order); err != nil {
			return nil, fmt.Errorf("UpdatePickupStatus: update order: %w", err)
		}
		go func() {
			old := string(oldStatus)
			nw := string(order.Status)
			_ = s.logRepo.Create(context.Background(), &model.OrderLog{
				OrderID:  order.ID,
				UserID:   user.ID,
				Action:   model.OrderActionStatusChanged,
				OldValue: &old,
				NewValue: &nw,
			})
		}()
	}

	if err := s.deliveryRepo.Update(ctx, delivery); err != nil {
		return nil, fmt.Errorf("UpdatePickupStatus: %w", err)
	}
	return s.toResponse(ctx, delivery, order), nil
}

// ── UpdateDeliveryStatus ──────────────────────────────────────────────────────

// UpdateDeliveryStatus memperbarui status delivery oleh kurir.
// Transisi valid: pending → on_the_way; on_the_way → delivered | failed.
// delivered → set delivered_at + update order status ke 'diantar'.
// failed    → notes wajib diisi; order status tetap 'selesai'.
func (s *DeliveryService) UpdateDeliveryStatus(ctx context.Context, user *model.UserClaims, deliveryID uuid.UUID, newStatus model.DeliveryStatus, notes string) (*DeliveryDetailResponse, error) {
	delivery, err := s.deliveryRepo.GetByID(ctx, deliveryID)
	if err != nil {
		return nil, err
	}
	order, err := s.orderRepo.GetByID(ctx, delivery.OrderID, user.OutletID)
	if err != nil {
		return nil, apperrors.ErrNotFound.New(fmt.Errorf("delivery tidak ditemukan"))
	}

	if delivery.CourierID == nil || *delivery.CourierID != user.ID {
		return nil, apperrors.ErrForbidden.New(fmt.Errorf("Anda tidak di-assign ke delivery ini"))
	}

	if err := validateDeliveryTransition(delivery.DeliveryStatus, newStatus); err != nil {
		return nil, err
	}
	if newStatus == model.DeliveryStatusFailed && notes == "" {
		return nil, apperrors.ErrBadRequest.New(fmt.Errorf("notes wajib diisi jika status = failed"))
	}

	delivery.DeliveryStatus = newStatus
	if notes != "" {
		delivery.DeliveryNotes = &notes
	}
	if newStatus == model.DeliveryStatusDelivered {
		now := time.Now()
		delivery.DeliveredAt = &now

		oldStatus := order.Status
		order.Status = model.OrderStatusDiantar
		if err := s.orderRepo.Update(ctx, order); err != nil {
			return nil, fmt.Errorf("UpdateDeliveryStatus: update order: %w", err)
		}
		go func() {
			old := string(oldStatus)
			nw := string(order.Status)
			_ = s.logRepo.Create(context.Background(), &model.OrderLog{
				OrderID:  order.ID,
				UserID:   user.ID,
				Action:   model.OrderActionStatusChanged,
				OldValue: &old,
				NewValue: &nw,
			})
		}()
	}

	if err := s.deliveryRepo.Update(ctx, delivery); err != nil {
		return nil, fmt.Errorf("UpdateDeliveryStatus: %w", err)
	}
	return s.toResponse(ctx, delivery, order), nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (s *DeliveryService) toResponse(ctx context.Context, d *model.Delivery, order *model.Order) *DeliveryDetailResponse {
	res := &DeliveryDetailResponse{
		ID:              d.ID,
		OrderID:         d.OrderID,
		CourierID:       d.CourierID,
		PickupAddress:   d.PickupAddress,
		DeliveryAddress: d.DeliveryAddress,
		PickupStatus:    d.PickupStatus,
		DeliveryStatus:  d.DeliveryStatus,
		PickupNotes:     d.PickupNotes,
		DeliveryNotes:   d.DeliveryNotes,
		PickedUpAt:      d.PickedUpAt,
		DeliveredAt:     d.DeliveredAt,
		CreatedAt:       d.CreatedAt,
		UpdatedAt:       d.UpdatedAt,
	}
	if order != nil {
		res.OrderNumber = order.OrderNumber
		if customer, err := s.customerRepo.GetByID(ctx, order.CustomerID, order.OutletID); err == nil {
			res.Customer = &DeliveryCustomerInfo{
				Name:  customer.Name,
				Phone: customer.Phone,
			}
		}
	}
	return res
}

func validatePickupTransition(current, next model.PickupStatus) error {
	valid := map[model.PickupStatus]map[model.PickupStatus]bool{
		model.PickupStatusPending:  {model.PickupStatusOnTheWay: true},
		model.PickupStatusOnTheWay: {model.PickupStatusPickedUp: true, model.PickupStatusFailed: true},
	}
	if !valid[current][next] {
		return apperrors.ErrUnprocessable.New(fmt.Errorf(
			"transisi pickup_status dari '%s' ke '%s' tidak valid", current, next,
		))
	}
	return nil
}

func validateDeliveryTransition(current, next model.DeliveryStatus) error {
	valid := map[model.DeliveryStatus]map[model.DeliveryStatus]bool{
		model.DeliveryStatusPending:  {model.DeliveryStatusOnTheWay: true},
		model.DeliveryStatusOnTheWay: {model.DeliveryStatusDelivered: true, model.DeliveryStatusFailed: true},
	}
	if !valid[current][next] {
		return apperrors.ErrUnprocessable.New(fmt.Errorf(
			"transisi delivery_status dari '%s' ke '%s' tidak valid", current, next,
		))
	}
	return nil
}
