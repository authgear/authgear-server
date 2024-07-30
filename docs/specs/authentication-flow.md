- [Authentication Flow](#authentication-flow)
  * [Goals](#goals)
  * [Non-goals](#non-goals)
  * [Concepts](#concepts)
    + [Signup Flow in essence](#signup-flow-in-essence)
    + [Login Flow in essence](#login-flow-in-essence)
    + [Reauth Flow in essence](#reauth-flow-in-essence)
  * [Design](#design)
    + [Design Principles](#design-principles)
    + [Design of the configuration](#design-of-the-configuration)
    + [type: signup](#type-signup)
    + [type: login](#type-login)
    + [type: signup_login](#type-signup_login)
    + [type: reauth](#type-reauth)
    + [type: account_recovery](#type-account_recovery)
  * [Use case examples](#use-case-examples)
    + [Use case example 1: Latte](#use-case-example-1-latte)
    + [Use case example 2: Uber](#use-case-example-2-uber)
    + [Use case example 3: Google](#use-case-example-3-google)
    + [Use case example 4: The Club](#use-case-example-4-the-club)
    + [Use case example 5: Manulife MPF](#use-case-example-5-manulife-mpf)
    + [Use case example 6: Comprehensive example](#use-case-example-6-comprehensive-example)
  * [HTTP API](#http-api)
    + [The response](#the-response)
    + [Create a Authentication Flow](#create-a-authentication-flow)
    + [Execute the Authentication Flow](#execute-the-authentication-flow)
    + [Get the Authentication Flow](#get-the-authentication-flow)
  * [Mobile apps using the Default UI](#mobile-apps-using-the-default-ui)
    + [Ordinary Authentication Flow](#ordinary-authentication-flow)
    + [Authentication Flow involving OAuth](#authentication-flow-involving-oauth)
    + [Authentication Flow involving passkey](#authentication-flow-involving-passkey)
  * [Mobile apps using a Custom UI](#mobile-apps-using-a-custom-ui)
    + [Ordinary Authentication Flow](#ordinary-authentication-flow-1)
    + [Authentication Flow involving OAuth](#authentication-flow-involving-oauth-1)
    + [Authentication Flow involving passkey](#authentication-flow-involving-passkey-1)
  * [Mobile apps using native UI](#mobile-apps-using-native-ui)
    + [The start and the end of a Authentication Flow](#the-start-and-the-end-of-a-authentication-flow)
    + [Facebook Login](#facebook-login)
    + [Sign in with Apple](#sign-in-with-apple)
    + [Any other OAuth providers supported by Authgear](#any-other-oauth-providers-supported-by-authgear)
  * [Appendix](#appendix)
    + [Review on the authentication UI / UX of existing consumer apps](#review-on-the-authentication-ui--ux-of-existing-consumer-apps)
    + [Review on the design of various competitors](#review-on-the-design-of-various-competitors)
      - [Auth0](#auth0)
      - [Okta](#okta)
      - [Azure AD B2C](#azure-ad-b2c)
      - [Zitadel](#zitadel)
      - [Supertokens](#supertokens)
    + [JSON schema](#json-schema)
    + [Action Data](#action-data)
      - [identification_data](#identification_data)
      - [authentication_data](#authentication_data)
      - [oauth_data](#oauth_data)
      - [create_authenticator_data](#create_authenticator_data)
      - [view_recovery_code_data](#view_recovery_code_data)
      - [select_oob_otp_channels_data](#select_oob_otp_channels_data)
      - [verify_oob_otp_data](#verify_oob_otp_data)
      - [create_passkey_data](#create_passkey_data)
      - [create_totp_data](#create_totp_data)
      - [new_password_data](#new_password_data)
      - [account_recovery_identification_data](#account_recovery_identification_data)
      - [account_recovery_select_destination_data](#account_recovery_select_destination_data)
      - [account_recovery_verify_code_data](#account_recovery_verify_code_data)

# Authentication Flow

Authentication Flow allows the developer to specify the authentication in a declarative way.

The primary way to create and execute an Authentication Flow is via its HTTP API.

How Authentication Flow is implemented is intentionally left unspecified in this document. Instead, this document specifies the public API of Authentication Flow.

## Goals

- Support Signup Flow.
- Support Login Flow.
- Support Reauth Flow.
- Support more than 1 Signup Flow, Login Flow, or Reauth Flow.
- Support SignupLogin Flow, a flow which switches to a Signup Flow, or a Login Flow, depending on the claimed Identity.
- The Default UI is driven by generated Authentication Flows, according to the configuration of the app.
- The developer can use the HTTP API on both the Web platform, and the mobile platforms (iOS and Android).

## Non-goals

- Build a generic workflow engine.

## Concepts

This section clarifies how Authentication Flow is related to our core concepts, like User, Identity, and Authenticator.

### Signup Flow in essence
- Generate a new user ID for the User.
- Create 1 or more Identities. Later on, the User identify themselves with one of the Identities.
- Create 0 or more Authenticators. The User authenticates themselves with one of the Authenticators if needed.
- (Optional) Collect user profile (i.e. standard attributes and custom attributes)

### Login Flow in essence
- Identify the User with an Identity.
- Depending on the Identity, authenticate the User with their Authenticators.
- (Optional) Further authenticate the User with **other** Authenticators.

Suppose the User has a Email Login ID Identity `johndoe@gmail.com`, a Email OOB-OTP Authenticator `johndoe@gmail.com`, a Phone Login ID Identity `+85298765432`, a Phone OOB-OTP Authenticator `+85298765432`, a Password Authenticator, and a OAuth Identity `johndoe@gmail.com`.

If the User identifies themselves with the Email Login ID Identity `johndoe@gmail.com`, then the User can authenticate themselves with:
1. Email OOB-OTP Authenticator `johndoe@gmail.com`
2. Password Authenticator
3. Phone OOB-OTP Authenticator `+85298765432`. Note that this Authenticator is NOT associated with the identifying Identity, but it can also be used to authenticate the User.

If the User identifies themselves with the OAuth Identity `johndoe@gmail.com`, then the User DOES NOT need to authenticate themselves. This is how most other applications work.

### Reauth Flow in essence
- Authenticate the User with any Authenticators.


## Design

### Design Principles

- We want to keep the existing configuration. The Default UI is driven by on-the-fly generated Authentication Flows.
- We want to be able to fulfill the authentication flows in existing consumer apps

### Design of the configuration

- A flow has one or more `steps`.
- A step MAY optionally have an `id`.
- A step must have a `type`.
- A `type` step is specific to the kind of the flow. For example, only SignupFlow has the `type: user_profile` step.
- Some steps allow branching. Those steps have `one_of`.
- The branch of a step MAY optionally have zero or more `steps`.

### type: signup

Example:

```yaml
signup_flows:
- name: default_signup_flow
  steps:
  - name: setup_identity
    type: identify
    one_of:
    - identification: phone
      steps:
      - type: authenticate
        one_of:
        - authentication: primary_oob_otp_sms
          target_step: setup_identity
      - type: verify
        target_step: setup_identity
    - identification: email
      steps:
      - type: authenticate
        one_of:
        - authentication: primary_oob_otp_email
          target_step: setup_identity
      - type: verify
        target_step: setup_identity
  - type: authenticate
    one_of:
    - authentication: primary_password
  - name: setup_phone_2fa
    type: authenticate
    one_of:
    - authentication: secondary_oob_otp_sms
  - type: verify
    target_step: setup_phone_2fa
  # Generate and show the recovery code.
  - type: recovery_code
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

### type: login

```yaml
login_flows:
# Sign in with a phone number and OTP via SMS to any phone number the account has.
- name: phone_otp_to_any_phone
  steps:
  - type: identify
    one_of:
    - identification: phone
  - type: authenticate
    one_Of:
    - authentication: primary_oob_otp_sms

# Sign in with a phone number and OTP via SMS to the same phone number.
- name: phone_otp_to_same_phone
  steps:
  - name: identify
    type: identify
    one_of:
    - identification: phone
  - type: authenticate
    one_of:
    - authentication: primary_oob_otp_sms
      target_step: identify

# Sign in with a phone number and a password
- name: phone_password
  steps:
  - type: identify
    one_of:
    - identification: phone
  - type: authenticate
    one_of:
    - authentication: primary_password

# Sign in with a phone number, or an email address, with a password
- name: phone_email_password
  steps:
  - type: identify
    one_of:
    - identification: phone
    - identification: email
  - type: authenticate
    one_of:
    - authentication: primary_password

# Sign in with an email address, a password and a TOTP
- name: email_password_totp
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

# Sign in with an email address, a password. Perform 2FA if the end-user has configured.
- name: email_password_optional_2fa
  steps:
  - type: identify
    one_of:
    - identification: email
  - type: authenticate
    one_of:
    - authentication: primary_password
  - type: authenticate
    # If the end-user has no applicable authentication method,
    # then this step is optional.
    # optional by default is false, meaning that the step is required.
    optional: true
    one_of:
    - authentication: secondary_totp
    # If recovery_code is present, the end-user can use recovery_code.
    - authentication: recovery_code
    # If device_token is present, the end-user can use device token.
    - authentication: device_token

# Sign in with an email address, a password. Require the end-user to change password,
# if the password does not fulfill password requirements.
- name: forced_password_update
  steps:
  - type: identify
    one_of:
    - identification: email
  - type: authenticate
    name: step1
    one_of:
    - authentication: primary_password
  # Require the end-user to change the password,
  # if the actual authenticator used in target_step is a password authenticator,
  # and the password does not fulfill password requirements.
  # If the condition does not hold, this step is no-op.
  - type: change_password
    target_step: step1

# Sign in with an email address, a password. Enforce 2FA but allow the end-user to enroll to proceed.
# More explanation on enrollment_allowed is in [2FA Grace Period](./2fa-grace-period.md).
- name: email_password_optional_2fa
  steps:
  - type: identify
    one_of:
    - identification: email
  - type: authenticate
    one_of:
    - authentication: primary_password
  - type: authenticate
    # Requires user to satisfy one of the following authentication.
    optional: false
    # enrollment_allowed by default is false, meaning user with no applicable method beforehand will be blocked from proceeding.
    enrollment_allowed: true
    one_of:
    - authentication: secondary_totp
    # If recovery_code is present, the end-user can use recovery_code.
    - authentication: recovery_code
    # If device_token is present, the end-user can use device token.
    - authentication: device_token
```

### type: signup_login

Example:

```yaml
signup_login_flows:
- name: default_signup_login_flow
  steps:
  - type: identify
    one_of:
    - identification: phone
      signup_flow: default_signup_flow
      login_flow: default_login_flow
    - identification: email
      signup_flow: default_signup_flow
      login_flow: default_login_flow
```

### type: reauth

Example:

```yaml
reauth_flows:
# Re-authenticate with primary password.
- name: reauth_password
  steps:
  - type: authenticate
    one_of:
    - authentication: primary_password

# Re-authenticate with any 2nd factor, assuming that 2FA is required in signup flow.
- name: reauth_2fa
  steps:
  - type: authenticate
    one_of:
    - authentication: secondary_totp
    - authentication: secondary_sms_code

# Re-authenticate with the 1st factor AND the 2nd factor.
- name: reauth_full
  steps:
  - type: authenticate
    one_of:
    - authentication: primary_password
  - type: authenticate
    one_of:
    - authentication: secondary_totp
    - authentication: secondary_sms_code
```

### type: account_recovery

Example:

```yaml
account_recovery_flows:
# Reset password with email+link, sms+code
- name: default
  steps:
  - type: identify
    one_of:
    - identification: email
    - identification: phone
  - type: select_destination
    enumerate_destinations: false
    allowed_channels:
    - channel: email
      otp_form: link
    - channel: sms
      otp_form: code
  - type: verify_account_recovery_code
  - type: reset_password
```

## Use case examples

### Use case example 1: Latte

```yaml
signup_flows:
- name: default_signup_flow
  steps:
  - name: setup_phone
    type: identify
    one_of:
    - identification: phone
  - type: authenticate
    one_of:
    - authentication: primary_oob_otp_sms
      target_step: setup_phone
  - type: verify
    target_step: setup_phone
  - name: setup_email
    type: identify
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
- name: default_login_flow
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

### Use case example 2: Uber

```yaml
signup_flows:
- name: default_signup_flow
  steps:
  - name: setup_first_identity
    type: identify
    one_of:
    - identification: phone
      steps:
      - type: authenticate
        one_of:
        - authentication: primary_oob_otp_sms
          target_step: setup_first_identity
      - type: verify
        target_step: setup_first_identity
      - name: setup_second_identity
        type: identify
        one_of:
        - identification: email
      - type: authenticate
        one_of:
        - authentication: primary_oob_otp_email
          target_step: setup_second_identity
      - type: verify
        target_step: setup_second_identity
    - identification: email
      - type: authenticate
        one_of:
        - authentication: primary_oob_otp_email
          target_step: setup_first_identity
      - type: verify
        target_step: setup_first_identity
      - name: setup_second_identity
        type: identify
        one_of:
        - identification: phone
      - type: authenticate
        one_of:
        - authentication: primary_oob_otp_sms
          target_step: setup_second_identity
      - type: verify
        target_step: setup_second_identity
  - type: authenticate
    one_of:
    - authentication: primary_password

login_flows:
- name: default_login_flow
  steps:
  - type: identify
    one_of:
    - identification: phone
      steps:
      - type: authenticate
        one_of:
        - authentication: primary_oob_otp_sms
        - authentication: primary_password
    - identification: email
      steps:
      - type: authenticate
        one_of:
        - authentication: primary_oob_otp_email
        - authentication: primary_oob_otp_sms
        - authentication: primary_password

signup_login_flows:
- name: default_signup_login_flow
  steps:
  - type: identify
    one_of:
    - identification: phone
      login_flow: default_login_flow
      signup_flow: default_signup_flow
    - identification: email
      login_flow: default_login_flow
      signup_flow: default_signup_flow
```

### Use case example 3: Google

```yaml
signup_flows:
- name: default_signup_flow
  steps:
  - type: identify
    one_of:
    - identification: email
  - type: authenticate
    one_of:
    - authentication: primary_password

login_flows:
- name: default_login_flow
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
    - authentication: secondary_oob_otp_sms
```

### Use case example 4: The Club

```yaml
# signup_flows is omitted here because the exact signup flow is unknown.

login_flows:
- name: default_login_flow
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

### Use case example 5: Manulife MPF

```yaml
# signup_flows are omitted because it does not have public signup.

login_flows:
- name: default_login_flow
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

### Use case example 6: Comprehensive example

```yaml
signup_flows:
# The end user sign up with OAuth without password or 2FA.
# Or the end user sign up with verified email with password and 2FA.
- name: default_signup_flow
  steps:
  - name: setup_identity
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
      - type: authenticate
        one_of:
        - authentication: primary_password
      - type: authenticate
        one_of:
        - authentication: secondary_totp

login_flows:
# The end user can sign in with OAuth.
# The end user can sign in with passkey directly.
# The end user can sign in with email with OTP, password, or passkey, and with 2FA.
- name: default_login_flow
  steps:
  - type: identify
    one_of:
    - identification: oauth
    - identification: passkey
    - identification: email
      steps:
      - type: authenticate
        one_of:
        - authentication: primary_passkey
        - authentication: primary_password
          steps:
          - type: authenticate
            one_of:
            - authentication: secondary_totp
        - authentication: primary_oob_otp_email
          steps:
          - type: authenticate
            one_of:
            - authentication: secondary_totp
```

## HTTP API

### The response

The HTTP API always return a JSON response.

Example of a successful response.

```json
{
  "result": {
    "state_token": "authflowstate_blahblahblah",
    "type": "login",
    "name": "default",
    "action": {
      "type": "authenticate",
      "authentication": "primary_oob_otp_email"
      "data": {}
    }
  }
}
```

- `result.state_token`: The token that refers to a particular state of an Authentication Flow. You must keep this for the next request. This token changes every time you give an input to the flow. As a result, you can back-track by associating the token with your application navigation backstack very easily.
- `result.type`: The type of the flow. Valid values are `signup`, `login`, `signup_login`, `reauth`, and `account_recovery`.
- `result.name`: The name of the flow. Use the special value `default` to refer to the flow generated according to configuration.
- `result.action.type`: The action to be taken. Valid values are `identify`, `authenticate`, `verify`, `user_profile`, `recovery_code`, `change_password`, and `prompt_create_passkey`, and `finished`.
- `result.action.identification`: The taken branch in this action. It is only present when `result.action.type=identify`. Valid values are `email`, `phone`, and `username`.
- `result.action.authentication`: The taken branch in this action. It is only present when `result.action.type=authenticate`. Valid values are `primary_password`, `primary_oob_otp_email`, `primary_oob_otp_sms`, `secondary_password`, `secondary_totp`, `secondary_oob_otp_email`, `secondary_oob_otp_sms`, `recovery_code`.
- `result.action.data`: The data associated with the current step of the Authentication Flow. For example, if the flow is currently waiting for the User to enter a OTP, then the data contains information like resend cooldown. For list of possible values in action data, please read [the action data section](#action-data).

Example of a finished response.

```json
{
  "result": {
    "state_token": "authflowstate_blahblahblah",
    "type": "login",
    "name": "default",
    "action": {
      "type": "finished",
      "data": {
        "finish_redirect_uri": "https://example.com"
      }
    }
  }
}
```

- `result.action.type`: If the flow has finished, the value is `finished`.
- `result.action.data.finish_redirect_uri`: If `result.action.type=finished`, then this key MAY be present. You MUST redirect the end-user to this URI, where Authgear will take control and continue.

### Create a Authentication Flow

To create a Authentication Flow, specify the type and the name.

```
POST /api/v1/authentication_flows
Content-Type: application/json

{
  "type": "login",
  "name": "default"
}
```

### Execute the Authentication Flow

To execute the Authentication Flow, specify a valid input.

```
POST /api/v1/authentication_flows/states/input
Content-Type: application/json

{
  "state_token": "{{ STATE_TOKEN }}"
  "input": {}
}
```

- `state_token`: The `state_token` of a previous request.
- `input`: A JSON object specific to the current step of the flow.
- `batch_input`: An array of input. This allows you to execute the flow multiple steps in a single request. In order to do this, you must know in advance how the flow is going.

### Get the Authentication Flow

In case you want to get the Authentication Flow again.

```
POST /api/v1/authentication_flows/states
Content-Type: application/json

{
  "state_token": "{{ STATE_TOKEN }}"
}
```

## Mobile apps using the Default UI

### Ordinary Authentication Flow

[![](https://mermaid.ink/img/pako:eNqNU9tq20AQ_ZVhnxoax9BHQQImdkIopcVyKBRB2OyOpKXWjrqXtmnIv3dWliXFoiUvusyec-bsXJ6FIo0iEx5_RLQK10ZWTjaFBWilC0aZVtoAq7Y9DeXrj6ehe49uVaENM3oMdYXSpThLweLqKvEzkHzAeKNkwHcOtXGowv327rJgY81F8yQ5sThLPMZ3vCFJBtEjrPKv-LgaZQzZHL3nF5CD6-gDNTv56JPEQO2EjqYyuN3sYEnJy4dlepIzf7Dz2kNgcZJ521uFQLCcSb-faG9-o4oB4bVFuNnTLwhmD2h14ltijDNVHYDKCX1XI6yxlHHP5b0D7cxP9MBSULLCzOPinyb76ymyvm_QW6oxgb-pFmPTXicYGr7FEJ2FY5UPxUgzOG3x6OXL53w0E-g72rmVqa7D0qGvHzooSKtBKsXj8DBw5zmm9-WJcsaW9P8sU1Sn1wm2bQb96OWB5_m6lrZKUy092ctJ_1GfiXPRoGuk0bx7z0mnEHzeYCEy_tSHjheisC8Mja1m1kabQE5kpdx7PBdsmPInq0QWXMQjqN_fAcXr941o_MdO5NNh6bvdf_kLVbVgzQ?type=png)](https://mermaid.live/edit#pako:eNqNU9tq20AQ_ZVhnxoax9BHQQImdkIopcVyKBRB2OyOpKXWjrqXtmnIv3dWliXFoiUvusyec-bsXJ6FIo0iEx5_RLQK10ZWTjaFBWilC0aZVtoAq7Y9DeXrj6ehe49uVaENM3oMdYXSpThLweLqKvEzkHzAeKNkwHcOtXGowv327rJgY81F8yQ5sThLPMZ3vCFJBtEjrPKv-LgaZQzZHL3nF5CD6-gDNTv56JPEQO2EjqYyuN3sYEnJy4dlepIzf7Dz2kNgcZJ521uFQLCcSb-faG9-o4oB4bVFuNnTLwhmD2h14ltijDNVHYDKCX1XI6yxlHHP5b0D7cxP9MBSULLCzOPinyb76ymyvm_QW6oxgb-pFmPTXicYGr7FEJ2FY5UPxUgzOG3x6OXL53w0E-g72rmVqa7D0qGvHzooSKtBKsXj8DBw5zmm9-WJcsaW9P8sU1Sn1wm2bQb96OWB5_m6lrZKUy092ctJ_1GfiXPRoGuk0bx7z0mnEHzeYCEy_tSHjheisC8Mja1m1kabQE5kpdx7PBdsmPInq0QWXMQjqN_fAcXr941o_MdO5NNh6bvdf_kLVbVgzQ)

### Authentication Flow involving OAuth

[![](https://mermaid.ink/img/pako:eNqNVGFr2zAQ_SuHPq2sbWAfDekITVbKGBtxymAYgiKfbVFb50nyuq70v-9kO46TNCVfYlt579270929CEUpikg4_N2gUTjXMreySgxALa3XStfSeJjV9eFRPP96ePTg0M5yNP6I3vgiR2kPz--I8hLDKQeAq5uboBqBZDiraCU9frCYaovKPyzvpwnbra6rZ8l2xEXgMb7lDaEjaBzCLP6Jm9lORpOJ0Tl-AFm4bZynaiU3LkgM1FZoazWCu8UKJhS8fJqEX7L6X-e1h8DVQeRlbxU8weRI-uNIe_EXVeMR9i3Cl5KewOsS0KSBb4gxVueFB8pG9FWBMMdMNiUX_R5Sq_-gA5aCjBXeiNwVusspb9-vuZJddp-3FV43Vk_l6Kr2ww8a7XPvkrrYXPiW1gM48NXJ6jhHEyXLciPV4yQf2uDUXfz4Hq9Oks66D0cVPhVoMXxYdE31ZvWP9E6n0LeGIuP6lj-nk0bws3zvGn4_wDAsS_SNNbDt0C6VMNXj8TisZG_G0yOaYytjXYsZV6tYt1CQJgWpFI_SeuAexxjnG5pCm4zejzJGtXqtYF1H0I9t7LnNbgtp8rARpCMzHd0ephfiUlRoK6lT3mYvQScR_H-FiYj4Ne2mJRGJeWVoU6fMWqTakxVRJkuHl4INU_xslIi8bXAL6jfigOLF9Yto942tyLdujbbb9PU_HuvYRA?type=png)](https://mermaid.live/edit#pako:eNqNVGFr2zAQ_SuHPq2sbWAfDekITVbKGBtxymAYgiKfbVFb50nyuq70v-9kO46TNCVfYlt579270929CEUpikg4_N2gUTjXMreySgxALa3XStfSeJjV9eFRPP96ePTg0M5yNP6I3vgiR2kPz--I8hLDKQeAq5uboBqBZDiraCU9frCYaovKPyzvpwnbra6rZ8l2xEXgMb7lDaEjaBzCLP6Jm9lORpOJ0Tl-AFm4bZynaiU3LkgM1FZoazWCu8UKJhS8fJqEX7L6X-e1h8DVQeRlbxU8weRI-uNIe_EXVeMR9i3Cl5KewOsS0KSBb4gxVueFB8pG9FWBMMdMNiUX_R5Sq_-gA5aCjBXeiNwVusspb9-vuZJddp-3FV43Vk_l6Kr2ww8a7XPvkrrYXPiW1gM48NXJ6jhHEyXLciPV4yQf2uDUXfz4Hq9Oks66D0cVPhVoMXxYdE31ZvWP9E6n0LeGIuP6lj-nk0bws3zvGn4_wDAsS_SNNbDt0C6VMNXj8TisZG_G0yOaYytjXYsZV6tYt1CQJgWpFI_SeuAexxjnG5pCm4zejzJGtXqtYF1H0I9t7LnNbgtp8rARpCMzHd0ephfiUlRoK6lT3mYvQScR_H-FiYj4Ne2mJRGJeWVoU6fMWqTakxVRJkuHl4INU_xslIi8bXAL6jfigOLF9Yto942tyLdujbbb9PU_HuvYRA)

### Authentication Flow involving passkey

[![](https://mermaid.ink/img/pako:eNqNVG1r2zAQ_iuHPq2sWWAfDS2EJBtljI04ZTAMRZHPtogteXpZl5X-950UO3HsresXv0jP89yju9M9MaFzZAmz-MOjEriSvDS8yRRAy42TQrZcOVi07XgpXX0aL91bNIsSlZvQvatK5OZvhLBG8jC7vQ2aCXACk4YU3OEbg7k0KNz95u4mI7PNu-bAyQy7CjzCR94pcALeIizSb7hbnGWkVilaSy_QBpbeOt1s-c4GiRM1CvVGE_i43sJcBy_v5-GpjfyN0WsHgdko8qazCk7DfCL9dqC9_oXCO4RLi_Ch1o_gZA2o8sBXmjBGlpUDXQzo2wphhQX3NWXwDnIjf6IFkoKCFPrI48Qsayn2MT0tt3aPh5HFHp4MbWGU9UHuUboKdlI36IwUL6Xu65d0CwtLu-FYG7StVvZ_ubO-wYszXGBn_0x0VyIRQhwb7zUVHcBfVc9z402TFpt2g84bBX2nHAsa7tawTUcp6s04vUc1tTLUNVgYtNVDhAJXOXAhqKUfTtxpjOF5QwWlKvTLUYaoqBcF2zaB7vqkjlpiWXFVhpvJrVY3w2bJr9g1a9A0XOY0U56CTsZov8GMJfSZH7s2Y5l6Jqhvc2Ktc-m0YUnBa4vXjAzr9KAES5zx2IO6uXRC0fj4rvX5H6PI5-MwizPt-Q8iAKv-?type=png)](https://mermaid.live/edit#pako:eNqNVG1r2zAQ_iuHPq2sWWAfDS2EJBtljI04ZTAMRZHPtogteXpZl5X-950UO3HsresXv0jP89yju9M9MaFzZAmz-MOjEriSvDS8yRRAy42TQrZcOVi07XgpXX0aL91bNIsSlZvQvatK5OZvhLBG8jC7vQ2aCXACk4YU3OEbg7k0KNz95u4mI7PNu-bAyQy7CjzCR94pcALeIizSb7hbnGWkVilaSy_QBpbeOt1s-c4GiRM1CvVGE_i43sJcBy_v5-GpjfyN0WsHgdko8qazCk7DfCL9dqC9_oXCO4RLi_Ch1o_gZA2o8sBXmjBGlpUDXQzo2wphhQX3NWXwDnIjf6IFkoKCFPrI48Qsayn2MT0tt3aPh5HFHp4MbWGU9UHuUboKdlI36IwUL6Xu65d0CwtLu-FYG7StVvZ_ubO-wYszXGBn_0x0VyIRQhwb7zUVHcBfVc9z402TFpt2g84bBX2nHAsa7tawTUcp6s04vUc1tTLUNVgYtNVDhAJXOXAhqKUfTtxpjOF5QwWlKvTLUYaoqBcF2zaB7vqkjlpiWXFVhpvJrVY3w2bJr9g1a9A0XOY0U56CTsZov8GMJfSZH7s2Y5l6Jqhvc2Ktc-m0YUnBa4vXjAzr9KAES5zx2IO6uXRC0fj4rvX5H6PI5-MwizPt-Q8iAKv-)

## Mobile apps using a Custom UI

### Ordinary Authentication Flow

[![](https://mermaid.ink/img/pako:eNqNU21L3EAQ_ivDfqroVfBjQOE4r-UQUS5KoQTCupncLU124r5QrfjfO7sml3hXpV_yMvu87e7Mi1BUociEw8eARuGllhsr28IAdNJ6rXQnjYd51-2X8sur_dIiOE_t_eqAHfx2g9LGOivB7OIi0jOQvICGcdLjF4uVtqj8_Xp1XnCu9mv7LNlXHEUe4xNv8MggOIR5_gMf5qOKJpOjc_wCsj32Tj64qDAwk8wQKYPvyzs4pZjk7DQ-yeo_mJL2EJgx4XhivO6Dgid4KlWql0GXwepPfJZPqIJHeJ8WvjX0G7xuAE0V2YYY02DtgeqJ50I2zT-p89vVQdgPsg67VGQcy-xnnX1wKBP4_7iMF_dOf3fna_TBGhiO-m0rsQuntzwmub3JxyiefqE5DDLVtVhbdNsyQUGaCqRS3BLljnvoMd0td5XVpqbPXaaopJcEuy6Dvv1yzy292EqziY0tHZnzye1hdSRORIu2lbri6XuJOoXg9RYLkfFnhbUMjS9EYV4ZGrqKWctKe7Iiq2Xj8ERwYMqfjRKZtwEHUD_BOxRP4E-i8R-TyPXb2Kfpf_0LAHlhuA?type=png)](https://mermaid.live/edit#pako:eNqNU21L3EAQ_ivDfqroVfBjQOE4r-UQUS5KoQTCupncLU124r5QrfjfO7sml3hXpV_yMvu87e7Mi1BUociEw8eARuGllhsr28IAdNJ6rXQnjYd51-2X8sur_dIiOE_t_eqAHfx2g9LGOivB7OIi0jOQvICGcdLjF4uVtqj8_Xp1XnCu9mv7LNlXHEUe4xNv8MggOIR5_gMf5qOKJpOjc_wCsj32Tj64qDAwk8wQKYPvyzs4pZjk7DQ-yeo_mJL2EJgx4XhivO6Dgid4KlWql0GXwepPfJZPqIJHeJ8WvjX0G7xuAE0V2YYY02DtgeqJ50I2zT-p89vVQdgPsg67VGQcy-xnnX1wKBP4_7iMF_dOf3fna_TBGhiO-m0rsQuntzwmub3JxyiefqE5DDLVtVhbdNsyQUGaCqRS3BLljnvoMd0td5XVpqbPXaaopJcEuy6Dvv1yzy292EqziY0tHZnzye1hdSRORIu2lbri6XuJOoXg9RYLkfFnhbUMjS9EYV4ZGrqKWctKe7Iiq2Xj8ERwYMqfjRKZtwEHUD_BOxRP4E-i8R-TyPXb2Kfpf_0LAHlhuA)

### Authentication Flow involving OAuth

This is not possible if the custom UI is a single-page application.

### Authentication Flow involving passkey

Not documented at the moment.

## Mobile apps using native UI

> This use case is not intended to be supported at the moment.
> Instead, we want to put resources on making using the Default UI or using Custom UI solve majority of the use cases.

When the mobile apps want to use native UI, they can consume the HTTP API directly.
This implies a user agent is **NOT** involved in this use case.

Some authentication experience that involves third parties, however, require the mobile apps to cooperate.
This section outlines each cooperation in details.

### The start and the end of a Authentication Flow

RFC 6749 defines [Resource Owner Password Credentials Grant](https://datatracker.ietf.org/doc/html/rfc6749#section-4.3).
It requires the client to collect the username and the password from the end-user, and then call the token endpoint with the collected credentials.
If we compare this grant type with Authentication Flow, we can see that Authentication Flow is doing the same thing,
except that Authentication Flow is flexible enough to handle all kinds of authentication experience, not limited to username with password.

It follows naturally that we need to support a new grant type in the token endpoint to facilitate the integration between Authentication Flow and OAuth2, without a user agent.

When an Authentication Flow ends, it includes a `code` in the data.
This `code` is one-time use, just like an ordinary `authorization_code`.
The intended usage of this `code` is to call the token endpoint with grant type [`urn:authgear:params:oauth:grant-type:authorization_code`](./oidc.md#urnauthgearparamsoauthgrant-typeauthorization_code)

### Facebook Login

Mobile apps that use Facebook Login typically use Facebook SDK.

> In addition to the default short-lived access token, Facebook also supports Session Info Access Token.
> But getting a Session Info Access Token requires a long-lived token.
> A long-lived token requires app secret.
> Therefore, a long-lived token **CANNOT** be obtained in the mobile app.
> So the approach outlined here requires the mobile app to send the short-lived access token to Authgear.

[![](https://mermaid.ink/img/pako:eNp9UstqwzAQ_JVFZ-cHfAgE0pTQFkpDL8WXrbSORWxJlVYJJeTfK79SO6W5CHt3dnZmmbOQVpHIRaCvSEbSWuPeY1MYAIeetdQODcPKudvSBiV9WnvYrZ_-oCNXe0L_38ijR1etXrdtPzHDYrmc0uXwHuhagGe716aFTiCwaGfSbA5vxNEbCJX1vKj1kRSglBQCsD2Qme4YdeWwI5NgXYFM0oesLWxqe4KtcZHhpLmC1LtPO9DN9I_eehP3KNILJbGsIAby4LwtdU1ToyPX4PaqfrA8HcsAjUpPSK48hVhz1raOvYTZWh3giLVWcwu390QIsZsqYz2_k-kPlbY4awKJTDTkG9QqxejckhYioRsqRJ4-FZWY1BSiMJcEjU4h04PSbL3IS6wDZQIj2923kSJnH2kEDVG8olKEPqz9_aeO5KXPbxfjyw-yHfui?type=png)](https://mermaid.live/edit#pako:eNp9UstqwzAQ_JVFZ-cHfAgE0pTQFkpDL8WXrbSORWxJlVYJJeTfK79SO6W5CHt3dnZmmbOQVpHIRaCvSEbSWuPeY1MYAIeetdQODcPKudvSBiV9WnvYrZ_-oCNXe0L_38ijR1etXrdtPzHDYrmc0uXwHuhagGe716aFTiCwaGfSbA5vxNEbCJX1vKj1kRSglBQCsD2Qme4YdeWwI5NgXYFM0oesLWxqe4KtcZHhpLmC1LtPO9DN9I_eehP3KNILJbGsIAby4LwtdU1ToyPX4PaqfrA8HcsAjUpPSK48hVhz1raOvYTZWh3giLVWcwu390QIsZsqYz2_k-kPlbY4awKJTDTkG9QqxejckhYioRsqRJ4-FZWY1BSiMJcEjU4h04PSbL3IS6wDZQIj2923kSJnH2kEDVG8olKEPqz9_aeO5KXPbxfjyw-yHfui)

### Sign in with Apple

iOS apps that use Sign in with Apple typically use [ASAuthorizationController](https://developer.apple.com/documentation/authenticationservices/asauthorizationcontroller).

[![](https://mermaid.ink/img/pako:eNqNVNtuEzEQ_ZWRn9N8wIpGQg2gPiCkBnhAkTbGns1aOLaxZ6lC1X9n7L00t6a8eb1njs-cuTwJ5TWKSiT83aFTuDRyG-Vu7QCCjGSUCdIRvA_h9Mp8WZ2hOmq3KOOFaIv5kg9ws1hMwApWaFExmUZHpjFKkvEO1kLmkNr4VDu--oNrUeKHOLgpLCFU8IDURQey_MskA8VH6x8hYgreJYRHQy1oSRI2znOam8wmFTNLwjE75_lssSHwTU_-CR3GjJCg4j6QZ29Cy09Yu4eEqosIUTrtd7BJxMDN_N3PuFgh57zKgnw0f4ueYsH98iHbnGheRAB5YM2jJDCu_8w6N_OsSOOpxtFBdr-Cb5xZjrj6Vp_7CeTOO4reWiy1YrJzRy-T3kUspZL2_xz8jtE0-9EduL19k3deoNezP-wfp18pvnGhu5z9-Zt9_9H-q_-F3EzM-WaIPPZT47EhB6NQXIlm2_a2TNoHb3IF75fcDPz0adoHLC-93xfKYgVLD3l-QbXSbYcuV9awwtro2XjkRo1IMzhSXOfAWUm1tN9gr8WhEyaVQztEbHia2pomh6RSmFI96b42nKkr4KbLg3NtUtdOzMQO404azWvpKROvBeN3vAIqPmpsZGcp74NnhnaBpwU_aEM-iqqRNuFMcKJ-tXdKVBQ7HEHDaptQvJp-eP_yjYXkc78Py1p8_gcFvtXb?type=png)](https://mermaid.live/edit#pako:eNqNVNtuEzEQ_ZWRn9N8wIpGQg2gPiCkBnhAkTbGns1aOLaxZ6lC1X9n7L00t6a8eb1njs-cuTwJ5TWKSiT83aFTuDRyG-Vu7QCCjGSUCdIRvA_h9Mp8WZ2hOmq3KOOFaIv5kg9ws1hMwApWaFExmUZHpjFKkvEO1kLmkNr4VDu--oNrUeKHOLgpLCFU8IDURQey_MskA8VH6x8hYgreJYRHQy1oSRI2znOam8wmFTNLwjE75_lssSHwTU_-CR3GjJCg4j6QZ29Cy09Yu4eEqosIUTrtd7BJxMDN_N3PuFgh57zKgnw0f4ueYsH98iHbnGheRAB5YM2jJDCu_8w6N_OsSOOpxtFBdr-Cb5xZjrj6Vp_7CeTOO4reWiy1YrJzRy-T3kUspZL2_xz8jtE0-9EduL19k3deoNezP-wfp18pvnGhu5z9-Zt9_9H-q_-F3EzM-WaIPPZT47EhB6NQXIlm2_a2TNoHb3IF75fcDPz0adoHLC-93xfKYgVLD3l-QbXSbYcuV9awwtro2XjkRo1IMzhSXOfAWUm1tN9gr8WhEyaVQztEbHia2pomh6RSmFI96b42nKkr4KbLg3NtUtdOzMQO404azWvpKROvBeN3vAIqPmpsZGcp74NnhnaBpwU_aEM-iqqRNuFMcKJ-tXdKVBQ7HEHDaptQvJp-eP_yjYXkc78Py1p8_gcFvtXb)

### Any other OAuth providers supported by Authgear

These providers include Google, GitHub, etc.

Prerequisite
- The developer visit the portal of the provider and set a `redirect_uri` of custom URI scheme, e.g. `com.myapp://callback`. This redirect approach is recommended in RFC 8252 Appendix B.

[![](https://mermaid.ink/img/pako:eNqVVF1PGzEQ_CsrP7USkPdTSRXR9omKioAqVZHCxt5LrN7Zrj-gFPHfu_Yll9wF0fLmj_HMeLzrJyGtIlGJQL8SGUmfNK49tgsD4NBHLbVDE2Hm3HjpNpCfrcnEI2yKmzWhH69f5Y1v3t5rRWWTSeF0Ou0PVDCnhmQEBpioay0xamtgISwyZGl4ek8LUc5uz8BpYXCugmuKyRvAspcJtse_NPYBPAVnTSCIG4xAv13WQQPauBTB1iySk-jIUbIQRtpd21geN1QXYNG6YK7oE3OwFGR31us_nd7t9eXZh5Wf3vAOj6FNIYK0JqJmd5GJMC80mj0utWIXiu0p7dnSMnl9Lm171j5Wk4nEplmh_Jk9KBq72sXXP0QFV47My47gQccNZ5MYDJjR8G42_06rYVhzCiGfmMAFu7btDa7C-yzWixTJwVPm5Dv3RbooRFvGBQfu4M37SxxVQwnZ6_WmpDxSOHBJvcooliPGwUJXKAdZ7U1b6BJHftl96B9zPZwPglyWEhmmMS6_t1L9R601NuwvvX08XYMhUqRer43D1uIye7k3uh4oBXKXbd39q8FCkpILpU7Nq722MOJEtORb1Iq_mKdMuxCMb7nPKh4qqjE1MTfdM0OTU3yFz0pH60VVYxPoRHBqdv5opKi432gH2n5TPYo_mB_W7udUSL52f1v54p7_At9Ev6c?type=png)](https://mermaid.live/edit#pako:eNqVVF1PGzEQ_CsrP7USkPdTSRXR9omKioAqVZHCxt5LrN7Zrj-gFPHfu_Yll9wF0fLmj_HMeLzrJyGtIlGJQL8SGUmfNK49tgsD4NBHLbVDE2Hm3HjpNpCfrcnEI2yKmzWhH69f5Y1v3t5rRWWTSeF0Ou0PVDCnhmQEBpioay0xamtgISwyZGl4ek8LUc5uz8BpYXCugmuKyRvAspcJtse_NPYBPAVnTSCIG4xAv13WQQPauBTB1iySk-jIUbIQRtpd21geN1QXYNG6YK7oE3OwFGR31us_nd7t9eXZh5Wf3vAOj6FNIYK0JqJmd5GJMC80mj0utWIXiu0p7dnSMnl9Lm171j5Wk4nEplmh_Jk9KBq72sXXP0QFV47My47gQccNZ5MYDJjR8G42_06rYVhzCiGfmMAFu7btDa7C-yzWixTJwVPm5Dv3RbooRFvGBQfu4M37SxxVQwnZ6_WmpDxSOHBJvcooliPGwUJXKAdZ7U1b6BJHftl96B9zPZwPglyWEhmmMS6_t1L9R601NuwvvX08XYMhUqRer43D1uIye7k3uh4oBXKXbd39q8FCkpILpU7Nq722MOJEtORb1Iq_mKdMuxCMb7nPKh4qqjE1MTfdM0OTU3yFz0pH60VVYxPoRHBqdv5opKi432gH2n5TPYo_mB_W7udUSL52f1v54p7_At9Ev6c)

## Appendix

### Review on the authentication UI / UX of existing consumer apps

This notion records the authentication flows of existing consumer apps in Hong Kong.
https://www.notion.so/oursky/Common-Signup-Login-Flows-f62e48724dc041d29aa0a77ec1dae806

Some important observations drawn from this review.
- Most consumer apps do not support 2FA.
- The authentication method is not necessarily tied to the identification method. For example, in The Club app, user can first enter their email address, and then receive a Phone OTP to sign in.

### Review on the design of various competitors

#### Auth0

Auth0 offers Triggers, Actions and Flows. https://auth0.com/docs/customize/actions/flows-and-triggers Auth0 does not support fully customized flows. Instead, it defines some Triggers, and allow the developer to write their own Actions to build Flows.

#### Okta

Okta is based on Workflows. But the Workflows it offer are mainly for building business workflows, instead of customizing the authentication flow. In the documentation, it only documents how to customize a step in the authentication flow. https://help.okta.com/wf/en-us/Content/Topics/Workflows/connector-builder/authentication-custom.htm

#### Azure AD B2C

Azure AD B2C allows customization via custom policy. Custom policy is configured by configuration files. The custom policy has a few key concepts.

- Claims are the foundation of a custom policy.
- User Journey defines how the user authenticates themselves.
- A User Journey contains several Orchestration Steps.
- Each Orchestration Step can be executed conditionally.
- An Orchestration Step must refer to a Technical profile.
- A Technical Profile defines its input Claims and output Claims.

Therefore, the end-user goes through the User Journey, with more and more Claims being collected in each Orchestration Step.

https://learn.microsoft.com/en-us/azure/active-directory-b2c/custom-policy-overview

#### Zitadel

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

#### Supertokens

Supertokens requires the developer to host a backend server to interactive with the Core Driver Interface (CDI) https://app.swaggerhub.com/apis/supertokens/CDI/2.21.1 The CDI is not very flexible. For example, it only supports some pre-defined recipe like EmailPassword Recipe, Passwordless Recipe.

### JSON schema

Refer to the source code.
