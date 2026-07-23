package user

import (
	"context"
	"net/http"

	"github.com/Saurrabhh/splittr_be/internal/auth"
	"github.com/Saurrabhh/splittr_be/internal/pagination"
	"github.com/Saurrabhh/splittr_be/internal/request"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for user domain.
type Handler struct {
	uc *Usecase
}

// NewHandler creates a new Handler instance.
func NewHandler(uc *Usecase) *Handler {
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
		})
	})

	r.Route("/friends", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Use(h.UserContext)
		r.Post("/", h.AddFriend)
		r.Get("/", h.GetFriends)
		r.Delete("/{friendId}", h.RemoveFriend)
	})
}

// Register registers the authenticated user.
// @Summary      Register user
// @Description  Create a new user profile using Firebase identities.
// @Tags         users
// @Accept       JSON
// @Produce      JSON
// @Param        request body registerRequest true "Registration data"
// @Success      201  {object}  User
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
			Message: "unauthorized: missing auth credentials",
		})
		return
	}

	request.Run(w, r, http.StatusCreated, func(ctx context.Context, req registerRequest) (*User, error) {
		if req.Name == "" {
			return nil, &response.AppError{
				Type:    response.TypeValidation,
				Message: "name is required",
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
// @Produce      JSON
// @Success      200  {object}  User
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
// @Accept       JSON
// @Produce      JSON
// @Param        request body updateProfileRequest true "Profile details to update"
// @Success      200  {object}  User
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /users/me [put]
// @Security     BearerAuth
func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	currUser := MustFrom(r.Context())

	request.Run(w, r, http.StatusOK, func(ctx context.Context, req updateProfileRequest) (*User, error) {
		return h.uc.UpdateProfile(ctx, currUser.ID, req.Name, req.DefaultCurrency)
	})
}

// AddFriend adds a user as a friend by email or phone.
// @Summary      Add friend
// @Description  Create a friendship link with another user by their email or phone.
// @Tags         friends
// @Accept       JSON
// @Produce      JSON
// @Param        request body addFriendRequest true "Friend email or phone"
// @Success      200  {object}  User
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /friends [post]
// @Security     BearerAuth
func (h *Handler) AddFriend(w http.ResponseWriter, r *http.Request) {
	currUser := MustFrom(r.Context())

	request.Run(w, r, http.StatusOK, func(ctx context.Context, req addFriendRequest) (*User, error) {
		return h.uc.AddFriendByEmailOrPhone(ctx, currUser.ID, req.FriendEmail, req.FriendPhone)
	})
}

// GetFriends returns all friends of the current user.
// @Summary      List friends
// @Description  Get a cursor-paginated list of the current user's friends.
// @Tags         friends
// @Produce      JSON
// @Param        limit   query  int     false  "Items per page (max 100, default 20)"
// @Param        cursor  query  string  false  "Opaque cursor token from a previous response"
// @Success      200  {object}  ListFriendsResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /friends [get]
// @Security     BearerAuth
func (h *Handler) GetFriends(w http.ResponseWriter, r *http.Request) {
	currUser := MustFrom(r.Context())
	p := pagination.ParseParams(r, 20, 100)
	result, err := h.uc.ListFriends(r.Context(), currUser.ID, p)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
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

	friendID := chi.URLParam(r, "friendId")
	if friendID == "" {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeValidation,
			Message: "friendId is required",
		})
		return
	}

	err := h.uc.RemoveFriend(r.Context(), currUser.ID, friendID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
