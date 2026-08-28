# DCR Part 3 — Resolve DCR Clients at Runtime

Spec: [docs/specs/dcr.md — Storage Architecture](../../specs/dcr.md#storage-architecture), [docs/specs/client.md](../../specs/client.md).

Depends on Part 2 (`_auth_oauth_client` table, the `pkg/lib/oauthclient` package with `Client`/`Store.GetClientByClientID`/`ToClientConfig`, `config.OAuthDynamicClientRegistrationConfig`).

> Everything in this part is written against `pkg/lib/oauthclient`, not `pkg/lib/dcr` — see [Part 2 §4](2026-08-17-02-client-registration.md) for the package split. The resolver resolves *any* dynamic client from the shared table; DCR is simply the only `source` that exists yet.

## 1. Goal / Scope

After Part 2, a DCR client can be registered and inspected via Admin API, but `/oauth2/authorize`, `/oauth2/token`, and every other place that resolves a `client_id` still only knows about `authgear.yaml` clients — a `dcrc_...` client_id is simply "not found" everywhere else. This part makes the **existing, single central resolver** (`pkg/lib/oauthclient.Resolver`) DB-aware, so a DCR client behaves identically to a static client of the same `kind`/`application_type` everywhere in the codebase, per client.md: "The runtime behavior of a DCR client ... is identical to that of a static client with the same `kind` and `application_type`."

Out of scope: `access_policy`/resource-indicator support (Part 4); actually revoking a DCR client's outstanding tokens on `deleteDynamicClient` (flagged as a gap in Part 2 §6.4 — depends on this part's resolver change existing first, but the revocation wiring itself is not included here either; call this out again at the end of this file).

## 2. Central Finding — There Is No Existing "Third-Party + Public" Client Type

`config.OAuthClientConfig`'s trust/confidentiality/PII rules are all switch statements keyed on the single `ApplicationType OAuthClientApplicationType` field (`pkg/lib/config/oauth.go:353-375`, delegating to `OAuthClientApplicationType.IsThirdParty/IsConfidential/IsClientCredentialsFlowAllowed/HasFullAccessScope/PIIAllowedInIDToken`, `pkg/lib/config/oauth.go:41-142`). Confirmed exact behavior of the six enum values:

| `x_application_type` | IsThirdParty | IsConfidential | IsClientCredentialsFlowAllowed | HasFullAccessScope | PIIAllowedInIDToken |
|---|---|---|---|---|---|
| `spa` | false | false | false | true | false |
| `traditional_webapp` | false | false | false | true | false |
| `native` | false | false | false | true | false |
| `confidential` | false | true | true | false | true |
| `third_party_app` | **true** | **true** | false | false | true |
| `m2m` | false | true | true | false | false |

**There is no existing value with `IsThirdParty=true, IsConfidential=false`.** Every third-party client the codebase has ever known (`third_party_app`, statically configured) is confidential. But per dcr.md, **all** DCR clients — including third-party ones — are public ("DCR clients are always public, so `isConfidential` is always `false`"). Naively mapping a DCR third-party client onto `OAuthClientApplicationTypeThirdPartyApp` would make it incorrectly report `IsConfidential() == true` everywhere `OAuthClientConfig.IsConfidential()`/`.IsPublic()` is consulted (token endpoint client authentication requirements, full-access scope gating, etc.).

**Decision:** add a new, internal-only enum value used exclusively for synthesizing a *dynamic* third-party client's config — never accepted from `authgear.yaml`. It is named for the client *shape* (third-party + public), not for DCR, because a CIMD client is the same shape: cimd.md fixes `kind: THIRD_PARTY` and `isConfidential: false` for every CIMD client, so CIMD reuses this value unchanged rather than adding a second one. The value is synthetic and never persisted anywhere, so this naming choice costs nothing to make now and would be a churn-y rename later.

### 2.1 `pkg/lib/config/oauth.go` changes

```go
const (
	OAuthClientApplicationTypeSPA            OAuthClientApplicationType = "spa"
	OAuthClientApplicationTypeTraditionalWeb OAuthClientApplicationType = "traditional_webapp"
	OAuthClientApplicationTypeNative         OAuthClientApplicationType = "native"
	OAuthClientApplicationTypeConfidential   OAuthClientApplicationType = "confidential"
	OAuthClientApplicationTypeThirdPartyApp  OAuthClientApplicationType = "third_party_app"
	OAuthClientApplicationTypeM2M            OAuthClientApplicationType = "m2m"
	OAuthClientApplicationTypeUnspecified    OAuthClientApplicationType = ""

	// OAuthClientApplicationTypeDynamicThirdParty is synthetic: it is never
	// present in authgear.yaml and is NOT added to the "OAuthClientConfig"
	// JSON Schema's x_application_type enum (pkg/lib/config/oauth.go:153).
	// It exists solely so a dynamically-resolved third-party client — DCR
	// (dcr.md) or CIMD (cimd.md) — can be represented as a
	// *config.OAuthClientConfig: the one client shape this codebase has never
	// had, third-party AND public. See
	// docs/plans/dcr/2026-08-17-03-client-resolution.md §2.
	OAuthClientApplicationTypeDynamicThirdParty OAuthClientApplicationType = "x_dynamic_third_party"
)
```

Add an explicit `case OAuthClientApplicationTypeDynamicThirdParty:` to **every** switch in `pkg/lib/config/oauth.go`, matching third-party-client.md's table exactly (not copied from `third_party_app` — the confidentiality differs):

