# Rate Limits

## Background

Authgear enforces various rate limits to address potential threats:
- Excessive resource consumption
- Brute-force attempts

## Algorithm

Rate limit in Authgear uses a variant of token bucket rate limiting,
configured by 2 variables:
- Size: Maximum amount of tokens
- Reset period: Duration to re-fill the bucket

When rate limit is requested, a token is taken from bucket; rate limit is
exceeded is tokens are exhausted. The bucket is re-filled fully after the
reset period is elapsed from the time first token is taken.

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

| Operation | global | per-tenant | per-IP | per-target | per-user-per-IP | per-user | Rationales |
| ----------|--------|------------|--------|------------|-----------------|----------|------------|
| **Authentication**
| Verify password authenticator | | | default: 60/minute | | default: 10/minute | | Mitigate credential brute-forcing. Per-user rate limit is not used to avoid DoS on actual user login.
| Verify TOTP authenticator | | | default: 60/minute | | default: 10/minute
| Verify OOB-OTP authenticator (email/SMS) | | | default: 60/minute | | default: 10/minute
| Verify MFA recovery code | | | default: 60/minute | | default: 10/minute
| Verify MFA device token | | | default: 60/minute | | default: 10/minute
| Verify Passkey authenticator | | | default: 60/minute | | | | Since Authgear uses discoverable credentials, user is derived from passkey and per-user-per-IP rate limit is N/A.
| SWIE nonce request | | | default: 60/minute | | | | Mitigate credential brute-force. User is not known at this point for Web3 login so only per-IP rate limit is used.
| Signup new user | | | default: 10/minute | | | | Mitigate resource exhaustion by rapid registration of new user.
| Signup new anonymous user | | | default: 60/minute | | | | A more generous limit is given for anonymous user signup, since it usually occurs on app startup of new installation.
| Account enumeration (check login ID existence) | | | default: 10/minute | | | | Mitigate account enumeration.
| **Features** 
| Send OOB-OTP message | | | default: 10/hour; disabled | default: 1 minute cooldown | | | Mitigate mass/targeted message spam.
| Send forgot password message | | | default: 10/hour; disabled | default: 1 minute cooldown
| Send verification message | | | default: 10/hour; disabled | default: 1 minute cooldown | | default: 10/hour; disabled | Per-IP rate limit is used for verification during signup, per-user rate limit is used otherwise.
| OOB-OTP failed verify attempts | | | | default: 5 attempts; disabled | | | Mitigate credential brute-force; revoke OTP when set limit exceeded.
| Verification OTP failed verify attempts | | | | default: 5 attempts; disabled
| Verify verification OTP | | | default: 60/minute | | | | Mitigate credential brute-force.
| Verify forgot password OTP | | | default: 60/minute | | |
| **Messaging**
| Send SMS | environment variables | feature config | default: 60/minute; disabled | default: 50/day; disabled | | | If tenant would like rate limits for all SMS/email, contact server operator to adjust feature config.
| Send Email | environment variables | feature config | default: 60/minute; disabled | default: 50/day; disabled | |
| **Misc.** 
| Presign upload image request | | | | | fixed: 10/hour | | Configuration not needed for now.

## Future Works

- We may want to apply request-level rate limits (e.g. admin API, OIDC endpoints)
- We may want to exclude certain users (e.g. by IP) from applying rate limit.