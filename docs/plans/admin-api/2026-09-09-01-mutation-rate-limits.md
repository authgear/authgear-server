# Admin API Mutation Rate Limits

Spec: [docs/specs/rate-limit.md — Admin API mutations](../../specs/rate-limit.md#admin-api-mutations).

## 1. Goal / Scope

Add one token bucket to the Admin API GraphQL endpoint, consumed once per top-level mutation field, configured in `authgear.features.yaml` only.

There is deliberately no per-IP bucket. The Admin API authenticates as the project, so `app_id` is the caller identity and the one dimension a caller cannot choose; IP would be a weaker proxy for it, sidesteppable by spreading the same key across hosts. DCR registration and the CIMD fetch do carry per-IP buckets, but both of those endpoints are unauthenticated, where IP is the only handle available — and the per-IP limits on the authentication flows exist to stop credential brute-forcing, which has no analogue here.

| Rate limit name | Scope | Default |
| --- | --- | --- |
| `admin_api.mutation.all.per_project` | Per project (`app_id`) | 1000 / minute |

**In scope:** GraphQL mutations on `POST /graphql` and `POST /_api/admin/graphql` of the Admin API server (`pkg/admin`), reached both directly with an Admin API key and through the portal's reverse proxy.

**Out of scope, deliberately:** GraphQL queries; the Admin API's REST endpoints (user import/export are bounded by `user_import_usage` / `user_export_usage`, presign upload by its fixed 10/hour); the portal's own GraphQL API at `/api/graphql` (`pkg/portal/graphql`), which is not app-scoped and so has no feature config to hang limits on. Both are recorded under the spec's Future Works.

### 1.1 Two findings from the codebase that shape this plan

**Mutation fields per request are already capped at 5.** `pkg/util/graphqlutil/handler.go:36` defines `maxMutationFieldsPerRequest = 5`, enforced by `validateMutationFieldCount` at line 150 before `graphql.Do`, with fragment-spread resolution and cycle detection already handled by `countTopLevelFields` (line 267). Consequences:

- The counting logic does not need to be written, only exposed. §3.1.
- `n` passed to the limiter is always in `[1, 5]`, so the "a document larger than `burst` can never succeed" trap is bounded: it only bites if a tier configures `burst < 5`. §8 D5.
- **The spec is currently wrong on this point** and must be corrected in the same PR: it says "a 100-field document costs 100 tokens, not 1", which cannot happen. §6.6.

**`ratelimit.Limiter` is already in the Admin API's wire graph** (`pkg/admin/wire_gen.go:680`), as is `adminAPIFeatureConfig` (line 243). No new providers are needed — only new fields on `transport.GraphQLHandler` and a regenerated `wire_gen.go`.

## 2. Config model and schema

> Read the `update-feature-config` skill before touching this file. Field-level merge is a hard requirement and `pkg/lib/config/testdata/merge_feature.yaml` coverage is mandatory.

All config changes are in **`pkg/lib/config/feature_admin_api.go`**. There is no `authgear.yaml` counterpart — the limit bounds what a project's own administrators can do to that project's storage, so the tenant must not be able to raise it. This is the CIMD `fetch` precedent, not the DCR one; see the spec's Feature Config Rate Limits section.

### 2.1 Structs and schema

`AdminAPIFeatureConfig` gains a `RateLimits` field, appended **after** `UserExportUsage` (struct order is cosmetic here — `TestParseFeatureConfig` unmarshals `testdata/default_feature.yaml` into a struct and compares with `ShouldResemble` (`feature_test.go:29-33`), so YAML key order does not matter — but keep it last for readability).

```go
var _ = FeatureConfigSchema.Add("AdminAPIFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"create_session_enabled": { "type": "boolean" },
		"user_import_usage": { "$ref": "#/$defs/UsageLimitConfig" },
		"user_export_usage": { "$ref": "#/$defs/UsageLimitConfig" },
		"rate_limits": { "$ref": "#/$defs/AdminAPIRateLimitsFeatureConfig" }
	}
}
`)

var _ = FeatureConfigSchema.Add("AdminAPIRateLimitsFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"mutation": { "$ref": "#/$defs/AdminAPIRateLimitsMutationFeatureConfig" }
	}
}
`)

var _ = FeatureConfigSchema.Add("AdminAPIRateLimitsMutationFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"all": { "$ref": "#/$defs/AdminAPIRateLimitsMutationScopeFeatureConfig" }
	}
}
`)

