# Site Admin Plans API

## Goal / scope

Implement two endpoints:

- `GET /api/v1/plans` — list all plans from `_portal_plan`
- `POST /api/v1/apps/{app_id}/plan` — change the plan assigned to an app on `_portal_config_source`

Handler stubs and OpenAPI models are already generated (Stage 3). This plan covers the service layer through to the wired `ServeHTTP` bodies.

---

## Data sources

### List plans

Table: `_portal_plan`  
Read via: `plan.Store.List(ctx context.Context) ([]*plan.Plan, error)`  
Returns: `id`, `name`, `feature_config` (feature config is not exposed in the API response).

### Change app plan

Tables touched:
- `_portal_plan` — read to verify `planName` exists (return 404 if not)
- `_portal_config_source` — read by `app_id`, then update `plan_name` and `updated_at`

No schema migration is needed. The `plan_name` column and `plan_change` PostgreSQL NOTIFY trigger already exist.

---

## Runtime flow

### `GET /api/v1/plans`

```
PlansListHandler.ServeHTTP
  → PlansListService.ListPlans(ctx)
      → PlanStore.List(ctx)           // SELECT id, name, feature_config FROM _portal_plan
      ← []*plan.Plan
      → map each plan.Plan → siteadmin.Plan{Name: p.Name}
      ← []siteadmin.Plan
  ← SiteAdminAPISuccessResponse{Body: siteadmin.PlansListResponse{Plans: plans}}.WriteTo(w)
```

No DB transaction needed (read-only, no writes).

### `POST /api/v1/apps/{app_id}/plan`

```
AppPlanChangeHandler.ServeHTTP
  → parseAppPlanChangeParams(r) — already validates plan_name is non-empty
  → AppPlanChangeService.ChangeAppPlan(ctx, appID, planName)
      1. PlanStore.GetPlan(ctx, planName)
           → ErrPlanNotFound → apierrors.NotFound
           ← *plan.Plan (discard, only used for existence check)
      2. GlobalDatabase.WithTx:
           a. ConfigSourceStore.GetDatabaseSourceByAppID(ctx, appID)
                → ErrAppNotFound → apierrors.NotFound (returned from tx)
                ← *configsource.DatabaseSource (capture as `dbs`)
           b. dbs.PlanName = planName
              dbs.UpdatedAt = Clock.NowUTC()
           c. ConfigSourceStore.UpdateDatabaseSource(ctx, dbs)
      3. OwnerStore.GetOwnerByAppID(ctx, appID) — outside TX
           → ErrOwnerNotFound → ownerUserID = ""
      4. if ownerUserID != "":
           AdminAPI.ResolveUserEmails(ctx, []string{ownerUserID})
           ownerEmail = emailMap[ownerUserID]
      5. return &siteadmin.App{
             Id:           appID,
             Plan:         planName,
             CreatedAt:    dbs.CreatedAt,
             OwnerEmail:   ownerEmail,
             LastMonthMau: 0,  // plan change does not fetch MAU; client re-fetches list if needed
         }
  ← SiteAdminAPISuccessResponse{Body: app}.WriteTo(w)
```

The existing PostgreSQL `plan_change` NOTIFY trigger on `_portal_config_source` fires automatically on UPDATE, invalidating cached app configs — no extra code needed.

---

## Error mapping

| Condition | Error |
|---|---|
| `plan_name` does not exist in `_portal_plan` | `apierrors.NotFound` |
| `app_id` does not exist in `_portal_config_source` | `apierrors.NotFound` |

Both are mapped via `writeError(w, r, err)` in the handler — using `errors.Is(err, plan.ErrPlanNotFound)` and `errors.Is(err, configsource.ErrAppNotFound)` in the service to wrap them as `apierrors.NotFound`.

---

## File-level change plan

### Create `pkg/siteadmin/service/plan.go`

