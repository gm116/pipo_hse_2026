package password

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const cost = bcrypt.DefaultCost

func Hash(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("password is empty")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(raw), cost)
	if err != nil {
		return "", fmt.Errorf("bcrypt hash: %w", err)
	}
	return string(hash), nil
}

func Compare(stored, raw string) bool {
	if strings.TrimSpace(stored) == "" || strings.TrimSpace(raw) == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(stored), []byte(raw)) == nil
}
