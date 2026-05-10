package sagas

import (
	"cfk/internal/finance/pocket"
	"cfk/pkg/saga"
	"context"

	"github.com/samber/mo"
)

type CashflowInSagaDeps struct {
	PocketService *pocket.Service
}

func NewCashflowInSaga(deps CashflowInSagaDeps) saga.Definition {
	return saga.Definition{
		Name: "cashflowin",
		Steps: []saga.Step{
			{
				Name: "credit_pocket",
				Execute: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					pocketID, _ := payload["pocket_id"].(string)
					amount, _ := payload["amount"].(float64)
					if pocketID == "" || amount == 0 {
						return mo.Err[struct{}](pocket.ErrNotFound)
					}
					r := deps.PocketService.ChangeBalance(pocketID, amount)
					if r.IsError() {
						return mo.Err[struct{}](r.Error())
					}
					return saga.OkStep()
				},
				Compensate: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					pocketID, _ := payload["pocket_id"].(string)
					amount, _ := payload["amount"].(float64)
					if pocketID == "" || amount == 0 {
						return saga.OkStep()
					}
					r := deps.PocketService.ChangeBalance(pocketID, -amount)
					if r.IsError() {
						return mo.Err[struct{}](r.Error())
					}
					return saga.OkStep()
				},
			},
		},
	}
}

type CashflowOutSagaDeps struct {
	PocketService *pocket.Service
}

func NewCashflowOutSaga(deps CashflowOutSagaDeps) saga.Definition {
	return saga.Definition{
		Name: "cashflowout",
		Steps: []saga.Step{
			{
				Name: "debit_pocket",
				Execute: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					pocketID, _ := payload["pocket_id"].(string)
					amount, _ := payload["amount"].(float64)
					if pocketID == "" || amount == 0 {
						return mo.Err[struct{}](pocket.ErrNotFound)
					}
					r := deps.PocketService.ChangeBalance(pocketID, -amount)
					if r.IsError() {
						return mo.Err[struct{}](r.Error())
					}
					return saga.OkStep()
				},
				Compensate: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					pocketID, _ := payload["pocket_id"].(string)
					amount, _ := payload["amount"].(float64)
					if pocketID == "" || amount == 0 {
						return saga.OkStep()
					}
					r := deps.PocketService.ChangeBalance(pocketID, amount)
					if r.IsError() {
						return mo.Err[struct{}](r.Error())
					}
					return saga.OkStep()
				},
			},
		},
	}
}
