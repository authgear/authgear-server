# Fraud Protection

- [SMS Pumping](#sms-pumping)
  - [Config](#config)
  - [Warnings](#warnings)
    - [Naming Convention](#naming-convention)
    - [SMS__PHONE_COUNTRIES__BY_IP__DAILY_THRESHOLD_EXCEEDED](#sms__phone_countries__by_ip__daily_threshold_exceeded)
    - [SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__DAILY_THRESHOLD_EXCEEDED](#sms__unverified_otps__by_phone_country__daily_threshold_exceeded)
    - [SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__HOURLY_THRESHOLD_EXCEEDED](#sms__unverified_otps__by_phone_country__hourly_threshold_exceeded)
    - [SMS__UNVERIFIED_OTPS__BY_IP__DAILY_THRESHOLD_EXCEEDED](#sms__unverified_otps__by_ip__daily_threshold_exceeded)
    - [SMS__UNVERIFIED_OTPS__BY_IP__HOURLY_THRESHOLD_EXCEEDED](#sms__unverified_otps__by_ip__hourly_threshold_exceeded)
    - [Notes](#notes)
  - [Decision Record](#decision-record)
  - [API Error](#api-error)
  - [Examples](#examples)
- [Future Work](#future-work)
  - [Country Based Risk Classification](#country-based-risk-classification)
  - [Decision: Challenge](#decision-challenge)
  - [Warning: Custom](#warning-custom)

## SMS Pumping

### Config

**authgear.yaml**

An example:

```yaml
fraud_protection:
  enabled: true
  sms:
    unverified_otp_budget:
      daily_ratio: 0.3
      hourly_ratio: 0.2
      by_phone_country:
        - country_codes: ["HK", "SG"]
          daily_ratio: 0.15
          hourly_ratio: 0.1
        - country_codes: ["JP"]
          daily_ratio: 0.25
  warnings:
    - type: SMS__PHONE_COUNTRIES__BY_IP__DAILY_THRESHOLD_EXCEEDED
    - type: SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__DAILY_THRESHOLD_EXCEEDED
    - type: SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__HOURLY_THRESHOLD_EXCEEDED
    - type: SMS__UNVERIFIED_OTPS__BY_IP__DAILY_THRESHOLD_EXCEEDED
    - type: SMS__UNVERIFIED_OTPS__BY_IP__HOURLY_THRESHOLD_EXCEEDED
  decision:
    always_allow:
      # If any rule matches, the request is always allowed regardless of warnings.
      ip_address:
        cidrs: ["123.123.1.1/32"]
        geo_location_codes: ["HK"]
      phone_number:
        geo_location_codes: ["HK", "US"]
        regex: ["^\\+852\\d*$"]
    action: deny_if_any_warning # record_only or deny_if_any_warning
```

The default:


```yaml
fraud_protection:
  enabled: true
  sms:
    unverified_otp_budget:
      daily_ratio: 0.3
      hourly_ratio: 0.2
  warnings:
    - type: SMS__PHONE_COUNTRIES__BY_IP__DAILY_THRESHOLD_EXCEEDED
    - type: SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__DAILY_THRESHOLD_EXCEEDED
    - type: SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__HOURLY_THRESHOLD_EXCEEDED
    - type: SMS__UNVERIFIED_OTPS__BY_IP__DAILY_THRESHOLD_EXCEEDED
    - type: SMS__UNVERIFIED_OTPS__BY_IP__HOURLY_THRESHOLD_EXCEEDED
  decision:
    always_allow: {}
    action: record_only
```

`fraud_protection.sms.unverified_otp_budget` (object) configures the adaptive budget for SMS unverified OTP warnings.

- `daily_ratio` (number, optional): A decimal ratio applied to verified SMS OTP counts when computing daily thresholds. `0.3` means 30%. Default `0.3`.
- `hourly_ratio` (number, optional): A decimal ratio applied to verified SMS OTP counts when computing hourly thresholds. `0.2` means 20%. Default `0.2`.
- `by_phone_country` (array of objects, optional): Per phone-country overrides. Each item contains:
  - `country_codes` (array of strings): ISO 3166-1 alpha-2 country codes of the recipient phone number countries that share the same ratios.
  - `daily_ratio` (number, optional): Overrides the global `daily_ratio` for this phone country.
  - `hourly_ratio` (number, optional): Overrides the global `hourly_ratio` for this phone country.
- Optional fields should be treated as nullable in the schema so explicit `null` is accepted the same as omission.

The `by_phone_country` list is ordered. For a given phone country, use the first override item whose `country_codes` list contains that country code. Each `country_codes` list must not contain duplicates.

The effective ratios are resolved as follows:

1. For `SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__DAILY_THRESHOLD_EXCEEDED`, use the matching override's `daily_ratio` if present; otherwise use the global `daily_ratio`.
2. For `SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__HOURLY_THRESHOLD_EXCEEDED`, use the matching override's `hourly_ratio` if present; otherwise use the global `hourly_ratio`.
3. For `SMS__UNVERIFIED_OTPS__BY_IP__DAILY_THRESHOLD_EXCEEDED` and `SMS__UNVERIFIED_OTPS__BY_IP__HOURLY_THRESHOLD_EXCEEDED`, always use the global ratios. Phone-country overrides do not apply to IP-based metrics.

**authgear.features.yaml**

Fraud protection configurability is controlled by a feature flag. By default, fraud protection settings are **not modifiable** by project admins. To allow customization, set:

```yaml
fraud_protection:
  is_modifiable: true
```

When `is_modifiable` is `false` (default), projects must use the default configuration above and cannot customize warnings or decisions.

### Warnings

We have identified two primary patterns of SMS pumping attacks:

1. **Non-rotating IP attacks**: Attackers send a large number of OTP requests from the same IP address to different phone numbers across multiple countries. These are effectively blocked by IP-based metrics:
   - `SMS__PHONE_COUNTRIES__BY_IP__DAILY_THRESHOLD_EXCEEDED` - Detects when many countries are targeted from a single IP
   - `SMS__UNVERIFIED_OTPS__BY_IP__DAILY_THRESHOLD_EXCEEDED` and `SMS__UNVERIFIED_OTPS__BY_IP__HOURLY_THRESHOLD_EXCEEDED` - Detects when many unverified OTPs are requested from a single IP

2. **Rotating IP attacks**: Attackers change their IP address frequently to evade IP-based detection while targeting the same phone number. These are effectively blocked by phone-number-country based metrics:
   - `SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__DAILY_THRESHOLD_EXCEEDED` and `SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__HOURLY_THRESHOLD_EXCEEDED` - Detects when many unverified OTPs are requested for phone numbers in a specific country, regardless of IP

#### Naming Convention

Warning names follow the pattern: `{SERVICE}__{METRIC}__BY_{DIMENSION}__{TIME_PERIOD}_THRESHOLD_EXCEEDED`

- `{SERVICE}`: The service type (e.g., `SMS`, `EMAIL`)
- `{METRIC}`: What is being measured (e.g., `PHONE_COUNTRIES`, `UNVERIFIED_OTPS`)
- `{DIMENSION}`: The scope for counting (e.g., `IP`, `PHONE_COUNTRY`)
- `{TIME_PERIOD}`: The time window (e.g., `DAILY`, `HOURLY`)

The double underscore (`__`) separates logical sections for improved readability.

Examples:
- `SMS__PHONE_COUNTRIES__BY_IP__DAILY_THRESHOLD_EXCEEDED` - Phone countries detected per IP per day
- `SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__HOURLY_THRESHOLD_EXCEEDED` - Unverified OTPs detected per phone country per hour

#### SMS__PHONE_COUNTRIES__BY_IP__DAILY_THRESHOLD_EXCEEDED
Check if the number of distinct countries of requested phone numbers from a single IP exceeds the threshold in 24 hours.

Only countries with no verified SMS OTP from the same IP in the same 24h window are counted. If an IP has at least one verified SMS OTP for a country during that window, that country is excluded from the distinct-country count.

The threshold is 3.


#### SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__DAILY_THRESHOLD_EXCEEDED
Check if the number of unverified OTPs for a specific phone number country exceeds the daily threshold in 24 hours.

Let `daily_ratio` be the effective ratio for the recipient phone number country.

```
threshold = max(
  20,                                                    # (1)
  verified_otps_to_country_past_14_days_rolling_max * daily_ratio,  # (2)
  verified_otps_to_country_past_24h * daily_ratio               # (3)
)
```

- **(1) Constant lower bound**: Allows a minimum of 20 unverified OTPs regardless of history. This handles the initial launch period when there is no verified OTP data yet.
- **(2) 14-day rolling max × daily ratio**: Provides a stable baseline quota derived from historical traffic. Using the rolling max (rather than average) ensures the threshold does not drop too aggressively after a high-traffic day.
- **(3) Past 24h verified × daily ratio**: Adapts to sudden traffic spikes. Using the same ratio as the historical baseline ensures factor (3) only becomes the binding factor on true spike days, when today's verified OTP volume significantly exceeds the 14-day max.


#### SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__HOURLY_THRESHOLD_EXCEEDED
Check if the number of unverified OTPs for a specific phone number country exceeds the hourly threshold.

Let `hourly_ratio` be the effective ratio for the recipient phone number country.

```
threshold = max(
  3,                                                     # (1) lower bound
  verified_otps_to_country_past_14_days_rolling_max / 6 * hourly_ratio,  # (2)
  verified_otps_to_country_past_1h * hourly_ratio,      # (3)
)
```

- **(1) Constant lower bound**: Allows a minimum of 3 unverified OTPs per hour regardless of history.
- **(2) 14-day rolling max / 6 × hourly ratio**: Provides a stable hourly baseline derived directly from historical daily traffic, without coupling the hourly threshold to the daily threshold.
- **(3) Past 1h verified × hourly ratio**: Handles traffic concentrated within a single hour (e.g., initial launch). Without this, a burst of legitimate traffic in one hour could still exceed the historical hourly baseline even though the traffic is legitimate.

The simulation script [`fraud-protection-simulate.py`](./fraud-protection-simulate.py) demonstrates the formula behavior with mock data using the default global ratios (`0.3` for daily and `0.2` for hourly) and no phone-country overrides. Summary:

### ~1k SMS/day

| Scenario                                                         | Daily threshold | Daily     | Hourly threshold | Hourly    |
| ---------------------------------------------------------------- | --------------- | --------- | ---------------- | --------- |
| Initial launch (no historical data, ~300 verified in first hour) | 60              | ok        | 60               | ok        |
| Normal traffic (~1k/day, peak hour ~200)                         | 200             | ok        | 40               | ok        |
| Spike day (~2x normal = 2k/day, peak hour ~400)                  | 400             | ok        | 80               | ok        |
| Attack: quiet day (~1/2 normal = 500/day)                        | 200             | TRIGGERED | 33               | TRIGGERED |
| Attack: during spike (~2x normal = 2k/day)                       | 400             | TRIGGERED | 80               | TRIGGERED |

### Low traffic country (<20 SMS/day)

| Scenario                                                                      | Daily threshold | Daily     | Hourly threshold | Hourly    |
| ----------------------------------------------------------------------------- | --------------- | --------- | ---------------- | --------- |
| [Low traffic] Initial launch (no historical data, ~10 verified in first hour) | 20              | ok        | 3                | ok        |
| [Low traffic] Normal traffic (~15/day, peak hour ~5)                          | 20              | ok        | 3                | ok        |
| [Low traffic] Spike day (~2x normal = 30/day, peak hour ~10)                  | 20              | ok        | 3                | ok        |
| [Low traffic] Attack: quiet day (~1/2 normal = 7/day)                         | 20              | TRIGGERED | 3                | TRIGGERED |
| [Low traffic] Attack: during spike (~2x normal = 30/day)                      | 20              | TRIGGERED | 3                | TRIGGERED |


#### SMS__UNVERIFIED_OTPS__BY_IP__DAILY_THRESHOLD_EXCEEDED
Check if the number of unverified OTPs from a single IP exceeds the threshold in the past 24 hours.

Let `daily_ratio` be the global daily ratio.

```
threshold = max(10, daily_ratio * verified OTPs in the past 24 hours)
```


#### SMS__UNVERIFIED_OTPS__BY_IP__HOURLY_THRESHOLD_EXCEEDED
Check if the number of unverified OTPs from a single IP exceeds the hourly threshold.

Let `hourly_ratio` be the global hourly ratio.

```
threshold = max(5, hourly_ratio * verified OTPs in the past 24 hours / 6)
```


#### Notes

Q: Why do not support IP range?

A: By observation, if the attacker is capable to switch IP during an attack, usually it is difficult to define a meaningful attempt threshold for a ip range to block. If the attacker does not switch IP address, then per IP metrics can be used.

Q: How are unverified OTP counts calculated?

A: Unverified OTP counts include OTPs that were sent but not verified by the user. However, OTPs sent during a login, signup, or forgot password flow are excluded from the count if the flow was completed successfully using an alternative authentication method (e.g., passkey, password). This prevents legitimate flows where the user chose a different method from being counted as SMS pumping attempts.

### Decision Record

Each sms send request (No matter success or not) will produce a decision record.

```jsonc
{
  "timestamp": "2026-02-05T11:11:11.025Z",
  "decision": "blocked",
  "action": "send_sms",
  "action_detail": {
    "recipient": "+12341234",
    "type": "verification",
  },
  "triggered_warnings": [
    "SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__DAILY_THRESHOLD_EXCEEDED",
    "SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__HOURLY_THRESHOLD_EXCEEDED",
    "SMS__UNVERIFIED_OTPS__BY_IP__DAILY_THRESHOLD_EXCEEDED",
    "SMS__UNVERIFIED_OTPS__BY_IP__HOURLY_THRESHOLD_EXCEEDED",
  ],
  "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X)",
  "ip_address": "203.0.113.42",
  "http_url": "https://example.authgear-apps.com/",
  "http_referer": "https://example.authgear-apps.com/login",
  "user_id": "97a0c0bb-6662-4905-9d12-e0a3ac3033d9",
  "geo_location_code": "US"
}
```

And audit log:

```jsonc
{
  "context": {
    "app_id": "example",
    "audit_context": {
      "http_url": "https://example.authgear-apps.com/",
    },
    "client_id": "tester",
    "geo_location_code": "HK",
    "ip_address": "123.123.123.123",
    "language": "en",
    "oauth": {
      "state": "xxxxx",
    },
    "preferred_languages": ["en-GB", "en-US", "en"],
    "timestamp": 1767869403,
    "triggered_by": "user",
    "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
    "user_id": "97a0c0bb-6662-4905-9d12-e0a3ac3033d9",
  },
  "id": "00000000006be272",
  "payload": {
    "record": {
      "timestamp": "2026-02-05T11:11:11.025Z",
      "decision": "blocked",
      "triggered_warnings": [
        "SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__DAILY_THRESHOLD_EXCEEDED",
        "SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__HOURLY_THRESHOLD_EXCEEDED",
        "SMS__UNVERIFIED_OTPS__BY_IP__DAILY_THRESHOLD_EXCEEDED",
        "SMS__UNVERIFIED_OTPS__BY_IP__HOURLY_THRESHOLD_EXCEEDED",
      ],
      "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X)",
      "ip_address": "203.0.113.42",
      "geo_location_code": "US",
    },
  },
  "seq": 7070322,
  "type": "fraud_protection.decision_recorded",
}
```

### API Error

When `action: deny_if_any_warning` and a warning is triggered, an API error will be returned.

```json
{
  "name": "TooManyRequest",
  "reason": "BlockedByFraudProtection",
  "code": 429
}
```

When `action: record_only`, warnings are logged but the request is always allowed.

### Examples

#### 1. Turn on all warnings, deny if any warning triggered

```yaml
fraud_protection:
  enabled: true
  sms:
    unverified_otp_budget:
      daily_ratio: 0.3
      hourly_ratio: 0.2
  warnings:
    - type: SMS__PHONE_COUNTRIES__BY_IP__DAILY_THRESHOLD_EXCEEDED
    - type: SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__DAILY_THRESHOLD_EXCEEDED
    - type: SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__HOURLY_THRESHOLD_EXCEEDED
    - type: SMS__UNVERIFIED_OTPS__BY_IP__DAILY_THRESHOLD_EXCEEDED
    - type: SMS__UNVERIFIED_OTPS__BY_IP__HOURLY_THRESHOLD_EXCEEDED
  decision:
    action: deny_if_any_warning
```

#### 2. Record only (Diagnose mode). No denial

Triggered warnings will be recorded in logs but requests will always be allowed.

```yaml
fraud_protection:
  enabled: true
  sms:
    unverified_otp_budget:
      daily_ratio: 0.3
      hourly_ratio: 0.2
  warnings:
    - type: SMS__PHONE_COUNTRIES__BY_IP__DAILY_THRESHOLD_EXCEEDED
    - type: SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__DAILY_THRESHOLD_EXCEEDED
    - type: SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__HOURLY_THRESHOLD_EXCEEDED
    - type: SMS__UNVERIFIED_OTPS__BY_IP__DAILY_THRESHOLD_EXCEEDED
    - type: SMS__UNVERIFIED_OTPS__BY_IP__HOURLY_THRESHOLD_EXCEEDED
  decision:
    action: record_only
```

### Future Work

#### Country Based Risk Classification

This feature is not introduced for now for simplicity. In the future, we may introduce a 3-level country risk classification (High, Mid, Low) to apply different thresholds per country based on their historical association with SMS pumping. This would allow the ratios to vary by risk level rather than using a single universal ratio. The explicit overrides in `fraud_protection.sms.unverified_otp_budget.by_phone_country` are the manual version of this idea.

The classification could be configured via project config:

```yaml
fraud_protection:
  geo_location_risks:
    high:
      - EG
      - UA
    low:
      - HK
```

#### Decision: Challenge

 ```yaml
 fraud_protection:
   enabled: true
   warnings:
     - type: #...
   decisions:
     challenge:
       challenge_mode: bot_protection # or email_verification
 ```

#### Warning: Custom

```yaml
fraud_protection:
  enabled: true
  warnings:
    - type: CUSTOM
      id: MY_CUSTOM_WARNING
      hook:
        url: authgeardeno:///deno/script.ts
```
