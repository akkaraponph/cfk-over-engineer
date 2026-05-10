package sagas

import (
	"context"
	"cfk/internal/finance/pocket"
	"cfk/internal/finance/transfer"
	"cfk/pkg/saga"
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
				Execute: func(ctx context.Context, payload map[string]interface{}) error {
					fromPocketID, _ := payload["from_pocket_id"].(string)
					amount, _ := payload["amount"].(float64)
					if fromPocketID == "" || amount == 0 {
						return pocket.ErrNotFound
					}
					result := deps.PocketService.ChangeBalance(fromPocketID, -amount)
					if result.IsError() {
						return result.Error()
					}
					return nil
				},
				Compensate: func(ctx context.Context, payload map[string]interface{}) error {
					fromPocketID, _ := payload["from_pocket_id"].(string)
					amount, _ := payload["amount"].(float64)
					if fromPocketID == "" || amount == 0 {
						return nil
					}
					result := deps.PocketService.ChangeBalance(fromPocketID, amount)
					if result.IsError() {
						return result.Error()
					}
					return nil
				},
			},
			{
				Name: "credit_destination",
				Execute: func(ctx context.Context, payload map[string]interface{}) error {
					toPocketID, _ := payload["to_pocket_id"].(string)
					amount, _ := payload["amount"].(float64)
					if toPocketID == "" || amount == 0 {
						return pocket.ErrNotFound
					}
					result := deps.PocketService.ChangeBalance(toPocketID, amount)
					if result.IsError() {
						return result.Error()
					}
					return nil
				},
				Compensate: func(ctx context.Context, payload map[string]interface{}) error {
					toPocketID, _ := payload["to_pocket_id"].(string)
					amount, _ := payload["amount"].(float64)
					if toPocketID == "" || amount == 0 {
						return nil
					}
					result := deps.PocketService.ChangeBalance(toPocketID, -amount)
					if result.IsError() {
						return result.Error()
					}
					return nil
				},
			},
			{
				Name: "complete_transfer",
				Execute: func(ctx context.Context, payload map[string]interface{}) error {
					transferID, _ := payload["transfer_id"].(string)
					if transferID == "" {
						return transfer.ErrNotFound
					}
					result := deps.TransferService.CompleteTransfer(transferID)
					if result.IsError() {
						return result.Error()
					}
					return nil
				},
				Compensate: func(ctx context.Context, payload map[string]interface{}) error {
					transferID, _ := payload["transfer_id"].(string)
					if transferID == "" {
						return nil
					}
					result := deps.TransferService.FailTransfer(transferID, "saga compensation")
					if result.IsError() {
						return result.Error()
					}
					return nil
				},
			},
		},
	}
}
