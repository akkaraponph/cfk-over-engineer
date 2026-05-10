package main

import (
	"cfk/internal/asset"
	"cfk/internal/balancesheet"
	"cfk/internal/cashflowin"
	"cfk/internal/cashflowout"
	"cfk/internal/category"
	"cfk/internal/debt"
	"cfk/internal/pocket"
	"cfk/internal/tenant"
	"cfk/internal/transfer"
	"cfk/internal/user"
	"cfk/pkg/database"
	"cfk/pkg/event"
	"cfk/pkg/handlers"
	"cfk/pkg/middleware"
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

	eventBus := event.NewBus()

	projectionHandler := handlers.NewProjectionHandler(db)
	eventBus.Subscribe("*", projectionHandler.Handle)

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
