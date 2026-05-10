# CashFlowKub — Event-Sourcing CQRS Financial Management System

> Over-engineered modular monolith for personal & small-business financial management, built with Domain-Driven Design, Event Sourcing, CQRS, and the Saga pattern.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Fiber v3 HTTP API                            │
│  /tenants  /users  /pockets  /cashflowins  /cashflowouts  ...      │
├─────────────────────────────────────────────────────────────────────┤
│                     Middleware Layer                                 │
│  TenantResolver → FeatureGuard → Auth → RequestLogger               │
├──────────┬──────────┬──────────┬──────────┬────────────────────────┤
│ Identity │ Finance  │  Wealth  │Observab.  │   Infrastructure       │
│ Boundary  │ Boundary │ Boundary │ Boundary  │                        │
│          │          │          │          │                        │
│ tenant   │ pocket   │ asset   │requestlog │  event.Bus (async)     │
│ user     │ cashflow │ debt    │           │  saga.Orchestrator     │
│          │ transfer │ balance │           │  event.Store (GORM)    │
│          │ category │ sheet   │           │  saga.Store (GORM)     │
├──────────┴──────────┴──────────┴──────────┴────────────────────────┤
│                     PostgreSQL (GORM)                                │
│  event_store │ *_projections │ saga_instances                       │
└─────────────────────────────────────────────────────────────────────┘
```

## Design Principles

### Event-Driven Development

Every state change in the system is captured as an immutable event. Services never mutate state directly — they validate, construct an event, and publish it. Downstream handlers react to events asynchronously.

**Event flow:**

```
HTTP Request → Handler → Service.Validate()
                       → Service.Publish(event) → [buffered channel]
                                                  → Worker pool dispatches:
                                                    ├─ EventStoreHandler  (append to event_store)
                                                    ├─ ProjectionHandler  (update read models)
                                                    └─ Saga trigger       (orchestrate cross-domain)

                                                    On handler failure:
                                                    → Retry with exponential backoff
                                                    → Dead letter channel after max retries
```

**Key event bus properties:**

- Fire-and-forget: `Publish()` enqueues to a buffered channel and returns immediately
- Configurable goroutine worker pool (default: 4 workers)
- Retry with exponential backoff (default: 3 retries, 100ms base, 5s cap)
- Dead letter channel for permanently failed events
- Graceful shutdown via `context.Context` + `sync.WaitGroup`
- Wildcard subscription (`*`) for cross-cutting handlers (event store, logging)

### Domain-Driven Design

Each bounded context is a self-contained Go package with four layers:

| Layer | File | Responsibility |
|-------|------|----------------|
| **Domain** | `*_domain.go` | Entities, value objects, domain events, constants |
| **Port** | `*_port.go` | `Repository` interface (PortOut) |
| **UseCase** | `*_usecase.go` | `Service` — validation, event construction, orchestration |
| **Adapter** | `*_adapter.go` | GORM repository implementation, projection models |
| **Handler** | `*_handler.go` | Fiber HTTP handlers |
| **Projection** | `projections/handler.go` | Read-model update logic |

**Bounded contexts:**

| Boundary | Domains | Responsibility |
|----------|---------|----------------|
| `identity/` | tenant, user | Authentication, multi-tenancy, feature flags |
| `finance/` | pocket, cashflowin, cashflowout, transfer, category | Core money movement |
| `wealth/` | asset, debt, balancesheet | Portfolio and liability tracking |
| `observability/` | requestlog | Cross-cutting logging |

**Boundary rules:**

- No direct imports between boundaries (identity ↔ finance ↔ wealth)
- Communication only via events or saga orchestration
- Each boundary owns its projection handlers
- Cross-boundary coordination lives in `pkg/saga/` and saga definitions in `internal/finance/sagas/`

### Event Sourcing

No UPDATE or DELETE operations. Every state change appends a new event to the `event_store` table:

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Event ID (auto-generated) |
| aggregate_type | VARCHAR(100) | Domain type (e.g., `pocket`, `cashflowin`) |
| aggregate_id | UUID | Entity ID this event belongs to |
| event_type | VARCHAR(100) | Event type (e.g., `pocket.created`) |
| version | INT | Version for optimistic concurrency |
| payload | JSONB | Event data |
| metadata | JSONB | Tenant ID, user ID, timestamp, correlation ID |
| created_at | TIMESTAMP | Event timestamp |

Projection tables are rebuilt from events and serve read queries. They can be dropped and replayed at any time.

### CQRS

**Command side:** Services validate, construct events, publish to bus → events appended to `event_store`

**Query side:** Projection handlers subscribe to events → update projection tables → handlers query projections via repositories

### Saga Pattern

Cross-domain transactions are orchestrated by the saga engine:

```
Transfer Saga:
  Step 1: debit_source      → PocketService.ChangeBalance(from, -amount)
  Step 2: credit_destination → PocketService.ChangeBalance(to, +amount)
  Step 3: complete_transfer  → TransferService.CompleteTransfer(id)

  On step failure:
    → Compensate in reverse order
    → E.g., step 2 fails → credit source back → mark transfer failed
