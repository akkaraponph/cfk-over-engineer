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
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

const (
	EventRecorded                 = "asset.recorded"
	EventValueChanged             = "asset.value_changed"
	EventAssignedToBalanceSheet   = "asset.assigned_to_balancesheet"
	EventUnassignedFromBalanceSheet = "asset.unassigned_from_balancesheet"
)
