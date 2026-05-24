# Shopping Cart API

REST API for a shopping cart system built with Go and Gin. Handles authentication, user management, product management, cart, and orders.

## Tech Stack

- **Language:** Go 1.24
- **Framework:** Gin
- **Database:** PostgreSQL 16
- **Auth:** JWT (HS256, 24h expiry)
- **Password hashing:** bcrypt
- **Container:** Docker + Docker Compose

## Project Structure

```
cmd/             # Entry point
config/          # Database connection setup
internal/
  app/           # Wiring (setup + route aggregation)
  auth/          # Auth module (register, login)
  user/
    domain/      # User entity
    shared/      # Repository (shared across features)
    features/
      get_me/              # GET /users/me
      update_me/           # PUT /users/me
      admin_get_user/      # GET /admin/users/:id
      admin_list_users/    # GET /admin/users
      admin_manage_user/   # PATCH /admin/users/:id/ban|timeout|unban
      admin_activity_log/  # GET /admin/users/:id/activity (stub)
    routes.go    # Client and admin route mapping
  product/       # Product module
  cart/          # Cart module (in progress)
pkg/
  apperrors/     # Typed error handling
  jwt/           # JWT generate/parse
  middleware/    # Auth and role middleware
sdd/specs/       # Business rules and module specifications
db/migrations/   # SQL migration files
```

## Getting Started

### Prerequisites

- Docker and Docker Compose

### Environment Variables

Create a `.env` file at the project root:

```env
JWT_SECRET=your_secret_key_here
```

The database connection is configured in `docker-compose.yml`. For local development, the defaults work out of the box — no changes needed.

### Running

```bash
# First run (creates database schema automatically)
docker-compose down -v && docker-compose up --build

# Subsequent runs
docker-compose up --build
```

> The database migration runs automatically on first startup via `initdb.d`. To re-run it, bring the volume down with `docker-compose down -v`.

The API will be available at `http://localhost:8080`.

## API Reference

All routes are prefixed with `/api/v1`.

### Authentication

Public routes — no token required.

#### Register

```
POST /api/v1/auth/register
```

**Body:**

```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "minlength8"
}
```

**Validation:**
- All fields required
- `email` must be a valid email address
- `password` minimum 8 characters

**Response `201`:**

```json
{
  "user": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "role": "client",
    "status": "active"
  },
  "token": "<jwt>"
}
```

**Errors:**

| Status | Reason |
|---|---|
| 400 | Missing or invalid fields |
| 409 | Email already registered |

---

#### Login

```
POST /api/v1/auth/login
```

**Body:**

```json
{
  "email": "john@example.com",
  "password": "minlength8"
}
```

**Response `200`:**

```json
{
  "user": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "role": "client",
    "status": "active"
  },
  "token": "<jwt>"
}
```

**Errors:**

| Status | Reason |
|---|---|
| 400 | Missing fields |
| 401 | Invalid email or password |
| 403 | Account restricted |

---

### Protected Routes

All routes below require the `Authorization` header:

```
Authorization: Bearer <token>
```

| Status | Reason |
|---|---|
| 401 | Token missing, malformed, or expired |
| 403 | Token valid but insufficient role |

---

### User — Client

Requires a valid token (any role).

#### Get own profile

```
GET /api/v1/users/me
```

**Response `200`:**

```json
{
  "id": 1,
  "name": "John Doe",
  "email": "john@example.com",
  "role": "client",
  "status": "active",
  "created_at": "2025-01-01T00:00:00Z"
}
```

---

#### Update own profile

```
PUT /api/v1/users/me
```

Partial update — send only the fields you want to change.

**Body:**

```json
{
  "name": "New Name",
  "email": "newemail@example.com"
}
```

**Validation:**
- At least one field (`name` or `email`) must be provided
- `email`, if provided, must be a valid email address
- `role` and `status` cannot be changed through this endpoint

**Response `200`:** updated user profile (same shape as GET /users/me, plus `updated_at`)

**Errors:**

| Status | Reason |
|---|---|
| 400 | Invalid or missing fields |
| 409 | Email already in use |

---

### User — Admin

Requires a valid token with `admin` role.

#### List all users

```
GET /api/v1/admin/users
```

**Query params:**

| Param | Default | Max | Description |
|---|---|---|---|
| `page` | `1` | — | Page number |
| `limit` | `20` | `100` | Items per page |

**Response `200`:**

```json
{
  "data": [
    {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "role": "client",
      "status": "active",
      "created_at": "2025-01-01T00:00:00Z"
    }
  ],
  "total": 100,
  "page": 1,
  "limit": 20,
  "total_pages": 5
}
```

---

#### Get user by ID

```
GET /api/v1/admin/users/:id
```

**Response `200`:** full user object including `timeout_until` and `updated_at`

**Errors:**

| Status | Reason |
|---|---|
| 400 | Invalid ID |
| 404 | User not found |

---

#### Ban user

```
PATCH /api/v1/admin/users/:id/ban
```

**Errors:**

| Status | Reason |
|---|---|
| 403 | Target user is an admin |
| 404 | User not found |

---

#### Timeout user

```
PATCH /api/v1/admin/users/:id/timeout
```

**Body:**

```json
{
  "duration_hours": 24
}
```

**Validation:**
- `duration_hours` must be a positive integer

**Errors:**

| Status | Reason |
|---|---|
| 400 | Missing or invalid `duration_hours` |
| 403 | Target user is an admin |
| 404 | User not found |

---

#### Remove restriction (unban / undo timeout)

```
PATCH /api/v1/admin/users/:id/unban
```

Sets the user's status back to `active`.

**Errors:**

| Status | Reason |
|---|---|
| 403 | Target user is an admin |
| 404 | User not found |

---

#### Activity log

```
GET /api/v1/admin/users/:id/activity
```

> **Note:** This endpoint is available but returns an empty list until the activity log module is fully implemented.

**Query params:** `page`, `limit` (same defaults as list users)

**Response `200`:**

```json
{
  "data": [],
  "total": 0,
  "page": 1,
  "limit": 20,
  "total_pages": 0
}
```

---

### Products

#### List all products

```
GET /api/v1/products
```

**Response `200`:**

```json
{
  "data": [
    {
      "id": 1,
      "photo": "url",
      "title": "Product Name",
      "description": "...",
      "price": 49.90,
      "quantity": 10
    }
  ]
}
```

#### Create product

```
POST /api/v1/products
```

#### Update product

```
PUT /api/v1/products/:productId
```

---

## User Roles

| Role | Description |
|---|---|
| `client` | Default role assigned on register |
| `admin` | Extended permissions — assigned directly in the database |

---

## Running Tests

```bash
go test ./...
```

Tests use `testify` with mocked dependencies — no database connection required.

---

## Modules

| Module | Status | Spec |
|---|---|---|
| Auth | Done | [01-auth.md](sdd/specs/modules/01-auth.md) |
| User | Done | [02-user.md](sdd/specs/modules/02-user.md) |
| Product | Partial | [03-product.md](sdd/specs/modules/03-product.md) |
| Cart | In progress | [04-cart.md](sdd/specs/modules/04-cart.md) |
| Order | Planned | [05-order.md](sdd/specs/modules/05-order.md) |
| Activity Log | Planned | [06-activity-log.md](sdd/specs/modules/06-activity-log.md) |
