# Usage Implementation Plan

## Goal

Implement [`docs/specs/usage.md`](../specs/usage.md) against the current codebase.

The existing code only supports the deprecated feature-config fields:

- `messaging.sms_usage`
- `messaging.email_usage`
- `messaging.whatsapp_usage`
- `admin_api.user_import_usage`
- `admin_api.user_export_usage`

The current spec is materially different:

- usage is configured under a unified `usage` section
- `authgear.features.yaml` supports `usage.hooks`
- `authgear.yaml` supports `usage.alerts`
- both configs share the same `usage.limits` shape
- each usage can have multiple limits
- each limit has an `action` of `alert` or `block`
- `usage.alert.triggered` is emitted whenever a configured limit is triggered
- feature-config merge behavior is now explicit:
  - `usage.hooks` appends from lower precedence to higher precedence
  - `usage.limits.<usage_name>` is overridden by higher precedence
- deprecated legacy feature fields must continue to work during migration

This document replaces the previous soft-limit-only plan.

## Design Summary

Implementation is split into 5 workstreams:

1. Add a shared `usage` config model for feature config and app config.
2. Extend the runtime limiter from single hard-limit support to ordered multi-limit evaluation.
3. Deliver notifications through:
   - feature `usage.hooks`
   - app `usage.alerts`
   - normal event hooks listening to `usage.alert.triggered`
4. Preserve backward compatibility for deprecated feature fields.
5. Add server-side test coverage.

## 1. Config Model and Schema

### 1.1 Shared usage model

Introduce shared usage enums in `pkg/api/model`, and keep all config structs in `pkg/lib/config` following the existing codebase convention.

Add these types in `pkg/api/model/usage.go`:

```go
type UsageName string

const (
    UsageNameUserExport UsageName = "user_export"
    UsageNameUserImport UsageName = "user_import"
    UsageNameEmail      UsageName = "email"
    UsageNameWhatsapp   UsageName = "whatsapp"
    UsageNameSMS        UsageName = "sms"
)

type UsageLimitPeriod string

const (
    UsageLimitPeriodDay   UsageLimitPeriod = "day"
    UsageLimitPeriodMonth UsageLimitPeriod = "month"
)

type UsageLimitAction string

const (
    UsageLimitActionAlert UsageLimitAction = "alert"
    UsageLimitActionBlock UsageLimitAction = "block"
)
```

All config structs should stay in `pkg/lib/config`, and app-config structs should be separate from feature-config structs even when their shapes are parallel.

Follow the existing file placement pattern in this repository:

- feature-config structs belong in `feature_xx.go` files
- app-config structs belong in non-feature config files
- do not place feature-config structs in a shared non-feature file if the repo pattern already gives them a dedicated `feature_xx.go` home

Add these types in `pkg/lib/config/feature_usage.go` and `pkg/lib/config/usage.go`:

```go
// pkg/lib/config/feature_usage.go
type FeatureUsageLimitConfig struct {
    Quota  int                       `json:"quota"`
    Period model.UsageLimitPeriod `json:"period"`
    Action model.UsageLimitAction `json:"action"`
}

type FeatureUsageLimitsConfig struct {
    UserExport []FeatureUsageLimitConfig `json:"user_export,omitempty"`
    UserImport []FeatureUsageLimitConfig `json:"user_import,omitempty"`
    Email      []FeatureUsageLimitConfig `json:"email,omitempty"`
    Whatsapp   []FeatureUsageLimitConfig `json:"whatsapp,omitempty"`
    SMS        []FeatureUsageLimitConfig `json:"sms,omitempty"`
}

// pkg/lib/config/usage.go
type UsageLimitConfig struct {
    Quota  int                       `json:"quota"`
    Period model.UsageLimitPeriod `json:"period"`
    Action model.UsageLimitAction `json:"action"`
}

type UsageLimitsConfig struct {
    UserExport []UsageLimitConfig `json:"user_export,omitempty"`
    UserImport []UsageLimitConfig `json:"user_import,omitempty"`
    Email      []UsageLimitConfig `json:"email,omitempty"`
    Whatsapp   []UsageLimitConfig `json:"whatsapp,omitempty"`
    SMS        []UsageLimitConfig `json:"sms,omitempty"`
}
```

Add these helper methods:

```go
func (c *FeatureUsageLimitsConfig) Limits(name model.UsageName) []FeatureUsageLimitConfig
func (c *UsageLimitsConfig) Limits(name model.UsageName) []UsageLimitConfig
```

### 1.2 Feature config: `usage.hooks` + `usage.limits`

Add `Usage *FeatureUsageConfig \`json:"usage,omitempty"\`` to `pkg/lib/config/feature.go`.

Define `FeatureUsageHookConfig` and `FeatureUsageConfig` in `pkg/lib/config/feature_usage.go`:

```go
type FeatureUsageHookConfig struct {
    URL   string `json:"url"`
    Match string `json:"match"`
}

type FeatureUsageConfig struct {
    Hooks  []FeatureUsageHookConfig `json:"hooks,omitempty"`
    Limits *FeatureUsageLimitsConfig `json:"limits,omitempty"`
}

type FeatureConfig struct {
    // existing fields...
    Usage *FeatureUsageConfig `json:"usage,omitempty"`
}
```

Schema rules:

- `usage.hooks[].url`: required, `x_hook_uri`
- `usage.hooks[].match`: required, enum of supported usage names plus `*`
- `usage.limits.<usage_name>[]`:
  - `quota`: required integer, `minimum: 0`
  - `period`: required enum `day | month`
  - `action`: required enum `alert | block`

Merge behavior in `FeatureUsageConfig.Merge(...)` must follow the spec:

- `hooks` append lower-precedence entries first, higher-precedence entries second
- each per-usage limit list is overridden as a whole by higher precedence

### 1.3 App config: `usage.alerts` + `usage.limits`

Add `Usage *UsageConfig \`json:"usage,omitempty"\`` to `pkg/lib/config/config.go`.

Define `UsageAlertConfig` and `UsageConfig` in `pkg/lib/config/usage.go`:

```go
type UsageAlertConfig struct {
    Type  string `json:"type"`
    Email string `json:"email,omitempty"`
    Match string `json:"match"`
}

type UsageConfig struct {
    Alerts []UsageAlertConfig `json:"alerts,omitempty"`
    Limits *UsageLimitsConfig `json:"limits,omitempty"`
}

type AppConfig struct {
    // existing fields...
    Usage *UsageConfig `json:"usage,omitempty"`
}
```

