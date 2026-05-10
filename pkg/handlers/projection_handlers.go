package handlers

import (
	"cfk/pkg/event"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProjectionHandlers struct {
	db *gorm.DB
}

func NewProjectionHandlers(db *gorm.DB) *ProjectionHandlers {
	return &ProjectionHandlers{db: db}
}

func (h *ProjectionHandlers) HandleTenant(evt event.Event) error {
	switch evt.EventType {
	case "tenant.created":
		proj := map[string]interface{}{
			"id":         evt.AggregateID,
			"name":       payloadStr(evt.Payload, "name"),
			"slug":       payloadStr(evt.Payload, "slug"),
			"is_active":  payloadBool(evt.Payload, "is_active"),
			"created_at": payloadTime(evt.Payload, "created_at"),
			"updated_at": payloadTime(evt.Payload, "updated_at"),
		}
		return h.db.Table("tenant_projections").Create(proj).Error
	case "tenant.activated", "tenant.deactivated":
		return h.db.Table("tenant_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"is_active":  payloadBool(evt.Payload, "is_active"),
				"updated_at": payloadTime(evt.Payload, "updated_at"),
			}).Error
	}
	return nil
}

func (h *ProjectionHandlers) HandleUser(evt event.Event) error {
	switch evt.EventType {
	case "user.registered":
		proj := map[string]interface{}{
			"id":              evt.AggregateID,
			"tenant_id":       payloadStr(evt.Payload, "tenant_id"),
			"username":        payloadStr(evt.Payload, "username"),
			"hashed_password": payloadStr(evt.Payload, "hashed_password"),
			"first_name":      payloadStr(evt.Payload, "first_name"),
			"last_name":       payloadStr(evt.Payload, "last_name"),
			"phone":           payloadStr(evt.Payload, "phone"),
			"email":           payloadStr(evt.Payload, "email"),
			"role":            payloadStr(evt.Payload, "role"),
			"is_active":       payloadBool(evt.Payload, "is_active"),
			"created_at":      payloadTime(evt.Payload, "created_at"),
			"updated_at":      payloadTime(evt.Payload, "updated_at"),
		}
		return h.db.Table("user_projections").Create(proj).Error
	case "user.activated", "user.deactivated":
		return h.db.Table("user_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"is_active":  payloadBool(evt.Payload, "is_active"),
				"updated_at": payloadTime(evt.Payload, "updated_at"),
			}).Error
	case "user.role_changed":
		return h.db.Table("user_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"role":        payloadStr(evt.Payload, "role"),
				"updated_at": payloadTime(evt.Payload, "updated_at"),
			}).Error
	case "user.profile_updated":
		return h.db.Table("user_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"first_name": payloadStr(evt.Payload, "first_name"),
				"last_name":  payloadStr(evt.Payload, "last_name"),
				"phone":      payloadStr(evt.Payload, "phone"),
				"updated_at": payloadTime(evt.Payload, "updated_at"),
			}).Error
	}
	return nil
}

