package db

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrationSchemaContract(t *testing.T) {
	path := filepath.Join("migrations", "001_init.sql")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := strings.ToLower(string(raw))

	requireContainsAll(t, sql, []string{
		"create table if not exists users",
		"id bigserial",
		"email text",
		"name text",
		"password_hash text",
		"created_at timestamptz",
	})

	requireContainsAll(t, sql, []string{
		"create table if not exists tasks",
		"id bigserial",
		"user_id bigint",
		"title text",
		"description text",
		"status text",
		"created_at timestamptz",
		"updated_at timestamptz",
	})
}

func requireContainsAll(t *testing.T, body string, needles []string) {
	t.Helper()
	for _, needle := range needles {
		if !strings.Contains(body, needle) {
			t.Fatalf("schema contract mismatch: missing %q", needle)
		}
	}
}
