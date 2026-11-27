# Module: Cart

## Dependencies
- `01-auth` — JWT middleware
- `02-user` — `users` table (buyer_id FK)
- `03-product` — `products` table (product_id FK, stock validation, self-purchase check)

## Entities

**Cart**

| Field | Type | Notes |
|---|---|---|
| id | SERIAL PK | |
| buyer_id | INTEGER FK | REFERENCES users(id) NOT NULL |
| status | VARCHAR(15) | `active` / `checked_out`, DEFAULT `active` |
| created_at | TIMESTAMPTZ | DEFAULT NOW() |
| updated_at | TIMESTAMPTZ | DEFAULT NOW() |

**CartItem**

| Field | Type | Notes |
|---|---|---|
| id | SERIAL PK | |
| cart_id | INTEGER FK | REFERENCES carts(id) NOT NULL |
| product_id | INTEGER FK | REFERENCES products(id) NOT NULL |
| quantity | INTEGER | NOT NULL, > 0 |
| price_snapshot | NUMERIC(10,2) | NOT NULL — price locked at time of add |
| created_at | TIMESTAMPTZ | DEFAULT NOW() |

> `(cart_id, product_id)` must be UNIQUE — duplicate products are handled by summing quantities.
> A buyer has at most one `active` cart (partial unique index), even under concurrent requests.
> Deleting a product removes it from carts (`ON DELETE CASCADE`); purchase history lives in `order_items`.

## DB Schema

```sql
CREATE TABLE carts (
    id         SERIAL PRIMARY KEY,
    buyer_id   INTEGER     NOT NULL REFERENCES users(id),
    status     VARCHAR(15) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'checked_out')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_carts_buyer_id ON carts(buyer_id);
CREATE UNIQUE INDEX ux_carts_one_active_per_buyer ON carts(buyer_id) WHERE status = 'active';

CREATE TABLE cart_items (
    id             SERIAL PRIMARY KEY,
    cart_id        INTEGER       NOT NULL REFERENCES carts(id),
    product_id     INTEGER       NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity       INTEGER       NOT NULL CHECK (quantity > 0),
    price_snapshot NUMERIC(10,2) NOT NULL,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    UNIQUE (cart_id, product_id)
);
```

## Money

Cart and order amounts use `pkg/money.Money` (integer cents): sums and multiplications are exact. It
scans `NUMERIC(10,2)` without going through `float64` and marshals to JSON as a decimal number
(`89.90`), so the API contract uses plain numbers. The product module still stores `float64`; its price
enters money once, through `domain.Product.SalePrice()` (discount applied on cents, rounded half up).

## Transactions

Mutations and checkout run through `shared.UnitOfWork`, which gives the usecase cart, product and order
repositories bound to one transaction (`pkg/database.WithinTx`). Usecases depend on the port, so tests
use a fake unit of work.

## Endpoints

| Rule | Endpoint | Method | Auth |
|---|---|---|---|
| BR-CART-001 | `/carts/me` | GET | Client |
| BR-CART-002 | `/carts/me/items` | POST | Client |
| BR-CART-003 | `/carts/me/items/:itemId` | DELETE | Client |
| BR-CART-004 | `/carts/me/items/:itemId` | PATCH | Client |
| BR-CART-005 | `/carts/me/checkout` | POST | Client |

**BR-CART-001 — Get active cart**
- Returns or creates the single `active` cart for `buyer_id = current_user_id`
- Response shape:
```json
{
  "id": 1,
  "status": "active",
  "items": [
    {
      "id": 1,
      "product_id": 5,
      "title": "...",
      "quantity": 2,
      "price_snapshot": 89.90
    }
  ],
  "subtotal": 179.80,
  "discount_total": 0.00,
  "total": 179.80
}
```
> `subtotal` = sum of `price_snapshot * quantity` for all items.
> `discount_total` = 0 here because discount is already baked into `price_snapshot` at add time.

