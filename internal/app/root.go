package app

import (
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
}

func (app *Application) rootHandler(w http.ResponseWriter, r *http.Request) {
	res := rootResponse{
		Name:        "Splittr API",
		Version:     "1.0.0",
		Environment: os.Getenv("APP_ENV"),
		Status:      "healthy",
		Docs:        "/health", // This can be updated to point to a documentation route (like /docs or Swagger) later.
	}

	response.JSON(w, http.StatusOK, res)
}
