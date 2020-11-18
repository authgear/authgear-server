# User Model

## User

A user has a set of standard attributes.
The standard attributes can contribute to the computation of the claims of the user.

A user has a set of custom attributes.
The custom attributes can contribute to the computation of the claims of the user.

A user has many identities.
Identity has [Standard Claims](https://openid.net/specs/openid-connect-core-1_0.html#StandardClaims).
The claims of an identity can affect the standard attributes of the user.

A user has many authenticators.
The claims of an authenticators do **NOT** contribute to the computation of the claims of the user.

A user has a set of claims, which

- Are computed from the standard attributes and the custom attributes.
- Are the information returned in the UserInfo endpoint.
- Are included in the ID Token.

## Standard attributes

A user has the following standard attributes.

- `email`
- `phone_number`
- `preferred_username`

All of them are nullable.

The user can select their `email`, `phone_number` and `preferred_username` from the identity claims in the settings page.

The admin can do the same thing in the portal.

### Standard attributes - email

When the `email` standard attribute is null, adding or updating identity will copy the `email` claim to the `email` standard attribute.

When the user loses the last ownership of an `email` identity claim, the `email` standard attribute is set to the `email` claim of the oldest identity if present, or null if absent.

### Standard attributes - phone\_number

When the `phone_number` standard attribute is null, adding or updating identity will copy the `phone_number` claim to the `phone_number` standard attribute.

When the user loses the last ownership of an `phone_number` identity claim, the `phone_number` standard attribute is set to the `phone_number` claim of the oldest identity if present, or null if absent.

### Standard attributes - preferred\_username

When the `preferred_username` standard attribute is null, adding or updating identity will copy the `preferred_username` claim to the `preferred_username` standard attribute.

When the user loses the last ownership of an `preferred_username` identity claim, the `preferred_username` standard attribute is set to the `preferred_username` claim of the oldest identity if present, or null if absent.

> TODO: Should we promote the `email` identity claim into `preferred_username` standard attribute?

## Custom Attributes

Custom Attributes is a JSON Object. The developer can update it via Admin API.

Custom Attributes can be any valid JSON Object. There is no limitation on the nesting level. However, a sensible size limit (e.g. 10MiB) is enforced.

The developer can save whatever they want in Custom Attributes.

The user cannot read or write Custom Attributes directly.

> NOTE: The initial implementation is very likely to be a single textarea for editing the whole custom attributes object.
> Much more fancier features like generated form fields requires us to interpret the JSON schema and generate form fields.
> That could be a very useful standalone product.

### Custom Attributes validation

The validation on custom attributes can be done with the JSON schema provided by the developer in the configuration.

The JSON schema must be written against [the version 2019-09](https://json-schema.org/specification-links.html#2019-09-formerly-known-as-draft-8).

For example,

```yaml
custom_attributes:
  json_schema:
    type: object
    properties:
      email:
        type: string
        format: email
    required: ["email"]
```

Any update on custom attributes must be valid against the given schema.

> NOTE: In the future, we can add our own JSON schema keyword to annotate the access control of individual fields in the custom attributes.

## Identity

An identity is used to look up a user.

3 types of identity are supported.

- Login ID (`login_id`)
- OAuth (`oauth`)
- Anonymous (`anonymous`)

### Identity Claims

- OAuth identity has `email` claim.
- Email Login ID identity has `email` claim.
- Phone Login ID identity has `phone_number` claim.
- Username Login ID identity has `preferred_username` claim.

> TODO: In the future, we want to allow the developer to customize the scope when we perform OAuth flow with external OAuth provider.
> Then we can support more Standard Claims, such as `picture`.

The claim `email` is used to detect duplicate identity. For example, an Email Login ID and the email claim of an OAuth Identity. This prevents duplicate user when the user forgets the original authentication method.

### OAuth Identity

OAuth identity is external identity from supported OAuth 2 IdPs. Only authorization code flow is supported. If the provider supports OIDC, OIDC is preferred over provider-specific OAuth 2 protocol.

OAuth identity is updated every time authentication is performed.

> TODO: In the future, we may want to support frozen oauth identity, that is, the identity is never updated.

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
- Disallow developer-provided reserved usernames (Configurable, default empty list)
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

Authgear supports various types of authenticator. Authenticator must either be primary or secondary.

A primary or secondary authenticator can be set as default. The default authenticator is used first.

When performing authentication, all authenticators possessed by the user can be
used, regardless of the configured authenticator types.

When an identity is removed, all matching authenticators are also removed. For
example, removing a login ID identity would also remove the OOB-OTP 
authenticators using same login ID as target.

### Primary Authenticator

Primary authenticators authenticate the identity. Each identity has specific applicable primary authenticators. For example, OAuth Identity does not have any applicable primary authenticators.

### Secondary Authenticator

Secondary authenticators are additional authentication methods to ensure higher degree of confidence in authenticity.

### Authenticator Types

#### Password Authenticator

Password authenticator can be primary or secondary. Every user has at most 1 primary password authenticator, and at most 1 secondary password authenticator.

#### TOTP Authenticator

TOTP authenticator can only be secondary.

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

Out-of-band One-time-password authenticator can be primary or secondary.

OOB-OTP authenticator is bound to a recipient address. The recipient can be an email address or phone number that can receive SMS messages.

An primary OOB-OTP authenticator is associated with a login ID identity.
When the user no longer owns an email address or phone number, the primary OOB-OTP authentictor with the same email address or phone number is said to be orphaned.
Orphaned authenticators are deleted along with the last identity owning an email address or phone number.

Secondary OOB-OTP authenticators are not associated with identities.

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

## Claims

Claims are computed information about the user.
The computed claims is controlled by [claims mapping](#claims-mapping).

## Claims Mapping

The claims mapping is a list in the app configuration.
The mapping affects the claims of all users of the app.
The mapping in the app configuration can override [builtin claims mapping](#builtin-claims-mapping) with the same `name_pointer`.

### Builtin claims mapping

Builtin claims mapping is a list of mapping entries with `kind: "system"` and supported standard claims name.

The full list is:

```
- kind: "system"
  name_pointer: "#/email"
- kind: "system"
  name_pointer: "#/email_verified"
- kind: "system"
  name_pointer: "#/phone_number"
- kind: "system"
  name_pointer: "#/phone_number_verified"
- kind: "system"
  name_pointer: "#/preferred_username"
```

That is, the claims `email`, `email_verified`, `phone_number`, `phone_number_verified` and `preferred_username` are supported natively without any configuration.

#### Builtin claims mapping - email

```
- kind: "system"
  name_pointer: "#/email"
```

When this mapping is present, the `email` claim of the user is the `email` standard attribute.

#### Builtin claims mapping - email\_verified

```
- kind: "system"
  name_pointer: "#/email_verified"
```

When this mapping is present, the `email_verified` claim of the user is computed using the information of [user verification](./verification.md)

#### Builtin claims mapping - phone\_number

```
- kind: "system"
  name_pointer: "#/phone_number"
```

When this mapping is present, the `phone_number` claim of the user is the `phone_number` standard attribute.

#### Builtin claims mapping - phone\_number\_verified

```
- kind: "system"
  name_pointer: "#/phone_number_verified"
```

When this mapping is present, the `phone_number_verified` claim of the user is computed using the information of [user verification](./verification.md)

#### Builtin claims mapping - preferred\_username

```
- kind: "system"
  name_pointer: "#/preferred_username"
```

When this mapping is present, the `preferred_username` claim of the user is the `preferred_username` standard attribute.

### Custom attributes claims mapping

```
- kind: "custom_attributes"
  name_pointer: "#/zoneinfo"
  value_pointer: "#/profile/preferred_timezone"
```

Custom attributes claims mapping project the value pointed by `value_pointer` in the custom attributes into the location pointed by `name_pointer` in the claims.

### Claims Mapping Example

The following example illustrates the full capability of claims mapping.

Given the following effective claims mapping

```yaml
mapping:
- kind: "system"
  name_pointer: "#/email"
- kind: "system"
  name_pointer: "#/email_verified"
- kind: "system"
  name_pointer: "#/phone_number"
- kind: "system"
  name_pointer: "#/phone_number_verified"
- kind: "system"
  name_pointer: "#/preferred_username"

- kind: "custom_attributes"
  name_pointer: "#/zoneinfo"
  value_pointer: "#/profile/preferred_timezone"
- kind: "custom_attributes"
  name_pointer: "#/picture"
  value_pointer: "#/profile/profile_image_url"
- kind: "custom_attributes"
  name_pointer: "#/app:rbac"
  value_pointer: "#/rbac"
```

and the user has the following standard attributes

```json
{
  "email": "user@example.com"
}
```

and the user has the following custom attributes

```json
{
  "profile": {
    "preferred_timezone": "Asia/Hong_Kong",
    "profile_image_url": "https://cdn.example.com/u/user-a.jpg"
  },
  "rbac": [
    "product:list",
    "product:get",
    "product:delete"
  ]
}
```

The computed claims of the user is

```json
{
  "email": "user@example.com",
  "email_verified": true,
  "zoneinfo": "Asia/Hong_Kong",
  "picture": "https://cdn.example.com/u/user-a.jpg",
  "app:rbac": [
    "product:list",
    "product:get",
    "product:delete"
  ]
}
```

### Claims mapping JSON schema

This is the JSON schema defining claims mapping

```json
{
  "type": "array",
  "items": {
    "oneOf": [
      {
        "type": "object",
        "properties": {
          "kind": { "const": "system" },
          "name_pointer": {
            "type": "string",
            "enum": [
              "#/email",
              "#/email_verified",
              "#/phone_number",
              "#/phone_number_verified",
              "#/preferred_username"
            ]
          },
        },
        "required": ["kind", "name_pointer"]
      },
      {
        "type": "object",
        "properties": {
          "kind": { "const": "custom_attributes" },
          "name_pointer": { "type": "string" },
          "value_pointer": { "type": "string" }
        },
        "required": ["kind", "name_pointer", "value_pointer"]
      }
    ]
  }
}
```

- `kind`: The kind of the mapping, either `system` or `custom_attributes`.
- `name_pointer`: The JSON pointer to indicate which part of the claims is mapped to.
- `value_pointer`: The JSON pointer to indicate which part of the source claims is mapped from.

> NOTE: In the initial design, we also have the kind `specified`, which simply means a provided value.
> This was dropped because the similar effect can be achieved by storing a custom attribute and then use the kind `custom_attributes`.
