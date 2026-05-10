package pocket

import "time"

type Pocket struct {
	ID        string
	TenantID  string
	Name      string
	Balance   float64
	UserID    string
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
	IsDeleted bool
}

const (
	EventCreated        = "pocket.created"
	EventNameChanged    = "pocket.name_changed"
	EventBalanceChanged = "pocket.balance_changed"
	EventDeleted        = "pocket.deleted"
)

type CreatedPayload struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	Balance   float64   `json:"balance"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NameChangedPayload struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BalanceChangedPayload struct {
	ID         string    `json:"id"`
	Amount     float64   `json:"amount"`
	NewBalance float64   `json:"new_balance"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type DeletedPayload struct {
	ID        string    `json:"id"`
	UpdatedAt time.Time `json:"updated_at"`
}
