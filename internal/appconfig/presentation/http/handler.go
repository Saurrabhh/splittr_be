package http

import (
	"net/http"

	"github.com/Saurrabhh/splittr_be/internal/appconfig/domain"
	"github.com/Saurrabhh/splittr_be/internal/auth"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	usecase domain.Usecase
}

func NewHandler(usecase domain.Usecase) *Handler {
	return &Handler{usecase: usecase}
}

func (h *Handler) RegisterRoutes(r chi.Router, optionalAuth func(http.Handler) http.Handler) {
	r.With(optionalAuth).Get("/app-config", h.GetAppConfig)
}

// GetAppConfig godoc
// @Summary Fetch application startup configuration
// @Description Returns app version rules, maintenance status, expense categories, currencies, limits, feature flags, and legal links. Supports optional Bearer auth and ETag caching via If-None-Match.
// @Tags appConfig
// @Accept json
// @Produce json
// @Param If-None-Match header string false "ETag hash for cache validation"
// @Success 200 {object} domain.AppConfigData
// @Failure 304 "Not Modified"
// @Failure 500 {object} response.ErrorResponse
// @Router /app-config [get]
func (h *Handler) GetAppConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	clientETag := r.Header.Get("If-None-Match")

	var userID string
	if identity := auth.IdentityFrom(ctx); identity != nil {
		userID = identity.UserID
	}

	config, notModified, err := h.usecase.GetAppConfig(ctx, userID, clientETag)
	if err != nil {
		response.InternalServerError(w, "failed to fetch app config")
		return
	}

	if notModified {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("ETag", config.Meta.ConfigVersion)
	w.Header().Set("Cache-Control", "public, max-age=300")
	response.JSON(w, http.StatusOK, config)
}
