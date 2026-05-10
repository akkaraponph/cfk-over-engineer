package pocket

import "github.com/samber/mo"

type Repository interface {
	AppendEvent(eventType string, aggregateID string, payload any, metadata map[string]interface{}) error
	FindByID(id string) mo.Option[Pocket]
	FindByUser(tenantID, userID string) mo.Result[[]Pocket]
}
