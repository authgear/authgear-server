# DCR Part 2 — Client Registration, Storage, Query & Deletion

Spec: [docs/specs/dcr.md](../../specs/dcr.md), [docs/specs/client.md](../../specs/client.md).

Depends on Part 1 (`pkg/lib/dcr` package, `dcr.Queries.ValidateAndGetByToken`).

> **Table and package naming.** The persisted client record is `_auth_oauth_client`, owned by the existing `pkg/lib/oauthclient` package (which today holds only the `Resolver`), so CIMD-resolved clients reuse both rather than needing a second table and store. See §3.1 for why the specs require this and §4 for the package split — including `DCRClientIDPrefix` moving out of `pkg/lib/dcr`.

## 1. Goal / Scope

In scope:

- `oauth.dynamic_client_registration.{enabled,initial_access_token_required,default_client_config}` in `authgear.yaml`.
- `_auth_oauth_client` DB table + migration (source-agnostic, so CIMD reuses it — §3.1).
- `pkg/lib/oauthclient` (existing package, currently resolver-only): source-agnostic `Client` record, store/commands/queries/cache, and the `client_id` shape helpers moved from `dcr` (§4).
- `pkg/lib/dcr`: DCR registration request validation (redirect URIs, grant/response type consistency, application_type, https-only URI fields) and `Commands.RegisterClient`.
- `POST /oauth2/register` HTTP endpoint.
- Unified `OAuthClient` GraphQL type (per client.md), `dynamicClients` query, `deleteDynamicClient` mutation.

Out of scope (later parts / explicitly deferred):

- Making `/oauth2/authorize`, `/oauth2/token`, etc. actually resolve a `dcrc_...` client_id at runtime (Part 3). After this part, a client can be registered and inspected via Admin API, but **cannot yet complete an authorization flow** — `oauthclient.Resolver` still only knows about `authgear.yaml` clients.
- `access_policy` / resource-indicator support (Part 4).
- The DCR client limit — deferred to Part 5. **Superseded design note:** at the time this file was written, dcr.md still specified this as a standalone `oauth.dynamic_client_registration.maximum_clients` config key; as of the `spec` branch commit "Fold DCR/CIMD client limits into the usage limit config," it is instead `usage.limits.oauth_client_dcr` (a new "standing" usage name, see docs/specs/usage.md). Part 5 implements it against the current spec, not the superseded key named here.
- Portal UI screens for any of this (Admin GraphQL API only, per user's framing of this as a backend implementation plan).

## 2. Config Model & Schema

### 2.1 `pkg/lib/config/oauth.go` — extend `OAuthConfig`

Add a field and schema property (existing schema block at `pkg/lib/config/oauth.go:3-14`):

```go
var _ = Schema.Add("OAuthConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"clients": {
			"type": "array",
			"items": { "$ref": "#/$defs/OAuthClientConfig" }
		},
		"dynamic_client_registration": { "$ref": "#/$defs/OAuthDynamicClientRegistrationConfig" }
	}
}
`)

type OAuthConfig struct {
	Clients                   []OAuthClientConfig                   `json:"clients,omitempty"`
	DynamicClientRegistration *OAuthDynamicClientRegistrationConfig `json:"dynamic_client_registration,omitempty"`
}
```

No `nullable:"true"` here — see §2.2 for why. `OAuthConfig.DynamicClientRegistration` is always non-nil once `config.SetFieldDefaults` has run.

### 2.2 New file `pkg/lib/config/oauth_dynamic_client_registration.go`

```go
package config

var _ = Schema.Add("OAuthDynamicClientRegistrationConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"initial_access_token_required": { "type": "boolean" },
		"default_client_config": { "$ref": "#/$defs/OAuthDynamicClientRegistrationDefaultClientConfig" }
	}
}
`)

type OAuthDynamicClientRegistrationConfig struct {
	Enabled                    bool                                               `json:"enabled,omitempty"`
	InitialAccessTokenRequired *bool                                              `json:"initial_access_token_required,omitempty"`
	DefaultClientConfig        *OAuthDynamicClientRegistrationDefaultClientConfig `json:"default_client_config,omitempty"`
}

// No nullable tag on this type or its parent field (see §2.2), so
// config.SetFieldDefaults force-allocates it whenever the section is absent
// and recurses into DefaultClientConfig, which gets its own SetDefaults()
// below. IsEnabled/IsInitialAccessTokenRequired stay nil-safe only for code
// paths that read a config before SetFieldDefaults has run; at runtime the
// receiver is always non-nil.
func (c *OAuthDynamicClientRegistrationConfig) IsInitialAccessTokenRequired() bool {
	return c == nil || c.InitialAccessTokenRequired == nil || *c.InitialAccessTokenRequired
}

func (c *OAuthDynamicClientRegistrationConfig) IsEnabled() bool {
	return c != nil && c.Enabled
}

var _ = Schema.Add("OAuthDynamicClientRegistrationDefaultClientConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"access_token_lifetime_seconds": { "type": "integer", "minimum": 1 },
		"refresh_token_lifetime_seconds": { "type": "integer", "minimum": 1 },
		"refresh_token_idle_timeout_enabled": { "type": "boolean" },
		"refresh_token_idle_timeout_seconds": { "type": "integer", "minimum": 1 }
	}
}
`)

// Field names/types/json tags match the corresponding subset of
// OAuthClientConfig (pkg/lib/config/oauth.go:263-266) exactly, since these
// values are later copied verbatim onto the synthetic OAuthClientConfig
// built for a resolved DCR client (Part 3).
type OAuthDynamicClientRegistrationDefaultClientConfig struct {
	AccessTokenLifetime            DurationSeconds `json:"access_token_lifetime_seconds,omitempty"`
	RefreshTokenLifetime           DurationSeconds `json:"refresh_token_lifetime_seconds,omitempty"`
	RefreshTokenIdleTimeoutEnabled *bool           `json:"refresh_token_idle_timeout_enabled,omitempty"`
	RefreshTokenIdleTimeout        DurationSeconds `json:"refresh_token_idle_timeout_seconds,omitempty"`
}

// SetDefaults mirrors OAuthClientConfig.SetDefaults()'s token-lifetime
// fallback exactly (same constants, same zero-means-unset per-field
// semantics), so a DCR client with no configured override resolves to the
// identical values a static client that omits every lifetime field would
// get. Not a copy of OAuthClientConfig.SetDefaults() itself: that method
// also has an ApplicationType-driven branch (IssueJWTAccessToken) that does
// not apply here.
func (c *OAuthDynamicClientRegistrationDefaultClientConfig) SetDefaults() {
	if c.AccessTokenLifetime == 0 {
		c.AccessTokenLifetime = DefaultAccessTokenLifetime
	}
	if c.RefreshTokenLifetime == 0 {
		c.RefreshTokenLifetime = max(c.AccessTokenLifetime, DefaultRefreshTokenLifetime)
	}
	if c.RefreshTokenIdleTimeoutEnabled == nil {
		b := DefaultRefreshTokenIdleTimeoutEnabled
		c.RefreshTokenIdleTimeoutEnabled = &b
	}
	if c.RefreshTokenIdleTimeout == 0 {
		c.RefreshTokenIdleTimeout = DefaultRefreshTokenIdleTimeout
	}
}
```

**No `nullable:"true"` anywhere in this section — both levels are always non-nil once `config.SetFieldDefaults` has run.**

`config.SetFieldDefaults` (`pkg/lib/config/default.go`) is a reflection walker. For a nil pointer-to-struct field it **allocates a zero value** before descending (`default.go:33-37`: `ele = reflect.New(t.Elem()); v.Set(ele)`) and then calls `SetDefaults()` on it if the type implements `defaulter` — *unless* the field is tagged `nullable:"true"`, in which case it skips the field entirely (`default.go:27-31`). Neither `DynamicClientRegistration` nor `DefaultClientConfig` carries that tag, so both get force-allocated and `DefaultClientConfig.SetDefaults()` (above) always runs.

An earlier draft of this plan tagged the *parent* field `nullable:"true"` on the theory that "nil means absent" needed to be preserved at every level so `defaults != nil` in §4.1/Part 3 §3 could mean "the admin configured an override." That reasoning doesn't survive scrutiny:

- `IsEnabled()`/`IsInitialAccessTokenRequired()` are already nil-safe and already read every "unset" state correctly regardless of whether the parent is nil or a zero-valued struct — the nullable tag bought nothing they didn't already have.
- `defaults != nil` in §4.1's `ToClientConfig` was never really about "did the admin configure an override" in the first place — it's `ResolveTokenLifetimes`'s way of saying "this *source* (DCR) has a default-client-config concept at all," as opposed to a source that doesn't (e.g. a future CIMD before it implements one). That check is unaffected by this section being non-nil.
- The "portal round-trip" concern doesn't apply: the Portal's config editor never round-trips a `SetFieldDefaults`-processed `AppConfig` — it loads and saves the *raw*, undefaulted YAML bytes directly (`LoadRawAppConfig` / the `updateApp` mutation), a separate code path from `Parse()`'s defaulted struct.

So: no accessor is needed for `DefaultClientConfig` at all. Once defaults have run, `GetDefaultClientConfig()` (or just reading `.DefaultClientConfig` directly) always returns a fully-populated struct with real token-lifetime values — using it to detect "was an override configured" was already fragile (it only worked because the zero value happened to be all-zero and `OAuthClientConfig.SetDefaults()` happened to treat `0` as unset elsewhere); giving `DefaultClientConfig` its own `SetDefaults()` that reuses the *exact same* constants (`DefaultAccessTokenLifetime` et al.) removes that fragility rather than working around it.

`IsInitialAccessTokenRequired()`/`IsEnabled()` stay nil-safe purely as defense for code paths that read a config before `SetFieldDefaults` has run (e.g. a test that `yaml.Unmarshal`s a snippet directly) — not because the parent is expected to be nil at runtime.

Add a config unit test for all three states: section absent → `IsEnabled()` false, `IsInitialAccessTokenRequired()` true, and `DefaultClientConfig` non-nil with the four built-in fallback values (§2.2.1); section present but empty → same; section present with `initial_access_token_required: false` → false.

Register the new file's schema fragments in whatever central schema-registration list `oauth.go`'s `Schema.Add` calls already rely on (they self-register via `var _ = Schema.Add(...)` at package init, same as every other config file — no separate registration step needed).

### 2.2.1 Resolved token lifetimes when `default_client_config` is absent

client.md says only "Token lifetime fields are populated from `default_client_config` when set, otherwise from the project defaults." That phrase is misleading: **there is no project-level token lifetime configuration in `authgear.yaml`** — token lifetimes exist only per client. The real source is the built-in constants applied by `OAuthClientConfig.SetDefaults()` (`pkg/lib/config/oauth.go:376-398`), i.e. exactly what a static client that omits every lifetime field gets:

| Field | Value with `default_client_config` absent/null | Constant |
|---|---|---|
| `access_token_lifetime_seconds` | `1800` (30 minutes) | `DefaultAccessTokenLifetime` (`config/session.go:35`) |
| `refresh_token_lifetime_seconds` | `31449600` (52 weeks / 364 days) | `max(access, DefaultRefreshTokenLifetime)`, where `DefaultRefreshTokenLifetime = DefaultIDPSessionLifetime = 52*7*86400` (`config/session.go:27,19`) |
| `refresh_token_idle_timeout_enabled` | `true` | `DefaultRefreshTokenIdleTimeoutEnabled` (`config/session.go:31`) |
| `refresh_token_idle_timeout_seconds` | `2592000` (30 days) | `DefaultRefreshTokenIdleTimeout` (`config/session.go:29`) |

Note these are **not** the numbers in client.md's DCR example JSON (`1800` / `2592000` / `true` / `1209600`) — those come from dcr.md's Configuration snippet, which sets `default_client_config` explicitly. The example is self-consistent, but a reader can easily mistake it for the default. The doc-fix commit (§9) replaces "otherwise from the project defaults" with a pointer to these four values.

**Partial overrides interact with `SetDefaults()`, and one combination is unsafe.** Each field falls back independently, so setting only `access_token_lifetime_seconds: 7200` yields `refresh_token_lifetime_seconds = max(7200, 31449600) = 31449600`. But setting **both** with refresh below access — e.g. `access: 7200, refresh: 3600` — is *not* corrected: `SetDefaults()` only applies the `max()` when `RefreshTokenLifetime == 0`. Static clients are protected from this by `AppConfig.validateTokenLifetime` (`pkg/lib/config/config.go:154-162`), which iterates `c.OAuth.Clients` **only** — so `default_client_config` bypasses it entirely and every DCR client in the project would get a refresh token shorter than its access token.

Extend that validator (it is the natural home, and it already produces the right error message):

```go
func (c *AppConfig) validateTokenLifetime(ctx *validation.Context) {
	for i, client := range c.OAuth.Clients {
		// ... existing loop, unchanged ...
	}

	if dcr := c.OAuth.DynamicClientRegistration; dcr != nil {
		if d := dcr.DefaultClientConfig; d != nil {
			if d.RefreshTokenLifetime != 0 && d.AccessTokenLifetime != 0 &&
				d.RefreshTokenLifetime < d.AccessTokenLifetime {
				ctx.Child("oauth", "dynamic_client_registration", "default_client_config", "refresh_token_lifetime_seconds").
					EmitErrorMessage("refresh token lifetime must be greater than or equal to access token lifetime")
			}
		}
	}
}
```

The `!= 0` guards matter: a zero field means "unset", and the `max()` in `SetDefaults()` already makes those cases safe. Add a config validation test for all three shapes (both set and valid, both set and inverted → error, only access set → no error).

**Timing note (see §2.2's revised design):** `config.Parse` runs `PopulateDefaultValues` (which now includes `OAuthDynamicClientRegistrationDefaultClientConfig.SetDefaults()`) *before* `validateTokenLifetime` above. So for the "only access set" case, `RefreshTokenLifetime` is no longer `0` by the time this validator runs — it has already been filled to `max(access, DefaultRefreshTokenLifetime)`, which by construction is always `>= access`. The `!= 0` guards still do no harm (both fields are non-zero for every DCR-enabled config now, always-true rather than sometimes-true), but the actual safety net for the "only one field set" case is the `max()` in `SetDefaults()`, not the guard — exactly as it already was for static clients. The guard remains load-bearing only for the "both explicitly set, inverted" case, where neither field is ever `0` and `SetDefaults()` leaves both untouched.

### 2.2.2 `OAuthDynamicClientRegistrationRateLimitsConfig` — project-configurable registration rate limits

A new field on `OAuthDynamicClientRegistrationConfig` (§2.2), sibling to `DefaultClientConfig`:

```go
type OAuthDynamicClientRegistrationConfig struct {
	// ... existing fields ...
	RateLimits *OAuthDynamicClientRegistrationRateLimitsConfig `json:"rate_limits,omitempty"`
}

