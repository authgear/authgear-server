# Fraud Protection Implementation Plan

## Context

This plan implements the SMS pumping fraud protection feature as specified in `docs/specs/fraud-protection.md`. The feature detects and optionally blocks suspicious SMS OTP patterns (non-rotating IP attacks and rotating IP attacks) by evaluating 5 warning types against adaptive thresholds. It is a new, greenfield feature with no existing code.

Implementation is planned in 3 parts.

---

## Part 1: Configuration

### 1.1 App Config (`authgear.yaml`)

**New file:** `pkg/lib/config/fraud_protection.go`

Define Go structs + JSON schema (using `Schema.Add` pattern like `bot_protection.go`):

```go
type FraudProtectionConfig struct {
    Enabled  bool                       `json:"enabled,omitempty"`
    Warnings []*FraudProtectionWarning  `json:"warnings,omitempty"`
    Decision *FraudProtectionDecision   `json:"decision,omitempty"`
}

type FraudProtectionWarning struct {
    Type FraudProtectionWarningType `json:"type"`
}

type FraudProtectionWarningType string
const (
    WarningTypeSMSPhoneCountriesByIPDaily            FraudProtectionWarningType = "SMS__PHONE_COUNTRIES__BY_IP__DAILY_THRESHOLD_EXCEEDED"
    WarningTypeSMSUnverifiedOTPsByPhoneCountryDaily  FraudProtectionWarningType = "SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__DAILY_THRESHOLD_EXCEEDED"
    WarningTypeSMSUnverifiedOTPsByPhoneCountryHourly FraudProtectionWarningType = "SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__HOURLY_THRESHOLD_EXCEEDED"
    WarningTypeSMSUnverifiedOTPsByIPDaily            FraudProtectionWarningType = "SMS__UNVERIFIED_OTPS__BY_IP__DAILY_THRESHOLD_EXCEEDED"
    WarningTypeSMSUnverifiedOTPsByIPHourly           FraudProtectionWarningType = "SMS__UNVERIFIED_OTPS__BY_IP__HOURLY_THRESHOLD_EXCEEDED"
)

type FraudProtectionDecision struct {
    AlwaysAllow FraudProtectionAlwaysAllow    `json:"always_allow,omitempty"`
    Action      FraudProtectionDecisionAction `json:"action,omitempty"`
}

type FraudProtectionAlwaysAllow struct {
    IPAddress   *FraudProtectionIPAlwaysAllow          `json:"ip_address,omitempty"`
    PhoneNumber *FraudProtectionPhoneNumberAlwaysAllow `json:"phone_number,omitempty"`
}

type FraudProtectionIPAlwaysAllow struct {
    CIDRs            []string `json:"cidrs,omitempty"`
    GeoLocationCodes []string `json:"geo_location_codes,omitempty"`
}

type FraudProtectionPhoneNumberAlwaysAllow struct {
    GeoLocationCodes []string `json:"geo_location_codes,omitempty"`
    Regex            []string `json:"regex,omitempty"`
}

type FraudProtectionDecisionAction string
const (
    FraudProtectionDecisionActionRecordOnly       FraudProtectionDecisionAction = "record_only"
    FraudProtectionDecisionActionDenyIfAnyWarning FraudProtectionDecisionAction = "deny_if_any_warning"
)
```

**`SetDefaults()`** on `FraudProtectionConfig`:
- `Enabled = true`
- All 5 warning types populated
- `Action = record_only`
- `AlwaysAllow = {}`

**Modify `pkg/lib/config/config.go`:**
- Add `"fraud_protection": { "$ref": "#/$defs/FraudProtectionConfig" }` to AppConfig JSON schema
- Add `FraudProtection *FraudProtectionConfig \`json:"fraud_protection,omitempty"\`` to `AppConfig` struct

### 1.2 Feature Config (`authgear.features.yaml`)

**New file:** `pkg/lib/config/feature_fraud_protection.go`

```go
type FraudProtectionFeatureConfig struct {
    IsModifiable *bool `json:"is_modifiable,omitempty"`
}

func (c *FraudProtectionFeatureConfig) SetDefaults() {
    if c.IsModifiable == nil {
        c.IsModifiable = newBool(false)  // not modifiable by default
    }
}

// Implement MergeableFeatureConfig interface (same pattern as TestModeFeatureConfig)
func (c *FraudProtectionFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
    if layer.FraudProtection == nil {
        return c
    }
    return layer.FraudProtection
}
```

**Modify `pkg/lib/config/feature.go`:**
- Add `"fraud_protection": { "$ref": "#/$defs/FraudProtectionFeatureConfig" }` to FeatureConfig JSON schema
- Add `FraudProtection *FraudProtectionFeatureConfig \`json:"fraud_protection,omitempty"\`` to `FeatureConfig` struct

### 1.3 Effective Config Helper

When `IsModifiable = false`, the app's `fraud_protection` YAML is ignored; the hardcoded default is used instead. This is evaluated at service level:

```go
// In pkg/lib/fraudprotection/service.go
func effectiveFraudProtectionConfig(
    appCfg *config.FraudProtectionConfig,
    featureCfg *config.FraudProtectionFeatureConfig,
) *config.FraudProtectionConfig {
    if !*featureCfg.IsModifiable {
        return config.DefaultFraudProtectionConfig() // returns hardcoded default
    }
    return appCfg
}
```

### 1.4 Config Source Guard

**Modify `pkg/lib/config/configsource/resources.go`** — add a check in `validateBasedOnFeatureConfig()`:

If `!*fc.FraudProtection.IsModifiable` and the incoming `FraudProtection` config differs from the original, emit a validation error. This is the centralized enforcement point for all feature flag guards; all config updates flow through `AuthgearYAMLDescriptor.UpdateResource()` → `validateBasedOnFeatureConfig()`. This follows the same pattern as the existing biometric, password policy, and OAuth guards.

Unit tests in `pkg/lib/config/configsource/resources_test.go`:
- Modifying `fraud_protection` is blocked when `is_modifiable=false`
- Modifying `fraud_protection` is allowed when `is_modifiable=true`
- Saving with unchanged `fraud_protection` is allowed regardless of `is_modifiable`

---

## Part 2: Warning Implementation (Metrics & Evaluation)

### 2.1 Metrics Storage: `_audit_metrics` Table in `auditdb`

**New migration** `cmd/authgear/cmd/cmdaudit/migrations/audit/{timestamp}-add_audit_metrics_table.sql`:

```sql
-- +migrate Up
CREATE TABLE _audit_metrics (
    -- Normally this should be PRIMARY KEY, but a partitioned table cannot have
    -- a unique index on a column that is not part of the partition key.
    id          TEXT                        NOT NULL,
    app_id      TEXT                        NOT NULL,
    name TEXT                        NOT NULL,
    period      TEXT                        NOT NULL,
    key         TEXT                        NOT NULL,
    start_time  TIMESTAMP WITHOUT TIME ZONE NOT NULL
) PARTITION BY RANGE (start_time);

CREATE INDEX _audit_metrics_idx ON _audit_metrics (app_id, name, period, key, start_time);

CREATE TABLE _audit_metrics_template (LIKE _audit_metrics);
ALTER TABLE _audit_metrics_template ADD PRIMARY KEY (id);

-- +migrate StatementBegin
DO LANGUAGE 'plpgsql' $$DECLARE
  pg_partman_version text;
BEGIN
  SELECT extversion INTO pg_partman_version FROM pg_extension WHERE extname = 'pg_partman';
  IF pg_partman_version like '5.%' THEN
    PERFORM create_parent(
      p_parent_table := '{{ .SCHEMA }}._audit_metrics',
      p_control := 'start_time',
      p_interval := '1 month',
      p_template_table := '{{ .SCHEMA }}._audit_metrics_template'
    );
    UPDATE part_config SET retention = '90 days', retention_keep_table = FALSE
    WHERE parent_table = '{{ .SCHEMA }}._audit_metrics';
  ELSIF pg_partman_version like '4.%' THEN
    PERFORM create_parent(
      '{{ .SCHEMA }}._audit_metrics', 'start_time', 'native', 'monthly',
      p_template_table := '{{ .SCHEMA }}._audit_metrics_template'
    );
    UPDATE part_config SET retention = '90 days', retention_keep_table = FALSE
    WHERE parent_table = '{{ .SCHEMA }}._audit_metrics';
  ELSE
    RAISE EXCEPTION 'unsupported pg_partman version %', pg_partman_version;
  END IF;
END$$;
-- +migrate StatementEnd

-- +migrate Down
SELECT undo_partition('{{ .SCHEMA }}._audit_metrics', p_keep_table := FALSE,
  p_target_table := '{{ .SCHEMA }}._audit_metrics_default');
DROP TABLE _audit_metrics;
```

