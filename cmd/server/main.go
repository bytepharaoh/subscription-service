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

	"github.com/bytepharaoh/subscription-service/internal/config"
	"github.com/bytepharaoh/subscription-service/internal/db"
	"github.com/bytepharaoh/subscription-service/internal/handler"
	"github.com/bytepharaoh/subscription-service/internal/logger"
	dbgen "github.com/bytepharaoh/subscription-service/internal/repository/db"
	"github.com/bytepharaoh/subscription-service/internal/repository/postgres"
	"github.com/bytepharaoh/subscription-service/internal/server"
	"github.com/bytepharaoh/subscription-service/internal/service"

	_ "github.com/bytepharaoh/subscription-service/docs"
)

// @title           Subscription Service API
// @version         1.0
// @description     REST API for managing user subscriptions
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	// config
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	// logger
	log := logger.New(cfg.AppEnv)
	slog.SetDefault(log)

	// database pool
	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseDSN, log)
	if err != nil {
		log.Error("failed to connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	defer pool.Close()

	// wire layers
	queries := dbgen.New(pool)
	repo := postgres.NewSubscriptionRepo(queries, log)
	svc := service.NewSubscriptionService(repo, log)
	h := handler.New(svc)
	srv := server.New(cfg.AppPort, h, log, cfg.AppEnv)

	// start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		if err := srv.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		log.Error("server error", slog.Any("error", err))
	case sig := <-quit:
		log.Info("received shutdown signal", slog.String("signal", sig.String()))
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("forced shutdown", slog.Any("error", err))
		os.Exit(1)
	}

	log.Info("server stopped cleanly")
}
