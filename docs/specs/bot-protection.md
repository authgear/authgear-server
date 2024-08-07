- [Bot protection](#bot-protection)
  * [Old configuration](#old-configuration)
  * [New Configuration](#new-configuration)
    + [authgear.yaml](#authgearyaml)
      - [Risk level `mode`](#risk-level-mode)
    + [authgear.secrets.yaml](#authgearsecretsyaml)
  * [Authentication Flow](#authentication-flow)
    + [Bot protection in Authentication Flow configuration](#bot-protection-in-authentication-flow-configuration)
    + [Behavior of builtin flows](#behavior-of-builtin-flows)
    + [Bot protection in Authentication Flow API](#bot-protection-in-authentication-flow-api)
    + [Advanced use case: Require challenged-base bot protection at a specific branch only](#advanced-use-case-require-challenged-base-bot-protection-at-a-specific-branch-only)
    + [Advanced use case: Use fail-open instead of fail-close](#advanced-use-case-use-fail-open-instead-of-fail-close)
    + [Advanced use case: Allow internal staff to bypass bot protection](#advanced-use-case-allow-internal-staff-to-bypass-bot-protection)
    + [Advanced use case: Require challenge-based bot protection only when risk level is high](#advanced-use-case-require-challenge-based-bot-protection-only-when-risk-level-is-high)
    + [Unsupported use case: Use different challenge-based providers in different branches](#unsupported-use-case-use-different-challenge-based-providers-in-different-branches)
  * [Audit log](#audit-log)
  * [Study on bot protection providers](#study-on-bot-protection-providers)
    + [Geetest v4](#geetest-v4)
    + [Geetest v3](#geetest-v3)
    + [hCaptcha](#hcaptcha)
    + [reCAPTCHA v2](#recaptcha-v2)
    + [reCAPTCHA v3](#recaptcha-v3)
    + [reCAPTCHA Enterprise](#recaptcha-enterprise)
    + [Cloudflare Turnstile](#cloudflare-turnstile)
    + [Arkose Labs Bot Manager](#arkose-labs-bot-manager)
    + [Tencent Captcha](#tencent-captcha)

# Bot protection

## Old configuration

> The old configuration **IS NOT** used by Authentication Flow.
> To configure a bot protection provider for an Authentication Flow, the new configuration must be used.

See [here](./captcha_legacy.md#configuration) for the documentation of the old configuration.

## New Configuration

The section documents the new configuration.

### authgear.yaml

```yaml
bot_protection:
  enabled: true
  ip_allowlist:
  - "192.168.0.0/24"
  - "127.0.0.1"
  provider:
    type: cloudflare
    site_key: "SITE_KEY"
  risk_assessment:
    enabled: true
    provider:
      type: recaptchav3
      site_key: "SITE_KEY"
      risk_level:
        high_if_gte: 0.7
        medium_if_gte: 0.5
  requirements:
    signup_or_login:
      mode: "always" # "never" | "always" | "risk_level_medium" | "risk_level_high"
    account_recovery:
      mode: "always" # "never" | "always" | "risk_level_medium" | "risk_level_high"
    password:
      mode: "always" # "never" | "always" | "risk_level_medium" | "risk_level_high"
    oob_otp_email:
      mode: "always" # "never" | "always" | "risk_level_medium" | "risk_level_high"
    oob_otp_sms:
      mode: "always" # "never" | "always" | "risk_level_medium" | "risk_level_high"
```

- `bot_protection.enabled`: If it is true, the new configuration is used.
- `bot_protection.ip_allowlist`: A list of IPv4/IPv6 CIDR notations or addresses. If the incoming request matches any entry in the allowlist, the request bypasses bot protection.
- `bot_protection.provider`: A challenge-based provider configuration. The actual shape depends on the `type` property.
- `bot_protection.provider.type`: Required. The type of the challenge-based provider. Valid values are `cloudflare` and `recaptchav2`.
- `bot_protection.risk_assessment.enabled`: If it is true, then risk assessment is enabled.
- `bot_protection.risk_assessment.provider`: A risk assessment provider configuration. The actual shape depends on the `type` property.
- `bot_protection.risk_assessment.provider.type`: Required. The type of the risk assessment provider. Valid values are `recaptchav3`.
- `bot_protection.risk_assessment.provider.risk_level.high_if_gte`: Required. A floating number. If the provider-specific score is greater than or equal to this number, then the risk level is high. Otherwise, it is medium or low.
- `bot_protection.risk_assessment.provider.risk_level.medium_if_gte`: Required. A floating number. If the provider-specific score is greater than or equal to this number, then the risk level is medium. Otherwise, it is low.
- `bot_protection.requirements.signup_or_login.mode`: Optional. [Risk level mode](#risk-level-mode). Default `never`. See [Behavior of builtin flows](#behavior-of-builtin-flows).
- `bot_protection.requirements.account_recovery.mode`: Optional. [Risk level mode](#risk-level-mode). Default `never`. See [Behavior of builtin flows](#behavior-of-builtin-flows).
- `bot_protection.requirements.password`: Optional. [Risk level mode](#risk-level-mode). Default `never`. See [Behavior of builtin flows](#behavior-of-builtin-flows).
- `bot_protection.requirements.oob_otp_email`: Optional. [Risk level mode](#risk-level-mode). Default `never`. See [Behavior of builtin flows](#behavior-of-builtin-flows).
- `bot_protection.requirements.oob_otp_sms`: Optional. [Risk level mode](#risk-level-mode). Default `never`. See [Behavior of builtin flows](#behavior-of-builtin-flows).

Type specific fields:

- `bot_protection.provider.type=cloudflare.site_key`: Required. The site key of Cloudflare Turnstile.
- `bot_protection.provider.type=recaptchav2.site_key`: Required. The site key of reCAPTCHA v2.
- `bot_protection.risk_assessment.provider.type=recaptchav3.site_key`: Required. The site key of reCAPTCHA v3.

#### Risk level `mode`

| Value               | Effect                                                                                                                                                                                                                                                                                                                                                   |
|---------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `never`             | Bot protection is never required. It is the default.                                                                                                                                                                                                                                                                                                     |
| `always`            | Bot protection is always required. Risk level is ignored. Note this is equivalent to  `risk_level_low`                                                                                                                                                                                                                                                   |
| `risk_level_medium` | Bot protection is required when the risk level obtained by risk assessment is medium or above. If risk assessment is not enabled, then it means `always`. If risk assessment is service unavailable, then it means `always`. There is no `risk_level_low` because risk level is always low, medium, or high. `risk_level_low` is equivalent to `always`. |
| `risk_level_high`   | Bot protection is required when the risk level obtained by risk assessment is high. If risk assessment is not enabled, then it means `always`. If risk assessment is service unavailable, then it means `always`. There is no `risk_level_low` because risk level is always low, medium, or high.                                                        |


### authgear.secrets.yaml

```yaml
- data:
    type: recaptchav3
    secret_key: RECAPTCHAV3_SECRET_KEY
  key: bot_protection.risk_assessment.provider

- data:
    type: cloudflare
    secret_key: TURNSTILE_SECRET_KEY
  key: bot_protection.provider
```

- `key=bot_protection.risk_assessment.provider.type`: Required. It is the same as `bot_protection.risk_assessment.provider.type`.

Type specific fields:

- `key=bot_protection.risk_assessment.provider.type=recaptchav3.secret_key`: Required. The secret key of reCAPTCHA v3.

---

- `key=bot_protection.provider.type`: Required. It is the same as `bot_protection.provider.type`.

Type specific fields:

- `key=bot_protection.provider.type=cloudflare.secret_key`: Required. The secret key of Cloudflare Turnstile.
- `key=bot_protection.provider.type=recaptchav2.secret_key`: Required. The secret key of reCAPTCHA v2.

## Authentication Flow

This section specifies how bot protection works in a Authentication Flow.

### Bot protection in Authentication Flow configuration

Bot protection is supported in the following flow types:

- `signup`
- `promote`
- `login`
- `signup_login`
- `reauth`
- `account_recovery`

Bot protection is supported only in the following step types:

- `identify` in `signup`, `promote`, `login`, `signup_login`, and `account_recovery`.
- `authenticate` in `login` and `reauth`.
- `create_authenticator` in `signup` and `promote`

To enable bot protection in a branch, add `bot_protection` to the branch.

The configuration is as follows:

```yaml
bot_protection:
  mode: "always" # "never" | "always" | "risk_level_medium" | "risk_level_high"
  fail_open: true
  provider:
    type: cloudflare
  risk_assessment:
    enabled: true
    provider:
      type: recaptchav3
```

- `bot_protection.mode`: When bot protection is required. See [Risk level `mode`](#risk-level-mode).
- `bot_protection.fail_open`: If it is true, then if the challenge-based provider is service unavailable, access is granted. It is false by default.
- `bot_protection.provider.type`: If `mode` is not `never`, then it is required. Specify the challenge-based provider to be used in this branch.
- `bot_protection.risk_assessment.enabled`: Whether risk assessment is enabled.
- `bot_protection.risk_assessment.provider.type`: It `enabled` is true, then it is required. Specify the risk assessment provider to be used in this branch.

For example,

```yaml
authentication_flow:
  login_flows:
  - name: default
    steps:
    - type: identify
      one_of:
      - identification: email
        # Identify with email requires bot protection.
        bot_protection:
          mode: "always"
          provider:
            type: cloudflare
    - type: authenticate
      one_of:
      - authentication: primary_password
      - authentication: primary_oob_otp_email
```

### Behavior of builtin flows

Given `bot_protection.risk_assessment.enabled=true`,

1. All the branches of the first step (that is, the `identify` step, or the `authenticate` step in reauth flow) has `bot_protection.risk_assessment.enabled=true`.
2. The configured provider is used as `bot_protection.risk_assessment.provider.type`

The bot protection behavior in builtin flows depend on

- `bot_protection.enabled`
- `bot_protection.requirements`

Given `bot_protection.enabled=true`,

<table>
   <tr>
      <th>Configuration</th>
      <th>Behavior</th>
   </tr>
   <tr>
      <td>
        <code>bot_protection.requirements.signup_or_login.mode</code>
      </td>
      <td>
        Contribute to flows:
          <ol>
            <li>
              <code>signup</code>
              <ul>
                <li><code>identify</code></li>
                <ul>
                  <li><code>email</code></li>
                  <li><code>phone</code></li>
                  <li><code>username</code></li>
                </ul>
              </ul>
            </li>
            <li>
              <code>login</code>
              <ul>
                <li><code>identify</code></li>
                <ul>
                  <li><code>email</code></li>
                  <li><code>phone</code></li>
                  <li><code>username</code></li>
                </ul>
              </ul>
            </li>
            <li>
              <code>signup_login</code>
              <ul>
                <li><code>identify</code></li>
                <ul>
                  <li><code>email</code></li>
                  <li><code>phone</code></li>
                  <li><code>username</code></li>
                </ul>
              </ul>
            </li>
            <li>
              <code>promote</code>
              <ul>
                <li><code>identify</code></li>
                <ul>
                  <li><code>email</code></li>
                  <li><code>phone</code></li>
                  <li><code>username</code></li>
                </ul>
              </ul>
            </li>
          </ol>
   </tr>
   <tr>
      <td><code>bot_protection.requirements.account_recovery.mode</code></td>
      <td>
        Contribute to flows:
          <ol>
            <li>
              <code>account_recovery</code>
              <ul>
                <li><code>identify (all branches) </code></li>
              </ul>
            </li>
          </ol>
      </td>
   </tr>
   <tr>
      <td><code>bot_protection.requirements.password.mode</code></td>
      <td>
        Contribute to flows:
          <ol>
            <li>
              <code>login</code>
              <ul>
                <li><code>authenticate</code></li>
                <ul>
                  <li><code>primary_password</code></li>
                  <li><code>secondary_password</code></li>
                </ul>
              </ul>
            </li>
          </ol>
      </td>
   </tr>
   <tr>
      <td><code>bot_protection.requirements.oob_otp_email.mode</code></td>
      <td>
        Contribute to flows:
          <ol>
            <li>
            <!-- Promote only have identify, create_authenticator will reuse signup -->
              <code>signup (promote) </code>
              <ul>
                <li><code>identify</code></li>
                <ul>
                  <li><code>email</code></li>
                  <ul>
                    <li>if <code>verification.claims.email.required=true</code></li>
                  </ul>
                </ul>
                <li><code>create_authenticator</code></li>
                <ul>
                  <li><code>primary_oob_otp_email</code></li>
                  <li><code>secondary_oob_otp_email</code></li>
                </ul>
              </ul>
            </li>
            <li>
              <code>login</code>
              <ul>
                <li><code>authenticate</code></li>
                <ul>
                  <li><code>primary_oob_otp_email</code></li>
                  <li><code>secondary_oob_otp_email</code></li>
                </ul>
              </ul>
            </li>
          </ol>
      </td>
   </tr>
   <tr>
      <td><code>bot_protection.requirements.oob_otp_sms.mode</code></td>
       <td>
        Contribute to flows:
          <ol>
            <li>
            <!-- Promote only have identify, create_authenticator will reuse signup -->
              <code>signup (promote)</code>
              <ul>
                <li><code>identify</code></li>
                <ul>
                  <li><code>phone</code></li>
                  <ul>
                    <li>if <code>verification.claims.phone_number.required=true</code></li>
                  </ul>
                </ul>
              </ul>
              <ul>
                <li><code>create_authenticator</code></li>
                <ul>
                  <li><code>primary_oob_otp_sms</code></li>
                  <li><code>secondary_oob_otp_sms</code></li>
                </ul>
              </ul>
            </li>
            <li>
              <code>login</code>
              <ul>
                <li><code>authenticate</code></li>
                <ul>
                  <li><code>primary_oob_otp_sms</code></li>
                  <li><code>secondary_oob_otp_sms</code></li>
                </ul>
              </ul>
            </li>
          </ol>
      </td>
   </tr>
</table>


Some branches have multiple contributions from various configuration. The most strict `mode` is used in this case.

### Bot protection in Authentication Flow API

Please refer to [Bot protection](./authentication-flow-api-reference.md#bot-protection).

### Advanced use case: Require challenged-base bot protection at a specific branch only

Suppose Project A configures email login with password or OTP. The developer may only want to enable bot protection if OTP is used, to reduce friction.

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
      # Must pass bot protection BEFORE selecting this branch.
      # That is, before the OTP is sent.
      - authentication: primary_oob_otp_email
        bot_protection:
          mode: "always"
          provider:
            type: cloudflare
```

### Advanced use case: Use fail-open instead of fail-close

By default, bot protection is fail-close, meaning that bot protection must be passed in order to gain access.
If bot protection is fail-open, then the bot protection provider service unavailable grants access.
Note that access is still denied if the bot protection provider returns a failed verification result.

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
        bot_protection:
          mode: "always"
          fail_open: true
          provider:
            type: cloudflare
```

### Advanced use case: Allow internal staff to bypass bot protection

If internal staff is connected to a private network, thus having an IP address in a specific range,
they can bypass bot protection. This is generally for convenience.

Here is an example configuration:

```yaml
bot_protection:
  enabled: true
  ip_allowlist:
  - "10.0.0.0/16"
  provider:
    type: cloudflare
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
        bot_protection:
          mode: "always"
          provider:
            type: cloudflare
```

If the incoming request has an IP address of `10.0.0.1`, it is granted access automatically.

### Advanced use case: Require challenge-based bot protection only when risk level is high

To minimize friction in UX, it is common to require challenge-based bot protection only when the risk level is high.

Here is an example configuration:

```yaml
bot_protection:
  enabled: true
  provider:
    type: cloudflare
    site_key: "SITE_KEY"
  risk_assessment:
    enabled: true
    provider:
      type: recaptchav3
      site_key: "SITE_KEY"
      risk_score:
        low: 0.2
        medium: 0.5
        high: 0.7

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
        bot_protection:
          mode: "risk_level_high"
          provider:
            type: cloudflare
          risk_assessment:
            enabled: true
            provider:
              type: recaptchav3
```

When authenticating with password, a risk assessment has to be done first.
If the risk level is low or medium, access is granted.
Otherwise, challenge-based bot protection is required.

### Unsupported use case: Use different challenge-based providers in different branches

Since the configuration allow one provider only. It is impossible to use different providers in different branches.
The developer must use a provider that is generally enough for their use case.
For example, if the project targets to support both web and mobile platform, they have to use a provider that supports both.

## Audit log

When a verification failure is detected, the event [bot_protection.verification.failed](./event.md#bot_protectionverificationfailed) is logged.

## Study on bot protection providers

### Geetest v4

Geetest v4 is a challenge-based bot protection provider.

(https://mermaid.live/edit#pako:eNp1kktrwzAQhP-K2GsTeulJh0Bf9BQoSU_FF2GNE4Etuet1-gj5713HtR1KrJOk_XY0y-hIefIgSw0-WsQcT8Ht2FVZNLoey4Aoy9Xq5r6V_Q6OrckZTmCcXhRl-uzBoaxo32PNQ5JXToJcQoobVQ8Mb8Cc-L_4C_CGRg531oQYJLgy_KCHxpJyy0F6ggyjaUuZd3vVxazCxRiXpm4Pinqd-oqn6akBmlMc7Ud8iWkENS2oAlcueA3g2DVkJHtUyMjq1qNwnRBl8aRoW3fqzz5IYrKFKxssSGNI2--YkxVuMUB_IY5U7eJ7StMZZ5F1n_z5A5x-AdWRtBU)

Sequence diagram by Geetest: https://docs.geetest.com/BehaviorVerification/overview/communicationProcess/

Server API reference: https://docs.geetest.com/BehaviorVerification/deploy/server/

### Geetest v3

Geetest v3 is a challenge-based bot protection provider. Note that with Geetest v3, the process has to be initiated by the server, as opposed to other challenge-based bot protection providers.

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

reCAPTCHA v2 is a challenge-based bot protection provider.

reCAPTCHA v2 comes in 2 flavors:

- [Checkbox](https://developers.google.com/recaptcha/docs/display)
- [Invisible](https://developers.google.com/recaptcha/docs/invisible)

reCAPTCHA v2 does not support iOS. It supports [Android](https://developer.android.com/privacy-and-security/safetynet/recaptcha) though.

### reCAPTCHA v3

reCAPTCHA v3 is NOT challenge-based bot protection provider.

It does not support mobile platforms out-of-the-box.

It is worth to note that it has [Actions](https://developers.google.com/recaptcha/docs/v3#actions).

### reCAPTCHA Enterprise

reCAPTCHA Enterprise is reCAPTCHA v2 and reCAPTCHA v3 offered in a package. Additionally, the hightest pricing tier offers mobile platforms support.

### Cloudflare Turnstile

Cloudflare Turnstile is a challenge-based bot protection provider.

Cloudflare Turnstile comes in 3 flavors:

- [Managed](https://developers.cloudflare.com/turnstile/concepts/widget-types/#managed-recommended). Let Cloudflare to decide whether to show checkbox.
- [Non-interactive](https://developers.cloudflare.com/turnstile/concepts/widget-types/#non-interactive). It is just a badge.
- [Invisible](https://developers.cloudflare.com/turnstile/concepts/widget-types/#invisible). The end-user does not see anything visible.

Cloudflare Turnstile does not support mobile platforms natively.

### Arkose Labs Bot Manager

Arkose Labs Bot Manager CAN BE a challenge-basd bot protection provider.

It supports prompting interactive challenge if necessary.

It supports mobile platforms with webview (packaged as a SDK).

### Tencent Captcha

Tencent Captcha is a challenge-based bot protection provider.