**Columns:**
- `id` (`text`): surrogate PK — UUID string generated in Go via `uuid.New().String()` and passed as a parameter
- `app_id` (`text`): the app tenant identifier
- `name` (`text`): e.g. `sms_otp_verified`
- `period` (`text`): time bucket granularity — `"1h"` (hourly index scan window; daily sums aggregated at query time)
- `key` (`text`): dimension and value — `{dimension}:{value}`, e.g. `ip:1.2.3.4` or `phone_country:SG`
- `start_time` (`timestamp without time zone`): exact event time (UTC); no truncation

**Primary key:** `id` only — declared on the template table so pg_partman propagates it to each child partition. The parent table has no PK constraint (partitioned tables cannot have a unique index on a non-partition-key column). No `count` column; each row represents one verified OTP event (append-only).

**Index:** `(app_id, name, period, key, start_time)` — on the parent table `_audit_metrics`; PostgreSQL propagates it to child partitions automatically.

**`key` column format:** `{dimension}:{value}`
- Examples: `ip:1.2.3.4`, `phone_country:SG`

**Records written to `_audit_metrics` — for threshold computation only:**

Two rows inserted per event in a single `INSERT ... VALUES (...), (...)` statement:

| Event | `app_id` | `name` | `period` | `key` | `start_time` |
|-------|----------|--------------|---------|-------|-------------|
| OTP verified / alt-auth | `{appID}` | `sms_otp_verified` | `1h` | `ip:{ip}` | event time (UTC) |
| OTP verified / alt-auth | `{appID}` | `sms_otp_verified` | `1h` | `phone_country:{alpha2}` | event time (UTC) |

`sms_otp_sent` is **not** written to PostgreSQL — the leaky bucket (Redis) tracks the real-time unverified count. Only `sms_otp_verified` goes to PostgreSQL because it is the historical denominator needed for adaptive threshold computation.

`phone_country` (alpha2) is derived by calling `phone.ParsePhoneNumberWithUserInput(e164)` and taking `result.Alpha2[0]` (first entry).

### 2.2 Service Interface & Method API

Two separate structs handle the two storage tiers. `Service` coordinates them.

| Struct | Storage | Purpose |
|--------|---------|---------|
| `MetricsStore` | PostgreSQL `_audit_metrics` + 5-min Redis cache | Threshold computation — historical verified counts |
| `LeakyBucketStore` | Redis HASH/SET | Real-time unverified count — always fresh |

#### `pkg/lib/fraudprotection/metrics_store.go`

PostgreSQL-only (append-only). Handles writing verified OTP events and querying threshold baselines. All read methods cache their result in Redis for 5 minutes.

```go
type MetricsStore struct {
    AuditWriteDatabase *auditdb.WriteHandle
    AuditReadDatabase  *auditdb.ReadHandle
    SQLBuilder         *auditdb.SQLBuilderApp
    WriteSQLExecutor   *auditdb.WriteSQLExecutor
    ReadSQLExecutor    *auditdb.ReadSQLExecutor
    Redis              *appredis.Handle  // for 5-min threshold cache only
    AppID              config.AppID
    Clock              clock.Clock
}

// RecordVerified inserts 2 rows into _audit_metrics in a single statement —
// one for the IP dimension and one for the phone country dimension.
// Two UUIDs are generated in Go via uuid.New().String() and passed as parameters.
// Called after an OTP is verified or alt-auth completes (fire-and-forget).
//
// SQL:
//   INSERT INTO _audit_metrics (id, app_id, name, period, key, start_time)
//   VALUES
//     ($1, $2, 'sms_otp_verified', '1h', $3, $6),
//     ($4, $2, 'sms_otp_verified', '1h', $5, $6)
//   -- $1 = uuid.New().String()        (id for ip row)
//   -- $2 = app_id
//   -- $3 = "ip:{ip}"                  e.g. "ip:1.2.3.4"
//   -- $4 = uuid.New().String()        (id for phone_country row)
//   -- $5 = "phone_country:{alpha2}"   e.g. "phone_country:SG"
//   -- $6 = clock.NowUTC()
func (s *MetricsStore) RecordVerified(ctx context.Context, ip, phoneCountry string) error

// GetVerifiedByCountry24h returns the number of verified OTPs for a phone country
// in the past 24 hours. Result cached in Redis for 5 minutes.
//
// SQL:
//   SELECT COUNT(*)
//   FROM _audit_metrics
//   WHERE app_id = $1
//     AND name = 'sms_otp_verified'
//     AND period = '1h'
//     AND key = $2
//     AND start_time >= $3
//   -- $1 = app_id
//   -- $2 = "phone_country:{country}"
//   -- $3 = clock.NowUTC().Add(-24 * time.Hour)
func (s *MetricsStore) GetVerifiedByCountry24h(ctx context.Context, country string) (int64, error)

// GetVerifiedByCountry1h returns the number of verified OTPs for a phone country
// in the past 1 hour. Result cached in Redis for 5 minutes.
//
// SQL:
//   SELECT COUNT(*)
//   FROM _audit_metrics
//   WHERE app_id = $1
//     AND name = 'sms_otp_verified'
//     AND period = '1h'
//     AND key = $2
//     AND start_time >= $3
//   -- $2 = "phone_country:{country}"
//   -- $3 = clock.NowUTC().Add(-1 * time.Hour)
func (s *MetricsStore) GetVerifiedByCountry1h(ctx context.Context, country string) (int64, error)

// GetVerifiedByIP24h returns the number of verified OTPs from a specific IP
// in the past 24 hours. Result cached in Redis for 5 minutes.
//
// SQL:
//   SELECT COUNT(*)
//   FROM _audit_metrics
//   WHERE app_id = $1
//     AND name = 'sms_otp_verified'
//     AND period = '1h'
//     AND key = $2
//     AND start_time >= $3
//   -- $2 = "ip:{ip}"
//   -- $3 = clock.NowUTC().Add(-24 * time.Hour)
func (s *MetricsStore) GetVerifiedByIP24h(ctx context.Context, ip string) (int64, error)

// GetVerifiedByCountryPast14DaysRollingMax returns the maximum single-day verified count
// for a phone country across the past 14 days. Used as the baseline for adaptive thresholds.
// Result cached in Redis for 5 minutes.
//
// SQL:
//   SELECT COALESCE(MAX(daily_count), 0)
//   FROM (
//     SELECT DATE_TRUNC('day', start_time) AS day, COUNT(*) AS daily_count
//     FROM _audit_metrics
//     WHERE app_id = $1
//       AND name = 'sms_otp_verified'
//       AND period = '1h'
//       AND key = $2
//       AND start_time >= $3
//     GROUP BY day
//   ) t
//   -- $2 = "phone_country:{country}"
//   -- $3 = clock.NowUTC().Add(-14 * 24 * time.Hour)
func (s *MetricsStore) GetVerifiedByCountryPast14DaysRollingMax(ctx context.Context, country string) (int64, error)
```

**Redis cache key format:** `{appID}:fraud_protection:threshold_cache:{name}:{period}:{key}` — TTL **5 minutes**.

**Go-side key construction examples:**
```go
appID   := string(s.AppID)
pgKey   := fmt.Sprintf("phone_country:%s", country)  // or fmt.Sprintf("ip:%s", ip)

cacheKey := fmt.Sprintf("%s:fraud_protection:threshold_cache:sms_otp_verified:1h:%s", appID, pgKey)
```

#### `pkg/lib/fraudprotection/leaky_bucket_store.go`

Redis-only. Handles fill/drain operations and returns triggered state atomically.

```go
type LeakyBucketStore struct {
    Redis *appredis.Handle
    AppID config.AppID
    Clock clock.Clock
}

// RecordSMSOTPSent atomically fills all 4 leaky buckets and updates the ip_countries ZSET.
// Each Lua call fills AND returns {new_level, triggered} in a single Redis round-trip.
// Returns LeakyBucketTriggered indicating which warning conditions are now exceeded.
// Called INSIDE Service.CheckAndRecord BEFORE the SMS send (not after).
// Blocked requests are intentionally counted — a blocked attempt is still an attack signal.
func (s *LeakyBucketStore) RecordSMSOTPSent(ctx context.Context, ip, phoneCountry string, thresholds LeakyBucketThresholds) (LeakyBucketTriggered, error)

// RecordSMSOTPVerified drains all 4 leaky buckets by count units (fire-and-forget).
// Passes n = -count to the Lua script, which handles any integer n natively.
func (s *LeakyBucketStore) RecordSMSOTPVerified(ctx context.Context, ip, phoneCountry string, thresholds LeakyBucketThresholds, count int) error

// LeakyBucketThresholds holds per-bucket adaptive threshold values.
// The Go wrapper for each bucket supplies the matching period constant (3600 or 86400).
type LeakyBucketThresholds struct {
    CountryHourly float64 // used with period=3600
    CountryDaily  float64 // used with period=86400
    IPHourly      float64 // used with period=3600
    IPDaily       float64 // used with period=86400
}

// LeakyBucketTriggered holds the per-bucket warning trigger state returned by RecordSMSOTPSent.
// A bucket is triggered when new_level > threshold after the fill.
type LeakyBucketTriggered struct {
    CountryHourly    bool // SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__HOURLY_THRESHOLD_EXCEEDED
    CountryDaily     bool // SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__DAILY_THRESHOLD_EXCEEDED
    IPHourly         bool // SMS__UNVERIFIED_OTPS__BY_IP__HOURLY_THRESHOLD_EXCEEDED
    IPDaily          bool // SMS__UNVERIFIED_OTPS__BY_IP__DAILY_THRESHOLD_EXCEEDED
    IPCountriesDaily bool // SMS__PHONE_COUNTRIES__BY_IP__DAILY_THRESHOLD_EXCEEDED (from SET Lua)
}
```

