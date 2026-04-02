# Usage

- [Supported Usage Names](#supported-usage-names)
- [Usage](#usage)
  - [authgear.features.yaml](#authgearfeaturesyaml)
  - [authgear.yaml](#authgearyaml)
- [Usage Limits](#usage-limits)
- [Alert Delivery](#alert-delivery)
  - [hooks](#hooks)
  - [alerts](#alerts)
- [Event and Hook](#event-and-hook)
- [Feature Config Merging of Usage](#feature-config-merging-of-usage)
- [Portal Configurations](#portal-configurations)
- [Deprecated Legacy Usage Limit Configs](#deprecated-legacy-usage-limit-configs)

## Supported Usage Names

Supported usage names are:

- `user_export`
- `user_import`
- `email`
- `whatsapp`
- `sms`

## Usage

`authgear.features.yaml` and `authgear.yaml` share the same `usage.limits` shape. Delivery differs through `usage.hooks` and `usage.alerts`:

- In `authgear.features.yaml`, `usage.hooks` is supported.
- In `authgear.yaml`, `usage.alerts` is supported.

The shared fields are:

`usage.limits`: A map from usage name to a list of configured limits. Read [Usage Limits](#usage-limits) for detail.

`usage.limits.<usage_name>`: A list of limits for the corresponding usage name. Supported `<usage_name>` values are listed in [Supported Usage Names](#supported-usage-names). Read [Usage Limits](#usage-limits) for detail.

### authgear.features.yaml

Use the `usage` section in `authgear.features.yaml` to configure hooks and limits.

```yaml
usage:
  hooks:
    - url: https://example.com/your_webhook
      match: email
    - url: deno:///app/usage-alert.ts
      match: "*"
  limits:
    user_export:
      - quota: 5
        period: day
        action: alert
    user_import:
      - quota: 1000
        period: day
        action: alert
    email:
      - quota: 100
        period: month
        action: alert
    whatsapp:
      - quota: 50
        period: month
        action: alert
    sms:
      - quota: 4
        period: month
        action: alert
      - quota: 900
        period: month
        action: block
```

`usage.hooks`: The list of hook deliveries configured in `authgear.features.yaml`. Read [Alert Delivery](#alert-delivery) for detail.

`usage.hooks[].url`: Required. The endpoint to receive the hook request. The url can also be a deno script url. Read [hooks](#hooks) for detail.

`usage.hooks[].match`: Required. The usage name this hook subscribes to. Supported values are listed in [Supported Usage Names](#supported-usage-names), and `*`. `*` means all usage names. Read [Alert Delivery](#alert-delivery) for detail.

### authgear.yaml

Use the `usage` section in `authgear.yaml` to configure alerts and limits.

```yaml
usage:
  alerts:
    - type: email
      email: email@example.com
      match: email
  limits:
    user_export:
      - quota: 5
        period: day
        action: alert
    user_import:
      - quota: 1000
        period: day
        action: alert
    email:
      - quota: 100
        period: month
        action: alert
    whatsapp:
      - quota: 50
        period: month
        action: alert
    sms:
      - quota: 4
        period: month
        action: alert
      - quota: 900
        period: month
        action: block
```

`usage.alerts`: The list of alerts configured in `authgear.yaml`. Read [Alert Delivery](#alert-delivery) for detail.

`usage.alerts[].type`: Required. In `authgear.yaml`, the only supported value is `email`. Read [alerts](#alerts) for detail.

`usage.alerts[].email`: Required when `type` is `email`. Read [alerts](#alerts) for detail.

`usage.alerts[].match`: Required. The usage name this alert subscribes to. Supported values are listed in [Supported Usage Names](#supported-usage-names), and `*`. `*` means all usage names. Read [Alert Delivery](#alert-delivery) for detail.

## Usage Limits

`usage.limits.<usage_name>`: A list of limits for the corresponding usage name. Supported `<usage_name>` values are listed in [Supported Usage Names](#supported-usage-names).

`usage.limits.<usage_name>[].quota`: Required. Integer. The usage value that triggers this limit.

`usage.limits.<usage_name>[].period`: Required. Depends on the usage name. For example, messaging usage may use `month`, while admin API usage may use `day`.

`usage.limits.<usage_name>[].action`: Required. The action to take when usage reaches the quota. Supported values are `alert` and `block`.

## Alert Delivery

When a usage limit is triggered, only hook or alert entries with the same `match`, or `match: "*"`, receive the notification.

### hooks

In `authgear.features.yaml`, `usage.hooks` configures hook delivery.

`usage.hooks[].url`: Required. The endpoint to receive the hook request. The url can also be a deno script url.

`usage.hooks[].match`: Required. The usage name this hook subscribes to. Supported values are listed in [Supported Usage Names](#supported-usage-names), and `*`.

When a matching usage limit is triggered, the configured `url` is used to trigger a hook for the [`usage.alert.triggered`](./event.md#usagealerttriggered) event. Read [Event](./event.md#usagealerttriggered) and [Hook](./hook.md) for details.

### alerts

In `authgear.yaml`, `usage.alerts` configures alert delivery.

`usage.alerts[].type`: Required. Must be `email`.

`usage.alerts[].email`: Required when `type` is `email`. The email address to receive the alert.

`usage.alerts[].match`: Required. The usage name this alert subscribes to. Supported values are listed in [Supported Usage Names](#supported-usage-names), and `*`.

These alert emails are not counted in `email` usage, to prevent recurring usage caused by the alert itself.

## Event and Hook

When a usage limit is triggered, Authgear emits the [`usage.alert.triggered`](./event.md#usagealerttriggered) event, regardless of whether the configured action is `alert` or `block`.

For `usage.hooks`, the configured `url` is used to trigger a hook for this event. Read [Hook](./hook.md) for details.

In `authgear.yaml`, `hook.non_blocking_handlers[].events` can also be used to listen to this event.

Read [Event](./event.md#usagealerttriggered) for the event payload.

## Feature Config Merging of Usage

The `usage` section in feature config can be defined at site level, plan level, or project level.

`usage.hooks` is merged by appending entries from lower precedence to higher precedence.

Each `usage.limits.<usage_name>` list is also overridden by higher-precedence feature config.

The precedence order is:

1. project-level feature config
2. plan-level feature config
3. site-level feature config

See the below example:

A feature config of a plan:

```yaml
usage:
  hooks:
    - url: https://internal.authgear.cloud/notification
      match: sms
  limits:
    sms:
      - quota: 900
        period: month
        action: block
```

A feature config of a project:

```yaml
usage:
  hooks:
    - url: https://example.com/another_notification
      match: sms
  limits:
    sms:
      - quota: 800
        period: month
        action: block
```

The resulting config should be:

```yaml
usage:
  hooks:
    - url: https://internal.authgear.cloud/notification
      match: sms
    - url: https://example.com/another_notification
      match: sms
  limits:
    sms:
      - quota: 800
        period: month
        action: block
```

In other words, `usage.hooks` is merged by appending project-level feature config entries after plan-level feature config entries, and each `usage.limits.<usage_name>` list is overridden by higher-precedence feature config.

## Portal Configurations

The portal should automatically fill `usage.alerts` with the project owner's email address.

## Deprecated Legacy Usage Limit Configs

The old usage limit configs in feature config are deprecated in favor of the `usage` section and will be removed.

The deprecated configs are:

- `admin_api.user_export_usage`
- `admin_api.user_import_usage`
- `messaging.email_usage`
- `messaging.whatsapp_usage`
- `messaging.sms_usage`

Example:

```yaml
admin_api:
  user_export_usage:
    enabled: true
    period: day
    quota: 24
  user_import_usage:
    enabled: true
    period: day
    quota: 10000
messaging:
  email_usage:
    enabled: true
    period: month
    quota: 1000
  whatsapp_usage:
    enabled: true
    period: month
    quota: 1000
  sms_usage:
    enabled: true
    period: month
    quota: 1000
```
