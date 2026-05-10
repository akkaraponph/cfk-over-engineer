package balancesheet

import "time"

type BalanceSheet struct {
	ID        string
	TenantID  string
	UserID    string
	Year      int
	CreatedAt time.Time
	UpdatedAt time.Time
}

const (
	EventCreated = "balancesheet.created"
	EventUpdated = "balancesheet.updated"
)
