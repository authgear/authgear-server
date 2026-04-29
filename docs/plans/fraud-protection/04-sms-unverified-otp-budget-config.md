# SMS Unverified OTP Budget Config

## Summary

Add additive SMS fraud-protection budget configuration under `fraud_protection.sms.unverified_otp_budget`.

The new config is a ratio-based budget:

- `daily_ratio` defaults to `0.3`
- `hourly_ratio` defaults to `0.2`
- `by_phone_country` provides per-recipient-country overrides via grouped country lists

The hourly threshold is no longer derived from the daily threshold. It uses the 14-day rolling max directly.

This is a backward-compatible config change. No database migration, Redis migration, or generated code change is required.

---

## Config Model And Schema

### Target YAML

```yaml
fraud_protection:
  enabled: true
  sms:
    unverified_otp_budget:
      daily_ratio: 0.3
      hourly_ratio: 0.2
      by_phone_country:
        - geo_location_codes: ["HK", "SG"]
          daily_ratio: 0.15
          hourly_ratio: 0.1
        - geo_location_codes: ["JP"]
          daily_ratio: 0.25
  warnings:
    - type: SMS__PHONE_COUNTRIES__BY_IP__DAILY_THRESHOLD_EXCEEDED
    - type: SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__DAILY_THRESHOLD_EXCEEDED
    - type: SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__HOURLY_THRESHOLD_EXCEEDED
    - type: SMS__UNVERIFIED_OTPS__BY_IP__DAILY_THRESHOLD_EXCEEDED
    - type: SMS__UNVERIFIED_OTPS__BY_IP__HOURLY_THRESHOLD_EXCEEDED
  decision:
    action: record_only
```

### Config structs

- `pkg/lib/config/fraud_protection.go` keeps `FraudProtectionConfig` as the top-level app config type.
- Add `SMS *FraudProtectionSMSConfig` to `FraudProtectionConfig`.
- Add `UnverifiedOTPBudget *FraudProtectionSMSUnverifiedOTPBudgetConfig` to `FraudProtectionSMSConfig`.
- Add `DailyRatio *float64`, `HourlyRatio *float64`, and `ByPhoneCountry []*FraudProtectionSMSUnverifiedOTPBudgetByPhoneCountryConfig` to `FraudProtectionSMSUnverifiedOTPBudgetConfig`.
- Add `GeoLocationCodes []string`, `DailyRatio *float64`, and `HourlyRatio *float64` to `FraudProtectionSMSUnverifiedOTPBudgetByPhoneCountryConfig`.
- Use `omitempty` on every new field.
- Mark the new optional pointer fields with `nullable:"true"` so `SetFieldDefaults()` treats explicit `null` the same as omission.
- Use pointers for the ratio fields that are optional per country so omitted values can fall back independently.

### JSON Schema

- Add `FraudProtectionSMSConfig`, `FraudProtectionSMSUnverifiedOTPBudgetConfig`, and `FraudProtectionSMSUnverifiedOTPBudgetByPhoneCountryConfig` schema definitions in `pkg/lib/config/fraud_protection.go`.
- `daily_ratio` and `hourly_ratio` must be numeric values in the range `0..1`.
- `by_phone_country` remains a list, not a map.
- `geo_location_codes` remains a list of two-letter ISO 3166-1 alpha-2 codes.
- Mark the optional new fields as `nullable` in the JSON schema and `nullable:"true"` in the Go struct tags so explicit `null` is accepted the same as omission.

### Defaulting

- `FraudProtectionConfig.SetDefaults()` must populate `sms.unverified_otp_budget.daily_ratio = 0.3` and `hourly_ratio = 0.2`.
- `by_phone_country` defaults to `nil`.
- The existing top-level defaults remain unchanged: `enabled = true`, `warnings` contains the 5 existing warning types, and `decision.action = record_only`.

### Validation

