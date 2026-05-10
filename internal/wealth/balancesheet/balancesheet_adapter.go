package balancesheet

import (
	"cfk/pkg/database"
	"encoding/json"
	"time"

	"github.com/samber/mo"
	"gorm.io/gorm"
)

type BalanceSheetProjection struct {
	ID        string `gorm:"type:uuid;primaryKey"`
	TenantID  string `gorm:"type:uuid;not null;index"`
	UserID    string `gorm:"type:uuid;index"`
	Year      int    `gorm:"not null"`
	Version   int    `gorm:"not null;default:1"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (BalanceSheetProjection) TableName() string {
	return "balancesheet_projections"
}

type GORMRepository struct {
	db            *gorm.DB
	aggregateType string
}

func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db, aggregateType: "balancesheet"}
}

func (r *GORMRepository) AppendEvent(eventType string, aggregateID string, payload any, metadata map[string]interface{}) error {
	payloadJSON, _ := json.Marshal(payload)
	es := database.EventStore{
		AggregateType: r.aggregateType,
		AggregateID:   aggregateID,
		EventType:     eventType,
		Version:       1,
		Payload:       string(payloadJSON),
	}
	if metadata != nil {
		metadataJSON, _ := json.Marshal(metadata)
		es.Metadata = string(metadataJSON)
	}
	return r.db.Create(&es).Error
}

func (r *GORMRepository) FindByID(id string) mo.Option[BalanceSheet] {
	var proj BalanceSheetProjection
	if err := r.db.Where("id = ?", id).First(&proj).Error; err != nil {
		return mo.None[BalanceSheet]()
	}
	return mo.Some(toDomain(proj))
}

func toDomain(p BalanceSheetProjection) BalanceSheet {
	return BalanceSheet{
		ID:        p.ID,
		TenantID:  p.TenantID,
		UserID:    p.UserID,
		Year:      p.Year,
		Version:   p.Version,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
