package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fadeedan/pipo_hse_2026/internal/db/sqlcgen"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	queries *sqlcgen.Queries
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{queries: sqlcgen.New(pool)}
}

func (r *PostgresRepository) CreateUser(ctx context.Context, p CreateUserParams) (User, error) {
	u, err := r.queries.CreateUser(ctx, sqlcgen.CreateUserParams{
		Email:        p.Email,
		Name:         p.Name,
		PasswordHash: p.PasswordHash,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return User{}, ErrUserExists
		}
		return User{}, fmt.Errorf("create user: %w", err)
	}
	return toUser(u), nil
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (User, error) {
	u, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("get user by email: %w", err)
	}
	return toUser(u), nil
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id int64) (User, error) {
	u, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("get user by id: %w", err)
	}
	return toUser(u), nil
}

func toUser(u sqlcgen.User) User {
	return User{
		ID:           u.ID,
		Email:        u.Email,
		Name:         u.Name,
		PasswordHash: u.PasswordHash,
		CreatedAt:    fromDBTime(u.CreatedAt),
	}
}

func fromDBTime(ts pgtype.Timestamptz) time.Time {
	if !ts.Valid {
		return time.Time{}
	}
	return ts.Time
}
