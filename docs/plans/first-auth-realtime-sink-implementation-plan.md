# Implementation Plan — Real-time `application.first_auth` event sink

Companion to `docs/plans/first-auth-realtime-sink-design.md`. Replaces the batch
`analytic posthog first-auth` subcommand with an in-process event sink that emits
`application.first_auth` at the moment a client first authenticates.

## 1. Goal / scope

- Add a new `event.Sink` (`analytic.FirstAuthSink`) that, on `user.authenticated`
  and `m2m.token.created`, forwards one `application.first_auth` event to PostHog
  per (app_id, client_id), deduped via analytic Redis and made idempotent via a
  deterministic UUID.
- Wire PostHog credentials into every server binary that dispatches events (auth,
  admin, resolver, redisqueue) via `EnvironmentConfig`.
- Delete the batch subcommand and its audit-DB query.
- Preserve the PostHog event shape so Insight B in
  `docs/analytics/create-application-wizard-metrics.md` is unchanged.

## 2. Config model and schema

Add PostHog credentials to the shared server env config so `*config.AnalyticConfig`
is providable in the common DI graph.

- `pkg/lib/config/environment.go` — add field to `EnvironmentConfig`:
  ```go
  // Analytic configures analytics forwarding (PostHog).
  Analytic AnalyticConfig `envconfig:"ANALYTIC"`
  ```
  (`AnalyticConfig` already exists in `pkg/lib/config/analytic.go` with
  `POSTHOG_ENDPOINT` / `POSTHOG_APIKEY`; envconfig prefix becomes
  `ANALYTIC_POSTHOG_ENDPOINT` / `ANALYTIC_POSTHOG_APIKEY`.)

- No `authgear.yaml` / secret-config schema change. This is server-level env only,
  so `make export-schemas` is not required.

- DI provider — `pkg/lib/deps/deps_config.go`:
  ```go
  func ProvideAnalyticConfig(cfg *config.EnvironmentConfig) *config.AnalyticConfig {
      return &cfg.Analytic
  }
  ```
  Add `ProvideAnalyticConfig` to the `ConfigDeps` wire set in that file.
  `*config.EnvironmentConfig` is already exposed (deps_provider.go:66-67).

## 3. Credentials provider (shared)

Extract the portal's private `ProvidePosthogCredential` into a shared, exported
constructor in `pkg/lib/analytic` so both portal and the sink graph use one impl.

- `pkg/lib/analytic/posthog.go` — add:
  ```go
  func NewPosthogCredentials(c *config.AnalyticConfig) *PosthogCredentials {
      if c.PosthogEndpoint != "" && c.PosthogAPIKey != "" {
          return &PosthogCredentials{Endpoint: c.PosthogEndpoint, APIKey: c.PosthogAPIKey}
      }
      return nil
  }
  ```
- `pkg/portal/deps.go` — replace the body of `ProvidePosthogCredential` to call
  `analytic.NewPosthogCredentials(analyticConfig)` (keep the portal wrapper name to
  avoid touching portal wiring), OR swap portal's wire set to use
  `analytic.NewPosthogCredentials` directly. Prefer the former (smaller diff).

Returning `nil` when unset is the graceful-degradation contract: `PosthogService.Batch`
already logs-and-skips when credentials are nil (`posthog.go:412-419`).

## 4. Runtime plan — the sink

New file `pkg/lib/analytic/first_auth_sink.go` (same package, reuses
`firstAuthUUIDNamespace` and the event builder; verified no import cycle —
`pkg/lib/analytic` does not import `pkg/lib/event`, and the `event.Sink` interface
is defined over `pkg/api/event`).

