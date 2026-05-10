package main

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
	"cfk/pkg/database"
	"cfk/pkg/event"
	"context"
	"fmt"
	"log"

	"gorm.io/gorm"
)

func main() {
	db, err := database.NewPostgresDB()
	if err != nil {
		log.Fatal(err)
	}

	if err := database.AutoMigrate(db,
		&tenant.TenantProjection{},
		&user.UserProjection{},
		&pocket.PocketProjection{},
		&cashflowin.CashflowInProjection{},
		&cashflowout.CashflowOutProjection{},
		&transfer.TransferProjection{},
		&category.CategoryProjection{},
		&asset.AssetProjection{},
		&debt.DebtProjection{},
		&balancesheet.BalanceSheetProjection{},
	); err != nil {
		log.Fatal(err)
	}

	seed(db)
}

func seed(db *gorm.DB) {
	ctx := context.Background()
	eventBus := event.NewBus(event.WithWorkerPool(2), event.WithBufferSize(512))
	eventBus.Start(ctx)
	defer eventBus.Stop()

	tenantRepo := tenant.NewGORMRepository(db)
	tenantService := tenant.NewService(tenantRepo, eventBus)

	userRepo := user.NewGORMRepository(db)
	userService := user.NewService(userRepo, eventBus)

	pocketRepo := pocket.NewGORMRepository(db)
	pocketService := pocket.NewService(pocketRepo, eventBus)

	categoryRepo := category.NewGORMRepository(db)
	categoryService := category.NewService(categoryRepo, eventBus)

	cashflowinRepo := cashflowin.NewGORMRepository(db)
	cashflowinService := cashflowin.NewService(cashflowinRepo, eventBus)

	cashflowoutRepo := cashflowout.NewGORMRepository(db)
	cashflowoutService := cashflowout.NewService(cashflowoutRepo, eventBus)

	transferRepo := transfer.NewGORMRepository(db)
	transferService := transfer.NewService(transferRepo, eventBus)

	assetRepo := asset.NewGORMRepository(db)
	assetService := asset.NewService(assetRepo, eventBus)

	debtRepo := debt.NewGORMRepository(db)
	debtService := debt.NewService(debtRepo, eventBus)

	balancesheetRepo := balancesheet.NewGORMRepository(db)
	balancesheetService := balancesheet.NewService(balancesheetRepo, eventBus)

	fmt.Println("=== Seeding demo data ===")

	t, err := tenantService.CreateTenant("CashFlowKub Demo", "cfk-demo", "premium").Get()
	if err != nil {
		log.Fatalf("seed tenant: %v", err)
	}
	fmt.Printf("Tenant: %s (%s) plan=%s\n", t.Name, t.ID, t.Plan)

	tenantService.EnableFeature(t.ID, "balance_sheet", t.ID)
	tenantService.EnableFeature(t.ID, "debt", t.ID)
	tenantService.EnableFeature(t.ID, "asset", t.ID)
	tenantService.EnableFeature(t.ID, "transfer", t.ID)
	fmt.Println("Features: balance_sheet, debt, asset, transfer enabled")

	u, err := userService.RegisterUser(t.ID, "akira", "akira@cfk.demo", "password123", "Akira", "Ph", "0812345678", "admin").Get()
	if err != nil {
		log.Fatalf("seed user: %v", err)
	}
	fmt.Printf("User: %s %s (%s)\n", u.FirstName, u.LastName, u.ID)

	pocketWallet, err := pocketService.CreatePocket(t.ID, "Wallet", u.ID).Get()
	if err != nil {
		log.Fatalf("seed pocket wallet: %v", err)
	}
	fmt.Printf("Pocket: Wallet (%s)\n", pocketWallet.ID)

	pocketSavings, err := pocketService.CreatePocket(t.ID, "Savings", u.ID).Get()
	if err != nil {
		log.Fatalf("seed pocket savings: %v", err)
	}
	fmt.Printf("Pocket: Savings (%s)\n", pocketSavings.ID)

	pocketInvestment, err := pocketService.CreatePocket(t.ID, "Investment", u.ID).Get()
	if err != nil {
		log.Fatalf("seed pocket investment: %v", err)
	}
	fmt.Printf("Pocket: Investment (%s)\n", pocketInvestment.ID)

	catSalary, err := categoryService.CreateCategory(t.ID, "Salary", "Monthly salary", "income", false, u.ID).Get()
	if err != nil {
		log.Fatalf("seed category: %v", err)
	}
	catFreelance, _ := categoryService.CreateCategory(t.ID, "Freelance", "Freelance income", "income", true, u.ID).Get()
	catFood, _ := categoryService.CreateCategory(t.ID, "Food", "Food and dining", "expense", false, u.ID).Get()
	catRent, _ := categoryService.CreateCategory(t.ID, "Rent", "Monthly rent", "expense", false, u.ID).Get()
	catTransport, _ := categoryService.CreateCategory(t.ID, "Transport", "Transportation", "expense", false, u.ID).Get()
	catInvestmentCat, _ := categoryService.CreateCategory(t.ID, "Investment", "Stock & crypto", "investment", true, u.ID).Get()
	catSaving, _ := categoryService.CreateCategory(t.ID, "Emergency Fund", "Emergency savings", "saving", true, u.ID).Get()
	fmt.Printf("Categories: %s, %s, %s, %s, %s, %s, %s\n", catSalary.ID, catFreelance.ID, catFood.ID, catRent.ID, catTransport.ID, catInvestmentCat.ID, catSaving.ID)

	pocketService.ChangeBalance(pocketWallet.ID, 50000)
	fmt.Println("Wallet initial balance: 50,000")

	cfIn1, _ := cashflowinService.RecordCashflowIn(t.ID, u.ID, pocketWallet.ID, 1, 30000, "January salary", "").Get()
	cfIn2, _ := cashflowinService.RecordCashflowIn(t.ID, u.ID, pocketWallet.ID, 1, 30000, "February salary", "").Get()
	cfIn3, _ := cashflowinService.RecordCashflowIn(t.ID, u.ID, pocketWallet.ID, 2, 15000, "Freelance project", "").Get()
	fmt.Printf("CashflowIn: %s, %s, %s\n", cfIn1.ID, cfIn2.ID, cfIn3.ID)

	cfOut1, _ := cashflowoutService.RecordCashflowOut(t.ID, u.ID, pocketWallet.ID, 3, 5000, "Monthly rent", "", "fixed").Get()
	cfOut2, _ := cashflowoutService.RecordCashflowOut(t.ID, u.ID, pocketWallet.ID, 4, 3000, "Food and dining", "", "variable").Get()
	cfOut3, _ := cashflowoutService.RecordCashflowOut(t.ID, u.ID, pocketWallet.ID, 5, 2000, "Gas and transport", "", "variable").Get()
	fmt.Printf("CashflowOut: %s, %s, %s\n", cfOut1.ID, cfOut2.ID, cfOut3.ID)

	tr, _ := transferService.InitiateTransfer(t.ID, u.ID, pocketWallet.ID, pocketSavings.ID, 10000).Get()
	fmt.Printf("Transfer: %s (Wallet->Savings 10,000)\n", tr.ID)

	a1, _ := assetService.RecordAsset(t.ID, "liquid", "Emergency fund", u.ID, 200000, 5000).Get()
	a2, _ := assetService.RecordAsset(t.ID, "investment", "Stock portfolio", u.ID, 500000, 30000).Get()
	fmt.Printf("Assets: %s, %s\n", a1.ID, a2.ID)

	d1, _ := debtService.RecordDebt(t.ID, "long", "Home mortgage", u.ID, 3000000, 6.5, 15000, 1).Get()
	d2, _ := debtService.RecordDebt(t.ID, "short", "Credit card", u.ID, 50000, 18.0, 5000, 2).Get()
	fmt.Printf("Debts: %s, %s\n", d1.ID, d2.ID)

	bs, _ := balancesheetService.CreateBalanceSheet(t.ID, u.ID, 2026).Get()
	assetService.AssignToBalanceSheet(a1.ID, bs.ID)
	assetService.AssignToBalanceSheet(a2.ID, bs.ID)
	debtService.AssignToBalanceSheet(d1.ID, bs.ID)
	debtService.AssignToBalanceSheet(d2.ID, bs.ID)
	fmt.Printf("BalanceSheet: %s (2026)\n", bs.ID)

	fmt.Println("=== Seed complete ===")
}