**Leaky bucket state** stored as a Redis HASH:

```
key:    {appID}:fraud_protection:leaky_bucket:{period}:{dimension}:{value}
fields: level (float), last_updated (unix seconds)
```

The adaptive threshold is passed directly from `Service` into the Lua script along with the period. The Lua script derives `drain_rate = threshold / period` internally. Because the threshold can change between calls, the level is capped at the current threshold before any operation to prevent stale high values from persisting after the threshold decreases.

| Bucket | `period` arg | `threshold` arg |
|--------|-------------|----------------|
| `country` | `3600` (1h) | `country_hourly_threshold` |
| `country` | `86400` (1d) | `country_daily_threshold` |
| `ip` | `3600` (1h) | `ip_hourly_threshold` |
| `ip` | `86400` (1d) | `ip_daily_threshold` |

**Single unified Lua script** — `n>0` fill by n, `n<0` drain by |n|, `n=0` read:

```lua
-- KEYS[1] = bucket key
-- ARGV[1] = now (unix seconds, float)
-- ARGV[2] = threshold (current adaptive threshold for this bucket)
-- ARGV[3] = period (window_seconds: 3600 for 1h buckets, 86400 for 1d buckets)
-- ARGV[4] = n (positive=fill by n, negative=drain by |n|, 0=read)
-- ARGV[5] = ttl_seconds (2 * window_seconds; only applied when n=1)

local now       = tonumber(ARGV[1])
local threshold = tonumber(ARGV[2])
local period    = tonumber(ARGV[3])
local n         = tonumber(ARGV[4])
local ttl       = tonumber(ARGV[5])

local drain_rate = threshold / period

local data = redis.call('HMGET', KEYS[1], 'level', 'last_updated')
local level        = tonumber(data[1]) or 0
local last_updated = tonumber(data[2]) or now

-- Cap level at the current threshold before any operation.
-- This handles threshold decreases: if the threshold drops from 100 to 50
-- and the stored level is 80, we treat it as 50 so warnings don't persist
-- beyond the new threshold indefinitely.
level = math.min(level, threshold)

local elapsed   = now - last_updated
local leaked    = elapsed * drain_rate
local new_level = math.max(0, level - leaked + n)  -- time drain + event delta

if n ~= 0 then
    redis.call('HMSET', KEYS[1], 'level', new_level, 'last_updated', now)
    if n == 1 then
        redis.call('EXPIRE', KEYS[1], ttl)  -- extend TTL only on fill
    end
end

-- Return both the new level and whether it exceeds the threshold.
-- The caller uses this to determine if the warning rule is triggered.
return {new_level, (new_level > threshold) and 1 or 0}
```

**Return value:** A two-element Redis array `[new_level (bulk string), triggered (integer 0 or 1)]`. `triggered = 1` when `new_level > threshold`.

**TTL** (`2 * window_seconds`) is refreshed on every fill. If no sends occur for `2 × window`, the key expires and the level is treated as 0 on next access — providing time-based recovery.

**For distinct countries per IP** — uses a Redis ZSET (sorted set) with its own Lua script.

Each country code is a member; its score is the most recent timestamp it was seen from this IP. `ZADD GT` updates the score only if the new timestamp is greater (i.e. always refreshes to the latest). `ZREMRANGEBYSCORE` prunes countries not seen in the past 24h before counting, implementing a true 24h sliding window.

```
key: {appID}:fraud_protection:ip_countries:{ip}   (Redis ZSET)
```

```lua
-- KEYS[1] = sorted set key
-- ARGV[1] = alpha2 country code
-- ARGV[2] = now (unix timestamp)
-- ARGV[3] = threshold (fixed = 3)
-- ARGV[4] = ttl_seconds (2 * 86400 — cleanup if IP goes idle)
-- Returns: {count, triggered_int}

local now    = tonumber(ARGV[2])
local cutoff = now - 86400  -- 24h ago

-- Add/update country with latest timestamp as score
redis.call('ZADD', KEYS[1], 'GT', now, ARGV[1])
-- Prune countries not seen in the past 24h
redis.call('ZREMRANGEBYSCORE', KEYS[1], '-inf', cutoff)
-- Refresh TTL so key expires if IP goes idle for 2 days
redis.call('EXPIRE', KEYS[1], ARGV[4])
-- Count distinct active countries
local count = redis.call('ZCARD', KEYS[1])
return {count, (count > tonumber(ARGV[3])) and 1 or 0}
```

- A country stays in the set only while it has been seen within the last 24h
- Key expires `2 × 24h` after the last send from this IP

**Per-event operations summary:**

| Event | Operations |
|-------|-----------|
| SMS OTP sent | `fill(leaky_bucket:1h:country:{alpha2})`, `fill(leaky_bucket:1d:country:{alpha2})`, `fill(leaky_bucket:1h:ip:{ip})`, `fill(leaky_bucket:1d:ip:{ip})`, `ZADD ip_countries:{ip} {alpha2}` |
| OTP verified / alt-auth | PostgreSQL write + `drain(leaky_bucket:1h:country:{alpha2})`, `drain(leaky_bucket:1d:country:{alpha2})`, `drain(leaky_bucket:1h:ip:{ip})`, `drain(leaky_bucket:1d:ip:{ip})` |

#### `pkg/lib/fraudprotection/service.go`

`Service` is the coordinator. It queries `MetricsStore` for thresholds, passes them to `LeakyBucketStore` for atomic fill/check, then dispatches the audit event.

```go
// Interface consumed by messaging.Sender (defined in messaging package)
type FraudProtectionService interface {
    CheckAndRecord(ctx context.Context, phoneNumber, messageType string) error
}

type EventService interface {
    DispatchEventImmediately(ctx context.Context, payload event.NonBlockingPayload) error
}

type Service struct {
    Metrics       *MetricsStore
    LeakyBucket   *LeakyBucketStore
    Config        *config.FraudProtectionConfig
    FeatureConfig *config.FraudProtectionFeatureConfig
    RemoteIP      httputil.RemoteIP
    Clock         clock.Clock
    EventService  EventService
}

// CheckAndRecord is the main entry point called BEFORE sending an SMS.
// It: (1) computes thresholds via MetricsStore, (2) atomically fills leaky buckets + checks
//     warnings via LeakyBucketStore, (3) dispatches audit event, (4) returns error if blocked.
func (s *Service) CheckAndRecord(ctx context.Context, phoneNumber, messageType string) error

// RecordSMSOTPVerified is called from otp.Service.VerifyOTP() when code.OOBChannel==SMS (fire-and-forget).
// It writes to PostgreSQL metrics AND drains the leaky bucket by calling RevertSMSOTPSent(ctx, phoneNumber, 1) internally.
func (s *Service) RecordSMSOTPVerified(ctx context.Context, phoneNumber string)

// RevertSMSOTPSent drains all 4 leaky buckets by count units — no PostgreSQL write.
// count is the number of unverified sends to cancel out. The drain is executed in a single
// Lua call per bucket (n = -count), so callers must NOT loop; pass the total count directly.
// Must NOT be called when OTP is actually verified (use RecordSMSOTPVerified instead).
// RecordSMSOTPVerified calls this internally with count=1 for the single verified drain.
func (s *Service) RevertSMSOTPSent(ctx context.Context, phoneNumber string, count int)

// ComputeThresholds queries MetricsStore for all 4 adaptive thresholds.
func (s *Service) ComputeThresholds(ctx context.Context, ip, phoneCountry string) LeakyBucketThresholds

// evaluateWarnings maps LeakyBucketTriggered to []FraudProtectionWarningType, filtered by config.Warnings
func (s *Service) evaluateWarnings(triggered LeakyBucketTriggered) []FraudProtectionWarningType

// isAlwaysAllowed checks IP CIDRs, IP geo codes, phone geo codes, phone regex
func (s *Service) isAlwaysAllowed(ip, phoneNumber, phoneCountry string) bool
```

