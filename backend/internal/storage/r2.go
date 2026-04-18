package storage

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	appcfg "github.com/stefanuspet/solekita/backend/internal/config"
)

type R2Storage struct {
	client     *s3.Client
	presigner  *s3.PresignClient
	bucketName string
}

func NewR2(cfg *appcfg.Config) (*R2Storage, error) {
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.R2AccountID)

	r2Config, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("auto"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.R2AccessKeyID,
			cfg.R2SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("NewR2: load config: %w", err)
	}

	client := s3.NewFromConfig(r2Config, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	return &R2Storage{
		client:     client,
		presigner:  s3.NewPresignClient(client),
		bucketName: cfg.R2BucketName,
	}, nil
}

// Upload mengunggah raw bytes ke R2 dengan key dan content type yang ditentukan.
func (r *R2Storage) Upload(ctx context.Context, key string, data []byte, contentType string) error {
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("R2Storage.Upload: %w", err)
	}
	return nil
}

// GetSignedURL menghasilkan presigned URL untuk mengakses file secara sementara.
func (r *R2Storage) GetSignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	req, err := r.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("R2Storage.GetSignedURL: %w", err)
	}
	return req.URL, nil
}

// Delete menghapus file dari R2 berdasarkan key.
func (r *R2Storage) Delete(ctx context.Context, key string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("R2Storage.Delete: %w", err)
	}
	return nil
}

// ObjectInfo merepresentasikan metadata satu objek di R2.
type ObjectInfo struct {
	Key          string
	LastModified time.Time
	Size         int64
}

// ListObjects mencantumkan semua objek dengan prefix tertentu dari bucket yang ditentukan.
func (r *R2Storage) ListObjects(ctx context.Context, bucket, prefix string) ([]ObjectInfo, error) {
	var result []ObjectInfo
	var continuationToken *string

	for {
		resp, err := r.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(bucket),
			Prefix:            aws.String(prefix),
			ContinuationToken: continuationToken,
		})
		if err != nil {
			return nil, fmt.Errorf("R2Storage.ListObjects: %w", err)
		}

		for _, obj := range resp.Contents {
			lm := time.Time{}
			if obj.LastModified != nil {
				lm = *obj.LastModified
			}
			sz := int64(0)
			if obj.Size != nil {
				sz = *obj.Size
			}
			result = append(result, ObjectInfo{
				Key:          aws.ToString(obj.Key),
				LastModified: lm,
				Size:         sz,
			})
		}

		if !aws.ToBool(resp.IsTruncated) {
			break
		}
		continuationToken = resp.NextContinuationToken
	}

	// Sort ascending by key (keys are timestamped, sehingga chronological)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Key < result[j].Key
	})
	return result, nil
}

// UploadToBucket mengunggah raw bytes ke bucket tertentu (berbeda dari default bucket).
// Digunakan untuk upload backup ke solekita-backup.
func (r *R2Storage) UploadToBucket(ctx context.Context, bucket, key string, data []byte, contentType string) error {
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("R2Storage.UploadToBucket: %w", err)
	}
	return nil
}

// DeleteFromBucket menghapus file dari bucket tertentu.
func (r *R2Storage) DeleteFromBucket(ctx context.Context, bucket, key string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("R2Storage.DeleteFromBucket: %w", err)
	}
	return nil
}
