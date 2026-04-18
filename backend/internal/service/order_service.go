package service

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/stefanuspet/solekita/backend/internal/errors"
	"github.com/stefanuspet/solekita/backend/internal/model"
	"github.com/stefanuspet/solekita/backend/internal/repository"
	"github.com/stefanuspet/solekita/backend/internal/storage"
)

type OrderService struct {
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

func NewOrderService(
	db *sql.DB,
	orderRepo *repository.OrderRepository,
	photoRepo *repository.PhotoRepository,
	paymentRepo *repository.PaymentRepository,
	deliveryRepo *repository.DeliveryRepository,
	logRepo *repository.OrderLogRepository,
	treatmentRepo *repository.TreatmentRepository,
	customerRepo *repository.CustomerRepository,
	r2 *storage.R2Storage,
) *OrderService {
	return &OrderService{
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

type ListOrdersFilters struct {
	Status      *model.OrderStatus
	DateFrom    *time.Time
	DateTo      *time.Time
	KasirID     *uuid.UUID
	TreatmentID *uuid.UUID
	Search      string
}

type ListOrdersResponse struct {
	Orders []*OrderResponse `json:"orders"`
	Total  int              `json:"total"`
	Page   int              `json:"page"`
	Limit  int              `json:"limit"`
}

type CreateOrderRequest struct {
	CustomerID      uuid.UUID     `form:"customer_id" binding:"required"`
	TreatmentID     uuid.UUID     `form:"treatment_id" binding:"required"`
	ConditionNotes  *string       `form:"condition_notes"`
	IsPickup        bool          `form:"is_pickup"`
	IsDelivery      bool          `form:"is_delivery"`
	PickupAddress   *string       `form:"pickup_address"`
	DeliveryAddress *string       `form:"delivery_address"`
	DeliveryFee     int           `form:"delivery_fee"`
	EstimatedDoneAt *time.Time    `form:"estimated_done_at" time_format:"2006-01-02T15:04:05Z07:00"`
	PaymentMethod   string        `form:"payment_method" binding:"required"`
	PaymentAmount   int           `form:"payment_amount" binding:"required,min=1"`
}

type OrderCustomerData struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Phone string    `json:"phone"`
}

type OrderPhotosData struct {
	BeforeURL *string `json:"before_url"`
	AfterURL  *string `json:"after_url"`
}

type OrderResponse struct {
	ID              uuid.UUID           `json:"id"`
	OrderNumber     string              `json:"order_number"`
	Customer        *OrderCustomerData  `json:"customer"`
	TreatmentName   string              `json:"treatment_name"`
	Material        string              `json:"material"`
	Status          model.OrderStatus   `json:"status"`
	BasePrice       int                 `json:"base_price"`
	DeliveryFee     int                 `json:"delivery_fee"`
	TotalPrice      int                 `json:"total_price"`
	IsPriceEdited   bool                `json:"is_price_edited"`
	OriginalPrice   *int                `json:"original_price"`
	ConditionNotes  *string             `json:"condition_notes"`
	IsPickup        bool                `json:"is_pickup"`
	IsDelivery      bool                `json:"is_delivery"`
	EstimatedDoneAt *time.Time          `json:"estimated_done_at"`
	Photos          OrderPhotosData     `json:"photos"`
	Payment         *PaymentResponse    `json:"payment"`
	CreatedAt       time.Time           `json:"created_at"`
}

type PaymentResponse struct {
	ID     uuid.UUID           `json:"id"`
	Amount int                 `json:"amount"`
	Method model.PaymentMethod `json:"method"`
	Status model.PaymentStatus `json:"status"`
	PaidAt *time.Time          `json:"paid_at"`
}

type DeliveryResponse struct {
	ID              uuid.UUID              `json:"id"`
	CourierID       *uuid.UUID             `json:"courier_id"`
	PickupAddress   *string                `json:"pickup_address"`
	DeliveryAddress *string                `json:"delivery_address"`
	PickupStatus    model.PickupStatus     `json:"pickup_status"`
	DeliveryStatus  model.DeliveryStatus   `json:"delivery_status"`
	PickupNotes     *string                `json:"pickup_notes"`
	DeliveryNotes   *string                `json:"delivery_notes"`
	PickedUpAt      *time.Time             `json:"picked_up_at"`
	DeliveredAt     *time.Time             `json:"delivered_at"`
}

type OrderLogResponse struct {
	ID        uuid.UUID           `json:"id"`
	UserID    uuid.UUID           `json:"user_id"`
	Action    model.OrderAction   `json:"action"`
	OldValue  *string             `json:"old_value"`
	NewValue  *string             `json:"new_value"`
	Notes     *string             `json:"notes"`
	CreatedAt time.Time           `json:"created_at"`
}

type OrderDetailResponse struct {
	OrderResponse
	Delivery *DeliveryResponse   `json:"delivery"`
	Logs     []*OrderLogResponse `json:"logs"`
}

// ── GetOrder ──────────────────────────────────────────────────────────────────

func (s *OrderService) GetOrder(ctx context.Context, orderID, outletID uuid.UUID) (*OrderDetailResponse, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID, outletID)
	if err != nil {
		return nil, fmt.Errorf("GetOrder: %w", err)
	}

	customer, _ := s.customerRepo.GetByID(ctx, order.CustomerID, outletID)
	var customerData *OrderCustomerData
	if customer != nil {
		customerData = &OrderCustomerData{
			ID:    customer.ID,
			Name:  customer.Name,
			Phone: customer.Phone,
		}
	}

	photos, _ := s.photoRepo.GetByOrderID(ctx, order.ID)
	photosData := buildPhotosData(ctx, s.r2, photos)

	payment, _ := s.paymentRepo.GetByOrderID(ctx, order.ID)

	var deliveryRes *DeliveryResponse
	if delivery, err := s.deliveryRepo.GetByOrderID(ctx, order.ID, outletID); err == nil {
		deliveryRes = &DeliveryResponse{
			ID:              delivery.ID,
			CourierID:       delivery.CourierID,
			PickupAddress:   delivery.PickupAddress,
			DeliveryAddress: delivery.DeliveryAddress,
			PickupStatus:    delivery.PickupStatus,
			DeliveryStatus:  delivery.DeliveryStatus,
			PickupNotes:     delivery.PickupNotes,
			DeliveryNotes:   delivery.DeliveryNotes,
			PickedUpAt:      delivery.PickedUpAt,
			DeliveredAt:     delivery.DeliveredAt,
		}
	}

	rawLogs, _ := s.logRepo.GetByOrderID(ctx, order.ID)
	logs := make([]*OrderLogResponse, 0, len(rawLogs))
	for _, l := range rawLogs {
		logs = append(logs, &OrderLogResponse{
			ID:        l.ID,
			UserID:    l.UserID,
			Action:    l.Action,
			OldValue:  l.OldValue,
			NewValue:  l.NewValue,
			Notes:     l.Notes,
			CreatedAt: l.CreatedAt,
		})
	}

	return &OrderDetailResponse{
		OrderResponse: *buildOrderResponse(order, payment, customerData, photosData),
		Delivery:      deliveryRes,
		Logs:          logs,
	}, nil
}

// ── ListOrders ────────────────────────────────────────────────────────────────

func (s *OrderService) ListOrders(ctx context.Context, user *model.UserClaims, f ListOrdersFilters, page, limit int) (*ListOrdersResponse, error) {
	repoFilters := repository.OrderFilters{
		Status:      f.Status,
		KasirID:     f.KasirID,
		TreatmentID: f.TreatmentID,
		DateFrom:    f.DateFrom,
		DateTo:      f.DateTo,
		Search:      f.Search,
	}

	orders, total, err := s.orderRepo.ListByOutletID(ctx, user.OutletID, repoFilters, page, limit)
	if err != nil {
		return nil, fmt.Errorf("ListOrders: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	res := make([]*OrderResponse, 0, len(orders))
	for _, o := range orders {
		customer, _ := s.customerRepo.GetByID(ctx, o.CustomerID, user.OutletID)
		var customerData *OrderCustomerData
		if customer != nil {
			customerData = &OrderCustomerData{
				ID:    customer.ID,
				Name:  customer.Name,
				Phone: customer.Phone,
			}
		}

		photos, _ := s.photoRepo.GetByOrderID(ctx, o.ID)
		photosData := buildPhotosData(ctx, s.r2, photos)

		payment, _ := s.paymentRepo.GetByOrderID(ctx, o.ID)
		res = append(res, buildOrderResponse(o, payment, customerData, photosData))
	}

	return &ListOrdersResponse{
		Orders: res,
		Total:  total,
		Page:   page,
		Limit:  limit,
	}, nil
}

// ── CreateOrder ───────────────────────────────────────────────────────────────

func (s *OrderService) CreateOrder(ctx context.Context, user *model.UserClaims, req CreateOrderRequest, photoBefore io.Reader) (*OrderResponse, error) {
	// 1. Validasi treatment ada dan aktif
	treatment, err := s.treatmentRepo.GetByID(ctx, req.TreatmentID, user.OutletID)
	if err != nil {
		return nil, fmt.Errorf("CreateOrder: treatment tidak ditemukan: %w", err)
	}
	if !treatment.IsActive {
		return nil, apperrors.ErrUnprocessable.New(fmt.Errorf("treatment tidak aktif"))
	}

	// 2. Validasi foto before wajib ada
	if photoBefore == nil {
		return nil, apperrors.ErrBadRequest.New(fmt.Errorf("foto before wajib diupload"))
	}

	// 3. Upload foto before ke R2 sebelum transaction
	// (upload di luar tx agar tidak memperpanjang durasi lock DB)
	tempOrderID := uuid.New() // placeholder, diganti setelah order tersimpan
	photoKey, err := s.r2.UploadPhoto(ctx, photoBefore, tempOrderID, user.OutletID, model.PhotoTypeBefore)
	if err != nil {
		return nil, fmt.Errorf("CreateOrder: upload foto before: %w", err)
	}

	// 4. Begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("CreateOrder: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			// Best-effort: hapus foto yang sudah terupload jika tx gagal
			_ = s.r2.Delete(ctx, photoKey)
		}
	}()

	// 5. Generate order number (atomic dalam tx)
	orderNumber, err := s.orderRepo.GenerateOrderNumber(ctx, tx, user.OutletID, user.OutletCode)
	if err != nil {
		return nil, fmt.Errorf("CreateOrder: generate order number: %w", err)
	}

	// 6. Tentukan status awal
	status := model.OrderStatusBaru
	if req.IsPickup {
		status = model.OrderStatusDijemput
	}

	totalPrice := treatment.Price + req.DeliveryFee

	material := ""
	if treatment.Material != nil {
		material = *treatment.Material
	}

	order := &model.Order{
		OrderNumber:     orderNumber,
		OutletID:        user.OutletID,
		CustomerID:      req.CustomerID,
		KasirID:         user.ID,
		TreatmentID:     treatment.ID,
		TreatmentName:   treatment.Name,
		Material:        material,
		Status:          status,
		BasePrice:       treatment.Price,
		DeliveryFee:     req.DeliveryFee,
		TotalPrice:      totalPrice,
		ConditionNotes:  req.ConditionNotes,
		IsPickup:        req.IsPickup,
		IsDelivery:      req.IsDelivery,
		EstimatedDoneAt: req.EstimatedDoneAt,
	}

	// 7. Simpan order
	if err = s.orderRepo.Create(ctx, tx, order); err != nil {
		return nil, fmt.Errorf("CreateOrder: simpan order: %w", err)
	}

	// 8. Simpan foto before (pakai order.ID yang sudah dapat dari DB)
	photo := &model.Photo{
		OrderID: order.ID,
		Type:    model.PhotoTypeBefore,
		R2Key:   photoKey,
	}
	if err = s.photoRepo.Create(ctx, tx, photo); err != nil {
		return nil, fmt.Errorf("CreateOrder: simpan foto before: %w", err)
	}

	// 9. Simpan payment
	now := time.Now()
	payment := &model.Payment{
		OrderID: order.ID,
		Amount:  req.PaymentAmount,
		Method:  model.PaymentMethod(req.PaymentMethod),
		Status:  model.PaymentStatusPaid,
		PaidAt:  &now,
	}
	if err = s.paymentRepo.Create(ctx, tx, payment); err != nil {
		return nil, fmt.Errorf("CreateOrder: simpan payment: %w", err)
	}

	// 10. Jika is_pickup → buat record delivery
	var delivery *model.Delivery
	if req.IsPickup {
		delivery = &model.Delivery{
			OrderID:        order.ID,
			PickupAddress:  req.PickupAddress,
			DeliveryAddress: req.DeliveryAddress,
			PickupStatus:   model.PickupStatusPending,
			DeliveryStatus: model.DeliveryStatusPending,
		}
		if err = s.deliveryRepo.Create(ctx, tx, delivery); err != nil {
			return nil, fmt.Errorf("CreateOrder: buat delivery: %w", err)
		}
	}

	// 11. Commit
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("CreateOrder: commit: %w", err)
	}

	// 12. Catat order log (setelah commit, best-effort)
	logNote := fmt.Sprintf("order %s dibuat", orderNumber)
	_ = s.logRepo.Create(ctx, &model.OrderLog{
		OrderID:  order.ID,
		UserID:   user.ID,
		Action:   model.OrderActionCreated,
		NewValue: &logNote,
	})

	// 13. Fetch customer untuk response
	customer, _ := s.customerRepo.GetByID(ctx, order.CustomerID, user.OutletID)

	// 14. Generate signed URL foto before
	beforeURL, urlErr := s.r2.GetSignedURL(ctx, photoKey, time.Hour)
	if urlErr != nil {
		beforeURL = ""
	}
	err = nil // reset agar defer tidak rollback (GetSignedURL bukan error fatal)

	var customerData *OrderCustomerData
	if customer != nil {
		customerData = &OrderCustomerData{
			ID:    customer.ID,
			Name:  customer.Name,
			Phone: customer.Phone,
		}
	}

	res := buildOrderResponse(order, payment, customerData, OrderPhotosData{
		BeforeURL: &beforeURL,
		AfterURL:  nil,
	})
	return res, nil
}

