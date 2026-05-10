package cashflowin

import "time"

type CashflowIn struct {
	ID          string
	TenantID    string
	Amount      float64
	Description string
	UserID      string
	PocketID    string
	CategoryID  int
	Receipt     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	IsDeleted   bool
}

const (
	EventRecorded = "cashflowin.recorded"
	EventUpdated  = "cashflowin.updated"
	EventDeleted  = "cashflowin.deleted"
)