```

**Saga engine properties:**

- `saga.Instance` persisted to PostgreSQL (`saga_instances` table)
- State machine: `pending` → `executing` → `completed` / `compensating` → `failed`
- Crash recovery: `Recover()` loads incomplete instances on startup and resumes them
- Each step defines `Execute` and `Compensate` functions
- All methods use `mo.Result[T]` / `mo.Option[T]` for consistent error handling

### SaaS Multi-Tenancy

- Every tenant has a plan (`free`, `premium`, `enterprise`)
- Feature flags per tenant (`balance_sheet`, `debt`, `asset`, `transfer`, etc.)
- `TenantMiddleware` resolves tenant from `X-Tenant-Slug` header, validates active status
- `FeatureGuard` middleware blocks access to features not enabled for the tenant
- All read queries scoped by `tenant_id`

## Domain Events

| Domain | Events |
|--------|--------|
| Tenant | `tenant.created`, `tenant.activated`, `tenant.deactivated`, `tenant.plan_changed`, `tenant.feature_enabled`, `tenant.feature_disabled` |
| User | `user.registered`, `user.activated`, `user.deactivated`, `user.role_changed`, `user.profile_updated` |
| Pocket | `pocket.created`, `pocket.name_changed`, `pocket.balance_changed`, `pocket.deleted` |
| CashflowIn | `cashflowin.recorded`, `cashflowin.updated`, `cashflowin.deleted` |
| CashflowOut | `cashflowout.recorded`, `cashflowout.updated`, `cashflowout.deleted` |
| Transfer | `transfer.initiated`, `transfer.completed`, `transfer.failed`, `transfer.deleted` |
| Category | `category.created`, `category.updated`, `category.deleted` |
| Asset | `asset.recorded`, `asset.value_changed`, `asset.assigned_to_balancesheet`, `asset.unassigned_from_balancesheet` |
| Debt | `debt.recorded`, `debt.amount_changed`, `debt.assigned_to_balancesheet`, `debt.unassigned_from_balancesheet` |
| BalanceSheet | `balancesheet.created`, `balancesheet.updated` |
| RequestLog | `requestlog.recorded` |

## Sagas

| Saga | Steps | Trigger |
|------|-------|---------|
| `transfer` | debit_source → credit_destination → complete_transfer | `InitiateTransfer()` |
| `cashflowin` | credit_pocket | `RecordCashflowIn()` |
| `cashflowout` | debit_pocket | `RecordCashflowOut()` |

## Project Structure

```
cfk/
├── cmd/
│   ├── api/main.go              — API server entry point
│   └── seed/main.go             — Demo data seeder
├── internal/
│   ├── identity/
│   │   ├── tenant/              — Tenant domain (plan, features, SaaS)
│   │   ├── user/                — User domain (auth, profile)
│   │   └── projections/         — Identity projection handlers
│   ├── finance/
│   │   ├── pocket/              — Wallet/balance domain
│   │   ├── cashflowin/          — Income domain
│   │   ├── cashflowout/         — Expense domain
│   │   ├── transfer/            — Transfer domain
│   │   ├── category/            — Category domain
│   │   ├── sagas/               — Finance saga definitions
│   │   └── projections/         — Finance projection handlers
│   ├── wealth/
│   │   ├── asset/               — Asset domain
│   │   ├── debt/                — Debt domain
│   │   ├── balancesheet/        — Balance sheet domain
│   │   └── projections/         — Wealth projection handlers
│   └── observability/
│       ├── requestlog/          — Request logging domain
│       └── projections/         — Observability projection handlers
├── pkg/
│   ├── event/                   — Async event bus (goroutine worker pool)
│   ├── saga/                    — Saga engine (orchestrator, store, state machine)
│   ├── database/                — GORM setup, AutoMigrate, event store model
│   ├── handlers/                — Event store write handler
│   └── middleware/              — TenantResolver, FeatureGuard, Auth, RequestLogger
├── grafana/                     — Grafana + Loki provisioning
├── docs/
│   ├── plan.md
│   └── db.md
├── compose.yml
├── go.mod
└── README.md
```

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.26 |
| HTTP Framework | Fiber v3 |
| ORM | GORM + PostgreSQL |
| Functional Types | samber/mo (Result, Option) |
| Observability | Grafana + Loki |
| Containerization | Docker Compose |

## Getting Started

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- PostgreSQL (or use compose)

### Run with Docker Compose

```bash
docker compose up -d
```

This starts:
- PostgreSQL on `:5432`
- Loki on `:3100`
- Grafana on `:3001` (admin/admin)

### Run API Server

```bash
# Set database URL
export DATABASE_URL="host=localhost port=5432 user=postgres password=postgres dbname=cfk sslmode=disable TimeZone=UTC"

