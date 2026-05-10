package saga

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/samber/mo"
)

type mockSagaStore struct {
	instances map[string]*Instance
}

func newMockSagaStore() *mockSagaStore {
	return &mockSagaStore{
		instances: make(map[string]*Instance),
	}
}

func (s *mockSagaStore) Save(instance *Instance) mo.Result[struct{}] {
	s.instances[instance.ID] = instance
	return OkStep()
}

func (s *mockSagaStore) FindByID(id string) mo.Option[*Instance] {
	if inst, ok := s.instances[id]; ok {
		return mo.Some(inst)
	}
	return mo.None[*Instance]()
}

func (s *mockSagaStore) FindByState(state InstanceState) mo.Result[[]*Instance] {
	var result []*Instance
	for _, inst := range s.instances {
		if inst.State == state {
			result = append(result, inst)
		}
	}
	return mo.Ok(result)
}

func (s *mockSagaStore) Update(instance *Instance) mo.Result[struct{}] {
	s.instances[instance.ID] = instance
	return OkStep()
}

func TestExecute_AllStepsSucceed(t *testing.T) {
	store := newMockSagaStore()
	orchestrator := NewOrchestrator(store)

	executed := []string{}
	def := Definition{
		Name: "test-saga",
		Steps: []Step{
			{
				Name: "step-1",
				Execute: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					executed = append(executed, "step-1")
					return OkStep()
				},
				Compensate: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					return OkStep()
				},
			},
			{
				Name: "step-2",
				Execute: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					executed = append(executed, "step-2")
					return OkStep()
				},
				Compensate: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					return OkStep()
				},
			},
		},
	}
	orchestrator.Register(def)

	instance, err := orchestrator.Execute(context.Background(), "test-saga", map[string]interface{}{"key": "value"}).Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if instance.State != StateCompleted {
		t.Errorf("expected state '%s', got '%s'", StateCompleted, instance.State)
	}
	if instance.CurrentStep != 2 {
		t.Errorf("expected current step 2, got %d", instance.CurrentStep)
	}
	if len(executed) != 2 {
		t.Fatalf("expected 2 steps executed, got %d", len(executed))
	}
	if executed[0] != "step-1" || executed[1] != "step-2" {
		t.Errorf("expected execution order [step-1, step-2], got %v", executed)
	}
}

func TestExecute_StepFails_Compensates(t *testing.T) {
	store := newMockSagaStore()
	orchestrator := NewOrchestrator(store)

	executed := []string{}
	compensated := []string{}
	stepErr := errors.New("step 2 failed")

	def := Definition{
		Name: "failing-saga",
		Steps: []Step{
			{
				Name: "step-1",
				Execute: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					executed = append(executed, "step-1")
					return OkStep()
				},
				Compensate: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					compensated = append(compensated, "step-1")
					return OkStep()
				},
			},
			{
				Name: "step-2",
				Execute: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					executed = append(executed, "step-2")
					return mo.Err[struct{}](stepErr)
				},
				Compensate: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					compensated = append(compensated, "step-2")
					return OkStep()
				},
			},
			{
				Name: "step-3",
				Execute: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					executed = append(executed, "step-3")
					return OkStep()
				},
				Compensate: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					compensated = append(compensated, "step-3")
					return OkStep()
				},
			},
		},
	}
	orchestrator.Register(def)

	instance, err := orchestrator.Execute(context.Background(), "failing-saga", map[string]interface{}{}).Get()
	if err == nil {
		t.Fatal("expected error when step 2 fails")
	}
	if !errors.Is(err, stepErr) {
		t.Errorf("expected step error, got %v", err)
	}
	if instance != nil {
		if instance.State != StateFailed {
			t.Errorf("expected state '%s', got '%s'", StateFailed, instance.State)
		}
	}
	if len(executed) != 2 {
		t.Errorf("expected 2 steps executed, got %d", len(executed))
	}
	if len(compensated) != 1 {
		t.Errorf("expected 1 compensation (step-1), got %d: %v", len(compensated), compensated)
	}
	if len(compensated) > 0 && compensated[0] != "step-1" {
		t.Errorf("expected step-1 to be compensated, got %v", compensated)
	}
}

func TestExecute_StepFails_MultipleCompensations(t *testing.T) {
	store := newMockSagaStore()
	orchestrator := NewOrchestrator(store)

	compensated := []string{}
	stepErr := errors.New("step 3 failed")

	def := Definition{
		Name: "multi-compensate",
		Steps: []Step{
			{
				Name: "step-1",
				Execute: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					return OkStep()
				},
				Compensate: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					compensated = append(compensated, "step-1")
					return OkStep()
				},
			},
			{
				Name: "step-2",
				Execute: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					return OkStep()
				},
				Compensate: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					compensated = append(compensated, "step-2")
					return OkStep()
				},
			},
			{
				Name: "step-3",
				Execute: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					return mo.Err[struct{}](stepErr)
				},
				Compensate: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					compensated = append(compensated, "step-3")
					return OkStep()
				},
			},
		},
	}
	orchestrator.Register(def)

	instance, err := orchestrator.Execute(context.Background(), "multi-compensate", map[string]interface{}{}).Get()
	if err == nil {
		t.Fatal("expected error")
	}
	if instance != nil && instance.State != StateFailed {
		t.Errorf("expected state '%s', got '%s'", StateFailed, instance.State)
	}
	if len(compensated) != 2 {
		t.Errorf("expected 2 compensations, got %d: %v", len(compensated), compensated)
	}
	if len(compensated) >= 2 {
		if compensated[0] != "step-2" {
			t.Errorf("expected first compensation to be step-2, got %s", compensated[0])
		}
		if compensated[1] != "step-1" {
			t.Errorf("expected second compensation to be step-1, got %s", compensated[1])
		}
	}
}

