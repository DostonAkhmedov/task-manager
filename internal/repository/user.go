package repository

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/DostonAkhmedov/task-manager/internal/models"
)

// UserRepository handles user database operations
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(user *models.User) error {
	user.ID = uuid.New().String()
	
	query := `INSERT INTO users (id, email, username, password, created_at, updated_at)
	          VALUES (?, ?, ?, ?, NOW(), NOW())`
	
	_, err := r.db.Exec(query, user.ID, user.Email, user.Username, user.Password)
	return err
}

// GetByEmail finds a user by email
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	query := `SELECT id, email, username, password, created_at, updated_at
	          FROM users WHERE email = ?`
	
	user := &models.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.Username, &user.Password,
		&user.CreatedAt, &user.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	
	return user, nil
}

// GetByID finds a user by ID
func (r *UserRepository) GetByID(id string) (*models.User, error) {
	query := `SELECT id, email, username, password, created_at, updated_at
	          FROM users WHERE id = ?`
	
	user := &models.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Email, &user.Username, &user.Password,
		&user.CreatedAt, &user.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	
	return user, nil
}

// Update updates a user
func (r *UserRepository) Update(user *models.User) error {
	query := `UPDATE users SET email = ?, username = ?, password = ?, updated_at = NOW()
	          WHERE id = ?`
	
	_, err := r.db.Exec(query, user.Email, user.Username, user.Password, user.ID)
	return err
}
