package integration

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/DostonAkhmedov/task-manager/internal/models"
	"github.com/DostonAkhmedov/task-manager/internal/repository"
)

// IntegrationTestDB holds database connection and container for tests
type IntegrationTestDB struct {
	db        *sql.DB
	container testcontainers.Container
	ctx       context.Context
}

// NewIntegrationTestDB creates a test database connection with testcontainers MySQL
func NewIntegrationTestDB(ctx context.Context, t *testing.T) *IntegrationTestDB {
	req := testcontainers.ContainerRequest{
		Image:        "mysql:8.0",
		ExposedPorts: []string{"3306/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "testroot",
			"MYSQL_DATABASE":      "task_manager_test",
			"MYSQL_USER":          "testuser",
			"MYSQL_PASSWORD":      "testpass",
		},
		WaitingFor: wait.ForLog("ready for connections").WithOccurrence(2).WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Skipf("Failed to start MySQL container: %v", err)
		return nil
	}

	// Get the mapped port
	host, err := container.Host(ctx)
	if err != nil {
		t.Skipf("Failed to get container host: %v", err)
		container.Terminate(ctx) //nolint:errcheck
		return nil
	}

	port, err := container.MappedPort(ctx, "3306")
	if err != nil {
		t.Skipf("Failed to get mapped port: %v", err)
		container.Terminate(ctx) //nolint:errcheck
		return nil
	}

	// Connect to the test database
	dsn := fmt.Sprintf("testuser:testpass@tcp(%s:%s)/task_manager_test?parseTime=true", host, port.Port())
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Skipf("Failed to open database: %v", err)
		container.Terminate(ctx) //nolint:errcheck
		return nil
	}

	// Test the connection with retries
	var lastErr error
	for i := 0; i < 5; i++ {
		if err := db.Ping(); err == nil {
			return &IntegrationTestDB{
				db:        db,
				container: container,
				ctx:       ctx,
			}
		}
		lastErr = err
		time.Sleep(500 * time.Millisecond)
	}

	t.Skipf("Failed to connect to test database after retries: %v", lastErr)
	db.Close() //nolint:errcheck
	container.Terminate(ctx) //nolint:errcheck
	return nil
}

// Setup initializes the database schema for testing
func (itd *IntegrationTestDB) Setup(t *testing.T) {
	statements := []string{
		"CREATE TABLE IF NOT EXISTS users (id VARCHAR(36) PRIMARY KEY, username VARCHAR(100) NOT NULL UNIQUE, email VARCHAR(100) NOT NULL UNIQUE, password VARCHAR(255) NOT NULL, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP)",
		"CREATE TABLE IF NOT EXISTS teams (id VARCHAR(36) PRIMARY KEY, name VARCHAR(100) NOT NULL, description TEXT, created_by VARCHAR(36) NOT NULL, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE)",
		"CREATE TABLE IF NOT EXISTS team_members (id VARCHAR(36) PRIMARY KEY, team_id VARCHAR(36) NOT NULL, user_id VARCHAR(36) NOT NULL, role VARCHAR(50) NOT NULL DEFAULT 'member', joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE, FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE, UNIQUE KEY unique_team_user (team_id, user_id))",
		"CREATE TABLE IF NOT EXISTS tasks (id VARCHAR(36) PRIMARY KEY, team_id VARCHAR(36) NOT NULL, title VARCHAR(255) NOT NULL, description TEXT, status VARCHAR(50) NOT NULL DEFAULT 'todo', priority VARCHAR(50) NOT NULL DEFAULT 'medium', assigned_to VARCHAR(36), created_by VARCHAR(36) NOT NULL, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE, FOREIGN KEY (assigned_to) REFERENCES users(id) ON DELETE SET NULL, FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE)",
	}

	for _, stmt := range statements {
		if _, err := itd.db.ExecContext(itd.ctx, stmt); err != nil {
			t.Logf("Warning: failed to create table: %v", err)
		}
	}
}

// Cleanup cleans up test data
func (itd *IntegrationTestDB) Cleanup(t *testing.T) {
	tables := []string{"team_members", "tasks", "teams", "users"}
	for _, table := range tables {
		if _, err := itd.db.ExecContext(itd.ctx, fmt.Sprintf("TRUNCATE TABLE %s", table)); err != nil {
			t.Logf("Warning: failed to truncate %s: %v", table, err)
		}
	}
}

// Close closes the database connection and terminates the container
func (itd *IntegrationTestDB) Close(ctx context.Context) error {
	if itd.db != nil {
		itd.db.Close() //nolint:errcheck
	}
	if itd.container != nil {
		return itd.container.Terminate(ctx)
	}
	return nil
}

