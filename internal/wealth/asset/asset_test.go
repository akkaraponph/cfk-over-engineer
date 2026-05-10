package asset

import (
	"cfk/pkg/event"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/samber/mo"
)

type mockAssetRepo struct {
	assets map[string]Asset
}

func newMockAssetRepo() *mockAssetRepo {
	return &mockAssetRepo{
		assets: make(map[string]Asset),
	}
}

func (r *mockAssetRepo) AppendEvent(eventType string, aggregateID string, payload map[string]interface{}, metadata map[string]interface{}) error {
	return nil
}

func (r *mockAssetRepo) FindByID(id string) mo.Option[Asset] {
	if a, ok := r.assets[id]; ok {
		return mo.Some(a)
	}
	return mo.None[Asset]()
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

func TestRecordAsset_Valid(t *testing.T) {
	repo := newMockAssetRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	a, err := svc.RecordAsset("t-1", "stock", "Apple stock", "u-1", 10000.0, 500.0).Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if a.Type != "stock" {
		t.Errorf("expected type 'stock', got '%s'", a.Type)
	}
	if a.Value != 10000.0 {
		t.Errorf("expected value 10000.0, got %f", a.Value)
	}
	if a.CashflowPerYear != 500.0 {
		t.Errorf("expected cashflow 500.0, got %f", a.CashflowPerYear)
	}
	if a.TenantID != "t-1" {
		t.Errorf("expected tenant ID 't-1', got '%s'", a.TenantID)
	}
	if a.BalanceSheetID != "" {
		t.Errorf("expected empty balance sheet ID, got '%s'", a.BalanceSheetID)
	}
	if a.ID == "" {
		t.Error("expected non-empty ID")
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventRecorded {
		t.Errorf("expected event type '%s', got '%s'", EventRecorded, (*recorded)[0].EventType)
	}
	if (*recorded)[0].AggregateType != "asset" {
		t.Errorf("expected aggregate type 'asset', got '%s'", (*recorded)[0].AggregateType)
	}
}

func TestRecordAsset_EmptyType(t *testing.T) {
	repo := newMockAssetRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.RecordAsset("t-1", "", "Some asset", "u-1", 1000.0, 0).Get()
	if !errors.Is(err, ErrInvalidType) {
		t.Errorf("expected ErrInvalidType, got %v", err)
	}
}

func TestRecordAsset_NegativeValue(t *testing.T) {
	repo := newMockAssetRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.RecordAsset("t-1", "stock", "Bad asset", "u-1", -100.0, 0).Get()
	if !errors.Is(err, ErrInvalidValue) {
		t.Errorf("expected ErrInvalidValue, got %v", err)
	}
}

func TestRecordAsset_ZeroValue(t *testing.T) {
	repo := newMockAssetRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	a, err := svc.RecordAsset("t-1", "cash", "Cash", "u-1", 0, 0).Get()
	if err != nil {
		t.Fatalf("expected no error for zero value, got %v", err)
	}
	if a.Value != 0 {
		t.Errorf("expected value 0, got %f", a.Value)
	}
}

func TestChangeValue(t *testing.T) {
	repo := newMockAssetRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := Asset{
		ID:              "a-1",
		TenantID:        "t-1",
		Type:            "stock",
		Description:     "Apple stock",
		Value:           10000.0,
		CashflowPerYear: 500.0,
		BalanceSheetID:  "",
		UserID:          "u-1",
	}
	repo.assets["a-1"] = existing

	a, err := svc.ChangeValue("a-1", 12000.0, 600.0).Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if a.Value != 12000.0 {
		t.Errorf("expected value 12000.0, got %f", a.Value)
	}
	if a.CashflowPerYear != 600.0 {
		t.Errorf("expected cashflow 600.0, got %f", a.CashflowPerYear)
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventValueChanged {
		t.Errorf("expected event type '%s', got '%s'", EventValueChanged, (*recorded)[0].EventType)
	}
}

func TestChangeValue_NegativeValue(t *testing.T) {
	repo := newMockAssetRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.ChangeValue("a-1", -100.0, 0).Get()
	if !errors.Is(err, ErrInvalidValue) {
		t.Errorf("expected ErrInvalidValue, got %v", err)
	}
}

func TestChangeValue_NotFound(t *testing.T) {
	repo := newMockAssetRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.ChangeValue("nonexistent", 100.0, 0).Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestAssignToBalanceSheet(t *testing.T) {
	repo := newMockAssetRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := Asset{
		ID:             "a-1",
		TenantID:       "t-1",
		Type:           "stock",
		Value:          10000.0,
		BalanceSheetID: "",
	}
	repo.assets["a-1"] = existing

	a, err := svc.AssignToBalanceSheet("a-1", "bs-1").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if a.BalanceSheetID != "bs-1" {
		t.Errorf("expected balance sheet ID 'bs-1', got '%s'", a.BalanceSheetID)
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventAssignedToBalanceSheet {
		t.Errorf("expected event type '%s', got '%s'", EventAssignedToBalanceSheet, (*recorded)[0].EventType)
	}
}

func TestAssignToBalanceSheet_NotFound(t *testing.T) {
	repo := newMockAssetRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.AssignToBalanceSheet("nonexistent", "bs-1").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestUnassignFromBalanceSheet(t *testing.T) {
	repo := newMockAssetRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := Asset{
		ID:             "a-1",
		TenantID:       "t-1",
		Type:           "stock",
		Value:          10000.0,
		BalanceSheetID: "bs-1",
	}
	repo.assets["a-1"] = existing

	a, err := svc.UnassignFromBalanceSheet("a-1").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if a.BalanceSheetID != "" {
		t.Errorf("expected empty balance sheet ID, got '%s'", a.BalanceSheetID)
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventUnassignedFromBalanceSheet {
		t.Errorf("expected event type '%s', got '%s'", EventUnassignedFromBalanceSheet, (*recorded)[0].EventType)
	}
}

func TestUnassignFromBalanceSheet_NotFound(t *testing.T) {
	repo := newMockAssetRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.UnassignFromBalanceSheet("nonexistent").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetAssetByID_Found(t *testing.T) {
	repo := newMockAssetRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := Asset{ID: "a-1", Type: "stock", Value: 10000.0}
	repo.assets["a-1"] = existing

	a, err := svc.GetAssetByID("a-1").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if a.ID != "a-1" {
		t.Errorf("expected ID 'a-1', got '%s'", a.ID)
	}
}

func TestGetAssetByID_NotFound(t *testing.T) {
	repo := newMockAssetRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.GetAssetByID("nonexistent").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
