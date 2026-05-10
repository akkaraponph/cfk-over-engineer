package cashflowout

import (
	"cfk/pkg/event"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/samber/mo"
)

type mockCashflowOutRepo struct {
	items map[string]CashflowOut
}

func newMockCashflowOutRepo() *mockCashflowOutRepo {
	return &mockCashflowOutRepo{
		items: make(map[string]CashflowOut),
	}
}

func (r *mockCashflowOutRepo) AppendEvent(eventType string, aggregateID string, payload map[string]interface{}, metadata map[string]interface{}) error {
	return nil
}

func (r *mockCashflowOutRepo) FindByID(id string) mo.Option[CashflowOut] {
	if cf, ok := r.items[id]; ok {
		return mo.Some(cf)
	}
	return mo.None[CashflowOut]()
}

func (r *mockCashflowOutRepo) FindByPocket(tenantID, pocketID string) mo.Result[[]CashflowOut] {
	var result []CashflowOut
	for _, cf := range r.items {
		if cf.TenantID == tenantID && cf.PocketID == pocketID && !cf.IsDeleted {
			result = append(result, cf)
		}
	}
	return mo.Ok(result)
}

func setupTestBus(t *testing.T) (*event.Bus, *[]event.Event) {
	t.Helper()
	recorded := &[]event.Event{}
	bus := event.NewBus(event.WithWorkerPool(1), event.WithBufferSize(64))
	bus.Subscribe("*", func(evt event.Event) mo.Result[struct{}] {
		*recorded = append(*recorded, evt)
		return event.OkHandle()
	})
	ctx := context.Background()
	bus.Start(ctx)
	t.Cleanup(func() {
		bus.Stop()
	})
	return bus, recorded
}

func waitForEvents() {
	time.Sleep(50 * time.Millisecond)
}

func TestRecordCashflowOut_Valid(t *testing.T) {
	repo := newMockCashflowOutRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	cf, err := svc.RecordCashflowOut("t-1", "u-1", "p-1", 1, 50.0, "Groceries", "receipt.jpg", "expense").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cf.Amount != 50.0 {
		t.Errorf("expected amount 50.0, got %f", cf.Amount)
	}
	if cf.Type != "expense" {
		t.Errorf("expected type 'expense', got '%s'", cf.Type)
	}
	if cf.TenantID != "t-1" {
		t.Errorf("expected tenant ID 't-1', got '%s'", cf.TenantID)
	}
	if cf.PocketID != "p-1" {
		t.Errorf("expected pocket ID 'p-1', got '%s'", cf.PocketID)
	}
	if cf.IsDeleted {
		t.Error("expected IsDeleted to be false")
	}
	if cf.ID == "" {
		t.Error("expected non-empty ID")
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventRecorded {
		t.Errorf("expected event type '%s', got '%s'", EventRecorded, (*recorded)[0].EventType)
	}
	if (*recorded)[0].AggregateType != "cashflowout" {
		t.Errorf("expected aggregate type 'cashflowout', got '%s'", (*recorded)[0].AggregateType)
	}
}

func TestRecordCashflowOut_ZeroAmount(t *testing.T) {
	repo := newMockCashflowOutRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.RecordCashflowOut("t-1", "u-1", "p-1", 1, 0, "Groceries", "", "expense").Get()
	if !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestRecordCashflowOut_NegativeAmount(t *testing.T) {
	repo := newMockCashflowOutRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.RecordCashflowOut("t-1", "u-1", "p-1", 1, -10.0, "Groceries", "", "expense").Get()
	if !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestRecordCashflowOut_EmptyType(t *testing.T) {
	repo := newMockCashflowOutRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.RecordCashflowOut("t-1", "u-1", "p-1", 1, 50.0, "Groceries", "", "").Get()
	if !errors.Is(err, ErrInvalidType) {
		t.Errorf("expected ErrInvalidType, got %v", err)
	}
}

func TestUpdateCashflowOut(t *testing.T) {
	repo := newMockCashflowOutRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := CashflowOut{
		ID:          "cf-1",
		TenantID:    "t-1",
		Amount:      50.0,
		Description: "Old",
		Type:        "expense",
		PocketID:    "p-1",
		CategoryID:  1,
		IsDeleted:   false,
	}
	repo.items["cf-1"] = existing

	cf, err := svc.UpdateCashflowOut("cf-1", 75.0, "Updated", 2, "new.jpg").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cf.Amount != 75.0 {
		t.Errorf("expected amount 75.0, got %f", cf.Amount)
	}
	if cf.Description != "Updated" {
		t.Errorf("expected description 'Updated', got '%s'", cf.Description)
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventUpdated {
		t.Errorf("expected event type '%s', got '%s'", EventUpdated, (*recorded)[0].EventType)
	}
}

func TestUpdateCashflowOut_InvalidAmount(t *testing.T) {
	repo := newMockCashflowOutRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := CashflowOut{ID: "cf-1", Amount: 50.0, IsDeleted: false}
	repo.items["cf-1"] = existing

	_, err := svc.UpdateCashflowOut("cf-1", -10.0, "Updated", 1, "").Get()
	if !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestUpdateCashflowOut_NotFound(t *testing.T) {
	repo := newMockCashflowOutRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.UpdateCashflowOut("nonexistent", 75.0, "Updated", 1, "").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteCashflowOut(t *testing.T) {
	repo := newMockCashflowOutRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := CashflowOut{ID: "cf-1", TenantID: "t-1", Amount: 50.0, IsDeleted: false}
	repo.items["cf-1"] = existing

	cf, err := svc.DeleteCashflowOut("cf-1").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !cf.IsDeleted {
		t.Error("expected IsDeleted to be true")
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventDeleted {
		t.Errorf("expected event type '%s', got '%s'", EventDeleted, (*recorded)[0].EventType)
	}
}

func TestDeleteCashflowOut_NotFound(t *testing.T) {
	repo := newMockCashflowOutRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.DeleteCashflowOut("nonexistent").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetCashflowOutByID_Found(t *testing.T) {
	repo := newMockCashflowOutRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := CashflowOut{ID: "cf-1", Amount: 50.0, IsDeleted: false}
	repo.items["cf-1"] = existing

	cf, err := svc.GetCashflowOutByID("cf-1").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cf.ID != "cf-1" {
		t.Errorf("expected ID 'cf-1', got '%s'", cf.ID)
	}
}

func TestGetCashflowOutByID_NotFound(t *testing.T) {
	repo := newMockCashflowOutRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.GetCashflowOutByID("nonexistent").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestListCashflowOutsByPocket(t *testing.T) {
	repo := newMockCashflowOutRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	repo.items["cf-1"] = CashflowOut{ID: "cf-1", TenantID: "t-1", PocketID: "p-1", Amount: 50.0, IsDeleted: false}
	repo.items["cf-2"] = CashflowOut{ID: "cf-2", TenantID: "t-1", PocketID: "p-1", Amount: 30.0, IsDeleted: false}
	repo.items["cf-3"] = CashflowOut{ID: "cf-3", TenantID: "t-1", PocketID: "p-2", Amount: 20.0, IsDeleted: false}
	repo.items["cf-4"] = CashflowOut{ID: "cf-4", TenantID: "t-1", PocketID: "p-1", Amount: 10.0, IsDeleted: true}

	items, err := svc.ListCashflowOutsByPocket("t-1", "p-1").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}