**Threshold formulas** — threshold from PostgreSQL/cache, current count from Redis leaky bucket:

Thresholds are computed from PostgreSQL (5-min cache), collected into `LeakyBucketThresholds`, then passed to `RecordSMSOTPSent`. Each Lua call atomically fills and returns whether triggered. The results are mapped to warning types via `LeakyBucketTriggered`.

```
SMS__PHONE_COUNTRIES__BY_IP__DAILY_THRESHOLD_EXCEEDED:
  threshold = 3  (fixed)
  → triggered: LeakyBucketTriggered.IPCountriesDaily  (from distinct-countries Lua script)

SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__DAILY_THRESHOLD_EXCEEDED:
  threshold = max(20, GetVerifiedByCountryPast14DaysRollingMax(country)*0.2,
                      GetVerifiedByCountry24h(country)*0.2)   // PG, 5-min cache
  period = 86400
  → triggered: LeakyBucketTriggered.CountryDaily

SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__HOURLY_THRESHOLD_EXCEEDED:
  daily_threshold = (as above)
  threshold = max(3, daily_threshold/6, GetVerifiedByCountry1h(country)*0.2)  // PG, cached
  period = 3600
  → triggered: LeakyBucketTriggered.CountryHourly

SMS__UNVERIFIED_OTPS__BY_IP__DAILY_THRESHOLD_EXCEEDED:
  threshold = max(10, GetVerifiedByIP24h(ip)*0.2)    // PG, 5-min cache
  period = 86400
  → triggered: LeakyBucketTriggered.IPDaily

SMS__UNVERIFIED_OTPS__BY_IP__HOURLY_THRESHOLD_EXCEEDED:
  threshold = max(5, GetVerifiedByIP24h(ip)*0.2/6)   // PG, 5-min cache
  period = 3600
  → triggered: LeakyBucketTriggered.IPHourly
```

### 2.3 DI / Wire Setup

**New file: `pkg/lib/fraudprotection/deps.go`**

```go
var DependencySet = wire.NewSet(
    wire.Struct(new(Service), "*"),
    wire.Struct(new(MetricsStore), "*"),
    wire.Struct(new(LeakyBucketStore), "*"),
)
```

**Modify `pkg/lib/deps/deps_common.go`**:
- Add `fraudprotection.DependencySet` to `CommonDependencySet`
- Bind `*fraudprotection.Service` to the `FraudProtectionService` interface consumed by `messaging.Sender` and `otp.Service`

`auditdb.WriteHandle`, `auditdb.ReadHandle`, `auditdb.SQLBuilderApp`, `auditdb.WriteSQLExecutor`, `auditdb.ReadSQLExecutor` are already provided via `AppRootDeps` in `deps_provider.go`. `appredis.Handle` and `clock.Clock` are also already in scope.

**Modify `pkg/lib/messaging/sender.go`**: Add `FraudProtection FraudProtectionService` field.

**Modify `pkg/lib/authn/otp/service.go`**: Add `FraudProtection FraudProtectionService` field. This is where `RecordSMSOTPVerified` is called — centralised here so all current and future SMS OTP verification paths are covered automatically.

**Modify `pkg/lib/authenticationflow/deps.go`** (`authflow.Dependencies`): Add `FraudProtection *fraudprotection.Service` so the `OnCommitEffect` in the 6 public flow intents can call `RevertSMSOTPSent`.

### 2.4 Call Flow (Where Each Method Is Called)

```
messaging.Sender.sendSMS()
  └── [BEFORE send] s.FraudProtection.CheckAndRecord(ctx, opts.To, string(msgType))
        ├── effectiveCfg := effectiveFraudProtectionConfig(cfg, featureCfg)
        ├── if !effectiveCfg.Enabled → return nil
        ├── phoneCountry := phone.ParsePhoneNumberWithUserInput(phoneNumber).Alpha2[0]
        ├── if s.isAlwaysAllowed(ip, phoneNumber, phoneCountry) → return nil
        ├── thresholds := s.ComputeThresholds(ctx, ip, phoneCountry)
        │     └── MetricsStore.GetVerified* (PG + 5-min cache)
        ├── triggered, err := s.LeakyBucket.RecordSMSOTPSent(ctx, ip, phoneCountry, thresholds)
        │     └── Atomically: fill all 4 leaky buckets (n=1) + ZADD ip_countries (ZSET sliding window)
        │         Each Lua call returns {new_level, triggered} — no separate read step
        ├── warnings := s.evaluateWarnings(triggered)  // maps LeakyBucketTriggered → []WarningType
        ├── dispatch FraudProtectionDecisionRecorded event (always, see §3)
        └── if action==deny_if_any_warning && len(warnings)>0 → return ErrBlockedByFraudProtection

// Note: RecordSMSOTPSent is no longer called after the send.
// Blocked requests are intentionally counted — a blocked attempt is still an attack signal.

otp.Service.VerifyOTP() [after isCodeValid = true, code.OOBChannel==SMS]
  └── s.FraudProtection.RecordSMSOTPVerified(ctx, target)
        ├── MetricsStore.RecordVerified(ctx, ip, phoneCountry)   // PostgreSQL write
        └── RevertSMSOTPSent(ctx, target, 1)                    // leaky bucket drain (internal, count=1)

// RecordSMSOTPVerified fires automatically for ALL SMS OTP verification paths because every
// path converges at otp.Service.VerifyOTP() — node_authn_oob.go (via Authenticators facade),
// node_verify_claim.go (direct call), forgotpassword/account_recovery (via ResetPassword).
// No per-site hooks needed. Future SMS OTP paths are covered automatically.
//
// Email and WhatsApp OTPs are intentionally out of scope — all 5 warning types have the
// SMS__ prefix, and code.OOBChannel filters them out at the source.

All 6 public flow intents GetEffects() [OnCommitEffect — alt-auth exclusion, see §2.5]
(intent_login_flow.go, intent_signup_flow.go, intent_signup_login_flow.go,
 intent_reauth_flow.go, intent_promote_flow.go, intent_account_recovery_flow.go)
  └── skip if flows.Nearest != flows.Root  (sub-flows do nothing; root flow executes)
      if unverifiedCount = sentCount - verifiedCount > 0:
        deps.FraudProtection.RevertSMSOTPSent(ctx, phoneNumber, unverifiedCount)  // drain only, no PG write
```

> **Note on `thresholds`:** `Service.ComputeThresholds(ctx, ip, phoneCountry)` queries `MetricsStore.GetVerified*` (PG + 5-min cache) to build `LeakyBucketThresholds`. Thresholds are used by the Lua script to compute `drain_rate = threshold / period` and apply `level = min(level, threshold)`. Both `RecordSMSOTPVerified` and `RevertSMSOTPSent` call `ComputeThresholds` internally before delegating the drain to `LeakyBucketStore`. `RevertSMSOTPSent` passes its `count` argument through to `LeakyBucketStore.RecordSMSOTPVerified`, which passes `n = -count` to the Lua script. Since the Lua script handles any integer `n` natively (`math.max(0, level - leaked + n)`), no looping is needed.

### 2.5 Alt-Auth Exclusion Detail

**Problem:** When a user is shown an SMS OTP step, the SMS is sent and `RecordSMSOTPSent` is called. If the user then authenticates via an alternative method (password, passkey, recovery code, device token) without verifying the OTP, all sends would incorrectly inflate the "unverified" leaky bucket count. We cancel them out by calling `RevertSMSOTPSent` once with the total unverified count (no PostgreSQL write — the OTP was never verified).

**Why counts cannot be stored on the node:** Authflow supports back-navigation via `state_token`. When the user navigates back, the entire flow state (including node fields) reverts to the previous snapshot. This means any counter stored on `NodeAuthenticationOOB` would be silently reset, causing an undercount. The SMS was physically sent regardless of flow state — the counter must survive state reverts.

**Solution: Store counts in the authflow `Session`.** The `Session` is stored separately in Redis under its own key (keyed by `FlowID`) and is never rolled back by state token changes. It is the correct durability boundary for accumulating cross-step counters.

**Why `ErrorSMSOTPSent` signal from `doAccept` does NOT work for the sent count:**
`sendSMS()` is called inside `DelayedOneTimeFunction` (both initial send via `NewNodeAuthenticationOOB` and resend via `ErrReplaceNode`). `DelayedOneTimeFunctions` execute in `processAcceptResult` AFTER `doAccept` returns. Any signal set in `doAccept` would fire BEFORE the SMS is actually sent, making the count update premature (and wrong if the send fails).

**Unified mechanism: `UpdatedSession` on both `DelayedOneTimeFunctionResult` and `NodeWithDelayedOneTimeFunction`**

**Session stored in context:**

