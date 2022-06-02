# Billing

This document describes

- How Authgear collects usage.
- How Authgear stores usage.
- How Authgear imposes rate limit.
- How Authgear reports usage to Stripe.

## The Meter

The Meter is capable of calculating distinct-count.
It uses [HyperLogLog](https://redis.com/redis-best-practices/counting/hyperloglog/) of Redis.

It keeps track of counts for different periods.

- The monthly count includes `YYYY-MM` as part of the key, e.g. `2022-06`.
- The weekly count includes `YYYY-Www` as part of the key, e.g. `2022-W20`.
- The daily count includes `YYYY-MM-DD` as part of the key, e.g. `2022-06-01`.

## The rate limiter

The rate limiter uses simplified version of [Token Bucket](https://en.wikipedia.org/wiki/Token_bucket).
It reads app-specific bucket configuration and stores the bucket in Redis.

### How does the rate limiter work

```golang
type Bucket struct {
    Key         string
    Size        int
    ResetPeriod time.Duration
}
```

A bucket is identified by a key.
It has a size and a reset period.

```golang
type RateLimiter interface {
    TakeToken(key string) error
}
```

The rate limiter reads app-specific bucket configuration.
If the bucket does not exist in Redis, a full bucket is created first.
The bucket has an expiration of the reset period.
When the bucket expires, it disappear,
so next time when a token is about to be taken, the token is taken from a full bucket.

## The non-blocking event sink

The non-blocking event sink writes non-blocking events into PostgreSQL database.
The event data is for audit log purpose.

## The usage record table

```sql
CREATE _portal_usage_record (
    id string PRIMARY KEY,
    app_id string NOT NULL,
    name string NOT NULL,
    period string NOT NULL,
    start_time timestamp without time zone NOT NULL,
    end_time timestamp without time zone NOT NULL,
    count integer NOT NULL,
    alert_data jsonb,
    stripe_timestamp timestamp without time zone

    UNIQUE (app_id, tag, period, date)
);
```

- `period` is one of `monthly`, `weekly` or `daily`.
- `name` is a usage-specific name, e.g. `sms-count.north-america`, which could stand for the number of sent SMS messages to North American phone numbers.
- `start_time` is the timestamp at the beginning of the period.
  - For `monthly` usage record, it is the midnight of the first day of the month in UTC.
  - For `weekly` usage record, it is the midnight of the Monday of the week in UTC.
  - For `daily` usage record, it is the midnight of the day in UTC.
- `end_time` is the timestamp at the end of the period.
  - For `monthly` usage record, it is the midnight of the first day of the NEXT month in UTC.
  - For `weekly` usage record, it is the midnight of the Monday of the NEXT week in UTC.
  - For `daily` usage record, it is the midnight of the next day in UTC.
- `alert_data` is for [usage alert](#usage-alert).

## The cron-based aggregator

The cron-based aggregator is usage specific.
For each kind of usage, there is an aggregator.
The aggregator reads from its source and write to [the usage record table](#the-usage-record-table) in PostgreSQL database.

### Active user aggregator

The active user aggregator reads data from [the Meter](#the-meter) and writes to [the usage record table](#the-usage-record-table).

For example

```csv
,myapp,active-user,monthly,2022-06-01,20
,myapp,active-user,weekly,2022-06-06,10
,myapp,active-user,daily,2022-06-01,5
,myapp,active-user,daily,2022-06-02,4,
```

### SMS count aggregator

The SMS count aggregator reads from the event data written by [the non-blocking event sink](#the-non-blocking-event-sink), aggregate into counts by regions, and writes to [the usage record table](#the-usage-record-table).

For example

```csv
,myapp,sms-count.north-america,daily,2022-06-01,54
,myapp,sms-count.other-regions,daily,2022-06-01,68
```

### Email count aggregator

The email count aggregator reads from the event data written by [the non-blocking event sink](#the-non-blocking-event-sink), and writes to [the usage record table](#the-usage-record-table).

For example

```csv
,myapp,email-count,daily,2022-06-01,101
```

### Whatsapp OTP aggregator

The Whatsapp OTP aggregator reads from the event data written by [the non-blocking event sink](#the-non-blocking-event-sink), and writes to [the usage record table](#the-usage-record-table).

Since junk messages can be sent to our Whatsapp number,
only valid TOP code is counted as 1.

For example

```csv
,myapp,whatsapp-otp-count,daily,2022-06-01,101
```

## Usage alert

Usage alert periodically compares the usage record against the limits of the app.
If the usage exceeds the limits, an alert is sent.
The column `alert_data` is reserved for usage alert to store its state.
So it does not send alert every time it runs.

## The Stripe integration

The Stripe integration reads from [the usage record table](#the-usage-record-table) and calls Stripe API to reports usage to Stripe.

### The subscription table

```SQL
CREATE _portal_subscription (
    id string PRIMARY KEY,
    app_id string NOT NULL,
    stripe_customer_id string NOT NULL,
    stripe_subscription_id string NOT NULL
);
```

### Stable usage reporting to Stripe

When the reporting job reports the usage for a specific app, it does the following.

- Set `now` as `time.Now().UTC()`
- Set `midnight` to `now` adjusted to the midnight of the day.
- Get the `_portal_subscription`
- Fetch the Stripe Subscription
- If `midnight` is NOT within [current\_period\_start](https://stripe.com/docs/api/subscriptions/object#subscription_object-current_period_start) and [current\_period\_end](https://stripe.com/docs/api/subscriptions/object#subscription_object-current_period_end), exit.
- Fetch the daily usage records where `stripe_timestamp` is NULL and `end_time` is less than `now`. Those records are finalized and ready for reporting.
- Report the records to Stripe using [set](https://stripe.com/docs/api/usage_records/create#usage_record_create-action) with timestamp equal to `midnight`.
- Update the records to set `stripe_timestamp` to `midnight`.

> Figure out how to identify the correct subscription item to report usage to
> It is possible that Sales want to have different pricing for different apps.
> For example, App A may get a pricing of first 100 SMS for free, and USD N per message thereafter.
> While App B has a different pricing of USD N per message at the time.
> In Stripe this situation is modeled as different Stripe Price objects.
