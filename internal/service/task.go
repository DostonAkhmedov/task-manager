package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DostonAkhmedov/task-manager/internal/cache"
	"github.com/DostonAkhmedov/task-manager/internal/models"
	"github.com/DostonAkhmedov/task-manager/internal/repository"
)

// TaskService handles business logic for tasks
type TaskService struct {
	taskRepo  *repository.TaskRepository
	teamRepo  *repository.TeamRepository
	redis     *cache.Redis
	cacheTTL  time.Duration
}

// NewTaskService creates a new task service
func NewTaskService(taskRepo *repository.TaskRepository, teamRepo *repository.TeamRepository, redis *cache.Redis, cacheTTL time.Duration) *TaskService {
	return &TaskService{
		taskRepo:  taskRepo,
		teamRepo:  teamRepo,
		redis:     redis,
		cacheTTL:  cacheTTL,
	}
}

// CreateTask creates a new task
func (s *TaskService) CreateTask(userID, teamID, title, description, priority, assigneeID string) (*models.Task, error) {
	// Check if user is team member
	if _, err := s.teamRepo.GetMemberRole(userID, teamID); err != nil {
		return nil, fmt.Errorf("not a team member")
	}

	task := &models.Task{
		TeamID:      teamID,
		Title:       title,
		Description: description,
		Status:      "todo",
		Priority:    priority,
		AssigneeID:  assigneeID,
		CreatedBy:   userID,
	}

	if err := s.taskRepo.Create(task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Clear cache for all tasks of this team
s.redis.DeleteByPattern(context.Background(), fmt.Sprintf("team_tasks:%s:*", teamID)) //nolint:errcheck

	return task, nil
}

// GetTasks retrieves tasks with filtering and pagination
func (s *TaskService) GetTasks(userID, teamID, status, assigneeID string, page, pageSize int) (*models.PagedResponse, error) {
	// Check if user is team member
	if _, err := s.teamRepo.GetMemberRole(userID, teamID); err != nil {
		return nil, fmt.Errorf("not a team member")
	}

	ctx := context.Background()
	// Generate cache key based on filters
	cacheKey := fmt.Sprintf("team_tasks:%s:status:%s:assignee:%s:page:%d:size:%d", teamID, status, assigneeID, page, pageSize)

	// Try to get from cache
	cachedData, err := s.redis.Get(ctx, cacheKey)
	if err == nil && cachedData != "" {
		// Parse cached response
		var response models.PagedResponse
		if err := json.Unmarshal([]byte(cachedData), &response); err == nil {
			return &response, nil
		}
	}

	offset := (page - 1) * pageSize

	tasks, total, err := s.taskRepo.GetTeamTasks(teamID, status, assigneeID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	response := &models.PagedResponse{
		Data:      tasks,
		Page:      page,
		PageSize:  pageSize,
		Total:     total,
		TotalPage: totalPages,
	}

	// Cache the response with configured TTL
	if data, err := json.Marshal(response); err == nil {
		s.redis.Set(ctx, cacheKey, string(data), s.cacheTTL) //nolint:errcheck
	}

	return response, nil
}

// GetTask retrieves a single task
func (s *TaskService) GetTask(taskID string) (*models.Task, error) {
	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		return nil, fmt.Errorf("task not found")
	}
	return task, nil
}

// UpdateTask updates a task
func (s *TaskService) UpdateTask(userID, taskID, title, description, status, priority, assigneeID string) (*models.Task, error) {
	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		return nil, fmt.Errorf("task not found")
	}

	// Check if user is team member
	if _, err := s.teamRepo.GetMemberRole(userID, task.TeamID); err != nil {
		return nil, fmt.Errorf("not a team member")
	}

	oldTask := *task

	// Update fields
	if title != "" {
		task.Title = title
	}
	if description != "" {
		task.Description = description
	}
	if status != "" {
		task.Status = status
	}
	if priority != "" {
		task.Priority = priority
	}
	task.AssigneeID = assigneeID

	if err := s.taskRepo.Update(task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	// Record history changes
	if oldTask.Status != task.Status {
		s.taskRepo.CreateHistory(&models.TaskHistory{ //nolint:errcheck
			TaskID:    taskID,
			Field:     "status",
			OldValue:  oldTask.Status,
			NewValue:  task.Status,
			ChangedBy: userID,
		})
	}

	// Clear cache for all tasks of this team
s.redis.DeleteByPattern(context.Background(), fmt.Sprintf("team_tasks:%s:*", task.TeamID)) //nolint:errcheck

	return task, nil
}

// GetTaskHistory retrieves task history
func (s *TaskService) GetTaskHistory(taskID string) ([]*models.TaskHistory, error) {
	history, err := s.taskRepo.GetTaskHistory(taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}
	return history, nil
}
