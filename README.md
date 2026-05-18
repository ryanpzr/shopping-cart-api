# Shopping Cart API

REST API for a shopping cart system built with Go and Gin. Handles authentication, product management, cart, and orders.

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
  user/          # User domain and repository
  product/       # Product module
  cart/          # Cart module (in progress)
pkg/
  apperrors/     # Typed error handling
  jwt/           # JWT generate/parse
  middleware/    # Auth middleware
sdd/specs/       # Business rules and module specifications
db/migrations/   # SQL migration files (not committed)
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
| 403 | Account is banned or in timeout |

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
| `admin` | Extended permissions (managed via database) |

## Running Tests

```bash
go test ./...
```

Tests use `testify` with mocked dependencies — no database connection required.

## Modules

| Module | Status | Spec |
|---|---|---|
| Auth | Done | [01-auth.md](sdd/specs/modules/01-auth.md) |
| User | Planned | [02-user.md](sdd/specs/modules/02-user.md) |
| Product | Partial | [03-product.md](sdd/specs/modules/03-product.md) |
| Cart | In progress | [04-cart.md](sdd/specs/modules/04-cart.md) |
| Order | Planned | [05-order.md](sdd/specs/modules/05-order.md) |
| Activity Log | Planned | [06-activity-log.md](sdd/specs/modules/06-activity-log.md) |
