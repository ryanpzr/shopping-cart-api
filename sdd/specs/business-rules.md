# Business Rules

Marketplace-style shopping cart API. Reference this document before implementing any feature.

---

## Entities

| Entity | Fields |
|---|---|
| **User** | id, name, email, password_hash, role (`admin`/`client`), status (`active`/`banned`/`timeout`), timeout_until, created_at, updated_at |
| **Product** | id, seller_id (FK User), photo, title, description, price, discount_percentage (0–100), quantity, status (`active`/`inactive`), category, created_at, updated_at |
| **Cart** | id, buyer_id (FK User), status (`active`/`checked_out`), created_at, updated_at |
| **CartItem** | id, cart_id, product_id, quantity, price_snapshot, created_at |
| **Order** | id, cart_id, buyer_id, total_price, status (`pending`/`paid`/`shipped`/`delivered`/`cancelled`), created_at, updated_at |
| **OrderItem** | id, order_id, product_id, quantity, unit_price_snapshot, discount_snapshot |
| **ActivityLog** | id, user_id, event_type, metadata (JSON), created_at |

---

## Auth

| Rule | Endpoint | Method | Auth |
|---|---|---|---|
| BR-AUTH-001 | `/auth/register` | POST | Public |
| BR-AUTH-002 | `/auth/login` | POST | Public |

**BR-AUTH-001 — Register**
- Required: `name`, `email`, `password`
- `email` must be unique → `ErrConflict`
- `password` stored as bcrypt hash
- Default role: `client`
- Returns: user info + JWT token

**BR-AUTH-002 — Login**
- Required: `email`, `password`
- Returns JWT (payload: `user_id`, `role`, `email`, `exp`)
- `401` if credentials invalid
- `403` if `status = banned`
- `403` if `status = timeout` AND `timeout_until > now`

**BR-AUTH-003 — JWT Middleware**
- All protected routes require `Authorization: Bearer <token>`
- `401` if token missing or invalid
- `403` if token valid but role insufficient

---

## Users

| Rule | Endpoint | Method | Auth |
|---|---|---|---|
| BR-USER-001 | `/users/me` | GET | Client |
| BR-USER-002 | `/users/me` | PUT | Client |
| BR-ADMIN-001 | `/admin/users` | GET | Admin |
| BR-ADMIN-002 | `/admin/users/:id` | GET | Admin |
| BR-ADMIN-003 | `/admin/users/:id/ban` | PATCH | Admin |
| BR-ADMIN-004 | `/admin/users/:id/timeout` | PATCH | Admin |
| BR-ADMIN-005 | `/admin/users/:id/unban` | PATCH | Admin |
| BR-ADMIN-006 | `/admin/users/:id/activity` | GET | Admin |

**BR-USER-002 — Update own profile**
- Cannot change `role` or `status`
- Cannot change `email` to one already in use → `ErrConflict`

**BR-ADMIN-003 — Ban user**
- Sets `status = banned`
- Cannot ban another admin → `ErrForbidden`

**BR-ADMIN-004 — Timeout user**
- Required: `duration_hours` (integer > 0)
- Sets `status = timeout`, `timeout_until = now + duration_hours`
- Cannot timeout another admin → `ErrForbidden`

**BR-ADMIN-005 — Unban / remove timeout**
- Sets `status = active`, clears `timeout_until`

**BR-ADMIN-006 — Activity history**
- Returns paginated log of user events
- Event types: `login`, `product_created`, `product_updated`, `product_deleted`, `order_placed`, `order_cancelled`, `cart_checkout`

---

## Products

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
- Supports query filters: `category`, `min_price`, `max_price`, `search` (matches title)

**BR-PROD-002 — Get product by ID**
- `ErrNotFound` if not found or `status ≠ active`

**BR-PROD-003 — Create product**
- Required: `title`, `price` (> 0), `quantity` (>= 0)
- Optional: `photo`, `description`, `category`, `discount_percentage` (0–100)
- `seller_id` is set from the authenticated user's token (not from request body)
- Default `status`: `active`

