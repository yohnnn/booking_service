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

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *models.User) error {
	query := `
        INSERT INTO users (email, password_hash)
        VALUES ($1, $2)
        RETURNING id, created_at
    `
	if err := tx.Executor(ctx, r.db).QueryRow(ctx, query, user.Email, user.PasswordHash).Scan(&user.ID, &user.CreatedAt); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
        SELECT id, email, password_hash, created_at
        FROM users
        WHERE email = $1
    `
	var user models.User
	if err := pgxscan.Get(ctx, tx.Executor(ctx, r.db), &user, query, email); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
        SELECT id, email, password_hash, created_at
        FROM users
        WHERE id = $1
    `
	var user models.User
	if err := pgxscan.Get(ctx, tx.Executor(ctx, r.db), &user, query, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return &user, nil
}
