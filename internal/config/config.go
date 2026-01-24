package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Env      string `env:"APP_ENV" envDefault:"local"`
	Port     int    `env:"HTTP_PORT" envDefault:"8080"`
	Postgres PostgresConfig
	Redis    RedisConfig
	Kafka    KafkaConfig
	JWT      JWTConfig
}

type JWTConfig struct {
	SecretKey string        `env:"JWT_SECRET"`
	TokenTTL  time.Duration `env:"JWT_TTL" envDefault:"24h"`
}

type PostgresConfig struct {
	Host     string `env:"DB_HOST" envDefault:"localhost"`
	Port     string `env:"DB_PORT" envDefault:"5432"`
	User     string `env:"DB_USER" envDefault:"postgres"`
	Password string `env:"DB_PASSWORD" envDefault:"password"`
	DBName   string `env:"DB_NAME" envDefault:"booking_service"`
	SSLMode  string `env:"DB_SSLMODE" envDefault:"disable"`
}

func (p *PostgresConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.DBName, p.SSLMode,
	)
}

type RedisConfig struct {
	Host string `env:"REDIS_HOST" envDefault:"localhost"`
	Port string `env:"REDIS_PORT" envDefault:"6379"`
}

func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

type KafkaConfig struct {
	Brokers []string `env:"KAFKA_BROKERS" envDefault:"localhost:29092"`
	Topic   string   `env:"KAFKA_TOPIC" envDefault:"bookings.created"`
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	if cfg.JWT.SecretKey == "" {
		return nil, errors.New("JWT_SECRET is required")
	}

	return cfg, nil
}
