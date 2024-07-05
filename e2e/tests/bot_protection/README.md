# E2E Tests - Bot Protection
All e2e tests related to bot protection.

## Table of Contents
- [Context](#context)
- [Non-coverage](#non-coverage)
  * [Justification](#justification)
- [Coverage](#coverage)
- [Mocking Verification](#mocking-verification)
  * [Cloudflare Turnstile](#cloudflare-turnstile)
  * [Recaptcha V2](#recaptcha-v2)

## Context

> Bot protection is supported only in the following step types:

    identify in signup, promote, login, signup_login, and account_recovery.
    authenticate in login and reauth.

Referenced from [Bot Protection Spec](../../../docs/specs/botprobot-protection.md)

## Non-coverage
- all steps under `promote` flow
- all steps involving `passkey`

### Justification
Promote flow will need to mock oauth session for anonymous user, which is not trivial effort

Passkey will need to implement reverse of [`webauthncose.VerifySignature`](https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.10.2/protocol/webauthncose#VerifySignature), which is not trivial effort


## Coverage

- [x] signup flow
  - [x] identify
    - [x] loginid
    - [x] oauth
- [ ] promote flow
  - [ ] identify
    - [ ] loginid
    - [ ] oauth
- [x] login flow
  - [x] identify
    - [x] loginid
    - [x] oauth
    - [ ] passkey
  - [x] authenticate
    - [x] password
    - [ ] passkey
    - [x] oobotp
    - [x] totp
    - [x] recoverycode
- [x] signup_login flow
  - [x] switch flow success
  - [x] identify
    - [x] loginid
    - [x] oauth
    - [ ] passkey
- [x] reauth flow
  - [x] authentication
    - [x] password
    - [ ] passkey
    - [x] oobotp
    - [x] totp
- [x] account_recovery_flow
  - [x] identify

## Mocking Verification
For convenience, we use some magic phrases for imitating captcha verification success/failed/service unavailable

### Cloudflare Turnstile
#### Usage
Set `bot_protection.response` as magic word

| Magic Word          | Effect                                   |
|---------------------|------------------------------------------|
| pass                | Always passes                            |
| service-unavailable | Always fail with `internal-error`        |
| (Any other string)  | Always fail with `invalid-input-response`|

### Recaptcha V2
#### Usage
Set `bot_protection.response` as magic word

| Magic Word          | Effect                                   |
|---------------------|------------------------------------------|
| pass                | Always passes                            |
| (Any other string)  | Always fail with `invalid-input-response`|