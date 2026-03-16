package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/DostonAkhmedov/task-manager/internal/service"
	"github.com/DostonAkhmedov/task-manager/internal/transport/errors"
	"github.com/DostonAkhmedov/task-manager/internal/transport/response"
)

// TeamHandler handles team-related HTTP endpoints
type TeamHandler struct {
	teamService *service.TeamService
}

// NewTeamHandler creates a new team handler
func NewTeamHandler(teamService *service.TeamService) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
	}
}

// CreateTeam creates a new team
// POST /api/v1/teams
func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDOrRespond(w, r)
	if !ok {
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	if !DecodeJSON(w, r, &req) {
		return
	}

	team, err := h.teamService.CreateTeam(userID, req.Name)
	if err != nil {
		response.Error(w, http.StatusBadRequest, errors.ErrCreateTeamFailed, err.Error())
		return
	}

	response.Success(w, http.StatusCreated, team)
}

// GetTeams retrieves all teams for the current user
// GET /api/v1/teams
func (h *TeamHandler) GetTeams(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDOrRespond(w, r)
	if !ok {
		return
	}

	teams, err := h.teamService.GetTeams(userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, errors.ErrGetTeamsFailed, err.Error())
		return
	}

	response.Success(w, http.StatusOK, teams)
}

// GetTeam retrieves a single team by ID
// GET /api/v1/teams/{id}
func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "id")

	team, err := h.teamService.GetTeam(teamID)
	if err != nil {
		response.Error(w, http.StatusNotFound, errors.ErrTeamNotFound)
		return
	}

	response.Success(w, http.StatusOK, team)
}

// InviteTeamMember invites a user to a team
// POST /api/v1/teams/{id}/invite
func (h *TeamHandler) InviteTeamMember(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "id")

	userID, ok := GetUserIDOrRespond(w, r)
	if !ok {
		return
	}

	var req struct {
		UserEmail string `json:"user_email"`
		Role      string `json:"role"`
	}

	if !DecodeJSON(w, r, &req) {
		return
	}

	member, err := h.teamService.InviteTeamMember(userID, teamID, req.UserEmail, req.Role)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	response.Success(w, http.StatusCreated, member)
}

// GetTeamMembers retrieves all members of a team
// GET /api/v1/teams/{id}/members
func (h *TeamHandler) GetTeamMembers(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "id")

	userID, ok := GetUserIDOrRespond(w, r)
	if !ok {
		return
	}

	members, err := h.teamService.GetTeamMembers(userID, teamID)
	if err != nil {
		response.Error(w, http.StatusForbidden, errors.ErrForbidden, err.Error())
		return
	}

	response.Success(w, http.StatusOK, members)
}
