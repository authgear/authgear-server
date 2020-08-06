# User Verification

  * [Definitions](#definitions)
  * [Criteria](#criteria)
  * [Requirement](#requirement)
  * [Interaction with OOB-OTP authentication](#interaction-with-oob-otp-authentication)
  * [Code &amp; Message](#code--message)
  * [Status Flag](#status-flag)
  * [Future enhancement](#future-enhancement)

## Definitions

A user is verified if the user fulfill the condition specified by criteria.

An identity is verifiable if:
- it is an SSO identity; or
- it is a login ID identity with a configured verifiable login ID key.

An identity is verified if:
- it is an SSO identity; or
- it is a login ID identity, and the user has a matching OOB-OTP authenticator.

## Criteria

Developer can configure the criteria used to determine verification status
of a user. There are two possible criteria:

- `any`: User has at least one verifiable identity and at least one verified verifiable identity.
- `all`: User has at least one verifiable identity and all the verifiable identities are verified.

By default, criteria `any` is used.

```yaml
# Use 'all' criteria to determine user verification status
verification:
    criteria: all
```

## Requirement

Developer can configure verification requirement for specific login ID keys.
The specified login ID keys must have type `email` or `phone`.

When a login ID identity with a login ID key requiring verification is created
(e.g. during sign up), the user is required to verify the login ID using a
one-time-password sent to the login ID before proceeding. A matching OOB-OTP
authenticator would be created in the verification process.

If a login ID key has optional verification requirement, user does not need to
verify it when creating identity. Instead, user can choose to verify it in
settings page after creation.

By default, user must verify login ID key 'email' and 'phone'.

```yaml
# Require verification for login ID key 'email'
identity:
  login_id:
    keys:
    - key: email
      type: email
      verification:       # Default value if not specified; verification is required
        enabled: true
        required: true
    - key: phone
      type: phone
      verification:       # verification is optional, can be performed in settings page
        enabled: true
        required: false
    - key: username
      type: username
      verification:
        enabled: false    # verification is disabled
```

## Interaction with authenticators

Verifying an identity would create an OOB-OTP authenticator without tag
of primary/secondary authenticator, therefore it cannot be used in
authentication.

However, removing the matching identities would also remove the OOB-OTP
authenticator created by the verification process. This ensures re-verification
is required when re-adding the identity.

## Message

The OTP message can be customized in configuration and templates.

```yaml
verification:
    sms:
      message:
        sender: "+85200000000"
    email:
      message:
        sender: "no-reply@example.com"
```

## Status Flag

The verification status flag of an identity would be shown in the UI of
account settings page.

The verification status flag of a user would be available in:
- the user info model (e.g. in webhook); and
- [OIDC ID token](./oidc.md#httpsauthgearcomuseris_verified); and
- [resolved session headers](./api-resolver.md#x-authgear-user-verified).

## Future enhancement

- Manual verification
