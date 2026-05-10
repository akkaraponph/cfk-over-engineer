package main

import (
	"cfk/cfk"
	"cfk/pkg/database"
	"fmt"
	"log"
)

func main() {
	db, err := database.NewPostgresDB()
	if err != nil {
		log.Fatal(err)
	}

	services, err := cfk.NewSeedServices(db)
	if err != nil {
		log.Fatal(err)
	}
	defer services.Close()

	seed(services)
}

func seed(s *cfk.SeedServices) {
	fmt.Println("=== Seeding demo data ===")

	t, err := s.TenantService.CreateTenant("CashFlowKub Demo", "cfk-demo", "premium").Get()
	if err != nil {
		log.Fatalf("seed tenant: %v", err)
	}
	fmt.Printf("Tenant: %s (%s) plan=%s\n", t.Name, t.ID, t.Plan)

	s.TenantService.EnableFeature(t.ID, "balance_sheet", t.ID)
	s.TenantService.EnableFeature(t.ID, "debt", t.ID)
	s.TenantService.EnableFeature(t.ID, "asset", t.ID)
	s.TenantService.EnableFeature(t.ID, "transfer", t.ID)
	fmt.Println("Features: balance_sheet, debt, asset, transfer enabled")

	u, err := s.UserService.RegisterUser(t.ID, "akira", "akira@cfk.demo", "password123", "Akira", "Ph", "0812345678", "admin").Get()
	if err != nil {
		log.Fatalf("seed user: %v", err)
	}
	fmt.Printf("User: %s %s (%s)\n", u.FirstName, u.LastName, u.ID)

	pocketWallet, err := s.PocketService.CreatePocket(t.ID, "Wallet", u.ID).Get()
	if err != nil {
		log.Fatalf("seed pocket wallet: %v", err)
	}
	fmt.Printf("Pocket: Wallet (%s)\n", pocketWallet.ID)

	pocketSavings, err := s.PocketService.CreatePocket(t.ID, "Savings", u.ID).Get()
	if err != nil {
		log.Fatalf("seed pocket savings: %v", err)
	}
	fmt.Printf("Pocket: Savings (%s)\n", pocketSavings.ID)

	pocketInvestment, err := s.PocketService.CreatePocket(t.ID, "Investment", u.ID).Get()
	if err != nil {
		log.Fatalf("seed pocket investment: %v", err)
	}
	fmt.Printf("Pocket: Investment (%s)\n", pocketInvestment.ID)

	catSalary, err := s.CategoryService.CreateCategory(t.ID, "Salary", "Monthly salary", "income", false, u.ID).Get()
	if err != nil {
		log.Fatalf("seed category: %v", err)
	}
	catFreelance, _ := s.CategoryService.CreateCategory(t.ID, "Freelance", "Freelance income", "income", true, u.ID).Get()
	catFood, _ := s.CategoryService.CreateCategory(t.ID, "Food", "Food and dining", "expense", false, u.ID).Get()
	catRent, _ := s.CategoryService.CreateCategory(t.ID, "Rent", "Monthly rent", "expense", false, u.ID).Get()
	catTransport, _ := s.CategoryService.CreateCategory(t.ID, "Transport", "Transportation", "expense", false, u.ID).Get()
	catInvestmentCat, _ := s.CategoryService.CreateCategory(t.ID, "Investment", "Stock & crypto", "investment", true, u.ID).Get()
	catSaving, _ := s.CategoryService.CreateCategory(t.ID, "Emergency Fund", "Emergency savings", "saving", true, u.ID).Get()
	fmt.Printf("Categories: %s, %s, %s, %s, %s, %s, %s\n", catSalary.ID, catFreelance.ID, catFood.ID, catRent.ID, catTransport.ID, catInvestmentCat.ID, catSaving.ID)

	s.PocketService.ChangeBalance(pocketWallet.ID, 50000)
	fmt.Println("Wallet initial balance: 50,000")

	cfIn1, _ := s.CashflowInService.RecordCashflowIn(t.ID, u.ID, pocketWallet.ID, 1, 30000, "January salary", "").Get()
	cfIn2, _ := s.CashflowInService.RecordCashflowIn(t.ID, u.ID, pocketWallet.ID, 1, 30000, "February salary", "").Get()
	cfIn3, _ := s.CashflowInService.RecordCashflowIn(t.ID, u.ID, pocketWallet.ID, 2, 15000, "Freelance project", "").Get()
	fmt.Printf("CashflowIn: %s, %s, %s\n", cfIn1.ID, cfIn2.ID, cfIn3.ID)

	cfOut1, _ := s.CashflowOutService.RecordCashflowOut(t.ID, u.ID, pocketWallet.ID, 3, 5000, "Monthly rent", "", "fixed").Get()
	cfOut2, _ := s.CashflowOutService.RecordCashflowOut(t.ID, u.ID, pocketWallet.ID, 4, 3000, "Food and dining", "", "variable").Get()
	cfOut3, _ := s.CashflowOutService.RecordCashflowOut(t.ID, u.ID, pocketWallet.ID, 5, 2000, "Gas and transport", "", "variable").Get()
	fmt.Printf("CashflowOut: %s, %s, %s\n", cfOut1.ID, cfOut2.ID, cfOut3.ID)

	tr, _ := s.TransferService.InitiateTransfer(t.ID, u.ID, pocketWallet.ID, pocketSavings.ID, 10000).Get()
	fmt.Printf("Transfer: %s (Wallet->Savings 10,000)\n", tr.ID)

	a1, _ := s.AssetService.RecordAsset(t.ID, "liquid", "Emergency fund", u.ID, 200000, 5000).Get()
	a2, _ := s.AssetService.RecordAsset(t.ID, "investment", "Stock portfolio", u.ID, 500000, 30000).Get()
	fmt.Printf("Assets: %s, %s\n", a1.ID, a2.ID)

	d1, _ := s.DebtService.RecordDebt(t.ID, "long", "Home mortgage", u.ID, 3000000, 6.5, 15000, 1).Get()
	d2, _ := s.DebtService.RecordDebt(t.ID, "short", "Credit card", u.ID, 50000, 18.0, 5000, 2).Get()
	fmt.Printf("Debts: %s, %s\n", d1.ID, d2.ID)

	bs, _ := s.BalanceSheetService.CreateBalanceSheet(t.ID, u.ID, 2026).Get()
	s.AssetService.AssignToBalanceSheet(a1.ID, bs.ID)
	s.AssetService.AssignToBalanceSheet(a2.ID, bs.ID)
	s.DebtService.AssignToBalanceSheet(d1.ID, bs.ID)
	s.DebtService.AssignToBalanceSheet(d2.ID, bs.ID)
	fmt.Printf("BalanceSheet: %s (2026)\n", bs.ID)

	fmt.Println("=== Seed complete ===")
}
