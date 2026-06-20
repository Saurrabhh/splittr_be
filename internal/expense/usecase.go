package expense

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/Saurrabhh/splittr_be/internal/activity"
	"github.com/Saurrabhh/splittr_be/internal/db"
	"github.com/Saurrabhh/splittr_be/internal/group"
	"github.com/Saurrabhh/splittr_be/internal/notification"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/google/uuid"
)

// Repository defines the storage contract for the expense domain.
type Repository interface {
	CreateExpense(ctx context.Context, e *Expense) error
	CreateExpenseSplit(ctx context.Context, s *ExpenseSplit) error
	GetExpenseByID(ctx context.Context, id string) (*Expense, error)
	ListExpenseSplits(ctx context.Context, expenseID string) ([]ExpenseSplit, error)
	ListExpensesByGroup(ctx context.Context, groupID string) ([]Expense, error)
	ListUserPersonalExpenses(ctx context.Context, userID string) ([]Expense, error)
	ListUserFriendExpenses(ctx context.Context, userID string) ([]Expense, error)
	DeleteExpense(ctx context.Context, id string) error
	GetGroupBalances(ctx context.Context, groupID string) ([]UserBalance, error)
	GetFriendBalances(ctx context.Context, userID string) ([]UserBalance, error)
	GetGroupPairwiseDebts(ctx context.Context, groupID string) ([]PairwiseDebt, error)
}

// GroupService defines the contract required to validate group membership.
type GroupService interface {
	GetGroupDetails(ctx context.Context, groupID, userID string) (*group.Group, []group.GroupMember, error)
}

type ActivityLogger interface {
	LogActivity(ctx context.Context, actorID string, groupID *string, actionType string, description string, visibleToUserIDs []string) (*activity.Activity, error)
}

type NotificationSender interface {
	CreateAlert(ctx context.Context, userID string, actorID *string, activityID *string, title, content string) (*notification.Notification, error)
}

// Usecase manages business logic for expenses, splits, and balances.
type Usecase struct {
	repo         Repository
	tx           db.Transactor
	groupSvc     GroupService
	activity     ActivityLogger
	notification NotificationSender
}

// NewUsecase instantiates a new Usecase.
func NewUsecase(repo Repository, tx db.Transactor, groupSvc GroupService, activitySvc ActivityLogger, notificationSvc NotificationSender) *Usecase {
	return &Usecase{
		repo:         repo,
		tx:           tx,
		groupSvc:     groupSvc,
		activity:     activitySvc,
		notification: notificationSvc,
	}
}

