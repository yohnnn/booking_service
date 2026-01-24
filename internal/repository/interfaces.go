package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/yohnnn/booking_service/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
}

type ConcertRepository interface {
	GetAll(ctx context.Context) ([]models.Concert, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Concert, error)
	DecrementSeats(ctx context.Context, id uuid.UUID) error
}

type BookingRepository interface {
	Create(ctx context.Context, booking *models.Booking) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Booking, error)
}
