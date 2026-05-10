package cashflowout

import (
	"cfk/pkg/database"
	"encoding/json"
	"time"

	"github.com/samber/mo"
	"gorm.io/gorm"
)

type CashflowOutProjection struct {
	ID          string  `gorm:"type:uuid;primaryKey"`
	TenantID    string  `gorm:"type:uuid;not null;index"`
	Amount      float64 `gorm:"type:decimal(15,2);not null"`
	Description string  `gorm:"type:text"`
	Type        string  `gorm:"size:50;not null"`
	UserID      string  `gorm:"type:uuid;index"`
	PocketID    string  `gorm:"type:uuid;index"`
	CategoryID  int     `gorm:"index"`
	Receipt     string  `gorm:"type:text"`
	IsDeleted   bool    `gorm:"not null;default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (CashflowOutProjection) TableName() string {
	return "cashflowout_projections"
}

type GORMRepository struct {
	db            *gorm.DB
	aggregateType string
}

func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db, aggregateType: "cashflowout"}
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

func (r *GORMRepository) FindByID(id string) mo.Option[CashflowOut] {
	var proj CashflowOutProjection
	if err := r.db.Where("id = ? AND is_deleted = false", id).First(&proj).Error; err != nil {
		return mo.None[CashflowOut]()
	}
	return mo.Some(toDomain(proj))
}

func (r *GORMRepository) FindByPocket(tenantID, pocketID string) mo.Result[[]CashflowOut] {
	var projs []CashflowOutProjection
	if err := r.db.Where("tenant_id = ? AND pocket_id = ? AND is_deleted = false", tenantID, pocketID).Order("created_at DESC").Find(&projs).Error; err != nil {
		return mo.Err[[]CashflowOut](err)
	}
	cfs := make([]CashflowOut, len(projs))
	for i, p := range projs {
		cfs[i] = toDomain(p)
	}
	return mo.Ok(cfs)
}

func toDomain(p CashflowOutProjection) CashflowOut {
	return CashflowOut{
		ID:          p.ID,
		TenantID:    p.TenantID,
		Amount:      p.Amount,
		Description: p.Description,
		Type:        p.Type,
		UserID:      p.UserID,
		PocketID:    p.PocketID,
		CategoryID:  p.CategoryID,
		Receipt:     p.Receipt,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
		IsDeleted:   p.IsDeleted,
	}
}
