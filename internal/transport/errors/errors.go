package errors

// HTTP Error Messages
const (
	// Validation errors
	ErrInvalidRequest = "invalid request"
	ErrMissingFields  = "missing required fields"
	ErrInvalidInput   = "invalid input"

	// Authentication errors
	ErrUnauthorized           = "unauthorized"
	ErrInvalidToken           = "invalid token"
	ErrMissingAuthHeader      = "missing authorization header"
	ErrInvalidAuthHeader      = "invalid authorization header"
	ErrUserAlreadyExists      = "user already exists"
	ErrInvalidCredentials     = "invalid credentials"

	// Authorization errors
	ErrForbidden         = "forbidden"
	ErrNotTeamMember     = "not a team member"
	ErrInsufficientPerms = "insufficient permissions"

	// Resource errors
	ErrNotFound        = "not found"
	ErrTaskNotFound    = "task not found"
	ErrTeamNotFound    = "team not found"
	ErrUserNotFound    = "user not found"
	ErrConflict        = "resource already exists"

	// Server errors
	ErrInternal = "internal server error"
	ErrDatabase = "database error"
)

// Detailed error messages for logging
const (
	ErrCreateTaskFailed    = "failed to create task"
	ErrGetTasksFailed      = "failed to get tasks"
	ErrUpdateTaskFailed    = "failed to update task"
	ErrDeleteTaskFailed    = "failed to delete task"
	ErrGetTaskHistoryFailed = "failed to get task history"

	ErrCreateTeamFailed  = "failed to create team"
	ErrGetTeamsFailed    = "failed to get teams"
	ErrGetTeamFailed     = "failed to get team"
	ErrUpdateTeamFailed  = "failed to update team"
	ErrGetTeamMembersFailed = "failed to get team members"

	ErrRegisterUserFailed = "failed to register user"
	ErrLoginFailed        = "failed to login"
	ErrHashPasswordFailed = "failed to hash password"
)
