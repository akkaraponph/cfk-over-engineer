package balancesheet

import (
	"cfk/pkg/event"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/samber/mo"
)

var (
	ErrInvalidYear = errors.New("invalid year")
	ErrNotFound    = errors.New("balancesheet not found")
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

func (s *Service) CreateBalanceSheet(tenantID, userID string, year int) mo.Result[BalanceSheet] {
	if year < 2000 || year > 2100 {
		return mo.Err[BalanceSheet](ErrInvalidYear)
	}

	id := uuid.New().String()
	now := time.Now()

	eventPayload := map[string]interface{}{
		"id":         id,
		"tenant_id":  tenantID,
		"user_id":    userID,
		"year":       year,
		"created_at": now,
		"updated_at": now,
	}

	evt := event.Event{
		AggregateType: "balancesheet",
		AggregateID:   id,
		EventType:     EventCreated,
		Version:       1,
		Payload:       eventPayload,
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[BalanceSheet](r.Error())
	}

	return mo.Ok(BalanceSheet{
		ID:        id,
		TenantID:  tenantID,
		UserID:    userID,
		Year:      year,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func (s *Service) UpdateBalanceSheet(id string, year int) mo.Result[BalanceSheet] {
	if year < 2000 || year > 2100 {
		return mo.Err[BalanceSheet](ErrInvalidYear)
	}

	bsOpt := s.repo.FindByID(id)
	bs, ok := bsOpt.Get()
	if !ok {
		return mo.Err[BalanceSheet](ErrNotFound)
	}

	now := time.Now()

	eventPayload := map[string]interface{}{
		"id":         id,
		"year":       year,
		"updated_at": now,
	}

	evt := event.Event{
		AggregateType: "balancesheet",
		AggregateID:   id,
		EventType:     EventUpdated,
		Version:       1,
		Payload:       eventPayload,
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[BalanceSheet](r.Error())
	}

	bs.Year = year
	bs.UpdatedAt = now
	return mo.Ok(bs)
}

func (s *Service) GetBalanceSheetByID(id string) mo.Result[BalanceSheet] {
	bsOpt := s.repo.FindByID(id)
	bs, ok := bsOpt.Get()
	if !ok {
		return mo.Err[BalanceSheet](ErrNotFound)
	}
	return mo.Ok(bs)
}
