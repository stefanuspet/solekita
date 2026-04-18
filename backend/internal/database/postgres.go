package database

import (
	"database/sql"
	"log/slog"
	"time"

	"github.com/stefanuspet/solekita/backend/internal/config"

	_ "github.com/lib/pq"
)

func Connect(cfg *config.Config) *sql.DB {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		slog.Error("Gagal membuka koneksi database", "error", err)
		panic(err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	if err := db.Ping(); err != nil {
		slog.Error("Gagal ping database", "error", err)
		panic(err)
	}

	slog.Info("Database terhubung",
		"host", cfg.DBHost,
		"port", cfg.DBPort,
		"db", cfg.DBName,
	)

	return db
}
