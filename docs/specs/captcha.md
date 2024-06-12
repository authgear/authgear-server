# Captcha

- [Captcha](#captcha)
  * [Old configuration](#old-configuration)
    + [authgear.yaml](#authgearyaml)
    + [authgear.secrets.yaml](#authgearsecretsyaml)
  * [New Configuration](#new-configuration)
    + [authgear.yaml](#authgearyaml-1)
      - [`type: cloudflare`](#type-cloudflare)
    + [authgear.secrets.yaml](#authgearsecretsyaml-1)
      - [`type: cloudflare`](#type-cloudflare-1)
  * [Authentication Flow](#authentication-flow)
    + [Captcha in Authentication Flow configuration](#captcha-in-authentication-flow-configuration)
    + [Default behavior](#default-behavior)
    + [Captcha in Authentication Flow API](#captcha-in-authentication-flow-api)
    + [Advanced use case: Require Captcha at a specific branch, instead of right after flow creation](#advanced-use-case-require-captcha-at-a-specific-branch-instead-of-right-after-flow-creation)
    + [Future use cases](#future-use-cases)
      - [Future use case: Use different Captcha providers in different flows](#future-use-case-use-different-captcha-providers-in-different-flows)
      - [Future use case: Allow fallback in providers](#future-use-case-allow-fallback-in-providers)
      - [Future use case: Fail-open instead of Fail-close](#future-use-case-fail-open-instead-of-fail-close)
      - [Future use case: Use risk assessment to determine whether captcha is required](#future-use-case-use-risk-assessment-to-determine-whether-captcha-is-required)
  * [Audit log](#audit-log)
  * [Study on Captcha providers](#study-on-captcha-providers)
    + [Geetest v4](#geetest-v4)
    + [Geetest v3](#geetest-v3)
    + [hCaptcha](#hcaptcha)
    + [reCAPTCHA v2](#recaptcha-v2)
    + [reCAPTCHA v3](#recaptcha-v3)
    + [reCAPTCHA Enterprise](#recaptcha-enterprise)
    + [Cloudflare Turnstile](#cloudflare-turnstile)
    + [Arkose Labs Bot Manager](#arkose-labs-bot-manager)
    + [Tencent Captcha](#tencent-captcha)

## Old configuration

This section documents the old configuration.

> The old configuration **IS NOT** used by Authentication Flow.
> To configure Captcha providers for an Authentication Flow, the new configuration must be used.

### authgear.yaml

```
captcha:
  provider: cloudflare
```

- `captcha.provider`: The only possible value is `cloudflare`. The default is null.

### authgear.secrets.yaml

```yaml
- data:
    secret: YOUR_TURNSTILE_SECRET_KEY
  key: captcha.cloudflare
```

## New Configuration

The section documents the new configuration.

### authgear.yaml

```yaml
captcha:
  enabled: true
  providers:
  - type: cloudflare
    alias: cloudflare
```

- `captcha.enabled`: Boolean. If it is true, the new configuration is used.
- `captcha.providers`: An array of Captcha provider configuration. The actual shape depends on the `type` property.
- `captcha.providers.type`: Required. The type of the Captcha provider. Valid values are `cloudflare` and `recaptchav2`.
- `captcha.providers.alias`: Required. The unique identifier to the Captcha provider. It is used in other parts of the configuration to refer to this particular Captcha provider. For example, the project can configured a number of Captcha providers, but only uses one of them in a particular Authentication Flow.

Other fields are specific to `type`.

#### `type: cloudflare`

There is no specific fields.

### authgear.secrets.yaml

```yaml
- data:
    items:
    - type: cloudflare
      alias: cloudflare
      secret: TURNSTILE_SECRET_KEY
  key: captcha.providers
```

- `items.type`: Required. It is the same as `captcha.providers.type`.
- `items.alias`: Required. It is the same as `captcha.providers.alias`.

Other fields are specific to `type`.

#### `type: cloudflare`

- `secret_key`: Required. The Cloudflare Turnstile secret key.

## Authentication Flow

This section specifies how Captcha works in a Authentication Flow.

### Captcha in Authentication Flow configuration

Captcha is supported in the following flow types:

- `signup`
- `promote`
- `login`
- `signup_login`
- `reauth`
- `account_recovery`

Captcha is supported only in the following step types:

- `identify` in `signup`, `promote`, `login`, `signup_login`, and `account_recovery`.
- `create_authenticator` in `signup` and `promote`.
- `authenticate` in `login` and `reauth`.

We can see that all supported step types have branches.
To specify Captcha is required, add `captcha.required=true` to the branch.

For example,

```
authentication_flow:
  signup_flows:
  - name: default
    steps:
    - type: identify
      one_of:
      - identification: email
        # Identify with email requires captcha.
        captcha:
          required: true
    - type: authenticate
      one_of:
      - authentication: primary_password
      - authentication: primary_oob_otp_email
```

### Default behavior

When Captcha is enabled and there is at least one configured provider,
all the branches of the first step (that is, the `identify` step) of the generated flows will have `captcha.required=true`.

This means whether Captcha is enabled, every flow requires captcha at the beginning of the flow.

### Captcha in Authentication Flow API

Please refer to [captcha](./authentication-flow-api-reference.md#captcha).

### Advanced use case: Require Captcha at a specific branch, instead of right after flow creation

Suppose Project A configures email login with password or OTP. The developer may only want to require captcha if OTP is used, to reduce friction.

This can be achieved by customizing the flow.

```
authentication_flow:
  signup_flows:
  - name: default
    steps:
    - type: identify
      one_of:
      - identification: email
    - type: authenticate
      one_of:
      - authentication: primary_password
      # Captcha is required BEFORE selecting this branch.
      # That is, before the OTP is sent.
      - authentication: primary_oob_otp_email
        captcha:
          required: true
```

### Future use cases

This section documents future use cases and their imaginary configuration.

#### Future use case: Use different Captcha providers in different flows

Given that different Captcha providers have different support for platforms, the developer may want to use different Captcha providers in different flows.

For example, reCAPTCHA v3 is only available in web, while hCaptcha supports mobile platforms.
The developer may want to use reCAPTCHA v3 in web, and use hCaptcha in mobile.

The [new configuration](#new-configuration) supports configuring multiple Captcha providers already, even multiple providers of the same type.

What is missing is a way to specify which providers can be used in a flow.
This could be specified in the following imaginary configuration.

```
authentication_flow:
  signup_flows:
  - name: web
    captcha:
      providers:
      - alias: recaptchav3
  - name: mobile
    captcha:
      providers:
      - alias: hcaptcha
```

#### Future use case: Allow fallback in providers

reCAPTCHA v3 is a score-based provider. The developer may want to

1. Configure the minimum score.
2. Fallback to another provider in case the score is less than the minimum.

This could be specified in the following imaginary configuration.

```
authentication_flow:
  signup_flows:
  - name: default
    captcha:
      providers:
      - alias: recaptchav3
        # The score be must >= 0.5 in order to be considered as passed.
        minimum_score_inclusive: 0.5
      - alias: recaptchav2
```

In this configuration, reCAPTCHA v2 can actually be used before reCAPTCHA v3. It is up to the UI to implement the fallback mechanism.

#### Future use case: Fail-open instead of Fail-close

By default, when captcha verification fails, the end-user cannot proceed.
The developer may want the end-user to proceed instead.

This could be specified in the following imaginary configuration.

```
authentication_flow:
  signup_flows:
  - name: default
    captcha:
      providers:
      # fail_open is false by default.
      - alias: recaptchav3
        minimum_score_inclusive: 0.5
      - alias: recaptchav2
        fail_open: true
```

With the above configuration, reCAPTCHA v3 is not fail-open. If the score of reCAPTCHA v3 is low than the minimum, reCAPTCHA v2 can be tried.
If that fails too, it is opened, the flow can proceed.

#### Future use case: Use risk assessment to determine whether captcha is required

This use case is very advanced, and potentially related to adaptive MFA.
They both depend on another feature, namely risk assessment, to determine the risk level, and
finally decide which protective measures, like captcha, MFA, are required.

This is out-of-scope in this document.

## Audit log

When a verification failure is detected, the event [captcha.failed](./event.md#captchafailed) is logged.

## Study on Captcha providers

### Geetest v4

Geetest v4 is a challenge-based Captcha provider.

(https://mermaid.live/edit#pako:eNp1kktrwzAQhP-K2GsTeulJh0Bf9BQoSU_FF2GNE4Etuet1-gj5713HtR1KrJOk_XY0y-hIefIgSw0-WsQcT8Ht2FVZNLoey4Aoy9Xq5r6V_Q6OrckZTmCcXhRl-uzBoaxo32PNQ5JXToJcQoobVQ8Mb8Cc-L_4C_CGRg531oQYJLgy_KCHxpJyy0F6ggyjaUuZd3vVxazCxRiXpm4Pinqd-oqn6akBmlMc7Ud8iWkENS2oAlcueA3g2DVkJHtUyMjq1qNwnRBl8aRoW3fqzz5IYrKFKxssSGNI2--YkxVuMUB_IY5U7eJ7StMZZ5F1n_z5A5x-AdWRtBU)

Sequence diagram by Geetest: https://docs.geetest.com/BehaviorVerification/overview/communicationProcess/

Server API reference: https://docs.geetest.com/BehaviorVerification/deploy/server/

### Geetest v3

Geetest v3 is a challenge-based Captcha provider. Note that with Geetest v3, the process has to be initiated by the server, as opposed to other challenge-based Captcha providers.

(https://mermaid.live/edit#pako:eNqFkt9LwzAQx_-VkFc3h_gW2MBf-CSI-iQFOdPv1kCb1OQynWP_u-lqW8Fuy1PSfu6Tu8ttpXY5pJIBHxFW49bQylOVWZHWTWlgebpYnF1FLlYgr4T2IIag9GFZus8W7H436D3wgsDrSyVm64vZ-6amEN4CE8dwXhd1G9FTKWQ62FtsHqLWCOG43GNlAsOflHZg2oRY8j9rW6YS144fvWNoNs4-pYYYj1zkxCTm45KhQX8SM9awodJ8YySr7q4BGhcO2Y9mdVBw8CUSmirByWZ14CFrX4HFF6f3Qi1mAt47n1k5kRV8RSZPE7VtQjPJBSpkUqVtjiU1SpnZXUJj3dxzlxt2XqollQETmebKPW-slop9RAf9TmVP1WRfnRvO2Ese2lHeT_TuB13s-_g)

Server API reference: https://docs.geetest.com/captcha/apirefer/api/server/#SDK-GeeTest-Server-Communication-API-Reference

### hCaptcha

hCaptcha comes in 3 flavors:

- Checkbox
- [Invisible](https://docs.hcaptcha.com/invisible#invisible-vs-passive)
- [Passive](https://docs.hcaptcha.com/invisible#invisible-vs-passive)

Checkbox and Invisible are challenge-based, while Passive is not.

hCaptcha supports major mobile platforms with [SDKs](https://docs.hcaptcha.com/mobile_app_sdks).

Sequence diagram: https://docs.hcaptcha.com/#request-flow

### reCAPTCHA v2

reCAPTCHA v2 is a challenge-based Captcha provider.

reCAPTCHA v2 comes in 2 flavors:

- [Checkbox](https://developers.google.com/recaptcha/docs/display)
- [Invisible](https://developers.google.com/recaptcha/docs/invisible)

reCAPTCHA v2 does not support iOS. It supports [Android](https://developer.android.com/privacy-and-security/safetynet/recaptcha) though.

### reCAPTCHA v3

reCAPTCHA v3 is NOT challenge-based Captcha provider.

It does not support mobile platforms out-of-the-box.

It is worth to note that it has [Actions](https://developers.google.com/recaptcha/docs/v3#actions).

### reCAPTCHA Enterprise

reCAPTCHA Enterprise is reCAPTCHA v2 and reCAPTCHA v3 offered in a package. Additionally, the hightest pricing tier offers mobile platforms support.

### Cloudflare Turnstile

Cloudflare Turnstile is a challenge-based Captcha provider.

Cloudflare Turnstile comes in 3 flavors:

- [Managed](https://developers.cloudflare.com/turnstile/concepts/widget-types/#managed-recommended). Let Cloudflare to decide whether to show checkbox.
- [Non-interactive](https://developers.cloudflare.com/turnstile/concepts/widget-types/#non-interactive). It is just a badge.
- [Invisible](https://developers.cloudflare.com/turnstile/concepts/widget-types/#invisible). The end-user does not see anything visible.

Cloudflare Turnstile does not support mobile platforms natively.

### Arkose Labs Bot Manager

Arkose Labs Bot Manager CAN BE a challenge-basd Captcha provider.

It supports prompting interactive challenge if necessary.

It supports mobile platforms with webview (packaged as a SDK).

### Tencent Captcha

Tencent Captcha is a challenge-based Captcha provider.
