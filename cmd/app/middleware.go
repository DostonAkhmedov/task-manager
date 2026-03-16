package app

import (
	"net/http"
	"strings"

	"github.com/DostonAkhmedov/task-manager/internal/transport/errors"
	"github.com/DostonAkhmedov/task-manager/internal/transport/response"
	util "github.com/DostonAkhmedov/task-manager/util/auth"
)

// authMiddleware validates JWT tokens and adds claims to context
func authMiddleware(jwtManager *util.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, http.StatusUnauthorized, errors.ErrMissingAuthHeader)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				response.Error(w, http.StatusUnauthorized, errors.ErrInvalidAuthHeader)
				return
			}

			claims, err := jwtManager.ValidateToken(parts[1])
			if err != nil {
				response.Error(w, http.StatusUnauthorized, errors.ErrInvalidToken)
				return
			}

			// Add claims and user_id to context for authenticated requests
			ctx := util.AddClaimsAndUserIDToContext(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
