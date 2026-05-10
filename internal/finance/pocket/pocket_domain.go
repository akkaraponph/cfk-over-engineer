package pocket

import "time"

type Pocket struct {
	ID        string
	TenantID  string
	Name      string
	Balance   float64
	UserID    string
	CreatedAt time.Time
	UpdatedAt time.Time
	IsDeleted bool
}

const (
	EventCreated      = "pocket.created"
	EventNameChanged  = "pocket.name_changed"
	EventBalanceChanged = "pocket.balance_changed"
	EventDeleted      = "pocket.deleted"
)