**BR-CART-002 — Add item to cart**
- Required body: `product_id`, `quantity` (> 0)
- Product must exist and `status = active` → `ErrNotFound` (404)
- `product.seller_id` must ≠ `current_user_id` → `ErrForbidden` (403)
- `product.quantity` >= requested quantity → `ErrConflict` (409) if insufficient
- If `(cart_id, product_id)` row already exists: add to existing quantity, re-validate combined total against stock, and refresh `price_snapshot` to the current sale price
- `price_snapshot = product.price * (1 - product.discount_percentage / 100)`, rounded half up to cents (`domain.Product.SalePrice()`)
- Missing `product_id` / `quantity`, or `quantity <= 0` → `ErrBadRequest` (400)
- Runs in a transaction with the cart row locked; returns the updated cart

**BR-CART-003 — Remove item from cart**
- `ErrNotFound` (404) if `itemId` does not belong to `current_user`'s active cart

**BR-CART-004 — Update item quantity**
- Required body: `quantity`
- `quantity` must be > 0 → `ErrBadRequest` (400) if 0 or negative (use DELETE to remove)
- Re-validates new quantity against current `product.quantity` → `ErrConflict` (409) if insufficient
- `ErrNotFound` (404) if `itemId` does not belong to `current_user`'s active cart
- Keeps the price snapshot; returns the updated cart

**BR-CART-005 — Checkout**
- One transaction: lock the cart (`FOR UPDATE`) and its products (`FOR UPDATE`, ordered by id, so concurrent checkouts cannot deadlock). Concurrent checkouts wait on the locks and **cannot oversell**; a second checkout of the same cart finds no active cart → `409`
- Empty cart (or no active cart) → `ErrConflict` (409)
- Re-validates every item: product exists and `status = active` → `409`; `product.seller_id ≠ current_user_id` → `ErrForbidden` (403); `product.quantity >= item.quantity` → `409`
- **Prices:** if a product's current sale price differs from the item's `price_snapshot`, the snapshots are refreshed and committed, no order is placed, and the response is `409` ("prices changed"). The buyer never pays a price they did not see; checking out again uses the new prices
- For each item: decrement `product.quantity`
- Create `Order` (`status = pending`, `total_price = sum of price_snapshot * quantity`) + `OrderItem` records (`unit_price_snapshot` = price before the discount, `discount_snapshot` = discount; both equal to the cart snapshot at this point)
- Mark `cart.status = checked_out`; the next `GET /carts/me` opens a new cart
- Returns the created order summary (`201`)
- Emits `cart_checkout` and `order_placed` activity log events (module 06; `TODO(module-06)` in the usecase)

## Implementation Checklist

- [x] Migration SQL — `003_create_carts.sql` (`carts`, `cart_items`, one-active-cart index) and `004_create_orders.sql` (schema from `05-order.md`, needed by checkout)
- [x] `pkg/money` (integer cents) and `pkg/database` (`DBTX`, `WithinTx`)
- [x] `internal/cart/domain/` — `Cart`, `Item`, totals, quantity and stock rules, sentinel errors
- [x] `internal/cart/shared/repository.go` — `FindByBuyer`, `LockByBuyer`, `Create`, `UpdateStatus`, `AddItem`, `UpdateItem`, `RemoveItem` (domain errors, no business rules in SQL)
- [x] `internal/cart/shared/unit_of_work.go` — cart, product and order repositories in one transaction
- [x] `internal/order/domain` and `internal/order/shared` — order entity and `Create`
- [x] `internal/cart/features/get_cart/`, `add_item_to_cart/`, `change_item_quantity/`, `remove_item/`, `checkout/` — handler, usecase, ports, dtos, tests
- [x] `internal/cart/routes.go` — map all routes; wired in `setup.go` and `routes.go`
- [x] Product delete guard uses the order repository (`HasOrdersWithStatus`, moved to `internal/order/shared`)
- [x] Manual test: get cart → add item → add same item again (quantity sums) → update quantity → remove item → price change → checkout; concurrent checkouts for the last units
- [ ] Activity log events (`cart_checkout`, `order_placed`) — module 06, through the outbox