// ── GetOrderPhotos ────────────────────────────────────────────────────────────

func (s *OrderService) GetOrderPhotos(ctx context.Context, orderID, outletID uuid.UUID) (*OrderPhotosData, error) {
	// Validasi order ada dan milik outlet
	if _, err := s.orderRepo.GetByID(ctx, orderID, outletID); err != nil {
		return nil, fmt.Errorf("GetOrderPhotos: %w", err)
	}

	photos, err := s.photoRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("GetOrderPhotos: ambil foto: %w", err)
	}

	data := buildPhotosData(ctx, s.r2, photos)
	return &data, nil
}

// ── UploadAfterPhoto ──────────────────────────────────────────────────────────

type PhotoURLResponse struct {
	ID         uuid.UUID       `json:"id"`
	Type       model.PhotoType `json:"type"`
	SignedURL  string          `json:"signed_url"`
	UploadedAt time.Time       `json:"uploaded_at"`
}

func (s *OrderService) UploadAfterPhoto(ctx context.Context, user *model.UserClaims, orderID, outletID uuid.UUID, file io.Reader) (*PhotoURLResponse, error) {
	if s.r2 == nil {
		return nil, apperrors.ErrUnprocessable.New(fmt.Errorf("storage belum dikonfigurasi"))
	}

	// 1. Validasi order ada dan milik outlet
	order, err := s.orderRepo.GetByID(ctx, orderID, outletID)
	if err != nil {
		return nil, fmt.Errorf("UploadAfterPhoto: ambil order: %w", err)
	}

	if order.Status == model.OrderStatusDibatalkan {
		return nil, apperrors.ErrUnprocessable.New(fmt.Errorf("tidak bisa upload foto ke order yang dibatalkan"))
	}

	// 2. Validasi foto after belum ada
	existing, err := s.photoRepo.GetByOrderIDAndType(ctx, orderID, model.PhotoTypeAfter)
	if err != nil {
		return nil, fmt.Errorf("UploadAfterPhoto: cek foto after: %w", err)
	}
	if existing != nil {
		return nil, apperrors.ErrConflict.New(fmt.Errorf("foto after sudah ada untuk order ini"))
	}

	// 3. Upload ke R2
	photoKey, err := s.r2.UploadPhoto(ctx, file, orderID, outletID, model.PhotoTypeAfter)
	if err != nil {
		return nil, fmt.Errorf("UploadAfterPhoto: upload: %w", err)
	}

	// 4. Simpan ke DB
	photo := &model.Photo{
		OrderID: order.ID,
		Type:    model.PhotoTypeAfter,
		R2Key:   photoKey,
	}
	if err := s.photoRepo.CreateDirect(ctx, photo); err != nil {
		// Best-effort hapus file yang sudah terupload
		_ = s.r2.Delete(ctx, photoKey)
		return nil, fmt.Errorf("UploadAfterPhoto: simpan foto: %w", err)
	}

	// 5. Catat log (best-effort)
	action := "foto after diupload"
	_ = s.logRepo.Create(ctx, &model.OrderLog{
		OrderID:  order.ID,
		UserID:   user.ID,
		Action:   model.OrderActionPhotoAdded,
		NewValue: &action,
	})

	signedURL, _ := s.r2.GetSignedURL(ctx, photoKey, time.Hour)

	return &PhotoURLResponse{
		ID:         photo.ID,
		Type:       photo.Type,
		SignedURL:  signedURL,
		UploadedAt: photo.UploadedAt,
	}, nil
}