`MakeContext` stores the live `*Session` pointer in context so all callers can construct a complete updated session:
```go
// In session.go MakeContext:
ctx = context.WithValue(ctx, contextKeySession, s)

// In context.go:
func GetSession(ctx context.Context) *Session { return ctx.Value(contextKeySession).(*Session) }
```

Since the pointer is stored by reference, any in-place mutations (e.g., `SetBotProtectionVerificationResult`) made after `MakeContext` are visible to later callers of `GetSession(ctx)`.

**`Session.PatchFrom` copies all fields** (callers must provide a complete session):
```go
func (s *Session) PatchFrom(updated *Session) {
    *s = *updated
}
```

Callers always construct `UpdatedSession` by copying the live session and modifying the target field:
```go
updated := *authflow.GetSession(ctx)
updated.SMSOTPSentCount++
return &updated
```

**Part 1 — SMS sent (`SMSOTPSentCount`):** via `DelayedOneTimeFunctionResult`

The `DelayedOneTimeFunction` type is changed to return `(DelayedOneTimeFunctionResult, error)`:
```go
// In pkg/lib/authenticationflow/node.go:
type DelayedOneTimeFunctionResult struct {
    // If non-nil, processAcceptResult patches the session and calls UpdateSession.
    UpdatedSession *Session
}
type DelayedOneTimeFunction func(ctx context.Context, deps *Dependencies) (DelayedOneTimeFunctionResult, error)
```

The `sendSMS()` delayed function constructs a complete updated session after the send succeeds:
```go
err := n.SendCode(ctx, deps, code)
if err != nil {
    return DelayedOneTimeFunctionResult{}, err
}
updated := *authflow.GetSession(ctx)
updated.SMSOTPSentCount++
return DelayedOneTimeFunctionResult{UpdatedSession: &updated}, nil
```

In `processAcceptResult`, after each delayed function:
```go
result, err := fn(ctx, s.Deps)
if result.UpdatedSession != nil {
    session.PatchFrom(result.UpdatedSession)
    _ = s.Store.UpdateSession(ctx, session)
}
if err != nil { /* ... existing error handling ... */ }
```

**All 5 existing `DelayedOneTimeFunction` usages** must be updated to return `(DelayedOneTimeFunctionResult, error)`:
- `node_authn_oob.go` (initial send + resend) — send closures return `UpdatedSession` with incremented count
- `node_verify_claim.go`, `node_pre_authenticate.go`, `node_pre_initialize.go`, `node_post_identified.go` — return `DelayedOneTimeFunctionResult{}` (no session update needed)

**Part 2 — OTP verified (`SMSOTPVerifiedCount`):** via `NodeWithDelayedOneTimeFunction.UpdatedSession`

`VerifyOneWithSpec` is synchronous inside `ReactTo()`. Instead of an error signal, `NodeWithDelayedOneTimeFunction` is extended with an `UpdatedSession` field:
```go
// In pkg/lib/authenticationflow/node.go:
type NodeWithDelayedOneTimeFunction struct {
    Node                   *Node
    DelayedOneTimeFunction DelayedOneTimeFunction  // may be nil
    UpdatedSession         *Session                // optional; processed before delayed functions
}
```

When OTP verification succeeds, `ReactTo()` returns:
```go
updated := *authflow.GetSession(ctx)
updated.SMSOTPVerifiedCount++
return &authflow.NodeWithDelayedOneTimeFunction{
    Node:           nextNode,
    UpdatedSession: &updated,
}, nil
```

`doAccept` adds `UpdatedSession` to a new `AcceptResult.PendingSessionPatches` slice when handling `NodeWithDelayedOneTimeFunction` (in BOTH the `ErrReplaceNode` path and the normal nil-error path):
```go
// Add to AcceptResult struct:
PendingSessionPatches []*Session

// In doAccept, both NodeWithDelayedOneTimeFunction paths:
if nwdf.UpdatedSession != nil {
    result.PendingSessionPatches = append(result.PendingSessionPatches, nwdf.UpdatedSession)
}
```

`processAcceptResult` applies all pending patches BEFORE running delayed functions:
```go
for _, patch := range acceptResult.PendingSessionPatches {
    session.PatchFrom(patch)
}
if len(acceptResult.PendingSessionPatches) > 0 {
    _ = s.Store.UpdateSession(ctx, session)
}
// Then run DelayedOneTimeFunctions...
```

**`Session` struct** additions in `pkg/lib/authenticationflow/session.go`:
```go
SMSOTPSentCount     int `json:"sms_otp_sent_count,omitempty"`
SMSOTPVerifiedCount int `json:"sms_otp_verified_count,omitempty"`
```
Added to `MakeContext()` and retrieved via context getters in `context.go`:
```go
func GetSMSOTPSentCount(ctx context.Context) int { ... }
func GetSMSOTPVerifiedCount(ctx context.Context) int { ... }
func GetSession(ctx context.Context) *Session { ... }
```

**Milestone used:**

**`MilestoneOOBOTPSentPhoneTarget`** — exposes the phone number. The count comes from the session via context. `NodeAuthenticationOOB` implements this by returning `n.Info.OOBOTP.ToTarget()`.

```go
// New interface in milestone.go
type MilestoneOOBOTPSentPhoneTarget interface {
    authflow.Milestone
    MilestoneOOBOTPSentPhoneTarget() string   // returns E.164 phone number
}
```

**OnCommitEffect implementation** — added to `GetEffects()` of all 6 public flow intents:

```go
authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
    // Only execute for the root flow. Sub-flows (e.g. login embedded in signup for
    // account linking) skip this — the root flow's effect sees all accumulated counts.
    if flows.Nearest != flows.Root {
        return nil
    }
    sentCount     := authflow.GetSMSOTPSentCount(ctx)
    verifiedCount := authflow.GetSMSOTPVerifiedCount(ctx)
    unverifiedCount := sentCount - verifiedCount
    if unverifiedCount <= 0 {
        return nil
    }
    phoneNode, _, ok := authflow.FindMilestoneInCurrentFlow[MilestoneOOBOTPSentPhoneTarget](flows)
    if !ok {
        return nil
    }
    phoneNumber := phoneNode.MilestoneOOBOTPSentPhoneTarget()
    deps.FraudProtection.RevertSMSOTPSent(ctx, phoneNumber, unverifiedCount)
    return nil
}),
```

**Why `flows.Nearest != flows.Root`:** `SMSOTPSentCount`/`SMSOTPVerifiedCount` are global session counters that accumulate across the entire flow tree. When login is embedded inside signup (account linking), OTPs may be sent both inside the login sub-flow and in the remaining signup steps after the login sub-flow completes. Only the root flow's `OnCommitEffect` fires after all steps are fully done, so it sees the final totals. The embedded login flow's own `OnCommitEffect` skips (`Nearest != Root`) to avoid double-counting.

The 6 public flows: `intent_login_flow.go`, `intent_signup_flow.go`, `intent_signup_login_flow.go`, `intent_reauth_flow.go`, `intent_promote_flow.go`, `intent_account_recovery_flow.go`.

### 2.6 Integration Points (File-Level)

