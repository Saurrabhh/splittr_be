package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
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

type ErrorType string

const (
	TypeValidation   ErrorType = "VALIDATION"
	TypeNotFound     ErrorType = "NOT_FOUND"
	TypeUnauthorized ErrorType = "UNAUTHORIZED"
	TypeForbidden    ErrorType = "FORBIDDEN"
	TypeConflict     ErrorType = "CONFLICT"
	TypeInternal     ErrorType = "INTERNAL"
)

// AppError represents a generic application error returned by the usecases.
type AppError struct {
	Type    ErrorType
	Message string
	Err     error // Underlying trace / system error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

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
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, ErrBadRequest, message)
}

// Unauthorized sends a 401 Unauthorized JSON response.
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, ErrUnauthorized, message)
}

// Forbidden sends a 403 Forbidden JSON response.
func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, ErrForbidden, message)
}

// NotFound sends a 404 Not Found JSON response.
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, ErrNotFound, message)
}

// InternalServerError sends a 500 Internal Server Error JSON response.
func InternalServerError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, ErrInternalServerError, message)
}

// HandleError centralizes application error mapping to standard HTTP responses.
func HandleError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		switch appErr.Type {
		case TypeValidation:
			BadRequest(w, appErr.Message)
		case TypeNotFound:
			NotFound(w, appErr.Message)
		case TypeUnauthorized:
			Unauthorized(w, appErr.Message)
		case TypeForbidden:
			Forbidden(w, appErr.Message)
		case TypeConflict:
			Error(w, http.StatusConflict, "CONFLICT", appErr.Message)
		case TypeInternal:
			log.Printf("[INTERNAL ERROR] %v", appErr.Err)
			InternalServerError(w, "an internal database/system error occurred")
		default:
			InternalServerError(w, "an unexpected error occurred")
		}
		return
	}

	// Unhandled error fallback
	log.Printf("[UNHANDLED ERROR] %v", err)
	InternalServerError(w, "an unexpected error occurred")
}
