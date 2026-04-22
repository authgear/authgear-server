---
name: new-siteadmin-api
description: Full pipeline for adding a new Site Admin API feature — from OpenAPI spec through implementation plan to working service. Use when adding a new endpoint or filling in real data for an existing stub.
---

Add a new Site Admin API feature: $ARGUMENTS

Follow the stages below in order. Confirm with the user before moving from one stage to the next.

---

## Stage 1: Design and confirm OpenAPI spec

Add the endpoint to `docs/api/siteadmin-api.yaml`:
- Add the path under `paths:` with appropriate HTTP method(s)
- Use `$ref: "#/components/responses/BadRequest"`, `Forbidden`, `NotFound` for error responses
- Add any new request/response schemas to `components/schemas:`
- Follow existing conventions: snake_case fields, hyphenated paths, `format: date` for date strings

> **Stop here.** Present the spec diff to the user and wait for confirmation before proceeding.

---

## Stage 2: Generate models

```bash
make generate
```

> **Avoid external type imports**: `pkg/siteadmin/model/oapi-codegen.yaml` maps `format: date` → `string` and `format: email` → `string` so that no `github.com/oapi-codegen/runtime/types` import is generated. If you add a new OpenAPI format that would otherwise pull in an external type, add a mapping to `output-options.type-mapping` in that file before regenerating.

Commit: `"[Site Admin] Generate models for <feature>"`

---

## Stage 3: Create and wire the handler file

Create `pkg/siteadmin/transport/handler_<name>.go`. **No dummy data** — leave `ServeHTTP` as a stub that returns `http.NotFound`:

```go
package transport

import "net/http"

func Configure<Name>Route(route httproute.Route) httproute.Route {
    return route.WithMethods("OPTIONS", "GET"). // always include "OPTIONS" for CORS preflight
        WithPathPattern("/api/v1/...")
}

type <Name>Handler struct {
    // Service dependencies added in Stage 5
}

type <Name>Params struct {
    // path and query params
}

func parse<Name>Params(r *http.Request) (<Name>Params, error) {
    // parse path params via httproute.GetParam(r, "paramName")
    // query params: use helpers from params.go (getIntParam, getDateParam, etc.)
    // POST body: decode JSON, validate with validation.NewSimpleSchema
}

func (h *<Name>Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    _, err := parse<Name>Params(r)
    if err != nil {
        writeError(w, r, err)
        return
    }
    http.NotFound(w, r)
}
```

**Param validation rules:**
- All routes (GET, POST, DELETE, …) must include `"OPTIONS"` in `WithMethods` for CORS preflight
- When multiple handlers share the same path (e.g. GET + POST), only one of them should declare `"OPTIONS"` to avoid duplicate preflight registrations — conventionally the first route registered for that path
- Use `writeError(w, r, err)` for all error responses
- Use helpers in `params.go`: `getIntParam`, `getDateParam`, `validateMonth`, `makeValidationError`
- Do NOT use `apierrors.NewBadRequest` for query param errors
- Range checks (e.g. start ≤ end, range ≤ 1 year): use `makeValidationError` in the transport layer — not the service layer
- 1-year range check: `end.After(start.AddDate(1, 0, 0))` — handles leap years; do NOT use `end.Sub(start) > 365*24*time.Hour`

**Wire the handler:**

`pkg/siteadmin/transport/deps.go` — add:
```go
wire.Struct(new(<Name>Handler), "*"),
```

`pkg/siteadmin/wire.go` — add injector:
```go
func new<Name>Handler(p *deps.RequestProvider) http.Handler {
    panic(wire.Build(DependencySet, wire.Bind(new(http.Handler), new(*transport.<Name>Handler))))
}
```

`pkg/siteadmin/routes.go` — register:
```go
router.Add(transport.Configure<Name>Route(route), p.Handler(new<Name>Handler))
```

Then regenerate:
```bash
wire gen ./pkg/siteadmin/...
go build ./pkg/siteadmin/...
```

Commit: `"Add <name> handler scaffolding"`

---

## Stage 4: Write, review, and confirm implementation plan

Create `docs/plans/siteadmin-api/<N>-<name>.md` using the `/write-implementation-plan` skill.

The plan must specify:
- Data sources (table, column, filter conditions)
- Service interface and struct
- Flow for each method (transport validation → service → DB → response)
- Files to create and modify, with exact code samples
- Atomic commits with verification steps per commit

