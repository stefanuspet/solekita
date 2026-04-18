package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/stefanuspet/solekita/backend/internal/notification"
	"github.com/stefanuspet/solekita/backend/internal/repository"
)

type DashboardService struct {
	dashboardRepo *repository.DashboardRepository
	outletRepo    *repository.OutletRepository
	fonnte        *notification.FonnteClient
}

func NewDashboardService(dashboardRepo *repository.DashboardRepository, outletRepo *repository.OutletRepository, fonnte *notification.FonnteClient) *DashboardService {
	return &DashboardService{dashboardRepo: dashboardRepo, outletRepo: outletRepo, fonnte: fonnte}
}

// ── Response types ────────────────────────────────────────────────────────────

type KasirSummaryResponse struct {
	KasirID    uuid.UUID `json:"kasir_id"`
	KasirName  string    `json:"kasir_name"`
	OrderCount int       `json:"order_count"`
	Revenue    int       `json:"revenue"`
}

type PaymentMethodSummaryResponse struct {
	Method     string `json:"method"`
	OrderCount int    `json:"order_count"`
	Amount     int    `json:"amount"`
}

type OverdueOrderResponse struct {
	ID            uuid.UUID `json:"id"`
	OrderNumber   string    `json:"order_number"`
	CustomerID    uuid.UUID `json:"customer_id"`
	CustomerName  string    `json:"customer_name"`
	CustomerPhone string    `json:"customer_phone"`
	DaysOverdue   int       `json:"days_overdue"`
	TotalPrice    int       `json:"total_price"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type DailySummaryResponse struct {
	Date           time.Time                       `json:"date"`
	Revenue        int                             `json:"revenue"`
	OrderCount     int                             `json:"order_count"`
	StaffSummary   []*KasirSummaryResponse         `json:"staff_summary"`
	PaymentSummary []*PaymentMethodSummaryResponse `json:"payment_summary"`
	OverdueOrders  []*OverdueOrderResponse         `json:"overdue_orders"`
}

type DailyStatPointResponse struct {
	Date       time.Time `json:"date"`
	Revenue    int       `json:"revenue"`
	OrderCount int       `json:"order_count"`
}

// ── GetSummary ────────────────────────────────────────────────────────────────

// GetSummary mengambil summary harian beserta daftar order overdue.
// date: tanggal yang diminta (default: hari ini).
func (s *DashboardService) GetSummary(ctx context.Context, outletID uuid.UUID, date time.Time) (*DailySummaryResponse, error) {
	outlet, err := s.outletRepo.GetByID(ctx, outletID)
	if err != nil {
		return nil, err
	}

	summary, err := s.dashboardRepo.GetDailySummary(ctx, outletID, date)
	if err != nil {
		return nil, err
	}

	overdue, err := s.dashboardRepo.GetOverdueOrders(ctx, outletID, outlet.OverdueThresholdDays)
	if err != nil {
		return nil, err
	}

	res := &DailySummaryResponse{
		Date:       summary.Date,
		Revenue:    summary.Revenue,
		OrderCount: summary.OrderCount,
	}

	res.StaffSummary = make([]*KasirSummaryResponse, 0, len(summary.StaffSummary))
	for _, k := range summary.StaffSummary {
		res.StaffSummary = append(res.StaffSummary, &KasirSummaryResponse{
			KasirID:    k.KasirID,
			KasirName:  k.KasirName,
			OrderCount: k.OrderCount,
			Revenue:    k.Revenue,
		})
	}

	res.PaymentSummary = make([]*PaymentMethodSummaryResponse, 0, len(summary.PaymentSummary))
	for _, p := range summary.PaymentSummary {
		res.PaymentSummary = append(res.PaymentSummary, &PaymentMethodSummaryResponse{
			Method:     p.Method,
			OrderCount: p.OrderCount,
			Amount:     p.Amount,
		})
	}

	res.OverdueOrders = make([]*OverdueOrderResponse, 0, len(overdue))
	for _, o := range overdue {
		res.OverdueOrders = append(res.OverdueOrders, &OverdueOrderResponse{
			ID:            o.ID,
			OrderNumber:   o.OrderNumber,
			CustomerID:    o.CustomerID,
			CustomerName:  o.CustomerName,
			CustomerPhone: o.CustomerPhone,
			DaysOverdue:   o.DaysOverdue,
			TotalPrice:    o.TotalPrice,
			UpdatedAt:     o.UpdatedAt,
		})
	}

	return res, nil
}

// ── SendOverdueReminders ──────────────────────────────────────────────────────

type ReminderResult struct {
	Sent   int `json:"sent"`
	Failed int `json:"failed"`
}

// SendOverdueReminders mengirim WA reminder ke setiap customer dengan order menggantung.
// Setiap pengiriman dijalankan dalam goroutine terpisah; method menunggu semua selesai.
func (s *DashboardService) SendOverdueReminders(ctx context.Context, outletID uuid.UUID) (*ReminderResult, error) {
	if s.fonnte == nil {
		return nil, fmt.Errorf("SendOverdueReminders: Fonnte belum dikonfigurasi")
	}

	outlet, err := s.outletRepo.GetByID(ctx, outletID)
	if err != nil {
		return nil, err
	}

	orders, err := s.dashboardRepo.GetOverdueOrders(ctx, outletID, outlet.OverdueThresholdDays)
	if err != nil {
		return nil, err
	}

	type result struct{ err error }
	ch := make(chan result, len(orders))

	for _, o := range orders {
		o := o
		go func() {
			msg := fmt.Sprintf(
				"Halo %s! 👟\n\nOrder laundry sepatu Anda (*%s*) sudah selesai %d hari yang lalu dan belum diambil.\n\nMohon segera diambil. Terima kasih!",
				o.CustomerName, o.OrderNumber, o.DaysOverdue,
			)
			ch <- result{err: s.fonnte.Send(ctx, o.CustomerPhone, msg)}
		}()
	}

	res := &ReminderResult{}
	for range orders {
		r := <-ch
		if r.err != nil {
			res.Failed++
			slog.Warn("SendOverdueReminders: gagal kirim WA", "error", r.err)
		} else {
			res.Sent++
		}
	}

	return res, nil
}

// ── GetStats ──────────────────────────────────────────────────────────────────

// GetStats mengambil data statistik harian untuk grafik.
// period: jumlah hari (7 atau 30).
func (s *DashboardService) GetStats(ctx context.Context, outletID uuid.UUID, period int) ([]*DailyStatPointResponse, error) {
	if period != 7 && period != 30 {
		period = 7
	}

	points, err := s.dashboardRepo.GetOrderStats(ctx, outletID, period)
	if err != nil {
		return nil, err
	}

	res := make([]*DailyStatPointResponse, 0, len(points))
	for _, p := range points {
		res = append(res, &DailyStatPointResponse{
			Date:       p.Date,
			Revenue:    p.Revenue,
			OrderCount: p.OrderCount,
		})
	}
	return res, nil
}
