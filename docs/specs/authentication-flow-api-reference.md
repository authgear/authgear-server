- [Authentication Flow API](#authentication-flow-api)
- [State and Branching](#state-and-branching)
- [HTTP API](#http-api)
  * [HTTP response](#http-response)
    + [Successful response](#successful-response)
    + [Finish response](#finish-response)
    + [Error response](#error-response)
  * [Create an authentication flow](#create-an-authentication-flow)
  * [Pass an input to a state of an authentication flow](#pass-an-input-to-a-state-of-an-authentication-flow)
  * [Retrieve a state again](#retrieve-a-state-again)
  * [Listen for change with Websocket](#listen-for-change-with-websocket)
- [Reference on input and output](#reference-on-input-and-output)
  * [type: signup; action.type: identify](#type-signup-actiontype-identify)
    + [captcha](#captcha)
      - [captcha provider: cloudflare](#captcha-provider-cloudflare)
      - [captcha provider: recaptchav2](#captcha-provider-recaptchav2)
      - [captcha input](#captcha-input)
      - [captcha input; type: cloudflare](#captcha-input-type-cloudflare)
      - [captcha input; type: recaptchav2](#captcha-input-type-recaptchav2)
      - [captcha error](#captcha-error)
    + [identification: email](#identification-email)
    + [identification: phone](#identification-phone)
    + [identification: username](#identification-username)
    + [identification: oauth](#identification-oauth)
    + [type: signup; action.type: identify; data.type: account_linking_identification_data](#type-signup-actiontype-identify-datatype-account_linking_identification_data)
  * [type: signup; action.type: verify](#type-signup-actiontype-verify)
  * [type: signup; action.type: create_authenticator](#type-signup-actiontype-create_authenticator)
    + [captcha](#captcha-1)
    + [authentication: primary_password](#authentication-primary_password)
    + [authentication: primary_oob_otp_email](#authentication-primary_oob_otp_email)
    + [authentication: primary_oob_otp_sms](#authentication-primary_oob_otp_sms)
    + [authentication: secondary_password](#authentication-secondary_password)
    + [authentication: secondary_oob_otp_email](#authentication-secondary_oob_otp_email)
    + [authentication: secondary_oob_otp_sms](#authentication-secondary_oob_otp_sms)
    + [authentication: secondary_totp](#authentication-secondary_totp)
  * [type: signup; action.type: view_recovery_code](#type-signup-actiontype-view_recovery_code)
  * [type: signup; action.type: prompt_create_passkey](#type-signup-actiontype-prompt_create_passkey)
  * [type: login; action.type: identify](#type-login-actiontype-identify)
  * [type: login; action.type: authenticate](#type-login-actiontype-authenticate)
    + [captcha](#captcha-2)
    + [authentication: primary_password](#authentication-primary_password-1)
    + [authentication: primary_oob_otp_email](#authentication-primary_oob_otp_email-1)
    + [authentication: primary_oob_otp_sms](#authentication-primary_oob_otp_sms-1)
    + [authentication: primary_passkey](#authentication-primary_passkey)
    + [authentication: secondary_password](#authentication-secondary_password-1)
    + [authentication: secondary_oob_otp_email](#authentication-secondary_oob_otp_email-1)
    + [authentication: secondary_oob_otp_sms](#authentication-secondary_oob_otp_sms-1)
    + [authentication: secondary_totp](#authentication-secondary_totp-1)
  * [type: login; action.type: change_password](#type-login-actiontype-change_password)
  * [type: login; action.type: prompt_create_passkey](#type-login-actiontype-prompt_create_passkey)
  * [type: signup_login; action.type: identify](#type-signup_login-actiontype-identify)
  * [type: account_recovery; action.type: identify](#type-account_recovery-actiontype-identify)
    + [captcha](#captcha-3)
    + [identification: email](#identification-email-1)
    + [identification: phone](#identification-phone-1)
  * [type: account_recovery; action.type: select_destination](#type-account_recovery-actiontype-select_destination)
  * [type: account_recovery; action.type: verify_account_recovery_code](#type-account_recovery-actiontype-verify_account_recovery_code)
  * [type: account_recovery; action.type: reset_password](#type-account_recovery-actiontype-reset_password)
- [Reference on action data](#reference-on-action-data)
  * [identification_data](#identification_data)
  * [authentication_data](#authentication_data)
  * [oauth_data](#oauth_data)
  * [create_authenticator_data](#create_authenticator_data)
  * [view_recovery_code_data](#view_recovery_code_data)
  * [select_oob_otp_channels_data](#select_oob_otp_channels_data)
  * [verify_oob_otp_data](#verify_oob_otp_data)
  * [create_passkey_data](#create_passkey_data)
  * [create_totp_data](#create_totp_data)
  * [new_password_data](#new_password_data)
  * [account_recovery_identification_data](#account_recovery_identification_data)
  * [account_recovery_select_destination_data](#account_recovery_select_destination_data)
  * [account_recovery_verify_code_data](#account_recovery_verify_code_data)

# Authentication Flow API

Authentication Flow API is a HTTP API to create and run an authentication flow. It is the same API that powers that the default UI of Authgear. With Authentication Flow API, you can build your own UI while preserving the capability of running complicated authentication flow as the default UI does.

# State and Branching

An authentication flow has a constant ID that never changes. When an authentication flow is created, it has one state. A state of an authentication flow is identified by its unique state token. A particular state of authentication flow reacts to an input, and produce a new state. You keep track of the latest state token and feed an input to it to obtain another state token. In doing this you move forward in the authentication flow.

In some steps in an authentication flow, you can take any one branch to continue. For example, your project may be configured to let the end-user to sign in with email address or phone number. In this case, there are two branches. Assume the current state is **StateA**. You pass an input to **StateA** to select the email address branch, you get a new state **StateB** with the email address branch selected. If the end-user changes their mind and taps the back button, we have to allow them to select phone number. This can be done by passing an input to **StateA** to select the phone number branch, resulting in a new state **StateC**. What if the end-user changes their mind again? All you need to do is to pass an input to **StateA** to select the email address branch, and get a new state **StateB’**. **StateB** and **StateB’** are equal in their contents, only the state tokens are different.

As long as you associate the state token with the navigation, you can easily build multi-step UI.

- On the web where the [History API](https://developer.mozilla.org/en-US/docs/Web/API/History_API) is usually used to implement navigation, you can store the state ID in the `state` of a history entry.
- On iOS where [UIViewController](https://developer.apple.com/documentation/uikit/uiviewcontroller) usually represents a screen, you can store the state ID as a property of the view controller.
- On Android where Activity or Fragment usually represents a screen, you can store the state token as a property of the Activity or the Fragment, and implement onSaveInstanceState and onRestoreInstanceState to ensure the state token is persisted.

# HTTP API

## HTTP response

Authentication Flow API always returns a JSON response of the same shape.

### Successful response

```json
{
  "result": {
    "state_token": "authflowstate_blahblahblah",
    "type": "login",
    "name": "default",
    "action": {
      "type": "authenticate",
      "authentication": "primary_oob_otp_email",
      "data": {}
    }
  }
}
```

- `state_token`: The token that refers to this particular state of an authentication flow.
- `id`: The ID of the authentication flow. It is a constant for a particular authentication flow.
- `type`: The type of the authentication flow. Possible values are
  - `signup`: The flow to sign up as a new user.
  - `login`: The flow to sign in as a new user.
  - `signup_login`: This flow will either become `signup` or `login` depending on the input. If the end-user enters an existing login ID, then the flow will becomes `login`, otherwise, it is `signup`.
  - `account_recovery`: The flow to recover an account. Currently it can request a reset password link / reset password code to reset primary password.
- `name`: The name of the authentication flow. See [Create an authentication flow](#create-an-authentication-flow)
- `action`: An object containing information about the current action.
  - `action.type`: The type of step. See [Reference on input and output](#reference-on-input-and-output)
  - `action.authentication`: The taken authentication branch.
  - `action.identification`: The taken identification branch.
  - `action.data`: An object containing action-specific data. See [Reference on input and output](#reference-on-input-and-output)

### Finish response

```json
{
  "result": {
    "state_token": "authflowstate_blahblahblah",
    "type": "login",
    "name": "default",
    "action": {
      "type": "finished",
      "data": {
        "finish_redirect_uri": "https://myapp.authgear.cloud/..."
      }
    }
  }
}
```

- `action.type`: When the flow has finished, the value is `finished`.
- `action.type.data.finish_redirect_uri`: When the flow has finished, you must redirect to this URI to return the control back to Authgear.

### Error response

```json
{
  "error": {
    "name": "Unauthorized",
    "reason": "InvalidCredentials",
    "message": "invalid credentials",
    "code": 401,
    "info": {}
  }
}
```

- `reason`: You use this string to distinguish between different errors. Do NOT use `message` as it could change anytime.
- `info`: An object containing extra information about the error. It can be absent (i.e. not `null`, but absent)

## Create an authentication flow

```
POST /api/v1/authentication_flows
Content-Type: application/json

{
  "type": "login",
  "name": "default"
}
```

Create an authentication flow by specifying the `type` and the `name`. Use the name `default` to refer to the generated flow according to your project configuration. This is the same flow that the default UI runs.

## Pass an input to a state of an authentication flow

```
POST /api/v1/authentication_flows/states/input
Content-Type: application/json

{
  "state_token": "{{ STATE_TOKEN }}"
  "input": {}
}
```

```
POST /api/v1/authentication_flows/states/input
Content-Type: application/json

{
  "state_token": "{{ STATE_TOKEN }}"
  "batch_input": [{}, {}]
}
```

Pass an input to a state of an authentication flow by specifying `state_token` and `input`. See [Reference on input and output](#reference-on-input-and-output) for details on `input`.

Or if you want to pass multiple input at once, replace `input` with `batch_input`. `batch_input` must be an array with at least one element.

## Retrieve a state again

```
POST /api/v1/authentication_flows/states
Content-Type: application/json

{
  "state_token": "{{ state_token }}"
}
```

Retrieve a state by by specifying `state_token`. Typically you do not need this because the state is returned after creation or after input was passed.

## Listen for change with Websocket

```
GET /api/v1/authentication_flows/ws?flow_id={{ FLOW_ID }}
Connection: Upgrade
```

Connect to the websocket by specifying `flow_id`. The only message you will receive is `{"kind":"refresh"}`. Upon receiving the message, you should retrieve the state again with [Retrieve a state again](#retrieve-a-state-again). The `step.data` should contain updated information.

# Reference on input and output

## type: signup; action.type: identify

When you are in this step of this flow, you will see a response like the following.

```json
{
  "result": {
    "state_token": "authflowstate_5R6NM7HGGKV64538R0QEGY9RQBDM4PZD",
    "type": "signup",
    "name": "default",
    "action": {
      "type": "identify",
      "data": {
        "type": "identification_data",
        "options": [
          {
            "identification": "email"
          },
          {
            "identification": "phone",
            "captcha": {
              "providers": [
                {
                  "type": "cloudflare",
                  "alias": "cloudflare",
                  "site_key": "SITE_KEY"
                }
              ]
            }
          },
          {
            "identification": "oauth",
            "provider_type": "google",
            "alias": "google"
          },
          {
            "identification": "oauth",
            "provider_type": "wechat",
            "alias": "wechat_mobile",
            "wechat_app_type": "mobile"
          }
        ]
      }
    }
  }
}
```

### captcha

Each option may contain the key `captcha`.
If this key is present, that means selecting the option requires captcha.

The `providers` key contain a list of captcha provider you can use.
In each provider, necessary information is included for you to perform captcha in the frontend.

#### captcha provider: cloudflare

- `site_key`: The site key you use to initialize the Turnstile client-side library.

#### captcha provider: recaptchav2

- `site_key`: The site key you use to initialize the reCAPTCHA v2 client-side library.

#### captcha input

To pass captcha input, use the following input shape

```json
{
  "captcha": {
    "alias": "cloudflare",
    "type": "cloudflare",
    "response": { ... }
  }
}
```

- `captcha.alias`: The alias of the captcha provider.
- `captcha.type`: The type of the captcha provider.

Other fields are provider-specific.

#### captcha input; type: cloudflare

- `captcha.response`: The response provided by the Turnstile client-side library.

#### captcha input; type: recaptchav2

- `captcha.response`: The response provided by the reCAPTCHA v2 client-side library.

#### captcha error

When you submit an input without verifying captcha, you will receive the following error.

```json
{
  "error": {
    "name": "Forbidden",
    "reason": "CaptchaRequired",
    "message': "captcha required",
    "code": 403,
    "info": {}
  }
}
```

### identification: email

The presence of this means you can sign up with an email address.

```json
{
  "identification": "email"
}
```

The corresponding input is

```json
{
  "identification": "email",
  "login_id": "johndoe@example.com"
}
```

### identification: phone

The presence of this means you can sign up with a phone number.

```json
{
  "identification": "phone"
}
```

The corresponding input is

```json
{
  "identification": "phone",
  "login_id": "+85298765432"
}
```

Note that the phone number **MUST BE** in **E.164** format without any separators nor spaces.

### identification: username

The presence of this means you can sign up with a username.

```json
{
  "identification": "username"
}
```

The corresponding input is

```json
{
  "identification": "username",
  "login_id": "johndoe"
}
```

### identification: oauth

The presence of this means you can sign up with an OAuth provider.

```json
{
  "identification": "oauth",
  "provider_type": "google",
  "alias": "google"
}
```

- `provider_type`: The type of the OAuth provider. Possible values are
  - `google`
  - `facebook`
  - `github`
  - `linkedin`
  - `azureadv2`
  - `azureadb2c`
  - `adfs`
  - `apple`
  - `wechat`
- `alias`: The identifier of the OAuth provider. You pass this in the input.

The corresponding input is

```json
{
  "identification": "oauth",
  "alias": "google",
  "redirect_uri": "<https://example.com/oauth/redirect/google>"
}
```

- `alias`: The `alias` you see in the response. You pass this to tell Authgear which OAuth provider you choose.
- `redirect_uri`: The redirect URI after the provider has finished authenticating the end-user. This should be an URL to your website, where you must continue the authentication flow.

After passing this input, you will see a response like this

```json
{
  "result": {
    "state_token": "authflowstate_PZMX4FG4N82WGSSY0Y398YH0F9BX4FPX",
    "type": "signup",
    "name": "default",
    "action": {
      "type": "identify",
      "identification": "oauth",
      "data": {
        "type": "oauth_data",
        "alias": "google",
        "oauth_provider_type": "google",
        "oauth_authorization_url": "<https://google.com/oauth2>"
      }
    }
  }
}
```

You must redirect the end user to `oauth_authorization_url`. This is typically done by `window.location.href = {{ oauth_authorization_url }}`. Before you perform redirection, you typically need to add the query parameter `state` to `oauth_authorization_url`, so that you can resume the authentication flow.

The OAuth provider will authenticate the end-user, and then redirect back to the `redirect_uri` you provided.
You parse the callback URI and extract the `state` parameter in the query to resume your authentication flow.
You pass the URL encoded query as the next input.

Here are some examples:

```json
{
  "query": "state=mystate&code=some_authorization_code"
}
```

```json
{
  "query": "state=mystate&error=some_error&error_description=this+is+url+encoded+spaces+become+plus+sign"
}
```

### type: signup; action.type: identify; data.type: account_linking_identification_data

During identification steps in signup flow, an account linking could be triggered. In this case, you will see a response like the following:

```json
{
  "result": {
    "state_token": "authflowstate_9A2FKQJ9YWBM85255632SFQT6RQ41P5V",
    "type": "signup",
    "name": "default",
    "action": {
      "type": "identify",
      "identification": "oauth",
      "data": {
        "type": "account_linking_identification_data",
        "options": [
          {
            "identification": "email",
            "action": "login_and_link",
            "masked_display_name": "exam****@gmail.com"
          },
          {
            "identification": "oauth",
            "action": "login_and_link",
            "masked_display_name": "exam****@gmail.com",
            "provider_type": "github",
            "alias": "github"
          }
        ]
      }
    }
  }
}
```

This means account linking was triggered by the previously identified identity. You can find the followings in `action.data`:

- `options`: Contains options that you can use to continue the account linking flow. The items contains the following fields:
  - `identification`: See [type: signup; action.type: identify](#type-signup-actiontype-identify). They are having the same meaning.
  - `action`: This field specify what is going to happen when this option is selected. The only possible value in current version is `login_and_link`.
      - `login_and_link`: You need to login to one of the account in `options`. After that, the identity you have just created in previous steps will be linked to the logged in account.
  - `masked_display_name`: The display name of the identity to use. Different from signup flow, during account linking, you must use an existing identity to start account linking. The display name here is the display name of the referred identity of this option. If it is an `email`, a masked email will be displayed. If it is a `phone`, a masked phone number will be displayed. If it is a `username`, the username will be displayed without masking. If it is a `oauth` identity, the display name will be a name which you should be able to recongize the account in that provider.

  - `provider_type`: Only exist if `identification` is `oauth`. It is the type of the oauth provider. Read [identification: oauth](#identification-oauth) for details.
  - `alias`: Only exist if `identification` is `oauth`. It is the alias of the oauth provider. Read [identification: oauth](#identification-oauth) for details.

You should pass an input to choose an option to continue for the account linking, here is an example of the corresponding input:

```json
{
  "index": 0
}
```

- `index`: The index of the option you choose.

In case the option you are choosing is `"identification": "oauth"`, `redirect_uri` must be included in the input as well:

```json
{
  "index": 0,
  "redirect_uri": "http://localhost:3000/sso/oauth2/callback/github"
}
```

- `redirect_uri`: The redirect URI after the provider has finished authenticating the end-user. Read [identification: oauth](#identification-oauth) for details.

## type: signup; action.type: verify

When you are in this step, you **MAY** see a response like the following

```json
{
  "result": {
    "state_token": "authflowstate_PZMX4FG4N82WGSSY0Y398YH0F9BX4FPX",
    "type": "signup",
    "name": "default",
    "action": {
      "type": "verify",
      "data": {
        "type": "select_oob_otp_channels_data",
        "channels": [
          "sms",
          "whatsapp"
        ]
      }
    }
  }
}
```

It is asking how to deliver the OTP. You pass the following input

```json
{
  "channel": "sms"
}
```

When you are in this step, you WILL see a response like the following if the otp is a code.

```json
{
  "result": {
    "state_token": "authflowstate_blahblahblah",
    "type": "signup",
    "name": "default",
    "action": {
      "type": "verify",
      "data": {
        "type": "verify_oob_otp_data",
        "channel": "email",
        "otp_form": "code",
        "masked_claim_value": "john******@example.com",
        "code_length": 6,
        "can_resend_at": "2023-09-21T00:00:00+08:00",
        "can_check": false,
        "failed_attempt_rate_limit_exceeded": false
      }
    }
  }
}
```

If `otp_form` is `code`, a OTP will be sent to the end-user at `masked_claim_value`.

To request a resend, pass this input

```json
{
  "resend": true
}
```

After the end-user has entered the code in your UI, pass this input

```json
{
  "code": "000000"
}
```

Or you WILL see a response like the following if the otp is a link.

```json
{
  "result": {
    "state_token": "authflowstate_blahblahblah",
    "type": "signup",
    "name": "default",
    "action": {
      "type": "verify",
      "data": {
        "type": "verify_oob_otp_data",
        "channel": "email",
        "otp_form": "link",
        "websocket_url": "wss://...",
        "masked_claim_value": "john******@example.com",
        "code_length": 32,
        "can_resend_at": "2023-09-21T00:00:00+08:00",
        "can_check": false,
        "failed_attempt_rate_limit_exceeded": false
      }
    }
  }
}
```

if `otp_form` is `link`, `can_check` initially is `false` and `websocket_url` will be present in `data`.
You can connect to a websocket with this URL to listen for the event of the link being approved.

The link will be sent to the end-user at `masked_claim_value`. Clicking the link will open an approval page in the default UI.
When the user has approved the link, a websocket message of a JSON object `{"type": "refresh"}` is sent.
Upon receiving the message, you can [retrieve a state again](#retrieve-a-state-again).
The retrieved state should have `can_check=true`.
Now you can pass this input to check if the link has been approved.

```json
{
  "check": true
}
```

Alternatively, you can have a button in the UI to send the above input per tap.

To request a resend, pass this input

```json
{
  "resend": true
}
```

`can_resend_at` tells you the earliest time you can trigger resend without encountering rate limit error. Use this information to implement a cooldown counter in your UI.

`code_length` tells you the length of the OTP. It is typically relevant when `otp_form` is `code`, because it gives an hint to the end-user how long the OTP is. When `otp_form` is `link`, the OTP is included in the link, the length is not an important information to the end-user.

## type: signup; action.type: create_authenticator

When you are in this step, you will see the following response if you are setting up a primary authenticator.

```json
{
  "result": {
    "state_token": "authflowstate_DVW3H3Q9YDB3BRAA15D74V1PYGX6XYJB",
    "type": "signup",
    "name": "default",
    "action": {
      "type": "create_authenticator",
      "data": {
        "type": "create_authenticator_data",
        "options": [
          {
            "authentication": "primary_oob_otp_email",
            "otp_form": "code",
            "channels": [
              "email"
            ],
            "target": {
              "masked_display_name": "+852*****123",
              "verification_required": false
            }
          },
          {
            "authentication": "primary_password",
            "password_policy": {
              "minimum_length": 8,
              "alphabet_required": true,
              "digit_required": true,
              "history": {
                "enabled": false
              }
            }
          }
        ]
      }
    }
  }
}
```

Or this response if you are setting up 2FA.

```json
{
  "result": {
    "state_token": "authflowstate_blahblahblah",
    "type": "signup",
    "name": "default",
    "action": {
      "type": "create_authenticator",
      "data": {
        "type": "create_authenticator_data",
        "options": [
          {
            "authentication": "secondary_totp"
          },
          {
            "authentication": "secondary_password",
            "password_policy": {
              "minimum_length": 8,
              "alphabet_required": true,
              "digit_required": true
            }
          },
          {
            "authentication": "secondary_oob_otp_email",
            "otp_form": "code",
            "channels": [
              "email"
            ]
          },
          {
            "authentication": "secondary_oob_otp_sms",
            "otp_form": "code",
            "channels": [
              "sms"
            ]
          }
        ]
      }
    }
  }
}
```

### captcha

Each option may contain the key `captcha`.
If this key is present, that means selecting the option requires captcha.

See [captcha](#captcha) for details.

### authentication: primary_password

The presence of this means you can create a primary password.

```json
{
  "authentication": "primary_password",
  "password_policy": {
    "minimum_length": 8
  }
}
```

`password_policy` tells you the requirements on the password. Here is the full version of it

```json
{
  "minimum_length": 8,
  "uppercase_required": true,
  "lowercase_required": true,
  "alphabet_required": true,
  "digit_required": true,
  "symbol_required": true,
  "minimum_zxcvbn_score": 4
}
```

Any of the properties can be absent. If a property is absent, then the requirement indicated by the property DOES NOT apply.

- `minimum_length`: The minimum length of the password.
- `uppercase_required`: The password must contain at least one uppercase character.
- `lowercase_required`: The password must contain at least one lowercase character.
- `alphabet_required`: The password must contain at least one uppercase or lowercase character.
- `digit_required`: The password must contain at least one digit.
- `symbol_required`: The password must contain at least one non-alphanumeric character.
- `minimum_zxcvbn_score`: The minimum [zxcvbn](https://github.com/dropbox/zxcvbn#usage) score. Possible values are 0,1,2,3,4.

The corresponding input is

```json
{
  "authentication": "primary_password",
  "new_password": "some.very.secure.password"
}
```

### authentication: primary_oob_otp_email

The presence of this means you can create a primary Out-of-band (OOB) One-time-password (OTP) authenticator using an email address.

```json
{
  "authentication": "primary_oob_otp_email",
  "otp_form": "code",
  "channels": [
    "email"
  ]
}
```

The corresponding input is

```json
{
  "authentication": "primary_oob_otp_email",
  "channel": "email"
}
```

In case `target` is present and `target.verification_required` is false, you do not need to verify the email address. Otherwise, you **MAY** enter a state where you need to verify the email address.

### authentication: primary_oob_otp_sms

The presence of this means you can create a primary OOB OTP authenticator using phone number.

```json
{
  "authentication": "primary_oob_otp_sms",
  "otp_form": "code",
  "channels": [
    "sms"
  ]
}
```

The corresponding input is

```json
{
  "authentication": "primary_oob_otp_sms",
  "channel": "sms"
}
```

In case `target` is present and `target.verification_required` is false, you do not need to verify the phone number. Otherwise, you **MAY** enter a state where you need to verify the phone number.

### authentication: secondary_password

The presence of this means you can create a secondary password.

```json
{
  "authentication": "secondary_password",
  "password_policy": {
    "minimum_length": 8
  }
}
```

Use `password_policy` to implement your password strength validator in the UI. The corresponding input is

```json
{
  "authentication": "secondary_password",
  "new_password": "some.very.secure.password"
}
```

### authentication: secondary_oob_otp_email

The presence of this means you can create a secondary Out-of-band (OOB) One-time-password (OTP) authenticator using an email address.

```json
{
  "authentication": "secondary_oob_otp_email",
  "otp_form": "code",
  "channels": [
    "email"
  ]
}
```

The corresponding input is

```json
{
  "authentication": "secondary_oob_otp_email",
  "channel": "email",
  "target": "johndoe@example.com"
}
```

`target` can be different (and is usually different) from the email address the end-user uses to sign in.

After passing the input, you **WILL** enter a state where you need to verify the email address.

### authentication: secondary_oob_otp_sms

The presence of this means you can create a secondary OOB OTP authenticator using phone number.

```json
{
  "authentication": "secondary_oob_otp_sms",
  "otp_form": "code",
  "channels": [
    "sms"
  ]
}
```

The corresponding input is

```json
{
  "authentication": "secondary_oob_otp_sms",
  "channel": "sms",
  "target": "+85298765432"
}
```

`target` **MUST BE** in **E.164** format without any separators nor spaces. It can be different (and is usually different) from the phone number the end-user uses to sign in.

After passing the input, you **WILL** enter a state where you need to verify the phone number.

### authentication: secondary_totp

The presence of this means you can create a secondary Time-based One-time-password (TOTP) authenticator.

```json
{
  "authentication": "secondary_totp"
}
```

The corresponding input is

```json
{
  "authentication": "secondary_totp"
}
```

After passing the above input, you will see a response like this

```json
{
  "result": {
    "state_token": "authflowstate_blahblahblah",
    "type": "signup",
    "name": "default",
    "action": {
      "type": "authenticate",
      "authentication": "secondary_totp",
      "data": {
        "type": "create_totp_data",
        "secret": "SEURUM6364TM7TRL5SSGDVURZRHZY34O",
        "otpauth_uri": "otpauth://totp/johndoe@example.com?algorithm=SHA1&digits=6&issuer=http%3A%2F%2Flocalhost%3A3100&period=30&secret=SEURUM6364TM7TRL5SSGDVURZRHZY34O"
      }
    }
  }
}
```

- `secret`: It is the value the end-user need to enter if they want to set up TOTP manually.
- `otpauth_uri`: The intended usage of this URI is construct a QR code image of it. Present the QR code image to the end-user and ask them to scan the code with their TOTP authenticator application, such as Google Authenticator.

After the end-user has set up the TOTP, they have to verify once to prove that the setup is fine. Collect the TOTP from the end-user and pass this input.

```json
{
  "code": "000000"
}
```

## type: signup; action.type: view_recovery_code

When you are in this step of this flow, you will see a response like the following.

```json
{
  "result": {
    "state_token": "authflowstate_VN0JDCRTFJBPW230WXVX17RD0FKHC23B",
    "type": "signup",
    "name": "default",
    "action": {
      "type": "view_recovery_code",
      "data": {
        "type": "view_recovery_code_data",
        "recovery_codes": [
          "94X5NST2VM",
          "ZTC1BQJSMX",
          "R6NA5BS8Z0",
          "WFKDRJPHXB",
          "K6V6EWJ6NZ",
          "0XHS2ARPDM",
          "4Q0GPJTC9H",
          "7MWXG4SJFN",
          "PN5DX4B9JV",
          "NRW9NP8MXK",
          "WPJQARRRKN",
          "QDS53NPH8D",
          "SC1AVJYT9Z",
          "KY1D2EXZM2",
          "ZVG3HMEFTC",
          "0Z6YXC5W95"
        ]
      }
    }
  }
}
```

You need to present `recovery_codes` to the end-user, preferably allow them to download the recovery codes. Ask confirmation from the end-user that they have saved the recovery codes. After that pass this input

```json
{
  "confirm_recovery_code": true
}
```

## type: signup; action.type: prompt_create_passkey

When you are in this step of this flow, you will see a response like the following

```json
{
  "result": {
    "state_token": "authflowstate_blahblahblah",
    "type": "signup",
    "name": "default",
    "action": {
      "type": "prompt_create_passkey",
      "data": {
        "type": "create_passkey_data",
        "creation_options": {
          "publicKey": {
            "challenge": "muG_Yk_VyupxTyF6A9v1RO3fwBLfYxZ4N1JtVZ6OtlU",
            "rp": {
              "id": "localhost",
              "name": "redacted"
            },
            "user": {
              "id": "ZDAzZjg2YTktMDA2MS00NDFiLTk1NjQtYTk3ZmVmMzFhM2E0",
              "name": "johndoe@oursky.com",
              "displayName": "johndoe@oursky.com"
            },
            "pubKeyCredParams": [
            {
              "type": "public-key",
              "alg": -7
            },
            {
              "type": "public-key",
              "alg": -257
            }
            ],
            "timeout": 300000,
            "authenticatorSelection": {
              "residentKey": "preferred",
              "userVerification": "preferred"
            },
            "attestation": "direct",
            "extensions": {
              "credProps": true,
              "uvm": true
            }
          }
        }
      }
    }
  }
}
```

To skip creation, pass this input

```json
{
  "skip": true
}
```

To create the passkey, you need to run some javascript

```jsx
function b64ToUint6(nChr) {
  return nChr > 64 && nChr < 91
    ? nChr - 65
    : nChr > 96 && nChr < 123
    ? nChr - 71
    : nChr > 47 && nChr < 58
    ? nChr + 4
    : nChr === 43
    ? 62
    : nChr === 47
    ? 63
    : 0;
}

function base64DecToArr(sBase64, nBlocksSize) {
  var sB64Enc = sBase64.replace(/[^A-Za-z0-9\\+\\/]/g, ""),
    nInLen = sB64Enc.length,
    nOutLen = nBlocksSize
      ? Math.ceil(((nInLen * 3 + 1) >> 2) / nBlocksSize) * nBlocksSize
      : (nInLen * 3 + 1) >> 2,
    taBytes = new Uint8Array(nOutLen);

  for (
    var nMod3, nMod4, nUint24 = 0, nOutIdx = 0, nInIdx = 0;
    nInIdx < nInLen;
    nInIdx++
  ) {
    nMod4 = nInIdx & 3;
    nUint24 |= b64ToUint6(sB64Enc.charCodeAt(nInIdx)) << (6 * (3 - nMod4));
    if (nMod4 === 3 || nInLen - nInIdx === 1) {
      for (nMod3 = 0; nMod3 < 3 && nOutIdx < nOutLen; nMod3++, nOutIdx++) {
        taBytes[nOutIdx] = (nUint24 >>> ((16 >>> nMod3) & 24)) & 255;
      }
      nUint24 = 0;
    }
  }

  return taBytes;
}

function uint6ToB64(nUint6) {
  return nUint6 < 26
    ? nUint6 + 65
    : nUint6 < 52
    ? nUint6 + 71
    : nUint6 < 62
    ? nUint6 - 4
    : nUint6 === 62
    ? 43
    : nUint6 === 63
    ? 47
    : 65;
}

function base64EncArr(aBytes) {
  var nMod3 = 2,
    sB64Enc = "";

  for (var nLen = aBytes.length, nUint24 = 0, nIdx = 0; nIdx < nLen; nIdx++) {
    nMod3 = nIdx % 3;
    if (nIdx > 0 && ((nIdx * 4) / 3) % 76 === 0) {
      sB64Enc += "\\r\\n";
    }
    nUint24 |= aBytes[nIdx] << ((16 >>> nMod3) & 24);
    if (nMod3 === 2 || aBytes.length - nIdx === 1) {
      sB64Enc += String.fromCodePoint(
        uint6ToB64((nUint24 >>> 18) & 63),
        uint6ToB64((nUint24 >>> 12) & 63),
        uint6ToB64((nUint24 >>> 6) & 63),
        uint6ToB64(nUint24 & 63),
      );
      nUint24 = 0;
    }
  }

  return (
    sB64Enc.substr(0, sB64Enc.length - 2 + nMod3) +
    (nMod3 === 2 ? "" : nMod3 === 1 ? "=" : "==")
  );
}

function base64URLToBase64(base64url) {
  let base64 = base64url.replace(/-/g, "+").replace(/_/g, "/");
  if (base64.length % 4 !== 0) {
    const count = 4 - (base64.length % 4);
    base64 += "=".repeat(count);
  }
  return base64;
}

function base64ToBase64URL(base64) {
  return base64.replace(/\\+/g, "-").replace(/\\//g, "_").replace(/=/g, "");
}

function trimNewline(str) {
  return str.replace(/\\r/g, "").replace(/\\n/g, "");
}

function deserializeCreationOptions(creationOptions) {
  const base64URLChallenge = creationOptions.publicKey.challenge;
  const challenge = base64DecToArr(base64URLToBase64(base64URLChallenge));
  creationOptions.publicKey.challenge = challenge;

  const base64URLUserID = creationOptions.publicKey.user.id;
  const userID = base64DecToArr(base64URLToBase64(base64URLUserID));
  creationOptions.publicKey.user.id = userID;

  if (creationOptions.publicKey.excludeCredentials != null) {
    for (const c of creationOptions.publicKey.excludeCredentials) {
      c.id = base64DecToArr(base64URLToBase64(c.id));
    }
  }
  return creationOptions;
}

function serializeAttestationResponse(credential) {
  const response = credential.response;

  const attestationObject = trimNewline(
    base64ToBase64URL(base64EncArr(new Uint8Array(response.attestationObject))),
  );
  const clientDataJSON = trimNewline(
    base64ToBase64URL(base64EncArr(new Uint8Array(response.clientDataJSON))),
  );

  let transports = [];
  if (typeof response.getTransports === "function") {
    transports = response.getTransports();
  }

  const clientExtensionResults = credential.getClientExtensionResults();

  return {
    id: credential.id,
    rawId: credential.id,
    type: credential.type,
    response: {
      attestationObject,
      clientDataJSON,
      transports,
    },
    clientExtensionResults,
  };
}

// Basically you need to deserialize the creation_options, and
// pass it to window.navigator.credentials.create(), and then
// serialize the return value and pass it back to the API.
async function main(creationOptions) {
  creationOptions = deserializeCreationOptions(options);
  const rawResponse = await window.navigator.credentials.create(creationOptions);
  if (rawResponse instanceof PublicKeyCredential) {
    const response = serializeAttestationResponse(rawResponse);
    return response;
  }
}
```

Pass `creation_options` to `main` and then pass this input

```json
{
  "creation_response": {{ resolved return value of main }}
}
```

## type: login; action.type: identify

See [type: signup; action.type: identify](#type-signup-actiontype-identify). They are the same except that `type` is `login`.

## type: login; action.type: authenticate

When you are in this step, you will see a response like the following if you are performing primary authentication.

```json
{
  "result": {
    "state_token": "authflowstate_X0BJ22Y0P4MB6A98X75AMQ8ADVQC94MK",
    "type": "login",
    "name": "default",
    "action": {
      "type": "authenticate",
      "data": {
        "type": "authentication_data",
        "options": [
          {
            "authentication": "primary_passkey",
            "request_options": {
              "publicKey": {
                "challenge": "3PzOb9VvB54BIdrOC5b88ewjYt1wEOmKbCd0IM8FQSA",
                "timeout": 300000,
                "rpId": "localhost",
                "userVerification": "preferred",
                "allowCredentials": [],
                "extensions": {
                  "uvm": true
                }
              }
            }
          },
          {
            "authentication": "primary_oob_otp_email",
            "otp_form": "code",
            "masked_display_name": "loui*****@oursky.com",
            "channels": [
              "email"
            ]
          },
          {
            "authentication": "primary_password"
          }
        ],
        "device_token_enabled": false
      }
    }
  }
}
```

Or this response if you are performing secondary authentication.

```json
{
  "result": {
    "state_token": "authflowstate_HYQ33WWMZM2AV91VPQWJE2M0HXWT02AK",
    "type": "login",
    "name": "default",
    "action": {
      "type": "authenticate",
      "data": {
        "type": "authentication_data",
        "options": [
          {
            "authentication": "secondary_totp"
          },
          {
            "authentication": "secondary_password"
          },
          {
            "authentication": "recovery_code"
          }
        ],
        "device_token_enabled": true
      }
    }
  }
}
```

### captcha

Each option may contain the key `captcha`.
If this key is present, that means selecting the option requires captcha.

See [captcha](#captcha) for details.

### authentication: primary_password

The presence of this means you can sign in with primary password.

```json
{
  "authentication": "primary_password"
}
```

The corresponding input is

```json
{
  "authentication": "primary_password",
  "password": "12345678"
}
```

### authentication: primary_oob_otp_email

The presence of this means you can sign in by receiving a OOB OTP via email.

```json
{
  "authentication": "primary_oob_otp_email",
  "otp_form": "code",
  "masked_display_name": "john****@example.com",
  "channels": ["email"]
}
```

To reference this authentication, use its index in `options` array.
`otp_form` tells you what kind of OTP will be sent. `masked_display_name` tells you what email address the OTP will be sent to. `channels` tells you the available channels you must choose from.

The corresponding input is

```json
{
  "authentication": "primary_oob_otp_email",
  "index": 1,
  "channel": "email"
}
```

After passing the input, you **WILL** enter a state where you need to verify the OTP. [type: signup; action.type: verify](#type-signup-actiontype-verify)

### authentication: primary_oob_otp_sms

The presence of this means you can sign in by receiving a OOB OTP via phone number.

```json
{
  "authentication": "primary_oob_otp_sms",
  "otp_form": "code",
  "masked_display_name": "+8529876****",
  "channels": ["sms", "whatsapp"]
}
```

To reference this authentication, use its index in `options` array.
`otp_form` tells you what kind of OTP will be sent. `masked_display_name` tells you what phone number the OTP will be sent to. `channels` tells you the available channels you must choose from.

The corresponding input is

```json
{
  "authentication": "primary_oob_otp_sms",
  "index": 2,
  "channel": "sms"
}
```

After passing the input, you **WILL** enter a state where you need to verify the OTP. [type: signup; action.type: verify](#type-signup-actiontype-verify)

### authentication: primary_passkey

The presence of this means you can sign in with passkey.

```json
{
  "authentication": "primary_passkey",
  "request_options": {
    "publicKey": {
      "challenge": "2tVbbyG9dJ0KuM1yHlXeah1fZ6grtP4YyOIORYxIzUM",
      "timeout": 300000,
      "rpId": "localhost",
      "userVerification": "preferred",
      "allowCredentials": [
        {
          "type": "public-key",
          "id": "dFcL6B0cTujk-mONTRqsP4TXVrLWWvzWfa7oG_b36T8"
        }
      ],
      "extensions": {
        "uvm": true
      }
    }
  }
}
```

To use passkey, you need to run some javascript

```jsx
function b64ToUint6(nChr) {
  return nChr > 64 && nChr < 91
    ? nChr - 65
    : nChr > 96 && nChr < 123
    ? nChr - 71
    : nChr > 47 && nChr < 58
    ? nChr + 4
    : nChr === 43
    ? 62
    : nChr === 47
    ? 63
    : 0;
}

function base64DecToArr(sBase64, nBlocksSize) {
  var sB64Enc = sBase64.replace(/[^A-Za-z0-9\\+\\/]/g, ""),
    nInLen = sB64Enc.length,
    nOutLen = nBlocksSize
      ? Math.ceil(((nInLen * 3 + 1) >> 2) / nBlocksSize) * nBlocksSize
      : (nInLen * 3 + 1) >> 2,
    taBytes = new Uint8Array(nOutLen);

  for (
    var nMod3, nMod4, nUint24 = 0, nOutIdx = 0, nInIdx = 0;
    nInIdx < nInLen;
    nInIdx++
  ) {
    nMod4 = nInIdx & 3;
    nUint24 |= b64ToUint6(sB64Enc.charCodeAt(nInIdx)) << (6 * (3 - nMod4));
    if (nMod4 === 3 || nInLen - nInIdx === 1) {
      for (nMod3 = 0; nMod3 < 3 && nOutIdx < nOutLen; nMod3++, nOutIdx++) {
        taBytes[nOutIdx] = (nUint24 >>> ((16 >>> nMod3) & 24)) & 255;
      }
      nUint24 = 0;
    }
  }

  return taBytes;
}

function uint6ToB64(nUint6) {
  return nUint6 < 26
    ? nUint6 + 65
    : nUint6 < 52
    ? nUint6 + 71
    : nUint6 < 62
    ? nUint6 - 4
    : nUint6 === 62
    ? 43
    : nUint6 === 63
    ? 47
    : 65;
}

function base64EncArr(aBytes) {
  var nMod3 = 2,
    sB64Enc = "";

  for (var nLen = aBytes.length, nUint24 = 0, nIdx = 0; nIdx < nLen; nIdx++) {
    nMod3 = nIdx % 3;
    if (nIdx > 0 && ((nIdx * 4) / 3) % 76 === 0) {
      sB64Enc += "\\r\\n";
    }
    nUint24 |= aBytes[nIdx] << ((16 >>> nMod3) & 24);
    if (nMod3 === 2 || aBytes.length - nIdx === 1) {
      sB64Enc += String.fromCodePoint(
        uint6ToB64((nUint24 >>> 18) & 63),
        uint6ToB64((nUint24 >>> 12) & 63),
        uint6ToB64((nUint24 >>> 6) & 63),
        uint6ToB64(nUint24 & 63),
      );
      nUint24 = 0;
    }
  }

  return (
    sB64Enc.substr(0, sB64Enc.length - 2 + nMod3) +
    (nMod3 === 2 ? "" : nMod3 === 1 ? "=" : "==")
  );
}

function base64URLToBase64(base64url) {
  let base64 = base64url.replace(/-/g, "+").replace(/_/g, "/");
  if (base64.length % 4 !== 0) {
    const count = 4 - (base64.length % 4);
    base64 += "=".repeat(count);
  }
  return base64;
}

function base64ToBase64URL(base64) {
  return base64.replace(/\\+/g, "-").replace(/\\//g, "_").replace(/=/g, "");
}

function trimNewline(str) {
  return str.replace(/\\r/g, "").replace(/\\n/g, "");
}

function deserializeRequestOptions(requestOptions) {
  const base64URLChallenge = requestOptions.publicKey.challenge;
  const challenge = base64DecToArr(base64URLToBase64(base64URLChallenge));
  requestOptions.publicKey.challenge = challenge;
  if (requestOptions.publicKey.allowCredentials) {
    for (const c of requestOptions.publicKey.allowCredentials) {
      c.id = base64DecToArr(base64URLToBase64(c.id));
    }
  }
  return requestOptions;
}

function serializeAssertionResponse(credential) {
  const response = credential.response;
  const authenticatorData = trimNewline(
    base64ToBase64URL(base64EncArr(new Uint8Array(response.authenticatorData))),
  );
  const clientDataJSON = trimNewline(
    base64ToBase64URL(base64EncArr(new Uint8Array(response.clientDataJSON))),
  );
  const signature = trimNewline(
    base64ToBase64URL(base64EncArr(new Uint8Array(response.signature))),
  );
  const userHandle =
    response.userHandle == null
      ? undefined
      : trimNewline(
          base64ToBase64URL(base64EncArr(new Uint8Array(response.userHandle))),
        );
  const clientExtensionResults = credential.getClientExtensionResults();
  return {
    id: credential.id,
    rawId: credential.id,
    type: credential.type,
    response: {
      authenticatorData,
      clientDataJSON,
      signature,
      userHandle,
    },
    clientExtensionResults,
  };
}

async function main(options) {
  options = deserializeRequestOptions(options);
  const rawResponse = await window.navigator.credentials.get(options);
  if (rawResponse instanceof PublicKeyCredential) {
    const response = serializeAssertionResponse(rawResponse);
    return response
  }
}
```

Pass `request_options` to `main`, and then pass this input

```json
{
  "assertion_response": {{ resolved return value of main }}
}
```

### authentication: secondary_password

The presence of this means you can sign in with secondary password.

```json
{
  "authentication": "secondary_password"
}
```

The corresponding input is

```json
{
  "authentication": "secondary_password",
  "password": "12345678"
}
```

### authentication: secondary_oob_otp_email

The presence of this means you can sign in by receiving a OOB OTP via email.

```json
{
  "authentication": "secondary_oob_otp_email",
  "otp_form": "code",
  "masked_display_name": "john****@example.com",
  "channels": ["email"]
}
```

To reference this authentication, use its index in `options` array.

The corresponding input is

```json
{
  "authentication": "secondary_oob_otp_email",
  "index": 1,
  "channel": "email"
}
```

After passing the input, you **WILL** enter a state where you need to verify the OTP. [type: signup; action.type: verify](#type-signup-actiontype-verify)

### authentication: secondary_oob_otp_sms

The presence of this means you can sign in by receiving a OOB OTP via phone number.

```json
{
  "authentication": "secondary_oob_otp_sms",
  "otp_form": "code",
  "masked_display_name": "+8529876****",
  "channels": ["sms", "whatsapp"]
}
```

To reference this authentication, use its index in `options` array.

The corresponding input is

```json
{
  "authentication": "secondary_oob_otp_sms",
  "index": 2,
  "channel": "sms"
}
```

After passing the input, you **WILL** enter a state where you need to verify the OTP. [type: signup; action.type: verify](#type-signup-actiontype-verify)

### authentication: secondary_totp

The presence of this means you can sign in with TOTP.

```json
{
  "authentication": "secondary_totp"
}
```

The corresponding input is

```json
{
  "authentication": "secondary_totp",
  "code": "000000"
}
```

## type: login; action.type: change_password

When you are in this step, you will see a response like the following

```json
{
  "result": {
    "state_token": "authflowstate_blahblahblah",
    "type": "login",
    "name": "default",
    "action": {
      "type": "change_password",
      "data": {
        "type": "new_password_data",
        "password_policy": {
          "minimum_length": 8,
          "alphabet_required": true,
          "digit_required": true
        }
      }
    }
  }
}
```

The end-user is forced to change their password because their current password does not meet the password policy.

The corresponding input is

```json
{
  "new_password": "a.new.password.that.meet.the.password.policy"
}
```

## type: login; action.type: prompt_create_passkey

See [type: signup; action.type: prompt_create_passkey](#type-signup-actiontype-prompt_create_passkey). They are the same except that `type` is `login`.

## type: signup_login; action.type: identify

See [type: signup; action.type: identify](#type-signup-actiontype-identify). They are the same except that `type` is `signup_login`.

## type: account_recovery; action.type: identify

When you are in this step of this flow, you will see a response like the following.

```json
{
  "result": {
    "state_token": "authflowstate_5R6NM7HGGKV64538R0QEGY9RQBDM4PZD",
    "type": "signup",
    "name": "default",
    "action": {
      "type": "identify",
      "data": {
        "type": "account_recovery_identification_data",
        "options": [
          {
            "identification": "email"
          },
          {
            "identification": "phone"
          }
        ]
      }
    }
  }
}
```

### captcha

Each option may contain the key `captcha`.
If this key is present, that means selecting the option requires captcha.

See [captcha](#captcha) for details.

### identification: email

The presence of this means you can receive an account recovery code with an email address.

```json
{
  "identification": "email"
}
```

The corresponding input is

```json
{
  "identification": "email",
  "login_id": "johndoe@example.com"
}
```

### identification: phone

The presence of this means you can receive an account recovery code with a phone number.

```json
{
  "identification": "phone"
}
```

The corresponding input is

```json
{
  "identification": "phone",
  "login_id": "+85298765432"
}
```

Note that the phone number **MUST BE** in **E.164** format without any separators nor spaces.

## type: account_recovery; action.type: select_destination

When you are in this step of this flow, you will see a response like the following.

```json
{
  "result": {
    "state_token": "authflowstate_5R6NM7HGGKV64538R0QEGY9RQBDM4PZD",
    "type": "signup",
    "name": "default",
    "action": {
      "type": "select_destination",
      "data": {
        "type": "account_recovery_select_destination_data",
        "options": [
          {
            "masked_display_name": "+8529876****",
            "channel": "sms",
            "otp_form": "code"
          },
          {
            "masked_display_name": "+8529876****",
            "channel": "whatsapp",
            "otp_form": "code"
          },
          {
            "masked_display_name": "john****@example.com",
            "channel": "email",
            "otp_form": "link"
          }
        ]
      }
    }
  }
}
```

It is asking where to deliver the account recovery code.

`otp_form` can be `code` or `link`. `code` is a 6-digit otp code, and `link` is a long code which is attached to a link.

`channel` is the channel to receiving the account recovery code. Current supported channels are `sms`, `whatsapp` and `email`.

You pass the following input to indicate your choice:

```json
{
  "index": 0
}
```

`index` is the index of the option in `options` array. For `0`, it sends an sms with a 6-digit account recovery code to `+8529876****`.

## type: account_recovery; action.type: verify_account_recovery_code

When you are in this step of this flow, you will see a response like the following.

```json
{
  "result": {
    "state_token": "authflowstate_5R6NM7HGGKV64538R0QEGY9RQBDM4PZD",
    "type": "signup",
    "name": "default",
    "action": {
      "type": "verify_account_recovery_code",
      "data": {
        "type": "account_recovery_verify_code_data",
        "masked_display_name": "+8529876****",
        "channel": "sms",
        "otp_form": "code",
        "code_length": 6,
        "can_resend_at": "1970-01-01T08:00:00+08:00",
        "failed_attempt_rate_limit_exceeded": false
      }
    }
  }
}
```

In previous steps, you should have selected the destination to receive the account recovery code. You should get the code in your selected destination, such as your mailbox, and pass it with the following input:

```jsonc
{
  "account_recovery_code": "123456" // Put your account recovery code here
}
```

```jsonc
{
  "account_recovery_code": "M6CGA4WV6M9XTXNWFYFHRQDWF6VFR7K4" // OR the code in the link.
}
```

Note that `state_token` can be omitted in this step, if and only if your selected destination have `otp_form=link`. Using only the `account_recovery_code` is enough for proceeding to the next step in this case:

```jsonc
// POST /api/v1/authentication_flows/states/input
// Content-Type: application/json
{
  "input": {
    "account_recovery_code": "M6CGA4WV6M9XTXNWFYFHRQDWF6VFR7K4"
  }
}
```

## type: account_recovery; action.type: reset_password

When you are in this step of this flow, you will see a response like the following.

```json
{
  "result": {
    "state_token": "authflowstate_5R6NM7HGGKV64538R0QEGY9RQBDM4PZD",
    "type": "signup",
    "name": "default",
    "action": {
      "type": "reset_password",
      "data": {
        "type": "reset_password_data",
        "password_policy": {
          "minimum_length": 8,
          "digit_required": true,
          "history": {
            "enabled": false
          }
        }
      }
    }
  }
}
```

You can reset the password of the user in this step.

The corresponding input is

```json
{
  "new_password": "a.new.password.that.meet.the.password.policy"
}
```

# Reference on action data

This section lists all possible types of data of `result.action.data`.

Developer could identify the data type by checking the type key in `result.action.data.type`. All possible values are listed below:

## identification_data

The data contains identification options.

- `options`: The list of usable identification options.

## authentication_data

The data contains authentication options.

- `options`: The list of usable authentication options.

## oauth_data

The data contains information for initiating an oauth authentication.

- `alias`: The configured alias of the selected oauth provider.
- `oauth_provider_type`: The type of the oauth provider, such as `google`.
- `oauth_authorization_url`: The authorization url of the oauth provider.
- `wechat_app_type`: The wechat app type. Only used when provider is `wechat`.

## create_authenticator_data

The data contains options for creating new authenticator.

- `options`: The list of creatable authenticators.

## view_recovery_code_data

The data contains recovery codes of the user.

- `recovery_codes`: The recovery codes of the user.

## select_oob_otp_channels_data

The data contains usable channels of the oob authenticator, with information of the selected oob authenticator.

- `channels`: The list of usable channels for receiving the OTP.
- `masked_claim_value`: The masked phone number or email address that is going to recieve the OTP.

## verify_oob_otp_data

The data contains information about the otp verification step.

- `channel`: The selected channel.
- `otp_form`: The otp form. `code` for a 6-digit otp code, or `link` for a long otp embedded in a link.
- `websocket_url`: The websocket url for listening to the change of the otp verification status.
- `masked_claim_value`: The masked phone number or email address that is going to recieve the OTP.
- `code_length`: The length of the sent code.
- `can_resend_at`: A timestamp. Resend can be triggered after this timestamp.
- `can_check`: Used when otp_form is `link` only. If `true`, you can check the latest verification state.
- `failed_attempt_rate_limit_exceeded`: If `true`, the maximum number of fail attempt has been exceeded, therefore the OTP becomes invalid. You should request for a new OTP.

## create_passkey_data

The data contains information used for creating passkey.

- `creation_options`: The options used to create the passkey in the browser.

## create_totp_data

The data contains information of the totp.

- `secret`: The totp secret.
- `otpauth_uri`: The uri for constructing a QR code image, which can be read by authenticator apps.

## new_password_data

The data contains requirements of the new password.

- `password_policy`: The password policy requirements.

## account_recovery_identification_data

The data contains identification options for triggering account recovery flow.

- `options`: The list of usable identification options.

## account_recovery_select_destination_data

The data contains options of destinations for receiving the account recovery code.

- `options`: The list of destinations, such as phone number and emails, with the corresponding channel.

## account_recovery_verify_code_data

The data contains information about the account recovery code verification step.

- `channel`: The selected channel.
- `otp_form`: The otp form. `code` for a 6-digit otp code, or `link` for a long otp embedded in a link.
- `masked_display_name`: The masked phone number or email address that is going to recieve the code.
- `code_length`: The length of the sent code.
- `can_resend_at`: A timestamp. Resend can be triggered after this timestamp.
- `failed_attempt_rate_limit_exceeded`: If `true`, the maximum number of fail attempt has been exceeded, therefore the code becomes invalid. You should request for a new code.