func (h *ProjectionHandlers) HandlePocket(evt event.Event) error {
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
				"name":        payloadStr(evt.Payload, "name"),
				"updated_at":  payloadTime(evt.Payload, "updated_at"),
			}).Error
	case "pocket.balance_changed":
		return h.db.Table("pocket_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"balance":     payloadFloat(evt.Payload, "new_balance"),
				"updated_at":  payloadTime(evt.Payload, "updated_at"),
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

func (h *ProjectionHandlers) HandleCashflowIn(evt event.Event) error {
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

func (h *ProjectionHandlers) HandleCashflowOut(evt event.Event) error {
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

func (h *ProjectionHandlers) HandleTransfer(evt event.Event) error {
	switch evt.EventType {
	case "transfer.initiated":
		proj := map[string]interface{}{
			"id":              evt.AggregateID,
			"tenant_id":       payloadStr(evt.Payload, "tenant_id"),
			"amount":          payloadFloat(evt.Payload, "amount"),
			"from_pocket_id":  payloadStr(evt.Payload, "from_pocket_id"),
			"to_pocket_id":    payloadStr(evt.Payload, "to_pocket_id"),
			"user_id":         payloadStr(evt.Payload, "user_id"),
			"status":          payloadStr(evt.Payload, "status"),
			"is_deleted":      false,
			"created_at":      payloadTime(evt.Payload, "created_at"),
			"updated_at":      payloadTime(evt.Payload, "updated_at"),
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

func (h *ProjectionHandlers) HandleCategory(evt event.Event) error {
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

func (h *ProjectionHandlers) HandleAsset(evt event.Event) error {
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
		return h.db.Table("asset_projections").Create(proj).Error
	case "asset.value_changed":
		return h.db.Table("asset_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"value":             payloadFloat(evt.Payload, "value"),
				"cashflow_per_year": payloadFloat(evt.Payload, "cashflow_per_year"),
				"updated_at":        payloadTime(evt.Payload, "updated_at"),
			}).Error
	case "asset.assigned_to_balancesheet":
		return h.db.Table("asset_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"balance_sheet_id": payloadStr(evt.Payload, "balance_sheet_id"),
				"updated_at":       payloadTime(evt.Payload, "updated_at"),
			}).Error
	case "asset.unassigned_from_balancesheet":
		return h.db.Table("asset_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"balance_sheet_id": "",
				"updated_at":       payloadTime(evt.Payload, "updated_at"),
			}).Error
	}
	return nil
}

func (h *ProjectionHandlers) HandleDebt(evt event.Event) error {
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
		return h.db.Table("debt_projections").Create(proj).Error
	case "debt.amount_changed":
		return h.db.Table("debt_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"amount":     payloadFloat(evt.Payload, "amount"),
				"updated_at": payloadTime(evt.Payload, "updated_at"),
			}).Error
	case "debt.assigned_to_balancesheet":
		return h.db.Table("debt_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"balance_sheet_id": payloadStr(evt.Payload, "balance_sheet_id"),
				"updated_at":       payloadTime(evt.Payload, "updated_at"),
			}).Error
	case "debt.unassigned_from_balancesheet":
		return h.db.Table("debt_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"balance_sheet_id": "",
				"updated_at":       payloadTime(evt.Payload, "updated_at"),
			}).Error
	}
	return nil
}

func (h *ProjectionHandlers) HandleBalanceSheet(evt event.Event) error {
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
		return h.db.Table("balancesheet_projections").Create(proj).Error
	case "balancesheet.updated":
		return h.db.Table("balancesheet_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"year":       payloadInt(evt.Payload, "year"),
				"updated_at": payloadTime(evt.Payload, "updated_at"),
			}).Error
	}
	return nil
}

func (h *ProjectionHandlers) HandleRequestLog(evt event.Event) error {
	if evt.EventType == "requestlog.recorded" {
		proj := map[string]interface{}{
			"id":               evt.AggregateID,
			"tenant_id":        payloadStr(evt.Payload, "tenant_id"),
			"user_id":          payloadStr(evt.Payload, "user_id"),
			"method":           payloadStr(evt.Payload, "method"),
			"path":             payloadStr(evt.Payload, "path"),
			"query_params":      payloadStr(evt.Payload, "query_params"),
			"request_headers":  payloadStr(evt.Payload, "request_headers"),
			"request_body":     payloadStr(evt.Payload, "request_body"),
			"response_status":  payloadInt(evt.Payload, "response_status"),
			"response_body":    payloadStr(evt.Payload, "response_body"),
			"response_time_ms": payloadInt(evt.Payload, "response_time_ms"),
			"ip_address":       payloadStr(evt.Payload, "ip_address"),
			"user_agent":       payloadStr(evt.Payload, "user_agent"),
			"error_message":    payloadStr(evt.Payload, "error_message"),
			"error_stack":      payloadStr(evt.Payload, "error_stack"),
			"created_at":       payloadTime(evt.Payload, "created_at"),
		}
		return h.db.Table("request_log_projections").Create(proj).Error
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

func onConflictDoNothing(db *gorm.DB) *gorm.DB {
	return db.Clauses(clause.OnConflict{DoNothing: true})
}
