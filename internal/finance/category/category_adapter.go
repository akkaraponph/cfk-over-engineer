package category

import (
	"cfk/pkg/database"
	"encoding/json"
	"time"

	"github.com/samber/mo"
	"gorm.io/gorm"
)

type CategoryProjection struct {
	ID          string `gorm:"type:uuid;primaryKey"`
	TenantID    string `gorm:"type:uuid;not null;index"`
	Name        string `gorm:"size:255;not null"`
	Description string `gorm:"type:text"`
	Type        string `gorm:"size:50;not null"`
	IsCustom    bool   `gorm:"not null;default:false"`
	UserID      string `gorm:"type:uuid"`
	IsDeleted   bool   `gorm:"not null;default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (CategoryProjection) TableName() string {
	return "category_projections"
}

type GORMRepository struct {
	db            *gorm.DB
	aggregateType string
}

func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db, aggregateType: "category"}
}

func (r *GORMRepository) AppendEvent(eventType string, aggregateID string, payload map[string]interface{}, metadata map[string]interface{}) error {
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

func (r *GORMRepository) FindByID(id string) mo.Option[Category] {
	var proj CategoryProjection
	if err := r.db.Where("id = ? AND is_deleted = false", id).First(&proj).Error; err != nil {
		return mo.None[Category]()
	}
	return mo.Some(toDomain(proj))
}

func (r *GORMRepository) FindByTenant(tenantID string) mo.Result[[]Category] {
	var projs []CategoryProjection
	if err := r.db.Where("tenant_id = ? AND is_deleted = false", tenantID).Find(&projs).Error; err != nil {
		return mo.Err[[]Category](err)
	}
	cats := make([]Category, len(projs))
	for i, p := range projs {
		cats[i] = toDomain(p)
	}
	return mo.Ok(cats)
}

func toDomain(p CategoryProjection) Category {
	return Category{
		ID:          p.ID,
		TenantID:    p.TenantID,
		Name:        p.Name,
		Description: p.Description,
		Type:        p.Type,
		IsCustom:    p.IsCustom,
		UserID:      p.UserID,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
		IsDeleted:   p.IsDeleted,
	}
}
