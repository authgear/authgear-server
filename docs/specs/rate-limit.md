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

## Rate Limits

Rate limits are checked right-to-left, with short-circuit on failure.

Rate limits without default are hard-coded (non-configurable).

| Operation | global | per-tenant | per-IP | per-target | per-user |
| ----------|--------|------------|--------|------------|----------|
| **Authentication**
| Verify Password | | | | | default: 10/minute |
| Verify TOTP | | | | | 10/minute |
| Verify Passkey | | | | | 10/minute |
| Verify OOB-OTP (email/SMS) | | | | | 10/minute |
| Verify MFA recovery code | | | | | 10/minute |
| Verify MFA device token | | | | | 10/minute |
| SWIE nonce request | | | 10/minute | | |
| Signup new user | | | 10/minute | | |
| Signup new anonymous user | | | 60/hour | | |
| Account enumeration (check login ID existence) | | | 10/minute | | |
| **Features**
| Send OTP (OOB-OTP/verification) message (i.e. cooldown) | | | | 1/minute | |
| OTP (OOB-OTP/verification) failed attempts | | | | default: 5/20 minutes; disabled | |
| Verify verification OTP | | | 10/minute | | |
| Send forgot password message | | | | | 5/5 minutes (per login ID) |
| Verify forgot password OTP | | | 10/minute | | |
| **Messaging**
| Send SMS | | feature config | default: 120/minute; disabled | default: 10/day; disabled | |
| Send Email | | | | 10/minute | |
| **Misc.** 
| Presign upload image request | | | | | 10/hour |
