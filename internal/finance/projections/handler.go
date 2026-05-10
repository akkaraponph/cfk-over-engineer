package projections

import (
	"cfk/pkg/event"

	"gorm.io/gorm"
)

type FinanceProjectionHandler struct {
	db *gorm.DB
}

func NewFinanceProjectionHandler(db *gorm.DB) *FinanceProjectionHandler {
	return &FinanceProjectionHandler{db: db}
}

func (h *FinanceProjectionHandler) HandlePocket(evt event.Event) error {
	switch evt.EventType {
	case "pocket.created":
		proj := map[string]interface{}{
			"id":         evt.AggregateID,
			"tenant_id":  payloadStr(evt.Payload, "tenant_id"),
			"name":       payloadStr(evt.Payload, "name"),
			"balance":    payloadFloat(evt.Payload, "balance"),
			"user_id":    payloadStr(evt.Payload, "user_id"),
			"is_deleted": false,
			"created_at": payloadTime(evt.Payload, "created_at"),
			"updated_at": payloadTime(evt.Payload, "updated_at"),
		}
		return h.db.Table("pocket_projections").Create(proj).Error
	case "pocket.name_changed":
		return h.db.Table("pocket_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"name":       payloadStr(evt.Payload, "name"),
				"updated_at": payloadTime(evt.Payload, "updated_at"),
			}).Error
	case "pocket.balance_changed":
		return h.db.Table("pocket_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"balance":    payloadFloat(evt.Payload, "new_balance"),
				"updated_at": payloadTime(evt.Payload, "updated_at"),
			}).Error
	case "pocket.deleted":
		return h.db.Table("pocket_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"is_deleted": true,
				"updated_at": payloadTime(evt.Payload, "updated_at"),
			}).Error
	}
	return nil
}

func (h *FinanceProjectionHandler) HandleCashflowIn(evt event.Event) error {
	switch evt.EventType {
	case "cashflowin.recorded":
		proj := map[string]interface{}{
			"id":          evt.AggregateID,
			"tenant_id":   payloadStr(evt.Payload, "tenant_id"),
			"amount":      payloadFloat(evt.Payload, "amount"),
			"description": payloadStr(evt.Payload, "description"),
			"user_id":     payloadStr(evt.Payload, "user_id"),
			"pocket_id":   payloadStr(evt.Payload, "pocket_id"),
			"category_id": payloadInt(evt.Payload, "category_id"),
			"receipt":     payloadStr(evt.Payload, "receipt"),
			"is_deleted":  false,
			"created_at":  payloadTime(evt.Payload, "created_at"),
			"updated_at":  payloadTime(evt.Payload, "updated_at"),
		}
		return h.db.Table("cashflowin_projections").Create(proj).Error
	case "cashflowin.updated":
		return h.db.Table("cashflowin_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"amount":      payloadFloat(evt.Payload, "amount"),
				"description": payloadStr(evt.Payload, "description"),
				"category_id": payloadInt(evt.Payload, "category_id"),
				"receipt":     payloadStr(evt.Payload, "receipt"),
				"updated_at":  payloadTime(evt.Payload, "updated_at"),
			}).Error
	case "cashflowin.deleted":
		return h.db.Table("cashflowin_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"is_deleted": true,
				"updated_at": payloadTime(evt.Payload, "updated_at"),
			}).Error
	}
	return nil
}

