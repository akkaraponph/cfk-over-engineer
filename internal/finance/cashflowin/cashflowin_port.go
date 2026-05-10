package cashflowin

import "github.com/samber/mo"

type Repository interface {
	AppendEvent(eventType string, aggregateID string, payload any, metadata map[string]interface{}) error
	FindByID(id string) mo.Option[CashflowIn]
	FindByPocket(tenantID, pocketID string) mo.Result[[]CashflowIn]
}
