# Fraud Protection: Exclude Verified Countries from IP-Country Warning

## Summary

Adjust `SMS__PHONE_COUNTRIES__BY_IP__DAILY_THRESHOLD_EXCEEDED` so it counts only countries that have **no verified SMS OTP in the same 24h window**.

If an IP has at least one verified OTP for a country during that window, that country is excluded from the distinct-country count. The threshold remains `3`.

This is a behavior change only. No backward-compatibility work is required.

## Runtime Change

### Current behavior

`pkg/lib/fraudprotection/leaky_bucket_store.go` currently tracks distinct phone countries per IP in the Redis ZSET keyed by:

`app:{appID}:fraud_protection:ip_countries:{ip}`

`RecordSMSOTPSent(...)` updates that ZSET and triggers `IPCountriesDaily` when the raw distinct-country count exceeds the fixed threshold.

### New behavior

Add a second Redis ZSET per IP to record countries that have at least one verified SMS OTP in the same 24h window:

`app:{appID}:fraud_protection:ip_verified_countries:{ip}`

The send-path warning becomes:

1. Count distinct countries seen from the IP in the last 24h.
2. Remove any country that appears in the verified-country ZSET for that IP.
3. Trigger `SMS__PHONE_COUNTRIES__BY_IP__DAILY_THRESHOLD_EXCEEDED` only when the filtered count is `> 3`.

### Interface Details

#### `pkg/lib/fraudprotection/service.go`

`LeakyBucketer` becomes:

```go
type LeakyBucketer interface {
    RecordSMSOTPSent(ctx context.Context, ip, phoneCountry string, thresholds LeakyBucketThresholds) (LeakyBucketTriggered, LeakyBucketLevels, error)
    RecordSMSOTPVerified(ctx context.Context, ip, phoneCountry string, thresholds LeakyBucketThresholds, count int) error
    RecordSMSOTPVerifiedCountry(ctx context.Context, ip, phoneCountry string) error
}
```

`Service.RecordSMSOTPVerified(ctx, phoneNumber)` keeps the current parse/write/drain flow, but now invokes the bucket store in this order:

1. read `ip` from the request context and parse `phoneNumber` to get `phoneCountry`
2. `Metrics.RecordVerified(ctx, ip, phoneCountry)`
3. `LeakyBucket.RecordSMSOTPVerifiedCountry(ctx, ip, phoneCountry)`
4. `RevertSMSOTPSent(ctx, phoneNumber, 1)`

The verified-country update is a side effect of a successful SMS OTP consumption only. It is not part of the alt-auth revert path.

This split matters because the lower-level `LeakyBucketStore.RecordSMSOTPVerified(...)` method is also used by `RevertSMSOTPSent(...)` to drain unverified OTPs that were sent during a flow but never consumed. If the same method also marked a country as verified, alt-auth cleanup would incorrectly promote unverified sends into the verified-country set.

So the invariant is:

- service-level `RecordSMSOTPVerified(...)` = actual verification event
- store-level `RecordSMSOTPVerified(...)` = drain-only bookkeeping for verified and reverted counts
- store-level `RecordSMSOTPVerifiedCountry(...)` = explicit marker for a real verified OTP

#### `pkg/lib/fraudprotection/leaky_bucket_store.go`

`leakyBucketScript` remains unchanged and continues to be used only for the four leaky buckets on the send path.

`ipCountriesScript` remains the script that computes the IP-country warning on the send path. That is where the filtered distinct-country count is evaluated.

The verified-country marker is implemented by `RecordSMSOTPVerifiedCountry(ctx, ip, phoneCountry)`, which writes to the IP-scoped verified-country ZSET. It can reuse the same 24h retention model as the send-path country ZSET, but it is intentionally a separate store method so the service can call it only for real OTP consumption, not for alt-auth cleanup.

`RecordSMSOTPVerified(ctx, ip, phoneCountry, thresholds, count)` remains drain-only and is still used by `RevertSMSOTPSent(...)` for unverified OTP cleanup.

The filtered IP-country count is computed in the send-path script, not in Go:

```go
res, err := conn.Eval(ctx, ipCountriesScript,
    []string{s.ipCountriesKey(ip), s.ipVerifiedCountriesKey(ip)},
    phoneCountry, now, ipCountriesThreshold, 2*bucketWindowDaily,
).Slice()
```

`ipCountriesScript` is responsible for:

