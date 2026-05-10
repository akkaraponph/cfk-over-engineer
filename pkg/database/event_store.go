package database

import (
	"cfk/pkg/saga"
	"time"

	"gorm.io/gorm"
)

type EventStore struct {
	ID            string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AggregateType string    `gorm:"type:varchar(100);not null;index:idx_event_store_aggregate"`
	AggregateID   string    `gorm:"type:uuid;not null;index:idx_event_store_aggregate"`
	EventType     string    `gorm:"type:varchar(100);not null;index"`
	Version       int       `gorm:"not null"`
	Payload       string    `gorm:"type:jsonb;not null"`
	Metadata      string    `gorm:"type:jsonb"`
	CreatedAt     time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;index"`
}

func (EventStore) TableName() string {
	return "event_store"
}

func AutoMigrate(db *gorm.DB, models ...interface{}) error {
	allModels := []interface{}{&EventStore{}, &saga.SagaInstanceProjection{}}
	allModels = append(allModels, models...)
	return db.AutoMigrate(allModels...)
}
