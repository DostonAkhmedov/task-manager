package handlers

import (
	"net/http"

	"github.com/DostonAkhmedov/task-manager/internal/repository"
	"github.com/DostonAkhmedov/task-manager/internal/transport/response"
)

// ComplexQueryHandler handles complex query endpoints
type ComplexQueryHandler struct {
	complexQueryRepo *repository.ComplexQueryRepository
}

// NewComplexQueryHandler creates a new complex query handler
func NewComplexQueryHandler(complexQueryRepo *repository.ComplexQueryRepository) *ComplexQueryHandler {
	return &ComplexQueryHandler{
		complexQueryRepo: complexQueryRepo,
	}
}

// GetTeamStatistics retrieves statistics for all teams
// GET /api/v1/queries/team-statistics
// Returns: team name, member count, completed tasks (last 7 days)
func (h *ComplexQueryHandler) GetTeamStatistics(w http.ResponseWriter, r *http.Request) {
	stats, err := h.complexQueryRepo.GetTeamStatistics()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch team statistics", err.Error())
		return
	}

	response.Success(w, http.StatusOK, stats)
}

// GetTopTaskCreators retrieves top 3 users by task creation per team
// GET /api/v1/queries/top-task-creators
// Returns: top-3 users per team ordered by task creation count (last 30 days)
// Uses ROW_NUMBER() window function with CTE
func (h *ComplexQueryHandler) GetTopTaskCreators(w http.ResponseWriter, r *http.Request) {
	topCreators, err := h.complexQueryRepo.GetTopTaskCreators()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch top task creators", err.Error())
		return
	}

	response.Success(w, http.StatusOK, topCreators)
}

// GetInvalidAssignments retrieves tasks with invalid assignments
// GET /api/v1/queries/invalid-assignments
// Returns: tasks where assignee is not a team member (data integrity check)
func (h *ComplexQueryHandler) GetInvalidAssignments(w http.ResponseWriter, r *http.Request) {
	invalid, err := h.complexQueryRepo.GetInvalidAssignments()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch invalid assignments", err.Error())
		return
	}

	response.Success(w, http.StatusOK, invalid)
}
