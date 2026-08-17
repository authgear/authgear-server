# Real-time `application.first_auth` event sink

## Context

PR #5806 ("Add Create Application baseline metrics") forwards a per-client
`application.first_auth` event to PostHog. It powers Insight B in
`docs/analytics/create-application-wizard-metrics.md`: a funnel
`createApplication.created → application.first_auth` aggregated by `client_id`,
yielding activation rate and time-to-integration.

The original implementation derives the event **in batch**: an
`analytic posthog first-auth` subcommand scans the audit DB over a 35-day
lookback window and forwards one event per client. That requires a cronjob.

Review feedback (tung2744): *"Instead of adding extra command which require a
cronjob to run it, can we simply send an event at the moment the first auth
occurs?"*

This spec replaces the batch forwarder with a **real-time event sink** that
emits `application.first_auth` at the moment a client first authenticates.

## Goal

- Emit `application.first_auth` (properties: `client_id`, `app_id`) once, at the
  first successful auth per OAuth client, from the auth server runtime.
- Preserve the metric semantics consumed by Insight B: the event carries the
  real auth timestamp and a deterministic UUID so PostHog aggregation is
  unchanged.
- Remove the batch subcommand and its supporting query so no cronjob is needed.

Non-goals: retroactive backfill (baseline still accrues forward from ship), any
change to the portal-emitted `createApplication.*` events.

## Design

### Hook point — a new event Sink

Auth success is already dispatched as non-blocking events:
- `user.authenticated` (`pkg/api/event/nonblocking/user_authenticated.go`)
- `m2m.token.created` (`pkg/api/event/nonblocking/m2m_token_created.go`)

Both flow through `event.Service`, which fans every non-blocking event out to
its `Sink`s in `DidCommitTx` (after the DB transaction commits). Existing sinks:
`hook`, `audit`, `search/reindex`, `userinfo` (registered in
`pkg/lib/event/deps.go`).

Add a **new sink** (in `pkg/lib/analytic`) implementing the `event.Sink`
interface. Modeled on `pkg/lib/audit/sink.go`:
- `ReceiveBlockingEvent` — no-op.
- `ReceiveNonBlockingEvent` — act only on event types `user.authenticated` and
  `m2m.token.created`; ignore everything else.

`client_id` and `app_id` are read from `event.Context` (`context.go:44-45`),
which is already populated for both event types. The event's real timestamp is
`e.Context` / payload time.

### Dedup — SETNX on analytic Redis

To fire ~once per client (not on every auth), the sink does a `SetNX` on the
analytic Redis:

```
key:   posthog:first_auth:<app_id>:<client_id>
value: <first-auth timestamp>
TTL:   90 days  (matches audit retention; bounds re-sends)
```

Pattern: `pkg/lib/infra/whatsapp/message_store_redis.go` (`SetNX(...).Result()`
returns `keyWasSet`). Only forward to PostHog when `keyWasSet == true`. On a
lost race (key exists), the sink returns immediately after one Redis round-trip.

### Idempotency backstop — deterministic UUID

The event keeps the deterministic UUIDv5 of `app_id:client_id` (as the batch
version already computed). Redis is best-effort (evictable, flushable); the UUID
is the correctness backstop — if the key is lost and a later auth re-sends, the
identical UUID lets PostHog dedupe. Redis = efficiency; UUID = correctness.

### Delivery — fire-and-forget goroutine

Sinks run synchronously in `DidCommitTx`, in-request. An analytics side-effect
must never add latency to (or fail) a user's successful login. So delivery is
detached, mirroring non-blocking webhook delivery
(`pkg/lib/hook/hook_web.go` `DeliverNonBlockingEvent` → `PerformNoResponse`):

```go
ctx = context.WithoutCancel(ctx)   // survive request cancellation
go func() {
    // build single application.first_auth event, POST to PostHog
    // recover panics; log (never propagate) errors
}()
```

Uses an external HTTP client with a short timeout. Errors are logged, never
returned to the caller.

The single-event payload reuses the existing
`analytic.PosthogService.Batch` / event-building helper (extract the
per-event construction from `makeFirstAuthEvents` so batch-shape and sink-shape
share one builder).

### Wiring PostHog credentials into the auth server

The auth server's `EnvironmentConfig` (`pkg/lib/config/environment.go`) does
**not** currently carry PostHog credentials — only the portal server does
(`cmd/portal/server/config.go`). Add the `ANALYTIC` config
(`POSTHOG_ENDPOINT`, `POSTHOG_APIKEY`) to the auth server so the sink can read
them.

- When credentials are empty (default), the sink is a no-op — it must degrade
  gracefully, exactly like `PosthogService.Batch` already does when the endpoint
  is unset.

Register the new sink in `event.NewService` (`pkg/lib/event/deps.go`) and add it
to the wire `DependencySet`, then regenerate wire. The sink appears in every
graph that builds `event.Service`: `pkg/auth`, `pkg/admin`, `pkg/resolver`,
`pkg/redisqueue`.

**Operational note:** the PostHog env vars must be deployed to the auth server
for the event to fire. This is the deployment cost of moving emission from the
portal batch job to the auth runtime.

## Removals

- `analytic posthog first-auth` subcommand
  (`cmd/portal/cmd/cmdanalytic/posthog.go` lines ~184-256) and its registration
  in `cmd/portal/cmd/cmdanalytic/analytic.go`.
- `PosthogIntegration.ForwardFirstAuthEvents` and the batch-only bits in
  `pkg/lib/analytic/posthog.go` (`firstAuthLookback`; keep the shared
  single-event builder + UUID namespace + activity-type constants as needed by
  the sink).
- `MeterAuditDBReadStore.GetFirstAuthTimeByClientID` — interface entry in
  `count_collector.go` and impl in `pkg/lib/meter/auditdb_read_store.go`.
- First-auth cases in `pkg/lib/analytic/posthog_test.go`; add sink tests instead.

## Spec doc update

Update `docs/analytics/create-application-wizard-metrics.md`:
- `application.first_auth` source: `auth server (real-time event sink)` instead
  of the batch job.
- Remove the "Freshness lag" and "90-day audit retention + lookback" caveats
  (no longer a batch job / lookback window).
- Rewrite the "Raw first_auth counts vs. min()" caveat: duplicates are now only
  possible on Redis key loss, deduped by deterministic UUID; still aggregate
  with `min(timestamp)` / funnel first-match.

## Testing

- **Sink unit tests** (`pkg/lib/analytic`): given a `user.authenticated` /
  `m2m.token.created` event, first call forwards (SETNX wins) and builds the
  correct event (client_id, app_id, deterministic UUID, timestamp); second call
  for the same client is a no-op (SETNX loses); non-auth events ignored; empty
  PostHog credentials → no-op.
- Reuse the existing PostHog HTTP-shape assertions from `posthog_test.go` for
  the single-event builder.
- `go test ./pkg/lib/analytic/... ./pkg/lib/event/...` and `make generate`
  (wire) must pass.

## Compatibility

- PostHog event name, properties, timestamp, and UUID are unchanged, so Insight
  B and its HogQL work without modification.
- No config-schema change for tenant `authgear.yaml`; the new config is
  server-level env only.
