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
	Version        int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

const (
	EventRecorded                   = "debt.recorded"
	EventAmountChanged              = "debt.amount_changed"
	EventAssignedToBalanceSheet     = "debt.assigned_to_balancesheet"
	EventUnassignedFromBalanceSheet = "debt.unassigned_from_balancesheet"
)

type DebtRecordedPayload struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id"`
	Type           string    `json:"type"`
	Description    string    `json:"description"`
	Amount         float64   `json:"amount"`
	Interest       float64   `json:"interest"`
	MinimumPay     float64   `json:"minimum_pay"`
	Priority       int       `json:"priority"`
	BalanceSheetID string    `json:"balance_sheet_id"`
	UserID         string    `json:"user_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type DebtAmountChangedPayload struct {
	ID        string    `json:"id"`
	Amount    float64   `json:"amount"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DebtAssignedToBalanceSheetPayload struct {
	ID             string    `json:"id"`
	BalanceSheetID string    `json:"balance_sheet_id"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type DebtUnassignedFromBalanceSheetPayload struct {
	ID        string    `json:"id"`
	UpdatedAt time.Time `json:"updated_at"`
}
