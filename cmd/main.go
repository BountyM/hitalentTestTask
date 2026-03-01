package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BountyM/hitalentTestTask/internal/config"
	"github.com/BountyM/hitalentTestTask/internal/handler"
	"github.com/BountyM/hitalentTestTask/internal/logger"
	"github.com/BountyM/hitalentTestTask/internal/repository"
	"github.com/BountyM/hitalentTestTask/internal/server"
	"github.com/BountyM/hitalentTestTask/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration", "error", err)
		return
	}
	log := logger.New(cfg.Logger)
	log.Info("Logger initialized")
	db, err := repository.NewPostgresDB(cfg.DB, log)
	if err != nil {
		log.Error("Failed to initialize database", "error", err)
		return
	}
	defer func() {
		sqlDB, err := db.DB()
		if err != nil {
			log.Warn("Error getting generic DB: %v", "error", err)
			return
		}
		if err := sqlDB.Close(); err != nil {
			log.Warn("Error closing database: %v", "error", err)
		}
	}()

	repo := repository.New(db)
	services := service.New(repo)
	handlers := handler.NewDepartmentHandler(services, log) // переименовано для избежания конфликта с пакетом

	srv := &server.Server{}

	// Канал для ошибок от HTTP сервера
	serverErr := make(chan error, 1)

	go func() {
		defer close(serverErr)
		addr := ":" + cfg.Port
		if err := srv.Run(addr, handler.SetupRoutes(handlers, log)); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	log.Info("API started")

	// Ожидание сигнала завершения или ошибки сервера
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-quit:
		log.Info("API shutting down")
	case err := <-serverErr:
		log.Error("Server failed with error", "error", err)
		return
	}

	// Graceful shutdown сервера
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Error occurred on server shutdown", "error", err)
	} else {
		log.Info("Server stopped gracefully")
	}
}
