# DCR Part 5 — `oauth_client_dcr` Usage Limit

Spec: [docs/specs/usage.md](../../specs/usage.md), [docs/specs/dcr.md — Client Limit](../../specs/dcr.md#client-limit). Supersedes the `oauth.dynamic_client_registration.maximum_clients` design excluded from Parts 1–4 — as of `spec` branch commit "Fold DCR/CIMD client limits into the usage limit config," the DCR client cap is a new **standing** (period-less) entry in the general `usage.limits` engine, not a standalone `dynamic_client_registration` config key.

Depends on Part 2 (`oauthclient.Store.CountClientsBySource`, `RegistrationHandlerDCRService.CountClientsBySource` already declared but unused; `pkg/lib/oauth/handler/handler_register.go`).

## 1. Goal / Scope

Implement `usage.limits.oauth_client_dcr` end to end: config schema, the new "standing" concept in the usage engine (`pkg/lib/usage`), and wiring it into `POST /oauth2/register` so the project's DCR client count is capped per plan tier.

Confirmed by research: the existing `usage.Limiter` (`pkg/lib/usage/limit.go`) is structurally periodic-only — `Reserve` hardcodes a `day`/`month` period loop over Redis-backed atomic counters, and `legacyLimitName` **panics** on any `model.UsageName` it doesn't recognize. There is **no existing precedent** for a limit whose "usage" is a live `COUNT(*)` over a database table rather than an incrementing counter. This part adds that as a new, parallel code path — it does not touch `Reserve` or any of its periodic call sites (`email`/`sms`/`whatsapp`/`user_export`/`user_import` are untouched).

Out of scope: `oauth_client_cimd` (CIMD is not implemented in this codebase yet — the config schema additions below are structured so adding `oauth_client_cimd` later is a same-shape follow-up, but it is not added here to avoid dead code for a feature that doesn't exist).

## 2. Config Model & Schema

### 2.1 `pkg/api/model/usage.go` — new usage name

```go
const (
	UsageNameUserExport      UsageName = "user_export"
	UsageNameUserImport      UsageName = "user_import"
	UsageNameEmail           UsageName = "email"
	UsageNameWhatsapp        UsageName = "whatsapp"
	UsageNameSMS             UsageName = "sms"
	UsageNameOAuthClientDCR  UsageName = "oauth_client_dcr" // new
)
```

### 2.2 `pkg/lib/config/feature_usage.go` (authgear.features.yaml) — new standing shape

A **separate** schema/struct from the periodic `FeatureUsageLimitConfig`, since a standing entry has no `period`:

```go
var _ = FeatureConfigSchema.Add("StandingFeatureUsageLimitConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"quota": { "type": "integer", "minimum": 0 },
		"action": { "$ref": "#/$defs/UsageLimitAction" }
	},
	"required": ["quota", "action"]
}
`)

type StandingFeatureUsageLimitConfig struct {
	Quota  int                    `json:"quota"`
	Action model.UsageLimitAction `json:"action"`
}
```

Extend `FeatureUsageLimitsConfig` (`pkg/lib/config/feature_usage.go:38-58`):

```go
var _ = FeatureConfigSchema.Add("FeatureUsageLimitsConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"user_export": { "type": "array", "items": { "$ref": "#/$defs/FeatureUsageLimitConfig" } },
		"user_import": { "type": "array", "items": { "$ref": "#/$defs/FeatureUsageLimitConfig" } },
		"email": { "type": "array", "items": { "$ref": "#/$defs/FeatureUsageLimitConfig" } },
		"whatsapp": { "type": "array", "items": { "$ref": "#/$defs/FeatureUsageLimitConfig" } },
		"sms": { "type": "array", "items": { "$ref": "#/$defs/FeatureUsageLimitConfig" } },
		"oauth_client_dcr": { "type": "array", "items": { "$ref": "#/$defs/StandingFeatureUsageLimitConfig" } }
	}
}
`)

type FeatureUsageLimitsConfig struct {
	UserExport     []FeatureUsageLimitConfig         `json:"user_export,omitempty"`
	UserImport     []FeatureUsageLimitConfig         `json:"user_import,omitempty"`
	Email          []FeatureUsageLimitConfig         `json:"email,omitempty"`
	Whatsapp       []FeatureUsageLimitConfig         `json:"whatsapp,omitempty"`
	SMS            []FeatureUsageLimitConfig         `json:"sms,omitempty"`
	OAuthClientDCR []StandingFeatureUsageLimitConfig `json:"oauth_client_dcr,omitempty"` // new
}
```

New accessor (parallel to, not replacing, `Limits(name)` at `feature_usage.go:60-79`, since the element type differs):

