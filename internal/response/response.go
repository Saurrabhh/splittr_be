package response

import (
	"encoding/json"
	"net/http"
)

// ErrorCode represents a centralized standard API error code.
type ErrorCode string

const (
	// Generic error codes
	ErrBadRequest          ErrorCode = "BAD_REQUEST"
	ErrUnauthorized        ErrorCode = "UNAUTHORIZED"
	ErrForbidden           ErrorCode = "FORBIDDEN"
	ErrNotFound            ErrorCode = "NOT_FOUND"
	ErrInternalServerError ErrorCode = "INTERNAL_SERVER_ERROR"

	// Domain-specific error codes
	ErrInvalidBody  ErrorCode = "INVALID_BODY"
	ErrNameRequired ErrorCode = "NAME_REQUIRED"
	ErrUserNotFound ErrorCode = "USER_NOT_FOUND"
)

// ErrorResponse represents a standard JSON error response structure.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// JSON sends a raw JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// Error sends a JSON error response with the HTTP status, custom error code, and error message.
func Error(w http.ResponseWriter, status int, errorCode ErrorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	res := ErrorResponse{
		Code:    string(errorCode),
		Message: message,
	}

	_ = json.NewEncoder(w).Encode(res)
}

// BadRequest sends a 400 Bad Request JSON response.
func BadRequest(w http.ResponseWriter, code ErrorCode, message string) {
	Error(w, http.StatusBadRequest, code, message)
}

// Unauthorized sends a 401 Unauthorized JSON response.
func Unauthorized(w http.ResponseWriter, code ErrorCode, message string) {
	Error(w, http.StatusUnauthorized, code, message)
}

// Forbidden sends a 403 Forbidden JSON response.
func Forbidden(w http.ResponseWriter, code ErrorCode, message string) {
	Error(w, http.StatusForbidden, code, message)
}

// NotFound sends a 404 Not Found JSON response.
func NotFound(w http.ResponseWriter, code ErrorCode, message string) {
	Error(w, http.StatusNotFound, code, message)
}

// InternalServerError sends a 500 Internal Server Error JSON response.
func InternalServerError(w http.ResponseWriter, code ErrorCode, message string) {
	Error(w, http.StatusInternalServerError, code, message)
}