- `pkg/lib/config/config.go` should validate `geo_location_codes` arrays, including type and element-level constraints.
- The override list is order-sensitive, so resolution uses the first matching item rather than rejecting duplicate country codes across items.
- No new feature-flag schema is needed. The existing `fraud_protection.is_modifiable` guard already covers the entire `fraud_protection` subtree.

---

## Runtime Flow

### Entry point

`pkg/lib/fraudprotection/service.go` keeps the current entry points:

- `CheckAndRecord(ctx, phoneNumber, messageType string) error`
- `RecordSMSOTPVerified(ctx, phoneNumber string) error`
- `RevertSMSOTPSent(ctx, phoneNumber string, count int) error`

### Threshold resolution

- Add a helper in `Service` that resolves the effective SMS budget for a phone country.
- The helper should return one global ratio pair and the first matching country-specific override pair when present.
- Country overrides are matched by exact country code using list order in `by_phone_country` and any entry in `geo_location_codes`.
- If a country override omits one ratio, that dimension falls back to the global ratio.
- IP-based thresholds always use the global ratios.

### Threshold formulas

`ComputeThresholds(ctx, ip, phoneCountry)` keeps the same historical metric queries, but the formulas change to:

```text
country_daily = max(
  20,
  verified_otps_to_country_past_14_days_rolling_max * effective_daily_ratio,
  verified_otps_to_country_past_24h * effective_daily_ratio,
)

country_hourly = max(
  3,
  verified_otps_to_country_past_14_days_rolling_max / 6 * effective_hourly_ratio,
  verified_otps_to_country_past_1h * effective_hourly_ratio,
)

ip_daily = max(
  10,
  verified_otps_by_ip_past_24h * global_daily_ratio,
)

ip_hourly = max(
  5,
  verified_otps_by_ip_past_24h / 6 * global_hourly_ratio,
)
```

- `country_daily` uses the effective per-country daily ratio.
- `country_hourly` uses the effective per-country hourly ratio and no longer depends on `country_daily`.
- `ip_daily` and `ip_hourly` use the global ratios only.
- The existing floor values remain unchanged.

### Call flow

1. `CheckAndRecord` parses the phone number and resolves the phone country.
2. `ComputeThresholds` queries the 14-day max, 24h verified counts, 1h verified counts, and IP verified counts.
3. `ComputeThresholds` resolves the global ratios and any matching `by_phone_country` override.
4. `ComputeThresholds` computes the four leaky-bucket thresholds from those ratios.
5. `RecordSMSOTPVerified` and `RevertSMSOTPSent` continue to reuse `ComputeThresholds` and the existing leaky-bucket methods.

### Constants

- Remove the hardcoded `thresholdScaleFactor` dependency from `pkg/lib/fraudprotection/service.go`.
- Keep the `thresholdHoursPerDay` scaling constant for the `/ 6` hourly conversion.

---

## Compatibility And Deployment

- The change is additive for the config schema.
- Apps that omit `fraud_protection.sms` or `fraud_protection.sms.unverified_otp_budget` will receive the default ratios from `SetDefaults()`.
- No legacy config key needs migration.
- No Redis key format changes are needed.
- No SQL migration is needed.
- The existing `fraud_protection` feature flag behavior stays the same because the guard already resets the full subtree to defaults when `is_modifiable=false`.

---

## File-Level Change Plan

