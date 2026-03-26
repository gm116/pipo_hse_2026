package tasks

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fadeedan/pipo_hse_2026/internal/db/sqlcgen"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	queries *sqlcgen.Queries
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{queries: sqlcgen.New(pool)}
}

func (r *PostgresRepository) CreateTask(ctx context.Context, userID int64, in CreateTaskInput) (Task, error) {
	t, err := r.queries.CreateTask(ctx, sqlcgen.CreateTaskParams{
		UserID:      userID,
		Title:       in.Title,
		Description: in.Description,
		Status:      in.Status,
	})
	if err != nil {
		return Task{}, fmt.Errorf("create task: %w", err)
	}
	return toTask(t), nil
}

func (r *PostgresRepository) GetTask(ctx context.Context, userID, taskID int64) (Task, error) {
	t, err := r.queries.GetTask(ctx, sqlcgen.GetTaskParams{
		ID:     taskID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Task{}, ErrTaskNotFound
		}
		return Task{}, fmt.Errorf("get task: %w", err)
	}
	return toTask(t), nil
}

func (r *PostgresRepository) UpdateTask(ctx context.Context, userID, taskID int64, in UpdateTaskInput) (Task, error) {
	t, err := r.queries.UpdateTask(ctx, sqlcgen.UpdateTaskParams{
		Title:       in.Title,
		Description: in.Description,
		Status:      in.Status,
		ID:          taskID,
		UserID:      userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Task{}, ErrTaskNotFound
		}
		return Task{}, fmt.Errorf("update task: %w", err)
	}
	return toTask(t), nil
}

func (r *PostgresRepository) DeleteTask(ctx context.Context, userID, taskID int64) error {
	rowsAffected, err := r.queries.DeleteTask(ctx, sqlcgen.DeleteTaskParams{
		ID:     taskID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	if rowsAffected == 0 {
		return ErrTaskNotFound
	}
	return nil
}

func (r *PostgresRepository) ListTasks(ctx context.Context, userID int64) ([]Task, error) {
	items, err := r.queries.ListTasks(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	out := make([]Task, 0, len(items))
	for _, item := range items {
		out = append(out, toTask(item))
	}
	return out, nil
}

func toTask(t sqlcgen.Task) Task {
	return Task{
		ID:          t.ID,
		UserID:      t.UserID,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		CreatedAt:   fromDBTime(t.CreatedAt),
		UpdatedAt:   fromDBTime(t.UpdatedAt),
	}
}

func fromDBTime(ts pgtype.Timestamptz) time.Time {
	if !ts.Valid {
		return time.Time{}
	}
	return ts.Time
}
