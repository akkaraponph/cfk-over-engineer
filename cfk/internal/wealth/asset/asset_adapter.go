package asset

import (
	"cfk/pkg/database"
	"encoding/json"
	"time"

	"github.com/samber/mo"
	"gorm.io/gorm"
)

type AssetProjection struct {
	ID              string  `gorm:"type:uuid;primaryKey"`
	TenantID        string  `gorm:"type:uuid;not null;index"`
	Type            string  `gorm:"size:255;not null"`
	Description     string  `gorm:"size:1024"`
	Value           float64 `gorm:"type:decimal(15,2);not null;default:0"`
	CashflowPerYear float64 `gorm:"type:decimal(15,2);not null;default:0"`
	BalanceSheetID  string  `gorm:"type:uuid;index"`
	UserID          string  `gorm:"type:uuid;index"`
	Version         int     `gorm:"not null;default:1"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (AssetProjection) TableName() string {
	return "asset_projections"
}

type GORMRepository struct {
	db            *gorm.DB
	aggregateType string
}

func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db, aggregateType: "asset"}
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

func (r *GORMRepository) FindByID(id string) mo.Option[Asset] {
	var proj AssetProjection
	if err := r.db.Where("id = ?", id).First(&proj).Error; err != nil {
		return mo.None[Asset]()
	}
	return mo.Some(toDomain(proj))
}

func toDomain(p AssetProjection) Asset {
	return Asset{
		ID:              p.ID,
		TenantID:        p.TenantID,
		Type:            p.Type,
		Description:     p.Description,
		Value:           p.Value,
		CashflowPerYear: p.CashflowPerYear,
		BalanceSheetID:  p.BalanceSheetID,
		UserID:          p.UserID,
		Version:         p.Version,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}
