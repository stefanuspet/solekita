package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// App
	AppEnv  string
	AppPort string

	// Database
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string

	// JWT
	JWTSecret          string
	JWTRefreshSecret   string
	JWTAccessExpiry    string
	JWTRefreshExpiry   string

	// Cloudflare R2
	R2AccountID         string
	R2AccessKeyID       string
	R2SecretAccessKey   string
	R2BucketName        string
	R2PublicURL         string
	R2BackupBucketName  string

	// Fonnte
	FonnteToken  string
	FonnteSender string

	// Tripay
	TripayAPIKey      string
	TripayPrivateKey  string
	TripayMerchantCode string
	TripayBaseURL     string

	// Admin
	AdminSecretKey string

	// CORS
	AllowedOrigins string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		// .env opsional di production (env var bisa di-inject langsung)
	}

	cfg := &Config{
		AppEnv:  getEnv("APP_ENV", "development"),
		AppPort: getEnv("APP_PORT", "8080"),

		DBHost:     mustGetEnv("DB_HOST"),
		DBPort:     mustGetEnv("DB_PORT"),
		DBName:     mustGetEnv("DB_NAME"),
		DBUser:     mustGetEnv("DB_USER"),
		DBPassword: mustGetEnv("DB_PASSWORD"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		JWTSecret:        mustGetEnv("JWT_SECRET"),
		JWTRefreshSecret: mustGetEnv("JWT_REFRESH_SECRET"),
		JWTAccessExpiry:  getEnv("JWT_ACCESS_EXPIRY", "15m"),
		JWTRefreshExpiry: getEnv("JWT_REFRESH_EXPIRY", "720h"),

		R2AccountID:        getEnv("R2_ACCOUNT_ID", ""),
		R2AccessKeyID:      getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretAccessKey:  getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2BucketName:       getEnv("R2_BUCKET_NAME", "solekita-dev"),
		R2PublicURL:        getEnv("R2_PUBLIC_URL", ""),
		R2BackupBucketName: getEnv("R2_BACKUP_BUCKET_NAME", "solekita-backup"),

		FonnteToken:  getEnv("FONNTE_TOKEN", ""),
		FonnteSender: getEnv("FONNTE_SENDER", ""),

		TripayAPIKey:       getEnv("TRIPAY_API_KEY", ""),
		TripayPrivateKey:   getEnv("TRIPAY_PRIVATE_KEY", ""),
		TripayMerchantCode: getEnv("TRIPAY_MERCHANT_CODE", ""),
		TripayBaseURL:      getEnv("TRIPAY_BASE_URL", "https://tripay.co.id/api-sandbox"),

		AdminSecretKey: getEnv("ADMIN_SECRET_KEY", ""),

		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:3001"),
	}

	return cfg
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBName, c.DBUser, c.DBPassword, c.DBSSLMode,
	)
}

func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

// mustGetEnv panic jika env var tidak di-set atau kosong.
func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("env var %q wajib diisi tapi kosong", key))
	}
	return val
}

// getEnv mengembalikan nilai env var atau fallback jika kosong.
func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