- `pkg/lib/config/fraud_protection.go`: add the new SMS budget structs, schema, and defaulting logic.
- `pkg/lib/config/config.go`: validate `by_phone_country.country_codes` arrays in `AppConfig.Validate()`.
- `pkg/lib/config/testdata/default_config.yaml`: add the default `fraud_protection.sms.unverified_otp_budget` block.
- `pkg/lib/config/testdata/fraud_protection_tests.yaml`: add schema cases for the new nested config, valid overrides, and invalid ratio values.
- `pkg/lib/config/testdata/config_tests.yaml`: add app-config validation cases for first-match override resolution.
- `pkg/lib/config/config_test.go`: update the default round-trip and `ApplyFeatureConfigConstraints` coverage for the new nested defaults.
- `pkg/lib/fraudprotection/service.go`: add the ratio-resolution helper and update `ComputeThresholds()` to use the new formulas.
- `pkg/lib/fraudprotection/service_test.go`: update threshold expectations and add override-resolution coverage.
- `e2e/tests/fraud_protection/verified_otp_history_raises_threshold.test.yaml`: update the explanatory comment and expected threshold math to reflect the new default ratios and hourly formula.
- `e2e/tests/fraud_protection/sms_unverified_by_phone_country_hourly.test.yaml`: keep as the baseline unverified-country hourly regression, but verify it still matches the new config defaults.
- `e2e/tests/fraud_protection/` new test: add a dedicated country-override regression that proves `by_phone_country` changes the threshold only for the matching recipient country.

---

## Test Plan

### Config tests

- Verify the new nested schema accepts `sms.unverified_otp_budget`.
- Verify the default config round-trips with `daily_ratio = 0.3` and `hourly_ratio = 0.2`.
- Verify per-country overrides parse correctly.
- Verify ordered first-match override resolution.
- Verify out-of-range ratios are rejected.

### Service tests

- Verify `ComputeThresholds()` uses `daily_ratio = 0.3` by default.
- Verify `ComputeThresholds()` uses `hourly_ratio = 0.2` by default.
- Verify `country_hourly` is computed from `verified_otps_to_country_past_14_days_rolling_max / 6 * hourly_ratio`, not from `country_daily / 6`.
- Verify `by_phone_country` overrides apply only to the matching recipient country.
- Verify IP thresholds ignore `by_phone_country` overrides and continue to use the global ratios.

### E2E tests

- Update the existing hourly-history regression so the explanatory threshold math matches the decoupled formula.
- Add one e2e case that sets a phone-country override and proves the override changes the trigger point for that country only.
- Keep the existing verified-OTP metric coverage unchanged except for any updated comments or expected numbers that depend on the new defaults.

---

## Fixed Behavioral Decisions

- Default daily ratio is `0.3`.
- Default hourly ratio is `0.2`.
- `by_phone_country` is keyed by the recipient phone country, not the sender IP country.
- Daily and hourly ratios are independent.
- Country overrides can change the daily and hourly ratio independently.
- IP-based warnings never consult `by_phone_country`.

---

## Implementation Order

1. Add the nested SMS budget config structs, schema, and defaulting logic in `pkg/lib/config/fraud_protection.go`.
2. Validate `by_phone_country.country_codes` arrays in `pkg/lib/config/config.go`.
3. Update `pkg/lib/fraudprotection/service.go` to resolve ratios from config and compute thresholds with the new formulas.
4. Update config fixtures and config unit tests.
5. Update service unit tests and the existing fraud-protection e2e comments and expectations.
6. Update the existing fraud-protection e2e comments and expectations to match the final implementation.

---

## Atomic Commits

1. `fraud: add SMS unverified OTP budget config`
   Files: `pkg/lib/config/fraud_protection.go`, `pkg/lib/config/config.go`, `pkg/lib/config/testdata/default_config.yaml`, `pkg/lib/config/testdata/fraud_protection_tests.yaml`, `pkg/lib/config/testdata/config_tests.yaml`, `pkg/lib/config/config_test.go`
   Scope: config structs, schema, defaulting, and validation.
2. `fraud: use configurable SMS OTP ratios in threshold computation`
   Files: `pkg/lib/fraudprotection/service.go`, `pkg/lib/fraudprotection/service_test.go`
   Scope: runtime threshold resolution and unit coverage.
3. `e2e: verify SMS OTP budget ratios`
   Files: `e2e/tests/fraud_protection/verified_otp_history_raises_threshold.test.yaml`, `e2e/tests/fraud_protection/sms_unverified_by_phone_country_hourly.test.yaml`, `e2e/tests/fraud_protection/*.test.yaml` for the new override regression
   Scope: end-to-end coverage.
