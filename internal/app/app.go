package app

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

type App struct {
	logger *slog.Logger
	server *http.Server
	pool   *pgxpool.Pool
	redis  *redis.Client
	kafka  *event.KafkaProducer
}

func NewApp(ctx context.Context, logger *slog.Logger, cfg *config.Config) (*App, error) {

	pool, err := pgxpool.New(ctx, cfg.Postgres.DSN())
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	logger.Info("Connected to PostgreSQL")

	redisClient := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr()})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		pool.Close()
		return nil, err
	}
	logger.Info("Connected to Redis")

	kafkaProducer := event.NewKafkaProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic)

	validate := validator.New()
	txManager := tx.NewManager(pool)
	cache := cache_redis.NewConcertCache(redisClient, 5*time.Minute)

	userRepo := postgres.NewUserRepo(pool)
	concertRepo := postgres.NewConcertRepo(pool)
	bookingRepo := postgres.NewBookingRepo(pool)

	authService := service.NewAuthService(userRepo, cfg.JWT.SecretKey, cfg.JWT.TokenTTL)
	concertService := service.NewConcertService(concertRepo, cache)
	bookingService := service.NewBookingService(bookingRepo, concertRepo, cache, txManager, kafkaProducer)

	authHandler := v1.NewAuthHandler(logger, validate, authService)
	concertHandler := v1.NewConcertHandler(logger, concertService)
	bookingHandler := v1.NewBookingHandler(logger, validate, bookingService)

	router := handler.NewRouter(logger, authService, authHandler, concertHandler, bookingHandler)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router.InitRoutes(),
	}

	return &App{
		logger: logger,
		server: server,
		pool:   pool,
		redis:  redisClient,
		kafka:  kafkaProducer,
	}, nil
}

func (a *App) Run() error {
	go func() {
		a.logger.Info("Server starting", "address", a.server.Addr)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	a.logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		return err
	}

	a.logger.Info("Server exited")
	return nil
}

func (a *App) Close() {
	if a.kafka != nil {
		a.kafka.Close()
	}
	if a.redis != nil {
		a.redis.Close()
	}
	if a.pool != nil {
		a.pool.Close()
	}
}
