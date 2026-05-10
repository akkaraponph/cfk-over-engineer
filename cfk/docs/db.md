// Cashflow DB schema (DBML) — Event Sourcing + CQRS
// Conventions: event_store is append-only; projections are read models rebuilt from events.
// Application enforces transaction FK rules and transfer self-transfer check.

// Enum definitions
Enum "role_enum" {
  "user"
  "premium"
  "admin"
}

Enum "category_enum" {
  "income"
  "expense"
  "investment"
  "saving"
}

Enum "cashflowout_enum" {
  "variable"
  "fixed"
  "investment"
  "saving"
}

Enum "transaction_enum" {
  "in"
  "out"
  "transfer"
}

Enum "asset_enum" {
  "liquid"
  "private"
  "investment"
  "intangible"
}

Enum "debt_enum" {
  "long"
  "short"
}

Enum "feature_enum" {
  "balance_sheet"
  "debt"
  "asset"
  "advanced_reporting"
  "api_access"
  "custom_categories"
}

Enum "log_level_enum" {
  "debug"
  "info"
  "warn"
  "error"
  "fatal"
}

Enum "log_action_enum" {
  "create"
  "read"
  "update"
  "delete"
  "login"
  "logout"
  "authorize"
  "unauthorized"
  "feature_enabled"
  "feature_disabled"
}

// ============================================
// EVENT STORE (Append-Only, Source of Truth)
// ============================================

Table "event_store" {
  id uuid [pk, default: `gen_random_uuid()`, note: "Unique event ID"]
  aggregate_type VARCHAR(100) [not null, note: "Domain type (e.g., 'pocket', 'cashflowin')"]
  aggregate_id uuid [not null, note: "Entity ID this event belongs to"]
  event_type VARCHAR(100) [not null, note: "Event type (e.g., 'pocket.created', 'pocket.balance_changed')"]
  version INT [not null, note: "Event version for optimistic concurrency per aggregate"]
  payload JSONB [not null, note: "Event data as JSON"]
  metadata JSONB [note: "Tenant ID, user ID, timestamp, correlation ID, etc."]
  created_at TIMESTAMP [default: `CURRENT_TIMESTAMP`, note: "Event timestamp"]
  indexes {
    (aggregate_type, aggregate_id, version) [unique, name: "uq_event_store_aggregate_version"]
    (aggregate_type, aggregate_id, created_at) [name: "idx_event_store_aggregate_created"]
    (event_type, created_at) [name: "idx_event_store_type_created"]
    (created_at) [name: "idx_event_store_created"]
  }
  Note: "APPEND-ONLY: No UPDATE or DELETE allowed. All state changes are new events."
}

// ============================================
// PROJECTION TABLES (Read Models)
// Rebuilt from event_store by projection handlers.
// Can be dropped and rebuilt at any time.
// ============================================

Table "tenant_projections" {
  id uuid [pk, note: "Tenant ID (from tenant.created event)"]
  name VARCHAR(255) [not null, note: "Name of the tenant/organization"]
  slug VARCHAR(255) [unique, not null, note: "Unique slug identifier"]
  is_active BOOLEAN [not null, default: true, note: "Active status"]
  created_at TIMESTAMP [not null, note: "From tenant.created event"]
  updated_at TIMESTAMP [not null, note: "From latest tenant.* event"]
}

Table "tenant_feature_projections" {
  id uuid [pk, default: `gen_random_uuid()`, note: "Unique ID"]
  tenant_id uuid [ref: > tenant_projections.id, not null, note: "References tenant"]
  feature feature_enum [not null, note: "Feature name"]
  is_enabled BOOLEAN [not null, default: true, note: "Enabled status"]
  enabled_at TIMESTAMP [note: "From tenant.feature_enabled event"]
  disabled_at TIMESTAMP [note: "From tenant.feature_disabled event"]
  enabled_by uuid [note: "User who enabled"]
  disabled_by uuid [note: "User who disabled"]
  created_at TIMESTAMP [not null, note: "From tenant.feature_enabled event"]
  updated_at TIMESTAMP [not null, note: "From latest tenant_feature.* event"]
  indexes {
    (tenant_id, feature) [unique, name: "uq_tenant_feature"]
  }
}

