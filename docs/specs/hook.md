- [Hooks](#hooks)
- [Kinds of Hooks](#kinds-of-hooks)
- [Event Delivery](#event-delivery)
- [Lifecycle of Event Delivery](#lifecycle-of-event-delivery)
- [Blocking Events](#blocking-events)
  * [Blocking Event Mutations](#blocking-event-mutations)
  * [Blocking Event Response](#blocking-event-response)
- [Non-blocking Events](#non-blocking-events)
  * [Future works of non-blocking events](#future-works-of-non-blocking-events)
- [Webhook](#webhook)
  * [Webhook Signature](#webhook-signature)
- [Hooks Event Management](#hooks-event-management)
  * [Hooks Event Alerts](#hooks-event-alerts)
  * [Hooks Past Events](#hooks-past-events)
  * [Hooks Manual Re-delivery](#hooks-manual-re-delivery)
- [Considerations](#considerations)
  * [Recursive Hooks](#recursive-hooks)
  * [Delivery Reliability](#delivery-reliability)
  * [Eventual Consistency](#eventual-consistency)
  * [CAP Theorem](#cap-theorem)
- [Configuration in `authgear.yaml`](#configuration-in-authgearyaml)
- [Blocking Event Actions](#blocking-event-actions)
  * [Responding to `user.identified`](#responding-to-user.identified)

# Hooks

Hooks is a mechanism to notify external services about [events](./event.md).

# Kinds of Hooks

There are 2 kinds of Hooks.

- [Webhook](#webhook)
- [Deno Hook](#deno-hook)

# Event Delivery

Each event can have many Hooks. The order of delivery is unspecified for [non-blocking event](#non-blocking-events). [Blocking events](#blocking-events) are delivered in the source order as in the configuration.

Hooks should be idempotent, since non-blocking events may be delivered multiple times due to retries.

# Lifecycle of Event Delivery

1. Begin transaction
1. Perform operation
1. Deliver blocking events to Hooks.
1. If failed, rollback the transaction.
1. Perform mutations
1. Commit transaction
1. Deliver non-blocking events to Hooks.

# Blocking Events

Blocking events are delivered to hooks synchronously, right before committing changes to the database.

Hooks must return a JSON document to indicate whether the operation should continue.

To let the operation to proceed, respond with `is_allowed` set to `true`.

```json
{
  "is_allowed": true
}
```

To fail the operation, respond with `is_allowed` set to `false`, and a non-empty `title` and `reason`.

```json
{
  "is_allowed": false,
  "title": "any title",
  "reason": "any string"
}
```

If any hook fails the operation, the operation is failed. The operation fails with error

```json
{
  "error": {
    "name": "Forbidden",
    "reason": "WebHookDisallowed",
    "info": {
      "reasons": [
        {
          "title": "any title",
          "reason": "any string"
        }
      ]
    }
  }
}
```

> For backward compatibility, the reason is called "WebHookDisallowed".

The time spent in a blocking event delivery must not exceed 5 seconds, otherwise it will be considered as a failed delivery. Also, the total time spent in all deliveries of the event must not exceed 10 seconds, otherwise it would also be considered as failed delivery. Both timeouts are configurable.

Blocking events are not persisted, and their failed deliveries are not retried.

## Blocking Event Mutations

Hooks can optionally mutate the object in the Event payload.

Hooks cannot request mutation if the operation is failed by them.

Hooks specify the mutations in the JSON document they return.

Given the event

```json
{
  "payload": {
    "user": {
      "standard_attributes": {
        "name": "John"
      }
    }
  }
}
```

Hooks can mutate the user with the following JSON document.

```json
{
  "is_allowed": true,
  "mutations": {
    "user": {
      "standard_attributes": {
        "name": "Jane"
      },
      "roles": ["store_manager", "salesperson"],
      "groups": ["manager"]
    }
  }
}
```

Objects not appearing in `mutations` are left intact.

The mutated objects do NOT merge with the original ones.

The mutated payload are NOT validated and are propagated along the Hooks chain.
The payload will only be validated after traversing the Hooks chain.

Mutations do NOT generate extra events to avoid infinite loop.

Currently, only `standard_attributes`, `custom_attributes`, `roles` and `groups` of the user object are mutable.

## Blocking Event Action

Hooks can response to a blocking event with a specific action if the corresponding event supports it.

```json5
{
  "is_allowed": true,
  "action": { /* */ }
}
```

Interpretation and format of the `action` object is different in different events. Please read the corresponding spec for the meaning and format of `action` of each event:

- [`user.identified`](#responding-to-user-identified)

Events not listed above does not support the `action` object. The `action` key in the response of the hook will be simply ignored.

# Non-blocking Events

Non-blocking events are delivered to Hooks asynchronously after the operation is performed (i.e. changes committed to the database).

The time spent in an non-blocking event delivery must no exceed 60 seconds, otherwise it would be considered as a failed delivery.

The return value of non-blocking event Hooks is ignored.

## Future works of non-blocking events

All non-blocking events with registered Hooks are persisted in the database, with minimum retention period of 30 days.

If any delivery failed, all deliveries will be retried after some time, regardless of whether some deliveries may have succeeded. The retry is performed with a variant of exponential back-off algorithm. Specifically for Webhooks, if `Retry-After:` HTTP header is present in the response, the delivery will not be retried before the specific time.

If the delivery keeps on failing after 3 days from the time of first attempted delivery, the event will be marked as permanently failed and will not be retried automatically.

# Webhook

Webhook is a kind of Hook via the HTTPS protocol. This ensures integrity and confidentiality of the delivery.

Events are POSTed to the Webhook.

The endpoint of the Webhook must be an absolute URL.

The Webhook must return a status code within the 2xx range. Other status code is considered as a failed delivery.

## Webhook Signature

Each request is signed with a secret key shared between Authgear and the Webhook. The developer must validate the signature and reject requests with invalid signature to ensure the request originates from Authgear.

The signature is calculated as the hex encoded value of HMAC-SHA256 of the request body.

The signature is included in the header `x-authgear-body-signature:`.

> For advanced end-to-end security scenario, some network admin may wish to
> use mTLS for authentication. It is not supported at the moment.

# Deno Hook

Deno Hook is a kind of Hook in form of a TypeScript / JavaScript module. The module is executed by [Deno](https://deno.land/).

The module **MUST** have a [default export](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Statements/export#description) of a function taking 1 argument. The argument is the event payload. The function can either be synchronous or asynchronous. An asynchronous function is a function returning [Promise](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise), or an [async function](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Statements/async_function)

If the Deno Hook is registered for a [blocking event](#blocking-events), the function **MUST** return a value according to the [specification](#blocking-events).

If the Deno hook is registered for a [non-blocking event](#non-blocking-events), the return value is ignored.

Program run with Deno has [no access](https://deno.land/manual@v1.27.2/basics/permissions) to file, network or environment by default. In case of Deno Hook, it only has access to external network. Other access is blocked. For example, A Deno Hook is **NOT** allowed to read or write the file system.

The stdout and the stderr of the Deno Hook is ignored currently.
The arguments and the stdin is intentionally unspecified. A Deno Hook **MUST NOT** assume anything on them.

Deno Hooks are stored along with other app resources, such as `authgear.yaml` and templates.
The size limit of a Deno Hook is 100KiB. A module larger than 100KiB **CANNOT** be registered as a Deno Hook.

## Example of a blocking Deno Hook

```typescript
import { HookEvent, HookResponse } from "https://deno.land/x/authgear-deno-hook@0.1.0/mod.ts"

export default async function(e: HookEvent): Promise<HookResponse> {
  return { is_allowed: true };
}
```

## Example of a non-blocking Deno Hook

```typescript
import { HookEvent } from "https://deno.land/x/authgear-deno-hook@0.1.0/mod.ts"

export default async function(e: HookEvent): Promise<void> {
  // Do something with e.
}
```

# Hooks Event Management

## Hooks Event Alerts

If an event delivery is permanently failed, an ERROR log is generated to notify developers.

## Hooks Past Events

An API is provided to list past events. This can be used to reconcile self-managed database with the failed events.

> NOTE: Blocking events are not persisted, regardless of success or failure.

## Hooks Manual Re-delivery

The developer can manually trigger a re-delivery of failed event, bypassing the retry interval limit.

> NOTE: Blocking events cannot be re-delivered.

# Considerations

## Recursive Hooks

An ill-designed Hook may be triggered recursively. For example, calling API that will trigger other events.

The developer is responsible for ensuring that:
- Hooks would not be triggered recursively; or
- Recursive Hooks have well-defined termination condition.

## Delivery Reliability

The main purpose of Hooks is to allow external services to observe state changes.

Therefore, AFTER events are persistent, immutable, and delivered reliably. Otherwise, external services may observe inconsistent changes.

It is not recommended to perform side effects in blocking event Hooks. Otherwise, the developer should consider how to compensate for the side effects of potential failed operation.

## Eventual Consistency

Fundamentally, Hooks is a distributed system. When Hooks have side effects, we need to choose between guaranteeing consistency or availability of the system (See [CAP Theorem](#cap-theorem)).

We decided to ensure the availability of the system. To maintain consistency, the developer should take eventual consistency into account when designing their system.

The developer should regularly check the past events for unprocessed events to ensure consistency.

## CAP Theorem

To simplify, the CAP theorem states that a distributed data store can satisfy
only two of the three properties simultaneously:
- Consistency
- Availability
- Network Partition Tolerance

Since network partition cannot be avoided practically, distributed system would
need to choose between consistency and availability. Most microservice
architecture prefer availability over strong consistency, and instead application
state is eventually consistent.

# Configuration in `authgear.yaml`

```
hook:
  blocking_handlers:
    - event: "user.pre_create"
      url: "https://myapp.com/check_user_create"
    - event: "user.pre_create"
      url: "authgeardeno:///deno/randomstring.ts"
  non_blocking_handlers:
    - events: ["*"]
      url: 'https://myapp.com/all_events'
    - events: ["*"]
      url: "authgeardeno:///deno/randomstring.ts"
    - events: ["user.created"]
      url: 'https://myapp.com/sync_user_creation'
```

# Blocking Event Actions

This section describe the meaning of `action` to different type of blocking events.

## Responding to `user.identified`

### Usecases

#### Force user to use a preferred authentication method if available

Assume you have a portal app which supports two signup methods: Email with Password, and Google oauth login.

User of the portal can signup to the portal by one of the two methods, and login with that method afterwards.

If the user has signed up with email with password, the user can connect to a Google account at any time after he logged in to the portal (For example, through [Account Linking](./account-linking.md)). In your persepctive, Google oauth login is the preferred login method because it is more secure. Therefore once user connected their google account, the portal should only accept google oauth login for the same account, and do not accept email with password logins.

This can be done by implementing a hook for the `user.identified` event:

1. The hook will be triggered once any user is trying to login with any identifier.
2. Get the identity which the user is trying to use in the login flow from the event payload `identity`. If it is an `oauth` identity, he is already using Google oauth logic which you want, so stop and return `{ "is_allowed": true }` as the hook respond.
3. Get all identities of the user who is trying to login, and find if there is at least one `oauth` identity. If no, it means the user is unable to use Google oauth login, so you should allow him to continue with email with password. Return `{ "is_allowed": true }` as the hook respond and stop.
4. Now, we know that the user already has one Google oauth identity, but he is trying to use another identity to login. You want him to use Google login instead. So respond the following result to block the user from logging in:

    ```json
    {
      "is_allowed": false,
      "title": "Please use Google Login instead",
      "reason": "Please use Google Login instead"
    }
    ```

    However, the user will see an error message in the UI and this might not be a good UX. So, use the `action` object to switch the user to the login flow with google login selected by responding:

    ```json
    {
      "is_allowed": true,
      "action": {
        "type": "switch_authentication_flow",
        "authentication_flow": {
          "name": "default",
          "type": "login",
          "input": {
            "identification": "oauth",
            "alias": "google",
            "redirect_uri": "http://example.com/sso/oauth2/callback/google"
          }
        }
      }
    }
    ```

    The fields inside the action object has the following meansings:

      - `type`: Type of the action object defines what type of action it is. The only possible values are:
          - `switch_authentication_flow`: Switch the user to a new authentication flow. The `authentication_flow` object must be provided to specify the target authentication flow to switch to.
      - `authentication_flow`: When `action.type` is `switch_authentication_flow`, this object must be provided. It is the target authentication flow to switch to.
          - `authentication_flow.name`: The name of the target flow. Please read the [authentication flow spec](./authentication-flow-api-reference.md) for details.
          - `authentication_flow.type`: The type of the target flow. Please read the [authentication flow spec](./authentication-flow-api-reference.md) for details.
          - `authentication_flow.input`: The initial input to feed into the new flow. Please read the [input and output section of authentication flow spec](./authentication-flow-api-reference.md#reference-on-input-and-output) for details.

    So the above `action` actually means:

      1. Switch the user to a new authentication flow, with flow type `login` and flow name `default`.
      2. In the new authentication flow, proceed the flow with an input, which selects Google oauth as the identification option.
      3. In the UI, the user should be redirected to the Google oauth login page, and continue the flow.
