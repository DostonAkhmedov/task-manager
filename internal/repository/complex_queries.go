package repository

import (
	"database/sql"
	"fmt"
)

// ComplexQueryRepository handles complex database queries
type ComplexQueryRepository struct {
	db *sql.DB
}

// NewComplexQueryRepository creates a new complex query repository
func NewComplexQueryRepository(db *sql.DB) *ComplexQueryRepository {
	return &ComplexQueryRepository{db: db}
}

// TeamStatisticsResult represents team statistics
type TeamStatisticsResult struct {
	TeamID        string
	TeamName      string
	MemberCount   int64
	TasksDone7d   int64
	CreatedBy     string
}

// GetTeamStatistics returns statistics for teams
// Query with 3+ JOINs + aggregation:
// "Get for each team: name, number of participants, number of completed tasks in the last 7 days"
func (r *ComplexQueryRepository) GetTeamStatistics() ([]*TeamStatisticsResult, error) {
	actualQuery := `
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

	rows, err := r.db.Query(actualQuery)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	var results []*TeamStatisticsResult
	for rows.Next() {
		result := &TeamStatisticsResult{}
		err := rows.Scan(&result.TeamID, &result.TeamName, &result.MemberCount, 
			&result.TasksDone7d, &result.CreatedBy)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		results = append(results, result)
	}

	return results, rows.Err()
}

// UserTopTasksResult represents user task creation statistics
type UserTopTasksResult struct {
	UserID       string
	Username     string
	TeamID       string
	TeamName     string
	TasksCreated int64
}

// GetTopTaskCreators returns top users by task creation count per team
// Recursive query with window function:
// "Get top-3 users by number of tasks created in each team for the month"
func (r *ComplexQueryRepository) GetTopTaskCreators() ([]*UserTopTasksResult, error) {
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

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	var results []*UserTopTasksResult
	for rows.Next() {
		result := &UserTopTasksResult{}
		err := rows.Scan(&result.UserID, &result.Username, &result.TeamID, 
			&result.TeamName, &result.TasksCreated)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		results = append(results, result)
	}

	return results, rows.Err()
}

// InvalidAssignmentResult represents a task with invalid assignment
type InvalidAssignmentResult struct {
	TaskID        string
	TaskTitle     string
	TeamID        string
	AssigneeID    string
	AssigneeName  string
}

// GetInvalidAssignments returns tasks where assignee is not a team member
// Query with condition on related tables:
// "Find tasks where assignee is not a member of this task's team"
func (r *ComplexQueryRepository) GetInvalidAssignments() ([]*InvalidAssignmentResult, error) {
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

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	var results []*InvalidAssignmentResult
	for rows.Next() {
		result := &InvalidAssignmentResult{}
		err := rows.Scan(&result.TaskID, &result.TaskTitle, &result.TeamID, 
			&result.AssigneeID, &result.AssigneeName)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		results = append(results, result)
	}

	return results, rows.Err()
}

// ActivityStatistics represents user activity
type ActivityStatistics struct {
	UserID    string
	Username  string
	TeamID    string
	TaskCount int64
	CommentCount int64
}

// GetUserActivity returns user activity statistics
func (r *ComplexQueryRepository) GetUserActivity(teamID string) ([]*ActivityStatistics, error) {
	query := `
	SELECT 
		u.id as user_id,
		u.username,
		t.id as team_id,
		COUNT(DISTINCT task_tasks.id) as task_count,
		COUNT(DISTINCT task_comments.id) as comment_count
	FROM users u
	JOIN team_members tm ON u.id = tm.user_id
	JOIN teams t ON tm.team_id = t.id
	LEFT JOIN tasks task_tasks ON u.id = task_tasks.created_by AND task_tasks.team_id = t.id
	LEFT JOIN task_comments ON u.id = task_comments.user_id AND task_tasks.team_id = t.id
	WHERE t.id = ?
	GROUP BY u.id, u.username, t.id
	ORDER BY task_count DESC
	`

	rows, err := r.db.Query(query, teamID)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	var results []*ActivityStatistics
	for rows.Next() {
		result := &ActivityStatistics{}
		err := rows.Scan(&result.UserID, &result.Username, &result.TeamID, 
			&result.TaskCount, &result.CommentCount)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		results = append(results, result)
	}

	return results, rows.Err()
}

// GetTaskChangeFrequency returns which tasks are changed most frequently
func (r *ComplexQueryRepository) GetTaskChangeFrequency(teamID string, limit int) ([]map[string]interface{}, error) {
	query := `
	SELECT 
		t.id,
		t.title,
		COUNT(th.id) as change_count,
		MAX(th.created_at) as last_changed,
		COUNT(DISTINCT tc.id) as comment_count
	FROM tasks t
	LEFT JOIN task_history th ON t.id = th.task_id
	LEFT JOIN task_comments tc ON t.id = tc.task_id
	WHERE t.team_id = ?
	GROUP BY t.id, t.title
	ORDER BY change_count DESC
	LIMIT ?
	`

	rows, err := r.db.Query(query, teamID, limit)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	var results []map[string]interface{}
	for rows.Next() {
		var id, title string
		var changeCount, commentCount int64
		var lastChanged sql.NullTime

		err := rows.Scan(&id, &title, &changeCount, &lastChanged, &commentCount)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		result := map[string]interface{}{
			"id":            id,
			"title":         title,
			"change_count":  changeCount,
			"comment_count": commentCount,
		}
		if lastChanged.Valid {
			result["last_changed"] = lastChanged.Time
		}

		results = append(results, result)
	}

	return results, rows.Err()
}
