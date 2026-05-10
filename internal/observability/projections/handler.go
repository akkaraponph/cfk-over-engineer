package projections

import (
	"cfk/pkg/event"

	"github.com/samber/mo"
	"gorm.io/gorm"
)

type ObservabilityProjectionHandler struct {
	db *gorm.DB
}

func NewObservabilityProjectionHandler(db *gorm.DB) *ObservabilityProjectionHandler {
	return &ObservabilityProjectionHandler{db: db}
}

func (h *ObservabilityProjectionHandler) HandleRequestLog(evt event.Event) mo.Result[struct{}] {
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
		if err := h.db.Table("request_log_projections").Create(proj).Error; err != nil {
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
