package tenant

import "time"

type Tenant struct {
	ID        string
	Name      string
	Slug      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

const (
	EventCreated     = "tenant.created"
	EventActivated   = "tenant.activated"
	EventDeactivated = "tenant.deactivated"
)
