package app

import (
	"net/http"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/response"
)

type healthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Checks    map[string]string `json:"checks,omitempty"`
}

func (app *Application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	status := "UP"
	checks := make(map[string]string)

	// Check Database
	if err := app.DB.Ping(r.Context()); err != nil {
		app.Logger.Error("database health check failed", "error", err)
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

	if status == "DOWN" {
		response.JSON(w, http.StatusServiceUnavailable, res)
	} else {
		response.JSON(w, http.StatusOK, res)
	}
}
