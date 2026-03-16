package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/DostonAkhmedov/task-manager/internal/transport/errors"
	"github.com/DostonAkhmedov/task-manager/internal/transport/response"
	util "github.com/DostonAkhmedov/task-manager/util/auth"
)

// GetUserIDOrRespond extracts user ID from context or responds with error
func GetUserIDOrRespond(w http.ResponseWriter, r *http.Request) (string, bool) {
	userID := util.GetUserIDFromContext(r)
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, errors.ErrUnauthorized)
		return "", false
	}
	return userID, true
}

// DecodeJSON decodes JSON request body or responds with error
func DecodeJSON(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		response.Error(w, http.StatusBadRequest, errors.ErrInvalidRequest, err.Error())
		return false
	}
	return true
}

// GetIntParam retrieves an integer query parameter with default
func GetIntParam(r *http.Request, name string, defaultValue int) int {
	value := r.URL.Query().Get(name)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// GetStringParam retrieves a string query parameter
func GetStringParam(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}
