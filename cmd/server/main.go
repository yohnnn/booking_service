package main

import (
	"context"
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	application, err := app.NewApp(ctx, logger, cfg)
	if err != nil {
		logger.Error("failed to create app", "error", err)
		os.Exit(1)
	}
	defer application.Close()

	if err := application.Run(); err != nil {
		logger.Error("app run failed", "error", err)
		os.Exit(1)
	}
}
