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
