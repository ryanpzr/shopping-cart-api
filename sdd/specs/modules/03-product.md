# Module: Product

## Dependencies
- `01-auth` — JWT middleware
- `02-user` — `users` table must exist (seller_id FK)

## Entities

**Product**

| Field | Type | Notes |
|---|---|---|
| id | SERIAL PK | |
| seller_id | INTEGER FK | REFERENCES users(id) NOT NULL |
| photo | TEXT | nullable |
| title | VARCHAR(255) | NOT NULL |
| description | TEXT | nullable |
| price | NUMERIC(10,2) | NOT NULL, > 0 |
| discount_percentage | INTEGER | NOT NULL, DEFAULT 0, 0–80 |
| quantity | INTEGER | NOT NULL, DEFAULT 0, >= 0 |
| status | VARCHAR(10) | `active` / `inactive`, DEFAULT `active` |
| category | VARCHAR(100) | nullable; one of the catalog categories (validated in the domain) |
| created_at | TIMESTAMPTZ | DEFAULT NOW() |
| updated_at | TIMESTAMPTZ | DEFAULT NOW() |

## DB Schema

```sql
CREATE TABLE products (
    id                   SERIAL PRIMARY KEY,
    seller_id            INTEGER       NOT NULL REFERENCES users(id),
    photo                TEXT,
    title                VARCHAR(255)  NOT NULL,
    description          TEXT,
    price                NUMERIC(10,2) NOT NULL CHECK (price > 0),
    discount_percentage  INTEGER       NOT NULL DEFAULT 0 CHECK (discount_percentage >= 0 AND discount_percentage <= 100),
    quantity             INTEGER       NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    status               VARCHAR(10)   NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    category             VARCHAR(100),
    created_at           TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_seller_id ON products(seller_id);
CREATE INDEX idx_products_status    ON products(status);
CREATE INDEX idx_products_category  ON products(category);
```

## Business Limits

| Rule | Value | Reason |
|---|---|---|
| `discount_percentage` | 0–80 | The marketplace keeps a minimum margin. Enforced by the domain; the schema only keeps the structural `0–100` check, so changing the cap needs no migration |
| `price` | > 0 and <= 1,000,000 | Guards against input errors and abuse |
| `quantity` | >= 0 | 0 means sold out: checkout can bring stock to 0 and sellers can set it; buying is blocked by the cart stock check (`ErrConflict`) |

## Categories

Fixed catalog taxonomy (`domain.Categories`), matched exactly and case-sensitively:
`Electronics`, `Clothing`, `Books`, `Home & Garden`, `Sports`, `Beauty`, `Toys`,
`Food & Beverages`, `Health`, `Automotive`, `Music & Movies`, `Office Supplies`.

The list lives in the domain, not in a database `CHECK`, so adding a category is a code change
without a migration.

## Endpoints

| Rule | Endpoint | Method | Auth |
|---|---|---|---|
| BR-PROD-001 | `/products` | GET | Public |
| BR-PROD-002 | `/products/:id` | GET | Public |
| BR-PROD-003 | `/products` | POST | Client |
| BR-PROD-004 | `/products/:id` | PUT | Client (owner) |
| BR-PROD-005 | `/products/:id/status` | PATCH | Client (owner) |
| BR-PROD-006 | `/products/:id` | DELETE | Client (owner) or Admin |

**BR-PROD-001 — List products**
- Returns only `status = active` products
- Query params (all optional): `category`, `min_price`, `max_price`, `search` (matches title, case-insensitive; `%` and `_` match literally)
- Pagination: `limit` (default 20, capped at 100) and `offset` (default 0); response has `data` and `meta { total, limit, offset }`
- `min_price`/`max_price` must be >= 0 and `min_price <= max_price`; `category` must be a catalog category → `ErrBadRequest` (400)
- Response includes `final_price = price * (1 - discount_percentage / 100)`, rounded to 2 decimals, as computed field (`domain.Product.FinalPrice`)

**BR-PROD-002 — Get product by ID**
- `ErrNotFound` (404) if not found or `status = inactive`

**BR-PROD-003 — Create product**
- Required: `title` (max 255 chars), `price` (> 0, max 1,000,000), `quantity` (>= 0)
- `category` optional; if provided, must be one of the catalog categories (see Categories) → `ErrBadRequest` (400)
- Optional: `photo`, `description`, `category`, `discount_percentage` (0–80, defaults to 0)
- `seller_id` is injected from JWT — not accepted from request body
- Default `status`: `active`
- Emits `product_created` activity log event

**BR-PROD-004 — Update product (partial)**
- Only owner (`seller_id = current_user_id`) can edit → `ErrForbidden` (403)
- `ErrNotFound` (404) if product does not exist
- At least one field must be provided → `ErrBadRequest` (400)
- `price` must be > 0 and at most 1,000,000 if provided → `ErrBadRequest` (400)
- `discount_percentage` must be 0–80 if provided → `ErrBadRequest` (400)
- `quantity` must be >= 0 if provided → `ErrBadRequest` (400)
- `seller_id` is ignored even if sent in body
- Emits `product_updated` activity log event

**BR-PROD-005 — Toggle visibility**
- Only owner can toggle → `ErrForbidden` (403)
- `active` → `inactive` and vice versa
- `inactive` products are hidden from BR-PROD-001 listing

**BR-PROD-006 — Delete product**
- Hard delete from database
- Cannot delete if product has orders with `status IN ('pending', 'paid', 'shipped')` → `ErrConflict` (409)
- Any other order item that still references the product (e.g. a delivered or cancelled order) also blocks the delete → `ErrConflict` (409); the long-term policy is decided in `05-order.md`
- Removes the product from carts (`cart_items` cascade)
- Owner can delete own product; admin can delete any product
- Non-owner client → `ErrForbidden` (403)
- Emits `product_deleted` activity log event

## Implementation Checklist

- [x] Migration SQL — `db/migrations/002_create_products.sql` (`products` table + indexes)
- [x] `internal/product/domain/product.go` — full entity, `FinalPrice()`; field rules in `domain/rules.go`
- [x] `internal/product/shared/repository.go` — `FindAll` (with filters, including `Status`), `FindByID`, `Create`, `Update` (whole entity), `UpdateStatus`, `Delete`; takes `context.Context` and a `DBTX` (`*sql.DB` or `*sql.Tx`); returns `domain.ErrProductNotFound` or wrapped errors, never `apperrors`
- [x] Delete guard through the `OrderChecker` port, implemented by the order repository (`internal/order/shared`, `HasOrdersWithStatus`); a `NoOrders` stub stood in until the order tables arrived in Phase 2. A product still referenced by any order item (e.g. delivered/cancelled) answers `409` (`domain.ErrProductInUse`)
- [x] `internal/product/features/list_products/` — handler, usecase, ports, dtos, tests
- [x] `internal/product/features/get_product/` — handler, usecase, ports, dtos, tests
- [x] `internal/product/features/create_product/` — handler, usecase, ports, dtos, tests
- [x] `internal/product/features/update_product/` — handler, usecase, ports, dtos, tests
- [x] `internal/product/features/toggle_status/` — handler, usecase, ports, dtos, tests
- [x] `internal/product/features/delete_product/` — handler, usecase, ports, dtos, tests
- [x] `internal/product/routes.go` — map all routes
- [x] Wire in `internal/app/setup/setup.go` and `internal/app/route/routes.go`
- [x] Manual test: create → list → filter → update → toggle → delete (check conflict guard)
- [ ] Activity log events (`product_created`, `product_updated`, `product_deleted`) — module 06, via the outbox
