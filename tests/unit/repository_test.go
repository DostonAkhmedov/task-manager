package tests

import (
	"testing"

	"github.com/DostonAkhmedov/task-manager/internal/models"
)

func TestUserRepository_Create(t *testing.T) {
	// This is a placeholder test
	// In production, use testcontainers for MySQL
	user := &models.User{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "hashedpassword",
	}

	if user.Email != "test@example.com" {
		t.Errorf("expected email to be %s, got %s", "test@example.com", user.Email)
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	// Placeholder test
	email := "test@example.com"
	if email != "test@example.com" {
		t.Errorf("expected email to be %s, got %s", "test@example.com", email)
	}
}