- **`pkg/lib/authenticationflow/node.go`**: Add `DelayedOneTimeFunctionResult` struct; change `DelayedOneTimeFunction` return type to `(DelayedOneTimeFunctionResult, error)`; add `UpdatedSession *Session` field to `NodeWithDelayedOneTimeFunction`.
- **`pkg/lib/authenticationflow/session.go`**: Add `SMSOTPSentCount int` and `SMSOTPVerifiedCount int` fields; add `PatchFrom(*Session)` method (`*s = *updated`); update `MakeContext` to store `*Session` pointer in context.
- **`pkg/lib/authenticationflow/context.go`**: Add `GetSession(ctx)` returning `*Session`; add `GetSMSOTPSentCount` / `GetSMSOTPVerifiedCount` getter functions.
- **`pkg/lib/authenticationflow/accept.go`**: Add `PendingSessionPatches []*Session` to `AcceptResult`; in `doAccept`, for both `NodeWithDelayedOneTimeFunction` code paths, append `nwdf.UpdatedSession` to `result.PendingSessionPatches` when non-nil.
- **`pkg/lib/authenticationflow/service.go`**: In `processAcceptResult`, (a) apply `PendingSessionPatches` via `PatchFrom` + `UpdateSession` before running delayed functions, and (b) after each delayed function, if `result.UpdatedSession != nil`, call `session.PatchFrom` then `UpdateSession`.
- **`pkg/lib/messaging/sender.go`**: Add `FraudProtection FraudProtectionService` field; call `CheckAndRecord` before send.
- **`pkg/lib/authenticationflow/declarative/milestone.go`**: Add `MilestoneOOBOTPSentPhoneTarget` interface.
- **`pkg/lib/authn/otp/service.go`**: Add `FraudProtection FraudProtectionService` field. In `VerifyOTP()`, after `isCodeValid = true`, call `s.FraudProtection.RecordSMSOTPVerified(ctx, target)` when `code.OOBChannel == model.AuthenticatorOOBChannelSMS`. This is the single centralised hook covering all current and future SMS OTP verification paths.
- **`pkg/lib/authenticationflow/declarative/node_authn_oob.go`**: Update both `DelayedOneTimeFunction` closures (initial send + resend) to return `DelayedOneTimeFunctionResult{UpdatedSession: &updated}` after `SendCode` succeeds; when OTP verification succeeds in `ReactTo`, return `NodeWithDelayedOneTimeFunction{Node: nextNode, UpdatedSession: &updated}` with `SMSOTPVerifiedCount++`; implement `MilestoneOOBOTPSentPhoneTarget`. No `RecordSMSOTPVerified` call here — handled centrally in `otp.Service`.
- **`pkg/lib/authenticationflow/declarative/node_verify_claim.go`**: Update `DelayedOneTimeFunction` closure signature only. No `RecordSMSOTPVerified` call — handled centrally.
- **`pkg/lib/authenticationflow/declarative/node_pre_authenticate.go`**, **`node_pre_initialize.go`**, **`node_post_identified.go`**: Update `DelayedOneTimeFunction` closure signatures to return `(DelayedOneTimeFunctionResult{}, error)`.
- **`pkg/lib/authenticationflow/declarative/intent_account_recovery_flow_step_verify_account_recovery_code.go`**: When `verifyCode()` succeeds and `dest.ForgotPasswordCodeChannel() == SMS`, increment `SMSOTPVerifiedCount` in session (return `NodeWithDelayedOneTimeFunction{Node: nextNode, UpdatedSession: &updated}`). No `RecordSMSOTPVerified` call — handled centrally in `otp.Service`.
- **`pkg/lib/authenticationflow/declarative/intent_login_flow.go`**: Add `OnCommitEffect` with root-flow guard to `GetEffects()`.
- **`pkg/lib/authenticationflow/declarative/intent_signup_flow.go`**: Same.
- **`pkg/lib/authenticationflow/declarative/intent_signup_login_flow.go`**: Same.
- **`pkg/lib/authenticationflow/declarative/intent_reauth_flow.go`**: Same.
- **`pkg/lib/authenticationflow/declarative/intent_promote_flow.go`**: Same.
- **`pkg/lib/authenticationflow/declarative/intent_account_recovery_flow.go`**: Same.

---

## Part 3: Decision Record & Audit Log

### 3.1 API Error

In `pkg/lib/fraudprotection/service.go`:
```go
var ErrBlockedByFraudProtection = apierrors.Forbidden.WithReason("BlockedByFraudProtection").New("request blocked by fraud protection")
```
→ `{"name":"Forbidden","reason":"BlockedByFraudProtection","code":403}`

### 3.2 Audit Log Event

**New file:** `pkg/api/event/nonblocking/fraud_protection_decision_recorded.go`

```go
const FraudProtectionDecisionRecorded event.Type = "fraud_protection.decision_recorded"

type FraudProtectionDecisionRecord struct {
    Timestamp         string            `json:"timestamp"`
    Decision          string            `json:"decision"`             // "allowed" or "blocked"
    BlockMode         string            `json:"block_mode,omitempty"` // "error" when blocked
    Action            string            `json:"action"`               // "send_sms"
    ActionDetail      map[string]string `json:"action_detail"`
    TriggeredWarnings []string          `json:"triggered_warnings"`
    UserAgent         string            `json:"user_agent,omitempty"`
    IPAddress         string            `json:"ip_address,omitempty"`
    HTTPUrl           string            `json:"http_url,omitempty"`
    HTTPReferer       string            `json:"http_referer,omitempty"`
    UserID            string            `json:"user_id,omitempty"`
    GeoLocationCode   string            `json:"geo_location_code,omitempty"`
}

type FraudProtectionDecisionRecordedEventPayload struct {
    Record FraudProtectionDecisionRecord `json:"record"`
}

func (e *FraudProtectionDecisionRecordedEventPayload) NonBlockingEventType() event.Type {
    return FraudProtectionDecisionRecorded
}
func (e *FraudProtectionDecisionRecordedEventPayload) ForAudit() bool { return true }
func (e *FraudProtectionDecisionRecordedEventPayload) ForHook() bool  { return false }
```

### 3.3 Event Dispatch

Dispatched inside `Service.CheckAndRecord()` after warning evaluation, before the block-or-allow decision:

```go
s.EventService.DispatchEventImmediately(ctx, &nonblocking.FraudProtectionDecisionRecordedEventPayload{
    Record: FraudProtectionDecisionRecord{
        Timestamp:         s.Clock.NowUTC().Format(time.RFC3339Nano),
        Decision:          decision,   // "allowed" or "blocked"
        BlockMode:         blockMode,  // "error" if blocked
        Action:            "send_sms",
        ActionDetail:      map[string]string{"recipient": phoneNumber, "type": messageType},
        TriggeredWarnings: triggeredWarningStrings,
        IPAddress:         string(s.RemoteIP),
        GeoLocationCode:   geoCode,  // from geoip.IPString(ip)
    },
})
```

---

## Critical Files

| File | Action |
|------|--------|
| `pkg/lib/config/fraud_protection.go` | **Create** — AppConfig structs + JSON schema |
| `pkg/lib/config/feature_fraud_protection.go` | **Create** — FeatureConfig struct + Merge |
| `pkg/lib/config/config.go` | **Modify** — add `FraudProtection` field + schema entry |
| `pkg/lib/config/feature.go` | **Modify** — add `FraudProtection` field + schema entry |
| `pkg/lib/config/configsource/resources.go` | **Modify** — add `fraud_protection` guard in `validateBasedOnFeatureConfig()` |
| `pkg/lib/config/configsource/resources_test.go` | **Modify** — add test cases for the `is_modifiable` guard |
| `pkg/lib/authenticationflow/node.go` | **Modify** — add `DelayedOneTimeFunctionResult` struct; change `DelayedOneTimeFunction` return type; add `UpdatedSession` to `NodeWithDelayedOneTimeFunction` |
| `pkg/lib/authenticationflow/session.go` | **Modify** — add `SMSOTPSentCount` / `SMSOTPVerifiedCount` fields + `PatchFrom` method + store `*Session` in `MakeContext` |
| `pkg/lib/authenticationflow/context.go` | **Modify** — add `GetSession` + `GetSMSOTPSentCount` / `GetSMSOTPVerifiedCount` getters |
| `pkg/lib/authenticationflow/accept.go` | **Modify** — add `PendingSessionPatches` to `AcceptResult`; handle `NodeWithDelayedOneTimeFunction.UpdatedSession` in `doAccept` |
| `pkg/lib/authenticationflow/service.go` | **Modify** — update `processAcceptResult` to apply `PendingSessionPatches` + `DelayedOneTimeFunctionResult.UpdatedSession` |
| `pkg/lib/fraudprotection/service.go` | **Create** — `Service`, `CheckAndRecord`, `RecordSMSOTPVerified`, `ComputeThresholds`, `evaluateWarnings`, `isAlwaysAllowed`, error var |
| `pkg/lib/fraudprotection/metrics_store.go` | **Create** — `MetricsStore`: PostgreSQL writes + threshold queries (with 5-min Redis cache) |
| `pkg/lib/fraudprotection/leaky_bucket_store.go` | **Create** — `LeakyBucketStore`: Redis fill/drain + `LeakyBucketThresholds` / `LeakyBucketTriggered` types |
| `pkg/lib/fraudprotection/deps.go` | **Create** — Wire DI set |
| `pkg/lib/deps/deps_common.go` | **Modify** — add `fraudprotection.DependencySet` + interface binding |
| `pkg/lib/messaging/sender.go` | **Modify** — `FraudProtection` field + call `CheckAndRecord` / `RecordSMSOTPSent` |
| `pkg/lib/authenticationflow/declarative/milestone.go` | **Modify** — add `MilestoneOOBOTPSentPhoneTarget` interface |
| `pkg/lib/authn/otp/service.go` | **Modify** — add `FraudProtection` field; call `RecordSMSOTPVerified` in `VerifyOTP()` when `code.OOBChannel == SMS` |
| `pkg/lib/authenticationflow/declarative/node_authn_oob.go` | **Modify** — session counter (`SMSOTPSentCount++` on send, `SMSOTPVerifiedCount++` on verify) + implement `MilestoneOOBOTPSentPhoneTarget` |
| `pkg/lib/authenticationflow/declarative/node_verify_claim.go` | **Modify** — `DelayedOneTimeFunction` signature update only |
| `pkg/lib/authenticationflow/declarative/intent_account_recovery_flow_step_verify_account_recovery_code.go` | **Modify** — `SMSOTPVerifiedCount++` in session when SMS channel |
| `pkg/lib/authenticationflow/declarative/intent_login_flow.go` | **Modify** — add `OnCommitEffect` with root-flow guard to `GetEffects()` |
| `pkg/lib/authenticationflow/declarative/intent_signup_flow.go` | **Modify** — same |
| `pkg/lib/authenticationflow/declarative/intent_signup_login_flow.go` | **Modify** — same |
| `pkg/lib/authenticationflow/declarative/intent_reauth_flow.go` | **Modify** — same |
| `pkg/lib/authenticationflow/declarative/intent_promote_flow.go` | **Modify** — same |
| `pkg/lib/authenticationflow/declarative/intent_account_recovery_flow.go` | **Modify** — same |
| `pkg/api/event/nonblocking/fraud_protection_decision_recorded.go` | **Create** — audit event payload |
| `cmd/authgear/cmd/cmdaudit/migrations/audit/{timestamp}-add_audit_metrics_table.sql` | **Create** — partitioned table, 90-day retention |

