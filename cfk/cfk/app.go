package cfk

import (
	"cfk/internal/finance/cashflowin"
	"cfk/internal/finance/cashflowout"
	"cfk/internal/finance/category"
	"cfk/internal/finance/pocket"
	"cfk/internal/finance/sagas"
	"cfk/internal/finance/transfer"
	"cfk/internal/identity/tenant"
	"cfk/internal/identity/user"
	"cfk/internal/observability/requestlog"
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
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

type App struct {
	db        *gorm.DB
	eventBus  *event.Bus
	cancel    context.CancelFunc
	fiber     *fiber.App
}

type Config struct {
	Port         string
	WorkerPool   int
	BufferSize   int
	MaxRetries   int
}

func DefaultConfig() Config {
	return Config{
		Port:       ":3000",
		WorkerPool: 4,
		BufferSize: 1024,
		MaxRetries: 3,
	}
}

func NewApp(cfg Config) (*App, error) {
	db, err := database.NewPostgresDB()
	if err != nil {
		return nil, err
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
		&requestlog.RequestLogProjection{},
	); err != nil {
		return nil, err
	}

	if err := db.Use(tracing.NewPlugin()); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	eventBus := event.NewBus(
		event.WithWorkerPool(cfg.WorkerPool),
		event.WithBufferSize(cfg.BufferSize),
		event.WithMaxRetries(cfg.MaxRetries),
	)

	SubscribeProjectionHandlers(eventBus, db)

	projectionHandler := handlers.NewProjectionHandler(db)
	eventBus.Subscribe("*", projectionHandler.Handle)

	sagaStore := saga.NewGORMStore(db)
	sagaOrchestrator := saga.NewOrchestrator(sagaStore)

	tenantService := tenant.NewService(tenant.NewGORMRepository(db), eventBus)
	userService := user.NewService(user.NewGORMRepository(db), eventBus)
	pocketService := pocket.NewService(pocket.NewGORMRepository(db), eventBus)
	cashflowinService := cashflowin.NewService(cashflowin.NewGORMRepository(db), eventBus)
	cashflowoutService := cashflowout.NewService(cashflowout.NewGORMRepository(db), eventBus)
	transferService := transfer.NewService(transfer.NewGORMRepository(db), eventBus)
	categoryService := category.NewService(category.NewGORMRepository(db), eventBus)
	assetService := asset.NewService(asset.NewGORMRepository(db), eventBus)
	debtService := debt.NewService(debt.NewGORMRepository(db), eventBus)
	balancesheetService := balancesheet.NewService(balancesheet.NewGORMRepository(db), eventBus)

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

	if r := sagaOrchestrator.Recover(ctx); r.IsError() {
		log.Printf("saga recovery warning: %v", r.Error())
	}

	go func() {
		for failed := range eventBus.DeadLetters() {
			log.Printf("dead letter event: type=%s err=%v", failed.Event.EventType, failed.Err)
		}
	}()

	app := fiber.New()
	app.Use(middleware.TraceMiddleware())
	app.Use(middleware.RequestLoggerMiddleware())

	RegisterRoutes(app, RoutesDeps{
		TenantService:       tenantService,
		UserService:         userService,
		PocketService:       pocketService,
		CashflowInService:   cashflowinService,
		CashflowOutService:  cashflowoutService,
		TransferService:     transferService,
		CategoryService:     categoryService,
		AssetService:        assetService,
		DebtService:         debtService,
		BalanceSheetService: balancesheetService,
	})

	return &App{
		db:       db,
		eventBus: eventBus,
		cancel:   cancel,
		fiber:    app,
	}, nil
}

func (a *App) Run(addr string) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := a.fiber.Listen(addr); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	<-quit
	return a.Shutdown()
}

func (a *App) Shutdown() error {
	log.Println("shutting down server...")

	if err := a.fiber.Shutdown(); err != nil {
		log.Printf("server shutdown error: %v", err)
		return err
	}

	log.Println("draining event bus...")
	a.cancel()
	a.eventBus.Stop()

	log.Println("shutdown complete")
	return nil
}

func (a *App) DB() *gorm.DB {
	return a.db
}

func (a *App) EventBus() *event.Bus {
	return a.eventBus
}

func (a *App) Fiber() *fiber.App {
	return a.fiber
}
