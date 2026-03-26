package tasks

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"
)

type memoryTaskRepo struct {
	nextID int64
	items  map[int64]Task
}

func newMemoryTaskRepo() *memoryTaskRepo {
	return &memoryTaskRepo{nextID: 1, items: map[int64]Task{}}
}

func (r *memoryTaskRepo) CreateTask(_ context.Context, userID int64, in CreateTaskInput) (Task, error) {
	task := Task{
		ID:          r.nextID,
		UserID:      userID,
		Title:       in.Title,
		Description: in.Description,
		Status:      in.Status,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	r.nextID++
	r.items[task.ID] = task
	return task, nil
}

func (r *memoryTaskRepo) GetTask(_ context.Context, userID, taskID int64) (Task, error) {
	task, ok := r.items[taskID]
	if !ok || task.UserID != userID {
		return Task{}, ErrTaskNotFound
	}
	return task, nil
}

func (r *memoryTaskRepo) UpdateTask(_ context.Context, userID, taskID int64, in UpdateTaskInput) (Task, error) {
	task, ok := r.items[taskID]
	if !ok || task.UserID != userID {
		return Task{}, ErrTaskNotFound
	}
	task.Title = in.Title
	task.Description = in.Description
	task.Status = in.Status
	task.UpdatedAt = time.Now()
	r.items[taskID] = task
	return task, nil
}

func (r *memoryTaskRepo) DeleteTask(_ context.Context, userID, taskID int64) error {
	task, ok := r.items[taskID]
	if !ok || task.UserID != userID {
		return ErrTaskNotFound
	}
	delete(r.items, taskID)
	return nil
}

func (r *memoryTaskRepo) ListTasks(_ context.Context, userID int64) ([]Task, error) {
	out := make([]Task, 0)
	for _, task := range r.items {
		if task.UserID == userID {
			out = append(out, task)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out, nil
}

func TestService_CreateUpdateDelete(t *testing.T) {
	repo := newMemoryTaskRepo()
	svc := NewService(repo)

	created, err := svc.CreateTask(context.Background(), 1, CreateTaskInput{Title: "Task #1", Status: "todo"})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("task id is empty")
	}

	updated, err := svc.UpdateTask(context.Background(), 1, created.ID, UpdateTaskInput{
		Title:       "Task #1 Updated",
		Description: "updated",
		Status:      "done",
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Status != StatusDone {
		t.Fatalf("unexpected status: %s", updated.Status)
	}

	if err := svc.DeleteTask(context.Background(), 1, created.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	_, err = svc.GetTask(context.Background(), 1, created.ID)
	if !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("expected not found after delete, got %v", err)
	}
}
