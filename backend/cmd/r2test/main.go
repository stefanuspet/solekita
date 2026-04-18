package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/stefanuspet/solekita/backend/internal/config"
	"github.com/stefanuspet/solekita/backend/internal/model"
	"github.com/stefanuspet/solekita/backend/internal/storage"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	r2, err := storage.NewR2(cfg)
	if err != nil {
		slog.Error("gagal init R2", "error", err)
		os.Exit(1)
	}

	// ── T118: Upload dummy file ───────────────────────────────────────────────
	slog.Info("=== T118: Upload dummy file ===")

	dummyData := []byte("solekita r2 test " + time.Now().String())
	dummyKey := fmt.Sprintf("test/dummy-%s.txt", uuid.New())

	if err := r2.Upload(ctx, dummyKey, dummyData, "text/plain"); err != nil {
		slog.Error("upload gagal", "error", err)
		os.Exit(1)
	}
	slog.Info("upload berhasil", "key", dummyKey)

	signedURL, err := r2.GetSignedURL(ctx, dummyKey, time.Hour)
	if err != nil {
		slog.Error("get signed URL gagal", "error", err)
		os.Exit(1)
	}
	slog.Info("signed URL", "url", signedURL)

	resp, err := http.Get(signedURL)
	if err != nil {
		slog.Error("akses URL gagal", "error", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		slog.Info("akses URL berhasil", "status", resp.StatusCode)
	} else {
		slog.Error("akses URL gagal", "status", resp.StatusCode)
		os.Exit(1)
	}

	// Cleanup dummy
	if err := r2.Delete(ctx, dummyKey); err != nil {
		slog.Warn("hapus dummy gagal", "error", err)
	} else {
		slog.Info("dummy dihapus dari bucket")
	}

	// ── T119: Upload foto besar, cek kompresi < 500KB ─────────────────────────
	slog.Info("=== T119: Test kompresi foto ===")

	// Generate gambar JPEG ~2MB (3000x2000 solid color)
	bigImg := image.NewRGBA(image.Rect(0, 0, 3000, 2000))
	for y := 0; y < 2000; y++ {
		for x := 0; x < 3000; x++ {
			bigImg.Set(x, y, color.RGBA{
				R: uint8(x % 256),
				G: uint8(y % 256),
				B: 128,
				A: 255,
			})
		}
	}

	var rawBuf bytes.Buffer
	if err := jpeg.Encode(&rawBuf, bigImg, &jpeg.Options{Quality: 95}); err != nil {
		slog.Error("encode gambar test gagal", "error", err)
		os.Exit(1)
	}
	rawSizeKB := rawBuf.Len() / 1024
	slog.Info("ukuran foto asli", "size_kb", rawSizeKB)

	dummyOrderID := uuid.New()
	dummyOutletID := uuid.New()

	key, err := r2.UploadPhoto(ctx, &rawBuf, dummyOrderID, dummyOutletID, model.PhotoTypeBefore)
	if err != nil {
		slog.Error("UploadPhoto gagal", "error", err)
		os.Exit(1)
	}
	slog.Info("foto terupload", "key", key)

	// Cek ukuran file yang tersimpan via HEAD request
	photoURL, err := r2.GetSignedURL(ctx, key, time.Hour)
	if err != nil {
		slog.Error("get signed URL foto gagal", "error", err)
		os.Exit(1)
	}

	headResp, err := http.Head(photoURL)
	if err != nil {
		slog.Error("HEAD request gagal", "error", err)
		os.Exit(1)
	}
	defer headResp.Body.Close()

	storedSizeKB := headResp.ContentLength / 1024
	slog.Info("ukuran foto di R2", "size_kb", storedSizeKB, "content_length", headResp.ContentLength)

	if storedSizeKB <= 500 {
		slog.Info("kompresi BERHASIL — foto < 500KB", "size_kb", storedSizeKB)
	} else {
		slog.Error("kompresi GAGAL — foto masih > 500KB", "size_kb", storedSizeKB)
		os.Exit(1)
	}

	// Cleanup foto test
	if err := r2.Delete(ctx, key); err != nil {
		slog.Warn("hapus foto test gagal", "error", err)
	} else {
		slog.Info("foto test dihapus dari bucket")
	}

	slog.Info("=== Semua test lulus ===")
}