```go
package service

import (
    "context"
    "errors"

    "github.com/authgear/authgear-server/pkg/api/apierrors"
    "github.com/authgear/authgear-server/pkg/api/siteadmin"
    "github.com/authgear/authgear-server/pkg/lib/config/configsource"
    "github.com/authgear/authgear-server/pkg/lib/config/plan"
    "github.com/authgear/authgear-server/pkg/util/clock"
)

// Narrow interfaces

type PlanServiceGlobalDatabase interface {
    WithTx(ctx context.Context, do func(ctx context.Context) error) error
}

type PlanServicePlanStore interface {
    GetPlan(ctx context.Context, name string) (*plan.Plan, error)
    List(ctx context.Context) ([]*plan.Plan, error)
}

type PlanServiceConfigSourceStore interface {
    GetDatabaseSourceByAppID(ctx context.Context, appID string) (*configsource.DatabaseSource, error)
    UpdateDatabaseSource(ctx context.Context, dbs *configsource.DatabaseSource) error
}

type PlanServiceOwnerStore interface {
    GetOwnerByAppID(ctx context.Context, appID string) (string, error)
}

// PlanService

type PlanService struct {
    GlobalDatabase    PlanServiceGlobalDatabase
    PlanStore         PlanServicePlanStore
    ConfigSourceStore PlanServiceConfigSourceStore
    OwnerStore        PlanServiceOwnerStore
    AdminAPI          *AdminAPIService
    Clock             clock.Clock
}

func (s *PlanService) ListPlans(ctx context.Context) ([]siteadmin.Plan, error) {
    plans, err := s.PlanStore.List(ctx)
    if err != nil {
        return nil, err
    }
    result := make([]siteadmin.Plan, len(plans))
    for i, p := range plans {
        result[i] = siteadmin.Plan{Name: p.Name}
    }
    return result, nil
}

func (s *PlanService) ChangeAppPlan(ctx context.Context, appID string, planName string) (*siteadmin.App, error) {
    // 1. Verify plan exists.
    _, err := s.PlanStore.GetPlan(ctx, planName)
    if errors.Is(err, plan.ErrPlanNotFound) {
        return nil, apierrors.NotFound.WithReason("PlanNotFound").New("plan not found")
    }
    if err != nil {
        return nil, err
    }

    // 2. Update config source in a transaction.
    var dbs *configsource.DatabaseSource
    err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
        var e error
        dbs, e = s.ConfigSourceStore.GetDatabaseSourceByAppID(ctx, appID)
        if errors.Is(e, configsource.ErrAppNotFound) {
            return apierrors.NotFound.WithReason("AppNotFound").New("app not found")
        }
        if e != nil {
            return e
        }
        dbs.PlanName = planName
        dbs.UpdatedAt = s.Clock.NowUTC()
        return s.ConfigSourceStore.UpdateDatabaseSource(ctx, dbs)
    })
    if err != nil {
        return nil, err
    }

    // 3. Resolve owner email (outside TX).
    ownerEmail := ""
    ownerUserID, err := s.OwnerStore.GetOwnerByAppID(ctx, appID)
    if err != nil && !errors.Is(err, ErrOwnerNotFound) {
        return nil, err
    }
    if ownerUserID != "" {
        emailMap, err := s.AdminAPI.ResolveUserEmails(ctx, []string{ownerUserID})
        if err != nil {
            return nil, err
        }
        ownerEmail = emailMap[ownerUserID]
    }

    return &siteadmin.App{
        Id:           appID,
        Plan:         planName,
        CreatedAt:    dbs.CreatedAt,
        OwnerEmail:   ownerEmail,
        LastMonthMau: 0,
    }, nil
}
```

### Modify `pkg/siteadmin/service/deps.go`

Add to `DependencySet`:
```go
wire.Struct(new(PlanService), "*"),
```

### Modify `pkg/siteadmin/transport/handler_plans_list.go`

Add service interface and field; implement `ServeHTTP`:

```go
type PlansListService interface {
    ListPlans(ctx context.Context) ([]siteadmin.Plan, error)
}

type PlansListHandler struct {
    Service PlansListService
}

func (h *PlansListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    plans, err := h.Service.ListPlans(r.Context())
    if err != nil {
        writeError(w, r, err)
        return
    }
    SiteAdminAPISuccessResponse{Body: siteadmin.PlansListResponse{Plans: plans}}.WriteTo(w)
}
```

### Modify `pkg/siteadmin/transport/handler_app_plan_change.go`

Add service interface and field; implement `ServeHTTP`:

