# CIMD Part 1 — Config, Client Identifier URL & Discovery Metadata

Spec: [docs/specs/cimd.md — Configuration](../../specs/cimd.md#configuration), [Client ID Format](../../specs/cimd.md#client-id-format), [OIDC Discovery Metadata](../../specs/cimd.md#oidc-discovery-metadata), [Domain Trust](../../specs/cimd.md#domain-trust).

Depends on: DCR Parts 1–5 (all landed — `7b873ba74 Implement DCR #5870`). Specifically it depends on `pkg/lib/oauthclient` (the `Client`/`Store`/`Queries`/`Resolver`/`ClientCache` set), the `_auth_oauth_client` table with its `source`/`last_fetched_at` columns, and `config.OAuthClientApplicationTypeDynamicThirdParty`.

## 1. Goal / Scope

This part adds every piece of CIMD that needs **no network access and no new runtime behavior**:

- the `oauth.client_id_metadata_document` config section,
- the Client Identifier URL format validation (spec §3) as a pure predicate,
- the `allowed_domains` policy matcher,
- turning a valid, allowed CIMD URL into a *dynamic client ID candidate* so the existing resolver read path (`oauthclient.Resolver` → Redis → Postgres) will resolve a persisted CIMD row once one exists,
- `ResolveTokenLifetimes`' CIMD case, so a persisted CIMD row resolves to the right token lifetimes,
- `client_id_metadata_document_supported` in the two discovery documents.

After this part, a CIMD client that somehow already had a row in `_auth_oauth_client` would work end-to-end everywhere (`/oauth2/authorize`, `/oauth2/token`, Admin API, resource indicators) — but nothing yet creates such a row. Fetching is [Part 2](2026-08-28-02-document-fetching.md); creating the row is [Part 3](2026-08-28-03-authorize-time-resolution.md).

Out of scope here: the fetcher, document validation, the authorize-time hook, rate limits, the usage limit, and any UI.

### 1.1 What DCR already built that CIMD reuses unchanged

Worth stating up front, because it removes a large amount of work people expect to find in these plans:

| Already exists | Where |
|---|---|
| `_auth_oauth_client` table with `source` and `last_fetched_at` columns, `(app_id, client_id)` unique index, `(app_id, source)` index | `cmd/authgear/cmd/cmddatabase/migrations/authgear/20260817120001-add_oauth_dynamic_client.sql` |
| `model.OAuthClientSourceCIMD` | `pkg/api/model/oauth_client.go:10` |
| `oauthclient.Client.LastFetchedAt`, `RegisteredAt()` returning `nil` for non-DCR | `pkg/lib/oauthclient/client.go:21,52` |
| `Client.ToModel` already reports `LastFetchedAt` and `RegisteredAt` correctly for CIMD | `pkg/lib/oauthclient/client_model.go:224-225` |
| `config.OAuthClientApplicationTypeDynamicThirdParty` (third-party **and** public) | `pkg/lib/config/oauth.go:49` |
| Redis resolver cache keyed by `sha256(client_id)` — already hashed *specifically* because CIMD client IDs are URLs | `pkg/lib/oauthclient/cache_client.go:898-907` |
| `Store.LockForClientCount(source)` / `CountClientsBySource(source)`, already scoped per source | `pkg/lib/oauthclient/store_client.go:749,762` |
| Admin GraphQL `OAuthClient.source` with a `CIMD` enum value, and `lastFetchedAt` | `pkg/admin/graphql/oauth_client.go:19,125`; `portal/src/graphql/adminapi/schema.graphql:1782,1869` |
| `dynamicClients` query and `deleteDynamicClient` mutation, both source-agnostic | `pkg/admin/graphql/query.go`, `pkg/admin/graphql/oauth_client_mutation.go` |
| `usage.limits` schema enum already accepts `oauth_client_cimd` | `pkg/lib/config/feature_usage.go:8` |
| `access_policy.allow_dynamic_third_party_client_access` and the whole resource-indicator path, gated on `IsDynamicClient() && IsThirdParty()` | `pkg/lib/oauth/handler/handler_authz.go:199,271` |
| Opaque-token-by-default for third-party clients, resource-bound JWT `aud`, `/resolve` opaque-token gate | `pkg/lib/oauth/token_encoding.go`, `pkg/resolver/handler/resolve.go` |

**Consequences.** There is **no migration** in any part of this plan set. There is **no Admin API work** in any part of this plan set — `dynamicClients` starts returning CIMD clients the moment Part 3 writes the first row, with `source: CIMD`, `registeredAt: null` and a real `lastFetchedAt`, exactly as [cimd.md § Reading a CIMD Client's Stored Config](../../specs/cimd.md#reading-a-cimd-clients-stored-config) describes. And there is **no new access-policy flag** — spec § Access Token Audience Binding says so explicitly, and the existing gate is already written in terms of the client *shape* (dynamic + third-party), not the source.

### 1.2 Code style — applies to every part

The Go in these plans is annotated so a reviewer can follow the reasoning **in the plan**. The shipped code should not carry it all: keep comments to what is genuinely non-obvious from the code — a security invariant, a spec citation, a rejected alternative that would otherwise be re-litigated — and drop the rest, matching the comment density of the file being edited. `pkg/lib/dcr/validate.go` is the right calibration for the new `pkg/lib/cimd` files; `pkg/lib/oauthclient/*.go` is heavier than necessary and should not be treated as the target.

Where a plan's comment explains *why* rather than *what*, that reasoning belongs in the commit body or this document, not in a 12-line block above a 4-line function.

**No unused parameters, and no unused struct fields.** Not "for symmetry", not "so the seam is ready", not "the interface already has it". A parameter nothing reads misleads the next reader about what the function depends on, and it survives refactors long after the hypothetical caller it was kept for has been forgotten. If a future control needs an argument, add it then, against a real caller. The same goes for a field wired in by `wire.Struct(new(X), "*")` that no method touches — it makes the dependency graph claim a dependency that does not exist.

## 2. Config Model & Schema

### 2.1 Rename the shared token-lifetimes type first

`config.OAuthDynamicClientRegistrationDefaultClientConfig` is about to be shared by two config sections, and its DCR-specific name would be actively misleading under `client_id_metadata_document.client_config`. Rename it to a source-neutral name **before** adding the CIMD section:

```
OAuthDynamicClientRegistrationDefaultClientConfig  ->  OAuthDynamicClientTokenLifetimesConfig
```

Ten references, all compiler-checked (`grep -rn "OAuthDynamicClientRegistrationDefaultClientConfig" --include=*.go .`):

| File | What changes |
|---|---|
| `pkg/lib/config/oauth_dynamic_client_registration.go:10` | `"$ref": "#/$defs/OAuthDynamicClientTokenLifetimesConfig"` |
| `pkg/lib/config/oauth_dynamic_client_registration.go:26,44,61,77,91` | field type, `GetDefaultClientConfig()` return type, `Schema.Add` name, `type` declaration, `SetDefaults()` receiver |
| `pkg/lib/oauthclient/client_config.go:13,43,60` | `ToClientConfig` param, comment, `ResolveTokenLifetimes` return type |
| `pkg/lib/oauthclient/client_model.go:12` | `ToModel` param |

**What does not change:** the Go field name `DefaultClientConfig`, its `json:"default_client_config,omitempty"` tag, and every field name/tag inside the struct. So no `authgear.yaml` written against DCR changes meaning, and no config migration is needed. The only externally visible effect is the `$defs` key in the exported app-config JSON Schema — which `make export-schemas` writes to `tmp/app-config.schema.json`, an untracked path, so nothing checked in moves.

> Do this as its own commit (commit 1 of §10). A rename mixed into a feature commit is the kind of diff nobody reads.

### 2.2 `pkg/lib/config/oauth.go` — extend `OAuthConfig`

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
		"dynamic_client_registration": { "$ref": "#/$defs/OAuthDynamicClientRegistrationConfig" },
		"client_id_metadata_document": { "$ref": "#/$defs/OAuthClientIDMetadataDocumentConfig" }
	}
}
`)

type OAuthConfig struct {
	Clients                   []OAuthClientConfig                   `json:"clients,omitempty"`
	DynamicClientRegistration *OAuthDynamicClientRegistrationConfig `json:"dynamic_client_registration,omitempty"`
	ClientIDMetadataDocument  *OAuthClientIDMetadataDocumentConfig  `json:"client_id_metadata_document,omitempty"`
}
```

