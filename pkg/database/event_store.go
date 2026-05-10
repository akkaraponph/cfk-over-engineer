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

type TenantFeatureProjection struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID    string    `gorm:"type:uuid;not null;index:idx_tenant_feature_tenant"`
	Feature     string    `gorm:"type:varchar(100);not null"`
	IsEnabled   bool      `gorm:"not null;default:true"`
	EnabledAt   time.Time
	DisabledAt  time.Time
	EnabledBy   string    `gorm:"type:uuid"`
	DisabledBy  string    `gorm:"type:uuid"`
	CreatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (TenantFeatureProjection) TableName() string {
	return "tenant_feature_projections"
}

func AutoMigrate(db *gorm.DB, models ...interface{}) error {
	allModels := []interface{}{
		&EventStore{},
		&saga.SagaInstanceProjection{},
		&TenantFeatureProjection{},
	}
	allModels = append(allModels, models...)
	return db.AutoMigrate(allModels...)
}