Table "user_projections" {
  id uuid [pk, note: "User ID (from user.registered event)"]
  tenant_id uuid [ref: > tenant_projections.id, not null, note: "References tenant"]
  username VARCHAR(255) [not null, note: "Username"]
  hashed_password VARCHAR(255) [not null, note: "Hashed password"]
  first_name VARCHAR(255) [note: "First name"]
  last_name VARCHAR(255) [note: "Last name"]
  phone VARCHAR(20) [note: "Phone number"]
  email VARCHAR(255) [not null, note: "Email address"]
  role role_enum [not null, note: "User role"]
  is_active BOOLEAN [not null, default: true, note: "Active status"]
  created_at TIMESTAMP [not null, note: "From user.registered event"]
  updated_at TIMESTAMP [not null, note: "From latest user.* event"]
  indexes {
    (tenant_id, email) [unique, name: "uq_user_email_per_tenant"]
  }
}

Table "category_projections" {
  id INT [pk, increment, note: "Category ID (from category.created event)"]
  tenant_id uuid [ref: > tenant_projections.id, not null, note: "References tenant"]
  name VARCHAR(255) [not null, note: "Category name"]
  description TEXT [note: "Category description"]
  type category_enum [not null, note: "Category type"]
  is_custom BOOLEAN [not null, default: false, note: "Custom category flag"]
  user_id uuid [note: "User who created"]
  is_deleted BOOLEAN [not null, default: false, note: "From category.deleted event"]
  created_at TIMESTAMP [not null, note: "From category.created event"]
  updated_at TIMESTAMP [not null, note: "From latest category.* event"]
  indexes {
    (tenant_id, name) [unique, name: "uq_category_name_per_tenant", note: "Only for non-deleted"]
  }
}

Table "pocket_projections" {
  id uuid [pk, note: "Pocket ID (from pocket.created event)"]
  tenant_id uuid [ref: > tenant_projections.id, not null, note: "References tenant"]
  name VARCHAR(255) [not null, note: "Pocket name"]
  balance DECIMAL(15,2) [not null, default: 0, note: "Current balance (from pocket.balance_changed events)"]
  user_id uuid [note: "Owner user"]
  is_deleted BOOLEAN [not null, default: false, note: "From pocket.deleted event"]
  created_at TIMESTAMP [not null, note: "From pocket.created event"]
  updated_at TIMESTAMP [not null, note: "From latest pocket.* event"]
  indexes {
    (tenant_id, user_id, name) [unique, name: "uq_pocket_name_per_user_tenant", note: "Only for non-deleted"]
  }
}

Table "cashflowin_projections" {
  id uuid [pk, note: "CashflowIn ID (from cashflowin.recorded event)"]
  tenant_id uuid [ref: > tenant_projections.id, not null, note: "References tenant"]
  amount DECIMAL(15,2) [not null, note: "Inflow amount"]
  description TEXT [note: "Description"]
  user_id uuid [note: "Associated user"]
  pocket_id uuid [ref: > pocket_projections.id, note: "Source pocket"]
  category_id INT [ref: > category_projections.id, note: "Category"]
  receipt TEXT [note: "Receipt file path/URL"]
  is_deleted BOOLEAN [not null, default: false, note: "From cashflowin.deleted event"]
  created_at TIMESTAMP [not null, note: "From cashflowin.recorded event"]
  updated_at TIMESTAMP [not null, note: "From latest cashflowin.* event"]
  indexes {
    (tenant_id, user_id, created_at) [name: "idx_cashflowin_tenant_user_created"]
    (pocket_id, created_at) [name: "idx_cashflowin_pocket_created"]
  }
}