```go
func (c *FeatureUsageLimitsConfig) StandingLimits(name model.UsageName) []StandingFeatureUsageLimitConfig {
	if c == nil {
		return nil
	}
	switch name {
	case model.UsageNameOAuthClientDCR:
		return c.OAuthClientDCR
	default:
		return nil
	}
}
```

`UsageMatch` (`feature_usage.go:5-10`, used by `usage.hooks[].match`) gains **both** `"oauth_client_dcr"` and `"oauth_client_cimd"` in its enum, so a hook can subscribe to this usage name's crossing events like any other.

Adding `oauth_client_cimd` to the enum is deliberate even though CIMD is not implemented (see the out-of-scope note in §1): usage.md lists it under [Supported Usage Names](../../specs/usage.md#supported-usage-names) and says every name there is a valid `match` value, so omitting it would make a spec-valid `authgear.features.yaml` fail schema validation. This is enum-only — no `oauth_client_cimd` field on `FeatureUsageLimitsConfig`, no `model.UsageName` constant, no `StandingLimits` case, so no dead code path. The same applies to `UsageMatch` in `usage.go` (§2.3).

**Field-level merge** (`mergeFeatureUsageLimits`, `feature_usage.go:186-213`) — add the same whole-list-per-name override pattern already used for every other field, per the `update-feature-config` skill's hard requirement:

```go
if layer.OAuthClientDCR != nil {
	merged.OAuthClientDCR = layer.OAuthClientDCR
}
```

### 2.3 `pkg/lib/config/usage.go` (authgear.yaml) — deliberately NOT extended

Per usage.md: "`oauth_client_dcr` ... [is] only meaningful ... in the feature-config hierarchy ... not in the project-editable `authgear.yaml` section ... since a project admin doesn't set their own DCR client cap." So `UsageLimitsConfig` (`usage.go:59-65`) gets **no** new field — a project cannot set this quota themselves, matching the existing precedent that all `*_usage`-style plan caps are features-only. `UsageMatch` in `usage.go:19-24` (used by `usage.alerts[].match`) **does** gain `"oauth_client_dcr"` and `"oauth_client_cimd"`, though — a project admin can't set the quota, but can still ask to be emailed when it's crossed, exactly as they can for the other plan-tier-influenced usage names today. See §2.2 for why the CIMD value is added to the enums despite CIMD being unimplemented.

### 2.4 `make export-schemas`

Regenerate the JSON Schema artifacts in the same commit as the Go changes.

## 3. Runtime — `pkg/lib/usage` Standing Limit Path

New file `pkg/lib/usage/standing.go`. Does not modify `limit.go`'s `Reserve`/`reservePeriod`/Redis machinery at all — reuses only the already period-agnostic pieces (`usageHookURLs`, `usageAlertRecipients`, `dispatchEventImmediately`, `maybeDispatchUsageAlert`, all still on `*Limiter` in `limit.go`).

```go
package usage

type EffectiveStandingUsageLimit struct {
	Name   model.UsageName
	Quota  int
	Action model.UsageLimitAction
}

func (l *Limiter) effectiveStandingUsageLimits(name model.UsageName) []EffectiveStandingUsageLimit {
	var limits []EffectiveStandingUsageLimit
	if l.EffectiveConfig.FeatureConfig.Usage != nil && l.EffectiveConfig.FeatureConfig.Usage.Limits != nil {
		for _, limit := range l.EffectiveConfig.FeatureConfig.Usage.Limits.StandingLimits(name) {
			limits = append(limits, EffectiveStandingUsageLimit{Name: name, Quota: limit.Quota, Action: limit.Action})
		}
	}
	// Deliberately no AppConfig.Usage.Limits lookup here (see config §2.3 —
	// standing limits are never project-editable).
	return limits
}

func (l *Limiter) minBlockStandingQuota(limits []EffectiveStandingUsageLimit) (int, bool) {
	minQuota := 0
	found := false
	for _, limit := range limits {
		if limit.Action != model.UsageLimitActionBlock {
			continue
		}
		if !found || limit.Quota < minQuota {
			minQuota = limit.Quota
			found = true
		}
	}
	return minQuota, found
}

func crossedStandingUsageLimits(before, after int, limits []EffectiveStandingUsageLimit) []EffectiveStandingUsageLimit {
	var crossed []EffectiveStandingUsageLimit
	for _, limit := range limits {
		if before < limit.Quota && after >= limit.Quota {
			crossed = append(crossed, limit)
		}
	}
	return crossed
}

// CheckStanding returns ErrStandingUsageLimitExceeded if creating one more
// record of this usage name, on top of currentCount, would exceed the
// configured block quota. It does not create or reserve anything — the
// caller creates the record itself only if this returns nil (see §4 for the
// full check-then-insert sequence and its concurrency handling).
func (l *Limiter) CheckStanding(ctx context.Context, name model.UsageName, currentCount int) error {
	limits := l.effectiveStandingUsageLimits(name)
	if blockQuota, ok := l.minBlockStandingQuota(limits); ok && currentCount+1 > blockQuota {
		return ErrStandingUsageLimitExceeded(name)
	}
	return nil
}

// ReportStandingCreated fires alert/hook/event triggers for any quota
// crossed by the creation that just succeeded. Call only after the new
// record has actually been committed, with the count observed immediately
// before that creation — mirrors Reserve's before/after crossing detection
// (limit.go:159,166) but against a live COUNT(*) instead of a Redis counter.
func (l *Limiter) ReportStandingCreated(ctx context.Context, name model.UsageName, countBeforeCreate int) {
	limits := l.effectiveStandingUsageLimits(name)
	for _, standing := range crossedStandingUsageLimits(countBeforeCreate, countBeforeCreate+1, limits) {
		// Period is empty for a standing limit; UsageAlertPayload.Period
		// becomes the zero value model.UsageLimitPeriod(""), which
		// pkg/admin's/pkg/portal's payload consumers must treat as "not
		// applicable" for this usage name (see §6 for the payload note).
		_ = l.maybeDispatchUsageAlert(ctx, EffectiveUsageLimit{
			Name:   standing.Name,
			Quota:  standing.Quota,
			Period: "",
			Action: standing.Action,
		}, countBeforeCreate+1)
	}
}
```

`maybeDispatchUsageAlert` (`limit.go:371-384`) is reused unmodified — it only reads `limit.Name`/`limit.Action` (via `usageHookURLs`/`usageAlertRecipients`) and passes `limit` straight into `makeUsageAlertTriggeredPayload`, which is a plain struct copy; no periodic-only logic lives in that call chain.

### 3.1 `pkg/lib/usage/errors.go` — new error constructor

`ErrUsageLimitExceeded` (`errors.go:9-15`) calls `legacyLimitName(name)`, which **panics** on any name it doesn't have a case for (`limit.go:460-475`) — `oauth_client_dcr` has no such case and must not be routed through it. Add a dedicated constructor instead of adding a fake `LimitName` mapping for a usage name that has no Redis key at all:

```go
func ErrStandingUsageLimitExceeded(name model.UsageName) error {
	return UsageLimitExceeded.NewWithInfo("usage limit exceeded", apierrors.Details{
		"usage_name": name,
	})
}
```

Reuses the same `UsageLimitExceeded` (`apierrors.TooManyRequest.WithReason("UsageLimitExceeded")`) kind for consistency in logs/metrics; the DCR registration handler (§4) translates this into the RFC 7591-mandated `access_denied` (403), so the exact `apierrors.Kind` here is an internal implementation detail, not what the API caller sees.

## 4. Wiring into `POST /oauth2/register`

Extends Part 2 §5.1's `RegistrationHandler.Handle` call sequence (`pkg/lib/oauth/handler/handler_register.go`), inserting the check between step 5 (`dcr.ValidateAndNormalize`) and step 6 (`h.DCR.CreateClient`):

```go
type RegistrationHandlerUsageLimiter interface {
	CheckStanding(ctx context.Context, name model.UsageName, currentCount int) error
	ReportStandingCreated(ctx context.Context, name model.UsageName, countBeforeCreate int)
}
```

Add `UsageLimiter RegistrationHandlerUsageLimiter` to the `RegistrationHandler` struct (Part 2 §5.1), bound to `*usage.Limiter` the same way the three existing consumers (`pkg/admin/transport/handler_user_export_create.go`, `pkg/lib/userimport/job.go`, `pkg/lib/messaging/limits.go`) already bind their own local `UsageLimiter` interfaces to it.

Updated step 6 in `Handle` (all still inside the same `h.Database.WithTx` per Part 2 §5.1):

```go
// Close the check-then-insert race between concurrent registrations for the
// same app: a plain "SELECT COUNT(*) then INSERT" has a TOCTOU window where
// two concurrent requests both observe a count under quota and both proceed,
// which Reserve's Redis Lua script avoids for free via atomicity that a
// Postgres COUNT(*) does not get automatically. Serialize per-app with a
// transaction-scoped advisory lock — cheap, and registration is not a hot
// path like token issuance.
if err := h.DCR.LockForClientCount(ctx); err != nil {
	return nil, err
}

count, err := h.DCR.CountClientsBySource(ctx, model.OAuthClientSourceDCR)
if err != nil {
	return nil, err
}
if err := h.UsageLimiter.CheckStanding(ctx, model.UsageNameOAuthClientDCR, int(count)); err != nil {
	return nil, protocol.NewErrorStatusCode("access_denied", "the project has reached its dynamic client registration limit", 403)
}

client, err := h.DCR.CreateClient(ctx, options)
if err != nil {
	return nil, err
}
h.UsageLimiter.ReportStandingCreated(ctx, model.UsageNameOAuthClientDCR, int(count))
```

Confirmed: there is **no** advisory-lock helper anywhere in the codebase (`grep -rn pg_advisory` over `*.go` returns nothing), so this is new. It does **not** belong on `*appdb.Handle` — that is `db.HookHandle` (`pkg/lib/infra/db/hook_handle.go`), which owns transaction lifecycle (`WithTx`/`ReadOnly`/`IsInTx`) but no `SQLBuilder`/`SQLExecutor` and so cannot issue a query. Put it on `oauthclient.Store` instead (Part 2 §4), which has both:

```go
// pkg/lib/oauthclient/store_client.go
//
// LockForClientCount serializes concurrent DCR registrations for this app so
// the CountClients-then-CreateClient sequence in the registration handler is
// atomic with respect to the configured quota. Must be called inside a
// transaction: pg_advisory_xact_lock releases at transaction end, unlike the
// session-scoped pg_advisory_lock, which would leak the lock across the
// pooled connection's next user.
func (s *Store) LockForClientCount(ctx context.Context) error {
	_, err := s.SQLExecutor.ExecWith(ctx, s.SQLBuilder.
		Select("pg_advisory_xact_lock(hashtext(?))", "oauth_client_dcr:"+<app id>))
	return err
}
```

Exposed to the handler through `oauthclient.Commands` and added to `RegistrationHandlerDCRService` (Part 2 §5.1) alongside `CountClientsBySource`. Getting the app id inside `Store` follows whatever `appdb.SQLBuilderApp` already does for its implicit `app_id` scoping — check whether it exposes the id directly; if not, add `AppID config.AppID` to `oauthclient.Store` (several `pkg/lib` stores already carry it) rather than plumbing it through the handler, since Part 2's `RegistrationHandler` has no `AppID` field.

The lock is named `oauth_client_dcr:<app id>` — **per source, not per table.** CIMD's quota is counted and enforced independently (§5), so a CIMD resolution must not serialize behind a DCR registration or vice versa. When CIMD lands, its check-then-insert takes `oauth_client_cimd:<app id>`.

The `access_denied`/403 mapping here is the one place this part's error handling **must** deviate from `ErrStandingUsageLimitExceeded`'s underlying 429 `apierrors.Kind` — dcr.md's error table (already implemented in Part 2 §5.1/§Errors) is explicit: `access_denied` / 403 for "the project's client limit has been reached," not a generic rate-limit-shaped 429. Do not let `ErrStandingUsageLimitExceeded` propagate raw past this handler.

## 5. `CountClients` Semantics Check

Part 2 §4.2 specifies `Store.CountClientsBySource(ctx, source) (uint64, error)` as `SELECT COUNT(*) FROM _auth_oauth_client WHERE app_id = ? AND source = ?`, served by the `(app_id, source)` index. No changes needed to that method's design — this part is its first caller.

**The source filter is not optional.** `_auth_oauth_client` holds DCR and CIMD rows in one table (Part 2 §3.1), while dcr.md counts the `oauth_client_dcr` quota over "`OAuthClient` records with `source: DCR`" and cimd.md counts `oauth_client_cimd` over `source: CIMD`, against separate quotas. A source-less count would make each source's clients consume the other's quota — and because CIMD rows are created by unauthenticated `/oauth2/authorize` traffic, that would let CIMD activity silently exhaust the DCR cap. Pass `model.OAuthClientSourceDCR` explicitly at the call site; do not add a convenience `CountClients()` that omits it.

## 6. `usage.alert.triggered` Event Payload — Period Field for a Standing Name

`nonblocking.UsageAlertPayload{Name, Action, Period, Quota, CurrentValue, PlanName}` (confirmed present, exact file not modified by this plan) has a `Period model.UsageLimitPeriod` field that is meaningful for every existing (periodic) usage name. For `oauth_client_dcr`, §3's `ReportStandingCreated` passes `Period: ""` (the zero value). This is **not** a new field or schema change — `model.UsageLimitPeriod` is already just a `string` type, so `""` serializes as an empty string in the event payload / webhook body. Document this explicitly in the code (a comment on the `ReportStandingCreated` call site, already included in §3) so a future reader doesn't mistake it for a bug. No consumer-side code in this codebase currently branches on `Period` in a way that would break — verify this assumption holds by grepping `UsageAlertPayload` consumers before implementation, since this plan does not do so.

## 7. Test Plan

Unit tests (Convey):

- `pkg/lib/config/feature_usage_test.go` (extend if present, else new) — `FeatureUsageLimitsConfig.StandingLimits`/`Limits` return the right slice per name; `mergeFeatureUsageLimits` whole-list-overrides `OAuthClientDCR` independently of the other fields.
- `pkg/lib/config/testdata/merge_feature.yaml` (extend, per the `update-feature-config` skill's required test coverage) — add a new layer under a `usage:` entry with `limits.oauth_client_dcr`, and assert the merged `result.usage.limits.oauth_client_dcr` reflects only the highest-precedence layer's list (matching the existing `sms`/`email` entries' whole-list-override behavior already asserted in that file).
- `pkg/lib/usage/standing_test.go` (new) — `CheckStanding`: no configured limit → always nil; `action: alert` only (no block) → always nil regardless of count; `action: block, quota: 20` → nil for `currentCount<20`, `ErrStandingUsageLimitExceeded` for `currentCount>=20`. `ReportStandingCreated`: fires exactly once when `countBeforeCreate+1` crosses a configured quota (block or alert), does not fire again on a repeated call with the same already-crossed count (mirrors `crossedUsageLimits`'s existing tested behavior for the periodic path — check `pkg/lib/usage/limit_test.go` for the equivalent periodic test shape to mirror, if it exists).

e2e tests (extend `e2e/tests/dcr_register.yaml` from Part 2, or a new `e2e/tests/dcr_usage_limit.yaml`):

1. Set `authgear.features.yaml`'s `usage.limits.oauth_client_dcr` to `[{quota: 2, action: block}]` (via the e2e test harness's feature-config override mechanism — check `write-e2e-test` skill for how other tests already override plan-tier feature config, e.g. an existing `admin_api.user_export_usage`-based e2e test, and mirror that exact mechanism).
2. Register 2 DCR clients successfully.
3. Register a 3rd → `POST /oauth2/register` returns `access_denied` (403), not `invalid_client_metadata` or any other error.
4. Delete one DCR client via `deleteDynamicClient` (Part 2), then register again → succeeds (count dropped back under quota).
5. Configure an additional lower `{quota: 1, action: alert}` entry alongside the `{quota: 2, action: block}` entry; register 1 client and assert (via whatever mechanism existing usage-alert e2e tests already use to observe dispatched events/hooks — mirror it) that `usage.alert.triggered` fired once for the `alert` entry and, after the 2nd registration, once more for the `block` entry.

## 8. Fixed Behavioral Decisions

- Standing limits are **feature-config-only** — `authgear.yaml`'s `usage.limits` never gains `oauth_client_dcr`; only its `usage.alerts[].match` enum does.
- The check-then-insert sequence is serialized per-app via a transaction-scoped Postgres advisory lock (§4), not left as a best-effort race — unlike some "soft cap" limits elsewhere, closing this was cheap enough to just do rather than document as an accepted gap.
- `Reserve`/`reservePeriod`/the Redis Lua scripts in `limit.go` are completely untouched by this part; the standing path is fully additive.
- `oauth_client_cimd` is not implemented here — adding it later is a same-shape follow-up to §2/§3 once CIMD itself exists, not a redesign. Its **enum value** in both `UsageMatch` schemas is added now, so a spec-valid features.yaml validates (§2.2).
- The advisory lock lives on `oauthclient.Store`, not on `*appdb.Handle`, which has no SQL executor (§4), and its key is scoped per usage name so DCR and CIMD never serialize against each other.

## 9. Atomic Commit Plan

1. **`[DCR] Add oauth_client_dcr usage name and standing limit config`** — §2 (config schema/structs/merge) + unit tests + `merge_feature.yaml` fixture + `make export-schemas`.
2. **`[DCR] Add standing usage limit check to pkg/lib/usage`** — §3, §3.1 (new `standing.go`, error constructor) + unit tests. No callers yet.
3. **`[DCR] Enforce oauth_client_dcr usage limit in POST /oauth2/register`** — §4 (handler wiring, advisory lock helper if new) + regenerated `wire_gen.go` if the `UsageLimiter` binding requires new wiring.
4. **`[DCR] Add e2e tests for the DCR client usage limit`** — e2e YAML from §7.
