package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"steam-observer/internal/modules/auth/domain"
	"steam-observer/internal/modules/auth/ports/out_ports"
)

var ErrNotFound = errors.New("user not found")

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) out_ports.UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	const q = `
		SELECT id, email, google_id, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, q, id)
	return scanUser(row)
}

func (r *UserRepository) FindByGoogleID(ctx context.Context, googleID string) (*domain.User, error) {
	const q = `
		SELECT id, email, google_id, created_at, updated_at
		FROM users
		WHERE google_id = $1
	`
	row := r.pool.QueryRow(ctx, q, googleID)
	return scanUser(row)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	const q = `
		SELECT id, email, google_id, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	row := r.pool.QueryRow(ctx, q, email)
	return scanUser(row)
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	const q = `
		INSERT INTO users (email, google_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	now := time.Now().UTC()
	var id string
	var createdAt, updatedAt time.Time

	err := r.pool.QueryRow(ctx, q, u.Email, u.GoogleID, now, now).
		Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return err
	}

	u.ID = domain.UserID(id)
	u.CreatedAt = createdAt
	u.UpdatedAt = updatedAt
	return nil
}

func (r *UserRepository) Update(ctx context.Context, u *domain.User) error {
	const q = `
		UPDATE users
		SET email = $1,
		    google_id = $2,
		    updated_at = $3
		WHERE id = $4
	`

	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx, q, u.Email, u.GoogleID, now, u.ID)
	if err != nil {
		return err
	}
	u.UpdatedAt = now
	return nil
}

func scanUser(row pgx.Row) (*domain.User, error) {
	var (
		id        string
		email     *string
		googleID  *string
		createdAt time.Time
		updatedAt time.Time
	)

	if err := row.Scan(&id, &email, &googleID, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, out_ports.ErrNotFound
		}
		return nil, err
	}

	return &domain.User{
		ID:        domain.UserID(id),
		Email:     email,
		GoogleID:  googleID,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}
