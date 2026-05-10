package pocket

import (
	"cfk/pkg/event"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/samber/mo"
)

var (
	ErrInvalidName = errors.New("invalid pocket name")
	ErrNotFound    = errors.New("pocket not found")
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

func (s *Service) CreatePocket(tenantID, name, userID string) mo.Result[Pocket] {
	if name == "" {
		return mo.Err[Pocket](ErrInvalidName)
	}

	id := uuid.New().String()
	now := time.Now()

	eventPayload := map[string]interface{}{
		"id":         id,
		"tenant_id":  tenantID,
		"name":       name,
		"balance":    0.0,
		"user_id":    userID,
		"created_at": now,
		"updated_at": now,
	}

	evt := event.Event{
		AggregateType: "pocket",
		AggregateID:   id,
		EventType:     EventCreated,
		Version:       1,
		Payload:       eventPayload,
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[Pocket](r.Error())
	}

	return mo.Ok(Pocket{
		ID:        id,
		TenantID:  tenantID,
		Name:      name,
		Balance:   0.0,
		UserID:    userID,
		CreatedAt: now,
		UpdatedAt: now,
		IsDeleted: false,
	})
}

func (s *Service) ChangeName(id, name string) mo.Result[Pocket] {
	if name == "" {
		return mo.Err[Pocket](ErrInvalidName)
	}

	pocketOpt := s.repo.FindByID(id)
	pocket, ok := pocketOpt.Get()
	if !ok {
		return mo.Err[Pocket](ErrNotFound)
	}

	now := time.Now()
	eventPayload := map[string]interface{}{
		"id":         id,
		"name":       name,
		"updated_at": now,
	}

	evt := event.Event{
		AggregateType: "pocket",
		AggregateID:   id,
		EventType:     EventNameChanged,
		Version:       1,
		Payload:       eventPayload,
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[Pocket](r.Error())
	}

	pocket.Name = name
	pocket.UpdatedAt = now
	return mo.Ok(pocket)
}

func (s *Service) ChangeBalance(id string, amount float64) mo.Result[Pocket] {
	pocketOpt := s.repo.FindByID(id)
	pocket, ok := pocketOpt.Get()
	if !ok {
		return mo.Err[Pocket](ErrNotFound)
	}

	now := time.Now()
	newBalance := pocket.Balance + amount

	eventPayload := map[string]interface{}{
		"id":          id,
		"amount":      amount,
		"new_balance": newBalance,
		"updated_at":  now,
	}

	evt := event.Event{
		AggregateType: "pocket",
		AggregateID:   id,
		EventType:     EventBalanceChanged,
		Version:       1,
		Payload:       eventPayload,
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[Pocket](r.Error())
	}

	pocket.Balance = newBalance
	pocket.UpdatedAt = now
	return mo.Ok(pocket)
}

func (s *Service) DeletePocket(id string) mo.Result[Pocket] {
	pocketOpt := s.repo.FindByID(id)
	pocket, ok := pocketOpt.Get()
	if !ok {
		return mo.Err[Pocket](ErrNotFound)
	}

	now := time.Now()
	eventPayload := map[string]interface{}{
		"id":         id,
		"updated_at": now,
	}

	evt := event.Event{
		AggregateType: "pocket",
		AggregateID:   id,
		EventType:     EventDeleted,
		Version:       1,
		Payload:       eventPayload,
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[Pocket](r.Error())
	}

	pocket.IsDeleted = true
	pocket.UpdatedAt = now
	return mo.Ok(pocket)
}

func (s *Service) GetPocketByID(id string) mo.Result[Pocket] {
	pocketOpt := s.repo.FindByID(id)
	pocket, ok := pocketOpt.Get()
	if !ok {
		return mo.Err[Pocket](ErrNotFound)
	}
	return mo.Ok(pocket)
}

func (s *Service) ListPocketsByUser(tenantID, userID string) mo.Result[[]Pocket] {
	return s.repo.FindByUser(tenantID, userID)
}
