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
| discount_percentage | INTEGER | NOT NULL, DEFAULT 0, 0–100 |
| quantity | INTEGER | NOT NULL, DEFAULT 0, >= 0 |
| status | VARCHAR(10) | `active` / `inactive`, DEFAULT `active` |
| category | VARCHAR(100) | nullable |
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
- Query params (all optional): `category`, `min_price`, `max_price`, `search` (matches title, case-insensitive)
- Response includes `final_price = price * (1 - discount_percentage / 100)` as computed field

**BR-PROD-002 — Get product by ID**
- `ErrNotFound` (404) if not found or `status = inactive`

**BR-PROD-003 — Create product**
- Required: `title`, `price` (> 0), `quantity` (>= 0)
- Optional: `photo`, `description`, `category`, `discount_percentage` (0–100, defaults to 0)
- `seller_id` is injected from JWT — not accepted from request body
- Default `status`: `active`
- Emits `product_created` activity log event

**BR-PROD-004 — Update product (partial)**
- Only owner (`seller_id = current_user_id`) can edit → `ErrForbidden` (403)
- `ErrNotFound` (404) if product does not exist
- `price` must be > 0 if provided → `ErrBadRequest` (400)
- `discount_percentage` must be 0–100 if provided → `ErrBadRequest` (400)
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
- Owner can delete own product; admin can delete any product
- Non-owner client → `ErrForbidden` (403)
- Emits `product_deleted` activity log event

## Implementation Checklist

- [ ] Migration SQL — create `products` table + indexes
- [ ] `internal/product/domain/product.go` — update existing struct to match full entity
- [ ] `internal/product/shared/repository.go` — `FindAll` (with filters), `FindById`, `Create`, `Update`, `Delete`, `HasActiveOrders`
- [ ] `internal/product/features/list_products/` — handler, usecase, ports, dtos
- [ ] `internal/product/features/get_product/` — handler, usecase, ports, dtos
- [ ] `internal/product/features/create_product/` — handler, usecase, ports, dtos
- [ ] `internal/product/features/update_product/` — handler, usecase, ports, dtos
- [ ] `internal/product/features/toggle_status/` — handler, usecase, ports, dtos
- [ ] `internal/product/features/delete_product/` — handler, usecase, ports, dtos
- [ ] `internal/product/routes.go` — map all routes
- [ ] Wire in `internal/app/setup/setup.go` and `internal/app/route/routes.go`
- [ ] Manual test: create → list → filter → update → toggle → delete (check conflict guard)
