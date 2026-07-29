# 09 — App Feature Config API

## Goal / Scope

Implement the three endpoints specified in `docs/api/siteadmin-api.yaml`:

| Endpoint | operationId | Method |
|---|---|---|
| `/api/v1/apps/{app_id}/feature-config` | `getAppFeatureConfig` | GET |
| `/api/v1/apps/{app_id}/feature-config` | `updateAppFeatureConfig` | PUT |
| `/api/v1/apps/{app_id}/feature-config/preview` | `previewAppFeatureConfig` | POST |

Nothing under `pkg/siteadmin/` currently references feature config — there is
no handler stub, no generated types yet. This plan covers spec-to-code
generation, handler scaffolding, and full implementation in one pass (Stages
2, 3, and 5 of `new-siteadmin-api` combined — the spec is already written and
reviewed, so Stage 1 is done), plus DI wiring, unit tests, and e2e coverage.

**Baseline assumption**: feature config merge is field-level for every
section — setting one field never resets a sibling field. See
`pkg/lib/config/feature_*.go` for the `Merge` implementations. This plan's
test fixtures (§"Test plan") use `collaborator` and `oauth.client` simply as
two structurally different examples (a two-field section and a
four-mixed-type-field section), not to contrast merge strategies.

One new service (`FeatureConfigService`), one new audit payload
(`site_admin.app.feature_config.updated`), three new transport handlers. No
new exported symbols needed from `pkg/lib/config/configsource`.

---

## Merge computation (shared by GET / PUT / POST)

New private helpers in `pkg/siteadmin/service/feature_config.go`, used by all
three public methods. No new exported symbols are added to
`pkg/lib/config/configsource` — everything needed is already exported.

Layer stack to fold, matching `dbApp.doLoad`
(`pkg/lib/config/configsource/database.go:469-470`):
**`BaseResources ⟵ plan ⟵ app`**, not just `[plan, app]`. `FeatureConfigService`
gets a new `BaseResources deps.AppBaseResources` field (already provided by
`deps.DependencySet`, no new wiring needed — see "DI wiring" below).
`deps.AppBaseResources` is a distinct named type over `*resource.Manager`,
not an alias, so convert with `(*resource.Manager)(s.BaseResources)` before
calling `.Overlay`.

```go
func featureConfigOverrideKey() string {
    return filepathutil.EscapePath(configsource.AuthgearFeatureYAML)
}

// computeLayers folds BaseResources, the plan named planName, and a
// candidate app-override document (nil/empty overrideYAML = no override)
// exactly the way dbApp.doLoad computes an app's effective feature config —
// same layer stack, same order, same FS-construction helpers. No merge
// logic is reimplemented.
func (s *FeatureConfigService) computeLayers(
    ctx context.Context,
    planName string,
    overrideYAML []byte,
) (planEffective *config.FeatureConfig, effective *config.FeatureConfig, err error) {
    p, err := s.PlanStore.GetPlan(ctx, planName)
    if err != nil && !errors.Is(err, plan.ErrPlanNotFound) {
        return nil, nil, err
    }
    // On ErrPlanNotFound, p stays nil. MakePlanFSFromPlan(nil) produces an
    // empty plan layer — the same tolerance dbApp.doLoad has for an app
    // referencing a missing plan.

    planFs, err := configsource.MakePlanFSFromPlan(p)
    if err != nil {
        return nil, nil, err
    }

    appData := map[string][]byte{}
    if len(overrideYAML) > 0 {
        appData[featureConfigOverrideKey()] = overrideYAML
    }
    appFs, err := configsource.MakeAppFSFromDatabaseSource(&configsource.DatabaseSource{Data: appData})
    if err != nil {
        return nil, nil, err
    }

    base := (*resource.Manager)(s.BaseResources)
    planEffective, err = readEffectiveFeatureConfig(ctx, base.Overlay(planFs))
    if err != nil {
        return nil, nil, err
    }
    effective, err = readEffectiveFeatureConfig(ctx, base.Overlay(planFs).Overlay(appFs))
    if err != nil {
        return nil, nil, err
    }
    return planEffective, effective, nil
}

// readEffectiveFeatureConfig mirrors configsource.LoadConfig's feature config
// resolution (pkg/lib/config/configsource/loader.go:31): read the
// FeatureConfig resource from the given manager, falling back to server
// defaults when no layer has an authgear.features.yaml file.
func readEffectiveFeatureConfig(ctx context.Context, mgr *resource.Manager) (*config.FeatureConfig, error) {
    result, err := mgr.Read(ctx, configsource.FeatureConfig, resource.EffectiveResource{})
    if errors.Is(err, resource.ErrResourceNotFound) {
        return config.NewEffectiveDefaultFeatureConfig(), nil
    }
    if err != nil {
        return nil, err
    }
    return result.(*config.FeatureConfig), nil
}
```

`plan.Plan.RawFeatureConfig` is a parsed `*config.FeatureConfig`
(`plan/store.go:147`), sparse, no defaults — `MakePlanFSFromPlan` re-marshals
it to YAML, matching `dbApp.doLoad`'s round-trip.

---

## Validation error contract

Existing server-wide convention, not new for this endpoint: a JSON Schema
validation failure returns `*validation.AggregatedError`
(`pkg/util/validation/error.go`), which `pkg/api/apierrors`'s `asAPIError`
automatically converts into a 400 response with `info.causes` — an array of
`{location, kind, details}`, `location` being an RFC 6901 JSON pointer to the
failing field. This is the `ValidationErrorCause` OpenAPI schema. No new
conversion code is written for this endpoint.

Specific to this plan:

1. **This service never calls `config.ParseFeatureConfigWithoutDefaults`
   directly on the raw submitted document.** It's called twice, both times
   inside `viewEffectiveResource` (`pkg/lib/config/configsource/resources.go:694-716`),
   which `computeLayers`/`readEffectiveFeatureConfig` invoke via
   `Manager.Read`: once per layer while folding (the submitted override is
   one of those layers), and once more on the fully merged YAML.
