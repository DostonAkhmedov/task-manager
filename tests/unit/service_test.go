package tests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/DostonAkhmedov/task-manager/internal/models"
)

// TestModelTeamCreation tests the Team model creation
func TestModelTeamCreation(t *testing.T) {
	team := &models.Team{
		ID:        uuid.New().String(),
		Name:      "Test Team",
		CreatedBy: "user123",
	}

	if team.Name != "Test Team" {
		t.Errorf("expected team name 'Test Team', got '%s'", team.Name)
	}

	if team.CreatedBy != "user123" {
		t.Errorf("expected creator 'user123', got '%s'", team.CreatedBy)
	}
}

// TestModelUserCreation tests the User model creation
func TestModelUserCreation(t *testing.T) {
	user := &models.User{
		ID:       uuid.New().String(),
		Email:    "test@example.com",
		Username: "testuser",
		Password: "hashedpassword",
	}

	if user.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", user.Email)
	}

	if user.Username != "testuser" {
		t.Errorf("expected username 'testuser', got '%s'", user.Username)
	}
}

// TestModelTeamMemberCreation tests the TeamMember model creation
func TestModelTeamMemberCreation(t *testing.T) {
	member := &models.TeamMember{
		ID:     uuid.New().String(),
		UserID: "user123",
		TeamID: "team456",
		Role:   "owner",
	}

	if member.Role != "owner" {
		t.Errorf("expected role 'owner', got '%s'", member.Role)
	}

	if member.UserID != "user123" {
		t.Errorf("expected user ID 'user123', got '%s'", member.UserID)
	}
}
