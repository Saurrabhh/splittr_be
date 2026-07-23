package app

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Saurrabhh/splittr_be/internal/config"
	"github.com/Saurrabhh/splittr_be/internal/db"
)

func TestHealthCheckHandler_Success(t *testing.T) {
	mockDB := &db.DB{
		PingFunc: func(ctx context.Context) error {
			return nil
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	app := NewApplication(&config.Config{}, logger, mockDB)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	app.healthCheckHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var res healthResponse
	if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode health response: %v", err)
	}

	if res.Status != "UP" {
		t.Errorf("expected status 'UP', got %q", res.Status)
	}

	if res.Checks == nil || res.Checks["database"] != "UP" {
		t.Errorf("expected database check 'UP', got %v", res.Checks)
	}
}

func TestHealthCheckHandler_DBFailure(t *testing.T) {
	mockDB := &db.DB{
		PingFunc: func(ctx context.Context) error {
			return errors.New("database connection refused")
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	app := NewApplication(&config.Config{}, logger, mockDB)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	app.healthCheckHandler(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status code %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	var res healthResponse
	if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode health response: %v", err)
	}

	if res.Status != "DOWN" {
		t.Errorf("expected status 'DOWN', got %q", res.Status)
	}

	if res.Checks == nil || res.Checks["database"] != "DOWN" {
		t.Errorf("expected database check 'DOWN', got %v", res.Checks)
	}
}
