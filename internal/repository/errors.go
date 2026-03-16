package repository

import "errors"

// Repository error definitions
var (
	ErrUserNotFound        = errors.New("user not found")
	ErrTeamNotFound        = errors.New("team not found")
	ErrTaskNotFound        = errors.New("task not found")
	ErrUserNotTeamMember   = errors.New("user is not a team member")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserAlreadyExists   = errors.New("user already exists")
)
