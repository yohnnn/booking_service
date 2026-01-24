package cache

import (
	"context"

	"github.com/yohnnn/booking_service/internal/models"
)

type ConcertCacheRepository interface {
	Get(ctx context.Context) ([]models.Concert, error)
	Set(ctx context.Context, concerts []models.Concert) error
	Delete(ctx context.Context) error
}
