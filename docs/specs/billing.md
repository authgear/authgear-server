# Billing

This document describes how Authgear keeps track of usage information for billing purpose.

## Database schema

TODO

## Overview of the Meter system

### Diagram of the Meter system
```
                                          ┌────────────────┐
                                          │                │
                                          │    Authgear    │
                                          │                │
 ┌────────────┐                           │  ┌──────────┐  │                              ┌───────┐
 │            │       Read for portal     │  │          │  │      Track on demand         │       │
 │ PostgreSQL │◄──────────────────────────┼──┤   HTTP   ├──┼─────────────────────────────►│ Redis │
 │            │                           │  │          │  │     Read for ratelimit       │       │
 └────────────┘                           │  └──────────┘  │                              └───────┘
       ▲                                  │                │                                  ▲
       │                                  │  ┌──────────┐  │                                  │
       │                                  │  │          │  │                                  │
       └──────────────────────────────────┼──┤ Cronjob  ├──┼──────────────────────────────────┘
           Read and persist to PostgreSQL │  │          │  │  Read and persist to PostgreSQL
                                          │  └──────────┘  │
                                          │                │
                                          │                │
                                          └────────────────┘
```

Authgear stores usage information in Redis real-time.
Usage information includes count and unique count.

### Key convention

- Monthly count includes `YYYY-MM` as part of the key, e.g. `2022-06`.
- Weekly count includes `YYYY-Www` as part of the key, e.g. `2022-W20`.
- Daily count includes `YYYY-MM-DD` as part of the key, e.g. `2022-06-01`.

### Unique count

Unique count counts the number of occurrence of unique values within a certain period.
The implementation uses [HyperLogLog](https://redis.com/redis-best-practices/counting/hyperloglog/).

#### Active User

Active user keeps track of unique active user of a given app within a certain period.
It is metered monthly, weekly and daily.
Only the monthly active user is used 

### Count

Count is an incremental counter within a certain period.

#### SMS count

> We want to bill SMS by region. https://github.com/authgear/authgear-server/issues/2043
> So the count here is only useful for imposing hard rate limit.
> For billing purpose, would it be more flexible if we store the raw log entry and derive the usage?
