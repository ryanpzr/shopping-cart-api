# Module: Order

## Dependencies
- `01-auth` — JWT middleware
- `02-user` — `users` table (buyer_id FK)
- `03-product` — `products` table (product_id FK in order_items)
- `04-cart` — Order is created by the checkout flow; `carts` table referenced

## Entities

**Order**

| Field | Type | Notes |
|---|---|---|
| id | SERIAL PK | |
| cart_id | INTEGER FK | REFERENCES carts(id) NOT NULL |
| buyer_id | INTEGER FK | REFERENCES users(id) NOT NULL |
| total_price | NUMERIC(10,2) | NOT NULL |
| status | VARCHAR(15) | `pending` / `paid` / `shipped` / `delivered` / `cancelled` |
| created_at | TIMESTAMPTZ | DEFAULT NOW() |
| updated_at | TIMESTAMPTZ | DEFAULT NOW() |

**OrderItem**

| Field | Type | Notes |
|---|---|---|
| id | SERIAL PK | |
| order_id | INTEGER FK | REFERENCES orders(id) NOT NULL |
| product_id | INTEGER FK | REFERENCES products(id) NOT NULL |
| quantity | INTEGER | NOT NULL, > 0 |
| unit_price_snapshot | NUMERIC(10,2) | NOT NULL — price before the discount, at checkout time |
| discount_snapshot | INTEGER | NOT NULL, 0–100 — discount percentage at checkout time |

## DB Schema

```sql
CREATE TABLE orders (
    id          SERIAL PRIMARY KEY,
    cart_id     INTEGER       NOT NULL UNIQUE REFERENCES carts(id),  -- one order per cart
    buyer_id    INTEGER       NOT NULL REFERENCES users(id),
    total_price NUMERIC(10,2) NOT NULL,
    status      VARCHAR(15)   NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'paid', 'shipped', 'delivered', 'cancelled')),
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_buyer_id ON orders(buyer_id);
CREATE INDEX idx_orders_status   ON orders(status);

CREATE TABLE order_items (
    id                   SERIAL PRIMARY KEY,
    order_id             INTEGER       NOT NULL REFERENCES orders(id),
    product_id           INTEGER       NOT NULL REFERENCES products(id),
    quantity             INTEGER       NOT NULL CHECK (quantity > 0),
    unit_price_snapshot  NUMERIC(10,2) NOT NULL,
    discount_snapshot    INTEGER       NOT NULL DEFAULT 0
);
```

## Status Transitions

```
pending → paid → shipped → delivered
    └──────────────────────→ cancelled  (only from pending)
```

- Forward transitions only — cannot go backwards
- `cancelled` is a terminal state — no further transitions allowed

## Endpoints

| Rule | Endpoint | Method | Auth |
|---|---|---|---|
| BR-ORDER-001 | `/orders/me` | GET | Client |
| BR-ORDER-002 | `/orders/:id` | GET | Client (owner) or Admin |
| BR-ORDER-003 | `/orders/:id/cancel` | PATCH | Client (owner) |
| BR-ADMIN-ORDER-001 | `/admin/orders/:id/status` | PATCH | Admin |

**BR-ORDER-001 — List own orders**
- Returns orders where `buyer_id = current_user_id`, ordered by `created_at DESC`
- Includes order items in response

**BR-ORDER-002 — Get order by ID**
- Client can only fetch own order (`buyer_id = current_user_id`) → `ErrForbidden` (403)
- Admin can fetch any order
- `ErrNotFound` (404) if order does not exist

**BR-ORDER-003 — Cancel order**
- Must be owner → `ErrForbidden` (403)
- Only cancellable when `status = pending` → `ErrConflict` (409) otherwise
- For each order item: restore `product.quantity += item.quantity`
- Sets `status = cancelled`, `updated_at = NOW()`
- Emits `order_cancelled` activity log event

**BR-ADMIN-ORDER-001 — Update order status**
- Required body: `status`
- Valid transitions only (see diagram above) → `ErrConflict` (409) for invalid transitions
- Cannot modify a `cancelled` order → `ErrConflict` (409)

## Note on Order Creation

Orders are not created via a direct POST /orders endpoint. They are created by the checkout flow in `BR-CART-005`. The Order module only exposes read and status-update operations.

Since Phase 2 the tables exist (`004_create_orders.sql`) and checkout writes them through
`internal/order/shared.Repository.Create`, in the checkout transaction. `unit_price_snapshot` is the
product price **before** the discount and `discount_snapshot` the discount percentage, both taken at
checkout; the line price is `unit_price_snapshot` with the discount applied (rounded half up to cents),
and `total_price` is the sum of the lines. Amounts use `pkg/money` (integer cents).

## Cross-Module Notes

**Product delete guard (BR-PROD-006).** The delete usecase asks an `OrderChecker` whether the product
is in orders with `domain.ActiveOrderStatuses` (`pending`, `paid`, `shipped`).
`HasOrdersWithStatus` in the order repository (`internal/order/shared`) joins `order_items`
(by `product_id`) with `orders` and filters `orders.status = ANY($2)`. It is wired in `setup.go` since
Phase 2, when checkout started creating orders. Keep the table and column names above
(`order_items.product_id`, `order_items.order_id`, `orders.status`).

**Open question — delete is check-then-act.** The usecase checks for active orders and then deletes,
in two statements, while checkout can now create orders concurrently. Run both inside one transaction
(the product repository already accepts a `*sql.Tx` through `DBTX`) and lock the product row.

**Open question — deleting products with past orders.** `order_items.product_id` references
`products(id)` with the default `NO ACTION`, so a hard delete of a product that appears in a
`delivered` or `cancelled` order fails on the foreign key. Since Phase 2 the product repository maps
that violation to `ErrProductInUse` and the delete answers `409` instead of `500`; this also covers an
order placed between the active-order check and the delete. Decide in this module whether that stays
the policy or becomes a soft delete / `ON DELETE SET NULL` on a nullable `product_id`.

## Implementation Checklist

- [x] Migration SQL — create `orders` and `order_items` tables + indexes (Phase 2, `004_create_orders.sql`)
- [x] `internal/order/domain/order.go` — Order and OrderItem structs (Phase 2)
- [x] `internal/order/shared/repository.go` — `Create` (Phase 2, used by checkout)
- [ ] `internal/order/shared/repository.go` — `FindByBuyer`, `FindById`, `UpdateStatus`, `Cancel` (with stock restore)
- [ ] `internal/order/features/list_orders/` — handler, usecase, ports, dtos
- [ ] `internal/order/features/get_order/` — handler, usecase, ports, dtos
- [ ] `internal/order/features/cancel_order/` — handler, usecase, ports, dtos
- [ ] `internal/order/features/admin_update_status/` — handler, usecase, ports, dtos
- [ ] `internal/order/routes.go` — map all routes
- [ ] Wire in `internal/app/setup/setup.go` and `internal/app/route/routes.go`
- [ ] Manual test: checkout → list orders → get order → cancel (check stock restored) → admin advance status