**Reusable existing utilities:**
- `phone.ParsePhoneNumberWithUserInput(e164)` → `result.Alpha2[0]` — phone country code (`pkg/util/phone/`)
- `geoip.IPString(ip)` → `info.CountryCode` — IP geo lookup (`pkg/util/geoip/geoip.go`)
- `httputil.RemoteIP` — string type (`pkg/util/httputil/ip.go`)
- `auditdb.WriteSQLExecutor`, `auditdb.ReadSQLExecutor`, `auditdb.SQLBuilderApp` — follow pattern in `pkg/lib/audit/write_store.go` and `read_store.go`
- `appredis.Handle.WithConnContext()` — Redis access pattern from `pkg/lib/userinfo/userinfo.go`
- `authflow.FindMilestoneInCurrentFlow` — milestone search in flow tree
- Migration pattern from `cmd/authgear/cmd/cmdaudit/migrations/audit/20210531162744-add_audit_log_table.sql`

---

## Implementation Roadmap

Atomic commits in dependency order. No commit mixes content from two different parts.

### Part 1 — Configuration

**Commit 1: `config: add FraudProtectionConfig app config`**
Create `pkg/lib/config/fraud_protection.go` with all Go structs, JSON schema, `SetDefaults()`, and warning type constants. Add `FraudProtection *FraudProtectionConfig` field + schema reference to `AppConfig` in `config.go`.

**Commit 2: `tests: update app config testdata for fraud_protection`**
Update `testdata/default_config.yaml` to include the `fraud_protection` section with all defaults (enabled=true, all 5 warning types, action=record_only, empty always_allow). Add fraud_protection parse validation test cases to `testdata/config_tests.yaml` (unknown key, invalid decision action, invalid warning type). Create `testdata/fraud_protection_tests.yaml` with `part: FraudProtectionConfig` schema-level test cases and register it in `schema_test.go`'s `testFiles` list.

**Commit 3: `config: add FraudProtectionFeatureConfig feature config`**
Create `pkg/lib/config/feature_fraud_protection.go` with `FraudProtectionFeatureConfig`, `SetDefaults()` (default `IsModifiable = false`), and `Merge()`. Add `FraudProtection *FraudProtectionFeatureConfig` field + schema reference to `FeatureConfig` in `feature.go`.

**Commit 4: `tests: update feature config testdata for fraud_protection`**
Update `testdata/default_feature.yaml` to include `fraud_protection: {is_modifiable: false}`. Add fraud_protection test cases to `testdata/parse_feature_tests.yaml` (valid is_modifiable true, valid is_modifiable false, unknown key).

**Commit 5: `configsource: guard fraud_protection config when is_modifiable is false`**
In `pkg/lib/config/configsource/resources.go`, add a check inside `validateBasedOnFeatureConfig()`: if `!*fc.FraudProtection.IsModifiable` and the incoming `FraudProtection` config differs from the original, emit a validation error. This follows the same pattern as the existing biometric, password policy, and OAuth guards in that method.

**Commit 5b: `tests: add resources_test cases for fraud_protection is_modifiable guard`**
In `pkg/lib/config/configsource/resources_test.go`, add test cases for the new guard: modifying `fraud_protection` is blocked when `is_modifiable=false`; modifying `fraud_protection` is allowed when `is_modifiable=true`; saving with unchanged `fraud_protection` is allowed regardless of `is_modifiable`.

### Part 2 — Warning Implementation

**Commit 6: `audit/db: add _audit_metrics partitioned table`**
Add SQL migration creating the `_audit_metrics` table partitioned by `start_time` (monthly, 90-day retention via pg_partman), template table for PK propagation, and `(app_id, name, period, key, start_time)` index.

**Commit 7: `fraudprotection: implement MetricsStore`**
Create `pkg/lib/fraudprotection/metrics_store.go`. Implements `RecordVerified` (2-row INSERT), `GetVerifiedByCountry24h`, `GetVerifiedByCountry1h`, `GetVerifiedByIP24h`, `GetVerifiedByCountryPast14DaysRollingMax`. All read methods use a 5-minute Redis cache.

**Commit 8: `fraudprotection: implement LeakyBucketStore`**
Create `pkg/lib/fraudprotection/leaky_bucket_store.go`. Implements `RecordSMSOTPSent` (atomic Lua fill across 4 buckets + `ZADD` ip_countries ZSET, returns `LeakyBucketTriggered`) and `RecordSMSOTPVerified` (drain). Defines `LeakyBucketThresholds` and `LeakyBucketTriggered` types.

**Commit 9: `fraudprotection: implement Service`**
Create `pkg/lib/fraudprotection/service.go`. Implements `Service`, `CheckAndRecord`, `RecordSMSOTPVerified`, `RevertSMSOTPSent`, `ComputeThresholds`, `evaluateWarnings`, `isAlwaysAllowed`, `effectiveFraudProtectionConfig`, and `ErrBlockedByFraudProtection` error var.

**Commit 10: `fraudprotection: wire into dependency graph`**
Create `pkg/lib/fraudprotection/deps.go` (`DependencySet`). Add `FraudProtection *fraudprotection.Service` to `authflow.Dependencies` (`pkg/lib/authenticationflow/deps.go`) — needed by the `OnCommitEffect` `RevertSMSOTPSent` call. Add `FraudProtection FraudProtectionService` (interface) field to `messaging.Sender` (for `CheckAndRecord`) and to `otp.Service` (for `RecordSMSOTPVerified`). Wire into `pkg/lib/deps/deps_common.go`.

**Commit 11: `authflow: change DelayedOneTimeFunction to return a result struct`**
In `node.go`, introduce `DelayedOneTimeFunctionResult` struct and change `DelayedOneTimeFunction` to `func(ctx, deps) (DelayedOneTimeFunctionResult, error)`. Update all 5 existing call sites (`node_authn_oob.go`, `node_verify_claim.go`, `node_pre_authenticate.go`, `node_pre_initialize.go`, `node_post_identified.go`) to return `(DelayedOneTimeFunctionResult{}, err)` — pure signature refactor, no functional change.

**Commit 12: `authflow: add session SMS OTP count tracking infrastructure`**
In `node.go`, add `UpdatedSession *Session` to `NodeWithDelayedOneTimeFunction`. In `session.go`, add `SMSOTPSentCount`/`SMSOTPVerifiedCount` fields, `PatchFrom(*Session)` method (`*s = *updated`), store `*Session` pointer in `MakeContext`. In `context.go`, add `GetSession`, `GetSMSOTPSentCount`, `GetSMSOTPVerifiedCount`. In `accept.go`, add `PendingSessionPatches []*Session` to `AcceptResult` and handle `UpdatedSession` in both `NodeWithDelayedOneTimeFunction` paths in `doAccept`. In `service.go`, update `processAcceptResult` to apply `PendingSessionPatches` before delayed functions and `DelayedOneTimeFunctionResult.UpdatedSession` after each delayed function.

**Commit 13: `authflow/declarative: track SMS OTP sent and verified counts in session`**
In `milestone.go`, add `MilestoneOOBOTPSentPhoneTarget` interface. In `node_authn_oob.go`, update both send `DelayedOneTimeFunction` closures to return `DelayedOneTimeFunctionResult{UpdatedSession}` with `SMSOTPSentCount++` after `SendCode` succeeds; when OTP verification succeeds in `ReactTo`, return `NodeWithDelayedOneTimeFunction{Node: nextNode, UpdatedSession}` with `SMSOTPVerifiedCount++`; implement `MilestoneOOBOTPSentPhoneTarget`. In `intent_account_recovery_flow_step_verify_account_recovery_code.go`, when `verifyCode()` succeeds and channel is SMS, return `NodeWithDelayedOneTimeFunction{Node: nextNode, UpdatedSession}` with `SMSOTPVerifiedCount++`.

