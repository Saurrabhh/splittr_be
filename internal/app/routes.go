package app

import (
	"net/http"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// routes configures global middlewares, defines route groups, and mounts handlers.
func (app *Application) routes(deps *dependencies) http.Handler {
	r := chi.NewRouter()

	// Global Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromRemoteAddr)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Default/root route
	r.Get("/", app.rootHandler)

	// Health check route
	r.Get("/health", app.healthCheckHandler)

	// API version 1 routes
	r.Route("/v1", func(r chi.Router) {
		// Register domain-specific routes
		deps.userHandler.RegisterRoutes(r, deps.authMiddleware.Authenticate)
	})

	// Custom 404 Not Found handler using response package
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		response.NotFound(w, response.ErrNotFound, "endpoint not found")
	})

	return r
}