var _ = FeatureConfigSchema.Add("AdminAPIRateLimitsMutationScopeFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"per_project": { "$ref": "#/$defs/RateLimitConfig" }
	}
}
`)

type AdminAPIRateLimitsFeatureConfig struct {
	Mutation *AdminAPIRateLimitsMutationFeatureConfig `json:"mutation,omitempty"`
}

type AdminAPIRateLimitsMutationFeatureConfig struct {
	All *AdminAPIRateLimitsMutationScopeFeatureConfig `json:"all,omitempty"`
}

type AdminAPIRateLimitsMutationScopeFeatureConfig struct {
	PerProject *RateLimitConfig `json:"per_project,omitempty"`
}
```

`RateLimitConfig` is already registered in `FeatureConfigSchema` (`pkg/lib/config/rate_limit.go:19`), so `$ref` resolves with no new registration.

No plain-scalar fields are added, so the `omitempty`-on-a-real-default trap from the skill does not apply here — every new leaf is a pointer, where `nil` genuinely means "this layer said nothing".

### 2.2 Why the `all` level exists

`per_project` sits under a scope object rather than directly under `mutation` so that gating an individual mutation later — `admin_api.rate_limits.mutation.create_group.*`, yielding `admin_api.mutation.create_group.per_project` — needs no rename of anything shipped here. Every child of `mutation` is a scope; every child of a scope is a bucket. `all` is the reserved scope meaning "every mutation field". Same reasoning as `OAuthClientIDMetadataDocumentRateLimitsFeatureConfig` nesting under `fetch` (`feature_oauth.go:202-208`).

This plan does **not** build per-mutation scopes. When they are built they will be consumed in addition to `all`, not as a fallback — see spec, and §8 D6.

### 2.3 Defaults

```go
// SetDefaults mirrors OAuthClientIDMetadataDocumentRateLimitsFetchFeatureConfig's
// pattern: PerProject is already non-nil when this runs, because
// SetFieldDefaults force-allocates every pointer without a nullable tag, so
// checking Enabled == nil is what detects "no layer configured this bucket".
//
// The default is deliberately loose: a backstop against runaway or abusive
// volume, not a tuned throttle. Tiers are expected to set it far lower.
func (c *AdminAPIRateLimitsMutationScopeFeatureConfig) SetDefaults() {
	if c.PerProject.Enabled == nil {
		c.PerProject = &RateLimitConfig{
			Enabled: new(true),
			Period:  "1m",
			Burst:   1000,
		}
	}
}
```

Defaults are applied **once, after the merge fold**, not per layer: `AuthgearFeatureYAMLDescriptor.viewEffectiveResource` (`pkg/lib/config/configsource/resources.go:695`) parses each layer with `ParseFeatureConfigWithoutDefaults`, merges, marshals, then re-parses with `ParseFeatureConfig`, which runs `SetFieldDefaults`. A layer that omits `rate_limits` therefore really is `nil` during the merge.

No tier's values are hardcoded anywhere in the repo. Plan-tier tightening is an operator action on the plan records — §9.

### 2.4 Merge — field-level at every level

`AdminAPIFeatureConfig.Merge` today merges `CreateSessionEnabled` / `UserImportUsage` / `UserExportUsage` field-by-field. It gains one line:

```go
	merged.RateLimits = merged.RateLimits.Merge(layer.AdminAPI.RateLimits)
```

and three cascading `Merge` methods, each with the nil-guard prologue from `OAuthClientFeatureConfig.Merge`:

```go
func (c *AdminAPIRateLimitsFeatureConfig) Merge(layer *AdminAPIRateLimitsFeatureConfig) *AdminAPIRateLimitsFeatureConfig {
	if c == nil && layer == nil { return nil }
	if c == nil { return layer }
	if layer == nil { return c }
	c.Mutation = c.Mutation.Merge(layer.Mutation)
	return c
}

func (c *AdminAPIRateLimitsMutationFeatureConfig) Merge(layer *AdminAPIRateLimitsMutationFeatureConfig) *AdminAPIRateLimitsMutationFeatureConfig {
	if c == nil && layer == nil { return nil }
	if c == nil { return layer }
	if layer == nil { return c }
	c.All = c.All.Merge(layer.All)
	return c
}

// Merge replaces each bucket wholesale, not field-by-field: enabled/period/burst
// are one unit, and merging them field-wise would let two layers jointly produce
// a bucket neither one wrote.
func (c *AdminAPIRateLimitsMutationScopeFeatureConfig) Merge(layer *AdminAPIRateLimitsMutationScopeFeatureConfig) *AdminAPIRateLimitsMutationScopeFeatureConfig {
	if c == nil && layer == nil { return nil }
	if c == nil { return layer }
	if layer == nil { return c }
	if layer.PerProject != nil { c.PerProject = layer.PerProject }
	return c
}
```

