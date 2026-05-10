package asset

import "github.com/samber/mo"

type Repository interface {
	AppendEvent(eventType string, aggregateID string, payload map[string]interface{}, metadata map[string]interface{}) error
	FindByID(id string) mo.Option[Asset]
}