```go
package analytic

const firstAuthDedupTTL = 90 * 24 * time.Hour // matches audit retention; bounds re-sends

var FirstAuthSinkLogger = slogutil.NewLogger("posthog-first-auth-sink")

type FirstAuthSink struct {
    Clock         clock.Clock
    AnalyticRedis *analyticredis.Handle // may be nil
    Posthog       *PosthogService       // holds credentials + HTTP client
}

func (s *FirstAuthSink) ReceiveBlockingEvent(ctx context.Context, e *event.Event) error {
    return nil
}

func (s *FirstAuthSink) ReceiveNonBlockingEvent(ctx context.Context, e *event.Event) error {
    logger := FirstAuthSinkLogger.GetLogger(ctx)

    // 1. Only auth events.
    if e.Type != nonblocking.UserAuthenticated && e.Type != nonblocking.M2MTokenCreated {
        return nil
    }
    // 2. Need identifiers.
    appID := e.Context.AppID
    clientID := e.Context.ClientID
    if appID == "" || clientID == "" {
        return nil
    }
    // 3. No credentials or no analytic redis -> no-op (graceful).
    if s.Posthog.PosthogCredentials == nil || s.AnalyticRedis == nil {
        return nil
    }
    // 4. Dedup: only the first auth per client wins.
    firstAuthAt := s.Clock.NowUTC()
    won, err := s.markFirstAuth(ctx, appID, clientID, firstAuthAt)
    if err != nil {
        logger.WithError(err).Error(ctx, "failed to mark first auth")
        return nil // never fail the auth
    }
    if !won {
        return nil
    }
    // 5. Fire-and-forget delivery; detach from request cancellation.
    detachedCtx := context.WithoutCancel(ctx)
    go func() {
        defer func() {
            if r := recover(); r != nil {
                logger.Error(detachedCtx, "panic forwarding first_auth", slog.Any("recovered", r))
            }
        }()
        evt, err := buildFirstAuthEvent(appID, clientID, firstAuthAt)
        if err != nil {
            logger.WithError(err).Error(detachedCtx, "failed to build first_auth event")
            return
        }
        if err := s.Posthog.Batch(detachedCtx, []json.RawMessage{evt}); err != nil {
            logger.WithError(err).Error(detachedCtx, "failed to forward first_auth to posthog")
        }
    }()
    return nil
}

func (s *FirstAuthSink) markFirstAuth(ctx context.Context, appID, clientID string, at time.Time) (bool, error) {
    key := firstAuthDedupKey(appID, clientID)
    var keyWasSet bool
    err := s.AnalyticRedis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
        var err error
        keyWasSet, err = conn.SetNX(ctx, key, at.UTC().Format(time.RFC3339), firstAuthDedupTTL).Result()
        return err
    })
    return keyWasSet, err
}

func firstAuthDedupKey(appID, clientID string) string {
    return fmt.Sprintf("app:%s:posthog-first-auth:%s", appID, clientID)
}
```

- `var _ event.Sink = &FirstAuthSink{}` — assert interface (import `pkg/lib/event`
  ONLY for this assertion would create a cycle; instead assert structurally by the
  method set matching, and rely on wire binding in `event.NewService`). Do **not**
  add the `event.Sink` compile assertion in the `analytic` package. Correctness is
  enforced by `event.NewService`'s typed parameter.
- Redis idiom mirrors `pkg/lib/infra/whatsapp/message_store_redis.go:48-64` and the
  `WithConnContext` + `redis.Redis_6_0_Cmdable` pattern used across the codebase.
- Key format follows the meter convention `app:%s:...`
  (`pkg/lib/meter/write_store_redis.go:78-96`).
- Detached-goroutine delivery mirrors
  `pkg/lib/hook/hook_web.go:152-161` (`context.WithoutCancel` → `go func()`).

### Event builder (refactor from batch)

`pkg/lib/analytic/posthog.go` — extract per-event construction so the sink and any
remaining code share one builder:

```go
func buildFirstAuthEvent(appID, clientID string, firstAuthAt time.Time) (json.RawMessage, error) {
    eventUUID := uuid.NewSHA1(firstAuthUUIDNamespace, []byte(appID+":"+clientID)).String()
    event := map[string]any{
        "event":       "application.first_auth",
        "distinct_id": clientID,
        "uuid":        eventUUID,
        "timestamp":   firstAuthAt.UTC().Format(time.RFC3339),
        "properties": map[string]any{
            "client_id":               clientID,
            "app_id":                  appID,
            "$geoip_disable":          true,
            "$process_person_profile": false,
        },
    }
    return json.Marshal(event)
}
```

Keep `firstAuthUUIDNamespace` (posthog.go:38). Delete `makeFirstAuthEvents`,
`firstAuthActivityTypes`, `firstAuthLookback`, and `ForwardFirstAuthEvents`.

## 5. Event / delivery flow (call sequence)

1. Auth completes; a handler calls `Events.DispatchEventOnCommit(ctx, &nonblocking.UserAuthenticatedEventPayload{...})`
   (e.g. `pkg/lib/oauth/handler/handler_token.go:1433`) or, for M2M,
   `&nonblocking.M2MTokenCreatedEventPayload{...}` (`pkg/lib/oauth/handler/service_token.go:445`).
