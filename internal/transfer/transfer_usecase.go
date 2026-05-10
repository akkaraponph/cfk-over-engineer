package transfer

import (
	"cfk/pkg/event"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/samber/mo"
)

var (
	ErrInvalidAmount = errors.New("invalid amount")
	ErrSamePocket    = errors.New("cannot transfer to same pocket")
	ErrNotFound      = errors.New("transfer not found")
	ErrInvalidStatus = errors.New("transfer cannot be completed or failed in current status")
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

func (s *Service) InitiateTransfer(tenantID, userID, fromPocketID, toPocketID string, amount float64) mo.Result[Transfer] {
	if amount <= 0 {
		return mo.Err[Transfer](ErrInvalidAmount)
	}
	if fromPocketID == toPocketID {
		return mo.Err[Transfer](ErrSamePocket)
	}

	id := uuid.New().String()
	now := time.Now()

	eventPayload := map[string]interface{}{
		"id":             id,
		"tenant_id":      tenantID,
		"user_id":        userID,
		"from_pocket_id": fromPocketID,
		"to_pocket_id":   toPocketID,
		"amount":         amount,
		"status":         "pending",
		"created_at":     now,
		"updated_at":     now,
	}

	evt := event.Event{
		AggregateType: "transfer",
		AggregateID:   id,
		EventType:     EventInitiated,
		Version:       1,
		Payload:       eventPayload,
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if err := s.eventBus.Publish(evt); err != nil {
		return mo.Err[Transfer](err)
	}

	return mo.Ok(Transfer{
		ID:           id,
		TenantID:     tenantID,
		UserID:       userID,
		FromPocketID: fromPocketID,
		ToPocketID:   toPocketID,
		Amount:       amount,
		Status:       "pending",
		CreatedAt:    now,
		UpdatedAt:    now,
		IsDeleted:    false,
	})
}

func (s *Service) CompleteTransfer(id string) mo.Result[Transfer] {
	tOpt := s.repo.FindByID(id)
	t, ok := tOpt.Get()
	if !ok {
		return mo.Err[Transfer](ErrNotFound)
	}
	if t.Status != "pending" {
		return mo.Err[Transfer](ErrInvalidStatus)
	}

	now := time.Now()

	eventPayload := map[string]interface{}{
		"id":         id,
		"status":     "completed",
		"updated_at": now,
	}

	evt := event.Event{
		AggregateType: "transfer",
		AggregateID:   id,
		EventType:     EventCompleted,
		Version:       1,
		Payload:       eventPayload,
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if err := s.eventBus.Publish(evt); err != nil {
		return mo.Err[Transfer](err)
	}

	t.Status = "completed"
	t.UpdatedAt = now
	return mo.Ok(t)
}

func (s *Service) FailTransfer(id string, reason string) mo.Result[Transfer] {
	tOpt := s.repo.FindByID(id)
	t, ok := tOpt.Get()
	if !ok {
		return mo.Err[Transfer](ErrNotFound)
	}
	if t.Status != "pending" {
		return mo.Err[Transfer](ErrInvalidStatus)
	}

	now := time.Now()

	eventPayload := map[string]interface{}{
		"id":         id,
		"status":     "failed",
		"reason":     reason,
		"updated_at": now,
	}

	evt := event.Event{
		AggregateType: "transfer",
		AggregateID:   id,
		EventType:     EventFailed,
		Version:       1,
		Payload:       eventPayload,
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if err := s.eventBus.Publish(evt); err != nil {
		return mo.Err[Transfer](err)
	}

	t.Status = "failed"
	t.UpdatedAt = now
	return mo.Ok(t)
}

func (s *Service) DeleteTransfer(id string) mo.Result[Transfer] {
	tOpt := s.repo.FindByID(id)
	t, ok := tOpt.Get()
	if !ok {
		return mo.Err[Transfer](ErrNotFound)
	}

	now := time.Now()

	eventPayload := map[string]interface{}{
		"id":         id,
		"updated_at": now,
	}

	evt := event.Event{
		AggregateType: "transfer",
		AggregateID:   id,
		EventType:     EventDeleted,
		Version:       1,
		Payload:       eventPayload,
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if err := s.eventBus.Publish(evt); err != nil {
		return mo.Err[Transfer](err)
	}

	t.IsDeleted = true
	t.UpdatedAt = now
	return mo.Ok(t)
}

func (s *Service) GetTransferByID(id string) mo.Result[Transfer] {
	tOpt := s.repo.FindByID(id)
	t, ok := tOpt.Get()
	if !ok {
		return mo.Err[Transfer](ErrNotFound)
	}
	return mo.Ok(t)
}
