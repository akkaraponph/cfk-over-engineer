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
	Version     int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	IsDeleted   bool
}

const (
	EventRecorded = "cashflowout.recorded"
	EventUpdated  = "cashflowout.updated"
	EventDeleted  = "cashflowout.deleted"
)

type CashflowOutRecordedPayload struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	UserID      string    `json:"user_id"`
	PocketID    string    `json:"pocket_id"`
	CategoryID  int       `json:"category_id"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	Receipt     string    `json:"receipt"`
	Type        string    `json:"type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CashflowOutUpdatedPayload struct {
	ID          string    `json:"id"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	CategoryID  int       `json:"category_id"`
	Receipt     string    `json:"receipt"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CashflowOutDeletedPayload struct {
	ID        string    `json:"id"`
	UpdatedAt time.Time `json:"updated_at"`
}