2. `UpdateAppFeatureConfig`/`PreviewAppFeatureConfig` must return whatever
   error `computeLayers` produces **unmodified** (or `%w`-wrapped —
   `errors.As` still unwraps it). **Do not** catch and rebuild it; that would
   silently drop `causes`/`location`.
3. **One hand-built exception**: the multi-document-YAML guard (see
   "Validation" below) returns `apierrors.ValidationFailed.New("...")`
   directly — same `reason: ValidationFailed`, but **no `causes`**, since the
   failure is about the whole document, not one field. Tests must cover both
   shapes (with `causes`, and without).

---

## Validation (PUT and POST share this)

Schema validation of the submitted document is **not** a standalone
pre-check. This service only pre-checks what JSON schema can't express —
emptiness, multi-document-YAML — and leaves `computeLayers` (next, see
"Runtime flow") as the single source of truth for schema validity, including
of the submitted document itself.

```go
// parseAppFeatureConfigOverride handles the two things that must be checked
// before a candidate override can even be considered for merging. Returns
// (nil, nil) when raw is empty or whitespace-only, meaning "clear the
// override" — this is the one case callers must treat specially (not an
// error, and not "store an empty file"). Does NOT run schema validation —
// that happens inside computeLayers (see "Runtime flow"), which validates
// this data as one of the layers it merges, plus the merged result.
func parseAppFeatureConfigOverride(raw string) ([]byte, error) {
    if strings.TrimSpace(raw) == "" {
        return nil, nil
    }

    data := []byte(raw)

    n, err := countYAMLDocuments(data)
    if err != nil {
        return nil, err
    }
    if n > 1 {
        return nil, apierrors.ValidationFailed.New("app_feature_config_yaml must contain at most one YAML document")
    }

    return data, nil
}

// countYAMLDocuments counts YAML documents separated by "---". Malformed
// YAML is not an error here — countYAMLDocuments returns whatever count it
// reached and nil; computeLayers (via ParseFeatureConfigWithoutDefaults)
// produces the canonical validation error for malformed input.
func countYAMLDocuments(data []byte) (int, error) {
    dec := goyaml.NewDecoder(bytes.NewReader(data))
    count := 0
    for {
        var doc any
        if err := dec.Decode(&doc); err != nil {
            if errors.Is(err, io.EOF) {
                return count, nil
            }
            return count, nil
        }
        count++
    }
}
```

`goyaml` is `gopkg.in/yaml.v3` (already an indirect dependency per `go.mod`;
this makes it direct). Used only for document counting — parsing the config
itself still goes through `config.ParseFeatureConfigWithoutDefaults` inside
`computeLayers` (`sigs.k8s.io/yaml` + the real JSON schema), so the
`causes`/`location` contract above is unaffected. Multi-document input never
reaches that point: `sigs.k8s.io/yaml.YAMLToJSON` silently keeps only the
first document rather than erroring, so this guard must run first.

---

## Runtime flow

### `GetAppFeatureConfig(ctx, appID string) (*AppFeatureConfigResult, error)`

1. `s.GlobalDatabase.ReadOnly(ctx, func(ctx) error { ... })`
2. Inside: `dbs, e := s.ConfigSourceStore.GetDatabaseSourceByAppID(ctx, appID)`.
   `errors.Is(e, configsource.ErrAppNotFound)` → return
   `apierrors.NotFound.WithReason("AppNotFound").New("app not found")` (same
   pattern as `PlanService.ChangeAppPlan`, `pkg/siteadmin/service/plan.go:84`).
3. `overrideYAML := dbs.Data[featureConfigOverrideKey()]` (nil if absent — `""`
   when stringified, matching the spec).
4. `planEffective, effective, e := s.computeLayers(ctx, dbs.PlanName, overrideYAML)`.
5. Outside the transaction: build
   `&AppFeatureConfigResult{PlanName: dbs.PlanName, EffectivePlanFeatureConfig:
   planEffective, AppFeatureConfigYAML: string(overrideYAML), EffectiveAppFeatureConfig:
   effective}`.

### `UpdateAppFeatureConfig(ctx, appID string, rawYAML string) (*AppFeatureConfigResult, error)`

Validate-via-merge happens **before** writing, so an invalid override is
never persisted even transiently (and so a failed write attempt is never
even issued):

1. `overrideBytes, err := parseAppFeatureConfigOverride(rawYAML)` — **outside**
   any DB transaction; only the empty/whitespace-clear case and the
   multi-document guard, no schema validation yet (see "Validation").
2. `s.GlobalDatabase.WithTx(ctx, func(ctx) error { ... })`
3. Inside:
   a. `dbs, e := s.ConfigSourceStore.GetDatabaseSourceByAppID(ctx, appID)` →
      404 `AppNotFound` on `configsource.ErrAppNotFound`, same as GET.
   b. `key := featureConfigOverrideKey()`; `oldYAML := string(dbs.Data[key])`.
   c. **`planEffective, effective, e := s.computeLayers(ctx, dbs.PlanName, overrideBytes)`
      — this is the actual schema validation** (see "Validation error
      contract"). Return `e` as-is on failure (400 `ValidationFailed` /
      `causes`); nothing below this line runs, and the transaction rolls
      back.
   d. If `dbs.Data == nil`, initialize `dbs.Data = map[string][]byte{}` (the
      real DB rows always unmarshal to a non-nil map from the `NOT NULL`
      jsonb column, but this is a defensive guard against assigning to a nil
      map panicking).
   e. If `overrideBytes == nil` (parsed as "clear"): `delete(dbs.Data, key)`.
      Else: `dbs.Data[key] = overrideBytes`.
   f. `dbs.UpdatedAt = s.Clock.NowUTC()`.
   g. `s.ConfigSourceStore.UpdateDatabaseSource(ctx, dbs)` — the plain SQL
      `UPDATE` that fires the `notify_config_source_change` trigger;
      propagation needs no other action.
