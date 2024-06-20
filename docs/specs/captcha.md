# Risk assessment and Captcha

- [Risk assessment and Captcha](#risk-assessment-and-captcha)
  * [Old configuration](#old-configuration)
  * [New Configuration](#new-configuration)
    + [authgear.yaml](#authgearyaml)
    + [authgear.secrets.yaml](#authgearsecretsyaml)
  * [Authentication Flow](#authentication-flow)
    + [Risk assessment and Captcha in Authentication Flow configuration](#risk-assessment-and-captcha-in-authentication-flow-configuration)
    + [Behavior of generated flows](#behavior-of-generated-flows)
    + [Risk assessment and Captcha in Authentication Flow API](#risk-assessment-and-captcha-in-authentication-flow-api)
    + [Advanced use case: Require Captcha at a specific branch only](#advanced-use-case-require-captcha-at-a-specific-branch-only)
    + [Advanced use case: Use different Captcha providers in different branches](#advanced-use-case-use-different-captcha-providers-in-different-branches)
    + [Advanced use case: Use fail-open instead of fail-close](#advanced-use-case-use-fail-open-instead-of-fail-close)
    + [Advanced use case: Allow internal staff to bypass Captcha](#advanced-use-case-allow-internal-staff-to-bypass-captcha)
    + [Advanced use case: Require Captcha only when risk level is high](#advanced-use-case-require-captcha-only-when-risk-level-is-high)
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

> The old configuration **IS NOT** used by Authentication Flow.
> To configure Captcha providers for an Authentication Flow, the new configuration must be used.

See [here](./captcha_legacy.md#configuration) for the documentation of the old configuration.

## New Configuration

The section documents the new configuration.

### authgear.yaml

```yaml
risk_assessment:
  enabled: true
  providers:
  - type: recaptchav3
    alias: recaptchav3
    site_key: "SITE_KEY"
    risk_level:
      high_if_gte: 0.7
      medium_if_gte: 0.5

captcha:
  enabled: true
  ip_allowlist:
  - "192.168.0.0/24"
  - "127.0.0.1"
  providers:
  - type: cloudflare
    alias: cloudflare
    site_key: "SITE_KEY"
```

- `risk_assessment.enabled`: If it is true, then risk assessment is enabled.
- `risk_assessment.providers`: A list of risk assessment provider configuration. The actual shape depends on the `type` property.
- `risk_assessment.providers.type`: Required. The type of the risk assessment provider. Valid values are `recaptchav3`.
- `risk_assessment.providers.alias`. Required. The unique identifier of the risk assessment provider. It is used in other parts of the configuration to refer to this particular risk assessment provider. For example, the project can configured a number of risk assessment providers, but only uses one of them in a particular Authentication Flow.
- `risk_assessment.providers.risk_level.high_if_gte`: Required. A floating number. If the provider-specific score is greater than or equal to this number, then the risk level is high. Otherwise, it is medium or low.
- `risk_assessment.providers.risk_level.medium_if_gte`: Required. A floating number. If the provider-specific score is greater than or equal to this number, then the risk level is medium. Otherwise, it is low.

Type specific fields:

- `risk_assessment.providers.type=recaptchav3.site_key`: Required. The site key of reCAPTCHA v3.

---

- `captcha.enabled`: If it is true, the new configuration is used.
- `captcha.ip_allowlist`: A list of IPv4 CIDR notations or IPv4 addresses. If the incoming request matches any entry in the allowlist, the request bypasses Captcha.
- `captcha.providers`: A list of Captcha provider configuration. The actual shape depends on the `type` property.
- `captcha.providers.type`: Required. The type of the Captcha provider. Valid values are `cloudflare` and `recaptchav2`.
- `captcha.providers.alias`: Required. The unique identifier of the Captcha provider. It is used in other parts of the configuration to refer to this particular Captcha provider. For example, the project can configured a number of Captcha providers, but only uses one of them in a particular Authentication Flow.


Type specific fields:

- `captcha.providers.type=cloudflare.site_key`: Required. The site key of Cloudflare Turnstile.
- `captcha.providers.type=recaptchav2.site_key`: Required. The site key of reCAPTCHA v2.

### authgear.secrets.yaml

```yaml
- data:
    items:
    - type: recaptchav3
      alias: recaptchav3
      secret_key: RECAPTCHAV3_SECRET_KEY
  key: risk_assessment.providers

- data:
    items:
    - type: cloudflare
      alias: cloudflare
      secret_key: TURNSTILE_SECRET_KEY
  key: captcha.providers
```

- `key=risk_assessment.providers.items.type`: Required. It is the same as `risk_assessment.providers.type`.
- `key=risk_assessment.providers.items.alias`: Required. It is the same as `risk_assessment.providers.alias`.

Type specific fields:

- `key=risk_assessment.providers.items.type=recaptchav3.secret_key`: Required. The secret key of reCAPTCHA v3.

---

- `key=captcha.providers.items.type`: Required. It is the same as `captcha.providers.type`.
- `key=captcha.providers.items.alias`: Required. It is the same as `captcha.providers.alias`.

Type specific fields:

- `key=captcha.providers.items.type=cloudflare.secret_key`: Required. The secret key of Cloudflare Turnstile.
- `key=captcha.providers.items.type=recaptchav2.secret_key`: Required. The secret key of reCAPTCHA v2.

## Authentication Flow

This section specifies how risk assessment and Captcha works in a Authentication Flow.

### Risk assessment and Captcha in Authentication Flow configuration

Risk assessment and Captcha are supported in the following flow types:

- `signup`
- `promote`
- `login`
- `signup_login`
- `reauth`
- `account_recovery`

Risk assessment and Captcha are supported only in the following step types:

- `identify` in `signup`, `promote`, `login`, `signup_login`, and `account_recovery`.
- `create_authenticator` in `signup` and `promote`.
- `authenticate` in `login` and `reauth`.

We can see that all supported step types have branches.

To enable risk assessment and Captcha in a branch, add `risk_assessment` and `captcha` to the branch.

The configuration is as follows:

```yaml
risk_assessment:
  enabled: true
  provider:
    alias: recaptchav3

captcha:
  mode: "always" # "never" | "always" | "risk_level_low" | "risk_level_medium" | "risk_level_high"
  fail_open: true
  provider:
    alias: cloudflare
```

- `risk_assessment.enabled`: Whether risk assessment is enabled.
- `risk_assessment.provider.alias`: It `enabled` is true, then it is required. Specify the risk assessment provider to be used in this branch.

- `captcha.mode`: When Captcha is required.
  - `never`: Captcha is never required. It is the default.
  - `always`: Captcha is always required. Risk level is ignored.
  - `risk_level_low`: Captcha is required when the risk level obtained by risk assessment is low. If risk assessment is not enabled, then it means `always`. If risk assessment is service unavailable, then it means `always`.
  - `risk_level_medium`: Captcha is required when the risk level obtained by risk assessment is medium. If risk assessment is not enabled, then it means `always`. If risk assessment is service unavailable, then it means `always`.
  - `risk_level_high`: Captcha is required when the risk level obtained by risk assessment is high. If risk assessment is not enabled, then it means `always`. If risk assessment is service unavailable, then it means `always`.
- `captcha.fail_open`: If it is true, then if the Captcha provider is service unavailable, access is granted. It is false by default.
- `captcha.provider.alias`: If `mode` is not `never`, then it is required. Specify the Captcha provider to be used in this branch.

For example,

```yaml
authentication_flow:
  login_flows:
  - name: default
    steps:
    - type: identify
      one_of:
      - identification: email
        # Identify with email requires captcha.
        captcha:
          mode: "always"
          provider:
            alias: cloudflare
    - type: authenticate
      one_of:
      - authentication: primary_password
      - authentication: primary_oob_otp_email
```

### Behavior of generated flows

Given `risk_assessment.enabled=true` and `risk_assessment.providers` is non-empty,

1. All the branches of the first step (that is, the `identify` step, or the `authenticate` step in reauth flow) has `risk_assessment.enabled=true`.
2. The first provider in `risk_assessment.providers` is used as `risk_assessment.provider.alias`

Given `captcha.enabled=true` and `captcha.providers` is non-empty,

1. All the branches of the first step (that is, the `identify` step, or the `authenticate` step in reauth flow) has `captcha.mode=always`.
2. The first provider in `captcha.providers` is used as `captcha.provider.alias`

In terms of UX, when Captcha is enabled and configured, every generated flow requires captcha at the beginning of the flow.

### Risk assessment and Captcha in Authentication Flow API

Please refer to [Risk assessment and Captcha](./authentication-flow-api-reference.md#risk-assessment-and-captcha).

### Advanced use case: Require Captcha at a specific branch only

Suppose Project A configures email login with password or OTP. The developer may only want to enable captcha if OTP is used, to reduce friction.

This can be achieved by customizing the flow.

```yaml
authentication_flow:
  login_flows:
  - name: default
    steps:
    - type: identify
      one_of:
      - identification: email
    - type: authenticate
      one_of:
      - authentication: primary_password
      # Must pass Captcha BEFORE selecting this branch.
      # That is, before the OTP is sent.
      - authentication: primary_oob_otp_email
        captcha:
          mode: "always"
          provider:
            alias: cloudflare
```

### Advanced use case: Use different Captcha providers in different branches

The developer can specify different Captcha provider to be used in different branches.

```yaml
authentication_flow:
  login_flows:
  - name: default
    steps:
    - type: identify
      one_of:
      - identification: email
    - type: authenticate
      one_of:
      - authentication: primary_password
        captcha:
          mode: "always"
          provider:
            alias: recaptchav2
      - authentication: primary_oob_otp_email
        captcha:
          mode: "always"
          provider:
            alias: cloudflare
```

### Advanced use case: Use fail-open instead of fail-close

By default, Captcha is fail-close, meaning that Captcha must be passed in order to gain access.
If Captcha is fail-open, then the Captcha provider service unavailable grants access.
Note that access is still denied if the Captcha provider returns a failed verification result.

Here is an example configuration:

```yaml
authentication_flow:
  login_flows:
  - name: default
    steps:
    - type: identify
      one_of:
      - identification: email
    - type: authenticate
      one_of:
      - authentication: primary_password
        captcha:
          mode: "always"
          fail_open: true
          provider:
            alias: cloudflare
```

### Advanced use case: Allow internal staff to bypass Captcha

If internal staff is connected to a private network, thus having an IP address in a specific range,
they can bypass Captcha. This is generally for convenience.

Here is an example configuration:

```yaml
captcha:
  enabled: true
  ip_allowlist:
  - "10.0.0.0/16"
  providers:
  - type: cloudflare
    alias: cloudflare
    site_key: "SITE_KEY"

authentication_flow:
  login_flows:
  - name: default
    steps:
    - type: identify
      one_of:
      - identification: email
    - type: authenticate
      one_of:
      - authentication: primary_password
        captcha:
          mode: "always"
          provider:
            alias: cloudflare
```

If the incoming request has an IP address of `10.0.0.1`, it is granted access automatically.

### Advanced use case: Require Captcha only when risk level is high

To minimize friction in UX, it is common to require Captcha only when the risk level is high.

Here is an example configuration:

```yaml
risk_assessment:
  enabled: true
  providers:
  - type: recaptchav3
    alias: recaptchav3
    site_key: "SITE_KEY"
    risk_score:
      low: 0.2
      medium: 0.5
      high: 0.7

captcha:
  enabled: true
  providers:
  - type: cloudflare
    alias: cloudflare
    site_key: "SITE_KEY"

authentication_flow:
  login_flows:
  - name: default
    steps:
    - type: identify
      one_of:
      - identification: email
    - type: authenticate
      one_of:
      - authentication: primary_password
        risk_assessment:
          enabled: true
          provider:
            alias: recaptchav3
        captcha:
          mode: "risk_level_high"
          provider:
            alias: cloudflare
```

When authenticating with password, a risk assessment has to be done first.
If the risk level is low or medium, access is granted.
Otherwise, Captcha is required.

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