**Commit 14: `messaging: call fraud protection check before sending SMS`**
In `messaging/sender.go`, call `s.FraudProtection.CheckAndRecord(ctx, opts.To, messageType)` before the SMS send. Return the error immediately if the request is blocked.

**Commit 15: `otp: record verified SMS OTPs through fraud protection`**
In `pkg/lib/authn/otp/service.go`, add `FraudProtection FraudProtectionService` field. In `VerifyOTP()`, after `isCodeValid = true`, call `s.FraudProtection.RecordSMSOTPVerified(ctx, target)` when `code.OOBChannel == model.AuthenticatorOOBChannelSMS`. This single change covers all existing paths (auth OOB, claim verification, forgot password, account recovery) and all future SMS OTP paths automatically. No changes to authflow nodes, `forgotpassword/service.go`, or other call sites.

Note: email and WhatsApp OTPs are filtered out by the `code.OOBChannel == SMS` check at the source.

**Commit 16: `authflow/declarative: revert unverified SMS OTPs on flow completion`**
Add `OnCommitEffect` with `flows.Nearest != flows.Root` guard to `GetEffects()` of all 6 public flow intents (`intent_login_flow.go`, `intent_signup_flow.go`, `intent_signup_login_flow.go`, `intent_reauth_flow.go`, `intent_promote_flow.go`, `intent_account_recovery_flow.go`). Effect computes `unverifiedCount = sentCount - verifiedCount` and calls `FraudProtection.RevertSMSOTPSent(ctx, phoneNumber, unverifiedCount)` once with the total count.

**Commit 17: `e2e: extend test runner with audit database support`**
The existing `action: query` and `before: custom_sql` only target the GlobalDatabase. Since `_audit_metrics` lives in the audit database, two new step types are needed to test verified OTP behaviour:

Add `action: audit_query` to `pkg/testrunner/` and `e2e/cmd/e2e/pkg/sql_select.go` — identical to `action: query` but connects to `cfg.AuditDatabase` instead of `cfg.GlobalDatabase`. Output format and `query_output` assertion syntax are the same.

Add `before: custom_audit_sql` to `pkg/testrunner/` — identical to `before: custom_sql` but executes the SQL file against the audit database. The SQL template context (`{{ .AppID }}`, `{{ uuidv4 }}`, etc.) is unchanged.

**Commit 18: `e2e: add fraud protection unverified OTP blocking e2e tests`**
Create `e2e/tests/fraud_protection/` with three test files. Each configures `authgear.yaml` via `override` with `fraud_protection`, `test_mode.oob_otp` (fixed code `111111`), and `test_mode.sms.suppressed`. Flows are run sequentially within a single test case using multiple `action: create` blocks, leaving OTPs unverified between flows. Because each test has a fresh app ID, the leaky bucket starts at 0 and `_audit_metrics` has no history, so adaptive thresholds fall to their minimums (country hourly = 3, IP countries = 3), making each warning reachable in exactly 4 flows.

`sms_unverified_by_phone_country_hourly.test.yaml` — Tests `SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__HOURLY_THRESHOLD_EXCEEDED` with `deny_if_any_warning`. Three flows send to different SG numbers (`+6591230001`–`+6591230003`), each OTP send step succeeds (`verify_oob_otp_data`). The 4th flow's OTP send step for `+6591230004` returns `{"name": "Forbidden", "reason": "BlockedByFraudProtection", "code": 403}`.

`sms_phone_countries_by_ip_daily.test.yaml` — Tests `SMS__PHONE_COUNTRIES__BY_IP__DAILY_THRESHOLD_EXCEEDED` with `deny_if_any_warning`. Three flows send to distinct countries (SG `+6591230001`, HK `+85291230001`, MY `+60123450001`); each succeeds. The 4th flow to JP `+819012340001` is blocked with the same error.

`record_only.test.yaml` — Tests that `action: record_only` does NOT block. Same four-flow SG scenario but with `decision.action: record_only`; the 4th OTP send succeeds, proving the default configuration never blocks.

`enabled_false.test.yaml` — Tests that `fraud_protection.enabled: false` fully disables the feature. Configures `deny_if_any_warning` (which would block in the enabled case) but sets `enabled: false`. Runs the same four-flow SG scenario; all 4 OTP send steps succeed. Then uses `action: audit_query` against `_audit_log` to assert zero rows with `activity_type = 'fraud_protection.decision_recorded'` for this app ID, confirming no audit event was dispatched either. This test depends on Commit 17 (audit DB query support) and is only meaningful once Commit 21 (event dispatch) is implemented.

**Commit 19: `e2e: add fraud protection verified OTP e2e tests`**
Two more test files in `e2e/tests/fraud_protection/` that exercise the verified OTP path, using the audit DB infrastructure from Commit 17.

`verified_otp_writes_to_metrics.test.yaml` — Completes a full SMS OTP signup (identify phone → select channel → submit code `111111` → flow finishes). Then uses `action: audit_query` to assert that exactly 2 rows exist in `_audit_metrics` for this app ID with `name = 'sms_otp_verified'` — one with `key = 'ip:{ip}'` and one with `key = 'phone_country:SG'`. This confirms `RecordSMSOTPVerified` → `MetricsStore.RecordVerified` is wired correctly end-to-end.

`verified_otp_history_raises_threshold.test.yaml` — Uses `before: custom_audit_sql` to insert 30 `sms_otp_verified` rows for `phone_country:SG` with `start_time` in the past 30 minutes. With this history, `GetVerifiedByCountry1h(SG) = 30` → country hourly threshold = `max(3, 20/6, 30×0.2)` = 6. The test then runs 4 signup flows to SG numbers with `deny_if_any_warning`; all 4 OTP send steps succeed (level 4 ≤ threshold 6). A 7th flow is blocked. This directly proves that verified OTP history raises the adaptive threshold: without the pre-populated history the 4th would already be blocked (as shown in `sms_unverified_by_phone_country_hourly.test.yaml`).

### Part 3 — Decision Record & Audit Log

**Commit 20: `event: add fraud_protection.decision_recorded audit event type`**
Create `pkg/api/event/nonblocking/fraud_protection_decision_recorded.go` with `FraudProtectionDecisionRecord` struct, `FraudProtectionDecisionRecordedEventPayload`, and implementations of `NonBlockingEventType()`, `ForAudit() = true`, `ForHook() = false`.

**Commit 21: `fraudprotection: dispatch decision_recorded event from CheckAndRecord`**
Add `EventService` interface and field to `Service` struct. In `CheckAndRecord`, after warning evaluation and before the block-or-allow decision, call `EventService.DispatchEventImmediately` with `FraudProtectionDecisionRecordedEventPayload` (timestamp, decision, action, triggered warnings, IP, geo code).

**Dependency notes:**
- Commits 1, 3 before 9 (Service reads app config and feature config)
- Commits 7–8 before 9 (Service uses both stores)
- Commit 9 before 10 (DI wiring references the concrete type)
- Commit 10 before 11–19 (FraudProtection field must exist on deps/Sender)
- Commit 11 before 12–13 (type change must compile first)
- Commit 12 before 13 (session infrastructure required for count tracking)
- Commit 13 before 16 (`MilestoneOOBOTPSentPhoneTarget` required for the effect)
- Commits 14–16 before 18–19 (e2e tests require full Part 2 implementation)
- Commit 17 before 18 (audit DB query required by `enabled_false.test.yaml`)
- Commit 17 before 19 (audit DB step types required by verified OTP tests)
- Commits 20–21 before `enabled_false.test.yaml` is meaningful (event dispatch must exist for the zero-rows assertion to be non-trivial; the test file can be committed earlier but only proves the property after Commit 21)
- Commit 20 before 21 (event type must exist before dispatch)

---

## Verification

1. **Unit tests** for threshold formulas and leaky bucket drain calculations in `pkg/lib/fraudprotection/`
2. **Config tests**: `SetDefaults()` produces correct defaults; `is_modifiable=false` ignores app config
3. **Config source guard**: updating `authgear.yaml` with a changed `fraud_protection` section returns a validation error when `is_modifiable=false`; allowed when `is_modifiable=true`
4. **Integration**: send >3 countries from same IP → `SMS__PHONE_COUNTRIES__BY_IP__DAILY_THRESHOLD_EXCEEDED` triggers; send >20 unverified OTPs to a country → country unverified warning triggers
5. **Leaky bucket recovery**: verify that after the window elapses with no new sends, the leaky bucket level decays to 0 and warnings stop triggering
6. **Audit log**: `fraud_protection.decision_recorded` event appears after a blocked/allowed SMS attempt
7. **API error**: `{"name":"Forbidden","reason":"BlockedByFraudProtection","code":403}` when `action: deny_if_any_warning`
8. **Always-allow**: whitelisted IP/phone bypasses the check entirely
