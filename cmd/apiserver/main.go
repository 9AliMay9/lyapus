package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/9AliMay9/lyapus/internal/platform/config"
	"github.com/9AliMay9/lyapus/internal/platform/database"
	transporthttp "github.com/9AliMay9/lyapus/internal/platform/transport/http"
)

const (
	databaseStartupTimeout = 5 * time.Second
	shutdownTimeout        = 10 * time.Second
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := run(logger); err != nil {
		logger.Error("apiserver exited", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	databaseCtx, cancelDatabase := context.WithTimeout(context.Background(), databaseStartupTimeout)
	pool, err := database.Open(databaseCtx, cfg.DatabaseURL)
	cancelDatabase()
	if err != nil {
		return fmt.Errorf("connect to PostgreSQL: %w", err)
	}
	defer pool.Close()

	server := transporthttp.NewServer(cfg, logger, pool)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("http_server_started", "http_addr", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- fmt.Errorf("serve HTTP: %w", err)
			return
		}
		serverErr <- nil
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		logger.Info("shutdown_signal_received")
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}
	if err := <-serverErr; err != nil {
		return err
	}

	logger.Info("http_server_stopped")
	return nil
}
