* [SMS Gateway](#sms-gateway)
  * [Configuration](#configuration)
    * [Webhook](#webhook)
    * [Deno Hook](#deno-hook)
    * [Request](#request)
    * [Response](#response)
      * [Backward Compatibility](#backward-compatibility)

# SMS Gateway

This document describes configuration of sms gateway.

## Configuration

The following configs must be specified in `authgear.yaml`:

```yaml
messaging:
  sms_gateway:
    provider: custom
    use_config_from: authgear.secrets.yaml
```

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

### Request

The request is a JSON object. Its TypeScript equivalent definition is as follows:

```typescript
interface CustomSMSGatewayPayload {
  // The recipient phone number in E.164 format.
  // For example, "+85298765432"
  to: string;

  // The fully formatted message body ready to be sent.
  // This is already localized.
  // For example, "123456 is your Myapp verification code"
  body: string;

  // IETF BCP 47 language tag describing the language of body.
  language_tag: string;

  // The Authgear project ID.
  // For example, "myapp"
  app_id: string;

  // One of the listed literal.
  template_name:
    | "authenticate_primary_oob_sms.txt"
    | "authenticate_secondary_oob_sms.txt"
    | "forgot_password_oob_sms.txt"
    | "forgot_password_sms.txt"
    | "setup_primary_oob_sms.txt"
    | "setup_secondary_oob_sms.txt"
    | "verification_sms.txt"
    ;

  template_variables: {
    // This is present when template_name is
    // - "authenticate_primary_oob_sms.txt"
    // - "authenticate_secondary_oob_sms.txt"
    // - "forgot_password_oob_sms.txt"
    // - "setup_primary_oob_sms.txt"
    // - "setup_secondary_oob_sms.txt"
    // - "verification_sms.txt"
    code?: string;
    // This is present when template_name is
    // - "forgot_password_sms.txt"
    link?: string;
  };
}
```

#### Use case 1: Simply send the body

In case you have configured the SMS template in the portal, you can just send `body` to `to`, ignoring all other fields.

#### Use case 2: Send SMS to +86 phone numbers

You usually cannot send arbitrary SMS messages to +86 phone numbers.
The service provider capable of sending SMS messages to +86 phone numbers typically require you to
register pre-defined templates.

You can use `template_name` to select a suitable pre-defined template registered in your service provider.
And then use `template_variables` to interpolate the template.

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
  // The error message of any error occurred in the gateway. This will not be exposed to user and is only for debug purpose.
  go_error?: string;

  // This field is only set by deployment-wise sms gateway.
  // The dumped response of the sms provider. This will not be exposed to user and is only for debug purpose.
  dumped_response?: string;
}
```

#### Backward Compatibility

If a deno hook produces an output conforming to the above interface, the delivery result will be determined by the `code`. Else, any function which does not throw error during execution will be treated as success for backward compatibility.

If a webhook produces a json response conforming to the above interface, the delivery result will be determined by the `code`. Else, any webhook which returns a 200-299 status code will be treated as success for backward compatibility.
