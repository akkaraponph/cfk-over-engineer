package transfer

import (
	"cfk/pkg/database"
	"encoding/json"
	"time"

	"github.com/samber/mo"
	"gorm.io/gorm"
)

type TransferProjection struct {
	ID           string  `gorm:"type:uuid;primaryKey"`
	TenantID     string  `gorm:"type:uuid;not null;index"`
	Amount       float64 `gorm:"type:decimal(15,2);not null"`
	FromPocketID string  `gorm:"type:uuid;not null;index"`
	ToPocketID   string  `gorm:"type:uuid;not null;index"`
	UserID       string  `gorm:"type:uuid;not null;index"`
	Status       string  `gorm:"size:50;not null"`
	IsDeleted    bool    `gorm:"not null;default:false"`
	Version      int     `gorm:"not null;default:1"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (TransferProjection) TableName() string {
	return "transfer_projections"
}

type GORMRepository struct {
	db            *gorm.DB
	aggregateType string
}

func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db, aggregateType: "transfer"}
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

func (r *GORMRepository) FindByID(id string) mo.Option[Transfer] {
	var proj TransferProjection
	if err := r.db.Where("id = ? AND is_deleted = false", id).First(&proj).Error; err != nil {
		return mo.None[Transfer]()
	}
	return mo.Some(toDomain(proj))
}

func toDomain(p TransferProjection) Transfer {
	return Transfer{
		ID:           p.ID,
		TenantID:     p.TenantID,
		Amount:       p.Amount,
		FromPocketID: p.FromPocketID,
		ToPocketID:   p.ToPocketID,
		UserID:       p.UserID,
		Status:       p.Status,
		Version:      p.Version,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
		IsDeleted:    p.IsDeleted,
	}
}
