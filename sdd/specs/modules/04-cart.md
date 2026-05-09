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

CREATE TABLE cart_items (
    id             SERIAL PRIMARY KEY,
    cart_id        INTEGER       NOT NULL REFERENCES carts(id),
    product_id     INTEGER       NOT NULL REFERENCES products(id),
    quantity       INTEGER       NOT NULL CHECK (quantity > 0),
    price_snapshot NUMERIC(10,2) NOT NULL,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    UNIQUE (cart_id, product_id)
);
```

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
- If `(cart_id, product_id)` row already exists: add to existing quantity, re-validate combined total against stock
- `price_snapshot = product.price * (1 - product.discount_percentage / 100)` locked at this moment

**BR-CART-003 — Remove item from cart**
- `ErrNotFound` (404) if `itemId` does not belong to `current_user`'s active cart

**BR-CART-004 — Update item quantity**
- Required body: `quantity`
- `quantity` must be > 0 → `ErrBadRequest` (400) if 0 or negative (use DELETE to remove)
- Re-validates new quantity against current `product.quantity` → `ErrConflict` (409) if insufficient

**BR-CART-005 — Checkout**
- Re-validates every item: `product.status = active`, `product.quantity >= item.quantity`, `product.seller_id ≠ current_user_id`
- For each item: decrement `product.quantity`
- Create `Order` (`status = pending`, `total_price = sum of price_snapshot * quantity`) + `OrderItem` records
- Mark `cart.status = checked_out`
- Returns the created Order summary
- Emits `cart_checkout` activity log event

## Implementation Checklist

- [ ] Migration SQL — create `carts` and `cart_items` tables
- [ ] `internal/cart/domain/cart.go` — Cart and CartItem structs (replace empty interface)
- [ ] `internal/cart/shared/repository.go` — `FindActiveByBuyer`, `CreateCart`, `AddItem`, `RemoveItem`, `UpdateItemQuantity`, `Checkout`
- [ ] `internal/cart/features/get_cart/` — handler, usecase, ports, dtos
- [ ] `internal/cart/features/add_item_to_cart/` — complete the existing stub
- [ ] `internal/cart/features/remove_item/` — handler, usecase, ports, dtos
- [ ] `internal/cart/features/change_item_quantity/` — complete the existing stub
- [ ] `internal/cart/features/checkout/` — handler, usecase, ports, dtos
- [ ] `internal/cart/routes.go` — map all routes
- [ ] Wire in `internal/app/setup/setup.go` and `internal/app/route/routes.go`
- [ ] Manual test: get cart → add item → add same item again (quantity sums) → update quantity → remove item → checkout