### 2.3 New file `pkg/lib/config/oauth_client_id_metadata_document.go`

```go
package config

var _ = Schema.Add("OAuthClientIDMetadataDocumentConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"allowed_domains": {
			"type": "array",
			"items": { "$ref": "#/$defs/OAuthClientIDMetadataDocumentAllowedDomain" }
		},
		"client_config": { "$ref": "#/$defs/OAuthDynamicClientTokenLifetimesConfig" }
	}
}
`)

// OAuthClientIDMetadataDocumentAllowedDomain constrains an allowed_domains
// entry to the two shapes §3.1 actually implements: an exact hostname, or a
// single leading "*." wildcard. A pattern this schema rejects would
// otherwise be accepted by config validation and then silently never match
// anything -- an admin would believe they had allowlisted a domain while
// every request from it was refused. Failing at config load is the only
// place that mistake is cheap to see.
//
// The regex allows: optional "*." prefix, then one or more dot-separated
// LDH labels. A SINGLE-label host such as "localhost" is deliberately
// ALLOWED: rejecting it would be address policy leaking into shape
// validation, which contradicts D4 below -- whether a host is reachable is
// the address filter's job (Part 2 §2.3), not the allowlist pattern's. It
// also has to be allowed for a test or local-development project to
// allowlist its own document host. This is why urlutil.ValidateHTTPSStrict's
// ErrSingleLabel rule is NOT mirrored here.
var _ = Schema.Add("OAuthClientIDMetadataDocumentAllowedDomain", `
{
	"type": "string",
	"minLength": 1,
	"maxLength": 253,
	"pattern": "^(\\*\\.)?[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*$"
}
`)

// OAuthClientIDMetadataDocumentConfig carries no nullable tag, so
// config.SetFieldDefaults force-allocates it (and ClientConfig below)
// whenever the whole section is absent from authgear.yaml -- exactly like
// OAuthDynamicClientRegistrationConfig. The nil-safe accessors below exist
// only for code paths that read a config before SetFieldDefaults has run
// (e.g. a test that yaml.Unmarshals a snippet directly); at runtime the
// receiver is always non-nil.
type OAuthClientIDMetadataDocumentConfig struct {
	Enabled bool `json:"enabled,omitempty"`
	// AllowedDomains is empty by default, which means "no domain
	// restriction" -- NOT "no domain allowed". See spec § Domain Trust: the
	// MCP use case is specifically about accepting arbitrary client domains,
	// so an empty allowlist must be permissive. Enforced in
	// pkg/lib/oauthclient.IsCIMDClientIDAllowed (§3.1).
	AllowedDomains []string `json:"allowed_domains,omitempty"`
	// ClientConfig is deliberately not named "default_client_config" the way
	// DCR's is. There is no per-client override path for CIMD, planned or
	// otherwise: a CIMD row is a system-maintained mirror of externally
	// hosted data, overwritten on every refetch, so it is not a place an
	// admin can attach anything. This is simply *the* config for every CIMD
	// client. See spec § Configuration.
	ClientConfig *OAuthDynamicClientTokenLifetimesConfig `json:"client_config,omitempty"`
}

func (c *OAuthClientIDMetadataDocumentConfig) IsEnabled() bool {
	return c != nil && c.Enabled
}

func (c *OAuthClientIDMetadataDocumentConfig) GetAllowedDomains() []string {
	if c == nil {
		return nil
	}
	return c.AllowedDomains
}

// GetClientConfig is nil-safe for the same pre-SetFieldDefaults reason as
// IsEnabled. Once defaults have run, ClientConfig is always non-nil with
// real token-lifetime values (OAuthDynamicClientTokenLifetimesConfig's own
// SetDefaults()) -- it is never used to detect "was an override
// configured", because there are no overrides.
func (c *OAuthClientIDMetadataDocumentConfig) GetClientConfig() *OAuthDynamicClientTokenLifetimesConfig {
	if c == nil {
		return nil
	}
	return c.ClientConfig
}
```

No `SetDefaults()` on `OAuthClientIDMetadataDocumentConfig` itself: `Enabled` defaults to Go's `false`, which is the spec default, and `ClientConfig` gets its values from its own (already existing) `SetDefaults()`, force-allocated by `config.SetFieldDefaults`. The example values in the spec's Configuration block (`access_token_lifetime_seconds: 1800`, `refresh_token_lifetime_seconds: 2592000`, `refresh_token_idle_timeout_enabled: true`, `refresh_token_idle_timeout_seconds: 1209600`) are illustrative, **not** CIMD-specific defaults — the actual defaults are the shared `DefaultAccessTokenLifetime` / `DefaultRefreshTokenLifetime` / … constants, identical to DCR and to a static client that omits every lifetime field. Do not hard-code the spec's example numbers anywhere.

Fetch timeout, max document size, redirect count and refetch interval are **not** in this section — spec § Configuration: "Fetch timeout, maximum document size, and cache lifetime are fixed values, not project-configurable." They are constants in `pkg/lib/cimd` ([Part 2](2026-08-28-02-document-fetching.md) §2.1).

The DoS rate limits are **not** in this section either, but for the opposite reason: they are configurable, in the *feature*-config section below, per plan tier. See [Part 4 §1.3](2026-08-28-04-rate-limits.md) — `authgear.yaml` must not be able to raise them, because a CIMD fetch's cost lands on a third party, but the operator needs a lever. This is the deliberate asymmetry with `oauth.dynamic_client_registration.rate_limits`, whose load lands on Authgear's own database and is therefore the tenant's to tune.

### 2.4 `authgear.features.yaml` — the two insecure-fetch escape hatches

> **Read the `update-feature-config` skill before writing this section.** Field-level merge is a hard requirement and `pkg/lib/config/testdata/merge_feature.yaml` coverage is mandatory.
>
> Two other parts extend feature config: [Part 4 §2](2026-08-28-04-rate-limits.md) adds `rate_limits` to this same `oauth.client_id_metadata_document` section, and [Part 5](2026-08-28-05-usage-limit.md) adds `usage.limits.oauth_client_cimd` elsewhere in the file. Part 4's addition means this struct's `Merge` gains a third delegated line, so write `Merge` in a shape that extends cleanly.

Two protections have to be defeatable in a test or local-development project, and **only** there:

| Protection | Where it lives | Why it must be defeatable |
|---|---|---|
| **`https` is required** — for the `client_id` (§3, `ParseCIMDClientID`), for the document's `logo_uri`/`client_uri`/`tos_uri`/`policy_uri` ([Part 2](2026-08-28-02-document-fetching.md) §3 rule 8), and for the logo fetch ([Part 7](2026-08-28-07-logo-proxy.md) §4) | three sites, one rule | A test or local document host has no publicly-trusted certificate, and making the server trust a private CA is an environment-specific, OS-specific problem (`SSL_CERT_FILE` is Unix-only in Go; macOS ignores it) |
| the resolved address must be publicly routable | [Part 2](2026-08-28-02-document-fetching.md) §2.3 | A test or local document host is on loopback or a container-network private IP |

`insecure_http_allowed` covers **every** place CIMD would otherwise require `https`, not just the `client_id`. An earlier draft scoped it to the `client_id` alone; that left [Part 7](2026-08-28-07-logo-proxy.md)'s logo-proxy e2e test unwritable, because the e2e document host is plain HTTP and so cannot serve an `https` `logo_uri` either. One flag with one meaning — "plaintext `http` is acceptable in this project" — is also easier to reason about than a per-field set. It does **not** relax anything other than the scheme: every other rule on those fields (absolute URI, no fragment, and the `client_id`'s path/userinfo/dot-segment rules) still applies.

**These are feature-config fields, not `DEV_MODE`, and not `authgear.yaml`.** The reasoning, since this is the most security-sensitive design choice in the whole plan set:

- **Not `authgear.yaml`.** A project admin must never be able to turn off an SSRF control. `authgear.yaml` is tenant-editable through the portal.
- **Not `DEV_MODE`** (which is what spec §8.6's dev exception nominally asks for, and what an earlier draft of these plans specified). Three reasons it does not work:
  1. **It is process-wide, so it cannot express "this project may, that project may not."** The e2e suite needs *both* postures in one run — a permissive project to exercise the fetch path end to end, and a strict project to assert that `http://` and private addresses are actually refused. A single global switch makes the second test impossible, which means the enforcement path would ship with no end-to-end coverage at all.
  2. **e2e cannot turn it on.** `e2e/.env:1` sets `DEV_MODE=true` and then line 74 overrides it to `DEV_MODE=false # required to send email` (`godotenv` is last-wins — verify with `godotenv.Read`). `pkg/lib/messaging/sender.go:100,233,382` and `pkg/lib/usage/usage_alert_email_service.go:53` short-circuit delivery under `DevMode`, so e2e genuinely requires it off. A `DEV_MODE`-gated exception would be dead code in e2e.
  3. It would couple an OAuth SSRF policy to an unrelated email-suppression switch, so neither can move without considering the other.
