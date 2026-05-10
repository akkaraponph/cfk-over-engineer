package projections

import (
	"cfk/pkg/event"

	"github.com/samber/mo"
	"gorm.io/gorm"
)

type FinanceProjectionHandler struct {
	db *gorm.DB
}

func NewFinanceProjectionHandler(db *gorm.DB) *FinanceProjectionHandler {
	return &FinanceProjectionHandler{db: db}
}

func (h *FinanceProjectionHandler) HandlePocket(evt event.Event) mo.Result[struct{}] {
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
		if err := h.db.Table("pocket_projections").Create(proj).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	case "pocket.name_changed":
		if err := h.db.Table("pocket_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"name":       payloadStr(evt.Payload, "name"),
				"updated_at": payloadTime(evt.Payload, "updated_at"),
			}).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	case "pocket.balance_changed":
		if err := h.db.Table("pocket_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"balance":    payloadFloat(evt.Payload, "new_balance"),
				"updated_at": payloadTime(evt.Payload, "updated_at"),
			}).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	case "pocket.deleted":
		if err := h.db.Table("pocket_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"is_deleted": true,
				"updated_at": payloadTime(evt.Payload, "updated_at"),
			}).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	}
	return event.OkHandle()
}

func (h *FinanceProjectionHandler) HandleCashflowIn(evt event.Event) mo.Result[struct{}] {
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
		if err := h.db.Table("cashflowin_projections").Create(proj).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	case "cashflowin.updated":
		if err := h.db.Table("cashflowin_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"amount":       payloadFloat(evt.Payload, "amount"),
				"description":  payloadStr(evt.Payload, "description"),
				"category_id": payloadInt(evt.Payload, "category_id"),
				"receipt":     payloadStr(evt.Payload, "receipt"),
				"updated_at":  payloadTime(evt.Payload, "updated_at"),
			}).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	case "cashflowin.deleted":
		if err := h.db.Table("cashflowin_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"is_deleted": true,
				"updated_at": payloadTime(evt.Payload, "updated_at"),
			}).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	}
	return event.OkHandle()
}

func (h *FinanceProjectionHandler) HandleCashflowOut(evt event.Event) mo.Result[struct{}] {
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
		if err := h.db.Table("cashflowout_projections").Create(proj).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	case "cashflowout.updated":
		if err := h.db.Table("cashflowout_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"amount":       payloadFloat(evt.Payload, "amount"),
				"description":  payloadStr(evt.Payload, "description"),
				"category_id": payloadInt(evt.Payload, "category_id"),
				"receipt":     payloadStr(evt.Payload, "receipt"),
				"updated_at":  payloadTime(evt.Payload, "updated_at"),
			}).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	case "cashflowout.deleted":
		if err := h.db.Table("cashflowout_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"is_deleted": true,
				"updated_at": payloadTime(evt.Payload, "updated_at"),
			}).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	}
	return event.OkHandle()
}

func (h *FinanceProjectionHandler) HandleTransfer(evt event.Event) mo.Result[struct{}] {
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
		if err := h.db.Table("transfer_projections").Create(proj).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	case "transfer.completed", "transfer.failed":
		if err := h.db.Table("transfer_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"status":     payloadStr(evt.Payload, "status"),
				"updated_at": payloadTime(evt.Payload, "updated_at"),
			}).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	case "transfer.deleted":
		if err := h.db.Table("transfer_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"is_deleted": true,
				"updated_at": payloadTime(evt.Payload, "updated_at"),
			}).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	}
	return event.OkHandle()
}

func (h *FinanceProjectionHandler) HandleCategory(evt event.Event) mo.Result[struct{}] {
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
		if err := h.db.Table("category_projections").Create(proj).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	case "category.updated":
		if err := h.db.Table("category_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"name":        payloadStr(evt.Payload, "name"),
				"description": payloadStr(evt.Payload, "description"),
				"updated_at":  payloadTime(evt.Payload, "updated_at"),
			}).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	case "category.deleted":
		if err := h.db.Table("category_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"is_deleted": true,
				"updated_at": payloadTime(evt.Payload, "updated_at"),
			}).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	}
	return event.OkHandle()
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
