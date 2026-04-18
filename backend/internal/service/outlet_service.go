package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/stefanuspet/solekita/backend/internal/model"
	"github.com/stefanuspet/solekita/backend/internal/repository"
)

type OutletService struct {
	outletRepo *repository.OutletRepository
}

func NewOutletService(outletRepo *repository.OutletRepository) *OutletService {
	return &OutletService{outletRepo: outletRepo}
}

// ── Request ───────────────────────────────────────────────────────────────────

type UpdateOutletRequest struct {
	Name                 string  `json:"name" binding:"required"`
	Address              *string `json:"address"`
	Phone                *string `json:"phone"`
	OverdueThresholdDays *int    `json:"overdue_threshold_days" binding:"omitempty,min=1,max=30"`
}

// ── GetMyOutlet ───────────────────────────────────────────────────────────────

func (s *OutletService) GetMyOutlet(ctx context.Context, outletID uuid.UUID) (*OutletResponse, error) {
	outlet, err := s.outletRepo.GetByID(ctx, outletID)
	if err != nil {
		return nil, fmt.Errorf("GetMyOutlet: %w", err)
	}
	return outletToResponse(outlet), nil
}

// ── UpdateOutlet ──────────────────────────────────────────────────────────────

func (s *OutletService) UpdateOutlet(ctx context.Context, outletID uuid.UUID, req UpdateOutletRequest) (*OutletResponse, error) {
	outlet, err := s.outletRepo.GetByID(ctx, outletID)
	if err != nil {
		return nil, fmt.Errorf("UpdateOutlet: ambil outlet: %w", err)
	}

	outlet.Name = req.Name
	outlet.Address = req.Address
	outlet.Phone = req.Phone
	if req.OverdueThresholdDays != nil {
		outlet.OverdueThresholdDays = *req.OverdueThresholdDays
	}

	if err := s.outletRepo.Update(ctx, outlet); err != nil {
		return nil, fmt.Errorf("UpdateOutlet: simpan: %w", err)
	}

	return outletToResponse(outlet), nil
}

// ── Response ──────────────────────────────────────────────────────────────────

type OutletResponse struct {
	ID                   uuid.UUID `json:"id"`
	Name                 string    `json:"name"`
	Code                 string    `json:"code"`
	Address              *string   `json:"address"`
	Phone                *string   `json:"phone"`
	OwnerID              uuid.UUID `json:"owner_id"`
	SubscriptionStatus   string    `json:"subscription_status"`
	OverdueThresholdDays int       `json:"overdue_threshold_days"`
	IsActive             bool      `json:"is_active"`
}

func outletToResponse(o *model.Outlet) *OutletResponse {
	return &OutletResponse{
		ID:                   o.ID,
		Name:                 o.Name,
		Code:                 o.Code,
		Address:              o.Address,
		Phone:                o.Phone,
		OwnerID:              o.OwnerID,
		SubscriptionStatus:   string(o.SubscriptionStatus),
		OverdueThresholdDays: o.OverdueThresholdDays,
		IsActive:             o.IsActive,
	}
}