| Method | Existing `third_party_app` value | New `x_dynamic_third_party` value | Source |
|---|---|---|---|
| `IsThirdParty()` (line 41) | `true` | **`true`** | unchanged from third_party_app |
| `IsConfidential()` (line 64) | `true` | **`false`** | dcr.md: "DCR clients are always public" |
| `IsClientCredentialsFlowAllowed()` (line 83) | `false` | `false` | third-party-client.md: `client_credentials` "Not allowed" |
| `HasFullAccessScope()` (line 106) | `false` | `false` | third-party-client.md: full-access "Not allowed" for third-party |
| `PIIAllowedInIDToken()` (line 125) | `true` | `true` | third-party-client.md: "Third-party clients ... PII in ID token: Yes" |

Explicit `case` statements are required, not reliance on `default` — `HasFullAccessScope()`'s `default` branch returns `true` (line 121), the opposite of what a third-party client needs; an omitted case here would silently grant full-access scope eligibility to every dynamic third-party client.

A first-party DCR client (`web` or `native`) needs **no new enum value** — `OAuthClientApplicationTypeSPA` (for `web`) / `OAuthClientApplicationTypeNative` (for `native`) already have exactly the right values (`IsThirdParty=false, IsConfidential=false`), matching dcr.md's client.md mapping table for `FIRST_PARTY` DCR clients.

## 3. `pkg/lib/oauthclient` — Dynamic Client → `config.OAuthClientConfig` Conversion

> **Moved to Part 2.** `ToClientConfig` is a pure function (it builds a `*config.OAuthClientConfig` and calls its `SetDefaults()`) with no runtime wiring, and Part 2's `Client.ToModel` needs it in order to report the same resolved token lifetimes the runtime enforces — see [Part 2 §4.1](2026-08-17-02-client-registration.md) and §2.2.1 there. It is therefore introduced in Part 2 and merely *consumed* here. The listing below is retained as the reference for what it does; commit 2 of §9 becomes a no-op if Part 2 already landed it.

`pkg/lib/oauthclient/client_config.go`:

```go
package oauthclient

import "github.com/authgear/authgear-server/pkg/lib/config"

// ToClientConfig synthesizes a *config.OAuthClientConfig for this dynamic client,
// making it usable everywhere a static-config client is (see
// docs/plans/dcr/2026-08-17-03-client-resolution.md). defaults comes from
// config.OAuthConfig.DynamicClientRegistration.DefaultClientConfig, which is
// always non-nil for DCR once config defaults have run (Part 2 §2.2) --
// ToClientConfig's nil branch below exists for other sources only.
func (c *Client) ToClientConfig(defaults *config.OAuthDynamicClientRegistrationDefaultClientConfig) *config.OAuthClientConfig {
	var appType config.OAuthClientApplicationType
	switch {
	case c.Kind == model.OAuthClientKindThirdParty:
		appType = config.OAuthClientApplicationTypeDynamicThirdParty
	case c.ApplicationType == "native":
		appType = config.OAuthClientApplicationTypeNative
	default: // "web", first-party
		appType = config.OAuthClientApplicationTypeSPA
	}

	cfg := &config.OAuthClientConfig{
		ClientID:                        c.ClientID,
		ClientName:                      derefOr(c.ClientName, ""),
		Name:                            c.DisplayName(),
		ApplicationType:                 appType,
		RedirectURIs:                    c.RedirectURIs,
		GrantTypes_do_not_use_directly:  c.GrantTypes,
		ResponseTypes:                   c.ResponseTypes,
		ClientURI:                       derefOr(c.ClientURI, ""),
		LogoURI:                         derefOr(c.LogoURI, ""),
		TOSURI:                          derefOr(c.TOSURI, ""),
		PolicyURI:                       derefOr(c.PolicyURI, ""),
		IssueJWTAccessToken:             false, // fixed per client.md's DCR mapping table
	}
	// nil here means this source has no default_client_config concept at all
	// (e.g. a not-yet-implemented CIMD), not "the admin configured no
	// override" -- for DCR, defaults is always non-nil with real
	// token-lifetime values once config defaults have run. See Part 2 §2.2.
	if defaults != nil {
		cfg.AccessTokenLifetime = defaults.AccessTokenLifetime
		cfg.RefreshTokenLifetime = defaults.RefreshTokenLifetime
		cfg.RefreshTokenIdleTimeoutEnabled = defaults.RefreshTokenIdleTimeoutEnabled
		cfg.RefreshTokenIdleTimeout = defaults.RefreshTokenIdleTimeout
	}
	cfg.SetDefaults() // reuse the exact same fallback logic static clients get (pkg/lib/config/oauth.go:376-398)
	return cfg
}
```

Reusing `OAuthClientConfig.SetDefaults()` (rather than duplicating its fallback arithmetic) means a DCR client with no `default_client_config` override gets exactly the same computed lifetimes a static client with no explicit lifetime fields would — and it is what makes the always-non-nil-but-zero `defaults` above behave identically to a true absence, since `SetDefaults()` tests each field against `0`/`nil`.

### 3.1 `pkg/lib/oauthclient/queries.go` — add `GetClientConfigByClientID`