`AdminAPIRateLimitsFeatureConfig` and `AdminAPIRateLimitsMutationFeatureConfig` each have a single field today, but the cascade must not stop there — `All` has two real siblings. This is the `Authenticator → Password → Policy` cascade rule from the skill.

**This cascade is what makes the app-level override in §9 work.** A file containing only `admin_api.rate_limits.mutation.all.per_project` must not reset `create_session_enabled` or the usage limits inherited from the plan layer. §7.1 tests exactly that.

### 2.5 Nil-safe accessors

Three accessors, each used exactly once in the §3.2 call flow:

```go
func (c *AdminAPIFeatureConfig) GetRateLimits() *AdminAPIRateLimitsFeatureConfig
func (c *AdminAPIRateLimitsFeatureConfig) GetMutation() *AdminAPIRateLimitsMutationFeatureConfig
func (c *AdminAPIRateLimitsMutationFeatureConfig) GetAll() *AdminAPIRateLimitsMutationScopeFeatureConfig
```

Each returns `nil` on a `nil` receiver, mirroring `OAuthClientIDMetadataDocumentFeatureConfig.GetRateLimits` (`feature_oauth.go:158`). At runtime the chain is always non-nil because `SetFieldDefaults` force-allocates it; the guards exist only for pre-defaults callers such as a test that unmarshals a YAML snippet directly. §3.2 still handles a `nil` result, because `RateLimitConfig.IsEnabled()` has no nil-receiver guard (`rate_limit.go:57`) and would panic.

## 3. Runtime flow

### 3.1 Exposing the mutation field count — `pkg/util/graphqlutil/handler.go`

The count is already computed and thrown away. Extract it:

```go
// CountTopLevelMutationFields returns the number of top-level fields in the
// document's selected operation when that operation is a mutation, and 0
// otherwise (a query, or an operation that cannot be resolved). Fragment
// spreads are expanded; a fragment cycle is an error.
func CountTopLevelMutationFields(query string, operationName string) (int, error)
```

Body is the current `validateMutationFieldCount` (line 209) with the `count > limit` check removed and `count` returned; the limit check moves to `ContextHandler` (§3.2) and `validateMutationFieldCount` is deleted. Externally observable behaviour — same 400, same message, same fragment-cycle error — is unchanged.

### 3.2 A pre-execution hook — `pkg/util/graphqlutil/handler.go`

The Admin API handler cannot count for itself: `NewRequestOptions` consumes `r.Body`, so parsing in `pkg/admin/transport` would mean buffering and re-attaching the body. Instead `Handler` gains an optional hook, called after options are parsed and validated and before `graphql.Do`:

```go
type Handler struct {
	Schema           *graphql.Schema
	ResultCallbackFn ResultCallbackFn
	// BeforeExecuteFn, when set, is called after the request options are
	// parsed and validated and before the operation executes.
	// mutationFieldCount is 0 for queries. Returning an error rejects the
	// request without executing it; the error is written with
	// writeErrorResponse, so its apierrors kind determines the status code.
	BeforeExecuteFn func(ctx context.Context, mutationFieldCount int) error
}
```

`ContextHandler` changes at line 150 from

```go
	if err := validateMutationFieldCount(opts.Query, opts.OperationName, maxMutationFieldsPerRequest); err != nil {
```

to computing the count once and reusing it:

```go
	mutationFieldCount, err := CountTopLevelMutationFields(opts.Query, opts.OperationName)
	if err != nil {
		writeErrorResponse(ctx, w, err)
		return
	}
	if mutationFieldCount > maxMutationFieldsPerRequest {
		writeErrorResponse(ctx, w, apierrors.NewBadRequest(
			fmt.Sprintf("too many mutation fields in one request: got %d, limit is %d",
				mutationFieldCount, maxMutationFieldsPerRequest)))
		return
	}
	if h.BeforeExecuteFn != nil {
		if err := h.BeforeExecuteFn(ctx, mutationFieldCount); err != nil {
			writeErrorResponse(ctx, w, err)
			return
		}
	}
```