// CreateExpense calculates splits, validates constraints, and inserts the expense inside a transaction.
func (u *Usecase) CreateExpense(ctx context.Context, desc string, amount float64, currency string, category string, groupID *string, paidBy string, splitType SplitType, inputs []InputSplit, createdBy string) (*Expense, []ExpenseSplit, error) {
	if desc == "" {
		return nil, nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: "description is required",
		}
	}
	if amount <= 0 {
		return nil, nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: "amount must be greater than zero",
		}
	}
	if len(inputs) == 0 {
		return nil, nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: "expense must be split with at least one user",
		}
	}
	if currency == "" {
		currency = "INR"
	}

	// 1. Group validation & Membership Access Control
	if groupID != nil && *groupID != "" {
		_, members, err := u.groupSvc.GetGroupDetails(ctx, *groupID, createdBy)
		if err != nil {
			return nil, nil, err // bubble up group access validation error
		}

		// Ensure paidBy is in the group
		memberMap := make(map[string]bool)
		for _, m := range members {
			memberMap[m.UserID] = true
		}

		if !memberMap[paidBy] {
			return nil, nil, &response.AppError{
				Type:    response.TypeValidation,
				Message: "payer must be a member of the group",
			}
		}

		// Ensure all split users are in the group
		for _, split := range inputs {
			if !memberMap[split.UserID] {
				return nil, nil, &response.AppError{
					Type:    response.TypeValidation,
					Message: fmt.Sprintf("split user %s is not a member of the group", split.UserID),
				}
			}
		}
	}

	// 2. Perform Split Calculations (using cent-level integers to avoid floats precision bugs)
	calculatedSplits, err := calculateSplits(amount, splitType, inputs)
	if err != nil {
		return nil, nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: err.Error(),
		}
	}

	newExpense := &Expense{
		ID:          uuid.New().String(),
		Description: desc,
		Amount:      amount,
		Currency:    currency,
		Category:    category,
		GroupID:     groupID,
		PaidBy:      paidBy,
		CreatedBy:   createdBy,
		IsPayment:   false,
		SpentAt:     time.Now(),
	}

	// 3. Persist everything atomically inside a transaction
	err = u.tx.RunInTx(ctx, func(txCtx context.Context) error {
		if err := u.repo.CreateExpense(txCtx, newExpense); err != nil {
			return err
		}

		for _, split := range calculatedSplits {
			split.ExpenseID = newExpense.ID
			split.SplitType = splitType
			if err := u.repo.CreateExpenseSplit(txCtx, &split); err != nil {
				return err
			}
		}

		// Log Activity
		activityDesc := fmt.Sprintf("added expense '%s' of %.2f %s", desc, amount, currency)
		var visibleTo []string
		if groupID == nil || *groupID == "" {
			visibilityMap := make(map[string]bool)
			visibilityMap[paidBy] = true
			visibilityMap[createdBy] = true
			for _, sp := range inputs {
				visibilityMap[sp.UserID] = true
			}
			visibleTo = make([]string, 0, len(visibilityMap))
			for uID := range visibilityMap {
				visibleTo = append(visibleTo, uID)
			}
		}

		act, err := u.activity.LogActivity(txCtx, createdBy, groupID, "EXPENSE_CREATED", activityDesc, visibleTo)
		if err != nil {
			return err
		}

		// Trigger Notifications
		notificationTitle := "New Expense"
		notificationContent := fmt.Sprintf("New expense '%s' of %.2f %s added", desc, amount, currency)

		if groupID != nil && *groupID != "" {
			_, members, err := u.groupSvc.GetGroupDetails(txCtx, *groupID, createdBy)
			if err == nil {
				for _, m := range members {
					if m.UserID != createdBy {
						_, _ = u.notification.CreateAlert(txCtx, m.UserID, &createdBy, &act.ID, notificationTitle, notificationContent)
					}
				}
			}
		} else {
			for _, sp := range inputs {
				if sp.UserID != createdBy {
					_, _ = u.notification.CreateAlert(txCtx, sp.UserID, &createdBy, &act.ID, notificationTitle, notificationContent)
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: "create expense transaction failed",
			Err:     err,
		}
	}

	// Fetch enriched splits with user names and profiles for the response
	enrichedSplits, err := u.repo.ListExpenseSplits(ctx, newExpense.ID)
	if err != nil {
		return newExpense, nil, nil // Return expense anyway if only list fails
	}

	return newExpense, enrichedSplits, nil
}

// SettleUp creates a payment record to clear or reduce debt between a payer and a payee.
func (u *Usecase) SettleUp(ctx context.Context, amount float64, currency string, groupID *string, paidBy string, receivedBy string, createdBy string) (*Expense, *ExpenseSplit, error) {
	if amount <= 0 {
		return nil, nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: "settlement amount must be greater than zero",
		}
	}
	if paidBy == receivedBy {
		return nil, nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: "payer and payee must be different users",
		}
	}
	if currency == "" {
		currency = "INR"
	}

	// Group validation & Membership Access Control
	if groupID != nil && *groupID != "" {
		_, members, err := u.groupSvc.GetGroupDetails(ctx, *groupID, createdBy)
		if err != nil {
			return nil, nil, err // bubble up group access validation error
		}

		memberMap := make(map[string]bool)
		for _, m := range members {
			memberMap[m.UserID] = true
		}

		if !memberMap[paidBy] || !memberMap[receivedBy] {
			return nil, nil, &response.AppError{
				Type:    response.TypeValidation,
				Message: "both payer and payee must be members of the group",
			}
		}
	}

	newExpense := &Expense{
		ID:          uuid.New().String(),
		Description: "Settle Up",
		Amount:      amount,
		Currency:    currency,
		Category:    "Payment",
		GroupID:     groupID,
		PaidBy:      paidBy,
		CreatedBy:   createdBy,
		IsPayment:   true,
		SpentAt:     time.Now(),
	}

	split := &ExpenseSplit{
		ExpenseID: newExpense.ID,
		UserID:    receivedBy,
		Amount:    amount,
		SplitType: SplitTypeExact,
	}

	err := u.tx.RunInTx(ctx, func(txCtx context.Context) error {
		if err := u.repo.CreateExpense(txCtx, newExpense); err != nil {
			return err
		}
		if err := u.repo.CreateExpenseSplit(txCtx, split); err != nil {
			return err
		}

		// Log Activity
		activityDesc := fmt.Sprintf("settled %.2f %s", amount, currency)
		var visibleTo []string
		if groupID == nil || *groupID == "" {
			visibleTo = []string{paidBy, receivedBy, createdBy}
		}

		act, err := u.activity.LogActivity(txCtx, createdBy, groupID, "SETTLEMENT", activityDesc, visibleTo)
		if err != nil {
			return err
		}

		// Notify payee
		_, _ = u.notification.CreateAlert(
			txCtx,
			receivedBy,
			&paidBy,
			&act.ID,
			"Payment Received",
			fmt.Sprintf("Payment of %.2f %s received", amount, currency),
		)

		return nil
	})
	if err != nil {
		return nil, nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: "settle up transaction failed",
			Err:     err,
		}
	}

	// Fetch enriched splits details
	splits, err := u.repo.ListExpenseSplits(ctx, newExpense.ID)
	if err == nil && len(splits) > 0 {
		return newExpense, &splits[0], nil
	}

	return newExpense, split, nil
}

