package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/yohnnn/booking_service/internal/app"
	"github.com/yohnnn/booking_service/internal/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if err := run(logger, cfg); err != nil {
		logger.Error("fatal error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger, cfg *config.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	application, err := app.NewApp(ctx, logger, cfg)
	if err != nil {
		return fmt.Errorf("failed to create app: %w", err)
	}
	defer application.Close()

	return application.Run()
}
