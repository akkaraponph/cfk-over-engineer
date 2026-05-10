package user

import (
	"cfk/pkg/database"
	"encoding/json"
	"time"

	"github.com/samber/mo"
	"gorm.io/gorm"
)

type UserProjection struct {
	ID             string `gorm:"type:uuid;primaryKey"`
	TenantID       string `gorm:"type:uuid;not null;index"`
	Username       string `gorm:"size:255;not null"`
	HashedPassword string `gorm:"size:255;not null"`
	FirstName      string `gorm:"size:255"`
	LastName       string `gorm:"size:255"`
	Phone          string `gorm:"size:20"`
	Email          string `gorm:"size:255;not null;index:idx_user_tenant_email,unique"`
	Role           string `gorm:"size:50;not null"`
	IsActive       bool   `gorm:"not null;default:true"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (UserProjection) TableName() string {
	return "user_projections"
}

type GORMRepository struct {
	db            *gorm.DB
	aggregateType string
}

func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db, aggregateType: "user"}
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

func (r *GORMRepository) FindByID(id string) mo.Option[User] {
	var proj UserProjection
	if err := r.db.Where("id = ?", id).First(&proj).Error; err != nil {
		return mo.None[User]()
	}
	return mo.Some(toDomain(proj))
}

func (r *GORMRepository) FindByEmail(tenantID, email string) mo.Option[User] {
	var proj UserProjection
	if err := r.db.Where("tenant_id = ? AND email = ?", tenantID, email).First(&proj).Error; err != nil {
		return mo.None[User]()
	}
	return mo.Some(toDomain(proj))
}

func toDomain(p UserProjection) User {
	return User{
		ID:             p.ID,
		TenantID:       p.TenantID,
		Username:       p.Username,
		HashedPassword: p.HashedPassword,
		FirstName:      p.FirstName,
		LastName:       p.LastName,
		Phone:          p.Phone,
		Email:          p.Email,
		Role:           p.Role,
		IsActive:       p.IsActive,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}