# Run with auto-migration
go run ./cmd/api
```

### Seed Demo Data

```bash
go run ./cmd/seed
```

Creates a premium tenant, admin user, 3 pockets, 7 categories, cashflow records, a transfer, assets, debts, and a balance sheet.

### Run Tests

```bash
go test ./...                    # all tests
go test -v ./internal/finance/transfer/  # specific domain
go test -count=1 ./...          # no cache
```

104 tests across 12 packages covering domain validation, use case orchestration, event publishing, saga compensation, and bus worker pool behavior.

## API Endpoints

### Tenant Management (no auth required)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/tenants` | Create tenant (with plan) |
| GET | `/api/v1/tenants/:slug` | Get tenant by slug |
| PUT | `/api/v1/tenants/:id/plan` | Change plan |
| POST | `/api/v1/tenants/:id/activate` | Activate tenant |
| POST | `/api/v1/tenants/:id/deactivate` | Deactivate tenant |
| POST | `/api/v1/tenants/:id/features` | Enable feature |
| DELETE | `/api/v1/tenants/:id/features` | Disable feature |
| GET | `/api/v1/tenants/:id/features?feature=X` | Check feature |

### Tenant-Scoped Endpoints (require `X-Tenant-Slug` header)

| Method | Path | Feature Guard | Description |
|--------|------|---------------|-------------|
| POST | `/api/v1/users` | — | Register user |
| GET | `/api/v1/users/email/:email` | — | Get user by email |
| POST | `/api/v1/pockets` | — | Create pocket |
| GET | `/api/v1/pockets/:id` | — | Get pocket |
| POST | `/api/v1/cashflowins` | — | Record income |
| POST | `/api/v1/cashflowouts` | — | Record expense |
| POST | `/api/v1/transfers` | `transfer` | Initiate transfer |
| POST | `/api/v1/categories` | — | Create category |
| POST | `/api/v1/assets` | `balance_sheet` | Record asset |
| POST | `/api/v1/debts` | `balance_sheet` | Record debt |
| POST | `/api/v1/balancesheets` | `balance_sheet` | Create balance sheet |

## Type System

The project uses `samber/mo` throughout for type-safe error handling:

```go
// Services return mo.Result[T]
func (s *Service) CreatePocket(...) mo.Result[Pocket]

// Repositories return mo.Option[T] for single reads
func (r *Repo) FindByID(id string) mo.Option[Pocket]

// Repositories return mo.Result[[]T] for list reads
func (r *Repo) FindByUser(...) mo.Result[[]Pocket]

// Event handlers return mo.Result[struct{}]
func (h *Handler) HandlePocket(evt event.Event) mo.Result[struct{}]

// Saga steps return mo.Result[struct{}]
func Execute(ctx context.Context, payload map[string]interface{}) mo.Result[struct{}]
```

Unwrapping: `result, err := service.Method(args).Get()`

## Key Decisions

1. **Append-only event store** — no UPDATE or DELETE, all state changes are events
2. **Pre-generated UUIDs** — IDs created before event publish, callers get complete entities
3. **Soft deletes via events** — delete creates a tombstone event, projections filter `is_deleted`
4. **Tenant isolation** — all queries scoped by `tenant_id`, middleware resolves tenant from header
5. **Feature flags** — SaaS plan system with per-tenant feature gates
6. **Saga persistence** — saga state in PostgreSQL for crash recovery
7. **Fire-and-forget publishing** — async event bus, errors go to dead letter channel
8. **Projection rebuild** — can replay all events to rebuild any projection table
9. **Boundary isolation** — no direct imports between identity, finance, wealth
10. **TDD** — 104 tests, domain validation before orchestration, mock repos + async event recording
