package tasks

import (
	"context"
	"errors"
)

var (
	ErrTaskNotFound = errors.New("task not found")
)

type Repository interface {
	CreateTask(ctx context.Context, userID int64, in CreateTaskInput) (Task, error)
	GetTask(ctx context.Context, userID, taskID int64) (Task, error)
	UpdateTask(ctx context.Context, userID, taskID int64, in UpdateTaskInput) (Task, error)
	DeleteTask(ctx context.Context, userID, taskID int64) error
	ListTasks(ctx context.Context, userID int64) ([]Task, error)
}
