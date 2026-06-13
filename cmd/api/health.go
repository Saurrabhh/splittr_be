package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type healthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Checks    map[string]string `json:"checks,omitempty"`
}

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	status := "UP"
	checks := make(map[string]string)

	// Check Database
	if err := app.db.Ping(r.Context()); err != nil {
		app.logger.Error("database health check failed", "error", err)
		status = "DOWN"
		checks["database"] = "DOWN"
	} else {
		checks["database"] = "UP"
	}

	res := healthResponse{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Checks:    checks,
	}

	w.Header().Set("Content-Type", "application/json")
	if status == "DOWN" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	_ = json.NewEncoder(w).Encode(res)
}
