- [Webhook](#webhook)
  * [Webhook Delivery](#webhook-delivery)
  * [Webhook Event Lifecycle](#webhook-event-lifecycle)
  * [Webhook Blocking Events](#webhook-blocking-events)
  * [Webhook Non-blocking Events](#webhook-non-blocking-events)
    + [Future works of non-blocking events](#future-works-of-non-blocking-events)
  * [Webhook Event Management](#webhook-event-management)
    + [Webhook Event Alerts](#webhook-event-alerts)
    + [Webhook Past Events](#webhook-past-events)
    + [Webhook Manual Re-delivery](#webhook-manual-re-delivery)
    + [Webhook Delivery Security](#webhook-delivery-security)
      - [Webhook HTTPS](#webhook-https)
      - [Webhook Signature](#webhook-signature)
  * [Webhook Considerations](#webhook-considerations)
    + [Recursive Webhooks](#recursive-webhooks)
    + [Webhook Delivery Reliability](#webhook-delivery-reliability)
    + [Webhook Eventual Consistency](#webhook-eventual-consistency)
    + [CAP Theorem](#cap-theorem)
  * [authgear.yaml](#authgearyaml)

# Webhook

Webhook is the mechanism to notify external services about events.

For the definition of events, see [Event](./event.md)

## Webhook Delivery

The webhook event is POSTed to the webhook handler endpoint.

The webhook handler endpoint must be an absolute URL.

Each event can have many handlers. The order of delivery is unspecified for non-blocking event. Blocking events are delivered in the source order as in the configuration.

Webhook handler should be idempotent, since non-blocking events may be delivered multiple times due to retries.

Webhook handler must return a status code within the 2xx range. Other status code is considered as a failed delivery.

## Webhook Event Lifecycle

1. Begin transaction
1. Perform operation
1. Deliver blocking events to webhook handlers
1. If failed, rollback the transaction.
1. Commit transaction
1. Deliver non-blocking events to webhook handlers

## Webhook Blocking Events

Blocking events are delivered to webhook handlers synchronously, right before committing changes to the database.

Webhook handler must respond with a JSON body to indicate whether the operation should continue.

To let the operation to proceed, respond with `is_allowed` being set to `true`.

```json
{
  "is_allowed": true
}
```

To fail the operation, respond with `is_allowed` being set to `false` and a non-empty `title` and `reason`.

```json
{
  "is_allowed": false,
  "title": "any title",
  "reason": "any string"
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
          "title": "any title",
          "reason": "any string"
        }
      ]
    }
  }
}
```

The time spent in a blocking event delivery must not exceed 5 seconds, otherwise it would be considered as a failed delivery. Also, the total time spent in all deliveries of the event must not exceed 10 seconds, otherwise it would also be considered as a failed delivery. Both timeouts are configurable.

Blocking events are not persisted, and their failed deliveries are not retried.

## Webhook Non-blocking Events

Non-blocking events are delivered to webhook handlers asynchronously after the operation is performed (i.e. committed into the database).

The time spent in an non-blocking event delivery must not exceed 60 seconds, otherwise it would be considered as a failed delivery.

The response body of non-blocking event webhook handler is ignored.

### Future works of non-blocking events

All non-blocking events with registered webhook handlers are persisted into the database, with minimum retention period of 30 days.

If any delivery failed, all deliveries will be retried after some time, regardless of whether some deliveries may have succeeded. The retry is performed with a variant of exponential back-off algorithm. If `Retry-After:` HTTP header is present in the response, the delivery will not be retried before the specific time.

If the delivery keeps on failing after 3 days from the time of first attempted delivery, the event will be marked as permanently failed and will not be retried automatically.

## Webhook Event Management

### Webhook Event Alerts

If an event delivery is permanently failed, an ERROR log is generated to notify developers.

### Webhook Past Events

An API is provided to list past events. This can be used to reconcile self-managed database with the failed events.

> NOTE: Blocking events are not persisted, regardless of success or failure.

### Webhook Manual Re-delivery

The developer can manually trigger a re-delivery of failed event, bypassing the retry interval limit.

> NOTE: Blocking events cannot be re-delivered.

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

It is not recommended to perform side effects in blocking event handlers. Otherwise, the developer should consider how to compensate for the side effects of potential failed operation.

### Webhook Eventual Consistency

Fundamentally, webhook is a distributed system. When webhook handlers have side effects, we need to choose between guaranteeing consistency or availability of the system (See [CAP Theorem](#cap-theorem)).

We decided to ensure the availability of the system. To maintain consistency, the developer should take eventual consistency into account when designing their system.

The developer should regularly check the past events for unprocessed events to ensure consistency.

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

## authgear.yaml

```
hook:
  blocking_handlers:
    - event: "user.pre_create"
      url: 'https://myapp.com/check_user_create'
  non_blocking_handlers:
    - events: ["*"]
      url: 'https://myapp.com/all_events'
    - events: ["user.created"]
      url: 'https://myapp.com/sync_user_creation'
```