4. Outside the transaction, if `s.AuditService != nil`: emit
   `*siteadminauditlog.AppFeatureConfigUpdatedPayload{AppID: appID, OldAppFeatureConfigYAML:
   oldYAML, NewAppFeatureConfigYAML: string(overrideBytes)}` via
   `s.AuditService.LogEvent(ctx, appID, payload)`. Log-and-continue on error
   (never fail the mutation because audit logging failed) — same pattern as
   `PlanService.ChangeAppPlan`.
5. Return `&AppFeatureConfigResult{...}` using `planEffective`/`effective`
   from step 3c (already computed, no need to recompute) and `overrideBytes`
   instead of the pre-write `overrideYAML`.

### `PreviewAppFeatureConfig(ctx, appID string, rawYAML string) (*AppFeatureConfigResult, error)`

1. `overrideBytes, err := parseAppFeatureConfigOverride(rawYAML)` — same as
   PUT step 1.
2. `s.GlobalDatabase.ReadOnly(ctx, func(ctx) error { ... })`
3. Inside: `dbs, e := s.ConfigSourceStore.GetDatabaseSourceByAppID(ctx, appID)`
   → 404 `AppNotFound` on `configsource.ErrAppNotFound`, same as GET.
   `planEffective, effective, e := s.computeLayers(ctx, dbs.PlanName, overrideBytes)`
   — same validation, same error shape as PUT step 3c. Return `e` as-is on
   failure.
4. **No write.** Return `&AppFeatureConfigResult{PlanName: dbs.PlanName,
   EffectivePlanFeatureConfig: planEffective, AppFeatureConfigYAML:
   string(overrideBytes), EffectiveAppFeatureConfig: effective}`.

---

## Service layer — file-level plan

### New file: `pkg/siteadmin/service/feature_config.go`

```go
package service

import (
    "bytes"
    "context"
    "errors"
    "io"
    "strings"

    goyaml "gopkg.in/yaml.v3"

    "github.com/authgear/authgear-server/pkg/api/apierrors"
    "github.com/authgear/authgear-server/pkg/api/event"
    "github.com/authgear/authgear-server/pkg/lib/config"
    "github.com/authgear/authgear-server/pkg/lib/config/configsource"
    "github.com/authgear/authgear-server/pkg/lib/config/plan"
    "github.com/authgear/authgear-server/pkg/portal/deps"
    siteadminauditlog "github.com/authgear/authgear-server/pkg/siteadmin/auditlog"
    "github.com/authgear/authgear-server/pkg/util/clock"
    "github.com/authgear/authgear-server/pkg/util/filepathutil"
    "github.com/authgear/authgear-server/pkg/util/resource"
)

// ---- Narrow interfaces -------------------------------------------------------

type FeatureConfigServiceGlobalDatabase interface {
    WithTx(ctx context.Context, do func(ctx context.Context) error) error
    ReadOnly(ctx context.Context, do func(ctx context.Context) error) error
}

type FeatureConfigServicePlanStore interface {
    GetPlan(ctx context.Context, name string) (*plan.Plan, error)
}

type FeatureConfigServiceConfigSourceStore interface {
    GetDatabaseSourceByAppID(ctx context.Context, appID string) (*configsource.DatabaseSource, error)
    UpdateDatabaseSource(ctx context.Context, dbs *configsource.DatabaseSource) error
}

type FeatureConfigServiceAuditService interface {
    LogEvent(ctx context.Context, appID string, payload event.NonBlockingPayload) error
}

// ---- Domain types -------------------------------------------------------------

// AppFeatureConfigResult is the internal computation result — pointer-typed
// because config.FeatureConfig sections are pointer fields throughout
// pkg/lib/config. Transport dereferences EffectivePlanFeatureConfig /
// EffectiveAppFeatureConfig into the (value-typed, generated) wire schema.
type AppFeatureConfigResult struct {
    PlanName                   string
    EffectivePlanFeatureConfig *config.FeatureConfig
    AppFeatureConfigYAML       string
    EffectiveAppFeatureConfig     *config.FeatureConfig
}

// ---- FeatureConfigService -------------------------------------------------------

type FeatureConfigService struct {
    GlobalDatabase    FeatureConfigServiceGlobalDatabase
    PlanStore         FeatureConfigServicePlanStore
    ConfigSourceStore FeatureConfigServiceConfigSourceStore
    AuditService      FeatureConfigServiceAuditService
    BaseResources     deps.AppBaseResources // already provided by deps.DependencySet, no new wiring
    Clock             clock.Clock
}

func (s *FeatureConfigService) GetAppFeatureConfig(ctx context.Context, appID string) (*AppFeatureConfigResult, error) {
    // ... (see "Runtime flow" above)
}

func (s *FeatureConfigService) UpdateAppFeatureConfig(ctx context.Context, appID string, rawYAML string) (*AppFeatureConfigResult, error) {
    // ... (see "Runtime flow" above)
}

func (s *FeatureConfigService) PreviewAppFeatureConfig(ctx context.Context, appID string, rawYAML string) (*AppFeatureConfigResult, error) {
    // ... (see "Runtime flow" above)
}

// ---- Merge computation (see "Merge computation" section above for full bodies) --

func featureConfigOverrideKey() string { /* ... */ }
func (s *FeatureConfigService) computeLayers(ctx context.Context, planName string, overrideYAML []byte) (*config.FeatureConfig, *config.FeatureConfig, error) { /* ... */ }
func readEffectiveFeatureConfig(ctx context.Context, fs []resource.Fs) (*config.FeatureConfig, error) { /* ... */ }

// ---- Validation (see "Validation" section above for full bodies) ---------------

func parseAppFeatureConfigOverride(raw string) ([]byte, error) { /* ... */ }
func countYAMLDocuments(data []byte) (int, error) { /* ... */ }
```

