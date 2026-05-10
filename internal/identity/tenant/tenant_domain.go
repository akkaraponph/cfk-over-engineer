package tenant

import "time"

type Plan string

const (
	PlanFree       Plan = "free"
	PlanPremium    Plan = "premium"
	PlanEnterprise Plan = "enterprise"
)

type Feature string

const (
	FeatureBalanceSheet    Feature = "balance_sheet"
	FeatureDebt            Feature = "debt"
	FeatureAsset           Feature = "asset"
	FeatureAdvancedReport  Feature = "advanced_reporting"
	FeatureAPIClient       Feature = "api_access"
	FeatureCustomCategory  Feature = "custom_categories"
	FeatureMultiUser       Feature = "multi_user"
	FeatureTransfer        Feature = "transfer"
)

var PlanFeatures = map[Plan][]Feature{
	PlanFree: {
		FeatureCustomCategory,
	},
	PlanPremium: {
		FeatureBalanceSheet,
		FeatureDebt,
		FeatureAsset,
		FeatureAdvancedReport,
		FeatureCustomCategory,
		FeatureMultiUser,
		FeatureTransfer,
	},
	PlanEnterprise: {
		FeatureBalanceSheet,
		FeatureDebt,
		FeatureAsset,
		FeatureAdvancedReport,
		FeatureAPIClient,
		FeatureCustomCategory,
		FeatureMultiUser,
		FeatureTransfer,
	},
}

type Tenant struct {
	ID        string
	Name      string
	Slug      string
	Plan      Plan
	IsActive  bool
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TenantFeature struct {
	ID         string
	TenantID   string
	Feature    Feature
	IsEnabled  bool
	EnabledAt  time.Time
	DisabledAt time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

const (
	EventCreated         = "tenant.created"
	EventActivated       = "tenant.activated"
	EventDeactivated     = "tenant.deactivated"
	EventFeatureEnabled  = "tenant.feature_enabled"
	EventFeatureDisabled = "tenant.feature_disabled"
	EventPlanChanged     = "tenant.plan_changed"
)

type TenantCreatedPayload struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Plan      string    `json:"plan"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TenantActivatedPayload struct {
	ID        string    `json:"id"`
	IsActive  bool      `json:"is_active"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TenantDeactivatedPayload struct {
	ID        string    `json:"id"`
	IsActive  bool      `json:"is_active"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TenantPlanChangedPayload struct {
	ID        string    `json:"id"`
	Plan      string    `json:"plan"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TenantFeatureEnabledPayload struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Feature   string    `json:"feature"`
	IsEnabled bool      `json:"is_enabled"`
	EnabledBy string    `json:"enabled_by"`
	EnabledAt time.Time `json:"enabled_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TenantFeatureDisabledPayload struct {
	TenantID   string    `json:"tenant_id"`
	Feature    string    `json:"feature"`
	IsEnabled  bool      `json:"is_enabled"`
	DisabledBy string    `json:"disabled_by"`
	DisabledAt time.Time `json:"disabled_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
