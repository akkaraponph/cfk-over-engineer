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
	Version      int
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

type TransferInitiatedPayload struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	UserID       string    `json:"user_id"`
	FromPocketID string    `json:"from_pocket_id"`
	ToPocketID   string    `json:"to_pocket_id"`
	Amount       float64   `json:"amount"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type TransferCompletedPayload struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TransferFailedPayload struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Reason    string    `json:"reason"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TransferDeletedPayload struct {
	ID        string    `json:"id"`
	UpdatedAt time.Time `json:"updated_at"`
}