(The plan repeats the full bodies from the sections above rather than eliding
them in the actual implementation — this file listing is grouped by concern
for readability, not to be copy-pasted with the elisions left in.)

### New file: `pkg/siteadmin/auditlog/app_feature_config_updated.go`

Copy `app_plan_updated.go` verbatim (same `event.NonBlockingPayload` method
set: `NonBlockingEventType`, `UserID` → `""`, `GetTriggeredBy` →
`event.TriggeredBySiteAdmin`, `FillContext` → no-op, `ForHook` → `false`,
`ForAudit` → `true`, `RequireReindexUserIDs`/`DeletedUserIDs` → `nil`), only
changing:

```go
const AppFeatureConfigUpdated event.Type = "site_admin.app.feature_config.updated"

type AppFeatureConfigUpdatedPayload struct {
    AppID                   string `json:"app_id"`
    OldAppFeatureConfigYAML string `json:"old_app_feature_config_yaml"`
    NewAppFeatureConfigYAML string `json:"new_app_feature_config_yaml"`
}
```

Storing full old/new YAML documents (not a diff): documents are small (a few
dozen fields at most) and contain no end-user PII, only per-app operational
settings.

### `pkg/siteadmin/service/deps.go`

Add:
```go
wire.Struct(new(FeatureConfigService), "*"),
```

---

## Transport layer — file-level plan

Three new handler files, all following the existing
`handler_app_plan_change.go`-style shape. GET has no request body; PUT and
POST are structurally identical (envelope-validate → call service → same
response builder) and differ only in the columns below.

### New file: `pkg/siteadmin/transport/handler_app_feature_config_get.go`

```go
package transport

import (
    "context"
    "net/http"

    "github.com/authgear/authgear-server/pkg/api/siteadmin"
    "github.com/authgear/authgear-server/pkg/siteadmin/service"
    "github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureAppFeatureConfigGetRoute(route httproute.Route) httproute.Route {
    return route.WithMethods("OPTIONS", "GET").
        WithPathPattern("/api/v1/apps/:appID/feature-config")
}

type AppFeatureConfigGetService interface {
    GetAppFeatureConfig(ctx context.Context, appID string) (*service.AppFeatureConfigResult, error)
}

type AppFeatureConfigGetHandler struct {
    Service AppFeatureConfigGetService
}

func (h *AppFeatureConfigGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    appID := httproute.GetParam(r, "appID")

    result, err := h.Service.GetAppFeatureConfig(r.Context(), appID)
    if err != nil {
        writeError(w, r, err)
        return
    }

    SiteAdminAPISuccessResponse{Body: featureConfigResultToResponse(result)}.WriteTo(w)
}

// featureConfigResultToResponse is shared by the get/update/preview handlers
// (defined here since this is the first of the three files registered).
func featureConfigResultToResponse(result *service.AppFeatureConfigResult) siteadmin.AppFeatureConfigResponse {
    return siteadmin.AppFeatureConfigResponse{
        PlanName:                   result.PlanName,
        EffectivePlanFeatureConfig: *result.EffectivePlanFeatureConfig,
        AppFeatureConfigYaml:       result.AppFeatureConfigYAML,
        EffectiveAppFeatureConfig:     *result.EffectiveAppFeatureConfig,
    }
}
```

### New files: `handler_app_feature_config_update.go` (PUT) and `handler_app_feature_config_preview.go` (POST)

Canonical shape (shown once — both files are this same template, substitute
the column values below):

```go
package transport

import (
    "context"
    "encoding/json"
    "net/http"

    "github.com/authgear/authgear-server/pkg/api/siteadmin"
    "github.com/authgear/authgear-server/pkg/siteadmin/service"
    "github.com/authgear/authgear-server/pkg/util/httproute"
    "github.com/authgear/authgear-server/pkg/util/validation"
)

func Configure<Route>(route httproute.Route) httproute.Route {
    return route.WithMethods(<methods>).
        WithPathPattern(<path>)
}

var <RequestSchema> = validation.NewSimpleSchema(`
{
    "type": "object",
    "properties": { "app_feature_config_yaml": { "type": "string" } },
    "required": ["app_feature_config_yaml"]
}
`)

type <Service> interface {
    <Method>(ctx context.Context, appID string, rawYAML string) (*service.AppFeatureConfigResult, error)
}

type <Handler> struct {
    Service <Service>
}

type <Params> struct {
    AppID string
    siteadmin.<RequestType>
}

func parse<Params>(r *http.Request) (<Params>, error) {
    var body siteadmin.<RequestType>
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        return <Params>{}, err
    }
    if err := <RequestSchema>.Validator().ValidateValue(r.Context(), body); err != nil {
        return <Params>{}, err
    }
    return <Params>{AppID: httproute.GetParam(r, "appID"), <RequestType>: body}, nil
}

func (h *<Handler>) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    params, err := parse<Params>(r)
    if err != nil {
        writeError(w, r, err)
        return
    }
    result, err := h.Service.<Method>(r.Context(), params.AppID, params.AppFeatureConfigYaml)
    if err != nil {
        writeError(w, r, err)
        return
    }
    SiteAdminAPISuccessResponse{Body: featureConfigResultToResponse(result)}.WriteTo(w)
}
```

