package tenant

import (
	"cfk/pkg/event"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/samber/mo"
)

var (
	ErrInvalidName = errors.New("invalid tenant name")
	ErrInvalidSlug = errors.New("invalid tenant slug")
	ErrNotFound    = errors.New("tenant not found")
)

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

func (s *Service) CreateTenant(name, slug string) mo.Result[Tenant] {
	if name == "" {
		return mo.Err[Tenant](ErrInvalidName)
	}
	if slug == "" {
		return mo.Err[Tenant](ErrInvalidSlug)
	}

	id := uuid.New().String()
	now := time.Now()

	evt := event.Event{
		AggregateType: "tenant",
		AggregateID:   id,
		EventType:     EventCreated,
		Version:       1,
		Payload: map[string]interface{}{
			"id":         id,
			"name":       name,
			"slug":       slug,
			"is_active":  true,
			"created_at": now,
			"updated_at": now,
		},
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if err := s.eventBus.Publish(evt); err != nil {
		return mo.Err[Tenant](err)
	}

	return mo.Ok(Tenant{
		ID:        id,
		Name:      name,
		Slug:      slug,
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
		Version:       1,
		Payload: map[string]interface{}{
			"id":         id,
			"is_active":  active,
			"updated_at": now,
		},
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if err := s.eventBus.Publish(evt); err != nil {
		return mo.Err[Tenant](err)
	}

	t.IsActive = active
	t.UpdatedAt = now
	return mo.Ok(t)
}