var _ = Schema.Add("OAuthDynamicClientRegistrationRateLimitsConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"per_ip": { "$ref": "#/$defs/RateLimitConfig" },
		"per_project": { "$ref": "#/$defs/RateLimitConfig" }
	}
}
`)

type OAuthDynamicClientRegistrationRateLimitsConfig struct {
	PerIP      *RateLimitConfig `json:"per_ip,omitempty"`
	PerProject *RateLimitConfig `json:"per_project,omitempty"`
}
```

`SetDefaults()` follows `AuthenticationRateLimitsSignupConfig.SetDefaults()`'s exact pattern (`pkg/lib/config/authentication_rate_limits.go`) — by the time this runs, `config.SetFieldDefaults`'s walker has already force-allocated `PerIP`/`PerProject` (neither field carries a `nullable` tag), so checking `Enabled == nil` safely detects "the project never configured this bucket" and replaces the whole zero-valued struct with the built-in default:

```go
func (c *OAuthDynamicClientRegistrationRateLimitsConfig) SetDefaults() {
	if c.PerIP.Enabled == nil {
		c.PerIP = &RateLimitConfig{Enabled: new(true), Period: "1m", Burst: 10}
	}
	if c.PerProject.Enabled == nil {
		c.PerProject = &RateLimitConfig{Enabled: new(true), Period: "1h", Burst: 1000}
	}
}
```

A `GetRateLimits()` nil-safe accessor on `OAuthDynamicClientRegistrationConfig` mirrors `GetDefaultClientConfig()` — nil-safe only for the pre-`SetFieldDefaults` case; at runtime `RateLimits` (and its two fields) are always non-nil.

`pkg/lib/oauth/handler/ratelimit.go`'s `NewBucketSpecOAuthRegisterPerIP`/`NewBucketSpecOAuthRegisterPerProject` (§5.1.1) take this resolved config directly instead of constructing a `config.RateLimitConfig` literal inline — the hard-coded defaults live in `SetDefaults()` above now, not duplicated in the handler package.

Add a config test mirroring `TestOAuthDynamicClientRegistrationConfig` (§7): section absent → `RateLimits.PerIP`/`.PerProject` resolve to the built-in defaults (10/minute, 1000/hour); section present with only `per_project` set → `PerIP` still defaults, `PerProject` uses the configured value.

### 2.3 JSON Schema export

Run `make export-schemas` to regenerate the checked-in `authgear.yaml` JSON Schema artifact — commit the regenerated file alongside the Go changes.

## 3. Data Model & Migration

### 3.1 Table design — one table for all dynamic clients, not a DCR-specific one

dcr.md gives no explicit `CREATE TABLE` for the client itself (only for the IAT table), so the shape is designed here, following the `_auth_resource` migration's conventions (`cmd/authgear/cmd/cmddatabase/migrations/authgear/20250710143552-add_resource_scope.sql`).

**The table is `_auth_oauth_client`, keyed by `source` — not a DCR-only `_auth_oauth_dcr_client`, and not `_auth_oauth_dynamic_client`.** Three independent parts of the specs already require CIMD-resolved clients to live in the same place:

- client.md defines **one** `OAuthClient` type with `source: STATIC | DCR | CIMD`, and dcr.md's `dynamicClients` query returns "DCR-registered clients and, when enabled, CIMD-resolved clients … Distinguish the two via `source`" — a single paginated, `created_at`-ordered Connection over both. Two tables would mean either a UNION query or a merge-and-re-sort in Go, and relay cursor pagination over a merged result set is genuinely painful.
- `deleteDynamicClient` takes one `clientID: String!` and deletes "a DCR-registered or CIMD-resolved client" — one lookup, one table.
- cimd.md's persisted record is described in exactly the same terms as a DCR client's: "a **persisted record keyed by `client_id`**, shared across every user and every grant", read by `/oauth2/token` and the Authorized Apps page as "a plain lookup, never a live fetch". That is the same read path Part 3's resolver builds.

Doing this now costs one extra column and two lines of index; doing it later costs a data migration plus a rewrite of the store, the cache, the resolver and the Admin API.

**Why `_auth_oauth_client` and not `_auth_oauth_dynamic_client`.** Three reasons:

1. **It matches the repo's naming convention,** which names tables after the entity, not after a subset of it — `_auth_resource`, `_auth_user`, `_auth_oauth_authorization`. Where a qualifier appears it denotes a *subtype of a parent table* (`_auth_identity` → `_auth_identity_oauth`, `_auth_authenticator` → `_auth_authenticator_oob`), which is not the relationship here. It also matches the model name, `model.OAuthClient`.
2. **"Dynamic" is ambiguous in the specs as they stand.** client.md's source list calls item 2 "**Dynamic clients** — registered at runtime via DCR", i.e. DCR *only*, with CIMD as a separate item 3 — while dcr.md's `dynamicClients` query returns both. Baking a word the specs use in two conflicting senses into a table name is asking for confusion. (The client.md wording is a doc bug regardless; see §9.)
3. **`source` already carries the distinction,** so encoding it in the table name is the same redundancy that got `registered_at` dropped below. And if static clients are ever moved into the database, `_auth_oauth_client` is already the right name, whereas `_auth_oauth_dynamic_client` would need a rename.

The one genuine cost: today the table holds only non-static clients, so `SELECT * FROM _auth_oauth_client` silently omits every client declared in `authgear.yaml`. Mitigate with the column comment in the DDL below rather than with the table name — a support engineer reading the schema needs the "static clients live in authgear.yaml" fact spelled out either way, and a name can't convey it.

The table name, the owning package (`pkg/lib/oauthclient`, §4) and the API model (`model.OAuthClient`) all agree, which is the main reason to prefer this name over a qualified one.

New file: `cmd/authgear/cmd/cmddatabase/migrations/authgear/20260817120001-add_oauth_dynamic_client.sql`

```sql
-- +migrate Up

CREATE TABLE _auth_oauth_client (
  id text PRIMARY KEY,
  app_id text NOT NULL,
  client_id text NOT NULL,
  -- 'DCR' | 'CIMD'. Matches model.OAuthClientSource verbatim.
  --
  -- NOTE: this table does NOT contain every OAuth client of a project.
  -- Statically configured clients (model.OAuthClientSource 'STATIC') live in
  -- authgear.yaml under oauth.clients and never appear here, so 'STATIC' is
  -- currently not a value this column takes. Any query intended to cover all
  -- clients must also read the project's authgear.yaml.
  source text NOT NULL,
  created_at timestamp without time zone NOT NULL,
  updated_at timestamp without time zone NOT NULL,
  -- CIMD only: timestamp of the most recent successful metadata fetch.
  -- Always NULL for source = 'DCR'.
  last_fetched_at timestamp without time zone,
  kind text NOT NULL,
  application_type text NOT NULL,
  client_name text,
  client_uri text,
  logo_uri text,
  tos_uri text,
  policy_uri text,
  redirect_uris jsonb NOT NULL,
  grant_types jsonb NOT NULL,
  response_types jsonb NOT NULL
);
CREATE UNIQUE INDEX _auth_oauth_client_client_id_unique ON _auth_oauth_client USING btree (app_id, client_id);
CREATE INDEX _auth_oauth_client_app_id_created_at ON _auth_oauth_client USING btree (app_id, created_at);
CREATE INDEX _auth_oauth_client_app_id_source ON _auth_oauth_client USING btree (app_id, source);

-- +migrate Down

DROP INDEX IF EXISTS _auth_oauth_client_app_id_source;
DROP INDEX IF EXISTS _auth_oauth_client_app_id_created_at;
DROP INDEX IF EXISTS _auth_oauth_client_client_id_unique;
DROP TABLE IF EXISTS _auth_oauth_client;
```

Notes:

