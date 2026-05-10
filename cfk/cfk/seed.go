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
	"cfk/pkg/database"
	"cfk/pkg/event"
	"context"

	"gorm.io/gorm"
)

type SeedServices struct {
	TenantService       *tenant.Service
	UserService         *user.Service
	PocketService       *pocket.Service
	CategoryService     *category.Service
	CashflowInService   *cashflowin.Service
	CashflowOutService  *cashflowout.Service
	TransferService     *transfer.Service
	AssetService        *asset.Service
	DebtService         *debt.Service
	BalanceSheetService *balancesheet.Service
	EventBus            *event.Bus
	Cancel              context.CancelFunc
}

func NewSeedServices(db *gorm.DB) (*SeedServices, error) {
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
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	eventBus := event.NewBus(event.WithWorkerPool(2), event.WithBufferSize(512))
	eventBus.Start(ctx)

	return &SeedServices{
		TenantService:       tenant.NewService(tenant.NewGORMRepository(db), eventBus),
		UserService:         user.NewService(user.NewGORMRepository(db), eventBus),
		PocketService:       pocket.NewService(pocket.NewGORMRepository(db), eventBus),
		CategoryService:     category.NewService(category.NewGORMRepository(db), eventBus),
		CashflowInService:   cashflowin.NewService(cashflowin.NewGORMRepository(db), eventBus),
		CashflowOutService:  cashflowout.NewService(cashflowout.NewGORMRepository(db), eventBus),
		TransferService:     transfer.NewService(transfer.NewGORMRepository(db), eventBus),
		AssetService:        asset.NewService(asset.NewGORMRepository(db), eventBus),
		DebtService:         debt.NewService(debt.NewGORMRepository(db), eventBus),
		BalanceSheetService: balancesheet.NewService(balancesheet.NewGORMRepository(db), eventBus),
		EventBus:            eventBus,
		Cancel:              cancel,
	}, nil
}

func (s *SeedServices) Close() {
	s.Cancel()
	s.EventBus.Stop()
}
