package projections

import (
	"cfk/pkg/event"
	"encoding/json"

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
		var p struct {
			TenantID  string  `json:"tenant_id"`
			Name      string  `json:"name"`
			Balance   float64 `json:"balance"`
			UserID    string  `json:"user_id"`
			CreatedAt any     `json:"created_at"`
			UpdatedAt any     `json:"updated_at"`
		}
		if err := decodePayload(evt, &p); err != nil {
			return mo.Err[struct{}](err)
		}
		proj := map[string]interface{}{
			"id":         evt.AggregateID,
			"tenant_id":  p.TenantID,
			"name":       p.Name,
			"balance":    p.Balance,
			"user_id":    p.UserID,
			"is_deleted": false,
			"version":    evt.Version,
			"created_at": p.CreatedAt,
			"updated_at": p.UpdatedAt,
		}
		if err := h.db.Table("pocket_projections").Create(proj).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	case "pocket.name_changed":
		var p struct {
			Name      string `json:"name"`
			UpdatedAt any    `json:"updated_at"`
		}
		if err := decodePayload(evt, &p); err != nil {
			return mo.Err[struct{}](err)
		}
		if err := h.db.Table("pocket_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"name":       p.Name,
				"version":    evt.Version,
				"updated_at": p.UpdatedAt,
			}).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	case "pocket.balance_changed":
		var p struct {
			NewBalance float64 `json:"new_balance"`
			UpdatedAt  any     `json:"updated_at"`
		}
		if err := decodePayload(evt, &p); err != nil {
			return mo.Err[struct{}](err)
		}
		if err := h.db.Table("pocket_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"balance":    p.NewBalance,
				"version":    evt.Version,
				"updated_at": p.UpdatedAt,
			}).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	case "pocket.deleted":
		var p struct {
			UpdatedAt any `json:"updated_at"`
		}
		if err := decodePayload(evt, &p); err != nil {
			return mo.Err[struct{}](err)
		}
		if err := h.db.Table("pocket_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"is_deleted": true,
				"version":    evt.Version,
				"updated_at": p.UpdatedAt,
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

func decodePayload(evt event.Event, target interface{}) error {
	b, err := json.Marshal(evt.Payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, target)
}

func toMap(payload any) map[string]interface{} {
	b, _ := json.Marshal(payload)
	m := make(map[string]interface{})
	json.Unmarshal(b, &m)
	return m
}

func payloadStr(payload any, key string) string {
	m := toMap(payload)
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func payloadBool(payload any, key string) bool {
	m := toMap(payload)
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func payloadFloat(payload any, key string) float64 {
	m := toMap(payload)
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

func payloadInt(payload any, key string) int {
	m := toMap(payload)
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

func payloadTime(payload any, key string) interface{} {
	m := toMap(payload)
	if v, ok := m[key]; ok {
		return v
	}
	return nil
}
