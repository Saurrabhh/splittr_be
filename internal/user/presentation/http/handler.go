package http

import (
	"context"
	"net/http"
	"strconv"

	"github.com/Saurrabhh/splittr_be/internal/auth"
	"github.com/Saurrabhh/splittr_be/internal/pagination"
	"github.com/Saurrabhh/splittr_be/internal/request"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/Saurrabhh/splittr_be/internal/user/domain"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for user domain.
type Handler struct {
	uc *domain.UseCase
}

// NewHandler creates a new Handler instance.
func NewHandler(uc *domain.UseCase) *Handler {
	return &Handler{uc: uc}
}

// RegisterRoutes registers endpoints on the router.
func (h *Handler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Route("/users", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Post("/", h.Register)
		r.Group(func(r chi.Router) {
			r.Use(h.UserContext)
			r.Get("/me", h.GetMe)
			r.Put("/me", h.UpdateMe)
			r.Get("/me/settings", h.GetSettings)
			r.Put("/me/settings", h.UpdateSettings)
		})
	})

	r.Route("/friends", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Use(h.UserContext)
		r.Post("/", h.AddFriend)
		r.Get("/", h.GetFriends)
		r.Patch("/{friendId}", h.UpdateFriendStatus)
		r.Delete("/{friendId}", h.RemoveFriend)
		r.Get("/sync", h.SyncFriends)
	})
}

// Register registers the authenticated user.
// @Summary      Register user
// @Description  Create a new user profile using Firebase identities.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "Registration data"
// @Success      201  {object}  domain.User
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /users [post]
// @Security     BearerAuth
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	identity := auth.IdentityFrom(r.Context())
	if identity == nil {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeUnauthorized,
			Message: response.MsgUnauthorized,
		})
		return
	}

	request.Run(w, r, http.StatusCreated, func(ctx context.Context, req RegisterRequest) (*domain.User, error) {
		if req.Name == "" {
			return nil, &response.AppError{
				Type:    response.TypeValidation,
				Message: response.MsgMissingName,
			}
		}

		var emailPtr *string
		if identity.Email != "" {
			emailPtr = &identity.Email
		}

		var phonePtr *string
		if identity.Phone != "" {
			phonePtr = &identity.Phone
		}

		return h.uc.RegisterUser(ctx, identity.UserID, emailPtr, phonePtr, req.Name)
	})
}

// GetMe retrieves the profile of the currently authenticated user.
// @Summary      Get current user profile
// @Description  Retrieve the profile details of the logged-in user.
// @Tags         users
// @Produce      json
// @Success      200  {object}  domain.User
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /users/me [get]
// @Security     BearerAuth
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	u := MustFrom(r.Context())
	response.JSON(w, http.StatusOK, u)
}

// UpdateMe updates the profile metadata of the current user.
// @Summary      Update user profile
// @Description  Update name or default currency for the current user.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body UpdateProfileRequest true "Profile details to update"
// @Success      200  {object}  domain.User
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /users/me [put]
// @Security     BearerAuth
func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	currUser := MustFrom(r.Context())

	request.Run(w, r, http.StatusOK, func(ctx context.Context, req UpdateProfileRequest) (*domain.User, error) {
		return h.uc.UpdateProfile(ctx, currUser.ID, req.Name, req.DefaultCurrency)
	})
}

// GetSettings retrieves settings for the current user.
// @Summary      Get user settings
// @Description  Retrieve privacy and app settings for the current user.
// @Tags         users
// @Produce      json
// @Success      200  {object}  UserSettingsResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /users/me/settings [get]
// @Security     BearerAuth
func (h *Handler) GetSettings(w http.ResponseWriter, r *http.Request) {
	currUser := MustFrom(r.Context())
	settings, err := h.uc.GetUserSettings(r.Context(), currUser.ID)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, UserSettingsResponse{
		AutoAcceptFriendRequests: settings.AutoAcceptFriendRequests,
	})
}

// UpdateSettings updates settings for the current user.
// @Summary      Update user settings
// @Description  Update privacy preferences such as autoAcceptFriendRequests.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body UpdateUserSettingsRequest true "Settings to update"
// @Success      200  {object}  UserSettingsResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /users/me/settings [put]
// @Security     BearerAuth
func (h *Handler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	currUser := MustFrom(r.Context())

	request.Run(w, r, http.StatusOK, func(ctx context.Context, req UpdateUserSettingsRequest) (*UserSettingsResponse, error) {
		settings, err := h.uc.UpdateUserSettings(ctx, currUser.ID, req.AutoAcceptFriendRequests)
		if err != nil {
			return nil, err
		}
		return &UserSettingsResponse{
			AutoAcceptFriendRequests: settings.AutoAcceptFriendRequests,
		}, nil
	})
}

