# Module: Auth

## Dependencies
None — this is the foundation. All other modules depend on it.

## Entities

**User** (auth-relevant fields only; full profile managed in `02-user`)

| Field | Type | Notes |
|---|---|---|
| id | SERIAL PK | |
| name | VARCHAR(255) | NOT NULL |
| email | VARCHAR(255) | NOT NULL, UNIQUE |
| password_hash | VARCHAR(255) | NOT NULL, bcrypt |
| role | VARCHAR(10) | `admin` / `client`, DEFAULT `client` |
| status | VARCHAR(10) | `active` / `banned` / `timeout`, DEFAULT `active` |
| timeout_until | TIMESTAMPTZ | NULL when not in timeout |
| created_at | TIMESTAMPTZ | DEFAULT NOW() |
| updated_at | TIMESTAMPTZ | DEFAULT NOW() |

## DB Schema

```sql
CREATE TABLE users (
    id            SERIAL PRIMARY KEY,
    name          VARCHAR(255)  NOT NULL,
    email         VARCHAR(255)  NOT NULL UNIQUE,
    password_hash VARCHAR(255)  NOT NULL,
    role          VARCHAR(10)   NOT NULL DEFAULT 'client'  CHECK (role IN ('admin', 'client')),
    status        VARCHAR(10)   NOT NULL DEFAULT 'active'  CHECK (status IN ('active', 'banned', 'timeout')),
    timeout_until TIMESTAMPTZ,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);
```

## Endpoints

| Rule | Endpoint | Method | Auth |
|---|---|---|---|
| BR-AUTH-001 | `/auth/register` | POST | Public |
| BR-AUTH-002 | `/auth/login` | POST | Public |

**BR-AUTH-001 — Register**
- Required body: `name`, `email`, `password`
- `email` must be unique → `ErrConflict` (409)
- `password` stored as bcrypt hash — never stored in plain text
- Default role: `client`
- Returns: user info (no password_hash) + JWT token

**BR-AUTH-002 — Login**
- Required body: `email`, `password`
- Returns JWT (payload: `user_id`, `role`, `email`, `exp`)
- `401` if email not found or password doesn't match
- `403` if `status = banned`
- `403` if `status = timeout` AND `timeout_until > NOW()`

**BR-AUTH-003 — JWT Middleware** (used by all other modules)
- Header: `Authorization: Bearer <token>`
- `401` if header missing or token invalid/expired
- `403` if token valid but role insufficient for the route
- Injects `user_id`, `role`, `email` into `gin.Context` for downstream handlers

## Implementation Checklist

- [ ] Migration SQL — create `users` table
- [ ] `internal/user/domain/user.go` — User struct
- [ ] `internal/user/shared/repository.go` — `FindByEmail`, `Create`
- [ ] `internal/auth/features/register/` — handler, usecase, ports, dtos
- [ ] `internal/auth/features/login/` — handler, usecase, ports, dtos
- [ ] `pkg/middleware/auth.go` — JWT validation middleware (reused by all modules)
- [ ] `internal/auth/routes.go` — `POST /register`, `POST /login`
- [ ] Wire in `internal/app/setup/setup.go` and `internal/app/route/routes.go`
- [ ] Manual test: register → login → receive token → use token on a protected route
