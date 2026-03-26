package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fadeedan/pipo_hse_2026/internal/platform/jwt"
)

type memoryRepo struct {
	nextID  int64
	byEmail map[string]User
	byID    map[int64]User
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{
		nextID:  1,
		byEmail: map[string]User{},
		byID:    map[int64]User{},
	}
}

func (r *memoryRepo) CreateUser(_ context.Context, p CreateUserParams) (User, error) {
	if _, ok := r.byEmail[p.Email]; ok {
		return User{}, ErrUserExists
	}
	u := User{ID: r.nextID, Email: p.Email, Name: p.Name, PasswordHash: p.PasswordHash, CreatedAt: time.Now()}
	r.nextID++
	r.byEmail[p.Email] = u
	r.byID[u.ID] = u
	return u, nil
}

func (r *memoryRepo) GetUserByEmail(_ context.Context, email string) (User, error) {
	u, ok := r.byEmail[email]
	if !ok {
		return User{}, ErrUserNotFound
	}
	return u, nil
}

func (r *memoryRepo) GetUserByID(_ context.Context, id int64) (User, error) {
	u, ok := r.byID[id]
	if !ok {
		return User{}, ErrUserNotFound
	}
	return u, nil
}

func TestService_RegisterAndLogin(t *testing.T) {
	repo := newMemoryRepo()
	svc := NewService(repo, jwt.NewManager("secret", "test", time.Hour))

	reg, err := svc.Register(context.Background(), RegisterInput{
		Email:    "user@example.com",
		Name:     "Test User",
		Password: "123456",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if reg.Token == "" {
		t.Fatal("empty token")
	}
	if reg.User.ID == 0 {
		t.Fatal("empty user id")
	}

	login, err := svc.Login(context.Background(), LoginInput{Email: "user@example.com", Password: "123456"})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if login.Token == "" {
		t.Fatal("empty login token")
	}
}

func TestService_LoginInvalidPassword(t *testing.T) {
	repo := newMemoryRepo()
	svc := NewService(repo, jwt.NewManager("secret", "test", time.Hour))

	_, err := svc.Register(context.Background(), RegisterInput{
		Email:    "user@example.com",
		Name:     "Test User",
		Password: "123456",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	_, err = svc.Login(context.Background(), LoginInput{Email: "user@example.com", Password: "badpass"})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
