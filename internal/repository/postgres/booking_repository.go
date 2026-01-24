package postgres

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yohnnn/booking_service/internal/models"
	"github.com/yohnnn/booking_service/internal/repository/tx"
)

type BookingRepo struct {
	db *pgxpool.Pool
}

func NewBookingRepo(db *pgxpool.Pool) *BookingRepo {
	return &BookingRepo{db: db}
}

func (r *BookingRepo) Create(ctx context.Context, booking *models.Booking) error {
	query := `
		INSERT INTO bookings (user_id, concert_id, seat_number, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	err := tx.Executor(ctx, r.db).QueryRow(ctx, query,
		booking.UserID,
		booking.ConcertID,
		booking.SeatNumber,
		booking.Status,
	).Scan(&booking.ID, &booking.CreatedAt)

	if err != nil {
		if IsUnique(err) {
			return models.ErrAlreadyExists
		}
		return fmt.Errorf("failed to create booking: %w", err)
	}

	return nil
}

func (r *BookingRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Booking, error) {
	query := `
		SELECT id, user_id, concert_id, seat_number, status, created_at
		FROM bookings
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	var bookings []models.Booking
	if err := pgxscan.Select(ctx, tx.Executor(ctx, r.db), &bookings, query, userID); err != nil {
		return nil, fmt.Errorf("failed to get user bookings: %w", err)
	}
	return bookings, nil
}
