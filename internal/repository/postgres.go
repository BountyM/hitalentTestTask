package repository

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"runtime"

	"github.com/BountyM/hitalentTestTask/internal/config"
	"github.com/pressly/goose"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgresDB(cfg config.DB, log *slog.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s dbname=%s password=%s port=%s sslmode=%s",
		cfg.Host,
		cfg.Username,
		cfg.Dbname,
		cfg.Password,
		cfg.Port,
		cfg.Sslmode,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Запуск миграции Goose
	sqlDB, err := db.DB()
	if err != nil {
		log.Error("ошибка получения sql.DB",
			slog.String("error", err.Error()))
		return nil, err
	}

	log.Info("запуск миграций")

	// Определяем путь к текущему файлу (postgres.go)
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("failed to get current file path")
	}
	basepath := filepath.Dir(filename) // internal/repository
	migrationPath := filepath.Join(basepath, "migrations")
	if err := goose.Up(sqlDB, migrationPath); err != nil {
		log.Error("ошибка выполнения миграций",
			slog.String("error", err.Error()))
		return nil, err
	}

	log.Info("миграции успешно выполнены")

	return db, nil
}
