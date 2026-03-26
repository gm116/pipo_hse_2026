package password

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
)

const saltSize = 16

func Hash(raw string) (string, error) {
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}
	h := digest(salt, []byte(raw))
	return base64.RawURLEncoding.EncodeToString(salt) + "." + base64.RawURLEncoding.EncodeToString(h), nil
}

func Compare(stored, raw string) bool {
	parts := splitTwo(stored)
	if len(parts) != 2 {
		return false
	}
	salt, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	hash, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	candidate := digest(salt, []byte(raw))
	return subtle.ConstantTimeCompare(hash, candidate) == 1
}

func digest(salt, pass []byte) []byte {
	buf := append(append([]byte{}, salt...), pass...)
	h := sha256.Sum256(buf)
	return h[:]
}

func splitTwo(s string) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return nil
}
