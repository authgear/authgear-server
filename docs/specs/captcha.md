# Captcha

- [Configurations](#configurations)
  - [authgear.yaml](#authgearyaml)
  - [authgear.secrets.yaml](#authgearsecretsyaml)
- [Supported Workflows](#supported-workflows)

## Configurations

### authgear.yaml

Before using any workflow with captcha enabled, the captcha provider must be configured in `authgear.yaml`:

```yaml
captcha:
  provider: cloudflare
```

- `provider`
  - The captcha provider. Currently, only `cloudflare` is supported.

### authgear.secrets.yaml

When the captcha provider is configured, the corresponding secret must exist in `authgear.secrets.yaml`.

```yaml
- data:
    secret: YOUR_TURNSTILE_SECRET_KEY
  key: captcha.cloudflare
```

- Provider: `cloudflare`
  - `key`: `captcha.cloudflare`
  - `data.secret`: The turnstile secret key

## Supported Workflows

The following intents supports captcha.

- latte.IntentVerifyIdentity
- latte.IntentAuthenticateOOBOTPPhone