// AddFriend adds a user as a friend or sends a pending request by email or phone.
// @Summary      Add friend or send friend request
// @Description  Create a friendship link (PENDING or ACCEPTED based on target user settings) using email or phone.
// @Tags         friends
// @Accept       json
// @Produce      json
// @Param        request body AddFriendRequest true "Friend email or phone"
// @Success      200  {object}  AddFriendResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /friends [post]
// @Security     BearerAuth
func (h *Handler) AddFriend(w http.ResponseWriter, r *http.Request) {
	currUser := MustFrom(r.Context())

	request.Run(w, r, http.StatusOK, func(ctx context.Context, req AddFriendRequest) (*AddFriendResponse, error) {
		friend, status, err := h.uc.AddFriendByEmailOrPhone(ctx, currUser.ID, req.FriendEmail, req.FriendPhone)
		if err != nil {
			return nil, err
		}
		return &AddFriendResponse{
			Friend: *friend,
			Status: status,
		}, nil
	})
}

// GetFriends returns friends of the current user, optionally filtered by status.
// @Summary      List friends
// @Description  Get friends list. Use optional status parameter to filter by ACCEPTED, PENDING, or BLOCKED.
// @Tags         friends
// @Produce      json
// @Param        status  query  domain.FriendshipStatus  false  "Filter by status: ACCEPTED, PENDING, BLOCKED"
// @Param        limit   query  int                      false  "Items per page (max 100, default 20)"
// @Param        cursor  query  string                   false  "Opaque cursor token from a previous response"
// @Success      200  {object}  pagination.Response[domain.User]
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /friends [get]
// @Security     BearerAuth
func (h *Handler) GetFriends(w http.ResponseWriter, r *http.Request) {
	currUser := MustFrom(r.Context())
	statusParam := r.URL.Query().Get("status")

	if statusParam != "" {
		friends, err := h.uc.ListFriendsByStatus(r.Context(), currUser.ID, domain.FriendshipStatus(statusParam))
		if err != nil {
			response.HandleError(w, err)
			return
		}
		response.JSON(w, http.StatusOK, friends)
		return
	}

	p := pagination.ParseParams(r, 20, 100)
	result, err := h.uc.ListFriends(r.Context(), currUser.ID, p)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

// UpdateFriendStatus updates friendship state (ACCEPTED, DECLINED, BLOCKED).
// @Summary      Update friendship status
// @Description  Accept, decline, or block a friend request / friendship.
// @Tags         friends
// @Accept       json
// @Produce      json
// @Param        friendId path string true "Friend User ID"
// @Param        request body UpdateFriendStatusRequest true "New status (ACCEPTED, DECLINED, BLOCKED)"
// @Success      204  "No Content"
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /friends/{friendId} [patch]
// @Security     BearerAuth
func (h *Handler) UpdateFriendStatus(w http.ResponseWriter, r *http.Request) {
	currUser := MustFrom(r.Context())

	friendID, ok := request.URLParam(w, r, "friendId")
	if !ok {
		return
	}

	request.Run(w, r, http.StatusNoContent, func(ctx context.Context, req UpdateFriendStatusRequest) (any, error) {
		err := h.uc.UpdateFriendshipStatus(ctx, currUser.ID, friendID, req.Status)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
}

// RemoveFriend deletes a friendship link.
// @Summary      Remove friend
// @Description  Delete a friendship link by friend ID.
// @Tags         friends
// @Param        friendId path string true "Friend ID"
// @Success      204  "No Content"
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /friends/{friendId} [delete]
// @Security     BearerAuth
func (h *Handler) RemoveFriend(w http.ResponseWriter, r *http.Request) {
	currUser := MustFrom(r.Context())

	friendID, ok := request.URLParam(w, r, "friendId")
	if !ok {
		return
	}

	err := h.uc.RemoveFriend(r.Context(), currUser.ID, friendID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SyncFriends retrieves delta friendship updates using a monotonic sequence counter.
// @Summary      Sync friends
// @Description  Retrieve friendship changes modified after a given sequence version.
// @Tags         friends
// @Produce      json
// @Param        lastVersion query int64 false "Last received sequence version"
// @Param        limit       query int   false "Maximum items to return (default 100)"
// @Success      200  {object}  domain.FriendSyncResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /friends/sync [get]
// @Security     BearerAuth
func (h *Handler) SyncFriends(w http.ResponseWriter, r *http.Request) {
	currUser := MustFrom(r.Context())

	lastVersionStr := r.URL.Query().Get("lastVersion")
	var lastVersion int64
	if lastVersionStr != "" {
		if v, err := strconv.ParseInt(lastVersionStr, 10, 64); err == nil {
			lastVersion = v
		}
	}

	limitStr := r.URL.Query().Get("limit")
	var limit int32 = 100
	if limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 32); err == nil && l > 0 {
			limit = int32(l)
		}
	}

	res, err := h.uc.SyncFriends(r.Context(), lastVersion, currUser.ID, limit)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, res)
}
