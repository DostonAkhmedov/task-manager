package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	authutil "github.com/DostonAkhmedov/task-manager/util/auth"
)

// TestAddClaimsToContext tests adding claims to context
func TestAddClaimsToContext(t *testing.T) {
	tests := []struct {
		name   string
		claims *authutil.Claims
	}{
		{
			name: "valid claims",
			claims: &authutil.Claims{
				UserID:   "user123",
				Email:    "user@example.com",
				Username: "testuser",
			},
		},
		{
			name: "claims with empty user id",
			claims: &authutil.Claims{
				UserID: "",
				Email:  "user@example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result := authutil.AddClaimsToContext(ctx, tt.claims)

			if result == nil {
				t.Error("AddClaimsToContext returned nil context")
			}
		})
	}
}

// TestGetClaimsFromContext tests retrieving claims from context
func TestGetClaimsFromContext(t *testing.T) {
	claims := &authutil.Claims{
		UserID:   "user123",
		Email:    "user@example.com",
		Username: "testuser",
	}

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := authutil.AddClaimsToContext(req.Context(), claims)
	req = req.WithContext(ctx)

	retrieved := authutil.GetClaimsFromContext(req)
	if retrieved == nil {
		t.Error("GetClaimsFromContext returned nil")
	} else if retrieved.UserID != claims.UserID {
		t.Errorf("UserID mismatch: got %s, want %s", retrieved.UserID, claims.UserID)
	}
}

// TestAddClaimsAndUserIDToContext tests adding both claims and user_id to context
func TestAddClaimsAndUserIDToContext(t *testing.T) {
	claims := &authutil.Claims{
		UserID:   "user123",
		Email:    "user@example.com",
		Username: "testuser",
	}

	ctx := context.Background()
	result := authutil.AddClaimsAndUserIDToContext(ctx, claims)

	if result == nil {
		t.Error("AddClaimsAndUserIDToContext returned nil context")
	}

	// Verify the claims are properly set by using the exported getter
	req := httptest.NewRequest("GET", "/test", nil)
	req = req.WithContext(result)
	retrieved := authutil.GetClaimsFromContext(req)
	if retrieved == nil || retrieved.UserID != claims.UserID {
		t.Errorf("Claims not properly set in context: got %v", retrieved)
	}
}

// TestGetUserIDFromContext tests extracting user ID from HTTP request context
func TestGetUserIDFromContext(t *testing.T) {
	tests := []struct {
		name         string
		setupRequest func() *http.Request
		expected     string
		shouldExist  bool
	}{
		{
				name: "valid user id",
			setupRequest: func() *http.Request {
				claims := &authutil.Claims{UserID: "user123"}
				req := httptest.NewRequest("GET", "/test", nil)
				ctx := authutil.AddClaimsToContext(req.Context(), claims)
				return req.WithContext(ctx)
			},
			expected:    "user123",
			shouldExist: true,
		},
		{
			name: "no claims in context",
			setupRequest: func() *http.Request {
				return httptest.NewRequest("GET", "/test", nil)
			},
			expected:    "",
			shouldExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupRequest()
			userID := authutil.GetUserIDFromContext(req)
			
			if tt.shouldExist && userID != tt.expected {
				t.Errorf("UserID mismatch: got %s, want %s", userID, tt.expected)
			} else if !tt.shouldExist && userID != "" {
				t.Errorf("Expected empty userID, got %s", userID)
			}
		})
	}
}

// TestJWTValidation tests JWT token validation
func TestJWTValidation(t *testing.T) {
	manager := authutil.NewJWTManagerWithSecret("test-secret-key", 1)

	t.Run("generate and validate token", func(t *testing.T) {
		userID := "user123"
		email := "user@example.com"
		username := "testuser"

		// Generate token
		token, err := manager.GenerateToken(userID, email, username, 1)
		if err != nil {
			t.Errorf("GenerateToken failed: %v", err)
			return
		}

		if token == "" {
			t.Error("GenerateToken returned empty token")
		}

		// Validate token
		validated, err := manager.ValidateToken(token)
		if err != nil {
			t.Errorf("ValidateToken failed: %v", err)
			return
		}

		if validated.UserID != userID {
			t.Errorf("UserID mismatch: got %s, want %s", validated.UserID, userID)
		}
		if validated.Email != email {
			t.Errorf("Email mismatch: got %s, want %s", validated.Email, email)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		_, err := manager.ValidateToken("invalid.token.here")
		if err == nil {
			t.Error("ValidateToken should fail for invalid token")
		}
	})
}