Schema rules:

- `usage.alerts[].type`: required, only `email`
- `usage.alerts[].email`: required when `type == email`
- `usage.alerts[].match`: required, enum of supported usage names plus `*`
- `usage.limits` shares the same schema as feature config

### 1.4 Backward compatibility for deprecated feature fields

Do not remove the old fields yet.

Keep parsing:

- `messaging.sms_usage`
- `messaging.email_usage`
- `messaging.whatsapp_usage`
- `admin_api.user_import_usage`
- `admin_api.user_export_usage`

Rename the existing legacy config type in `pkg/lib/config/feature_usage_limit.go`:

- `UsageLimitConfig` -> `Deprecated_UsageLimitConfig`

Implementation plan:

1. Parse both the new `usage` section and the deprecated fields.
2. During defaulting/normalization, translate deprecated fields into the unified in-memory usage representation.
3. If both old and new config define limits for the same usage name, the new `usage.limits.<name>` should win.
4. Keep the deprecated schema and tests until a later cleanup.

This keeps runtime logic on one model while preserving backward compatibility.

### 1.5 Config normalization call plan

The new `usage` section is the single runtime-facing representation, even while deprecated fields remain parseable.

Add these helper methods:

```go
// pkg/lib/config/usage.go
func (c *FeatureConfig) Migrate() *FeatureConfig
```

Config parsing flow:

1. `ParseFeatureConfig(...)`
   - schema validation accepts both deprecated fields and new `usage`.
   - decode YAML into `FeatureConfig`.
2. `SetFieldDefaults(config)`
   - existing field defaults still run.
   - `FeatureUsageConfig` defaults are initialized.
3. `(*FeatureConfig).Migrate()`
   - if `usage.limits.<name>` already exists, keep it.
   - otherwise convert deprecated `messaging.*_usage` / `admin_api.*_usage` into `usage.limits.<name>`.
   - legacy conversion maps:
     - enabled false -> no limits
     - enabled true -> one `block` limit with the same `quota` and `period`
4. `(*FeatureConfig).Merge(layer)`
   - existing legacy field merge behavior remains for compatibility.
   - `FeatureUsageConfig.Merge(...)` applies spec-defined append/override behavior for the new section.

App config flow:

1. `Parse(...)`
2. `SetFieldDefaults(appConfig)`

No app-config migration method is needed because deprecated usage config exists only on feature config.

## 2. Runtime Limit Evaluation

### 2.1 Current gap

Today `pkg/lib/usage/limit.go` only supports:

- one effective limit per usage name
- block-only behavior
- no alert-only thresholds
- no event dispatch

The new spec requires evaluating multiple configured limits for the same usage name, potentially mixing:

- `alert`
- `block`

Example:

```yaml
sms:
  - quota: 4
    period: month
    action: alert
  - quota: 900
    period: month
    action: block
```

### 2.2 Runtime model

Add this runtime descriptor in `pkg/lib/usage/limit.go`:

```go
type EffectiveUsageLimit struct {
    Name   model.UsageName
    Quota  int
    Period model.UsageLimitPeriod
    Action model.UsageLimitAction
}
```

Add these lookup helpers on `pkg/lib/config/config.go`:

```go
func (c *Config) UsageLimits(name model.UsageName) []EffectiveUsageLimit
func (c *Config) FeatureUsageHooks(name model.UsageName) []config.FeatureUsageHookConfig
func (c *Config) AppUsageAlerts(name model.UsageName) []config.UsageAlertConfig
```

`Config.UsageLimits(...)` converts from the separate config structs:

- `config.FeatureUsageLimitConfig`
- `config.UsageLimitConfig`

into one runtime-only `EffectiveUsageLimit` slice. The runtime layer is where the feature/app shapes are unified.

Add these limiter-local helpers in `pkg/lib/usage/limit.go`:

```go
// pkg/lib/usage/limit.go
func (l *Limiter) effectiveUsageLimits(name model.UsageName) []EffectiveUsageLimit
func (l *Limiter) usagePeriods() []model.UsageLimitPeriod
func (l *Limiter) limitsForPeriod(limits []EffectiveUsageLimit, period model.UsageLimitPeriod) []EffectiveUsageLimit
func (l *Limiter) minBlockQuota(limits []EffectiveUsageLimit) (int, bool)
func (l *Limiter) redisLimitKey(name model.UsageName, period model.UsageLimitPeriod) string
func (l *Limiter) usageHookURLs(name model.UsageName) []string
func (l *Limiter) usageAlertRecipients(name model.UsageName) []string
func (l *Limiter) makeUsageAlertTriggeredPayload(limit EffectiveUsageLimit, currentValue int) *nonblocking.UsageAlertTriggeredEventPayload
```

`redisLimitKey(name, period)` takes `model.UsageName`, translates it to the existing Redis key name used by current code, and builds the Redis key.

Those existing Redis key-name strings come from `usage.LimitName` constants. As part of this change, move those constants into the `usage` package itself, so they are defined in one place instead of being distributed across call-site files.

Add these constants in `pkg/lib/usage/limit_name.go`:

```go
const (
    LimitNameEmail      LimitName = "Email"
    LimitNameSMS        LimitName = "SMS"
    LimitNameWhatsapp   LimitName = "Whatsapp"
    LimitNameUserImport LimitName = "UserImport"
    LimitNameUserExport LimitName = "UserExport"
)
```

The new implementation keeps the translation from `model.UsageName` to `usage.LimitName` private inside `redisLimitKey(...)`; it does not expose `usage.LimitName` in the new limiter API.

The translation is:

- `sms` -> `SMS`
- `email` -> `Email`
- `whatsapp` -> `Whatsapp`
- `user_import` -> `UserImport`
- `user_export` -> `UserExport`

### 2.3 Reservation flow

Support different periods for the same usage name.

The runtime uses one Redis counter per usage name and period. There are two key shapes during and after migration:

```text
app:{app_id}:usage-limit:{translated_key_name}
app:{app_id}:usage-limit:{translated_key_name}:{period}
```

Compatibility rule:

- the limiter continues to use the existing legacy key without `:{period}` for monthly limits only
- the limiter uses the new `:{period}` key for non-monthly limits

Counting rule:

- the limiter always counts usage for every supported period of a usage name
- the supported periods are fixed by runtime policy, not derived from current config
- in this implementation, the counted periods are `day` and `month`
- this remains true even when only one of those periods has a configured limit
- this prevents incorrect counters when config is changed in the middle of a period