**BR-PROD-004 — Update product (partial)**
- Only the owner can edit → `ErrForbidden`
- `price` must be > 0 if provided
- `discount_percentage` must be 0–100 if provided
- `quantity` must be >= 0 if provided
- `seller_id` cannot be changed

**BR-PROD-005 — Toggle visibility**
- Switches `active` ↔ `inactive`
- `inactive` products are hidden from the public listing but remain in the database

**BR-PROD-006 — Delete product**
- Hard delete from database
- Cannot delete if product has orders with `status = pending`, `paid`, or `shipped` → `ErrConflict`
- Admin can delete any product; client can only delete their own → `ErrForbidden`

---

## Cart

| Rule | Endpoint | Method | Auth |
|---|---|---|---|
| BR-CART-001 | `/carts/me` | GET | Client |
| BR-CART-002 | `/carts/me/items` | POST | Client |
| BR-CART-003 | `/carts/me/items/:itemId` | DELETE | Client |
| BR-CART-004 | `/carts/me/items/:itemId` | PATCH | Client |
| BR-CART-005 | `/carts/me/checkout` | POST | Client |

**BR-CART-001 — Get active cart**
- Returns the single `active` cart for the authenticated user
- If no active cart exists, creates one automatically
- Response includes: items, quantities, `price_snapshot` per item, calculated `subtotal`, `discount_total`, `total`

**BR-CART-002 — Add item to cart**
- Required: `product_id`, `quantity` (> 0)
- Product must exist and be `active` → `ErrNotFound`
- `product.seller_id` must ≠ `current_user_id` → `ErrForbidden` (no self-purchase)
- `product.quantity` must be >= requested quantity → `ErrConflict` if insufficient stock
- If the item is already in the cart: add quantities together, re-validate combined total against stock
- Stores `price_snapshot = price * (1 - discount_percentage / 100)` at the moment of add

**BR-CART-003 — Remove item from cart**
- `ErrNotFound` if item does not belong to the user's active cart

**BR-CART-004 — Update item quantity**
- `quantity` must be > 0; to remove use DELETE → `ErrBadRequest` if 0 or negative
- Re-validates new quantity against current product stock → `ErrConflict` if insufficient

**BR-CART-005 — Checkout**
- Re-validates all items: product active, stock sufficient, no self-purchase
- Decrements `product.quantity` for each item
- Creates `Order` (`status = pending`) + `OrderItem` records with price/discount snapshots
- Marks cart `status = checked_out`
- Returns order summary

---

## Orders

| Rule | Endpoint | Method | Auth |
|---|---|---|---|
| BR-ORDER-001 | `/orders/me` | GET | Client |
| BR-ORDER-002 | `/orders/:id` | GET | Client (owner) or Admin |
| BR-ORDER-003 | `/orders/:id/cancel` | PATCH | Client (owner) |
| BR-ADMIN-ORDER-001 | `/admin/orders/:id/status` | PATCH | Admin |

**BR-ORDER-003 — Cancel order**
- Only cancellable if `status = pending` → `ErrConflict` otherwise
- Restores `product.quantity` for each item
- Sets `status = cancelled`

**BR-ADMIN-ORDER-001 — Update order status**
- Valid transitions only: `pending → paid → shipped → delivered`
- Cannot go backwards → `ErrConflict`
- Cannot modify a `cancelled` order → `ErrConflict`

---

## Error Behavior

| `apperrors` sentinel | HTTP | When to use |
|---|---|---|
| `ErrBadRequest` | 400 | Invalid input, missing required fields, constraint violations (price ≤ 0, quantity < 0) |
| `ErrUnauthorized` | 401 | Missing or invalid JWT |
| `ErrForbidden` | 403 | Insufficient role, banned/timeout user, self-purchase, editing another user's resource |
| `ErrNotFound` | 404 | Entity not found or not accessible to the caller |
| `ErrConflict` | 409 | Duplicate email, insufficient stock, invalid state transition, delete blocked by active orders |
| `ErrTooManyRequests` | 429 | Rate limit exceeded (future) |
| `ErrInternalServer` | 500 | Unhandled errors |
