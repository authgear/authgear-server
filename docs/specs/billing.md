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
- `name` is a usage-specific name, e.g. `sms-sent.north-america`, which could stand for the number of sent SMS messages to North American phone numbers.
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
,myapp,sms-sent.north-america,daily,2022-06-01,54
,myapp,sms-sent.other-regions,daily,2022-06-01,68
,myapp,sms-total.other-regions,daily,2022-06-01,122
```

### Email count aggregator

The email count aggregator reads from the event data written by [the non-blocking event sink](#the-non-blocking-event-sink), and writes to [the usage record table](#the-usage-record-table).

For example

```csv
,myapp,email-sent,daily,2022-06-01,101
```

### Whatsapp OTP aggregator

The Whatsapp OTP aggregator reads from the event data written by [the non-blocking event sink](#the-non-blocking-event-sink), and writes to [the usage record table](#the-usage-record-table).

For example

```csv
,myapp,whatsapp-sent,2022-06-01,101
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
    cancelled_at                timestamp WITHOUT TIME ZONE,
    end_at                      timestamp WITHOUT TIME ZONE,
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

A Stripe Product represents a single billable item.
For example, the Business Plan has 4 billable items,
namely the base cost, 2 SMS costs and the MAU costs.

Each Stripe Product can have one or more Stripe Prices.

The following metadata are known to Authgear.

- `price_type`: Valid values are `fixed` and `usage`. Only appear in Product.
- `plan_name`: Valid values are `developers`, `startups` and `business`. Only appear in Product.
- `usage_type`: Valid values are `sms` and `mau`. Only appear in Product.
- `sms_region`: Valid values are `north-america` and `other-regions`.Only appear in Price.
- `free_quantity`: valid values are non-negative integers. Only appear in Price.
- `subscription_item_type`: Only appear in Product. Valid values are:
  - `plan`: The base cost of plan
  - `sms-north-america`: The sms cost in north america region
  - `sms-other-region`: The sms cost in other region
  - `whatsapp-north-america`: The whatsapp cost in north america region
  - `whatsapp-other-region`: The whatsapp cost in other region
  - `mau`: The mau cost

Each Stripe Product MUST have `price_type` and `subscription_item_type`.
Products with the same `subscription_item_type` are interchangable when upgrading or downgrading plan.
All plans must have exactly one product with each `subscription_item_type`.
If a Stripe Product has `plan_name`, then it is applicable ONLY to that plan.
Otherwise a Stripe Product is applicable to every plan.
Only the default Price of a Stripe Product is used for creating new subscriptions.

When the Price of a Product needs adjustment,
create a new Price and set it as the default.
Future subscription will use that new Price.
Existing subscriptions still reference the old Prices.

#### Fixed Price

Fixed Price is a Stripe Price with `recurring.usage_type=licensed`.
It is used for billing the base price of a subscription plan.

