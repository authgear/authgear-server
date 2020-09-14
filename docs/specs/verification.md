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

A claim is verifiable if its verification is enabled in the configuration.

## Criteria

Developer can configure the criteria used to determine verification status
of a user. There are two possible criteria:

- `any`: User has at least one verifiable claim and at least one verified verifiable claim.
- `all`: User has at least one verifiable claim and all the verifiable claims are verified.

By default, criteria `any` is used.

```yaml
# Use 'all' criteria to determine user verification status
verification:
    criteria: all
```

## Requirement

Developer can configure verification requirement for 'email' and 'phone_number'
OIDC standard claims.

When an identity with a claim requiring verification is created
(e.g. during sign up), the user is required to verify the claim using a
one-time-password associated with the claim before proceeding.

If a claim has optional verification requirement, user does not need to
verify it when creating identity. Instead, user can choose to verify it in
settings page after creation.

By default, user must verify both 'email' and 'phone_number' claims when
creating new identity.

An OOB-OTP authenticator may be associated with an email/phone number. Enrolling
in OOB-OTP authentication always require verification on the email/phone number.

```yaml
verification:
  claims:
    email:  # Default value if not specified; verification is required
      enabled: true
      required: true
    phone_number:  # Verification is optional, can be performed in settings page
      enabled: true
      required: false
```

## Interaction with Identity/Authenticator

An identity can have multiple verifiable claims (e.g. phone_number and email),
all the claims would be inspected when checking verification requirement.
For example, if both phone_number and email claim requires verification, user
is required to perform verification on both claims before creating the identity.

When user adds an OOB-OTP authenticator, its associated claim is marked as
verified since the enrollment process verifies the claim.

A verified claim must match at least one identity/authenticator; it will be
removed when the last matching identity/authenticator is removed.

## Status Flag

The verification status flag of an identity would be shown in the UI of
account settings page.

The verification status flag of a user would be available in:
- the user info model (e.g. in webhook); and
- [OIDC ID token](./oidc.md#httpsauthgearcomuseris_verified); and
- [resolved session headers](./api-resolver.md#x-authgear-user-verified).

## Future enhancement

- Manual verification
