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

	eventPayload := map[string]interface{}{
		"id":          id,
		"tenant_id":   tenantID,
		"user_id":     userID,
		"pocket_id":   pocketID,
		"category_id": categoryID,
		"amount":      amount,
		"description": description,
		"receipt":     receipt,
		"created_at":  now,
		"updated_at":  now,
	}

	evt := event.Event{
		AggregateType: "cashflowin",
		AggregateID:   id,
		EventType:     EventRecorded,
		Version:       1,
		Payload:       eventPayload,
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if err := s.eventBus.Publish(evt); err != nil {
		return mo.Err[CashflowIn](err)
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
	eventPayload := map[string]interface{}{
		"id":          id,
		"amount":      amount,
		"description": description,
		"category_id": categoryID,
		"receipt":     receipt,
		"updated_at":  now,
	}

	evt := event.Event{
		AggregateType: "cashflowin",
		AggregateID:   id,
		EventType:     EventUpdated,
		Version:       1,
		Payload:       eventPayload,
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if err := s.eventBus.Publish(evt); err != nil {
		return mo.Err[CashflowIn](err)
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
	eventPayload := map[string]interface{}{
		"id":         id,
		"updated_at": now,
	}

	evt := event.Event{
		AggregateType: "cashflowin",
		AggregateID:   id,
		EventType:     EventDeleted,
		Version:       1,
		Payload:       eventPayload,
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if err := s.eventBus.Publish(evt); err != nil {
		return mo.Err[CashflowIn](err)
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
