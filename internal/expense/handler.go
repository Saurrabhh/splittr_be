package expense

import (
	"encoding/json"
	"net/http"
	"strconv"

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
// @Summary      Create expense
// @Description  Create a new expense with equal/exact/percentage splits.
// @Tags         expenses
// @Accept       json
// @Produce      json
// @Param        request body createExpenseRequest true "Expense details and splits structure"
// @Success      201  {object}  CreateExpenseResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /expenses [post]
// @Security     Bearer
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	currUser := user.UserFrom(r.Context())
	if currUser == nil {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeUnauthorized,
			Message: "unauthorized: missing user profile",
		})
		return
	}

	var req createExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, response.ErrInvalidBody, "invalid request body")
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
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, CreateExpenseResponse{
		Expense: exp,
		Splits:  splits,
	})
}

// Settle creates a payment record to clear or reduce debt.
// @Summary      Settle balance
// @Description  Create a settlement payment between two users.
// @Tags         expenses
// @Accept       json
// @Produce      json
// @Param        request body settleExpenseRequest true "Settlement details"
// @Success      201  {object}  SettleExpenseResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /expenses/settle [post]
// @Security     Bearer
func (h *Handler) Settle(w http.ResponseWriter, r *http.Request) {
	currUser := user.UserFrom(r.Context())
	if currUser == nil {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeUnauthorized,
			Message: "unauthorized: missing user profile",
		})
		return
	}

	var req settleExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, response.ErrInvalidBody, "invalid request body")
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
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, SettleExpenseResponse{
		Expense: exp,
		Split:   split,
	})
}

// List lists expenses based on filters (group, personal, or friend).
// @Summary      List expenses
// @Description  Retrieve a list of expenses filtered by group, personal=true, or friendId.
// @Tags         expenses
// @Produce      json
// @Param        groupId query string false "Filter by Group ID"
// @Param        personal query boolean false "Filter for personal only (true/false)"
// @Param        friendId query string false "Filter by Friend ID"
// @Success      200  {array}   Expense
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /expenses [get]
// @Security     Bearer
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	currUser := user.UserFrom(r.Context())
	if currUser == nil {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeUnauthorized,
			Message: "unauthorized: missing user profile",
		})
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
		response.HandleError(w, &response.AppError{
			Type:    response.TypeValidation,
			Message: "missing filter parameter: must supply groupId, personal=true, or friendId",
		})
		return
	}

	expenses, err := h.uc.ListExpenses(r.Context(), filterType, filterID, currUser.ID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, expenses)
}

// GetDetails retrieves a specific expense and its details.
// @Summary      Get expense details
// @Description  Get a specific expense's details including all splits.
// @Tags         expenses
// @Produce      json
// @Param        id path string true "Expense ID"
// @Success      200  {object}  GetExpenseDetailsResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /expenses/{id} [get]
// @Security     Bearer
func (h *Handler) GetDetails(w http.ResponseWriter, r *http.Request) {
	expenseID := chi.URLParam(r, "id")
	if expenseID == "" {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeValidation,
			Message: "expense id is required",
		})
		return
	}

	currUser := user.UserFrom(r.Context())
	if currUser == nil {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeUnauthorized,
			Message: "unauthorized: missing user profile",
		})
		return
	}

	exp, splits, err := h.uc.GetExpenseDetails(r.Context(), expenseID, currUser.ID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, GetExpenseDetailsResponse{
		Expense: exp,
		Splits:  splits,
	})
}

// Delete soft deletes an expense.
// @Summary      Delete expense
// @Description  Soft-delete an expense by ID.
// @Tags         expenses
// @Param        id path string true "Expense ID"
// @Success      204  "No Content"
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /expenses/{id} [delete]
// @Security     Bearer
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	expenseID := chi.URLParam(r, "id")
	if expenseID == "" {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeValidation,
			Message: "expense id is required",
		})
		return
	}

	currUser := user.UserFrom(r.Context())
	if currUser == nil {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeUnauthorized,
			Message: "unauthorized: missing user profile",
		})
		return
	}

	err := h.uc.DeleteExpense(r.Context(), expenseID, currUser.ID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetBalances calculates balances either inside a group or globally.
// @Summary      Get user balances
// @Description  Calculate net balances and recommended settlement transactions.
// @Tags         expenses
// @Produce      json
// @Param        groupId query string false "Filter by Group ID. If omitted, returns global balances."
// @Param        simplified query boolean false "Simplify debts algorithm (true/false)"
// @Success      200  {object}  BalanceResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /balances [get]
// @Security     Bearer
func (h *Handler) GetBalances(w http.ResponseWriter, r *http.Request) {
	currUser := user.UserFrom(r.Context())
	if currUser == nil {
		response.HandleError(w, &response.AppError{
			Type:    response.TypeUnauthorized,
			Message: "unauthorized: missing user profile",
		})
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
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, balances)
}
