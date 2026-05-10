package tenant

import (
	"cfk/pkg/event"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/samber/mo"
)

type mockTenantRepo struct {
	tenants  map[string]Tenant
	bySlug   map[string]Tenant
	features map[string]map[string]bool
}

func newMockTenantRepo() *mockTenantRepo {
	return &mockTenantRepo{
		tenants:  make(map[string]Tenant),
		bySlug:   make(map[string]Tenant),
		features: make(map[string]map[string]bool),
	}
}

func (r *mockTenantRepo) FindByID(id string) mo.Option[Tenant] {
	if t, ok := r.tenants[id]; ok {
		return mo.Some(t)
	}
	return mo.None[Tenant]()
}

func (r *mockTenantRepo) FindBySlug(slug string) mo.Option[Tenant] {
	if t, ok := r.bySlug[slug]; ok {
		return mo.Some(t)
	}
	return mo.None[Tenant]()
}

func (r *mockTenantRepo) HasFeature(tenantID, feature string) bool {
	if features, ok := r.features[tenantID]; ok {
		return features[feature]
	}
	return false
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

func TestCreateTenant_Valid(t *testing.T) {
	repo := newMockTenantRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	result := svc.CreateTenant("Test Tenant", "test-slug", "free")
	tn, err := result.Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tn.Name != "Test Tenant" {
		t.Errorf("expected name 'Test Tenant', got '%s'", tn.Name)
	}
	if tn.Slug != "test-slug" {
		t.Errorf("expected slug 'test-slug', got '%s'", tn.Slug)
	}
	if tn.Plan != PlanFree {
		t.Errorf("expected plan '%s', got '%s'", PlanFree, tn.Plan)
	}
	if !tn.IsActive {
		t.Error("expected IsActive to be true")
	}
	if tn.ID == "" {
		t.Error("expected non-empty ID")
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventCreated {
		t.Errorf("expected event type '%s', got '%s'", EventCreated, (*recorded)[0].EventType)
	}
	if (*recorded)[0].AggregateType != "tenant" {
		t.Errorf("expected aggregate type 'tenant', got '%s'", (*recorded)[0].AggregateType)
	}
}

func TestCreateTenant_AllPlans(t *testing.T) {
	repo := newMockTenantRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	plans := []Plan{PlanFree, PlanPremium, PlanEnterprise}
	for _, plan := range plans {
		tn, err := svc.CreateTenant("Test", "slug-"+string(plan), string(plan)).Get()
		if err != nil {
			t.Errorf("plan '%s': expected no error, got %v", plan, err)
		}
		if tn.Plan != plan {
			t.Errorf("plan '%s': expected plan '%s', got '%s'", plan, plan, tn.Plan)
		}
	}
}

func TestCreateTenant_EmptyName(t *testing.T) {
	repo := newMockTenantRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.CreateTenant("", "test-slug", "free").Get()
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
}

func TestCreateTenant_EmptySlug(t *testing.T) {
	repo := newMockTenantRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.CreateTenant("Test", "", "free").Get()
	if err == nil {
		t.Fatal("expected error for empty slug")
	}
	if !errors.Is(err, ErrInvalidSlug) {
		t.Errorf("expected ErrInvalidSlug, got %v", err)
	}
}

func TestCreateTenant_InvalidPlan(t *testing.T) {
	repo := newMockTenantRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.CreateTenant("Test", "test-slug", "invalid").Get()
	if err == nil {
		t.Fatal("expected error for invalid plan")
	}
	if !errors.Is(err, ErrInvalidPlan) {
		t.Errorf("expected ErrInvalidPlan, got %v", err)
	}
}

func TestActivateTenant(t *testing.T) {
	repo := newMockTenantRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := Tenant{ID: "t-1", Name: "Test", Slug: "test", Plan: PlanFree, IsActive: false}
	repo.tenants["t-1"] = existing

	tn, err := svc.ActivateTenant("t-1").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !tn.IsActive {
		t.Error("expected IsActive to be true after activation")
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventActivated {
		t.Errorf("expected event type '%s', got '%s'", EventActivated, (*recorded)[0].EventType)
	}
}

func TestActivateTenant_NotFound(t *testing.T) {
	repo := newMockTenantRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.ActivateTenant("nonexistent").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeactivateTenant(t *testing.T) {
	repo := newMockTenantRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := Tenant{ID: "t-1", Name: "Test", Slug: "test", Plan: PlanFree, IsActive: true}
	repo.tenants["t-1"] = existing

	tn, err := svc.DeactivateTenant("t-1").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tn.IsActive {
		t.Error("expected IsActive to be false after deactivation")
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventDeactivated {
		t.Errorf("expected event type '%s', got '%s'", EventDeactivated, (*recorded)[0].EventType)
	}
}

func TestDeactivateTenant_NotFound(t *testing.T) {
	repo := newMockTenantRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.DeactivateTenant("nonexistent").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestChangePlan(t *testing.T) {
	repo := newMockTenantRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := Tenant{ID: "t-1", Name: "Test", Slug: "test", Plan: PlanFree, IsActive: true}
	repo.tenants["t-1"] = existing

	tn, err := svc.ChangePlan("t-1", "premium").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tn.Plan != PlanPremium {
		t.Errorf("expected plan '%s', got '%s'", PlanPremium, tn.Plan)
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventPlanChanged {
		t.Errorf("expected event type '%s', got '%s'", EventPlanChanged, (*recorded)[0].EventType)
	}
}

func TestChangePlan_InvalidPlan(t *testing.T) {
	repo := newMockTenantRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := Tenant{ID: "t-1", Name: "Test", Slug: "test", Plan: PlanFree, IsActive: true}
	repo.tenants["t-1"] = existing

	_, err := svc.ChangePlan("t-1", "invalid").Get()
	if !errors.Is(err, ErrInvalidPlan) {
		t.Errorf("expected ErrInvalidPlan, got %v", err)
	}
}

func TestChangePlan_NotFound(t *testing.T) {
	repo := newMockTenantRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.ChangePlan("nonexistent", "premium").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetTenantBySlug_Found(t *testing.T) {
	repo := newMockTenantRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := Tenant{ID: "t-1", Name: "Test", Slug: "test-slug", Plan: PlanFree, IsActive: true}
	repo.tenants["t-1"] = existing
	repo.bySlug["test-slug"] = existing

	tn, err := svc.GetTenantBySlug("test-slug").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tn.Slug != "test-slug" {
		t.Errorf("expected slug 'test-slug', got '%s'", tn.Slug)
	}
}

func TestGetTenantBySlug_NotFound(t *testing.T) {
	repo := newMockTenantRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.GetTenantBySlug("nonexistent").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestEnableFeature(t *testing.T) {
	repo := newMockTenantRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := Tenant{ID: "t-1", Name: "Test", Slug: "test", Plan: PlanPremium, IsActive: true}
	repo.tenants["t-1"] = existing

	tf, err := svc.EnableFeature("t-1", "balance_sheet", "u-1").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tf.TenantID != "t-1" {
		t.Errorf("expected tenant ID 't-1', got '%s'", tf.TenantID)
	}
	if tf.Feature != FeatureBalanceSheet {
		t.Errorf("expected feature '%s', got '%s'", FeatureBalanceSheet, tf.Feature)
	}
	if !tf.IsEnabled {
		t.Error("expected IsEnabled to be true")
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventFeatureEnabled {
		t.Errorf("expected event type '%s', got '%s'", EventFeatureEnabled, (*recorded)[0].EventType)
	}
}

func TestEnableFeature_NotFound(t *testing.T) {
	repo := newMockTenantRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.EnableFeature("nonexistent", "balance_sheet", "u-1").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDisableFeature(t *testing.T) {
	repo := newMockTenantRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := Tenant{ID: "t-1", Name: "Test", Slug: "test", Plan: PlanPremium, IsActive: true}
	repo.tenants["t-1"] = existing

	tf, err := svc.DisableFeature("t-1", "balance_sheet", "u-1").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tf.IsEnabled {
		t.Error("expected IsEnabled to be false")
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventFeatureDisabled {
		t.Errorf("expected event type '%s', got '%s'", EventFeatureDisabled, (*recorded)[0].EventType)
	}
}

func TestHasFeature(t *testing.T) {
	repo := newMockTenantRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	repo.features["t-1"] = map[string]bool{"balance_sheet": true}

	if !svc.HasFeature("t-1", "balance_sheet") {
		t.Error("expected HasFeature to return true")
	}
	if svc.HasFeature("t-1", "debt") {
		t.Error("expected HasFeature to return false for missing feature")
	}
	if svc.HasFeature("nonexistent", "balance_sheet") {
		t.Error("expected HasFeature to return false for missing tenant")
	}
}
