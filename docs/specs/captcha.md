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
