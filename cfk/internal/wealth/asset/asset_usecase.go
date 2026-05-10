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

	evt := event.Event{
		AggregateType: "asset",
		AggregateID:   id,
		EventType:     EventRecorded,
		Version:       1,
		Payload: AssetRecordedPayload{
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
		},
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

	evt := event.Event{
		AggregateType: "asset",
		AggregateID:   id,
		EventType:     EventValueChanged,
		Version:       asset.Version + 1,
		Payload: AssetValueChangedPayload{
			ID:              id,
			Value:           value,
			CashflowPerYear: cashflowPerYear,
			UpdatedAt:       now,
		},
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

	evt := event.Event{
		AggregateType: "asset",
		AggregateID:   id,
		EventType:     EventAssignedToBalanceSheet,
		Version:       asset.Version + 1,
		Payload: AssetAssignedToBalanceSheetPayload{
			ID:             id,
			BalanceSheetID: balanceSheetID,
			UpdatedAt:      now,
		},
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

	evt := event.Event{
		AggregateType: "asset",
		AggregateID:   id,
		EventType:     EventUnassignedFromBalanceSheet,
		Version:       asset.Version + 1,
		Payload: AssetUnassignedFromBalanceSheetPayload{
			ID:        id,
			UpdatedAt: now,
		},
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
