* [Glossary](#glossary)
  * [User](#user)
    + [Standard Attributes](#standard-attributes)
    + [Custom Attributes](#custom-attributes)
    + [Identity](#identity)
      + [Identity Attributes](#identity-attributes)
    + [Authenticator](#authenticator)
    + [User Info](#user-info)
  * [OAuth User Profile](#oauth-user-profile)
  * [Authentication Flow](#authentication-flow)
    + [Identification](#identification)
    + [Authentication](#authentication)

# Glossary

## User

[User](./user-model.md#user)

### Standard Attributes

[Standard Attributes](./user-profile/design.md#standard-attributes)

### Custom Attributes

[Custom Attributes](./user-profile/design.md#custom-attributes)

### Identity

[Identity](./user-model.md#identity)

#### Identity Attributes

[Identity Attributes](./account-linking.md#identity-attributes)

### Authenticator

[Authenticator](./user-model.md#authenticator)

### User Info

[User Info](./user-profile/design.md#user-info-endpoint)

## OAuth User Profile

[OAuth User Profile](./sso-providers.md#oauth-user-profile)

## Authentication Flow

[Authentication Flow](./authentication-flow.md#authentication-flow)

### Identification

Identification is the method the user uses to identify themself. For example, an email, phone, or username.

Read the [authentication flow API reference](./authentication-flow-api-reference.md) for details.

### Authentication

Authentication is the means the user uses to prove they are the user identified with the identification method. For example, by using a password, OTP, or biometrics.

Read the [authentication flow API reference](./authentication-flow-api-reference.md) for details.
