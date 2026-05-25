# Module: Activity Log

## Dependencies
- `01-auth` — JWT middleware
- `02-user` — `users` table (user_id FK); BR-ADMIN-006 endpoint is surfaced via the user module routes
- All other modules emit events into this table

## Entity

**ActivityLog**

| Field | Type | Notes |
|---|---|---|
| id | SERIAL PK | |
| user_id | INTEGER FK | REFERENCES users(id) NOT NULL |
| event_type | VARCHAR(50) | NOT NULL |
| metadata | JSONB | nullable — contextual data per event |
| created_at | TIMESTAMPTZ | DEFAULT NOW() |

## DB Schema

```sql
CREATE TABLE activity_logs (
    id         SERIAL PRIMARY KEY,
    user_id    INTEGER     NOT NULL REFERENCES users(id),
    event_type VARCHAR(50) NOT NULL,
    metadata   JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_activity_logs_user_id    ON activity_logs(user_id);
CREATE INDEX idx_activity_logs_event_type ON activity_logs(event_type);
CREATE INDEX idx_activity_logs_created_at ON activity_logs(created_at DESC);
```

## Event Catalog

| Event Type | Emitted by | Metadata fields |
|---|---|---|
| `login` | Auth — BR-AUTH-002 | `ip` (optional) |
| `product_created` | Product — BR-PROD-003 | `product_id`, `title` |
| `product_updated` | Product — BR-PROD-004 | `product_id`, `changed_fields` |
| `product_deleted` | Product — BR-PROD-006 | `product_id`, `title` |
| `cart_checkout` | Cart — BR-CART-005 | `order_id`, `total_price` |
| `order_placed` | Cart — BR-CART-005 | `order_id` |
| `order_cancelled` | Order — BR-ORDER-003 | `order_id` |

## Endpoint

The read endpoint is defined in `02-user` as BR-ADMIN-006 and registered under the `/admin/users/:id/activity` route.

| Rule | Endpoint | Method | Auth |
|---|---|---|---|
| BR-ADMIN-006 | `/admin/users/:id/activity` | GET | Admin |

- Returns paginated activity log for the given `user_id`
- Supports optional filter: `event_type`
- Ordered by `created_at DESC`

## How to Emit Events

Each module that emits events should call a shared logger, not import this module directly. The recommended approach is a `pkg/activitylog` package with a single function:

```go
// pkg/activitylog/logger.go
func Log(db *sql.DB, userID int, eventType string, metadata map[string]any) error
```

Each usecase that needs to log an event calls this function after the main operation succeeds. If logging fails, it should not block the main operation (log the error, continue).

## Implementation Checklist

- [ ] Migration SQL — create `activity_logs` table + indexes
- [ ] `pkg/activitylog/logger.go` — shared `Log` function used by all modules
- [ ] `internal/activitylog/domain/log.go` — ActivityLog struct
- [ ] `internal/activitylog/shared/repository.go` — `FindByUser` (with pagination + optional event_type filter)
- [ ] `internal/activitylog/features/list_activity/` — handler, usecase, ports, dtos
- [ ] Substituir o stub em `internal/user/features/admin_activity_log/usecase.go`:
      trocar `usecase struct{}` por `usecase struct{ repo Repository }` onde
      `Repository = activitylog_shared.Repository`. O handler e as rotas NÃO mudam.
- [ ] Add `pkg/activitylog.Log(...)` calls nos seguintes pontos (TODOs já inseridos no código):
  - **Auth login** — `internal/auth/features/login/usecase.go` (evento `login`)
  - **Product create** — `internal/product/features/create_product/usecase.go` (evento `product_created`, metadata: `product_id`, `title`)
  - **Product update** — `internal/product/features/update_product/usecase.go` (evento `product_updated`, metadata: `product_id`, `changed_fields`)
  - **Product delete** — `internal/product/features/delete_product/usecase.go` (evento `product_deleted`, metadata: `product_id`, `title`)
  - **Cart checkout** — `internal/cart/...` (eventos `cart_checkout` e `order_placed`)
  - **Order cancel** — `internal/order/features/cancel_order/usecase.go` (evento `order_cancelled`)
- [ ] Substituir stub `internal/user/features/admin_activity_log/usecase.go` pelo repositório real de activity log
- [ ] Manual test: perform each logged action → call GET /admin/users/:id/activity → verify events appear
