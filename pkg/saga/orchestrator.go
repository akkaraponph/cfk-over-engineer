package saga

import "context"

type Store interface {
	Save(instance *Instance) error
	FindByID(id string) (*Instance, error)
	FindByState(state InstanceState) ([]*Instance, error)
	Update(instance *Instance) error
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

func (o *Orchestrator) Execute(ctx context.Context, sagaName string, payload map[string]interface{}) (*Instance, error) {
	def, ok := o.definitions[sagaName]
	if !ok {
		return nil, ErrSagaNotFound
	}

	instance := NewInstance(sagaName, payload)
	instance.State = StateExecuting
	if err := o.store.Save(instance); err != nil {
		return nil, err
	}

	for i, step := range def.Steps {
		instance.CurrentStep = i
		if err := o.store.Update(instance); err != nil {
			return instance, err
		}

		if err := step.Execute(ctx, instance.Payload); err != nil {
			instance.Error = err.Error()
			instance.State = StateCompensating
			if updateErr := o.store.Update(instance); updateErr != nil {
				return instance, updateErr
			}
			o.compensate(ctx, def, instance)
			return instance, err
		}
	}

	instance.State = StateCompleted
	instance.CurrentStep = len(def.Steps)
	_ = o.store.Update(instance)
	return instance, nil
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

func (o *Orchestrator) Recover(ctx context.Context) error {
	executing, err := o.store.FindByState(StateExecuting)
	if err != nil {
		return err
	}
	compensating, err := o.store.FindByState(StateCompensating)
	if err != nil {
		return err
	}

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
	return nil
}

func (o *Orchestrator) resumeExecution(ctx context.Context, def Definition, instance *Instance) {
	for i := instance.CurrentStep; i < len(def.Steps); i++ {
		step := def.Steps[i]
		instance.CurrentStep = i
		if err := o.store.Update(instance); err != nil {
			return
		}
		if err := step.Execute(ctx, instance.Payload); err != nil {
			instance.Error = err.Error()
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
