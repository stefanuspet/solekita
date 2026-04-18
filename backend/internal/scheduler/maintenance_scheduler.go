package scheduler

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"time"

	"github.com/stefanuspet/solekita/backend/internal/model"
)

const (
	retentionPhotoBeforeMonths = 6
	retentionPhotoAfterMonths  = 3
	retentionSelfieMonths      = 1
	backupKeepCount            = 30
	backupPrefix               = "db/"
)

// CleanupExpiredPhotos — dipanggil cron Minggu 02:00.
// Hapus foto before > 6 bulan dan foto after > 3 bulan dari R2, set is_deleted = true.
func (s *Scheduler) CleanupExpiredPhotos() {
	if s.r2 == nil {
		slog.Warn("CleanupExpiredPhotos: R2 tidak dikonfigurasi, skip")
		return
	}

	ctx := context.Background()
	now := time.Now()

	type cleanupTarget struct {
		photoType model.PhotoType
		olderThan time.Time
		label     string
	}

	targets := []cleanupTarget{
		{model.PhotoTypeBefore, now.AddDate(0, -retentionPhotoBeforeMonths, 0), "before"},
		{model.PhotoTypeAfter, now.AddDate(0, -retentionPhotoAfterMonths, 0), "after"},
	}

	for _, t := range targets {
		photos, err := s.photoRepo.ListExpired(ctx, t.photoType, t.olderThan)
		if err != nil {
			slog.Error("CleanupExpiredPhotos: query gagal", "type", t.label, "error", err)
			continue
		}

		deleted, failed := 0, 0
		for _, p := range photos {
			if err := s.r2.Delete(ctx, p.R2Key); err != nil {
				slog.Warn("CleanupExpiredPhotos: hapus R2 gagal",
					"photo_id", p.ID, "key", p.R2Key, "error", err)
				failed++
				continue
			}
			if err := s.photoRepo.MarkDeleted(ctx, p.ID); err != nil {
				slog.Warn("CleanupExpiredPhotos: mark deleted gagal",
					"photo_id", p.ID, "error", err)
				failed++
				continue
			}
			deleted++
		}

		slog.Info("CleanupExpiredPhotos selesai",
			"type", t.label, "deleted", deleted, "failed", failed)
	}
}

// CleanupExpiredSelfies — dipanggil cron Minggu 02:00.
// Hapus selfie attendance > 1 bulan dari R2, set is_selfie_deleted = true.
func (s *Scheduler) CleanupExpiredSelfies() {
	if s.r2 == nil {
		slog.Warn("CleanupExpiredSelfies: R2 tidak dikonfigurasi, skip")
		return
	}

	ctx := context.Background()
	olderThan := time.Now().AddDate(0, -retentionSelfieMonths, 0)

	selfies, err := s.attendanceRepo.ListExpiredSelfies(ctx, olderThan)
	if err != nil {
		slog.Error("CleanupExpiredSelfies: query gagal", "error", err)
		return
	}

	deleted, failed := 0, 0
	for _, sel := range selfies {
		if err := s.r2.Delete(ctx, sel.R2Key); err != nil {
			slog.Warn("CleanupExpiredSelfies: hapus R2 gagal",
				"attendance_id", sel.ID, "key", sel.R2Key, "error", err)
			failed++
			continue
		}
		if err := s.attendanceRepo.MarkSelfieDeleted(ctx, sel.ID); err != nil {
			slog.Warn("CleanupExpiredSelfies: mark deleted gagal",
				"attendance_id", sel.ID, "error", err)
			failed++
			continue
		}
		deleted++
	}

	slog.Info("CleanupExpiredSelfies selesai", "deleted", deleted, "failed", failed)
}

// BackupDatabase — dipanggil cron setiap hari 03:00.
// Jalankan pg_dump, kompres dengan gzip, upload ke R2 backup bucket,
// kemudian pertahankan hanya 30 backup terbaru.
func (s *Scheduler) BackupDatabase() {
	if s.r2 == nil {
		slog.Warn("BackupDatabase: R2 tidak dikonfigurasi, skip")
		return
	}

	ctx := context.Background()
	cfg := s.cfg
	backupBucket := cfg.R2BackupBucketName

	filename := fmt.Sprintf("backup_%s.dump.gz", time.Now().Format("20060102_150405"))
	r2Key := backupPrefix + filename

	slog.Info("BackupDatabase: mulai pg_dump", "file", filename)

	// ── 1. Jalankan pg_dump ───────────────────────────────────────────────────
	pgDump := exec.CommandContext(ctx,
		"pg_dump",
		"-h", cfg.DBHost,
		"-p", cfg.DBPort,
		"-U", cfg.DBUser,
		"-d", cfg.DBName,
		"--format=plain",
		"--no-password",
	)
	pgDump.Env = append(pgDump.Environ(), "PGPASSWORD="+cfg.DBPassword)

	dumpOutput, err := pgDump.Output()
	if err != nil {
		slog.Error("BackupDatabase: pg_dump gagal", "error", err)
		return
	}

	// ── 2. Kompres dengan gzip ────────────────────────────────────────────────
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(dumpOutput); err != nil {
		slog.Error("BackupDatabase: gzip write gagal", "error", err)
		return
	}
	if err := gz.Close(); err != nil {
		slog.Error("BackupDatabase: gzip close gagal", "error", err)
		return
	}

	// ── 3. Upload ke R2 backup bucket ─────────────────────────────────────────
	if err := s.r2.UploadToBucket(ctx, backupBucket, r2Key, buf.Bytes(), "application/gzip"); err != nil {
		slog.Error("BackupDatabase: upload R2 gagal", "key", r2Key, "error", err)
		return
	}
	slog.Info("BackupDatabase: upload berhasil", "key", r2Key, "size_kb", buf.Len()/1024)

	// ── 4. Prune: pertahankan hanya 30 backup terbaru ─────────────────────────
	objects, err := s.r2.ListObjects(ctx, backupBucket, backupPrefix)
	if err != nil {
		slog.Warn("BackupDatabase: list objek gagal, skip prune", "error", err)
		return
	}

	if len(objects) <= backupKeepCount {
		return
	}

	// Objects sudah diurutkan ascending by key — hapus yang paling lama (indeks awal)
	toDelete := objects[:len(objects)-backupKeepCount]
	for _, obj := range toDelete {
		if err := s.r2.DeleteFromBucket(ctx, backupBucket, obj.Key); err != nil {
			slog.Warn("BackupDatabase: hapus backup lama gagal", "key", obj.Key, "error", err)
		} else {
			slog.Info("BackupDatabase: backup lama dihapus", "key", obj.Key)
		}
	}
}