// ── EditOrderPrice ────────────────────────────────────────────────────────────

type EditOrderPriceRequest struct {
	NewPrice int     `json:"new_price" binding:"required,min=1"`
	Notes    *string `json:"notes"`
}

func (s *OrderService) EditOrderPrice(ctx context.Context, user *model.UserClaims, orderID, outletID uuid.UUID, req EditOrderPriceRequest) (*OrderResponse, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID, outletID)
	if err != nil {
		return nil, fmt.Errorf("EditOrderPrice: ambil order: %w", err)
	}

	if order.Status == model.OrderStatusDibatalkan {
		return nil, apperrors.ErrUnprocessable.New(fmt.Errorf("harga order yang dibatalkan tidak bisa diubah"))
	}

	// Simpan original_price hanya pada edit pertama kali
	if !order.IsPriceEdited {
		order.OriginalPrice = &order.TotalPrice
	}

	oldPrice := order.TotalPrice
	order.TotalPrice = req.NewPrice
	order.IsPriceEdited = true

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("EditOrderPrice: simpan: %w", err)
	}

	// Catat log (best-effort)
	oldVal := fmt.Sprintf("%d", oldPrice)
	newVal := fmt.Sprintf("%d", req.NewPrice)
	_ = s.logRepo.Create(ctx, &model.OrderLog{
		OrderID:  order.ID,
		UserID:   user.ID,
		Action:   model.OrderActionPriceEdited,
		OldValue: &oldVal,
		NewValue: &newVal,
		Notes:    req.Notes,
	})

	customer, _ := s.customerRepo.GetByID(ctx, order.CustomerID, outletID)
	var customerData *OrderCustomerData
	if customer != nil {
		customerData = &OrderCustomerData{ID: customer.ID, Name: customer.Name, Phone: customer.Phone}
	}
	photos, _ := s.photoRepo.GetByOrderID(ctx, order.ID)
	payment, _ := s.paymentRepo.GetByOrderID(ctx, order.ID)

	return buildOrderResponse(order, payment, customerData, buildPhotosData(ctx, s.r2, photos)), nil
}