**Review checklist before confirming:**
- [ ] All date/time bounds are explicit about inclusive vs exclusive
- [ ] Interface parameter names make bound inclusivity explicit (e.g. `toExclusive`, `toInclusive`) when it is not obvious from context
- [ ] Range validations (ordering, max range) assigned to transport layer
- [ ] No open `for {}` loops — bounded iteration with early return on invalid range
- [ ] Test fakes mirror real SQL semantics (filter by time range, not just by name)
- [ ] The final implementation commit (handler `ServeHTTP` bodies) includes `goanalysis`, `make sort-vettedpositions`, `make check-tidy`

> **Stop here.** Present the plan to the user and wait for confirmation before proceeding.

---

## Stage 5: Run the implementation plan

Follow the commit sequence below.

### Always read before writing

Read the actual source files — do not rely solely on plan code samples, which can drift:
- `pkg/siteadmin/service/deps.go` and `pkg/siteadmin/deps.go` before touching wire sets
- `pkg/siteadmin/service/app_test.go` for test package and fake patterns (use `package service`, not `package service_test`)

### Response pattern

```go
SiteAdminAPISuccessResponse{Body: result}.WriteTo(w)
```
Not `json.NewEncoder` or manual `w.WriteHeader` + `w.Write`.

### Service layer conventions

- Define narrow interfaces in the service file (not in `deps.go`)
- For partial struct wiring: `wire.Struct(new(X), "Field1", "Field2")` in `pkg/siteadmin/deps.go`
- `wire.Bind` goes in `pkg/siteadmin/deps.go`, not `pkg/siteadmin/service/deps.go`

### Date/time conventions

- Daily exclusive upper bound: `end.AddDate(0, 0, 1)`
- Monthly exclusive upper bound: `time.Date(y, time.Month(m)+1, 1, ...)` — no December special case
- Interface parameter name for exclusive bound: `toEndTimeExclusive`

### Defensive iteration

```go
total := (endYear-startYear)*12 + (endMonth-startMonth) + 1
if total <= 0 {
    return &T{Counts: nil}, nil
}
counts := make([]Item, 0, total)
for i := 0; i < total; i++ { ... }
```

### Test fakes

Fakes must filter by the time range parameters (mirror real SQL):

```go
func (f *fakeStore) FetchUsageRecordsInRange(_ context.Context, _ string, name RecordName, _ periodical.Type, from, toExclusive time.Time) ([]*UsageRecord, error) {
    var out []*UsageRecord
    for _, r := range f.byName[name] {
        t := r.StartTime.UTC().Truncate(24 * time.Hour)
        if !t.Before(from) && t.Before(toExclusive) {
            out = append(out, r)
        }
    }
    return out, nil
}
```

Always set `StartTime` on fake records when date-range filtering is involved.

### DI wiring

After any `deps.go` change:
```bash
go run github.com/google/wire/cmd/wire gen ./pkg/siteadmin/...
go mod tidy
go build ./pkg/siteadmin/...
go build ./cmd/portal/...
```

### `.vettedpositions`

Every new `r.Context()` in `pkg/siteadmin/transport/` is flagged. After adding `ServeHTTP` bodies:
```bash
go run ./devtools/goanalysis ./cmd/... ./pkg/...
# add each new position to .vettedpositions
make sort-vettedpositions
go run ./devtools/goanalysis ./cmd/... ./pkg/...  # must be clean
```

### Final gate: `make check-tidy`

Run once, on the last commit only (`make check-tidy` is slow — it regenerates everything). It reruns `wire gen` (globally) and `make fmt` — the output files will be dirty. Stage those regenerated/formatted files and include them in the final commit. Do not re-run `make check-tidy` after staging.

### Commit sequence

Commits must be **atomic** and **ordered by dependency** — each commit must build and test cleanly on its own. Do not mix content from two different logical parts in one commit. The number of commits will vary with complexity.

Typical ordering (split further if a step is large):

| Layer | Typical contents | Verification |
|---|---|---|
| Service | Service file + test + `service/deps.go` | `go test ./pkg/siteadmin/service/...` · `make fmt` |
| Transport interfaces | Narrow interfaces on handlers + `Service` fields (no `ServeHTTP` bodies yet) | `go build ./pkg/siteadmin/...` · `make fmt` |
| DI wiring | `pkg/siteadmin/deps.go` + `wire gen` + `go mod tidy` | build both packages · `make fmt` |
| Handler bodies | `ServeHTTP` implementations + `.vettedpositions` + `make check-tidy` output | `go test ./pkg/siteadmin/...` · `make fmt` · `make lint` · `goanalysis` · `make sort-vettedpositions` · `make check-tidy` |

If a feature requires additional layers (e.g. a new DB store method, a shared helper, a schema migration, a refactor of existing code), add commits for those between the layers above. Never combine, for example, the service layer and the DI wiring in one commit. Refactors to existing code always get their own commit.
