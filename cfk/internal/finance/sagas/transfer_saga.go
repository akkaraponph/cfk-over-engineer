package sagas

import (
	"cfk/internal/finance/pocket"
	"cfk/internal/finance/transfer"
	"cfk/pkg/saga"
	"context"

	"github.com/samber/mo"
)

type TransferSagaDeps struct {
	TransferService *transfer.Service
	PocketService   *pocket.Service
}

func NewTransferSaga(deps TransferSagaDeps) saga.Definition {
	return saga.Definition{
		Name: "transfer",
		Steps: []saga.Step{
			{
				Name: "debit_source",
				Execute: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					fromPocketID, _ := payload["from_pocket_id"].(string)
					amount, _ := payload["amount"].(float64)
					if fromPocketID == "" || amount == 0 {
						return mo.Err[struct{}](pocket.ErrNotFound)
					}
					r := deps.PocketService.ChangeBalance(fromPocketID, -amount)
					if r.IsError() {
						return mo.Err[struct{}](r.Error())
					}
					return saga.OkStep()
				},
				Compensate: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					fromPocketID, _ := payload["from_pocket_id"].(string)
					amount, _ := payload["amount"].(float64)
					if fromPocketID == "" || amount == 0 {
						return saga.OkStep()
					}
					r := deps.PocketService.ChangeBalance(fromPocketID, amount)
					if r.IsError() {
						return mo.Err[struct{}](r.Error())
					}
					return saga.OkStep()
				},
			},
			{
				Name: "credit_destination",
				Execute: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					toPocketID, _ := payload["to_pocket_id"].(string)
					amount, _ := payload["amount"].(float64)
					if toPocketID == "" || amount == 0 {
						return mo.Err[struct{}](pocket.ErrNotFound)
					}
					r := deps.PocketService.ChangeBalance(toPocketID, amount)
					if r.IsError() {
						return mo.Err[struct{}](r.Error())
					}
					return saga.OkStep()
				},
				Compensate: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					toPocketID, _ := payload["to_pocket_id"].(string)
					amount, _ := payload["amount"].(float64)
					if toPocketID == "" || amount == 0 {
						return saga.OkStep()
					}
					r := deps.PocketService.ChangeBalance(toPocketID, -amount)
					if r.IsError() {
						return mo.Err[struct{}](r.Error())
					}
					return saga.OkStep()
				},
			},
			{
				Name: "complete_transfer",
				Execute: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					transferID, _ := payload["transfer_id"].(string)
					if transferID == "" {
						return mo.Err[struct{}](transfer.ErrNotFound)
					}
					r := deps.TransferService.CompleteTransfer(transferID)
					if r.IsError() {
						return mo.Err[struct{}](r.Error())
					}
					return saga.OkStep()
				},
				Compensate: func(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}] {
					transferID, _ := payload["transfer_id"].(string)
					if transferID == "" {
						return saga.OkStep()
					}
					r := deps.TransferService.FailTransfer(transferID, "saga compensation")
					if r.IsError() {
						return mo.Err[struct{}](r.Error())
					}
					return saga.OkStep()
				},
			},
		},
	}
}