1. `ZADD` the sent country into `ip_countries`
2. prune expired entries from both `ip_countries` and `ip_verified_countries`
3. `ZRANGE` both ZSETs and build a Lua lookup table for the verified countries
4. count the distinct sent countries that are not present in the verified-country lookup table
5. return `{filtered_count, triggered_int}`

This keeps the send-path warning atomic with the country update and avoids a race between sent-country recording and verified-country exclusion.

## File-Level Changes

### `pkg/lib/fraudprotection/leaky_bucket_store.go`

- Keep the existing `ip_countries` ZSET and the four leaky buckets unchanged.
- Add a verified-country ZSET helper named `ipVerifiedCountriesKey(ip string) string`.
- Add a new store method `RecordSMSOTPVerifiedCountry(ctx context.Context, ip, phoneCountry string) error`.
- Keep `RecordSMSOTPVerified(...)` drain-only.
- Update `RecordSMSOTPSent(...)` so the IP-country warning uses the filtered count described above.

### `pkg/lib/fraudprotection/service.go`

- Extend `LeakyBucketer` with the new verified-country recording method.
- Update `Service.RecordSMSOTPVerified(...)` so the verified-OTP flow becomes:
  1. parse the phone number
  2. write the `sms_otp_verified` metric
  3. call `RecordSMSOTPVerifiedCountry(...)`
  4. call `RevertSMSOTPSent(..., 1)` to drain the leaky buckets through the existing path
- Leave `CheckAndRecord(...)`, threshold computation, and warning mapping unchanged.

### `docs/specs/fraud-protection.md`

- Rewrite the `SMS__PHONE_COUNTRIES__BY_IP__DAILY_THRESHOLD_EXCEEDED` section to state that the warning counts only countries without a verified SMS OTP in the same 24h window.
- Keep the threshold at `3`.

## Test Plan

### Unit tests

#### `pkg/lib/fraudprotection/leaky_bucket_store_test.go`

- Add coverage for the new verified-country key helper.
- Add a case proving that a verified country is excluded from the IP-country count.
- Add a case proving the verified-country marker respects the same 24h expiry behavior as the existing country set.
- Keep the current regression that proves four unverified countries from one IP still trigger the warning.
- Add a case proving `RecordSMSOTPVerifiedCountry(...)` is called only for actual verification, not for alt-auth cleanup.
- Add a case proving `RevertSMSOTPSent(...)` still drains the buckets and does not mark verified countries.

#### `pkg/lib/fraudprotection/service_test.go`

- Extend the leaky-bucket stub with the new verified-country method.
- Add a test that `RecordSMSOTPVerified(...)` records the verified-country marker and still drains the buckets.

### E2E tests

- Keep `e2e/tests/fraud_protection/sms_phone_countries_by_ip_daily.test.yaml` as the baseline regression for the unverified-country case.
- Add a new e2e test under `e2e/tests/fraud_protection/` that:
  - verifies one country first
  - sends unverified OTPs to three other countries successfully
  - blocks on the 4th unverified country
  - proves the verified country does not contribute to the threshold

## Assumptions

- “In the period” means the existing 24h sliding window.
- The verified-country marker is keyed by IP because the warning itself is IP-scoped.
- Old Redis keys can expire naturally; no migration or backfill is needed.
- No config schema, database schema, or generated code changes are required.

## Implementation Order

1. Add the verified-country Redis storage and filtered counting logic in `pkg/lib/fraudprotection/leaky_bucket_store.go`.
2. Wire the new method through `pkg/lib/fraudprotection/service.go`.
3. Update unit tests for store and service behavior.
4. Update the fraud-protection spec text.
5. Add the e2e regression test for the verified-country exclusion case.

## Atomic Commits

1. `fraud: exclude verified countries from SMS IP-country counting`
   - Files: `pkg/lib/fraudprotection/leaky_bucket_store.go`, `pkg/lib/fraudprotection/service.go`, `pkg/lib/fraudprotection/leaky_bucket_store_test.go`, `pkg/lib/fraudprotection/service_test.go`
   - Scope: storage, service wiring, and unit coverage.
2. `doc,e2e: update SMS IP-country fraud protection semantics`
   - Files: `docs/specs/fraud-protection.md`, `e2e/tests/fraud_protection/*.test.yaml`
   - Scope: spec wording and end-to-end regression coverage.
