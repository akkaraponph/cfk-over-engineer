package balancesheet

import "time"

type BalanceSheet struct {
	ID        string
	TenantID  string
	UserID    string
	Year      int
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
}

const (
	EventCreated = "balancesheet.created"
	EventUpdated = "balancesheet.updated"
)

type BalanceSheetCreatedPayload struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	UserID    string    `json:"user_id"`
	Year      int       `json:"year"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BalanceSheetUpdatedPayload struct {
	ID        string    `json:"id"`
	Year      int       `json:"year"`
	UpdatedAt time.Time `json:"updated_at"`
}
