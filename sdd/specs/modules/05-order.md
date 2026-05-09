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
| unit_price_snapshot | NUMERIC(10,2) | NOT NULL — price at checkout time |
| discount_snapshot | INTEGER | NOT NULL, 0–100 — discount at checkout time |

## DB Schema

```sql
CREATE TABLE orders (
    id          SERIAL PRIMARY KEY,
    cart_id     INTEGER       NOT NULL REFERENCES carts(id),
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

## Implementation Checklist

- [ ] Migration SQL — create `orders` and `order_items` tables + indexes
- [ ] `internal/order/domain/order.go` — Order and OrderItem structs
- [ ] `internal/order/shared/repository.go` — `FindByBuyer`, `FindById`, `UpdateStatus`, `Cancel` (with stock restore)
- [ ] `internal/order/features/list_orders/` — handler, usecase, ports, dtos
- [ ] `internal/order/features/get_order/` — handler, usecase, ports, dtos
- [ ] `internal/order/features/cancel_order/` — handler, usecase, ports, dtos
- [ ] `internal/order/features/admin_update_status/` — handler, usecase, ports, dtos
- [ ] `internal/order/routes.go` — map all routes
- [ ] Wire in `internal/app/setup/setup.go` and `internal/app/route/routes.go`
- [ ] Manual test: checkout → list orders → get order → cancel (check stock restored) → admin advance status
