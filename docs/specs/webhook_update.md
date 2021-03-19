## Update webhooks

### Concepts

Instead of having before or after hooks, introduce new concept: blocking and non-blocking hooks.

Actually blocking hooks are the before books and non-blocking is after. But they don't need to be exists in pair anymore.

Blocking hooks accept below response body, operation will be aborted when `is_allowed` is `false`. `title` and `reason` will be shown in auth ui.

```json5
{
  "is_allowed": false,
  "title": "any title",
  "reason": "any string",
}
```

Non-blocking hooks are the same as the after event hooks, response body will be ignored.


### Naming convention

For blocking hooks, we create them by different scenarios, so no specific naming convention.

For non-blocking hooks, the main purposes are for data synchronization / app server notification, we will follow the pattern of `<entity>.<action>.<reason>`.


### Proposed hooks

#### Blocking hooks

- pre_registration
- admin_create_user

#### Non-blocking hooks

- user.created.user_signup
- user.created.admin_api_create
- identity.created.user_create
- identity.created.admin_api_create
- identity.deleted.user_delete
- identity.deleted.admin_api_delete
- identity.updated.user_update
- session.created.user_signup
- session.created.user_login
- session.deleted.user_revoke
- session.deleted.user_logout
- session.deleted.admin_api_revoke
- user.promoted.user_promote

### Webhook Event hooks shape

The current webhook shape looks good to me, i want to keep it with some minor update only.

Quick preview of event:

```json5
{
  "id": "0E1E9537-DF4F-4AF6-8B48-3DB4574D4F24",
  "seq": 435,
  "type": "after_user_create",
  "payload": { /* ... */ },
  "context": { /* ... */ }
}
```

1. `type` will be the new event type
1. Add `locale` to context
1. Remove `reason` from `payload`


Please check the below list for webhook shape
- [Webhook shape](./webhook.md#webhook-event-shape)
- [Payload of events](./webhook.md#blocking-webhook-event-list)


### Discussion questions

1. Seems we cannot solve developers may missed some of the events after we add new events problem, unless we support subscribing event type `user.created.*`. But we cannot do it in my blocking hooks definition, so eventually they will miss. I found the existing services don't have this problem, since they didn't put the reason in the event name, even in the request payload. Developer has no way to know the reason why the event is triggered. e.g.

    - Tested auth0, pre-registration will be called in both user signup and portal create user. No reason in the context, developers have no way to identify the reason.
    - In stripe, [customer.created](https://stripe.com/docs/api/events/types#event_types-customer.created) will be triggered. Seems no way to determine the reason how the customer is created.

    I don't have better solution for this yet, see if you guys have better idea of naming that can fulfill this? Or how to balance the trade off?

1. The current implementation, portal is calling the admin api, so we don't know it is triggered by admin api / admin portal. Do you think we need to distinguish them?
