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

| Group                                     | Name                                           | Operation                               | Rate                                                 | Rationales                                                                                                            |
| ----------------------------------------- | ---------------------------------------------- | --------------------------------------- | ---------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| **authentication.general**                | `authentication.general.per_ip`                | Verify any credentials                  | 60/minute                                            | Mitigate credential brute-forcing. Per-user rate limit is not used to avoid DoS on actual user login.                 |
|                                           | `authentication.general.per_user_per_ip`       |                                         | 10/minute                                            |                                                                                                                       |
| **authentication.password**               | `authentication.password.per_ip`               | Verify Password / Additional PW         | Fallback to `authentication.general.per_ip`          |                                                                                                                       |
|                                           | `authentication.password.per_user_per_ip`      |                                         | Fallback to `authentication.general.per_user_per_ip` |                                                                                                                       |
| **authentication.oob_otp.email.trigger**  | `authentication.oob_otp.email.trigger.per_ip`  | Send Email OTP                          | Disabled                                             | Mitigate mass/targeted message spam.                                                                                  |
| **authentication.oob_otp.email.validate** | `authentication.oob_otp.email.validate.per_ip` | Verify Email OTP                        | Fallback to `authentication.general.per_ip`          | Mitigate credential brute-force                                                                                       |
|                                           |                                                |                                         | Fallback to `authentication.general.per_user_per_ip` |                                                                                                                       |
| **authentication.oob_otp.sms.trigger**    | `authentication.oob_otp.sms.trigger.per_ip`    | Send SMS OTP                            | Disabled                                             | Mitigate mass/targeted message spam.                                                                                  |
| **authentication.oob_otp.sms.validate**   | `authentication.oob_otp.sms.validate.per_ip`   | Verify SMS OTP                          | Fallback to `authentication.general.per_ip`          | Mitigate credential brute-force                                                                                       |
|                                           |                                                |                                         | Fallback to `authentication.general.per_user_per_ip` |                                                                                                                       |
| **authentication.totp**                   | `authentication.totp.per_ip`                   | Verify TOTP                             | Fallback to `authentication.general.per_ip`          |                                                                                                                       |
|                                           | `authentication.totp.per_user_per_ip`          |                                         | Fallback to `authentication.general.per_user_per_ip` |                                                                                                                       |
| **authentication.recovery_code**          | `authentication.recovery_code.per_ip`          | Verify MFA recovery code                | Fallback to `authentication.general.per_ip`          |                                                                                                                       |
|                                           | `authentication.recovery_code.per_user_per_ip` |                                         | Fallback to `authentication.general.per_user_per_ip` |                                                                                                                       |
| **authentication.device_token**           | `authentication.device_token.per_ip`           | Verify MFA device tokens                | Fallback to `authentication.general.per_ip`          |                                                                                                                       |
|                                           | `authentication.device_token.per_user_per_ip`  |                                         | Fallback to `authentication.general.per_user_per_ip` |                                                                                                                       |
| **authentication.passkey**                | `authentication.passkey.per_ip`                | Verify Passkey                          | Fallback to `authentication.general.per_ip`          | Since Authgear uses discoverable credentials, user is derived from passkey and per-user-per-IP rate limit is N/A.     |
| **authentication.siwe**                   | `authentication.siwe.per_ip`                   | SWIE nonce request                      | Fallback to `authentication.general.per_ip`          | Mitigate credential brute-force. User is not known at this point for Web3 login so only per-IP rate limit is used.    |
| **authentication.signup**                 | `authentication.signup.per_ip`                 | Signup new user                         | 10/minute                                            | Mitigate resource exhaustion by rapid registration of new user.                                                       |
| **authentication.signup_anonymous**       | `authentication.signup_anonymous.per_ip`       | Signup new anonymous user               | 60/minute                                            | A more generous limit is given for anonymous user signup, since it usually occurs on app startup of new installation. |
| **authentication.account_enumeration**    | `authentication.account_enumeration.per_ip`    | Check login ID existence                | 10/minute                                            | Mitigate account enumeration.                                                                                         |
| **Features**                              |                                                |                                         |                                                      |                                                                                                                       |
| -                                         | -                                              | Login OTP failed verify attempts        | 5 attempts (per-target)                              | Mitigate credential brute-force; revoke OTP when set limit exceeded.                                                  |
| -                                         | -                                              | Verification OTP failed verify attempts | 5 attempts (per-target)                              |                                                                                                                       |
| **verification.email.trigger**            | `verification.email.trigger.per_ip`            | Send Verification Email                 | Disabled                                             | Mitigate mass/targeted message spam.                                                                                  |
|                                           | `verification.email.trigger.per_user`          |                                         | Disabled                                             |                                                                                                                       |
| **verification.email.validate**           | `verification.email.validate.per_ip`           | Verify Verification Email               | 60/minute                                            | Mitigate credential brute-force.                                                                                      |
| **verification.sms.trigger**              | `verification.sms.trigger.per_ip`              | Send Verification SMS                   | Disabled                                             | Mitigate mass/targeted message spam.                                                                                  |
|                                           | `verification.sms.trigger.per_user`            |                                         | Disabled                                             |                                                                                                                       |
| **verification.sms.validate**             | `verification.sms.validate.per_ip`             | Verify Verification SMS                 | 60/minute                                            | Mitigate credential brute-force.                                                                                      |
| **forgot_password.email.trigger**         | `forgot_password.email.trigger.per_ip`         | Send Forgot Password Email              | Disabled                                             | Mitigate mass/targeted message spam.                                                                                  |
| **forgot_password.email.validate**        | `forgot_password.email.validate.per_ip`        | Verify Forgot Password Email            | 60/minute                                            | Mitigate credential brute-force.                                                                                      |
| **forgot_password.sms.trigger**           | `forgot_password.sms.trigger.per_ip`           | Send Forgot Password SMS                | Disabled                                             | Mitigate mass/targeted message spam.                                                                                  |
| **forgot_password.sms.validate**          | `forgot_password.sms.validate.per_ip`          | Verify Forgot Password SMS              | 60/minute                                            | Mitigate credential brute-force.                                                                                      |
|                                           |                                                | Presign upload image request            | fixed: 10/hour                                       | Configuration not needed for now.                                                                                     |

