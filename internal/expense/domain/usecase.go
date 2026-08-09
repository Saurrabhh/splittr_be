package domain

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
	"github.com/Saurrabhh/splittr_be/internal/pagination"
	"github.com/Saurrabhh/splittr_be/internal/response"
	"github.com/google/uuid"
)

// GroupService defines the contract required to validate group membership.
type GroupService interface {
	GetGroupDetails(ctx context.Context, groupID, userID string) (*group.Group, error)
}

type ActivityLogger interface {
	LogEvent(
		ctx context.Context,
		actorID string,
		groupID *string,
		visibleToUserIDs []string,
		event activity.Event,
	) (*activity.Activity, error)
}

type NotificationSender interface {
	CreateAlert(ctx context.Context, userID string, actorID *string, activityID *string, alert notification.Alert) error
}

// BalanceResponse contains a list of member balances and a list of recommended settlement transactions.
type BalanceResponse struct {
	Balances    []UserBalance `json:"balances"`
	Settlements []Settlement  `json:"settlements"`
} // @name BalanceResponse

// UseCase manages business logic for expenses, splits, and balances.
type UseCase struct {
	repo         Repository
	tx           db.Transactor
	groupSvc     GroupService
	activity     ActivityLogger
	notification NotificationSender
}

// NewUseCase instantiates a new UseCase.
func NewUseCase(repo Repository, tx db.Transactor, groupSvc GroupService, activitySvc ActivityLogger, notificationSvc NotificationSender) *UseCase {
	return &UseCase{
		repo:         repo,
		tx:           tx,
		groupSvc:     groupSvc,
		activity:     activitySvc,
		notification: notificationSvc,
	}
}

// CreateExpense calculates splits, validates constraints, and inserts the expense inside a transaction.
func (u *UseCase) CreateExpense(ctx context.Context, desc string, amount float64, currency string, category string, groupID *string, paidBy string, splitType SplitType, inputs []InputSplit, createdBy string) (*ExpenseWithSplits, error) {
	if desc == "" {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgMissingDescription,
		}
	}
	if amount <= 0 {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgInvalidAmount,
		}
	}
	if len(inputs) == 0 {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgInvalidSplit,
		}
	}
	if currency == "" {
		currency = "INR"
	}

	var groupMembers []group.Member
	if groupID != nil && *groupID != "" {
		g, err := u.groupSvc.GetGroupDetails(ctx, *groupID, createdBy)
		if err != nil {
			return nil, err
		}
		groupMembers = g.Members

		memberMap := make(map[string]bool)
		for _, m := range groupMembers {
			memberMap[m.UserID] = true
		}

		if !memberMap[paidBy] {
			return nil, &response.AppError{
				Type:    response.TypeValidation,
				Message: response.MsgPayerNotGroupMember,
			}
		}

		for _, split := range inputs {
			if !memberMap[split.UserID] {
				return nil, &response.AppError{
					Type:    response.TypeValidation,
					Message: response.MsgSplitUserNotMember,
				}
			}
		}
	}

	calculatedSplits, err := calculateSplits(amount, splitType, inputs)
	if err != nil {
		return nil, &response.AppError{
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

		enrichedSplits, err := u.repo.ListExpenseSplits(txCtx, newExpense.ID)
		if err != nil {
			return err
		}

		actPayload := activity.ExpensePayload{
			Expense: newExpense,
			Splits:  enrichedSplits,
		}
		actEvent := activity.NewExpenseCreatedEvent(newExpense.ID, actPayload, activityDesc)
		act, err := u.activity.LogEvent(
			txCtx, createdBy, groupID, visibleTo, actEvent,
		)
		if err != nil {
			return err
		}

		expenseAlert := notification.NewExpenseAddedAlert(desc, amount, currency)

		if groupID != nil && *groupID != "" {
			for _, m := range groupMembers {
				if m.UserID != createdBy {
					_ = u.notification.CreateAlert(txCtx, m.UserID, &createdBy, &act.ID, expenseAlert)
				}
			}
		} else {
			for _, sp := range inputs {
				if sp.UserID != createdBy {
					_ = u.notification.CreateAlert(txCtx, sp.UserID, &createdBy, &act.ID, expenseAlert)
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogCreateExpenseTx,
			Err:     err,
		}
	}

	enrichedSplits, err := u.repo.ListExpenseSplits(ctx, newExpense.ID)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogLoadSplits,
			Err:     err,
		}
	}

	return &ExpenseWithSplits{
		Expense: *newExpense,
		Splits:  enrichedSplits,
	}, nil
}

