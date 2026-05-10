package debt

import (
	"cfk/pkg/database"
	"encoding/json"
	"time"

	"github.com/samber/mo"
	"gorm.io/gorm"
)

type DebtProjection struct {
	ID             string  `gorm:"type:uuid;primaryKey"`
	TenantID       string  `gorm:"type:uuid;not null;index"`
	Type           string  `gorm:"size:255;not null"`
	Description    string  `gorm:"size:1024"`
	Amount         float64 `gorm:"type:decimal(15,2);not null;default:0"`
	Interest       float64 `gorm:"type:decimal(5,2);not null;default:0"`
	MinimumPay     float64 `gorm:"type:decimal(15,2);not null;default:0"`
	Priority       int     `gorm:"not null;default:0"`
	BalanceSheetID string  `gorm:"type:uuid;index"`
	UserID         string  `gorm:"type:uuid;index"`
	Version        int     `gorm:"not null;default:1"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (DebtProjection) TableName() string {
	return "debt_projections"
}

type GORMRepository struct {
	db            *gorm.DB
	aggregateType string
}

func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db, aggregateType: "debt"}
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

func (r *GORMRepository) FindByID(id string) mo.Option[Debt] {
	var proj DebtProjection
	if err := r.db.Where("id = ?", id).First(&proj).Error; err != nil {
		return mo.None[Debt]()
	}
	return mo.Some(toDomain(proj))
}

func toDomain(p DebtProjection) Debt {
	return Debt{
		ID:             p.ID,
		TenantID:       p.TenantID,
		Type:           p.Type,
		Description:    p.Description,
		Amount:         p.Amount,
		Interest:       p.Interest,
		MinimumPay:     p.MinimumPay,
		Priority:       p.Priority,
		BalanceSheetID: p.BalanceSheetID,
		UserID:         p.UserID,
		Version:        p.Version,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}
