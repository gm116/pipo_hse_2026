package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/fadeedan/pipo_hse_2026/internal/platform/jwt"
	"github.com/fadeedan/pipo_hse_2026/internal/platform/password"
)

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Service struct {
	repo   Repository
	tokens *jwt.Manager
}

func NewService(repo Repository, tokens *jwt.Manager) *Service {
	return &Service{repo: repo, tokens: tokens}
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (AuthResponse, error) {
	if err := validateRegister(in); err != nil {
		return AuthResponse{}, err
	}

	hash, err := password.Hash(in.Password)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("hash password: %w", err)
	}

	u, err := s.repo.CreateUser(ctx, CreateUserParams{
		Email:        strings.TrimSpace(strings.ToLower(in.Email)),
		Name:         strings.TrimSpace(in.Name),
		PasswordHash: hash,
	})
	if err != nil {
		return AuthResponse{}, err
	}

	token, err := s.tokens.Issue(u.ID)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("issue token: %w", err)
	}

	u.PasswordHash = ""
	return AuthResponse{Token: token, User: u}, nil
}

func (s *Service) Login(ctx context.Context, in LoginInput) (AuthResponse, error) {
	if strings.TrimSpace(in.Email) == "" || strings.TrimSpace(in.Password) == "" {
		return AuthResponse{}, ErrInvalidInput
	}

	u, err := s.repo.GetUserByEmail(ctx, strings.TrimSpace(strings.ToLower(in.Email)))
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return AuthResponse{}, ErrInvalidCredentials
		}
		return AuthResponse{}, err
	}

	if !password.Compare(u.PasswordHash, in.Password) {
		return AuthResponse{}, ErrInvalidCredentials
	}

	token, err := s.tokens.Issue(u.ID)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("issue token: %w", err)
	}

	u.PasswordHash = ""
	return AuthResponse{Token: token, User: u}, nil
}

func (s *Service) Me(ctx context.Context, userID int64) (User, error) {
	u, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return User{}, err
	}
	u.PasswordHash = ""
	return u, nil
}

func (s *Service) ParseToken(raw string) (int64, error) {
	claims, err := s.tokens.Parse(raw)
	if err != nil {
		return 0, err
	}
	return claims.Subject, nil
}

func validateRegister(in RegisterInput) error {
	email := strings.TrimSpace(in.Email)
	name := strings.TrimSpace(in.Name)
	if !strings.Contains(email, "@") || len(email) < 5 {
		return ErrInvalidInput
	}
	if len(name) < 2 {
		return ErrInvalidInput
	}
	if len(in.Password) < 6 {
		return ErrInvalidInput
	}
	return nil
}
