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
	Version     int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	IsDeleted   bool
}

const (
	EventRecorded = "cashflowin.recorded"
	EventUpdated  = "cashflowin.updated"
	EventDeleted  = "cashflowin.deleted"
)

type CashflowInRecordedPayload struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	UserID      string    `json:"user_id"`
	PocketID    string    `json:"pocket_id"`
	CategoryID  int       `json:"category_id"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	Receipt     string    `json:"receipt"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CashflowInUpdatedPayload struct {
	ID          string    `json:"id"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	CategoryID  int       `json:"category_id"`
	Receipt     string    `json:"receipt"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CashflowInDeletedPayload struct {
	ID        string    `json:"id"`
	UpdatedAt time.Time `json:"updated_at"`
}