| Placeholder | PUT (`handler_app_feature_config_update.go`) | POST (`handler_app_feature_config_preview.go`) |
|---|---|---|
| `<Route>`/file symbol prefix | `AppFeatureConfigUpdateRoute` | `AppFeatureConfigPreviewRoute` |
| `<methods>` | `"PUT"` (GET already registered `OPTIONS` for this shared path — see routes.go) | `"OPTIONS", "POST"` (separate path, declares its own) |
| `<path>` | `/api/v1/apps/:appID/feature-config` | `/api/v1/apps/:appID/feature-config/preview` |
| `<RequestSchema>` | `AppFeatureConfigUpdateRequestSchema` | `AppFeatureConfigPreviewRequestSchema` |
| `<Service>`/`<Method>` | `AppFeatureConfigUpdateService` / `UpdateAppFeatureConfig` | `AppFeatureConfigPreviewService` / `PreviewAppFeatureConfig` |
| `<Handler>` | `AppFeatureConfigUpdateHandler` | `AppFeatureConfigPreviewHandler` |
| `<Params>` | `AppFeatureConfigUpdateParams` | `AppFeatureConfigPreviewParams` |
| `<RequestType>` | `siteadmin.UpdateAppFeatureConfigRequest` | `siteadmin.PreviewAppFeatureConfigRequest` |

`<RequestSchema>` only validates the request envelope shape
(`app_feature_config_yaml` is a present string) — it is **not** where the
feature-config document itself gets validated. That happens inside
`s.computeLayers` (via `parseAppFeatureConfigOverride` for the
empty/multi-document checks, then `viewEffectiveResource` for schema
validation during the merge), which is what produces the `causes`/`location`
errors described above.

`configsource.ErrAppNotFound` is mapped to 404 **inside the service** for all
three methods (matching `PlanService.ChangeAppPlan`'s convention), so no
transport-layer `errors.Is(err, configsource.ErrAppNotFound)` branch is
needed here.

### `pkg/siteadmin/transport/deps.go`

Add:
```go
wire.Struct(new(AppFeatureConfigGetHandler), "*"),
wire.Struct(new(AppFeatureConfigUpdateHandler), "*"),
wire.Struct(new(AppFeatureConfigPreviewHandler), "*"),
```

---

## DI wiring — file-level plan

### `pkg/siteadmin/deps.go`

Add, in the "siteadmin service layer" section (alongside the existing
`PlanService*` bindings):

```go
// FeatureConfigService bindings
wire.Bind(new(siteadminservice.FeatureConfigServiceGlobalDatabase), new(*globaldb.Handle)),
wire.Bind(new(siteadminservice.FeatureConfigServicePlanStore), new(*plan.Store)),
wire.Bind(new(siteadminservice.FeatureConfigServiceConfigSourceStore), new(*configsource.Store)),
wire.Bind(new(siteadminservice.FeatureConfigServiceAuditService), new(*siteadminservice.SiteAdminAuditService)),

// transport bindings
wire.Bind(new(transport.AppFeatureConfigGetService), new(*siteadminservice.FeatureConfigService)),
wire.Bind(new(transport.AppFeatureConfigUpdateService), new(*siteadminservice.FeatureConfigService)),
wire.Bind(new(transport.AppFeatureConfigPreviewService), new(*siteadminservice.FeatureConfigService)),
```

`*globaldb.Handle`, `*plan.Store`, `*configsource.Store`, and
`*siteadminservice.SiteAdminAuditService` are all already constructed
elsewhere in this same `DependencySet` (used by `PlanService`); no new
concrete providers are needed, only the additional `wire.Bind` lines above.

`FeatureConfigService.BaseResources` needs **no wiring change at all**:
its field type is `deps.AppBaseResources`, already provided by
`deps.DependencySet` (imported at `pkg/siteadmin/deps.go:45`) — wire resolves
`wire.Struct(new(FeatureConfigService), "*")` fields by type, so this one is
satisfied automatically alongside the explicit binds above.

### `pkg/siteadmin/wire.go`

Add three injectors, one per handler, same template as every existing
injector in this file:

```go
func newAppFeatureConfig<X>Handler(p *deps.RequestProvider) http.Handler {
    panic(wire.Build(
        DependencySet,
        wire.Bind(new(http.Handler), new(*transport.AppFeatureConfig<X>Handler)),
    ))
}
```

`<X>` ∈ `{Get, Update, Preview}`.

### `pkg/siteadmin/routes.go`

Add, after `ConfigureAppPlanChangeRoute` (GET registered before PUT so GET's
route owns the shared path's `OPTIONS` registration):

```go
router.Add(transport.ConfigureAppFeatureConfigGetRoute(route), p.Handler(newAppFeatureConfigGetHandler))
router.Add(transport.ConfigureAppFeatureConfigUpdateRoute(route), p.Handler(newAppFeatureConfigUpdateHandler))
router.Add(transport.ConfigureAppFeatureConfigPreviewRoute(route), p.Handler(newAppFeatureConfigPreviewHandler))
```

### Spec-to-code generation

`pkg/api/siteadmin/gen.go` has not been regenerated yet for this feature —
`make generate` (Stage 2 of `new-siteadmin-api`) must run before any handler
code referencing the new generated types will compile. Verify after
generation that:

- `pkg/api/siteadmin/gen.go` contains a `FeatureConfig` type alias to
  `config.FeatureConfig`, matching the existing `APIError` alias pattern.
- `AppFeatureConfigResponse`'s `EffectivePlanFeatureConfig` /
  `EffectiveAppFeatureConfig` fields generate as **value** types (`FeatureConfig`,
  not `*FeatureConfig`).
- `ValidationErrorCause` generates as a plain struct with `Location`,
  `Keyword`, `Details` fields (or equivalent per the generator's naming
  convention) — it isn't referenced directly by any Go code in this plan (the
  `causes` array is built generically by `pkg/api/apierrors`), it exists
  purely for API-consumer documentation.

If `x-go-type` codegen for `FeatureConfig` doesn't produce a clean alias,
adjust the **spec** (not the generated file).

After the spec regeneration and the `deps.go`/`wire.go` changes above:
```bash
make generate
go generate ./pkg/siteadmin/...
go mod tidy   # adds gopkg.in/yaml.v3 as a direct dependency
go build ./pkg/siteadmin/...
go build ./cmd/portal/...
```

`wire_gen.go` and `pkg/api/siteadmin/gen.go` must never be hand-edited — only
regenerated by the commands above.

---

## Test plan

### Unit tests — `pkg/siteadmin/service/feature_config_test.go` (new, package
`service`, Convey BDD style — matches `plan_test.go`, `collaborator_test.go`)

