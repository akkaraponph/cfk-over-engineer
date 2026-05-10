package cashflowout

import "time"

type CashflowOut struct {
	ID          string
	TenantID    string
	Amount      float64
	Description string
	Type        string
	UserID      string
	PocketID    string
	CategoryID  int
	Receipt     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	IsDeleted   bool
}

const (
	EventRecorded = "cashflowout.recorded"
	EventUpdated  = "cashflowout.updated"
	EventDeleted  = "cashflowout.deleted"
)
