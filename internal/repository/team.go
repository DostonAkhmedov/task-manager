package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/DostonAkhmedov/task-manager/internal/models"
)

// TeamRepository handles team database operations
type TeamRepository struct {
	db *sql.DB
}

// NewTeamRepository creates a new team repository
func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

// Create creates a new team
func (r *TeamRepository) Create(team *models.Team) error {
	team.ID = uuid.New().String()
	team.CreatedAt = time.Now()
	team.UpdatedAt = time.Now()
	
	query := `INSERT INTO teams (id, name, created_by, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?)`
	
	_, err := r.db.Exec(query, team.ID, team.Name, team.CreatedBy, team.CreatedAt, team.UpdatedAt)
	return err
}

// GetByID finds a team by ID
func (r *TeamRepository) GetByID(id string) (*models.Team, error) {
	query := `SELECT id, name, created_by, created_at, updated_at FROM teams WHERE id = ?`
	
	team := &models.Team{}
	err := r.db.QueryRow(query, id).Scan(
		&team.ID, &team.Name, &team.CreatedBy, &team.CreatedAt, &team.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("team not found")
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	
	return team, nil
}

// GetUserTeams gets all teams for a user
func (r *TeamRepository) GetUserTeams(userID string) ([]*models.Team, error) {
	query := `SELECT t.id, t.name, t.created_by, t.created_at, t.updated_at
	          FROM teams t
	          JOIN team_members tm ON t.id = tm.team_id
	          WHERE tm.user_id = ?
	          ORDER BY t.created_at DESC`
	
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	var teams []*models.Team
	for rows.Next() {
		team := &models.Team{}
		err := rows.Scan(&team.ID, &team.Name, &team.CreatedBy, &team.CreatedAt, &team.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		teams = append(teams, team)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	
	return teams, nil
}

// Update updates a team
func (r *TeamRepository) Update(team *models.Team) error {
	team.UpdatedAt = time.Now()
	query := `UPDATE teams SET name = ?, updated_at = ? WHERE id = ?`
	
	_, err := r.db.Exec(query, team.Name, team.UpdatedAt, team.ID)
	return err
}

// AddMember adds a user to a team
func (r *TeamRepository) AddMember(member *models.TeamMember) error {
	member.ID = uuid.New().String()
	member.CreatedAt = time.Now()
	
	query := `INSERT INTO team_members (id, user_id, team_id, role, created_at)
	          VALUES (?, ?, ?, ?, ?)`
	
	_, err := r.db.Exec(query, member.ID, member.UserID, member.TeamID, member.Role, member.CreatedAt)
	return err
}

// GetTeamMembers gets all members of a team
func (r *TeamRepository) GetTeamMembers(teamID string) ([]*models.TeamMember, error) {
	query := `SELECT id, user_id, team_id, role, created_at FROM team_members WHERE team_id = ?`
	
	rows, err := r.db.Query(query, teamID)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	var members []*models.TeamMember
	for rows.Next() {
		member := &models.TeamMember{}
		err := rows.Scan(&member.ID, &member.UserID, &member.TeamID, &member.Role, &member.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		members = append(members, member)
	}
	
	return members, rows.Err()
}

// GetMemberRole gets a user's role in a team
func (r *TeamRepository) GetMemberRole(userID, teamID string) (string, error) {
	query := `SELECT role FROM team_members WHERE user_id = ? AND team_id = ?`
	
	var role string
	err := r.db.QueryRow(query, userID, teamID).Scan(&role)
	
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("user not member of team")
	}
	if err != nil {
		return "", fmt.Errorf("database error: %w", err)
	}
	
	return role, nil
}
