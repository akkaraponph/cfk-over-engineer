package database

import (
	"cfk/pkg/event"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type EventStoreRecord struct {
	ID            string    `gorm:"type:uuid;primaryKey"`
	AggregateType string    `gorm:"type:varchar(100);not null"`
	AggregateID   string    `gorm:"type:uuid;not null"`
	EventType     string    `gorm:"type:varchar(100);not null"`
	Version       int       `gorm:"not null"`
	Payload       string    `gorm:"type:jsonb;not null"`
	Metadata      string    `gorm:"type:jsonb"`
	CreatedAt     time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (EventStoreRecord) TableName() string {
	return "event_store"
}

func ReplayEvents(db *gorm.DB, handler func(event.Event) error) error {
	var events []EventStoreRecord
	if err := db.Order("created_at ASC, version ASC").Find(&events).Error; err != nil {
		return fmt.Errorf("load events: %w", err)
	}

	for _, e := range events {
		var payload any
		if e.Payload != "" {
			if err := json.Unmarshal([]byte(e.Payload), &payload); err != nil {
				return fmt.Errorf("unmarshal payload for event %s: %w", e.ID, err)
			}
		}

		var metadata map[string]interface{}
		if e.Metadata != "" {
			if err := json.Unmarshal([]byte(e.Metadata), &metadata); err != nil {
				return fmt.Errorf("unmarshal metadata for event %s: %w", e.ID, err)
			}
		}

		evt := event.Event{
			AggregateType: e.AggregateType,
			AggregateID:   e.AggregateID,
			EventType:     e.EventType,
			Version:       e.Version,
			Payload:       payload,
			Metadata:      metadata,
		}

		if err := handler(evt); err != nil {
			return fmt.Errorf("handle event %s (%s): %w", e.ID, e.EventType, err)
		}
	}

	return nil
}

func TruncateProjections(db *gorm.DB) error {
	tables := []string{
		"tenant_projections",
		"tenant_feature_projections",
		"user_projections",
		"pocket_projections",
		"cashflowin_projections",
		"cashflowout_projections",
		"transfer_projections",
		"category_projections",
		"asset_projections",
		"debt_projections",
		"balancesheet_projections",
		"request_log_projections",
	}
	for _, t := range tables {
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", t)).Error; err != nil {
			return fmt.Errorf("truncate %s: %w", t, err)
		}
	}
	return nil
}
