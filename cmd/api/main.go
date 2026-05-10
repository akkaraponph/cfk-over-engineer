package main

import (
	"cfk/internal/finance/cashflowin"
	"cfk/internal/finance/cashflowout"
	"cfk/internal/finance/category"
	"cfk/internal/finance/pocket"
	financeprojections "cfk/internal/finance/projections"
	"cfk/internal/finance/sagas"
	"cfk/internal/finance/transfer"
	identityprojections "cfk/internal/identity/projections"
	"cfk/internal/identity/tenant"
	"cfk/internal/identity/user"
	obsprojections "cfk/internal/observability/projections"
	wealthprojections "cfk/internal/wealth/projections"
	"cfk/internal/wealth/asset"
	"cfk/internal/wealth/balancesheet"
	"cfk/internal/wealth/debt"
	"cfk/pkg/database"
	"cfk/pkg/event"
	"cfk/pkg/handlers"
	"cfk/pkg/middleware"
	"cfk/pkg/saga"
	"context"
	"log"

	"github.com/gofiber/fiber/v3"
)

func main() {
	db, err := database.NewPostgresDB()
	if err != nil {
		log.Fatal(err)
	}

	if err := database.AutoMigrate(db); err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eventBus := event.NewBus(
		event.WithWorkerPool(4),
		event.WithBufferSize(1024),
		event.WithMaxRetries(3),
	)

	projectionHandler := handlers.NewProjectionHandler(db)
	eventBus.Subscribe("*", projectionHandler.Handle)

	identityProjHandler := identityprojections.NewIdentityProjectionHandler(db)
	eventBus.Subscribe("tenant.created", identityProjHandler.HandleTenant)
	eventBus.Subscribe("tenant.activated", identityProjHandler.HandleTenant)
	eventBus.Subscribe("tenant.deactivated", identityProjHandler.HandleTenant)
	eventBus.Subscribe("user.registered", identityProjHandler.HandleUser)
	eventBus.Subscribe("user.activated", identityProjHandler.HandleUser)
	eventBus.Subscribe("user.deactivated", identityProjHandler.HandleUser)
	eventBus.Subscribe("user.role_changed", identityProjHandler.HandleUser)
	eventBus.Subscribe("user.profile_updated", identityProjHandler.HandleUser)

	financeProjHandler := financeprojections.NewFinanceProjectionHandler(db)
	eventBus.Subscribe("pocket.created", financeProjHandler.HandlePocket)
	eventBus.Subscribe("pocket.name_changed", financeProjHandler.HandlePocket)
	eventBus.Subscribe("pocket.balance_changed", financeProjHandler.HandlePocket)
	eventBus.Subscribe("pocket.deleted", financeProjHandler.HandlePocket)
	eventBus.Subscribe("cashflowin.recorded", financeProjHandler.HandleCashflowIn)
	eventBus.Subscribe("cashflowin.updated", financeProjHandler.HandleCashflowIn)
	eventBus.Subscribe("cashflowin.deleted", financeProjHandler.HandleCashflowIn)
	eventBus.Subscribe("cashflowout.recorded", financeProjHandler.HandleCashflowOut)
	eventBus.Subscribe("cashflowout.updated", financeProjHandler.HandleCashflowOut)
	eventBus.Subscribe("cashflowout.deleted", financeProjHandler.HandleCashflowOut)
	eventBus.Subscribe("transfer.initiated", financeProjHandler.HandleTransfer)
	eventBus.Subscribe("transfer.completed", financeProjHandler.HandleTransfer)
	eventBus.Subscribe("transfer.failed", financeProjHandler.HandleTransfer)
	eventBus.Subscribe("transfer.deleted", financeProjHandler.HandleTransfer)
	eventBus.Subscribe("category.created", financeProjHandler.HandleCategory)
	eventBus.Subscribe("category.updated", financeProjHandler.HandleCategory)
	eventBus.Subscribe("category.deleted", financeProjHandler.HandleCategory)

	wealthProjHandler := wealthprojections.NewWealthProjectionHandler(db)
	eventBus.Subscribe("asset.recorded", wealthProjHandler.HandleAsset)
	eventBus.Subscribe("asset.value_changed", wealthProjHandler.HandleAsset)
	eventBus.Subscribe("asset.assigned_to_balancesheet", wealthProjHandler.HandleAsset)
	eventBus.Subscribe("asset.unassigned_from_balancesheet", wealthProjHandler.HandleAsset)
	eventBus.Subscribe("debt.recorded", wealthProjHandler.HandleDebt)
	eventBus.Subscribe("debt.amount_changed", wealthProjHandler.HandleDebt)
	eventBus.Subscribe("debt.assigned_to_balancesheet", wealthProjHandler.HandleDebt)
	eventBus.Subscribe("debt.unassigned_from_balancesheet", wealthProjHandler.HandleDebt)
	eventBus.Subscribe("balancesheet.created", wealthProjHandler.HandleBalanceSheet)
	eventBus.Subscribe("balancesheet.updated", wealthProjHandler.HandleBalanceSheet)

	obsProjHandler := obsprojections.NewObservabilityProjectionHandler(db)
	eventBus.Subscribe("requestlog.recorded", obsProjHandler.HandleRequestLog)

	sagaStore := saga.NewGORMStore(db)
	sagaOrchestrator := saga.NewOrchestrator(sagaStore)

	tenantRepo := tenant.NewGORMRepository(db)
	tenantService := tenant.NewService(tenantRepo, eventBus)
	tenantHandler := tenant.NewHandler(tenantService)

	userRepo := user.NewGORMRepository(db)
	userService := user.NewService(userRepo, eventBus)
	userHandler := user.NewHandler(userService)

	pocketRepo := pocket.NewGORMRepository(db)
	pocketService := pocket.NewService(pocketRepo, eventBus)
	pocketHandler := pocket.NewHandler(pocketService)

	cashflowinRepo := cashflowin.NewGORMRepository(db)
	cashflowinService := cashflowin.NewService(cashflowinRepo, eventBus)
	cashflowinHandler := cashflowin.NewHandler(cashflowinService)

	cashflowoutRepo := cashflowout.NewGORMRepository(db)
	cashflowoutService := cashflowout.NewService(cashflowoutRepo, eventBus)
	cashflowoutHandler := cashflowout.NewHandler(cashflowoutService)

	transferRepo := transfer.NewGORMRepository(db)
	transferService := transfer.NewService(transferRepo, eventBus)
	transferHandler := transfer.NewHandler(transferService)

	categoryRepo := category.NewGORMRepository(db)
	categoryService := category.NewService(categoryRepo, eventBus)
	categoryHandler := category.NewHandler(categoryService)

	assetRepo := asset.NewGORMRepository(db)
	assetService := asset.NewService(assetRepo, eventBus)
	assetHandler := asset.NewHandler(assetService)

	debtRepo := debt.NewGORMRepository(db)
	debtService := debt.NewService(debtRepo, eventBus)
	debtHandler := debt.NewHandler(debtService)

	balancesheetRepo := balancesheet.NewGORMRepository(db)
	balancesheetService := balancesheet.NewService(balancesheetRepo, eventBus)
	balancesheetHandler := balancesheet.NewHandler(balancesheetService)

	sagaOrchestrator.Register(sagas.NewTransferSaga(sagas.TransferSagaDeps{
		TransferService: transferService,
		PocketService:   pocketService,
	}))
	sagaOrchestrator.Register(sagas.NewCashflowInSaga(sagas.CashflowInSagaDeps{
		PocketService: pocketService,
	}))
	sagaOrchestrator.Register(sagas.NewCashflowOutSaga(sagas.CashflowOutSagaDeps{
		PocketService: pocketService,
	}))

	transferService.SetSagaOrchestrator(sagaOrchestrator)
	cashflowinService.SetSagaOrchestrator(sagaOrchestrator)
	cashflowoutService.SetSagaOrchestrator(sagaOrchestrator)

	eventBus.Start(ctx)

	if err := sagaOrchestrator.Recover(ctx); err != nil {
		log.Printf("saga recovery warning: %v", err)
	}

	go func() {
		for failed := range eventBus.DeadLetters() {
			log.Printf("dead letter event: type=%s err=%v", failed.Event.EventType, failed.Err)
		}
	}()

	app := fiber.New()

	app.Use(middleware.RequestLoggerMiddleware())

	api := app.Group("/api/v1")

	tenants := api.Group("/tenants")
	tenants.Post("/", tenantHandler.CreateTenant)
	tenants.Get("/:slug", tenantHandler.GetTenantBySlug)

	users := api.Group("/users")
	users.Post("/", userHandler.RegisterUser)
	users.Get("/email/:email", userHandler.GetUserByEmail)

	pockets := api.Group("/pockets")
	pockets.Post("/", pocketHandler.CreatePocket)
	pockets.Get("/:id", pocketHandler.GetPocketByID)
	pockets.Get("/user/:userId", pocketHandler.ListPocketsByUser)

	cashflowins := api.Group("/cashflowins")
	cashflowins.Post("/", cashflowinHandler.RecordCashflowIn)
	cashflowins.Get("/:id", cashflowinHandler.GetCashflowInByID)
	cashflowins.Get("/pocket/:pocketId", cashflowinHandler.ListCashflowInsByPocket)

	cashflowouts := api.Group("/cashflowouts")
	cashflowouts.Post("/", cashflowoutHandler.RecordCashflowOut)
	cashflowouts.Get("/:id", cashflowoutHandler.GetCashflowOutByID)
	cashflowouts.Get("/pocket/:pocketId", cashflowoutHandler.ListCashflowOutsByPocket)

	transfers := api.Group("/transfers")
	transfers.Post("/", transferHandler.InitiateTransfer)
	transfers.Get("/:id", transferHandler.GetTransferByID)

	categories := api.Group("/categories")
	categories.Post("/", categoryHandler.CreateCategory)
	categories.Get("/:id", categoryHandler.GetCategoryByID)
	categories.Get("/", categoryHandler.ListCategoriesByTenant)

	assets := api.Group("/assets")
	assets.Post("/", assetHandler.RecordAsset)
	assets.Get("/:id", assetHandler.GetAssetByID)
	assets.Put("/:id/value", assetHandler.ChangeValue)
	assets.Post("/:id/balancesheet", assetHandler.AssignToBalanceSheet)
	assets.Delete("/:id/balancesheet", assetHandler.UnassignFromBalanceSheet)

	debts := api.Group("/debts")
	debts.Post("/", debtHandler.RecordDebt)
	debts.Get("/:id", debtHandler.GetDebtByID)
	debts.Put("/:id/amount", debtHandler.ChangeAmount)
	debts.Post("/:id/balancesheet", debtHandler.AssignToBalanceSheet)
	debts.Delete("/:id/balancesheet", debtHandler.UnassignFromBalanceSheet)

	balancesheets := api.Group("/balancesheets")
	balancesheets.Post("/", balancesheetHandler.CreateBalanceSheet)
	balancesheets.Get("/:id", balancesheetHandler.GetBalanceSheetByID)
	balancesheets.Put("/:id", balancesheetHandler.UpdateBalanceSheet)

	log.Fatal(app.Listen(":3000"))
}