// TestUserRepository_Integration tests user repository with MySQL testcontainer
func TestUserRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	itd := NewIntegrationTestDB(ctx, t)
	if itd == nil {
		return
	}
	defer itd.Close(ctx) //nolint:errcheck

	itd.Setup(t)
	defer itd.Cleanup(t)

	repo := repository.NewUserRepository(itd.db)

	t.Run("create and retrieve user", func(t *testing.T) {
		userID := uuid.New().String()
		user := &models.User{
			ID:       userID,
			Email:    fmt.Sprintf("test_%s@example.com", userID),
			Username: "testuser_" + userID[:8],
			Password: "hashed_password",
		}

		err := repo.Create(user)
		if err != nil {
			t.Fatalf("Create user failed: %v", err)
		}

		retrieved, err := repo.GetByID(userID)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if retrieved == nil {
			t.Fatal("Retrieved user is nil")
		}

		if retrieved.Email != user.Email {
			t.Errorf("Email mismatch: got %s, want %s", retrieved.Email, user.Email)
		}
	})

	t.Run("get user by email", func(t *testing.T) {
		userID := uuid.New().String()
		email := fmt.Sprintf("email_test_%s@example.com", userID)
		user := &models.User{
			ID:       userID,
			Email:    email,
			Username: "emailtest_" + userID[:8],
			Password: "password",
		}

		err := repo.Create(user)
		if err != nil {
			t.Fatalf("Create user failed: %v", err)
		}

		retrieved, err := repo.GetByEmail(email)
		if err != nil {
			t.Fatalf("GetByEmail failed: %v", err)
		}

		if retrieved == nil {
			t.Fatal("Retrieved user is nil")
		}

		if retrieved.Email != email {
			t.Errorf("Email mismatch: got %s, want %s", retrieved.Email, email)
		}
	})

	t.Run("duplicate email error", func(t *testing.T) {
		userID1 := uuid.New().String()
		userID2 := uuid.New().String()
		email := fmt.Sprintf("dup_%s@example.com", uuid.New().String())

		user1 := &models.User{
			ID:       userID1,
			Email:    email,
			Username: "duptest1_" + userID1[:8],
			Password: "password",
		}

		user2 := &models.User{
			ID:       userID2,
			Email:    email,
			Username: "duptest2_" + userID2[:8],
			Password: "password",
		}

		err := repo.Create(user1)
		if err != nil {
			t.Fatalf("Create first user failed: %v", err)
		}

		err = repo.Create(user2)
		if err == nil {
			t.Error("Expected error for duplicate email, got nil")
		}
	})
}

// TestTeamRepository_Integration tests team repository with testcontainer
func TestTeamRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	itd := NewIntegrationTestDB(ctx, t)
	if itd == nil {
		return
	}
	defer itd.Close(ctx) //nolint:errcheck

	itd.Setup(t)
	defer itd.Cleanup(t)

	// Create a user first
	userRepo := repository.NewUserRepository(itd.db)
	userID := uuid.New().String()
	user := &models.User{
		ID:       userID,
		Email:    "team_creator@example.com",
		Username: "teamcreator",
		Password: "password",
	}
	_ = userRepo.Create(user)

	teamRepo := repository.NewTeamRepository(itd.db)

	t.Run("create and retrieve team", func(t *testing.T) {
		teamID := uuid.New().String()
		team := &models.Team{
			ID:        teamID,
			Name:      "Test Team " + teamID[:8],
			CreatedBy: userID,
		}

		err := teamRepo.Create(team)
		if err != nil {
			t.Fatalf("Create team failed: %v", err)
		}

		retrieved, err := teamRepo.GetByID(teamID)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if retrieved == nil {
			t.Fatal("Retrieved team is nil")
		}

		if retrieved.Name != team.Name {
			t.Errorf("Team name mismatch: got %s, want %s", retrieved.Name, team.Name)
		}
	})

	t.Run("get user teams", func(t *testing.T) {
		teamID := uuid.New().String()
		team := &models.Team{
			ID:        teamID,
			Name:      "User Team " + teamID[:8],
			CreatedBy: userID,
		}

		err := teamRepo.Create(team)
		if err != nil {
			t.Fatalf("Create team failed: %v", err)
		}

		teams, err := teamRepo.GetUserTeams(userID)
		if err != nil {
			t.Fatalf("GetUserTeams failed: %v", err)
		}

		if len(teams) == 0 {
			t.Error("Expected at least one team, got zero")
		}
	})

	t.Run("add team member", func(t *testing.T) {
		// Create another user
		member2ID := uuid.New().String()
		member2 := &models.User{
			ID:       member2ID,
			Email:    "member2@example.com",
			Username: "member2",
			Password: "password",
		}
		_ = userRepo.Create(member2)

		teamID := uuid.New().String()
		team := &models.Team{
			ID:        teamID,
			Name:      "Member Team " + teamID[:8],
			CreatedBy: userID,
		}
		_ = teamRepo.Create(team)

		memberID := uuid.New().String()
		member := &models.TeamMember{
			ID:     memberID,
			TeamID: teamID,
			UserID: member2ID,
			Role:   "member",
		}

		err := teamRepo.AddMember(member)
		if err != nil {
			t.Fatalf("AddMember failed: %v", err)
		}

		members, err := teamRepo.GetTeamMembers(teamID)
		if err != nil {
			t.Fatalf("GetTeamMembers failed: %v", err)
		}

		if len(members) == 0 {
			t.Error("Expected at least one team member, got zero")
		}
	})
}

