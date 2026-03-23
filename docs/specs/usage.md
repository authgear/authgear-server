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

## Usage Soft Limits

Define soft limits to send an alert when a usage threshold is reached.

```yaml
alerts:
  urls:
    - https://example.com/your_webhook
  usages:
    sms:
      soft_limits:
        - threshold: 4
```

`alerts.urlss`: Required when usage alerts are enabled. A list of urls we send requests to when a configured feature config soft limit or the hard limit is triggered. Each url can also be a deno script url.

`alerts.usages.<name>.soft_limits`: A list of soft limits to trigger when usage crosses from below to at least the configured threshold. Supported `<name>` values are listed in [Supported Usage Types](#supported-usage-types).

`alerts.usages.<name>.soft_limits[].threshold`: Required. Integer. The usage value to trigger this soft limit.

### The Alert Request

We send a HTTP request to the configured `alerts.urls` when usage alerts are enabled and usage crosses from below to at least the configured threshold.

The same event type is also used when usage crosses the hard limit.

The request body follows the [Event](./event.md) specification.

The event type is [`usage.alert.triggered`](./event.md#usagealerttriggered).

When usage crosses the hard limit, it also sends alerts to urls configured in `alerts.urls`.

### Merging of usage limit soft limits

Soft limits can be defined on plan level feature config, or in project level feature config.

When both are present, the resulting `alerts.usages.<name>.soft_limits` list is formed by appending the app-level feature config entries after the plan-level feature config entries.

`alerts.urls` follows the same append merge behavior. If both plan-level and project-level feature config define it, the resulting list is formed by appending the project-level entries after the plan-level entries.

See the below example:

A feature config of a plan:

```yaml
alerts:
  urls:
    - https://internal.authgear.cloud/notification
  usages:
    sms:
      soft_limits:
        - threshold: 900
```

A feature config of a project:

```yaml
alerts:
  urls:
    - https://example.com/another_notification
  usages:
    sms:
      soft_limits:
        - threshold: 800
```

The resulting config should be:

```yaml
alerts:
  urls:
    - https://internal.authgear.cloud/notification
    - https://example.com/another_notification
  usages:
    sms:
      soft_limits:
        - threshold: 900
        - threshold: 800
```

In other words, both `alerts.urls` and `alerts.usages.<name>.soft_limits` are merged by appending the project-level entries after the plan-level entries.

## Send Usage Alert To Project Owner

We can configure usage alerts to project owners through `authgear.yaml`.

```yaml
alerts:
  emails:
    - <project_owner_email_1>
    - <project_collaborator_email_2>
  usages:
    sms:
      soft_limits:
        - threshold: 900
```

`alerts.emails`: Optional. A list of email addresses that receive usage alerts for this project.

`alerts.usages.<name>.soft_limits`: A list of soft limits to trigger when usage crosses from below to at least the configured threshold. Supported `<name>` values are listed in [Supported Usage Types](#supported-usage-types).

The same values are used in `usage_limit.name` in [`usage.alert.triggered`](./event.md#usagealerttriggered).

If configured, `alerts.usages.<name>.soft_limits` in `authgear.yaml` trigger [`usage.alert.triggered`](./event.md#usagealerttriggered) and [hooks](./hook.md) when usage reaches a configured soft limit threshold. Project collaborators can configure deno hooks or webhooks in order to receive usage alerts.

When usage reaches the hard limit, it also triggers [`usage.alert.triggered`](./event.md#usagealerttriggered) and [hooks](./hook.md). If `alerts.emails` is set, hard limit alerts are also sent to the configured email addresses. If `alerts.urls` is configured in feature config, the hard limit also triggers those alerts.

### Portal Configurations

The portal should automatically fill `alerts.emails` with the project owner's email when usage alert is enabled.
