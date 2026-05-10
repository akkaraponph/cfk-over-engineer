package pocket

import (
	"cfk/pkg/event"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/samber/mo"
)

type mockPocketRepo struct {
	pockets map[string]Pocket
}

func newMockPocketRepo() *mockPocketRepo {
	return &mockPocketRepo{
		pockets: make(map[string]Pocket),
	}
}

func (r *mockPocketRepo) AppendEvent(eventType string, aggregateID string, payload any, metadata map[string]interface{}) error {
	return nil
}

func (r *mockPocketRepo) FindByID(id string) mo.Option[Pocket] {
	if p, ok := r.pockets[id]; ok {
		return mo.Some(p)
	}
	return mo.None[Pocket]()
}

func (r *mockPocketRepo) FindByUser(tenantID, userID string) mo.Result[[]Pocket] {
	var result []Pocket
	for _, p := range r.pockets {
		if p.TenantID == tenantID && p.UserID == userID && !p.IsDeleted {
			result = append(result, p)
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

func TestCreatePocket_Valid(t *testing.T) {
	repo := newMockPocketRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	p, err := svc.CreatePocket("t-1", "Wallet", "u-1").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.Name != "Wallet" {
		t.Errorf("expected name 'Wallet', got '%s'", p.Name)
	}
	if p.TenantID != "t-1" {
		t.Errorf("expected tenant ID 't-1', got '%s'", p.TenantID)
	}
	if p.UserID != "u-1" {
		t.Errorf("expected user ID 'u-1', got '%s'", p.UserID)
	}
	if p.Balance != 0.0 {
		t.Errorf("expected balance 0.0, got %f", p.Balance)
	}
	if p.IsDeleted {
		t.Error("expected IsDeleted to be false")
	}
	if p.ID == "" {
		t.Error("expected non-empty ID")
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventCreated {
		t.Errorf("expected event type '%s', got '%s'", EventCreated, (*recorded)[0].EventType)
	}
}

func TestCreatePocket_EmptyName(t *testing.T) {
	repo := newMockPocketRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.CreatePocket("t-1", "", "u-1").Get()
	if !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
}

func TestChangeName(t *testing.T) {
	repo := newMockPocketRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := Pocket{ID: "p-1", TenantID: "t-1", Name: "Old Name", Balance: 100.0, UserID: "u-1", IsDeleted: false}
	repo.pockets["p-1"] = existing

	p, err := svc.ChangeName("p-1", "New Name").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.Name != "New Name" {
		t.Errorf("expected name 'New Name', got '%s'", p.Name)
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventNameChanged {
		t.Errorf("expected event type '%s', got '%s'", EventNameChanged, (*recorded)[0].EventType)
	}
}

func TestChangeName_EmptyName(t *testing.T) {
	repo := newMockPocketRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.ChangeName("p-1", "").Get()
	if !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
}

func TestChangeName_NotFound(t *testing.T) {
	repo := newMockPocketRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.ChangeName("nonexistent", "New Name").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestChangeBalance(t *testing.T) {
	repo := newMockPocketRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := Pocket{ID: "p-1", TenantID: "t-1", Name: "Wallet", Balance: 100.0, UserID: "u-1", IsDeleted: false}
	repo.pockets["p-1"] = existing

	p, err := svc.ChangeBalance("p-1", 50.0).Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.Balance != 150.0 {
		t.Errorf("expected balance 150.0, got %f", p.Balance)
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventBalanceChanged {
		t.Errorf("expected event type '%s', got '%s'", EventBalanceChanged, (*recorded)[0].EventType)
	}
}

func TestChangeBalance_Negative(t *testing.T) {
	repo := newMockPocketRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := Pocket{ID: "p-1", TenantID: "t-1", Name: "Wallet", Balance: 100.0, UserID: "u-1", IsDeleted: false}
	repo.pockets["p-1"] = existing

	p, err := svc.ChangeBalance("p-1", -30.0).Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.Balance != 70.0 {
		t.Errorf("expected balance 70.0, got %f", p.Balance)
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
}

func TestChangeBalance_NotFound(t *testing.T) {
	repo := newMockPocketRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.ChangeBalance("nonexistent", 50.0).Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeletePocket(t *testing.T) {
	repo := newMockPocketRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := Pocket{ID: "p-1", TenantID: "t-1", Name: "Wallet", Balance: 100.0, UserID: "u-1", IsDeleted: false}
	repo.pockets["p-1"] = existing

	p, err := svc.DeletePocket("p-1").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !p.IsDeleted {
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

func TestDeletePocket_NotFound(t *testing.T) {
	repo := newMockPocketRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.DeletePocket("nonexistent").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetPocketByID_Found(t *testing.T) {
	repo := newMockPocketRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := Pocket{ID: "p-1", TenantID: "t-1", Name: "Wallet", Balance: 100.0, UserID: "u-1"}
	repo.pockets["p-1"] = existing

	p, err := svc.GetPocketByID("p-1").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.ID != "p-1" {
		t.Errorf("expected ID 'p-1', got '%s'", p.ID)
	}
}

func TestGetPocketByID_NotFound(t *testing.T) {
	repo := newMockPocketRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.GetPocketByID("nonexistent").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestListPocketsByUser(t *testing.T) {
	repo := newMockPocketRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	repo.pockets["p-1"] = Pocket{ID: "p-1", TenantID: "t-1", Name: "Wallet 1", UserID: "u-1", IsDeleted: false}
	repo.pockets["p-2"] = Pocket{ID: "p-2", TenantID: "t-1", Name: "Wallet 2", UserID: "u-1", IsDeleted: false}
	repo.pockets["p-3"] = Pocket{ID: "p-3", TenantID: "t-1", Name: "Wallet 3", UserID: "u-2", IsDeleted: false}
	repo.pockets["p-4"] = Pocket{ID: "p-4", TenantID: "t-2", Name: "Other", UserID: "u-1", IsDeleted: false}
	repo.pockets["p-5"] = Pocket{ID: "p-5", TenantID: "t-1", Name: "Deleted", UserID: "u-1", IsDeleted: true}

	pockets, err := svc.ListPocketsByUser("t-1", "u-1").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(pockets) != 2 {
		t.Errorf("expected 2 pockets, got %d", len(pockets))
	}
}