Reuse `fakeDatabase{}` from `app_test.go` (already provides `WithTx`/`ReadOnly`
no-ops) and `fakeAuditService{}` from `plan_test.go`.

New fakes needed:

```go
type fakeFeatureConfigPlanStore struct {
    plans map[string]*plan.Plan
}

func (f *fakeFeatureConfigPlanStore) GetPlan(_ context.Context, name string) (*plan.Plan, error) {
    if p, ok := f.plans[name]; ok {
        return p, nil
    }
    return nil, plan.ErrPlanNotFound
}

type fakeFeatureConfigConfigSourceStore struct {
    sources map[string]*configsource.DatabaseSource
    updated *configsource.DatabaseSource
}

func (f *fakeFeatureConfigConfigSourceStore) GetDatabaseSourceByAppID(_ context.Context, appID string) (*configsource.DatabaseSource, error) {
    if s, ok := f.sources[appID]; ok {
        cp := *s
        cp.Data = maps.Clone(s.Data) // deep-enough copy so mutation in the service doesn't corrupt the fixture between test cases
        return &cp, nil
    }
    return nil, configsource.ErrAppNotFound
}

func (f *fakeFeatureConfigConfigSourceStore) UpdateDatabaseSource(_ context.Context, dbs *configsource.DatabaseSource) error {
    f.updated = dbs
    return nil
}
```

Two field-level fixtures, chosen for structural variety:

- **`collaborator`** (two independent `*int` fields, `pkg/lib/config/feature_collaborator.go`).
  Plan sets `maximum = 3`; override sets only `soft_maximum = 1` → effective
  `maximum` stays **3** (inherited from the plan), `soft_maximum` is `1`.
- **`oauth.client`** (four independent fields of mixed `*int`/`*bool` type,
  `pkg/lib/config/feature_oauth.go`). Plan sets `maximum = 10`; override sets
  only `custom_ui_enabled = true` → effective `maximum` stays **10**,
  `custom_ui_enabled` is `true`.

**Test cases:**

| Method | Case | Assertion |
|---|---|---|
| `GetAppFeatureConfig` | app has no override | `AppFeatureConfigYAML == ""`, `EffectiveAppFeatureConfig` equals plan-only fold |
| `GetAppFeatureConfig` | app has an override with a comment | `AppFeatureConfigYAML` matches the stored bytes **verbatim, including the comment** (not re-serialized) |
| `GetAppFeatureConfig` | `collaborator` fixture | effective `maximum` inherited from plan, `soft_maximum` from override |
| `GetAppFeatureConfig` | `oauth.client` fixture | effective `maximum` inherited from plan, `custom_ui_enabled` from override |
| `GetAppFeatureConfig` | app references a plan that does not exist (`plan.ErrPlanNotFound`) | no error; `EffectivePlanFeatureConfig` equals server defaults |
| `GetAppFeatureConfig` | app does not exist | `apierrors.NotFound` with condition `Kind.Name == apierrors.NotFound` |
| `UpdateAppFeatureConfig` | valid YAML | stored bytes in `csStore.updated.Data[key]` match input exactly (byte-for-byte, including comments/key order); response reflects new effective config |
| `UpdateAppFeatureConfig` | empty string clears an existing override | `csStore.updated.Data` no longer contains the key (not present at all, not an empty-string value) |
| `UpdateAppFeatureConfig` | whitespace-only string (`"  \n\t"`) clears | same as above |
| `UpdateAppFeatureConfig` | invalid YAML — unknown top-level key (schema `additionalProperties: false` violation) | `errors.As(err, &aggErr)` where `aggErr` is `*validation.AggregatedError`; `aggErr.Errors[0].Location` is the JSON pointer to the offending key (verify the actual value the schema validator produces rather than assuming); `csStore.updated` stays nil |
| `UpdateAppFeatureConfig` | invalid YAML — field-level type error, e.g. `oauth: {client: {maximum: "not-a-number"}}` | `errors.As(err, &aggErr)`; `aggErr.Errors[0].Location == "/oauth/client/maximum"`, `aggErr.Errors[0].Keyword == "type"`; `csStore.updated` stays nil. **This is the test that pins the JSON-pointer contract** — if the schema validator's location format ever changes, this test catches it before the frontend does. |
| `UpdateAppFeatureConfig` | multi-document YAML (`"a: 1\n---\nb: 2\n"`) | `apierrors.IsAPIErrorWithCondition` with `Kind.Name == apierrors.Invalid` and `Kind.Reason == "ValidationFailed"`; error does **not** satisfy `errors.As(err, &aggErr)` (no `causes` present); `csStore.updated` stays nil |
| `UpdateAppFeatureConfig` | app does not exist | `apierrors.NotFound`; `csStore.updated` stays nil |
| `UpdateAppFeatureConfig` | emits audit log with old and new YAML | `fakeAuditService.logged[0]` is `*siteadminauditlog.AppFeatureConfigUpdatedPayload` with correct `AppID`/`OldAppFeatureConfigYAML`/`NewAppFeatureConfigYAML` |
| `UpdateAppFeatureConfig` | audit failure does not affect mutation result | `AuditService` nil → still succeeds |
| `PreviewAppFeatureConfig` | valid candidate | response matches what PUT would have produced, but `csStore.updated` stays nil (nothing persisted) |
| `PreviewAppFeatureConfig` | invalid candidate (field-level type error) | same `Location`/`Keyword` assertions as PUT's field-level-type-error case; `csStore.updated` stays nil |
| `PreviewAppFeatureConfig` | app does not exist | `apierrors.NotFound` |
| `countYAMLDocuments` | 0 docs (empty input) | `0, nil` |
| `countYAMLDocuments` | 1 doc | `1, nil` |
| `countYAMLDocuments` | 2 docs separated by `---` | `2, nil` |

