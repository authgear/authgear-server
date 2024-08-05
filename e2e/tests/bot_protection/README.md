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
    authenticate in login.
    <!-- TODO: Add signup verify & create_authenticator -->

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
    - [ ] oauth
- [ ] promote flow
  - [ ] identify
    - [ ] loginid
    - [ ] oauth
- [x] login flow
  - [x] identify
    - [x] loginid
    - [ ] oauth
    - [ ] passkey
  - [x] authenticate
    - [x] password
    - [ ] passkey
    - [x] oobotp
    - [ ] totp
    - [ ] recoverycode
- [x] signup_login flow
  - [x] switch flow success
  - [x] identify
    - [x] loginid
    - [ ] oauth
    - [ ] passkey
- [ ] reauth flow
  - [ ] authentication
    - [ ] password
    - [ ] passkey
    - [ ] oobotp
    - [ ] totp
- [x] account_recovery_flow
  - [x] identify
- [x] general
  - [x] should not require bot protection if previous steps already have `success` verification
  - [x] should reject bot protection provider not aligned to `authgear.yaml` `bot_protection.provider`
      For example, if `authgear.yaml` has `cloudflare` configured, but input has `recaptchav2`, should reject on json schema validation
  

## Mocking Verification
For convenience, we use some magic phrases for imitating captcha verification success/failed/service unavailable

### Cloudflare Turnstile
#### Usage
Set input `bot_protection.response` as magic word

| Magic Word          | Effect                                   |
|---------------------|------------------------------------------|
| pass                | Always passes                            |
| service-unavailable | Always fail with `internal-error`        |
| (Any other string)  | Always fail with `invalid-input-response`|

### Recaptcha V2
#### Usage
Set input `bot_protection.response` as magic word

| Magic Word          | Effect                                   |
|---------------------|------------------------------------------|
| pass                | Always passes                            |
| (Any other string)  | Always fail with `invalid-input-response`|