- **Feature config is operator-only.** `authgear.features.yaml` is written through the **Site Admin** surface (`pkg/siteadmin/service/feature_config.go`), never the tenant portal, and merges `code defaults ← cluster ← plan ← app override` (see `pkg/api/siteadmin/gen.go:191`). So the trust level is the same as `DEV_MODE`'s — whoever deploys Authgear — while the granularity is per project.
- **There is direct precedent for exactly this shape:** `config.RateLimitsFeatureConfig.Disabled` (`pkg/lib/config/feature_rate_limits.go:13-17`) is a plan-tier feature-config boolean that disables a protective control wholesale for one project on a shared process. Same conceptual mismatch (a project-scoped flag for a process-scoped capability), same trust boundary, already accepted in this codebase.

Extend the existing `OAuthFeatureConfig` in `pkg/lib/config/feature_oauth.go`. `FeatureConfig.OAuth` already exists (`pkg/lib/config/feature.go:46`) and is already fanned out by wire (`pkg/lib/deps/deps_config.go:74`), so `*config.OAuthFeatureConfig` is injectable with no wire change.

```go
var _ = FeatureConfigSchema.Add("OAuthFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"client": { "$ref": "#/$defs/OAuthClientFeatureConfig" },
		"client_id_metadata_document": { "$ref": "#/$defs/OAuthClientIDMetadataDocumentFeatureConfig" }
	}
}
`)

type OAuthFeatureConfig struct {
	Client                   *OAuthClientFeatureConfig                   `json:"client,omitempty"`
	ClientIDMetadataDocument *OAuthClientIDMetadataDocumentFeatureConfig `json:"client_id_metadata_document,omitempty"`
}

var _ = FeatureConfigSchema.Add("OAuthClientIDMetadataDocumentFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"insecure_http_allowed": { "type": "boolean" },
		"insecure_fetch_address_allowed": { "type": "boolean" }
	}
}
`)

// OAuthClientIDMetadataDocumentFeatureConfig defeats the CIMD fetch's
// transport protections, for test and local-development projects only.
// Never set on a project serving real traffic, and never at the cluster or
// plan layer -- app-specific override only.
type OAuthClientIDMetadataDocumentFeatureConfig struct {
	// InsecureHTTPAllowed permits http:// wherever CIMD requires https: the
	// client_id, the document's uri fields, and the logo fetch. Scheme only.
	InsecureHTTPAllowed *bool `json:"insecure_http_allowed,omitempty"`
	// InsecureFetchAddressAllowed permits connecting to a
	// non-publicly-routable address, including 169.254.169.254.
	InsecureFetchAddressAllowed *bool `json:"insecure_fetch_address_allowed,omitempty"`
}

func (c *OAuthClientIDMetadataDocumentFeatureConfig) SetDefaults() {
	if c.InsecureHTTPAllowed == nil {
		c.InsecureHTTPAllowed = new(false)
	}
	if c.InsecureFetchAddressAllowed == nil {
		c.InsecureFetchAddressAllowed = new(false)
	}
}

func (c *OAuthFeatureConfig) GetClientIDMetadataDocument() *OAuthClientIDMetadataDocumentFeatureConfig {
	if c == nil {
		return nil
	}
	return c.ClientIDMetadataDocument
}

func (c *OAuthClientIDMetadataDocumentFeatureConfig) IsInsecureHTTPAllowed() bool {
	return c != nil && c.InsecureHTTPAllowed != nil && *c.InsecureHTTPAllowed
}

func (c *OAuthClientIDMetadataDocumentFeatureConfig) IsInsecureFetchAddressAllowed() bool {
	return c != nil && c.InsecureFetchAddressAllowed != nil && *c.InsecureFetchAddressAllowed
}
```

Merge, field-level, delegating per sub-section the way `OAuthFeatureConfig.Merge` (`feature_oauth.go:18-31`) already does:

```go
func (c *OAuthFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.OAuth == nil {
		return c
	}
	var merged *OAuthFeatureConfig = c
	if merged == nil {
		merged = &OAuthFeatureConfig{}
	}
	merged.Client = merged.Client.Merge(layer.OAuth.Client)
	merged.ClientIDMetadataDocument = merged.ClientIDMetadataDocument.Merge(layer.OAuth.ClientIDMetadataDocument)
	return merged
}

func (c *OAuthClientIDMetadataDocumentFeatureConfig) Merge(layer *OAuthClientIDMetadataDocumentFeatureConfig) *OAuthClientIDMetadataDocumentFeatureConfig {
	if c == nil && layer == nil {
		return nil
	}
	if c == nil {
		return layer
	}
	if layer == nil {
		return c
	}
	if layer.InsecureHTTPAllowed != nil {
		c.InsecureHTTPAllowed = layer.InsecureHTTPAllowed
	}
	if layer.InsecureFetchAddressAllowed != nil {
		c.InsecureFetchAddressAllowed = layer.InsecureFetchAddressAllowed
	}
	return c
}
```

