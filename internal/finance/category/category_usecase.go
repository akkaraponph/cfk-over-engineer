package category

import (
	"cfk/pkg/event"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/samber/mo"
)

var (
	ErrInvalidName = errors.New("invalid category name")
	ErrNotFound    = errors.New("category not found")
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

func (s *Service) CreateCategory(tenantID, name, description, catType string, isCustom bool, userID string) mo.Result[Category] {
	if name == "" {
		return mo.Err[Category](ErrInvalidName)
	}

	id := uuid.New().String()
	now := time.Now()

	evt := event.Event{
		AggregateType: "category",
		AggregateID:   id,
		EventType:     EventCreated,
		Version:       1,
		Payload: CategoryCreatedPayload{
			ID:          id,
			TenantID:    tenantID,
			Name:        name,
			Description: description,
			Type:        catType,
			IsCustom:    isCustom,
			UserID:      userID,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[Category](r.Error())
	}

	return mo.Ok(Category{
		ID:          id,
		TenantID:    tenantID,
		Name:        name,
		Description: description,
		Type:        catType,
		IsCustom:    isCustom,
		UserID:      userID,
		CreatedAt:   now,
		UpdatedAt:   now,
		IsDeleted:   false,
	})
}

func (s *Service) UpdateCategory(id, name, description string) mo.Result[Category] {
	if name == "" {
		return mo.Err[Category](ErrInvalidName)
	}

	catOpt := s.repo.FindByID(id)
	cat, ok := catOpt.Get()
	if !ok {
		return mo.Err[Category](ErrNotFound)
	}

	now := time.Now()

	evt := event.Event{
		AggregateType: "category",
		AggregateID:   id,
		EventType:     EventUpdated,
		Version:       cat.Version + 1,
		Payload: CategoryUpdatedPayload{
			ID:          id,
			Name:        name,
			Description: description,
			UpdatedAt:   now,
		},
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[Category](r.Error())
	}

	cat.Name = name
	cat.Description = description
	cat.UpdatedAt = now
	return mo.Ok(cat)
}

func (s *Service) DeleteCategory(id string) mo.Result[Category] {
	catOpt := s.repo.FindByID(id)
	cat, ok := catOpt.Get()
	if !ok {
		return mo.Err[Category](ErrNotFound)
	}

	now := time.Now()

	evt := event.Event{
		AggregateType: "category",
		AggregateID:   id,
		EventType:     EventDeleted,
		Version:       cat.Version + 1,
		Payload: CategoryDeletedPayload{
			ID:        id,
			UpdatedAt: now,
		},
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[Category](r.Error())
	}

	cat.IsDeleted = true
	cat.UpdatedAt = now
	return mo.Ok(cat)
}

func (s *Service) GetCategoryByID(id string) mo.Result[Category] {
	catOpt := s.repo.FindByID(id)
	cat, ok := catOpt.Get()
	if !ok {
		return mo.Err[Category](ErrNotFound)
	}
	return mo.Ok(cat)
}

func (s *Service) ListCategoriesByTenant(tenantID string) mo.Result[[]Category] {
	return s.repo.FindByTenant(tenantID)
}
