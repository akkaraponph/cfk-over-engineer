package user

import "github.com/samber/mo"

type Repository interface {
	AppendEvent(eventType string, aggregateID string, payload any, metadata map[string]interface{}) error
	FindByID(id string) mo.Option[User]
	FindByEmail(tenantID, email string) mo.Option[User]
}
