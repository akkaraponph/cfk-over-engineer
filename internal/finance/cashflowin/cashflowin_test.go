package cashflowin

import (
	"cfk/pkg/event"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/samber/mo"
)

type mockCashflowInRepo struct {
	items map[string]CashflowIn
}

func newMockCashflowInRepo() *mockCashflowInRepo {
	return &mockCashflowInRepo{
		items: make(map[string]CashflowIn),
	}
}

func (r *mockCashflowInRepo) AppendEvent(eventType string, aggregateID string, payload map[string]interface{}, metadata map[string]interface{}) error {
	return nil
}

func (r *mockCashflowInRepo) FindByID(id string) mo.Option[CashflowIn] {
	if cf, ok := r.items[id]; ok {
		return mo.Some(cf)
	}
	return mo.None[CashflowIn]()
}

func (r *mockCashflowInRepo) FindByPocket(tenantID, pocketID string) mo.Result[[]CashflowIn] {
	var result []CashflowIn
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

func TestRecordCashflowIn_Valid(t *testing.T) {
	repo := newMockCashflowInRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	cf, err := svc.RecordCashflowIn("t-1", "u-1", "p-1", 1, 100.50, "Salary", "receipt.jpg").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cf.Amount != 100.50 {
		t.Errorf("expected amount 100.50, got %f", cf.Amount)
	}
	if cf.TenantID != "t-1" {
		t.Errorf("expected tenant ID 't-1', got '%s'", cf.TenantID)
	}
	if cf.PocketID != "p-1" {
		t.Errorf("expected pocket ID 'p-1', got '%s'", cf.PocketID)
	}
	if cf.Description != "Salary" {
		t.Errorf("expected description 'Salary', got '%s'", cf.Description)
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
	if (*recorded)[0].AggregateType != "cashflowin" {
		t.Errorf("expected aggregate type 'cashflowin', got '%s'", (*recorded)[0].AggregateType)
	}
}

func TestRecordCashflowIn_ZeroAmount(t *testing.T) {
	repo := newMockCashflowInRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.RecordCashflowIn("t-1", "u-1", "p-1", 1, 0, "Salary", "").Get()
	if !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestRecordCashflowIn_NegativeAmount(t *testing.T) {
	repo := newMockCashflowInRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.RecordCashflowIn("t-1", "u-1", "p-1", 1, -50.0, "Salary", "").Get()
	if !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestUpdateCashflowIn(t *testing.T) {
	repo := newMockCashflowInRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := CashflowIn{
		ID:          "cf-1",
		TenantID:    "t-1",
		UserID:      "u-1",
		PocketID:    "p-1",
		CategoryID:  1,
		Amount:      100.0,
		Description: "Old",
		Receipt:     "old.jpg",
		IsDeleted:   false,
	}
	repo.items["cf-1"] = existing

	cf, err := svc.UpdateCashflowIn("cf-1", 200.0, "Updated", 2, "new.jpg").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cf.Amount != 200.0 {
		t.Errorf("expected amount 200.0, got %f", cf.Amount)
	}
	if cf.Description != "Updated" {
		t.Errorf("expected description 'Updated', got '%s'", cf.Description)
	}
	if cf.CategoryID != 2 {
		t.Errorf("expected category ID 2, got %d", cf.CategoryID)
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventUpdated {
		t.Errorf("expected event type '%s', got '%s'", EventUpdated, (*recorded)[0].EventType)
	}
}

func TestUpdateCashflowIn_InvalidAmount(t *testing.T) {
	repo := newMockCashflowInRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := CashflowIn{ID: "cf-1", Amount: 100.0, IsDeleted: false}
	repo.items["cf-1"] = existing

	_, err := svc.UpdateCashflowIn("cf-1", 0, "Updated", 1, "").Get()
	if !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestUpdateCashflowIn_NotFound(t *testing.T) {
	repo := newMockCashflowInRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.UpdateCashflowIn("nonexistent", 200.0, "Updated", 1, "").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteCashflowIn(t *testing.T) {
	repo := newMockCashflowInRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := CashflowIn{ID: "cf-1", TenantID: "t-1", Amount: 100.0, IsDeleted: false}
	repo.items["cf-1"] = existing

	cf, err := svc.DeleteCashflowIn("cf-1").Get()
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

func TestDeleteCashflowIn_NotFound(t *testing.T) {
	repo := newMockCashflowInRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.DeleteCashflowIn("nonexistent").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetCashflowInByID_Found(t *testing.T) {
	repo := newMockCashflowInRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := CashflowIn{ID: "cf-1", Amount: 100.0, IsDeleted: false}
	repo.items["cf-1"] = existing

	cf, err := svc.GetCashflowInByID("cf-1").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cf.ID != "cf-1" {
		t.Errorf("expected ID 'cf-1', got '%s'", cf.ID)
	}
}

func TestGetCashflowInByID_NotFound(t *testing.T) {
	repo := newMockCashflowInRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.GetCashflowInByID("nonexistent").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestListCashflowInsByPocket(t *testing.T) {
	repo := newMockCashflowInRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	repo.items["cf-1"] = CashflowIn{ID: "cf-1", TenantID: "t-1", PocketID: "p-1", Amount: 100.0, IsDeleted: false}
	repo.items["cf-2"] = CashflowIn{ID: "cf-2", TenantID: "t-1", PocketID: "p-1", Amount: 200.0, IsDeleted: false}
	repo.items["cf-3"] = CashflowIn{ID: "cf-3", TenantID: "t-1", PocketID: "p-2", Amount: 300.0, IsDeleted: false}
	repo.items["cf-4"] = CashflowIn{ID: "cf-4", TenantID: "t-1", PocketID: "p-1", Amount: 400.0, IsDeleted: true}

	items, err := svc.ListCashflowInsByPocket("t-1", "p-1").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}
