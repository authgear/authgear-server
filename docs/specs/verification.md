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
- it is a login ID identity, and the user has a OOB-OTP authenticator bound to
  the login ID identity.

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

## Interaction with OOB-OTP authentication

Verifying a login ID identity is equivalent to adding an OOB-OTP authenticator.
Therefore:
- Removing the OOB-OTP authenticator would cause the corresponding login ID
  identity to become unverified.
- Enrolling in OOB-OTP authentication would cause the corresponding login ID
  identity to become verified.
- Verifying a login ID identity would allow it to be used in OOB-OTP
  authentication if enabled in the configuration.

Note that even if OOB-OTP authentication is not enabled, user can still perform
user verification. However, the added OOB-OTP authenticator cannot be used in
authentication unless it is enabled in the configuration.

## Code & Message

The OTP format and message has same configuration as specified by [OOB-OTP authenticator](./user-model.md#oob-otp-authenticator).

```yaml
verification:
    sms:
      code_format: numeric      # SMS OTP defaults to 'numeric' format
      message:
        sender: "+85200000000"
    email:
      code_format: complex      # Email OTP defaults to 'complex' format
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
