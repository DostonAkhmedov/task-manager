package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/DostonAkhmedov/task-manager/internal/service"
	"github.com/DostonAkhmedov/task-manager/internal/transport/errors"
	"github.com/DostonAkhmedov/task-manager/internal/transport/response"
)

// TaskHandler handles task-related HTTP endpoints
type TaskHandler struct {
	taskService *service.TaskService
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

// CreateTask creates a new task
// POST /api/v1/tasks
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDOrRespond(w, r)
	if !ok {
		return
	}

	var req struct {
		TeamID      string `json:"team_id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Priority    string `json:"priority"`
		AssigneeID  string `json:"assignee_id"`
	}

	if !DecodeJSON(w, r, &req) {
		return
	}

	if req.TeamID == "" || req.Title == "" {
		response.Error(w, http.StatusBadRequest, errors.ErrMissingFields)
		return
	}

	task, err := h.taskService.CreateTask(userID, req.TeamID, req.Title, req.Description, req.Priority, req.AssigneeID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, errors.ErrCreateTaskFailed, err.Error())
		return
	}

	response.Success(w, http.StatusCreated, task)
}

// GetTasks retrieves tasks with filtering and pagination
// GET /api/v1/tasks?team_id=xxx&status=xxx&assignee_id=xxx&page=1&page_size=20
func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDOrRespond(w, r)
	if !ok {
		return
	}

	teamID := GetStringParam(r, "team_id")
	if teamID == "" {
		response.Error(w, http.StatusBadRequest, "team_id is required")
		return
	}

	status := GetStringParam(r, "status")
	assigneeID := GetStringParam(r, "assignee_id")
	page := GetIntParam(r, "page", 1)
	pageSize := GetIntParam(r, "page_size", 20)

	result, err := h.taskService.GetTasks(userID, teamID, status, assigneeID, page, pageSize)
	if err != nil {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}

	response.Success(w, http.StatusOK, result)
}

// GetTask retrieves a single task by ID
// GET /api/v1/tasks/{id}
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	task, err := h.taskService.GetTask(taskID)
	if err != nil {
		response.Error(w, http.StatusNotFound, errors.ErrTaskNotFound)
		return
	}

	response.Success(w, http.StatusOK, task)
}

// UpdateTask updates an existing task
// PUT /api/v1/tasks/{id}
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	userID, ok := GetUserIDOrRespond(w, r)
	if !ok {
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
		Priority    string `json:"priority"`
		AssigneeID  string `json:"assignee_id"`
	}

	if !DecodeJSON(w, r, &req) {
		return
	}

	task, err := h.taskService.UpdateTask(userID, taskID, req.Title, req.Description, req.Status, req.Priority, req.AssigneeID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, errors.ErrUpdateTaskFailed, err.Error())
		return
	}

	response.Success(w, http.StatusOK, task)
}

// GetTaskHistory retrieves the history of changes for a task
// GET /api/v1/tasks/{id}/history
func (h *TaskHandler) GetTaskHistory(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	history, err := h.taskService.GetTaskHistory(taskID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, errors.ErrGetTaskHistoryFailed, err.Error())
		return
	}

	response.Success(w, http.StatusOK, history)
}
