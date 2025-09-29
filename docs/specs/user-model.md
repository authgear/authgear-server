* [User Model](#user-model)
  * [User](#user)
  * [Identity](#identity)
    * [Identity attributes](#identity-attributes)
    * [OAuth Identity](#oauth-identity)
    * [WebAuthn Identity](#webauthn-identity)
    * [Anonymous Identity](#anonymous-identity)
      * [Anonymous Identity JWT](#anonymous-identity-jwt)
      * [Anonymous Identity JWT headers](#anonymous-identity-jwt-headers)
      * [Anonymous Identity JWT payload](#anonymous-identity-jwt-payload)
      * [Anonymous Identity Promotion](#anonymous-identity-promotion)
    * [Biometric Identity](#biometric-identity)
      * [Biometric Identity JWT](#biometric-identity-jwt)
      * [Biometric Identity JWT headers](#biometric-identity-jwt-headers)
      * [Biometric Identity JWT payload](#biometric-identity-jwt-payload)
    * [Login ID Identity](#login-id-identity)
      * [Login ID Key](#login-id-key)
      * [Login ID Type](#login-id-type)
        * [Email Login ID](#email-login-id)
          * [Validation of Email Login ID](#validation-of-email-login-id)
          * [Normalization of Email Login ID](#normalization-of-email-login-id)
          * [Unique key generation of Email Login ID](#unique-key-generation-of-email-login-id)
        * [Username Login ID](#username-login-id)
          * [Validation of Username Login ID](#validation-of-username-login-id)
          * [Normalization of Username Login ID](#normalization-of-username-login-id)
          * [Unique key generation of Username Login ID](#unique-key-generation-of-username-login-id)
        * [Phone Login ID](#phone-login-id)
          * [Validation of Phone Login ID](#validation-of-phone-login-id)
          * [Normalization of Phone Login ID](#normalization-of-phone-login-id)
          * [Unique key generation of Phone Login ID](#unique-key-generation-of-phone-login-id)
      * [Optional Login ID Key during authentication](#optional-login-id-key-during-authentication)
      * [The purpose of unique key](#the-purpose-of-unique-key)
  * [Authenticator](#authenticator)
    * [Primary Authenticator](#primary-authenticator)
    * [Secondary Authenticator](#secondary-authenticator)
    * [Authenticator Types](#authenticator-types)
      * [Password Authenticator](#password-authenticator)
      * [WebAuthn Authenticator](#webauthn-authenticator)
      * [TOTP Authenticator](#totp-authenticator)
      * [OOB\-OTP Authenticator](#oob-otp-authenticator)
        * [Login Link](#login-link)
    * [Device Token](#device-token)
    * [Recovery Code](#recovery-code)
  * [Deleting a user](#deleting-a-user)
  * [Anonymizing a user](#anonymizing-a-user)
  * [Cached data of deleted or anonymized users](#cached-data-of-deleted-or-anonymized-users)
  * [Account Status](#account-status)
    * [Disabled user](#disabled-user)
    * [Deactivated user](#deactivated-user)
    * [Scheduled account deletion or anonymization](#scheduled-account-deletion-or-anonymization)
    * [Account Status Use case 1: Join](#account-status-use-case-1-join)
    * [Account Status Use case 2: Leave](#account-status-use-case-2-leave)
    * [Account Status Use case 3: Join and Leave](#account-status-use-case-3-join-and-leave)
    * [Account Status Use case 4: Block an account for a specific amount of time](#account-status-use-case-4-block-an-account-for-a-specific-amount-of-time)
    * [Account Status use case 5: Combination of Use case 3 and Use case 4](#account-status-use-case-5-combination-of-use-case-3-and-use-case-4)
    * [Changes on Admin API after the introduction of join\_at, leave\_at, disable\_at and enable\_at](#changes-on-admin-api-after-the-introduction-of-join_at-leave_at-disable_at-and-enable_at)
    * [Changes on the models seen in Hooks after the introduction of join\_at, leave\_at, disable\_at and enable\_at](#changes-on-the-models-seen-in-hooks-after-the-introduction-of-join_at-leave_at-disable_at-and-enable_at)
    * [Changes on user import after the introduction of join\_at, leave\_at, disable\_at and enable\_at](#changes-on-user-import-after-the-introduction-of-join_at-leave_at-disable_at-and-enable_at)
    * [Sessions](#sessions)
    * [Configuration](#configuration)

Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc.go)

# User Model

## User

A user has an unqiue user ID.

A user has many identities.

A user has many authenticators.

## Identity

An identity is used to look up a user.

5 types of identity are supported.

- Login ID
- OAuth
- WebAuthn
- Anonymous
- Biometric

A user either has no anonymous identity, or have exactly one anonymous identity.
A user with anonymous identity is considered as anonymous user.

A user must have at least 1 Login ID identity or 1 OAuth identity.

### Identity attributes

See [Identity Attributes](./glossary.md#identity-attributes)

The identity attributes are used to detect duplicate identity. For example, an Email Login ID and the email claim of an OAuth Identity. This prevents duplicate user when the user forgets the original authentication method.

### OAuth Identity

OAuth identity is external identity from supported OAuth 2 IdPs. Only authorization code flow is supported. If the provider supports OIDC, OIDC is preferred over provider-specific OAuth 2 protocol.

OAuth identity does not require primary authentication.

### WebAuthn Identity

WebAuthn identity is an identity backed a WebAuthn public key credential.

WebAuthn identity uses its associated WebAuthn authenticator only.

For the details of WebAuthn, please see [./webauthn.md](./webauthn.md).

### Anonymous Identity

Anonymous identity does not require primary authentication.

Anonymous identity has the following fields:

- Public Key: It is represented as a JWK and stored in the database.
- Private Key: It is kept privately and securely in the device storage.
- Key ID: A unique random string for efficient lookup.

The key-pair of an anonymous identity is optional. The anonymous identity which created through the web SDK should not has key-pair, as there is no encrypted store for storing key-pair in web browser. That means we won't be able to re-login the same anonymous user again in the web SDK, and the anonymous user account lifetime will be the same as the logged in session.

Anonymous user creation should be rate limited.

Re-login the same anonymous user is supported in the native SDK.

From the user point of view, they do not perform any explicit authentication. Therefore

- Anonymous user cannot have secondary authenticators
- Anonymous user cannot access the settings page

Anonymous users can be used only by [first-party public clients](./oidc.md#first-party-public-clients), since it allows
the client access of user credentials.

#### Anonymous Identity JWT

The server verifies the validity of the key-pair by verifying a JWT.
A challenge is requested by the SDK on demand, it is one-time use and short-lived.
The JWT is provided in the [login_hint](./oidc.md#login_hint).

#### Anonymous Identity JWT headers

- `typ`: Must be the string `vnd.authgear.anonymous-request`.

#### Anonymous Identity JWT payload

- `challenge`: The challenge returned by the server.
- `action`: either `auth` or `promote`

#### Anonymous Identity Promotion

Anonymous user can be promoted to normal user by adding a new identity. When an anonymous user is promoted:

- A new non-anonymous identity is added.
- The anonymous identity is deleted.
- A new session is created.

The promotion flow is the same as the normal OIDC authorization code flow.

The promotion flow can be triggered by using the signed JWT or promotion code.

### Biometric Identity

Biometric identity is a asymmetric key-pair.
Biometric identity can only be added on iOS and Android,
where those platforms provide necessary API to protect key material with biometric authentication.

Biometric identity does not require primary authentication.

Biometric identity collects necessary device info so that
nice name such as "iPhone 12 Mini" can be displayed to the user.

Biometric authentication can be used only by [first-party public clients](./oidc.md#first-party-public-clients), since it allows
the client access of user credentials.

Biometric authentication must NOT involve the usage of webview, in order to provide a smooth user experience.
The setup and the authentication is implemented by `/oauth2/challenge` and `/oauth2/token`.

#### Biometric Identity JWT

The server verifies the validity of the key-pair by verifying a JWT.
A challenge is requested by the SDK on demand, it is one-time use and short-lived.

#### Biometric Identity JWT headers

- `typ`: Must be the string `vnd.authgear.biometric-request`.

#### Biometric Identity JWT payload

- `challenge`: The challenge returned by the server.
- `action`: either `authenticate` or `setup`.
- `jwk`: When action is `setup`, it is the JWK of the public key.

### Login ID Identity

A login ID has the following attributes:

- Key
- Type
- Normalized value
- Original value
- Unique key

A user can have many login IDs. For example, a user can have both an email and phone number as their login IDs.

Login ID identity requires primary authentication.
Primary password authenticator, any primary Webauthn authenticator, or matching primary OOB-OTP authenticator can be used in primary authentication.

#### Login ID Key

Login ID key is a symbolic name assigned by the developer.

#### Login ID Type

Login ID type determines the validation, normalization and unique key generation rules.

##### Email Login ID

###### Validation of Email Login ID

- [RFC5322](https://tools.ietf.org/html/rfc5322#section-3.4.1)
- Disallow `+` sign in the local part (Configurable, default OFF)
- Domain blocklist / allowlist
  - Block domains in blocklist (Configurable, default OFF, can be ON only if *Allow domains in allowlist only* is OFF)
  - Block email addresses from free email provider domains (Configurable, default OFF, can be ON only if *Block domains in blocklist* is ON)
  - Allow domains in allowlist only (Configurable, default OFF, can be ON only if *Block domains in blocklist* is OFF)
  - Domain blocklist / allowlist only affect user signup, users created from portal or via admin API are not affected
  - Developer will need to provide their blocklist / allowlist in txt file, separated by newline.

###### Normalization of Email Login ID

- Case fold the domain part
- Case fold the local part (Configurable, default ON)
- Perform NFKC on the local part
- Remove all `.` signs in the local part (Configurable, default OFF)

###### Unique key generation of Email Login ID

- Encode the domain part of normalized value to punycode (IDNA 2008)

##### Username Login ID

###### Validation of Username Login ID

- Disallow confusing homoglyphs
- Validate against PRECIS IdentifierClass profile
- Disallow builtin reserved usernames (Configurable, default ON)
- Disallow username contains developer-provided keywords (Configurable, default OFF)
  - Developer will need to provide their exclude keywords in txt file, separated by newline.
- Check ASCII Only (`a-zA-Z0-9_-.`) (Configurable, default ON)

###### Normalization of Username Login ID

- Case fold (Configurable, default ON)
- Perform NFKC

###### Unique key generation of Username Login ID

The unique key is the normalized value.

##### Phone Login ID

###### Validation of Phone Login ID

- Check E.164 format

###### Normalization of Phone Login ID

Since well-formed phone login ID is in E.164 format, the normalized value is the original value.

###### Unique key generation of Phone Login ID

The unique key is the normalized value.

#### Optional Login ID Key during authentication

The login ID provided by the user is normalized against the configured set of login ID keys. If exact one identity is found, the user is identified. Otherwise, the login ID is ambiguous. Under default configuration, Email, Phone and Username login ID are disjoint sets so no ambiguity will occur. (Email must contain `@`; Username does not contain `@` or `+`; Phone must contain `+` and does not contain `@`)

#### The purpose of unique key

If the domain part of an Email login ID is internationalized, there are 2 ways to represent the login ID, either in Unicode or punycode-encoded. To ensure the same logical Email login ID always refer to the same user, unique key is generated.

## Authenticator

There are 2 kinds of authenticators, namely primary and secondary.
An authenticator is either primary or secondary, but not both.

Authenticator can be marked as default.
The primary authenticator created in user creation will be marked as default.

Authenticators have priorities.
A default authenticator has a higher priority.
Authenticators are further ordered by the configuration.

When performing authentication, all authenticators possessed by the user can be
used, regardless of the configured authenticator types.

Whether secondary authentication is needed or not depends on the primary authenticator
being used in the authentication.
If no primary authenticator is used, secondary authentication is NOT needed.
For example, signing in with OAuth does not require secondary authentication because
no primary authenticator is used.

When an identity is removed, all matching authenticators are also removed. For
example, removing a login ID identity would also remove the OOB-OTP
authenticators using same login ID as target.

### Primary Authenticator

Primary authenticators authenticate the identity.

### Secondary Authenticator

Secondary authenticators are additional authentication methods to ensure higher degree of confidence in authenticity.

The mode of secondary authentication is configurable. They are as follows:

- `disabled`: secondary authentication is disabled.
- `required`: secondary authentication is required. Every user must have at least one secondary authenticator.
- `if_exists`: secondary authentication is opt-in. If the user has at least one secondary authenticator, then the user must perform secondary authentication.

The default mode is `if_exists`.

### Authenticator Types

#### Password Authenticator

Password authenticator is either primary or secondary.
Each user has at most 1 primary password, and at most 1 secondary password.

Primary password authenticator requires secondary authentication.

#### WebAuthn Authenticator

WebAuthn authenticator is a primary authenticator.
It is always associated with 1 WebAuthn identity.
When the associated identity is deleted, it is deleted as well.

Primary WebAuthn authenticator DOES NOT require secondary authentication.

#### TOTP Authenticator

TOTP authenticator is a secondary authenticator.

TOTP authenticator is specified in [RFC6238](https://tools.ietf.org/html/rfc6238) and [RFC4226](https://tools.ietf.org/html/rfc4226).

In order to be compatible with existing authenticator applications like Google Authenticator, the following parameters are chosen:

- The algorithm is always HMAC-SHA1.
- The code is always 6-digit long.
- The valid period of a code is always 30 seconds.

To deal with clock skew, the code generated before or after the current time are also accepted.

Users may have multiple TOTP authenticators. In this case, the inputted TOTP
would be matched against all TOTP authenticators of user. However, a limit on
the maximum amount of secondary TOTP authenticators may be set in the
configuration.

#### OOB-OTP Authenticator

Out-of-band One-time-password authenticator is either primary or secondary.

OOB-OTP authenticator is bound to a recipient address. The recipient can be an email address or phone number that can receive SMS messages.

Primary OOB-OTP authenticator requires secondary authentication.

A primary OOB-OTP authenticator is associated with a login ID identity.
When the associated identity is deleted, the authenticator is deleted as well.

The OTP is a numeric code. The number of digits can be customized in the
configuration.

```yaml
authenticator:
  oob_otp:
    sms:
      message:
        sender: "+85200000000"
    email:
      message:
        sender: "no-reply@example.com"
```

The OTP message is rendered by a [customizable template](./templates.md#otp_message).

Users may have multiple OOB-OTP authenticators. In this case, user may select
which OOB-OTP authenticator to use when performing authentication. However, a
limit on the maximum amount of secondary OOB-OTP authenticators may be set in
the configuration.

##### Login Link

Login link is a sub-category under OOB-OTP authenticator. User receives a link instead of OTP code, which authenticates user on visit.

```yaml
authentication:
  oob_otp:
    email:
      email_otp_mode: "login_link" # default "code"

verification:
  code_expiry_seconds: 3600
```

The login link message is rendered by a [customizable template](./templates.md#oob_magic_link).

Unlike other OTP modes, login link can be opened on another device. Login will proceed on original device upon approval success.

### Device Token

Device tokens are used to indicate a trusted device.

A device token is generated when user opts in during secondary authentication.
The generated device token is stored in a cookie, and it allows the user to skip
secondary authentication as long as it remains valid.

The token is a cryptographically secure random string with at least 256 bits.

### Recovery Code

Recovery codes are used to bypass secondary authentication when a secondary
authenticator is lost or unusable.

Recovery codes are generated when the user adds a secondary authenticator first
time. It can be regenerated and listed (if configured) in settings page.

Once used, a recovery code is invalidated.

The codes are cryptographically secure random 10-letter string in Crockford's
Base32 alphabet.

## Deleting a user

Deleting a user will hard-delete all data from the database,
including identities, authenticators, sessions, etc.

The developer can delete a user via the Admin API, or
the admin can delete a user on the portal.

## Anonymizing a user

Anonymizing a user is similar to deleting a user.
Except for the user record, all data including identities,
authenticators, sessions, etc will be deleted from the database.
All the standard and custom attributes of the user will be cleared.
The user record's fields `is_anonymized` and `anonymized_at` will be updated when
the user is anonymized.

The developer can anonymize a user via the Admin API, or
the admin can anonymize a user on the portal.

Once the user is anonymized, it cannot be reverted.

## Cached data of deleted or anonymized users

Some internal data may still present in cache (Redis), such as OAuth states,
MFA device tokens, rate limit counter. There data will remain in the cache
until its natural expiry.

## Account Status

This section specifies Account Status, which includes the following:

- disabled user
- deactivated user
- anonymized user
- scheduled account deletion
- scheduled account anonymization
- `join_at`
- `leave_at`
- `disable_at` and `enable_at`

The following truth table defines Account Status.

|is\_disabled|is\_deactivated|is_anonymized|delete\_at|anonymize_at|anonymized_at|join_at|leave_at|disable_at|enable_at|Account Status|
|---|---|---|---|---|---|---|---|---|---|---|
|false|false|false|null|null|null|null|null|null|null|Normal|
|false|false|false|null|null|null|non-null|any|any|any|Disabled due to join_at|
|false|false|false|null|null|null|any|non-null|any|any|Disabled due to leave_at|
|false|false|false|null|null|null|any|any|non-null|non-null|Disabled due to disable_at and enable_at|
|true|false|false|null|null|null|any|any|any|any|Disabled|
|true|true|false|null|null|null|any|any|any|any|Deactivated|
|true|false|false|non-null|null|null|any|any|any|any|Scheduled deletion by admin|
|true|true|false|non-null|null|null|any|any|any|any|Scheduled deletion by end-user|
|true|false|false|null|non-null|null|any|any|any|any|Scheduled anonymization by admin|
|true|true|false|null|non-null|null|any|any|any|any|Scheduled anonymization by end-user|
|true|any|true|null|any|non-null|any|any|any|any|Anonymized|

List of valid state transitions:

- Normal --[Disable]--> Disabled
- Normal --[Deactivate]--> Deactivated
- Normal --[Schedule deletion by admin]--> Scheduled deletion by admin
- Normal --[Schedule deletion by end-user]--> Scheduled deletion by end-user
- Disabled --[Re-enable]--> Normal
- Deactivated --[Reactivate]--> Normal
- Deactivated --[Re-enable]--> Normal
- Scheduled deletion by admin --[Unschedule deletion]--> Normal
- Scheduled deletion by end-user --[Unschedule deletion]--> Normal
- Normal --[Schedule anonymization by admin]--> Scheduled anonymization by admin
- Normal --[Schedule anonymization by end-user]--> Scheduled anonymization by end-user
- Scheduled anonymization by admin --[Unschedule anonymization]--> Normal
- Scheduled anonymization by end-user --[Unschedule anonymization]--> Normal

Rules on `join_at`, `leave_at`, `disable_at` and `enable_at`:

- When the column `is_disabled=true`, `join_at`, `leave_at`, `disable_at` and `enable_at` are ignored.
  It is legal to set `join_at`, `leave_at`, `disable_at` and `enable_at` when `is_disabled=true` via Admin API.
- `join_at` can be set independent of `leave_at`. See use case 1.
- `leave_at` can be set independent of `join_at`. See Use case 2.
- `join_at` and `leave_at` can also be set together. See Use case 3.
- `disable_at` and `enable_at` **MUST** be set together, with `disable_at < enable_at` to form a disabled period. See Use case 4.
  - You may argue that you want to disable an account starting from a specific time indefinitely. This use case is actually Use case 2.
  - You may argue that you want to enable an account starting from a specific time indefinitely. This use case is actually Use case 1.
- The 4 columns form this invariant: `join_at < disable_at < enable_at < leave_at`.

### Disabled user

A user can be disabled by admins. A disabled user cannot sign in, and appropriate
error message will be shown when login is attempted.

Admin may optionally provide a reason when disabling a user. This reason will be
shown when the user attempted to sign in.

When a disabled user attempts to sign in, the user will be informed of disabled
status only after performing the whole authentication process, including MFA if required.

### Deactivated user

The end-user can deactivate their account. A deactivated user is considered as disabled.
When a deactivated user signs in, they can reactivate their account.

> Reactivating a user is NOT yet implemented!

### Scheduled account deletion or anonymization

Instead of deleting or anonymizing a user immediately, it can be scheduled.

The schedule is measured in terms of days. The default value is 30 days. Valid values are [1, 180].

When the deletion or anonymization is scheduled via Admin API or by admin, the user is disabled.
When the deletion or anonymization is unscheduled, the user is re-enabled.

When the deletion or anonymization is scheduled by the end-user, the user is deactivated.
To cancel the schedule, the end-user has to reactivate their account.
It is possible to cancel the schedule on behalf of the end-user.
Whether the end-user can schedule deletion or anonymization on their account is configurable.

> Scheduling anonymization by the end-user is not implemented yet.

### Account Status Use case 1: Join

I want to import an account on 2025-09-29 but I want the account to be able to sign in on 2025-10-02

```
////////////
------------+-----------
  Disabled  J  Enabled

where J === join_at === 2025-10-02
```

### Account Status Use case 2: Leave

An employee resigned and his final working day is 2025-10-31. So I want the account to not be able to sign in on 2025-10-31.

```
           /////////////
-----------+------------
  Enabled  L  Disabled

where L === leave_at === 2025-10-31
```

### Account Status Use case 3: Join and Leave

I have a part-time employee whom I know when he is joining and when he is leaving. He is joining on 2025-11-01 and he is leaving on 2025-11-30.

```
////////////             /////////////
------------+------------+------------
  Disabled  J  Enabled   L  Disabled

Where J === join_at === 2025-11-01
      L === leave_at === 2025-11-30
```

### Account Status Use case 4: Block an account for a specific amount of time

An employee is on leave during 2025-12-24 through 2026-01-01 (inclusive). I do not want him to be able to sign in during the period.

```
           //////////////
-----------+-------------+-----------
  Enabled  D  Disabled   E  Enabled

Where D === disable_at === 2025-12-24
      E === enable_at === 2026-01-02
```

### Account Status use case 5: Combination of Use case 3 and Use case 4

A contract employee is joining on 2026-04-01. His finally working date is 2027-03-31. He applied an annual leave from 2026-07-15 to 2026-07-31 (inclusive).

```
////////////             /////////////             /////////////
------------+------------+------------+------------+------------
  Disabled  J  Enabled   D  Disabled  E  Enabled   L  Disabled

Where J === join_at === 2026-04-01
      D === disable_at === 2026-07-15
      E === enable_at === 2026-08-01
      L === leave_at === 2027-04-01
```

### Changes on Admin API after the introduction of `join_at`, `leave_at`, `disable_at` and `enable_at`

- Add `joinAt`, `leaveAt`, `disableAt`, and `enableAt` to `User`.
- `User.isDisabled` becomes the effective disabled status. It is a derived field now.
	- It is intuitive to read `isDisabled` to see if the account is disabled or not.
	- For existing projects that do not utilize this new feature, the behavior of `isDisabled` does not change.
	- For new projects that utilize this new feature, if the developer wants to know whether an account was disabled manually via `setDisabledStatus`, they can read `isDisabledRaw`.
- A new field, `User.isDisabledRaw`, is the column `is_disabled`.
- Add the following mutations

```
type Mutation {
  // These are existing mutations.
  // They are included here for easier comparison.
  setDisabledStatus(input: {isDisabled: Boolean!, reason: String, userID: ID!})
  scheduleAccountAnonymization(input: {userID: ID!})
  scheduleAccountDeletion(input: {userID: ID!})
  unscheduleAccountAnonymization(input: {userID: ID!})
  unscheduleAccountDeletion(input: {userID: ID!})

  // These are new mutations.

  // If joinAt is null, set join_at to null. leave_at is left unchanged.
  setJoinAt(input: {joinAt: DateTime, userID: ID!})
  // If leaveAt is null, set leave_at to null. join_at is left unchanged.
  setLeaveAt(input: {leaveAt: DateTime, userID: ID!})
  // Set join_at and leave_at at the same time.
  setJoinAtLeaveAt(input: {joinAt: DateTime, leaveAt: DateTime, userID: ID!})
  // This sets disable_at and enable_at to the given values.
  scheduleAccountDisabled(input: {disableAt: Datetime!, enableAt: DateTime!, userID: ID!})
  // This sets disable_at and enable_at to null.
  unscheduleAccountDisabled(input: {userID: ID!})
}
```

### Changes on the models seen in Hooks after the introduction of `join_at`, `leave_at`, `disable_at` and `enable_at`

- Add `join_at`, `leave_at`, `disable_at`, and `enable_at` to user.
- Add `is_disabled_raw` to user.
- `is_disabled` becomes the effective disabled status.

### Changes on user import after the introduction of `join_at`, `leave_at`, `disable_at` and `enable_at`

- `disabled` is kept unchanged. It means setting the column `is_disabled`.
- Support `join_at` and `leave_at`, which set the column `join_at` and `leave_at` respectively.
- Support `disable_at` and `enable_at`, which set the column `disable_at` and `enable_at`.
- The rules of `join_at`, `leave_at`, `disable_at` and `enable_at` are enforced.

### Sessions

When a user is disabled, deactivated or scheduled for deletion, all sessions are deleted.

### Configuration

```yaml
account_deletion:
  scheduled_by_end_user_enabled: false
  grace_period_days: 30
account_anonymization:
  # Should be added when scheduling anonymization by the end-user is supported
  # scheduled_by_end_user_enabled: false
  grace_period_days: 30
```
