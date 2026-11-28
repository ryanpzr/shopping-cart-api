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
   └────→ cancelled   (only from pending)
```

- Forward transitions only, **one step at a time**: `pending → paid → shipped → delivered`. Skipping a step (`pending → shipped`), going backwards and repeating the current status are all conflicts (`ErrConflict`, 409)
- `cancelled` is reachable only from `pending`; `cancelled` and `delivered` are terminal states — no further transitions allowed
- The rules live in `internal/order/domain` (`CanTransition`); a full from/to matrix is tested

## Endpoints

| Rule | Endpoint | Method | Auth |
|---|---|---|---|
| BR-ORDER-001 | `/orders/me` | GET | Authenticated |
| BR-ORDER-002 | `/orders/:id` | GET | Owner or Admin |
| BR-ORDER-003 | `/orders/:id/cancel` | PATCH | Owner |
| BR-ADMIN-ORDER-001 | `/admin/orders/:id/status` | PATCH | Admin |

Every response of an order has `id`, `cart_id`, `status`, `total_price`, `created_at`, `updated_at` and
`items[]` with `product_id`, `title`, `quantity`, `unit_price`, `discount_percentage`, `sale_price` and
`subtotal` (the price and discount as they were at checkout; the title comes from the product).

**BR-ORDER-001 — List own orders**
- Returns orders where `buyer_id = current_user_id`, ordered by `created_at DESC` (then `id DESC`)
- Includes order items in response (loaded with one query for the whole page)
- Pagination: `limit` (default 20, capped at 100) and `offset` (default 0); response has `data` and `meta { total, limit, offset }`; an invalid query value → `ErrBadRequest` (400). `total` is counted by a separate query, so a checkout of the same buyer landing between the two can leave it one behind the page; a window count (`COUNT(*) OVER()`) was rejected because it loses the total when `offset` is past the last page

**BR-ORDER-002 — Get order by ID**
- `ErrNotFound` (404) if the order does not exist (for everyone, before any ownership check)
- Client can only fetch own order (`buyer_id = current_user_id`) → `ErrForbidden` (403)
- Admin can fetch any order
- Invalid id → `ErrBadRequest` (400)

> A `403` for someone else's order reveals that the id exists; the spec chose it over a `404`.

**BR-ORDER-003 — Cancel order**
- Must be owner → `ErrForbidden` (403); checked before the status, so the answer does not reveal it
- Only cancellable when `status = pending` → `ErrConflict` (409) otherwise
- For each order item: restore `product.quantity += item.quantity`
- Sets `status = cancelled`, `updated_at = NOW()`
- One transaction: the order row is locked (`SELECT … FOR UPDATE`), then the products (by id, like checkout), so two cancellations of the same order never give the stock back twice
- Emits `order_cancelled` activity log event (module 06; `TODO(module-06)` in the usecase)

**BR-ADMIN-ORDER-001 — Update order status**
- Required body: `status`, one of `pending`, `paid`, `shipped`, `delivered`, `cancelled` (exact, lower case) → `ErrBadRequest` (400) when missing or unknown
- Only the transitions above → `ErrConflict` (409) for any other change, including a `cancelled` or `delivered` order
- `cancelled` (from `pending`) **restores the stock** exactly like the buyer's cancellation, in the same transaction; the other statuses do not touch the stock
- The order row is locked, so concurrent changes run one after the other: of several identical requests, one succeeds and the rest get `409`
- `ErrNotFound` (404) if the order does not exist

## Note on Order Creation

Orders are not created via a direct POST /orders endpoint. They are created by the checkout flow in `BR-CART-005`. The Order module only exposes read and status-update operations.

Since Phase 2 the tables exist (`004_create_orders.sql`) and checkout writes them through
`internal/order/shared.Repository.Create`, in the checkout transaction. `unit_price_snapshot` is the
product price **before** the discount and `discount_snapshot` the discount percentage, both taken at
checkout; the line price is `unit_price_snapshot` with the discount applied (rounded half up to cents),
and `total_price` is the sum of the lines. Amounts use `pkg/money` (integer cents).

## Implementation Notes

- **Layers:** the order repository only talks to the database (domain errors, no business rule). Usecases decide who may see or change an order and which transitions are allowed.
- **Unit of work:** `internal/order/shared.UnitOfWork` gives usecases the order and product repositories in one transaction (cancel and admin status).
- **Cancellation** is shared by the buyer's cancel and the admin status change in `internal/order/cancellation`, so giving the stock back lives in one place. The amount to give back is a pure domain function (`RestockQuantities`).
- **Locking order:** checkout locks cart → products; cancellation locks order → products (by id); delete locks the product. Products are always locked in id order, so none of these can deadlock each other.

## Cross-Module Notes

**Product delete guard (BR-PROD-006).** The delete usecase asks an `OrderChecker` whether the product
is in orders with `domain.ActiveOrderStatuses` (`pending`, `paid`, `shipped`).
`HasOrdersWithStatus` in the order repository (`internal/order/shared`) joins `order_items`
(by `product_id`) with `orders` and filters `orders.status = ANY($2)`. Keep the table and column names
above (`order_items.product_id`, `order_items.order_id`, `orders.status`).

**Delete is transactional (Phase 3).** The delete usecase runs in one transaction: it locks the product
row (`SELECT … FOR UPDATE`), checks ownership, checks the orders and deletes. Checkout also locks the
product row (after the cart), so no order can be placed for the product between the check and the delete: a
concurrent checkout and delete end as (checkout `201`, delete `409`) or (checkout `409`, delete `200`),
never a `500`. The unit of work lives in the `delete_product` feature, because the order module already
depends on the product repository.

**Deleting products with past orders — decision (Phase 3): keep `409`.** `order_items.product_id`
references `products(id)` with the default `NO ACTION`, so a product that appears in any order (even a
`delivered` or `cancelled` one) cannot be deleted: the product repository maps the foreign key violation
to `ErrProductInUse` and the delete answers `409`. The seller deactivates the product instead
(`PATCH /products/:id/status`). Soft delete and `ON DELETE SET NULL` were considered and left out: they
need a schema change and filters in every product query, for no current need.

## Implementation Checklist

- [x] Migration SQL — create `orders` and `order_items` tables + indexes (Phase 2, `004_create_orders.sql`)
- [x] `internal/order/domain/` — Order and OrderItem structs (Phase 2); transitions, status validation, `RestockQuantities` (Phase 3)
- [x] `internal/order/shared/repository.go` — `Create` (Phase 2); `FindByBuyer`, `FindByID`, `LockByID`, `UpdateStatus`, `HasOrdersWithStatus` (Phase 3)
- [x] `internal/order/shared/unit_of_work.go` — order and product repositories in one transaction
- [x] `internal/order/cancellation/` — cancellation with stock restore, shared by the buyer and the admin
- [x] `internal/order/features/list_orders/`, `get_order/`, `cancel_order/`, `admin_update_status/` — handler, usecase, ports, dtos, tests
- [x] `internal/order/routes.go` — map all routes; wired in `setup.go` and `routes.go`
- [x] Product delete in one transaction with the product row locked
- [x] Manual test: checkout → list orders → get order → cancel (check stock restored) → admin advance status; concurrent cancellations, status changes and checkout × delete
- [ ] Activity log event `order_cancelled` — module 06, through the outbox
