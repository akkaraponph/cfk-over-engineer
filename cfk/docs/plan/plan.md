# CFK — Cashflow Event Sourcing CQRS

## Architecture

Modular monolith, Ports & Adapters (hexagonal) architecture, append-only event store with PostgreSQL + GORM + Fiber v3.

### Core Principles

- **Event Sourcing**: No UPDATE or DELETE. Every state change appends a new event record.
- **CQRS**: Command side appends events to `event_store`. Query side reads from projection tables.
- **Publish-Only Pattern**: Services never call repos directly. They validate, build entities, publish events, and return results.
- **Three Function Types**: Pure Logic (no I/O), Side Effect (I/O, returns `mo.Result[T]`), Orchestration (composes both).

### Folder Structure

```
cfk/
├── cmd/api/
│   └── main.go                 — Binary entry point; wires all adapters
├── internal/
│   ├── tenant/
│   │   ├── domain/             — Models, constants, value objects
│   │   ├── port/               — UseCase (PortIn) and Repository (PortOut) interfaces
│   │   ├── adapter/
│   │   │   ├── *.go            — GORM repository implementations
│   │   │   └── http/           — HTTP handlers (package httpadapter)
│   │   └── application/        — Service (orchestration + pure logic)
│   ├── user/
│   ├── pocket/
│   ├── cashflowin/
│   ├── cashflowout/
│   ├── transfer/
│   ├── category/
│   ├── asset/
│   ├── debt/
│   ├── balancesheet/
│   ├── auditlog/
│   └── requestlog/
├── pkg/
│   ├── event/                  — Synchronous in-process event bus
│   ├── database/               — GORM PostgreSQL connection
│   └── middleware/             — Fiber middleware (auth, tenant, logging)
├── docs/
│   ├── plan.md
│   └── db.md
├── go.mod
└── compose.yml
```

### Event Flow

```
HTTP Request
  → Fiber Handler (adapter/http/)
    → port.UseCase.Method()
      → application.Service.Method()
        ├── Pure Logic: validate input
        ├── Pure Logic: build domain entity (pre-generate UUID)
        ├── Side Effect: publisher.Publish(event)
        └── return entity to caller

EventBus dispatches to:
  ├── SaveHandler → repo.Append(event) → event_store table
  ├── ProjectionHandler → update projection tables (pocket, cashflowin, etc.)
  ├── AuditHandler → audit_log projection
  └── NotificationHandler → downstream side effects
```

### Event Store Schema

All events stored in a single `event_store` table:

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Event ID (primary key) |
| aggregate_type | VARCHAR | Domain type (e.g., "pocket", "cashflowin") |
| aggregate_id | UUID | Entity ID this event belongs to |
| event_type | VARCHAR | Event type (e.g., "pocket.created", "pocket.balance_changed") |
| version | INT | Event version for optimistic concurrency |
| payload | JSONB | Event data |
| metadata | JSONB | Tenant ID, user ID, timestamp, etc. |
| created_at | TIMESTAMP | Event timestamp |

### Projection Tables

Projection tables are updated by event handlers and serve read queries:

- `tenant_projections` — Current tenant state
- `user_projections` — Current user state
- `pocket_projections` — Current pocket balances
- `cashflowin_projections` — Cash inflow records
- `cashflowout_projections` — Cash outflow records
- `transfer_projections` — Transfer records
- `category_projections` — Category records
- `asset_projections` — Asset records
- `debt_projections` — Debt records
- `balancesheet_projections` — Balance sheet records
- `audit_log_projections` — Audit log records
- `request_log_projections` — Request log records

### Domain Events

#### Tenant
- `tenant.created`
- `tenant.activated`
- `tenant.deactivated`
- `tenant.feature_enabled`
- `tenant.feature_disabled`

#### User
- `user.registered`
- `user.activated`
- `user.deactivated`
- `user.role_changed`
- `user.profile_updated`

#### Pocket
- `pocket.created`
- `pocket.name_changed`
- `pocket.balance_changed`
- `pocket.deleted`

#### CashflowIn
- `cashflowin.recorded`
- `cashflowin.updated`
- `cashflowin.deleted`

#### CashflowOut
- `cashflowout.recorded`
- `cashflowout.updated`
- `cashflowout.deleted`

#### Transfer
- `transfer.initiated`
- `transfer.completed`
- `transfer.failed`
- `transfer.deleted`

#### Category
- `category.created`
- `category.updated`
- `category.deleted`

#### Asset
- `asset.recorded`
- `asset.value_changed`
- `asset.assigned_to_balance_sheet`
- `asset.unassigned_from_balance_sheet`

#### Debt
- `debt.recorded`
- `debt.amount_changed`
- `debt.assigned_to_balance_sheet`
- `debt.unassigned_from_balance_sheet`

#### BalanceSheet
- `balancesheet.created`
- `balancesheet.updated`

### Technology Stack

- **Framework**: Fiber v3 (`github.com/gofiber/fiber/v3`)
- **ORM**: GORM (`gorm.io/gorm` + `gorm.io/driver/postgres`)
- **Database**: PostgreSQL
- **Error Handling**: `github.com/samber/mo` (Result, Option, Either types)
- **UUID**: `github.com/google/uuid`
- **Password Hashing**: `golang.org/x/crypto/bcrypt`

### Key Decisions

1. **No shared services between domains** — each domain is self-contained
2. **Pre-generate UUIDs** before publishing events so callers get complete entities
3. **Append-only repositories** — no Update/Delete methods on any Repository interface
4. **Soft deletes via events** — delete creates a tombstone event
5. **Tenant isolation** — all queries scoped by tenant_id
6. **Optimistic concurrency** — version field in event_store prevents lost updates
7. **Projection rebuild** — can replay all events to rebuild projections at any time

### Commands

```bash
go build ./cmd/api          # build binary
go run ./cmd/api            # run application
go test ./...               # run all tests
go get <pkg>@latest && go mod tidy   # add dependency
```

### Implementation Order

1. Core infrastructure (event bus, GORM setup, Fiber app)
2. Tenant domain (foundation for multi-tenancy)
3. User domain (authentication & authorization)
4. Category domain (used by cashflow domains)
5. Pocket domain (wallet concept)
6. CashflowIn domain
7. CashflowOut domain
8. Transfer domain
9. Asset domain
10. Debt domain
11. BalanceSheet domain
12. AuditLog domain
13. RequestLog domain
14. Middleware (auth, tenant resolution, logging)
15. Integration tests