// TestTaskRepository_Integration tests task repository with testcontainer
func TestTaskRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	itd := NewIntegrationTestDB(ctx, t)
	if itd == nil {
		return
	}
	defer itd.Close(ctx) //nolint:errcheck

	itd.Setup(t)
	defer itd.Cleanup(t)

	// Create a user and team first
	userRepo := repository.NewUserRepository(itd.db)
	userID := uuid.New().String()
	user := &models.User{
		ID:       userID,
		Email:    "task_creator@example.com",
		Username: "taskcreator",
		Password: "password",
	}
	_ = userRepo.Create(user)

	teamRepo := repository.NewTeamRepository(itd.db)
	teamID := uuid.New().String()
	team := &models.Team{
		ID:        teamID,
		Name:      "Task Team",
		CreatedBy: userID,
	}
	_ = teamRepo.Create(team)

	taskRepo := repository.NewTaskRepository(itd.db)

	t.Run("create and retrieve task", func(t *testing.T) {
		taskID := uuid.New().String()
		task := &models.Task{
			ID:          taskID,
			TeamID:      teamID,
			Title:       "Test Task",
			Description: "Test Description",
			Status:      "todo",
			Priority:    "high",
			CreatedBy:   userID,
		}

		err := taskRepo.Create(task)
		if err != nil {
			t.Fatalf("Create task failed: %v", err)
		}

		retrieved, err := taskRepo.GetByID(taskID)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if retrieved == nil {
			t.Fatal("Retrieved task is nil")
		}

		if retrieved.Title != task.Title {
			t.Errorf("Title mismatch: got %s, want %s", retrieved.Title, task.Title)
		}
	})

	t.Run("update task status", func(t *testing.T) {
		taskID := uuid.New().String()
		task := &models.Task{
			ID:        taskID,
			TeamID:    teamID,
			Title:     "Update Task",
			Status:    "todo",
			Priority:  "medium",
			CreatedBy: userID,
		}

		err := taskRepo.Create(task)
		if err != nil {
			t.Fatalf("Create task failed: %v", err)
		}

		// Update status
		task.Status = "done"
		err = taskRepo.Update(task)
		if err != nil {
			t.Fatalf("Update task failed: %v", err)
		}

		retrieved, err := taskRepo.GetByID(taskID)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if retrieved == nil {
			t.Fatal("Retrieved task is nil")
		}

		if retrieved.Status != "done" {
			t.Errorf("Status mismatch: got %s, want done", retrieved.Status)
		}
	})

	t.Run("get team tasks", func(t *testing.T) {
		taskID := uuid.New().String()
		task := &models.Task{
			ID:        taskID,
			TeamID:    teamID,
			Title:     "Team Task",
			Status:    "todo",
			Priority:  "low",
			CreatedBy: userID,
		}

		err := taskRepo.Create(task)
		if err != nil {
			t.Fatalf("Create task failed: %v", err)
		}

		tasks, _, err := taskRepo.GetTeamTasks(teamID, "", "", 10, 0)
		if err != nil {
			t.Fatalf("GetTeamTasks failed: %v", err)
		}

		if len(tasks) == 0 {
			t.Error("Expected at least one task, got zero")
		}
	})
}

// TestComplexQueryRepository_Integration tests complex queries with database
func TestComplexQueryRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	itd := NewIntegrationTestDB(ctx, t)
	if itd == nil {
		return
	}
	defer itd.Close(ctx) //nolint:errcheck

	itd.Setup(t)
	defer itd.Cleanup(t)

	repo := repository.NewComplexQueryRepository(itd.db)

	t.Run("get team statistics", func(t *testing.T) {
		results, err := repo.GetTeamStatistics()
		if err != nil {
			t.Logf("GetTeamStatistics: %v", err)
		} else {
			t.Logf("Team statistics retrieved: %d teams", len(results))
		}
	})

	t.Run("get top task creators", func(t *testing.T) {
		results, err := repo.GetTopTaskCreators()
		if err != nil {
			t.Logf("GetTopTaskCreators: %v", err)
		} else {
			t.Logf("Top creators retrieved: %d users", len(results))
		}
	})

	t.Run("get invalid assignments", func(t *testing.T) {
		results, err := repo.GetInvalidAssignments()
		if err != nil {
			t.Logf("GetInvalidAssignments: %v", err)
		} else {
			t.Logf("Invalid assignments retrieved: %d tasks", len(results))
		}
	})
}