2. On tx commit, `event.Service.DidCommitTx` (`pkg/lib/event/service.go:179-195`)
   fans the event out to every sink, including `FirstAuthSink`.
3. `FirstAuthSink.ReceiveNonBlockingEvent` filters by type, reads
   `e.Context.AppID` / `e.Context.ClientID`, and calls `markFirstAuth` (one SETNX).
4. If SETNX won, it spawns a detached goroutine that builds the single event and
   calls `PosthogService.Batch`. The request returns without waiting.
5. On a lost SETNX (returning client) or empty credentials, it returns immediately.

Called once per auth event; the PostHog HTTP call happens at most once per client
(SETNX-guarded within the TTL window), and the deterministic UUID collapses any
duplicate that a Redis key loss could cause.

## 6. Wiring

- `pkg/lib/analytic/deps.go` — add a dedicated set:
  ```go
  var FirstAuthSinkDependencySet = wire.NewSet(
      NewPosthogHTTPClient,
      NewPosthogCredentials,
      wire.Struct(new(PosthogService), "*"),
      wire.Struct(new(FirstAuthSink), "*"),
  )
  ```
- `pkg/lib/deps/deps_common.go` — add `analytic.FirstAuthSinkDependencySet` and
  `ProvideAnalyticConfig` (via `ConfigDeps`) to the common set (near the existing
  `event.DependencySet` at line 169).
- `pkg/lib/event/deps.go` — add `firstAuthSink *analytic.FirstAuthSink` parameter to
  `NewService` and append it to the `Sinks` slice. Import `pkg/lib/analytic`.
- Regenerate wire: `make generate`. Affected generated files:
  `pkg/auth/wire_gen.go`, `pkg/admin/wire_gen.go`, `pkg/resolver/wire_gen.go`,
  `pkg/redisqueue/wire_gen.go` (each constructs `event.NewService`).

`config.AppID`, `*analyticredis.Handle`, and `clock.Clock` are already in the
app-scoped graph (`AppRootDeps`, deps_provider.go:83-94), so no new providers for
those.

## 7. Removals

- `cmd/portal/cmd/cmdanalytic/posthog.go` — delete `cmdAnalyticPosthogFirstAuth`
  (lines ~184-256).
- `cmd/portal/cmd/cmdanalytic/analytic.go` — delete the
  `cmdAnalyticPosthog.AddCommand(cmdAnalyticPosthogFirstAuth)` block and its
  `binder.BindString(...)` lines (lines ~65-72).
- `pkg/lib/analytic/posthog.go` — delete `ForwardFirstAuthEvents`,
  `makeFirstAuthEvents`, `firstAuthActivityTypes`, `firstAuthLookback`; keep
  `firstAuthUUIDNamespace` and add `buildFirstAuthEvent`.
- `pkg/lib/analytic/count_collector.go` — remove `GetFirstAuthTimeByClientID` from
  the `MeterAuditDBReadStore` interface (line 28).
- `pkg/lib/meter/auditdb_read_store.go` — delete the `GetFirstAuthTimeByClientID`
  method (lines ~40-71).
- Confirm no other references remain: `grep -rn "GetFirstAuthTimeByClientID\|ForwardFirstAuthEvents\|firstAuthLookback\|makeFirstAuthEvents\|firstAuthActivityTypes\|cmdAnalyticPosthogFirstAuth"`
  returns only deletions.

## 8. Spec doc update

`docs/analytics/create-application-wizard-metrics.md`:
- Events table: `application.first_auth` source → `auth server (real-time event sink)`.
- Remove the "Freshness lag" caveat and the "90-day audit retention + lookback"
  caveat.
- Rewrite the "Raw first_auth counts vs. min()" caveat: a duplicate is now only
  possible if the Redis dedup key is lost within the TTL window; the deterministic
  UUID lets PostHog dedupe; still aggregate with `min(timestamp)` / funnel
  first-match.

## 9. Test plan

Unit tests, Convey BDD style (matches `pkg/lib/analytic/posthog_test.go`, which uses
`github.com/smartystreets/goconvey/convey`). See the `add-go-test` skill.

