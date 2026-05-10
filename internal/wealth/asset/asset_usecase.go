package asset

import (
	"cfk/pkg/event"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/samber/mo"
)

var (
	ErrInvalidType  = errors.New("invalid asset type")
	ErrInvalidValue = errors.New("invalid asset value")
	ErrNotFound     = errors.New("asset not found")
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

func (s *Service) RecordAsset(tenantID, assetType, description, userID string, value, cashflowPerYear float64) mo.Result[Asset] {
	if assetType == "" {
		return mo.Err[Asset](ErrInvalidType)
	}
	if value < 0 {
		return mo.Err[Asset](ErrInvalidValue)
	}

	id := uuid.New().String()
	now := time.Now()

	eventPayload := map[string]interface{}{
		"id":                id,
		"tenant_id":         tenantID,
		"type":              assetType,
		"description":       description,
		"value":             value,
		"cashflow_per_year": cashflowPerYear,
		"balance_sheet_id":  "",
		"user_id":           userID,
		"created_at":        now,
		"updated_at":        now,
	}

	evt := event.Event{
		AggregateType: "asset",
		AggregateID:   id,
		EventType:     EventRecorded,
		Version:       1,
		Payload:       eventPayload,
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[Asset](r.Error())
	}

	return mo.Ok(Asset{
		ID:              id,
		TenantID:        tenantID,
		Type:            assetType,
		Description:     description,
		Value:           value,
		CashflowPerYear: cashflowPerYear,
		BalanceSheetID:  "",
		UserID:          userID,
		CreatedAt:       now,
		UpdatedAt:       now,
	})
}

func (s *Service) ChangeValue(id string, value, cashflowPerYear float64) mo.Result[Asset] {
	if value < 0 {
		return mo.Err[Asset](ErrInvalidValue)
	}

	assetOpt := s.repo.FindByID(id)
	asset, ok := assetOpt.Get()
	if !ok {
		return mo.Err[Asset](ErrNotFound)
	}

	now := time.Now()

	eventPayload := map[string]interface{}{
		"id":                id,
		"value":             value,
		"cashflow_per_year": cashflowPerYear,
		"updated_at":        now,
	}

	evt := event.Event{
		AggregateType: "asset",
		AggregateID:   id,
		EventType:     EventValueChanged,
		Version:       1,
		Payload:       eventPayload,
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[Asset](r.Error())
	}

	asset.Value = value
	asset.CashflowPerYear = cashflowPerYear
	asset.UpdatedAt = now
	return mo.Ok(asset)
}

func (s *Service) AssignToBalanceSheet(id, balanceSheetID string) mo.Result[Asset] {
	assetOpt := s.repo.FindByID(id)
	asset, ok := assetOpt.Get()
	if !ok {
		return mo.Err[Asset](ErrNotFound)
	}

	now := time.Now()

	eventPayload := map[string]interface{}{
		"id":               id,
		"balance_sheet_id": balanceSheetID,
		"updated_at":       now,
	}

	evt := event.Event{
		AggregateType: "asset",
		AggregateID:   id,
		EventType:     EventAssignedToBalanceSheet,
		Version:       1,
		Payload:       eventPayload,
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[Asset](r.Error())
	}

	asset.BalanceSheetID = balanceSheetID
	asset.UpdatedAt = now
	return mo.Ok(asset)
}

func (s *Service) UnassignFromBalanceSheet(id string) mo.Result[Asset] {
	assetOpt := s.repo.FindByID(id)
	asset, ok := assetOpt.Get()
	if !ok {
		return mo.Err[Asset](ErrNotFound)
	}

	now := time.Now()

	eventPayload := map[string]interface{}{
		"id":         id,
		"updated_at": now,
	}

	evt := event.Event{
		AggregateType: "asset",
		AggregateID:   id,
		EventType:     EventUnassignedFromBalanceSheet,
		Version:       1,
		Payload:       eventPayload,
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[Asset](r.Error())
	}

	asset.BalanceSheetID = ""
	asset.UpdatedAt = now
	return mo.Ok(asset)
}

func (s *Service) GetAssetByID(id string) mo.Result[Asset] {
	assetOpt := s.repo.FindByID(id)
	asset, ok := assetOpt.Get()
	if !ok {
		return mo.Err[Asset](ErrNotFound)
	}
	return mo.Ok(asset)
}
