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
	Version     int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	IsDeleted   bool
}

const (
	EventCreated = "category.created"
	EventUpdated = "category.updated"
	EventDeleted = "category.deleted"
)

type CategoryCreatedPayload struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	IsCustom    bool      `json:"is_custom"`
	UserID      string    `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CategoryUpdatedPayload struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CategoryDeletedPayload struct {
	ID        string    `json:"id"`
	UpdatedAt time.Time `json:"updated_at"`
}
