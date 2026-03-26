package tasks

import (
	"context"
	"errors"
	"strings"
)

var ErrInvalidInput = errors.New("invalid input")

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateTask(ctx context.Context, userID int64, in CreateTaskInput) (Task, error) {
	in = normalizeCreateInput(in)
	if err := validateCreateInput(in); err != nil {
		return Task{}, err
	}
	return s.repo.CreateTask(ctx, userID, in)
}

func (s *Service) GetTask(ctx context.Context, userID, taskID int64) (Task, error) {
	if taskID <= 0 {
		return Task{}, ErrInvalidInput
	}
	return s.repo.GetTask(ctx, userID, taskID)
}

func (s *Service) UpdateTask(ctx context.Context, userID, taskID int64, in UpdateTaskInput) (Task, error) {
	in = normalizeUpdateInput(in)
	if taskID <= 0 {
		return Task{}, ErrInvalidInput
	}
	if err := validateUpdateInput(in); err != nil {
		return Task{}, err
	}
	return s.repo.UpdateTask(ctx, userID, taskID, in)
}

func (s *Service) DeleteTask(ctx context.Context, userID, taskID int64) error {
	if taskID <= 0 {
		return ErrInvalidInput
	}
	return s.repo.DeleteTask(ctx, userID, taskID)
}

func (s *Service) ListTasks(ctx context.Context, userID int64) ([]Task, error) {
	return s.repo.ListTasks(ctx, userID)
}

func normalizeCreateInput(in CreateTaskInput) CreateTaskInput {
	in.Title = strings.TrimSpace(in.Title)
	in.Description = strings.TrimSpace(in.Description)
	in.Status = strings.TrimSpace(strings.ToLower(in.Status))
	if in.Status == "" {
		in.Status = StatusTodo
	}
	return in
}

func normalizeUpdateInput(in UpdateTaskInput) UpdateTaskInput {
	in.Title = strings.TrimSpace(in.Title)
	in.Description = strings.TrimSpace(in.Description)
	in.Status = strings.TrimSpace(strings.ToLower(in.Status))
	if in.Status == "" {
		in.Status = StatusTodo
	}
	return in
}

func validateCreateInput(in CreateTaskInput) error {
	if len(in.Title) < 2 {
		return ErrInvalidInput
	}
	if !isValidStatus(in.Status) {
		return ErrInvalidInput
	}
	return nil
}

func validateUpdateInput(in UpdateTaskInput) error {
	if len(in.Title) < 2 {
		return ErrInvalidInput
	}
	if !isValidStatus(in.Status) {
		return ErrInvalidInput
	}
	return nil
}

func isValidStatus(status string) bool {
	switch status {
	case StatusTodo, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}