```go
type Queries struct {
	Store    *Store
	Clock    clock.Clock
	Cache    *ClientCache   // §3.2
	Database *appdb.Handle  // §4.2
}

func (q *Queries) GetClientConfigByClientID(ctx context.Context, clientID string, defaults *config.OAuthDynamicClientRegistrationDefaultClientConfig) (*config.OAuthClientConfig, error) {
	c, err := q.getClientByClientIDCached(ctx, clientID)
	if err != nil {
		return nil, err // ErrDynamicClientNotFound
	}
	return c.ToClientConfig(defaults), nil
}

// getClientByClientIDCached consults Redis first and only touches Postgres on
// a miss. See §3.2 for the cache and §4.2 for why the DB scope is opened here
// rather than by the caller.
func (q *Queries) getClientByClientIDCached(ctx context.Context, clientID string) (*Client, error) {
	c, found, err := q.Cache.Get(ctx, clientID)
	if err != nil {
		// A cache failure must never take the endpoint down; fall through to
		// the database. Log at warn, same posture as
		// Limiter.maybeDispatchUsageAlert's non-fatal error handling.
		c, found = nil, false
	}
	if found {
		return c, nil // c == nil means a cached negative result
	}

	read := func(ctx context.Context) error {
		c, err = q.Store.GetClientByClientID(ctx, clientID)
		return err
	}
	if q.Database.IsInTx(ctx) {
		err = read(ctx)
	} else {
		err = q.Database.ReadOnly(ctx, read)
	}

	switch {
	case errors.Is(err, ErrDynamicClientNotFound):
		_ = q.Cache.SetNotFound(ctx, clientID)
		return nil, err
	case err != nil:
		return nil, err
	}
	_ = q.Cache.Set(ctx, c)
	return c, nil
}
```

On a cache hit this opens **no** database transaction and takes **no** connection from the Postgres pool — which is the point (§3.2).

### 3.2 `pkg/lib/oauthclient/cache_client.go` (new file) — Redis cache for resolved dynamic clients

Without a cache, every `/oauth2/authorize`, `/oauth2/token`, consent render and view-model construction involving a DCR client costs a Postgres connection checkout plus `BEGIN`/`SELECT`/`COMMIT`, since §4.2 opens a `ReadOnly` scope per resolve. A Redis `GET` replaces all of that. (An earlier draft of this plan recorded "no caching" as an accepted v1 tradeoff; that is reversed here — it is cheap enough not to accept, and §4.2 makes the per-resolve cost more visible than it first appeared.)

Follow `pkg/lib/oauth/redis/store.go` exactly: it is the established shape for an app-scoped Redis store in this codebase (`Redis *appredis.Handle`, `AppID config.AppID`, `WithConnContext`, JSON marshal, `errors.Is(err, goredis.Nil)` → not-found, keys built in a `keys.go` helper as `app:%s:...`).

```go
type ClientCache struct {
	Redis *appredis.Handle
	AppID config.AppID
	Clock clock.Clock
}

const (
	// A DCR client is immutable until RFC 7592 lands, so its only mutation is
	// deletion, which invalidates explicitly (below). This TTL is purely the
	// backstop for an invalidation that failed to run.
	//
	// CIMD rows ARE mutable (a refetch overwrites in place), so the CIMD
	// upsert path must invalidate too — see the note on Source below.
	dynamicClientCacheTTL = 5 * time.Minute
	// Negative entries expire faster: they exist to blunt read amplification,
	// not to be authoritative.
	dynamicClientCacheNotFoundTTL = 30 * time.Second
)

func redisKeyDynamicClient(appID string, clientID string) string {
	// The key is keyed by client_id alone, not by source: a client_id belongs
	// to exactly one source (§3.1's unique index), and the resolver looks up
	// by client_id without knowing the source yet.
	//
	// clientID is caller-influenced for CIMD (it is a URL), so it must be
	// hashed or escaped rather than interpolated raw — a ':' in a URL would
	// otherwise let one client_id's key collide with another's namespace.
	// Use crypto.SHA256String(clientID), matching how every other
	// caller-influenced Redis key in pkg/lib/oauth/redis/keys.go is a hash.
	return fmt.Sprintf("app:%s:dynamic-client:%s", appID, crypto.SHA256String(clientID))
}

func (c *ClientCache) Get(ctx context.Context, clientID string) (client *Client, found bool, err error)
func (c *ClientCache) Set(ctx context.Context, client *Client) error
func (c *ClientCache) SetNotFound(ctx context.Context, clientID string) error
func (c *ClientCache) Delete(ctx context.Context, clientID string) error
```

Four design decisions worth stating explicitly, because each has a wrong-looking alternative:

**Cache the `oauthclient.Client` row, not the synthesized `*config.OAuthClientConfig`.** The synthesized config folds in `oauth.dynamic_client_registration.default_client_config` from `authgear.yaml`, which changes on config deploy **without** the database row changing. Caching the synthesized config would pin stale token lifetimes for up to the TTL after an admin edits `default_client_config`. `ToClientConfig()` is pure in-memory work, so re-running it per resolve costs nothing and keeps config edits instantaneous.

**Negative caching is safe here, unlike in most caches.** A `client_id` is 16 bytes of server-generated entropy (§8 / Part 1 §5.2), so a negative entry can never collide with a client created later — there is no "cache says no, then it exists" race that matters. This closes a small amplification hole: `/oauth2/authorize` has no rate limit (Part 2 §5.1.1 only covers `/oauth2/register`), so without negative caching, `client_id=dcrc_<random>` forces one Postgres round-trip per request from an unauthenticated caller. Note the contrast with CIMD, where `client_id` is a caller-supplied URL — negative caching there would need different reasoning.

**Invalidate on delete via `DidCommitTx`, not inline.** `pkg/lib/infra/db.TransactionHook` already declares `WillCommitTx(ctx) error` / `DidCommitTx(ctx)`, and `HookHandle.UseHook` registers one. `Commands.DeleteClient` must invalidate from `DidCommitTx`, because the failure modes are asymmetric:

- invalidate before commit, transaction then rolls back → cache empty while the row still exists → a harmless extra miss;
- commit succeeds but invalidation never runs → **a deleted client keeps authenticating** until the TTL expires.

