package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	apperrors "github.com/stefanuspet/solekita/backend/internal/errors"
	"github.com/stefanuspet/solekita/backend/internal/model"
	"github.com/stefanuspet/solekita/backend/internal/repository"
)

type TreatmentService struct {
	treatmentRepo *repository.TreatmentRepository
}

func NewTreatmentService(treatmentRepo *repository.TreatmentRepository) *TreatmentService {
	return &TreatmentService{treatmentRepo: treatmentRepo}
}

// ── Request / Response ────────────────────────────────────────────────────────

type CreateTreatmentRequest struct {
	Name     string  `json:"name" binding:"required"`
	Material *string `json:"material"`
	Price    int     `json:"price" binding:"required,min=1"`
}

type UpdateTreatmentRequest struct {
	Name     string  `json:"name" binding:"required"`
	Material *string `json:"material"`
	Price    int     `json:"price" binding:"required,min=1"`
	IsActive *bool   `json:"is_active"`
}

type TreatmentResponse struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Material *string   `json:"material"`
	Price    int       `json:"price"`
	IsActive bool      `json:"is_active"`
}

// ── ListTreatments ────────────────────────────────────────────────────────────

func (s *TreatmentService) ListTreatments(ctx context.Context, outletID uuid.UUID, isActive *bool, material *string) ([]*TreatmentResponse, error) {
	treatments, err := s.treatmentRepo.ListByOutletID(ctx, outletID, isActive, material)
	if err != nil {
		return nil, fmt.Errorf("ListTreatments: %w", err)
	}

	res := make([]*TreatmentResponse, len(treatments))
	for i, t := range treatments {
		res[i] = treatmentToResponse(t)
	}
	return res, nil
}

// ── CreateTreatment ───────────────────────────────────────────────────────────

func (s *TreatmentService) CreateTreatment(ctx context.Context, outletID uuid.UUID, req CreateTreatmentRequest) (*TreatmentResponse, error) {
	t := &model.Treatment{
		OutletID: outletID,
		Name:     req.Name,
		Material: req.Material,
		Price:    req.Price,
		IsActive: true,
	}
	if err := s.treatmentRepo.Create(ctx, t); err != nil {
		return nil, fmt.Errorf("CreateTreatment: %w", err)
	}
	return treatmentToResponse(t), nil
}

// ── UpdateTreatment ───────────────────────────────────────────────────────────

func (s *TreatmentService) UpdateTreatment(ctx context.Context, treatmentID, outletID uuid.UUID, req UpdateTreatmentRequest) (*TreatmentResponse, error) {
	t, err := s.treatmentRepo.GetByID(ctx, treatmentID, outletID)
	if err != nil {
		return nil, fmt.Errorf("UpdateTreatment: ambil treatment: %w", err)
	}

	t.Name = req.Name
	t.Material = req.Material
	t.Price = req.Price
	if req.IsActive != nil {
		t.IsActive = *req.IsActive
	}

	if err := s.treatmentRepo.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("UpdateTreatment: simpan: %w", err)
	}
	return treatmentToResponse(t), nil
}

// ── DeleteTreatment ───────────────────────────────────────────────────────────

func (s *TreatmentService) DeleteTreatment(ctx context.Context, treatmentID, outletID uuid.UUID) error {
	if _, err := s.treatmentRepo.GetByID(ctx, treatmentID, outletID); err != nil {
		return fmt.Errorf("DeleteTreatment: ambil treatment: %w", err)
	}

	used, err := s.treatmentRepo.IsUsedInOrder(ctx, treatmentID)
	if err != nil {
		return fmt.Errorf("DeleteTreatment: cek penggunaan: %w", err)
	}
	if used {
		return apperrors.ErrUnprocessable.New(fmt.Errorf("treatment sudah pernah dipakai di order, tidak bisa dihapus"))
	}

	if err := s.treatmentRepo.Delete(ctx, treatmentID, outletID); err != nil {
		return fmt.Errorf("DeleteTreatment: hapus: %w", err)
	}
	return nil
}

// ── Helper ────────────────────────────────────────────────────────────────────

func treatmentToResponse(t *model.Treatment) *TreatmentResponse {
	return &TreatmentResponse{
		ID:       t.ID,
		Name:     t.Name,
		Material: t.Material,
		Price:    t.Price,
		IsActive: t.IsActive,
	}
}
