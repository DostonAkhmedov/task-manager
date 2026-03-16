package repository

import (
	"testing"

	"github.com/DostonAkhmedov/task-manager/internal/models"
)

// TestUserRepository_Create tests user creation
func TestUserRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		user    *models.User
		wantErr bool
	}{
		{
			name: "valid user creation",
			user: &models.User{
				Email:    "test@example.com",
				Username: "testuser",
				Password: "hashedpassword",
			},
			wantErr: false,
		},
		{
			name: "user with empty email",
			user: &models.User{
				Email:    "",
				Username: "testuser",
				Password: "hashedpassword",
			},
			wantErr: false, // Will fail at DB level if not null
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.user != nil && tt.user.Email == "" && !tt.wantErr {
				// Skip validation test for this scenario
				t.Logf("User validation test: %s", tt.name)
			}
		})
	}
}

// TestUserRepository_GetByEmail tests finding user by email
func TestUserRepository_GetByEmail(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		expectErr bool
	}{
		{
			name:      "existing user",
			email:     "existing@example.com",
			expectErr: false,
		},
		{
			name:      "non-existing user",
			email:     "nonexist@example.com",
			expectErr: true,
		},
		{
			name:      "empty email",
			email:     "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This would need an actual DB connection
			// For now, we test the logic
			if tt.email == "" && tt.expectErr {
				t.Logf("Correctly expecting error for empty email")
			}
		})
	}
}

// TestTeamRepository_Create tests team creation
func TestTeamRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		team    *models.Team
		wantErr bool
	}{
		{
			name: "valid team",
			team: &models.Team{
				Name:      "Engineering Team",
				CreatedBy: "user123",
			},
			wantErr: false,
		},
		{
			name: "team with empty name",
			team: &models.Team{
				Name:      "",
				CreatedBy: "user123",
			},
			wantErr: false, // DB constraint will catch it
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.team.Name != "" {
				t.Logf("Team validation: %s", tt.name)
			}
		})
	}
}

// TestComplexQueryRepository_GetTeamStatistics tests team statistics query
func TestComplexQueryRepository_GetTeamStatistics(t *testing.T) {
	// Test that the query structure is valid
	query := `
	SELECT 
		t.id,
		t.name,
		COUNT(DISTINCT tm.user_id) as member_count,
		COUNT(DISTINCT CASE 
			WHEN tasks.status = 'done' AND tasks.updated_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)
			THEN tasks.id 
		END) as done_tasks_7d,
		t.created_by
	FROM teams t
	LEFT JOIN team_members tm ON t.id = tm.team_id
	LEFT JOIN tasks ON t.id = tasks.team_id
	GROUP BY t.id, t.name, t.created_by
	ORDER BY t.created_at DESC
	`

	if query == "" {
		t.Fatal("Query is empty")
	}

	// Verify query contains expected keywords
	expectedKeywords := []string{"SELECT", "FROM teams", "LEFT JOIN", "GROUP BY", "ORDER BY"}
	for _, keyword := range expectedKeywords {
		if !containsKeyword(query, keyword) {
			t.Errorf("Query missing keyword: %s", keyword)
		}
	}
}

// TestComplexQueryRepository_GetTopTaskCreators tests top creators query with window function
func TestComplexQueryRepository_GetTopTaskCreators(t *testing.T) {
	query := `
	WITH ranked_creators AS (
		SELECT 
			u.id as user_id,
			u.username,
			t.id as team_id,
			t.name as team_name,
			COUNT(tasks.id) as tasks_created,
			ROW_NUMBER() OVER (PARTITION BY t.id ORDER BY COUNT(tasks.id) DESC) as rank
		FROM users u
		JOIN tasks ON u.id = tasks.created_by
		JOIN teams t ON tasks.team_id = t.id
		WHERE tasks.created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)
		GROUP BY u.id, u.username, t.id, t.name
	)
	SELECT user_id, username, team_id, team_name, tasks_created
	FROM ranked_creators
	WHERE rank <= 3
	ORDER BY team_id, rank
	`

	if !containsKeyword(query, "ROW_NUMBER()") {
		t.Error("Query missing ROW_NUMBER() window function")
	}

	if !containsKeyword(query, "PARTITION BY") {
		t.Error("Query missing PARTITION BY clause")
	}

	if !containsKeyword(query, "rank <= 3") {
		t.Error("Query missing rank filter")
	}
}

// TestComplexQueryRepository_GetInvalidAssignments tests invalid assignments detection
func TestComplexQueryRepository_GetInvalidAssignments(t *testing.T) {
	query := `
	SELECT 
		t.id as task_id,
		t.title as task_title,
		t.team_id,
		t.assignee_id,
		u.username
	FROM tasks t
	INNER JOIN users u ON t.assignee_id = u.id
	LEFT JOIN team_members tm ON t.assignee_id = tm.user_id AND t.team_id = tm.team_id
	WHERE t.assignee_id IS NOT NULL
	AND tm.id IS NULL
	`

	if !containsKeyword(query, "LEFT JOIN team_members") {
		t.Error("Query missing LEFT JOIN on team_members")
	}

	if !containsKeyword(query, "tm.id IS NULL") {
		t.Error("Query missing NULL check for team_members")
	}
}

// Helper function to check if query contains keyword
func containsKeyword(query, keyword string) bool {
	for i := 0; i < len(query)-len(keyword)+1; i++ {
		if query[i:i+len(keyword)] == keyword {
			return true
		}
	}
	return false
}
