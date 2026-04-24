# SMS OTP Unverified Count Drained Audit Metric

## Summary

Add a new PostgreSQL audit metric, `sms_otp_unverified_count_drained`, in `pkg/lib/fraudprotection/metrics_store.go`. This metric records every successful drain of the unverified SMS OTP bucket, whether the drain came from a real verification or from alt-auth cleanup after a flow completed without consuming the OTP.

The derived “reverted but actually not verified” count is computed from audit metrics as:

`sms_otp_unverified_count_drained - sms_otp_verified`

for the same `app_id`, `name`, `key`, and time window. No OTel metric changes are required.

At the same time, rename the low-level `LeakyBucketStore` methods so the API matches their actual behavior. This rename is scoped to the concrete store and its internal wiring; `fraudprotection.Service` keeps the same public methods.

## Runtime Flow

### Verified OTP path

`pkg/lib/authn/otp/service.go` already calls `FraudProtection.RecordSMSOTPVerified(ctx, phoneNumber)` when an SMS OTP is consumed. That flow remains the same:

1. parse the phone number
2. write the existing `sms_otp_verified` audit metric
3. mark the IP-scoped verified-country Redis ZSET
4. drain the leaky buckets by 1 unit

The drain step should now call `LeakyBucketer.DrainUnverifiedSMSOTPSent(...)`.

### Revert path

`pkg/lib/authenticationflow/declarative/utils_fraud_protection.go` remains the case-2 hook for flows that complete successfully through another factor. It computes:

`sentCount - verifiedCount`

and calls `FraudProtection.RevertSMSOTPSent(ctx, phoneNumber, unverifiedCount)` once.

`Service.RevertSMSOTPSent(...)` should then:

1. parse the phone number
2. compute thresholds
3. call `LeakyBucketer.DrainUnverifiedSMSOTPSent(...)`
4. if the drain succeeds, record `sms_otp_unverified_count_drained` for the same `count`

This keeps the metric aligned with the semantic operation: `RecordSMSOTPVerified(...)` is still the verified path, while `RevertSMSOTPSent(...)` records all cleanup drains.

## File-Level Changes

### `pkg/lib/fraudprotection/metrics_store.go`

- Add a new metric name constant, `metricsNameSMSOTPUnverifiedCountDrained`.
- Add `RecordUnverifiedSMSOTPCountDrained(ctx, ip, phoneCountry string, count int) error`.
- Reuse the existing `_audit_metrics` insert shape so both verified and reverted audit metrics write the same key structure:
  - `ip:{ip}`
  - `phone_country:{alpha2}`
- Insert `count` row-pairs in one transaction, so a one-unit drain writes 2 rows and a three-unit drain writes 6 rows.
- Factor the common insert logic into a private helper if needed, but keep the public API minimal and explicit.

### `pkg/lib/fraudprotection/service.go`

- Extend `MetricsQuerier` with `RecordUnverifiedSMSOTPCountDrained(ctx, ip, phoneCountry string, count int) error`.
- Rename the `LeakyBucketer` methods from `RecordSMSOTPSent(...)` / `RecordSMSOTPVerified(...)` to `RecordUnverifiedSMSOTPSent(...)` / `DrainUnverifiedSMSOTPSent(...)`.
- Update `Service.RecordSMSOTPVerified(...)` to call the renamed drain method.
- Update `Service.RevertSMSOTPSent(...)` so it records `sms_otp_unverified_count_drained` after a successful drain.
- Keep the parse-and-skip behavior for unparseable phone numbers unchanged.
- Keep `Service.RecordSMSOTPVerified(...)` and `Service.RevertSMSOTPSent(...)` as the public fraud-protection entry points; the rename is only in the concrete `LeakyBucketStore` implementation and the internal interface that wires it.

### `pkg/lib/fraudprotection/leaky_bucket_store.go`

- Rename the concrete `LeakyBucketStore` methods to `RecordUnverifiedSMSOTPSent(...)` and `DrainUnverifiedSMSOTPSent(...)`.
- Keep the Lua script, thresholds, Redis keys, and drain semantics unchanged.
- Update comments so the method clearly means “drain sent OTPs”, not “record verification”.

### `pkg/lib/fraudprotection/service_test.go`

