package debt

import (
	"cfk/pkg/event"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/samber/mo"
)

type mockDebtRepo struct {
	debts map[string]Debt
}

func newMockDebtRepo() *mockDebtRepo {
	return &mockDebtRepo{
		debts: make(map[string]Debt),
	}
}

func (r *mockDebtRepo) AppendEvent(eventType string, aggregateID string, payload any, metadata map[string]interface{}) error {
	return nil
}

func (r *mockDebtRepo) FindByID(id string) mo.Option[Debt] {
	if d, ok := r.debts[id]; ok {
		return mo.Some(d)
	}
	return mo.None[Debt]()
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

func TestRecordDebt_Valid(t *testing.T) {
	repo := newMockDebtRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	d, err := svc.RecordDebt("t-1", "credit_card", "Visa card", "u-1", 5000.0, 15.5, 200.0, 1).Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if d.Type != "credit_card" {
		t.Errorf("expected type 'credit_card', got '%s'", d.Type)
	}
	if d.Amount != 5000.0 {
		t.Errorf("expected amount 5000.0, got %f", d.Amount)
	}
	if d.Interest != 15.5 {
		t.Errorf("expected interest 15.5, got %f", d.Interest)
	}
	if d.MinimumPay != 200.0 {
		t.Errorf("expected minimum pay 200.0, got %f", d.MinimumPay)
	}
	if d.Priority != 1 {
		t.Errorf("expected priority 1, got %d", d.Priority)
	}
	if d.TenantID != "t-1" {
		t.Errorf("expected tenant ID 't-1', got '%s'", d.TenantID)
	}
	if d.BalanceSheetID != "" {
		t.Errorf("expected empty balance sheet ID, got '%s'", d.BalanceSheetID)
	}
	if d.ID == "" {
		t.Error("expected non-empty ID")
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventRecorded {
		t.Errorf("expected event type '%s', got '%s'", EventRecorded, (*recorded)[0].EventType)
	}
	if (*recorded)[0].AggregateType != "debt" {
		t.Errorf("expected aggregate type 'debt', got '%s'", (*recorded)[0].AggregateType)
	}
}

func TestRecordDebt_EmptyType(t *testing.T) {
	repo := newMockDebtRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.RecordDebt("t-1", "", "Some debt", "u-1", 1000.0, 5.0, 100.0, 1).Get()
	if !errors.Is(err, ErrInvalidType) {
		t.Errorf("expected ErrInvalidType, got %v", err)
	}
}

func TestRecordDebt_NegativeAmount(t *testing.T) {
	repo := newMockDebtRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.RecordDebt("t-1", "loan", "Bad debt", "u-1", -100.0, 5.0, 50.0, 1).Get()
	if !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestRecordDebt_ZeroAmount(t *testing.T) {
	repo := newMockDebtRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	d, err := svc.RecordDebt("t-1", "loan", "Zero debt", "u-1", 0, 0, 0, 1).Get()
	if err != nil {
		t.Fatalf("expected no error for zero amount, got %v", err)
	}
	if d.Amount != 0 {
		t.Errorf("expected amount 0, got %f", d.Amount)
	}
}

func TestChangeAmount(t *testing.T) {
	repo := newMockDebtRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := Debt{
		ID:             "d-1",
		TenantID:       "t-1",
		Type:           "credit_card",
		Amount:         5000.0,
		Interest:       15.5,
		MinimumPay:     200.0,
		Priority:       1,
		BalanceSheetID: "",
		UserID:         "u-1",
	}
	repo.debts["d-1"] = existing

	d, err := svc.ChangeAmount("d-1", 4000.0).Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if d.Amount != 4000.0 {
		t.Errorf("expected amount 4000.0, got %f", d.Amount)
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventAmountChanged {
		t.Errorf("expected event type '%s', got '%s'", EventAmountChanged, (*recorded)[0].EventType)
	}
}

func TestChangeAmount_NegativeAmount(t *testing.T) {
	repo := newMockDebtRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.ChangeAmount("d-1", -100.0).Get()
	if !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestChangeAmount_NotFound(t *testing.T) {
	repo := newMockDebtRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.ChangeAmount("nonexistent", 100.0).Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestAssignToBalanceSheet(t *testing.T) {
	repo := newMockDebtRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := Debt{
		ID:             "d-1",
		TenantID:       "t-1",
		Type:           "credit_card",
		Amount:         5000.0,
		BalanceSheetID: "",
	}
	repo.debts["d-1"] = existing

	d, err := svc.AssignToBalanceSheet("d-1", "bs-1").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if d.BalanceSheetID != "bs-1" {
		t.Errorf("expected balance sheet ID 'bs-1', got '%s'", d.BalanceSheetID)
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
	repo := newMockDebtRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.AssignToBalanceSheet("nonexistent", "bs-1").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestUnassignFromBalanceSheet(t *testing.T) {
	repo := newMockDebtRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := Debt{
		ID:             "d-1",
		TenantID:       "t-1",
		Type:           "credit_card",
		Amount:         5000.0,
		BalanceSheetID: "bs-1",
	}
	repo.debts["d-1"] = existing

	d, err := svc.UnassignFromBalanceSheet("d-1").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if d.BalanceSheetID != "" {
		t.Errorf("expected empty balance sheet ID, got '%s'", d.BalanceSheetID)
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
	repo := newMockDebtRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.UnassignFromBalanceSheet("nonexistent").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetDebtByID_Found(t *testing.T) {
	repo := newMockDebtRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := Debt{ID: "d-1", Type: "credit_card", Amount: 5000.0}
	repo.debts["d-1"] = existing

	d, err := svc.GetDebtByID("d-1").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if d.ID != "d-1" {
		t.Errorf("expected ID 'd-1', got '%s'", d.ID)
	}
}

func TestGetDebtByID_NotFound(t *testing.T) {
	repo := newMockDebtRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.GetDebtByID("nonexistent").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
