package expense

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/Saurrabhh/splittr_be/internal/user"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for expenses and balances.
type Handler struct {
	uc *Usecase
}

// NewHandler creates a new Handler.
func NewHandler(uc *Usecase) *Handler {
	return &Handler{uc: uc}
}

// RegisterRoutes mounts the routes on a Chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/expenses", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Post("/settle", h.Settle)
		r.Get("/", h.List)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetDetails)
			r.Delete("/", h.Delete)
		})
	})
	r.Get("/balances", h.GetBalances)
}

type createExpenseRequest struct {
	Description string       `json:"description"`
	Amount      float64      `json:"amount"`
	Currency    string       `json:"currency"`
	Category    string       `json:"category"`
	GroupID     *string      `json:"groupId"`
	PaidBy      string       `json:"paidBy"`
	SplitType   SplitType    `json:"splitType"`
	Splits      []InputSplit `json:"splits"`
}

type settleExpenseRequest struct {
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	GroupID    *string `json:"groupId"`
	PaidBy     string  `json:"paidBy"`
	ReceivedBy string  `json:"receivedBy"`
}

// Create logs a new expense and distributes the splits.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	currUser := user.UserFrom(r.Context())
	if currUser == nil {
		response.Unauthorized(w, response.ErrUnauthorized, "unauthorized: missing user profile")
		return
	}

	var req createExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrInvalidBody, "invalid request body")
		return
	}

	// Default paidBy to the current user if not supplied
	paidBy := req.PaidBy
	if paidBy == "" {
		paidBy = currUser.ID
	}

	// Default category to 'Other' if not supplied
	category := req.Category
	if category == "" {
		category = "Other"
	}

	exp, splits, err := h.uc.CreateExpense(
		r.Context(),
		req.Description,
		req.Amount,
		req.Currency,
		category,
		req.GroupID,
		paidBy,
		req.SplitType,
		req.Splits,
		currUser.ID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "must be") || strings.Contains(err.Error(), "split calculation") {
			response.BadRequest(w, response.ErrBadRequest, err.Error())
			return
		}
		if strings.Contains(err.Error(), "access validation") || strings.Contains(err.Error(), "member") {
			response.Forbidden(w, response.ErrForbidden, err.Error())
			return
		}
		response.InternalServerError(w, response.ErrInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, map[string]any{
		"expense": exp,
		"splits":  splits,
	})
}

// Settle creates a payment record to clear or reduce debt.
func (h *Handler) Settle(w http.ResponseWriter, r *http.Request) {
	currUser := user.UserFrom(r.Context())
	if currUser == nil {
		response.Unauthorized(w, response.ErrUnauthorized, "unauthorized: missing user profile")
		return
	}

	var req settleExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrInvalidBody, "invalid request body")
		return
	}

	// Default paidBy to current user if not specified
	paidBy := req.PaidBy
	if paidBy == "" {
		paidBy = currUser.ID
	}

	exp, split, err := h.uc.SettleUp(
		r.Context(),
		req.Amount,
		req.Currency,
		req.GroupID,
		paidBy,
		req.ReceivedBy,
		currUser.ID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "must be") || strings.Contains(err.Error(), "different users") {
			response.BadRequest(w, response.ErrBadRequest, err.Error())
			return
		}
		if strings.Contains(err.Error(), "access validation") || strings.Contains(err.Error(), "members") {
			response.Forbidden(w, response.ErrForbidden, err.Error())
			return
		}
		response.InternalServerError(w, response.ErrInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, map[string]any{
		"expense": exp,
		"split":   split,
	})
}

// List lists expenses based on filters (group, personal, or friend).
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	currUser := user.UserFrom(r.Context())
	if currUser == nil {
		response.Unauthorized(w, response.ErrUnauthorized, "unauthorized: missing user profile")
		return
	}

	groupID := r.URL.Query().Get("groupId")
	personalStr := r.URL.Query().Get("personal")
	friendID := r.URL.Query().Get("friendId")

	var filterType string
	var filterID string

	if groupID != "" {
		filterType = "group"
		filterID = groupID
	} else if isPersonal, _ := strconv.ParseBool(personalStr); isPersonal {
		filterType = "personal"
	} else if friendID != "" {
		filterType = "friend"
		filterID = friendID
	} else {
		response.BadRequest(w, response.ErrBadRequest, "missing filter parameter: must supply groupId, personal=true, or friendId")
		return
	}

	expenses, err := h.uc.ListExpenses(r.Context(), filterType, filterID, currUser.ID)
	if err != nil {
		if strings.Contains(err.Error(), "denied") {
			response.Forbidden(w, response.ErrForbidden, err.Error())
			return
		}
		response.InternalServerError(w, response.ErrInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, expenses)
}

// GetDetails retrieves a specific expense and its details.
func (h *Handler) GetDetails(w http.ResponseWriter, r *http.Request) {
	expenseID := chi.URLParam(r, "id")
	if expenseID == "" {
		response.BadRequest(w, response.ErrBadRequest, "expense id is required")
		return
	}

	currUser := user.UserFrom(r.Context())
	if currUser == nil {
		response.Unauthorized(w, response.ErrUnauthorized, "unauthorized: missing user profile")
		return
	}

	exp, splits, err := h.uc.GetExpenseDetails(r.Context(), expenseID, currUser.ID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, response.ErrNotFound, err.Error())
			return
		}
		if strings.Contains(err.Error(), "access denied") {
			response.Forbidden(w, response.ErrForbidden, err.Error())
			return
		}
		response.InternalServerError(w, response.ErrInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"expense": exp,
		"splits":  splits,
	})
}

// Delete soft deletes an expense.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	expenseID := chi.URLParam(r, "id")
	if expenseID == "" {
		response.BadRequest(w, response.ErrBadRequest, "expense id is required")
		return
	}

	currUser := user.UserFrom(r.Context())
	if currUser == nil {
		response.Unauthorized(w, response.ErrUnauthorized, "unauthorized: missing user profile")
		return
	}

	err := h.uc.DeleteExpense(r.Context(), expenseID, currUser.ID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, response.ErrNotFound, err.Error())
			return
		}
		if strings.Contains(err.Error(), "unauthorized") {
			response.Forbidden(w, response.ErrForbidden, err.Error())
			return
		}
		response.InternalServerError(w, response.ErrInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetBalances calculates balances either inside a group or globally.
func (h *Handler) GetBalances(w http.ResponseWriter, r *http.Request) {
	currUser := user.UserFrom(r.Context())
	if currUser == nil {
		response.Unauthorized(w, response.ErrUnauthorized, "unauthorized: missing user profile")
		return
	}

	groupIDStr := r.URL.Query().Get("groupId")
	var groupID *string
	if groupIDStr != "" {
		groupID = &groupIDStr
	}

	simplifiedStr := r.URL.Query().Get("simplified")
	simplified, _ := strconv.ParseBool(simplifiedStr)

	balances, err := h.uc.GetBalances(r.Context(), groupID, currUser.ID, simplified)
	if err != nil {
		if strings.Contains(err.Error(), "denied") {
			response.Forbidden(w, response.ErrForbidden, err.Error())
			return
		}
		response.InternalServerError(w, response.ErrInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, balances)
}
