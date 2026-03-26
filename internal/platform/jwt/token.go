package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var ErrInvalidToken = errors.New("invalid token")

type Manager struct {
	secret []byte
	issuer string
	ttl    time.Duration
}

type Claims struct {
	Subject int64
	Expiry  time.Time
	Issued  time.Time
	Issuer  string
}

func NewManager(secret, issuer string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), issuer: issuer, ttl: ttl}
}

func (m *Manager) Issue(subject int64) (string, error) {
	now := time.Now().UTC()
	payload := map[string]any{
		"sub": strconv.FormatInt(subject, 10),
		"iss": m.issuer,
		"iat": now.Unix(),
		"exp": now.Add(m.ttl).Unix(),
	}
	headerRaw, err := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		return "", fmt.Errorf("marshal header: %w", err)
	}
	payloadRaw, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	h := base64.RawURLEncoding.EncodeToString(headerRaw)
	p := base64.RawURLEncoding.EncodeToString(payloadRaw)
	signingInput := h + "." + p
	sig := sign(signingInput, m.secret)
	return signingInput + "." + sig, nil
}

func (m *Manager) Parse(token string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, ErrInvalidToken
	}
	signingInput := parts[0] + "." + parts[1]
	expected := sign(signingInput, m.secret)
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return Claims{}, ErrInvalidToken
	}

	payloadRaw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Claims{}, ErrInvalidToken
	}
	var payload struct {
		Sub string `json:"sub"`
		Iss string `json:"iss"`
		Iat int64  `json:"iat"`
		Exp int64  `json:"exp"`
	}
	if err := json.Unmarshal(payloadRaw, &payload); err != nil {
		return Claims{}, ErrInvalidToken
	}
	id, err := strconv.ParseInt(payload.Sub, 10, 64)
	if err != nil {
		return Claims{}, ErrInvalidToken
	}
	exp := time.Unix(payload.Exp, 0)
	if time.Now().After(exp) {
		return Claims{}, ErrInvalidToken
	}

	return Claims{
		Subject: id,
		Issuer:  payload.Iss,
		Issued:  time.Unix(payload.Iat, 0),
		Expiry:  exp,
	}, nil
}

func sign(payload string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