Table "cashflowout_projections" {
  id uuid [pk, note: "CashflowOut ID (from cashflowout.recorded event)"]
  tenant_id uuid [ref: > tenant_projections.id, not null, note: "References tenant"]
  amount DECIMAL(15,2) [not null, note: "Outflow amount"]
  description TEXT [note: "Description"]
  type cashflowout_enum [not null, note: "Outflow type"]
  user_id uuid [note: "Associated user"]
  pocket_id uuid [ref: > pocket_projections.id, note: "Source pocket"]
  category_id INT [ref: > category_projections.id, note: "Category"]
  receipt TEXT [note: "Receipt file path/URL"]
  is_deleted BOOLEAN [not null, default: false, note: "From cashflowout.deleted event"]
  created_at TIMESTAMP [not null, note: "From cashflowout.recorded event"]
  updated_at TIMESTAMP [not null, note: "From latest cashflowout.* event"]
  indexes {
    (tenant_id, user_id, created_at) [name: "idx_cashflowout_tenant_user_created"]
    (pocket_id, created_at) [name: "idx_cashflowout_pocket_created"]
  }
}

Table "transfer_projections" {
  id uuid [pk, note: "Transfer ID (from transfer.initiated event)"]
  tenant_id uuid [ref: > tenant_projections.id, not null, note: "References tenant"]
  amount DECIMAL(15,2) [not null, note: "Transfer amount"]
  from_pocket_id uuid [ref: > pocket_projections.id, not null, note: "Source pocket"]
  to_pocket_id uuid [ref: > pocket_projections.id, not null, note: "Destination pocket"]
  user_id uuid [not null, note: "User who initiated"]
  status VARCHAR(50) [not null, default: 'completed', note: "From transfer.* events (initiated, completed, failed)"]
  is_deleted BOOLEAN [not null, default: false, note: "From transfer.deleted event"]
  created_at TIMESTAMP [not null, note: "From transfer.initiated event"]
  updated_at TIMESTAMP [not null, note: "From latest transfer.* event"]
  Note: "Application must enforce CHECK: from_pocket_id != to_pocket_id"
}

Table "transaction_projections" {
  id uuid [pk, note: "Transaction ID"]
  tenant_id uuid [ref: > tenant_projections.id, not null, note: "References tenant"]
  type transaction_enum [not null, note: "Transaction type"]
  cashflowin_id uuid [ref: > cashflowin_projections.id, note: "Reference if type=in"]
  cashflowout_id uuid [ref: > cashflowout_projections.id, note: "Reference if type=out"]
  transfer_id uuid [ref: > transfer_projections.id, note: "Reference if type=transfer"]
  user_id uuid [note: "Associated user"]
  created_at TIMESTAMP [not null, note: "From transaction event"]
  updated_at TIMESTAMP [not null, note: "From latest transaction event"]
  Note: "Exactly one of cashflowin_id, cashflowout_id, transfer_id is set based on type"
  indexes {
    (tenant_id, created_at) [name: "idx_transaction_tenant_created"]
    (user_id, created_at) [name: "idx_transaction_user_created"]
  }
}

Table "asset_projections" {
  id uuid [pk, note: "Asset ID (from asset.recorded event)"]
  tenant_id uuid [ref: > tenant_projections.id, not null, note: "References tenant"]
  type asset_enum [not null, note: "Asset type"]
  description TEXT [note: "Description"]
  value DECIMAL(15,2) [not null, note: "Current value (from asset.value_changed events)"]
  cashflow_per_year DECIMAL(15,2) [note: "Annual cashflow"]
  balance_sheet_id uuid [note: "Assigned balance sheet"]
  user_id uuid [note: "Owner user"]
  created_at TIMESTAMP [not null, note: "From asset.recorded event"]
  updated_at TIMESTAMP [not null, note: "From latest asset.* event"]
}

