# Webhook

Webhook is the mechanism to notify external services about events.

  * [Webhook Events](#webhook-events)
  * [Webhook Event Shape](#webhook-event-shape)
    * [Webhook Event Shape Versioning](#webhook-event-shape-versioning)
    * [Webhook Event Context](#webhook-event-context)
  * [Webhook Delivery](#webhook-delivery)
  * [Webhook Event Lifecycle](#webhook-event-lifecycle)
  * [Webhook BEFORE Events](#webhook-before-events)
  * [Webhook AFTER Events](#webhook-after-events)
  * [Webhook Event List](#webhook-event-list)
    * [before_user_create, after_user_create](#before_user_create-after_user_create)
    * [before_identity_create, after_identity_create](#before_identity_create-after_identity_create)
    * [before_identity_update, after_identity_update](#before_identity_update-after_identity_update)
    * [before_identity_delete, after_identity_delete](#before_identity_delete-after_identity_delete)
    * [before_session_create, after_session_create](#before_session_create-after_session_create)
    * [before_session_delete, after_session_delete](#before_session_delete-after_session_delete)
  * [Webhook Event Management](#webhook-event-management)
    * [Webhook Event Alerts](#webhook-event-alerts)
    * [Webhook Past Events](#webhook-past-events)
    * [Webhook Manual Re-delivery](#webhook-manual-re-delivery)
    * [Webhook Delivery Security](#webhook-delivery-security)
      * [Webhook HTTPS](#webhook-https)
      * [Webhook Signature](#webhook-signature)
  * [Webhook Considerations](#webhook-considerations)
    * [Recursive Webhooks](#recursive-webhooks)
    * [Webhook Delivery Reliability](#webhook-delivery-reliability)
    * [Webhook Eventual Consistency](#webhook-eventual-consistency)
    * [Webhook Event Timing](#webhook-event-timing)
    * [CAP Theorem](#cap-theorem)
## Webhook Events

Webhook events are triggered when some mutating operation is performed.

Each operation will trigger two events: BEFORE and AFTER.

- BEFORE event is triggered before the operation is performed. The operation can be aborted by webhook handler.
- AFTER event is triggered after the operation is performed.

Additionally, a `user_sync` event is triggered along with the main event.

BEFORE and AFTER events have the same payload.

## Webhook Event Shape

All webhook events have the following shape:

```json5
{
  "id": "0E1E9537-DF4F-4AF6-8B48-3DB4574D4F24",
  "seq": 435,
  "type": "after_user_create",
  "payload": { /* ... */ },
  "context": { /* ... */ }
}
```

- `id`: The ID of the event.
- `seq`: A monotonically increasing signed 64-bit integer.
- `type`: The type of the webhook event.
- `payload`: The payload of the webhook event, varies with type.
- `context`: The context of the webhook event.

### Webhook Event Shape Versioning

All fields are guaranteed that only backward-compatible changes would be made.

- Existing fields would not be removed or changed in meaning.
- New fields may be added.

### Webhook Event Context

- `timestamp`: signed 64-bit UNIX timestamp of when this event is generated. Retried deliveries do not affect this field.
- `user_id`: The ID of the user associated with the event. It may be absent. For example, the user has not authenticated yet.

## Webhook Delivery

The webhook event is POSTed to the webhook handler endpoint.

The webhook handler endpoint must be an absolute URL.

Each event can have many handlers. The order of delivery is unspecified for AFTER event. BEFORE events are delivered in the source order as in the configuration.

BEFORE events are always delivered before AFTER events.

Webhook handler should be idempotent, since AFTER events may be delivered multiple times due to retries.

Webhook handler must return a status code within the 2xx range. Other status code is considered as a failed delivery.

## Webhook Event Lifecycle

1. Begin transaction
1. Perform operation
1. Deliver BEFORE events to webhook handlers
1. If failed, rollback the transaction.
1. If mutation requested, perform mutation.
1. Commit transaction
1. Deliver AFTER events to webhook handlers

## Webhook BEFORE Events

BEFORE events are delivered to webhook handlers synchronously, right before committing changes to the database.

Webhook handler must respond with a JSON body to indicate whether the operation should continue.

To let the operation to proceed, respond with `is_allowed` being set to `true`.

```json
{
  "is_allowed": true
}
```

To fail the operation, respond with `is_allowed` being set to `false` and a non-empty `reason`. Additional information can be included in `data`.

```json
{
  "is_allowed": false,
  "reason": "any string",
  "data": {
    "foobar": 42
  }
}
```

If any handler fails the operation, the operation is failed. The operation fails with error

```json
{
  "error": {
    "name": "Forbidden",
    "reason": "WebHookDisallowed",
    "info": {
      "reasons": [
        {
          "reason": "any string",
          "data": {
            "foobar": 42
          }
        }
      ]
    }
  }
}
```

The time spent in a BEFORE event delivery must not exceed 5 seconds, otherwise it would be considered as a failed delivery. Also, the total time spent in all deliveries of the event must not exceed 10 seconds, otherwise it would also be considered as a failed delivery. Both timeouts are configurable.

BEFORE events are not persisted, and their failed deliveries are not retried.

A failed operation does not trigger AFTER events.

## Webhook AFTER Events

AFTER events are delivered to webhook handlers asynchronously after the operation is performed (i.e. committed into the database).

The time spent in an AFTER event delivery must not exceed 60 seconds, otherwise it would be considered as a failed delivery.

All AFTER events with registered webhook handlers are persisted into the database, with minimum retention period of 30 days.

The response body of AFTER event webhook handler is ignored.

If any delivery failed, all deliveries will be retried after some time, regardless of whether some deliveries may have succeeded. The retry is performed with a variant of exponential back-off algorithm. If `Retry-After:` HTTP header is present in the response, the delivery will not be retried before the specific time.

If the delivery keeps on failing after 3 days from the time of first attempted delivery, the event will be marked as permanently failed and will not be retried automatically.

## Webhook Event List

### before_user_create, after_user_create

When a new user is being created.

```json5
{
  "payload": {
    "user": { /* ... */ },
    "identities": [ { /* ... */ } ]
  }
}
```

### before_identity_create, after_identity_create

When a new identity is being created for an existing user. So it does not trigger together with `before_user_create` and `after_user_create`.

```json5
{
  "payload": {
    "user": { /* ... */ },
    "identity": { /* ... */ }
  }
}
```

### before_identity_update, after_identity_update

When an identity is being updated.

```json5
{
  "payload": {
    "user": { /* ... */ },
    "old_identity": { /* ... */ },
    "new_identity": { /* ... */ }
  }
}
```

### before_identity_delete, after_identity_delete

When an identity is being deleted from an existing user.

```json5
{
  "payload": {
    "user": { /* ... */ },
    "identity": { /* ... */ }
  }
}
```

### before_session_create, after_session_create

When a session is being created for a new user or an existing user.

```json5
{
  "payload": {
    "reason": "signup",
    "user": { /* ... */ },
    "identity": { /* ... */ },
    "session": { /* ... */ }
  }
}
```

- `reason`: The reason for the creation of the session, can be `signup` or `login`.

### before_session_delete, after_session_delete

When a session is being deleted from an existing user, e.g. logging out.

```json5
{
  "payload": {
    "reason": "logout",
    "user": { /* ... */ },
    "session": { /* ... */ }
  }
}
```

- `reason`: The reason for the deletion of the session, can be `logout`.

## Webhook Event Management

### Webhook Event Alerts

If an event delivery is permanently failed, an ERROR log is generated to notify developers.

### Webhook Past Events

An API is provided to list past events. This can be used to reconcile self-managed database with the failed events.

> NOTE: BEFORE events are not persisted, regardless of success or failure.

### Webhook Manual Re-delivery

The developer can manually trigger a re-delivery of failed event, bypassing the retry interval limit.

> NOTE: BEFORE events cannot be re-delivered.

### Webhook Delivery Security

#### Webhook HTTPS

Webhook handlers must be HTTPS. This ensures integrity and confidentiality of the delivery.

#### Webhook Signature

Each webhook event request is signed with a secret key shared between Authgear and the webhook handler. The developer must validate the signature and reject requests with invalid signature to ensure the request originates from Authgear.

The signature is calculated as the hex encoded value of HMAC-SHA256 of the request body.

The signature is included in the header `x-authgear-body-signature:`.

> For advanced end-to-end security scenario, some network admin may wish to
> use mTLS for authentication. It is not supported at the moment.

## Webhook Considerations

### Recursive Webhooks

A ill-designed web-hook handler may be called recursively. For example, calling api that will trigger another events.

The developer is responsible for ensuring that:
- webhook handlers would not be called recursively; or
- recursive web-hook handlers have well-defined termination condition.

### Webhook Delivery Reliability

The main purpose of webhook is to allow external services to observe state changes.

Therefore, AFTER events are persistent, immutable, and delivered reliably. Otherwise, external services may observe inconsistent changes.

It is not recommended to perform side effects in BEFORE event handlers. Otherwise, the developer should consider how to compensate for the side effects of potential failed operation.

### Webhook Eventual Consistency

Fundamentally, webhook is a distributed system. When webhook handlers have side effects, we need to choose between guaranteeing consistency or availability of the system (See [CAP Theorem](#cap-theorem)).

We decided to ensure the availability of the system. To maintain consistency, the developer should take eventual consistency into account when designing their system.

The developer should regularly check the past events for unprocessed events to ensure consistency.

### Webhook Event Timing

There are four theoretically delivery timing of events: sync BEFORE, async BEFORE, sync AFTER and async AFTER.

Async BEFORE is mostly useless. The operation may not be successful, and the handler cannot affect the operation. So async BEFORE events do not exist.

Sync AFTER cannot be used safely due to the following reasoning:

- If it is not within the operation transaction, async AFTER can be used instead.
- If it is within the operation transaction and has no side effects, sync BEFORE can be used instead.
- If it is within the operation transaction and has side effects, async AFTER should be used instead.

So sync AFTER events do not exist.

### CAP Theorem

To simplify, the CAP theorem states that a distributed data store can satisfy
only two of the three properties simultaneously:
- Consistency
- Availability
- Network Partition Tolerance

Since network partition cannot be avoided practically, distributed system would
need to choose between consistency and availability. Most microservice
architecture prefer availability over strong consistency, and instead application
state is eventually consistent.
