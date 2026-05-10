package transfer

import (
	"cfk/pkg/event"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/samber/mo"
)

type mockTransferRepo struct {
	transfers map[string]Transfer
}

func newMockTransferRepo() *mockTransferRepo {
	return &mockTransferRepo{
		transfers: make(map[string]Transfer),
	}
}

func (r *mockTransferRepo) AppendEvent(eventType string, aggregateID string, payload map[string]interface{}, metadata map[string]interface{}) error {
	return nil
}

func (r *mockTransferRepo) FindByID(id string) mo.Option[Transfer] {
	if tf, ok := r.transfers[id]; ok {
		return mo.Some(tf)
	}
	return mo.None[Transfer]()
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

func TestInitiateTransfer_Valid(t *testing.T) {
	repo := newMockTransferRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	tf, err := svc.InitiateTransfer("t-1", "u-1", "p-from", "p-to", 100.0).Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tf.Amount != 100.0 {
		t.Errorf("expected amount 100.0, got %f", tf.Amount)
	}
	if tf.FromPocketID != "p-from" {
		t.Errorf("expected from pocket 'p-from', got '%s'", tf.FromPocketID)
	}
	if tf.ToPocketID != "p-to" {
		t.Errorf("expected to pocket 'p-to', got '%s'", tf.ToPocketID)
	}
	if tf.Status != "pending" {
		t.Errorf("expected status 'pending', got '%s'", tf.Status)
	}
	if tf.IsDeleted {
		t.Error("expected IsDeleted to be false")
	}
	if tf.TenantID != "t-1" {
		t.Errorf("expected tenant ID 't-1', got '%s'", tf.TenantID)
	}
	if tf.ID == "" {
		t.Error("expected non-empty ID")
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventInitiated {
		t.Errorf("expected event type '%s', got '%s'", EventInitiated, (*recorded)[0].EventType)
	}
	if (*recorded)[0].AggregateType != "transfer" {
		t.Errorf("expected aggregate type 'transfer', got '%s'", (*recorded)[0].AggregateType)
	}
}

func TestInitiateTransfer_ZeroAmount(t *testing.T) {
	repo := newMockTransferRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.InitiateTransfer("t-1", "u-1", "p-from", "p-to", 0).Get()
	if !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestInitiateTransfer_NegativeAmount(t *testing.T) {
	repo := newMockTransferRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.InitiateTransfer("t-1", "u-1", "p-from", "p-to", -50.0).Get()
	if !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestInitiateTransfer_SamePocket(t *testing.T) {
	repo := newMockTransferRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.InitiateTransfer("t-1", "u-1", "p-same", "p-same", 100.0).Get()
	if !errors.Is(err, ErrSamePocket) {
		t.Errorf("expected ErrSamePocket, got %v", err)
	}
}

func TestCompleteTransfer(t *testing.T) {
	repo := newMockTransferRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := Transfer{
		ID:           "tf-1",
		TenantID:     "t-1",
		FromPocketID: "p-from",
		ToPocketID:   "p-to",
		Amount:       100.0,
		Status:       "pending",
		IsDeleted:    false,
	}
	repo.transfers["tf-1"] = existing

	tf, err := svc.CompleteTransfer("tf-1").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tf.Status != "completed" {
		t.Errorf("expected status 'completed', got '%s'", tf.Status)
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventCompleted {
		t.Errorf("expected event type '%s', got '%s'", EventCompleted, (*recorded)[0].EventType)
	}
}

func TestCompleteTransfer_NotPending(t *testing.T) {
	repo := newMockTransferRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := Transfer{
		ID:           "tf-1",
		TenantID:     "t-1",
		FromPocketID: "p-from",
		ToPocketID:   "p-to",
		Amount:       100.0,
		Status:       "completed",
		IsDeleted:    false,
	}
	repo.transfers["tf-1"] = existing

	_, err := svc.CompleteTransfer("tf-1").Get()
	if !errors.Is(err, ErrInvalidStatus) {
		t.Errorf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestCompleteTransfer_NotFound(t *testing.T) {
	repo := newMockTransferRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.CompleteTransfer("nonexistent").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestFailTransfer(t *testing.T) {
	repo := newMockTransferRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := Transfer{
		ID:           "tf-1",
		TenantID:     "t-1",
		FromPocketID: "p-from",
		ToPocketID:   "p-to",
		Amount:       100.0,
		Status:       "pending",
		IsDeleted:    false,
	}
	repo.transfers["tf-1"] = existing

	tf, err := svc.FailTransfer("tf-1", "insufficient funds").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tf.Status != "failed" {
		t.Errorf("expected status 'failed', got '%s'", tf.Status)
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventFailed {
		t.Errorf("expected event type '%s', got '%s'", EventFailed, (*recorded)[0].EventType)
	}
}

func TestFailTransfer_NotPending(t *testing.T) {
	repo := newMockTransferRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := Transfer{
		ID:           "tf-1",
		TenantID:     "t-1",
		FromPocketID: "p-from",
		ToPocketID:   "p-to",
		Amount:       100.0,
		Status:       "completed",
		IsDeleted:    false,
	}
	repo.transfers["tf-1"] = existing

	_, err := svc.FailTransfer("tf-1", "reason").Get()
	if !errors.Is(err, ErrInvalidStatus) {
		t.Errorf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestFailTransfer_NotFound(t *testing.T) {
	repo := newMockTransferRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.FailTransfer("nonexistent", "reason").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteTransfer(t *testing.T) {
	repo := newMockTransferRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := Transfer{
		ID:           "tf-1",
		TenantID:     "t-1",
		FromPocketID: "p-from",
		ToPocketID:   "p-to",
		Amount:       100.0,
		Status:       "completed",
		IsDeleted:    false,
	}
	repo.transfers["tf-1"] = existing

	tf, err := svc.DeleteTransfer("tf-1").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !tf.IsDeleted {
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

func TestDeleteTransfer_NotFound(t *testing.T) {
	repo := newMockTransferRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.DeleteTransfer("nonexistent").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetTransferByID_Found(t *testing.T) {
	repo := newMockTransferRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := Transfer{ID: "tf-1", Amount: 100.0, Status: "pending", IsDeleted: false}
	repo.transfers["tf-1"] = existing

	tf, err := svc.GetTransferByID("tf-1").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tf.ID != "tf-1" {
		t.Errorf("expected ID 'tf-1', got '%s'", tf.ID)
	}
}

func TestGetTransferByID_NotFound(t *testing.T) {
	repo := newMockTransferRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.GetTransferByID("nonexistent").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
