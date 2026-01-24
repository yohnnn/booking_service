package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yohnnn/booking_service/internal/models"
	"github.com/yohnnn/booking_service/internal/repository/tx"
)

type ConcertRepo struct {
	db *pgxpool.Pool
}

func NewConcertRepo(db *pgxpool.Pool) *ConcertRepo {
	return &ConcertRepo{db: db}
}

func (r *ConcertRepo) GetAll(ctx context.Context) ([]models.Concert, error) {
	query := `
		SELECT id, name, place, date, price, total_seats, available_seats, created_at
		FROM concerts
		ORDER BY date ASC
	`
	var concerts []models.Concert

	if err := pgxscan.Select(ctx, tx.Executor(ctx, r.db), &concerts, query); err != nil {
		return nil, fmt.Errorf("failed to get all concerts: %w", err)
	}

	return concerts, nil
}

func (r *ConcertRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Concert, error) {
	query := `
		SELECT id, name, place, date, price, total_seats, available_seats, created_at
		FROM concerts
		WHERE id = $1
	`
	var concert models.Concert

	if err := pgxscan.Get(ctx, tx.Executor(ctx, r.db), &concert, query, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get concert by id: %w", err)
	}

	return &concert, nil
}

func (r *ConcertRepo) DecrementSeats(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE concerts 
		SET available_seats = available_seats - 1 
		WHERE id = $1 AND available_seats > 0
	`
	res, err := tx.Executor(ctx, r.db).Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to decrement seats: %w", err)
	}
	if res.RowsAffected() == 0 {
		return models.ErrNoSeats
	}
	return nil
}