// ── CancelOrder ───────────────────────────────────────────────────────────────

var cancellableStatuses = map[model.OrderStatus]bool{
	model.OrderStatusDijemput: true,
	model.OrderStatusBaru:     true,
	model.OrderStatusProses:   true,
}

func (s *OrderService) CancelOrder(ctx context.Context, user *model.UserClaims, orderID, outletID uuid.UUID, reason string) (*OrderResponse, error) {
	if reason == "" {
		return nil, apperrors.ErrBadRequest.New(fmt.Errorf("alasan pembatalan wajib diisi"))
	}

	order, err := s.orderRepo.GetByID(ctx, orderID, outletID)
	if err != nil {
		return nil, fmt.Errorf("CancelOrder: ambil order: %w", err)
	}

	// Validasi status bisa dibatalkan
	if !cancellableStatuses[order.Status] {
		return nil, apperrors.ErrUnprocessable.New(
			fmt.Errorf("order berstatus %q tidak bisa dibatalkan", order.Status),
		)
	}

	// Validasi permission: kasir hanya bisa batalkan ordernya sendiri
	if !user.IsOwner && !user.HasPermission(model.PermissionManageOutlet) {
		if order.KasirID != user.ID {
			return nil, apperrors.ErrForbidden.New(fmt.Errorf("kasir hanya bisa membatalkan order milik sendiri"))
		}
	}

	now := time.Now()
	order.Status = model.OrderStatusDibatalkan
	order.CancelReason = &reason
	order.CancelledBy = &user.ID
	order.CancelledAt = &now

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("CancelOrder: simpan: %w", err)
	}

	// Catat log (best-effort)
	_ = s.logRepo.Create(ctx, &model.OrderLog{
		OrderID:  order.ID,
		UserID:   user.ID,
		Action:   model.OrderActionCancelled,
		NewValue: &reason,
	})

	customer, _ := s.customerRepo.GetByID(ctx, order.CustomerID, outletID)
	var customerData *OrderCustomerData
	if customer != nil {
		customerData = &OrderCustomerData{ID: customer.ID, Name: customer.Name, Phone: customer.Phone}
	}
	photos, _ := s.photoRepo.GetByOrderID(ctx, order.ID)
	payment, _ := s.paymentRepo.GetByOrderID(ctx, order.ID)

	return buildOrderResponse(order, payment, customerData, buildPhotosData(ctx, s.r2, photos)), nil
}