| Group               | Name                         | Operation                        | Rate       | Notes                                                                                                                |
| ------------------- | ---------------------------- | -------------------------------- | ---------- | -------------------------------------------------------------------------------------------------------------------- |
|                     |                              | Send SMS (Global) (per_ip)       | Disabled   | Disabled by default; configured using environment variables                                                          |
|                     |                              | Send SMS (Global) (per_target)   | 50/day     | Configured using environment variables                                                                               |
|                     |                              | Send Email (Global) (per_ip)     | Disabled   | Disabled by default; configured using environment variables                                                          |
|                     |                              | Send Email (Global) (per_target) | 50/day     |                                                                                                                      |
| **messaging.sms**   | `messaging.sms.per_ip`       | Send SMS (Per Tenant)            | 60/minute  | Server operator can configured hard-limit using environment variable; tenant admin may set an additional rate limit. |
|                     | `messaging.sms.per_target`   | Send SMS (Per Tenant)            | 10/hour    |                                                                                                                      |
| **messaging.email** | `messaging.email.per_ip`     | Send Email (Per Tenant)          | 200/minute |                                                                                                                      |
|                     | `messaging.email.per_target` | Send Email (Per Tenant)          | 50/day     |                                                                                                                      |

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

| Name                                                    | Fallback To                              |
| ------------------------------------------------------- | ---------------------------------------- |
| `authentication.password.per_ip`                        | `authentication.general.per_ip`          |
| `authentication.password.per_user_per_ip`               | `authentication.general.per_user_per_ip` |
| `authentication.oob_otp.email.validate.per_ip`          | `authentication.general.per_ip`          |
| `authentication.oob_otp.email.validate.per_user_per_ip` | `authentication.general.per_user_per_ip` |
| `authentication.oob_otp.sms.validate.per_ip`            | `authentication.general.per_ip`          |
| `authentication.oob_otp.sms.validate.per_user_per_ip`   | `authentication.general.per_user_per_ip` |
| `authentication.totp.per_ip`                            | `authentication.general.per_ip`          |
| `authentication.totp.per_user_per_ip`                   | `authentication.general.per_user_per_ip` |
| `authentication.recovery_code.per_ip`                   | `authentication.general.per_ip`          |
| `authentication.recovery_code.per_user_per_ip`          | `authentication.general.per_user_per_ip` |
| `authentication.device_token.per_ip`                    | `authentication.general.per_ip`          |
| `authentication.device_token.per_user_per_ip`           | `authentication.general.per_user_per_ip` |
| `authentication.passkey.per_ip`                         | `authentication.general.per_ip`          |
| `authentication.siwe.per_ip`                            | `authentication.general.per_ip`          |

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
