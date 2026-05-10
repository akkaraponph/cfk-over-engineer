package category

import (
	"cfk/pkg/event"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/samber/mo"
)

type mockCategoryRepo struct {
	categories map[string]Category
}

func newMockCategoryRepo() *mockCategoryRepo {
	return &mockCategoryRepo{
		categories: make(map[string]Category),
	}
}

func (r *mockCategoryRepo) AppendEvent(eventType string, aggregateID string, payload any, metadata map[string]interface{}) error {
	return nil
}

func (r *mockCategoryRepo) FindByID(id string) mo.Option[Category] {
	if c, ok := r.categories[id]; ok {
		return mo.Some(c)
	}
	return mo.None[Category]()
}

func (r *mockCategoryRepo) FindByTenant(tenantID string) mo.Result[[]Category] {
	var result []Category
	for _, c := range r.categories {
		if c.TenantID == tenantID && !c.IsDeleted {
			result = append(result, c)
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

func TestCreateCategory_Valid(t *testing.T) {
	repo := newMockCategoryRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	cat, err := svc.CreateCategory("t-1", "Food", "Food expenses", "expense", true, "u-1").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cat.Name != "Food" {
		t.Errorf("expected name 'Food', got '%s'", cat.Name)
	}
	if cat.TenantID != "t-1" {
		t.Errorf("expected tenant ID 't-1', got '%s'", cat.TenantID)
	}
	if cat.Description != "Food expenses" {
		t.Errorf("expected description 'Food expenses', got '%s'", cat.Description)
	}
	if cat.Type != "expense" {
		t.Errorf("expected type 'expense', got '%s'", cat.Type)
	}
	if !cat.IsCustom {
		t.Error("expected IsCustom to be true")
	}
	if cat.IsDeleted {
		t.Error("expected IsDeleted to be false")
	}
	if cat.ID == "" {
		t.Error("expected non-empty ID")
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventCreated {
		t.Errorf("expected event type '%s', got '%s'", EventCreated, (*recorded)[0].EventType)
	}
	if (*recorded)[0].AggregateType != "category" {
		t.Errorf("expected aggregate type 'category', got '%s'", (*recorded)[0].AggregateType)
	}
}

func TestCreateCategory_EmptyName(t *testing.T) {
	repo := newMockCategoryRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.CreateCategory("t-1", "", "desc", "expense", false, "u-1").Get()
	if !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
}

func TestUpdateCategory(t *testing.T) {
	repo := newMockCategoryRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := Category{
		ID:          "c-1",
		TenantID:    "t-1",
		Name:        "Old Name",
		Description: "Old desc",
		Type:        "expense",
		IsCustom:    true,
		UserID:      "u-1",
		IsDeleted:   false,
	}
	repo.categories["c-1"] = existing

	cat, err := svc.UpdateCategory("c-1", "New Name", "New desc").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cat.Name != "New Name" {
		t.Errorf("expected name 'New Name', got '%s'", cat.Name)
	}
	if cat.Description != "New desc" {
		t.Errorf("expected description 'New desc', got '%s'", cat.Description)
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventUpdated {
		t.Errorf("expected event type '%s', got '%s'", EventUpdated, (*recorded)[0].EventType)
	}
}

func TestUpdateCategory_EmptyName(t *testing.T) {
	repo := newMockCategoryRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.UpdateCategory("c-1", "", "desc").Get()
	if !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
}

func TestUpdateCategory_NotFound(t *testing.T) {
	repo := newMockCategoryRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.UpdateCategory("nonexistent", "New Name", "New desc").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteCategory(t *testing.T) {
	repo := newMockCategoryRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := Category{
		ID:        "c-1",
		TenantID:  "t-1",
		Name:      "Food",
		Type:      "expense",
		IsDeleted: false,
	}
	repo.categories["c-1"] = existing

	cat, err := svc.DeleteCategory("c-1").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !cat.IsDeleted {
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

func TestDeleteCategory_NotFound(t *testing.T) {
	repo := newMockCategoryRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.DeleteCategory("nonexistent").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetCategoryByID_Found(t *testing.T) {
	repo := newMockCategoryRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := Category{ID: "c-1", Name: "Food", IsDeleted: false}
	repo.categories["c-1"] = existing

	cat, err := svc.GetCategoryByID("c-1").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cat.ID != "c-1" {
		t.Errorf("expected ID 'c-1', got '%s'", cat.ID)
	}
}

func TestGetCategoryByID_NotFound(t *testing.T) {
	repo := newMockCategoryRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.GetCategoryByID("nonexistent").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestListCategoriesByTenant(t *testing.T) {
	repo := newMockCategoryRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	repo.categories["c-1"] = Category{ID: "c-1", TenantID: "t-1", Name: "Food", IsDeleted: false}
	repo.categories["c-2"] = Category{ID: "c-2", TenantID: "t-1", Name: "Transport", IsDeleted: false}
	repo.categories["c-3"] = Category{ID: "c-3", TenantID: "t-2", Name: "Other", IsDeleted: false}
	repo.categories["c-4"] = Category{ID: "c-4", TenantID: "t-1", Name: "Deleted", IsDeleted: true}

	cats, err := svc.ListCategoriesByTenant("t-1").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cats) != 2 {
		t.Errorf("expected 2 categories, got %d", len(cats))
	}
}
