package cfk

import (
	"cfk/internal/finance/cashflowin"
	"cfk/internal/finance/cashflowout"
	"cfk/internal/finance/category"
	"cfk/internal/finance/pocket"
	"cfk/internal/finance/transfer"
	"cfk/internal/identity/tenant"
	"cfk/internal/identity/user"
	"cfk/internal/wealth/asset"
	"cfk/internal/wealth/balancesheet"
	"cfk/internal/wealth/debt"
	"cfk/pkg/middleware"

	"github.com/gofiber/fiber/v3"
)

type RoutesDeps struct {
	TenantService       *tenant.Service
	UserService         *user.Service
	PocketService       *pocket.Service
	CashflowInService   *cashflowin.Service
	CashflowOutService  *cashflowout.Service
	TransferService     *transfer.Service
	CategoryService     *category.Service
	AssetService        *asset.Service
	DebtService         *debt.Service
	BalanceSheetService *balancesheet.Service
}

func RegisterRoutes(app *fiber.App, deps RoutesDeps) {
	tenantHandler := tenant.NewHandler(deps.TenantService)
	userHandler := user.NewHandler(deps.UserService)
	pocketHandler := pocket.NewHandler(deps.PocketService)
	cashflowinHandler := cashflowin.NewHandler(deps.CashflowInService)
	cashflowoutHandler := cashflowout.NewHandler(deps.CashflowOutService)
	transferHandler := transfer.NewHandler(deps.TransferService)
	categoryHandler := category.NewHandler(deps.CategoryService)
	assetHandler := asset.NewHandler(deps.AssetService)
	debtHandler := debt.NewHandler(deps.DebtService)
	balancesheetHandler := balancesheet.NewHandler(deps.BalanceSheetService)

	api := app.Group("/api/v1")

	tenants := api.Group("/tenants")
	tenants.Post("/", tenantHandler.CreateTenant)
	tenants.Get("/:slug", tenantHandler.GetTenantBySlug)
	tenants.Put("/:id/plan", middleware.AuthMiddleware(), middleware.RequireRole("admin"), tenantHandler.ChangePlan)
	tenants.Post("/:id/activate", middleware.AuthMiddleware(), middleware.RequireRole("admin"), tenantHandler.ActivateTenant)
	tenants.Post("/:id/deactivate", middleware.AuthMiddleware(), middleware.RequireRole("admin"), tenantHandler.DeactivateTenant)
	tenants.Post("/:id/features", middleware.AuthMiddleware(), middleware.RequireRole("admin"), tenantHandler.EnableFeature)
	tenants.Delete("/:id/features", middleware.AuthMiddleware(), middleware.RequireRole("admin"), tenantHandler.DisableFeature)
	tenants.Get("/:id/features", tenantHandler.CheckFeature)

	tenantGroup := api.Group("/", middleware.TenantMiddleware(deps.TenantService))

	users := tenantGroup.Group("/users")
	users.Post("/", userHandler.RegisterUser)
	users.Post("/login", userHandler.Login)
	users.Get("/email/:email", userHandler.GetUserByEmail)

	authGroup := tenantGroup.Group("/", middleware.AuthMiddleware())

	pockets := authGroup.Group("/pockets")
	pockets.Post("/", pocketHandler.CreatePocket)
	pockets.Get("/:id", pocketHandler.GetPocketByID)
	pockets.Get("/user/:userId", pocketHandler.ListPocketsByUser)

	cashflowins := authGroup.Group("/cashflowins")
	cashflowins.Post("/", cashflowinHandler.RecordCashflowIn)
	cashflowins.Get("/:id", cashflowinHandler.GetCashflowInByID)
	cashflowins.Get("/pocket/:pocketId", cashflowinHandler.ListCashflowInsByPocket)

	cashflowouts := authGroup.Group("/cashflowouts")
	cashflowouts.Post("/", cashflowoutHandler.RecordCashflowOut)
	cashflowouts.Get("/:id", cashflowoutHandler.GetCashflowOutByID)
	cashflowouts.Get("/pocket/:pocketId", cashflowoutHandler.ListCashflowOutsByPocket)

	transfers := authGroup.Group("/transfers", middleware.FeatureGuard(deps.TenantService, "transfer"))
	transfers.Post("/", transferHandler.InitiateTransfer)
	transfers.Get("/:id", transferHandler.GetTransferByID)

	categories := authGroup.Group("/categories")
	categories.Post("/", categoryHandler.CreateCategory)
	categories.Get("/:id", categoryHandler.GetCategoryByID)
	categories.Get("/", categoryHandler.ListCategoriesByTenant)

	wealthGroup := authGroup.Group("/", middleware.FeatureGuard(deps.TenantService, "balance_sheet"))

	assets := wealthGroup.Group("/assets")
	assets.Post("/", assetHandler.RecordAsset)
	assets.Get("/:id", assetHandler.GetAssetByID)
	assets.Put("/:id/value", assetHandler.ChangeValue)
	assets.Post("/:id/balancesheet", assetHandler.AssignToBalanceSheet)
	assets.Delete("/:id/balancesheet", assetHandler.UnassignFromBalanceSheet)

	debts := wealthGroup.Group("/debts")
	debts.Post("/", debtHandler.RecordDebt)
	debts.Get("/:id", debtHandler.GetDebtByID)
	debts.Put("/:id/amount", debtHandler.ChangeAmount)
	debts.Post("/:id/balancesheet", debtHandler.AssignToBalanceSheet)
	debts.Delete("/:id/balancesheet", debtHandler.UnassignFromBalanceSheet)

	balancesheets := wealthGroup.Group("/balancesheets")
	balancesheets.Post("/", balancesheetHandler.CreateBalanceSheet)
	balancesheets.Get("/:id", balancesheetHandler.GetBalanceSheetByID)
	balancesheets.Put("/:id", balancesheetHandler.UpdateBalanceSheet)
}
