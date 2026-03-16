package models

import "time"

// User represents a user in the system
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Team represents a team
type Team struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TeamMember represents a user's membership in a team
type TeamMember struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	TeamID    string    `json:"team_id"`
	Role      string    `json:"role"` // owner, admin, member
	CreatedAt time.Time `json:"created_at"`
}

// Task represents a task in a team
type Task struct {
	ID          string    `json:"id"`
	TeamID      string    `json:"team_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // todo, in_progress, done
	Priority    string    `json:"priority"` // low, medium, high
	AssigneeID  string    `json:"assignee_id"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TaskHistory represents a change to a task
type TaskHistory struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	Field     string    `json:"field"`
	OldValue  string    `json:"old_value"`
	NewValue  string    `json:"new_value"`
	ChangedBy string    `json:"changed_by"`
	CreatedAt time.Time `json:"created_at"`
}

// TaskComment represents a comment on a task
type TaskComment struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	UserID    string    `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Auth response
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// PagedResponse for pagination
type PagedResponse struct {
	Data      interface{} `json:"data"`
	Page      int         `json:"page"`
	PageSize  int         `json:"page_size"`
	Total     int64       `json:"total"`
	TotalPage int64       `json:"total_page"`
}
