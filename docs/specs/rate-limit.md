# Rate Limits

## Background

Authgear enforces various rate limits to address potential threats:

- Excessive resource consumption
- Brute-force attempts

## Algorithm

Rate limit in Authgear uses a variant of token bucket rate limiting,
configured by 2 variables:

- Period: The minimum period between operations
- Burst: Number of operations before additional operations is denied

When rate limit is requested, a token is taken from bucket. Buckets are filled
with burst tokens initially. Rate limit is exceeded if tokens are exhausted.
The bucket is re-filled fully after the period is elapsed from the time first
token is taken.

For rate limit on credential verification (e.g. verify password, verify OTP),
token would be taken from bucket only for failed attempts. Bucket would still
be checked to ensure tokens are available to be taken before performing
verification.
Therefore, verifying a correct password would never exceed the rate limit.

## Rate Limits

Rate limits are checked right-to-left, with short-circuit on failure.

Some considerations for rate limits design:

- Per-IP rate limit may need to be higher, due to shared IP across public WiFi users.
- Per-user rate limit before authentication may cause DoS on actual user.

Rate limits without default are hard-coded (non-configurable).

| Name                                    | Operation                               | per-IP                               | per-target | per-user-per-IP                      | per-user | Rationales                                                                                                            |
| --------------------------------------- | --------------------------------------- | ------------------------------------ | ---------- | ------------------------------------ | -------- | --------------------------------------------------------------------------------------------------------------------- |
| **Authentication**                      |                                         |                                      |            |                                      |          |                                                                                                                       |
| `authentication.general`                | Verify any credentials                  | 60/minute                            |            | 10/minute                            |          | Mitigate credential brute-forcing. Per-user rate limit is not used to avoid DoS on actual user login.                 |
| `authentication.password`               | Verify Password / Additional PW         | Fallback to `authentication.general` |            | Fallback to `authentication.general` |          |                                                                                                                       |
| `authentication.oob_otp.email.trigger`  | Send Email OTP                          | Disabled                             |            |                                      | Disabled | Mitigate mass/targeted message spam.                                                                                  |
| `authentication.oob_otp.email.validate` | Verify Email OTP                        | Fallback to `authentication.general` |            | Fallback to `authentication.general` |          | Mitigate credential brute-force                                                                                       |
| `authentication.oob_otp.sms.trigger`    | Send SMS OTP                            | Disabled                             |            |                                      | Disabled | Mitigate mass/targeted message spam.                                                                                  |
| `authentication.oob_otp.sms.validate`   | Verify SMS OTP                          | Fallback to `authentication.general` |            | Fallback to `authentication.general` |          | Mitigate credential brute-force                                                                                       |
| `authentication.totp`                   | Verify TOTP                             | Fallback to `authentication.general` |            | Fallback to `authentication.general` |          |                                                                                                                       |
| `authentication.recovery_code`          | Verify MFA recovery code                | Fallback to `authentication.general` |            | Fallback to `authentication.general` |          |                                                                                                                       |
| `authentication.device_token`           | Verify MFA device tokens                | Fallback to `authentication.general` |            | Fallback to `authentication.general` |          |                                                                                                                       |
| `authentication.passkey`                | Verify Passkey                          | Fallback to `authentication.general` |            |                                      |          | Since Authgear uses discoverable credentials, user is derived from passkey and per-user-per-IP rate limit is N/A.     |
| `authentication.siwe`                   | SWIE nonce request                      | Fallback to `authentication.general` |            |                                      |          | Mitigate credential brute-force. User is not known at this point for Web3 login so only per-IP rate limit is used.    |
| `authentication.signup`                 | Signup new user                         | 10/minute                            |            |                                      |          | Mitigate resource exhaustion by rapid registration of new user.                                                       |
| `authentication.signup_anonymous`       | Signup new anonymous user               | 60/minute                            |            |                                      |          | A more generous limit is given for anonymous user signup, since it usually occurs on app startup of new installation. |
| `authentication.account_enumeration`    | Check login ID existence                | 10/minute                            |            |                                      |          | Mitigate account enumeration.                                                                                         |
| **Features**                            |                                         |                                      |            |                                      |          |                                                                                                                       |
| -                                       | Login OTP failed verify attempts        |                                      | 5 attempts |                                      |          | Mitigate credential brute-force; revoke OTP when set limit exceeded.                                                  |
| -                                       | Verification OTP failed verify attempts |                                      | 5 attempts |                                      |          |                                                                                                                       |
| `verification.email.trigger`            | Send Verification Email                 | Disabled                             |            |                                      |          | Mitigate mass/targeted message spam.                                                                                  |
| `verification.email.validate`           | Verify Verification Email               | 60/minute                            |            |                                      |          | Mitigate credential brute-force.                                                                                      |
| `verification.sms.trigger`              | Send Verification SMS                   | Disabled                             |            |                                      |          | Mitigate mass/targeted message spam.                                                                                  |
| `verification.sms.validate`             | Verify Verification SMS                 | 60/minute                            |            |                                      |          | Mitigate credential brute-force.                                                                                      |
| `forgot_password.email.trigger`         | Send Forgot Password Email              | Disabled                             |            |                                      |          | Mitigate mass/targeted message spam.                                                                                  |
| `forgot_password.email.validate`        | Verify Forgot Password Email            | 60/minute                            |            |                                      |          | Mitigate credential brute-force.                                                                                      |
| `forgot_password.sms.trigger`           | Send Forgot Password SMS                | Disabled                             |            |                                      |          | Mitigate mass/targeted message spam.                                                                                  |
| `forgot_password.sms.validate`          | Verify Forgot Password SMS              | 60/minute                            |            |                                      |          | Mitigate credential brute-force.                                                                                      |
| **Misc.**                               |                                         |                                      |            |                                      |          |                                                                                                                       |
|                                         | Presign upload image request            |                                      |            | fixed: 10/hour                       |          | Configuration not needed for now.                                                                                     |

