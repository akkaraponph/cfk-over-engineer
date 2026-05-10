package projections

import (
	"cfk/pkg/event"

	"github.com/samber/mo"
	"gorm.io/gorm"
)

type IdentityProjectionHandler struct {
	db *gorm.DB
}

func NewIdentityProjectionHandler(db *gorm.DB) *IdentityProjectionHandler {
	return &IdentityProjectionHandler{db: db}
}

func (h *IdentityProjectionHandler) HandleTenant(evt event.Event) mo.Result[struct{}] {
	switch evt.EventType {
	case "tenant.created":
		proj := map[string]interface{}{
			"id":         evt.AggregateID,
			"name":       payloadStr(evt.Payload, "name"),
			"slug":       payloadStr(evt.Payload, "slug"),
			"plan":       payloadStr(evt.Payload, "plan"),
			"is_active":  payloadBool(evt.Payload, "is_active"),
			"created_at": payloadTime(evt.Payload, "created_at"),
			"updated_at": payloadTime(evt.Payload, "updated_at"),
		}
		if err := h.db.Table("tenant_projections").Create(proj).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	case "tenant.activated", "tenant.deactivated":
		if err := h.db.Table("tenant_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"is_active":  payloadBool(evt.Payload, "is_active"),
				"updated_at": payloadTime(evt.Payload, "updated_at"),
			}).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	case "tenant.plan_changed":
		if err := h.db.Table("tenant_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"plan":       payloadStr(evt.Payload, "plan"),
				"updated_at": payloadTime(evt.Payload, "updated_at"),
			}).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	}
	return event.OkHandle()
}

func (h *IdentityProjectionHandler) HandleTenantFeature(evt event.Event) mo.Result[struct{}] {
	switch evt.EventType {
	case "tenant.feature_enabled":
		proj := map[string]interface{}{
			"id":         payloadStr(evt.Payload, "id"),
			"tenant_id":  payloadStr(evt.Payload, "tenant_id"),
			"feature":    payloadStr(evt.Payload, "feature"),
			"is_enabled": payloadBool(evt.Payload, "is_enabled"),
			"enabled_by": payloadStr(evt.Payload, "enabled_by"),
			"enabled_at": payloadTime(evt.Payload, "enabled_at"),
			"created_at": payloadTime(evt.Payload, "created_at"),
			"updated_at": payloadTime(evt.Payload, "updated_at"),
		}
		if err := h.db.Table("tenant_feature_projections").Create(proj).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	case "tenant.feature_disabled":
		if err := h.db.Table("tenant_feature_projections").
			Where("tenant_id = ? AND feature = ?", payloadStr(evt.Payload, "tenant_id"), payloadStr(evt.Payload, "feature")).
			Updates(map[string]interface{}{
				"is_enabled":  false,
				"disabled_by": payloadStr(evt.Payload, "disabled_by"),
				"disabled_at": payloadTime(evt.Payload, "disabled_at"),
				"updated_at": payloadTime(evt.Payload, "updated_at"),
			}).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	}
	return event.OkHandle()
}

func (h *IdentityProjectionHandler) HandleUser(evt event.Event) mo.Result[struct{}] {
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
			"email":          payloadStr(evt.Payload, "email"),
			"role":            payloadStr(evt.Payload, "role"),
			"is_active":       payloadBool(evt.Payload, "is_active"),
			"created_at":      payloadTime(evt.Payload, "created_at"),
			"updated_at":      payloadTime(evt.Payload, "updated_at"),
		}
		if err := h.db.Table("user_projections").Create(proj).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	case "user.activated", "user.deactivated":
		if err := h.db.Table("user_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"is_active":  payloadBool(evt.Payload, "is_active"),
				"updated_at": payloadTime(evt.Payload, "updated_at"),
			}).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	case "user.role_changed":
		if err := h.db.Table("user_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"role":       payloadStr(evt.Payload, "role"),
				"updated_at": payloadTime(evt.Payload, "updated_at"),
			}).Error; err != nil {
			return mo.Err[struct{}](err)
		}
		return event.OkHandle()
	case "user.profile_updated":
		if err := h.db.Table("user_projections").
			Where("id = ?", evt.AggregateID).
			Updates(map[string]interface{}{
				"first_name": payloadStr(evt.Payload, "first_name"),
				"last_name":  payloadStr(evt.Payload, "last_name"),
				"phone":      payloadStr(evt.Payload, "phone"),
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

func payloadTime(m map[string]interface{}, key string) interface{} {
	if v, ok := m[key]; ok {
		return v
	}
	return nil
}