Table "debt_projections" {
  id uuid [pk, note: "Debt ID (from debt.recorded event)"]
  tenant_id uuid [ref: > tenant_projections.id, not null, note: "References tenant"]
  type debt_enum [not null, note: "Debt type"]
  description TEXT [note: "Description"]
  amount DECIMAL(15,2) [not null, note: "Current amount (from debt.amount_changed events)"]
  interest FLOAT [not null, note: "Interest rate"]
  minimum_pay DECIMAL(15,2) [not null, note: "Minimum payment"]
  priority INT [note: "Payoff priority"]
  balance_sheet_id uuid [note: "Assigned balance sheet"]
  user_id uuid [note: "Owner user"]
  created_at TIMESTAMP [not null, note: "From debt.recorded event"]
  updated_at TIMESTAMP [not null, note: "From latest debt.* event"]
}

Table "balancesheet_projections" {
  id uuid [pk, note: "BalanceSheet ID (from balancesheet.created event)"]
  tenant_id uuid [ref: > tenant_projections.id, not null, note: "References tenant"]
  user_id uuid [note: "Owner user"]
  year INT [not null, note: "Year"]
  created_at TIMESTAMP [not null, note: "From balancesheet.created event"]
  updated_at TIMESTAMP [not null, note: "From latest balancesheet.* event"]
  indexes {
    (tenant_id, user_id, year) [unique, name: "uq_balancesheet_per_user_year_tenant"]
  }
}

Table "audit_log_projections" {
  id uuid [pk, note: "Log ID"]
  tenant_id uuid [ref: > tenant_projections.id, not null, note: "References tenant"]
  user_id uuid [note: "User who performed action"]
  action log_action_enum [not null, note: "Action type"]
  resource_type VARCHAR(100) [note: "Resource type"]
  resource_id TEXT [note: "Resource ID"]
  level log_level_enum [not null, default: "info", note: "Log level"]
  message TEXT [not null, note: "Log message"]
  metadata JSONB [note: "Additional metadata"]
  ip_address VARCHAR(45) [note: "Client IP"]
  user_agent TEXT [note: "User agent"]
  created_at TIMESTAMP [not null, note: "From audit event"]
  indexes {
    (tenant_id, created_at) [name: "idx_audit_log_tenant_created"]
    (tenant_id, user_id, created_at) [name: "idx_audit_log_tenant_user_created"]
    (resource_type, resource_id) [name: "idx_audit_log_resource"]
    (action, created_at) [name: "idx_audit_log_action_created"]
  }
}

Table "request_log_projections" {
  id uuid [pk, note: "Log ID"]
  tenant_id uuid [note: "References tenant (nullable for unauthenticated)"]
  user_id uuid [note: "User who made request (nullable)"]
  method VARCHAR(10) [not null, note: "HTTP method"]
  path VARCHAR(500) [not null, note: "Request path"]
  query_params JSONB [note: "Query parameters"]
  request_headers JSONB [note: "Request headers"]
  request_body JSONB [note: "Request body"]
  response_status INT [note: "Response status code"]
  response_body JSONB [note: "Response body"]
  response_time_ms INT [note: "Processing time in ms"]
  ip_address VARCHAR(45) [note: "Client IP"]
  user_agent TEXT [note: "User agent"]
  error_message TEXT [note: "Error message"]
  error_stack TEXT [note: "Error stack trace"]
  created_at TIMESTAMP [not null, note: "From request event"]
  indexes {
    (tenant_id, created_at) [name: "idx_request_log_tenant_created"]
    (tenant_id, user_id, created_at) [name: "idx_request_log_tenant_user_created"]
    (method, path, created_at) [name: "idx_request_log_method_path_created"]
    (response_status, created_at) [name: "idx_request_log_status_created"]
    (created_at) [name: "idx_request_log_created"]
  }
}
