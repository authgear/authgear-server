# E2E Tests - Bot Protection
All e2e tests related to bot protection.

## Context

> Bot protection is supported only in the following step types:

    identify in signup, promote, login, signup_login, and account_recovery.
    create_authenticator in signup and promote.
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
  - [x] create_authenticator
    - [x] password
    - [x] oobotp
    - [x] totp
- [ ] promote flow
  - [ ] identify
    - [ ] loginid
    - [ ] oauth
  - [ ] create_authenticator (same as signupflow > create_authenticator)
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
