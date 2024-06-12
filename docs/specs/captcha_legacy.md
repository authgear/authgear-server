# Captcha (Legacy)

This document specifies Captcha support in the abandoned [Workflow](./workflow.md).

## Configuration

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

## Supported Workflows

The following intents supports Captcha, if the [old configuration](./captcha.md#old-configuration) is used.

- latte.IntentVerifyIdentity
- latte.IntentAuthenticateOOBOTPPhone
