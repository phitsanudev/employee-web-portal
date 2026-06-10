package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"employee-portal/backend/internal/config"
	"employee-portal/backend/internal/database"
	"employee-portal/backend/internal/handler"
	"employee-portal/backend/internal/repository"
	"employee-portal/backend/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	if err := database.AutoMigrate(db); err != nil {
		logger.Error("database migration failed", "error", err)
		os.Exit(1)
	}
	if err := database.Seed(db, logger, cfg); err != nil {
		logger.Error("database seed failed", "error", err)
		os.Exit(1)
	}

	repo := repository.NewProfileRepository(db)
	authService := service.NewAuthService(cfg, repo)
	profileService := service.NewProfileService(cfg, repo)
	adminService := service.NewAdminService(repo)
	apiHandler := handler.New(cfg, authService, profileService, adminService, logger)
	router := handler.BuildRouter(cfg, apiHandler, logger)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("api server started", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("api server failed", "error", err)
			os.Exit(1)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ShutdownTimeoutSeconds)*time.Second)
	defer cancel()
	logger.Info("shutdown signal received")
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("http shutdown failed", "error", err)
	}
	if err := database.Close(db); err != nil {
		logger.Error("database close failed", "error", err)
	}
	logger.Info("api server stopped gracefully")
}
