# Fraud Protection

- [SMS Pumping](#sms-pumping)
  - [Config](#config)
  - [Warnings](#warnings)
    - [SMS_MANY_PHONE_NUMBER_COUNTRIES_PER_IP_PER_DAY](#sms_many_phone_number_countries_per_ip_per_day)
    - [SMS_MANY_UNVERIFIED_OTPS_PER_PHONE_NUMBER_COUNTRY_PER_DAY](#sms_many_unverified_otps_per_phone_number_country_per_day)
    - [SMS_MANY_UNVERIFIED_OTPS_PER_PHONE_NUMBER_COUNTRY_PER_HOUR](#sms_many_unverified_otps_per_phone_number_country_per_hour)
    - [SMS_MANY_UNVERIFIED_OTPS_PER_IP_PER_DAY](#sms_many_unverified_otps_per_ip_per_day)
    - [SMS_MANY_UNVERIFIED_OTPS_PER_IP_PER_HOUR](#sms_many_unverified_otps_per_ip_per_hour)
    - [Notes](#notes)
  - [Country Based Risk Classification](#country-based-risk-classification)
  - [Environment Variables](#environment-variables)
  - [Decision Record](#decision-record)
  - [API Error](#api-error)
  - [Examples](#examples)
- [Future Work](#future-work)
  - [Decision: Challenge](#decision-challenge)
  - [Warning: Custom](#warning-custom)

## SMS Pumping

### Config

**authgear.yaml**

An example:

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
    - type: SMS_MANY_PHONE_NUMBER_COUNTRIES_PER_IP_PER_DAY
    - type: SMS_MANY_UNVERIFIED_OTPS_PER_PHONE_NUMBER_COUNTRY_PER_DAY
    - type: SMS_MANY_UNVERIFIED_OTPS_PER_PHONE_NUMBER_COUNTRY_PER_HOUR
    - type: SMS_MANY_UNVERIFIED_OTPS_PER_IP_PER_DAY
    - type: SMS_MANY_UNVERIFIED_OTPS_PER_IP_PER_HOUR
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
  geo_location_risks:
    high:
      - DZ
      - AZ
      - BD
      - CU
      - IR
      - IL
      - NG
      - OM
      - PK
      - PS
      - LK
      - SY
      - TJ
      - TN
    low:
      - US
      - CA
  warnings:
    - type: SMS_MANY_PHONE_NUMBER_COUNTRIES_PER_IP_PER_DAY
    - type: SMS_MANY_UNVERIFIED_OTPS_PER_PHONE_NUMBER_COUNTRY_PER_DAY
    - type: SMS_MANY_UNVERIFIED_OTPS_PER_PHONE_NUMBER_COUNTRY_PER_HOUR
    - type: SMS_MANY_UNVERIFIED_OTPS_PER_IP_PER_DAY
    - type: SMS_MANY_UNVERIFIED_OTPS_PER_IP_PER_HOUR
  decision:
    always_allow: {}
    action: record_only
```

### Warnings

#### SMS_MANY_PHONE_NUMBER_COUNTRIES_PER_IP_PER_DAY
Check if the number of distinct countries of requested phone numbers from a single IP exceeds the threshold in 24 hours.

The threshold is 5.


#### SMS_MANY_UNVERIFIED_OTPS_PER_PHONE_NUMBER_COUNTRY_PER_DAY
Check if the number of unverified OTPs for a specific phone number country exceeds the daily threshold in 24 hours.

The threshold depends on the risk of the country.

For High risk countries:

```
threshold = max(15, 14 day rolling max of sms successfully verified to the country per day * 0.2)
```

For Low risk countries:

```
threshold = max(300, 14 day rolling max of sms successfully verified to the country per day * 1)
```

For Mid risk countries:

```
threshold = max(30, 14 day rolling max of sms successfully verified to the country per day * 0.5)
```


#### SMS_MANY_UNVERIFIED_OTPS_PER_PHONE_NUMBER_COUNTRY_PER_HOUR
Check if the number of unverified OTPs for a specific phone number country exceeds the hourly threshold.

The threshold is 1/6 of the corresponding daily threshold.

For High risk countries:

```
threshold = max(15, 14 day rolling max of sms successfully verified to the country per day * 0.2) / 6
```

For Low risk countries:

```
threshold = max(300, 14 day rolling max of sms successfully verified to the country per day * 1) / 6
```

For Mid risk countries:

```
threshold = max(30, 14 day rolling max of sms successfully verified to the country per day * 0.5) / 6
```


#### SMS_MANY_UNVERIFIED_OTPS_PER_IP_PER_DAY
Check if the number of unverified OTPs from a single IP exceeds the threshold in the past 24 hours.

```
threshold = max(20, 0.5 * verified OTPs in the past 24 hours)
```


#### SMS_MANY_UNVERIFIED_OTPS_PER_IP_PER_HOUR
Check if the number of unverified OTPs from a single IP exceeds the hourly threshold.

```
threshold = max(5, 0.5 * verified OTPs in the past 24 hours / 6)
```


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
 
### Environment Variables
 
 The default classification of countries can be overridden using the following environment variables. The value should be a comma-separated list of ISO 3166-1 alpha-2 country codes.
 
 - `FRAUD_PROTECTION_GEO_LOCATION_RISK_HIGH_DEFAULT`: Default list of High Risk countries.
 - `FRAUD_PROTECTION_GEO_LOCATION_RISK_LOW_DEFAULT`: Default list of Low Risk countries.

```shell
# Set high risk countries to Egypt and Ukraine, and low risk to Hong Kong
FRAUD_PROTECTION_GEO_LOCATION_RISK_HIGH_DEFAULT=EG,UA
FRAUD_PROTECTION_GEO_LOCATION_RISK_LOW_DEFAULT=HK
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
    "SMS_MANY_UNVERIFIED_OTPS_PER_PHONE_NUMBER_COUNTRY_PER_DAY",
    "SMS_MANY_UNVERIFIED_OTPS_PER_PHONE_NUMBER_COUNTRY_PER_HOUR",
    "SMS_MANY_UNVERIFIED_OTPS_PER_IP_PER_DAY",
    "SMS_MANY_UNVERIFIED_OTPS_PER_IP_PER_HOUR",
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
        "SMS_MANY_UNVERIFIED_OTPS_PER_PHONE_NUMBER_COUNTRY_PER_DAY",
        "SMS_MANY_UNVERIFIED_OTPS_PER_PHONE_NUMBER_COUNTRY_PER_HOUR",
        "SMS_MANY_UNVERIFIED_OTPS_PER_IP_PER_DAY",
        "SMS_MANY_UNVERIFIED_OTPS_PER_IP_PER_HOUR",
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
  "name": "Forbidden",
  "reason": "BlockedByFraudProtection",
  "code": 403
}
```

When `action: record_only`, warnings are logged but the request is always allowed.

### Examples

#### 1. Turn on all warnings, deny if any warning triggered

```yaml
fraud_protection:
  enabled: true
  warnings:
    - type: SMS_MANY_PHONE_NUMBER_COUNTRIES_PER_IP_PER_DAY
    - type: SMS_MANY_UNVERIFIED_OTPS_PER_PHONE_NUMBER_COUNTRY_PER_DAY
    - type: SMS_MANY_UNVERIFIED_OTPS_PER_PHONE_NUMBER_COUNTRY_PER_HOUR
    - type: SMS_MANY_UNVERIFIED_OTPS_PER_IP_PER_DAY
    - type: SMS_MANY_UNVERIFIED_OTPS_PER_IP_PER_HOUR
  decision:
    action: deny_if_any_warning
```

#### 2. Record only (Diagnose mode). No denial

Triggered warnings will be recorded in logs but requests will always be allowed.

```yaml
fraud_protection:
  enabled: true
  warnings:
    - type: SMS_MANY_PHONE_NUMBER_COUNTRIES_PER_IP_PER_DAY
    - type: SMS_MANY_UNVERIFIED_OTPS_PER_PHONE_NUMBER_COUNTRY_PER_DAY
    - type: SMS_MANY_UNVERIFIED_OTPS_PER_PHONE_NUMBER_COUNTRY_PER_HOUR
    - type: SMS_MANY_UNVERIFIED_OTPS_PER_IP_PER_DAY
    - type: SMS_MANY_UNVERIFIED_OTPS_PER_IP_PER_HOUR
  decision:
    action: record_only
```

### Future Work

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