| Name              | Operation               | per-IP     | per-target |                                                                                                                      |
| ----------------- | ----------------------- | ---------- | ---------- | -------------------------------------------------------------------------------------------------------------------- |
| **Messaging**     |                         |            |            |                                                                                                                      |
| -                 | Send SMS (Global)       | disabled   | 50/day     | Configured using environment variables                                                                               |
| -                 | Send Email (Global)     | disabled   | 50/day     |                                                                                                                      |
| `messaging.sms`   | Send SMS (Per Tenant)   | 60/minute  | 10/hour    | Server operator can configured hard-limit using environment variable; tenant admin may set an additional rate limit. |
| `messaging.email` | Send Email (Per Tenant) | 200/minute | 50/day     |                                                                                                                      |

## Cooldowns

Cooldowns are special rate limits, which always allow only 1 operations in a specific inverval.

Existing cooldowns are listed below:

| Name                                            | Operation                         | per-target | Rationales                           |
| ----------------------------------------------- | --------------------------------- | ---------- | ------------------------------------ |
| `authentication.oob_otp.email.trigger.cooldown` | Send authentication OOB OTP email | 1 minute   | Mitigate mass/targeted message spam. |
| `authentication.oob_otp.sms.trigger.cooldown`   | Send authentication OOB OTP SMS   | 1 minute   | Mitigate mass/targeted message spam. |
| `verification.email.trigger.cooldown`           | Send verification email           | 1 minute   | Mitigate mass/targeted message spam. |
| `verification.sms.trigger.cooldown`             | Send verification SMS             | 1 minute   | Mitigate mass/targeted message spam. |
| `forgot_password.email.trigger.cooldown`        | Send forgot password email        | 1 minute   | Mitigate mass/targeted message spam. |
| `forgot_password.sms.trigger.cooldown`          | Send forgot password SMS          | 1 minute   | Mitigate mass/targeted message spam. |

## Fallbacks

Some rate limits uses another rate limit config as a fallback if it is not set. See the following table for the mapping.

| Name                                    | Fallback To              |
| --------------------------------------- | ------------------------ |
| `authentication.password`               | `authentication.general` |
| `authentication.oob_otp.email.validate` | `authentication.general` |
| `authentication.oob_otp.sms.validate`   | `authentication.general` |
| `authentication.totp`                   | `authentication.general` |
| `authentication.recovery_code`          | `authentication.general` |
| `authentication.device_token`           | `authentication.general` |
| `authentication.passkey`                | `authentication.general` |
| `authentication.siwe`                   | `authentication.general` |

