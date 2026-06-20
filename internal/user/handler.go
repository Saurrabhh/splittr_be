package user

import (
	"encoding/json"
	"net/http"

	"github.com/Saurrabhh/splittr_be/internal/auth"
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

type registerRequest struct {
	Name string `json:"name"`
}

type updateProfileRequest struct {
	Name            string `json:"name"`
	DefaultCurrency string `json:"defaultCurrency"`
}

type addFriendRequest struct {
	FriendEmail string `json:"friendEmail"`
	FriendPhone string `json:"friendPhone"`
}

// Register registers the authenticated user.
// @Summary      Register user
// @Description  Create a new user profile using Firebase identities.
// @Tags         users
// @Accept       json
// @Produce      json
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

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, response.ErrInvalidBody, "invalid request body")
		return
	}

	if req.Name == "" {
		response.Error(w, http.StatusBadRequest, response.ErrNameRequired, "name is required")
		return
	}

	var emailPtr *string
	if identity.Email != "" {
		emailPtr = &identity.Email
	}

	var phonePtr *string
	if identity.Phone != "" {
		phonePtr = &identity.Phone
	}

	u, err := h.uc.RegisterUser(r.Context(), identity.UserID, emailPtr, phonePtr, req.Name)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, u)
}

// GetMe retrieves the profile of the currently authenticated user.
// @Summary      Get current user profile
// @Description  Retrieve the profile details of the logged-in user.
// @Tags         users
// @Produce      json
// @Success      200  {object}  User
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /users/me [get]
// @Security     BearerAuth
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	if u == nil {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeUnauthorized,
			Message: "unauthorized: missing user profile",
		})
		return
	}
	response.JSON(w, http.StatusOK, u)
}

// UpdateMe updates the profile metadata of the current user.
// @Summary      Update user profile
// @Description  Update name or default currency for the current user.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body updateProfileRequest true "Profile details to update"
// @Success      200  {object}  User
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /users/me [put]
// @Security     BearerAuth
func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	currUser := UserFrom(r.Context())
	if currUser == nil {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeUnauthorized,
			Message: "unauthorized: missing user profile",
		})
		return
	}

	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, response.ErrInvalidBody, "invalid request body")
		return
	}

	u, err := h.uc.UpdateProfile(r.Context(), currUser.ID, req.Name, req.DefaultCurrency)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, u)
}

// AddFriend adds a user as a friend by email or phone.
// @Summary      Add friend
// @Description  Create a friendship link with another user by their email or phone.
// @Tags         friends
// @Accept       json
// @Produce      json
// @Param        request body addFriendRequest true "Friend email or phone"
// @Success      200  {object}  User
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /friends [post]
// @Security     BearerAuth
func (h *Handler) AddFriend(w http.ResponseWriter, r *http.Request) {
	currUser := UserFrom(r.Context())
	if currUser == nil {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeUnauthorized,
			Message: "unauthorized: missing user profile",
		})
		return
	}

	var req addFriendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, response.ErrInvalidBody, "invalid request body")
		return
	}

	friend, err := h.uc.AddFriendByEmailOrPhone(r.Context(), currUser.ID, req.FriendEmail, req.FriendPhone)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, friend)
}

// GetFriends returns all friends of the current user.
// @Summary      List friends
// @Description  Get a list of all friends of the currently authenticated user.
// @Tags         friends
// @Produce      json
// @Success      200  {array}   User
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /friends [get]
// @Security     BearerAuth
func (h *Handler) GetFriends(w http.ResponseWriter, r *http.Request) {
	currUser := UserFrom(r.Context())
	if currUser == nil {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeUnauthorized,
			Message: "unauthorized: missing user profile",
		})
		return
	}

	friends, err := h.uc.ListFriends(r.Context(), currUser.ID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, friends)
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
	currUser := UserFrom(r.Context())
	if currUser == nil {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeUnauthorized,
			Message: "unauthorized: missing user profile",
		})
		return
	}

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
