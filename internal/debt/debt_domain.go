package debt

import "time"

type Debt struct {
	ID             string
	TenantID       string
	Type           string
	Description    string
	Amount         float64
	Interest       float64
	MinimumPay     float64
	Priority       int
	BalanceSheetID string
	UserID         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

const (
	EventRecorded                 = "debt.recorded"
	EventAmountChanged            = "debt.amount_changed"
	EventAssignedToBalanceSheet   = "debt.assigned_to_balancesheet"
	EventUnassignedFromBalanceSheet = "debt.unassigned_from_balancesheet"
)
