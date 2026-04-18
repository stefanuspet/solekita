package storage

import (
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"io"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/stefanuspet/solekita/backend/internal/model"
)

const (
	maxPhotoSizeKB  = 500
	maxSelfieSizeKB = 300
	// Lebar maksimum foto sebelum dikompresi — preserves aspect ratio
	maxPhotoWidth = 1280
)

// UploadPhoto mengkompresi foto ke < 500KB lalu upload ke R2.
// Return R2 key yang disimpan di DB (bukan URL).
// Key format: outlets/{outletID}/orders/{orderID}/{photoType}/{uuid}.jpg
func (r *R2Storage) UploadPhoto(ctx context.Context, file io.Reader, orderID, outletID uuid.UUID, photoType model.PhotoType) (string, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("UploadPhoto: baca file: %w", err)
	}

	compressed, err := compressImage(data, maxPhotoSizeKB, maxPhotoWidth)
	if err != nil {
		return "", fmt.Errorf("UploadPhoto: kompresi: %w", err)
	}

	key := fmt.Sprintf("outlets/%s/orders/%s/%s/%s.jpg", outletID, orderID, photoType, uuid.New())

	if err := r.Upload(ctx, key, compressed, "image/jpeg"); err != nil {
		return "", fmt.Errorf("UploadPhoto: upload: %w", err)
	}
	return key, nil
}

// UploadSelfie mengkompresi selfie absensi ke < 300KB lalu upload ke R2.
// Key format: outlets/{outletID}/selfies/{userID}/{uuid}.jpg
func (r *R2Storage) UploadSelfie(ctx context.Context, file io.Reader, userID, outletID uuid.UUID) (string, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("UploadSelfie: baca file: %w", err)
	}

	compressed, err := compressImage(data, maxSelfieSizeKB, maxPhotoWidth)
	if err != nil {
		return "", fmt.Errorf("UploadSelfie: kompresi: %w", err)
	}

	key := fmt.Sprintf("outlets/%s/selfies/%s/%s.jpg", outletID, userID, uuid.New())

	if err := r.Upload(ctx, key, compressed, "image/jpeg"); err != nil {
		return "", fmt.Errorf("UploadSelfie: upload: %w", err)
	}
	return key, nil
}

// compressImage mengkompresi gambar ke dalam batas ukuran tertentu (KB).
// Resize dulu jika lebar melebihi maxWidth, lalu turunkan JPEG quality secara bertahap.
func compressImage(data []byte, maxSizeKB, maxWidth int) ([]byte, error) {
	img, err := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
	if err != nil {
		return nil, fmt.Errorf("decode gambar: %w", err)
	}

	// Resize jika terlalu lebar, preserves aspect ratio
	if img.Bounds().Dx() > maxWidth {
		img = imaging.Resize(img, maxWidth, 0, imaging.Lanczos)
	}

	// Turunkan quality secara bertahap sampai ukuran di bawah batas
	for quality := 85; quality >= 30; quality -= 10 {
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
			return nil, fmt.Errorf("encode JPEG quality %d: %w", quality, err)
		}
		if buf.Len() <= maxSizeKB*1024 {
			return buf.Bytes(), nil
		}
	}

	// Quality 30 masih terlalu besar — resize lebih kecil lagi lalu encode ulang
	img = imaging.Resize(img, maxWidth/2, 0, imaging.Lanczos)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 30}); err != nil {
		return nil, fmt.Errorf("encode JPEG fallback: %w", err)
	}
	return buf.Bytes(), nil
}