// GetExpenseDetails retrieves an expense and its splits, checking view permissions.
func (u *Usecase) GetExpenseDetails(ctx context.Context, expenseID, userID string) (*Expense, []ExpenseSplit, error) {
	e, err := u.repo.GetExpenseByID(ctx, expenseID)
	if err != nil {
		return nil, nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to retrieve expense details",
			Err:     err,
		}
	}
	if e == nil {
		return nil, nil, &response.AppError{
			Type:    response.TypeNotFound,
			Message: "expense not found",
		}
	}

	splits, err := u.repo.ListExpenseSplits(ctx, expenseID)
	if err != nil {
		return nil, nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to retrieve expense splits",
			Err:     err,
		}
	}

	// Access control check: requester must be paidBy, creator, or a split participant
	hasAccess := e.PaidBy == userID || e.CreatedBy == userID
	if !hasAccess {
		for _, s := range splits {
			if s.UserID == userID {
				hasAccess = true
				break
			}
		}
	}

	if !hasAccess {
		return nil, nil, &response.AppError{
			Type:    response.TypeForbidden,
			Message: "access denied: not a participant of this expense",
		}
	}

	return e, splits, nil
}

// ListExpenses returns a list of expenses filtered by group or personal type.
func (u *Usecase) ListExpenses(ctx context.Context, filterType, filterID, userID string) ([]Expense, error) {
	if filterType == "group" {
		_, _, err := u.groupSvc.GetGroupDetails(ctx, filterID, userID)
		if err != nil {
			return nil, err // bubble up group access validation error
		}
		expenses, err := u.repo.ListExpensesByGroup(ctx, filterID)
		if err != nil {
			return nil, &response.AppError{
				Type:    response.TypeInternal,
				Message: "failed to list group expenses",
				Err:     err,
			}
		}
		return expenses, nil
	}

	if filterType == "personal" {
		expenses, err := u.repo.ListUserPersonalExpenses(ctx, userID)
		if err != nil {
			return nil, &response.AppError{
				Type:    response.TypeInternal,
				Message: "failed to list personal expenses",
				Err:     err,
			}
		}
		return expenses, nil
	}

	if filterType == "friend" {
		expenses, err := u.repo.ListUserFriendExpenses(ctx, userID)
		if err != nil {
			return nil, &response.AppError{
				Type:    response.TypeInternal,
				Message: "failed to list friend expenses",
				Err:     err,
			}
		}
		return expenses, nil
	}

	return nil, &response.AppError{
		Type:    response.TypeValidation,
		Message: "invalid filter type: must be group, personal, or friend",
	}
}