- `pkg/lib/analytic/posthog_test.go` — replace `TestMakeFirstAuthEvents` with
  `TestBuildFirstAuthEvent` asserting the same shape (event name, `distinct_id`,
  `timestamp`, `client_id`, `app_id`, `$geoip_disable`, `$process_person_profile`,
  deterministic UUID stability).
- `pkg/lib/analytic/first_auth_sink_test.go` (new):
  - non-auth event type → no SETNX, no forward;
  - `user.authenticated` with empty `Context.ClientID` → no-op;
  - `PosthogCredentials == nil` → no-op;
  - `AnalyticRedis == nil` → no-op;
  - event type filter covers both `user.authenticated` and `m2m.token.created`.
  Use a fake/mocked `analyticredis` connection to assert SETNX is called with the
  expected key and that a lost SETNX suppresses forwarding. Since delivery is a
  detached goroutine, unit-test `markFirstAuth` + `buildFirstAuthEvent`
  deterministically rather than asserting the async HTTP POST.
- Run: `go test ./pkg/lib/analytic/... ./pkg/lib/event/... ./pkg/lib/meter/...`
  and `make generate` (wire) must succeed with no diff afterward.

No e2e test: the sink's effect is an external PostHog POST with no assertable
in-product surface. Covered by unit tests. (If desired later, an e2e that asserts
the SETNX key is set after an auth could be added, but it is out of scope here.)

## 10. Compatibility and deployment

- PostHog event name, properties, timestamp, and deterministic UUID are unchanged →
  Insight B and its HogQL are unaffected.
- Deploy requirement: set `ANALYTIC_POSTHOG_ENDPOINT` / `ANALYTIC_POSTHOG_APIKEY`
  on the **auth server** (previously only the portal batch job had them). When
  unset, the sink is a no-op.
- No persisted-state migration. The new Redis keys are additive and self-expiring.
- Removing the batch subcommand means any cron currently invoking
  `analytic posthog first-auth` must be decommissioned (ops note for the PR).

## 11. Fixed behavioral decisions

- First-auth timestamp = `Clock.NowUTC()` at event receipt (fires immediately after
  auth commit; accurate to milliseconds). No dependence on audit `created_at`.
- Dedup TTL = 90 days. A client re-authenticating after 90 days may re-emit; the
  deterministic UUID keeps PostHog aggregation correct.
- Delivery is best-effort fire-and-forget; failures are logged, never surfaced to
  the authenticating user.
- Sink lives in package `pkg/lib/analytic` (no interface assertion there to avoid an
  import cycle; the `event.NewService` typed parameter enforces conformance).

## 12. Atomic commit plan

1. **Config + shared credentials provider**
   - `pkg/lib/config/environment.go` (add `Analytic` field),
     `pkg/lib/analytic/posthog.go` (add `NewPosthogCredentials`),
     `pkg/portal/deps.go` (reuse it),
     `pkg/lib/deps/deps_config.go` (add `ProvideAnalyticConfig` to `ConfigDeps`).
   - No behavior yet; compiles.
2. **Event builder refactor + sink implementation**
   - `pkg/lib/analytic/posthog.go` (extract `buildFirstAuthEvent`; delete
     `makeFirstAuthEvents`),
     `pkg/lib/analytic/first_auth_sink.go` (new sink),
     `pkg/lib/analytic/deps.go` (add `FirstAuthSinkDependencySet`).
   - Unit tests for builder + sink in the same commit.
3. **Wire the sink into the event service**
   - `pkg/lib/event/deps.go` (new `NewService` param + Sinks entry),
     `pkg/lib/deps/deps_common.go` (include the dependency set),
     regenerate `pkg/{auth,admin,resolver,redisqueue}/wire_gen.go` in the **same
     commit** (`make generate`).
4. **Remove batch machinery**
   - `cmd/portal/cmd/cmdanalytic/posthog.go`,
     `cmd/portal/cmd/cmdanalytic/analytic.go`,
     `pkg/lib/analytic/posthog.go` (delete `ForwardFirstAuthEvents`,
     `firstAuthActivityTypes`, `firstAuthLookback`),
     `pkg/lib/analytic/count_collector.go`,
     `pkg/lib/meter/auditdb_read_store.go`.
5. **Docs**
   - `docs/analytics/create-application-wizard-metrics.md` caveat/source updates.

Each commit compiles and passes tests independently; commit 3 keeps generated wiring
in lockstep with the `NewService` signature change so the tree is bisect-safe.
