package auth

import (
	"context"
	"errors"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
)

type CreateUserParams struct {
	Email        string
	Name         string
	PasswordHash string
}

type Repository interface {
	CreateUser(ctx context.Context, p CreateUserParams) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUserByID(ctx context.Context, id int64) (User, error)
}
