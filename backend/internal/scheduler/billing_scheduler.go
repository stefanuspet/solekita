package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/stefanuspet/solekita/backend/internal/model"
)

// GenerateMonthlyInvoices — dipanggil cron 09:00 setiap hari.
// Cari outlet aktif yang next_due_date = hari ini, buat tagihan Tripay, kirim WA #8.
func (s *Scheduler) GenerateMonthlyInvoices() {
	ctx := context.Background()

	outlets, err := s.subRepo.ListActiveWithDueDateOn(ctx, time.Now())
	if err != nil {
		slog.Error("GenerateMonthlyInvoices: query gagal", "error", err)
		return
	}

	if len(outlets) == 0 {
		return
	}

	slog.Info("GenerateMonthlyInvoices: memproses outlet", "count", len(outlets))

	for _, o := range outlets {
		o := o
		go func() {
			result, err := s.subscriptionService.GenerateBillingInvoice(ctx, o.OutletID)
			if err != nil {
				slog.Error("GenerateMonthlyInvoices: gagal buat invoice",
					"outlet_id", o.OutletID, "error", err)
				return
			}
			slog.Info("GenerateMonthlyInvoices: invoice dibuat",
				"outlet_id", o.OutletID, "invoice_id", result.InvoiceID)

			if s.fonnte == nil {
				return
			}
			msg := msgBillingH0(o.OutletName, result.Amount, result.DueDate, result.PaymentURL)
			if err := s.fonnte.Send(ctx, o.OwnerPhone, msg); err != nil {
				slog.Warn("GenerateMonthlyInvoices: WA gagal",
					"outlet_id", o.OutletID, "error", err)
			}
		}()
	}
}

// SendBillingReminderH3 — dipanggil cron 09:00 setiap hari.
// Kirim WA #6 ke outlet aktif yang next_due_date = hari ini + 3.
func (s *Scheduler) SendBillingReminderH3() {
	if s.fonnte == nil {
		return
	}
	s.sendBillingReminderForDate(context.Background(), time.Now().AddDate(0, 0, 3), msgBillingH3)
}

// SendBillingReminderH1 — dipanggil cron 09:00 setiap hari.
// Kirim WA #7 ke outlet aktif yang next_due_date = besok.
func (s *Scheduler) SendBillingReminderH1() {
	if s.fonnte == nil {
		return
	}
	s.sendBillingReminderForDate(context.Background(), time.Now().AddDate(0, 0, 1), msgBillingH1)
}

// sendBillingReminderForDate adalah helper untuk H-3 dan H-1 reminder.
func (s *Scheduler) sendBillingReminderForDate(ctx context.Context, date time.Time, msgFn func(outletName string, dueDate time.Time) string) {
	outlets, err := s.subRepo.ListActiveWithDueDateOn(ctx, date)
	if err != nil {
		slog.Error("sendBillingReminderForDate: query gagal", "date", date.Format("2006-01-02"), "error", err)
		return
	}

	for _, o := range outlets {
		o := o
		go func() {
			dueDate := time.Time{}
			if o.NextDueDate != nil {
				dueDate = *o.NextDueDate
			}
			msg := msgFn(o.OutletName, dueDate)
			if err := s.fonnte.Send(ctx, o.OwnerPhone, msg); err != nil {
				slog.Warn("billing reminder WA gagal",
					"outlet_id", o.OutletID, "date", date.Format("2006-01-02"), "error", err)
			} else {
				slog.Info("billing reminder WA terkirim",
					"outlet_id", o.OutletID, "date", date.Format("2006-01-02"))
			}
		}()
	}
}

// SuspendUnpaidOutlets — dipanggil cron 00:10 setiap hari.
// Cari outlet aktif yang next_due_date + 3 hari < now() (grace period habis),
// set status suspended dan kirim WA #9.
func (s *Scheduler) SuspendUnpaidOutlets() {
	ctx := context.Background()

	outlets, err := s.subRepo.ListActiveOverdue(ctx)
	if err != nil {
		slog.Error("SuspendUnpaidOutlets: query gagal", "error", err)
		return
	}

	if len(outlets) == 0 {
		return
	}

	slog.Info("SuspendUnpaidOutlets: memproses outlet", "count", len(outlets))

	for _, o := range outlets {
		if err := s.outletRepo.UpdateSubscriptionStatus(ctx, o.OutletID, model.SubscriptionStatusSuspended); err != nil {
			slog.Error("SuspendUnpaidOutlets: gagal suspend", "outlet_id", o.OutletID, "error", err)
			continue
		}
		if err := s.subRepo.UpdateSuspendedAt(ctx, o.OutletID); err != nil {
			slog.Warn("SuspendUnpaidOutlets: gagal catat suspended_at", "outlet_id", o.OutletID, "error", err)
		}
		slog.Info("SuspendUnpaidOutlets: outlet disuspend", "outlet_id", o.OutletID, "outlet_name", o.OutletName)

		if s.fonnte != nil {
			o := o
			go func() {
				dueDate := time.Time{}
				if o.NextDueDate != nil {
					dueDate = *o.NextDueDate
				}
				msg := msgBillingSuspended(o.OutletName, dueDate)
				if err := s.fonnte.Send(ctx, o.OwnerPhone, msg); err != nil {
					slog.Warn("SuspendUnpaidOutlets: WA gagal", "outlet_id", o.OutletID, "error", err)
				}
			}()
		}
	}
}

