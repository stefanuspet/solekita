package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	apperrors "github.com/stefanuspet/solekita/backend/internal/errors"
	"github.com/stefanuspet/solekita/backend/internal/model"
	"github.com/stefanuspet/solekita/backend/internal/repository"
)

type CustomerService struct {
	customerRepo *repository.CustomerRepository
}

func NewCustomerService(customerRepo *repository.CustomerRepository) *CustomerService {
	return &CustomerService{customerRepo: customerRepo}
}

// ── Request / Response ────────────────────────────────────────────────────────

type FindOrCreateCustomerRequest struct {
	Name  string  `json:"name" binding:"required"`
	Phone string  `json:"phone" binding:"required"`
	Notes *string `json:"notes"`
}

type UpdateCustomerRequest struct {
	Name            string  `json:"name" binding:"required"`
	Phone           string  `json:"phone" binding:"required"`
	Notes           *string `json:"notes"`
	IsBlacklisted   *bool   `json:"is_blacklisted"`
	BlacklistReason *string `json:"blacklist_reason"`
}

type CustomerResponse struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Phone           string    `json:"phone"`
	TotalOrders     int       `json:"total_orders"`
	Notes           *string   `json:"notes"`
	IsBlacklisted   bool      `json:"is_blacklisted"`
	BlacklistReason *string   `json:"blacklist_reason"`
}

type FindOrCreateResponse struct {
	Customer CustomerResponse `json:"customer"`
	IsNew    bool             `json:"is_new"`
}

type ListCustomersResponse struct {
	Customers []*CustomerResponse `json:"customers"`
	Total     int                 `json:"total"`
	Page      int                 `json:"page"`
	Limit     int                 `json:"limit"`
}

// ── FindOrCreate ──────────────────────────────────────────────────────────────

func (s *CustomerService) FindOrCreate(ctx context.Context, outletID uuid.UUID, req FindOrCreateCustomerRequest) (*FindOrCreateResponse, error) {
	existing, err := s.customerRepo.GetByPhone(ctx, outletID, req.Phone)
	if err == nil {
		return &FindOrCreateResponse{
			Customer: *customerToResponse(existing),
			IsNew:    false,
		}, nil
	}
	if !errors.Is(err, apperrors.ErrNotFound) {
		return nil, fmt.Errorf("FindOrCreate: cek pelanggan: %w", err)
	}

	c := &model.Customer{
		OutletID: outletID,
		Name:     req.Name,
		Phone:    req.Phone,
		Notes:    req.Notes,
	}
	if err := s.customerRepo.Create(ctx, c); err != nil {
		return nil, fmt.Errorf("FindOrCreate: buat pelanggan: %w", err)
	}

	return &FindOrCreateResponse{
		Customer: *customerToResponse(c),
		IsNew:    true,
	}, nil
}

// ── ListCustomers ─────────────────────────────────────────────────────────────

func (s *CustomerService) ListCustomers(ctx context.Context, outletID uuid.UUID, search string, page, limit int) (*ListCustomersResponse, error) {
	customers, total, err := s.customerRepo.ListByOutletID(ctx, outletID, search, page, limit)
	if err != nil {
		return nil, fmt.Errorf("ListCustomers: %w", err)
	}

	if limit < 1 {
		limit = 20
	}
	if page < 1 {
		page = 1
	}

	res := make([]*CustomerResponse, len(customers))
	for i, c := range customers {
		res[i] = customerToResponse(c)
	}

	return &ListCustomersResponse{
		Customers: res,
		Total:     total,
		Page:      page,
		Limit:     limit,
	}, nil
}

// ── GetCustomer ───────────────────────────────────────────────────────────────

func (s *CustomerService) GetCustomer(ctx context.Context, customerID, outletID uuid.UUID) (*CustomerResponse, error) {
	c, err := s.customerRepo.GetByID(ctx, customerID, outletID)
	if err != nil {
		return nil, fmt.Errorf("GetCustomer: %w", err)
	}
	return customerToResponse(c), nil
}

// ── UpdateCustomer ────────────────────────────────────────────────────────────

func (s *CustomerService) UpdateCustomer(ctx context.Context, customerID, outletID uuid.UUID, req UpdateCustomerRequest) (*CustomerResponse, error) {
	c, err := s.customerRepo.GetByID(ctx, customerID, outletID)
	if err != nil {
		return nil, fmt.Errorf("UpdateCustomer: ambil pelanggan: %w", err)
	}

	c.Name = req.Name
	c.Phone = req.Phone
	c.Notes = req.Notes
	if req.IsBlacklisted != nil {
		c.IsBlacklisted = *req.IsBlacklisted
	}
	if req.BlacklistReason != nil {
		c.BlacklistReason = req.BlacklistReason
	}

	if err := s.customerRepo.Update(ctx, c); err != nil {
		return nil, fmt.Errorf("UpdateCustomer: simpan: %w", err)
	}
	return customerToResponse(c), nil
}

// ── Helper ────────────────────────────────────────────────────────────────────

func customerToResponse(c *model.Customer) *CustomerResponse {
	return &CustomerResponse{
		ID:              c.ID,
		Name:            c.Name,
		Phone:           c.Phone,
		TotalOrders:     c.TotalOrders,
		Notes:           c.Notes,
		IsBlacklisted:   c.IsBlacklisted,
		BlacklistReason: c.BlacklistReason,
	}
}
