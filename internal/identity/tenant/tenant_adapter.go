package tenant

import (
	"time"

	"github.com/samber/mo"
	"gorm.io/gorm"
)

type TenantProjection struct {
	ID        string    `gorm:"type:uuid;primaryKey"`
	Name      string    `gorm:"size:255;not null"`
	Slug      string    `gorm:"size:255;uniqueIndex;not null"`
	Plan      string    `gorm:"size:50;not null;default:'free'"`
	IsActive  bool      `gorm:"not null;default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (TenantProjection) TableName() string {
	return "tenant_projections"
}

type GORMRepository struct {
	db *gorm.DB
}

func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db}
}

func (r *GORMRepository) FindByID(id string) mo.Option[Tenant] {
	var p TenantProjection
	if err := r.db.Where("id = ?", id).First(&p).Error; err != nil {
		return mo.None[Tenant]()
	}
	return mo.Some(toDomain(p))
}

func (r *GORMRepository) FindBySlug(slug string) mo.Option[Tenant] {
	var p TenantProjection
	if err := r.db.Where("slug = ?", slug).First(&p).Error; err != nil {
		return mo.None[Tenant]()
	}
	return mo.Some(toDomain(p))
}

func (r *GORMRepository) HasFeature(tenantID, feature string) bool {
	var count int64
	r.db.Table("tenant_feature_projections").
		Where("tenant_id = ? AND feature = ? AND is_enabled = true", tenantID, feature).
		Count(&count)
	return count > 0
}

func toDomain(p TenantProjection) Tenant {
	return Tenant{
		ID:        p.ID,
		Name:      p.Name,
		Slug:      p.Slug,
		Plan:      Plan(p.Plan),
		IsActive:  p.IsActive,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
