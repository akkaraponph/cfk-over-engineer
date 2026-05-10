package cashflowout

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
	ErrInvalidType   = errors.New("invalid cashflowout type")
	ErrNotFound      = errors.New("cashflowout not found")
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

func (s *Service) RecordCashflowOut(tenantID, userID, pocketID string, categoryID int, amount float64, description, receipt, outType string) mo.Result[CashflowOut] {
	if amount <= 0 {
		return mo.Err[CashflowOut](ErrInvalidAmount)
	}
	if outType == "" {
		return mo.Err[CashflowOut](ErrInvalidType)
	}

	id := uuid.New().String()
	now := time.Now()

	evt := event.Event{
		AggregateType: "cashflowout",
		AggregateID:   id,
		EventType:     EventRecorded,
		Version:       1,
		Payload: CashflowOutRecordedPayload{
			ID:          id,
			TenantID:    tenantID,
			UserID:      userID,
			PocketID:    pocketID,
			CategoryID:  categoryID,
			Amount:      amount,
			Description: description,
			Receipt:     receipt,
			Type:        outType,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[CashflowOut](r.Error())
	}

	if s.sagaOrchestrator != nil {
		sagaPayload := map[string]interface{}{
			"cashflowout_id": id,
			"pocket_id":      pocketID,
			"amount":         amount,
		}
		go s.sagaOrchestrator.Execute(context.Background(), "cashflowout", sagaPayload)
	}

	return mo.Ok(CashflowOut{
		ID:          id,
		TenantID:    tenantID,
		UserID:      userID,
		PocketID:    pocketID,
		CategoryID:  categoryID,
		Amount:      amount,
		Description: description,
		Type:        outType,
		Receipt:     receipt,
		CreatedAt:   now,
		UpdatedAt:   now,
		IsDeleted:   false,
	})
}

func (s *Service) UpdateCashflowOut(id string, amount float64, description string, categoryID int, receipt string) mo.Result[CashflowOut] {
	if amount <= 0 {
		return mo.Err[CashflowOut](ErrInvalidAmount)
	}

	cfOpt := s.repo.FindByID(id)
	cf, ok := cfOpt.Get()
	if !ok {
		return mo.Err[CashflowOut](ErrNotFound)
	}

	now := time.Now()

	evt := event.Event{
		AggregateType: "cashflowout",
		AggregateID:   id,
		EventType:     EventUpdated,
		Version:       cf.Version + 1,
		Payload: CashflowOutUpdatedPayload{
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
		return mo.Err[CashflowOut](r.Error())
	}

	cf.Amount = amount
	cf.Description = description
	cf.CategoryID = categoryID
	cf.Receipt = receipt
	cf.UpdatedAt = now
	return mo.Ok(cf)
}

func (s *Service) DeleteCashflowOut(id string) mo.Result[CashflowOut] {
	cfOpt := s.repo.FindByID(id)
	cf, ok := cfOpt.Get()
	if !ok {
		return mo.Err[CashflowOut](ErrNotFound)
	}

	now := time.Now()

	evt := event.Event{
		AggregateType: "cashflowout",
		AggregateID:   id,
		EventType:     EventDeleted,
		Version:       cf.Version + 1,
		Payload: CashflowOutDeletedPayload{
			ID:        id,
			UpdatedAt: now,
		},
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[CashflowOut](r.Error())
	}

	cf.IsDeleted = true
	cf.UpdatedAt = now
	return mo.Ok(cf)
}

func (s *Service) GetCashflowOutByID(id string) mo.Result[CashflowOut] {
	cfOpt := s.repo.FindByID(id)
	cf, ok := cfOpt.Get()
	if !ok {
		return mo.Err[CashflowOut](ErrNotFound)
	}
	return mo.Ok(cf)
}

func (s *Service) ListCashflowOutsByPocket(tenantID, pocketID string) mo.Result[[]CashflowOut] {
	return s.repo.FindByPocket(tenantID, pocketID)
}