// SettleUp creates a payment record to clear or reduce debt between a payer and a payee.
func (u *UseCase) SettleUp(ctx context.Context, amount float64, currency string, groupID *string, paidBy string, receivedBy string, createdBy string) (*ExpenseWithSplits, error) {
	if amount <= 0 {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgInvalidAmount,
		}
	}
	if receivedBy == "" {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgMissingRecipient,
		}
	}
	if paidBy == receivedBy {
		return nil, &response.AppError{
			Type:    response.TypeValidation,
			Message: response.MsgSamePayerPayee,
		}
	}
	if currency == "" {
		currency = "INR"
	}

	if groupID != nil && *groupID != "" {
		g, err := u.groupSvc.GetGroupDetails(ctx, *groupID, createdBy)
		if err != nil {
			return nil, err
		}

		memberMap := make(map[string]bool)
		for _, m := range g.Members {
			memberMap[m.UserID] = true
		}

		if !memberMap[paidBy] || !memberMap[receivedBy] {
			return nil, &response.AppError{
				Type:    response.TypeValidation,
				Message: response.MsgPayerPayeeGroupMember,
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

	split := &Split{
		ExpenseID: newExpense.ID,
		UserID:    receivedBy,
		Amount:    amount,
		SplitType: SplitTypeExact,
	}

	var finalSplit *Split

	err := u.tx.RunInTx(ctx, func(txCtx context.Context) error {
		if err := u.repo.CreateExpense(txCtx, newExpense); err != nil {
			return err
		}
		if err := u.repo.CreateExpenseSplit(txCtx, split); err != nil {
			return err
		}

		activityDesc := fmt.Sprintf("settled %.2f %s", amount, currency)
		var visibleTo []string
		if groupID == nil || *groupID == "" {
			visibleTo = []string{paidBy, receivedBy, createdBy}
		}

		enrichedSplits, err := u.repo.ListExpenseSplits(txCtx, newExpense.ID)
		if err != nil {
			return err
		}
		if len(enrichedSplits) > 0 {
			finalSplit = &enrichedSplits[0]
		} else {
			finalSplit = split
		}

		actPayload := activity.SettlementPayload{
			Expense: newExpense,
			Split:   finalSplit,
		}
		actEvent := activity.NewSettlementCreatedEvent(newExpense.ID, actPayload, activityDesc)
		act, err := u.activity.LogEvent(
			txCtx, createdBy, groupID, visibleTo, actEvent,
		)
		if err != nil {
			return err
		}

		_ = u.notification.CreateAlert(
			txCtx,
			receivedBy,
			&paidBy,
			&act.ID,
			notification.NewPaymentReceivedAlert(amount, currency),
		)

		return nil
	})
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogSettleUpTx,
			Err:     err,
		}
	}

	var splits []Split
	if finalSplit != nil {
		splits = []Split{*finalSplit}
	}

	return &ExpenseWithSplits{
		Expense: *newExpense,
		Splits:  splits,
	}, nil
}

// GetExpenseDetails retrieves an expense and its splits, checking view permissions.
func (u *UseCase) GetExpenseDetails(ctx context.Context, expenseID, userID string) (*ExpenseWithSplits, error) {
	e, err := u.repo.GetExpenseByID(ctx, expenseID)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogRetrieveExpense,
			Err:     err,
		}
	}
	if e == nil {
		return nil, &response.AppError{
			Type:    response.TypeNotFound,
			Message: response.MsgExpenseNotFound,
		}
	}

	splits, err := u.repo.ListExpenseSplits(ctx, expenseID)
	if err != nil {
		return nil, &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogRetrieveSplits,
			Err:     err,
		}
	}

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
		return nil, &response.AppError{
			Type:    response.TypeForbidden,
			Message: response.MsgNotExpenseParticipant,
		}
	}

	return &ExpenseWithSplits{
		Expense: *e,
		Splits:  splits,
	}, nil
}

