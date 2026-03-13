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

## Usage Limits Notifications

Define notifications to send when a usage threshold reached.

```yaml
enabled: true
period: "day"
quota: 5
notifications:
  - threshold:
      percentage: 90
    interval: 24h
    url: https://example.com/your_webhook
```

`notifications`: A list of notifications to trigger when usage reaches the configured threshold.

`notifications[].threshold`: Required. The threshold to trigger the notification. Only percentage threshold is supported at the moment.
`notifications[].threshold.percentage`: Required. Integer. 1 - 100. The percentage of the usage limit to trigger this notification.
`notifications[].interval`: Optional. Duration string. Default `24h`. The minimal interval to wait before the next trigger of the same notification.
`notifications[].url`: Required. The url we send a request to when the notification is triggered.

### The Notification Request

We send a HTTP request to the configured `notifications[].url` whenever a threshold is reached.

The request body follows the [Event](./event.md) specification.

The event type is [`usage.threshold.reached`](./event.md#usagethresholdreached).

### Merging of usage limit notifications

Notifications can be defined on plan level feature config, or in project level feature config.

Notifications defined in the two level will be merged.

See the below example:

A feature config of a plan:

```yaml
messaging:
  sms_usage:
    enabled: true
    period: "month"
    quota: 1000
    notifications:
      - threshold:
          percentage: 90
        url: https://internal.authgear.cloud/notification
```

A feature config of a project:

```yaml
messaging:
  sms_usage:
    notifications:
      - threshold:
          percentage: 80
        url: https://example.com/another_notification
```

The resulting config should be:

```yaml
messaging:
  sms_usage:
    enabled: true
    period: "month"
    quota: 1000
    notifications:
      - threshold:
          percentage: 90
        url: https://internal.authgear.cloud/notification
      - threshold:
          percentage: 80
        url: https://example.com/another_notification
```