Only the second is a correctness problem, so the invalidation must run after the commit is known to have succeeded, with `dynamicClientCacheTTL` as the bound on the residual window. When RFC 7592 adds `PUT /oauth2/register/{client_id}`, that path needs the same `DidCommitTx` invalidation — note it in the spec's Future Works so it is not forgotten.

**Do not cache anything else.** Specifically: `Store.CountClientsBySource` (Part 5's quota check) must always hit Postgres, or the standing limit becomes enforceable-in-name-only; and the Admin API's `dynamicClients` list (Part 2 §6.3) must read through, since an admin looking at the client list expects to see a just-registered client. The cache is exclusively for the by-`client_id` runtime resolution path.

**What CIMD will need from this cache, and why it already fits.** The cached value is a `oauthclient.Client` including its `Source` and `LastFetchedAt`, so nothing here is DCR-specific. Two obligations transfer to whoever implements CIMD, and both are consequences of CIMD rows being mutable where DCR rows are not:

- `Store.UpsertClient` (Part 2 §4.2) must invalidate the key from `DidCommitTx`, exactly as deletion does — otherwise a refetch that picks up new `redirect_uris` would not take effect for up to the TTL, and cimd.md is explicit that redirect URI validation reads "the client's current known state".
- cimd.md's 1-hour refetch interval is decided from `last_fetched_at`. Reading that from the cached row is fine and desirable — the 5-minute TTL is far shorter than the refetch interval, so a cache hit can never make Authgear think a stale document is fresh. Do **not** add a second cache keyed on fetch state.

Wiring: `ClientCache` joins `oauthclient.DependencySet` via `wire.Struct(new(ClientCache), "*")`. `*appredis.Handle` is already in every binary's provider set, as `pkg/lib/oauth/redis.Store` demonstrates.

## 4. `pkg/lib/oauthclient.Resolver` — the Actual Extension Point

Confirmed single choke point: `pkg/lib/oauthclient/resolver.go` is the concrete implementation bound (via `wire.Bind`, `pkg/lib/deps/deps_common.go:689-695`) to **six** separate `OAuthClientResolver`-shaped interfaces used across the codebase. This is the one place to change.

### 4.1 New signature

```go
// pkg/lib/oauthclient/resolver.go
//
// No DynamicClientQueries interface and no wire.Bind: Queries now lives in
// this same package (Part 2 §4), so Resolver just holds it. An earlier draft
// declared a cross-package interface because the store was going to live
// elsewhere.
type Resolver struct {
	OAuthConfig     *config.OAuthConfig
	TesterEndpoints tester.EndpointsProvider
	Queries         *Queries
}

func (r *Resolver) ResolveClient(ctx context.Context, clientID string) *config.OAuthClientConfig {
	if clientID == tester.ClientIDTester {
		return tester.NewTesterClient(r.TesterEndpoints.TesterURL().String())
	}
	if client, ok := r.OAuthConfig.GetClient(clientID); ok {
		return client
	}
	if !r.isDynamicClientIDCandidate(clientID) {
		return nil // fast path: never hits Redis or the DB for a static-shaped client_id
	}

	var defaults *config.OAuthDynamicClientRegistrationDefaultClientConfig
	if r.OAuthConfig.DynamicClientRegistration != nil {
		defaults = r.OAuthConfig.DynamicClientRegistration.DefaultClientConfig
	}

	// Redis-first, Postgres only on a miss, and the database scope is opened
	// inside Queries rather than here — see §3.1, §3.2 and §4.2.
	client, err := r.Queries.GetClientConfigByClientID(ctx, clientID, defaults)
	if err != nil {
		return nil
	}
	return client
}
```

The prefix guard is deliberate and does the work of the second half of the request: static-client resolution — the overwhelmingly common case — stays a zero-I/O in-memory map lookup exactly as today. Nothing downstream of that guard runs for a static client: **no Redis round-trip, no Postgres connection checkout, no `BEGIN`/`COMMIT`.** An unknown `client_id` that is not dynamic-shaped is likewise rejected in memory. The ladder is:

1. `client_id == tester.ClientIDTester` → synthesized tester client, in memory.
2. `OAuthConfig.GetClient(clientID)` hit → static client, in memory.
3. Not dynamic-shaped → `nil`, in memory.
4. Only now: Redis `GET`; and only on a cache miss, a Postgres `ReadOnly` scope.

**Express step 3 as a predicate, not an inline `strings.HasPrefix`,** because CIMD adds a second shape to it. cimd.md's resolution order is exactly this ladder — "Static and DCR clients are always checked first — a string is only ever treated as a CIMD candidate once neither matches" — so the extension point is step 3 and nowhere else:

```go
// pkg/lib/oauthclient/resolver.go
func (r *Resolver) isDynamicClientIDCandidate(clientID string) bool {
	// DCR: server-generated, always dcrc_-prefixed. IsDCRClientID and the
	// prefix both live in this package's client_id.go (Part 2 §4), so there
	// is no import of pkg/lib/dcr here.
	if IsDCRClientID(clientID) {
		return true
	}
	// CIMD adds: IsCIMDClientIDURL(clientID) && r.OAuthConfig.ClientIDMetadataDocument.IsEnabled()
	// (that predicate also lives in client_id.go — see Part 2 §4)
	// (https scheme, has a path, no fragment, no userinfo, no . or .. segments
	// — cimd.md Client ID Format, all checked before any network access).
	return false
}
```

Note the asymmetry to preserve: the DCR arm is unconditional, but the CIMD arm must be gated on CIMD being **enabled**, so that with CIMD off a URL-shaped `client_id` costs nothing. Do not "simplify" by dropping the config check into the store lookup.

The persisted-record lookup in step 4 is source-agnostic and needs no change for CIMD. What CIMD adds beyond it is the *fetch*, and per cimd.md that happens "**only** as a side effect of `/oauth2/authorize`" — never from this resolver, which is called from view models, middleware and the translation service where an outbound HTTP call would be indefensible. So CIMD's shape is: the authorize handler fetches-and-upserts, and `Resolver` keeps doing a plain read of whatever is persisted. Keeping the resolver read-only is what makes that possible; do not add fetching to it.

### 4.2 `ResolveClient` is called outside any database scope — someone must open one

**Without this, the first `/oauth2/authorize` request from a DCR client panics.** This was missing from the original draft of this plan.

`appdb.SQLExecutor`'s `ExecWith`/`QueryWith`/`QueryRowWith` all end in `mustGetTxLike(ctx)` (`pkg/lib/infra/db/sql_executor.go:71,111,151`), which calls `mustContextGetTxLike` → `panic(fmt.Errorf("programming_error: tx is not initialized"))` (`pkg/lib/infra/db/hook_handle.go:64-70`) when the context carries no transaction. A DB scope only exists inside `Handle.WithTx` / `Handle.ReadOnly` / `WithPrepareStatementsHandle`.

There is **no request-wide DB scope** to rely on. The session middleware opens `ReadOnly` only around its own `resolve()` and calls `next.ServeHTTP` after it returns (`pkg/lib/session/middleware.go:64`, and the `next.ServeHTTP` call outside it), so handlers start with an empty tx context.

At least two of the 15 call sites in §5.1 are provably outside a scope, and both are on the hottest DCR path:

- `pkg/lib/oauth/handler/resolve.go:26` (`resolveClient`) is called on the **very first line** of `AuthorizationHandler.ValidateRequestWithoutTx` (`handler_authz.go:883`) — the name says it: no transaction. Its caller is `AuthorizeHandler.ServeHTTP` (`pkg/auth/handler/oauth/authorize.go:46`), which passes `r.Context()` straight in. `doHandleRequestWithTx`'s `h.Database.WithTx` (`handler_authz.go:415`) only opens *after* validation has already resolved the client.
- `AuthorizationHandler.doHandleConsent` (`handler_authz.go:229`) similarly resolves the client via `prepareConsentRequest` with no `WithTx` anywhere in that path.

The rest (`viewmodels/base.go`, `webapp/redirect.go`, `client_like.go` ← userinfo, `translation/service.go`, …) are a mix and must not be audited one by one — the resolver has to be safe to call from anywhere.

**Fix, with existing precedent:** branch on `IsInTx` and open a `ReadOnly` scope when there is none, exactly as `usage.Limiter.dispatchEventImmediately` already does (`pkg/lib/usage/limit.go:394-404`):

```go
if l.Database.IsInTx(ctx) {
	return l.EventService.DispatchEventImmediately(ctx, payload)
}
return l.Database.ReadOnly(ctx, func(ctx context.Context) error { ... })
```

This is why `ResolveClient` needs `ctx` at all (§5) — not merely to pass it down to a query, but to inspect and possibly open a DB scope.

**The guard lives in `oauthclient.Queries.getClientByClientIDCached` (§3.1), not in `Resolver`.** Two reasons: it sits immediately next to the only code that touches Postgres, and it sits *after* the Redis lookup, so a cache hit opens no scope at all. `Resolver` therefore keeps no `Database` field and stays a thin dispatcher.

Two consequences to accept explicitly:

- **A cache-missing resolve outside a transaction takes a Postgres connection for the duration of one `SELECT`.** With §3.2's cache this is a cold-path cost, not a per-request one: an authorize request that resolves the same DCR client at validation, at consent and again in the view model does one Postgres read at most, and usually zero.
- **`ReadOnly` inside an already-open `WithTx` must be avoided,** which is what the `IsInTx` branch is for — nesting would attempt a second connection while the first transaction is open and can deadlock under pool pressure.

Add a resolver unit test for both branches: with a tx in context the read happens directly; without one, `ReadOnly` is opened (assert via a fake `Handle`, or by calling `ResolveClient` with a bare `context.Background()` and asserting it does not panic — that bare-context test is the actual regression guard).

### 4.3 Wiring

- `pkg/lib/oauthclient/deps.go` — extend the existing `wire.NewSet(...)` with `wire.Struct(new(Store), "*")`, `wire.Struct(new(ClientCache), "*")`, `wire.Struct(new(Queries), "*")` and `wire.Struct(new(Commands), "*")` alongside the existing `wire.Struct(new(Resolver), "*")`. **No `wire.Bind`** — `Resolver.Queries` is a concrete `*Queries` in the same package now, so there is no interface to bind. `Resolver` gains no `Database` field either: the `IsInTx`/`ReadOnly` guard lives in `Queries` (§4.2), whose `Database *appdb.Handle` and `Cache *ClientCache` are filled by `wire.Struct`, since `*appdb.Handle` and `*appredis.Handle` are already in every binary's provider set.
- Note `pkg/lib/oauthclient` gains Postgres and Redis dependencies where it previously had only `config` and `tester`. Every existing importer is a binary provider set or `pkg/auth/handler/webapp/tester.go`, all of which already have both, so no importer is newly burdened.
- No new binaries need `dcr.DependencySet` added individually — it was already added to `pkg/lib/deps/deps_common.go` in Part 1 (for the Admin API's own IAT/client facades) and that file is the shared wire set consumed by every binary that already binds `oauthclient.Resolver` (`pkg/admin`, `pkg/auth`, `cmd/authgear/background`, `pkg/redisqueue`, `pkg/resolver`, `e2e/cmd/e2e`) — confirmed these are exactly the binaries with generated `wire_gen.go` files referencing `oauthclient`. Run `make generate` — this regenerates **all** of their `wire_gen.go` files, not just `pkg/admin`'s; all must be committed together.

## 5. Mechanical Fan-Out — Every `OAuthClientResolver.ResolveClient` Caller Needs `ctx`

Adding `ctx` to `Resolver.ResolveClient` is a breaking change to **11 separate interface declarations** (all structurally identical, declared independently in each consuming package — there is no shared interface type) plus the concrete implementation. Confirmed exhaustive list (`grep -rn "ResolveClient(clientID string)"`, excluding the unrelated `sms.ClientResolver.ResolveClient()` in `pkg/lib/infra/sms` / `pkg/lib/messaging`, which is a same-named-but-unrelated interface for SMS provider clients — **do not touch it**):

| File | Line | Interface |
|---|---|---|
| `pkg/lib/oauthclient/resolver.go` | 13 | concrete `*Resolver` impl |
| `pkg/lib/oauth/client_like.go` | 21 | `OAuthClientResolver` |
| `pkg/lib/oauth/handler/resolve.go` | 22 | (local interface) |
| `pkg/lib/oauth/oidc/ui.go` | 91 | `UIInfoClientResolver`(-shaped) |
| `pkg/lib/interaction/context.go` | 216 | `OAuthClientResolver` |
| `pkg/lib/authenticationflow/service.go` | 65 | `OAuthClientResolver` |
| `pkg/admin/facade/oauth.go` | 46 | `OAuthClientResolver` |
| `pkg/auth/handler/webapp/authflow_controller.go` | 97 | (local interface) |
| `pkg/auth/handler/webapp/viewmodels/base.go` | 174 | (local interface) |
| `pkg/auth/handler/webapp/client.go` | 6 | `WebappOAuthClientResolver`(-shaped) |
| `pkg/auth/webapp/redirect.go` | 20 | (local interface) |
| `pkg/auth/handler/oauth/userinfo.go` | 29 | (local interface) |

That is 11 interface declarations plus the concrete `*Resolver`. Change each to `ResolveClient(ctx context.Context, clientID string) *config.OAuthClientConfig`.

Four test doubles also implement the same method and must be updated or regenerated (see §5.3): `pkg/auth/handler/webapp/authflow_controller_mock_test.go:391`, `pkg/lib/authenticationflow/service_mock_test.go:252`, `pkg/lib/oauth/handler/mock_test.go:104`, `pkg/lib/oauth/grant_offline_service_test.go:22`. The last two are hand-written, not generated — `go generate` will not fix them.

> **Line numbers in this file and Part 4 have drifted** — they were captured against an earlier tree and are now roughly 20-25 lines low in the larger files (e.g. `handler_token.go`'s `ResolveClient` call is at 1518, not 1494; `handleClientCredentials` starts at 2191, not 2167). The *structure* of every claim was re-verified and holds. Use the symbol names, not the line numbers, and rely on §5.2.

### 5.1 Known direct call sites (add `ctx` as first argument — all 15, `grep -rn "\.ResolveClient("`, excluding the two unrelated `SMSSender.ResolveClient()`/`ClientResolver.ResolveClient()` SMS-provider calls in `pkg/lib/messaging/sender.go:237` and `pkg/lib/infra/sms/sender.go:18`):

`pkg/admin/facade/oauth.go:70`, `pkg/auth/handler/webapp/authflow_controller.go:1012,1033`, `pkg/auth/webapp/redirect.go:34`, `pkg/auth/handler/webapp/viewmodels/base.go:216`, `pkg/auth/handler/webapp/logout.go:93`, `pkg/auth/handler/webapp/tester.go:193`, `pkg/lib/oauth/handler/resolve.go:26`, `pkg/lib/oauth/client_like.go:34`, `pkg/lib/oauth/grant_offline_service.go:122,327`, `pkg/lib/oauth/oidc/ui.go:151`, `pkg/lib/oauth/handler/handler_token.go:1518`, `pkg/lib/interaction/nodes/confirm_terminate_other_sessions_end.go:22`, `pkg/lib/authenticationflow/service.go:109`.

Most of these are inside functions that already take `ctx context.Context` — a mechanical one-line change. Two are **not**, and require their own signature change plus following their callers up the chain:

- **`oauth.SessionClientLike(s session.ResolvedSession, clientResolver OAuthClientResolver) *ClientLike`** (`pkg/lib/oauth/client_like.go:24`) → add `ctx context.Context` as the first parameter. Its only caller, `pkg/auth/handler/oauth/userinfo.go:42` (`oauth.SessionClientLike(s, h.OAuthClientResolver)`), is inside an HTTP handler `Handle(ctx context.Context, ...)` — trivial to thread through.
- **`OfflineGrantService.ComputeOfflineGrantExpiry(session *OfflineGrant) (expiry time.Time, err error)`** (`pkg/lib/oauth/grant_offline_service.go:121`) → add `ctx context.Context` as the first parameter. Its own interface duplicate at `pkg/lib/oauth/handler/service_token.go:78` (`ComputeOfflineGrantExpiry(session *oauth.OfflineGrant) (expiry time.Time, err error)`) needs the same change. Known call sites, all already inside `ctx`-bearing functions per a spot check of `authz_service.go:72`: `pkg/lib/oauth/authz_service.go:72`, `pkg/lib/oauth/grant_offline_service.go:91,138,200,252,291,351` (internal, same file), `pkg/lib/oauth/session_manager.go:87`, `pkg/lib/oauth/handler/service_token.go:183`.

### 5.2 Compiler-driven completeness check

The lists above are the result of static grep and are believed complete, but **this refactor must be finished by running `go build ./...` (and `go vet ./...`) repeatedly until clean**, not by treating §5.1 as exhaustive — Go's type system will surface any interface implementation or call site missed by grep (e.g. a test double, or a caller introduced between research and implementation). Any additional signature found this way follows the same rule: add `ctx context.Context` as the first parameter, thread it from the nearest enclosing function that already has one.

### 5.3 Generated mocks

`pkg/lib/oauth/handler` (and any other package under test that mocks `OAuthClientResolver`-shaped interfaces via `gomock`, confirmed present: `pkg/lib/oauth/handler/mock_test.go`, `handler_authz_mock_test.go`, `service_token_mock_test.go`, `handler_token_mock_test.go`) must be regenerated via `go generate ./...` (or the project's mock-generation target) in the same commit as the interface signature change — a stale mock would still compile against the old signature and silently not exercise the new `ctx` parameter.

## 6. Direct `OAuthConfig.GetClient` Bypass Sites

Four call sites read `config.OAuthConfig` directly instead of going through `oauthclient.Resolver`, so changing the Resolver alone does not make them DCR-aware. Assessed individually:

| File:line | What it does | Decision |
|---|---|---|
| `pkg/lib/oauth/oidc/handler/handler_end_session.go:101` | Iterates **all** `h.Config.Clients` to find any client whose `PostLogoutRedirectURIs` matches the request's `post_logout_redirect_uri` | **No change.** `PostLogoutRedirectURIs` is always empty for DCR clients (client.md: "Always empty for DCR clients") — a DCR client could never match this loop regardless of whether it's included, so leaving it static-only is correct, not an oversight. |
| `pkg/lib/resourcescope/commands.go:67,78,143,180,217` | Existence-check gate before `AddResourceToClientID`/`AddScopesToClientID`/etc. allow an explicit Client-Resource Association | **No change.** Per api-resource.md, explicit Client-Resource Association is the mechanism for admin-managed (in practice, M2M) clients; DCR/third-party clients gain Resource access exclusively through `access_policy` (Part 4), never through per-client association. Keeping this static-only is a deliberate scope boundary, not a gap — flagging it here so it isn't mistaken for one later. |
| `pkg/lib/authenticationflow/declarative/intent_login_flow_step_terminate_other_sessions.go:41` | Reads `client.MaxConcurrentSession == 1` to decide whether to force-terminate other sessions at login | **No change.** `MaxConcurrentSession` is always `0` for DCR clients (client.md: "Always 0 for DCR clients"), so `ok && client.MaxConcurrentSession == 1` evaluates identically (`false`) whether or not the DCR client is found — there is no behavioral difference to fix. |
| `pkg/lib/translation/service.go:281` | Resolves `client.Name` for use as a template variable (e.g. notification emails/SMS referencing the requesting client's display name) | **Fix required.** A DCR client's display name should appear in outgoing notification templates the same way a static client's does. Change `s.OAuthConfig.GetClient(uiParams.ClientID)` to go through the resolver: add an `OAuthClientResolver` dependency to `translation.Service` (interface `ResolveClient(ctx context.Context, clientID string) *config.OAuthClientConfig`, bound to `*oauthclient.Resolver` the same way every other consumer in §5 is), call `s.OAuthClientResolver.ResolveClient(ctx, uiParams.ClientID)`, use `client.Name` from the result (`ctx` is already available — `prepareTemplateVariables(ctx context.Context, ...)`). |

## 7. Test Plan

Unit tests (Convey):

- `pkg/lib/config/oauth_test.go` (extend if it exists, else new) — table-driven test asserting `OAuthClientApplicationTypeDynamicThirdParty.IsThirdParty()==true`, `.IsConfidential()==false`, `.IsClientCredentialsFlowAllowed()==false`, `.HasFullAccessScope()==false`, `.PIIAllowedInIDToken()==true`.
- `pkg/lib/oauthclient/client_config_test.go` — `Client.ToClientConfig`: first-party `web` → `OAuthClientApplicationTypeSPA`; first-party `native` → `OAuthClientApplicationTypeNative`; third-party (either) → `OAuthClientApplicationTypeDynamicThirdParty`; lifetime fields populated from `defaults` when non-nil, fall back to `SetDefaults()`'s project defaults when `defaults` is nil.
- `pkg/lib/oauthclient/resolver_test.go` (extend) — `ResolveClient` with a `dcrc_...`-prefixed ID not in static config reaches `Queries.GetClientConfigByClientID` and returns its result; a non-`dcrc_`-prefixed unknown ID returns `nil` **without** reaching it at all. With `Queries` now concrete and in-package, assert the fast path by faking at the `Store`/`ClientCache` level (`gomock.Times(0)` on both) rather than by mocking a `Queries` interface.
- `pkg/lib/oauthclient/queries_test.go` (new) — `getClientByClientIDCached`: cache hit returns the cached client and never calls `Store` **nor opens a DB scope** (`gomock.Times(0)` on both); cache miss reads through and calls `Cache.Set`; a not-found miss calls `Cache.SetNotFound` and returns `ErrDynamicClientNotFound`; a cached negative returns `ErrDynamicClientNotFound` without touching `Store`; and a `Cache.Get` **error** falls through to the database rather than failing the request. Plus the §4.2 regression guard — a cache-missing resolve with a bare `context.Background()` (no tx in context) opens a `ReadOnly` scope and does not panic.
- `pkg/lib/oauthclient/cache_client_test.go` (new) — round-trip `Set`/`Get`, `SetNotFound`/`Get` distinguishing "cached negative" from "not cached", `Delete`, and that the marshalled payload is the `Client` row rather than a resolved config (assert a lifetime field is absent from the JSON).

e2e tests (`e2e/tests/dcr_client_resolution.yaml`):

1. Register a `THIRD_PARTY` DCR client (via `POST /oauth2/register` from Part 2), then run a full authorization-code flow (`/oauth2/authorize` → consent → `/oauth2/token`) using its `client_id` — must succeed, consent screen must show, and (without `resource=`, per Part 4) the issued access token must be usable at `/oauth2/userinfo`.
2. Register a `FIRST_PARTY` DCR client (`iat_fp_...`), run the same flow — consent screen must be **skipped**.
3. `/oauth2/revoke` and `/oauth2/logout` with a DCR client's tokens/session — must succeed (exercises `pkg/lib/oauth/handler/resolve.go` and `handler_end_session.go` paths).
4. A refresh-token grant using a DCR client's refresh token — must succeed (exercises `grant_offline_service.go`, `client_like.go`).
5. An email/SMS notification triggered during a DCR client's auth flow (e.g. OTP) includes the DCR client's `client_name` in the rendered message (exercises the `translation/service.go` fix in §6).
6. Cache invalidation end to end (§3.2): register a DCR client, complete an authorization so the client is definitely cached, `deleteDynamicClient`, then immediately attempt `/oauth2/authorize` with the same `client_id` → must fail with `unauthorized_client`, **not** succeed from a stale cache entry. This is the test that a broken `DidCommitTx` invalidation would fail, and it must not be allowed to pass merely because the TTL happened to expire — run it back-to-back with no delay.

## 8. Fixed Behavioral Decisions

- Static-client resolution remains zero-I/O: tester client, static config hit, and non-`dcrc_` miss all return without touching Redis or Postgres, and without opening a transaction (§4.1).
- DCR client resolution is **Redis-first** with a 5-minute TTL, Postgres only on a miss (§3.2). Negative results are cached for 30 seconds, which is safe because `client_id`s are server-generated 16-byte random values.
- The cache holds the `oauthclient.Client` row, never the synthesized `*config.OAuthClientConfig`, so an edit to `default_client_config` in `authgear.yaml` takes effect immediately instead of after the TTL (§3.2).
- Cache invalidation on `deleteDynamicClient` runs from `DidCommitTx`, with the TTL as the backstop (§3.2). `CountClients` and the Admin API client list deliberately bypass the cache.
- The `IsInTx`/`ReadOnly` guard lives in `oauthclient.Queries`, after the cache lookup (§3.1, §4.2), rather than requiring all 15 `ResolveClient` call sites to be inside a transaction. Auditing those call sites individually was rejected: the resolver is called from middleware, view models and handlers alike, and any future call site would silently reintroduce the panic.
- `deleteDynamicClient` (Part 2) still does not revoke outstanding tokens/authorizations for the deleted client — that requires the resolver changes in this part to exist first, but the actual revocation call is not added here either. This remains an open gap after Part 3; flag to the user before considering the DCR project "done" even after Part 4.

## 9. Atomic Commit Plan

Because an interface signature change must compile as a whole, this is necessarily fewer, larger commits than Parts 1–2:

1. **`[DCR] Add synthetic third-party-public OAuthClientApplicationType`** — §2 only (`pkg/lib/config/oauth.go` new `x_dynamic_third_party` constant + 5 switch-statement cases) + unit test. Self-contained, no callers yet, safe to land alone.
2. **`[DCR] Add Redis cache for dynamic client resolution`** — §3.2 (`cache_client.go`, key helper, `ClientCache` in `dcr.DependencySet`) + `cache_client_test.go`. No callers yet.
3. **`[DCR] Add GetClientConfigByClientID to oauthclient.Queries`** — §3.1 (`getClientByClientIDCached` with the cache and the `IsInTx`/`ReadOnly` guard) + `queries_test.go`. `ToClientConfig` itself (§3) lands in Part 2 commit 4, since Part 2's `ToModel` depends on it; if for any reason it did not, add `client_config.go` here too. Still no runtime callers either way.
4. **`[DCR] Invalidate the DCR client cache on deletion`** — the `DidCommitTx` invalidation in `Commands.DeleteClient` (§3.2). Separate from Part 2's `deleteDynamicClient` commit because the cache does not exist until commit 2 here; if Part 3 lands before anyone can delete a client in anger, fold it into commit 2 instead.
5. **`[DCR] Thread ctx through OAuthClientResolver.ResolveClient`** — all of §5 (11 interface signatures, all known call sites, `ComputeOfflineGrantExpiry`/`SessionClientLike` cascades, regenerated mocks) in one commit, verified via `go build ./... && go vet ./...` clean. Purely mechanical — no behavior change yet (the concrete `Resolver.ResolveClient` body is untouched in this commit, only its signature).
6. **`[DCR] Resolve DCR clients from the database in oauthclient.Resolver`** — §4.1/§4.3 (`Resolver` body change, `isDynamicClientIDCandidate`, `deps.go` provider additions) + regenerated `wire_gen.go` for every affected binary + resolver unit tests.
7. **`[DCR] Resolve DCR client name in translation service`** — §6's `translation/service.go` fix, isolated since it's the one bypass site that actually changes.
8. **`[DCR] Add e2e tests for DCR client authorization flows`** — `e2e/tests/dcr_client_resolution.yaml`.