### e2e — `e2e/tests/siteadmin/feature_config.test.yaml` (new)

Follows the structure of `plans.test.yaml` exactly (same `generate_token` /
`get_access_token` steps, same `http://127.0.0.1:4003` base URL). Uses a
**dedicated** plan and app (not `e2e-siteadmin-app-alpha/beta/gamma`, which
`apps.test.yaml`/`plans.test.yaml` depend on staying in known states) to
avoid cross-test interference.

New fixture file `e2e/tests/siteadmin/fixture/seed_feature_config.sql`:
```sql
-- Dedicated plan + app for feature config e2e tests, isolated from the
-- apps/plans/collaborators fixtures so this test can freely mutate the
-- app's feature config override without affecting other suites.
INSERT INTO _portal_plan (id, name, feature_config, created_at, updated_at)
VALUES (
    gen_random_uuid()::text,
    'e2e-feature-config-plan',
    '{"collaborator": {"maximum": 3}, "oauth": {"client": {"maximum": 10}}}',
    NOW(), NOW()
)
ON CONFLICT (name) DO NOTHING;

INSERT INTO _portal_config_source (id, app_id, data, plan_name, created_at, updated_at)
VALUES (
    gen_random_uuid()::text,
    'e2e-siteadmin-app-feature-config',
    '{}',
    'e2e-feature-config-plan',
    NOW(), NOW()
)
ON CONFLICT (app_id) DO NOTHING;
```

`before:` block adds this fixture after the shared actor/apps seeds (same
`before:` shape as `plans.test.yaml`).

Steps (all against `http://127.0.0.1:4003/api/v1/apps/e2e-siteadmin-app-feature-config/feature-config...`):

1. `get_initial` — GET, expect 200, `plan_name: "e2e-feature-config-plan"`,
   `app_feature_config_yaml: ""`, `effective_plan_feature_config.collaborator.maximum:
   3`, `effective_app_feature_config.collaborator.maximum: 3` (no override yet).
2. `preview_candidate` — POST `/preview` with body
   `{"app_feature_config_yaml": "collaborator:\n  soft_maximum: 1\n"}`, expect
   200, `effective_app_feature_config.collaborator.maximum: 3` (inherited from
   the plan) and `effective_app_feature_config.collaborator.soft_maximum: 1`.
   Assert via `query` that `_portal_config_source.data` for this app is still
   `{}` (nothing persisted by preview).
3. `preview_invalid_returns_causes_with_location` — POST `/preview` with body
   `{"app_feature_config_yaml": "oauth:\n  client:\n    maximum: \"not-a-number\"\n"}`,
   expect 400, `error.reason: "ValidationFailed"`,
   `error.info.causes[0].location: "/oauth/client/maximum"`,
   `error.info.causes[0].kind: "type"`. This is the e2e-level pin of the
   JSON-pointer contract — it must be exercised over the real HTTP/JSON wire,
   not just asserted in Go unit tests.
4. `update_override` — PUT with the same body as step 2, expect 200, same
   `effective_app_feature_config` shape as the preview response.
5. `verify_override_stored` — plain SQL `query` action:
   `SELECT data->>'YXV0aGdlYXIuZmVhdHVyZXMueWFtbA==' AS override FROM
   _portal_config_source WHERE app_id = 'e2e-siteadmin-app-feature-config'`
   (the exact base64 key is computed at plan-writing time from
   `filepathutil.EscapePath("authgear.features.yaml")` — the implementer must
   verify the exact encoding `EscapePath` produces and substitute it here;
   `filepathutil.EscapePath` is `pkg/util/filepathutil/escape.go:11`), asserts
   the stored value equals the submitted YAML byte-for-byte.
6. `audit_update` — `audit_query` action (same shape as `plans.test.yaml`'s
   `audit_change_plan` step), asserting one row with `activity_type =
   'site_admin.app.feature_config.updated'`,
   `data->'payload'->>'app_id' = 'e2e-siteadmin-app-feature-config'`, and
   `data->'payload'->>'new_app_feature_config_yaml'` containing the submitted
   text.
7. `clear_override` — PUT with `{"app_feature_config_yaml": ""}`, expect 200,
   `app_feature_config_yaml: ""` in the response.
8. `verify_override_cleared` — SQL `query` asserting the override key is
   **absent** from `_portal_config_source.data` for this app (not merely an
   empty string) — e.g. `SELECT data ? '<key>' AS has_key FROM
   _portal_config_source WHERE app_id = '...'` expecting `has_key: false`
   (Postgres jsonb `?` containment operator).
9. `update_invalid_yaml_returns_400_no_causes` — PUT with
   `{"app_feature_config_yaml": "a: 1\n---\nb: 2\n"}` (multi-document), expect
   400, `error.reason: "ValidationFailed"`; do not assert `causes` is present
   here, since it deliberately isn't (the negative case for the hand-built
   exception in the validation error contract).
10. `update_non_existent_app_returns_404` — PUT to
    `.../apps/non-existent-app/feature-config`, expect 404.
11. `get_non_existent_app_returns_404` — GET on the same non-existent app,
    expect 404.
12. `preview_non_existent_app_returns_404` — POST `/preview` on the same
    non-existent app, expect 404.

---

## Fixed behavioral decisions

- Whitespace-only is defined as `strings.TrimSpace(raw) == ""` (covers empty
  string, spaces, tabs, newlines). Treated identically to the empty string:
  clears the override (deletes the map key), never stores a whitespace-only
  file.
- Multi-document YAML is rejected with 400 `ValidationFailed` in
  `parseAppFeatureConfigOverride`, before `computeLayers` ever runs, using a
  document count via `gopkg.in/yaml.v3`, because
  `sigs.k8s.io/yaml.YAMLToJSON` silently keeps only the first document
  otherwise. This is the **one** validation failure in this endpoint that
  does not carry `info.causes`.
