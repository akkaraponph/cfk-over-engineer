package cfk

import (
	financeprojections "cfk/internal/finance/projections"
	identityprojections "cfk/internal/identity/projections"
	obsprojections "cfk/internal/observability/projections"
	wealthprojections "cfk/internal/wealth/projections"
	"cfk/pkg/event"

	"gorm.io/gorm"
)

func SubscribeProjectionHandlers(bus *event.Bus, db *gorm.DB) {
	identityProjHandler := identityprojections.NewIdentityProjectionHandler(db)
	bus.Subscribe("tenant.created", identityProjHandler.HandleTenant)
	bus.Subscribe("tenant.activated", identityProjHandler.HandleTenant)
	bus.Subscribe("tenant.deactivated", identityProjHandler.HandleTenant)
	bus.Subscribe("tenant.plan_changed", identityProjHandler.HandleTenant)
	bus.Subscribe("tenant.feature_enabled", identityProjHandler.HandleTenantFeature)
	bus.Subscribe("tenant.feature_disabled", identityProjHandler.HandleTenantFeature)
	bus.Subscribe("user.registered", identityProjHandler.HandleUser)
	bus.Subscribe("user.activated", identityProjHandler.HandleUser)
	bus.Subscribe("user.deactivated", identityProjHandler.HandleUser)
	bus.Subscribe("user.role_changed", identityProjHandler.HandleUser)
	bus.Subscribe("user.profile_updated", identityProjHandler.HandleUser)

	financeProjHandler := financeprojections.NewFinanceProjectionHandler(db)
	bus.Subscribe("pocket.created", financeProjHandler.HandlePocket)
	bus.Subscribe("pocket.name_changed", financeProjHandler.HandlePocket)
	bus.Subscribe("pocket.balance_changed", financeProjHandler.HandlePocket)
	bus.Subscribe("pocket.deleted", financeProjHandler.HandlePocket)
	bus.Subscribe("cashflowin.recorded", financeProjHandler.HandleCashflowIn)
	bus.Subscribe("cashflowin.updated", financeProjHandler.HandleCashflowIn)
	bus.Subscribe("cashflowin.deleted", financeProjHandler.HandleCashflowIn)
	bus.Subscribe("cashflowout.recorded", financeProjHandler.HandleCashflowOut)
	bus.Subscribe("cashflowout.updated", financeProjHandler.HandleCashflowOut)
	bus.Subscribe("cashflowout.deleted", financeProjHandler.HandleCashflowOut)
	bus.Subscribe("transfer.initiated", financeProjHandler.HandleTransfer)
	bus.Subscribe("transfer.completed", financeProjHandler.HandleTransfer)
	bus.Subscribe("transfer.failed", financeProjHandler.HandleTransfer)
	bus.Subscribe("transfer.deleted", financeProjHandler.HandleTransfer)
	bus.Subscribe("category.created", financeProjHandler.HandleCategory)
	bus.Subscribe("category.updated", financeProjHandler.HandleCategory)
	bus.Subscribe("category.deleted", financeProjHandler.HandleCategory)

	wealthProjHandler := wealthprojections.NewWealthProjectionHandler(db)
	bus.Subscribe("asset.recorded", wealthProjHandler.HandleAsset)
	bus.Subscribe("asset.value_changed", wealthProjHandler.HandleAsset)
	bus.Subscribe("asset.assigned_to_balancesheet", wealthProjHandler.HandleAsset)
	bus.Subscribe("asset.unassigned_from_balancesheet", wealthProjHandler.HandleAsset)
	bus.Subscribe("debt.recorded", wealthProjHandler.HandleDebt)
	bus.Subscribe("debt.amount_changed", wealthProjHandler.HandleDebt)
	bus.Subscribe("debt.assigned_to_balancesheet", wealthProjHandler.HandleDebt)
	bus.Subscribe("debt.unassigned_from_balancesheet", wealthProjHandler.HandleDebt)
	bus.Subscribe("balancesheet.created", wealthProjHandler.HandleBalanceSheet)
	bus.Subscribe("balancesheet.updated", wealthProjHandler.HandleBalanceSheet)

	obsProjHandler := obsprojections.NewObservabilityProjectionHandler(db)
	bus.Subscribe("requestlog.recorded", obsProjHandler.HandleRequestLog)
}
