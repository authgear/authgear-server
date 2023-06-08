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
    + [`identification_methods`](#identification_methods)
    + [`authentication_methods`](#authentication_methods)
    + [`signup_flows`](#signup_flows)
    + [`login_flows`](#login_flows)
    + [`signup_login_flows`](#signup_login_flows)
    + [`reauth_flows`](#reauth_flows)
    + [Use case example 1: Latte](#use-case-example-1-latte)
    + [Use case example 2: Uber](#use-case-example-2-uber)
    + [Use case example 3: Google](#use-case-example-3-google)
    + [Use case example 4: The Club](#use-case-example-4-the-club)
- [Default UI](#default-ui)
  * [`login_flows` in Default UI](#login_flows-in-default-ui)
  * [`signup_login_flows` in Default UI](#signup_login_flows-in-default-ui)
- [Custom UI](#custom-ui)
- [Appendix](#appendix)
  * [JSON Schema of `identification_methods`](#json-schema-of-identification_methods)
  * [JSON Schema of `authentication_methods`](#json-schema-of-authentication_methods)
  * [JSON schema of `signup_flows`](#json-schema-of-signup_flows)
  * [JSON schema of `login_flows`](#json-schema-of-login_flows)
  * [JSON Schema of `signup_login_flows`](#json-schema-of-signup_login_flows)
  * [JSON Schema of `reauth_flows`](#json-schema-of-reauth_flows)

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

#### `identification_methods`

The developer defines all identification methods they allow with `identification_methods`. The `name` property is used in later configuration to refer to the identification method.

Detailed annotated configuration example:

```yaml
identification_methods:
# Identify the User by a Email Login ID Identity
- name: email
  identity:
    type: "login_id"
    login_id:
      type: "email"
# Identify the User by a Phone Login ID Identity
- name: phone
  identity:
    type: "login_id"
    login_id:
      type: "phone"
# Identify the User by a Username Login ID Identity
- name: username
  identity:
    type: "login_id"
    login_id:
      type: "username"
# Identify the User by a OAuth Identity
- name: oauth
  identity:
    type: "oauth"
    oauth:
      aliases: ["google", "apple"]
# Identify the User by a Anonymous Identity
- name: anonymous
  identity:
    type: "anonymous"
# Identify the User by a Biometric Identity
- name: biometric
  identity:
    type: "biometric"
# Identify the User by a Passkey Identity
- name: passkey
  identity:
    type: "passkey"
# Identify the User by a Sign-in with Ethereum Identity
- name: siwe
  identity:
    type: "siwe"
```

#### `authentication_methods`

The developer defines all authentication methods they allow with `authentication_methods`. The `name` property is used in later configuration to refer to the authentication method.

Detailed annotated configuration example:

```yaml
authentication_methods:
# Authenticate with a password
- name: primary_password
  kind: primary
  type: password
# Authenticate with a 6-digit code delivered via email
- name: primary_email_code
  kind: primary
  type: oob_otp_email
  email_otp_mode: "code"
# Authenticate with a link delivered via email
- name: primary_email_login_link
  kind: primary
  type: oob_otp_email
  email_otp_mode: "login_link"
# Authenticate with a 6-digit code delivered via SMS
- name: primary_sms_code
  kind: primary
  type: oob_otp_sms
  phone_otp_mode: "sms"
# Authenticate with a 6-digit code delivered via Whatsapp
- name: primary_whatsapp
  kind: primary
  type: oob_otp_sms
  phone_otp_mode: "whatsapp"

# 2FA with an additional password
- name: secondary_password
  kind: secondary
  type: password
# 2FA with a 6-digit code delivered via email
- name: secondary_email_code
  kind: secondary
  type: oob_otp_email
  email_otp_mode: "code"
# 2FA with a link delivered via email
- name: secondary_email_login_link
  kind: secondary
  type: oob_otp_email
  email_otp_mode: "login_link"
# 2FA with a 6-digit code delivered via SMS
- name: secondary_sms_code
  kind: secondary
  type: oob_otp_sms
  phone_otp_mode: "sms"
# 2FA with a 6-digit code delivered via Whatsapp
- name: secondary_whatsapp
  kind: secondary
  type: oob_otp_sms
  phone_otp_mode: "whatsapp"
# 2FA with a time-based 6-digit code
- name: secondary_totp
  kind: secondary
  type: totp
# 2FA with 10-letter one-time-use recovery code
- name: secondary_recovery_code
  kind: secondary
  type: recovery_code
# Skip 2FA on trusted device
- name: secondary_device_token
  kind: secondary
  type: device_token
```

#### `signup_flows`

The developer defines signup flows under `signup_flows`. The `name` property is used to refer to the signup flow.

Detailed annotated configuration example:

```yaml
signup_flows:
- name: default_flow
  steps:
  # Sign up with either a phone number or an email address.
  # The phone number or the email address can receive OTP to sign in.
  - name: setup_phone_or_email
    type: identification_method
    one_of:
    - identification_method:
        name: phone
      authentication_methods:
      - name: primary_sms_code
    - identification_method:
        name: email
      authentication_methods:
      - name: primary_email_code
  # Set up a primary password.
  - name: setup_password
    type: authentication_method
    one_of:
    - authentication_method:
        name: primary_password
  # Verify the phone number or the email address
  # If this step is not specified, the phone number or the email address is unverified.
  - name: verify_phone_or_email
    type: verification
    step:
      name: setup_phone_or_email
  # Set up another phone number for 2FA.
  - name: setup_phone_2fa
    type: authentication_method
    one_of:
    - authentication_method:
        name: secondary_sms_code
  # Verify the phone number in the previous step.
  - name: verify_phone_2fa
    type: verification
    step:
      name: setup_phone_2fa
  # Collect given name and family name.
  - name: fill_in_names
    type: user_profile
    user_profile:
    - pointer: /given_name
      required: true
    - pointer: /family_name
      required: true
  # Collect custom attributes.
  - name: fill_custom_attributes
    type: user_profile
    user_profile:
    - pointer: /x_age
      required: true
```

#### `login_flows`

The developer defines login flows under `login_flows`. The `name` property is used to refer to the login flow.

Detailed annotated configuration example:

```yaml
login_flows:
# Sign in with a phone number and OTP via SMS.
- name: phone_otp
  primary_authentication:
    one_of:
    - identification_method:
        name: phone
      authentication_methods:
      - name: primary_sms_code
# Sign in with a phone number and a password
- name: phone_password
  primary_authentication:
    one_of:
    - identification_method:
        name: phone
      authentication_methods:
      - name: primary_password
# Sign in with a phone number, or an email address, with a password
- name: phone_email_password
  primary_authentication:
    one_of:
    - identification_method:
        name: phone
      authentication_methods:
      - name: primary_password
    - identification_method:
        name: email
      authentication_methods:
      - name: primary_password
# Sign in with an email address, a password and a TOTP
- name: email_password_totp
  primary_authentication:
    one_of:
    - identification_method:
        name: email
      authentication_methods:
      - name: primary_password
  secondary_authentications:
  - one_of:
    - authentication_method:
        name: secondary_totp
```

#### `signup_login_flows`

The developer defines combined signup login flows under `signup_login_flows`. The `name` property is used to refer to the flow.

Detailed annotated configuration example:

```yaml
signup_login_flows:
# Identify the User by a phone number or an email address
# Depending on whether the phone number or the email address is registered,
# execute the signup flow or the login flow.
- name: default_signup_login_flow
  identification_methods:
    one_of:
    - name: phone
      signup_flow:
        name: default_signup_flow
      login_flow:
        name: default_login_flow
    - name: email
      signup_flow:
        name: default_signup_flow
      login_flow:
        name: default_login_flow
```

#### `reauth_flows`

The developer defines reauth flows under `reauth_flows`. The `name` property is used to refer to the flow.

Detailed annotated configuration example:

```yaml
reauth_flows:
# Re-authenticate with primary password.
- name: reauth_password
  authentications:
  - one_of:
    - authentication_method:
        name: primary_password

# Re-authenticate with any 2nd factor, assuming that 2FA is required in signup flow.
- name: reauth_2fa
  authentications:
  - one_of:
    - authentication_method:
        name: secondary_totp
    - authentication_method:
        name: secondary_sms_code

# Re-authenticate with the 1st factor AND the 2nd factor.
- name: reauth_full
  authentications:
  - one_of:
    - authentication_method:
        name: primary_password
  - one_of:
    - authentication_method:
        name: secondary_totp
    - authentication_method:
        name: secondary_sms_code
```

#### Use case example 1: Latte

```yaml
identification_methods:
- name: phone
  identity:
    type: "login_id"
    login_id:
      type: "phone"
- name: email
  identity:
    type: "login_id"
    login_id:
      type: "email"

authentication_methods:
- name: primary_sms_code
  kind: primary
  type: oob_otp_sms
  phone_otp_mode: "sms"
- name: primary_email_login_link
  kind: primary
  type: oob_otp_email
  email_otp_mode: "login_link"
- name: primary_password
  kind: primary
  type: password

signup_flows:
- name: default_signup_flow
  steps:
  - name: setup_phone
    type: identification_method
    one_of:
    - identification_method:
        name: phone
      authentication_methods:
      - name: primary_sms_code
  - name: verify_phone
    type: verification
    step:
      name: setup_phone
  - name: setup_email
    type: identification_method
    one_of:
    - identification_method:
        name: email
      authentication_methods:
      - name: primary_email_login_link
  - name: setup_password
    type: authentication_method
    one_of:
    - authentication_method:
        name: primary_password

login_flows:
- name: default_login_flow
  primary_authentication:
    one_of:
    - identification_method:
        name: phone
      authentication_methods:
      - name: primary_sms_code
  secondary_authentications:
  - one_of:
    - authentication_method:
        name: primary_email_login_link
    - authentication_method:
        name: primary_password
```

#### Use case example 2: Uber

```yaml
identification_methods:
- name: phone
  identity:
    type: "login_id"
    login_id:
      type: "phone"
- name: email
  identity:
    type: "login_id"
    login_id:
      type: "email"

authentication_methods:
- name: primary_sms_code
  kind: primary
  type: oob_otp_sms
  phone_otp_mode: "sms"
- name: primary_email_code
  kind: primary
  type: oob_otp_email
  email_otp_mode: "code"
- name: primary_password
  kind: primary
  type: password

signup_flows:
- name: phone_first
  steps:
  - name: setup_phone
    type: identification_method
    one_of:
    - identification_method:
        name: phone
      authentication_methods:
      - name: primary_sms_code
  - name: verify_phone
    type: verification
    step:
      name: setup_phone
  - name: setup_email
    type: identification_method
    one_of:
    - identification_method:
        name: email
      authentication_methods:
      - name: primary_email_code
  - name: verify_email
    type: verification
    step:
      name: setup_email
  - name: setup_password
    type: authentication_method
    one_of:
    - authentication_method:
        name: primary_password
- name: email_first
  steps:
  - name: setup_email
    type: identification_method
    one_of:
    - identification_method:
        name: email
      authentication_methods:
      - name: primary_email_code
  - name: verify_email
    type: verification
    step:
      name: setup_email
  - name: setup_phone
    type: identification_method
    one_of:
    - identification_method:
        name: phone
      authentication_methods:
      - name: primary_sms_code
  - name: verify_phone
    type: verification
    step:
      name: setup_phone
  - name: setup_password
    type: authentication_method
    one_of:
    - authentication_method:
        name: primary_password

login_flows:
- name: default_login_flow
  primary_authentication:
    one_of:
    - identification_method:
        name: phone
      authentication_methods:
      - name: primary_sms_code
      - name: primary_password
    - identification_method:
        name: email
      authentication_methods:
      - name: primary_email_code
      - name: primary_sms_code
      - name: primary_password

signup_login_flows:
- name: default_signup_login_flow
  identification_methods:
    one_of:
    - name: phone
      login_flow:
        name: default_login_flow
      signup_flow:
        name: phone_first
    - name: email
      login_flow:
        name: default_login_flow
      signup_flow:
        name: email_first
```

#### Use case example 3: Google

```yaml
identification_methods:
- name: email
  identity:
    type: "login_id"
    login_id:
      type: "email"

authentication_methods:
- name: primary_password
  kind: primary
  type: password
- name: secondary_sms_code
  kind: secondary
  type: oob_otp_sms
  phone_otp_mode: "sms"
- name: secondary_totp
  kind: secondary
  type: totp

signup_flows:
- name: default_signup_flow
  steps:
  - name: setup_email
    type: identification_method
    one_of:
    - identification_method:
        name: email
  - name: setup_password
    type: authentication_method
    one_of:
    - authentication_method:
        name: primary_password

login_flows:
- name: default_login_flow
  primary_authentication:
    one_of:
    - identification_method:
        name: email
      authentication_methods:
      - name: primary_password
  secondary_authentications:
  - one_of:
    - authentication_method:
        name: secondary_totp
    - authentication_method:
        name: secondary_sms_code
```

#### Use case example 4: The Club

```yaml
identification_methods:
- name: email
  identity:
    type: "login_id"
    login_id:
      type: "email"
- name: phone
  identity:
    type: "login_id"
    login_id:
      type: "phone"
- name: username
  identity:
    type: "username"
    login_id:
      type: "username"

authentication_methods:
- name: primary_password
  kind: primary
  type: password
- name: primary_sms_code
  kind: primary
  type: oob_otp_sms
  phone_otp_mode: "sms"

# signup_flows is omitted here because the exact signup flow is unknown.

login_flows:
- name: default_login_flow
  primary_authentication:
    one_of:
    - identification_method:
        name: email
      authentication_methods:
      - name: primary_password
      - name: primary_sms_code
    - identification_method:
        name: phone
      authentication_methods:
      - name: primary_password
      - name: primary_sms_code
    - identification_method:
        name: username
      authentication_methods:
      - name: primary_password
      - name: primary_sms_code
```

## Default UI

The configuration is declarative. It only specifies what identity or authenticator to create. But the exact execution is unspecified in the configuration. In this section, we specify how Default UI execute flows according to the configuration.

Each `signup_flow`, `login_flow`, and `signup_login_flow` will be executed with a [Workflow](./workflow.md#workflow).

### `login_flows` in Default UI

Default UI executes login flow in the following order:

- Perform `primary_authentication`.
  - Identify the User with one of the `identification_method`.
  - Authenticate the User with one of the `authentication_methods`.
  - The used authenticator cannot be used again in `secondary_authentications`.
- Perform `secondary_authentication` one by one.
  - The used authenticator in each `secondary_authentication` cannot be used again in subsequent `secondary_authentications`.

### `signup_login_flows` in Default UI

Default UI executes signup login flow in the following order:

- Identify the User with one of the `identification_method`.
- If the User is identified, execute the login flow.
- Otherwise, execute the signup flow.

## Custom UI

Custom UI and Default UI share the same Workflow. If Custom UI wants to collect user profile first, it can first collect the user profile, and store it temporarily. After finishing create the identification methods and authentication methods, provide the user profile to the Workflow.

## Appendix

### JSON Schema of `identification_methods`

```json
{
  "properties": {
    "identification_methods": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name": { "type": "string", "minLength": 1 },
          "identity": {
            "type": "object",
            "properties": {
              "type": {
                "type": "string",
                "enum": [
                  "login_id",
                  "oauth",
                  "anonymous",
                  "biometric",
                  "passkey",
                  "siwe"
                ]
              },
              "login_id": {
                "type": "object",
                "properties": {
                  "type": {
                    "type": "string",
                    "enum": [
                      "email",
                      "phone",
                      "username"
                    ]
                  }
                }
              },
              "oauth" {
                "type": "object",
                "properties": {
                  "aliases": {
                    "type": "array",
                    "items": {
                      "type": "string"
                    }
                  }
                }
              }
            },
            "required": ["type"]
          }
        },
        "required": ["name", "identity"]
      }
    }
  }
}
```

### JSON Schema of `authentication_methods`

```json
{
  "properties": {
    "authentication_methods": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name": { "type": "string", "minLength": 1 },
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
          },
          "phone_otp_mode": {
            "type": "string",
            "enum": [
              "sms",
            "whatsapp",
            "whatsapp_sms"
            ]
          },
          "email_otp_mode": {
            "type": "string",
            "enum": [
              "code",
            "login_link"
            ]
          }
        },
        "required": ["name", "kind", "type"]
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
        "required": ["name", "steps"],
        "properties": {
          "name": { "type": "string", "minLength": 1 },
          "steps": {
            "type": "array",
            "items": {
              "type": "object",
              "required": ["name", "type"],
              "properties": {
                "name": { "type": "string", "minLength": 1 },
                "type": {
                  "type": "string",
                  "enum": [
                    "identification_method",
                    "authentication_method",
                    "verification",
                    "user_profile"
                  ]
                }
              },
              "allOf": [
                {
                  "if": {
                    "properties": {
                      "type": { "const": "identification_method" }
                    }
                  },
                  "then": {
                    "required": ["one_of"],
                    "properties": {
                      "one_of": {
                        "type": "array",
                        "items": {
                          "type": "object",
                          "required": ["identification_method"],
                          "properties": {
                            "identification_method": {
                              "type": "object",
                              "required": ["name"],
                              "properties": {
                                "name": { "type": "string", "minLength": 1 }
                              }
                            },
                            "authentication_methods": {
                              "type": "array",
                              "items": {
                                "type": "object",
                                "required": ["name"],
                                "properties": {
                                  "name": { "type": "string", "minLength": 1 }
                                }
                              }
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
                      "type": { "const": "authentication_method" }
                    }
                  },
                  "then": {
                    "required": ["one_of"],
                    "properties": {
                      "one_of": {
                        "type": "array",
                        "items": {
                          "type": "object",
                          "required": ["authentication_method"],
                          "properties": {
                            "authentication_method": {
                              "type": "object",
                              "required": ["name"],
                              "properties": {
                                "name": { "type": "string", "minLength": 1 }
                              }
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
                      "type": { "const": "verification" }
                    }
                  },
                  "then": {
                    "required": ["step"],
                    "properties": {
                      "step": {
                        "type": "object",
                        "required": ["name"],
                        "properties": {
                          "name": { "type": "string", "minLength": 1 }
                        }
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
        "required": ["name", "primary_authentication", "secondary_authentications"],
        "properties": {
          "name": { "type": "string", "minLength": 1 },
          "primary_authentication": {
            "type": "object",
            "properties": {
              "one_of": {
                "type": "array",
                "items": {
                  "type": "object",
                  "properties": {
                    "identification_method": {
                      "type": "object",
                      "required": ["name"],
                      "properties": {
                        "name": { "type": "string", "minLength": 1 }
                      }
                    },
                    "authentication_methods": {
                      "type": "array",
                      "items": {
                        "type": "object",
                        "required": ["name"],
                        "properties": {
                          "name": { "type": "string", "minLength": 1 }
                        }
                      }
                    }
                  }
                }
              }
            }
          },
          "secondary_authentications": {
            "type": "array",
            "items": {
              "type": "object",
              "required": ["one_of"],
              "properties": {
                "one_of": {
                  "type": "array",
                  "items": {
                    "type": "object",
                    "required": ["authentication_method"],
                    "properties": {
                      "authentication_method": {
                        "type": "object",
                        "required": ["name"],
                        "properties": {
                          "name": { "type": "string", "minLength": 1 }
                        }
                      }
                    }
                  }
                }
              }
            }
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
        "required": ["name", "identification_methods"],
        "properties": {
          "name": { "type": "string", "minLength": 1 },
          "identification_methods": {
            "type": "object",
            "required": ["one_of"],
            "properties": {
              "one_of": {
                "type": "array",
                "items": {
                  "type": "object",
                  "required": ["name", "login_flow", "signup_flow"],
                  "properties": {
                    "name": { "type": "string", "minLength": 1 },
                    "login_flow": {
                      "type": "object",
                      "required": ["name"],
                      "properties": {
                        "name": { "type": "string", "minLength": 1 }
                      }
                    },
                    "signup_flow": {
                      "type": "object",
                      "required": ["name"],
                      "properties": {
                        "name": { "type": "string", "minLength": 1 }
                      }
                    }
                  }
                }
              }
            }
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
        "required": ["name", "authentications"],
        "properties": {
          "name": { "type": "string", "minLength": 1 },
          "authentications": {
            "type": "array",
            "items": {
              "type": "object",
              "required": ["one_of"],
              "properties": {
                "one_of": {
                  "type": "array",
                  "items": {
                    "type": "object",
                    "required": ["authentication_method"],
                    "properties": {
                      "authentication_method": {
                        "type": "object",
                        "required": ["name"],
                        "properties": {
                          "name": { "type": "string", "minLength": 1 }
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}
```
