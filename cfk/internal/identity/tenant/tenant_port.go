package tenant

import (
	"github.com/samber/mo"
)

type Repository interface {
	FindByID(id string) mo.Option[Tenant]
	FindBySlug(slug string) mo.Option[Tenant]
	HasFeature(tenantID, feature string) bool
}
