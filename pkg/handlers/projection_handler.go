package handlers

import (
	"cfk/pkg/database"
	"cfk/pkg/event"
	"encoding/json"
	"fmt"

	"github.com/samber/mo"
	"gorm.io/gorm"
)

type ProjectionHandler struct {
	db *gorm.DB
}

func NewProjectionHandler(db *gorm.DB) *ProjectionHandler {
	return &ProjectionHandler{db: db}
}

func (h *ProjectionHandler) Handle(evt event.Event) mo.Result[struct{}] {
	eventStore := database.EventStore{
		AggregateType: evt.AggregateType,
		AggregateID:   evt.AggregateID,
		EventType:     evt.EventType,
		Version:       evt.Version,
	}

	if evt.Payload != nil {
		payloadJSON, err := json.Marshal(evt.Payload)
		if err != nil {
			return mo.Err[struct{}](fmt.Errorf("marshal payload: %w", err))
		}
		eventStore.Payload = string(payloadJSON)
	}

	if evt.Metadata != nil {
		metadataJSON, _ := json.Marshal(evt.Metadata)
		eventStore.Metadata = string(metadataJSON)
	}

	if err := h.db.Create(&eventStore).Error; err != nil {
		return mo.Err[struct{}](err)
	}
	return event.OkHandle()
}
