package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `db:"id" json:"id"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

type Concert struct {
	ID             uuid.UUID `db:"id" json:"id"`
	Name           string    `db:"name" json:"name"`
	Place          string    `db:"place" json:"place"`
	Date           time.Time `db:"date" json:"date"`
	Price          float64   `db:"price" json:"price"`
	TotalSeats     int       `db:"total_seats" json:"total_seats"`
	AvailableSeats int       `db:"available_seats" json:"available_seats"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "PENDING"
	BookingStatusConfirmed BookingStatus = "CONFIRMED"
	BookingStatusCancelled BookingStatus = "CANCELLED"
)

type Booking struct {
	ID         uuid.UUID     `db:"id" json:"id"`
	UserID     uuid.UUID     `db:"user_id" json:"user_id"`
	ConcertID  uuid.UUID     `db:"concert_id" json:"concert_id"`
	SeatNumber int           `db:"seat_number" json:"seat_number"`
	Status     BookingStatus `db:"status" json:"status"`
	CreatedAt  time.Time     `db:"created_at" json:"created_at"`
}