The translated Redis key names are the existing key-name strings already used by the current code:

- `sms` -> `SMS`
- `email` -> `Email`
- `whatsapp` -> `Whatsapp`
- `user_import` -> `UserImport`
- `user_export` -> `UserExport`

This preserves existing monthly usage levels on deploy while allowing additional periods to coexist.

Change the reservation script contract to return both `usage_before` and `usage_after`.

The script result is:

```text
{pass, usage_before, usage_after}
```

Where:

- `usage_before` is the usage before this reservation attempt
- `usage_after` is:
  - `usage_before + n` if reserve succeeds
  - `usage_before` if reserve fails because a block limit would be exceeded

### 2.3.1 Lua script plan

Modify the existing script in `pkg/lib/usage/limit.go` to use one Lua script for both quota-checked and unconditional increment paths.

The script returns:

```text
{pass, usage_before, usage_after}
```

Script inputs:

- `KEYS[1]`: usage counter key
- `ARGV[1]`: `n`
- `ARGV[2]`: `reset_time_unix`
- `ARGV[3]`: `quota`

Quota sentinel:

- `quota == -1` means no quota check
- `quota >= 0` means enforce quota check

Script logic:

1. Read current counter from `KEYS[1]`.
2. Treat missing or expired key as `0`.
3. Set `usage_before` to the current value.
4. Parse `quota` from `ARGV[3]`.
5. If `quota >= 0` and `usage_before + n > quota`:
   - keep the stored value unchanged
   - set `usage_after = usage_before`
   - return `{false, usage_before, usage_after}`
6. Otherwise:
   - write `usage_after = usage_before + n`
   - `SET` the new value
   - `EXPIREAT` the key to `reset_time_unix`
   - return `{true, usage_before, usage_after}`

This one script is called by both wrapper methods:

```go
reserveWithQuota(ctx, key, n, quota, resetTime)
incrementWithoutQuota(ctx, key, n, resetTime)
```

`reserveWithQuota(...)` passes the real quota in `ARGV[3]`.

`incrementWithoutQuota(...)` calls the same script with `quota = -1` in `ARGV[3]`. In that case, the script always returns `pass = true`.

Implementation details:

- keep one script in `pkg/lib/usage/limit.go`
- keep the script operating on exactly one key per call
- do not combine multiple periods into one Redis script call
- one Go `reservePeriod(...)` call maps to one script execution against one key
- `Cancel(...)` remains a normal Redis decrement path and does not need a Lua script change

Exact method call logic:

1. `Reserve(ctx, name, n)`
   - call `limits := l.effectiveUsageLimits(name)`
   - call `periods := l.usagePeriods()`
2. For each `period` in `periods`, `Reserve(...)` calls:

```go
periodLimits := l.limitsForPeriod(limits, period)
result, err := l.reservePeriod(ctx, name, period, n, periodLimits)
```

3. `reservePeriod(ctx, name, period, n, periodLimits)` does:

```go
resetTime := ComputeResetTime(l.Clock.NowUTC(), period)
key := l.redisLimitKey(name, period)
blockQuota, hasBlockQuota := l.minBlockQuota(periodLimits)
```

4. Inside `reservePeriod(...)`:
   - if `hasBlockQuota`:

```go
pass, before, after, err := l.reserveWithQuota(ctx, key, n, blockQuota, resetTime)
```

   - if `!hasBlockQuota`:

```go
before, after, err := l.incrementWithoutQuota(ctx, key, n, resetTime)
pass = true
```

5. `reservePeriod(...)` returns one `periodReservationResult`:

```go
type periodReservationResult struct {
    Period     model.UsageLimitPeriod
    Limits     []EffectiveUsageLimit
    Key        string
    ResetTime  time.Time
    Pass       bool
    Before     int
    After      int
    Taken      int
}
```

6. Back in `Reserve(...)`:
   - append each successful `periodReservationResult` to the aggregate reservation
   - stop immediately when one `periodReservationResult.Pass == false`
7. If one period blocks:
   - call `Cancel(...)`-equivalent rollback for earlier successful period results from this request
   - call `evaluateUsageTriggers(ctx, name, blockedResult.Period, blockedResult.Before, blockedResult.After, true, blockedResult.Limits)`
   - return `ErrUsageLimitExceeded(name)`
8. If all periods succeed:
   - for each `periodReservationResult`, call:

```go
l.evaluateUsageTriggers(
    ctx,
    name,
    result.Period,
    result.Before,
    result.After,
    false,
    result.Limits,
)
```

   - return one aggregate reservation for later `Cancel(...)`

Important detail:

- `reserveWithQuota(...)` is called once per period, not once per limit.
- if a period has multiple block limits, `minBlockQuota(periodLimits)` picks the smallest block quota for admission control.
- all limits in that period are still evaluated afterward against the same `before` / `after` values.
- periods are not derived from current config; runtime always updates both day and month counters

This supports mixed periods, keeps block admission atomic within each period, and preserves existing usage counters across deployment.

The key point is that all configured period counters are updated on every successful request. For example, if SMS has both daily and monthly limits, a successful SMS send increments both counters.

### 2.3.2 Deployment compatibility for existing usage levels

Deployment compatibility is handled as follows:

1. Existing usage counters remain in the current Redis keys:
   - `app:{app_id}:usage-limit:{translated_key_name}`
2. The new version keeps using those keys for monthly limits.
3. If a usage name has additional non-monthly limits, the new version stores those counters in new period-specific keys:
   - `app:{app_id}:usage-limit:{translated_key_name}:{period}`
4. `FeatureConfig.Migrate()` maps legacy configs into the new `usage.limits` structure without changing the existing monthly semantics.
5. No Redis backfill job is needed.
6. No copy of existing counters is needed.
7. A deployment does not reset in-progress monthly usage counts.
8. Non-monthly counters are always maintained by the new runtime and are incremented on every successful usage event.

Example:

- before deployment, SMS monthly usage is stored in `app:{app_id}:usage-limit:SMS`
- after deployment:
  - monthly SMS limits continue reading `app:{app_id}:usage-limit:SMS`
  - daily SMS limits read `app:{app_id}:usage-limit:SMS:day`

This means a project that already used 700 monthly SMS before deployment still reads 700 monthly SMS after deployment. Daily SMS tracking uses its own key and continues to accumulate independently.

### 2.4 Trigger rules

For each configured limit:

- trigger when usage crosses from below quota to at least quota
- `before < quota && after >= quota`

Behavior by action:

- `alert`:
  - reservation succeeds
  - emit `usage.alert.triggered`
  - deliver matching `usage.hooks`
  - deliver matching `usage.alerts`
- `block`:
  - if reservation succeeds and crosses the quota, emit `usage.alert.triggered`
  - if reservation fails because the quota is already exhausted, do not emit `usage.alert.triggered`

No marker key is needed. Triggering is based only on threshold crossing:

```text
before < quota && after >= quota
```

### 2.5 Limiter API changes

Modify `pkg/lib/usage/limit.go` to define these interfaces and methods:

```go
type EventService interface {
    DispatchEventImmediately(ctx context.Context, payload event.NonBlockingPayload) error
}

type UsageAlertEmailService interface {
    Send(ctx context.Context, recipients []string, payload *nonblocking.UsageAlertTriggeredEventPayload) error
}

type Limiter struct {
    Clock        clock.Clock
    AppID        config.AppID
    Redis        *appredis.Handle
    EventService EventService
    UsageAlertEmailService UsageAlertEmailService

    EffectiveConfig *config.Config
}

func (l *Limiter) Reserve(ctx context.Context, name model.UsageName, n int) (*Reservation, error)
func (l *Limiter) evaluateUsageTriggers(ctx context.Context, name model.UsageName, period model.UsageLimitPeriod, before, after int, rejected bool, limits []EffectiveUsageLimit) error
```

Limiter callers stop passing a single `*config.Deprecated_UsageLimitConfig`. The limiter resolves configured limits internally from effective config, because the new model is multi-limit and shared across feature/app config.

Backward-compatibility requirement for API errors:

- update `ErrUsageLimitExceeded(...)` to return both:
  - `name`: the legacy `usage.LimitName` string
  - `usage_name`: the new `model.UsageName` string
- include `period`: the new `model.UsageLimitPeriod` value
- add a code comment on the old `name` field indicating that it is a legacy field kept for backward compatibility
- the conversion from `model.UsageName` to legacy `usage.LimitName` remains inside `pkg/lib/usage`

### 2.6 Runtime method call plan

#### A. Messaging flow

Current entry points:

- `pkg/lib/messaging/limits.go`
  - `checkEmail(...)`
  - `checkSMS(...)`
  - `checkWhatsapp(...)`

Call plan:

1. `checkSMS(...)`
   - call `l.UsageLimiter.Reserve(ctx, model.UsageNameSMS, 1)`
2. `pkg/lib/usage/limit.go`
   - `Reserve(ctx, name, 1)`
   - resolve all configured limits for `sms`
   - group limits by period
3. For each counted period:
   - resolve Redis key for that usage+period
   - reserve or increment against that key
   - collect per-period result
4. `Reserve(...)`
   - if one period blocks:
     - cancel earlier successful period reservations from this request
     - call `evaluateUsageTriggers(...)` only for the blocking period result
     - return `ErrUsageLimitExceeded(name)`
   - if all periods succeed:
     - call `evaluateUsageTriggers(...)` for each period result
     - return aggregate reservation
5. `checkSMS(...)`
   - continues with rate-limit checks
   - on downstream failure, `Cancel(...)` still rolls back the usage reservation when one was taken

Call-site cleanup in this file:

- remove local `usageLimitEmail`
- remove local `usageLimitSMS`
- remove local `usageLimitWhatsapp`
- use `usage.LimitNameEmail`, `usage.LimitNameSMS`, and `usage.LimitNameWhatsapp` only inside `pkg/lib/usage` if needed for translation or backward-compatibility internals

The same shape applies to:

- `checkEmail(...)`
- `checkWhatsapp(...)`

#### B. User import flow

Current entry point:

- `pkg/lib/userimport/job.go`
  - `JobManager.EnqueueJob(...)`

Call plan:

1. `EnqueueJob(...)`
   - call `m.UsageLimiter.Reserve(ctx, model.UsageNameUserImport, len(request.Records))`
2. `Limiter.Reserve(...)`
   - resolve `user_import` limits
   - evaluate all matching limits for the relevant period
3. On success:
   - enqueue tasks as today
4. On block:
   - return `ErrUsageLimitExceeded(name)`

This preserves the current charging unit for user import: number of imported records.

Call-site cleanup in this file:

- remove local `usageLimitUserImport`

#### C. User export flow

Current entry point:

- `pkg/admin/transport/handler_user_export_create.go`
  - `UserExportCreateHandler.handle(...)`

Call plan:

1. `handle(...)`
   - call `h.UsageLimiter.Reserve(ctx, model.UsageNameUserExport, 1)`
2. `Limiter.Reserve(...)`
   - resolve `user_export` limits
   - evaluate all matching limits
3. On success:
   - enqueue export task as today
4. On block:
   - return `ErrUsageLimitExceeded(name)`

This preserves the current charging unit for user export: number of export requests.

Call-site cleanup in this file:

- remove local `usageLimitUserExport`

#### D. Detailed limiter internals

Add these internal methods in `pkg/lib/usage/limit.go`:

```go
func (l *Limiter) Reserve(ctx context.Context, name model.UsageName, n int) (*Reservation, error)
func (l *Limiter) reservePeriod(ctx context.Context, name model.UsageName, period model.UsageLimitPeriod, n int, limits []EffectiveUsageLimit) (*periodReservationResult, error)
func (l *Limiter) reserveWithQuota(ctx context.Context, key string, n int, quota int, resetTime time.Time) (pass bool, before int64, after int64, err error)
func (l *Limiter) incrementWithoutQuota(ctx context.Context, key string, n int, resetTime time.Time) (before int64, after int64, err error)
func (l *Limiter) evaluateUsageTriggers(ctx context.Context, name model.UsageName, period model.UsageLimitPeriod, before, after int, rejected bool, limits []EffectiveUsageLimit) error
func (l *Limiter) maybeDispatchUsageAlert(ctx context.Context, limit EffectiveUsageLimit, currentValue int, rejected bool) error
```

Call sequence inside `Reserve(...)`:

1. Load `limits := l.effectiveUsageLimits(name)`.
2. If `len(limits) == 0`:
   - return zero-taken reservation or a simple no-op reservation.
3. Resolve `periods := l.usagePeriods()`.
4. For each `period` in `periods`:
   - resolve `periodLimits := l.limitsForPeriod(limits, period)`
   - call `reservePeriod(ctx, name, period, n, periodLimits)`
   - append successful results to the aggregate reservation
