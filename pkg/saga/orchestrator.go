package saga

import (
	"context"

	"github.com/samber/mo"
)

type Store interface {
	Save(instance *Instance) mo.Result[struct{}]
	FindByID(id string) mo.Option[*Instance]
	FindByState(state InstanceState) mo.Result[[]*Instance]
	Update(instance *Instance) mo.Result[struct{}]
}

type Orchestrator struct {
	store       Store
	definitions map[string]Definition
}

func NewOrchestrator(store Store) *Orchestrator {
	return &Orchestrator{
		store:       store,
		definitions: make(map[string]Definition),
	}
}

func (o *Orchestrator) Register(def Definition) {
	o.definitions[def.Name] = def
}

func (o *Orchestrator) Execute(ctx context.Context, sagaName string, payload map[string]interface{}) mo.Result[*Instance] {
	def, ok := o.definitions[sagaName]
	if !ok {
		return mo.Err[*Instance](ErrSagaNotFound)
	}

	instance := NewInstance(sagaName, payload)
	instance.State = StateExecuting
	if r := o.store.Save(instance); r.IsError() {
		return mo.Err[*Instance](r.Error())
	}

	for i, step := range def.Steps {
		instance.CurrentStep = i
		if r := o.store.Update(instance); r.IsError() {
			return mo.Err[*Instance](r.Error())
		}

		if r := step.Execute(ctx, instance.Payload); r.IsError() {
			instance.Error = r.Error().Error()
			instance.State = StateCompensating
			_ = o.store.Update(instance)
			o.compensate(ctx, def, instance)
			return mo.Err[*Instance](r.Error())
		}
	}

	instance.State = StateCompleted
	instance.CurrentStep = len(def.Steps)
	_ = o.store.Update(instance)
	return mo.Ok(instance)
}

func (o *Orchestrator) compensate(ctx context.Context, def Definition, instance *Instance) {
	for i := instance.CurrentStep - 1; i >= 0; i-- {
		step := def.Steps[i]
		if step.Compensate != nil {
			_ = step.Compensate(ctx, instance.Payload)
		}
		instance.CurrentStep = i
	}
	instance.State = StateFailed
	_ = o.store.Update(instance)
}

func (o *Orchestrator) Recover(ctx context.Context) mo.Result[struct{}] {
	executingR := o.store.FindByState(StateExecuting)
	if executingR.IsError() {
		return mo.Err[struct{}](executingR.Error())
	}
	executing, _ := executingR.Get()

	compensatingR := o.store.FindByState(StateCompensating)
	if compensatingR.IsError() {
		return mo.Err[struct{}](compensatingR.Error())
	}
	compensating, _ := compensatingR.Get()

	all := append(executing, compensating...)
	for _, instance := range all {
		def, ok := o.definitions[instance.SagaName]
		if !ok {
			continue
		}
		if instance.State == StateCompensating {
			o.compensate(ctx, def, instance)
			continue
		}
		go o.resumeExecution(ctx, def, instance)
	}
	return OkStep()
}

func (o *Orchestrator) resumeExecution(ctx context.Context, def Definition, instance *Instance) {
	for i := instance.CurrentStep; i < len(def.Steps); i++ {
		step := def.Steps[i]
		instance.CurrentStep = i
		if r := o.store.Update(instance); r.IsError() {
			return
		}
		if r := step.Execute(ctx, instance.Payload); r.IsError() {
			instance.Error = r.Error().Error()
			instance.State = StateCompensating
			_ = o.store.Update(instance)
			o.compensate(ctx, def, instance)
			return
		}
	}
	instance.State = StateCompleted
	instance.CurrentStep = len(def.Steps)
	_ = o.store.Update(instance)
}
