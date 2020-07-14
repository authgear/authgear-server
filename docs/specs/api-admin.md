# Admin API

> This section is purely imaginary

Admin API is intended to be consumed at server-side.

  * [Event Management API](#event-management-api)
    * [GET /admin/events](#get-adminevents)
    * [POST /admin/events/{seq}/retry](#post-admineventsseqretry)

## Event Management API

### GET /admin/events

Return a list of past events.

Query parameters:

- `cursor`: the `seq`. If omitted, the oldest events are returned.
- `limit`: optional integer within the range [1,20].
- `status`: optional comma-separated string of event statues to filter.

Response:

```json5
{
  "events": [
    {
      "status": "success",
      "event": { /* ... */ }
    }
  ]
}
```

- `status`: The delivery status of the event, can be one of:
  - `pending`: pending for delivery
  - `retrying`: failed to deliver, will be retried later on.
  - `failed`: permanently failed.
  - `success`: delivered successfully.

### POST /admin/events/{seq}/retry

The given event must be either `retrying` or `failed`.
