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


| Operation                               | per-IP    | per-target        | per-user-per-IP | per-user | Rationales                                                                                                            |
| --------------------------------------- | --------- | ----------------- | --------------- | -------- | --------------------------------------------------------------------------------------------------------------------- |
| **Authentication**                      |           |                   |                 |          |                                                                                                                       |
| Verify any credentials                  | 60/minute |                   | 10/minute       |          | Mitigate credential brute-forcing. Per-user rate limit is not used to avoid DoS on actual user login.                 |
| Verify Password / Additional PW         | NIL       |                   | NIL             |          |                                                                                                                       |
| Verify Email OTP                        | NIL       |                   | NIL             |          |                                                                                                                       |
| Verify SMS OTP                          | NIL       |                   | NIL             |          |                                                                                                                       |
| Verify TOTP                             | NIL       |                   | NIL             |          |                                                                                                                       |
| Verify MFA recovery code                | NIL       |                   | NIL             |          |                                                                                                                       |
| Verify MFA device tokens                | NIL       |                   | NIL             |          |                                                                                                                       |
| Verify Passkey                          | NIL       |                   |                 |          | Since Authgear uses discoverable credentials, user is derived from passkey and per-user-per-IP rate limit is N/A.     |
| SWIE nonce request                      | NIL       |                   |                 |          | Mitigate credential brute-force. User is not known at this point for Web3 login so only per-IP rate limit is used.    |
| Signup new user                         | 10/minute |                   |                 |          | Mitigate resource exhaustion by rapid registration of new user.                                                       |
| Signup new anonymous user               | 60/minute |                   |                 |          | A more generous limit is given for anonymous user signup, since it usually occurs on app startup of new installation. |
| Check login ID existence                | 10/minute |                   |                 |          | Mitigate account enumeration.                                                                                         |
| **Features**                            |           |                   |                 |          |                                                                                                                       |
| Send Login OTP by SMS                   | NIL       | 1 minute cooldown |                 | NIL      | Mitigate mass/targeted message spam.                                                                                  |
| Send Login OTP by Email                 | NIL       | 1 minute cooldown |                 | NIL      |                                                                                                                       |
| Send Forgot Password by SMS             | NIL       | 1 minute cooldown |                 |          |                                                                                                                       |
| Send Forgot Password by Email           | NIL       | 1 minute cooldown |                 |          |                                                                                                                       |
| Send Verification by SMS                | NIL       | 1 minute cooldown |                 | NIL      | Per-IP rate limit is used for verification during signup, per-user rate limit is used otherwise.                      |
| Send Verification by Email              | NIL       | 1 minute cooldown |                 | NIL      |                                                                                                                       | 
| Login OTP failed verify attempts        |           | 5 attempts        |                 |          | Mitigate credential brute-force; revoke OTP when set limit exceeded.                                                  |
| Verification OTP failed verify attempts |           | 5 attempts        |                 |          |                                                                                                                       |
| Verify Verification OTP                 | 60/minute |                   |                 |          | Mitigate credential brute-force.                                                                                      |
| Verify Forgot Password OTP              | 60/minute |                   |                 |          |                                                                                                                       |
| **Misc.**                               |           |                   |                 |          |                                                                                                                       |
| Presign upload image request            |           |                   | fixed: 10/hour  |          | Configuration not needed for now.                                                                                     |

| Operation               | per-IP     | per-target |                                                                                                                      |
| ----------------------- | ---------- | ---------- | -------------------------------------------------------------------------------------------------------------------- |
| **Messaging**           |            |            |                                                                                                                      |
| Send SMS (Global)       | disabled   | 50/day     | Configured using environment variables                                                                               |
| Send Email (Global)     | disabled   | 50/day     |                                                                                                                      |
| Send SMS (Per Tenant)   | 60/minute  | 10/hour    | Server operator can configured hard-limit using environment variable; tenant admin may set an additional rate limit. |
| Send Email (Per Tenant) | 200/minute | 50/day     |                                                                                                                      |

## Future Works

- We may want to apply request-level rate limits (e.g. admin API, OIDC endpoints)
- We may want to exclude certain users (e.g. by IP) from applying rate limit.


## Configuration

In general, rate limits are configured using two fields:
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
            period: 1h  # required
            # burst: 1  # default to 1
---
verification:
    rate_limits:
        # 5 validation attempts allowed per hour.
        validate_code_per_ip:
            enabled: true
            period: 1h  # required
            burst: 5    # default to 1
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
                send_message_per_ip: # default disabled
                send_message_per_user: # default disabled
                send_message_cooldown: 1m
                max_failed_attempts_revoke_otp: # 5 # default disabled
                validate_per_ip: # default disabled
                validate_per_user_per_ip: # default disabled
            sms:
                send_message_per_ip: # default disabled
                send_message_per_user: # default disabled
                send_message_cooldown: 1m
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

forgot_password:
    code_expiry: 20m
    rate_limits:
        email:
            send_message_per_ip: # default disabled
            send_message_cooldown: 1m
            validate_per_ip:
                enabled: true
                period: 1m
                burst: 60
        sms:
            send_message_per_ip: # default disabled
            send_message_cooldown: 1m
            validate_per_ip:
                enabled: true
                period: 1m
                burst: 60

verification:
    code_expiry: 1h
    rate_limits:
        email:
            send_message_per_ip: # default disabled
            send_message_per_user: # default disabled
            send_message_cooldown: 1m
            max_failed_attempts_revoke_otp: # 5 # default disabled
            validate_per_ip:
                enabled: true
                period: 1m
                burst: 60
        sms:
            send_message_per_ip: # default disabled
            send_message_per_user: # default disabled
            send_message_cooldown: 1m
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
