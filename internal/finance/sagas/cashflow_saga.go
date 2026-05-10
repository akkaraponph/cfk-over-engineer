package sagas

import (
	"context"
	"cfk/internal/finance/pocket"
	"cfk/pkg/saga"
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
				Execute: func(ctx context.Context, payload map[string]interface{}) error {
					pocketID, _ := payload["pocket_id"].(string)
					amount, _ := payload["amount"].(float64)
					if pocketID == "" || amount == 0 {
						return pocket.ErrNotFound
					}
					result := deps.PocketService.ChangeBalance(pocketID, amount)
					if result.IsError() {
						return result.Error()
					}
					return nil
				},
				Compensate: func(ctx context.Context, payload map[string]interface{}) error {
					pocketID, _ := payload["pocket_id"].(string)
					amount, _ := payload["amount"].(float64)
					if pocketID == "" || amount == 0 {
						return nil
					}
					result := deps.PocketService.ChangeBalance(pocketID, -amount)
					if result.IsError() {
						return result.Error()
					}
					return nil
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
				Execute: func(ctx context.Context, payload map[string]interface{}) error {
					pocketID, _ := payload["pocket_id"].(string)
					amount, _ := payload["amount"].(float64)
					if pocketID == "" || amount == 0 {
						return pocket.ErrNotFound
					}
					result := deps.PocketService.ChangeBalance(pocketID, -amount)
					if result.IsError() {
						return result.Error()
					}
					return nil
				},
				Compensate: func(ctx context.Context, payload map[string]interface{}) error {
					pocketID, _ := payload["pocket_id"].(string)
					amount, _ := payload["amount"].(float64)
					if pocketID == "" || amount == 0 {
						return nil
					}
					result := deps.PocketService.ChangeBalance(pocketID, amount)
					if result.IsError() {
						return result.Error()
					}
					return nil
				},
			},
		},
	}
}