// ── UpdateOrderStatus ─────────────────────────────────────────────────────────

// validTransitions mendefinisikan transisi status yang diizinkan.
var validTransitions = map[model.OrderStatus][]model.OrderStatus{
	model.OrderStatusDijemput: {model.OrderStatusBaru},
	model.OrderStatusBaru:     {model.OrderStatusProses},
	model.OrderStatusProses:   {model.OrderStatusSelesai},
	model.OrderStatusSelesai:  {model.OrderStatusDiambil, model.OrderStatusDiantar},
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, user *model.UserClaims, orderID, outletID uuid.UUID, newStatus model.OrderStatus) (*OrderResponse, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID, outletID)
	if err != nil {
		return nil, fmt.Errorf("UpdateOrderStatus: ambil order: %w", err)
	}

	// Validasi transisi
	allowed, ok := validTransitions[order.Status]
	if !ok {
		return nil, apperrors.ErrUnprocessable.New(fmt.Errorf("status %q tidak bisa diubah", order.Status))
	}
	valid := false
	for _, s := range allowed {
		if s == newStatus {
			valid = true
			break
		}
	}
	if !valid {
		return nil, apperrors.ErrUnprocessable.New(
			fmt.Errorf("transisi dari %q ke %q tidak diizinkan", order.Status, newStatus),
		)
	}

	// Validasi foto after wajib ada jika update ke selesai
	if newStatus == model.OrderStatusSelesai {
		afterPhoto, err := s.photoRepo.GetByOrderIDAndType(ctx, order.ID, model.PhotoTypeAfter)
		if err != nil {
			return nil, fmt.Errorf("UpdateOrderStatus: cek foto after: %w", err)
		}
		if afterPhoto == nil {
			return nil, apperrors.ErrUnprocessable.New(fmt.Errorf("foto after wajib ada sebelum order ditandai selesai"))
		}
	}

	oldStatus := order.Status
	order.Status = newStatus

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("UpdateOrderStatus: simpan: %w", err)
	}

	// Catat log (best-effort)
	oldVal := string(oldStatus)
	newVal := string(newStatus)
	_ = s.logRepo.Create(ctx, &model.OrderLog{
		OrderID:  order.ID,
		UserID:   user.ID,
		Action:   model.OrderActionStatusChanged,
		OldValue: &oldVal,
		NewValue: &newVal,
	})

	customer, _ := s.customerRepo.GetByID(ctx, order.CustomerID, outletID)
	var customerData *OrderCustomerData
	if customer != nil {
		customerData = &OrderCustomerData{ID: customer.ID, Name: customer.Name, Phone: customer.Phone}
	}
	photos, _ := s.photoRepo.GetByOrderID(ctx, order.ID)
	payment, _ := s.paymentRepo.GetByOrderID(ctx, order.ID)

	return buildOrderResponse(order, payment, customerData, buildPhotosData(ctx, s.r2, photos)), nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func buildPhotosData(ctx context.Context, r2 *storage.R2Storage, photos []*model.Photo) OrderPhotosData {
	data := OrderPhotosData{}
	if r2 == nil {
		return data
	}
	for _, p := range photos {
		url, err := r2.GetSignedURL(ctx, p.R2Key, time.Hour)
		if err != nil {
			continue
		}
		urlCopy := url
		switch p.Type {
		case model.PhotoTypeBefore:
			data.BeforeURL = &urlCopy
		case model.PhotoTypeAfter:
			data.AfterURL = &urlCopy
		}
	}
	return data
}

func buildOrderResponse(o *model.Order, p *model.Payment, customer *OrderCustomerData, photos OrderPhotosData) *OrderResponse {
	var payRes *PaymentResponse
	if p != nil {
		payRes = &PaymentResponse{
			ID:     p.ID,
			Amount: p.Amount,
			Method: p.Method,
			Status: p.Status,
			PaidAt: p.PaidAt,
		}
	}
	return &OrderResponse{
		ID:              o.ID,
		OrderNumber:     o.OrderNumber,
		Customer:        customer,
		TreatmentName:   o.TreatmentName,
		Material:        o.Material,
		Status:          o.Status,
		BasePrice:       o.BasePrice,
		DeliveryFee:     o.DeliveryFee,
		TotalPrice:      o.TotalPrice,
		IsPriceEdited:   o.IsPriceEdited,
		OriginalPrice:   o.OriginalPrice,
		ConditionNotes:  o.ConditionNotes,
		IsPickup:        o.IsPickup,
		IsDelivery:      o.IsDelivery,
		EstimatedDoneAt: o.EstimatedDoneAt,
		Photos:          photos,
		Payment:         payRes,
		CreatedAt:       o.CreatedAt,
	}
}
