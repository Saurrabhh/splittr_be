package http

import (
	"context"
	"net/http"
	"strconv"

	"github.com/Saurrabhh/splittr_be/internal/expense/domain"
	"github.com/Saurrabhh/splittr_be/internal/pagination"
	"github.com/Saurrabhh/splittr_be/internal/request"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/Saurrabhh/splittr_be/internal/user"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for expenses and balances.
type Handler struct {
	uc *domain.UseCase
}

// NewHandler creates a new Handler.
func NewHandler(uc *domain.UseCase) *Handler {
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

// Create logs a new expense and distributes the splits.
// @Summary      Create expense
// @Description  Create a new expense with equal/exact/percentage splits.
// @Tags         expenses
// @Accept       json
// @Produce      json
// @Param        request body CreateExpenseRequest true "Expense details and splits structure"
// @Success      201  {object}  domain.ExpenseWithSplits
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /expenses [post]
// @Security     BearerAuth
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	currUser := user.MustFrom(r.Context())

	request.Run(w, r, http.StatusCreated, func(ctx context.Context, req CreateExpenseRequest) (ExpenseResponse, error) {
		paidBy := req.PaidBy
		if paidBy == "" {
			paidBy = currUser.ID
		}

		category := "Other"
		if req.Category != nil && *req.Category != "" {
			category = *req.Category
		}

		exp, splits, err := h.uc.CreateExpense(
			ctx,
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
			return ExpenseResponse{}, err
		}

		return ExpenseResponse{
			Expense: *exp,
			Splits:  splits,
		}, nil
	})
}

// Settle creates a payment record to clear or reduce debt.
// @Summary      Settle balance
// @Description  Create a settlement payment between two users.
// @Tags         expenses
// @Accept       json
// @Produce      json
// @Param        request body SettleExpenseRequest true "Settlement details"
// @Success      201  {object}  domain.ExpenseWithSplits
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /expenses/settle [post]
// @Security     BearerAuth
func (h *Handler) Settle(w http.ResponseWriter, r *http.Request) {
	currUser := user.MustFrom(r.Context())

	request.Run(w, r, http.StatusCreated, func(ctx context.Context, req SettleExpenseRequest) (ExpenseResponse, error) {
		paidBy := req.PaidBy
		if paidBy == "" {
			paidBy = currUser.ID
		}

		exp, split, err := h.uc.SettleUp(
			ctx,
			req.Amount,
			req.Currency,
			req.GroupID,
			paidBy,
			req.ReceivedBy,
			currUser.ID,
		)
		if err != nil {
			return ExpenseResponse{}, err
		}

		var splits []domain.Split
		if split != nil {
			splits = []domain.Split{*split}
		}
		return ExpenseResponse{
			Expense: *exp,
			Splits:  splits,
		}, nil
	})
}

// List lists expenses based on filters (group, personal, or friend).
// @Summary      List expenses
// @Description  Retrieve a cursor-paginated list of expenses with splits, filtered by group, personal=true, or friendId.
// @Tags         expenses
// @Produce      json
// @Param        groupId   query  string  false  "Filter by Group ID"
// @Param        personal  query  boolean false  "Filter for personal only (true/false)"
// @Param        friendId  query  string  false  "Filter by Friend ID"
// @Param        limit     query  int     false  "Items per page (max 100, default 20)"
// @Param        cursor    query  string  false  "Opaque cursor token from a previous response"
// @Success      200  {object}  pagination.Response[domain.ExpenseWithSplits]
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /expenses [get]
// @Security     BearerAuth
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	currUser := user.MustFrom(r.Context())

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
			Message: response.MsgMissingFilterParam,
		})
		return
	}

	p := pagination.ParseParams(r, 20, 100)
	result, err := h.uc.ListExpenses(r.Context(), filterType, filterID, currUser.ID, p)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, result)

}

// GetDetails retrieves a specific expense and its details.
// @Summary      Get expense details
// @Description  Get a specific expense's details including all splits.
// @Tags         expenses
// @Produce      json
// @Param        id path string true "Expense ID"
// @Success      200  {object}  domain.ExpenseWithSplits
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      403  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /expenses/{id} [get]
// @Security     BearerAuth
func (h *Handler) GetDetails(w http.ResponseWriter, r *http.Request) {
	expenseID, ok := request.URLParam(w, r, "id")
	if !ok {
		return
	}

	currUser := user.MustFrom(r.Context())

	exp, splits, err := h.uc.GetExpenseDetails(r.Context(), expenseID, currUser.ID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, ExpenseResponse{
		Expense: *exp,
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
// @Failure      403  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /expenses/{id} [delete]
// @Security     BearerAuth
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	expenseID, ok := request.URLParam(w, r, "id")
	if !ok {
		return
	}

	currUser := user.MustFrom(r.Context())

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
// @Success      200  {object}  domain.BalanceResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /balances [get]
// @Security     BearerAuth
func (h *Handler) GetBalances(w http.ResponseWriter, r *http.Request) {
	currUser := user.MustFrom(r.Context())

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
