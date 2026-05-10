package category

import "time"

type Category struct {
	ID          string
	TenantID    string
	Name        string
	Description string
	Type        string
	IsCustom    bool
	UserID      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	IsDeleted   bool
}

const (
	EventCreated = "category.created"
	EventUpdated = "category.updated"
	EventDeleted = "category.deleted"
)
