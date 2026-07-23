package request_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Saurrabhh/splittr_be/internal/request"
	"github.com/Saurrabhh/splittr_be/internal/response"
)

type sampleRequest struct {
	Name string `json:"name"`
}

type sampleResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestDecodeBody_ValidJSON(t *testing.T) {
	body := `{"name":"alice"}`
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	w := httptest.NewRecorder()

	decoded, ok := request.DecodeBody[sampleRequest](w, req)
	if !ok {
		t.Fatal("expected DecodeBody to return true for valid JSON")
	}

	if decoded.Name != "alice" {
		t.Errorf("expected decoded name to be 'alice', got %q", decoded.Name)
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 (recorder default), got %d", w.Code)
	}
}

func TestDecodeBody_InvalidJSON(t *testing.T) {
	body := `{invalid-json`
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	w := httptest.NewRecorder()

	_, ok := request.DecodeBody[sampleRequest](w, req)
	if ok {
		t.Fatal("expected DecodeBody to return false for invalid JSON")
	}

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	var errResp response.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errResp.Code != string(response.ErrInvalidBody) {
		t.Errorf("expected error code %q, got %q", response.ErrInvalidBody, errResp.Code)
	}
}

func TestRun_Success(t *testing.T) {
	body := `{"name":"bob"}`
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	w := httptest.NewRecorder()

	action := func(ctx context.Context, req sampleRequest) (sampleResponse, error) {
		return sampleResponse{
			ID:   "user-123",
			Name: req.Name,
		}, nil
	}

	request.Run[sampleRequest, sampleResponse](w, req, http.StatusCreated, action)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	var res sampleResponse
	if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if res.ID != "user-123" || res.Name != "bob" {
		t.Errorf("unexpected response body: %+v", res)
	}
}

func TestRun_AppError(t *testing.T) {
	body := `{"name":"charlie"}`
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	w := httptest.NewRecorder()

	action := func(ctx context.Context, req sampleRequest) (sampleResponse, error) {
		return sampleResponse{}, &response.AppError{
			Type:    response.TypeValidation,
			Message: "invalid name provided",
		}
	}

	request.Run[sampleRequest, sampleResponse](w, req, http.StatusOK, action)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status code %d (TypeValidation), got %d", http.StatusBadRequest, w.Code)
	}

	var errResp response.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errResp.Code != string(response.ErrBadRequest) {
		t.Errorf("expected error code %q, got %q", response.ErrBadRequest, errResp.Code)
	}
	if errResp.Message != "invalid name provided" {
		t.Errorf("expected error message %q, got %q", "invalid name provided", errResp.Message)
	}
}