5. If one `reservePeriod(...)` returns blocked:
   - cancel all earlier successful period reservations from this request
   - call `evaluateUsageTriggers(ctx, name, period, before, after, true, periodLimits)` for the blocking period only
   - return `ErrUsageLimitExceeded(name)`
6. If all period reservations succeed:
   - for each period result, call `evaluateUsageTriggers(ctx, name, period, before, after, false, periodLimits)`
   - return aggregate reservation for later `Cancel(...)`

This matches the spec while preserving legacy usage counts, adding support for additional periods, and ensuring both day and month counters are updated on every successful usage event.

## 3. Event and Delivery

### 3.1 `usage.alert.triggered` payload

Align implementation with [`docs/specs/event.md`](../specs/event.md#usagealerttriggered).

Create or update `pkg/api/event/nonblocking/usage_alert_triggered.go`:

```go
const UsageAlertTriggered event.Type = "usage.alert.triggered"

type UsageAlertPayload struct {
    Name         model.UsageName        `json:"name"`
    Action       model.UsageLimitAction `json:"action"`
    Period       model.UsageLimitPeriod `json:"period"`
    Quota        int                       `json:"quota"`
    CurrentValue int                       `json:"current_value"`
}

type UsageAlertTriggeredEventPayload struct {
    Usage    UsageAlertPayload `json:"usage"`
    HookURLs []string          `json:"-"`
}
```

Event behavior:

- `context.triggered_by = system`
- `ForHook() == true`
- no audit log support unless explicitly required later

### 3.2 Feature `usage.hooks`

`usage.hooks` is not the same as `hook.non_blocking_handlers`.

Implementation:

1. Filter `usage.hooks` by `match == usage name` or `match == "*"` when a limit triggers.
2. Attach matched URLs to the event payload through `ExtraHookURLs()`.
3. Reuse the existing non-blocking event dispatch path so the same event can be:
   - delivered to `usage.hooks`
   - delivered to ordinary `hook.non_blocking_handlers` listening to `usage.alert.triggered`

This preserves the spec statement that the event is emitted regardless of whether special usage delivery is configured.

Detailed call chain:

1. `Limiter.evaluateUsageTriggers(...)`
2. For each crossed configured limit:
   - call `payload := makeUsageAlertTriggeredPayload(limit, currentValue)`
   - call `hookURLs := l.usageHookURLs(name)`
   - attach only URLs whose `match` equals the usage name or `*`
3. `Limiter.maybeDispatchUsageAlert(...)`
   - set payload extra hook URLs
   - call `EventService.DispatchEventImmediately(ctx, payload)`
4. `pkg/lib/event/service.go`
   - dispatch non-blocking event to sinks
5. `pkg/lib/hook/sink.go`
   - standard non-blocking handlers listening to `usage.alert.triggered` receive the event
   - payload-provided `ExtraHookURLs()` are also delivered

Important behavior:

- `usage.hooks` does not replace ordinary non-blocking event handlers.
- there is one event payload shape.
- hook delivery failures are non-blocking and should only be logged.

### 3.3 App `usage.alerts`

`usage.alerts` is app-config-only email delivery, separate from event hooks.

Implementation:

1. Filter `usage.alerts` by `match == usage name` or `match == "*"` when a limit triggers.
2. Collect email recipients from matching entries with `type == email`.
3. Send usage alert email through a dedicated service.
4. Deduplicate recipients before sending.
5. Do not count these emails toward `email` usage.

This reuses the existing mail infrastructure, but does not route through normal user-facing notification flows that affect messaging usage counters.

Detailed call chain:

1. `Limiter.evaluateUsageTriggers(...)`
2. For each crossed configured limit:
   - call `recipients := l.usageAlertRecipients(name)`
3. `usageAlertRecipients(name)`:
   - reads `appConfig.Usage.Alerts`
   - filters by `match == usage name` or `match == "*"`
   - keeps entries with `type == email`
   - deduplicates recipient email addresses
4. `Limiter.maybeDispatchUsageAlert(...)`
   - dispatch event through `EventService`
   - call `UsageAlertEmailService.Send(ctx, recipients, payload)` if recipients are non-empty
5. `UsageAlertEmailService.Send(...)`
   - renders one usage-alert email message
   - sends through a mail-sending path that bypasses usage checks entirely

Important behavior:

- email delivery is in addition to hook/event delivery
- email send failure is non-blocking for the main request
- send errors should be logged and should not turn an allowed usage reservation into a request failure
- this service must not go through the normal messaging sender path if that path performs usage checks, otherwise it would recurse into usage limiting again

### 3.3.1 `UsageAlertEmailService` implementation plan

Implement `UsageAlertEmailService` in `pkg/lib/usage/usage_alert_email_service.go`.

Add these types and methods:

```go
type UsageAlertEmailService interface {
    Send(ctx context.Context, recipients []string, payload *nonblocking.UsageAlertTriggeredEventPayload) error
}

type UsageAlertEmailServiceImpl struct {
    TranslationService TranslationService
    MailSender         MailSender
}

func (s *UsageAlertEmailServiceImpl) Send(ctx context.Context, recipients []string, payload *nonblocking.UsageAlertTriggeredEventPayload) error
```

`UsageAlertEmailServiceImpl.Send(...)` does:

1. Return immediately if `len(recipients) == 0`.
2. Build translation variables from the event payload:
   - `usage.name`
   - `usage.action`
   - `usage.period`
   - `usage.quota`
   - `usage.current_value`
3. Render the usage alert message through translation/message rendering.
4. If DEV mode would skip normal email delivery, skip usage alert email delivery in the same way.
5. Send the rendered email to all recipients through a lower-level mail sender that does not perform usage checks.
6. Do not call the normal messaging sender path if that path performs usage checks.
7. Do not call messaging usage counting code.

Add translation and template integration in these files:

- `pkg/lib/translation/message_declarations.go`
- `pkg/lib/translation/variables.go`
- `pkg/lib/translation/service.go`

Change `pkg/lib/translation/variables.go` explicitly:

```go
type UsageAlertTemplateVariables struct {
    Name         model.UsageName
    Action       model.UsageLimitAction
    Period       model.UsageLimitPeriod
    Quota        int
    CurrentValue int
}

type PartialTemplateVariables struct {
    // existing fields...
    Usage *UsageAlertTemplateVariables
}

type PreparedTemplateVariables struct {
    // existing fields...
    Usage *UsageAlertTemplateVariables
}
```

Change `pkg/lib/translation/service.go` explicitly:

- copy `PartialTemplateVariables.Usage` into `PreparedTemplateVariables.Usage`
- do not rename the key from `usage`; keep the template variable path aligned with the event payload shape

`UsageAlertEmailServiceImpl.Send(...)` builds this variable payload:

```go
vars := &translation.PartialTemplateVariables{
    Usage: &translation.UsageAlertTemplateVariables{
        Name:         payload.Usage.Name,
        Action:       payload.Usage.Action,
        Period:       payload.Usage.Period,
        Quota:        payload.Usage.Quota,
        CurrentValue: payload.Usage.CurrentValue,
    },
}
```

Template variable paths are therefore:

- `.Usage.name`
- `.Usage.action`
- `.Usage.period`
- `.Usage.quota`
- `.Usage.current_value`

Add these resources:

- `resources/authgear/templates/en/messages/usage_alert_email.txt.gotemplate`
- `resources/authgear/templates/en/messages/usage_alert_email.mjml.gotemplate`
- `resources/authgear/templates/en/translation.json`

Generated artifacts:

- `resources/authgear/templates/en/messages/usage_alert_email.txt`
- `resources/authgear/templates/en/messages/usage_alert_email.mjml`
- `resources/authgear/templates/en/messages/usage_alert_email.html`

The implementation must define one translation message spec for usage alert emails and wire it to the generated text/html templates.

Important implementation constraint:

- `UsageAlertEmailServiceImpl` must bypass usage checking entirely
- otherwise usage-alert email sending would recurse back into usage limiting and create an infinite loop
- the implementation therefore may not be able to reuse the existing messaging sender path in the repo
- `UsageAlertEmailServiceImpl` must also preserve the existing DEV-mode skip behavior for email delivery

### 3.4 Event dispatch method plan

Add these methods on `UsageAlertTriggeredEventPayload`:

```go
func (e *UsageAlertTriggeredEventPayload) NonBlockingEventType() event.Type
func (e *UsageAlertTriggeredEventPayload) UserID() string
func (e *UsageAlertTriggeredEventPayload) GetTriggeredBy() event.TriggeredByType
func (e *UsageAlertTriggeredEventPayload) FillContext(ctx *event.Context)
func (e *UsageAlertTriggeredEventPayload) ForHook() bool
func (e *UsageAlertTriggeredEventPayload) ForAudit() bool
func (e *UsageAlertTriggeredEventPayload) RequireReindexUserIDs() []string
func (e *UsageAlertTriggeredEventPayload) DeletedUserIDs() []string
func (e *UsageAlertTriggeredEventPayload) ExtraHookURLs() []string
```

Behavior:

- `NonBlockingEventType()` -> `usage.alert.triggered`
- `UserID()` -> `""`
- `GetTriggeredBy()` -> `event.TriggeredByTypeSystem`
- `ForHook()` -> `true`
- `ForAudit()` -> `false`
- `ExtraHookURLs()` -> feature `usage.hooks` URLs attached by limiter

## 4. Feature Config Merge Semantics

The current code merges deprecated usage fields individually inside `MessagingFeatureConfig.Merge(...)` and `AdminAPIFeatureConfig.Merge(...)`.

The new behavior must live in `FeatureUsageConfig.Merge(...)`.

Rules:

1. `usage.hooks` append from lower precedence to higher precedence.
2. `usage.limits.<usage_name>` is overridden as a whole by higher precedence.
3. Legacy fields continue to merge with existing behavior for backward compatibility, but normalization should prefer the new `usage` section.

Example:

- plan feature config defines `usage.limits.sms = [{quota: 900, period: month, action: block}]`
- project feature config defines `usage.limits.sms = [{quota: 800, period: month, action: block}]`
- effective config must keep only the project `sms` limits list

Detailed merge call plan:

1. Site feature config is parsed.
2. `SetFieldDefaults(siteFeatureConfig)`.
3. `siteFeatureConfig.Migrate()`.
4. Plan feature config is parsed.
5. `SetFieldDefaults(planFeatureConfig)`.
6. `planFeatureConfig.Migrate()`.
7. Project feature config is parsed.
8. `SetFieldDefaults(projectFeatureConfig)`.
9. `projectFeatureConfig.Migrate()`.
10. Existing feature-config merge pipeline calls:
   - `site.Merge(plan)`
   - `merged.Merge(project)`
11. `FeatureUsageConfig.Merge(layer)` runs during `FeatureConfig.Merge(...)`.
12. Inside `FeatureUsageConfig.Merge(...)`:
   - if `layer.Usage.Hooks` is non-empty:
     - append layer hooks after existing hooks
   - for each supported usage name:
     - if `layer.Usage.Limits.<name>` is non-nil, replace the whole current list

This preserves the spec ordering for `usage.hooks` while making migration happen exactly once per layer before merge.

## 5. Migration Strategy

Implement in phases so runtime behavior stays stable.

### Phase 1: Config model and normalization

- add shared usage types and schema
- parse `usage` in both feature and app config
- normalize deprecated feature fields into the new in-memory representation
- add merge/default tests

No runtime behavior change yet beyond config parsing.

### Phase 2: Event payload and hook delivery

- add `usage.alert.triggered`
- allow `hook.non_blocking_handlers[].events` to include `usage.alert.triggered`
- implement `usage.hooks` filtering and extra hook URL delivery

### Phase 3: Multi-limit runtime enforcement

- refactor limiter to resolve limits from effective config
- support multiple limits per usage
- support `alert` and `block`

### Phase 4: App alert emails

- implement `usage.alerts` filtering and email sending
- ensure usage alert emails do not increment email usage
- add templates and tests

Implementation order:

1. Land parsing/schema/types first.
2. Land event type and hook wiring second.
3. Refactor limiter call sites from `Reserve(ctx, name, cfg)` / `ReserveN(...)` to `Reserve(ctx, name, n)`.
4. Land multi-limit runtime logic.
5. Land app alert email delivery last.

## 6. File-Level Change Plan

### 6.1 Config

Create or modify:

- `pkg/lib/config/feature.go`
- `pkg/lib/config/config.go`
- `pkg/lib/config/feature_usage_limit.go`
- `pkg/lib/config/feature_usage.go`
- `pkg/lib/config/usage.go`
- `pkg/api/model/usage.go`
- `pkg/lib/config/feature_admin_api.go`
- `pkg/lib/config/hook.go`
- `pkg/lib/config/testdata/parse_feature_tests.yaml`
- `pkg/lib/config/testdata/merge_feature.yaml`
- `pkg/lib/config/feature_test.go`

Expected changes:

- add `usage` schemas to feature config and app config
- rename the existing legacy `UsageLimitConfig` to `Deprecated_UsageLimitConfig`
- keep feature/app usage config structs separate even when fields are parallel
- place feature usage config structs in `feature_usage.go` to match the repo pattern
- keep deprecated schemas
- add normalization helpers from legacy fields to new usage model
- add merge coverage for append-vs-override semantics

### 6.2 Runtime and eventing

Create or modify:

- `pkg/lib/usage/limit.go`
- `pkg/lib/usage/limit_name.go`
- `pkg/lib/usage/errors.go`
- `pkg/lib/usage/deps.go`
- `pkg/api/event/event.go`
- `pkg/api/event/nonblocking/usage_alert_triggered.go`
- `pkg/lib/event/service.go`
- `pkg/lib/hook/sink.go`
- `pkg/lib/usage/limit_test.go`
- `pkg/lib/event/service_test.go`
- `pkg/lib/hook/sink_test.go`
- `pkg/lib/messaging/limits.go`
- `pkg/lib/userimport/job.go`
- `pkg/admin/transport/handler_user_export_create.go`

Expected changes:

- multi-limit evaluation
- event emission
- extra hook URL delivery
- update caller interfaces to stop passing legacy single-limit config objects

### 6.3 Email delivery

Create or modify:

- `pkg/lib/usage/usage_alert_email_service.go`
- `pkg/lib/translation/message_declarations.go`
- `pkg/lib/translation/variables.go`
- `pkg/lib/translation/service.go`
- `resources/authgear/templates/en/messages/usage_alert_email.txt.gotemplate`
- `resources/authgear/templates/en/messages/usage_alert_email.mjml.gotemplate`
- `resources/authgear/templates/en/translation.json`

Generated artifacts:

- `resources/authgear/templates/en/messages/usage_alert_email.txt`
- `resources/authgear/templates/en/messages/usage_alert_email.mjml`
- `resources/authgear/templates/en/messages/usage_alert_email.html`

### 6.4 Wiring

Whenever a commit changes dependency wiring, run `make generate` in that same commit.

Generated files that may be updated by those commits:

- `pkg/auth/wire_gen.go`
- `pkg/admin/wire_gen.go`
- `cmd/authgear/background/wire_gen.go`

## 7. Test Plan

### 7.1 Config parsing

Add tests for:

- valid `usage.hooks`
- valid `usage.alerts`
- valid `usage.limits` for each supported usage name
- invalid `match`
- invalid `action`
- invalid `type`
- merge semantics for `usage.hooks`
- override semantics for each `usage.limits.<usage_name>`
- coexistence of legacy fields and new `usage` section

### 7.2 Runtime limits

Add tests for:

- alert-only limit crossing emits event but does not block
- block limit crossing emits event and then blocks subsequent over-quota requests
- mixed limits on the same usage
- multiple alert thresholds crossed by one large `ReserveN`
- day and month reset boundaries

### 7.3 Delivery

Add tests for:

- `usage.hooks.match` exact match
- `usage.hooks.match: "*"`
- ordinary `hook.non_blocking_handlers` still receive `usage.alert.triggered`
- `usage.alerts.match` exact match
- `usage.alerts.match: "*"`
- recipient deduplication
- usage alert email sending does not count toward `email` usage

### 7.4 Backward compatibility

Add tests ensuring legacy config still produces the same effective blocking behavior when no new `usage` section is present.

### 7.5 E2E coverage

Add e2e tests under `e2e/tests/` to cover:

- feature config soft limit
- feature config hard limit
- project config soft limit
- project config hard limit
- feature `usage.hooks` URL trigger
- project `usage.alerts` email trigger

The e2e coverage must verify both triggering behavior and non-triggering behavior:

- soft limit triggers alert delivery but does not block
- hard limit triggers alert delivery and then blocks
- matching hook URL receives `usage.alert.triggered`
- matching alert email is sent for `usage.alerts`
- non-matching hook and alert configs do not receive the event

## 8. Fixed Behavioral Decisions

The implementation uses these fixed decisions:

1. Multiple limits with the same `(period, quota, action)` are allowed in config and deduplicated at runtime before dispatch.
2. If no `block` limit exists for a usage, reservation always succeeds and only alert limits are evaluated.
3. A successful reservation that lands exactly on a `block` threshold emits `usage.alert.triggered` immediately.
4. `ErrUsageLimitExceeded(...)` returns `name`, `usage_name`, and `period`; `name` keeps the legacy `usage.LimitName` value for backward compatibility, and `usage_name` / `period` use the new model values.

## 9. Implementation Order

1. Land config schema and normalization.
2. Land event type and hook delivery support.
3. Refactor limiter to multi-limit evaluation.
4. Land app alert email delivery.
5. Remove deprecated fields only in a later follow-up after migration is complete.

## 10. Atomic Commit Plan

Make the implementation in these atomic commits:

### Commit 1: Add shared usage enums and config structs

Files:

- `pkg/api/model/usage.go`
- `pkg/lib/config/feature_usage.go`
- `pkg/lib/config/usage.go`
- `pkg/lib/config/feature.go`
- `pkg/lib/config/config.go`

Changes:

- add `model.UsageName`
- add `model.UsageLimitPeriod`
- add `model.UsageLimitAction`
- add feature-config usage structs in `feature_usage.go`
- add app-config usage structs in `usage.go`
- add `FeatureConfig.Usage`
- add `AppConfig.Usage`

This commit does not change runtime behavior.

### Commit 2: Rename legacy usage-limit config and add migration

Files:

- `pkg/lib/config/feature_usage_limit.go`
- `pkg/lib/config/feature_usage.go`
- `pkg/lib/config/feature.go`
- `pkg/lib/config/feature_messaging.go`
- `pkg/lib/config/feature_admin_api.go`

Changes:

- rename `UsageLimitConfig` to `Deprecated_UsageLimitConfig`
- update legacy feature-config fields to use `Deprecated_UsageLimitConfig`
- add `(*FeatureConfig).Migrate()`
- migrate deprecated usage fields into `FeatureConfig.Usage.Limits`

This commit still does not change runtime behavior.

### Commit 3: Add config schema, merge, and parsing tests

Files:

- `pkg/lib/config/feature_usage.go`
- `pkg/lib/config/usage.go`
- `pkg/lib/config/hook.go`
- `pkg/lib/config/testdata/parse_feature_tests.yaml`
- `pkg/lib/config/testdata/merge_feature.yaml`
- `pkg/lib/config/feature_test.go`

Changes:

- add feature-config usage schema
- add app-config usage schema
- add `FeatureUsageConfig.Merge(...)`
- allow `usage.alert.triggered` in hook config enum if needed by schema split
- add parsing and merge tests for the new usage config

This commit makes config parsing and merge behavior testable before runtime changes.

### Commit 4: Centralize legacy usage limit-name constants

Files:

- `pkg/lib/usage/limit_name.go`
- `pkg/lib/messaging/limits.go`
- `pkg/lib/userimport/job.go`
- `pkg/admin/transport/handler_user_export_create.go`

Changes:

- move distributed `usage.LimitName` constants into `pkg/lib/usage/limit_name.go`
- remove local duplicates from call sites

This commit is mechanical and keeps later runtime commits smaller.

### Commit 5: Add usage alert event payload and hook delivery support

Files:

- `pkg/api/event/event.go`
- `pkg/api/event/nonblocking/usage_alert_triggered.go`
- `pkg/lib/config/hook.go`
- `pkg/lib/event/service.go`
- `pkg/lib/hook/sink.go`
- `pkg/lib/event/service_test.go`
- `pkg/lib/hook/sink_test.go`

Changes:

- add `usage.alert.triggered` event payload
- make the payload hook-deliverable
- allow normal non-blocking handlers to subscribe to `usage.alert.triggered`
- support extra hook URLs on this event path
- run `make generate` in this commit if dependency wiring changes

This commit lands the event surface independently from limiter refactoring.

### Commit 6: Refactor limiter API to use `model.UsageName`

Files:

- `pkg/lib/usage/limit.go`
- `pkg/lib/usage/errors.go`
- `pkg/lib/usage/deps.go`
- `pkg/lib/messaging/limits.go`
- `pkg/lib/userimport/job.go`
- `pkg/admin/transport/handler_user_export_create.go`

Changes:

- change limiter entry point to `Reserve(ctx, name model.UsageName, n int)`
- remove passing legacy single-limit config objects into limiter
- update `ErrUsageLimitExceeded(...)` to return legacy `name` plus new `usage_name` and `period`
- add a code comment that `name` is the legacy field
- keep runtime behavior otherwise as close as possible before multi-period logic lands
- run `make generate` in this commit if dependency wiring changes

This commit isolates the call-site API migration from the deeper limiter changes.

### Commit 7: Implement Redis key translation and multi-period counting

Files:

- `pkg/lib/usage/limit.go`
- `pkg/lib/usage/limit_test.go`

Changes:

- add `redisLimitKey(name model.UsageName, period model.UsageLimitPeriod)`
- translate `model.UsageName` to existing Redis key-name strings internally
- count both day and month on every successful usage event
- preserve legacy monthly key compatibility
- add tests for key selection and deploy compatibility behavior

This commit lands the storage model for mixed periods.

### Commit 8: Implement single-script reservation and threshold evaluation

Files:

- `pkg/lib/usage/limit.go`
- `pkg/lib/usage/limit_test.go`

Changes:

- modify the Lua script to return `{pass, usage_before, usage_after}`
- use `quota = -1` to indicate no quota check
- add `reservePeriod(...)`
- evaluate alert/block threshold crossing with `before < quota && after >= quota`
- rollback earlier successful period reservations when a later period blocks

This commit lands the core limiter behavior required by the spec.

### Commit 9: Connect limiter to hooks and event dispatch

Files:

- `pkg/lib/usage/limit.go`
- `pkg/lib/event/service.go`
- `pkg/lib/hook/sink.go`
- `pkg/lib/usage/limit_test.go`

Changes:

- resolve matching feature `usage.hooks`
- dispatch `usage.alert.triggered` from limiter
- verify exact-match and wildcard hook delivery from limiter-triggered events

This commit completes feature-config delivery.

### Commit 10: Add app alert email delivery

Files:

- `pkg/lib/usage/usage_alert_email_service.go`
- `pkg/lib/usage/limit.go`
- `pkg/lib/translation/message_declarations.go`
- `pkg/lib/translation/variables.go`
- `pkg/lib/translation/service.go`
- `resources/authgear/templates/en/messages/usage_alert_email.txt.gotemplate`
- `resources/authgear/templates/en/messages/usage_alert_email.mjml.gotemplate`
- `resources/authgear/templates/en/translation.json`
- `pkg/lib/usage/limit_test.go`

Changes:

- resolve matching `usage.alerts`
- implement `UsageAlertEmailServiceImpl`
- build translation variables from `UsageAlertTriggeredEventPayload`
- add `UsageAlertTemplateVariables`
- add `Usage` to `PartialTemplateVariables` and `PreparedTemplateVariables`
- copy `Usage` through translation preparation unchanged
- render one usage alert email message through translation service
- preserve the existing DEV-mode skip behavior for email delivery
- send usage alert emails through a lower-level mail sender that bypasses usage checks
- deduplicate recipients
- ensure these emails do not count toward `email` usage
- run `make generate` in this commit because this commit is expected to change dependencies

This commit completes app-config delivery.

### Commit 11: Generate email artifacts

Files:

- `resources/authgear/templates/en/messages/usage_alert_email.txt`
- `resources/authgear/templates/en/messages/usage_alert_email.mjml`
- `resources/authgear/templates/en/messages/usage_alert_email.html`

Changes:

- generate email template artifacts

This commit is generated template output only.

### Commit 12: Final cleanup and backward-compatibility verification

Files:

- `pkg/lib/config/feature_test.go`
- `pkg/lib/usage/limit_test.go`
- any touched runtime/config files that need final cleanup

Changes:

- add any missing backward-compatibility assertions
- remove dead helper code or comments introduced during the refactor
- keep deprecated config support intact

This commit is a small cleanup pass only. It does not remove deprecated feature config fields.

### Commit 13: Add end-to-end tests for usage limits and delivery

Files:

- `e2e/tests/usage/*.test.yaml`
- any shared e2e fixtures needed by those tests

Changes:

- add e2e tests for feature-config soft limits
- add e2e tests for feature-config hard limits
- add e2e tests for project-config soft limits
- add e2e tests for project-config hard limits
- add e2e tests for hook URL delivery
- add e2e tests for alert email delivery

Required e2e coverage:

- feature soft limit + hook URL trigger
- feature hard limit + hook URL trigger
- project soft limit + email alert trigger
- project hard limit + email alert trigger

This commit is dedicated to end-to-end verification of the final behavior.
