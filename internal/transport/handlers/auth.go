package handlers

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"github.com/DostonAkhmedov/task-manager/internal/models"
	"github.com/DostonAkhmedov/task-manager/internal/repository"
	"github.com/DostonAkhmedov/task-manager/internal/transport/errors"
	"github.com/DostonAkhmedov/task-manager/internal/transport/response"
	util "github.com/DostonAkhmedov/task-manager/util/auth"
)

// AuthHandler handles authentication-related HTTP endpoints
type AuthHandler struct {
	userRepo   *repository.UserRepository
	jwtManager *util.JWTManager
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userRepo *repository.UserRepository, jwtManager *util.JWTManager) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// Register creates a new user account
// POST /api/v1/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if !DecodeJSON(w, r, &req) {
		return
	}

	// Validate input
	if req.Email == "" || req.Username == "" || req.Password == "" {
		response.Error(w, http.StatusBadRequest, errors.ErrMissingFields)
		return
	}

	// Check if user already exists
	_, err := h.userRepo.GetByEmail(req.Email)
	if err == nil {
		response.Error(w, http.StatusConflict, errors.ErrUserAlreadyExists)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, errors.ErrHashPasswordFailed)
		return
	}

	user := &models.User{
		Email:    req.Email,
		Username: req.Username,
		Password: string(hashedPassword),
	}

	if err := h.userRepo.Create(user); err != nil {
		response.Error(w, http.StatusInternalServerError, errors.ErrRegisterUserFailed, err.Error())
		return
	}

	// Generate token
	token, err := h.jwtManager.GenerateToken(user.ID, user.Email, user.Username, 24)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	result := models.AuthResponse{
		Token: token,
		User:  *user,
	}
	result.User.Password = ""

	response.Success(w, http.StatusCreated, result)
}

// Login authenticates a user and returns a JWT token
// POST /api/v1/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if !DecodeJSON(w, r, &req) {
		return
	}

	user, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, errors.ErrInvalidCredentials)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		response.Error(w, http.StatusUnauthorized, errors.ErrInvalidCredentials)
		return
	}

	// Generate token
	token, err := h.jwtManager.GenerateToken(user.ID, user.Email, user.Username, 24)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	result := models.AuthResponse{
		Token: token,
		User:  *user,
	}
	result.User.Password = ""

	response.Success(w, http.StatusOK, result)
}
