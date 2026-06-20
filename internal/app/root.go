package app

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Saurrabhh/splittr_be/internal/response"
)

type rootResponse struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
	Status      string `json:"status"`
	Docs        string `json:"docs"`
	Health        string `json:"health"`
}

func (app *Application) rootHandler(w http.ResponseWriter, r *http.Request) {
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	healthURL := fmt.Sprintf("%s://%s/health", scheme, r.Host)
	docsURL := fmt.Sprintf("%s://%s/docs", scheme, r.Host)

	res := rootResponse{
		Name:        "Splittr API",
		Version:     "1.0.0",
		Environment: os.Getenv("APP_ENV"),
		Status:      "healthy",
		Docs:        docsURL,
		Health:      healthURL,
	}

	response.JSON(w, http.StatusOK, res)
}