- `id` (uuid) is the internal row identity used for the GraphQL `Node` id; `client_id` is the externally-used OAuth `client_id` — `dcrc_...` for DCR, an `https://` URL for CIMD — looked up by the runtime resolver in Part 3, hence the separate unique index on `(app_id, client_id)`.
- The unique index spans **all** sources deliberately: `client_id` is one namespace, and for CIMD it is what enforces cimd.md's "a single shared record per `client_id`". A collision between the two shapes is impossible anyway (cimd.md's Client ID Format section notes DCR ids can never parse as `https://.../path`), so a single index is strictly better than two source-scoped ones.
- `_auth_oauth_client_app_id_source` exists for `COUNT(*) WHERE source = ?`. This is not speculative: dcr.md's client limit counts "`OAuthClient` records with `source: DCR`" and cimd.md's counts `source: CIMD`, against separate quotas — so Part 5's quota check **must** filter by source (see Part 5 §5). The `(app_id, created_at)` index stays for the unfiltered, `created_at`-ordered `dynamicClients` page.
- `updated_at` is required by CIMD even though DCR never uses it: a CIMD refetch "overwrites that same record in place". For DCR rows it is written equal to `created_at` and never changes (DCR clients are immutable until RFC 7592).
- **No `registered_at` column.** client.md's `registeredAt` is `created_at` for DCR and `null` for CIMD ("there is no registration event, only a fetch"), so it is derived from `source` + `created_at` rather than stored redundantly. Likewise `lastFetchedAt` is `last_fetched_at`, which is null for DCR.
- No `client_secret*` columns — both sources are always public (dcr.md: "Confidential clients are not supported via DCR"; cimd.md: "CIMD clients in v1 are always **public**").
- No `token_endpoint_auth_method` column — DCR rejects the field outright and CIMD requires it absent or `none`, so there is nothing per-client to record.
- No token-lifetime columns. They come from project config at resolution time, and note the two sources read **different** config keys: `oauth.dynamic_client_registration.default_client_config` for DCR, `oauth.client_id_metadata_document.client_config` for CIMD. This is why §4.1's `ToClientConfig` takes the resolved defaults as a parameter rather than reading config itself — the same row must be resolvable against whichever key its `source` dictates.
- `redirect_uris`, `grant_types`, `response_types` stored as `jsonb` arrays (`jsonb` is the natural, low-risk choice when there's no need to query into individual elements, and keeps the row schema stable if a future array field is added).

## 4. Package split: `pkg/lib/oauthclient` and `pkg/lib/dcr`

Since the table is source-agnostic (§3.1), the Go code that owns it should be too. A `_auth_oauth_client` table read and written from a package called `dcr` would force CIMD to either import `dcr` (wrong dependency direction — CIMD has nothing to do with the registration protocol) or duplicate the store.

**The persisted record goes into the existing `pkg/lib/oauthclient`**, alongside the `Resolver` that is its main consumer. That package currently holds only `resolver.go` + `deps.go` and imports nothing but `config` and `tester`, so it is a natural home rather than a crowded one. The result is one package answering "what is an OAuth client and how do I get one", with a name matching both the table (`_auth_oauth_client`) and the model (`model.OAuthClient`):

| File | Contents |
|---|---|
| `resolver.go` | `Resolver` (existing; Part 3 §4) |
| `client_id.go` | `DCRClientIDPrefix`, `IsDCRClientID()`, `GenerateDCRClientID()`, and later CIMD's URL-shape predicate |
| `client.go` | `Client` record type, `Source`/`Kind` aliases, `DisplayName()`, `RegisteredAt()`, `ToClientConfig()`, `ToModel()` |
| `store_client.go` | `Store` + all SQL against `_auth_oauth_client` |
| `cache_client.go` | Redis cache (Part 3 §3.2) |
| `queries.go` / `commands.go` | `GetClientConfigByClientID`, `ListClients`, `CountClientsBySource`, `DeleteClient` |
| `errors.go` | `ErrDynamicClientNotFound`, `ErrDynamicClientDuplicateClientID` |

**`pkg/lib/dcr` keeps what is genuinely DCR-specific:** the IAT domain from Part 1 (`initial_access_token.go`, `token.go`, `store_initial_access_token.go`), `validate.go` (§4.3 — DCR's metadata rules differ from CIMD's; e.g. CIMD accepts loopback redirect URIs regardless of `application_type`, DCR does not), and a thin `Commands.RegisterClient` that validates a registration request and writes through `oauthclient.Store` with `Source: DCR`.

CIMD later adds `pkg/lib/cimd` with its own fetcher and validator, writing the same store with `Source: CIMD`. `dcr` and `cimd` both depend on `oauthclient`; neither depends on the other, and `oauthclient` depends on neither.

**`DCRClientIDPrefix` moves out of `pkg/lib/dcr/token.go` (Part 1 §5.2) into `pkg/lib/oauthclient/client_id.go`.** This is what keeps the dependency one-directional. The prefix is not really a DCR-protocol secret: it is the *shape* of a `client_id`, and the party that most needs to recognise it is the resolver, in the same package. Putting it here also gives CIMD's URL-shape predicate an obvious home right next to it — the two together are exactly `isDynamicClientIDCandidate` (Part 3 §4.1). Generation moves with it as `GenerateDCRClientID()`, which `dcr.Commands.RegisterClient` calls; the IAT prefixes and `GenerateInitialAccessToken` stay in `dcr`, since an IAT is not a client id.

This supersedes Part 1 §5's note about `pkg/lib/dcr` hosting the client store, and Part 1 §5.2's note about the prefix staying in `dcr`.

**One consequence worth having on purpose:** with `Resolver` and `Queries` in the same package, Part 3's cross-package `DynamicClientQueries` interface and its `wire.Bind` disappear — `Resolver` holds `*Queries` directly. Keep a narrow interface only if the resolver's own unit tests want to fake the store; there is no longer a wiring reason for one.

The rest of this section reads as if these types were in `dcr` where it was written before the split; substitute `oauthclient` for `dcr` in §4.1 and §4.2, and keep §4.3 in `dcr`.

### 4.1 `pkg/lib/oauthclient/client.go` (new file)

```go
package oauthclient

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
)

// Source and Kind reuse the API-facing enums rather than redeclaring them:
// the stored strings are the GraphQL enum values verbatim, exactly as
// Part 1 does for the IAT's token_type column.
type Source = model.OAuthClientSource // "DCR" | "CIMD"
type Kind = model.OAuthClientKind     // "FIRST_PARTY" | "THIRD_PARTY"

type Client struct {
	ID              string // row uuid, GraphQL Node id
	ClientID        string // dcrc_... (DCR) or an https:// URL (CIMD)
	Source          Source
	CreatedAt       time.Time
	UpdatedAt       time.Time
	LastFetchedAt   *time.Time // CIMD only; always nil for Source == DCR
	Kind            Kind
	ApplicationType string // "web" | "native"
	ClientName      *string
	ClientURI       *string
	LogoURI         *string
	TOSURI          *string
	PolicyURI       *string
	RedirectURIs    []string
	GrantTypes      []string
	ResponseTypes   []string
}

// RegisteredAt is derived, not stored: client.md defines registeredAt as the
// DCR registration timestamp and null for CIMD ("there is no registration
// event, only a fetch"). See §3.1.
func (c *Client) RegisteredAt() *time.Time {
	if c.Source != model.OAuthClientSourceDCR {
		return nil
	}
	t := c.CreatedAt
	return &t
}

// DisplayName mirrors client.md's "name" field rule: client_name, or a
// generated "Client <clientID>" fallback. ClientName is never backfilled
// with this fallback in storage (Store.NewClient stores nil when omitted,
// §5.3) -- the fallback is a pure function of ClientID, so computing it
// here on every read is equivalent to storing it, without a stale copy to
// keep in sync if the generation rule ever changes.
func (c *Client) DisplayName() string {
	if c.ClientName != nil && *c.ClientName != "" {
		return *c.ClientName
	}
	return "Client " + c.ClientID
}
```

`ToModel` takes the DCR config rather than living on `Client` alone, because mapping to `model.OAuthClient` needs the resolved token-lifetime defaults from `*config.OAuthDynamicClientRegistrationConfig`, which `Client` itself does not hold.

The listing below shows the field-by-field mapping. **Its signature and the four token-lifetime lines are superseded** by the note that follows it — read both before implementing; the lifetimes come from `ToClientConfig()`, not from a separate resolved-defaults struct.

```go
// pkg/lib/oauthclient/client_model.go
func (c *Client) ToModel(defaults ResolvedDefaultClientConfig) *model.OAuthClient {
	return &model.OAuthClient{
		Meta: model.Meta{
			ID:        c.ID, // row uuid — the relay Node id, not ClientID
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt, // == CreatedAt for DCR; real for CIMD refetches
		},
		ClientID:                       c.ClientID,
		Source:                         c.Source,
		Kind:                           c.Kind,
		IsConfidential:                 false,
		IsServiceClient:                false,
		ApplicationType:                &c.ApplicationType,
		Name:                           c.DisplayName(),
		ClientName:                     c.ClientName,
		ClientURI:                      c.ClientURI,
		LogoURI:                        c.LogoURI,
		TOSURI:                         c.TOSURI,
		PolicyURI:                      c.PolicyURI,
		RedirectURIs:                   c.RedirectURIs,
		PostLogoutRedirectURIs:         []string{}, // always empty for both dynamic sources
		GrantTypes:                     c.GrantTypes,
		ResponseTypes:                  c.ResponseTypes,
		AccessTokenLifetimeSeconds:     int(defaults.AccessTokenLifetime),
		RefreshTokenLifetimeSeconds:    int(defaults.RefreshTokenLifetime),
		RefreshTokenIdleTimeoutEnabled: defaults.RefreshTokenIdleTimeoutEnabled,
		RefreshTokenIdleTimeoutSeconds: int(defaults.RefreshTokenIdleTimeout),
		RefreshTokenRotationEnabled:    false,
		IssueJWTAccessToken:            false, // fixed per client.md's DCR mapping table
		MaxConcurrentSession:           0,
		// all remaining "static clients only" fields (CustomUIURI, App2appEnabled,
		// App2appInsecureDeviceKeyBindingEnabled, DPoPDisabled,
		// AuthenticationFlowAllowlist, PreAuthenticatedURLEnabled,
		// PreAuthenticatedURLAllowedOrigins, ReplaceProjectLogoWithLogoURI)
		// are left at Go zero value (false/nil/empty), matching client.md's
		// "All Authgear extension fields are fixed at their zero values for
		// DCR clients" rule — cimd.md states the same for CIMD clients.
		RegisteredAt:   c.RegisteredAt(), // nil for CIMD, CreatedAt for DCR
		LastFetchedAt:  c.LastFetchedAt,  // nil for DCR
	}
}
```

**Do not hand-roll the lifetime fallbacks here.** `ResolvedDefaultClientConfig` must be produced by running the *same* code the runtime uses, otherwise `dynamicClients` reports token lifetimes that differ from the ones actually enforced at `/oauth2/token` — a silent, hard-to-notice divergence between the Admin API view and reality.

Concretely: move `Client.ToClientConfig(defaults)` — specified in [Part 3 §3](2026-08-17-03-client-resolution.md) — **forward into this part**. It is a pure function with no runtime wiring (it only builds a `*config.OAuthClientConfig` and calls its `SetDefaults()`), so nothing about it needs to wait for Part 3, and Part 3 then simply consumes it instead of introducing it. Define `ToModel` in terms of it:

```go
// tokenLifetimes is whichever project config governs this row's source:
// oauth.dynamic_client_registration.default_client_config for DCR,
// oauth.client_id_metadata_document.client_config for CIMD (§3.1). Resolving
// that from *config.OAuthConfig belongs in one helper —
// oauthclient.ResolveTokenLifetimes(oauthConfig, source) — so no caller has
// to remember which key applies to which source.
func (c *Client) ToModel(tokenLifetimes *config.OAuthDynamicClientRegistrationDefaultClientConfig) *model.OAuthClient {
	// One source of truth for resolved lifetimes: the same synthesized
	// config the resolver hands to the OAuth runtime (Part 3 §4).
	cfg := c.ToClientConfig(tokenLifetimes)
	return &model.OAuthClient{
		// ...
		AccessTokenLifetimeSeconds:     int(cfg.AccessTokenLifetime),
		RefreshTokenLifetimeSeconds:    int(cfg.RefreshTokenLifetime),
		RefreshTokenIdleTimeoutEnabled: *cfg.RefreshTokenIdleTimeoutEnabled, // non-nil after SetDefaults()
		RefreshTokenIdleTimeoutSeconds: int(cfg.RefreshTokenIdleTimeout),
		// ...
	}
}
```

`ResolvedDefaultClientConfig` as a separate struct is therefore **dropped** from this plan — it existed only to carry the four resolved numbers, which `*config.OAuthClientConfig` already carries. Add `GetDefaultClientConfig()` as a third nil-safe accessor alongside `IsEnabled()`/`IsInitialAccessTokenRequired()` (§2.2), returning nil on a nil receiver.

`ResolveTokenLifetimes(oauthConfig, source)` is the one place that maps `source` → config key. In this part it has a single `case model.OAuthClientSourceDCR` and a `default: return nil`; CIMD adds its case and nothing else changes. Note the two config structs are field-identical, so CIMD's `client_config` can reuse `OAuthDynamicClientRegistrationDefaultClientConfig` — or, better, rename that type to something source-neutral (`OAuthDynamicClientTokenLifetimesConfig`) when CIMD lands. Not worth renaming pre-emptively here, but do not bake the DCR name into `ToClientConfig`'s or `ToModel`'s *semantics*: they take "the token lifetimes for this client", not "the DCR defaults".

The concrete values this produces when `default_client_config` is absent are tabulated in §2.2.1. Assert them explicitly in `client_model_test.go` (§7) — a test that pins `1800` / `31449600` / `true` / `2592000` is what catches an accidental reintroduction of hand-rolled fallbacks.

### 4.2 `pkg/lib/oauthclient/store_client.go` (new file)

Same conventions as `pkg/lib/resourcescope/store_resource.go`. Every method is source-agnostic except where noted:

- `func (s *Store) NewClient(options *NewClientOptions) *Client` — builds the struct with `ID: uuid.NewString()`, `CreatedAt`/`UpdatedAt` from `s.Clock.NowUTC()`, and `Source`/`ClientID` taken from `options`. **The store does not generate the `client_id`**: DCR's caller passes `GenerateDCRClientID()` (§4, `client_id.go`), while a CIMD `client_id` is the caller-supplied URL. Keeping generation out of the store is what lets one store serve both.
- `func (s *Store) CreateClient(ctx context.Context, c *Client) error` — `INSERT INTO _auth_oauth_client (...)`, marshaling `RedirectURIs`/`GrantTypes`/`ResponseTypes` to JSON. There is no existing `jsonb` marshal/scan helper to copy in `pkg/lib/resourcescope` (the `metadata` columns there are never read or written — see Part 4 §2.2), so introduce a small `driver.Valuer`/`sql.Scanner` wrapper for a `[]string` jsonb column here and reuse it for `access_policy` in Part 4. `databaseutil.IsDuplicateKeyError` → `ErrDynamicClientDuplicateClientID`.
- `func (s *Store) UpsertClient(ctx context.Context, c *Client) error` — **CIMD only, not called in this part.** cimd.md requires a refetch to overwrite the same record in place; `INSERT ... ON CONFLICT (app_id, client_id) DO UPDATE` with `updated_at`/`last_fetched_at` refreshed. Listed here so the store's shape is settled, but implement it with CIMD, not now — an unused method is worse than a documented gap.
- `func (s *Store) GetClientByClientID(ctx context.Context, clientID string) (*Client, error)` — `SELECT ... WHERE client_id = ?`; `sql.ErrNoRows` → `ErrDynamicClientNotFound`. Deliberately **not** filtered by source: the runtime resolver looks up by `client_id` alone and learns the source from the row.
- `func (s *Store) GetClientByID(ctx context.Context, id string) (*Client, error)` — by row uuid, for the GraphQL Node loader.
- `func (s *Store) DeleteClientByClientID(ctx context.Context, clientID string) error` — `DELETE ... WHERE client_id = ?`; 0 rows → `ErrDynamicClientNotFound`. Also source-agnostic, matching `deleteDynamicClient`'s contract ("a DCR-registered or CIMD-resolved client").
- `func (s *Store) ListClients(ctx context.Context, pageArgs graphqlutil.PageArgs) (*storeListClientResult, error)` — same `db.ApplyPageArgs` pattern as `Store.ListResources` (`pkg/lib/resourcescope/store_resource.go:174-199`), ordered `created_at DESC`, **all sources**. No filter arguments — dcr.md's `dynamicClients` query takes only `first/after/last/before`, and the client distinguishes sources via the `source` field.
- `func (s *Store) CountClientsBySource(ctx context.Context, source Source) (uint64, error)` — `SELECT COUNT(*) ... WHERE source = ?`, served by the `(app_id, source)` index. **Must** take a source: dcr.md and cimd.md define separate quotas counted over `source: DCR` and `source: CIMD` respectively (Part 5 §5). A source-less `CountClients` would silently make each source's quota consume the other's.

Add to `pkg/lib/oauthclient/errors.go`:

```go
var ErrDynamicClientNotFound = errors.New("dynamic client not found")
var ErrDynamicClientDuplicateClientID = errors.New("duplicate dynamic client id")
```

### 4.3 `pkg/lib/dcr/validate.go` (new file) — Accepted Client Metadata validation

This one stays in `pkg/lib/dcr`, not `oauthclient`: the rules below are the DCR protocol's, and cimd.md deliberately diverges from several of them (loopback redirect URIs allowed regardless of `application_type`; `token_endpoint_auth_method: none` accepted rather than rejected outright). CIMD gets its own validator over the same `Client` record.


Implements every rule in [docs/specs/dcr.md — Accepted Client Metadata](../../specs/dcr.md#accepted-client-metadata) and [Errors](../../specs/dcr.md#errors). One function per RFC 7591 error family, matching the `resourcescope.FormatXxx{}.CheckFormat` style used for URI validation (`pkg/lib/resourcescope/formats.go`) where applicable:

```go
type RegistrationRequest struct {
	ClientName      *string
	RedirectURIs    []string
	GrantTypes      []string // nil means "not provided" (apply default)
	ResponseTypes   []string
	ApplicationType *string
	LogoURI         *string
	ClientURI       *string
	TOSURI          *string
	PolicyURI       *string
	// TokenEndpointAuthMethod is only used for rejecting the request if present.
	TokenEndpointAuthMethod *string
}

// ValidateAndNormalize applies defaults (grant_types, response_types,
// application_type) and returns invalid_client_metadata / invalid_redirect_uri
// protocol errors (see handler_register.go for how these map to HTTP status).
func ValidateAndNormalize(req *RegistrationRequest) (*NormalizedRegistration, error)
```

Rules implemented (all as sub-checks inside `ValidateAndNormalize`, each returning a specific `error` value so the HTTP handler layer can map to the exact `error` string from the spec's table):

- `redirect_uris` required, non-empty → else `ErrDCRRedirectURIsMissing` (`invalid_client_metadata`).
- Each redirect URI: absolute URI (RFC 3986 §4.3), no fragment → else `ErrDCRRedirectURIInvalid` (`invalid_redirect_uri`).
- Redirect URI scheme rules depend on the **effective** `application_type` (after defaulting to `web`): `web` → `https://` only; `native` → custom scheme or `http://localhost` — else `ErrDCRRedirectURIInvalid` (`invalid_redirect_uri`).
- `token_endpoint_auth_method` present at all → `ErrDCRTokenEndpointAuthMethodNotAccepted` (`invalid_client_metadata`).
- `grant_types` subset of `{authorization_code, refresh_token}` → else `ErrDCRGrantTypeUnsupported` (`invalid_client_metadata`). Default `["authorization_code", "refresh_token"]` when omitted.
- `response_types` subset of `{code}`, consistent with `grant_types` (code ⟺ authorization_code) → else `ErrDCRResponseTypeInconsistent` (`invalid_client_metadata`). Default `["code"]` when omitted.
- `application_type` ∈ `{web, native}` (omitted → `web`) → else `ErrDCRApplicationTypeUnsupported` (`invalid_client_metadata`).
- `logo_uri`, `client_uri`, `tos_uri`, `policy_uri`: each, if present, must be `https://` → else `ErrDCRURIFieldNotHTTPS` (`invalid_client_metadata`).

This intentionally does **not** implement `oauth.dynamic_client_registration.enabled`, IAT-required/type checks, or the client limit — those are request-level authorization concerns handled by the handler in §5, not client-metadata shape concerns.

## 5. Registration Endpoint

### 5.1 `pkg/lib/oauth/handler/handler_register.go` (new file)

Mirrors `handler_token.go`'s structure (dependencies injected, `Handle` method wraps a DB transaction via `Database *appdb.Handle`):

```go
type RegistrationHandlerDCRService interface {
	RegisterClient(ctx context.Context, options *dcr.RegisterClientOptions) (*model.OAuthClient, error)
	CountClientsBySource(ctx context.Context, source model.OAuthClientSource) (uint64, error) // present now for symmetry; wired up by Part 5's usage-limit check (docs/plans/dcr/2026-08-17-05-client-usage-limit.md)
}

type RegistrationHandlerIATService interface {
	ValidateAndGetByToken(ctx context.Context, plaintext string) (*model.OAuthInitialAccessToken, error)
}

type RegistrationHandler struct {
	Database    *appdb.Handle
	OAuthConfig *config.OAuthConfig
	DCR         RegistrationHandlerDCRService
	IAT         RegistrationHandlerIATService
	Clock       clock.Clock
}

func (h *RegistrationHandler) Handle(ctx context.Context, r *http.Request) (*RegistrationResponse, error)
```

Call sequence inside `Handle` (all inside one `h.Database.WithTx`, matching `handler_token.go`'s transaction wrapping style):

1. If `!h.OAuthConfig.DynamicClientRegistration.IsEnabled()` → return `protocol.NewErrorStatusCode("access_denied", "dynamic client registration is not enabled", 403)`. (Nil-safe accessor, §2.2 — the field really can be nil.)
2. Parse `Authorization: Bearer <token>` header, if present.
3. Determine `iatKind`:
   - No IAT presented, `IsInitialAccessTokenRequired() == true` → `protocol.NewErrorStatusCode("invalid_initial_access_token", "an initial access token is required", 401)`.
   - No IAT presented, `IsInitialAccessTokenRequired() == false` (open registration) → `iatKind = ThirdParty`, and `application_type` must resolve to `web`/`native` only (already guaranteed by §4.3's validation — no extra check needed since DCR only ever accepts those two values).
   - IAT presented → `h.IAT.ValidateAndGetByToken(ctx, token)`; `dcr.ErrInitialAccessTokenNotFound` → `protocol.NewErrorStatusCode("invalid_initial_access_token", "invalid or expired initial access token", 401)`. On success, `iatKind` = `THIRD_PARTY` or `FIRST_PARTY` per the token's `Type`.
4. Decode JSON body into `dcr.RegistrationRequest`; malformed JSON → `protocol.NewErrorStatusCode("invalid_client_metadata", "malformed JSON body", 400)`.
5. `dcr.ValidateAndNormalize(&req)` (§4.3) → maps each returned sentinel error to its `(error, status)` pair from the table in [dcr.md — Errors](../../specs/dcr.md#errors) via `protocol.NewErrorStatusCode`.
6. Build `dcr.RegisterClientOptions{ Kind: iatKind, ApplicationType, ClientName, RedirectURIs, GrantTypes, ResponseTypes, ClientURI, LogoURI, TOSURI, PolicyURI }` and call `h.DCR.RegisterClient(ctx, options)`. `dcr.Commands.RegisterClient` generates the `dcrc_` id, builds a `oauthclient.Client` with `Source: DCR`, and writes it via `oauthclient.Store.CreateClient`.
7. Map the returned `*model.OAuthClient` to the RFC 7591 response body (§5.3).

`RegistrationHandlerDCRService`/`RegistrationHandlerIATService` are both backed by `*dcr.Commands`/`*dcr.Queries` via the same wire-binding pattern as `facade.OAuthClientResolver` (`pkg/lib/deps/deps_common.go:689-695`). `CountClientsBySource` and the advisory lock (Part 5 §4) are re-exported from `dcr.Commands` so the handler depends on one collaborator rather than reaching into `oauthclient` directly.

### 5.1.1 Rate limiting `POST /oauth2/register`

Under open registration (`initial_access_token_required: false`) this is a **completely unauthenticated endpoint that writes a database row per call**. Part 5's `usage.limits.oauth_client_dcr` is a *standing* cap on the total client count, not a rate — it does not slow anyone down, and once the quota is full, `action: block` means no further client can register at all. Without a rate limit, a single caller can burn a project's entire client quota in seconds and lock out every legitimate MCP client, since nothing reclaims slots automatically. dcr.md specifies no rate limit today; this section designs one.

#### Design

Two buckets, both project-configurable under `oauth.dynamic_client_registration.rate_limits` in `authgear.yaml` (§2.2.2), following the existing project-configurable rate limit precedent (`AuthenticationRateLimitsSignupConfig`, `pkg/lib/config/authentication_rate_limits.go`) rather than the hard-coded OAuth token endpoint buckets in `pkg/lib/oauth/handler/ratelimit.go` — those stay hard-coded, but DCR registration volume is expected to scale with how many distinct legitimate clients self-register (e.g. an MCP-style integration where each new user's install registers once), which a flat non-configurable number can't accommodate across every plan tier.

| Name | Bucket key | Default rate | Rationale |
|---|---|---|---|
| `oauth.register.per_ip` | remote IP, per project | 10 / minute | Mirrors `authentication.signup.per_ip` (also 10/minute), the closest existing analogue: unauthenticated creation of a new persistent record. Generous enough for a CI system registering a batch of PR-preview clients, tight enough that hammering is pointless. |
| `oauth.register.per_project` | project | 1000 / hour | Caps how fast a project's client population can grow regardless of source IP — a per-IP limit alone is useless against a distributed caller. Earlier drafts fixed this at 60/hour, sized against the *example* `quota: 20` in §Client Limit — but `oauth_client_dcr`'s quota is plan-tier-controlled and can be set much higher, and legitimate DCR volume (many distinct one-time self-registrations, not repeat traffic from one caller) can spike well past 60/hour with zero abuse. A flat number can't track every plan's quota, so it's raised to a permissive default and made configurable instead. |

**What these limits do and do not achieve.** They bound *request volume* against an unauthenticated, DB-writing endpoint, which is their job. They do **not** meaningfully protect the standing quota on their own: with a small `quota` (e.g. the example `quota: 20`), the per-IP bucket alone (burst 10, refill 1m) lets one caller create 20 clients in about a minute, regardless of what `per_project` is set to. A project that wants the per-project bucket to meaningfully bind before its own `oauth_client_dcr` quota is exhausted should configure `per_project` below that quota's refill-equivalent rate itself — that is a per-project tuning decision, not something a single built-in default can get right for every quota. Do not describe these limits as quota protection in the PR or the spec — see the residual-risk note at the end of this section.

Both are consumed on **every** attempt, successful or not. This deliberately differs from credential-verification limits (which only take a token on failure, per rate-limit.md's Algorithm section): here the resource being protected is client creation itself, so successes must count.

#### `pkg/lib/ratelimit` additions

```go
// pkg/lib/ratelimit/ratelimits.go

// group
RateLimitGroupOAuthRegister RateLimitGroup = "oauth.register"

// names
RateLimitOAuthRegisterPerIP      RateLimitName = "oauth.register.per_ip"
RateLimitOAuthRegisterPerProject RateLimitName = "oauth.register.per_project"

// bucket names
OAuthRegisterPerIP      BucketName = "OAuthRegisterPerIP"
OAuthRegisterPerProject BucketName = "OAuthRegisterPerProject"
```

A new `RateLimitGroup` needs no further wiring: `RateLimitGroup.ResolveWeight` (`ratelimits.go:615-660`) ends in `default: weight = resolveWeight(r, "")`, so an unknown group resolves to weight 1 with no code change and no panic. Only the groups with a documented fallback are enumerated there, and `oauth.register` has none — consistent with `oauth.token.general`, which is also absent from that switch.

`oauth.register.per_project` takes no `args` in `NewBucketSpec`; the bucket key already includes the app id via `bucketKeyApp(l.AppID, spec)` (`pkg/lib/ratelimit/limiter.go:71-75`).

#### `pkg/lib/oauth/handler/ratelimit.go` additions

Unlike the two hard-coded OAuth token endpoint specs already in this file, these take the resolved `*config.OAuthDynamicClientRegistrationRateLimitsConfig` (§2.2.2) instead of constructing a `config.RateLimitConfig` literal — the built-in defaults live in that type's own `SetDefaults()`, not duplicated here:

```go
func NewBucketSpecOAuthRegisterPerIP(rateLimits *config.OAuthDynamicClientRegistrationRateLimitsConfig, ip string) ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(ratelimit.RateLimitOAuthRegisterPerIP, ratelimit.RateLimitGroupOAuthRegister, rateLimits.PerIP, ratelimit.OAuthRegisterPerIP, ip)
}

func NewBucketSpecOAuthRegisterPerProject(rateLimits *config.OAuthDynamicClientRegistrationRateLimitsConfig) ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(ratelimit.RateLimitOAuthRegisterPerProject, ratelimit.RateLimitGroupOAuthRegister, rateLimits.PerProject, ratelimit.OAuthRegisterPerProject)
}
```

Both call sites in `Handle` (§handler wiring below) resolve `rateLimits := h.OAuthConfig.DynamicClientRegistration.GetRateLimits()` once and pass it to both constructors — `GetRateLimits()` is always non-nil at request time (§2.2.2).

#### Handler wiring

`RegistrationHandler` (§5.1) gains two fields, mirroring `TokenHandler`'s (`handler_token.go:132-134,241`):

```go
type RegistrationHandlerRateLimiter interface {
	Allow(ctx context.Context, spec ratelimit.BucketSpec) (*ratelimit.FailedReservation, error)
}

type RegistrationHandler struct {
	// ... existing fields ...
	RemoteIP    httputil.RemoteIP
	RateLimiter RegistrationHandlerRateLimiter
}
```

plus a `checkRateLimit` copied from `TokenHandler.checkRateLimit` (`handler_token.go:2288-2302`) — `Allow`, then `failedReservation.Error()`, then map `ratelimit.RateLimited` to a protocol error.

Revised `Handle` call sequence — the two checks slot in as the **new steps 2 and 3**, renumbering the rest of §5.1's list:

1. DCR enabled check (unchanged; a pure in-memory config read, so it stays first and a disabled endpoint costs no Redis round-trip).
2. `rateLimits := h.OAuthConfig.DynamicClientRegistration.GetRateLimits()`, then `checkRateLimit(ctx, NewBucketSpecOAuthRegisterPerIP(rateLimits, string(h.RemoteIP)))`.
3. `checkRateLimit(ctx, NewBucketSpecOAuthRegisterPerProject(rateLimits))`.
4. … the existing steps 2-7 (bearer parse, IAT validation, body decode, metadata validation, usage limit, create).

Both checks run **before** the `Authorization` header is parsed and before any IAT lookup, so an attacker cannot use invalid-IAT attempts to probe or to force DB reads. Narrowest-scope-first ordering (per IP, then per project) matches rate-limit.md's short-circuit convention and means a single misbehaving IP is stopped without consuming the shared project bucket.

#### Error response

`checkRateLimit` returns `protocol.NewErrorStatusCode("x_rate_limited", "rate limit exceeded, please try again later.", http.StatusTooManyRequests)` — byte-identical to what `/oauth2/token` already returns (`handler_token.go:2299`). RFC 7591 §3.2.2 defines no rate-limit error code, and `x_rate_limited` is the established Authgear extension, so reusing it keeps the two OAuth endpoints consistent. §5.2's HTTP wrapper already renders any `*protocol.OAuthProtocolError` with its `StatusCode`, so no change is needed there.

#### Residual risk — accepted, not solved

Under open registration a caller can still fill a project's `oauth_client_dcr` quota (roughly a minute at `quota: 20`, per the note above), after which legitimate registration is refused until an admin deletes clients.

**Decision: accept this. No further mitigation is in scope for DCR.** A per-project usage limit plus a rate limit are the two controls Authgear applies to every comparable endpoint, both are now in place, and an availability attack against a deliberately open, unauthenticated endpoint cannot be fully prevented — a project that cannot tolerate that should leave `initial_access_token_required: true`, which is the default.

Two things to keep visible rather than treat as solved:

- **A standing limit does not self-heal, unlike every other limit in the system.** A rate-limit bucket refills; a periodic usage limit resets each period; an exhausted `oauth_client_dcr` quota stays exhausted until a human calls `deleteDynamicClient`. That asymmetry — not the attacker's ability to send requests — is what makes this failure mode worse than it looks. If it ever needs addressing, TTL-based eviction of DCR clients that were registered but never used in an authorization flow is the cheap fix, because it restores the self-healing property without adding a new control surface. Explicitly **not** built here.
- Recommend that projects enabling open registration configure a lower `action: alert` entry alongside their `action: block` entry (Part 5 §2.2), so an admin hears about exhaustion rather than discovering it from a support ticket.

The same asymmetry applies to `oauth_client_cimd`, and more sharply: cimd.md notes that deleting a CIMD client is not a durable ban, since the same URL can be presented again on the next authorization request.

### 5.2 `pkg/auth/handler/oauth/register.go` (new file) — thin HTTP wrapper, mirrors `token.go` exactly

```go
func ConfigureRegisterRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("POST", "OPTIONS").WithPathPattern("/oauth2/register")
}

type ProtocolRegistrationHandler interface {
	Handle(ctx context.Context, r *http.Request) (*handler.RegistrationResponse, error)
}

type RegistrationHandler struct {
	RegistrationHandler ProtocolRegistrationHandler
}

func (h *RegistrationHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	resp, err := h.RegistrationHandler.Handle(r.Context(), r)
	if err != nil {
		var oauthErr *protocol.OAuthProtocolError
		if errors.As(err, &oauthErr) {
			status := oauthErr.StatusCode
			if status == 0 {
				status = 400
			}
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(status)
			_ = json.NewEncoder(rw).Encode(oauthErr.Response)
			return
		}
		http.Error(rw, "Internal Server Error", 500)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(rw).Encode(resp)
}
```

### 5.3 `RegistrationResponse` — response DTO

`pkg/lib/oauth/handler/handler_register.go`:

```go
type RegistrationResponse struct {
	ClientID          string   `json:"client_id"`
	ClientIDIssuedAt  int64    `json:"client_id_issued_at"`
	ClientName        string   `json:"client_name,omitempty"`
	RedirectURIs      []string `json:"redirect_uris"`
	GrantTypes        []string `json:"grant_types"`
	ResponseTypes     []string `json:"response_types"`
	ApplicationType   string   `json:"application_type"`
	ClientURI         string   `json:"client_uri,omitempty"`
	LogoURI           string   `json:"logo_uri,omitempty"`
	TOSURI            string   `json:"tos_uri,omitempty"`
	PolicyURI         string   `json:"policy_uri,omitempty"`
}
```

Built directly from the `*model.OAuthClient` returned by `CreateClient` — no `client_secret*` fields at all (never issued, per spec).

`ClientName` on this DTO is populated from `model.OAuthClient.Name` — the computed display name (`Client.DisplayName()`, §4.1) — **not** from `model.OAuthClient.ClientName` (the raw, possibly-nil OIDC `client_name`). This is what satisfies dcr.md's UC2 (a request with no `client_name` echoes back `"client_name": "Client dcrc_AbCdEfGhIjKlMnOpQr"`) without needing to persist the generated fallback: `DisplayName()` is a pure function of `ClientID`, which is immutable per client, so computing it at read/response time produces byte-identical output to storing it, with no stale copy to keep in sync if the generation rule ever changes.

`Store.NewClient` (§4.2) therefore does **not** bake in a fallback — it normalizes an explicit empty string to nil (so "omitted" and "sent empty" both read as nil the same way) and otherwise stores `options.ClientName` verbatim:

```go
clientName := options.ClientName
if clientName != nil && *clientName == "" {
	clientName = nil
}
```

Consequences, all intentional:

- `_auth_oauth_client.client_name` **is** NULL for a row created by `POST /oauth2/register` when the request omits `client_name` — the column's nullability (§3.1) is meaningful, not vestigial.
- `Client.DisplayName()` (§4.1) is the sole authority for the generated fallback — every reader (registration response, `ToModel`'s `Name`, portal display) goes through it; nothing else re-implements the `"Client " + ClientID` rule.
- `model.OAuthClient.Name` and `.ClientName` **diverge** for a DCR client that omitted `client_name`: `Name` is the computed `"Client dcrc_..."`, `.ClientName` is `nil`. client.md's "Null for static clients that do not set client_name" note is extended to cover this DCR/CIMD case too (doc-fix, see §9) — it's no longer a static-client-only exception.
- `clientName` in the Admin API can be `null` for a DCR client, exactly like the static clients that don't set it — this is a genuine (if narrow) Admin API-visible behavior point, not just an internal storage detail.
- The `omitempty` on this DTO's `ClientName` field never actually fires in practice (a registration always has *some* display name, generated or explicit), but it costs nothing to leave in place.

### 5.4 Route wiring — `pkg/auth/routes.go`

```go
router.Add(oauthhandler.ConfigureRegisterRoute(oauthAPIRoute), p.Handler(newOAuthRegisterHandler))
```

Placed next to the existing `router.Add(oauthhandler.ConfigureTokenRoute(...))` / `ConfigureRevokeRoute(...)` block (`pkg/auth/routes.go:459-461`). Uses the plain `oauthAPIRoute` (declared at `pkg/auth/routes.go:279`) — **not** the `dpopOauthAPIRoute` that `/oauth2/token` and `/oauth2/revoke` actually use. Registration has no DPoP proof requirement per the spec, so the plain chain (CORS + public-origin + no-store, no DPoP) is correct; the nearest same-class precedent in that block is `ConfigureEndSessionRoute(oauthAPIRoute)` at line 461, not the token/revoke routes.

Add `newOAuthRegisterHandler` provider (wire) in `pkg/auth/deps.go`, following the exact pattern of `newOAuthTokenHandler`. Regenerate `pkg/auth/wire_gen.go` via `make generate` in the same commit.

### 5.5 OIDC Discovery Metadata

`pkg/auth/handler/oauth/metadata.go` is **not** where this goes — it is a generic aggregator that merges `map[string]any` from a list of injected `MetadataProvider`s (`MetadataHandler.ServeHTTP`, lines 30-42) and holds no field knowledge at all. Both `/.well-known/openid-configuration` and `/.well-known/oauth-authorization-server` are served by that same aggregator, so a field added to a provider appears in both, which is exactly what the spec asks for.

Three changes are needed:

1. `pkg/lib/endpoints/endpoints.go` — add `func (e *Endpoints) RegisterEndpointURL() *url.URL { return e.urlOf("oauth2/register") }`, next to `TokenEndpointURL`/`RevokeEndpointURL` (lines 69-70).
2. `pkg/lib/oauth/metadata.go` — add `RegisterEndpointURL() *url.URL` to the `EndpointsProvider` interface consumed by `MetadataProvider`, and add an `OAuthConfig *config.OAuthConfig` field to the `MetadataProvider` struct (it currently holds only `Endpoints`).
3. `pkg/lib/oauth/metadata.go`'s `PopulateMetadata` (line 12) — append:

   ```go
   if p.OAuthConfig.DynamicClientRegistration.IsEnabled() {
       meta["registration_endpoint"] = p.Endpoints.RegisterEndpointURL().String()
   }
   ```

Note there is **no existing conditionally-present field to mirror** in either provider: every field in `pkg/lib/oauth/metadata.go:12-24` and `pkg/lib/oauth/oidc/metadata.go:51-59` is set unconditionally (`end_session_endpoint` at `oidc/metadata.go:59` included), so `registration_endpoint` is the first one. Adding the `OAuthConfig` field means `make generate` must be run for the binaries that construct `MetadataProvider` via wire.

## 6. Unified `OAuthClient` GraphQL Type & Admin API

### 6.1 `pkg/api/model/oauth_client.go` (new file)

Full field set per [client.md — GraphQL Type](../../specs/client.md#graphql-type). Only DCR-sourced values are populated starting this part (CIMD is future work; static clients are not exposed through this type at all per dcr.md's "New query" note — "Static clients are managed via authgear.yaml and are not returned here").

```go
package model

import "time"

type OAuthClientSource string

const (
	OAuthClientSourceStatic OAuthClientSource = "STATIC"
	OAuthClientSourceDCR    OAuthClientSource = "DCR"
	OAuthClientSourceCIMD   OAuthClientSource = "CIMD" // reserved; unused until cimd.md ships
)

type OAuthClientKind string

const (
	OAuthClientKindFirstParty OAuthClientKind = "FIRST_PARTY"
	OAuthClientKindThirdParty OAuthClientKind = "THIRD_PARTY"
)

type OAuthClient struct {
	// model.Meta carries the internal row uuid (NOT the OAuth client_id) and
	// the creation timestamp. It is required for two independent reasons:
	//
	//   1. pkg/admin/graphql's entityIDField / entityCreatedAtField do an
	//      unchecked obj.(EntityRef) assertion, where EntityRef is
	//      `GetMeta() model.Meta` (pkg/admin/graphql/entity.go:41-70) — a
	//      model without Meta panics at resolve time, not compile time.
	//   2. The relay Node global id and the DataLoader key (§6.6) are both
	//      the row uuid, since `client_id` is an externally-chosen-looking
	//      string that does not belong in a global id.
	//
	// This is why client.md's `type OAuthClient` SDL has no `id` field and
	// does not implement Node: the Node id is a transport concern of the
	// Admin GraphQL layer, not part of the client model. No client.md change
	// is needed; the implemented schema simply has `id: ID!` in addition.
	//
	// Meta.UpdatedAt is set equal to Meta.CreatedAt for DCR clients (a DCR
	// client is immutable until RFC 7592 lands) and is never exposed —
	// nodeOAuthClient declares no "updatedAt" field and does not implement
	// entityInterface.
	model.Meta

	ClientID                                string
	Source                                  OAuthClientSource
	Kind                                    OAuthClientKind
	IsConfidential                          bool
	IsServiceClient                         bool
	ApplicationType                         *string
	Name                                    string
	ClientName                              *string
	ClientURI                               *string
	LogoURI                                 *string
	TOSURI                                  *string
	PolicyURI                               *string
	RedirectURIs                            []string
	PostLogoutRedirectURIs                  []string
	GrantTypes                              []string
	ResponseTypes                           []string
	AccessTokenLifetimeSeconds              int
	RefreshTokenLifetimeSeconds             int
	RefreshTokenIdleTimeoutEnabled          bool
	RefreshTokenIdleTimeoutSeconds          int
	RefreshTokenRotationEnabled             bool
	IssueJWTAccessToken                     bool
	MaxConcurrentSession                    int
	CustomUIURI                             *string
	App2appEnabled                          bool
	App2appInsecureDeviceKeyBindingEnabled  bool
	DPoPDisabled                            bool
	PreAuthenticatedURLEnabled              bool
	PreAuthenticatedURLAllowedOrigins       []string
	ReplaceProjectLogoWithLogoURI           bool
	RegisteredAt                            *time.Time
	LastFetchedAt                           *time.Time
}
```

`authenticationFlowAllowlist` (client.md) is omitted from this Go struct for now and hardcoded to `null` directly in the GraphQL resolver (§6.2) — it is "Static clients only; always null for DCR clients," and Part 2 only ever constructs `OAuthClient` values for DCR clients, so there is no live data to carry in the Go struct yet. Add the field to this struct (and thread it through) only when static clients are also mapped into this unified type, which is out of scope here.

### 6.2 `pkg/admin/graphql/oauth_client.go` (new file)

Pattern: `pkg/admin/graphql/resource.go`'s `node(...)` + `pkg/admin/graphql/authenticator.go`'s enum pattern. Two enums (`OAuthClientSource`, `OAuthClientKind`) plus one large object type with mostly direct field mappings and eight fields hardcoded per the "always X for DCR clients" column of client.md's mapping table:

```go
const typeOAuthClient = "OAuthClient"

var oauthClientSourceType = graphql.NewEnum(graphql.EnumConfig{
	Name:   "OAuthClientSource",
	Values: graphql.EnumValueConfigMap{
		"STATIC": &graphql.EnumValueConfig{Value: "STATIC"},
		"DCR":    &graphql.EnumValueConfig{Value: "DCR"},
		"CIMD":   &graphql.EnumValueConfig{Value: "CIMD"},
	},
})

var oauthClientKindType = graphql.NewEnum(graphql.EnumConfig{
	Name:   "OAuthClientKind",
	Values: graphql.EnumValueConfigMap{
		"FIRST_PARTY": &graphql.EnumValueConfig{Value: "FIRST_PARTY"},
		"THIRD_PARTY": &graphql.EnumValueConfig{Value: "THIRD_PARTY"},
	},
})

var nodeOAuthClient = node(
	graphql.NewObject(graphql.ObjectConfig{
		Name:       typeOAuthClient,
		Interfaces: []*graphql.Interface{nodeDefs.NodeInterface}, // uses clientID as external key, but Node id is still the relay global id derived from the row uuid — see loader below
		Fields: graphql.Fields{
			"id":                                     entityIDField(typeOAuthClient),
			"clientID":                                &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"source":                                  &graphql.Field{Type: graphql.NewNonNull(oauthClientSourceType)},
			"kind":                                    &graphql.Field{Type: graphql.NewNonNull(oauthClientKindType)},
			"isConfidential":                          &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"isServiceClient":                         &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"applicationType":                         &graphql.Field{Type: graphql.String},
			"name":                                    &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"clientName":                              &graphql.Field{Type: graphql.String},
			"clientURI":                               &graphql.Field{Type: graphql.String},
			"logoURI":                                 &graphql.Field{Type: graphql.String},
			"tosURI":                                  &graphql.Field{Type: graphql.String},
			"policyURI":                               &graphql.Field{Type: graphql.String},
			"redirectURIs":                            &graphql.Field{Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String)))},
			"postLogoutRedirectURIs":                  &graphql.Field{Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String)))},
			"grantTypes":                              &graphql.Field{Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String)))},
			"responseTypes":                           &graphql.Field{Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String)))},
			"accessTokenLifetimeSeconds":              &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"refreshTokenLifetimeSeconds":             &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"refreshTokenIdleTimeoutEnabled":           &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"refreshTokenIdleTimeoutSeconds":           &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"refreshTokenRotationEnabled":             &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"issueJWTAccessToken":                     &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"maxConcurrentSession":                    &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"customUIURI":                             &graphql.Field{Type: graphql.String},
			"app2appEnabled":                          &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"app2appInsecureDeviceKeyBindingEnabled":  &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"dpopDisabled":                            &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"authenticationFlowAllowlist":             &graphql.Field{Type: authenticationFlowAllowlistType, Resolve: func(p graphql.ResolveParams) (any, error) { return nil, nil }},
			"preAuthenticatedURLEnabled":               &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"preAuthenticatedURLAllowedOrigins":        &graphql.Field{Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String)))},
			"replaceProjectLogoWithLogoURI":            &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"registeredAt":                             &graphql.Field{Type: graphql.DateTime},
			"lastFetchedAt":                            &graphql.Field{Type: graphql.DateTime},
		},
	}),
	&model.OAuthClient{},
	func(ctx context.Context, gqlCtx *Context, id string) (any, error) {
		return gqlCtx.DynamicClients.Load(ctx, id).Value, nil
	},
)

// authenticationFlowAllowlistType: a minimal placeholder GraphQL object
// (AuthenticationFlowAllowlist { groups { name } flows { type name } }) with
// no backing Go data yet — always resolves to nil for every OAuthClient this
// part can produce (all DCR-sourced). Defining the shape now (rather than
// omitting the field) keeps the schema conformant with client.md.
```

### 6.3 `pkg/admin/graphql/query.go` — `dynamicClients` field

```go
"dynamicClients": &graphql.Field{
	Description: "Clients that exist outside authgear.yaml: DCR-registered (and, once implemented, CIMD-resolved) clients.",
	Type:        connDynamicClient.ConnectionType,
	Args:        relay.NewConnectionArgs(graphql.FieldConfigArgument{}),
	Resolve: func(p graphql.ResolveParams) (any, error) {
		ctx := p.Context
		gqlCtx := GQLContext(ctx)
		pageArgs := graphqlutil.NewPageArgs(relay.NewConnectionArguments(p.Args))

		refs, result, err := gqlCtx.DCRFacade.ListClients(ctx, pageArgs)
		if err != nil {
			return nil, err
		}
		var lazyItems []graphqlutil.LazyItem
		for _, ref := range refs {
			lazyItems = append(lazyItems, graphqlutil.LazyItem{
				Lazy:   gqlCtx.DynamicClients.Load(ctx, ref.ID),
				Cursor: graphqlutil.Cursor(ref.Cursor),
			})
		}
		return graphqlutil.NewConnectionFromResult(lazyItems, result)
	},
},
```

**Naming deviation from the literal spec SDL:** `graphqlutil.NewConnectionDef` (`pkg/util/graphqlutil/connection.go:65-76`) always derives the Connection/Edge type name from the node object's own GraphQL `Name` (`schema.Name()`), so `connDynamicClient := graphqlutil.NewConnectionDef(nodeOAuthClient)` produces `OAuthClientConnection`/`OAuthClientEdge`, not the literal `DynamicClientConnection`/`DynamicClientEdge` named in dcr.md. Reusing the existing generic helper (rather than hand-rolling a differently-named Connection/Edge pair solely for this one query) is the smaller, more idiomatic change — flag this as a spec-wording fix for docs/specs/dcr.md in the same doc-fix commit pattern as Part 1's `token_type` deviation.

### 6.4 `pkg/admin/graphql/oauth_client_mutation.go` (new file) — `deleteDynamicClient`

```go
var deleteDynamicClientInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "DeleteDynamicClientInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"clientID": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
	},
})

var deleteDynamicClientPayload = graphql.NewObject(graphql.ObjectConfig{
	Name:   "DeleteDynamicClientPayload",
	Fields: graphql.Fields{"ok": &graphql.Field{Type: graphql.Boolean}},
})

var _ = registerMutationField(
	"deleteDynamicClient",
	&graphql.Field{
		Type: graphql.NewNonNull(deleteDynamicClientPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{Type: graphql.NewNonNull(deleteDynamicClientInput)},
		},
		Resolve: func(p graphql.ResolveParams) (any, error) {
			input := p.Args["input"].(map[string]any)
			clientID := input["clientID"].(string)

			ctx := p.Context
			gqlCtx := GQLContext(ctx)
			if err := gqlCtx.DCRFacade.DeleteClient(ctx, clientID); err != nil {
				return nil, err
			}
			return map[string]any{"ok": true}, nil
		},
	},
)
```

Note this mutation takes the **OAuth `client_id` string** (`dcrc_...`), not a relay global Node id — matches the spec's `DeleteDynamicClientInput { clientID: String! }` exactly (unlike `revokeInitialAccessToken`, which spec defines with a Node `ID!`). `DCRFacade.DeleteClient` resolves it via `Store.DeleteClientByClientID`. "Revokes all outstanding authorizations and tokens issued to it, for every user" (per the mutation's doc comment in dcr.md) is deferred: Part 2 only deletes the persisted client record; cascading token/authorization revocation requires the `oauth.Authorization`/offline-grant revocation machinery that Part 3 wires the resolver into, so add a `// TODO(DCR Part 3+): revoke outstanding authorizations/tokens for this client_id` at the call site rather than silently doing a partial job — **flag this gap explicitly to the user when this plan is reviewed**, since it means a deleted DCR client's previously-issued refresh tokens keep working until Part 3+ adds revocation.

### 6.5 `pkg/admin/graphql/context.go` — extend `Context`

```go
type DynamicClientLoader interface {
	graphqlutil.DataLoaderInterface
}

type DCRFacade interface {
	// Part 1 methods (IAT) already declared here; extend the same interface:
	CreateInitialAccessToken(ctx context.Context, options *dcr.NewInitialAccessTokenOptions) (string, *apimodel.OAuthInitialAccessToken, error)
	RevokeInitialAccessToken(ctx context.Context, id string) error
	ListInitialAccessTokens(ctx context.Context) ([]*apimodel.OAuthInitialAccessToken, error)

	// New in Part 2:
	ListClients(ctx context.Context, pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, *graphqlutil.PageResult, error)
	DeleteClient(ctx context.Context, clientID string) error
}
```

Add `DynamicClients DynamicClientLoader` to the `Context` struct, alongside `InitialAccessTokens`.

### 6.6 `pkg/admin/loader/oauth_client.go` (new file) — mirrors `pkg/admin/loader/resource.go`, backed by `oauthclient.Queries.GetClientModelByID`.

### 6.7 `pkg/admin/facade/dcr.go` — extend (from Part 1)

Add:

```go
func (f *DCRFacade) ListClients(ctx context.Context, pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, *graphqlutil.PageResult, error) {
	return f.Queries.ListClients(ctx, pageArgs)
}

func (f *DCRFacade) DeleteClient(ctx context.Context, clientID string) error {
	return f.Commands.DeleteClient(ctx, clientID)
}
```

`ListClients`/`DeleteClient` are new methods on `oauthclient.Queries`/`oauthclient.Commands`, wrapping `Store.ListClients`/`Store.DeleteClientByClientID` the same way Part 1's IAT methods wrap the IAT store methods. The facade is still named `DCRFacade` for continuity with Part 1's IAT methods, but note it now spans both packages; renaming it (e.g. `OAuthDynamicClientFacade`) is a reasonable tidy-up when CIMD lands.

`Queries.ListClients` must **not** read through the Redis cache introduced in [Part 3 §3.2](2026-08-17-03-client-resolution.md) — an admin looking at the client list expects to see a client registered a second ago. `Commands.DeleteClient` acquires the obligation to invalidate that cache from `DidCommitTx`; see Part 3 §3.2 for why after-commit and not inline. Until Part 3 lands there is no cache to invalidate, so this is a Part 3 commit, not a Part 2 one.

### 6.8 Wiring

`make generate` (wire) + `make export-schemas` (GraphQL SDL + portal gentype) in the same commit that adds §6, same as Part 1 §7.7.

## 7. Test Plan

Unit tests (Convey, in `pkg/lib/dcr` and `pkg/lib/oauthclient` per §4):

- `validate_test.go` — one `Convey` block per rule in §4.3: missing redirect_uris, fragment in redirect_uri, wrong scheme for `web`/`native`, `token_endpoint_auth_method` present, unsupported grant/response type, inconsistent grant/response type pairing, unsupported `application_type`, non-https `logo_uri`/`client_uri`/`tos_uri`/`policy_uri`. Also verify defaulting (`grant_types`/`response_types`/`application_type` omitted → correct defaults).
- `client_model_test.go` — `Client.ToModel` produces the exact zero-valued extension fields; with `client_name` omitted, `Name` is the generated `"Client <clientID>"` fallback while `ClientName` stays `nil` (they diverge, §5.3). Also pin the resolved token lifetimes with `default_client_config` nil: `1800` / `31449600` / `true` / `2592000` (§2.2.1). That assertion is what catches a reintroduction of hand-rolled fallbacks diverging from `OAuthClientConfig.SetDefaults()`.
- `pkg/lib/config/config_test.go` (extend) — `validateTokenLifetime` rejects `default_client_config` with `refresh < access` when both are non-zero, and accepts it when either is zero (§2.2.1).
- `store_client_test.go` — `Store.NewClient` with `ClientName` nil or `""` leaves `ClientName` nil (§5.3); with `ClientName` set to a non-empty string, leaves it untouched.

e2e tests (`e2e/tests/dcr_register.yaml`, `write-e2e-test` skill):

1. `enabled: false` → `POST /oauth2/register` returns `access_denied` (403).
2. `enabled: true`, `initial_access_token_required: true` (default), no `Authorization` header → `invalid_initial_access_token` (401).
3. Valid `iat_fp_...` IAT + `application_type: web` → `201`, response has no `client_secret`, `client_id` starts with `dcrc_`.
4. Valid `iat_tp_...` IAT → registered client is `THIRD_PARTY` (verify via `dynamicClients` Admin API query in the same test).
5. `initial_access_token_required: false`, no IAT → registers as `THIRD_PARTY`; `application_type: m2m` or any other unsupported value → `invalid_client_metadata`.
6. `redirect_uris` omitted → `invalid_client_metadata`.
7. `redirect_uris: ["https://x/callback#frag"]` → `invalid_redirect_uri`.
8. `application_type: native`, `redirect_uris: ["http://localhost/callback"]` → `201`.
9. `application_type: web`, `redirect_uris: ["http://localhost/callback"]` → `invalid_redirect_uri`.
10. `grant_types: ["implicit"]` → `invalid_client_metadata`.
11. `response_types: ["code"]` without `authorization_code` in `grant_types` → `invalid_client_metadata`.
12. Admin API: `dynamicClients` query returns the newly created client with correct fields; `deleteDynamicClient` removes it; a second `deleteDynamicClient` on the same `client_id` errors (not found).
13. Discovery: `/.well-known/openid-configuration` includes `registration_endpoint` when `enabled: true`, omits it when `false`.
14. Rate limit (§5.1.1): with open registration enabled, issue 11 `POST /oauth2/register` requests in quick succession from the same IP → the first 10 succeed (or fail on their own merits) and the 11th returns `x_rate_limited` with HTTP 429. Also assert the 429 body is the RFC 7591 `{error, error_description}` shape, not an HTML or apierrors-shaped body. Check `write-e2e-test` for how existing rate-limit e2e tests reset Redis buckets between cases (or scope the test to its own app id) — a leaked bucket makes neighbouring cases flaky.

## 8. Fixed Behavioral Decisions

- `client_id` uniqueness is enforced at the DB level (`(app_id, client_id)` unique index); a collision (astronomically unlikely given 16-byte entropy) surfaces as a generic 500, not retried — no existing precedent in this codebase retries on ID collision (e.g. `resourcescope.Store.CreateResource` doesn't retry on `ErrResourceDuplicateURI` either).
- `grant_types`/`response_types` are stored and echoed back verbatim as submitted (post-defaulting), but per `oauth.GetAllowedGrantTypes` (`pkg/lib/oauth/grant_type.go:37-67`), `authorization_code`/`refresh_token` are unconditionally in `whitelistedGrantTypes` for every non-M2M client regardless of this field's contents — so this field is validation/display-only for DCR clients in practice, not a runtime enforcement gate. Not something to "fix" here; noted so Part 3 doesn't over-engineer grant-type enforcement that doesn't exist for any other client type either.
- `deleteDynamicClient` in this part only deletes the persisted client row — it does **not** yet revoke outstanding tokens/authorizations (see §6.4). This is a known gap, not a silent omission — call it out to the user explicitly.
- `POST /oauth2/register` is rate limited by two buckets, `oauth.register.per_ip` (10/minute) and `oauth.register.per_project` (1000/hour), both project-configurable under `oauth.dynamic_client_registration.rate_limits` (§2.2.2, §5.1.1). Both are consumed on every attempt including successful ones, both run before the IAT is even parsed. They bound request volume; they are **not** quota protection.
- Also update `docs/specs/rate-limit.md`'s `oauth.register.*` rows to drop the `fixed:` marker (no longer applies) and reflect the new 1000/hour default.
- Quota exhaustion under open registration is an **accepted risk** (§5.1.1). The rate limit plus the per-project usage limit are the two controls Authgear applies everywhere else, and both are in place; no third mechanism is in scope. TTL eviction, bot protection and per-IP quota shares are recorded as options if it ever needs revisiting, not as pending work.
- `oauth.dynamic_client_registration` and its `default_client_config` are both always non-nil once `config.SetFieldDefaults` has run — no `nullable:"true"` tag anywhere in this section (§2.2). `IsEnabled()`/`IsInitialAccessTokenRequired()` stay nil-safe only for pre-`SetFieldDefaults` reads (e.g. a test unmarshaling a YAML snippet directly); `DefaultClientConfig` has its own `SetDefaults()` mirroring `OAuthClientConfig`'s exact constants, so it always resolves to real token-lifetime values.
- The Admin API's `dynamicClients` Connection type is named `OAuthClientConnection`/`OAuthClientEdge` in the actual schema, not the literal `DynamicClientConnection`/`DynamicClientEdge` from dcr.md's SDL (see §6.3).

## 9. Atomic Commit Plan

1. **`doc: Fix DCR registration and dynamic client inconsistencies`** — doc fixes in docs/specs/dcr.md and docs/specs/client.md:
   - **`dynamicClients` Connection shape.** Align the SDL with what `graphqlutil.NewConnectionDef` actually produces (see §6.3). Beyond the type *names* (`OAuthClientConnection`/`OAuthClientEdge`), `relay.ConnectionDefinitions` also yields nullable `edges: [OAuthClientEdge]`, nullable `node: OAuthClient`, and `totalCount: Int` — so every non-null marker in the spec's current `DynamicClientConnection`/`DynamicClientEdge` block is wrong too, not just the names.
   - **`application_type` acceptance under an IAT.** The Configuration section says of `initial_access_token_required`: "When `true` ... all `application_type` values are accepted. When `false`, open registration is permitted but only `application_type: web` and `application_type: native` are accepted." This contradicts §Accepted Client Metadata, which states Authgear accepts only the two standard OIDC DCR values (`web`, `native`) — and the `application_type` table there lists only those two for both IAT types. It is leftover wording from when `m2m`/`confidential` were registerable. The IAT *type* controls first-party vs third-party; it does not widen the accepted `application_type` set. Rewrite to: "When `true`, registration requires a valid IAT in the `Authorization: Bearer` header. When `false`, open registration is permitted and every client is registered as third-party." §4.3 implements the web/native-only reading.
   - **`client_secret` leftovers in §Response.** Two of the three bullets contradict the first: "`client_secret_expires_at: 0` means non-expiring (per RFC 7591 §3.2.1)" and "`client_secret` is returned **once only** and is not recoverable afterwards. The caller must store it securely." Neither can apply when "`client_secret` is not issued in this version. Confidential clients are not supported via DCR." Delete both, or move them under §Future Works alongside RFC 7592. §5.3 issues no secret fields at all.
   - **`deleteDynamicClient` token revocation.** Document the gap: this part deletes only the persisted client record. Until the revocation wiring lands (Part 3+), a deleted DCR client's outstanding authorizations, refresh tokens and access tokens keep working, which the mutation's current doc comment ("revokes all outstanding authorizations and tokens issued to it, for every user") promises.
   - No change is needed for client.md's `OAuthClient` lacking `id`/`implements Node` — the Node id is an Admin GraphQL transport concern, not part of the client model. See the `model.Meta` comment in §6.1.
   - **"Dynamic clients" is used in two conflicting senses** (client.md). Its source list defines item 2 as "**Dynamic clients** — registered at runtime via [DCR]" with CIMD as a separate item 3, but dcr.md's `dynamicClients` query and `deleteDynamicClient` mutation both use "dynamic" to mean "DCR **or** CIMD, i.e. anything not static" — and cimd.md relies on that reading ("CIMD clients are returned by DCR's `dynamicClients` query"). Rename client.md's item 2 to "**DCR clients**" and state explicitly that "dynamic client" means any client not declared in `authgear.yaml`. Without this, the `dynamicClients` query name reads as DCR-only and CIMD's reuse of it looks like a mistake.
   - **"Project defaults" for token lifetimes** (client.md). "Token lifetime fields are populated from `default_client_config` when set, otherwise from the project defaults" implies a project-level token lifetime setting, which does not exist — `authgear.yaml` has token lifetimes only per client. Replace with the built-in defaults and their values: `1800` / `31449600` / `true` / `2592000` (§2.2.1). Optionally also note that the example JSON in that section reflects dcr.md's `default_client_config` snippet, not the defaults.
   - **Registration rate limits** (§5.1.1). Add to dcr.md's §Errors table:

     | `error` value | HTTP status | Meaning |
     |---|---|---|
     | `x_rate_limited` | 429 | Too many registration attempts from this IP, or too many for this project — see [rate-limit.md](./rate-limit.md) |

     …and a short §Rate Limits section under Configuration noting the two limits, their defaults, and that both are project-configurable under `oauth.dynamic_client_registration.rate_limits`. Add to rate-limit.md's main table (no `fixed:` marker — both are configurable):

     | Group | Name | Operation | Rate | Rationales |
     |---|---|---|---|---|
     | **oauth.register** | `oauth.register.per_ip` | Register an OAuth client via DCR | 10/minute | Mitigate resource exhaustion by rapid client registration, especially under open registration where the endpoint is unauthenticated. Mirrors `authentication.signup.per_ip`. Project-configurable. |
     | | `oauth.register.per_project` | | 1000/hour | Bounds how fast a project's DCR client population can grow regardless of source IP. Not a substitute for the `oauth_client_dcr` standing quota. Project-configurable — a legitimate integration with many distinct one-time self-registrations (e.g. MCP-style clients) may need to raise this. |

     Neither appears in rate-limit.md's Fallbacks table (no fallback), matching `oauth.token.general.*`.
2. **`[DCR] Add dynamic_client_registration config`** — §2 (config structs, schema, nil-safe accessors, `DefaultClientConfig.SetDefaults()`, the `validateTokenLifetime` extension from §2.2.1) + config unit tests + `make export-schemas`.
3. **`[DCR] Add dynamic client table and migration`** — §3 migration file only (`_auth_oauth_client`, source-agnostic).
4. **`[DCR] Add pkg/lib/oauthclient for persisted dynamic clients`** — §4.1/§4.2 (`client.go`, `client_config.go`, `client_model.go`, `store_client.go`, `errors.go`, `deps.go`) + unit tests. Source-agnostic and CIMD-ready; no DCR-specific logic in it.
5. **`[DCR] Add DCR registration validation and RegisterClient`** — §4.3 (`pkg/lib/dcr/validate.go`, `RegisterClientOptions`, `Commands.RegisterClient`) + unit tests.
6. **`[DCR] Add rate limits for the client registration endpoint`** — §5.1.1's `pkg/lib/ratelimit` group/name/bucket constants and the two `NewBucketSpecOAuthRegister*` helpers, plus a unit test asserting the configured period/burst. No callers yet, so it lands independently of the handler.
7. **`[DCR] Add POST /oauth2/register endpoint`** — §5 (`handler_register.go` including the rate-limit checks and `checkRateLimit`, `pkg/auth/handler/oauth/register.go`, route wiring, discovery metadata) + regenerated `wire_gen.go`.
8. **`[DCR] Add unified OAuthClient GraphQL type and dynamicClients/deleteDynamicClient Admin API`** — §6 in full + regenerated `wire_gen.go` + regenerated GraphQL schema/gentype artifacts.
9. **`[DCR] Add e2e tests for client registration and Admin API`** — `e2e/tests/dcr_register.yaml`.

## 10. Note on the spec doc fixes

Each part's commit 1 is a doc fix to `docs/specs/`. **These are the implementer's to apply, not open questions.** Every one is either an internal contradiction in a spec, or a spec statement that conflicts with already-shipped code, and each has its exact replacement wording written out at the commit-1 entry of the part that needs it:

| Part | Doc fixes |
|---|---|
| 1 | `token_type` column in dcr.md's IAT `CREATE TABLE`; UC1's `expiresAt` selection moved under `initialAccessToken` |
| 2 | `dynamicClients` Connection type names and nullability; `application_type` acceptance under an IAT; `client_secret` leftovers in §Response; `deleteDynamicClient` token-revocation gap; client.md's "Dynamic clients" used in two senses; client.md's "project defaults" for token lifetimes; the new `x_rate_limited` error row and rate-limit tables |
| 4 | `invalid_resource` → `invalid_target` in api-resource.md; first-party `resource` support and multi-`resource` marked not-yet-implemented in api-resource.md and access-token-audience-binding.md |

The only ones that assert a *product* decision rather than fixing an inconsistency are Part 4's two: deferring first-party and multi-resource support. Both were confirmed as intended, so they are settled — the doc fix records the decision rather than making it.

One coordination point: the specs on this branch are being actively authored, so if spec edits should stay with their author rather than landing from an implementation PR, say so and the commit-1 entries become "spec change required, applied separately" instead. Nothing else in the plans depends on which way that goes.
