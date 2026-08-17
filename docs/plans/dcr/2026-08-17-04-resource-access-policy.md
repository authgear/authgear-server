# DCR Part 4 — `access_policy` and Resource Indicators for Third-Party Clients

Spec: [docs/specs/api-resource.md — Access Policy](../../specs/api-resource.md#access-policy), [docs/specs/access-token-audience-binding.md](../../specs/access-token-audience-binding.md), [docs/specs/dcr.md — Security Considerations](../../specs/dcr.md#security-considerations).

Depends on Part 3 (DCR clients resolve correctly at `/oauth2/authorize` and `/oauth2/token`).

## 1. Goal / Scope

Confirmed by research: `access_policy` does not exist anywhere in code today (no column, no Go field, no GraphQL type) — it is entirely spec-only. Confirmed: `resource=` support at `/oauth2/authorize` does not exist at all; `resource` is read only by `handleClientCredentials` (`pkg/lib/oauth/handler/handler_token.go:2167-2237`) for the `client_credentials` grant. Confirmed: `protocol.AuthorizationRequest` is `map[string]string` (single value per key) and the HTTP layer (`pkg/auth/handler/oauth/authorize.go:42-44`) takes only `values[0]` per query parameter — **multi-`resource=` is not representable without a bigger, unrelated change to the request-parsing layer**. Per the user's decision, multi-resource support is first-party-only and out of scope for DCR entirely — so this part implements **single-resource only**, which fits the existing `map[string]string` shape with no request-parsing changes needed.

In scope (per user's "Full round-trip" decision):

- `access_policy` on Resource and Scope (DB column, Go model, Admin GraphQL CRUD).
- Allowing resource-specific scope values through `oauth.ValidateScopesByClientConfig`, which today rejects them outright (§5.2 — a prerequisite for everything else here).
- Relaxing `DecodeAccessToken`'s `aud` validation so a resource-bound JWT is usable at `/oauth2/userinfo` at all (§8.3 — the other prerequisite).
- `resource=` accepted and validated at `/oauth2/authorize` for **dynamic** third-party clients (DCR, and later CIMD) only, gated by `access_policy`. **Correction:** an earlier draft of this scope note said "DCR or static `third_party_app`" — that was wrong. The policy field is literally named `allow_dynamic_third_party_client_access`; a static `third_party_app` client has no mechanism to be granted access to a Resource for the `authorization_code`/`refresh_token` grants at all (it has no client-resource association like M2M's `client_credentials` grant does) — it must fall through to the same `invalid_target` every other static, non-M2M client type gets. This also means `validateResource` needs both `client.IsDynamicClient()` and `client.IsThirdParty()`, not `IsThirdParty()` alone: a dynamic first-party client (`Kind == FIRST_PARTY`) is `IsDynamicClient()` but not `IsThirdParty()` (it resolves to the ordinary `native`/`spa` `ApplicationType`), and must be rejected the same as a static first-party client, since first-party support for `resource=` is separately deferred (§5.5).
- `/resolve` rejecting third-party clients' access tokens (§5.6).
- `resource=` accepted and re-validated at `/oauth2/token` for `authorization_code` and `refresh_token` grants, with JWT issuance bound to the resource's `aud`.
- Per user's "All third-party clients" decision: **every** third-party client (DCR-sourced or legacy static `third_party_app`) defaults to an opaque access token when no `resource` is requested, regardless of its `issue_jwt_access_token` config flag.
- Correct error (`invalid_target`) when a **non-M2M static client** (`spa`, `traditional_webapp`, `native`, `confidential`) requests `resource=` — there is no mechanism to associate a Resource with these client types, so the request must always fail (see §5.5).

Out of scope: multi-resource requests (first-party-only, needs `AuthorizationRequest`/HTTP-parsing changes — flagged above, not attempted here); the `scope_by_aud` claim entirely — **correction:** an earlier draft of this part emitted a single-entry `scope_by_aud` claim, since only one resource can ever be requested. That was removed: with at most one resource URI in `aud`, every granted scope already applies to that single audience, so a per-audience breakdown carries no information the flat `scope` claim doesn't already have. `scope_by_aud` is deferred to whenever first-party multi-resource support is built, at which point it becomes necessary; `oauth.dynamic_client_registration.maximum_clients` (still deferred).

## 2. `access_policy` — Data Model

### 2.1 Migration

New file: `cmd/authgear/cmd/cmddatabase/migrations/authgear/20260817120002-add_resource_access_policy.sql`

```sql
-- +migrate Up

ALTER TABLE _auth_resource ADD COLUMN access_policy jsonb NOT NULL DEFAULT '{}';
ALTER TABLE _auth_resource_scope ADD COLUMN access_policy jsonb NOT NULL DEFAULT '{}';

-- +migrate Down

ALTER TABLE _auth_resource_scope DROP COLUMN access_policy;
ALTER TABLE _auth_resource DROP COLUMN access_policy;
```

`DEFAULT '{}'` means every existing Resource/Scope row gets an empty policy object on migrate — per api-resource.md, "missing keys default to false," so this is exactly "no third-party access" for every pre-existing Resource/Scope, a safe backward-compatible default (no resource is silently exposed to third-party clients by this migration).

### 2.2 `pkg/lib/resourcescope` — extend `Resource`/`Scope`

`pkg/lib/resourcescope/resource.go` / `scope.go`:

```go
type AccessPolicy struct {
	AllowDynamicThirdPartyClientAccess bool `json:"allow_dynamic_third_party_client_access,omitempty"`
}

type Resource struct {
	ID           string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	ResourceURI  string
	Name         *string
	AccessPolicy AccessPolicy // new
}
```

(Same addition to `Scope`.) `AccessPolicy` is marshaled to/from the `access_policy jsonb` column as a JSON object. Reuse the `jsonb` scan/marshal helper Part 2 introduces for `_auth_oauth_client`'s array columns (§4.2 of Part 2). Note that `_auth_resource.metadata` / `_auth_resource_scope.metadata` are **not** a usable precedent despite existing in the migration since `20250710143552-add_resource_scope.sql`: neither column appears in `selectResourceQuery`/`scanResource` (`store_resource.go:253-284`) or their `Scope` equivalents, and is never written by `CreateResource`/`UpdateResource` — there is no `jsonb` handling in `pkg/lib/resourcescope` at all today. `access_policy` will be the package's first.

Update `pkg/lib/resourcescope/store_resource.go` / `store_scope.go`: add `access_policy` to every `Columns(...)`/`Values(...)` in `CreateResource`/`CreateScope`, add it to `selectResourceQuery`/`selectScopeQuery`'s column list and `scanResource`/`scanScope`, and add an `UpdateAccessPolicy` branch to `UpdateResource`/`UpdateScope` (mirroring the existing `NewName`-optional-update pattern at `store_resource.go:65-71`): when the options carry a non-nil `AccessPolicy`, `Set("access_policy", ...)`; when nil, leave the column untouched — matches api-resource.md's mutation contract exactly ("If omitted, the existing access policy is unchanged").

`pkg/lib/resourcescope/resource.go`'s `NewResourceOptions`/`UpdateResourceOptions` (and the `Scope` equivalents) each get an `AccessPolicy *AccessPolicy` field (nil = default-false on create, unchanged on update).

`Resource.ToModel()` / `Scope.ToModel()` (`pkg/lib/resourcescope/resource.go:47-57`) copy `AccessPolicy` through to `model.Resource`/`model.Scope` (`pkg/api/model/resource.go`), which also gain an `AccessPolicy resourcescope.AccessPolicy`-shaped field — reuse the same `AccessPolicy` struct across both packages (define it once in `pkg/api/model` and alias/reuse from `resourcescope`, matching how `model.Resource`/`resourcescope.Resource` already split domain vs. API-facing structs — put the canonical `AccessPolicy` type in `pkg/api/model` since that's the boundary type, and have `resourcescope.Resource.AccessPolicy` be of that same `model.AccessPolicy` type directly rather than a duplicate type, since it carries no store-internal state).

### 2.3 New read path — third-party access check

New methods on `resourcescope.Store` (`pkg/lib/resourcescope/store_resource.go` / `store_scope.go`), used only by the OAuth runtime (Part 4 §5), not the Admin API:

```go
// GetResourceByURIForThirdPartyAccess returns the resource only if its
// access_policy allows third-party access; otherwise ErrResourceNotFound,
// deliberately reusing the same not-found error M2M's GetClientResourceByURI
// uses for "not associated" so both grant paths collapse to invalid_target
// identically at the call site (see §5.1/§6).
func (s *Store) GetResourceByURIForThirdPartyAccess(ctx context.Context, uri string) (*Resource, error) {
	q := s.selectResourceQuery("r").
		Where("r.uri = ? AND (r.access_policy->>'allow_dynamic_third_party_client_access')::boolean IS TRUE", uri)
	row, err := s.SQLExecutor.QueryRowWith(ctx, q)
	if err != nil {
		return nil, err
	}
	r, err := s.scanResource(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, err
	}
	return r, nil
}

// ListScopesForThirdPartyAccess returns only the scopes of resourceID whose
// access_policy allows third-party access.
func (s *Store) ListScopesForThirdPartyAccess(ctx context.Context, resourceID string) ([]*Scope, error) {
	q := s.selectScopeQuery("sc"). // mirror selectResourceQuery's alias convention
		Where("sc.resource_id = ? AND (sc.access_policy->>'allow_dynamic_third_party_client_access')::boolean IS TRUE", resourceID)
	return s.queryScopes(ctx, q) // reuse the existing multi-row query helper, mirroring queryResources
}
```

Add corresponding `Queries` wrappers (`pkg/lib/resourcescope/queries.go`): `GetResourceByURIForThirdPartyAccess`, `ListScopesForThirdPartyAccess`, both returning `*model.Resource`/`[]*model.Scope`.

## 3. Admin GraphQL — `AccessPolicy` CRUD

Per [api-resource.md — Admin API](../../specs/api-resource.md#admin-api), add:

### 3.1 `pkg/admin/graphql/resource.go` (extend)

```go
var accessPolicyType = graphql.NewObject(graphql.ObjectConfig{
	Name: "AccessPolicy",
	Fields: graphql.Fields{
		"allowDynamicThirdPartyClientAccess": &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
	},
})

var accessPolicyInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "AccessPolicyInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"allowDynamicThirdPartyClientAccess": &graphql.InputObjectFieldConfig{Type: graphql.Boolean},
	},
})
```

Add an `"accessPolicy"` field (`Type: graphql.NewNonNull(accessPolicyType)`) to the existing `nodeResource` object (`pkg/admin/graphql/resource.go:27-50`) and to the `Scope` object (wherever it's defined, presumably `pkg/admin/graphql/scope.go` — locate via `grep -n "typeScope\|nodeScope" pkg/admin/graphql/*.go` at implementation time).

### 3.2 `pkg/admin/graphql/resource_mutation.go` / scope mutation file — extend inputs

Add `"accessPolicy": &graphql.InputObjectFieldConfig{Type: accessPolicyInputType}` to `CreateResourceInput`, `UpdateResourceInput`, `CreateScopeInput`, `UpdateScopeInput` (locate the exact current field lists in `pkg/admin/graphql/resource_mutation.go` and the scope mutation file at implementation time — the shapes are given in api-resource.md's SDL). Each resolver decodes the optional `accessPolicy` input map into a `*model.AccessPolicy` (nil if the GraphQL argument was omitted — the `graphql-go` library distinguishes "not provided" from an explicit object via `p.Args["input"].(map[string]any)`'s key presence, matching the existing `NewName *string`-style optional-update pattern already used for `name`/`description`) and passes it through `resourcescope.NewResourceOptions.AccessPolicy` / `UpdateResourceOptions.AccessPolicy` (§2.2).

### 3.3 `pkg/admin/facade/resoucescope.go` (existing file — note the existing typo in its filename; do not rename it as an incidental drive-by change)

`CreateResource`/`UpdateResource`/`CreateScope`/`UpdateScope` already pass their `options` straight through to `resourcescope.Commands` (per `pkg/admin/facade` → `pkg/lib/resourcescope/commands.go`'s existing signatures) — once `NewResourceOptions`/`UpdateResourceOptions`/`NewScopeOptions`/`UpdateScopeOptions` gain the `AccessPolicy` field (§2.2), no facade-layer code change is needed here beyond whatever the Go struct-literal call sites require to keep compiling.

### 3.4 Wiring

`make export-schemas` (GraphQL SDL + portal gentype) in the same commit — no new wire.go changes needed (no new facade/loader, only field additions to existing types).

## 4. `AuthorizationRequest.Resource()` — New Accessor

`pkg/lib/oauth/protocol/authz.go` (add alongside the other single-value accessors, e.g. near `CodeChallenge()`):

```go
func (r AuthorizationRequest) Resource() string { return r["resource"] }
```

Since `AuthorizationRequest` is `map[string]string` and `pkg/auth/handler/oauth/authorize.go:42-44` already takes `values[0]` for every form field, a client sending more than one `resource=` simply has all but the first silently dropped — identical to how any other unexpected repeated parameter already behaves today for this endpoint. Not a new limitation introduced by this part; just now applies to `resource` too. No parsing changes needed.

## 5. `/oauth2/authorize` — Resource Validation

### 5.1 `pkg/lib/oauth/handler/handler_authz.go` — new `validateResource` method

Add a dependency the `AuthorizationHandler` doesn't have yet:

```go
type AuthorizationHandlerResourceScopeService interface {
	GetResourceByURIForThirdPartyAccess(ctx context.Context, uri string) (*resourcescope.Resource, error)
	ListScopesForThirdPartyAccess(ctx context.Context, resourceID string) ([]*resourcescope.Scope, error)
}
```

**Type family:** use `*resourcescope.Resource` / `[]*resourcescope.Scope`, not the `*model.*` variants. The already-shipped sibling interface in `handler_token.go` (`ClientResourceScopeService`, lines 189-192) declares `GetClientResourceByURI(...) (*resourcescope.Resource, error)` and `GetClientResourceScopes(...) ([]*resourcescope.Scope, error)`, and `handleClientCredentials` consumes them as such — mixing `model.*` into the new sibling would mean two conversions for no gain. `GetClientResourceByURI` is **not** redeclared here: the M2M association path lives entirely in `handler_token.go`'s `client_credentials` handler and is never reached from `/oauth2/authorize` (M2M is rejected at `handler_authz.go:891-896` before validation), so the `case` for it in `validateResource` below needs no service call.

Add `ResourceScopeService AuthorizationHandlerResourceScopeService` to the `AuthorizationHandler` struct (`pkg/lib/oauth/handler/handler_authz.go:135-161`), bound to the existing `*resourcescope.Queries`/facade type the same way `handler_token.go`'s `ClientResourceScopeService` already is (locate that binding in `pkg/lib/deps/deps_common.go` and mirror it).

New method, mirroring `handleClientCredentials`'s resource-validation block (`pkg/lib/oauth/handler/handler_token.go:2190-2214`) but with the third-party access-policy check substituted for the M2M association check, plus the new static-non-M2M-client rejection:

```go
// validateResource returns the resource-specific scopes the client is allowed
// to request for the requested resource. The returned slice is consumed by
// §5.2's scope validation; it is empty when no resource was requested.
func (h *AuthorizationHandler) validateResource(ctx context.Context, client *config.OAuthClientConfig, r protocol.AuthorizationRequest) (allowedResourceScopes []string, err error) {
	resourceURI := r.Resource()
	if resourceURI == "" {
		return nil, nil // no resource requested — default behavior, see §8
	}
	if strings.HasPrefix(resourceURI, h.IDTokenIssuer.Iss()) { // mirror handler_token.go:2217's project-endpoint-prefix check exactly
		return nil, protocol.NewError("invalid_target", "resource URI must not be a prefixed by authgear endpoint")
	}

	switch {
	case client.IsDynamicClient() && client.IsThirdParty():
		// Both checks are required, independently: IsThirdParty() alone is
		// also true for a static third_party_app client, but the
		// allow_dynamic_third_party_client_access policy below is named
		// (and must behave) as dynamic-only — see the corrected §1 scope
		// note. IsDynamicClient() alone is not enough either: a dynamic
		// first-party client (Kind == FIRST_PARTY) resolves to the ordinary
		// "native"/"spa" ApplicationType (IsDynamicClient() true,
		// IsThirdParty() false) — first-party support for the resource
		// parameter is separately deferred, same as for static first-party
		// clients. IsDynamicClient() is a new field on OAuthClientConfig
		// (`IsDynamic bool`, json:"-"), set only by
		// oauthclient.Client.ToClientConfig, since ApplicationType alone
		// cannot carry this signal for the first-party case.
		resource, err := h.ResourceScopeService.GetResourceByURIForThirdPartyAccess(ctx, resourceURI)
		if err != nil {
			if errors.Is(err, resourcescope.ErrResourceNotFound) {
				return nil, protocol.NewError("invalid_target", "resource not found or not accessible to third-party clients")
			}
			return nil, err
		}
		allowedScopes, err := h.ResourceScopeService.ListScopesForThirdPartyAccess(ctx, resource.ID)
		if err != nil {
			return nil, err
		}
		return slice.Map(allowedScopes, func(s *resourcescope.Scope) string { return s.Scope }), nil
	case client.ApplicationType == config.OAuthClientApplicationTypeM2M:
		// M2M clients do not use /oauth2/authorize at all — already rejected
		// earlier in ValidateRequestWithoutTx (handler_authz.go:891-896).
		// Unreachable in practice; included only so the switch is exhaustive
		// and self-documenting about why M2M isn't handled here.
		return nil, protocol.NewError("unauthorized_client", "m2m clients are not allowed to use the authorize endpoint")
	default:
		// Every client that isn't both dynamic and third-party — always
		// error. This covers every static client type, including
		// third_party_app (a static third-party client has no mechanism to
		// be associated with a Resource for this grant today — see the
		// corrected §1 scope note), and every first-party client, static or
		// dynamic (deferred to a later, separate piece of work; see §5.5).
		return nil, protocol.NewError("invalid_target", "this client is not permitted to use the resource parameter")
	}
}
```

Call it from `doValidateRequestWithoutTx` (`pkg/lib/oauth/handler/handler_authz.go:1012-1033`):

```go
func (h *AuthorizationHandler) doValidateRequestWithoutTx(
	ctx context.Context, // new parameter
	client *config.OAuthClientConfig,
	r protocol.AuthorizationRequest,
) error {
	if err := h.validateResponseTypeIsWhitelisted(r); err != nil {
		return err
	}
	if err := h.validatePrompt(r); err != nil {
		return err
	}
	if err := h.validateRequestParameters(client, r); err != nil {
		return err
	}
	if _, err := h.validateResource(ctx, client, r); err != nil { // new
		return err
	}
	if r.SSOEnabled() && client != nil && client.MaxConcurrentSession == 1 {
		return protocol.NewError("invalid_request", "'sso_enabled' must be false if config 'x_max_concurrent_session' is 1")
	}
	return nil
}
```

`doValidateRequestWithoutTx` gains a `ctx` parameter. It has **two** call sites, not one — `ValidateRequestWithoutTx` (`handler_authz.go:911`) and `doHandleConsentRequest` (`handler_authz.go:799`). Both already have `ctx` in scope, so this stays contained within this one file with no fan-out (unlike Part 3's `ResolveClient` change). Note the consequence: `validateResource` runs again on the consent step, so a resource-bound authorization does one extra `access_policy` read there. Accepted — consent is not a hot path, and re-validating means a Resource whose policy was revoked mid-flow is caught before the code is issued.

### 5.2 Resource-specific scopes must be allowed through `ValidateScopesByClientConfig`

**This is the change without which nothing else in this part works, and it was missing from the original draft of this plan.**

`oauth.ValidateScopesByClientConfig` (`pkg/lib/oauth/scope.go:219`) starts with `ValidateScopes(scopes, AllowedScopes)`, and `AllowedScopes` (`scope.go:40-56`) is a closed list: `offline_access`, `device_sso`, `openid`, `profile`, `email`, `address`, `phone`, plus the three `https://authgear.com/scopes/*` values. Any other scope value is rejected with `invalid_scope`.

It is called at both authorize entry points — `doHandleRequestWithTx` (`handler_authz.go:434`) and `doHandleConsentRequest` (`handler_authz.go:806`) — and on the token side at `handler_token.go:943` and `:1355`.

So today `scope=openid read:tools` is rejected outright, before `validateResource` is ever consulted. That is precisely the request in dcr.md's UC2 Step 3 and in every example in access-token-audience-binding.md. The M2M path avoids this by never going through `ValidateScopesByClientConfig` at all: `handleClientCredentials` calls bare `oauth.ValidateScopes(r.Scope(), allowedScopeStrs)` against the resource's own scope list (`handler_token.go:2239`).

**Change:** give `ValidateScopesByClientConfig` an extra allowed-scope set, sourced from the requested resource.

```go
// pkg/lib/oauth/scope.go
func ValidateScopesByClientConfig(
	client *config.OAuthClientConfig,
	scopes []string,
	allowedResourceScopes []string, // new; nil when no resource was requested
) error {
	allowedScopes := AllowedScopes
	if len(allowedResourceScopes) > 0 {
		allowedScopes = append(slices.Clone(AllowedScopes), allowedResourceScopes...)
	}
	if err := ValidateScopes(scopes, allowedScopes); err != nil {
		return err
	}
	// ... rest of the existing per-scope loop unchanged ...
}
```

Everything else in that function (the `offline_access` / full-access / `device_sso` / pre-authenticated-URL rules, and the mandatory-`openid` check at line 250) is untouched: resource scopes fall through the loop's `if` chain without matching anything, exactly as an unknown-but-allowed scope should.

This is what implements access-token-audience-binding.md's error-table row **"`scope` includes a resource-specific scope but no matching `resource` was requested → `invalid_scope`"** — with `allowedResourceScopes` nil, a resource scope is not in `AllowedScopes` and is rejected as `invalid_scope`, which is the specified behavior and also what happens today. No separate check is needed for that row.

Wiring at the two authorize call sites: `validateResource` now returns the allowed resource scopes, but it runs inside `doValidateRequestWithoutTx`, which is called *before* `ValidateScopesByClientConfig` at `handler_authz.go:806` and *after* it at `:434`. Rather than reorder, call `validateResource` directly at each of the two `ValidateScopesByClientConfig` call sites and thread its result in:

```go
allowedResourceScopes, err := h.validateResource(ctx, client, r)
if err != nil {
	return nil, err
}
if err := oauth.ValidateScopesByClientConfig(client, r.Scope(), allowedResourceScopes); err != nil {
	return nil, err
}
```

and drop the `validateResource` call from `doValidateRequestWithoutTx` (keeping the `ctx` parameter change, which is still needed by nothing else — so revert that too and leave `doValidateRequestWithoutTx` alone entirely). This keeps resource validation and scope validation adjacent and single-pass, and means the double-validation note above does not apply: each entry point validates once.

The two token-endpoint call sites (`handler_token.go:943`, `:1355`) pass the resource scopes derived from the resource bound to the code / refresh token (§6, §7). Both need the same treatment, otherwise the token exchange rejects a `scope` the authorize step already granted. Determine the allowed set there by looking up the bound `resourceURI`'s policy-enabled scopes via the same `ResourceScopeService` — the token handler will need the two new methods added to its own service interface as well.

All other callers of `ValidateScopesByClientConfig` pass `nil` for the new parameter.

**Database scope — `validateResource` cannot be called from an arbitrary point in this handler.** `appdb.SQLExecutor` panics with `programming_error: tx is not initialized` when no transaction is in the context (`pkg/lib/infra/db/sql_executor.go:71,111,151` → `hook_handle.go:64-70`), and there is no request-wide DB scope: the session middleware closes its `ReadOnly` before `next.ServeHTTP` (`pkg/lib/session/middleware.go:64`). Of the two authorize call sites:

- `doHandleRequestWithTx` (`handler_authz.go:434`) runs inside `h.Database.WithTx` (opened at `handler_authz.go:415`) — **safe**.
- `doHandleConsentRequest` (`handler_authz.go:806`) is reached via `doHandleConsent` (`handler_authz.go:229`), which has no `WithTx` anywhere in its path — **panics**.

This is the same defect Part 3 §4.3 fixes for the client resolver, and it takes the same fix: `validateResource` guards its own reads with the `IsInTx` / `ReadOnly` branch (`pkg/lib/usage/limit.go:394-404` is the precedent). `AuthorizationHandler.Database` is already the right handle type — its interface at `handler_authz.go:130` currently declares only `WithTx`, so add `ReadOnly` and `IsInTx` to `AuthorizationHandlerDatabase`.

Do **not** instead wrap `doHandleConsent` in a `WithTx`: it is deliberately outside a write transaction, and widening it would put the consent-screen render inside a transaction.

The token-endpoint sites are inside the token handler's existing transaction and need no guard — verify this when wiring rather than assuming it, using the same `IsInTx`-branch helper so the answer does not matter.

### 5.3 Error surface confirmation

`doValidateRequestWithoutTx`'s caller (`ValidateRequestWithoutTx`, `handler_authz.go:911-928`) already converts any returned `*protocol.OAuthProtocolError` into a proper `AuthorizationResultError` redirect with `error`/`error_description` query parameters (per RFC 6749 §4.1.2.1, matching [Error Cases — Authorization endpoint errors](../../specs/access-token-audience-binding.md#authorization-endpoint-errors)). No new error-handling plumbing is needed — reusing `protocol.NewError` is sufficient, exactly as the neighboring `validateResponseTypeIsWhitelisted`/`validatePrompt` methods already do.

### 5.4 Binding the resource to the issued code — no new field needed

`oauth.CodeGrant` already embeds the full `protocol.AuthorizationRequest` verbatim (`pkg/lib/oauth/grant_code.go:27`, `AuthorizationRequest protocol.AuthorizationRequest`). Since §4 adds `Resource()` to that same type, **the resource the user authorized is automatically persisted with the code** the moment `generateCodeResponse` (`handler_authz.go:1035`) constructs the `CodeGrant` — no new field, no new migration, no new serialization logic.

### 5.5 First-party clients rejecting `resource=` is intended, and the specs need to say so

The `default` branch of `validateResource` rejects every static non-M2M client (`spa`, `traditional_webapp`, `native`, `confidential`) with `invalid_target`. **Confirmed as the expected behavior:** first-party support for the `resource` parameter is a separate, later piece of work, not an omission in this part.

Two specs currently promise otherwise and must be amended in this part's doc-fix commit (§12):

- `api-resource.md`'s intro bullet: "**First-party clients** (all grant types) — first-party clients can request resource-bound tokens if they are explicitly associated with the Resource."
- `access-token-audience-binding.md`'s Background: "This spec extends support to all client types using the `authorization_code` and `refresh_token` grants", plus the Authorization Endpoint section's `resource` being shown as "optional, repeatable" with "multiple resources allowed".

Both should be marked as not-yet-implemented: **dynamic** third-party clients (DCR-registered, and later CIMD) only — not static `third_party_app`, which has no mechanism to be associated with a Resource for this grant (see §1's corrected scope note) — single `resource` value only, with first-party support and multi-resource support called out as future work. Leaving the specs as-is would mean shipping Part 4 in direct contradiction with two of the three specs it implements.

### 5.6 `/resolve` must reject **opaque** tokens from third-party clients

`access-token-audience-binding.md` (Authgear's decision — Third-party clients) states the opaque token issued to a third-party client "Cannot be used with the `/resolve` endpoint", and dcr.md's Security Considerations repeats it: a DCR client with no `resource` receives a token "scoped to the userinfo endpoint only". Nothing in the codebase enforces this today: `pkg/resolver/handler/resolve.go`'s `resolve()` reads whatever session the session middleware already resolved (`session.HasValidSession` / `GetSession`) and applies no client-kind, scope, or token-shape check at all.

**The gate is the conjunction of two conditions: the token is opaque AND the client is third-party.** Neither alone is correct:

- Gating on opacity alone would break every existing project. `issue_jwt_access_token` defaults to false (only `m2m` forces it true, `pkg/lib/config/oauth.go:394-398`), so the *default* first-party client already issues opaque access tokens and already uses `/resolve`.
- Gating on third-party alone would also reject a third-party client's *resource-bound JWT*, which the spec does not ask for — atab.md scopes the restriction to the opaque token issued when no `resource` was requested.

Implementation, in `pkg/resolver/handler/resolve.go`'s `resolve()` (not in the shared session resolver — see below):

```go
if !clientLike.IsFirstParty && !accessTokenIsJWT {
	return &model.SessionInfo{IsValid: false}, nil
}
```

- `clientLike` comes from `oauth.SessionClientLike(ctx, s, h.OAuthClientResolver)` (`pkg/lib/oauth/client_like.go:24`, already `ctx`-threaded by Part 3) — the same helper `/oauth2/userinfo` uses. It returns `ClientLikeNotFound` (`IsFirstParty: false`) for an unresolvable client, so it fails closed. This adds an `OAuthClientResolver` dependency to `ResolveHandler`.
- `accessTokenIsJWT` is the `isHash` return value of `AccessTokenEncoding.DecodeAccessToken` (`token_encoding.go:265`), which is `true` exactly when the presented token parsed as a project-signed JWT and `false` for an opaque token. That value is currently discarded inside `oauth.Resolver.resolveAccessToken` (`pkg/lib/oauth/resolver.go:79-90`). Propagate it rather than re-deriving it: set it on the request context from the resolver (a small `session.WithAccessTokenIsJWT(ctx, isHash)` / `session.GetAccessTokenIsJWT(ctx)` pair alongside the existing `session.HasValidSession`/`GetSession` accessors in `pkg/lib/session/context.go`). Do **not** re-sniff the `Authorization` header for an `eyJ` prefix in the resolver handler — that duplicates `DecodeAccessToken`'s discriminator and will drift.
- Sessions that are not access-token-backed (IDP session cookie, app session cookie) must be unaffected: `accessTokenIsJWT` is only meaningful when the session came from a bearer access token, so the context value should be a tri-state (unset / opaque / JWT) and the gate applies only in the "opaque" case.

**Why this lives in `pkg/resolver`, not in `oauth.Resolver`:** a third-party client's opaque token *must* keep working at `/oauth2/userinfo` (dcr.md: "scoped to the userinfo endpoint only"). Putting the rejection in the shared session resolver would break that. It is a `/resolve`-specific policy.

Landed as its own commit (§12) since it is a behavior change to an endpoint otherwise untouched by this part; `pkg/resolver`'s `wire_gen.go` must be regenerated.

## 6. `/oauth2/token` — `authorization_code` Grant

### 6.1 Re-validation (`doIssueTokensForAuthorizationCode`, `handler_token.go:1679-1963`)

Per [access-token-audience-binding.md — Token Endpoint](../../specs/access-token-audience-binding.md#token-endpoint): if `resource` is present on the token request, it must be a subset of what was bound to the code; for this part's single-resource scope, "subset" means "absent, or exactly equal to `code.AuthorizationRequest.Resource()`." Add near the top of `doIssueTokensForAuthorizationCode`:

```go
resourceURI := code.AuthorizationRequest.Resource()
if requested := r.Resource(); requested != "" { // r is the protocol.TokenRequest passed into this call chain — confirm exact param name at the call site in handleAuthorizationCode (handler_token.go:560-569)
	if requested != resourceURI {
		return nil, protocol.NewError("invalid_target", "resource must be a subset of the resource authorized by the code")
	}
}
```

(`resourceURI` remains `""` when the code carried no resource, matching "no resources were bound" in the spec.)

### 6.2 Threading `resourceURI` to token issuance

`PrepareUserAccessGrantOptions` (`pkg/lib/oauth/grant_access_service.go:20-27`) gains a `ResourceURI string` field. At `handler_token.go:1916` (the `PrepareUserAccessGrantOptions{...}` construction inside `doIssueTokensForAuthorizationCode`), set `ResourceURI: resourceURI`.

### 6.3 Persisting the resource for subsequent refresh — `OfflineGrantRefreshToken`

`pkg/lib/oauth/grant_offline.go`'s `OfflineGrantRefreshToken` struct (lines 19-36) gains:

```go
// ResourceURI was added on 2026-08-17. Refresh tokens created before that
// date have an empty ResourceURI (equivalent to no resource bound) — matches
// the existing nil-AccessInfo/nil-ExpireAt backward-compatibility comments
// on this same struct.
ResourceURI string `json:"resource_uri,omitempty"`
```

Set `ResourceURI: resourceURI` wherever `oauth.OfflineGrantRefreshToken{...}` is constructed for a **newly-issued** refresh token during `doIssueTokensForAuthorizationCode` — locate the exact construction site via `grep -n "OfflineGrantRefreshToken{" pkg/lib/oauth/handler/handler_token.go` (expected inside the same function, near where `issueOfflineGrant` is invoked).

**Rotation must carry the binding forward.** Every *other* `OfflineGrantRefreshToken{...}` construction site must copy `ResourceURI` from the token it supersedes — in particular the refresh-token rotation path (`refresh_token_rotation_enabled`) and any add-token path in `pkg/lib/oauth/grant_offline_service.go`. Find them all with the same grep across `pkg/lib/oauth/` (not just `handler_token.go`) and treat a missing copy as a bug, not a default: if a rotated token loses `ResourceURI`, the *next* refresh sees `originalResourceURI == ""` and silently downgrades a resource-bound session to an opaque, project-scoped token (§8.2), or rejects a correctly-repeated `resource=` with `invalid_target` (§7.1). Neither failure is visible at the moment of rotation — it surfaces one refresh later — so this needs an explicit test (§10).

## 7. `/oauth2/token` — `refresh_token` Grant

### 7.1 `issueTokensForRefreshToken` (`handler_token.go:1963-2020`)

Per [access-token-audience-binding.md — refresh_token grant](../../specs/access-token-audience-binding.md#refresh_token-grant): `resource` optional, downscoping allowed, upscoping rejected. For single-resource scope, "downscoping" only has one meaningful case (going from "has a resource" to "no resource" isn't meaningful for opaque-vs-JWT purposes here since a first-party-only multi-resource downscoping concept doesn't apply):

```go
originalResourceURI := offlineGrantSession.ResourceURI // read from the matched OfflineGrantRefreshToken (§6.3)
resourceURI := originalResourceURI
if requested := r.Resource(); requested != "" {
	if requested != originalResourceURI {
		return nil, protocol.NewError("invalid_target", "resource must be a subset of the resource originally authorized")
	}
	resourceURI = requested
}
```

If `r.Resource() == ""`, per spec ("If omitted, the new access token is issued for the same resources as the previous access token"), keep `resourceURI = originalResourceURI` unchanged — this is already the fallback above.

Thread `ResourceURI: resourceURI` into the `PrepareUserAccessGrantOptions{...}` construction at `handler_token.go:1997`, same as §6.2.

## 8. Token Encoding — Opaque-by-Default and Resource-Bound `aud`

### 8.1 `pkg/lib/oauth/token_encoding.go` — `EncodeUserAccessTokenOptions`

```go
type EncodeUserAccessTokenOptions struct {
	OriginalToken      string
	ClientConfig       *config.OAuthClientConfig
	ClientLike         *ClientLike
	AccessGrant        *AccessGrant
	AuthenticationInfo authenticationinfo.T
	ResourceURI        string // new; empty means no resource was requested/bound
}
```

`pkg/lib/oauth/grant_access_service.go:67-73`'s `EncodeUserAccessTokenOptions{...}` literal gains `ResourceURI: options.ResourceURI` (reading it from the new `PrepareUserAccessGrantOptions.ResourceURI`, §6.2/§7.1).

### 8.2 `PrepareUserAccessToken` (`token_encoding.go:92-171`) — decision logic rewrite

Replace the current line 93 check:

```go
if !options.ClientConfig.IssueJWTAccessToken {
	return &prepareUserAccessTokenResultOpaque{...}, nil
}
```

with:

```go
issueJWT := options.ClientConfig.IssueJWTAccessToken
if options.ResourceURI != "" {
	issueJWT = true // a resource-bound token is always a JWT, regardless of the flag
} else if options.ClientConfig.IsThirdParty() {
	issueJWT = false // per the user's decision: ALL third-party clients default to opaque without a resource, overriding issue_jwt_access_token
}

if !issueJWT {
	return &prepareUserAccessTokenResultOpaque{
		OriginalToken: options.OriginalToken,
		ClientConfig:  options.ClientConfig,
	}, nil
}
```

`aud` construction (currently line 105, `_ = claims.Set(jwt.AudienceKey, e.BaseURL.Origin().String())`):

```go
if options.ResourceURI != "" {
	_ = claims.Set(jwt.AudienceKey, []string{options.ResourceURI})
} else {
	_ = claims.Set(jwt.AudienceKey, e.BaseURL.Origin().String()) // unchanged first-party default
}
```

**Correction:** an earlier draft of this snippet also emitted a single-entry `https://authgear.com/claims/scope_by_aud` claim (via a small `excludeOIDCScopes`/`scopeAudienceMap` helper). That claim was removed: with at most one resource URI ever in `aud`, every granted scope already applies to that single audience, and the flat `scope` claim already lists them all — a per-audience breakdown adds no information. `scope_by_aud` is deferred to whenever first-party multi-resource support is built.

Top-level `scope` claim (line 113, `claims.Set("scope", strings.Join(options.AccessGrant.Scopes, " "))`) is **unchanged** — it always includes every granted scope (OIDC + resource-specific), matching the spec's "OIDC scopes ... appear in the top-level scope field even though there is no corresponding aud entry for the project endpoint."

### 8.3 `DecodeAccessToken` must stop validating `aud` against the project endpoint

**Without this, a resource-bound JWT access token cannot be used anywhere in Authgear — including `/oauth2/userinfo`.** This was missing from the original draft of this plan and is as blocking as §5.2.

`AccessTokenEncoding.DecodeAccessToken` (`pkg/lib/oauth/token_encoding.go:265-291`) is the single entry point through which every presented access token is turned into a session. For anything starting with `eyJ` it signature-verifies the JWT and then runs:

```go
err = jwt.Validate(token,
	jwt.WithClock(&jwtClock{e.Clock}),
	jwt.WithAudience(e.BaseURL.Origin().String()),   // <-- project endpoint
)
if err != nil {
	return "", false, err
}
```

§8.2 issues resource-bound tokens with `aud = [<resource_uri>]` and **no** project endpoint entry (that omission is the entire point — access-token-audience-binding.md's "The project endpoint is **not** included"). So `jwt.Validate` fails, `DecodeAccessToken` returns an error, and `oauth.Resolver.resolveAccessToken` (`pkg/lib/oauth/resolver.go:79-96`) converts that into `session.ErrInvalidSession`. Every consumer of the session resolver — `/oauth2/userinfo`, `/resolve`, the auth UI — rejects the token as unauthenticated.

That contradicts three specs, all of which state the opposite explicitly:

- api-resource.md: "The userinfo endpoint accepts tokens where `scope` contains OIDC scopes ..., regardless of whether the Authgear project endpoint is present in `aud`."
- access-token-audience-binding.md (Access Token Claims): "The userinfo endpoint accepts tokens where `scope` contains OIDC scopes ..., regardless of the `aud` claim."
- dcr.md UC2 Step 4: "If `openid` or other OIDC scopes were also requested and granted, the userinfo endpoint remains accessible via that token."

**Change:** drop the `jwt.WithAudience(...)` option, keeping signature verification and `jwt.WithClock` (i.e. `exp`/`nbf` validation):

```go
// aud is deliberately NOT validated here. This function is Authgear's own
// introspection path for a token Authgear itself minted and just
// signature-verified; the authority for "is this token live" is the
// jti -> AccessGrant lookup in oauth.Resolver, not aud. aud exists for
// *resource servers* to enforce (RFC 8707), which is exactly why a
// resource-bound token carries only the resource URI and not the project
// endpoint. Validating aud here would make every resource-bound token
// unusable at /oauth2/userinfo, contrary to
// access-token-audience-binding.md and api-resource.md.
err = jwt.Validate(token, jwt.WithClock(&jwtClock{e.Clock}))
```

This is not a loosening of trust: the JWT must still be signed by the project's own key set (`jwk.PublicSetOf(e.Secrets.Set)`), and the returned `token.JwtID()` must still match a live `AccessGrant` row whose `Authorization` has not been revoked. A cross-project or forged token still fails at the signature step. The only thing that stops being enforced is a check that Authgear was performing against itself.

Endpoint-level audience/scope enforcement stays where the specs put it:

- `/oauth2/userinfo` — accepts the token, and already gates the *claims* it returns on the granted scopes via `oauth.SessionClientLike` → `GetUserInfo(ctx, userID, clientLike)` (`pkg/auth/handler/oauth/userinfo.go:23,42`). No change needed there; that is the "accepts tokens where `scope` contains OIDC scopes" behavior, already implemented.
- `/resolve` — gets the new third-party gate, §5.6.
- Resource servers — validate `aud` themselves, per the spec's "Resource server validation" checklist.

Add a unit test in `pkg/lib/oauth/token_encoding_test.go`: a JWT minted with `aud=["https://api.example.com/orders"]` decodes successfully and returns `isHash=true` with the right `jti`; an expired one still fails; one signed by a foreign key still fails.

### 8.4 `EncodeClientAccessToken` (`token_encoding.go:228+`) — unaffected

This function is only used by `handleClientCredentials` (M2M) — already resource-bound by construction (M2M requires `resource`) and already opaque-vs-JWT gated by the same `IssueJWTAccessToken` flag pattern. Not touched by this part.

## 9. Error Code Confirmation

The user asked to verify the correct error code for a static non-M2M client using `resource=`. Confirmed via the **existing, already-shipped** M2M code path (`handler_token.go:2190-2227`): every resource-authorization failure there — resource not found, resource not associated with the client, resource required but missing, resource prefixed by the project endpoint — uses **`invalid_target`**, not `invalid_resource`. `invalid_target` is also what `access-token-audience-binding.md`'s own error tables use throughout. `api-resource.md`'s prose ("the server returns `invalid_resource`") is inconsistent with both the current implementation and the newer spec — flag this as a doc-fix commit (§12), consistent with the doc-fix pattern established in Parts 1–2. **Every** `invalid_target` in this plan (§5.1, §6.1, §7.1) is deliberately consistent with this existing convention, including the new static-non-M2M-client rejection.

## 10. Test Plan

Unit tests (Convey):

- `pkg/lib/resourcescope/store_resource_test.go` / `store_scope_test.go` (or the package's existing DB-integration test harness if one exists — check for a `_test.go` file that already talks to a real test Postgres before assuming none exists) — `GetResourceByURIForThirdPartyAccess`/`ListScopesForThirdPartyAccess` return only policy-enabled rows.
- `pkg/lib/oauth/token_encoding_test.go` (extend) — `PrepareUserAccessToken`: third-party client, no resource → opaque, regardless of `IssueJWTAccessToken`; third-party client, with resource → JWT with `aud=[resourceURI]`; first-party client, no resource → unchanged existing behavior (flag-gated); any client with resource → JWT with `aud=[resourceURI]`.
- `pkg/lib/oauth/handler/handler_authz_test.go` (extend) — `validateResource`: third-party client + policy-enabled resource → returns the policy-enabled scope list; third-party client + policy-disabled resource → `invalid_target`; static `spa`/`native`/`confidential` client + any resource → `invalid_target`; resource prefixed by project endpoint → `invalid_target`.
- `pkg/lib/oauth/scope_test.go` (extend) — `ValidateScopesByClientConfig` with the new `allowedResourceScopes` parameter (§5.2): `openid read:orders` + `allowedResourceScopes=["read:orders"]` → nil; the same scopes with `allowedResourceScopes=nil` → `invalid_scope` (this is the spec's "resource-specific scope requested without a matching `resource`" row); `openid read:orders` + `allowedResourceScopes=["read:inventory"]` → `invalid_scope`; every pre-existing case with `nil` → unchanged behavior.
- `pkg/lib/oauth/grant_offline_service_test.go` / `handler_token_test.go` (extend) — refresh-token **rotation** preserves `OfflineGrantRefreshToken.ResourceURI` (§6.3): rotate once, then refresh again with no `resource=` and assert the issued token is still a JWT with `aud=[resourceURI]`, not an opaque token. This is the failure mode that only shows up one refresh after rotation.

e2e tests (`e2e/tests/dcr_resource_indicator.yaml`):

1. Admin API: create a Resource with `accessPolicy: { allowDynamicThirdPartyClientAccess: true }` and a Scope on it with the same policy; verify `resource.accessPolicy`/`scope.accessPolicy` round-trip via the query.
2. DCR third-party client, full authorization-code flow with `resource=<that Resource>` and its scope → access token is a JWT with `aud` containing the resource URI (decode and assert).
3. Same client, **no** `resource=` → access token is opaque (not a valid JWT).
4. DCR third-party client requests a Resource **without** the policy enabled → `/oauth2/authorize` redirects with `error=invalid_target`.
5. DCR third-party client requests a Scope without the policy enabled (Resource policy on, Scope policy off) → `error=invalid_scope`.
6. A **static SPA client** (`x_application_type: spa`) requests `resource=<any Resource>` → `error=invalid_target`.
7. A **static M2M client** using `client_credentials` (not `/oauth2/authorize`) is unaffected by any of the above — existing M2M e2e coverage should still pass unmodified.
8. Refresh-token grant: request an access token bound to a resource, refresh without `resource=` → new token still bound to the same resource; refresh with a **different** `resource=` → `error=invalid_target`.
9. Legacy static `third_party_app` client (if a fixture already exists in the e2e suite) with `issue_jwt_access_token: true`, no `resource=` → access token is now opaque (behavior change, confirm intentionally, matching the user's "all third-party clients" decision).
10. `/resolve` and `/oauth2/userinfo` matrix (§5.6, §8.3) — four cases, the last two being the regressions these tests exist to catch:
    - third-party client, **opaque** token → `/resolve` invalid; `/oauth2/userinfo` **succeeds**.
    - third-party client, **resource-bound JWT** → both succeed (`/resolve` is not gated on JWTs, `/oauth2/userinfo` must accept an `aud` that lacks the project endpoint).
    - first-party client, default **opaque** token → both still succeed, unchanged.
    - first-party client with `issue_jwt_access_token: true`, no `resource=` → both still succeed, unchanged (`aud` is still the project endpoint).
11. DCR third-party client requests `scope=openid read:tools` **with** `resource=` pointing at the resource that defines `read:tools` → `201`/success; the identical request **without** `resource=` → `error=invalid_scope` (§5.2, and the spec's error-table row for a resource scope with no matching resource).

## 11. Fixed Behavioral Decisions

- Single-resource only; multi-`resource=` and `scope_by_aud`'s multi-entry form are out of scope (see §1) — would require changing `protocol.AuthorizationRequest` from `map[string]string` to a multi-value shape and updating `pkg/auth/handler/oauth/authorize.go`'s form-parsing loop, which is a larger, unrelated change reserved for whenever first-party multi-resource support is actually built.
- The opaque-by-default change applies to **every** third-party client regardless of source (static `third_party_app` included), per the user's explicit decision — this is a real behavior change for any existing project with a static third-party client that has `issue_jwt_access_token: true` set. Call this out prominently in the PR description, not just this plan.
- `api-resource.md`'s `invalid_resource` wording is treated as a doc bug, not followed — `invalid_target` is used everywhere in this part, consistent with the already-shipped M2M behavior and `access-token-audience-binding.md`.
- No new field is added to `oauth.CodeGrant` — the resource binding rides for free on the existing embedded `AuthorizationRequest`.
- Rejecting `resource=` from first-party static clients is intended, not a gap (§5.5); the two specs that promise otherwise are amended in this part's doc-fix commit.
- `/resolve` rejects only the conjunction "opaque token AND third-party client" (§5.6). Gating on opacity alone breaks every default first-party project; gating on client kind alone wrongly rejects third-party resource-bound JWTs. The rejection lives in `pkg/resolver`, not the shared session resolver, because the same opaque token must keep working at `/oauth2/userinfo`.
- `DecodeAccessToken` stops validating `aud` against the project endpoint (§8.3). Signature verification and `exp` still apply, and the `jti` → `AccessGrant` lookup remains the authority on liveness; `aud` becomes purely a resource-server-side check, as RFC 8707 intends.
- `ValidateScopesByClientConfig` gains an allowed-resource-scope parameter (§5.2) rather than resource scopes bypassing scope validation entirely; the mandatory-`openid` rule and all existing per-scope rules stay in force for resource-bound requests.

## 12. Atomic Commit Plan

1. **`doc: Correct resource indicator scope and error codes in the API specs`** — three doc fixes:
   - `api-resource.md`: `invalid_resource` → `invalid_target`, aligning with the already-shipped M2M behavior (§9);
   - `api-resource.md` + `access-token-audience-binding.md`: mark first-party `resource` support and multi-`resource` requests as not yet implemented (§5.5).
2. **`[DCR] Add access_policy to Resource and Scope`** — §2 (migration, `resourcescope` struct/store/queries changes) + unit tests.
3. **`[DCR] Add AccessPolicy to Admin GraphQL Resource/Scope API`** — §3 + `make export-schemas`.
4. **`[DCR] Allow resource-specific scopes in scope validation`** — §5.2's `ValidateScopesByClientConfig` signature change plus `nil` at every existing call site + unit tests. Purely additive: with `nil` everywhere, behavior is byte-identical to today, so this lands safely ahead of the handler work.
5. **`[DCR] Add resource parameter support to the authorization endpoint`** — §4, §5.1/§5.3/§5.4 (`Resource()` accessor, `validateResource`, threading its result into the two `ValidateScopesByClientConfig` call sites, wiring) + unit tests.
6. **`[DCR] Bind resource to authorization_code and refresh_token grants`** — §6, §7 (`OfflineGrantRefreshToken.ResourceURI` including the rotation carry-forward, grant handler changes, token-endpoint scope validation) — no behavior change yet for token *shape*, only binding/validation.
7. **`[DCR] Stop validating access token aud against the project endpoint`** — §8.3 only, plus its unit test. Must land **before** commit 8: without it, the first resource-bound JWT issued is unusable at every endpoint. Self-contained and independently revertable.
8. **`[DCR] Issue resource-bound JWTs and default third-party clients to opaque tokens`** — §8.1, §8.2, §8.4, the actual token-shape/`aud` behavior change — isolated as its own commit since it's the one with real backward-compatibility impact, making it easy to revert independently if needed.
9. **`[DCR] Reject third-party clients' opaque tokens at the resolver endpoint`** — §5.6 + the `accessTokenIsJWT` context plumbing + regenerated `pkg/resolver/wire_gen.go`.
10. **`[DCR] Add e2e tests for resource indicators and access policy`** — `e2e/tests/dcr_resource_indicator.yaml`.