func TestExecute_UnknownSaga(t *testing.T) {
	store := newMockSagaStore()
	orchestrator := NewOrchestrator(store)

	_, err := orchestrator.Execute(context.Background(), "nonexistent", map[string]interface{}{}).Get()
	if !errors.Is(err, ErrSagaNotFound) {
		t.Errorf("expected ErrSagaNotFound, got %v", err)
	}
}

func TestExecute_PayloadPreserved(t *testing.T) {
	store := newMockSagaStore()
	orchestrator := NewOrchestrator(store)

	var receivedPayload map[string]interface{}
	def := Definition{
		Name: "payload-test",
		Steps: []Step{
			{
				Name: "step-1",
				Execute: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					receivedPayload = payload
					return OkStep()
				},
			},
		},
	}
	orchestrator.Register(def)

	payload := map[string]interface{}{"transfer_id": "t-1", "amount": 100.0}
	_, err := orchestrator.Execute(context.Background(), "payload-test", payload).Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if receivedPayload["transfer_id"] != "t-1" {
		t.Errorf("expected payload transfer_id 't-1', got %v", receivedPayload["transfer_id"])
	}
}

func TestExecute_StepWithoutCompensation(t *testing.T) {
	store := newMockSagaStore()
	orchestrator := NewOrchestrator(store)

	compensated := []string{}
	stepErr := errors.New("step 2 failed")

	def := Definition{
		Name: "no-compensate",
		Steps: []Step{
			{
				Name: "step-1",
				Execute: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					return OkStep()
				},
				Compensate: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					compensated = append(compensated, "step-1")
					return OkStep()
				},
			},
			{
				Name: "step-2",
				Execute: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					return mo.Err[struct{}](stepErr)
				},
			},
		},
	}
	orchestrator.Register(def)

	_, err := orchestrator.Execute(context.Background(), "no-compensate", map[string]interface{}{}).Get()
	if err == nil {
		t.Fatal("expected error")
	}
	if len(compensated) != 1 {
		t.Errorf("expected 1 compensation, got %d", len(compensated))
	}
}

func TestRecover_ExecutingInstances(t *testing.T) {
	store := newMockSagaStore()

	executed := []string{}
	def := Definition{
		Name: "recover-saga",
		Steps: []Step{
			{
				Name: "step-1",
				Execute: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					executed = append(executed, "step-1")
					return OkStep()
				},
			},
			{
				Name: "step-2",
				Execute: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					executed = append(executed, "step-2")
					return OkStep()
				},
			},
		},
	}

	orchestrator := NewOrchestrator(store)
	orchestrator.Register(def)

	inst := NewInstance("recover-saga", map[string]interface{}{})
	inst.State = StateExecuting
	inst.CurrentStep = 1
	store.instances[inst.ID] = inst

	result := orchestrator.Recover(context.Background())
	if result.IsError() {
		t.Fatalf("expected no error, got %v", result.Error())
	}

	time.Sleep(50 * time.Millisecond)

	updated := store.instances[inst.ID]
	if updated.State != StateCompleted {
		t.Errorf("expected state '%s', got '%s'", StateCompleted, updated.State)
	}
}

func TestRecover_CompensatingInstances(t *testing.T) {
	store := newMockSagaStore()

	compensated := false
	def := Definition{
		Name: "compensate-recover",
		Steps: []Step{
			{
				Name: "step-1",
				Execute: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					return OkStep()
				},
				Compensate: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					compensated = true
					return OkStep()
				},
			},
		},
	}

	orchestrator := NewOrchestrator(store)
	orchestrator.Register(def)

	inst := NewInstance("compensate-recover", map[string]interface{}{})
	inst.State = StateCompensating
	inst.CurrentStep = 1
	store.instances[inst.ID] = inst

	result := orchestrator.Recover(context.Background())
	if result.IsError() {
		t.Fatalf("expected no error, got %v", result.Error())
	}

	time.Sleep(50 * time.Millisecond)

	updated := store.instances[inst.ID]
	if updated.State != StateFailed {
		t.Errorf("expected state '%s', got '%s'", StateFailed, updated.State)
	}
	if !compensated {
		t.Error("expected step-1 to be compensated")
	}
}

func TestRecover_NoInstances(t *testing.T) {
	store := newMockSagaStore()
	orchestrator := NewOrchestrator(store)

	result := orchestrator.Recover(context.Background())
	if result.IsError() {
		t.Fatalf("expected no error, got %v", result.Error())
	}
}

func TestNewInstance(t *testing.T) {
	inst := NewInstance("test-saga", map[string]interface{}{"key": "value"})
	if inst.ID == "" {
		t.Error("expected non-empty ID")
	}
	if inst.SagaName != "test-saga" {
		t.Errorf("expected saga name 'test-saga', got '%s'", inst.SagaName)
	}
	if inst.State != StatePending {
		t.Errorf("expected state '%s', got '%s'", StatePending, inst.State)
	}
	if inst.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if inst.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
}

func TestInstanceState_String(t *testing.T) {
	states := map[InstanceState]string{
		StatePending:      "pending",
		StateExecuting:    "executing",
		StateCompleted:    "completed",
		StateCompensating: "compensating",
		StateFailed:       "failed",
	}
	for state, expected := range states {
		if string(state) != expected {
			t.Errorf("expected state string '%s', got '%s'", expected, string(state))
		}
	}
}
