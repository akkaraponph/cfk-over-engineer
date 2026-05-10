package category

import "github.com/samber/mo"

type Repository interface {
	AppendEvent(eventType string, aggregateID string, payload map[string]interface{}, metadata map[string]interface{}) error
	FindByID(id string) mo.Option[Category]
	FindByTenant(tenantID string) mo.Result[[]Category]
}