// GetBalances returns direct or group balances and recommended settlements.
func (u *Usecase) GetBalances(ctx context.Context, groupID *string, userID string, simplified bool) (*BalanceResponse, error) {
	if groupID != nil && *groupID != "" {
		_, _, err := u.groupSvc.GetGroupDetails(ctx, *groupID, userID)
		if err != nil {
			return nil, err // bubble up group access validation error
		}

		balances, err := u.repo.GetGroupBalances(ctx, *groupID)
		if err != nil {
			return nil, &response.AppError{
				Type:    response.TypeInternal,
				Message: "failed to calculate group balances",
				Err:     err,
			}
		}

		var settlements []Settlement
		if simplified {
			settlements = simplifyDebts(balances)
		} else {
			pairwise, err := u.repo.GetGroupPairwiseDebts(ctx, *groupID)
			if err != nil {
				return nil, &response.AppError{
					Type:    response.TypeInternal,
					Message: "failed to calculate pairwise debts",
					Err:     err,
				}
			}
			settlements = directDebts(pairwise)
		}

		return &BalanceResponse{
			Balances:    balances,
			Settlements: settlements,
		}, nil
	}

	balances, err := u.repo.GetFriendBalances(ctx, userID)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to calculate friend balances",
			Err:     err,
		}
	}

	// Map direct friend balances to settlements
	settlements := make([]Settlement, 0)
	for _, b := range balances {
		cents := int64(math.Round(b.NetBalance * 100))
		if cents > 0 {
			settlements = append(settlements, Settlement{
				FromUserID:   b.UserID,
				FromUserName: b.UserName,
				ToUserID:     userID,
				ToUserName:   "You",
				Amount:       float64(cents) / 100.0,
			})
		} else if cents < 0 {
			settlements = append(settlements, Settlement{
				FromUserID:   userID,
				FromUserName: "You",
				ToUserID:     b.UserID,
				ToUserName:   b.UserName,
				Amount:       float64(-cents) / 100.0,
			})
		}
	}

	return &BalanceResponse{
		Balances:    balances,
		Settlements: settlements,
	}, nil
}

// DeleteExpense soft deletes the expense record. Only the creator of the expense can do this.
func (u *Usecase) DeleteExpense(ctx context.Context, expenseID, userID string) error {
	e, err := u.repo.GetExpenseByID(ctx, expenseID)
	if err != nil {
		return &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to retrieve expense details",
			Err:     err,
		}
	}
	if e == nil {
		return &response.AppError{
			Type:    response.TypeNotFound,
			Message: "expense not found",
		}
	}

	if e.CreatedBy != userID {
		return &response.AppError{
			Type:    response.TypeForbidden,
			Message: "unauthorized: only the creator can delete this expense",
		}
	}

	if err := u.repo.DeleteExpense(ctx, expenseID); err != nil {
		return &response.AppError{
			Type:    response.TypeInternal,
			Message: "failed to delete expense",
			Err:     err,
		}
	}
	return nil
}

// Helper to compute dynamic splits
func calculateSplits(totalAmount float64, splitType SplitType, inputs []InputSplit) ([]ExpenseSplit, error) {
	totalCents := int64(math.Round(totalAmount * 100))
	splits := make([]ExpenseSplit, 0, len(inputs))

	switch splitType {
	case SplitTypeEqual:
		numSplits := int64(len(inputs))
		baseCents := totalCents / numSplits
		remainder := totalCents % numSplits

		for i, in := range inputs {
			shareCents := baseCents
			// Distribute remainder pennies to the first few users
			if int64(i) < remainder {
				shareCents++
			}
			splits = append(splits, ExpenseSplit{
				UserID: in.UserID,
				Amount: float64(shareCents) / 100.0,
			})
		}

	case SplitTypeExact:
		var sumCents int64
		for _, in := range inputs {
			if in.Amount == nil {
				return nil, errors.New("amount is required for each user in exact split")
			}
			amtCents := int64(math.Round(*in.Amount * 100))
			sumCents += amtCents

			splits = append(splits, ExpenseSplit{
				UserID: in.UserID,
				Amount: *in.Amount,
			})
		}

		if sumCents != totalCents {
			return nil, fmt.Errorf("sum of splits (%.2f) does not match total expense amount (%.2f)", float64(sumCents)/100.0, totalAmount)
		}

	case SplitTypePercentage:
		var sumPercent float64
		for _, in := range inputs {
			if in.Percentage == nil {
				return nil, errors.New("percentage is required for each user in percentage split")
			}
			sumPercent += *in.Percentage
		}

		if math.Abs(sumPercent-100.0) > 0.01 {
			return nil, fmt.Errorf("sum of split percentages (%.2f%%) must equal 100%%", sumPercent)
		}

		allocatedCents := int64(0)
		for i, in := range inputs {
			// Calculate share in cents
			shareCents := int64(math.Round((totalAmount * (*in.Percentage)) / 100.0 * 100.0))

			// Adjust the last element for decimal rounding errors
			if i == len(inputs)-1 {
				shareCents = totalCents - allocatedCents
			} else {
				allocatedCents += shareCents
			}

			valCopy := *in.Percentage
			splits = append(splits, ExpenseSplit{
				UserID:     in.UserID,
				Amount:     float64(shareCents) / 100.0,
				SplitValue: &valCopy,
			})
		}

	default:
		return nil, fmt.Errorf("unsupported split type: %s", splitType)
	}

	return splits, nil
}

