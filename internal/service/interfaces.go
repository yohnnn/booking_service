package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/yohnnn/booking_service/internal/models"
)

type Auth interface {
	Register(ctx context.Context, email, password string) (*models.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	ParseToken(tokenString string) (uuid.UUID, error)
}

type Concert interface {
	GetAll(ctx context.Context) ([]models.Concert, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Concert, error)
}

type Booking interface {
	CreateBooking(ctx context.Context, userID, concertID uuid.UUID, seat int) (*models.Booking, error)
	GetUserBookings(ctx context.Context, userID uuid.UUID) ([]models.Booking, error)
}
