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

	// Swagger routing
	app.registerSwaggerRoutes(r)

	// API version 1 routes
	r.Route("/v1", func(r chi.Router) {
		// Register domain-specific public/optional routes
		deps.appConfigHandler.RegisterRoutes(r, deps.authMiddleware.OptionalAuthenticate)
		deps.userHandler.RegisterRoutes(r, deps.authMiddleware.Authenticate)

		// Authenticated API routes
		r.Group(func(r chi.Router) {
			r.Use(deps.authMiddleware.Authenticate)
			r.Use(deps.userHandler.UserContext)

			deps.groupHandler.RegisterRoutes(r)
			deps.expenseHandler.RegisterRoutes(r)
			deps.activityHandler.RegisterRoutes(r)
			deps.notificationHandler.RegisterRoutes(r)
			deps.syncHandler.RegisterRoutes(r)
		})
	})

	// Custom 404 Not Found handler using response package
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		response.NotFound(w, "endpoint not found")
	})

	return r
}
