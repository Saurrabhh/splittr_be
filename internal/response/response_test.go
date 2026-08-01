package response_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Saurrabhh/splittr_be/internal/response"
)

func TestJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"foo": "bar"}

	response.JSON(w, http.StatusCreated, data)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", contentType)
	}

	var res map[string]string
	if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if res["foo"] != "bar" {
		t.Errorf("expected foo=bar, got %v", res)
	}
}

func TestHandleError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedCode   response.ErrorCode
		expectedMsg    string
	}{
		{
			name:           "TypeValidation maps to 400 Bad Request",
			err:            &response.AppError{Type: response.TypeValidation, Message: "validation failed"},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   response.ErrBadRequest,
			expectedMsg:    "validation failed",
		},
		{
			name:           "TypeNotFound maps to 404 Not Found",
			err:            &response.AppError{Type: response.TypeNotFound, Message: "resource not found"},
			expectedStatus: http.StatusNotFound,
			expectedCode:   response.ErrNotFound,
			expectedMsg:    "resource not found",
		},
		{
			name:           "TypeUnauthorized maps to 401 Unauthorized",
			err:            &response.AppError{Type: response.TypeUnauthorized, Message: "unauthorized access"},
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   response.ErrUnauthorized,
			expectedMsg:    "unauthorized access",
		},
		{
			name:           "TypeForbidden maps to 403 Forbidden",
			err:            &response.AppError{Type: response.TypeForbidden, Message: "access forbidden"},
			expectedStatus: http.StatusForbidden,
			expectedCode:   response.ErrForbidden,
			expectedMsg:    "access forbidden",
		},
		{
			name:           "TypeInternal maps to 500 Internal Server Error",
			err:            &response.AppError{Type: response.TypeInternal, Message: "db error", Err: errors.New("connection failed")},
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   response.ErrInternalServerError,
			expectedMsg:    "an internal database/system error occurred",
		},
		{
			name:           "Generic unhandled error maps to 500 Internal Server Error",
			err:            errors.New("unexpected crash"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   response.ErrInternalServerError,
			expectedMsg:    "an unexpected error occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			response.HandleError(w, tt.err)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type application/json, got %q", contentType)
			}

			var errResp response.ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
				t.Fatalf("failed to decode error response: %v", err)
			}

			if errResp.Code != tt.expectedCode {
				t.Errorf("expected code %q, got %q", tt.expectedCode, errResp.Code)
			}
			if errResp.Message != tt.expectedMsg {
				t.Errorf("expected message %q, got %q", tt.expectedMsg, errResp.Message)
			}
		})
	}
}

func TestHandleError_NilError(t *testing.T) {
	w := httptest.NewRecorder()
	response.HandleError(w, nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected recorder default status 200 for nil error, got %d", w.Code)
	}
	if w.Body.Len() > 0 {
		t.Errorf("expected empty body for nil error, got %q", w.Body.String())
	}
}

func TestAppError_ErrorAndUnwrap(t *testing.T) {
	innerErr := errors.New("underlying cause")
	appErr := &response.AppError{
		Type:    response.TypeInternal,
		Message: "high-level message",
		Err:     innerErr,
	}

	if appErr.Error() != "high-level message: underlying cause" {
		t.Errorf("unexpected error string: %q", appErr.Error())
	}

	if !errors.Is(appErr, innerErr) {
		t.Error("expected errors.Is to match inner error")
	}

	simpleErr := &response.AppError{
		Type:    response.TypeValidation,
		Message: "simple error",
	}

	if simpleErr.Error() != "simple error" {
		t.Errorf("unexpected error string for simple error: %q", simpleErr.Error())
	}
}
