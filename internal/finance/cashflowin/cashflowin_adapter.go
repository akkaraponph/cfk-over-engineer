package cashflowin

import (
	"cfk/pkg/database"
	"encoding/json"
	"time"

	"github.com/samber/mo"
	"gorm.io/gorm"
)

type CashflowInProjection struct {
	ID          string  `gorm:"type:uuid;primaryKey"`
	TenantID    string  `gorm:"type:uuid;not null;index"`
	Amount      float64 `gorm:"type:decimal(15,2);not null"`
	Description string  `gorm:"type:text"`
	UserID      string  `gorm:"type:uuid;index"`
	PocketID    string  `gorm:"type:uuid;index"`
	CategoryID  int     `gorm:"index"`
	Receipt     string  `gorm:"type:text"`
	IsDeleted   bool    `gorm:"not null;default:false"`
	Version     int     `gorm:"not null;default:1"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (CashflowInProjection) TableName() string {
	return "cashflowin_projections"
}

type GORMRepository struct {
	db            *gorm.DB
	aggregateType string
}

func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db, aggregateType: "cashflowin"}
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

func (r *GORMRepository) FindByID(id string) mo.Option[CashflowIn] {
	var proj CashflowInProjection
	if err := r.db.Where("id = ? AND is_deleted = false", id).First(&proj).Error; err != nil {
		return mo.None[CashflowIn]()
	}
	return mo.Some(toDomain(proj))
}

func (r *GORMRepository) FindByPocket(tenantID, pocketID string) mo.Result[[]CashflowIn] {
	var projs []CashflowInProjection
	if err := r.db.Where("tenant_id = ? AND pocket_id = ? AND is_deleted = false", tenantID, pocketID).Order("created_at DESC").Find(&projs).Error; err != nil {
		return mo.Err[[]CashflowIn](err)
	}
	cfs := make([]CashflowIn, len(projs))
	for i, p := range projs {
		cfs[i] = toDomain(p)
	}
	return mo.Ok(cfs)
}

func toDomain(p CashflowInProjection) CashflowIn {
	return CashflowIn{
		ID:          p.ID,
		TenantID:    p.TenantID,
		Amount:      p.Amount,
		Description: p.Description,
		UserID:      p.UserID,
		PocketID:    p.PocketID,
		CategoryID:  p.CategoryID,
		Receipt:     p.Receipt,
		Version:     p.Version,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
		IsDeleted:   p.IsDeleted,
	}
}
