package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/stefanuspet/solekita/backend/internal/model"
)

// SendTrialReminderH3 — WA #2, dipanggil cron 09:00 setiap hari.
// Kirim reminder ke outlet yang trial-nya berakhir 3 hari lagi.
func (s *Scheduler) SendTrialReminderH3() {
	s.sendTrialReminderForDate(context.Background(), time.Now().AddDate(0, 0, 3), msgTrialH3)
}

// SendTrialReminderH1 — WA #3, dipanggil cron 09:00 setiap hari.
// Kirim reminder ke outlet yang trial-nya berakhir besok.
func (s *Scheduler) SendTrialReminderH1() {
	s.sendTrialReminderForDate(context.Background(), time.Now().AddDate(0, 0, 1), msgTrialH1)
}

// SendTrialExpiredNotif — WA #4, dipanggil cron 09:00 setiap hari.
// Kirim notif ke outlet yang trial-nya berakhir hari ini.
func (s *Scheduler) SendTrialExpiredNotif() {
	s.sendTrialReminderForDate(context.Background(), time.Now(), msgTrialH0)
}

// sendTrialReminderForDate mengambil outlet yang trial-nya berakhir pada tanggal tertentu
// lalu mengirim WA ke owner masing-masing secara goroutine (best-effort).
func (s *Scheduler) sendTrialReminderForDate(ctx context.Context, date time.Time, msgFn func(outletName string, endsAt time.Time) string) {
	if s.fonnte == nil {
		return
	}

	outlets, err := s.subRepo.ListTrialEndingOn(ctx, date)
	if err != nil {
		slog.Error("sendTrialReminderForDate: query gagal", "date", date.Format("2006-01-02"), "error", err)
		return
	}

	for _, o := range outlets {
		o := o
		go func() {
			msg := msgFn(o.OutletName, o.TrialEndsAt)
			if err := s.fonnte.Send(ctx, o.OwnerPhone, msg); err != nil {
				slog.Warn("trial reminder WA gagal", "outlet_id", o.OutletID, "phone", o.OwnerPhone, "error", err)
			} else {
				slog.Info("trial reminder WA terkirim", "outlet_id", o.OutletID, "date", date.Format("2006-01-02"))
			}
		}()
	}
}

// SuspendExpiredTrials dipanggil cron 00:00 setiap hari.
// Mencari outlet yang trial-nya sudah habis dan masih berstatus trial → set suspended.
// WA notifikasi tidak dikirim ulang di sini karena WA #4 sudah dikirim pagi hari sebelumnya
// (sendTrialReminders H-0) yang sudah menyebutkan "akses kasir sementara terkunci".
func (s *Scheduler) SuspendExpiredTrials() {
	ctx := context.Background()

	outlets, err := s.subRepo.ListExpiredTrials(ctx)
	if err != nil {
		slog.Error("suspendExpiredTrials: query gagal", "error", err)
		return
	}

	if len(outlets) == 0 {
		return
	}

	slog.Info("suspendExpiredTrials: memproses outlet", "count", len(outlets))

	for _, o := range outlets {
		if err := s.outletRepo.UpdateSubscriptionStatus(ctx, o.OutletID, model.SubscriptionStatusSuspended); err != nil {
			slog.Error("suspendExpiredTrials: gagal suspend outlet", "outlet_id", o.OutletID, "error", err)
			continue
		}
		// Catat waktu suspend agar MarkInactiveOutlets bisa hitung 30 hari
		if err := s.subRepo.UpdateSuspendedAt(ctx, o.OutletID); err != nil {
			slog.Warn("suspendExpiredTrials: gagal catat suspended_at", "outlet_id", o.OutletID, "error", err)
		}
		slog.Info("suspendExpiredTrials: outlet disuspend", "outlet_id", o.OutletID, "outlet_name", o.OutletName)
	}
}

// ── WA message templates (sesuai onboarding-flow.md Section 2) ───────────────

// msgTrialH3 — WA #2: Reminder Trial H-3
func msgTrialH3(outletName string, endsAt time.Time) string {
	return fmt.Sprintf(
		"Halo %s 👋\n\n"+
			"Trial Solekita Anda berakhir 3 hari lagi (%s).\n\n"+
			"Jangan sampai putus akses — pilih paket sekarang:\n\n"+
			"💳 Bulanan: Rp29.000/bulan\n"+
			"💰 6 Bulan: Rp156.600 (hemat 10%%)\n\n"+
			"Buka aplikasi > Langganan > Pilih Paket untuk aktifkan.\n\n"+
			"Ada pertanyaan? Balas WA ini 🙏",
		outletName,
		formatTanggal(endsAt),
	)
}

// msgTrialH1 — WA #3: Reminder Trial H-1
func msgTrialH1(outletName string, endsAt time.Time) string {
	return fmt.Sprintf(
		"Halo %s,\n\n"+
			"Trial Anda berakhir BESOK (%s).\n\n"+
			"Pilih paket agar akses tidak terputus:\n\n"+
			"💳 Bulanan — Rp29.000/bulan\n\n"+
			"💰 6 Bulan di Muka — Rp156.600 (hemat Rp17.400)\n\n"+
			"Buka aplikasi > Langganan > Pilih Paket.\n\n"+
			"Semua data Anda tetap aman apapun yang terjadi 🙏",
		outletName,
		formatTanggal(endsAt),
	)
}

// msgTrialH0 — WA #4: Trial Habis, Belum Bayar
func msgTrialH0(outletName string, _ time.Time) string {
	return fmt.Sprintf(
		"Halo %s,\n\n"+
			"Trial Solekita Anda sudah berakhir hari ini.\n\n"+
			"Akses kasir sementara terkunci, tapi semua data Anda aman.\n\n"+
			"Lanjutkan dengan memilih paket:\n\n"+
			"💳 Bulanan — Rp29.000/bulan\n\n"+
			"💰 6 Bulan di Muka — Rp156.600 (hemat 10%%)\n\n"+
			"Buka aplikasi > Langganan > Pilih Paket.\n\n"+
			"Akses pulih otomatis setelah pembayaran dikonfirmasi ✅",
		outletName,
	)
}

// formatTanggal mengformat waktu ke bahasa Indonesia, misal "14 April 2026".
func formatTanggal(t time.Time) string {
	bulan := [...]string{
		"Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}
	return fmt.Sprintf("%d %s %d", t.Day(), bulan[t.Month()-1], t.Year())
}
