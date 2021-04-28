# Webhook

Webhook is the mechanism to notify external services about events.

  * [Webhook Events](#webhook-events)
  * [Webhook Event Shape](#webhook-event-shape)
    * [Webhook Event Shape Versioning](#webhook-event-shape-versioning)
    * [Webhook Event Context](#webhook-event-context)
  * [Webhook Delivery](#webhook-delivery)
  * [Webhook Event Lifecycle](#webhook-event-lifecycle)
  * [Webhook Blocking Events](#webhook-blocking-events)
  * [Webhook Non-blocking Events](#webhook-non-blocking-events)
  * [Webhook Event List](#webhook-event-list)
    * [Blocking Events](#blocking-events)
    * [Non-blocking Events](#non-blocking-events)
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
    * [CAP Theorem](#cap-theorem)
  * [authgear.yaml](#authgear.yaml)
## Webhook Events

Webhook events are triggered when some mutating operation is performed.

Each operation will trigger two events: Blocking and Non-blocking.

- Blocking event is triggered before the operation is performed. The operation can be aborted by webhook handler.
- Non-blocking event is triggered after the operation is performed.

## Webhook Event Shape

All webhook events have the following shape:

```json5
{
  "id": "0E1E9537-DF4F-4AF6-8B48-3DB4574D4F24",
  "seq": 435,
  "type": "user.created",
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
- `preferred_languages`: User preferred languages which are inferred from the request. Return values of `ui_locales` query if it is provided in auth ui, otherwise return languages in `Accept-Language` request header.
- `language`: User locale which is derived based on user's preferred languages and app's languages config.
- `triggered_by`: Triggered by indicates who triggered the events, values can be `user` or `admin_api`. `user` means it is triggered by user in auth ui. `admin_api` means it is triggered by admin api or admin portal.

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


## Webhook Event List

### Blocking Events

- [user.pre_create](#userpre_create)

### Non-blocking Events

- [user.created](#usercreated)
- [user.authenticated](#userauthenticated)
- [user.anonymous.promoted](#useranonymouspromoted)
- [identity.email.added](#identityemailadded)
- [identity.email.removed](#identityemailremoved)
- [identity.email.updated](#identityemailupdated)
- [identity.phone.added](#identityphoneadded)
- [identity.phone.removed](#identityphoneremoved)
- [identity.phone.updated](#identityphoneupdated)
- [identity.username.added](#identityusernameadded)
- [identity.username.removed](#identityusernameremoved)
- [identity.username.updated](#identityusernameupdated)
- [identity.oauth.connected](#identityoauthconnected)
- [identity.oauth.disconnected](#identityoauthdisconnected)

### user.pre_create

Occurs right before the user creation. User can be created by user signup, user signup as anonymous user, admin api or admin portal create an user.
Operation can be aborted by providing specific response in your webhook, details see [Webhook Blocking Events](#webhook-blocking-events).

```json5
{
  "payload": {
    "user": { /* ... */ },
    "identities": [ { /* ... */ } ],
    "state": ""
  }
}
```

- `state`: OIDC state if the signup is triggered through authorize endpoint.

### user.created

Occurs after a new user is created. User can be created by user signup, user signup as anonymous user, admin api or admin portal create an user.

```json5
{
  "payload": {
    "user": { /* ... */ },
    "identities": [ { /* ... */ } ]
  }
}
```

### user.authenticated

Occurs after user logged in.

```json5
{
  "payload": {
    "user": { /* ... */ },
    "session": { /* ... */ }
  }
}
```

### user.anonymous.promoted

Occurs whenever an anonymous user is promoted to normal user.

```json5
{
  "payload": {
    "anonymous_user": { /* ... */ },
    "user": { /* ... */ },
    "identities": [{ /* ... */ }]
  }
}
```

### identity.email.added

Occurs when a new email is added to existing user. Email can be added by user in setting page, added by admin through admin api or portal.

```json5
{
  "payload": {
    "user": { /* ... */ },
    "identity": { /* ... */ }
  }
}
```


### identity.email.removed

Occurs when email is removed from existing user. Email can be removed by user in setting page, removed by admin through admin api or portal.

```json5
{
  "payload": {
    "user": { /* ... */ },
    "identity": { /* ... */ }
  }
}
```


### identity.email.updated

Occurs when email is updated. Email can be updated by user in setting page.

```json5
{
  "payload": {
    "user": { /* ... */ },
    "old_identity": { /* ... */ },
    "new_identity": { /* ... */ }
  }
}
```

### identity.phone.added

Occurs when a new phone number is added to existing user. Phone number can be added by user in setting page, added by admin through admin api or portal.

```json5
{
  "payload": {
    "user": { /* ... */ },
    "identity": { /* ... */ }
  }
}
```


### identity.phone.removed

Occurs when phone number is removed from existing user. Phone number can be removed by user in setting page, removed by admin through admin api or portal.

```json5
{
  "payload": {
    "user": { /* ... */ },
    "identity": { /* ... */ }
  }
}
```


### identity.phone.updated

Occurs when phone number is updated. Phone number can be updated by user in setting page.


```json5
{
  "payload": {
    "user": { /* ... */ },
    "old_identity": { /* ... */ },
    "new_identity": { /* ... */ }
  }
}
```

### identity.username.added

Occurs when a new username is added to existing user. Username can be added by user in setting page, added by admin through admin api or portal.

```json5
{
  "payload": {
    "user": { /* ... */ },
    "identity": { /* ... */ }
  }
}
```


### identity.username.removed

Occurs when username is removed from existing user. Username can be removed by user in setting page, removed by admin through admin api or portal.

```json5
{
  "payload": {
    "user": { /* ... */ },
    "identity": { /* ... */ }
  }
}
```


### identity.username.updated

Occurs when username is updated. Username can be updated by user in setting page.

```json5
{
  "payload": {
    "user": { /* ... */ },
    "old_identity": { /* ... */ },
    "new_identity": { /* ... */ }
  }
}
```

### identity.oauth.connected

Occurs when user connected to a new OAuth provider.

```json5
{
  "payload": {
    "user": { /* ... */ },
    "identity": { /* ... */ }
  }
}
```


### identity.oauth.disconnected

Occurs when user disconnected from an OAuth provider. It can be done by user disconnected OAuth provider in the setting page, or admin removed OAuth identity through admin api or portal.

```json5
{
  "payload": {
    "user": { /* ... */ },
    "identity": { /* ... */ }
  }
}
```




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
