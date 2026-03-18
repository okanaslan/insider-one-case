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

	"insider-one-case/internal/config"
	"insider-one-case/internal/db"
	"insider-one-case/internal/http/handler"
	"insider-one-case/internal/http/router"
	"insider-one-case/internal/idempotency"
	"insider-one-case/internal/repository"
	"insider-one-case/internal/service"
	appvalidator "insider-one-case/internal/validator"
	"insider-one-case/internal/worker"
	"insider-one-case/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	log := logger.New(cfg.LogLevel, cfg.AppEnv).With(
		"app", cfg.AppName,
		"env", cfg.AppEnv,
	)

	startupCtx, startupCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer startupCancel()

	clickHouseConn, err := db.NewClickHouseConn(startupCtx, cfg)
	if err != nil {
		if cfg.AllowStartWithoutInfra {
			log.Warn("clickhouse unavailable; continuing due to ALLOW_START_WITHOUT_INFRA", "error", err)
			clickHouseConn = nil
		} else {
			log.Error("failed to connect clickhouse", "error", err)
			os.Exit(1)
		}
	}

	redisClient, err := db.NewRedisClient(startupCtx, cfg)
	if err != nil {
		if cfg.AllowStartWithoutInfra {
			log.Warn("redis unavailable; continuing due to ALLOW_START_WITHOUT_INFRA", "error", err)
			redisClient = nil
		} else {
			log.Error("failed to connect redis", "error", err)
			os.Exit(1)
		}
	}

	eventValidator := appvalidator.NewEventValidator()
	idempotencyStore := idempotency.NewRedisStore(redisClient, log)

	eventRepo := repository.NewEventRepository(clickHouseConn, log)
	metricsRepo := repository.NewMetricsRepository(clickHouseConn, log)
	_ = eventRepo // reserved for worker batch writes

	ingestWorker := worker.NewIngestWorker(cfg, log)
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	go ingestWorker.Start(workerCtx)

	eventService := service.NewEventService(ingestWorker, idempotencyStore, log)
	metricsService := service.NewMetricsService(metricsRepo, log)

	healthHandler := handler.NewHealthHandler(cfg)
	eventHandler := handler.NewEventHandler(eventService, eventValidator)
	metricsHandler := handler.NewMetricsHandler(metricsService)

	engine := router.Build(cfg, log, healthHandler, eventHandler, metricsHandler)
	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErrCh := make(chan error, 1)
	go func() {
		log.Info("http server starting", "addr", server.Addr)
		if serveErr := server.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			serverErrCh <- serveErr
		}
	}()

	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-signalCtx.Done():
		log.Info("shutdown signal received")
	case serveErr := <-serverErrCh:
		log.Error("server terminated unexpectedly", "error", serveErr)
	}

	workerCancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	log.Info("server shutdown complete")
}
