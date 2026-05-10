package projections

import (
	"cfk/pkg/event"

	"github.com/samber/mo"
	"gorm.io/gorm"
)

type WealthProjectionHandler struct {
	db *gorm.DB
}

func NewWealthProjectionHandler(db *gorm.DB) *WealthProjectionHandler {
	return &WealthProjectionHandler{db: db}
}

func (h *WealthProjectionHandler) HandleAsset(evt event.Event) mo.Result[struct{}] {
	switch evt.EventType {
	case "asset.recorded":
		proj := map[string]interface{}{
			"id":                evt.AggregateID,
			"tenant_id":         payloadStr(evt.Payload, "tenant_id"),
			"type":              payloadStr(evt.Payload, "type"),
			"description":       payloadStr(evt.Payload, "description"),
			"value":             payloadFloat(evt.Payload, "value"),
			"cashflow_per_year": payloadFloat(evt.Payload, "cashflow_per_year"),
			"balance_sheet_id":  payloadStr(evt.Payload, "balance_sheet_id"),
			"user_id":           payloadStr(evt.Payload, "user_id"),
			"created_at":        payloadTime(evt.Payload, "created_at"),
			"updated_at":        payloadTime(evt.Payload, "updated_at"),
		}
		if err := h.db.Table("asset_projections").Create(proj).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	case "asset.value_changed":
		if err := h.db.Table("asset_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"value":             payloadFloat(evt.Payload, "value"),
				"cashflow_per_year": payloadFloat(evt.Payload, "cashflow_per_year"),
				"updated_at":        payloadTime(evt.Payload, "updated_at"),
			}).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	case "asset.assigned_to_balancesheet":
		if err := h.db.Table("asset_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"balance_sheet_id": payloadStr(evt.Payload, "balance_sheet_id"),
				"updated_at":       payloadTime(evt.Payload, "updated_at"),
			}).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	case "asset.unassigned_from_balancesheet":
		if err := h.db.Table("asset_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"balance_sheet_id": "",
				"updated_at":       payloadTime(evt.Payload, "updated_at"),
			}).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	}
	return event.OkHandle()
}

func (h *WealthProjectionHandler) HandleDebt(evt event.Event) mo.Result[struct{}] {
	switch evt.EventType {
	case "debt.recorded":
		proj := map[string]interface{}{
			"id":               evt.AggregateID,
			"tenant_id":        payloadStr(evt.Payload, "tenant_id"),
			"type":             payloadStr(evt.Payload, "type"),
			"description":      payloadStr(evt.Payload, "description"),
			"amount":           payloadFloat(evt.Payload, "amount"),
			"interest":         payloadFloat(evt.Payload, "interest"),
			"minimum_pay":      payloadFloat(evt.Payload, "minimum_pay"),
			"priority":         payloadInt(evt.Payload, "priority"),
			"balance_sheet_id": payloadStr(evt.Payload, "balance_sheet_id"),
			"user_id":          payloadStr(evt.Payload, "user_id"),
			"created_at":       payloadTime(evt.Payload, "created_at"),
			"updated_at":       payloadTime(evt.Payload, "updated_at"),
		}
		if err := h.db.Table("debt_projections").Create(proj).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	case "debt.amount_changed":
		if err := h.db.Table("debt_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"amount":     payloadFloat(evt.Payload, "amount"),
				"updated_at": payloadTime(evt.Payload, "updated_at"),
			}).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	case "debt.assigned_to_balancesheet":
		if err := h.db.Table("debt_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"balance_sheet_id": payloadStr(evt.Payload, "balance_sheet_id"),
				"updated_at":       payloadTime(evt.Payload, "updated_at"),
			}).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	case "debt.unassigned_from_balancesheet":
		if err := h.db.Table("debt_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"balance_sheet_id": "",
				"updated_at":       payloadTime(evt.Payload, "updated_at"),
			}).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	}
	return event.OkHandle()
}

func (h *WealthProjectionHandler) HandleBalanceSheet(evt event.Event) mo.Result[struct{}] {
	switch evt.EventType {
	case "balancesheet.created":
		proj := map[string]interface{}{
			"id":         evt.AggregateID,
			"tenant_id":  payloadStr(evt.Payload, "tenant_id"),
			"user_id":    payloadStr(evt.Payload, "user_id"),
			"year":       payloadInt(evt.Payload, "year"),
			"created_at": payloadTime(evt.Payload, "created_at"),
			"updated_at": payloadTime(evt.Payload, "updated_at"),
		}
		if err := h.db.Table("balancesheet_projections").Create(proj).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	case "balancesheet.updated":
		if err := h.db.Table("balancesheet_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"year":       payloadInt(evt.Payload, "year"),
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
