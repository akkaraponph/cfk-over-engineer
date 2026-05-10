package pocket

import (
	"cfk/pkg/database"
	"encoding/json"
	"time"

	"github.com/samber/mo"
	"gorm.io/gorm"
)

type PocketProjection struct {
	ID        string    `gorm:"type:uuid;primaryKey"`
	TenantID  string    `gorm:"type:uuid;not null;index"`
	Name      string    `gorm:"size:255;not null"`
	Balance   float64   `gorm:"type:decimal(15,2);not null;default:0"`
	UserID    string    `gorm:"type:uuid;index"`
	Version   int       `gorm:"not null;default:1"`
	IsDeleted bool      `gorm:"not null;default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (PocketProjection) TableName() string {
	return "pocket_projections"
}

type GORMRepository struct {
	db            *gorm.DB
	aggregateType string
}

func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db, aggregateType: "pocket"}
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

func (r *GORMRepository) FindByID(id string) mo.Option[Pocket] {
	var proj PocketProjection
	if err := r.db.Where("id = ? AND is_deleted = false", id).First(&proj).Error; err != nil {
		return mo.None[Pocket]()
	}
	return mo.Some(toDomain(proj))
}

func (r *GORMRepository) FindByUser(tenantID, userID string) mo.Result[[]Pocket] {
	var projs []PocketProjection
	if err := r.db.Where("tenant_id = ? AND user_id = ? AND is_deleted = false", tenantID, userID).Find(&projs).Error; err != nil {
		return mo.Err[[]Pocket](err)
	}
	pockets := make([]Pocket, len(projs))
	for i, p := range projs {
		pockets[i] = toDomain(p)
	}
	return mo.Ok(pockets)
}

func toDomain(p PocketProjection) Pocket {
	return Pocket{
		ID:        p.ID,
		TenantID:  p.TenantID,
		Name:      p.Name,
		Balance:   p.Balance,
		UserID:    p.UserID,
		Version:   p.Version,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
		IsDeleted: p.IsDeleted,
	}
}
