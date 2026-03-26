package tasks

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) CreateTask(ctx context.Context, userID int64, in CreateTaskInput) (Task, error) {
	const q = `
		INSERT INTO tasks(user_id, title, description, status)
		VALUES($1, $2, $3, $4)
		RETURNING id, user_id, title, description, status, created_at, updated_at
	`
	var t Task
	err := r.pool.QueryRow(ctx, q, userID, in.Title, in.Description, in.Status).Scan(
		&t.ID,
		&t.UserID,
		&t.Title,
		&t.Description,
		&t.Status,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		return Task{}, fmt.Errorf("create task: %w", err)
	}
	return t, nil
}

func (r *PostgresRepository) GetTask(ctx context.Context, userID, taskID int64) (Task, error) {
	const q = `
		SELECT id, user_id, title, description, status, created_at, updated_at
		FROM tasks
		WHERE id=$1 AND user_id=$2
	`
	var t Task
	err := r.pool.QueryRow(ctx, q, taskID, userID).Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Task{}, ErrTaskNotFound
		}
		return Task{}, fmt.Errorf("get task: %w", err)
	}
	return t, nil
}

func (r *PostgresRepository) UpdateTask(ctx context.Context, userID, taskID int64, in UpdateTaskInput) (Task, error) {
	const q = `
		UPDATE tasks
		SET title=$1, description=$2, status=$3, updated_at=NOW()
		WHERE id=$4 AND user_id=$5
		RETURNING id, user_id, title, description, status, created_at, updated_at
	`
	var t Task
	err := r.pool.QueryRow(ctx, q, in.Title, in.Description, in.Status, taskID, userID).Scan(
		&t.ID,
		&t.UserID,
		&t.Title,
		&t.Description,
		&t.Status,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Task{}, ErrTaskNotFound
		}
		return Task{}, fmt.Errorf("update task: %w", err)
	}
	return t, nil
}

func (r *PostgresRepository) DeleteTask(ctx context.Context, userID, taskID int64) error {
	const q = `DELETE FROM tasks WHERE id=$1 AND user_id=$2`
	res, err := r.pool.Exec(ctx, q, taskID, userID)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrTaskNotFound
	}
	return nil
}

func (r *PostgresRepository) ListTasks(ctx context.Context, userID int64) ([]Task, error) {
	const q = `
		SELECT id, user_id, title, description, status, created_at, updated_at
		FROM tasks
		WHERE user_id=$1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	out := make([]Task, 0)
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}
	return out, nil
}