**`*bool`, not `bool`.** With field-level merge, `nil` is the only way to distinguish "this layer said nothing" from "this layer said false", and that distinction is what lets a higher layer force a flag back **off** — which a plain `bool` cannot express. It also survives the pipeline: `AuthgearFeatureYAMLDescriptor.viewEffectiveResource` (`pkg/lib/config/configsource/resources.go:695-718`) parses every layer with `ParseFeatureConfigWithoutDefaults` (no defaults), merges them, then `yaml.Marshal`s the result and re-parses it with `ParseFeatureConfig` (which applies `SetFieldDefaults`). A plain `bool` with `omitempty` would lose an explicit `false` at the marshal step; a plain `bool` without `omitempty` would emit `false` from every layer and make the merge's nil-check meaningless. `*bool` + `omitempty` is correct at every stage.

This also matches the siblings in the same struct — `OAuthClientFeatureConfig.CustomUIEnabled` and `App2AppEnabled` are both `*bool` with exactly this merge shape. `RateLimitsFeatureConfig.Disabled` is a plain `bool` *without* `omitempty`, but its `Merge` replaces the whole section (`return layer.RateLimits`) rather than merging fields, so it needs `false` to survive the marshal — a different, internally consistent design that does not apply here.

**Two flags, not one.** They gate two checks in two unrelated layers — a pure string predicate with no I/O (§3) and a socket-level address policy (Part 2 §2.3). One boolean threaded to both would mean a reader at either site cannot tell what else it switches off, and it would remove the ability to assert each protection independently in e2e. They are nested under one section so they still read as a pair.

**`DEV_MODE` is not also consulted.** These flags replace the `DEV_MODE` gate an earlier draft put on the address filter, rather than being OR'd with it. Two overlapping exceptions with different scopes and different owners is how a hole gets missed; one mechanism is auditable. `SafeDialer` consequently loses its `config.DevMode` field entirely (Part 2 §2.4). Spec §8.6's dev exception is still implemented — just gated on feature config instead, which is a deviation to document (§9).

**Observability, required.** Every fetch that actually uses either relaxation logs at `Warn` with the `app_id`, the flag(s) in effect, and the target host. Volume is bounded by the [Part 4 rate limits](2026-08-28-04-rate-limits.md) to ≤10/min/project, so this cannot flood. Without it, a flag left on in a deployed project is invisible. Details in Part 2 §2.4.

### 2.5 `make export-schemas`

Run it after §2.1, §2.3 and §2.4. It regenerates `tmp/app-config.schema.json` and `tmp/secrets-config.schema.json` (both untracked) plus the two `schema.graphql` files and portal `gentype` (unchanged by this part — no GraphQL change). Verify `git status --porcelain` is clean afterwards, per the root `Makefile`'s `check-tidy` gate.

Note `pkg/lib/config/testdata/parse_feature_tests.yaml` pins some feature-config shapes; re-run `go test ./pkg/lib/config/...` and update any fixture that enumerates `OAuthFeatureConfig`'s properties.

## 3. Client Identifier URL — `pkg/lib/oauthclient/client_id.go`