```go
type AppPlanChangeService interface {
    ChangeAppPlan(ctx context.Context, appID string, planName string) (*siteadmin.App, error)
}

type AppPlanChangeHandler struct {
    Service AppPlanChangeService
}

func (h *AppPlanChangeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    params, err := parseAppPlanChangeParams(r)
    if err != nil {
        writeError(w, r, err)
        return
    }
    app, err := h.Service.ChangeAppPlan(r.Context(), params.AppID, params.PlanName)
    if err != nil {
        writeError(w, r, err)
        return
    }
    SiteAdminAPISuccessResponse{Body: app}.WriteTo(w)
}
```

### Modify `pkg/siteadmin/deps.go`

Add wiring for `plan.Store` and `PlanService` bindings:

```go
// plan.Store satisfies PlanServicePlanStore (Clock not needed for read-only methods)
wire.Struct(new(plan.Store), "SQLBuilder", "SQLExecutor"),
wire.Bind(new(siteadminservice.PlanServicePlanStore), new(*plan.Store)),

// PlanServiceGlobalDatabase reuses the existing *globaldb.Handle binding
wire.Bind(new(siteadminservice.PlanServiceGlobalDatabase), new(*globaldb.Handle)),

// PlanServiceConfigSourceStore reuses configsource.Store (already wired)
wire.Bind(new(siteadminservice.PlanServiceConfigSourceStore), new(*configsource.Store)),

// PlanServiceOwnerStore reuses AppOwnerStore (already wired in service/deps.go)
wire.Bind(new(siteadminservice.PlanServiceOwnerStore), new(*siteadminservice.AppOwnerStore)),

// transport bindings
wire.Bind(new(transport.PlansListService), new(*siteadminservice.PlanService)),
wire.Bind(new(transport.AppPlanChangeService), new(*siteadminservice.PlanService)),
```

---

## Test plan

Create `pkg/siteadmin/service/plan_test.go` using `package service`, GoConvey, and hand-written fakes matching `pkg/siteadmin/service/app_test.go` patterns.

### `ListPlans`

| Case | Setup | Expected |
|---|---|---|
| Returns all plans sorted as DB returns | `fakeStore` returns `[{Name: "free"}, {Name: "enterprise"}]` | `[]siteadmin.Plan{{Name: "free"}, {Name: "enterprise"}}` |
| Empty plans table | `fakeStore` returns `[]` | `[]siteadmin.Plan{}` (empty, not nil) |

### `ChangeAppPlan`

| Case | Setup | Expected |
|---|---|---|
| Plan not found | `fakePlanStore.GetPlan` returns `ErrPlanNotFound` | `apierrors.NotFound` |
| App not found | `fakeConfigSourceStore.GetDatabaseSourceByAppID` returns `ErrAppNotFound` | `apierrors.NotFound` |
| Success, no owner | Plan exists, app exists, `fakeOwnerStore` returns `ErrOwnerNotFound` | `App{Plan: newPlan, OwnerEmail: "", ...}` |
| Success, with owner | Plan exists, app exists, owner resolves to email | `App{Plan: newPlan, OwnerEmail: "owner@example.com", ...}` |
| Config source plan_name updated | Capture value passed to `fakeConfigSourceStore.UpdateDatabaseSource` | `dbs.PlanName == newPlan` |

---

## Atomic commit plan

| # | Layer | Files | Verification |
|---|---|---|---|
| 1 | Service | `pkg/siteadmin/service/plan.go` · `pkg/siteadmin/service/plan_test.go` · `pkg/siteadmin/service/deps.go` | `go test ./pkg/siteadmin/service/...` · `make fmt` |
| 2 | Transport interfaces | `pkg/siteadmin/transport/handler_plans_list.go` · `pkg/siteadmin/transport/handler_app_plan_change.go` (add interfaces + Service fields, leave ServeHTTP as stub) | `go build ./pkg/siteadmin/...` · `make fmt` |
| 3 | DI wiring | `pkg/siteadmin/deps.go` · `pkg/siteadmin/wire_gen.go` (via `go run github.com/google/wire/cmd/wire gen`) · `go mod tidy` | `go build ./pkg/siteadmin/... ./cmd/portal/...` · `make fmt` |
| 4 | Handler bodies | `pkg/siteadmin/transport/handler_plans_list.go` · `pkg/siteadmin/transport/handler_app_plan_change.go` (real ServeHTTP) · `.vettedpositions` · `make check-tidy` output | `go test ./pkg/siteadmin/...` · `make fmt` · `make lint` · `go run ./devtools/goanalysis ...` · `make sort-vettedpositions` · `make check-tidy` |