func (h *FinanceProjectionHandler) HandleCashflowOut(evt event.Event) error {
	switch evt.EventType {
	case "cashflowout.recorded":
		proj := map[string]interface{}{
			"id":          evt.AggregateID,
			"tenant_id":   payloadStr(evt.Payload, "tenant_id"),
			"amount":      payloadFloat(evt.Payload, "amount"),
			"description": payloadStr(evt.Payload, "description"),
			"type":        payloadStr(evt.Payload, "type"),
			"user_id":     payloadStr(evt.Payload, "user_id"),
			"pocket_id":   payloadStr(evt.Payload, "pocket_id"),
			"category_id": payloadInt(evt.Payload, "category_id"),
			"receipt":     payloadStr(evt.Payload, "receipt"),
			"is_deleted":  false,
			"created_at":  payloadTime(evt.Payload, "created_at"),
			"updated_at":  payloadTime(evt.Payload, "updated_at"),
		}
		return h.db.Table("cashflowout_projections").Create(proj).Error
	case "cashflowout.updated":
		return h.db.Table("cashflowout_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"amount":      payloadFloat(evt.Payload, "amount"),
				"description": payloadStr(evt.Payload, "description"),
				"category_id": payloadInt(evt.Payload, "category_id"),
				"receipt":     payloadStr(evt.Payload, "receipt"),
				"updated_at":  payloadTime(evt.Payload, "updated_at"),
			}).Error
	case "cashflowout.deleted":
		return h.db.Table("cashflowout_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"is_deleted": true,
				"updated_at": payloadTime(evt.Payload, "updated_at"),
			}).Error
	}
	return nil
}

func (h *FinanceProjectionHandler) HandleTransfer(evt event.Event) error {
	switch evt.EventType {
	case "transfer.initiated":
		proj := map[string]interface{}{
			"id":             evt.AggregateID,
			"tenant_id":      payloadStr(evt.Payload, "tenant_id"),
			"amount":         payloadFloat(evt.Payload, "amount"),
			"from_pocket_id": payloadStr(evt.Payload, "from_pocket_id"),
			"to_pocket_id":   payloadStr(evt.Payload, "to_pocket_id"),
			"user_id":        payloadStr(evt.Payload, "user_id"),
			"status":         payloadStr(evt.Payload, "status"),
			"is_deleted":     false,
			"created_at":     payloadTime(evt.Payload, "created_at"),
			"updated_at":     payloadTime(evt.Payload, "updated_at"),
		}
		return h.db.Table("transfer_projections").Create(proj).Error
	case "transfer.completed", "transfer.failed":
		return h.db.Table("transfer_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"status":     payloadStr(evt.Payload, "status"),
				"updated_at": payloadTime(evt.Payload, "updated_at"),
			}).Error
	case "transfer.deleted":
		return h.db.Table("transfer_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"is_deleted": true,
				"updated_at": payloadTime(evt.Payload, "updated_at"),
			}).Error
	}
	return nil
}

func (h *FinanceProjectionHandler) HandleCategory(evt event.Event) error {
	switch evt.EventType {
	case "category.created":
		proj := map[string]interface{}{
			"id":          evt.AggregateID,
			"tenant_id":   payloadStr(evt.Payload, "tenant_id"),
			"name":        payloadStr(evt.Payload, "name"),
			"description": payloadStr(evt.Payload, "description"),
			"type":        payloadStr(evt.Payload, "type"),
			"is_custom":   payloadBool(evt.Payload, "is_custom"),
			"user_id":     payloadStr(evt.Payload, "user_id"),
			"is_deleted":  false,
			"created_at":  payloadTime(evt.Payload, "created_at"),
			"updated_at":  payloadTime(evt.Payload, "updated_at"),
		}
		return h.db.Table("category_projections").Create(proj).Error
	case "category.updated":
		return h.db.Table("category_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"name":        payloadStr(evt.Payload, "name"),
				"description": payloadStr(evt.Payload, "description"),
				"updated_at":  payloadTime(evt.Payload, "updated_at"),
			}).Error
	case "category.deleted":
		return h.db.Table("category_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"is_deleted": true,
				"updated_at": payloadTime(evt.Payload, "updated_at"),
			}).Error
	}
	return nil
}

func payloadStr(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func payloadBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func payloadFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case int:
			return float64(val)
		case int64:
			return float64(val)
		}
	}
	return 0
}

func payloadInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return int(val)
		case int:
			return val
		case int64:
			return int(val)
		}
	}
	return 0
}

func payloadTime(m map[string]interface{}, key string) interface{} {
	if v, ok := m[key]; ok {
		return v
	}
	return nil
}
