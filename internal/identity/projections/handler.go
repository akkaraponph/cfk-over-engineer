package projections

import (
	"cfk/pkg/event"

	"gorm.io/gorm"
)

type IdentityProjectionHandler struct {
	db *gorm.DB
}

func NewIdentityProjectionHandler(db *gorm.DB) *IdentityProjectionHandler {
	return &IdentityProjectionHandler{db: db}
}

func (h *IdentityProjectionHandler) HandleTenant(evt event.Event) error {
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

func (h *IdentityProjectionHandler) HandleUser(evt event.Event) error {
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
				"role":       payloadStr(evt.Payload, "role"),
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
