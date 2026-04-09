---
name: new_siteadmin_api
description: Scaffold a new Site Admin API endpoint. Use when adding a new route to the site admin server.
disable-model-invocation: true
---

Scaffold a new Site Admin API endpoint: $ARGUMENTS

Follow these steps in order:

## 1. Update the OpenAPI spec

Add the new endpoint to `docs/api/siteadmin-api.yaml`:
- Add the path under `paths:`
- Use `$ref: "#/components/responses/BadRequest"`, `Forbidden`, `NotFound` for error responses
- Add any new request/response schemas to `components/schemas:`
- Follow existing naming conventions (snake_case fields, hyphenated paths)

## 2. Regenerate models

```
make generate
```

> **Avoid external type imports**: `pkg/siteadmin/model/oapi-codegen.yaml` maps `format: date` → `string` and `format: email` → `string` so that no `github.com/oapi-codegen/runtime/types` import is generated. If you add a new OpenAPI format that would otherwise pull in an external type, add a mapping to `output-options.type-mapping` in that file before regenerating.

## 3. Create the handler file

Create `pkg/siteadmin/transport/handler_<name>.go` following this pattern:

```go
package transport

func Configure<Name>Route(route httproute.Route) httproute.Route {
    return route.WithMethods("GET"). // add "OPTIONS" for POST/DELETE
        WithPathPattern("/api/v1/...")
}

type <Name>Handler struct {
    // Add service dependencies here as needed
}

type <Name>Params struct {
    // path and query params
}

func parse<Name>Params(r *http.Request) (<Name>Params, error) {
    // parse path params via httproute.GetParam(r, "paramName")
    // for GET query params: use helpers from params.go (getIntParam, getDateParam, etc.)
    // for POST: decode JSON body into generated request model, validate with validation.NewSimpleSchema
}

func (h *<Name>Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    params, err := parse<Name>Params(r)
    if err != nil {
        writeError(w, r, err)
        return
    }
    _ = params
    // TODO: implement
    http.NotFound(w, r)
}
```

Key rules:
- POST routes must include `"OPTIONS"` in `WithMethods` for CORS preflight
- DELETE routes must include `"OPTIONS"` in `WithMethods` for CORS preflight
- Use `writeError(w, r, err)` for all error responses
- Use `apierrors.NewNotFound(...)` for 404s
- For GET query param validation, use the helpers in `pkg/siteadmin/transport/params.go` — they return `ValidationFailed` errors:
  - `getIntParam(q, name)` / `getOptionalIntParam(q, name)` — required/optional integer
  - `getDateParam(q, name)` / `getOptionalDateParam(q, name)` — required/optional `YYYY-MM-DD` date
  - `validateMonth(name, v)` — validates value is 1–12
  - For custom range/business-rule errors, use `makeValidationError(func(ctx *validation.Context) { ... })`
  - Do NOT use `apierrors.NewBadRequest` for query param errors
  - Do NOT use JSON schema for query params — query params arrive as strings so schema type/format checks don't apply cleanly
- For POST body validation, use `validation.NewSimpleSchema` and `ValidateValue(r.Context(), body)`
- Embed generated request model in params struct instead of copying fields

## 4. Wire up the handler

Update these files:

**`pkg/siteadmin/transport/deps.go`** — add to `DependencySet`:
```go
wire.Struct(new(<Name>Handler), "*"),
```

**`pkg/siteadmin/wire.go`** — add injector:
```go
func new<Name>Handler(p *deps.RequestProvider) http.Handler {
    panic(wire.Build(
        DependencySet,
        wire.Bind(new(http.Handler), new(*transport.<Name>Handler)),
    ))
}
```

**`pkg/siteadmin/wire_gen.go`** — add generated function:
```go
func new<Name>Handler(p *deps.RequestProvider) http.Handler {
    <name>Handler := &transport.<Name>Handler{}
    return <name>Handler
}
```

**`pkg/siteadmin/routes.go`** — register the route:
```go
router.Add(transport.Configure<Name>Route(route), p.Handler(new<Name>Handler))
```

## 5. Build and verify

```
go build ./pkg/siteadmin/...
```

## 6. Commit

Create two commits:
1. `Add <name> handler scaffolding with params parsing`
2. `Return dummy data from <name> handler` (if dummy data is needed)
