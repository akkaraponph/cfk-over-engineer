package tenant

import (
	"cfk/pkg/event"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/samber/mo"
)

var (
	ErrInvalidName    = errors.New("invalid tenant name")
	ErrInvalidSlug    = errors.New("invalid tenant slug")
	ErrInvalidPlan    = errors.New("invalid tenant plan")
	ErrNotFound       = errors.New("tenant not found")
	ErrFeatureDenied  = errors.New("feature not enabled for tenant")
)

var validPlans = map[Plan]bool{
	PlanFree:       true,
	PlanPremium:    true,
	PlanEnterprise: true,
}

type Service struct {
	repo     Repository
	eventBus *event.Bus
}

func NewService(repo Repository, eventBus *event.Bus) *Service {
	return &Service{
		repo:     repo,
		eventBus: eventBus,
	}
}

func (s *Service) CreateTenant(name, slug, plan string) mo.Result[Tenant] {
	if name == "" {
		return mo.Err[Tenant](ErrInvalidName)
	}
	if slug == "" {
		return mo.Err[Tenant](ErrInvalidSlug)
	}
	if !validPlans[Plan(plan)] {
		return mo.Err[Tenant](ErrInvalidPlan)
	}

	id := uuid.New().String()
	now := time.Now()
	tenantPlan := Plan(plan)

	evt := event.Event{
		AggregateType: "tenant",
		AggregateID:   id,
		EventType:     EventCreated,
		Version:       1,
		Payload: TenantCreatedPayload{
			ID:        id,
			Name:      name,
			Slug:      slug,
			Plan:      string(tenantPlan),
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[Tenant](r.Error())
	}

	return mo.Ok(Tenant{
		ID:        id,
		Name:      name,
		Slug:      slug,
		Plan:      tenantPlan,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func (s *Service) ActivateTenant(id string) mo.Result[Tenant] {
	return s.updateStatus(id, true, EventActivated)
}

func (s *Service) DeactivateTenant(id string) mo.Result[Tenant] {
	return s.updateStatus(id, false, EventDeactivated)
}

func (s *Service) ChangePlan(id, plan string) mo.Result[Tenant] {
	if !validPlans[Plan(plan)] {
		return mo.Err[Tenant](ErrInvalidPlan)
	}

	opt := s.repo.FindByID(id)
	t, ok := opt.Get()
	if !ok {
		return mo.Err[Tenant](ErrNotFound)
	}

	now := time.Now()
	evt := event.Event{
		AggregateType: "tenant",
		AggregateID:   id,
		EventType:     EventPlanChanged,
		Version:       t.Version + 1,
		Payload: TenantPlanChangedPayload{
			ID:        id,
			Plan:      plan,
			UpdatedAt: now,
		},
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[Tenant](r.Error())
	}

	t.Plan = Plan(plan)
	t.UpdatedAt = now
	return mo.Ok(t)
}

func (s *Service) EnableFeature(tenantID, feature, userID string) mo.Result[TenantFeature] {
	opt := s.repo.FindByID(tenantID)
	t, ok := opt.Get()
	if !ok {
		return mo.Err[TenantFeature](ErrNotFound)
	}

	id := uuid.New().String()
	now := time.Now()

	evt := event.Event{
		AggregateType: "tenant",
		AggregateID:   tenantID,
		EventType:     EventFeatureEnabled,
		Version:       t.Version + 1,
		Payload: TenantFeatureEnabledPayload{
			ID:        id,
			TenantID:  tenantID,
			Feature:   feature,
			IsEnabled: true,
			EnabledBy: userID,
			EnabledAt: now,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[TenantFeature](r.Error())
	}

	return mo.Ok(TenantFeature{
		ID:        id,
		TenantID:  tenantID,
		Feature:   Feature(feature),
		IsEnabled: true,
		EnabledAt: now,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func (s *Service) DisableFeature(tenantID, feature, userID string) mo.Result[TenantFeature] {
	opt := s.repo.FindByID(tenantID)
	t, ok := opt.Get()
	if !ok {
		return mo.Err[TenantFeature](ErrNotFound)
	}

	now := time.Now()

	evt := event.Event{
		AggregateType: "tenant",
		AggregateID:   tenantID,
		EventType:     EventFeatureDisabled,
		Version:       t.Version + 1,
		Payload: TenantFeatureDisabledPayload{
			TenantID:   tenantID,
			Feature:    feature,
			IsEnabled:  false,
			DisabledBy: userID,
			DisabledAt: now,
			UpdatedAt:  now,
		},
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[TenantFeature](r.Error())
	}

	return mo.Ok(TenantFeature{
		TenantID:   tenantID,
		Feature:    Feature(feature),
		IsEnabled:  false,
		DisabledAt: now,
		UpdatedAt:  now,
	})
}

func (s *Service) HasFeature(tenantID, feature string) bool {
	return s.repo.HasFeature(tenantID, feature)
}

func (s *Service) GetTenantBySlug(slug string) mo.Result[Tenant] {
	opt := s.repo.FindBySlug(slug)
	t, ok := opt.Get()
	if !ok {
		return mo.Err[Tenant](ErrNotFound)
	}
	return mo.Ok(t)
}

func (s *Service) updateStatus(id string, active bool, eventType string) mo.Result[Tenant] {
	opt := s.repo.FindByID(id)
	t, ok := opt.Get()
	if !ok {
		return mo.Err[Tenant](ErrNotFound)
	}

	now := time.Now()
	evt := event.Event{
		AggregateType: "tenant",
		AggregateID:   id,
		EventType:     eventType,
		Version:       t.Version + 1,
		Payload: TenantActivatedPayload{
			ID:        id,
			IsActive:  active,
			UpdatedAt: now,
		},
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[Tenant](r.Error())
	}

	t.IsActive = active
	t.UpdatedAt = now
	return mo.Ok(t)
}