// ListExpenses returns a cursor-paginated list of expenses filtered by group, personal, or friend type.
func (u *UseCase) ListExpenses(ctx context.Context, filterType, filterID, userID string, p pagination.Params) (pagination.Response[ExpenseWithSplits], error) {
	cursor := pagination.ParseCursor(p.Cursor)
	encodeFn := func(e ExpenseWithSplits) string { return pagination.EncodeCursor(e.Expense.CreatedAt, e.Expense.ID) }

	var expenses []Expense
	var err error

	switch filterType {
	case "group":
		_, err = u.groupSvc.GetGroupDetails(ctx, filterID, userID)
		if err != nil {
			return pagination.Response[ExpenseWithSplits]{}, err
		}
		expenses, err = u.repo.ListExpensesByGroup(ctx, filterID, p.Limit+1, cursor.LastTime, cursor.LastID)
		if err != nil {
			return pagination.Response[ExpenseWithSplits]{}, &response.AppError{Type: response.TypeInternal, Message: response.ErrLogListGroupExpenses, Err: err}
		}
	case "personal":
		expenses, err = u.repo.ListUserPersonalExpenses(ctx, userID, p.Limit+1, cursor.LastTime, cursor.LastID)
		if err != nil {
			return pagination.Response[ExpenseWithSplits]{}, &response.AppError{Type: response.TypeInternal, Message: response.ErrLogListPersonalExp, Err: err}
		}
	case "friend":
		expenses, err = u.repo.ListUserFriendExpenses(ctx, userID, p.Limit+1, cursor.LastTime, cursor.LastID)
		if err != nil {
			return pagination.Response[ExpenseWithSplits]{}, &response.AppError{Type: response.TypeInternal, Message: response.ErrLogListFriendExp, Err: err}
		}
	default:
		return pagination.Response[ExpenseWithSplits]{}, &response.AppError{Type: response.TypeValidation, Message: response.MsgInvalidExpenseFilter}
	}

	// Bulk-fetch all splits in a single query — always 2 DB round-trips total, never N+1.
	ids := make([]string, len(expenses))
	for i, e := range expenses {
		ids[i] = e.ID
	}
	allSplits, err := u.repo.ListExpenseSplitsByIDs(ctx, ids)
	if err != nil {
		return pagination.Response[ExpenseWithSplits]{}, &response.AppError{Type: response.TypeInternal, Message: response.ErrLogLoadSplits, Err: err}
	}

	// Group splits by expense ID.
	splitsByExpense := make(map[string][]Split, len(expenses))
	for _, s := range allSplits {
		splitsByExpense[s.ExpenseID] = append(splitsByExpense[s.ExpenseID], s)
	}

	rows := make([]ExpenseWithSplits, len(expenses))
	for i, e := range expenses {
		rows[i] = ExpenseWithSplits{
			Expense: e,
			Splits:  splitsByExpense[e.ID],
		}
	}

	return pagination.BuildResponse(rows, p.Limit, encodeFn), nil
}

// GetBalances returns direct or group balances and recommended settlements.
func (u *UseCase) GetBalances(ctx context.Context, groupID *string, userID string, simplified bool) (*BalanceResponse, error) {
	if groupID != nil && *groupID != "" {
		_, err := u.groupSvc.GetGroupDetails(ctx, *groupID, userID)
		if err != nil {
			return nil, err
		}

		balances, err := u.repo.GetGroupBalances(ctx, *groupID)
		if err != nil {
			return nil, &response.AppError{
				Type:    response.TypeInternal,
				Message: response.ErrLogCalcGroupBalances,
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
					Message: response.ErrLogCalcPairwiseDebts,
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
			Message: response.ErrLogCalcFriendBalances,
			Err:     err,
		}
	}

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
func (u *UseCase) DeleteExpense(ctx context.Context, expenseID, userID string) error {
	e, err := u.repo.GetExpenseByID(ctx, expenseID)
	if err != nil {
		return &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogRetrieveExpense,
			Err:     err,
		}
	}
	if e == nil {
		return &response.AppError{
			Type:    response.TypeNotFound,
			Message: response.MsgExpenseNotFound,
		}
	}

	if e.CreatedBy != userID {
		return &response.AppError{
			Type:    response.TypeForbidden,
			Message: response.MsgExpenseCreatorOnly,
		}
	}

	if err := u.repo.DeleteExpense(ctx, expenseID); err != nil {
		return &response.AppError{
			Type:    response.TypeInternal,
			Message: response.ErrLogDeleteExpense,
			Err:     err,
		}
	}
	return nil
}

func calculateSplits(totalAmount float64, splitType SplitType, inputs []InputSplit) ([]Split, error) {
	totalCents := int64(math.Round(totalAmount * 100))
	splits := make([]Split, 0, len(inputs))

	switch splitType {
	case SplitTypeEqual:
		numSplits := int64(len(inputs))
		baseCents := totalCents / numSplits
		remainder := totalCents % numSplits

		for i, in := range inputs {
			shareCents := baseCents
			if int64(i) < remainder {
				shareCents++
			}
			splits = append(splits, Split{
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

			splits = append(splits, Split{
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
			shareCents := int64(math.Round((totalAmount * (*in.Percentage)) / 100.0 * 100.0))

			if i == len(inputs)-1 {
				shareCents = totalCents - allocatedCents
			} else {
				allocatedCents += shareCents
			}

			splits = append(splits, Split{
				UserID:     in.UserID,
				Amount:     float64(shareCents) / 100.0,
				SplitValue: in.Percentage,
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

		amountCents := min(c.bal, d.bal)

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
