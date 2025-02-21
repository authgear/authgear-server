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

export default async function (
  e: CustomSMSGatewayPayload
): Promise<CustomSMSGatewayResponse> {
  const response = await fetch("https://some.sms.gateway");
  if (!response.ok) {
    throw new Error("Failed to send sms");
  }
  return {
    code: "ok",
  };
}
```

### Response

The deno hook and webhook shares the same response schema:

```typescript
interface CustomSMSGatewayResponse {
  code:
    | "ok" // Return this code if the sms is delivered successfully
    | "invalid_phone_number" // Return this code if the phone number is invalid
    | "rate_limited" // Return this code if some rate limit is reached and the user should retry the request
    | "authentication_failed" // Return this code if some authentication is failed, and the developer should check the current configurations.
    | "delivery_rejected"; // Return this code if the sms delivery service rejected the request for any reason the user cannot fix by retrying.

  // A string identifying the sms provider.
  // This field is only set by deployment-wise sms gateway.
  provider_name?: string;

  // Error code that could appear on portal to assist debugging.
  // For example, you may put the error code returned by twilio here.
  // The deployment-wise sms gateway always put the error code returned by the sms provider here.
  provider_error_code?: string;

  // This field is only set by deployment-wise sms gateway.
  // The error message of any error occured in the gateway. This will not be exposed to user and is only for debug purpose.
  go_error?: string;

  // This field is only set by deployment-wise sms gateway.
  // The dumped response of the sms provider. This will not be exposed to user and is only for debug purpose.
  dumped_response?: string;
}
```

#### Backward Compatibility

If a deno hook produces an output conforming to the above interface, the delivery result will be determined by the `code`. Else, any function which does not throw error during execution will be treated as success for backward compatibility.

If a webhook produces a json response conforming to the above interface, the delivery result will be determined by the `code`. Else, any webhook which returns a 200-299 status code will be treated as success for backward compatibility.
