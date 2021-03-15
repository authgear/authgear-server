# User Model

  * [User](#user)
  * [Identity](#identity)
    * [Identity Claims](#identity-claims)
    * [OAuth Identity](#oauth-identity)
      * [OIDC IdPs](#oidc-idps)
      * [OAuth 2 IdPs](#oauth-2-idps)
    * [Anonymous Identity](#anonymous-identity)
      * [Anonymous Identity JWT](#anonymous-identity-jwt)
      * [Anonymous Identity JWT headers](#anonymous-identity-jwt-headers)
      * [Anonymous Identity JWT payload](#anonymous-identity-jwt-payload)
      * [Anonymous Identity Promotion](#anonymous-identity-promotion)
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
        * [Raw Login ID](#raw-login-id)
      * [Optional Login ID Key during authentication](#optional-login-id-key-during-authentication)
      * [The purpose of unique key](#the-purpose-of-unique-key)
  * [Authenticator](#authenticator)
    * [Primary Authenticator](#primary-authenticator)
    * [Secondary Authenticator](#secondary-authenticator)
    * [Authenticator Tags](#authenticator-tags)
    * [Authenticator Types](#authenticator-types)
      * [Password Authenticator](#password-authenticator)
      * [TOTP Authenticator](#totp-authenticator)
      * [OOB-OTP Authenticator](#oob-otp-authenticator)
    * [Device Token](#device-token)
    * [Recovery Code](#recovery-code)

## User

A user has many identities. A user has many authenticators.

## Identity

An identity is used to look up a user.

3 types of identity are supported.

- Login ID
- OAuth
- Anonymous

### Identity Claims

The information of an identity are mapped to [Standard Claims](https://openid.net/specs/openid-connect-core-1_0.html#StandardClaims)

Currently, only `email` is mapped.

The claims are used to detect duplicate identity. For example, an Email Login ID and the email claim of an OAuth Identity. This prevents duplicate user when the user forgets the original authentication method.

### OAuth Identity

OAuth identity is external identity from supported OAuth 2 IdPs. Only authorization code flow is supported. If the provider supports OIDC, OIDC is preferred over provider-specific OAuth 2 protocol.

#### OIDC IdPs

The following IdPs are integrated with OIDC:

- Google
- Apple
- Azure AD

#### OAuth 2 IdPs

The following IdPs does not support OIDC. The integration is provider-specific.

- LinkedIn
- Facebook

### Anonymous Identity

A user either has no anonymous identity, or have exactly one anonymous identity. A user with anonymous identity is considered as anonymous user.

Anonymous identity has the following fields:

- Public Key: It is represented as a JWK and stored in the database.
- Private Key: It is kept privately and securely in the device storage.
- Key ID: A unique random string for efficient lookup.

From the user point of view, they do not perform any explicit authentication. Therefore

- Anonymous user cannot have secondary authenticators
- Anonymous user cannot access the settings page

Anonymous users can be used only by first-party OAuth clients, since it allows
the client access of user credentials.

#### Anonymous Identity JWT

The server verifies the validity of the key-pair by verify a JWT. A challenge is requested by the client on demand, it is one-time use and short-lived. The JWT is provided in the [login_hint](./oidc.md#login_hint).

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

### Login ID Identity

A login ID has the following attributes:

- Key
- Type
- Normalized value
- Original value
- Unique key

A user can have many login IDs. For example, a user can have both an email and phone number as their login IDs.

#### Login ID Key

Login ID key is a symbolic name assigned by the developer.

#### Login ID Type

Login ID type determines the validation, normalization and unique key generation rules.

##### Email Login ID

###### Validation of Email Login ID

- [RFC5322](https://tools.ietf.org/html/rfc5322#section-3.4.1)
- Disallow `+` sign in the local part (Configurable, default OFF)
- Domain blacklist / whitelist
  - Block blacklisted domains (Configurable, default OFF, can be ON only if *Allow whitelisted domains* is OFF)
  - Block email addresses from free email provider domains (Configurable, default OFF, can be ON only if *Block blacklisted domains* is ON)
  - Allow whitelisted domains only (Configurable, default OFF, can be ON only if *Block blacklisted domains* is OFF)
  - Domain blacklist / whitelist only affect user signup, users created from portal or via admin API are not affected
  - Developer will need to provide their blacklist / whitelist in txt file, separated by newline.

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

##### Raw Login ID

Raw login ID does not any validation or normalization. The unique key is the same as the original value. Most of the use case of login ID should be covered by the above login ID types.

#### Optional Login ID Key during authentication

The login ID provided by the user is normalized against the configured set of login ID keys. If exact one identity is found, the user is identified. Otherwise, the login ID is ambiguous. Under default configuration, Email, Phone and Username login ID are disjoint sets so no ambiguity will occur. (Email must contain `@`; Username does not contain `@` or `+`; Phone must contain `+` and does not contain `@`)

#### The purpose of unique key

If the domain part of an Email login ID is internationalized, there are 2 ways to represent the login ID, either in Unicode or punycode-encoded. To ensure the same logical Email login ID always refer to the same user, unique key is generated.

## Authenticator

Authgear supports various types of authenticator. Authenticator can be primary, secondary or both.

Authenticators have priorities. The first authenticator is the default authenticator in the UI.

When performing authentication, all authenticators possessed by the user can be
used, regardless of the configured authenticator types.

When an identity is removed, all matching authenticators are also removed. For
example, removing a login ID identity would also remove the OOB-OTP 
authenticators using same login ID as target.

### Primary Authenticator

Primary authenticators authenticate the identity. Each identity has specific applicable primary authenticators. For example, OAuth Identity does not have any applicable primary authenticators.

### Secondary Authenticator

Secondary authenticators are additional authentication methods to ensure higher degree of confidence in authenticity.

### Authenticator Tags

Each authenticator may have associated tags, they are used for determining:
- whether the authenticator is primary or secondary,
  or not used in authentication.
- whether the authenticator is the default when there are multiple authenticators.

The authenticator tags are persisted along with the authenticator, so changing
the configuration would not affect the interpretation of existing authenticators.

### Authenticator Types

#### Password Authenticator

Password authenticator is a primary authenticator. Every user has at most 1 password authenticator.

#### TOTP Authenticator

TOTP authenticator is either primary or secondary.

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

An OOB-OTP authenticator may matches a login ID identity. The normalized email
address/phone number is used to match login ID identities.

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
