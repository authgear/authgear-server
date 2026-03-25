# Thresholds

- [Supported Threshold Names](#supported-threshold-names)
- [Thresholds](#thresholds)
  - [authgear.features.yaml](#authgearfeaturesyaml)
  - [authgear.yaml](#authgearyaml)
- [Alert Channels](#alert-channels)
  - [url](#url)
  - [email](#email)
- [Event and Hook](#event-and-hook)
- [Feature Config Merging of Thresholds](#feature-config-merging-of-thresholds)
- [Portal Configurations](#portal-configurations)
- [Deprecated Usage Configs](#deprecated-usage-configs)

## Supported Threshold Names

Supported threshold names are:

- `user_export_usage`
- `user_import_usage`
- `email_usage`
- `whatsapp_usage`
- `sms_usage`

## Thresholds

`authgear.features.yaml` and `authgear.yaml` share the same `thresholds` section shape. The main difference is the supported channel types:

- In `authgear.features.yaml`, `thresholds.alert_channels[].type` only supports `url`.
- In `authgear.yaml`, `thresholds.alert_channels[].type` only supports `email`.

The shared shape is:

```yaml
thresholds:
  alert_channels:
    - type: url
      url: https://example.com/your_webhook
      threshold_name: email_usage
    - type: url
      url: deno:///app/usage-alert.ts
      threshold_name: "*"
  user_export_usage_thresholds:
    - threshold: 5
      period: "day"
      type: soft
  user_import_usage_thresholds:
    - threshold: 1000
      period: "day"
      type: soft
  email_usage_thresholds:
    - threshold: 100
      period: "month"
      type: soft
  whatsapp_usage_thresholds:
    - threshold: 50
      period: "month"
      type: soft
  sms_usage_thresholds:
    - threshold: 4
      period: "month"
      type: soft
    - threshold: 900
      period: "month"
      type: soft
```

`thresholds.alert_channels`: The list of alert channels to notify when a configured threshold is triggered. Read [Alert Channels](#alert-channels) for detail.

`thresholds.alert_channels[].type`: Required. In `authgear.features.yaml`, the only supported value is `url`. In `authgear.yaml`, the only supported value is `email`. Read [url](#url) and [email](#email) for detail.

`thresholds.alert_channels[].url`: Required when `type` is `url`. Read [url](#url) for detail.

`thresholds.alert_channels[].email`: Required when `type` is `email`. Read [email](#email) for detail.

`thresholds.alert_channels[].threshold_name`: Required. The threshold name this channel subscribes to. Supported values are listed in [Supported Threshold Names](#supported-threshold-names), and `*`. `*` means all threshold names. Read [Alert Channels](#alert-channels) for detail.

`thresholds.<threshold_name>_thresholds`: A list of thresholds to trigger when usage crosses from below to at least the configured threshold for the corresponding threshold name.

`thresholds.<threshold_name>_thresholds[].threshold`: Required. Integer. The usage value to trigger this threshold.

`thresholds.<threshold_name>_thresholds[].period`: Required. Depends on the threshold name. For example, messaging thresholds may use `month`, while admin API thresholds may use `day`.

`thresholds.<threshold_name>_thresholds[].type`: Required. Supported values are `soft` and `hard`.

### authgear.features.yaml

Use the `thresholds` section in `authgear.features.yaml` to configure threshold alerts delivered by `url` channels.

```yaml
thresholds:
  alert_channels:
    - type: url
      url: https://example.com/your_webhook
      threshold_name: email_usage
    - type: url
      url: deno:///app/usage-alert.ts
      threshold_name: "*"
  user_export_usage_thresholds:
    - threshold: 5
      period: "day"
      type: soft
  user_import_usage_thresholds:
    - threshold: 1000
      period: "day"
      type: soft
  email_usage_thresholds:
    - threshold: 100
      period: "month"
      type: soft
  whatsapp_usage_thresholds:
    - threshold: 50
      period: "month"
      type: soft
  sms_usage_thresholds:
    - threshold: 4
      period: "month"
      type: soft
    - threshold: 900
      period: "month"
      type: soft
```

### authgear.yaml

Use the `thresholds` section in `authgear.yaml` to configure threshold alerts delivered by `email` channels.

```yaml
thresholds:
  alert_channels:
    - type: email
      email: <project_owner_email_1>
      threshold_name: sms_usage
    - type: email
      email: <project_collaborator_email_2>
      threshold_name: "*"
  user_export_usage_thresholds:
    - threshold: 5
      period: "day"
      type: soft
  user_import_usage_thresholds:
    - threshold: 1000
      period: "day"
      type: soft
  sms_usage_thresholds:
    - threshold: 900
      period: "month"
      type: soft
  email_usage_thresholds:
    - threshold: 1000
      period: "month"
      type: soft
```

## Alert Channels

Each alert channel subscribes to a threshold by `threshold_name`. When a threshold is triggered, only alert channels with the same `threshold_name`, or `threshold_name: "*"`, receive the alert.

### url

In `authgear.features.yaml`, the only supported channel type is `url`.

`thresholds.alert_channels[].type`: Required. Must be `url`.

`thresholds.alert_channels[].url`: Required when `type` is `url`. The endpoint to receive the alert request. The url can also be a deno script url.

`thresholds.alert_channels[].threshold_name`: Required. The threshold name this channel subscribes to. Supported values are listed in [Supported Threshold Names](#supported-threshold-names), and `*`. `*` means all threshold names.

When a matching threshold is triggered, the configured `url` is used to trigger a hook for the [`threshold.alert.triggered`](./event.md#thresholdalerttriggered) event. Read [Event](./event.md#thresholdalerttriggered) and [Hook](./hook.md) for details.

### email

In `authgear.yaml`, the only supported channel type is `email`.

`thresholds.alert_channels[].type`: Required. Must be `email`.

`thresholds.alert_channels[].email`: Required when `type` is `email`. The email address to receive the alert.

`thresholds.alert_channels[].threshold_name`: Required. The threshold name this channel subscribes to. Supported values are listed in [Supported Threshold Names](#supported-threshold-names), and `*`.

These alert emails are not counted in `email_usage`, to prevent recurring usage caused by the alert itself.

## Event and Hook

When a threshold is triggered, Authgear emits the [`threshold.alert.triggered`](./event.md#thresholdalerttriggered) event.

For `url` alert channels, the configured `url` is used to trigger a hook for this event. Read [Hook](./hook.md) for details.

In `authgear.yaml`, `hook.non_blocking_handlers` can also be used to listen to this event.

Read [Event](./event.md#thresholdalerttriggered) for the event payload.

## Feature Config Merging of Thresholds

The `thresholds` section in feature config can be defined at site level, plan level, or project level.

`thresholds.alert_channels` is overridden by higher-precedence config.

Each `thresholds.<threshold_name>_thresholds` list is also overridden by higher-precedence config.

The precedence order is:

1. project-level feature config
2. plan-level feature config
3. site-level feature config

See the below example:

A feature config of a plan:

```yaml
thresholds:
  alert_channels:
    - type: url
      url: https://internal.authgear.cloud/notification
      threshold_name: sms_usage
  sms_usage_thresholds:
    - threshold: 900
      period: "month"
      type: soft
```

A feature config of a project:

```yaml
thresholds:
  alert_channels:
    - type: url
      url: https://example.com/another_notification
      threshold_name: sms_usage
  sms_usage_thresholds:
    - threshold: 800
      period: "month"
      type: soft
```

The resulting config should be:

```yaml
thresholds:
  alert_channels:
    - type: url
      url: https://example.com/another_notification
      threshold_name: sms_usage
  sms_usage_thresholds:
    - threshold: 800
      period: "month"
      type: soft
```

In other words, project-level feature config overrides plan-level feature config, and plan-level feature config overrides site-level feature config, for both `thresholds.alert_channels` and each `thresholds.<threshold_name>_thresholds` list.

## Portal Configurations

The portal should automatically fill `thresholds.alert_channels` with the project owner's email address.

## Deprecated Usage Configs

The old usage configs in feature config are deprecated in favor of the `thresholds` section and will be removed.

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
