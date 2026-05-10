package user

import "time"

type User struct {
	ID             string
	TenantID       string
	Username       string
	HashedPassword string
	FirstName      string
	LastName       string
	Phone          string
	Email          string
	Role           string
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

const (
	EventRegistered    = "user.registered"
	EventActivated     = "user.activated"
	EventDeactivated   = "user.deactivated"
	EventRoleChanged   = "user.role_changed"
	EventProfileUpdated = "user.profile_updated"
)
