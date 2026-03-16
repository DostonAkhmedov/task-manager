package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/DostonAkhmedov/task-manager/internal/models"
)

// TaskRepository handles task database operations
type TaskRepository struct {
	db *sql.DB
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// Create creates a new task
func (r *TaskRepository) Create(task *models.Task) error {
	task.ID = uuid.New().String()
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	
	query := `INSERT INTO tasks (id, team_id, title, description, status, priority, 
	                            assignee_id, created_by, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	_, err := r.db.Exec(query, task.ID, task.TeamID, task.Title, task.Description,
		task.Status, task.Priority, task.AssigneeID, task.CreatedBy,
		task.CreatedAt, task.UpdatedAt)
	return err
}

// GetByID finds a task by ID
func (r *TaskRepository) GetByID(id string) (*models.Task, error) {
	query := `SELECT id, team_id, title, description, status, priority, assignee_id,
	                 created_by, created_at, updated_at FROM tasks WHERE id = ?`
	
	task := &models.Task{}
	err := r.db.QueryRow(query, id).Scan(
		&task.ID, &task.TeamID, &task.Title, &task.Description,
		&task.Status, &task.Priority, &task.AssigneeID, &task.CreatedBy,
		&task.CreatedAt, &task.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found")
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	
	return task, nil
}

// GetTeamTasks gets all tasks for a team with filtering
func (r *TaskRepository) GetTeamTasks(teamID string, status, assigneeID string, limit, offset int) ([]*models.Task, int64, error) {
	query := `SELECT id, team_id, title, description, status, priority, assignee_id,
	                 created_by, created_at, updated_at FROM tasks WHERE team_id = ?`
	countQuery := `SELECT COUNT(*) FROM tasks WHERE team_id = ?`
	
	args := []interface{}{teamID}
	
	if status != "" {
		query += ` AND status = ?`
		countQuery += ` AND status = ?`
		args = append(args, status)
	}
	
	if assigneeID != "" {
		query += ` AND assignee_id = ?`
		countQuery += ` AND assignee_id = ?`
		args = append(args, assigneeID)
	}
	
	query += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)
	
	// Get total count
	var total int64
	countArgs := args[:len(args)-2]
	err := r.db.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil && err != sql.ErrNoRows {
		return nil, 0, fmt.Errorf("count query error: %w", err)
	}
	
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		err := rows.Scan(&task.ID, &task.TeamID, &task.Title, &task.Description,
			&task.Status, &task.Priority, &task.AssigneeID, &task.CreatedBy,
			&task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("scan error: %w", err)
		}
		tasks = append(tasks, task)
	}
	
	return tasks, total, rows.Err()
}

// Update updates a task
func (r *TaskRepository) Update(task *models.Task) error {
	task.UpdatedAt = time.Now()
	query := `UPDATE tasks SET title = ?, description = ?, status = ?, priority = ?,
	                           assignee_id = ?, updated_at = ? WHERE id = ?`
	
	_, err := r.db.Exec(query, task.Title, task.Description, task.Status,
		task.Priority, task.AssigneeID, task.UpdatedAt, task.ID)
	return err
}

// CreateHistory creates a task history entry
func (r *TaskRepository) CreateHistory(history *models.TaskHistory) error {
	history.ID = uuid.New().String()
	history.CreatedAt = time.Now()
	
	query := `INSERT INTO task_history (id, task_id, field, old_value, new_value, changed_by, created_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?)`
	
	_, err := r.db.Exec(query, history.ID, history.TaskID, history.Field,
		history.OldValue, history.NewValue, history.ChangedBy, history.CreatedAt)
	return err
}

// GetTaskHistory gets all changes for a task
func (r *TaskRepository) GetTaskHistory(taskID string) ([]*models.TaskHistory, error) {
	query := `SELECT id, task_id, field, old_value, new_value, changed_by, created_at
	          FROM task_history WHERE task_id = ? ORDER BY created_at DESC`
	
	rows, err := r.db.Query(query, taskID)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	var histories []*models.TaskHistory
	for rows.Next() {
		history := &models.TaskHistory{}
		err := rows.Scan(&history.ID, &history.TaskID, &history.Field,
			&history.OldValue, &history.NewValue, &history.ChangedBy, &history.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		histories = append(histories, history)
	}
	
	return histories, rows.Err()
}

// InvalidAssignees returns tasks where assignee is not a team member
func (r *TaskRepository) InvalidAssignees(teamID string) ([]*models.Task, error) {
	query := `SELECT t.id, t.team_id, t.title, t.description, t.status, t.priority, 
	                 t.assignee_id, t.created_by, t.created_at, t.updated_at
	          FROM tasks t
	          LEFT JOIN team_members tm ON t.assignee_id = tm.user_id AND t.team_id = tm.team_id
	          WHERE t.team_id = ? AND t.assignee_id IS NOT NULL AND tm.id IS NULL`
	
	rows, err := r.db.Query(query, teamID)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		err := rows.Scan(&task.ID, &task.TeamID, &task.Title, &task.Description,
			&task.Status, &task.Priority, &task.AssigneeID, &task.CreatedBy,
			&task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		tasks = append(tasks, task)
	}
	
	return tasks, rows.Err()
}

// GetTeamStats returns team statistics for the last 7 days
func (r *TaskRepository) GetTeamStats(teamID string) (map[string]int64, error) {
	stats := make(map[string]int64)
	
	// Count members
	query := `SELECT COUNT(*) FROM team_members WHERE team_id = ?`
	var count int64
	err := r.db.QueryRow(query, teamID).Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("members count error: %w", err)
	}
	stats["members"] = count
	
	// Count done tasks in last 7 days
	query = `SELECT COUNT(*) FROM tasks WHERE team_id = ? AND status = 'done' 
	          AND updated_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)`
	err = r.db.QueryRow(query, teamID).Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("tasks count error: %w", err)
	}
	stats["done_tasks_7d"] = count
	
	return stats, nil
}
