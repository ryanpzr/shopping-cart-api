# Module: User

## Dependencies
- `01-auth` — JWT middleware required for all routes; `users` table already created

## Entities

**User** (full definition)

| Field | Type | Notes |
|---|---|---|
| id | SERIAL PK | |
| name | VARCHAR(255) | NOT NULL |
| email | VARCHAR(255) | NOT NULL, UNIQUE |
| password_hash | VARCHAR(255) | NOT NULL, never returned in responses |
| role | VARCHAR(10) | `admin` / `client` |
| status | VARCHAR(10) | `active` / `banned` / `timeout` |
| timeout_until | TIMESTAMPTZ | NULL unless status = timeout |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

> DB schema was created in `01-auth`. No new migration needed.

## Endpoints

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

**BR-USER-001 — Get own profile**
- Returns: id, name, email, role, status, created_at (no password_hash)

**BR-USER-002 — Update own profile**
- Partial update: `name`, `email`
- Cannot change `role` or `status` — `ErrForbidden` (403)
- New `email` must not be taken → `ErrConflict` (409)

**BR-ADMIN-001 — List all users**
- Returns paginated list of users

**BR-ADMIN-002 — Get user by ID**
- `ErrNotFound` (404) if user does not exist

**BR-ADMIN-003 — Ban user**
- Sets `status = banned`, `updated_at = NOW()`
- Target cannot be another admin → `ErrForbidden` (403)

**BR-ADMIN-004 — Timeout user**
- Required body: `duration_hours` (integer > 0)
- Sets `status = timeout`, `timeout_until = NOW() + duration_hours`
- Target cannot be another admin → `ErrForbidden` (403)

**BR-ADMIN-005 — Unban / remove timeout**
- Sets `status = active`, `timeout_until = NULL`, `updated_at = NOW()`

**BR-ADMIN-006 — Activity history**
- Returns paginated activity log for the given user
- Implemented here but reads from `activity_logs` table created in `06-activity-log`
- Event types: `login`, `product_created`, `product_updated`, `product_deleted`, `order_placed`, `order_cancelled`, `cart_checkout`

> ⚠️ STUB — implementado em `internal/user/features/admin_activity_log/` retornando lista vazia paginada.
> A implementação real deve ser feita no módulo 06: substituir o usecase stub pelo real
> que injeta `activitylog.Repository`. Ver checklist em `06-activity-log.md`.

## Implementation Checklist

- [ ] `internal/user/features/get_profile/` — handler, usecase, ports, dtos
- [ ] `internal/user/features/update_profile/` — handler, usecase, ports, dtos
- [ ] `internal/user/features/admin_list_users/` — handler, usecase, ports, dtos
- [ ] `internal/user/features/admin_manage_user/` — ban, timeout, unban (can be one feature with sub-actions)
- [ ] `internal/user/shared/repository.go` — `FindById`, `FindAll`, `Update`, `UpdateStatus`
- [ ] `internal/user/routes.go` — map routes, apply JWT middleware + role check
- [ ] Wire in `internal/app/setup/setup.go` and `internal/app/route/routes.go`
- [ ] Manual test: get /users/me → update profile → admin ban/timeout/unban flow