- Extend the stub metrics implementation with `RecordUnverifiedSMSOTPCountDrained(...)`.
- Rename the stub leaky-bucket methods to match the `LeakyBucketStore` names.
- Add a test that `RecordSMSOTPVerified(...)` still:
  - writes the verified audit metric
  - records the verified-country marker
  - drains once
- Add a test that `RevertSMSOTPSent(ctx, ..., count)`:
  - drains `count` units
  - writes `sms_otp_unverified_count_drained`
- Add a failure-path test showing that if the drain fails, the revert audit metric is not written.

### `pkg/lib/fraudprotection/leaky_bucket_store_test.go`

- Rename the store-level drain tests to the new method name.
- Keep the existing regression coverage for:
  - draining an empty bucket
  - draining by `count`
  - not marking verified countries from the drain path

### `pkg/lib/authenticationflow/declarative/utils_fraud_protection_test.go`

- Add a focused unit test for `revertUnverifiedSMSOTPs(...)` so the case-2 path is explicit and the function still passes only the unverified remainder into `RevertSMSOTPSent(...)`.

### `e2e/tests/fraud_protection/verified_otp_writes_to_metrics.test.yaml`

- Extend the existing verified-OTP integration test to assert that one verified SMS OTP writes both:
  - `sms_otp_verified`
  - `sms_otp_unverified_count_drained`

### `e2e/tests/fraud_protection/account_recovery_verified_otp_reverts_leaky_bucket.test.yaml`

- Extend the existing cleanup-path regression to assert that:
  - successful account-recovery completion writes `sms_otp_unverified_count_drained`
  - the revert metric reflects the unverified cleanup amount
  - the verified metric still only counts real OTP consumption

## Compatibility and Deployment

- No schema migration is required.
- No Redis key format changes are required.
- The new audit metric is additive only and starts at zero after deployment.
- Historical dashboards that subtract `sms_otp_verified` from `sms_otp_unverified_count_drained` must treat rollout day as the start of reliable data for the new metric.
- The method rename is scoped to the `LeakyBucketStore` implementation and its internal wiring; `fraudprotection.Service` methods remain `RecordSMSOTPVerified(...)` and `RevertSMSOTPSent(...)`.

## Test Plan

- `go test ./pkg/lib/fraudprotection`
- `go test ./pkg/lib/authn/otp`
- `go test ./pkg/lib/authenticationflow/declarative`
- Run the two affected e2e cases:
  - `e2e/tests/fraud_protection/verified_otp_writes_to_metrics.test.yaml`
  - `e2e/tests/fraud_protection/account_recovery_verified_otp_reverts_leaky_bucket.test.yaml`

The critical regressions are:

- verified SMS OTPs still write the verified audit metric and drain one unit
- alt-auth cleanup still drains only the unverified remainder
- the new revert audit metric counts total drains, so the derived unverified-cleanup count can be computed from audit metrics

## Implementation Order

1. Add `sms_otp_unverified_count_drained` to `pkg/lib/fraudprotection/metrics_store.go`.
2. Rename the `LeakyBucketStore` methods to `RecordUnverifiedSMSOTPSent(...)` and `DrainUnverifiedSMSOTPSent(...)`, then wire the new metric into `Service.RevertSMSOTPSent(...)`.
3. Update service and store unit tests.
4. Extend the two existing e2e tests.

## Atomic Commits

1. `fraud: add SMS OTP unverified-drain audit metric and rename LeakyBucketStore methods`
   - Files: `pkg/lib/fraudprotection/metrics_store.go`, `pkg/lib/fraudprotection/service.go`, `pkg/lib/fraudprotection/leaky_bucket_store.go`, `pkg/lib/fraudprotection/service_test.go`, `pkg/lib/fraudprotection/leaky_bucket_store_test.go`, `pkg/lib/authenticationflow/declarative/utils_fraud_protection_test.go`
   - Scope: audit metric storage, drain rename, and unit coverage.
2. `doc,e2e: verify SMS OTP unverified-drain metrics`
   - Files: `e2e/tests/fraud_protection/verified_otp_writes_to_metrics.test.yaml`, `e2e/tests/fraud_protection/account_recovery_verified_otp_reverts_leaky_bucket.test.yaml`
   - Scope: end-to-end regression coverage.
