# Fraud Protection

- [SMS Pumping](#sms-pumping)
  - [Config](#config)
  - [Warnings](#warnings)
    - [SMS_MANY_PHONE_NUMBER_COUNTRIES_PER_IP](#sms_many_phone_number_countries_per_ip)
    - [SMS_MANY_FAILURES_PER_PHONE_NUMBER_COUNTRY](#sms_many_failures_per_phone_number_country)
    - [SMS_MANY_ATTEMPTS_PER_PHONE_NUMBER_COUNTRY](#sms_many_attempts_per_phone_number_country)
    - [SMS_MANY_UNVERIFIED_OTPS_PER_PHONE_NUMBER_COUNTRY](#sms_many_unverified_otps_per_phone_number_country)
    - [SMS_MANY_UNVERIFIED_OTPS_PER_IP](#sms_many_unverified_otps_per_ip)
    - [SMS_UNMATCHED_PHONE_NUMBER_COUNTRIES_IP_GEO_LOCATION](#sms_unmatched_phone_number_countries_ip_geo_location)
    - [Notes](#notes)
  - [Country Based Risk Classification](#country-based-risk-classification)
  - [Decision Record](#decision-record)
  - [API Error](#api-error)
  - [Future Work](#future-work)
    - [Decision: Challenge](#decision-challenge)
    - [Warning: Custom](#warning-custom)
    - [Risk Scoring](#risk-scoring)

## SMS Pumping

### Config

**authgear.yaml**

```yaml
fraud_protection:
  enabled: true
  geo_location_risks:
    high:
      - EG
      - UA
    low:
      - HK
  warnings:
    - type: SMS_MANY_PHONE_NUMBER_COUNTRIES_PER_IP
      enabled: true
    - type: SMS_MANY_FAILURES_PER_PHONE_NUMBER_COUNTRY
      enabled: true
    - type: SMS_MANY_ATTEMPTS_PER_PHONE_NUMBER_COUNTRY
      enabled: true
    - type: SMS_MANY_UNVERIFIED_OTPS_PER_PHONE_NUMBER_COUNTRY
      enabled: true
    - type: SMS_MANY_UNVERIFIED_OTPS_PER_IP
      enabled: true
    - type: SMS_UNMATCHED_PHONE_NUMBER_COUNTRIES_IP_GEO_LOCATION
      enabled: true
  decisions:
    # Decisions are evaluated in order. The first decision that matches will be executed, and further decisions will be ignored.
    # If no decisions match, the default behavior is to allow the request.
    - decision: allow
      name: always allow major business location
      allow_when_matches:
        ip_address:
          cidrs: ["123.123.1.1/32"]
          geo_location_codes: ["HK"]
        phone_number:
          geo_location_codes: ["HK", "US"]
          regex: ["^\\+852\\d*$"]
    - decision: block
      name: block if triggered 3 warnings
      block_mode: error
      block_with_warnings:
        - "*"
      block_with_thresholds:
        risk_score: 3
    - decision: block
      name: block if number of unverified otp is high
      block_mode: silent
      block_with_warnings:
        - SMS_MANY_UNVERIFIED_OTPS_PER_COUNTRY
        - SMS_MANY_UNVERIFIED_OTPS_PER_IP
      block_with_thresholds:
        risk_score: 1
```

### Warnings

#### SMS_MANY_PHONE_NUMBER_COUNTRIES_PER_IP
Check if the number of distinct countries of requested phone numbers from a single IP exceeds the threshold in 24 hours.

The threshold is 5.

`enabled`: boolean. Whether this warning is enabled.

#### SMS_MANY_FAILURES_PER_PHONE_NUMBER_COUNTRY
Check if the number of SMS delivery failures for a specific phone number country exceeds the threshold in 24 hours.

The threshold is 50.

`enabled`: boolean. Whether this warning is enabled.

#### SMS_MANY_ATTEMPTS_PER_PHONE_NUMBER_COUNTRY
Check if the total number of SMS requested for a specific phone number country exceeds the threshold in 24 hours.

The threshold depends on the risk of the country.

For High risk countries:

```
threshold = 7 day rolling mean of sms successfully sent to the country per day
```

For Low risk countries:

```
threshold = infinity
```

For Mid risk countries:

```
threshold = 7 day rolling mean of sms successfully sent to the country per day * 2
```

`enabled`: boolean. Whether this warning is enabled.

#### SMS_MANY_UNVERIFIED_OTPS_PER_PHONE_NUMBER_COUNTRY
Check if the number of unverified OTPs for a specific phone number country exceeds the threshold in 24 hours.

The threshold depends on the risk of the country.

For High risk countries:

```
threshold = 7 day rolling max of sms successfully verified to the country per day
```

For Low risk countries:

```
threshold = 7 day rolling max of sms successfully verified to the country per day * 3
```

For Mid risk countries:

```
threshold = 7 day rolling max of sms successfully verified to the country per day * 2
```

`enabled`: boolean. Whether this warning is enabled.

#### SMS_MANY_UNVERIFIED_OTPS_PER_IP
Check if the number of unverified OTPs from a single IP exceeds the threshold.

The threshold is 10.

`enabled`: boolean. Whether this warning is enabled.

#### SMS_UNMATCHED_PHONE_NUMBER_COUNTRIES_IP_GEO_LOCATION
Check if the country of the requested phone number matches the geo-location of the IP address.

`enabled`: boolean. Whether this warning is enabled.

#### Notes

Q: Why do not support IP range?

A: By observation, if the attacker is capable to switch IP during an attack, usually it is difficult to define a meaningful attempt threshold for a ip range to block. If the attacker does not switch IP address, then per IP metrics can be used.

### Country Based Risk Classification

We define 3 level of risk. High, Mid, Low.

By default, we classify countries as follows:

High Risk:

    - DZ # Algeria
    - AZ # Azerbaijan
    - BD # Bangladesh
    - CU # Cuba
    - IR # Iran
    - IL # Israel
    - NG # Nigeria
    - OM # Oman
    - PK # Pakistan
    - PS # Palestinian Territory
    - LK # Sri Lanka
    - SY # Syria
    - TJ # Tajikistan
    - TN # Tunisia

Low Risk:

    - US
    - CA

Mid Risk:

    - All remaining countries not listed as Low or High risk.

This can be configured in project config:

```yaml
fraud_protection:
  geo_location_risks:
    high:
      - EG
      - UA
    low:
      - HK
```

### Decision Record

Each sms send request (No matter success or not) will produce a decision record.

```jsonc
{
  "timestamp": "2026-02-05T11:11:11.025Z",
  "decision": "blocked",
  "block_mode": "error",
  "action": "send_sms",
  "action_detail": {
    "recipient": "+12341234",
    "type": "verification",
  },
  "triggered_warnings": [
    "SMS_MANY_UNVERIFIED_OTPS_PER_PHONE_NUMBER_COUNTRY",
    "SMS_MANY_UNVERIFIED_OTPS_PER_IP",
    "SMS_MANY_ATTEMPTS_PER_PHONE_NUMBER_COUNTRY",
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
      "block_mode": "error",
      "triggered_warnings": [
        "SMS_MANY_UNVERIFIED_OTPS_PER_PHONE_NUMBER_COUNTRY",
        "SMS_MANY_UNVERIFIED_OTPS_PER_IP",
        "SMS_MANY_ATTEMPTS_PER_PHONE_NUMBER_COUNTRY",
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

If `block_mode` is `error`, API error will be returned.

```json
{
  "name": "Forbidden",
  "reason": "BlockedByFraudProtection",
  "code": 403
}
```

If `block_mode` is `silent`, the API will pretends sms has been sent without returning error.

### Future Work

#### Decision: Challenge

```yaml
fraud_protection:
  enabled: true
  warnings:
    - type: #...
  decisions:
    - decision: challenge
      name: challenge if triggered 1 warnings
      warnings:
        - "*"
      thresholds:
        risk_score: 3
      challenge:
        bot_protection: # ...
        email_verification: #...
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

#### Risk Scoring

Each warning will support a `weight` config. Default `1`.

```yaml
fraud_protection:
  warnings:
    - type: SMS_MANY_PHONE_NUMBER_COUNTRIES_PER_IP
      weight: 2
      enabled: true
    - type: SMS_MANY_FAILURES_PER_PHONE_NUMBER_COUNTRY
      weight: 3
      enabled: true
    - type: SMS_MANY_ATTEMPTS_PER_PHONE_NUMBER_COUNTRY
      enabled: true
    - type: SMS_MANY_UNVERIFIED_OTPS_PER_PHONE_NUMBER_COUNTRY
      enabled: true
    - type: SMS_MANY_UNVERIFIED_OTPS_PER_IP
      enabled: true
    - type: SMS_UNMATCHED_PHONE_NUMBER_COUNTRIES_IP_GEO_LOCATION
      enabled: true
```

Each warning has a `score` (default `1`). The `risk_score` for a decision is the sum of `score * weight` for all its corresponding triggered warnings.

We can make decisions based on the risk score:

```yaml
fraud_protection:
  decisions:
    - decision: block
      block_mode: error
      name: block if high risk score
      block_with_thresholds:
        risk_score: 10
```