The pure `client_id`-shape predicates live next to `IsDCRClientID` in `pkg/lib/oauthclient`, **not** in `pkg/lib/cimd`: `oauthclient.Resolver` must consult them, and `pkg/lib/cimd` imports `pkg/lib/oauthclient` (mirroring `pkg/lib/dcr`'s direction), so putting them in `cimd` would invert that and cycle. Everything that is not a pure shape question — fetching, document validation, resolution orchestration — does live in `pkg/lib/cimd`.

```go
var (
	ErrCIMDClientIDNotHTTPS      = errors.New("oauthclient: cimd client_id must use the https scheme")
	ErrCIMDClientIDNoPath        = errors.New("oauthclient: cimd client_id must contain a path component")
	ErrCIMDClientIDHasFragment   = errors.New("oauthclient: cimd client_id must not contain a fragment")
	ErrCIMDClientIDHasUserInfo   = errors.New("oauthclient: cimd client_id must not contain userinfo")
	ErrCIMDClientIDHasDotSegment = errors.New("oauthclient: cimd client_id must not contain '.' or '..' path segments")
	ErrCIMDClientIDBadHost       = errors.New("oauthclient: cimd client_id has an invalid or missing host")
	ErrCIMDClientIDTooLong       = errors.New("oauthclient: cimd client_id is too long")
)

const MaxCIMDClientIDLength = urlutil.MaxURLLength

// ParseCIMDClientID implements docs/specs/cimd.md § Client ID Format. It
// performs no network access: a malformed client_id must never reach the
// fetch path. The returned URL is what Part 2's fetcher requests, unmodified
// -- no normalization, because the document's own client_id must equal this
// string byte-for-byte.
func ParseCIMDClientID(clientID string, allowInsecureHTTP bool) (*url.URL, error) {
	if len(clientID) > MaxCIMDClientIDLength {
		return nil, ErrCIMDClientIDTooLong
	}

	u, err := url.Parse(clientID)
	if err != nil {
		return nil, err
	}

	switch {
	case strings.EqualFold(u.Scheme, "https"):
	case allowInsecureHTTP && strings.EqualFold(u.Scheme, "http"):
	default:
		return nil, ErrCIMDClientIDNotHTTPS
	}
	if u.User != nil {
		return nil, ErrCIMDClientIDHasUserInfo
	}
	// RawFragment catches a bare trailing "#", which Fragment alone does not.
	if u.Fragment != "" || u.RawFragment != "" {
		return nil, ErrCIMDClientIDHasFragment
	}
	if u.Host == "" || u.Hostname() == "" {
		return nil, ErrCIMDClientIDBadHost
	}
	if u.Path == "" {
		return nil, ErrCIMDClientIDNoPath
	}
	// u.Path is the DECODED path, so this also catches "%2e%2e".
	for _, seg := range strings.Split(u.Path, "/") {
		if seg == "." || seg == ".." {
			return nil, ErrCIMDClientIDHasDotSegment
		}
	}

	return u, nil
}

func IsCIMDClientID(clientID string, allowInsecureHTTP bool) bool {
	_, err := ParseCIMDClientID(clientID, allowInsecureHTTP)
	return err == nil
}
```

`allowInsecureHTTP` is an explicit parameter rather than a config read inside the function, so the function stays pure and every call site is compiler-forced to state its posture. Callers with no business relaxing anything pass `false` as a literal.

**Plain `errors.New` sentinels, not `apierrors`.** This mirrors `pkg/lib/dcr/validate.go`, which uses plain sentinels for the same reason: these errors are mapped by their caller to an OAuth protocol error, and never rendered as an Authgear API error. For CIMD the point is sharper — spec § Authgear as an SSRF/Probing Oracle requires that no fetch- or shape-related failure be distinguishable in a response, so a `reason` string and HTTP status attached to these errors would be a liability, not a feature: any middleware that rendered one would leak exactly what must not leak. `apierrors` is used in `pkg/lib/oauthclient/errors.go` because those errors *do* surface through the Admin API (`ErrDynamicClientNotFound` → GraphQL `NotFound`); [Part 3](2026-08-28-03-authorize-time-resolution.md)'s `ErrDynamicClientSourceConflict` is an `apierror` for the same reason, being an internal invariant violation rather than a client-visible outcome.

Deliberately **not** rejected:

- **A bare-root path.** `https://example.com/` is accepted; only an empty path (`https://example.com`) is refused. See §7 D2.
- **A query string.** Spec §3 does not forbid one, and spec § Denial of Service depends on it being significant ("caching is keyed on the exact URL string, not the hostname"). It is part of the identity and is fetched verbatim.
- **An IP-literal host.** `urlutil.ValidateHTTPSStrict` is *not* reused: spec §3 enumerates five rules and an IP-literal host is not among them, while `ValidateHTTPSStrict` additionally blocks `example.com` and friends via its `strictBlocked` list, which would reject the spec's own examples. Non-publicly-routable targets are blocked where the spec puts that control — the resolved-address filter in [Part 2](2026-08-28-02-document-fetching.md) §2.3, which also catches a normal-looking hostname pointing at `169.254.169.254`. Do not "harden" this function by calling `ValidateHTTPSStrict`.
- **Trailing-dot hosts** (`https://example.com./x`). Left to the address filter and TLS verification.

### 3.1 `allowed_domains` matching — same file

```go
// IsCIMDClientIDAllowed reports whether u's host satisfies the project's
// allowed_domains policy (spec § Domain Trust). An empty policy allows
// everything. Matching is on Hostname() only -- never host:port -- and is
// case-insensitive.
//
// A leading "*." matches exactly ONE label, per the RFC 6125 / TLS
// certificate convention: "*.example.com" matches "a.example.com" but not
// "a.b.example.com" and not "example.com".
func IsCIMDClientIDAllowed(allowedDomains []string, u *url.URL) bool {
	if len(allowedDomains) == 0 {
		return true
	}
	host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
	for _, pattern := range allowedDomains {
		p := strings.ToLower(pattern)
		if suffix, ok := strings.CutPrefix(p, "*."); ok {
			rest, found := strings.CutSuffix(host, "."+suffix)
			if found && rest != "" && !strings.Contains(rest, ".") {
				return true
			}
			continue
		}
		if host == p {
			return true
		}
	}
	return false
}
```

**One label, not any depth.** An earlier draft had `*.example.com` span dots. That was the wrong default: the convention people's intuition is calibrated on is TLS certificates (RFC 6125 §6.4.3), where a wildcard matches a single label, and DNS wildcard records behave the same way. Suffix-at-any-depth matching exists too — cookie `Domain` attributes and CSP host-sources work that way — but for a *trust* allowlist the single-label reading is both the more expected and the safer one: an organisation that has delegated `dev.example.com` to a team would be surprised to find `evil.dev.example.com` admitted by `*.example.com`.

Consequence: an all-depths match is not expressible in one entry. That is acceptable — `allowed_domains` is empty for the MCP use case (spec § Domain Trust: "This is likely to matter more for higher-trust deployments"), and a higher-trust deployment knows its exact hosts. An admin who wants the apex as well as one level lists both `example.com` and `*.example.com`.

`path.Match` is not used: it would give meaning to `?`, `[a-z]` and `\` in a config field where those characters have none, and its `*` spans dots.

## 4. `oauthclient.Resolver` — the CIMD candidate branch

`pkg/lib/oauthclient/resolver.go` already has the extension point, with the intended shape in its comment (`resolver.go:271-285`). Fill it in:

```go
// isDynamicClientIDCandidate answers only "is a persisted-row lookup
// warranted for this string?". It applies no trust policy: allowed_domains
// and insecure_http_allowed gate FETCHING, not reading (see below), so http
// is accepted here unconditionally.
func (r *Resolver) isDynamicClientIDCandidate(clientID string) bool {
	if IsDCRClientID(clientID) {
		return true
	}
	if r.OAuthConfig.ClientIDMetadataDocument.IsEnabled() {
		return IsCIMDClientID(clientID, true)
	}
	return false
}
```

`Resolver` needs no new field — no feature config, no `allowed_domains`. `ResolveClient` itself is unchanged; its existing body (static lookup → tester → candidate guard → `Queries.GetClientConfigByClientID`) is already source-agnostic.

### 4.1 Trust policy gates fetching, not reading

An earlier draft of this plan enforced `allowed_domains` (and the `insecure_http_allowed` scheme rule) here, on every read path, so that tightening either took effect immediately everywhere. **That was wrong**, and the reason is worth stating because it is a product rule, not an implementation detail:

> Removing a domain from `allowed_domains` must not break clients that are already working.

An admin editing an allowlist is making an onboarding decision. Breaking every existing user's active session with an already-authorized client — potentially mid-flow, and at `/oauth2/token` where nothing can be re-fetched — is a far worse outcome than the domain remaining operational until the admin explicitly removes it. This also matches how spec § Domain Trust words the control: "which domains it **trusts as `client_id`s**", i.e. which client IDs it will *accept*, not which persisted clients may continue to operate.

So the division is:

| Check | Read path (`ResolveClient`, every endpoint) | Fetch path (`EnsureClientResolved`, `/oauth2/authorize` only) |
|---|---|---|
| `enabled` | **yes** — a feature switch; disabling CIMD stops CIMD | yes |
| `client_id` shape (§3, five rules) | yes, with `http` accepted | yes |
| `allowed_domains` | no | **yes** |
| `insecure_http_allowed` | no | **yes** |

`enabled` stays on the read path deliberately, and this is the one asymmetry: it is a feature kill switch, not a trust policy, and an operator who sets `enabled: false` expects CIMD to stop working, not to keep serving existing clients. If that is not wanted, move it to the fetch path too — but then disabling CIMD leaves it partly on, which is worse.

What this buys, and what it costs:

- An existing client keeps working after its domain is removed from the allowlist — including across refetches, since a refused fetch serves the stale record ([Part 3](2026-08-28-03-authorize-time-resolution.md) §5). Its metadata therefore **freezes** at whatever was last fetched: the domain's operator can no longer change its `redirect_uris` through Authgear. That is the right failure direction for a revoked-trust scenario.
- No new client from that domain can ever be created, because creation requires a fetch.
- `deleteDynamicClient` + removal from `allowed_domains` is therefore a genuinely **durable** ban: the row is gone and cannot be recreated. Removal alone is not — it is "no new clients", which is what an allowlist should mean.
- An admin who does want an existing client stopped now uses `deleteDynamicClient`. Recommend saying so in the spec (§9).

## 5. `ResolveTokenLifetimes` — the CIMD case

`pkg/lib/oauthclient/client_config.go:60-73` has a `default: return nil` branch with a comment saying "CIMD will add its own case when cimd.md ships". Add it:

```go
func ResolveTokenLifetimes(oauthConfig *config.OAuthConfig, source model.OAuthClientSource) *config.OAuthDynamicClientTokenLifetimesConfig {
	switch source {
	case model.OAuthClientSourceDCR:
		return oauthConfig.DynamicClientRegistration.GetDefaultClientConfig()
	case model.OAuthClientSourceCIMD:
		return oauthConfig.ClientIDMetadataDocument.GetClientConfig()
	default:
		return nil
	}
}
```

This single line is what makes `oauth.client_id_metadata_document.client_config` take effect on every read path at once, because every caller already routes through it: `GetClientConfigByClientID` (the runtime resolver), `GetClientModelByID`, `GetManyClientModels`, `ListClients` (Admin API). Note the property `queries.go:354-362` already documents — the *config* is re-read on every call while the *row* is cached, so an admin edit to `client_config` takes effect immediately, without waiting for a refetch or a cache expiry.

The `default: return nil` branch stays, and `ToClientConfig`'s `if defaults != nil` guard stays with it: `model.OAuthClientSourceStatic` is a legal `OAuthClientSource` value that this function must not panic on, even though no static client ever reaches a `Client` row.

`ToClientConfig` needs no other change. Trace it for a CIMD row to confirm — `Kind` is always `THIRD_PARTY` for CIMD (spec § Mapping to the Unified Client Model), so:

- `appType` → `OAuthClientApplicationTypeDynamicThirdParty` (third-party, **public**, no client-credentials, no full-access scope, PII allowed in ID token). Correct per spec § Client Authentication: "CIMD clients in v1 are always public".
- `IsDynamic: true` → the resource-indicator and access-policy path in `handler_authz.go:199` accepts it with no change.
- `IssueJWTAccessToken: false`, and every Authgear extension field at its zero value → matches spec § Mapping's "All Authgear-only extension fields ... are fixed at their zero values".
- `PostLogoutRedirectURIs` is not set by `ToClientConfig` at all, so it is `nil` — spec says `[]`. `ToModel` already writes a literal `[]string{}` for the API surface; the runtime `nil` and `[]` are equivalent for every consumer (`slices.Contains` over nil is false). No change.

### 5.1 One field to add: `Source` on the synthesized client config

[Part 6](2026-08-28-06-consent-and-authorized-apps.md) needs to know, from a `*config.OAuthClientConfig` alone, whether the client is CIMD — to decide whether to render the `client_id` hostname on the consent screen (spec § Phishing Mitigation). Add it here, where `ToClientConfig` already has the row in hand, rather than making Part 6 re-derive it by string-sniffing the `client_id`:

```go
// pkg/lib/config/oauth.go, next to IsDynamic

	// DynamicSource is set only by oauthclient.Client.ToClientConfig, from
	// the persisted row's own source column. It is "" for every statically
	// configured client, including one whose client_id happens to be an
	// https:// URL (spec § Client ID Format's "pre-registering Client
	// Identifier URLs" pattern) -- such a client is an ordinary static
	// client and must not pick up any CIMD-specific behavior. json:"-"
	// because it is synthetic and never present in authgear.yaml.
	DynamicSource model.OAuthClientSource `json:"-"`

func (c *OAuthClientConfig) IsCIMDClient() bool {
	return c.DynamicSource == model.OAuthClientSourceCIMD
}
```

and in `ToClientConfig`: `DynamicSource: c.Source,`.

`pkg/lib/config` already imports `pkg/api/model` (`feature_usage.go:3`), so this adds no new dependency edge.

## 6. OIDC Discovery Metadata

`pkg/lib/oauth/metadata.go`, `MetadataProvider.PopulateMetadata`, immediately after the existing `registration_endpoint` block:

```go
	if p.OAuthConfig.DynamicClientRegistration.IsEnabled() {
		meta["registration_endpoint"] = p.Endpoints.RegistrationEndpointURL().String()
	}
	if p.OAuthConfig.ClientIDMetadataDocument.IsEnabled() {
		// draft-ietf-oauth-client-id-metadata-document-02 §6. The property is
		// OPTIONAL for an AS to include; Authgear includes it whenever CIMD
		// is enabled, and omits it entirely otherwise -- an explicit
		// "false" would be indistinguishable from absent to a compliant
		// client, and per the MCP authorization spec's fixed priority order
		// (spec § UC1) absent is exactly what makes an MCP client fall back
		// to DCR.
		meta["client_id_metadata_document_supported"] = true
	}
```

`oauth.MetadataProvider` already holds `OAuthConfig`, so there is no wiring change. Both discovery documents are covered by one edit: `pkg/auth/handler/oauth/metadata.go` fans the same `[]MetadataProvider` out to `/.well-known/openid-configuration` and `/.well-known/oauth-authorization-server` (`pkg/auth/deps.go:64-65`).

Do **not** add anything to `token_endpoint_auth_methods_supported` — it already advertises `"none"`, which is the only method a CIMD client may use.

## 7. Fixed Behavioral Decisions

- **D1. The CIMD `client_id`-shape predicate lives in `pkg/lib/oauthclient`, not `pkg/lib/cimd`.** Required by the import direction: `oauthclient.Resolver` consults it, and `pkg/lib/cimd` imports `oauthclient` (mirroring `pkg/lib/dcr`). Only pure `client_id`-shape questions go there.
- **D2. `https://host/` (bare-root path) is accepted; only an empty path is rejected.** Spec §3 requires the URL to "contain a path component (i.e. not just `https://host`)", and `/` *is* a path component. An earlier draft rejected it on the grounds that RFC 3986 normalization makes `https://host` and `https://host/` the same resource — but the rule is about the `client_id` **string**, which is matched byte-for-byte against the document's own `client_id`, so the two are distinct client IDs regardless. Hosting a metadata document at the site root is unusual but perfectly valid, there is no security consequence to allowing it (nothing in the byte-for-byte match or the exact redirect-URI match depends on path depth), and rejecting it would break such a client behind CIMD's deliberately uninformative error. Where the spec is ambiguous, be permissive.
- **D3. A query string is allowed and is part of the client's identity.** Spec §3 does not forbid it, and spec § Denial of Service depends on it being significant ("caching is keyed on the exact URL string, not the hostname").
- **D4. `urlutil.ValidateHTTPSStrict` is not reused for `client_id` validation.** Its `strictBlocked` list (`example.com`, `test`, `localhost`, …) and its IP-host and single-label rules go well beyond spec §3 and would reject the spec's own examples. Non-publicly-routable targets are blocked at the resolved-address level in Part 2, which is strictly stronger anyway (it also catches a normal-looking hostname pointing at `169.254.169.254`).
- **D5. `allowed_domains` and `insecure_http_allowed` gate fetching, not reading.** Removing a domain from the allowlist must not break clients that already work: an allowlist is an onboarding control (spec § Domain Trust: "which domains it trusts **as `client_id`s**"), and breaking live sessions — including at `/oauth2/token`, which never fetches — is a worse outcome than the domain staying operational until an admin deletes the client. See §4.1 for the full table. `enabled` is the one exception and stays on the read path: it is a feature kill switch, and an operator disabling CIMD expects CIMD to stop.
- **D5a. `deleteDynamicClient` plus removal from `allowed_domains` is a durable ban**; removal alone is "no new clients from this domain", and an existing client's metadata freezes at its last successful fetch (since a refused refetch serves the stale record, Part 3 §5).
- **D6. `allowed_domains` matches `Hostname()` only, case-insensitively; a leading `*.` matches exactly ONE label; the apex is not matched by its own wildcard.** Single-label matching follows the RFC 6125 / TLS-certificate and DNS-wildcard convention, which is what intuition is calibrated on — not the suffix-at-any-depth behavior of cookie `Domain` or CSP host-sources. For a trust allowlist the single-label reading is both more expected and safer: `*.example.com` admitting `evil.dev.example.com` would surprise an organisation that has delegated `dev.example.com`. An all-depths match is therefore not expressible in one entry; that is accepted (§3.1). Malformed patterns fail JSON Schema validation at config load rather than silently never matching.
- **D7. `client_id_metadata_document_supported` is emitted only when enabled, never as `false`.** Absence is what makes a compliant MCP client fall back to DCR (spec § UC1); an explicit `false` conveys nothing extra.
- **D8. `client_config` is not named `default_client_config` and has no override path.** Spec § Configuration states the reasoning; the plan does not build an override surface, and reviewers should reject one added later without a spec change.
- **D9. The shared token-lifetimes type is renamed rather than reused under its DCR name.** Field names and JSON tags are untouched, so no `authgear.yaml` changes meaning; only an untracked generated schema file moves.
- **D10. The two insecure-fetch escape hatches are feature-config fields, not `DEV_MODE` and not `authgear.yaml`.** §2.4 gives the full reasoning: `authgear.yaml` is tenant-editable, `DEV_MODE` is process-wide (so it cannot express both postures at once, which the enforcement tests require) and is pinned `false` in e2e by the email-suppression requirement. Feature config is Site-Admin-only and per project. Precedent: `RateLimitsFeatureConfig.Disabled`.
- **D11. Two flags, not one**, because they gate two checks in two unrelated layers and because each protection must be assertable independently.
- **D12. `DEV_MODE` is not consulted at all by the CIMD fetch path.** The feature flags replace it rather than being OR'd with it; `SafeDialer` loses its `DevMode` field. One auditable mechanism.
- **D13. `insecure_http_allowed` is not restricted to loopback hosts.** Restricting the scheme rule by host would duplicate, over an unresolved string, a judgement the address filter makes correctly over resolved addresses — and it would break the e2e case, whose document host is a private container address rather than loopback. The two flags compose instead.
- **D13a. `insecure_http_allowed` covers every `https` requirement in CIMD**, not just the `client_id`: the document's URI fields and the logo fetch too. Scoping it to the `client_id` alone left Part 7's logo e2e test unwritable. It relaxes only the scheme.
- **D14. `OAuthClientIDMetadataDocumentFeatureConfig`'s fields are `*bool`, not `bool`.** Field-level merge needs `nil` to mean "this layer said nothing", which is also the only way a higher layer can force a flag back off; and `*bool` + `omitempty` is the only shape that survives the layer pipeline, which marshals the merged config back to YAML before applying defaults (`configsource/resources.go:695-718`). Matches the `*bool` siblings in the same struct. `RateLimitsFeatureConfig.Disabled` is a plain `bool` because its merge replaces the whole section, a different and internally consistent design.
- **D15. `allowed_domains` accepts single-label patterns** (e.g. `localhost`). Excluding them was address policy leaking into shape validation, contradicting D4, and it would prevent a test project from allowlisting its own document host.
- **D16. Every fetch using either relaxation logs at `Warn` with `app_id`, flags and target host.** A flag left on in a deployed project must not be invisible. Bounded by the Part 4 rate limits.
- **D17. The `ErrCIMDClientID*` sentinels are plain `errors.New`, not `apierrors`** — but the errors `cimd.Service` *returns* are `apierrors.Kind`s ([Part 3](2026-08-28-03-authorize-time-resolution.md) §3.2). The split is the point: a Kind is designed to be classified and rendered, which is right for a boundary outcome the handler must distinguish, and wrong for a per-rule shape error that must never be distinguishable in a response (spec § Authgear as an SSRF/Probing Oracle). These sentinels are collapsed into one Kind before leaving `pkg/lib/cimd`, so attaching a Kind to each would only create per-mode reasons that must never escape. They are also what the unit tests compare. Mirrors `pkg/lib/dcr/validate.go`; `pkg/lib/oauthclient/errors.go` uses `apierrors` because those errors genuinely surface through the Admin API.

## 8. Test Plan

**Unit — `pkg/lib/oauthclient/client_id_test.go` (extend the existing file)**

`ParseCIMDClientID`, table-driven, one case per spec §3 rule plus the decisions above:

| Input | Expect |
|---|---|
| `https://mcp-client.example.com/oauth/client-metadata.json` | ok |
| `https://mcp-client.example.com/a?b=c` | ok (query allowed, D3) |
| `HTTPS://Example.com/x` | ok (scheme case-insensitive), and the returned URL's `String()` is unchanged |
| `http://example.com/x` | `ErrCIMDClientIDNotHTTPS` |
| `https://example.com` | `ErrCIMDClientIDNoPath` |
| `https://example.com/` | **ok** (bare-root path allowed, D2) |
| `https://example.com/x#frag` | `ErrCIMDClientIDHasFragment` |
| `https://example.com/x#` | `ErrCIMDClientIDHasFragment` (bare `#`) |
| `https://user:pw@example.com/x` | `ErrCIMDClientIDHasUserInfo` |
| `https://example.com/a/../b` | `ErrCIMDClientIDHasDotSegment` |
| `https://example.com/a/./b` | `ErrCIMDClientIDHasDotSegment` |
| `https://example.com/a/%2e%2e/b` | `ErrCIMDClientIDHasDotSegment` (percent-encoded, caught via decoded `Path`) |
| `https:///x` (empty host) | `ErrCIMDClientIDBadHost` |
| `"https://example.com/" + strings.Repeat("a", 2001)` | `ErrCIMDClientIDTooLong` |
| `dcrc_abc`, `my-static-client`, `""` | error (not a CIMD shape) — asserts no collision with the other two client-id shapes |
| `https://127.0.0.1/x`, `https://localhost/x` | **ok** — asserts D4: shape validation does not do address policy |

Every case above is run with `allowInsecureHTTP: false`. Then the same table with `allowInsecureHTTP: true`, asserting only the scheme rule changed:

| Input | `allowInsecureHTTP: false` | `allowInsecureHTTP: true` |
|---|---|---|
| `http://example.com/x` | `ErrCIMDClientIDNotHTTPS` | **ok** |
| `http://localhost:2727/x.json` | `ErrCIMDClientIDNotHTTPS` | **ok** |
| `http://10.0.0.5:2727/x.json` | `ErrCIMDClientIDNotHTTPS` | **ok** (D13 — host is not the scheme rule's business) |
| `http://example.com/` | `ErrCIMDClientIDNotHTTPS` | **ok** (bare-root path, D2) |
| `https://example.com/x` | ok | ok |
| `ftp://example.com/x`, `//example.com/x`, `example.com/x` | error | **error** — the hatch permits `http` only, not "any scheme" |
| `http://example.com` (empty path) | error | `ErrCIMDClientIDNoPath` — every other rule still applies |
| `http://user:pw@example.com/x` | error | `ErrCIMDClientIDHasUserInfo` |

That last block is the important one: a reviewer needs to see that the flag widens exactly one rule and nothing else.

`IsCIMDClientIDAllowed`:

| Patterns | Host | Expect |
|---|---|---|
| (empty) | anything | allowed |
| `example.com` | `example.com` | allowed |
| `example.com` | `a.example.com` | refused |
| `*.example.com` | `a.example.com` | allowed |
| `*.example.com` | `a.b.example.com` | **refused** — one label only, D6 |
| `*.example.com` | `example.com` | refused (apex not matched by its own wildcard) |
| `*.example.com` | `.example.com` | refused (empty label) |
| `*.example.com` | `xexample.com` | refused (must match on a label boundary) |
| `example.com`, `*.example.com` | `example.com` and `a.example.com` | both allowed |
| `EXAMPLE.com` | `example.COM` | allowed (case-insensitive both sides) |
| `example.com` | `example.com.` | allowed (trailing dot stripped) |
| `localhost` | `localhost` | allowed (D15) |
| `example.com` | `other.test` | refused |

The `a.b.example.com` row is the one that changed from the earlier draft; give it its own named case so a regression says so.

**Unit — `pkg/lib/oauthclient/resolver_test.go` (extend)**

The existing file already stubs `Queries`. Add:

- CIMD disabled + URL-shaped `client_id` → `nil`, **and `Queries` is never called** (assert on the stub's call count — this is the "costs nothing with CIMD off" property).
- CIMD enabled + URL-shaped `client_id` → `Queries.GetClientConfigByClientID` called with the exact string.
- CIMD enabled + URL-shaped `client_id` whose host is **not** in `allowed_domains` → `Queries` **is** called. The resolver applies no trust policy (§4.1, D5); `allowed_domains` is the fetch path's job. This test is the guard against someone reinstating the earlier draft's behavior.
- CIMD enabled + `http://x.example.com/y` → `Queries` **is** called, regardless of `insecure_http_allowed`. Same reason: an existing `http://` row must keep resolving after the flag is revoked.
- `Resolver` has no feature-config or `allowed_domains` dependency at all — a compile-time property, so assert it by the struct's field list rather than a test.
- CIMD enabled + a `client_id` that is a *static* client and happens to be `https://…/x` → the static config is returned and `Queries` is never called (spec § Client ID Format's pre-registration pattern; the returned config's `DynamicSource` is `""`).
- CIMD enabled + `dcrc_…` → still resolved (no regression).

**Unit — `pkg/lib/config` (project config)**

Add a testdata pair under the existing config test harness: a `authgear.yaml` with the whole `client_id_metadata_document` section absent, asserting after `SetFieldDefaults` that `ClientIDMetadataDocument != nil`, `IsEnabled() == false`, and `GetClientConfig()` has the shared lifetime defaults (`DefaultAccessTokenLifetime`, …) — **not** the spec's illustrative `1800`/`2592000`. A second fixture with a malformed `allowed_domains` entry (`"exa*mple.com"`, `"*"`, `"*.*.example.com"`, `"-bad.example.com"`) asserting a schema validation error, and a third asserting `"localhost"` and `"*.example.com"` both **pass** (D15).

**Unit — `pkg/lib/config` (feature config) — the `update-feature-config` skill's mandatory coverage**

`pkg/lib/config/testdata/merge_feature.yaml`:

- a layer setting `oauth.client_id_metadata_document.insecure_fetch_address_allowed: true`, with the expected merged output — and asserting `oauth.client` fields from another layer **survive** the merge (this is what catches an implementation that replaces `OAuthFeatureConfig` wholesale instead of merging per sub-section);
- a layer setting only `insecure_http_allowed`, asserting the other flag stays `false`;
- both layers setting one flag each, asserting both end up `true`;
- a later layer with the section **absent**, asserting an earlier layer's `true` survives (the plain-`bool` merge semantics documented in §2.4 — if this test is written the other way round, the field must become `*bool`; decide before writing it, not after).

Plus a schema test: `insecure_fetch_address_allowed: "yes"` fails, and an unknown key under `client_id_metadata_document` fails (`additionalProperties: false`).

Nil-safety: `(*OAuthFeatureConfig)(nil).GetClientIDMetadataDocument().IsInsecureHTTPAllowed()` returns `false` without panicking — the chain that `isDynamicClientIDCandidate` relies on.

**Unit — `pkg/lib/oauthclient/client_config_test.go` / `client_model_test.go` (extend)**

A `Client{Source: CIMD, Kind: THIRD_PARTY}` through `ToClientConfig(ResolveTokenLifetimes(cfg, CIMD))`:
- `ApplicationType == OAuthClientApplicationTypeDynamicThirdParty`; `IsThirdParty()` true, `IsConfidential()` false, `IsPublic()` true, `HasFullAccessScope()` false, `IsClientCredentialsFlowAllowed()` false.
- `IsDynamicClient()` true, `IsCIMDClient()` true.
- Lifetimes come from `client_config`, not from `dynamic_client_registration.default_client_config` — set the two sections to *different* values and assert the CIMD one wins. This is the test that would catch the `ResolveTokenLifetimes` case being wired to the wrong config key.
- Through `ToModel`: `RegisteredAt == nil`, `LastFetchedAt` passed through, `Source == "CIMD"`, `PostLogoutRedirectURIs == []`.

**Unit — `pkg/lib/oauth/metadata_test.go` (new, or extend if one exists)**

`PopulateMetadata` with CIMD disabled → key absent. Enabled → `client_id_metadata_document_supported == true`. Assert `registration_endpoint` behavior is unchanged in both cases.

**e2e — `e2e/tests/cimd/discovery.test.yaml` (new)**

Mirrors `e2e/tests/dcr/register_discovery.test.yaml`: one app with CIMD enabled asserting `client_id_metadata_document_supported: true` at both `/.well-known/openid-configuration` and `/.well-known/oauth-authorization-server`; one with it disabled asserting the key is absent. Follow the `write-e2e-test` skill for the fixture layout.

**Commands to run**

```
go test ./pkg/lib/oauthclient/... ./pkg/lib/config/... ./pkg/lib/oauth/...
make lint
make export-schemas && git status --porcelain   # must be empty
make -C e2e run
```

## 9. Spec Updates

`doc:` commit against `docs/specs/cimd.md`:

1. **§ SSRF Protection** — replace the `DEV_MODE` paragraph. Current text: "Authgear implements this gated on `DEV_MODE`, the existing process-wide environment variable (default `false`) … not a per-project `authgear.yaml` setting, so no project admin can enable it. When `DEV_MODE` is `false` (always the case in production), the rules above apply unconditionally."

   It should instead describe the two `authgear.features.yaml` flags: `oauth.client_id_metadata_document.insecure_http_allowed` and `insecure_fetch_address_allowed`, both default `false`, both settable only through the Site Admin surface (never `authgear.yaml`, so no project admin can enable them), intended for test and local-development projects only, and never to be set at the cluster or plan layer. Keep the existing guarantee in the same shape: with both flags `false` — always the case for a project serving real traffic — the rules apply unconditionally.

   State why it is per-project rather than process-wide: it is the only form that lets the enforcement path itself be tested, since a global switch cannot express a permissive project and a strict project at the same time.

2. **§ Client ID Format** — note that the `https` requirement is defeatable by `insecure_http_allowed` for a test or local-development project, and that no other rule in that list is affected by it.

3. **§ Domain Trust** — three additions:
   - `allowed_domains` accepts a single-label hostname (e.g. `localhost`), so a test project can allowlist its own document host.
   - a leading `*.` matches exactly one label (RFC 6125 / TLS convention), so `*.example.com` does not match `a.b.example.com`; an admin who wants the apex too lists it separately.
   - **`allowed_domains` is an onboarding control, not an operating one.** Removing a domain prevents any *new* client from that domain being resolved, and prevents refetches, but does not stop clients that already have a persisted record — their metadata simply freezes at its last successful fetch. To stop an existing client now, use `deleteDynamicClient`; combined with removal from the allowlist that is a durable ban, since the record cannot be recreated. This is the concrete answer to the "closest thing to one" hand-wave in § Client Limit.

## 10. Atomic Commit Plan

1. `Rename OAuthDynamicClientRegistrationDefaultClientConfig to OAuthDynamicClientTokenLifetimesConfig` — §2.1 only, pure rename, no behavior change.
2. `[CIMD] Add the client_id_metadata_document config section` — §2.2, §2.3, §2.5; project-config unit tests.
3. `[CIMD] Add feature config flags for insecure metadata document fetches` — §2.4; `merge_feature.yaml` and feature-config schema/nil-safety tests. Its own commit: it is the security-sensitive change in this part and reviewers should see it alone, with §2.4's reasoning in the commit body.
4. `[CIMD] Add Client Identifier URL validation and allowed_domains matching` — §3, §3.1; `client_id_test.go` including both `allowInsecureHTTP` tables.
5. `[CIMD] Resolve persisted CIMD clients through the dynamic client resolver` — §4, §5, §5.1; `resolver_test.go`, `client_config_test.go`, `client_model_test.go`. After this commit a CIMD row would resolve end-to-end; nothing creates one yet.
6. `[CIMD] Advertise client_id_metadata_document_supported in discovery` — §6; metadata test.
7. `[CIMD] Add e2e test for the CIMD discovery metadata` — §8 e2e.

Body of each: `ref DEV-XXXX`. Commit 3's body should state, in prose, that the flags are Site-Admin-only, must never be set at the cluster or plan layer, and exist for test and local-development projects — so the reasoning survives in `git log` and not only in this file.

Run `make update-vettedpositions` if `goanalysis` line numbers shift (the `config`/`oauth` packages are covered by `.vettedpositions`).