- Every other validation failure (schema violations inside the document, or
  in the merged result) surfaces from **`computeLayers`**, which is the only
  place `config.ParseFeatureConfigWithoutDefaults` is called for this
  endpoint (via `viewEffectiveResource`, twice: once per layer during the
  fold, once more on the merged YAML). It produces `info.causes` with RFC
  6901 `location` pointers automatically via the existing
  `pkg/api/apierrors` conversion — **no new error-formatting code is written
  for this endpoint, and there is no separate standalone pre-validation
  step.**
- For `UpdateAppFeatureConfig`, `computeLayers` runs inside the transaction,
  after fetching `dbs` but before any mutation — an invalid override is
  rejected (and the transaction rolled back) before `dbs.Data` is touched or
  `UpdateDatabaseSource` is called.
- `configsource.ErrAppNotFound` is mapped to 404 inside `FeatureConfigService`
  (not in transport), for all three methods, for consistency with
  `PlanService.ChangeAppPlan` and because `UpdateAppFeatureConfig` must do
  this mapping inside its transaction regardless.
- `plan.ErrPlanNotFound` is tolerated (empty plan layer) in all three
  methods, never surfaced as an error to the caller.
- Audit payload carries the full old and new YAML documents, not a diff —
  both are always small.
- `dbs.UpdatedAt` is still bumped on every successful PUT even though it is
  not the propagation mechanism — kept for consistency with `ChangeAppPlan`
  and because `_portal_config_source` rows are otherwise inspected by admins
  expecting `updated_at` to reflect the latest write.
- No optimistic locking / `expected_updated_at` — matches the rest of the
  Site Admin API.
- No cross-validation of the new feature config against the app's current
  `authgear.yaml` — matches existing plan-change behavior.
- No new per-endpoint request body size limit — already generically capped
  at 1 MiB by `newBodyLimitMiddleware` in `pkg/siteadmin/routes.go`'s
  `rootChain`.
- No enum update needed anywhere for the new
  `site_admin.app.feature_config.updated` activity type: the audit read path
  filters `WHERE activity_type LIKE 'site_admin.%'` (prefix, not
  allow-list); the Admin API GraphQL `AuditLogActivityType` enum never
  included `site_admin.*` values.

---

## Implementation order

1. OpenAPI spec regeneration: `make generate` (produces `pkg/api/siteadmin/gen.go`
   with the new types — the spec itself, `docs/api/siteadmin-api.yaml`, is
   already written and is not part of this implementation plan's scope to
   change further unless generation surfaces a problem).
2. Service layer: `pkg/siteadmin/service/feature_config.go` +
   `pkg/siteadmin/service/feature_config_test.go` + `pkg/siteadmin/service/deps.go`.
   Fully testable in isolation with fakes; no transport or DI changes needed
   yet.
3. Audit payload: `pkg/siteadmin/auditlog/app_feature_config_updated.go`
   (small, no dependencies on the rest).
4. Transport handler scaffolding: the three new handler files, with full
   `ServeHTTP` bodies.
5. DI wiring: `pkg/siteadmin/deps.go`, `pkg/siteadmin/wire.go`,
   `pkg/siteadmin/routes.go`, `pkg/siteadmin/transport/deps.go`, then
   `go generate ./pkg/siteadmin/...` to produce `wire_gen.go`.
6. `.vettedpositions` update for the new `r.Context()` call sites in the three
   new handler files.
7. e2e test + fixture.

---

## Atomic commit plan

| # | Commit message | Files | Verification |
|---|---|---|---|
| 1 | `[Site Admin API] Generate models for app feature config` | `pkg/api/siteadmin/gen.go` (regenerated, not hand-edited) | `go build ./pkg/api/siteadmin/...` · `make fmt` |
| 2 | `[Site Admin API] Add FeatureConfigService` | `pkg/siteadmin/service/feature_config.go`, `pkg/siteadmin/service/feature_config_test.go`, `pkg/siteadmin/service/deps.go`, `go.mod`/`go.sum` (yaml.v3 becomes direct) | `go test ./pkg/siteadmin/service/...` · `make fmt` |
| 3 | `[Site Admin API] Add site_admin.app.feature_config.updated audit payload` | `pkg/siteadmin/auditlog/app_feature_config_updated.go` | `go build ./pkg/siteadmin/...` · `make fmt` |
| 4 | `[Site Admin API] Add app feature config handlers` | `pkg/siteadmin/transport/handler_app_feature_config_get.go`, `handler_app_feature_config_update.go`, `handler_app_feature_config_preview.go`, `pkg/siteadmin/transport/deps.go` | `go build ./pkg/siteadmin/...` · `make fmt` |
| 5 | `[Site Admin API] Wire app feature config handlers` | `pkg/siteadmin/deps.go`, `pkg/siteadmin/wire.go`, `pkg/siteadmin/routes.go`, `pkg/siteadmin/wire_gen.go` (regenerated, not hand-edited) | `go generate ./pkg/siteadmin/...` · `go build ./pkg/siteadmin/... ./cmd/portal/...` · `make fmt` |
| 6 | `[Site Admin API] Add app feature config e2e tests` | `e2e/tests/siteadmin/feature_config.test.yaml`, `e2e/tests/siteadmin/fixture/seed_feature_config.sql` | `make -C e2e run` (or the targeted subset per `write-e2e-test`) |
| 7 | final: `.vettedpositions` + `make check-tidy` output, folded into commit 4 or its own trailing commit per the `new-siteadmin-api` skill's final-gate rule | `.vettedpositions`, any files `make check-tidy` reformats/regenerates | `go run ./devtools/goanalysis ./cmd/... ./pkg/...` · `make sort-vettedpositions` · `go run ./devtools/goanalysis ./cmd/... ./pkg/...` (must be clean) · `make check-tidy` (run once, last) |

Each commit builds and tests cleanly on its own before moving to the next,
per the `new-siteadmin-api` skill's commit-sequence rule.
