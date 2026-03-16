package response

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Status  int    `json:"status"`
}

// SuccessResponse represents a standard success response
type SuccessResponse struct {
	Data   interface{} `json:"data"`
	Status int         `json:"status"`
}

// Error writes a JSON error response
func Error(w http.ResponseWriter, statusCode int, message string, details ...string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	detail := ""
	if len(details) > 0 {
		detail = details[0]
	}

	resp := ErrorResponse{
		Error:   message,
		Message: detail,
		Status:  statusCode,
	}

	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

// Success writes a JSON success response
func Success(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := SuccessResponse{
		Data:   data,
		Status: statusCode,
	}

	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

// JSON writes a JSON response with optional status code (defaults to 200)
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data) //nolint:errcheck
}
