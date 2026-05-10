package handlers

import (
	"cfk/pkg/database"
	"cfk/pkg/event"
	"encoding/json"

	"gorm.io/gorm"
)

type ProjectionHandler struct {
	db *gorm.DB
}

func NewProjectionHandler(db *gorm.DB) *ProjectionHandler {
	return &ProjectionHandler{db: db}
}

func (h *ProjectionHandler) Handle(evt event.Event) error {
	eventStore := database.EventStore{
		AggregateType: evt.AggregateType,
		AggregateID:   evt.AggregateID,
		EventType:     evt.EventType,
		Version:       evt.Version,
	}

	payloadJSON, _ := json.Marshal(evt.Payload)
	eventStore.Payload = string(payloadJSON)

	if evt.Metadata != nil {
		metadataJSON, _ := json.Marshal(evt.Metadata)
		eventStore.Metadata = string(metadataJSON)
	}

	return h.db.Create(&eventStore).Error
}