// MarkInactiveOutlets — dipanggil cron 00:15 setiap hari.
// Cari outlet yang sudah suspended > 30 hari, set status inactive.
func (s *Scheduler) MarkInactiveOutlets() {
	ctx := context.Background()

	outlets, err := s.subRepo.ListSuspendedOlderThan(ctx, 30)
	if err != nil {
		slog.Error("MarkInactiveOutlets: query gagal", "error", err)
		return
	}

	if len(outlets) == 0 {
		return
	}

	slog.Info("MarkInactiveOutlets: memproses outlet", "count", len(outlets))

	for _, o := range outlets {
		if err := s.outletRepo.UpdateSubscriptionStatus(ctx, o.OutletID, model.SubscriptionStatusInactive); err != nil {
			slog.Error("MarkInactiveOutlets: gagal tandai inactive", "outlet_id", o.OutletID, "error", err)
		} else {
			slog.Info("MarkInactiveOutlets: outlet dinonaktifkan", "outlet_id", o.OutletID, "outlet_name", o.OutletName)
		}
	}
}

// ── WA message templates (sesuai onboarding-flow.md Section 2) ───────────────

// msgBillingH3 — WA #6: Reminder Tagihan H-3
func msgBillingH3(outletName string, dueDate time.Time) string {
	return fmt.Sprintf(
		"Halo %s,\n\n"+
			"Tagihan Solekita Anda akan jatuh tempo dalam 3 hari (%s).\n\n"+
			"Jumlah: Rp29.000\n\n"+
			"Buka aplikasi > Langganan untuk melakukan pembayaran.\n\n"+
			"Terima kasih 🙏",
		outletName,
		formatTanggal(dueDate),
	)
}

// msgBillingH1 — WA #7: Reminder Tagihan H-1
func msgBillingH1(outletName string, dueDate time.Time) string {
	return fmt.Sprintf(
		"Halo %s,\n\n"+
			"Tagihan Solekita jatuh tempo BESOK (%s).\n"+
			"Jumlah: Rp29.000\n\n"+
			"Buka aplikasi > Langganan untuk bayar sekarang.\n\n"+
			"Jika tidak bayar dalam 3 hari setelah jatuh tempo,\n"+
			"akses akan ditangguhkan sementara.",
		outletName,
		formatTanggal(dueDate),
	)
}

// msgBillingH0 — WA #8: Tagihan Jatuh Tempo Hari Ini
func msgBillingH0(outletName string, amount int, dueDate time.Time, paymentURL string) string {
	return fmt.Sprintf(
		"Halo %s,\n\n"+
			"Tagihan Solekita hari ini:\n"+
			"Jumlah: Rp%s\n"+
			"Jatuh tempo: Hari ini (%s)\n\n"+
			"Bayar di sini: %s\n\n"+
			"Akses tetap aktif selama 3 hari ke depan.",
		outletName,
		formatRupiah(amount),
		formatTanggal(dueDate),
		paymentURL,
	)
}

// msgBillingSuspended — WA #9: Akses Ditangguhkan
func msgBillingSuspended(outletName string, dueDate time.Time) string {
	return fmt.Sprintf(
		"Halo %s,\n\n"+
			"Akses Solekita Anda sementara ditangguhkan karena\n"+
			"tagihan belum dibayar sejak %s.\n\n"+
			"Semua data Anda tetap aman.\n\n"+
			"Buka aplikasi > Langganan untuk aktifkan kembali.\n\n"+
			"Akses pulih otomatis setelah pembayaran ✅",
		outletName,
		formatTanggal(dueDate),
	)
}

// formatRupiah memformat angka sebagai nominal rupiah tanpa simbol, misal "29.000".
func formatRupiah(amount int) string {
	s := fmt.Sprintf("%d", amount)
	// sisipkan titik setiap 3 digit dari kanan
	result := make([]byte, 0, len(s)+len(s)/3)
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, '.')
		}
		result = append(result, byte(c))
	}
	return string(result)
}