Rate limits not mentioned in the table has no fallback.

## Future Works

- We may want to apply request-level rate limits (e.g. admin API, OIDC endpoints)
- We may want to exclude certain users (e.g. by IP) from applying rate limit.

## Configuration

In general, rate limits are configured using 3 fields:

```yaml
verification:
  rate_limits:
    # No rate limit is applied
    # validate_code_per_ip:
    #   enabled: true
    #   period: 1h  # required if enabled
    #   burst: 1    # default to 1
---
verification:
  rate_limits:
    # Turn off a default rate limit
    validate_code_per_ip:
      enabled: false
---
verification:
  rate_limits:
    # 1 validation attempt allowed per hour.
    validate_code_per_ip:
      enabled: true
      period: 1h # required
      # burst: 1  # default to 1
---
verification:
  rate_limits:
    # 5 validation attempts allowed per hour.
    validate_code_per_ip:
      enabled: true
      period: 1h # required
      burst: 5 # default to 1
```

The available rate limits can be configured as follow:

```yaml
authentication:
  rate_limits:
    general:
      per_ip:
        enabled: true
        period: 1m
        burst: 60
      per_user_per_ip:
        enabled: true
        period: 1m
        burst: 10
    password:
      per_ip: # default disabled
      per_user_per_ip: # default disabled
    oob_otp:
      email:
        trigger_per_ip: # default disabled
        trigger_per_user: # default disabled
        trigger_cooldown: 1m
        max_failed_attempts_revoke_otp: # 5 # default disabled
        validate_per_ip: # default disabled
        validate_per_user_per_ip: # default disabled
      sms:
        trigger_per_ip: # default disabled
        trigger_per_user: # default disabled
        trigger_cooldown: 1m
        max_failed_attempts_revoke_otp: # 5 # default disabled
        validate_per_ip: # default disabled
        validate_per_user_per_ip: # default disabled
      whatsapp: # TBC?
    totp:
      per_ip: # default disabled
      per_user_per_ip: # default disabled
    passkey:
      per_ip: # default disabled
    siwe:
      per_ip: # default disabled
    recovery_code:
      per_ip: # default disabled
      per_user_per_ip: # default disabled
    device_token:
      per_ip: # default disabled
      per_user_per_ip: # default disabled
    signup:
      per_ip:
        enabled: true
        period: 1m
        burst: 10
    signup_anonymous:
      per_ip:
        enabled: true
        period: 1m
        burst: 60
    account_enumeration:
      per_ip:
        enabled: true
        period: 1m
        burst: 10

authenticator:
  oob_otp:
    sms:
      code_valid_period: 20m
    email:
      code_valid_period: 20m

forgot_password:
  code_valid_period: 20m
  rate_limits:
    email:
      trigger_per_ip: # default disabled
      trigger_cooldown: 1m
      validate_per_ip:
        enabled: true
        period: 1m
        burst: 60
    sms:
      trigger_per_ip: # default disabled
      trigger_cooldown: 1m
      validate_per_ip:
        enabled: true
        period: 1m
        burst: 60

verification:
  code_valid_period: 1h
  rate_limits:
    email:
      trigger_per_ip: # default disabled
      trigger_per_user: # default disabled
      trigger_cooldown: 1m
      max_failed_attempts_revoke_otp: # 5 # default disabled
      validate_per_ip:
        enabled: true
        period: 1m
        burst: 60
    sms:
      trigger_per_ip: # default disabled
      trigger_per_user: # default disabled
      trigger_cooldown: 1m
      max_failed_attempts_revoke_otp: # 5 # default disabled
      validate_per_ip:
        enabled: true
        period: 1m
        burst: 60

messaging:
  rate_limits:
    sms: # disabled
    sms_per_ip:
      enabled: true
      period: 1m
      burst: 60
    sms_per_target:
      enabled: true
      period: 1h
      burst: 10
    email: # disabled
    email_per_ip:
      enabled: true
      period: 1m
      burst: 60
    email_per_target:
      enabled: true
      period: 1h
      burst: 10
```

## Audit Log

There will an audit log produced whenever any request blocked by a rate limit.

See [events](./event.md#rate_limitblocked) for details.
