# Infrastructure Architecture Spec

## Overview

The application uses a three-layer architecture to wire modules into the HTTP server:

| Layer | Path | Responsibility |
|---|---|---|
| **App Setup** | `internal/app/setup/` | Wires dependencies (repository → usecase → handler) for all modules |
| **App Route** | `internal/app/route/` | Aggregates modules under path group prefixes, delegates to module routes |
| **Module Routes** | `internal/<module>/routes.go` | Declares the HTTP routes of a specific module |

---

## Directory Structure

```
internal/
├── app/
│   ├── setup/
│   │   └── setup.go          # Dependency wiring for all modules
│   └── route/
│       └── routes.go         # Path group aggregator
└── <module>/
    ├── domain/
    │   └── <model>.go        # Domain structs
    ├── features/
    │   └── <feature>/
    │       ├── handler.go    # HTTP handler
    │       ├── usecase.go    # Business logic
    │       ├── ports.go      # Interfaces (Usecase, Repository)
    │       └── dtos.go       # Request/response types
    ├── shared/
    │   └── repository.go     # Database access interface + implementation
    ├── routes.go             # Module route definitions
    └── setup.go              # Optional: module-level stub (for future use)
```

---

## Data Flow

```
cmd/main.go
  └── setup.Setup(api, conn)
        │
        │  [internal/app/setup/setup.go]
        │  Creates: repository → usecase → handler
        │
        └── route.RegisterRoutes(api, hd<Module>)
              │
              │  [internal/app/route/routes.go]
              │  Groups: api.Group("/<module>")
              │
              └── <module>.MapRouters(group, hd)
                    │
                    │  [internal/<module>/routes.go]
                    │  Registers: GET, POST, PUT, DELETE ...
                    │
                    └── hd.<MethodName>(ctx)
                          [internal/<module>/features/<feature>/handler.go]
```

---

## Layer Descriptions

### `internal/app/setup/setup.go`

Single entry point called from `main()`. Instantiates all module dependencies in order and passes handlers to `route.RegisterRoutes`.

```go
func Setup(api *gin.RouterGroup, conn *sql.DB) {
    rp<Module> := <module>shared.NewRepository(conn)
    us<Module> := <feature>.NewUsecase(rp<Module>)
    hd<Module> := <feature>.NewHandler(us<Module>)
    route.RegisterRoutes(api, hd<Module>)
}
```

Rules:
- Receives `*gin.RouterGroup` and `*sql.DB` — no global state
- One block per module, separated by a blank line
- Does not define any routes

---

### `internal/app/route/routes.go`

Aggregates all module route registrations under their path group prefix. One line per module.

```go
func RegisterRoutes(api *gin.RouterGroup, hd<Module> <feature>.Handler /*, ... */) {
    <module>Group := api.Group("/<module>")
    <module>.MapRouters(<module>Group, hd<Module>)
}
```

Rules:
- Only calls `api.Group()` and `<module>.MapRouters()` — no handler logic here
- Grows one entry per new module added

---

### `internal/<module>/routes.go`

Owns all HTTP route declarations for the module. Uses the router group received from the aggregator (path prefix already applied).

```go
func MapRouters(r *gin.RouterGroup, hd Handler) {
    r.GET("", hd.GetAll)
    r.GET("/:id", hd.GetById)
    r.POST("", hd.Create)
    r.PUT("/:id", hd.Update)
    r.DELETE("/:id", hd.Delete)
}
```

Rules:
- Routes are relative — the prefix (e.g. `/products`) comes from the aggregator
- `Handler` here refers to the interface defined in `ports.go` of the feature package

---

## Checklist: Adding a New Module

- [ ] Create `internal/<module>/domain/<model>.go` with domain structs
- [ ] Create `internal/<module>/shared/repository.go` with `Repository` interface and `repository` struct implementing it
- [ ] Create `internal/<module>/features/<feature>/ports.go` with `Usecase` and `Repository` interfaces
- [ ] Create `internal/<module>/features/<feature>/dtos.go` with request/response types
- [ ] Create `internal/<module>/features/<feature>/usecase.go` implementing `Usecase`
- [ ] Create `internal/<module>/features/<feature>/handler.go` implementing `Handler`
- [ ] Create `internal/<module>/routes.go` with `MapRouters(r *gin.RouterGroup, hd Handler)`
- [ ] Add wiring block in `internal/app/setup/setup.go`
- [ ] Add `api.Group("/<module>")` + `<module>.MapRouters(...)` in `internal/app/route/routes.go`
- [ ] Run `go build ./...` — must compile with no errors

---

## Naming Conventions

| Artifact | Convention | Example |
|---|---|---|
| Module package | lowercase, no underscore | `product`, `cart`, `order` |
| Feature package | camelCase joined | `additemtocart`, `changeinfoproduct` |
| Repository var | `rp<Module>` | `rpProduct` |
| Usecase var | `us<Module>` | `usProduct` |
| Handler var | `hd<Module>` | `hdProduct` |
| Route group | `/<module>` (plural preferred) | `/products`, `/carts` |
| Route function | `MapRouters` | same across all modules |
| Setup function | `Setup` | same across all module-level stubs |
