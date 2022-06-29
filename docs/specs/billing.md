# Billing

This document describes

- How Authgear collects usage.
- How Authgear stores usage.
- How Authgear imposes rate limit.
- How Authgear reports usage to Stripe.
- How Authgear allows the developer to interact with Stripe.

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

type RateLimiter interface {
    TakeToken(bucket Bucket) error
}

// There are several ways to create Buckets.

// The first way is define a static Bucket.
var StaticBucket = Bucket{
    Key: "static"
}

// The second way is define a function that returns a bucket.
// This is the common way when we have the necessary argument in hand.

function IPBucket(ip string) Bucket {
    return Bucket{
        Key: ip,
    }
}

// The third way is define a struct that has a method to return a bucket.
// This is preferrable when the bucket requires so many data to construct,
// so dependency injection is required

type ComplexBucketFactory struct {
    InjectableA InjectableA
    INjectableB InjectableB
}
func (f *ComplexBucketFactory) MakeBucket() Bucket

// There is NO BucketMaker interface because
// sometimes it is necessary to inject more than 1 factory.
// Since Wire injects by type, having a single type is confusing.
// It is suggested that different parts of the code define their own BucketMaker interfaces.
// type BucketMaker interface {
//     MakeBucket() Bucket
// }
```

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

    UNIQUE (app_id, name, period, start_time)
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
only valid OTP code is counted as 1.

For example

```csv
,myapp,whatsapp-otp-count,daily,2022-06-01,101
```

## Usage alert

Usage alert periodically compares the usage record against the limits of the app.
If the usage exceeds the limits, an alert is sent.
The column `alert_data` is reserved for usage alert to store its state.
So it does not send alert every time it runs.

## The Stripe Integration

Authgear integrates with Stripe with [Checkout](https://stripe.com/docs/payments/checkout), [Customer Portal](https://stripe.com/docs/billing/subscriptions/integrating-customer-portal) and [webhooks](https://stripe.com/docs/webhooks).

### The subscription tables

```SQL
CREATE TABLE _portal_subscription (
    id                          text PRIMARY KEY,
    app_id                      text NOT NULL,
    stripe_customer_id          text NOT NULL,
    stripe_subscription_id      text NOT NULL,
    created_at                  timestamp WITHOUT TIME ZONE NOT NULL,
    updated_at                  timestamp WITHOUT TIME ZONE NOT NULL,
    UNIQUE (app_id)
);

CREATE TABLE _portal_subscription_checkout (
    id                          text PRIMARY KEY,
    app_id                      text NOT NULL,
    stripe_checkout_session_id  text NOT NULL,
    stripe_customer_id          text,
    status                      text NOT NULL,
    created_at                  timestamp WITHOUT TIME ZONE NOT NULL,
    updated_at                  timestamp WITHOUT TIME ZONE NOT NULL,
    expire_at                   timestamp WITHOUT TIME ZONE NOT NULL,
    UNIQUE (stripe_checkout_session_id),
    UNIQUE (stripe_customer_id)
);
```

### Pricing model

Create the necessary Products and Prices in Dashboard.
Authgear recognizes the Products and Prices with pre-defined metadata.

We use the following specifically configured Stripe Prices to represent our pricing model.

#### Fixed Price

Fixed Price is a Stripe Price with `recurring.usage_type=licensed`.
It is used for billing the base price of a subscription plan.

#### Usage Price

Usage Price is a Stripe Price with `recurring.usage_type=metered` and `recurring.aggregate_usage=sum`.
The quantity is calculated by summing all usage records within the billing period.
It is used for billing SMS cost.

#### MAU Price

MAU Price is a Stripe Price with `recurring.usage_type=metered`, `recurring.aggregate_usage=max`,
`transform_quantity.divide_by=5000`, and `transform_quantity.round=down`.
The quantity is the maximum value seen in the billing period.
It is used for billing MAU cost.

### Configure the Customer Portal

- ONLY turn on the following Functionality
  - Allow customers to view their Invoice history
  - Allow customers to update the following billing information: Email address and Billing address
  - Allow customer to update payment methods
- Leave anything else turned off

Reference: https://stripe.com/docs/billing/subscriptions/integrating-customer-portal?platform=billing#configure

### Configure Products and Prices

The metadata is for recognizing various Stripe Objects in Authgear.

- Create a Product for each pricing item we have.
  - A Product represents a single billable item, for example, Developer plan base price, SMS price for North America.
  - For products representing base price, attach the metadata `price_type=fixed,plan_name=PLAN_NAME`
  - For products representing usage price, attach the metadata `price_type=usage,usage_type=sms,sms_region=north-america` or `price_type=usage,usage_type=sms,sms_region=other-regions`
  - Only the default price is used by Authgear at the moment.

Reference: https://stripe.com/docs/products-prices/manage-prices

### Configure Webhooks

- Go to Stripe dashboard *Webhooks* section
- Add endpoint `https://PORTAL_DOMAIN/api/subscription/webhook/stripe`
- Select events: `checkout.session.completed`, `customer.subscription.created` and `customer.subscription.updated`

### Create Stripe Subscription

When the developer clicks to subscribe one of the plan, the portal does the following:

- Create Checkout Session with `mode=setup` to let the developer to create a Stripe Customer. Insert into `_portal_subscription_checkout` with `status=open`.
- Redirect the developer to the Checkout Session URL.
- The developer completes the Checkout Session.
- Listen `checkout.session.completed` and create a Stripe Subscription. Update `_portal_subscription_checkout` with `status=completed`.
- Listen `customer.subscription.created` and insert into `_portal_subscription`. Update `_portal_subscription_checkout` with `status=subscribed`.

### Switch plan

When the developer switches plan, the portal does the following:

- Update the Stripe Subscription Item's underlying Stripe Prices.
- Update the plan of the app.

Fixed Standard Price is subject to proration.
However, metered usage is billed using the updated price.
Therefore, if the prices are different, the developer could pay more or less.

See https://stripe.com/docs/billing/subscriptions/upgrade-downgrade

### Report [Usage Price](#usage-price) to Stripe

When the reporting job reports the usage for a specific app, it does the following.

- Set `now` as `time.Now().UTC()`
- Set `midnight` to `now` adjusted to the midnight of the day.
- Get the `_portal_subscription`
- Fetch the Stripe Subscription
- If `midnight` is NOT within [current\_period\_start](https://stripe.com/docs/api/subscriptions/object#subscription_object-current_period_start) and [current\_period\_end](https://stripe.com/docs/api/subscriptions/object#subscription_object-current_period_end), exit.
- For each kind of usage we keep track of, do the following
  - Identify the Stripe Subscription Item that contains the target Stripe Price for this usage. This is done via `metadata`. If the Stripe Subscription Item cannot be found, log an error telling the Stripe Subscription of which app is missing a Stripe Price for usage reporting, and then exit.
  - Fetch the daily usage records from [the usage record table](#the-usage-record-table) where `stripe_timestamp` is NULL and `end_time` is less than `now`. Those records are finalized and ready for reporting.
  - Set `quantity` to the sum of the count of the usage records.
  - Create a single Stripe Usage Record with `quantity=${quantity}`, `action=set` and `timestamp=${midnight}`.
  - Update the records to set `stripe_timestamp` to `midnight`.
