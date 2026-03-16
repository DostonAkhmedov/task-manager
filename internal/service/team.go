package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/DostonAkhmedov/task-manager/internal/models"
	"github.com/DostonAkhmedov/task-manager/internal/repository"
)

// TeamService handles business logic for teams
type TeamService struct {
	teamRepo *repository.TeamRepository
	userRepo *repository.UserRepository
}

// NewTeamService creates a new team service
func NewTeamService(teamRepo *repository.TeamRepository, userRepo *repository.UserRepository) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

// CreateTeam creates a new team
func (s *TeamService) CreateTeam(userID, name string) (*models.Team, error) {
	if name == "" {
		return nil, fmt.Errorf("team name required")
	}

	team := &models.Team{
		ID:        uuid.New().String(),
		Name:      name,
		CreatedBy: userID,
	}

	if err := s.teamRepo.Create(team); err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	// Add creator as team owner
	member := &models.TeamMember{
		ID:     uuid.New().String(),
		UserID: userID,
		TeamID: team.ID,
		Role:   "owner",
	}
	_ = s.teamRepo.AddMember(member)

	return team, nil
}

// GetTeams retrieves all teams for a user
func (s *TeamService) GetTeams(userID string) ([]*models.Team, error) {
	teams, err := s.teamRepo.GetUserTeams(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get teams: %w", err)
	}

	if teams == nil {
		teams = []*models.Team{}
	}

	return teams, nil
}

// GetTeam retrieves a single team
func (s *TeamService) GetTeam(teamID string) (*models.Team, error) {
	team, err := s.teamRepo.GetByID(teamID)
	if err != nil {
		return nil, fmt.Errorf("team not found")
	}
	return team, nil
}

// InviteTeamMember invites a user to a team
func (s *TeamService) InviteTeamMember(userID, teamID, userEmail, role string) (*models.TeamMember, error) {
	// Check if user is team owner/admin
	requesterRole, err := s.teamRepo.GetMemberRole(userID, teamID)
	if err != nil || (requesterRole != "owner" && requesterRole != "admin") {
		return nil, fmt.Errorf("insufficient permissions")
	}

	if userEmail == "" {
		return nil, fmt.Errorf("user email required")
	}

	// Find user by email
	user, err := s.userRepo.GetByEmail(userEmail)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Add user to team
	member := &models.TeamMember{
		ID:     uuid.New().String(),
		UserID: user.ID,
		TeamID: teamID,
		Role:   role,
	}

	if role == "" {
		member.Role = "member"
	}

	if err := s.teamRepo.AddMember(member); err != nil {
		return nil, fmt.Errorf("failed to add member: %w", err)
	}

	return member, nil
}

// GetTeamMembers retrieves all members of a team
func (s *TeamService) GetTeamMembers(userID, teamID string) ([]*models.TeamMember, error) {
	// Check if user is team member
	if _, err := s.teamRepo.GetMemberRole(userID, teamID); err != nil {
		return nil, fmt.Errorf("not a team member")
	}

	members, err := s.teamRepo.GetTeamMembers(teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get members: %w", err)
	}

	if members == nil {
		members = []*models.TeamMember{}
	}

	return members, nil
}
