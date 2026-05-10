package balancesheet

import (
	"cfk/pkg/event"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/samber/mo"
)

type mockBalanceSheetRepo struct {
	sheets map[string]BalanceSheet
}

func newMockBalanceSheetRepo() *mockBalanceSheetRepo {
	return &mockBalanceSheetRepo{
		sheets: make(map[string]BalanceSheet),
	}
}

func (r *mockBalanceSheetRepo) AppendEvent(eventType string, aggregateID string, payload any, metadata map[string]interface{}) error {
	return nil
}

func (r *mockBalanceSheetRepo) FindByID(id string) mo.Option[BalanceSheet] {
	if bs, ok := r.sheets[id]; ok {
		return mo.Some(bs)
	}
	return mo.None[BalanceSheet]()
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

func TestCreateBalanceSheet_Valid(t *testing.T) {
	repo := newMockBalanceSheetRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	bs, err := svc.CreateBalanceSheet("t-1", "u-1", 2024).Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if bs.Year != 2024 {
		t.Errorf("expected year 2024, got %d", bs.Year)
	}
	if bs.TenantID != "t-1" {
		t.Errorf("expected tenant ID 't-1', got '%s'", bs.TenantID)
	}
	if bs.UserID != "u-1" {
		t.Errorf("expected user ID 'u-1', got '%s'", bs.UserID)
	}
	if bs.ID == "" {
		t.Error("expected non-empty ID")
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventCreated {
		t.Errorf("expected event type '%s', got '%s'", EventCreated, (*recorded)[0].EventType)
	}
	if (*recorded)[0].AggregateType != "balancesheet" {
		t.Errorf("expected aggregate type 'balancesheet', got '%s'", (*recorded)[0].AggregateType)
	}
}

func TestCreateBalanceSheet_InvalidYear(t *testing.T) {
	repo := newMockBalanceSheetRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	tests := []struct {
		name string
		year int
	}{
		{"too low", 1999},
		{"too high", 2101},
		{"zero", 0},
		{"negative", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateBalanceSheet("t-1", "u-1", tt.year).Get()
			if !errors.Is(err, ErrInvalidYear) {
				t.Errorf("year %d: expected ErrInvalidYear, got %v", tt.year, err)
			}
		})
	}
}

func TestCreateBalanceSheet_BoundaryYears(t *testing.T) {
	repo := newMockBalanceSheetRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	tests := []struct {
		name string
		year int
	}{
		{"min valid", 2000},
		{"max valid", 2100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs, err := svc.CreateBalanceSheet("t-1", "u-1", tt.year).Get()
			if err != nil {
				t.Errorf("year %d: expected no error, got %v", tt.year, err)
			}
			if bs.Year != tt.year {
				t.Errorf("expected year %d, got %d", tt.year, bs.Year)
			}
		})
	}
}

func TestUpdateBalanceSheet(t *testing.T) {
	repo := newMockBalanceSheetRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := BalanceSheet{
		ID:       "bs-1",
		TenantID: "t-1",
		UserID:   "u-1",
		Year:     2023,
	}
	repo.sheets["bs-1"] = existing

	bs, err := svc.UpdateBalanceSheet("bs-1", 2024).Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if bs.Year != 2024 {
		t.Errorf("expected year 2024, got %d", bs.Year)
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventUpdated {
		t.Errorf("expected event type '%s', got '%s'", EventUpdated, (*recorded)[0].EventType)
	}
}

func TestUpdateBalanceSheet_InvalidYear(t *testing.T) {
	repo := newMockBalanceSheetRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := BalanceSheet{ID: "bs-1", TenantID: "t-1", Year: 2023}
	repo.sheets["bs-1"] = existing

	_, err := svc.UpdateBalanceSheet("bs-1", 1999).Get()
	if !errors.Is(err, ErrInvalidYear) {
		t.Errorf("expected ErrInvalidYear, got %v", err)
	}
}

func TestUpdateBalanceSheet_NotFound(t *testing.T) {
	repo := newMockBalanceSheetRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.UpdateBalanceSheet("nonexistent", 2024).Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetBalanceSheetByID_Found(t *testing.T) {
	repo := newMockBalanceSheetRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := BalanceSheet{ID: "bs-1", TenantID: "t-1", Year: 2024}
	repo.sheets["bs-1"] = existing

	bs, err := svc.GetBalanceSheetByID("bs-1").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if bs.ID != "bs-1" {
		t.Errorf("expected ID 'bs-1', got '%s'", bs.ID)
	}
}

func TestGetBalanceSheetByID_NotFound(t *testing.T) {
	repo := newMockBalanceSheetRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.GetBalanceSheetByID("nonexistent").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
