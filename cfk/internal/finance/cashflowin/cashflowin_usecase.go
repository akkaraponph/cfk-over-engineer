package cashflowin

import (
	"cfk/pkg/event"
	"cfk/pkg/saga"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/samber/mo"
)

var (
	ErrInvalidAmount = errors.New("invalid amount")
	ErrNotFound      = errors.New("cashflowin not found")
)

type Service struct {
	repo             Repository
	eventBus         *event.Bus
	sagaOrchestrator *saga.Orchestrator
}

func NewService(repo Repository, eventBus *event.Bus) *Service {
	return &Service{
		repo:     repo,
		eventBus: eventBus,
	}
}

func (s *Service) SetSagaOrchestrator(orchestrator *saga.Orchestrator) {
	s.sagaOrchestrator = orchestrator
}

func (s *Service) RecordCashflowIn(tenantID, userID, pocketID string, categoryID int, amount float64, description, receipt string) mo.Result[CashflowIn] {
	if amount <= 0 {
		return mo.Err[CashflowIn](ErrInvalidAmount)
	}

	id := uuid.New().String()
	now := time.Now()

	evt := event.Event{
		AggregateType: "cashflowin",
		AggregateID:   id,
		EventType:     EventRecorded,
		Version:       1,
		Payload: CashflowInRecordedPayload{
			ID:          id,
			TenantID:    tenantID,
			UserID:      userID,
			PocketID:    pocketID,
			CategoryID:  categoryID,
			Amount:      amount,
			Description: description,
			Receipt:     receipt,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[CashflowIn](r.Error())
	}

	if s.sagaOrchestrator != nil {
		sagaPayload := map[string]interface{}{
			"cashflowin_id": id,
			"pocket_id":     pocketID,
			"amount":        amount,
		}
		go s.sagaOrchestrator.Execute(context.Background(), "cashflowin", sagaPayload)
	}

	return mo.Ok(CashflowIn{
		ID:          id,
		TenantID:    tenantID,
		UserID:      userID,
		PocketID:    pocketID,
		CategoryID:  categoryID,
		Amount:      amount,
		Description: description,
		Receipt:     receipt,
		CreatedAt:   now,
		UpdatedAt:   now,
		IsDeleted:   false,
	})
}

func (s *Service) UpdateCashflowIn(id string, amount float64, description string, categoryID int, receipt string) mo.Result[CashflowIn] {
	if amount <= 0 {
		return mo.Err[CashflowIn](ErrInvalidAmount)
	}

	cfOpt := s.repo.FindByID(id)
	cf, ok := cfOpt.Get()
	if !ok {
		return mo.Err[CashflowIn](ErrNotFound)
	}

	now := time.Now()

	evt := event.Event{
		AggregateType: "cashflowin",
		AggregateID:   id,
		EventType:     EventUpdated,
		Version:       cf.Version + 1,
		Payload: CashflowInUpdatedPayload{
			ID:          id,
			Amount:      amount,
			Description: description,
			CategoryID:  categoryID,
			Receipt:     receipt,
			UpdatedAt:   now,
		},
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[CashflowIn](r.Error())
	}

	cf.Amount = amount
	cf.Description = description
	cf.CategoryID = categoryID
	cf.Receipt = receipt
	cf.UpdatedAt = now
	return mo.Ok(cf)
}

func (s *Service) DeleteCashflowIn(id string) mo.Result[CashflowIn] {
	cfOpt := s.repo.FindByID(id)
	cf, ok := cfOpt.Get()
	if !ok {
		return mo.Err[CashflowIn](ErrNotFound)
	}

	now := time.Now()

	evt := event.Event{
		AggregateType: "cashflowin",
		AggregateID:   id,
		EventType:     EventDeleted,
		Version:       cf.Version + 1,
		Payload: CashflowInDeletedPayload{
			ID:        id,
			UpdatedAt: now,
		},
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[CashflowIn](r.Error())
	}

	cf.IsDeleted = true
	cf.UpdatedAt = now
	return mo.Ok(cf)
}

func (s *Service) GetCashflowInByID(id string) mo.Result[CashflowIn] {
	cfOpt := s.repo.FindByID(id)
	cf, ok := cfOpt.Get()
	if !ok {
		return mo.Err[CashflowIn](ErrNotFound)
	}
	return mo.Ok(cf)
}

func (s *Service) ListCashflowInsByPocket(tenantID, pocketID string) mo.Result[[]CashflowIn] {
	return s.repo.FindByPocket(tenantID, pocketID)
}
