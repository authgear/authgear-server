- [Goals](#goals)
- [Non-goals](#non-goals)
- [Overview](#overview)
- [Review the authentication UI / UX of existing consumer apps](#review-the-authentication-ui--ux-of-existing-consumer-apps)
- [Review the design of various competitors](#review-the-design-of-various-competitors)
  * [Auth0](#auth0)
  * [Okta](#okta)
  * [Azure AD B2C](#azure-ad-b2c)
  * [Zitadel](#zitadel)
  * [Supertokens](#supertokens)
- [Design](#design)
  * [Design Principles](#design-principles)
  * [What is a signup flow](#what-is-a-signup-flow)
  * [What is a login flow](#what-is-a-login-flow)
  * [What is a reauth flow](#what-is-a-reauth-flow)
  * [Design of the configuration](#design-of-the-configuration)
    + [Design overview](#design-overview)
    + [Object: AuthenticationMethod](#object-authenticationmethod)
    + [Object: Flow](#object-flow)
    + [Object: Step](#object-step)
    + [Object: SignupFlow](#object-signupflow)
    + [Object: LoginFlow](#object-loginflow)
    + [Object: SignupLoginFlow](#object-signuploginflow)
    + [Object: ReauthFlow](#object-reauthflow)
    + [Use case example 1: Latte](#use-case-example-1-latte)
    + [Use case example 2: Uber](#use-case-example-2-uber)
    + [Use case example 3: Google](#use-case-example-3-google)
    + [Use case example 4: The Club](#use-case-example-4-the-club)
    + [Use case example 5: Manulife MPF](#use-case-example-5-manulife-mpf)
    + [Use case example 6: Comprehensive example](#use-case-example-6-comprehensive-example)
- [Expressions](#expressions)
  * [Literals](#literals)
  * [Operators](#operators)
  * [Functions](#functions)
    + [contains](#contains)
    + [fromJSON](#fromjson)
  * [Contexts](#contexts)
- [Appendix](#appendix)
  * [JSON Schema of `authentication_methods`](#json-schema-of-authentication_methods)
  * [JSON schema of `signup_flows`](#json-schema-of-signup_flows)
  * [JSON schema of `login_flows`](#json-schema-of-login_flows)
  * [JSON Schema of `signup_login_flows`](#json-schema-of-signup_login_flows)
  * [JSON Schema of `reauth_flows`](#json-schema-of-reauth_flows)
  * [Unused design that allow nested steps](#unused-design-that-allow-nested-steps)

## Goals

- Support customized signup flow
- Support customized login flow
- Support customized reauth flow
- Support more than 1 flows for signup / login / reauth
- Support combined signup and login flow
- The customized flows are supported by Default UI out of the box
- If Default UI does not suit the taste of the developer, the customized flows can be executed by a custom UI.
- (Future works) Support account linking flow

## Non-goals

- Build a generic workflow engine

## Overview

Before we get down to the design, we first

1. Review the authentication UI / UX of existing consumer apps
2. Review the design of various competitors

That will provide us insights into how to design our workflow

## Review the authentication UI / UX of existing consumer apps

This notion records the authentication flows of existing consumer apps in Hong Kong.
https://www.notion.so/oursky/Common-Signup-Login-Flows-f62e48724dc041d29aa0a77ec1dae806

Some important observations drawn from this review.
- Most consumer apps do not support 2FA.
- The authentication method is not necessarily tied to the identification method. For example, in The Club app, user can first enter their email address, and then receive a Phone OTP to sign in.

## Review the design of various competitors

### Auth0

Auth0 offers Triggers, Actions and Flows. https://auth0.com/docs/customize/actions/flows-and-triggers Auth0 does not support fully customized flows. Instead, it defines some Triggers, and allow the developer to write their own Actions to build Flows.

### Okta

Okta is based on Workflows. But the Workflows it offer are mainly for building business workflows, instead of customizing the authentication flow. In the documentation, it only documents how to customize a step in the authentication flow. https://help.okta.com/wf/en-us/Content/Topics/Workflows/connector-builder/authentication-custom.htm

### Azure AD B2C

Azure AD B2C allows customization via custom policy. Custom policy is configured by configuration files. The custom policy has a few key concepts.

- Claims are the foundation of a custom policy.
- User Journey defines how the user authenticates themselves.
- A User Journey contains several Orchestration Steps.
- Each Orchestration Step can be executed conditionally.
- An Orchestration Step must refer to a Technical profile.
- A Technical Profile defines its input Claims and output Claims.

Therefore, the end-user goes through the User Journey, with more and more Claims being collected in each Orchestration Step.

https://learn.microsoft.com/en-us/azure/active-directory-b2c/custom-policy-overview

### Zitadel

Zitadel is experimenting with a new Resource-based API. The Resource-based API has a Session API. The Session API is data-driven. For example, the developer can ask Zitadel to authenticate the user and verify the password by creating a session of the following shape

```json
{
  "checks": {
    "user": {
      "loginName": "mini@mouse.com"
    },
    "password": {
      "password": "V3ryS3cure!"
    }
  }
}
```

More complicated flows could be supported by supporting more `checks`, as proposed by [this comment](https://github.com/zitadel/zitadel/discussions/5875#discussioncomment-5985323)

https://github.com/zitadel/zitadel/discussions/5922

### Supertokens

Supertokens requires the developer to host a backend server to interactive with the Core Driver Interface (CDI) https://app.swaggerhub.com/apis/supertokens/CDI/2.21.1 The CDI is not very flexible. For example, it only supports some pre-defined recipe like EmailPassword Recipe, Passwordless Recipe.

## Design

### Design Principles

- We want to design a configuration such that the default UI can just read the configuration and execute the customized flows.
- We want to keep the guarantee that the existing configuration ensures every new user has some certain identities and authenticators.
- We want to be able to fulfill the authentication flows in existing consumer apps

### What is a signup flow
- Create 1 or more Identities. Later on, the User identify themselves with one of the Identities.
- Create 0 or more Authenticators. The User authenticates themselves with one of the Authenticators if needed.
- (Optional) Collect user profile (i.e. standard attributes and custom attributes)

### What is a login flow
- Identify the User with an Identity
- Depending on the Identity, authenticator the User with the Authenticators. Note that the pre-selected Authenticator is usually associated with the Identity.
- (Optional) Further authenticate the User with **other** Authenticators.

Suppose the User has a Email Login ID Identity `johndoe@gmail.com`, a Email OOB-OTP Authenticator `johndoe@gmail.com`, a Phone Login ID Identity `+85298765432`, a Phone OOB-OTP Authenticator `+85298765432`, a Password Authenticator, and a OAuth Identity `johndoe@gmail.com`.

If the User identifies themselves with the Email Login ID Identity `johndoe@gmail.com`, then the User can authenticate themselves with:
1. Email OOB-OTP Authenticator `johndoe@gmail.com`
2. Password Authenticator
3. Phone OOB-OTP Authenticator `+85298765432`. Note that this Authenticator is NOT associated with the identifying Identity, but it can also be used to authenticate the User.

If the User identifies themselves with the OAuth Identity `johndoe@gmail.com`, then the User DOES NOT need to authenticate themselves. This is how most other applications work.

### What is a reauth flow
- Authenticate the User with any Authenticators.

### Design of the configuration

#### Design overview

The configuration is a small DSL with a simple expression language.
Every object must have an `id`. The namespace in which the `id` is in varies by object.

The configuration maps to a user flow as performed by a end-user.
A user flow consists of one or more screens, and the configuration consists of one or more steps.
Screens are not nested, so steps are organized linearly, and not nested as well.

#### Object: AuthenticationMethod

We define available AuthenticationMethods under `authentication_methods`.
The `id` is in the global namespace.

Example:

```yaml
authentication_methods:
# Authenticate with a password
- id: primary_password
  kind: primary
  type: password
# Authenticate with Out-of-band One-time-password associated with an email address.
# Whether it is code or link depends on the configuration in another place.
- id: primary_oob_otp_email
  kind: primary
  type: oob_otp_email
# Authenticate with Out-of-band One-time-password associated with a phone number.
# Whether it is SMS or Whatsapp depends on the configuration in another place.
- id: primary_oob_otp_sms
  kind: primary
  type: oob_otp_sms

# 2FA with an additional password
- id: secondary_password
  kind: secondary
  type: password
# 2FA with Out-of-band One-time-password associated with an email address.
# Whether it is code or link depends on the configuration in another place.
- id: secondary_oob_otp_email
  kind: secondary
  type: oob_otp_email
# 2FA with Out-of-band One-time-password associated with a phone number.
# Whether it is SMS or Whatsapp depends on the configuration in another place.
- id: secondary_oob_otp_sms
  kind: secondary
  type: oob_otp_sms
# 2FA with a time-based 6-digit code
- id: secondary_totp
  kind: secondary
  type: totp
# 2FA with 10-letter one-time-use recovery code
- id: secondary_recovery_code
  kind: secondary
  type: recovery_code
# Skip 2FA on trusted device
- id: secondary_device_token
  kind: secondary
  type: device_token
```

#### Object: Flow

A Flow has one or more Steps.
The `id` is in the global namespace.

#### Object: Step

A Step is associated with one Flow.
The `id` is the namespace of the Flow.
So Steps in different Flows can share the same `id`.

A Step is executed conditionally if it has a `if`.
The value is an [Expression](#expressions).

The `id` of a Flow can be omitted if the developer need not reference it in a later step.
An `id` will be randomly generated in such case.

#### Object: SignupFlow

A SignupFlow is a Flow.

Example:

```yaml
signup_flows:
- id: default_flow
  steps:
  # Sign up with either a phone number or an email address.
  - id: setup_identity
    type: identify
    one_of:
    - identification: phone
    - identification: email
  # Set up a phone OTP authenticator for the phone number
  - type: authenticate
    if: steps.setup_identity.identification == "phone"
    one_of:
    - authentication: primary_oob_otp_sms
      target_step: setup_identity
  # Set up an email OTP authenticator for the email address.
  - type: authenticate
    if: steps.setup_identity.identification == "email"
    one_of:
    - authentication: primary_oob_otp_email
      target_step: setup_identity
  # Set up a primary password.
  - type: authenticate
    one_of:
    - authentication: primary_password
  # Verify the phone number or the email address
  # If this step is not specified, the phone number or the email address is unverified.
  - type: verify
    if: contains(fromJSON('["phone", "email"]'), steps.setup_identity.identification)
    target_step: setup_identity
  # Set up another phone number for 2FA.
  - type: authenticate
    one_of:
    - authentication: secondary_sms_code
  # Verify the phone number in the previous step.
  - type: verify
    target_step: setup_phone_2fa
  # Collect given name and family name.
  - type: user_profile
    user_profile:
    - pointer: /given_name
      required: true
    - pointer: /family_name
      required: true
  # Collect custom attributes.
  - type: user_profile
    user_profile:
    - pointer: /x_age
      required: true
```

#### Object: LoginFlow

A LoginFlow is a Flow.

Example:

```yaml
login_flows:
# Sign in with a phone number and OTP via SMS to any phone number the account has.
- id: phone_otp_to_any_phone
  steps:
  - type: identify
    one_of:
    - identification: phone
  - type: authenticate
    one_of:
    - authentication: primary_oob_otp_sms
# Sign in with a phone number and OTP via SMS to the same phone number.
- id: phone_otp_to_same_phone
  steps:
  - id: identify
    type: identify
    one_of:
    - identification: phone
  - type: authenticate
    one_of:
    - authentication: primary_oob_otp_sms
      target_step: identify
# Sign in with a phone number and a password
- id: phone_password
  steps:
  - type: identify
    one_of:
    - identification: phone
  - type: authenticate
    one_of:
    - authentication: primary_password
# Sign in with a phone number, or an email address, with a password
- id: phone_email_password
  steps:
  - type: identify
    one_of:
    - identification: phone
    - identification: email
  - id: authenticate
    type: authenticate
    one_of:
    - authentication: primary_password
# Sign in with an email address, a password and a TOTP
- id: email_password_totp
  steps:
  - type: identify
    one_of:
    - identification: email
  - id: first_factor
    type: authenticate
    one_of:
    - authentication: primary_password
  - type: authenticate
    one_of:
    - authentication: secondary_totp
```

#### Object: SignupLoginFlow

A SignupLoginFlow is a Flow.

Example:

```yaml
signup_login_flows:
- id: default_signup_login_flow
  steps:
  - id: step
    type: identify
    one_of:
    - identification: phone
      signup_flow: default_signup_flow
      login_flow: default_login_flow
    - identification: email
      signup_flow: default_signup_flow
      login_flow: default_login_flow
```

#### Object: ReauthFlow

A ReauthFlow is a Flow.

Example:

```yaml
reauth_flows:
# Re-authenticate with primary password.
- id: reauth_password
  steps:
  - id: password
    type: authenticate
    one_of:
    - authentication: primary_password

# Re-authenticate with any 2nd factor, assuming that 2FA is required in signup flow.
- id: reauth_2fa
  steps:
  - id: second_factor
    type: authenticate
    one_of:
    - authentication: secondary_totp
    - authentication: secondary_sms_code

# Re-authenticate with the 1st factor AND the 2nd factor.
- id: reauth_full
  steps:
  - id: first_factor
    type: authenticate
    one_of:
    - authentication: primary_password
  - id: second_factor
    type: authenticate
    one_of:
    - authentication: secondary_totp
    - authentication: secondary_sms_code
```

#### Use case example 1: Latte

```yaml
authentication_methods:
- id: primary_oob_otp_sms
  kind: primary
  type: oob_otp_sms
- id: primary_oob_otp_email
  kind: primary
  type: oob_otp_email
- id: primary_password
  kind: primary
  type: password

signup_flows:
- id: default_signup_flow
  steps:
  - type: identify
    id: setup_phone
    one_of:
    - identification: phone
  - type: authenticate
    one_of:
    - authentication: primary_oob_otp_sms
      target_step: setup_phone
  - type: verify
    target_step: setup_phone
  - type: identify
    id: setup_email
    one_of:
    - identification: email
  - type: authenticate
    one_of:
    - authentication: primary_oob_otp_email
      target_step: setup_email
  - type: authenticate
    one_of:
    - authentication: primary_password

login_flows:
- id: default_login_flow
  steps:
  - type: identify
    one_of:
    - identification: phone
  - type: authenticate
    one_of:
    - authentication: primary_oob_otp_sms
  - type: authenticate
    one_of:
    - authentication: primary_oob_otp_email
    - authentication: primary_password
```

#### Use case example 2: Uber

```yaml
authentication_methods:
- id: primary_oob_otp_sms
  kind: primary
  type: oob_otp_sms
- id: primary_oob_otp_email
  kind: primary
  type: oob_otp_email
- id: primary_password
  kind: primary
  type: password

signup_flows:
- id: phone_first
  steps:
  - type: identify
    id: setup_phone
    one_of:
    - identification: phone
  - type: authenticate
    one_of:
    - authentication: primary_oob_otp_sms
      target_step: setup_phone
  - type: verify
    target_step: setup_phone
  - type: identify
    id: setup_email
    one_of:
    - identification: email
  - type: authenticate
    one_of:
    - authentication: primary_oob_otp_email
      target_step: setup_email
  - type: verify
    target_step: setup_email
  - type: authenticate
    one_of:
    - authentication: primary_password
- id: email_first
  steps:
  - type: identify
    id: setup_email
    one_of:
    - identification: email
  - type: authenticate
    one_of:
    - authentication: primary_oob_otp_email
      target_step: setup_email
  - type: verify
    target_step: setup_email
  - type: identify
    id: setup_phone
    one_of:
    - identification: phone
  - type: authenticate
    one_of:
    - authentication: primary_oob_otp_sms
      target_step: setup_phone
  - type: verify
    target_step: setup_phone
  - type: authenticate
    one_of:
    - authentication: primary_password

login_flows:
- id: default_login_flow
  steps:
  - id: identify
    type: identify
    one_of:
    - identification: phone
    - identification: email
  - type: authenticate
    if: steps.identify.identification == "phone"
    one_of:
    - authentication: primary_oob_otp_sms
    - authentication: primary_password
  - type: authenticate
    if: steps.identify.identification == "email"
    one_of:
    - authentication: primary_oob_otp_email
    - authentication: primary_oob_otp_sms
    - authentication: primary_password

signup_login_flows:
- id: default_signup_login_flow
  steps:
  - id: step
    type: identify
    one_of:
    - identification: phone
      login_flow: default_login_flow
      signup_flow: phone_first
    - identification: email
      login_flow: default_login_flow
      signup_flow: default_signup_flow
```

#### Use case example 3: Google

```yaml
authentication_methods:
- id: primary_password
  kind: primary
  type: password
- id: secondary_sms_code
  kind: secondary
  type: oob_otp_sms
- id: secondary_totp
  kind: secondary
  type: totp

signup_flows:
- id: default_signup_flow
  steps:
  - type: identify
    one_of:
    - identification: email
  - type: authenticate
    one_of:
    - authentication: primary_password

login_flows:
- id: default_login_flow
  steps:
  - type: identify
    one_of:
    - identification: email
  - type: authenticate
    one_of:
    - authentication: primary_password
  - type: authenticate
    one_of:
    - authentication: secondary_totp
    - authentication: secondary_sms_code
```

#### Use case example 4: The Club

```yaml
authentication_methods:
- id: primary_password
  kind: primary
  type: password
- id: primary_oob_otp_sms
  kind: primary
  type: oob_otp_sms

# signup_flows is omitted here because the exact signup flow is unknown.

login_flows:
- id: default_login_flow
  steps:
  - type: identify
    one_of:
    - identification: email
    - identification: phone
    - identification: username
  - type: authenticate
    one_of:
    - authentication: primary_password
    - authentication: primary_oob_otp_sms
```

#### Use case example 5: Manulife MPF

```yaml
authentication_methods:
- id: primary_password
  kind: primary
  type: password
- id: primary_oob_otp_sms
  kind: primary
  type: oob_otp_sms
- id: primary_oob_otp_email
  kind: primary
  type: oob_otp_email

# signup_flows are omitted because it does not have public signup.

login_flows:
- id: default_login_flow
  steps:
  - type: identify
    one_of:
    - identification: username
  - type: authenticate
    one_of:
    - authentication: primary_password
  - type: authenticate
    one_of:
    - authentication: primary_oob_otp_sms
    - authentication: primary_oob_otp_email
```

#### Use case example 6: Comprehensive example

```yaml
authentication_methods:
- id: primary_password
  kind: primary
  type: password
- id: primary_passkey
  kind: primary
  type: passkey
- id: primary_oob_otp_email
  kind: primary
  type: oob_otp_email
- id: secondary_totp
  kind: secondary
  type: totp
- id: recovery_code
  kind: secondary
  type: recovery_code
- id: device_token
  kind: secondary
  type: device_token

signup_flows:
# The end user sign up with OAuth without password or 2FA.
# Or the end user sign up with verified email with password and 2FA.
- id: default_signup_flow
  steps:
  - id: setup_identity
    type: identify
    one_of:
    - identification: email
    - identification: oauth
  - type: authenticate
    if: steps.setup_identity.identification == "email"
    one_of:
    - authentication: primary_oob_otp_email
      target_step: setup_identity
  - type: verify
    if: steps.setup_identity.identification == "email"
    target_step: setup_identity
  - id: setup_first_factor
    type: authenticate
    if: steps.setup_identity.identification == "email"
    one_of:
    - authentication: primary_password
  - type: authenticate
    if: steps.setup_first_factor.authentication != null
    one_of:
    - authentication: secondary_totp

login_flows:
# The end user can sign in with OAuth.
# The end user can sign in with passkey directly.
# The end user can sign in with email with OTP, password, or passkey, and with 2FA.
- id: default_login_flow
  steps:
  - id: identify
    type: identify
    one_of:
    - identification: email
    - identification: oauth
    - identification: passkey
  - id: first_factor
    type: authenticate
    if: steps.identify.identification == "email"
    one_of:
    - authentication: primary_password
    - authentication: primary_oob_otp_email
    - authentication: primary_passkey
  - type: authenticate
    if: steps.first_factor.authentication != null && setup.first_factor.authentication != "primary_passkey"
    one_of:
    - authentication: secondary_totp
    - authentication: recovery_code
    - authentication: device_token
```

## Expressions

Expressions are used to determine whether a Step should run.

### Literals

|Data Type|Literal|
|---|---|
|`boolean`|`true` or `false`|
|`null`|`null`|
|`number`|[A JSON number](https://datatracker.ietf.org/doc/html/rfc8259#section-6)|
|`string`|[A JSON string](https://datatracker.ietf.org/doc/html/rfc8259#section-7)|

### Operators

|Operator|Meaning|
|---|---|
|`( )`|Grouping|
|`.`|Property access|
|`!`|Logical negation|
|`==`|Equal|
|`!=`|Not equal|
|`&&`|And|
|`||`|Or|

### Functions

The following built-in functions can be used in expressions.

#### contains

`contains(array, item)`

`contains` returns `true` if `item` is an element of `array`.

#### fromJSON

`fromJSON(jsonString)`

`fromJSON` returns a value represented by `jsonString`.

### Contexts

The only place an expression can appear is in the `if` of a Step.
The context of that place is described as follows.

|Expression|Type|Description|
|---|---|---|
|`steps`|`object`|The `steps` object|
|`steps.<id>`|`object`|The `step` object|
|`steps.<id>.identification`|`string` or `null`|The identification method selected in the step|
|`steps.<id>.authentication`|`string` or `null`|The `id` of the selected `authentication_method` in the step|

## Appendix

### JSON Schema of `authentication_methods`

```json
{
  "properties": {
    "authentication_methods": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "id": { "type": "string", "minLength": 1 },
          "kind": {
            "type": "string",
            "enum": ["primary", "secondary"]
          },
          "type": {
            "type": "string",
            "enum": [
              "password",
              "passkey",
              "oob_otp_email",
              "oob_otp_sms",
              "totp",
              "recovery_code",
              "device_token"
            ]
          }
        },
        "required": ["id", "kind", "type"]
      }
    }
  }
}
```

### JSON schema of `signup_flows`

```json
{
  "properties": {
    "signup_flows": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["id", "steps"],
        "properties": {
          "id": { "type": "string", "minLength": 1 },
          "steps": {
            "type": "array",
            "items": {
              "type": "object",
              "required": ["type"],
              "properties": {
                "id": { "type": "string", "minLength": 1 },
                "if": { "type": "string", "format": "x_expression"} },
                "type": {
                  "type": "string",
                  "enum": [
                    "identify",
                    "authenticate",
                    "verify",
                    "user_profile"
                  ]
                }
              },
              "allOf": [
                {
                  "if": {
                    "properties": {
                      "type": { "const": "identify" }
                    }
                  },
                  "then": {
                    "required": ["one_of"],
                    "properties": {
                      "one_of": {
                        "type": "array",
                        "items": {
                          "type": "object",
                          "required": ["identification"],
                          "properties": {
                            "identification": {
                              "type": "string",
                              "enum": [
                                "email",
                                "phone",
                                "username",
                                "oauth",
                                "passkey",
                                "siwe"
                              ]
                            }
                          }
                        }
                      }
                    }
                  }
                },
                {
                  "if": {
                    "properties": {
                      "type": { "const": "authenticate" }
                    }
                  },
                  "then": {
                    "required": ["one_of"],
                    "properties": {
                      "one_of": {
                        "type": "array",
                        "items": {
                          "type": "object",
                          "required": ["authentication"],
                          "properties": {
                            "authentication": {
                              "type": "string",
                              "minLength": 1
                            },
                            "target_step": {
                              "type": "string",
                              "minLength": 1
                            }
                          }
                        }
                      }
                    }
                  }
                },
                {
                  "if": {
                    "properties": {
                      "type": { "const": "verify" }
                    }
                  },
                  "then": {
                    "required": ["target_step"],
                    "properties": {
                      "target_step": {
                        "type": "string",
                        "minLength": 1
                      }
                    }
                  }
                },
                {
                  "if": {
                    "properties": {
                      "type": { "const": "user_profile" }
                    }
                  },
                  "then": {
                    "required": ["user_profile"],
                    "properties": {
                      "user_profile": {
                        "type": "array",
                        "items": {
                          "type": "object",
                          "required": ["pointer", "required"],
                          "properties": {
                            "pointer": { "type": "string" },
                            "required": { "type": "boolean" }
                          }
                        }
                      }
                    }
                  }
                }
              ]
            }
          }
        }
      }
    }
  }
}
```

### JSON schema of `login_flows`

```json
{
  "properties": {
    "login_flows": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["id", "steps"],
        "properties": {
          "id": { "type": "string", "minLength": 1 },
          "steps": {
            "type": "array",
            "items": {
              "type": "object",
              "required": ["type"],
              "properties": {
                "id": { "type": "string", "minLength": 1 },
                "if": { "type": "string", "format": "x_expression"} },
                "type": {
                  "type": "string",
                  "enum": [
                    "identify",
                    "authenticate"
                  ]
                }
              }
            },
            "allOf": [
              {
                "if": {
                  "properties": {
                      "type": { "const": "identify" }
                  }
                },
                "then": {
                  "required": ["one_of"],
                  "properties": {
                    "one_of": {
                      "type": "array",
                      "items": {
                        "type": "object",
                        "required": ["identification"],
                        "properties": {
                          "identification": {
                            "type": "string",
                            "enum": [
                              "email",
                              "phone",
                              "username",
                              "oauth",
                              "passkey",
                              "siwe"
                            ]
                          }
                        }
                      }
                    }
                  }
                }
              },
              {
                "if": {
                  "properties": {
                    "type": { "const": "authenticate" }
                  }
                },
                "then": {
                  "required": ["one_of"],
                  "properties": {
                    "one_of": {
                      "type": "array",
                      "items": {
                        "type": "object",
                        "required": ["authentication"],
                        "properties": {
                          "authentication": {
                            "type": "string",
                            "minLength": 1
                          },
                          "target_step": {
                            "type": "string",
                            "minLength": 1
                          }
                        }
                      }
                    }
                  }
                }
              }
            ]
          }
        }
      }
    }
  }
}
```

### JSON Schema of `signup_login_flows`

```json
{
  "properties": {
    "signup_login_flows": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["id", "steps"],
        "properties": {
          "id": { "type": "string", "minLength": 1 },
          "steps": {
            "type": "object",
            "required": ["type"],
            "properties": {
              "id": { "type": "string", "minLength": 1 },
              "if": { "type": "string", "format": "x_expression"} },
              "type": {
                "type": { "type": "string", "enum": ["identify"] }
              }
            },
            "allOf": [
              {
                "if": {
                  "properties": {
                    "type": { "const": "identify" }
                  }
                },
                "then": {
                  "required": ["one_of"],
                  "properties": {
                    "one_of": {
                      "type": "array",
                      "items": {
                        "type": "object",
                        "required": ["identification", "signup_flow", "login_flow"],
                        "properties": {
                          "identification": {
                            "type": "string",
                            "enum": [
                              "email",
                              "phone",
                              "username",
                              "oauth",
                              "passkey",
                              "siwe"
                            ]
                          },
                          "signup_flow": {
                            "type": "string",
                            "minLength": 1
                          },
                          "login_flow": {
                            "type": "string",
                            "minLength": 1
                          }
                        }
                      }
                    }
                  }
                }
              }
            ]
          }
        }
      }
    }
  }
}
```

### JSON Schema of `reauth_flows`

```json
{
  "properties": {
    "reauth_flows": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["id", "steps"],
        "properties": {
          "id": { "type": "string", "minLength": 1 },
          "steps": {
            "type": "array",
            "items": {
              "type": "object",
              "required": ["type"],
              "properties": {
                "id": { "type": "string", "minLength": 1 },
                "if": { "type": "string", "format": "x_expression"} },
                "type": { "type": "string", "enum": ["authenticate"] }
              },
              "allOf": [
                {
                  "if": {
                    "properties": {
                      "type": { "const": "authenticate" }
                    }
                  },
                  "then": {
                    "properties": {
                      "one_of": {
                        "type": "array",
                        "items": {
                          "type": "object",
                          "required": ["authentication"],
                          "properties": {
                            "authentication": {
                              "type": "string",
                              "minLength": 1
                            }
                          }
                        }
                      }
                    }
                  }
                }
              ]
            }
          }
        }
      }
    }
  }
}
```

### Unused design that allow nested steps

Here we show how would the configuration looks like if nested steps were allowed.
We take the [Use case example 6](#use-case-example-6-comprehensive-example) and rewrite it.

```yaml
authentication_methods:
- id: primary_password
  kind: primary
  type: password
- id: primary_passkey
  kind: primary
  type: passkey
- id: primary_oob_otp_email
  kind: primary
  type: oob_otp_email
- id: secondary_totp
  kind: secondary
  type: totp
- id: recovery_code
  kind: secondary
  type: recovery_code
- id: device_token
  kind: secondary
  type: device_token

signup_flows:
- id: default_signup_flow
  steps:
  - id: setup_identity
    type: identify
    one_of:
    - identification: oauth
    - identification: email
      steps:
      - type: authenticate
        one_of:
        - authentication: primary_oob_otp_email
          target_step: setup_identity
      - type: verify
        target_step: setup_identity
      - id: setup_first_factor
        type: authenticate
        one_of:
        - authentication: primary_password
      - type: authenticate
        one_of:
        - authentication: secondary_totp

login_flows:
- id: default_login_flow
  steps:
  - id: identify
    type: identify
    one_of:
    - identification: oauth
    - identification: passkey
    - identification: email
      steps:
      - type: authenticate
        one_of:
        - authentication: primary_password
          steps:
          - type: authenticate
            one_of:
            - authentication: secondary_totp
            - authentication: recovery_code
            - authentication: device_token
        - authentication: primary_oob_otp_email
          steps:
          - type: authenticate
            one_of:
            - authentication: secondary_totp
            - authentication: recovery_code
            - authentication: device_token
        - authentication: primary_passkey
```