See [Configure Products and Prices](#configure-products-and-prices) for details.

#### Usage Price

Usage Price is a Stripe Price with `recurring.usage_type=metered` and `recurring.aggregate_usage=sum`.
The quantity is calculated by summing all usage records within the billing period.
It is used for billing SMS cost.

See [Configure Products and Prices](#configure-products-and-prices) for details.

#### MAU Price

MAU Price is a Stripe Price with `recurring.usage_type=metered`, `recurring.aggregate_usage=last_during_period`,
`transform_quantity.divide_by=5000`, and `transform_quantity.round=down`.
The quantity is the last value being uploaded to Stripe within the billing period.
It is used for billing MAU cost.

If `free_quantity` is present, then `free_quantity` is subtracted from the actual MAU count.
If the result is positive, the result is uploaded as quantity.

MAU price must present for each plans even additional maus are not applicable for the plan. Set the price to 0 for this case.

See [Configure Products and Prices](#configure-products-and-prices) for details.

#### Clear usage rule

- Fixed Price DOES NOT clear usage because it has no usage.
- Usage Price DOES NOT clear usage because if we cleared the usage, the developer is charged less.
- MAU Price clears usage because if we did not clear the usage, the developer is charged more when they downgrade from Business plan.

### Configure the Customer Portal

- ONLY turn on the following Functionality
  - Allow customers to view their Invoice history
  - Allow customers to update the following billing information: Email address and Billing address
  - Allow customer to update payment methods
- Leave anything else turned off

Reference: https://stripe.com/docs/billing/subscriptions/integrating-customer-portal?platform=billing#configure

### Configure Products and Prices

The base cost of a plan

```
Product
metadata.subscription_item_type=plan
metadata.price_type=fixed
metadata.plan_name=PLAN_NAME

Price
recurring.usage_type=licensed
```

The SMS cost for North America

```
Product
metadata.subscription_item_type=sms-north-america
metadata.price_type=usage
metadata.usage_type=sms
metadata.sms_region=north-america

Price
recurring.usage_type=metered
recurring.aggregate_usage=sum
```

The SMS cost for other regions

```
Product
metadata.subscription_item_type=sms-other-region
metadata.price_type=usage
metadata.usage_type=sms
metadata.sms_region=other-regions

Price
recurring.usage_type=metered
recurring.aggregate_usage=sum
```

The Whatsapp cost for North America

```
Product
metadata.subscription_item_type=whatsapp-north-america
metadata.price_type=usage
metadata.usage_type=whatsapp
metadata.sms_region=north-america

Price
recurring.usage_type=metered
recurring.aggregate_usage=sum
```

The Whatsapp cost for other regions

```
Product
metadata.subscription_item_type=whatsapp-other-region
metadata.price_type=usage
metadata.usage_type=whatsapp
metadata.sms_region=other-regions

Price
recurring.usage_type=metered
recurring.aggregate_usage=sum
```

The MAU cost of Business Plan

```
Product
metadata.subscription_item_type=mau
metadata.price_type=usage
metadata.usage_type=mau
metadata.plan_name=business

Price
recurring.usage_type=metered
recurring.aggregate_usage=last_during_period
transform_quantity.divide_by=5000
transform_quantity.round=up
metadata.free_quantity=10000
```

Reference: https://stripe.com/docs/products-prices/manage-prices

### Configure Webhooks

- Go to Stripe dashboard _Webhooks_ section
- Add endpoint `https://PORTAL_DOMAIN/api/subscription/webhook/stripe`
- Select events:
  - `checkout.session.completed`
  - `customer.subscription.created`
  - `customer.subscription.updated`
  - `customer.subscription.deleted`

### Create subscription

When the developer clicks to subscribe one of the plan, the portal does the following:

- Create Checkout Session with `mode=setup` to let the developer to create a Stripe Customer. Insert into `_portal_subscription_checkout` with `status=open`.
- Redirect the developer to the Checkout Session URL.
- The developer completes the Checkout Session.
- Listen `checkout.session.completed` and create a Stripe Subscription. Update `_portal_subscription_checkout` with `status=completed`.
- Listen `customer.subscription.created` and insert into `_portal_subscription`. Update `_portal_subscription_checkout` with `status=subscribed`.

### Update subscription

When the developer switches plan, the following steps are taken:

- Let SetA be (the set of old Prices - the set of new Prices)
- Mark the SubscriptionItem whose price is in SetA as [deleted](https://stripe.com/docs/api/subscriptions/update#update_subscription-items-deleted). Set [clear_usage](https://stripe.com/docs/api/subscriptions/update#update_subscription-items-clear_usage) according to [Clear usage rule](#clear-usage-rule)
- Let SetB be (the set of new Prices - the set of old Prices)
- Add the price in SetB to the subscription.

> Fixed Price is subject to proration.
> Usage Price is billed using the updated price. Therefore, if the prices are different, the developer could pay more or less.

See https://stripe.com/docs/billing/subscriptions/upgrade-downgrade

### Cancel subscription

When the developer cancels subscription, the following steps are taken:

- Update the Stripe subscription to set `cancel_at_period_end` to true.
- Update `_portal_subscription` to set `cancelled_at` to now and `end_at` to the `current_period_end` of the Stripe subscription.
- Listen for `customer.subscription.deleted`. Downgrade the app to free plan.

### Report [Usage Price](#usage-price) to Stripe

The cronjob takes the following steps:

- Set `NOW` as `time.Now().UTC()`
- Set `MIDNIGHT` to `NOW` adjusted to the midnight of the day.
- Get the `_portal_subscription`
- Fetch the Stripe Subscription
- Set `SUBSCRIPTION_CREATED_AT` be the creation time of the Stripe subscription.
- If `MIDNIGHT` is NOT within [current_period_start](https://stripe.com/docs/api/subscriptions/object#subscription_object-current_period_start) and [current_period_end](https://stripe.com/docs/api/subscriptions/object#subscription_object-current_period_end), exit.
- For each kind of usage we keep track of, do the following
  - Identify the Stripe Subscription Item that contains the target Stripe Price for this usage with `metadata`. If not found, exit.
  - Fetch the daily usage records from [the usage record table](#the-usage-record-table) with this condition.
    ```sql
    -- We do not report usage prior to subscription creation.
    -- This ensures we do not charge the developer more.
    start_time > SUBSCRIPTION_CREATED_AT
    AND
    -- The 1st condition is to retrieve usage records that have not been uploaded.
    -- The 2nd condition is to retrieve usage records that have been uploaded on the same day. If the job ever re-runs, the quantity is still correct.
    ((stripe_timestamp IS NULL AND end_time <= MIDNIGHT) OR (stripe_timestamp IS NOT NULL and stripe_timestamp = MIDNIGHT))
    ```
  - Set `QUANTITY` to the sum of the count of the usage records.
  - Create a single Stripe Usage Record with `quantity=${QUANTITY}`, `action=set` and `timestamp=${MIDNIGHT}`.
  - Update the records to set `stripe_timestamp` to `MIDNIGHT`.

### Report [MAU Price](#mau-price) to Stripe

The cronjob takes the following steps:

- Set `NOW` as `time.Now().UTC()`
- Get the `_portal_subscription`
- Fetch the Stripe Subscription
- Set `CURRENT_PERIOD_START` be the `current_period_start` of the Stripe subscription
- Set `CURRENT_PERIOD_END` be the `current_period_end` of the Stripe subscription
- Identify the Stripe Subscription Item that contains the target Stripe Price with `metadata`. If not found, exit.
- Fetch the monthly usage records from [the usage record table](#the-usage-record-table) with this condition.
  ```sql
  end_time = CURRENT_PERIOD_END
  ```
- SET `QUANTITY` to the sum of the count of the usage records.
- Create a single Stripe Usage Record with `quantity=${QUANTITY}`, `action=set` and `timestamp=${CURRENT_PERIOD_START}`.
- Update the records to set `stripe_timestamp` to `CURRENT_PERIOD_START`.
