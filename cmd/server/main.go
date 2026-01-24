package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	cache_redis "github.com/yohnnn/booking_service/internal/cache/redis"
	"github.com/yohnnn/booking_service/internal/config"
	"github.com/yohnnn/booking_service/internal/event"
	"github.com/yohnnn/booking_service/internal/handler"
	v1 "github.com/yohnnn/booking_service/internal/handler/v1"
	"github.com/yohnnn/booking_service/internal/repository/postgres"
	"github.com/yohnnn/booking_service/internal/repository/tx"
	"github.com/yohnnn/booking_service/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Info("Starting Concert Booking Service...", "port", cfg.Port)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.Postgres.DSN())
	if err != nil {
		logger.Error("Unable to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Error("Database ping failed", "error", err)
		os.Exit(1)
	}
	logger.Info("Connected to PostgreSQL")

	validate := validator.New()

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Addr(),
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Error("Failed to connect to redis", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()
	logger.Info("Connected to Redis")

	txManager := tx.NewManager(pool)
	userRepo := postgres.NewUserRepo(pool)
	concertRepo := postgres.NewConcertRepo(pool)
	bookingRepo := postgres.NewBookingRepo(pool)
	cacheRepo := cache_redis.NewConcertCache(redisClient, 5*time.Minute)

	kafkaProducer := event.NewKafkaProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic)
	defer kafkaProducer.Close()

	authService := service.NewAuthService(userRepo, cfg.JWT.SecretKey, cfg.JWT.TokenTTL)
	concertService := service.NewConcertService(concertRepo, cacheRepo)
	bookingService := service.NewBookingService(bookingRepo, concertRepo, cacheRepo, txManager, kafkaProducer)

	authHandler := v1.NewAuthHandler(logger, validate, authService)
	concertHandler := v1.NewConcertHandler(logger, concertService)
	bookingHandler := v1.NewBookingHandler(logger, validate, bookingService)

	router := handler.NewRouter(logger, authService, authHandler, concertHandler, bookingHandler)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router.InitRoutes(),
	}

	go func() {
		logger.Info("Server starting", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("Server exiting")
}
