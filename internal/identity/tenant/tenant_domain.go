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
	EventCreated          = "tenant.created"
	EventActivated        = "tenant.activated"
	EventDeactivated      = "tenant.deactivated"
	EventFeatureEnabled   = "tenant.feature_enabled"
	EventFeatureDisabled  = "tenant.feature_disabled"
	EventPlanChanged      = "tenant.plan_changed"
)
