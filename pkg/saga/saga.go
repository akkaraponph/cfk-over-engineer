package saga

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type InstanceState string

const (
	StatePending      InstanceState = "pending"
	StateExecuting    InstanceState = "executing"
	StateCompleted    InstanceState = "completed"
	StateCompensating InstanceState = "compensating"
	StateFailed       InstanceState = "failed"
)

type Step struct {
	Name       string
	Execute    func(ctx context.Context, payload map[string]interface{}) error
	Compensate func(ctx context.Context, payload map[string]interface{}) error
}

type Definition struct {
	Name  string
	Steps []Step
}

type Instance struct {
	ID          string                 `json:"id"`
	SagaName    string                 `json:"saga_name"`
	State       InstanceState          `json:"state"`
	CurrentStep int                    `json:"current_step"`
	Payload     map[string]interface{} `json:"payload"`
	Error       string                 `json:"error,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

func NewInstance(sagaName string, payload map[string]interface{}) *Instance {
	now := time.Now()
	return &Instance{
		ID:        uuid.New().String(),
		SagaName:  sagaName,
		State:     StatePending,
		Payload:   payload,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
