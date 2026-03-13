# Usage

## Usage Limits

Usage limits are defined in `authgear.features.yaml` with the following object.

```yaml
enabled: true
period: "day" # "month" or "day"
quota: 5
```

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
enabled: true
period: "day"
quota: 5
soft_limits:
  - threshold: 4
    interval: 24h
    url: https://example.com/your_webhook
```

`soft_limits`: A list of soft limits to trigger when usage reaches the configured threshold.

`soft_limits[].threshold`: Required. Integer. The usage value to trigger this soft limit.
`soft_limits[].interval`: Optional. Duration string. Default `24h`. The minimal interval to wait before the next trigger of the same soft limit.
`soft_limits[].url`: Required. The url we send a request to when the soft limit is triggered.

### The Soft Limit Request

We send a HTTP request to the configured `soft_limits[].url` whenever a threshold is reached.

The request body follows the [Event](./event.md) specification.

The event type is [`usage.soft_limit.reached`](./event.md#usagesoft_limitreached).

### Merging of usage limit soft limits

Soft limits can be defined on plan level feature config, or in project level feature config.

Soft limits defined in the two level will be merged.

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
        url: https://internal.authgear.cloud/notification
```

A feature config of a project:

```yaml
messaging:
  sms_usage:
    soft_limits:
      - threshold: 800
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
        url: https://internal.authgear.cloud/notification
      - threshold: 800
        url: https://example.com/another_notification
```
