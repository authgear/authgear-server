# SMS Gateway

This document describes configuration of sms gateway.

- [Using Custom SMS Gateway](#using-custom-sms-gateway)
  - [Configuration](#configuration)
  - [Webhook Signature](#webhook-signature)
  - [Webhook Body](#webhook-body)
  - [Deno Script](#deno-script)

## Using Custom SMS Gateway

When using custom sms gateway, `messaging.sms_provider` must be set to value `custom` in configuration. A secret with key `sms.custom` must also exist in secret configuration.

### Configuration

The following configs must be specified in `authgear.yaml`:

```yaml
messaging:
  sms_provider: custom
```

- `messaging.sms_provider`
  - Must be `custom` when using custom sms gateway

And the following secrets must be specified in `authgear.secrets.yaml`:

```yaml
- data:
    url: authgeardeno:///deno/sms.ts
    timeout: 10
  key: sms.custom
```

- `data.url`
  - If it is an http / https URL, a webhook will be sent to the URL.
  - If it is an URL with scheme `authgeardeno://`, the referenced deno module will be executed.
- `data.timeout`
  - The request timeout of the webhook request, or the execution timeout of deno hook in second.

### Webhook

When `url` in config is an http / https URL, a request is sent to the specified url.

See [webhook](./hook.md#webhook) for details.

The request body is in json format with the following fields:

- `to`: The recipient of the sms.
- `body`: The body of the sms.

Example:

```json
{
  "to": "+85298765432",
  "body": "You otp is 123456"
}
```

### Deno Hook

When `url` in config is an URL with scheme `authgeardeno://`, the referenced deno module will be executed instead of sending a webhook request.

The argument of the function is an object same as the webhook request body.

See [deno hook](./hook.md#deno-hook) for details.

Example:

```typescript
import { CustomSMSGatewayPayload } from "https://deno.land/x/authgear-deno-hook@0.1.0/mod.ts";

export default async function (e: CustomSMSGatewayPayload): Promise<void> {
  const response = await fetch("https://some.sms.gateway");
  if (!response.ok) {
    throw new Error("Failed to send sms");
  }
}
```
