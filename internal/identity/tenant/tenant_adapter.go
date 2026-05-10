package tenant

import (
	"time"

	"github.com/samber/mo"
	"gorm.io/gorm"
)

type TenantProjection struct {
	ID        string `gorm:"type:uuid;primaryKey"`
	Name      string `gorm:"size:255;not null"`
	Slug      string `gorm:"size:255;uniqueIndex;not null"`
	IsActive  bool   `gorm:"not null;default:true"`
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
	return mo.Some(Tenant{
		ID:        p.ID,
		Name:      p.Name,
		Slug:      p.Slug,
		IsActive:  p.IsActive,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	})
}

func (r *GORMRepository) FindBySlug(slug string) mo.Option[Tenant] {
	var p TenantProjection
	if err := r.db.Where("slug = ?", slug).First(&p).Error; err != nil {
		return mo.None[Tenant]()
	}
	return mo.Some(Tenant{
		ID:        p.ID,
		Name:      p.Name,
		Slug:      p.Slug,
		IsActive:  p.IsActive,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	})
}
