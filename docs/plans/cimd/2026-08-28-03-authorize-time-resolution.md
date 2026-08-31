# CIMD Part 3 — Authorize-Time Resolution, Persistence & Refetch

Spec: [docs/specs/cimd.md — Client Resolution](../../specs/cimd.md#client-resolution), [Where resolution happens](../../specs/cimd.md#where-resolution-happens), [Error Handling](../../specs/cimd.md#error-handling), [Redirect URI Validation](../../specs/cimd.md#redirect-uri-validation), [Mapping to the Unified Client Model](../../specs/cimd.md#mapping-to-the-unified-client-model).

Depends on [Part 1](2026-08-28-01-config-and-client-id.md) (config, `ParseCIMDClientID`, `IsCIMDClientIDAllowed`, the resolver candidate branch, `ResolveTokenLifetimes`) and [Part 2](2026-08-28-02-document-fetching.md) (`cimd.Fetcher`, `cimd.ParseAndValidate`, `cimd.RefetchInterval`).

## 1. Goal / Scope

This is the part that makes CIMD work. It adds:

- `oauthclient.Store.UpsertCIMDClient` / `Commands.UpsertCIMDClient` — the persisted, per-`client_id`, shared record spec § Where resolution happens describes, created on first resolution and overwritten in place on every refetch.
- `cimd.Service.EnsureClientResolved` — the orchestration: is this a CIMD candidate, is its record fresh, single-flight, fetch, validate, persist, invalidate cache.
- One call site: `AuthorizationHandler.ValidateRequestWithoutTx`, i.e. `/oauth2/authorize` and nothing else.
- The error mapping for an unresolvable CIMD `client_id`.

Out of scope: rate limits ([Part 4](2026-08-28-04-rate-limits.md) — the hook point is prepared here as an interface with a no-op default), the client limit ([Part 5](2026-08-28-05-usage-limit.md) — same), consent screen and settings UI ([Part 6](2026-08-28-06-consent-and-authorized-apps.md)), and audit logging ([Part 8](2026-08-31-08-audit-log.md), which emits from the numbered steps in §3.3).

### 1.1 The architectural rule this part exists to enforce

> A CIMD fetch — and therefore any outbound network call — happens **only** as a side effect of `/oauth2/authorize`. […] Every other step reads that persisted record — a plain lookup, never a live fetch.
> — spec § Where resolution happens

`oauthclient.Resolver.ResolveClient` is called from **35+ places** (`grep -rn "\.ResolveClient("`): middleware, view models, the token handler, the end-session handler, the userinfo handler, the admin facade, SAML, the tester. If resolution triggered a fetch, an unauthenticated caller could make Authgear issue an outbound HTTP request from any of them, and a page render could block for 5 seconds on a third party's server.

So the split is absolute:

| Layer | Responsibility | Network? |
|---|---|---|
| `cimd.Service.EnsureClientResolved` | fetch + validate + upsert. Called from exactly one place. | **yes** |
| `oauthclient.Resolver.ResolveClient` | read the persisted record (Redis → Postgres) and synthesize a `*config.OAuthClientConfig`. Unchanged from Part 1. | never |

`/oauth2/authorize` calls the first, then the second. Everything else calls only the second. A reviewer should treat any new `EnsureClientResolved` call site as requiring an explicit spec change.

## 2. Persistence — `pkg/lib/oauthclient`

### 2.1 `Store.UpsertCIMDClient`

`Store.CreateClient` cannot be reused: a refetch must overwrite the existing row in place (spec: "A fresh fetch overwrites that same record in place"), and the natural key is `(app_id, client_id)`, which already has a unique index (`_auth_oauth_client_client_id_unique`).

```go
// UpsertCIMDClientOptions carries the CIMD-specific write. Source and Kind
// are not parameters: spec § Mapping to the Unified Client Model fixes
// source: CIMD and kind: THIRD_PARTY for every CIMD client, and making them
// arguments would invite a caller to write a first-party or DCR row through
// this path.
type UpsertCIMDClientOptions struct {
	ClientID        string
	ApplicationType string
	ClientName      *string
	ClientURI       *string
	LogoURI         *string
	TOSURI          *string
	PolicyURI       *string
	RedirectURIs    []string
	GrantTypes      []string
	ResponseTypes   []string
}

// UpsertCIMDClient inserts or updates the single row for options.ClientID
// and reports whether a new row was created, so the caller can apply the
// client limit only to creations (spec § Client Limit: "A client_id that
// already has a persisted record is unaffected regardless of the limit").
//
// The WHERE clause on the DO UPDATE branch is a guard, not an optimisation:
// it makes it structurally impossible for this path to overwrite a DCR row.
// In practice a dcrc_-prefixed client_id can never equal a CIMD URL so the
// collision is unreachable, but the guard means a future change to either
// id shape cannot silently let a fetched document take over a registered
// client's row. When the guard blocks the update, ON CONFLICT ... DO UPDATE
// affects zero rows, RETURNING yields no row, and this function returns
// ErrDynamicClientSourceConflict.
func (s *Store) UpsertCIMDClient(ctx context.Context, options *UpsertCIMDClientOptions) (client *Client, created bool, err error) {
	now := s.Clock.NowUTC()
	// ... json.Marshal redirectURIs / grantTypes / responseTypes, and the
	// same empty-string-to-nil normalization for ClientName that
	// Store.NewClient does (store_client.go:535-538) -- factor that into a
	// shared helper rather than copying it.

	q := s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_auth_oauth_client")).
		Columns("id", "client_id", "source", "created_at", "updated_at",
			"last_fetched_at", "kind", "application_type", "client_name",
			"client_uri", "logo_uri", "tos_uri", "policy_uri",
			"redirect_uris", "grant_types", "response_types").
		Values(uuid.NewString(), options.ClientID, string(model.OAuthClientSourceCIMD),
			now, now, now, string(model.OAuthClientKindThirdParty),
			options.ApplicationType, clientName, options.ClientURI,
			options.LogoURI, options.TOSURI, options.PolicyURI,
			redirectURIs, grantTypes, responseTypes).
		Suffix(`ON CONFLICT (app_id, client_id) DO UPDATE SET
			updated_at = EXCLUDED.updated_at,
			last_fetched_at = EXCLUDED.last_fetched_at,
			application_type = EXCLUDED.application_type,
			client_name = EXCLUDED.client_name,
			client_uri = EXCLUDED.client_uri,
			logo_uri = EXCLUDED.logo_uri,
			tos_uri = EXCLUDED.tos_uri,
			policy_uri = EXCLUDED.policy_uri,
			redirect_uris = EXCLUDED.redirect_uris,
			grant_types = EXCLUDED.grant_types,
			response_types = EXCLUDED.response_types
		WHERE ` + s.SQLBuilder.TableName("_auth_oauth_client") + `.source = ?
		RETURNING id, created_at, (xmax = 0) AS created`, string(model.OAuthClientSourceCIMD))

	// ...
}
```

Details that matter:

- **`created_at` and `id` are preserved on update** — they are not in the `SET` list. `created_at` is the first-resolution timestamp; `id` is the GraphQL Node id and the DataLoader key, so churning it on every refetch would break relay identity for a client an admin is looking at. `RETURNING created_at` therefore returns the *original* value on an update, which is what `Client.CreatedAt` must carry.
- **`RegisteredAt()` stays `nil`** with no extra work: `Client.RegisteredAt()` (`client.go:52`) already returns `nil` for any non-DCR source. Spec § Mapping: "`registeredAt`: always `null` — there is no registration event, only a fetch."
- **`(xmax = 0) AS created`** is the standard Postgres idiom for distinguishing an insert from an update inside `ON CONFLICT DO UPDATE`. It is used rather than a separate `SELECT` because the alternative has a TOCTOU window that would let two concurrent first-resolutions both report `created = true` and both consume a quota slot. The advisory lock in §4.4 closes that window too, but reporting the truth from the write itself means the limit logic does not depend on the lock being correct.
- **`app_id` is supplied by `SQLBuilderApp`** — `s.SQLBuilder.Insert` injects it, as it does for `CreateClient`. Verify by reading `pkg/lib/infra/db/appdb`'s builder before writing the code; if the app builder does not inject on `Insert` (`CreateClient` does not list `app_id` in its `Columns` either, so it must), the `ON CONFLICT (app_id, client_id)` target still needs `app_id` present in the insert.
- **No new migration.** Every column already exists, including `last_fetched_at`.

New error in `pkg/lib/oauthclient/errors.go`:

```go
var ErrDynamicClientSourceConflict = apierrors.BadRequest.WithReason("DynamicClientSourceConflict").New("dynamic client id belongs to another source")
```

### 2.2 `Commands.UpsertCIMDClient` — and the cache invalidation that must not be forgotten

```go
// UpsertCIMDClient writes the row and schedules a resolver-cache
// invalidation from DidCommitTx, for the same reason DeleteClient does
// (commands.go:455-465): invalidating before the commit would leave a
// harmless extra miss if the tx rolled back, while skipping invalidation
// after a successful commit would serve stale metadata -- or worse, a
// cached NEGATIVE entry -- for up to dynamicClientCacheTTL.
//
// The negative entry is the case that actually bites. Queries.
// getClientByClientIDCached calls Cache.SetNotFound with a 30s TTL every
// time a client_id misses in Postgres (queries.go:407-409). The very first
// /oauth2/authorize for a new CIMD client_id does exactly that -- the
// freshness read in §4.2 misses, caches "not found", then this upsert
// creates the row. Without invalidation here, the resolveClient call
// immediately after would read that 30s-old negative entry and reject a
// client that was just successfully resolved. pkg/lib/oauthclient/
// cache_client.go:889-891 already anticipates this ("CIMD rows ARE mutable
// ... so the CIMD upsert path must invalidate too").
func (c *Commands) UpsertCIMDClient(ctx context.Context, options *UpsertCIMDClientOptions) (*Client, bool, error) {
	client, created, err := c.Store.UpsertCIMDClient(ctx, options)
	if err != nil {
		return nil, false, err
	}

	if !c.hooked {
		c.Database.UseHook(ctx, c)
		c.hooked = true
	}
	c.pendingInvalidations = append(c.pendingInvalidations, options.ClientID)
	return client, created, nil
}
```

`hooked`/`pendingInvalidations`/`WillCommitTx`/`DidCommitTx` already exist on `Commands` and need no change.

> **This is the single most likely bug in the whole plan set.** An implementation that forgets the invalidation passes every unit test (they stub the cache) and fails intermittently in e2e with "unauthorized_client" on the first authorization request for each new client. The e2e test in §7 is written specifically to catch it.

### 2.3 `Queries.GetClientByClientID` — an exported raw-row read

`cimd.Service` needs the persisted row's `LastFetchedAt` to decide whether to refetch. `Queries.GetClientConfigByClientID` throws that away (it returns a synthesized `*config.OAuthClientConfig`), and `getClientByClientIDCached` is unexported. Export a thin wrapper:

```go
// GetClientByClientID returns the persisted dynamic-client row, Redis-first
// and Postgres only on a miss, exactly like GetClientConfigByClientID --
// they share getClientByClientIDCached. It exists for callers that need the
// row's own fields (Source, LastFetchedAt) rather than a synthesized client
// config; pkg/lib/cimd's freshness check is the only such caller today.
//
// Returning the CACHED row for the freshness decision is deliberate. The
// cache TTL is 5 minutes and the refetch interval is 1 hour, so at worst a
// refetch happens 5 minutes late -- and in exchange, the common case
// (a warm, fresh CIMD client) costs exactly one Redis GET on the
// /oauth2/authorize path: no Postgres connection, no transaction, no
// outbound request.
func (q *Queries) GetClientByClientID(ctx context.Context, clientID string) (*Client, error) {
	return q.getClientByClientIDCached(ctx, clientID)
}
```

## 3. `pkg/lib/cimd` — the resolution service

### 3.1 Collaborator interfaces

```go
// pkg/lib/cimd/service.go

type ServiceOAuthClientCommands interface {
	UpsertCIMDClient(ctx context.Context, options *oauthclient.UpsertCIMDClientOptions) (*oauthclient.Client, bool, error)
	LockForClientCount(ctx context.Context, source oauthclient.Source) error
	CountClientsBySource(ctx context.Context, source model.OAuthClientSource) (uint64, error)
}

type ServiceOAuthClientQueries interface {
	GetClientByClientID(ctx context.Context, clientID string) (*oauthclient.Client, error)
}

type ServiceDatabase interface {
	WithTx(ctx context.Context, do func(ctx context.Context) error) error
	IsInTx(ctx context.Context) bool
}

// ServiceRateLimiter is a seam for Part 4. Part 3 binds it to a no-op
// implementation so this part is independently shippable; Part 4 replaces
// the binding with the real bucket checks and deletes the no-op.
type ServiceRateLimiter interface {
	CheckFetchAllowed(ctx context.Context) error
}

// ServiceUsageLimiter is a seam for Part 5, bound to a no-op here for the
// same reason.
type ServiceUsageLimiter interface {
	CheckStanding(ctx context.Context, name model.UsageName, currentCount int) error
	ReportStandingCreated(ctx context.Context, name model.UsageName, countBeforeCreate int)
}

type Service struct {
	OAuthConfig *config.OAuthConfig
	// OAuthFeatureConfig supplies insecure_http_allowed (Part 1
	// §2.4). Already fanned out by wire (pkg/lib/deps/deps_config.go:74).
	OAuthFeatureConfig *config.OAuthFeatureConfig
	Clock              clock.Clock
	Fetcher      *Fetcher
	Commands     ServiceOAuthClientCommands
	Queries      ServiceOAuthClientQueries
	Database     ServiceDatabase
	SingleFlight *FetchSingleFlight
	RateLimiter  ServiceRateLimiter
	UsageLimiter ServiceUsageLimiter
}
```

### 3.2 Errors — `apierrors.Kind`s, following the house pattern

`Service`'s two outcomes are `apierrors.Kind`s with constructor functions, matching `ratelimit.RateLimited` (`pkg/lib/ratelimit/error.go:12`) and `usage.UsageLimitExceeded` (`pkg/lib/usage/errors.go:8`) — the established shape for an error that a caller classifies:

```go
// CIMDUnresolvable is returned for EVERY fetch and validation failure mode:
// blocked address, timeout, non-2xx, oversize, invalid JSON, client_id
// mismatch, failed field validation, and "not allowed by allowed_domains".
// One Kind, one reason, one message -- spec § Authgear as an SSRF/Probing
// Oracle. The concrete cause is logged, never attached.
var CIMDUnresolvable = apierrors.Invalid.WithReason("CIMDUnresolvable")

// CIMDClientLimitExceeded is deliberately a DIFFERENT Kind. Spec § Error
// Handling: the client limit "is not a fetch-outcome signal -- it doesn't
// vary by target host or reveal anything about network reachability -- so
// it falls outside the uniform-error rule".
var CIMDClientLimitExceeded = apierrors.Forbidden.WithReason("CIMDClientLimitExceeded")

func ErrUnresolvable() error {
	return CIMDUnresolvable.New("client_id is not resolvable")
}

func ErrClientLimitExceeded() error {
	return CIMDClientLimitExceeded.New("the project has reached its CIMD client limit")
}
```

Three things follow, and they are the reason this is `apierrors` rather than plain sentinels:

- **`cimdResolutionError` becomes one mechanism instead of two.** Its chain classifies a CIMD error, a rate-limit error and everything else; with plain sentinels it would mix `errors.Is` for the first and `apierrors.IsKind` for the second (§6.3).
- **Attaching a `Kind` is safe here, and would not be for the internals.** Every fetch/validation failure collapses to the *same* Kind with the same reason and message, so even if something rendered it, every mode renders identically. Giving each failure mode its own Reason is what would leak, which is why [Part 2](2026-08-28-02-document-fetching.md) §2.2's internal sentinels stay plain `errors.New` and are never returned outside the package.
- **Never use `NewWithCause`/`NewWithInfo` for `CIMDUnresolvable`.** `Details` is rendered into the JSON body, so attaching the underlying network error would defeat the whole design. Plain `.New(msg)`; log the cause separately.

`apierrors.Invalid` (400) and `apierrors.Forbidden` (403) are chosen so the HTTP status is sane if one ever *is* rendered directly; at `/oauth2/authorize` the handler converts both to OAuth protocol errors and the `Name` is not used.

### 3.3 `EnsureClientResolved` — the algorithm

`EnsureClientResolved` returns `nil` for every `client_id` that is **not** a CIMD candidate — a static client, a `dcrc_` client, an unknown opaque string, or any URL when CIMD is disabled. "Not a candidate" is not an error: the caller proceeds to ordinary resolution, which succeeds or produces the ordinary unknown-client error. Only a `client_id` that *is* a candidate and could not be resolved returns `CIMDUnresolvable`.

It must be called **outside** any database transaction — it performs an outbound HTTP request with a 5s timeout, and holding a pooled Postgres connection across that is an availability bug at any real traffic level. It opens its own short write transaction afterwards, for the upsert only.

```go
func (s *Service) EnsureClientResolved(ctx context.Context, clientID string) error {
	cfg := s.OAuthConfig.ClientIDMetadataDocument
	if !cfg.IsEnabled() {
		return nil
	}

	// (1) Shape. No network access. Must use the same allowInsecureHTTP as
	// the read path's candidate check, or a client_id could be fetched here
	// and be unresolvable immediately afterwards.
	allowInsecureHTTP := s.OAuthFeatureConfig.GetClientIDMetadataDocument().IsInsecureHTTPAllowed()
	u, err := oauthclient.ParseCIMDClientID(clientID, allowInsecureHTTP)
	if err != nil {
		return nil
	}

	// (2) Static config wins, and is never fetched: spec § Client ID Format's
	// "pre-registering Client Identifier URLs" pattern. DCR needs no
	// equivalent check -- a dcrc_ id cannot parse as a URL.
	if _, ok := s.OAuthConfig.GetClient(clientID); ok {
		return nil
	}

	// (3) Domain trust, before anything that could touch the network. Only
	// here -- not on the read path, so removing a domain stops new clients
	// and refetches without breaking existing ones (Part 1 §4.1).
	if !oauthclient.IsCIMDClientIDAllowed(cfg.GetAllowedDomains(), u) {
		return ErrUnresolvable()
	}

	// (4) Freshness. One Redis GET in the common case (§4.2).
	existing, err := s.Queries.GetClientByClientID(ctx, clientID)
	switch {
	case err == nil:
		if existing.Source != model.OAuthClientSourceCIMD {
			return nil
		}
		if s.isFresh(existing) {
			return nil
		}
	case errors.Is(err, oauthclient.ErrDynamicClientNotFound):
		existing = nil
	default:
		// An infrastructure failure is not an unresolvable client.
		return err
	}

	// (5) Single-flight (§4.3). A Redis failure degrades to a possible
	// stampede, which beats refusing every request.
	acquired, err := s.SingleFlight.Acquire(ctx, clientID)
	if err != nil {
		acquired = true
	}
	if !acquired {
		if existing != nil {
			return nil
		}
		return ErrUnresolvable()
	}

	// (6) Rate limits (Part 4). Consumed only when a fetch is actually about
	// to happen, so a popular fresh client never burns tokens.
	if err := s.RateLimiter.CheckFetchAllowed(ctx); err != nil {
		return err
	}

	// (7) The only network access in the feature.
	if allowInsecureHTTP && !strings.EqualFold(u.Scheme, "https") {
		ServiceLogger.GetLogger(ctx).Warn(ctx, "cimd: fetching a metadata document over plaintext http",
			slog.String("client_id", clientID),
			slog.String("flag", "oauth.client_id_metadata_document.insecure_http_allowed"))
	}
	body, fetchErr := s.Fetcher.Fetch(ctx, u)
	var doc *Document
	if fetchErr == nil {
		doc, fetchErr = ParseAndValidate(clientID, body, allowInsecureHTTP)
	}
	if fetchErr != nil {
		// The cause is logged here and nowhere else; it never reaches a
		// response (Part 2 §2.2).
		ServiceLogger.GetLogger(ctx).WithError(fetchErr).
			With(slog.String("client_id", clientID)).
			Info(ctx, "cimd: failed to resolve client metadata document")
		if existing != nil {
			return nil // serve the stale record, §5
		}
		return ErrUnresolvable()
	}

	// (8) Persist. Short write transaction, opened only now.
	return s.Database.WithTx(ctx, func(ctx context.Context) error {
		return s.upsert(ctx, clientID, doc)
	})
}

// A NULL last_fetched_at is never fresh: that row was written by something
// other than this service.
func (s *Service) isFresh(c *oauthclient.Client) bool {
	if c.LastFetchedAt == nil {
		return false
	}
	return s.Clock.NowUTC().Sub(*c.LastFetchedAt) < RefetchInterval
}
```

### 3.4 `upsert` — with the limit check and the advisory lock

```go
func (s *Service) upsert(ctx context.Context, clientID string, doc *Document) error {
	// Serialize concurrent first-resolutions for this app so the
	// count-then-create sequence is atomic with respect to the quota. Same
	// reasoning and the same helper POST /oauth2/register uses
	// (handler_register.go:172-176); the lock key is already scoped per
	// source (store_client.go:750), so a CIMD resolution never serializes
	// against a DCR registration.
	if err := s.Commands.LockForClientCount(ctx, model.OAuthClientSourceCIMD); err != nil {
		return err
	}

	countBefore, err := s.Commands.CountClientsBySource(ctx, model.OAuthClientSourceCIMD)
	if err != nil {
		return err
	}

	client, created, err := s.Commands.UpsertCIMDClient(ctx, &oauthclient.UpsertCIMDClientOptions{
		ClientID:        clientID,
		ApplicationType: doc.ApplicationType,
		ClientName:      doc.ClientName,
		ClientURI:       doc.ClientURI,
		LogoURI:         doc.LogoURI,
		TOSURI:          doc.TOSURI,
		PolicyURI:       doc.PolicyURI,
		RedirectURIs:    doc.RedirectURIs,
		GrantTypes:      doc.GrantTypes,
		ResponseTypes:   doc.ResponseTypes,
	})
	// ... Part 5 inserts the CheckStanding/ReportStandingCreated calls
	// around this; see Part 5 §3 for why the check has to straddle the
	// upsert rather than precede it.
	_ = client
	_ = created
	_ = countBefore
	return err
}
```

Part 5 fills in the limit logic; Part 3 leaves the count read in place (it is needed either way and its position is what Part 5 reasons about) with a no-op `UsageLimiter`.

## 4. Concurrency & freshness mechanics

### 4.1 Why the fetch is outside a transaction

`ValidateRequestWithoutTx` is, as its name says, called with no open transaction (`pkg/auth/handler/oauth/authorize.go:44`). That is exactly what CIMD needs: a 5-second outbound HTTP call inside a Postgres transaction would pin a pooled connection for the duration, and an attacker driving cold `client_id`s at `/oauth2/authorize` would exhaust the pool long before hitting any rate limit. The `Database.WithTx` in step (8) opens *after* the fetch returns and covers only the advisory lock, the count and the upsert — a few milliseconds.

Corollary for the implementation: `EnsureClientResolved` must **not** be moved inside `doHandleRequestWithTx`, and `Service` must not be handed a transaction-scoped context. Assert this with `s.Database.IsInTx(ctx)` at the top of `EnsureClientResolved` and `panic` if true — a programming error, not a runtime condition, so a panic is the right response and it will surface in the first test that gets it wrong.

### 4.2 The freshness read is cache-first and that is a feature

Step (4) reads through `ClientCache` (5 minute TTL). For a warm, fresh CIMD client the entire CIMD code path on `/oauth2/authorize` is: two in-memory config reads, one `url.Parse`, one linear scan of `oauth.clients`, one Redis `GET`. No Postgres, no transaction, no outbound request. That is the shape the hot path must have, because `/oauth2/authorize` is a user-facing redirect.

The cost is that a refetch can be up to 5 minutes late (cache TTL) on top of the 1 hour interval. Against a 1 hour interval that is noise, and spec § Fetching already accepts "a developer waiting up to an hour to see an intentional document change take effect".

### 4.3 `FetchSingleFlight` — `pkg/lib/cimd/singleflight.go`

Without this, N concurrent authorization requests for the same stale client produce N simultaneous fetches of the same document — a self-inflicted amplification attack on the client's own server, and N racing upserts.

```go
// fetchLockTTL must exceed FetchTimeout so a lock is never released while
// its holder is still fetching, and must be short enough that a holder that
// dies mid-fetch does not block refetches for long.
const fetchLockTTL = 10 * time.Second

type FetchSingleFlight struct {
	Redis *appredis.Handle
	AppID config.AppID
}

// Acquire reports whether the caller may perform the fetch. It is a plain
// SET NX with a TTL -- the established pattern in this repo (see
// pkg/lib/analytic/first_auth_sink.go:117) -- and deliberately NOT a
// blocking lock: a caller that loses the race serves the stale record (or
// fails, if there is none) immediately rather than queueing behind a 5s
// network call.
//
// There is no Release. Letting the key expire costs at most fetchLockTTL of
// extra staleness on a 1 hour interval, and an explicit release would have
// to be careful not to delete a lock a later holder now owns. Not worth the
// complexity here.
func (f *FetchSingleFlight) Acquire(ctx context.Context, clientID string) (bool, error) {
	key := fmt.Sprintf("app:%s:cimd-fetch:%s", f.AppID, crypto.SHA256String(clientID))
	// clientID is hashed for the same reason ClientCache hashes it
	// (cache_client.go:898-907): it is an attacker-influenced URL and a ':'
	// in it would otherwise let one client_id collide with another's key
	// namespace.
	var acquired bool
	err := f.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		var err error
		acquired, err = conn.SetNX(ctx, key, "1", fetchLockTTL).Result()
		return err
	})
	return acquired, err
}
```

### 4.4 Why there is no negative caching of failed fetches

Spec § Error Handling is explicit: "a failed or invalid fetch is never reused, so a document author who fixes their document sees the fix take effect on the very next authorization request that needs it." So a failure is retried on the next request — deliberately. The controls that stop that becoming a DoS vector are the rate limits (Part 4) and the single-flight lock above, not a negative cache.

Note the interaction with the existing `ClientCache.SetNotFound` (30 s negative TTL): step (4)'s read *does* write a negative entry when the row is absent, so the next request within 30 s reads "not found" from Redis rather than Postgres — but it still proceeds to fetch, because "not found" and "stale" take the same branch. The negative entry saves a Postgres round trip and changes nothing else. The one thing it must not do is survive a successful upsert, which is what §2.2's invalidation is for.

## 5. Behavior on a failed refetch of an existing record: serve the stale record

**Decision:** if a record exists and the refetch fails for any reason, the authorization request proceeds using the existing record. Only a `client_id` with *no* record at all becomes `CIMDUnresolvable`.

Reasoning. Spec § Error Handling's "a failed or invalid fetch is never reused" is a statement about *fetch results* — a failed fetch never becomes a record — not about invalidating a record that a previous successful fetch established. Reading it the other way would mean a client's brief hosting outage breaks every login for every user of that client, including users mid-flow, which is a far worse failure than serving metadata up to an hour old. Spec § Where resolution happens also frames the record as "the *current known state* of the client", which a stale-but-known record is and an absent one is not.

Consequences to state plainly:

- A client that publishes a broken document, or takes its document offline, keeps working on its last-known-good metadata **indefinitely** — the record is never garbage collected. In particular, a redirect URI that the developer removed from their document stays valid until a successful refetch replaces it. An admin who needs that stopped now uses `deleteDynamicClient` — removing the domain from `allowed_domains` on its own does **not** stop an existing client (Part 1 §4.1, D5), it only prevents new ones and freezes this one's metadata permanently. The durable ban is delete + allowlist removal together.
- Every failed refetch attempt is retried on the next authorization request that finds the record stale, bounded by the single-flight lock and the rate limits. A permanently-dead document therefore generates roughly one fetch attempt per `fetchLockTTL` under sustained traffic, until the per-host rate limit engages.
- This is a **divergence from the strictest reading of the spec**; spec § Error Handling should gain a sentence saying so (§8).

An earlier draft considered a hard staleness cap (serve stale up to 24 h, then fail). Rejected: it introduces a second magic constant the spec does not define, and it converts a client-side outage into an Authgear-side outage on a delay, which is the same failure with worse timing.

## 6. The call site — `/oauth2/authorize` and nothing else

### 6.1 `AuthorizationHandler` gains one collaborator

```go
// pkg/lib/oauth/handler/handler_authz.go

// AuthorizationHandlerCIMDService resolves a CIMD Client Identifier URL by
// fetching its metadata document, as a side effect of /oauth2/authorize and
// nowhere else (docs/specs/cimd.md § Where resolution happens). It is a
// no-op for every client_id that is not a CIMD candidate.
type AuthorizationHandlerCIMDService interface {
	EnsureClientResolved(ctx context.Context, clientID string) error
}

type AuthorizationHandler struct {
	// ... existing fields ...
	CIMDService AuthorizationHandlerCIMDService
}
```

`//go:generate go tool mockgen -source=handler_authz.go` is already declared at the top of the file, so `make generate` regenerates `handler_authz_mock_test.go` with the new interface.

### 6.2 `ValidateRequestWithoutTx` — the hook

```go
func (h *AuthorizationHandler) ValidateRequestWithoutTx(
	ctx context.Context,
	r protocol.AuthorizationRequest,
) (context.Context, *AuthorizationParams, *AuthorizationResultError) {
	// CIMD resolution runs BEFORE resolveClient and outside any
	// transaction. It is a no-op unless client_id is a CIMD candidate, in
	// which case it may perform one outbound HTTP request and persist the
	// result; resolveClient below then reads that record like any other
	// dynamic client's.
	if err := h.CIMDService.EnsureClientResolved(ctx, r.ClientID()); err != nil {
		return ctx, nil, h.cimdResolutionError(ctx, r, err)
	}

	ctx, client := resolveClient(ctx, h.ClientResolver, r.ClientID())
	if client == nil {
		return ctx, nil, &AuthorizationResultError{
			ResponseMode: r.ResponseMode(),
			Response:     protocol.NewErrorResponse("unauthorized_client", "invalid client ID"),
		}
	}
	// ... unchanged from here on
}
```

### 6.3 Error mapping

```go
// cimdResolutionError maps a cimd.Service failure onto an authorization
// error response. There is no RedirectURI on any of these: without a
// resolved client there are no registered redirect URIs to validate the
// request's redirect_uri against, so every one of them must be rendered
// directly to the user agent rather than redirected. This is the same
// treatment the existing client == nil branch already gets.
func (h *AuthorizationHandler) cimdResolutionError(
	ctx context.Context,
	r protocol.AuthorizationRequest,
	err error,
) *AuthorizationResultError {
	switch {
	case apierrors.IsKind(err, cimd.CIMDUnresolvable):
		// Spec § Error Handling: "the authorization request fails the same
		// way it does today for a client_id that matches no known client".
		// Byte-identical to the client == nil branch below it, deliberately:
		// spec § Authgear as an SSRF/Probing Oracle requires that the error
		// not distinguish WHY resolution failed, and that includes not
		// distinguishing "CIMD fetch failed" from "no such client".
		return &AuthorizationResultError{
			ResponseMode: r.ResponseMode(),
			Response:     protocol.NewErrorResponse("unauthorized_client", "invalid client ID"),
		}
	case apierrors.IsKind(err, cimd.CIMDClientLimitExceeded):
		// Part 5. Spec § Error Handling makes this a distinct, non-uniform
		// error precisely because it carries no information about the fetch
		// target.
		return &AuthorizationResultError{
			ResponseMode: r.ResponseMode(),
			Response:     protocol.NewErrorResponse("access_denied", "the project has reached its client limit"),
		}
	case apierrors.IsKind(err, ratelimit.RateLimited):
		// Part 4.
		return &AuthorizationResultError{
			ResponseMode: r.ResponseMode(),
			Response:     protocol.NewErrorResponse("x_rate_limited", "rate limit exceeded, please try again later."),
		}
	default:
		logger := AuthorizationHandlerLogger.GetLogger(ctx)
		logger.WithError(err).Error(ctx, "cimd: unexpected error resolving client")
		return &AuthorizationResultError{
			ResponseMode:  r.ResponseMode(),
			Response:      protocol.NewErrorResponse("server_error", "internal server error"),
			InternalError: true,
		}
	}
}
```

Note `unauthorized_client` rather than `invalid_client`: spec § Error Handling asks for an "`invalid_client`-shaped error", and the shape the existing code produces for an unknown `client_id` at `/oauth2/authorize` is `unauthorized_client` (`handler_authz.go:1049`). Matching the existing behavior byte-for-byte is what makes the two indistinguishable, which is the actual requirement. Spec § Error Handling should say `unauthorized_client` (§8).

### 6.4 Every other endpoint: confirmed read-only, no change

| Endpoint / surface | Resolution path | Change needed |
|---|---|---|
| `/oauth2/token`, `authorization_code` grant | `resolveClient` → `Resolver` → cache/DB | none |
| `/oauth2/token`, `refresh_token` grant | same | none |
| `/oauth2/consent` (`prepareConsentRequest`, `handler_authz.go:492`) | same; the record exists because `/oauth2/authorize` created it earlier in the same flow | none |
| `/oauth2/userinfo` | same | none |
| `/oauth2/end_session` | same | none |
| `/oauth2/revoke` | same | none |
| `/resolve` (resolver endpoint) | same; the opaque-token gate is already source-agnostic | none |
| Auth UI view models (`viewmodels/base.go:216`) | same | Part 6 (display only) |
| Authorized Apps settings page | same | Part 6 |
| Admin API `dynamicClients` / `deleteDynamicClient` | `oauthclient.Queries`/`Commands` directly | none |

The redirect URI check at `/oauth2/authorize` uses `parseRedirectURI` → `client.RedirectURIs` from the freshly-resolved record, which is exactly spec § Redirect URI Validation's "exact match against the resolved client's `redirect_uris` — read from the persisted, shared record […] which reflects the client's current known state". No change to `resolve.go`.

The `authorization_code` grant's `redirect_uri` re-check at `/oauth2/token` validates against the URI bound to the code at issuance, per RFC 6749 §4.1.3 — unrelated to CIMD and unchanged, as spec § Redirect URI Validation notes.

### 6.5 Wiring

- `pkg/lib/cimd/deps.go`: add `wire.Struct(new(Service), "*")`, `wire.Struct(new(FetchSingleFlight), "*")`, the no-op rate/usage limiter providers, and `wire.Bind` for `ServiceOAuthClientCommands`/`ServiceOAuthClientQueries`/`ServiceDatabase` to `*oauthclient.Commands`, `*oauthclient.Queries`, `*appdb.Handle`.
- `pkg/lib/deps/deps_common.go`: bind `handler.AuthorizationHandlerCIMDService` to `*cimd.Service`, following the `facade.OAuthClientResolver` pattern at `deps_common.go:689-695`.
- `make generate` for wire and the authz handler mocks.
- `e2e/` has its own wire graph (`fc26fb42d chore: Regenerate e2e module's wire_gen.go for oauthclient.Store.AppID` is precedent) — regenerate it too and expect a diff.

## 7. Test Plan

**Unit — `pkg/lib/cimd/service_test.go` (new)**

Stub `Fetcher`'s HTTP client with an `httptest` server, stub `Commands`/`Queries`/`Database`, use `clock.MockClock`. Every branch of §3.3:

| Case | Expect |
|---|---|
| CIMD disabled, URL `client_id` | `nil`; fetcher never called; `Queries` never called |
| non-URL `client_id` (`dcrc_x`, `my-client`, `""`) | `nil`; fetcher never called |
| `client_id` matches a static client in `oauth.clients` | `nil`; **fetcher never called** — the pre-registration pattern |
| host fails `allowed_domains` | `CIMDUnresolvable`; fetcher never called |
| no record, fetch+validate ok | `UpsertCIMDClient` called once with `created` observed; options match the document field-for-field |
| record exists, `LastFetchedAt` 30 min ago | `nil`; fetcher **never called** |
| record exists, `LastFetchedAt` 61 min ago, fetch ok | upsert called; `created == false` |
| record exists, `LastFetchedAt` NULL | fetch attempted (never fresh) |
| record exists, fetch fails (each of: timeout, 404, oversize, invalid JSON, `client_id` mismatch) | `nil` — stale served; upsert **not** called; the cause appears in the log record and not in the returned error |
| **no** record, fetch fails, each of the same five modes | the returned errors are **mutually indistinguishable**: same `apierrors.Kind`, same message, and `errors.As` reaches neither `*net.OpError` nor any `ErrDocument*`/`ErrResponse*` sentinel. This is the SSRF-oracle invariant, asserted in the one place that owns it (Part 2 §2.2) |
| no record, fetch fails | `CIMDUnresolvable` |
| single-flight not acquired, record exists | `nil`; fetcher never called |
| single-flight not acquired, no record | `CIMDUnresolvable`; fetcher never called |
| single-flight `Acquire` returns an error (Redis down) | fetch proceeds |
| `Queries.GetClientByClientID` returns an infrastructure error (not `ErrDynamicClientNotFound`) | that error is returned unchanged → `server_error`, **not** `CIMDUnresolvable`. This distinction matters: a Postgres outage must not present as "your client_id is invalid". |
| called with `IsInTx(ctx) == true` | panics (§4.1) |
| existing row with `Source: DCR` under a URL client_id | `nil`; fetcher never called |

**Unit — `pkg/lib/oauthclient/store_client_test.go` (extend)**

`UpsertCIMDClient` against the real database harness the existing store tests use:

- first call inserts, `created == true`, `LastFetchedAt` set, `RegisteredAt() == nil`, `Source == CIMD`, `Kind == THIRD_PARTY`;
- second call with different metadata updates in place, `created == false`, `id` and `created_at` **unchanged**, `updated_at` and `last_fetched_at` advanced, `redirect_uris` replaced (not merged);
- a pre-existing DCR row with the same `client_id` → `ErrDynamicClientSourceConflict`, and the DCR row is untouched;
- two apps with the same `client_id` string do not collide.

**Unit — `pkg/lib/oauthclient/cache_client_test.go` / `commands` test**

`Commands.UpsertCIMDClient` registers the DB hook and, on `DidCommitTx`, calls `Cache.Delete` with the exact `client_id`. Also: `Cache.SetNotFound` followed by `UpsertCIMDClient` + `DidCommitTx` followed by `Get` reports a miss, not the negative entry. This is the §2.2 bug, in unit form.

**Unit — `pkg/lib/oauth/handler/handler_authz_test.go` (extend)**

With a mocked `AuthorizationHandlerCIMDService`:

- `EnsureClientResolved` is called with the request's `client_id` **before** `ClientResolver.ResolveClient` (assert ordering via `gomock.InOrder`);
- `CIMDUnresolvable` → response is byte-identical to the existing unknown-`client_id` response (`unauthorized_client` / `invalid client ID`), with no `RedirectURI`;
- `CIMDClientLimitExceeded` → `access_denied`;
- a `ratelimit.RateLimited` error → `x_rate_limited`;
- any other error → `server_error` with `InternalError: true`;
- `nil` → the flow proceeds unchanged (regression guard for every existing static-client test in the file).

**e2e — `e2e/tests/cimd/` (new directory)**

CIMD is the first Authgear feature whose e2e coverage requires **authgear itself to make an outbound request to a test-controlled host**. The [Part 1 §2.4](2026-08-28-01-config-and-client-id.md) feature-config flags exist precisely to make this possible, and they are what make it straightforward.

What the harness looks like today (`e2e/run.sh`, `e2e/docker-compose.yaml`, `e2e/cmd/`, `e2e/.env`):

- Postgres, Redis, `deno` and a `hook` server run in compose; the `hook` server (`e2e/cmd/hookserver`, port `2626`) is **plain HTTP** and has a built-in request `recorder` — the model to copy.
- **authgear runs on the host**, not in compose (`run.sh`: `authgear start`, health-checked at `http://localhost:4000/healthz`), reading `e2e/.env` via `godotenv.Load()` from `cmd/authgear/main.go:45`.
- **`DEV_MODE` is `false`.** `e2e/.env:1` sets `true`, then line 74 overrides it to `DEV_MODE=false # required to send email`; `godotenv` is last-wins (verify with `godotenv.Read`), and `pkg/lib/messaging/sender.go:100,233,382` plus `pkg/lib/usage/usage_alert_email_service.go:53` suppress delivery under `DevMode`. So `DEV_MODE` is unavailable as a gate here — which is one of the reasons Part 1 §2.4 does not use it.
- `e2e/cmd/proxy` terminates TLS on `:8080` with `e2e/ssl/ca.crt`/`ca.key`, and `e2e/pkg/e2eclient/client.go:118` loads that CA into the **test client's** trust pool. Nothing makes **authgear** trust it — and with the feature flags in place, nothing needs to.
- Per-test config overrides already exist for **both** files: `authgear.yaml: override:` and `authgear.features.yaml: override:` (`pkg/testrunner/models.go:38`; see `tests/dcr/register_usage_limit.test.yaml` for a features override in use). Each test file gets its own project, so two test files can hold two different postures in the same run.

**Approach.** A plain-HTTP document server plus per-project feature flags — no TLS, no CA trust, no OS-specific certificate handling, and no test-only knob in production code:

1. New `e2e/cmd/cimdserver` — a plain **HTTP** server on `127.0.0.1:2727`. It serves `e2e/tests/cimd/documents/*.json` by path, keeps a per-path hit counter exposed on a control endpoint (copy `hookserver`'s `recorder`, `e2e/cmd/hookserver/main.go:15-38`), and supports per-path controls for the failure cases: respond with an arbitrary status, respond with N bytes, redirect, and go offline. Started and torn down in `run.sh` alongside `e2e-smtp`, with `kill_port 2727` in `teardown`.
2. Permissive test projects set, in their own `authgear.features.yaml: override:`:
   ```yaml
   oauth:
     client_id_metadata_document:
       insecure_http_allowed: true
       insecure_fetch_address_allowed: true
   ```
   and use `client_id: http://localhost:2727/<name>.json`. (`localhost` resolves to `127.0.0.1`, which the address filter blocks without the second flag — hence both.)
3. Strict test projects set **neither** flag, and are used to assert the protections actually hold (§7.1 below). This is the coverage a process-wide switch could not provide.

Note the fixture documents carry `"client_id": "http://localhost:2727/<name>.json"` — the byte-for-byte equality check (Part 2 §3) is unchanged and still exercised; only the scheme differs.

**Rejected alternatives**, both superseded by the feature flags:

- *HTTPS on loopback with `SSL_CERT_FILE` pointing at the e2e CA.* Works on Linux/CI, but Go's `crypto/x509` ignores `SSL_CERT_FILE` on macOS (the Security framework path), so `make -C e2e run` would fail on a developer Mac until the CA was added to the login keychain. Rejected as an unnecessary platform trap.
- *A test-only env var on the fetcher* (`CIMD_ALLOW_LOOPBACK_FETCH`-style). Same effect, but it is a knob in production code whose only purpose is testing, with no per-project scoping and therefore no way to test the strict path in the same run.

### 7.1 Tests asserting the protections still hold

These matter more than the happy path, because they are the ones that would silently rot. Each uses a project with **no** feature flags set:

| File | Asserts |
|---|---|
| `strict_rejects_http.test.yaml` | `client_id: http://localhost:2727/valid.json` → the ordinary unknown-client error, and the document server's hit counter for that path is **0**. Proves `insecure_http_allowed` defaults off and is actually load-bearing. |
| `strict_rejects_private_address.test.yaml` | A project with `insecure_http_allowed: true` but **not** `insecure_fetch_address_allowed`, using `client_id: http://localhost:2727/valid.json` → refused, hit counter **0**. This is the composition case from Part 1 D13: the scheme passes shape validation and the address filter still refuses. It is the single most valuable test in this file, because it proves the two flags are independent and that the address filter — not the scheme rule — is what stops SSRF. |
| `revoked_flags_keep_existing_clients.test.yaml` | Seed a CIMD row with an `http://` `client_id` via `custom_sql`, plus `last_fetched_at = now()` so it is fresh, against a project with **no** flags set. Authorizing with it **succeeds** — trust policy gates fetching, not reading (Part 1 §4.1, D5). Then a *different*, unseeded `http://` `client_id` in the same project is refused with **zero** hits on the document server. This pair is the whole product rule in one file: revoking a flag stops new clients, never existing ones. |
| `revoked_allowed_domains_keeps_existing.test.yaml` | Same shape for `allowed_domains`: seed a fresh row for `http://localhost:2727/a.json` in a project whose `allowed_domains` is `["*.example.com"]` → authorization **succeeds**; an unseeded `client_id` on that host is refused with zero fetch attempts. |

### 7.2 Happy-path and behavior tests

All against a permissive project. Tests to write:

| File | Asserts |
|---|---|
| `resolution.test.yaml` | Full UC1: `/oauth2/authorize` with a CIMD `client_id` → consent → code → `/oauth2/token` with `resource=` → tokens. Then `dynamicClients` shows one client with `source: CIMD`, `registeredAt: null`, non-null `lastFetchedAt`. |
| `first_request_cache.test.yaml` | Two authorization requests in a row for a brand-new `client_id`; the **first** must succeed. This is the §2.2 negative-cache bug: without the invalidation the first request creates the row and then fails to resolve it. |
| `redirect_uri.test.yaml` | A `redirect_uri` not in the document → the request is refused; a loopback `redirect_uri` from the document with no `application_type` → accepted. |
| `disabled.test.yaml` | CIMD disabled → a URL `client_id` gives the ordinary unknown-client error, and (assert on the document server's hit counter) **no fetch is attempted**. |
| `allowed_domains.test.yaml` | An app with `allowed_domains: ["*.example.com"]`: a `localhost:2727` `client_id` is refused with **zero** hits on the document server. Then a second app with `allowed_domains: ["localhost"]` → resolution succeeds, which also covers Part 1 D15 (single-label patterns are valid config). |
| `stale_record.test.yaml` | Seed a CIMD row via `custom_sql` with `last_fetched_at = now() - interval '2 hours'` (stale) and a `logo_uri`/`client_name` distinguishable from what the document server serves. Then: (a) with the document path returning 503 → authorization **succeeds** on the stale record (§5); (b) with the document path healthy → authorization succeeds and `dynamicClients` now shows the *document's* metadata, proving the refetch replaced the row in place. `custom_sql` is a `before` hook only (`pkg/testrunner/models.go:51,137`), so this needs two test files rather than two steps in one — or a step-level SQL action if one is added. `before`-hook seeding is what makes the refetch interval testable without any clock control or config knob; do **not** make `RefetchInterval` configurable to make this easier. |
| `insecure_fetch_is_logged.test.yaml` | Optional, and only if the harness can assert on `logs/authgear.log`: a successful insecure fetch emits the Part 1 D16 `Warn` records. If it cannot, the `clientFor` unit test (Part 2 §6) is the coverage and this file is skipped. |
| `invalid_document.test.yaml` | `client_id` mismatch, non-2xx, oversize, malformed JSON — each gives the *same* error body as an unknown `client_id`. |
| `static_client_precedence.test.yaml` | A static client whose `client_id` is `https://…/x`; authorization succeeds using the static config, and the document server records **zero** hits. |
| `token_endpoint_no_fetch.test.yaml` | After a successful authorize, take the document server offline (or assert its hit count), then run the `authorization_code` and `refresh_token` grants — both succeed with no further fetch. |

**Commands to run**

```
go test ./pkg/lib/cimd/... ./pkg/lib/oauthclient/... ./pkg/lib/oauth/...
make generate && git status --porcelain   # must be empty
make lint
make -C e2e run
```

## 8. Spec Updates

`doc:` commit against `docs/specs/cimd.md`:

1. **§ Error Handling**: state that a *refetch* failure for a `client_id` that already has a persisted record does **not** make the client unresolvable — the existing record is served and the refetch is retried on the next request that finds it stale. Only a `client_id` with no record at all fails. Note the consequence: a client whose document goes offline keeps working on its last-known metadata indefinitely, and `deleteDynamicClient` / `allowed_domains` are the levers.
2. **§ Error Handling**: name the actual error code — `unauthorized_client`, which is what `/oauth2/authorize` already returns for an unknown `client_id` — rather than "`invalid_client`-shaped".
3. **§ Denial of Service**: clarify that the limits are consumed per **fetch attempt**, not per `/oauth2/authorize` request, so a project with many popular, fresh CIMD clients is not throttled by its own legitimate traffic.
4. **§ Where resolution happens**: add that `allowed_domains` is evaluated only on the fetch path, so removing a domain stops new clients and refetches but leaves existing clients working (Part 1 §4.1); `enabled` *is* evaluated on every read path, being a feature switch.

## 9. Fixed Behavioral Decisions

- **D1. `EnsureClientResolved` has exactly one call site, `ValidateRequestWithoutTx`.** Any additional call site is a spec change. It panics if called inside a transaction.
- **D2. "Not a CIMD candidate" returns `nil`, not an error.** The caller falls through to ordinary resolution. Only a candidate that could not be resolved errors.
- **D3. A static client always wins over a fetch, and no fetch is attempted for it.** Spec § Client ID Format's pre-registration pattern depends on this.
- **D4. `allowed_domains` is checked before any network access, and only there** — not on the read path (Part 1 §4.1, D5).
- **D5. Rate limits are consumed per fetch attempt, not per request.** A fresh record short-circuits before the limiter.
- **D6. A failed refetch serves the stale record; only a missing record is unresolvable.** §5. Divergence from the strictest spec reading, documented in the spec.
- **D7. No negative caching of failed fetches.** Spec § Error Handling requires next-request retry; single-flight and rate limits are the controls.
- **D8. Single-flight is non-blocking.** A caller that loses the race serves stale or fails immediately rather than queueing behind a network call. The lock is never explicitly released.
- **D9. `id` and `created_at` survive a refetch.** They are relay identity, not fetch metadata.
- **D10. `CIMDUnresolvable` produces a response byte-identical to the unknown-`client_id` response.** Required by spec § Authgear as an SSRF/Probing Oracle; a test asserts equality rather than just the error code, and a second asserts every distinct failure mode yields the identical error out of `EnsureClientResolved`.
- **D11. An infrastructure error is not `CIMDUnresolvable`.** A Postgres or Redis failure must surface as `server_error`, never as "invalid client".
- **D12. `Service`'s two outcomes are `apierrors.Kind`s; `pkg/lib/cimd`'s internal errors are plain sentinels.** §3.2 and Part 2 §2.2. One Kind for every fetch/validation failure makes attaching a Kind safe; per-mode Kinds would leak. `NewWithCause`/`NewWithInfo` are forbidden on `CIMDUnresolvable`, since `Details` is rendered.

## 10. Atomic Commit Plan

1. `[CIMD] Add UpsertCIMDClient to the dynamic client store` — §2.1, §2.3 + store tests.
2. `[CIMD] Invalidate the resolver cache on a CIMD upsert` — §2.2 + commands/cache tests.
3. `[CIMD] Add the single-flight guard for metadata document fetches` — §4.3 + test.
4. `[CIMD] Add the CIMD client resolution service` — §3, §3.2, §3.3, §3.4 (with no-op rate/usage limiters), §6.5 wiring for the service itself + `service_test.go`.
5. `[CIMD] Resolve CIMD clients at the authorize endpoint` — §6.1, §6.2, §6.3, remaining §6.5 wiring, `make generate` for wire + authz mocks + `authz` handler tests.
6. `[CIMD] Add a metadata document host to the e2e environment` — the §7 `e2e/cmd/cimdserver` work plus its `run.sh`/teardown wiring, on its own so the fixture server is reviewable in isolation.
7. `[CIMD] Add e2e tests for the CIMD fetch protections` — §7.1. Deliberately **before** the happy-path tests: these are the assertions that the escape hatches are off by default and that the address filter is what stops SSRF, and they should land as their own reviewable unit.
8. `[CIMD] Add e2e tests for CIMD client resolution` — §7.2.
9. `doc: Clarify cimd.md on stale records, error codes and fetch-scoped limits` — §8.

Body of each: `ref DEV-XXXX`. Run `make update-vettedpositions` after commit 5 (`handler_authz.go` line numbers move).
