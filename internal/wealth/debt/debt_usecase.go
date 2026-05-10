package debt

import (
	"cfk/pkg/event"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/samber/mo"
)

var (
	ErrInvalidType   = errors.New("invalid debt type")
	ErrInvalidAmount = errors.New("invalid debt amount")
	ErrNotFound      = errors.New("debt not found")
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

func (s *Service) RecordDebt(tenantID, debtType, description, userID string, amount, interest, minimumPay float64, priority int) mo.Result[Debt] {
	if debtType == "" {
		return mo.Err[Debt](ErrInvalidType)
	}
	if amount < 0 {
		return mo.Err[Debt](ErrInvalidAmount)
	}

	id := uuid.New().String()
	now := time.Now()

	evt := event.Event{
		AggregateType: "debt",
		AggregateID:   id,
		EventType:     EventRecorded,
		Version:       1,
		Payload: DebtRecordedPayload{
			ID:             id,
			TenantID:       tenantID,
			Type:           debtType,
			Description:    description,
			Amount:         amount,
			Interest:       interest,
			MinimumPay:     minimumPay,
			Priority:       priority,
			BalanceSheetID: "",
			UserID:         userID,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[Debt](r.Error())
	}

	return mo.Ok(Debt{
		ID:             id,
		TenantID:       tenantID,
		Type:           debtType,
		Description:    description,
		Amount:         amount,
		Interest:       interest,
		MinimumPay:     minimumPay,
		Priority:       priority,
		BalanceSheetID: "",
		UserID:         userID,
		CreatedAt:      now,
		UpdatedAt:      now,
	})
}

func (s *Service) ChangeAmount(id string, amount float64) mo.Result[Debt] {
	if amount < 0 {
		return mo.Err[Debt](ErrInvalidAmount)
	}

	debtOpt := s.repo.FindByID(id)
	debt, ok := debtOpt.Get()
	if !ok {
		return mo.Err[Debt](ErrNotFound)
	}

	now := time.Now()

	evt := event.Event{
		AggregateType: "debt",
		AggregateID:   id,
		EventType:     EventAmountChanged,
		Version:       debt.Version + 1,
		Payload: DebtAmountChangedPayload{
			ID:        id,
			Amount:    amount,
			UpdatedAt: now,
		},
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[Debt](r.Error())
	}

	debt.Amount = amount
	debt.UpdatedAt = now
	return mo.Ok(debt)
}

func (s *Service) AssignToBalanceSheet(id, balanceSheetID string) mo.Result[Debt] {
	debtOpt := s.repo.FindByID(id)
	debt, ok := debtOpt.Get()
	if !ok {
		return mo.Err[Debt](ErrNotFound)
	}

	now := time.Now()

	evt := event.Event{
		AggregateType: "debt",
		AggregateID:   id,
		EventType:     EventAssignedToBalanceSheet,
		Version:       debt.Version + 1,
		Payload: DebtAssignedToBalanceSheetPayload{
			ID:             id,
			BalanceSheetID: balanceSheetID,
			UpdatedAt:      now,
		},
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[Debt](r.Error())
	}

	debt.BalanceSheetID = balanceSheetID
	debt.UpdatedAt = now
	return mo.Ok(debt)
}

func (s *Service) UnassignFromBalanceSheet(id string) mo.Result[Debt] {
	debtOpt := s.repo.FindByID(id)
	debt, ok := debtOpt.Get()
	if !ok {
		return mo.Err[Debt](ErrNotFound)
	}

	now := time.Now()

	evt := event.Event{
		AggregateType: "debt",
		AggregateID:   id,
		EventType:     EventUnassignedFromBalanceSheet,
		Version:       debt.Version + 1,
		Payload: DebtUnassignedFromBalanceSheetPayload{
			ID:        id,
			UpdatedAt: now,
		},
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[Debt](r.Error())
	}

	debt.BalanceSheetID = ""
	debt.UpdatedAt = now
	return mo.Ok(debt)
}

func (s *Service) GetDebtByID(id string) mo.Result[Debt] {
	debtOpt := s.repo.FindByID(id)
	debt, ok := debtOpt.Get()
	if !ok {
		return mo.Err[Debt](ErrNotFound)
	}
	return mo.Ok(debt)
}