type person struct {
	id   string
	name string
	bal  int64
}

// simplifyDebts calculates minimized settlement transactions using a greedy max-flow matching algorithm.
func simplifyDebts(balances []UserBalance) []Settlement {
	var debtors []person
	var creditors []person

	for _, b := range balances {
		cents := int64(math.Round(b.NetBalance * 100))
		if cents < 0 {
			debtors = append(debtors, person{id: b.UserID, name: b.UserName, bal: -cents})
		} else if cents > 0 {
			creditors = append(creditors, person{id: b.UserID, name: b.UserName, bal: cents})
		}
	}

	var settlements []Settlement

	for len(debtors) > 0 && len(creditors) > 0 {
		debtorIdx := findMaxIdx(debtors)
		creditorIdx := findMaxIdx(creditors)

		d := &debtors[debtorIdx]
		c := &creditors[creditorIdx]

		amountCents := d.bal
		if c.bal < amountCents {
			amountCents = c.bal
		}

		if amountCents > 0 {
			settlements = append(settlements, Settlement{
				FromUserID:   d.id,
				FromUserName: d.name,
				ToUserID:     c.id,
				ToUserName:   c.name,
				Amount:       float64(amountCents) / 100.0,
			})
		}

		d.bal -= amountCents
		c.bal -= amountCents

		newDebtors := make([]person, 0, len(debtors))
		for _, val := range debtors {
			if val.bal > 0 {
				newDebtors = append(newDebtors, val)
			}
		}
		debtors = newDebtors

		newCreditors := make([]person, 0, len(creditors))
		for _, val := range creditors {
			if val.bal > 0 {
				newCreditors = append(newCreditors, val)
			}
		}
		creditors = newCreditors
	}

	return settlements
}

func findMaxIdx(list []person) int {
	maxVal := int64(-1)
	maxIdx := 0
	for i, p := range list {
		if p.bal > maxVal {
			maxVal = p.bal
			maxIdx = i
		}
	}
	return maxIdx
}

// directDebts nets off pairwise splits to calculate direct debts.
func directDebts(pairwise []PairwiseDebt) []Settlement {
	type pair struct {
		userA     string
		userAName string
		userB     string
		userBName string
	}

	netBalances := make(map[string]int64)
	names := make(map[string]pair)

	for _, pd := range pairwise {
		uA, uB := pd.DebtorID, pd.CreditorID
		uAName, uBName := pd.DebtorName, pd.CreditorName

		isOrder := uA < uB
		var key string
		if isOrder {
			key = uA + "_" + uB
		} else {
			key = uB + "_" + uA
		}

		cents := int64(math.Round(pd.Amount * 100))
		if isOrder {
			netBalances[key] -= cents
		} else {
			netBalances[key] += cents
		}

		if isOrder {
			names[key] = pair{userA: uA, userAName: uAName, userB: uB, userBName: uBName}
		} else {
			names[key] = pair{userA: uB, userAName: uBName, userB: uA, userBName: uAName}
		}
	}

	var settlements []Settlement
	for key, balCents := range netBalances {
		p := names[key]
		if balCents > 0 {
			settlements = append(settlements, Settlement{
				FromUserID:   p.userB,
				FromUserName: p.userBName,
				ToUserID:     p.userA,
				ToUserName:   p.userAName,
				Amount:       float64(balCents) / 100.0,
			})
		} else if balCents < 0 {
			settlements = append(settlements, Settlement{
				FromUserID:   p.userA,
				FromUserName: p.userAName,
				ToUserID:     p.userB,
				ToUserName:   p.userBName,
				Amount:       float64(-balCents) / 100.0,
			})
		}
	}

	return settlements
}