`validateMutationFieldCount` is deleted rather than kept: `ContextHandler` is its only caller, and keeping it would mean parsing the document twice per request. Its test is retargeted at `CountTopLevelMutationFields` — §7.2.

`pkg/portal/transport/graphql_handler.go:37` constructs a `graphqlutil.Handler` too; it leaves `BeforeExecuteFn` nil and is unaffected.

### 3.3 `AllowN` — `pkg/lib/ratelimit/limiter.go`

`Limiter.reserveN` already takes a float delta, and `Storage.Update` passes it to the GCRA script as `n` (`gcra.go:36`). The script computes `increment = ceil(emission_interval * n)` and **only writes when conforming** (line 44), so an `n`-token take is atomic: all `n` or none, in one round trip.

```go
// AllowN is Allow, taking n tokens instead of 1. n is multiplied by the
// group's context weight, so AllowN(ctx, spec, 1) is exactly Allow(ctx, spec).
// The take is atomic: the GCRA script only writes when the whole n conforms,
// so a rejected call consumes nothing.
func (l *Limiter) AllowN(ctx context.Context, spec BucketSpec, n int) (*FailedReservation, error) {
	weight := spec.RateLimitGroup.ResolveWeight(ctx)
	_, failedReservation, err := l.reserveN(ctx, spec, float64(n)*weight)
	return failedReservation, err
}
```

