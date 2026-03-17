# Usage

## Usage Limits

Usage limits are defined in `authgear.features.yaml` with the following object.

```yaml
enabled: true
period: "day" # "month" or "day"
quota: 5
```

## Supported Usage Types

Supported usage types are:

- `user_export`
- `user_import`
- `email`
- `whatsapp`
- `sms`

### Admin API - Export User

```yaml
admin_api:
  user_export_usage:
    enabled: true
    period: "day"
    quota: 5
```

### Admin API - Import User

```yaml
admin_api:
  user_import_usage:
    enabled: true
    period: "day"
    quota: 1000
```

### Messaging - Email

```yaml
messaging:
  email_usage:
    enabled: true
    period: "month"
    quota: 1000
```

### Messaging - Whatsapp

```yaml
messaging:
  whatsapp_usage:
    enabled: true
    period: "month"
    quota: 1000
```

### Messaging - SMS

```yaml
messaging:
  sms_usage:
    enabled: true
    period: "month"
    quota: 1000
```

## Usage Limit Soft Limits

Define soft limits to send when a usage threshold is reached.

```yaml
usage:
  alert:
    enabled: true
    url: https://example.com/your_webhook

messaging:
  sms_usage:
    enabled: true
    period: "day"
    quota: 5
    soft_limits:
      - threshold: 4
```

`usage.alert.enabled`: Optional. Boolean. Explicitly enable or disable usage alerts for this plan or app. This is used to turn off alerts for a specific plan or app even when site-wise usage alert is enabled.

`usage.alert.url`: Required when usage alerts are enabled and any usage limit config contains `soft_limits`. The url we send a request to when any configured soft limit or hard limit is triggered.

`soft_limits`: A list of soft limits to trigger when usage crosses from below to at least the configured threshold.

`soft_limits[].threshold`: Required. Integer. The usage value to trigger this soft limit.

### The Soft Limit Request

We send a HTTP request to the configured `usage.alert.url` when usage alerts are enabled and usage crosses from below to at least the configured threshold.

If `usage.alert.enabled` is `false`, no alert request is sent.

The same event type is also used when usage crosses the hard limit.

The request body follows the [Event](./event.md) specification.

The event type is [`usage.alert.triggered`](./event.md#usagealerttriggered).

### Merging of usage limit soft limits

Soft limits can be defined on plan level feature config, or in project level feature config.

When both are present, the resulting `soft_limits` list is formed by appending the app-level feature config `soft_limits` after the plan-level feature config `soft_limits`.

`usage.alert` is a single shared config across all usage limits. If both plan-level and project-level feature config define it, the project-level value overrides the plan-level value.

See the below example:

A feature config of a plan:

```yaml
messaging:
  sms_usage:
    enabled: true
    period: "month"
    quota: 1000
    soft_limits:
      - threshold: 900
usage:
  alert:
    enabled: true
    url: https://internal.authgear.cloud/notification
```

A feature config of a project:

```yaml
messaging:
  sms_usage:
    soft_limits:
      - threshold: 800
usage:
  alert:
    enabled: false
    url: https://example.com/another_notification
```

The resulting config should be:

```yaml
messaging:
  sms_usage:
    enabled: true
    period: "month"
    quota: 1000
    soft_limits:
      - threshold: 900
      - threshold: 800
usage:
  alert:
    enabled: false
    url: https://example.com/another_notification
```

In other words, `soft_limits` are merged by appending the project-level entries after the plan-level entries, and all merged `soft_limits` share the same effective `usage.alert` config.

When site-wise usage alert is enabled, `usage.alert.enabled: false` in a plan or app explicitly turns off usage alerts for that plan or app.

## Send Usage Alert To Project Owner

We can configure usage alerts to project owners through `authgear.yaml`.

```yaml
usage:
  alert:
    emails:
      - <project_owner_email_1>
      - <project_collaborator_email_2>
  sms:
    soft_limits:
      - threshold: 900
```

`usage.alert.emails`: Optional. A list of email addresses that receive usage alert notifications for this project.

`usage.<name>.soft_limits` has the same shape and semantics as the feature config `soft_limits`. Supported `<name>` values are listed in [Supported Usage Types](#supported-usage-types).

The same values are used in `usage_limit.name` in [`usage.alert.triggered`](./event.md#usagealerttriggered).

If configured, [`usage.alert.triggered`](./event.md#usagealerttriggered) triggers [hooks](./hook.md). Project collaborators can configure deno hooks or webhooks in order to receive usage alerts. If `usage.alert.emails` is set, usage alerts are also sent to the configured email addresses.

### Portal Configurations

The portal should automatically fill `usage.alert.emails` with the project owner's email when usage alert is enabled.
