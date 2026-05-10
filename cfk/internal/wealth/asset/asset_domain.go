package asset

import "time"

type Asset struct {
	ID              string
	TenantID        string
	Type            string
	Description     string
	Value           float64
	CashflowPerYear float64
	BalanceSheetID  string
	UserID          string
	Version         int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

const (
	EventRecorded                   = "asset.recorded"
	EventValueChanged               = "asset.value_changed"
	EventAssignedToBalanceSheet     = "asset.assigned_to_balancesheet"
	EventUnassignedFromBalanceSheet = "asset.unassigned_from_balancesheet"
)

type AssetRecordedPayload struct {
	ID              string    `json:"id"`
	TenantID        string    `json:"tenant_id"`
	Type            string    `json:"type"`
	Description     string    `json:"description"`
	Value           float64   `json:"value"`
	CashflowPerYear float64   `json:"cashflow_per_year"`
	BalanceSheetID  string    `json:"balance_sheet_id"`
	UserID          string    `json:"user_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type AssetValueChangedPayload struct {
	ID              string    `json:"id"`
	Value           float64   `json:"value"`
	CashflowPerYear float64   `json:"cashflow_per_year"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type AssetAssignedToBalanceSheetPayload struct {
	ID              string    `json:"id"`
	BalanceSheetID  string    `json:"balance_sheet_id"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type AssetUnassignedFromBalanceSheetPayload struct {
	ID        string    `json:"id"`
	UpdatedAt time.Time `json:"updated_at"`
}
