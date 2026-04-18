package database

import (
	"database/sql"
	"errors"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(db *sql.DB, migrationsPath string) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		slog.Error("Gagal membuat migration driver", "error", err)
		panic(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		slog.Error("Gagal inisialisasi migrator", "error", err)
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("Migration: tidak ada perubahan")
			return
		}
		slog.Error("Gagal menjalankan migration", "error", err)
		panic(err)
	}

	slog.Info("Migration: semua migration berhasil dijalankan")
}