Only `AllowN` is added — **not** an exported `ReserveN`. The call flow never cancels a reservation (a rejected mutation still costs its tokens, matching DCR's "consumed by every attempt, successful or not"), so an exported `ReserveN` would be an unused export. This refines the earlier intent to "export `ReserveN`": same mechanism, minimal surface.

Looping `Allow` n times is rejected: `Allow` discards the successful `Reservation`, so a failure partway through leaves the earlier tokens consumed and uncancellable, and it costs n Redis round trips.

### 3.4 Bucket specs — `pkg/admin/transport/ratelimit.go` (new)

Placed beside the consumer, as `pkg/lib/oauth/handler/ratelimit.go` is for DCR and `pkg/lib/cimd/ratelimit.go` is for CIMD.

```go
package transport

// NewBucketSpecAdminAPIMutationAllPerProject bounds Admin API
// mutation volume (docs/specs/rate-limit.md § Admin API mutations). rateLimits
// is the resolved feature config scope and is non-nil at request time.
func NewBucketSpecAdminAPIMutationAllPerProject(rateLimits *config.AdminAPIRateLimitsMutationScopeFeatureConfig) ratelimit.BucketSpec {
	// No args: BucketSpec.IsGlobal is false, so Limiter keys by app id.
	return ratelimit.NewBucketSpec(
		ratelimit.RateLimitAdminAPIMutationAllPerProject,
		ratelimit.RateLimitGroupAdminAPIMutation,
		rateLimits.PerProject,
		ratelimit.AdminAPIMutationAllPerProject,
	)
}

```

### 3.5 Names and buckets — `pkg/lib/ratelimit/ratelimits.go`

```go
	// Admin API rate limits
	RateLimitGroupAdminAPIMutation RateLimitGroup = "admin_api.mutation"

	RateLimitAdminAPIMutationAllPerProject RateLimitName = "admin_api.mutation.all.per_project"

	AdminAPIMutationAllPerProject BucketName = "AdminAPIMutationAllPerProject"
```

Nothing is added to `ResolveBucketSpecs`, `resolvePerProject`, or the `perProjectName` helper — like DCR and CIMD, the specs are constructed directly by the consumer. `ResolveWeight`'s `default` branch already returns weight 1 for an unlisted group (`ratelimits.go:679`), so no change there either.

### 3.6 The handler — `pkg/admin/transport/handler_graphql.go`

```go
type MutationRateLimiter interface {
	AllowN(ctx context.Context, spec ratelimit.BucketSpec, n int) (*ratelimit.FailedReservation, error)
}

type GraphQLHandler struct {
	GraphQLContext        *graphql.Context
	AppDatabase           *appdb.Handle
	RateLimiter           MutationRateLimiter
	AdminAPIFeatureConfig *config.AdminAPIFeatureConfig
}
```

Wired inside the existing `AppDatabase.WithTx` block:

```go
		graphqlHandler := &graphqlutil.Handler{
			Schema: graphql.Schema,
			BeforeExecuteFn: func(ctx context.Context, mutationFieldCount int) error {
				return h.checkMutationRateLimit(ctx, mutationFieldCount)
			},
			ResultCallbackFn: ...unchanged...,
		}
```

```go
// checkMutationRateLimit takes mutationFieldCount tokens from the Admin API
// mutation bucket, which is keyed on app_id alone.
func (h *GraphQLHandler) checkMutationRateLimit(ctx context.Context, mutationFieldCount int) error {
	if mutationFieldCount <= 0 {
		return nil
	}
	rateLimits := h.AdminAPIFeatureConfig.GetRateLimits().GetMutation().GetAll()
	if rateLimits == nil {
		return nil
	}
	spec := NewBucketSpecAdminAPIMutationAllPerProject(rateLimits)
	failed, err := h.RateLimiter.AllowN(ctx, spec, mutationFieldCount)
	if err != nil {
		return err
	}
	if failed != nil {
		return failed.Error()
	}
	return nil
}
```

### 3.7 Exact call sequence

Entry: `POST /graphql` (or `/_api/admin/graphql`) → `pkg/admin/transport/handler_graphql.go:31` `GraphQLHandler.ServeHTTP`.

1. `GET` → GraphiQL, unchanged, no limiting.
2. `POST` → `h.AppDatabase.WithTx` opens the app DB transaction (line 47).
3. `graphqlutil.Handler.ContextHandler` parses request options (`NewRequestOptions`).
4. `CountTopLevelMutationFields(opts.Query, opts.OperationName)` → `n`. A query yields `n = 0`.
5. `n > 5` → `writeErrorResponse` with `BadRequest`, HTTP 400, return. Unchanged behaviour.
6. `BeforeExecuteFn(ctx, n)` → `checkMutationRateLimit`:
   - `n == 0` → return nil, no Redis call. Queries are never limited.
   - resolve the scope config from the feature config;
   - `AllowN(per_project, n)`: `Limiter.reserveN` → `Storage.Update(bucketKeyApp(appID, spec), 1m, 1000, n)` → one GCRA script call. Not conforming → `FailedReservation`;
   - on failure: `Limiter.doReserveN` logs, and because `RateLimitGroup` is non-empty (`limiter.go:119`) dispatches `rate_limit.blocked`; `l.Database.IsInTx(ctx)` is **true** here, so the event joins the open transaction;
   - `failed.Error()` → `ratelimit.ErrRateLimited(name, group, bucketName)` → `apierrors.TooManyRequest` / reason `RateLimited`, details carrying `rate_limit.name` and `rate_limit.group`;
   - otherwise repeat for `per_project`.
7. On rejection: `writeErrorResponse` writes HTTP **429** (`apiError.Code`, `handler.go:202`) and returns. `ResultCallbackFn` is never invoked, so `doRollback` stays false and `WithTx` **commits** — which is what persists the `rate_limit.blocked` audit log. No user-visible writes happened.
8. On success: `graphql.Do` executes as before.

## 4. Storage, deployment, and compatibility

**Keys.** `bucketKeyApp` (`limiter.go:171`) formats `app:<app_id>:rate-limit:<bucket_key>`, where `BucketSpec.Key()` joins the bucket name and arguments with `:` (`bucket.go:69`):

- `app:<app_id>:rate-limit:AdminAPIMutationAllPerProject`

Both are new. No existing key is read, written, or renamed; no backfill, dual-read, or dual-write. An IP cannot contain `:` in the form `GetIP` returns (it strips brackets and ports, `ip.go:24-29`), so no key-collision handling is needed.

**Rollout.** Stateless. Old and new binaries can run side by side: old ones simply do not consume the buckets. Redis keys expire via the script's `EXPIREAT`, so a rollback leaves nothing behind.

**Config compatibility.** Purely additive. An existing `authgear.features.yaml` with no `admin_api.rate_limits` gets the §2.3 defaults. `additionalProperties: false` at each new level means a typo in a plan document fails validation loudly rather than silently disabling the limit.

**API compatibility.** No GraphQL schema change, so `make export-schemas` is not required and `portal/src/graphql/adminapi/schema.graphql` does not move. No portal UI change: `PortalFeatureConfig` (`pkg/portal/model/app.go:15`) whitelists the sections exposed to the console and does not include `admin_api`, so the new fields are not surfaced there. The Site Admin API's `effective_app_feature_config` / `effective_plan_feature_config` marshal the whole `FeatureConfig`, so the new subtree appears there automatically — additive, no field renamed or removed.

**Error shape.** A rejected request returns the standard rate limit error already produced everywhere else (`ratelimit.ErrRateLimited`): `TooManyRequest` / `RateLimited`, with `rate_limit` (`{name, group}`) plus the deprecated `bucket_name` and `rate_limit_name` details preserved by `error.go:14-28`. Nothing new, nothing removed.

## 5. Audit log

No new event type. `rate_limit.blocked` is dispatched by the existing limiter path because `BucketSpec.RateLimitGroup` is non-empty (`limiter.go:119-137`), carrying `name: admin_api.mutation.all.per_project` and `group: admin_api.mutation`. `docs/specs/event.md` needs no change.

## 6. File-level change plan

### 6.1 `pkg/util/graphqlutil/handler.go`
Add exported `CountTopLevelMutationFields`; delete `validateMutationFieldCount` and inline its limit check in `ContextHandler`; add the `BeforeExecuteFn` field and its call site. `countTopLevelFields` and `findOperation` unchanged.

### 6.2 `pkg/lib/config/feature_admin_api.go`
Add the three structs, four schema registrations, `SetDefaults`, three `Merge` methods, three accessors, and the one-line cascade in `AdminAPIFeatureConfig.Merge`.

### 6.3 `pkg/lib/ratelimit/ratelimits.go`, `pkg/lib/ratelimit/limiter.go`
Add the group, two names, two bucket names; add `AllowN`.

### 6.4 `pkg/admin/transport/ratelimit.go` (new), `pkg/admin/transport/handler_graphql.go`
Bucket spec constructors; handler fields, `MutationRateLimiter` interface, `checkMutationRateLimit`, `BeforeExecuteFn` wiring.

### 6.5 `pkg/admin/wire.go`, `pkg/admin/wire_gen.go`
Bind `MutationRateLimiter` to `*ratelimit.Limiter`. `wire_gen.go` is regenerated with `make generate`, never hand-edited, and lands in the same commit as the struct change.

### 6.6 `docs/specs/rate-limit.md`
Correct the "What is counted" paragraph: a document carries at most `maxMutationFieldsPerRequest` (5) top-level mutation fields, enforced before execution, so per-request charging would undercount by at most 5x — still worth charging per field, but the current "a 100-field document costs 100 tokens" is impossible and must go. Add that a tier configuring `burst < 5` makes a maximal document permanently unsatisfiable (§8 D5).

### 6.7 `.vettedpositions`
`/pkg/admin/transport/handler_graphql.go:47:9: requestcontext` is a vetted position and line 47 moves. Run `go run ./devtools/goanalysis ./cmd/... ./pkg/...`, update the position, `make sort-vettedpositions`, per the `update-vettedpositions` skill. Same commit as the handler change.

## 7. Test plan

### 7.1 `pkg/lib/config` — Convey (`feature_test.go:12`)

- **`testdata/default_feature.yaml`** — add the resolved subtree under `admin_api`: `rate_limits.mutation.all.per_project` = `{enabled: true, period: 1m, burst: 1000}`. Without this `TestParseFeatureConfig`'s "default feature config" case fails.
- **`testdata/merge_feature.yaml`** — two cases, both with values that are **not** the code default so a whole-section regression cannot pass by accident:
  - *cross-section*: layer 1 sets only `admin_api.rate_limits.mutation.all.per_project` (`period: 1m, burst: 7`); layer 2 sets only `admin_api.create_session_enabled: true`. Assert the result has **both**. This is the case that proves the §9 app-level override is safe.
- **`testdata/parse_feature_tests.yaml`** — schema validation, mirroring the CIMD cases at lines 270-317: `/admin_api/rate_limits/mutation/all/per_project/burst: minimum` (burst 0), `/admin_api/rate_limits/mutation/all/per_project: required` (`enabled: true` with no `period`), `/admin_api/rate_limits/mutation/all/per_ip` (rejected — there is no such bucket), `/admin_api/rate_limits/mutation/all/unknown_key`, `/admin_api/rate_limits/mutation/unknown_key`.

### 7.2 `pkg/util/graphqlutil/handler_test.go` — standard table-driven `testing.T`

Retarget `TestValidateMutationFieldCount` at `CountTopLevelMutationFields`, keeping every existing case and asserting the returned count rather than only the error. Add: a query operation returns 0; a document with 3 aliased copies of one mutation returns 3; a fragment spread expanding to 2 fields returns 2; a fragment cycle still errors.

**Ambiguous-operation case.** `findOperation` returns `nil, nil` when a document holds more than one operation and `operationName` is empty (`handler.go:252-257`), so `CountTopLevelMutationFields` returns 0 and §3.6 skips the limiter entirely. That is only safe if such a document never executes a mutation. Add a handler-level test — not just a counter test — that posts a document with two mutation operations and no `operationName` and asserts the response contains a GraphQL error and no mutation took effect. If `graphql.Do` turns out to execute one of them, the fix is to treat an unresolvable operation as a rejection in `ContextHandler` rather than as `n = 0`; decide that against the observed behaviour rather than assuming it.

### 7.3 `pkg/lib/ratelimit` — Convey (`ratelimits_test.go:10`)

`AllowN`: n tokens are taken in one call; a call for more than the remaining allowance is rejected **and consumes nothing** (assert a subsequent single-token `Allow` still succeeds — this is the property that makes `AllowN` correct where looping `Allow` is not); `AllowN(ctx, spec, 1)` is equivalent to `Allow`; a disabled spec is allowed regardless of n.

### 7.4 `pkg/admin/transport/handler_graphql_test.go` (new) — Convey, matching the sibling `pkg/portal/transport/admin_api_handler_test.go:8`

`checkMutationRateLimit` against a fake `MutationRateLimiter`: count 0 makes no limiter call; count 3 charges 3 to both buckets; per-IP is charged before per-project; a per-IP rejection returns `RateLimited` and never reaches per-project; a nil scope config is a no-op.

### 7.6 e2e — `e2e/tests/admin_api/mutation_ratelimit.test.yaml` (new)

YAML-driven, per the `write-e2e-test` skill. Uses the `admin_api_graphql` action (`e2e/tests/admin_api/create_user.test.yaml:4`) and an `authgear.features.yaml` override, as `e2e/tests/cimd/ratelimit_per_project.test.yaml:7` does.

```yaml
authgear.features.yaml:
  override: |
    admin_api:
      rate_limits:
        mutation:
          all:
            per_project:
              enabled: true
              period: 1h
              burst: 2
```

Cases to cover:
1. Two `createGroup` mutations succeed; the third is rejected with `TooManyRequest` / `RateLimited` and `rate_limit.name == "admin_api.mutation.all.per_project"`.
2. A query (e.g. fetching the group list) still succeeds after the bucket is exhausted — proving queries are not charged.

**Multi-field document — `e2e/tests/admin_api/mutation_ratelimit_multi_field.test.yaml` (new).** The property under test is the one that makes per-field charging worth doing at all: a single document carrying several mutation fields must cost one token per field, not one per request. Without it, batching is a 5x bypass of whatever the bucket says.

With `per_project: {enabled: true, period: 1h, burst: 3}`:

1. One document containing three aliased `createGroup` fields succeeds and creates three groups — exhausting the bucket in a single request.
2. A following single-field `createGroup` is rejected with `TooManyRequest` / `RateLimited`. If tokens were charged per request rather than per field, this fourth request would be the second charge against a burst of 3 and would wrongly succeed — so this assertion is what actually pins the behaviour.
3. A following query still succeeds, confirming D1 holds on the exhausted bucket.

This is deliberately e2e rather than only a handler unit test: it is the one property that spans parsing, counting, and the atomic multi-token take, and a unit test on any single layer would not catch a regression in the seam between them.

The **ambiguous multi-operation** case from §7.2 (several operations, no `operationName`) stays a Go handler test rather than an e2e case — it asserts that nothing executes, which is a negative that the YAML runner expresses poorly.

Run with `cd e2e && make teardown && make setup`, then `go test ./pkg/testrunner/ -count 1 -v -timeout 10m -run "TestAuthflow/admin_api/mutation_ratelimit"`.

### 7.7 Gate

Per `CLAUDE.md`, run the `review-pr` skill on the finished diff and resolve every finding before calling this done. Package commands: `go test ./pkg/lib/config/... ./pkg/lib/ratelimit/... ./pkg/util/graphqlutil/... ./pkg/util/httputil/... ./pkg/admin/...`, plus `make lint`.

## 8. Fixed behavioral decisions

- **D1.** Queries are never charged. `mutationFieldCount == 0` short-circuits before any Redis call.
- **D2.** One token per top-level mutation field, including aliased repeats and fields reached through fragment spreads — whatever `CountTopLevelMutationFields` returns.
- **D3.** Tokens are consumed by every attempt, successful or not. A mutation that fails validation or errors during execution still cost its tokens; there is no cancel path.
- **D5.** A document with more mutation fields than `burst` can never succeed, because GCRA never writes when non-conforming and so the state never improves. This is unreachable at the default (5 ≤ 1000) and only becomes reachable if a tier sets `burst < 5`. Recorded in the spec rather than special-cased in code.
- **D6.** Per-mutation scopes are not built here. When built they are consumed **in addition to** `all`, not as a fallback, so a loosened per-mutation limit cannot punch through the system-wide bound.
- **D7.** No `per_user` bucket. The acting user is only known for portal-proxied requests (`actor_user_id` in the Admin API audit context) and is absent for every direct Admin API key call.
- **D8.** There is no per-IP bucket. The Admin API authenticates as the project, so `app_id` is the caller identity and IP would be a weaker, sidesteppable proxy for it. The per-IP buckets on DCR registration and the CIMD fetch are not a precedent: those endpoints are unauthenticated, where IP is the only handle on a caller.
- **D9.** Internal Admin API traffic is **not** exempted. `usage: internal` lives in the Admin API JWT's audit context, which is signed with the project's own Admin API key — a key tenants hold — so an exemption keyed on it would be forgeable by exactly the caller being limited. The portal's own project is handled by config instead, §9.

## 9. Rollout steps outside this repository

Neither can ship in the PR; both must be tracked separately.

1. **Plan-tier values.** Tighter buckets for lower tiers are set on the plan records with the portal plan CLI (`cmd/portal/plan/service.go:39`, `UpdatePlan`), not in this repo. Until that is done every tier runs at the 1000/minute default.
2. **An app-level override for the portal's own Authgear project.** The portal calls its own Admin API for `createAccount` when inviting a collaborator with no existing account (`pkg/portal/service/collaborator.go:877`) and `submitOnboardEntry` at onboarding (`pkg/portal/service/onboard.go:62`), both through `SelfDirector` against `AuthgearConfig.AppID`. Give that project an `authgear.features.yaml` at the app FS level raising or disabling `admin_api.rate_limits.mutation.all`. Note `AuthgearFeatureYAMLDescriptor.UpdateResource` refuses edits (`resources.go:721`), so this is written into the project's resource rows (DB source) or placed beside `authgear.yaml` (local FS source) — not through the portal UI. §2.4's cascade is what keeps this override from clobbering the rest of the project's `admin_api` feature config.

## 10. Atomic commit plan

Each commit builds and tests green on its own.

**C1 — `Return the mutation field count from graphqlutil`**
`pkg/util/graphqlutil/handler.go`, `pkg/util/graphqlutil/handler_test.go`. Add `CountTopLevelMutationFields`, delete `validateMutationFieldCount`, inline the limit check, retarget and extend the tests. Pure refactor: no behaviour change, no new dependency, no wiring.

**C2 — `Add AllowN to the rate limiter`**
`pkg/lib/ratelimit/limiter.go`, `pkg/lib/ratelimit/limiter_test.go` (or `ratelimits_test.go`, matching where limiter tests already live). Additive; no existing caller changes.

**C3 — `Add admin_api.rate_limits feature config`**
`pkg/lib/config/feature_admin_api.go` and the three `testdata` fixtures. Config only — nothing reads it yet. Must include the `default_feature.yaml` update or the package's own tests fail, and the `merge_feature.yaml` cases required by the `update-feature-config` skill.

**C4 — `Add admin_api.mutation rate limit names`**
`pkg/lib/ratelimit/ratelimits.go`. Constants only.

**C5 — `Rate limit Admin API GraphQL mutations`**
`pkg/admin/transport/ratelimit.go` (new), `pkg/admin/transport/handler_graphql.go`, `pkg/admin/transport/handler_graphql_test.go` (new), `pkg/admin/wire.go`, `pkg/admin/wire_gen.go`, `.vettedpositions`. The behavioural commit. Regenerated wiring (`make generate`) and the vetted-position update belong **in this commit**, not a follow-up, so the tree is green at every point.

**C6 — `Add e2e tests for Admin API mutation rate limits`**
`e2e/tests/admin_api/mutation_ratelimit.test.yaml` and the per-IP variant.

**C7 — `doc: Correct the mutation field count in the rate limit spec`**
`docs/specs/rate-limit.md`, per §6.6.
