package transfer

import "time"

type Transfer struct {
	ID           string
	TenantID     string
	Amount       float64
	FromPocketID string
	ToPocketID   string
	UserID       string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	IsDeleted    bool
}

const (
	EventInitiated = "transfer.initiated"
	EventCompleted = "transfer.completed"
	EventFailed    = "transfer.failed"
	EventDeleted   = "transfer.deleted"
)
