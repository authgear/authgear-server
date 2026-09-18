# Resource Indicator Part 3 — Enforcing the policy at `/oauth2/authorize` and `/oauth2/token`

Spec: [docs/specs/api-resource.md — Two-level check](../../specs/api-resource.md#two-level-check), [— Revocation](../../specs/api-resource.md#revocation), [docs/specs/access-token-audience-binding.md](../../specs/access-token-audience-binding.md), [docs/specs/client.md — Access Token Behavior by Client Kind](../../specs/client.md#access-token-behavior-by-client-kind).

Depends on Part 1 (`AccessPolicy.AllowsClient`, and the store reads no longer filtering on the policy). Independent of Part 2, though an admin cannot exercise the three new categories until Part 2 ships.

This is the part with behaviour. Two distinct changes land here, and they are worth keeping apart in review:

1. **Widening** — `resource` at `/oauth2/authorize` stops being dynamic-third-party-only and becomes available to whichever categories a Resource opens itself to.
2. **Revocation** — the policy starts being re-read on every access token issuance, which it is not today. Clearing a key currently does nothing to a live grant until it expires.

## 1. What is there now

`validateResource` (`pkg/lib/oauth/handler/handler_authz.go:200-266`) switches on the client: dynamic-and-third-party reads the policy, `m2m` returns `unauthorized_client`, and **everything else falls into a `default` arm that returns `invalid_target` unconditionally**. That arm is what this part deletes.

On the token side there is no policy read at all. `doIssueTokensForAuthorizationCode` (`handler_token.go:1764-1767`) and `issueTokensForRefreshToken` (`:2066-2070`) each compare the requested `resource` against the one bound to the code or refresh token and stop there. So a Resource whose keys an admin has just cleared keeps minting tokens on every refresh for as long as the offline grant lives. That is the bug the spec's Revocation section describes; it is not a regression introduced here.

## 2. One shared read path

`handler_authz.go` and `handler_token.go` are both `package handler`, so the service interface and the read helper can be shared outright rather than duplicated.

Rename `AuthorizationHandlerResourceScopeService` (`handler_authz.go:152-155`) to `ResourceAccessPolicyService` and give `TokenHandler` a field of the same type. One `wire.Bind(new(handler.ResourceAccessPolicyService), new(*resourcescope.Store))` replaces `deps_common.go:419` and serves both handlers; `TokenHandlerClientResourceScopeService` (the M2M association path, bound at `:418`) is untouched and stays a separate interface — the two mechanisms are partitioned by grant, and merging their service shapes is how that partition gets accidentally OR-ed later.

```go
// errResourceNotAvailable is what a client is told for three cases the
// callers deliberately do not distinguish: the Resource does not exist, it
// was deleted, or its access policy does not admit this client. Telling a
// client which of the three it hit would enumerate a project's Resources,
// so there is one constructor rather than three call sites that have to
// keep saying the same thing.
func errResourceNotAvailable() error {
	return protocol.NewError("invalid_target", "resource not found or not accessible to this client")
}

// allowedResourceScopes returns the scopes of resourceURI that its access
// policy currently opens to client, or errResourceNotAvailable if the
// Resource itself does not.
func allowedResourceScopes(
	ctx context.Context,
	svc ResourceAccessPolicyService,
	client *config.OAuthClientConfig,
	resourceURI string,
) ([]string, error) {
	isDynamic, isThirdParty := client.IsDynamicClient(), client.IsThirdParty()

	resource, err := svc.GetResourceByURI(ctx, resourceURI)
	if err != nil {
		if errors.Is(err, resourcescope.ErrResourceNotFound) {
			return nil, errResourceNotAvailable()
		}
		return nil, err
	}
	if !resource.AccessPolicy.AllowsClient(isDynamic, isThirdParty) {
		return nil, errResourceNotAvailable()
	}

	scopes, err := svc.ListScopesByResourceID(ctx, resource.ID)
	if err != nil {
		return nil, err
	}
	var allowed []string
	for _, s := range scopes {
		if s.AccessPolicy.AllowsClient(isDynamic, isThirdParty) {
			allowed = append(allowed, s.Scope)
		}
	}
	return allowed, nil
}
```

This is the spec's two-level check, and it is worth reading as one: the Resource level is an error, the Scope level is a filter. A Scope that admits a category its Resource does not is simply never reached — the spec's "valid state, not an error" falls out of the ordering rather than needing a rule.

Two things the shape has to preserve, both of which Part 1 §6 moved out of the database and into here:

- **The two failure paths return the identical error.** Part 1 deleted the SQL predicate that used to collapse "no such Resource" and "not permitted" into one `sql.ErrNoRows`; `errResourceNotAvailable` is what keeps them indistinguishable now that they are two branches. A reviewer should not have to compare two string literals to confirm it.
- **The message no longer names a client kind.** "not accessible to third-party clients" stops being true the moment a first-party client reaches this code, and a user-facing string that names the wrong client kind is a bug in its own right.

`errors.Is` against `ErrResourceNotFound` stays — `GetResourceByURI` still returns it, and it still has to be caught rather than surfaced as a 500.

### 2.1 Transactions

- **`/oauth2/authorize`** — `validateResource` is called from `doHandleRequestWithTx` (inside `h.Database.WithTx`, `handler_authz.go:595`) and from `doHandleConsentRequest` (deliberately outside one, `:970`). It already brackets its reads in an `IsInTx`/`ReadOnly` branch; keep that wrapper around the call to `allowedResourceScopes` rather than pushing the branch into the helper, since the token handler does not need it.
- **`/oauth2/token`** — verified: every token grant runs inside `h.Database.WithTx` (`handler_token.go:312` wrapping `doHandleWithTx`, the handler's only transaction). No guard needed; call the helper directly.

## 3. `/oauth2/authorize` — widening

`validateResource` keeps its URI-prefix check (`handler_authz.go:205-207`), its early return for an absent `resource`, and its existing `m2m` arm (`:249-252`, returning `unauthorized_client`). Only the dynamic-third-party arm changes, into a call to `allowedResourceScopes`.

Leave the `m2m` arm exactly where it is. It is pre-existing, it already carries a comment saying it is unreachable because `ValidateRequestWithoutTx` rejects `m2m` at `:1078`, and it returns before any policy is read — which is what keeps `AllowsClient` from needing an m2m case of its own (Part 1 §4).

The `default` arm is deleted, and with it Part 1 §6.1's placeholder `AllowsClient(true, true)`. Every non-`m2m` client now reaches the policy with its own `IsDynamicClient()`/`IsThirdParty()` pair, and one the Resource does not admit gets `invalid_target` — the same error as before, from a different branch. The `default` arm's long comment about static third-party clients having "their own client-resource association mechanism" for this grant goes with it: that was never true (associations are `client_credentials`-only) and the spec now says so explicitly.

`resourceScopeDisplayNames` (`handler_authz.go:282-...`) guards on `!(client.IsDynamicClient() && client.IsThirdParty())`. Drop the guard entirely and have it call `allowedResourceScopes`, keying its display names off the scopes that come back rather than re-deriving which ones to show. It still swallows errors into `nil`, which is correct for a render-time lookup running after validation has already passed — and it stops being a second hand-rolled copy of the policy check that can disagree with the first.

### 3.1 What widens, concretely

| Client | Before | After |
|---|---|---|
| Dynamic third-party | policy-checked | unchanged |
| Static third-party (`third_party_app`) | always `invalid_target` | policy-checked against `allow_static_third_party_client_access` |
| Dynamic first-party | always `invalid_target` | policy-checked against `allow_dynamic_first_party_client_access` |
| Static first-party (`spa`, `traditional_webapp`, `native`, `confidential`) | always `invalid_target` | policy-checked against `allow_static_first_party_client_access` |
| `m2m` | `unauthorized_client` | unchanged |

No Resource in any existing project has any of the three new keys set (Part 1 §2), so on upgrade every one of these still gets `invalid_target`. Nothing widens until an admin sets a key. **This is why the `false` default matters and why it is not a breaking change** — and it is worth saying in the PR description, because "first-party clients can now request resources" reads like a default-on change and is not one.

### 3.2 Two downstream behaviours that are already correct

Neither needs a code change; both need an e2e test, because this part is the first time either is reachable.

- **A first-party client's resource-bound token is a JWT with `aud = [<resource_uri>]` and no project endpoint.** `PrepareUserAccessToken` (`pkg/lib/oauth/token_encoding.go:92-118`) already keys on `ResourceURI` before it looks at `IsThirdParty()` or `IssueJWTAccessToken`, so this falls out unchanged.
- **That token still works at `/resolve`.** `resolve.go:110` gates on `clientLike.IsFirstParty` alone, and `DecodeAccessToken` no longer validates `aud`. Per [client.md's footnote](../../specs/client.md#access-token-behavior-by-client-kind), that is the decision: `/resolve` accepts any first-party client's token of any shape, and resource binding does not enter into it. Worth an explicit test precisely because it looks like an oversight and is not.

## 4. Revocation — re-reading the policy on every issuance

The spec: *"Both levels are re-read every time an access token is issued, not only when the grant is created."* Resource level failing is `invalid_target`; Scope level failing drops the scope and issues the rest.

### 4.1 Telling a resource scope from a project scope

`pkg/lib/oauth/scope.go`:

```go
// IsResourceScope reports whether scope is resource-specific rather than
// project-level. AllowedScopes is the closed set of project-level scopes,
// and ValidateScopesByClientConfig admits a resource scope by appending the
// Resource's own scopes to it -- so "not in AllowedScopes" is the same
// discriminator, used from the other side.
func IsResourceScope(scope string) bool {
	return !slices.Contains(AllowedScopes, scope)
}
```

Deriving it from `AllowedScopes` (`scope.go:40-56`) rather than from a prefix rule or a fresh list is what keeps the drop logic in step with the admit logic. A project scope is never dropped by any of this, which is the spec's "OIDC scopes are never affected".

### 4.2 The issuance-time check

```go
// resourceScopesForIssuance re-reads the access policy and returns granted
// minus any resource-specific scope the Resource no longer opens to the
// client's category. See docs/specs/api-resource.md § Revocation.
//
// Only the issued access token is narrowed. The grant keeps its full scope
// set, so re-enabling a key restores the scope on the next issuance without
// the user re-authorizing.
func (h *TokenHandler) resourceScopesForIssuance(
	ctx context.Context,
	client *config.OAuthClientConfig,
	resourceURI string,
	granted []string,
) ([]string, error) {
	if resourceURI == "" {
		return granted, nil
	}
	stillAllowed, err := allowedResourceScopes(ctx, h.ResourceAccessPolicyService, client, resourceURI)
	if err != nil {
		return nil, err
	}
	return slice.Filter(granted, func(s string) bool {
		return !oauth.IsResourceScope(s) || slices.Contains(stillAllowed, s)
	}), nil
}
```

A Resource-level failure surfaces as `allowedResourceScopes`' `invalid_target`, which is the spec's "the refresh fails with `invalid_target`. Authgear does not fall back to an unbound token". Do not rescue it into an unbound issuance — the grant stays bound, and handing the client a project-audience token it never asked for is the audience-confusion this whole mechanism exists to prevent.

### 4.3 Where it is applied — and where it must not be

**`doIssueTokensForAuthorizationCode`** — after the existing equality check (`handler_token.go:1764-1767`). Filter the value that reaches **`PrepareUserAccessGrantOptions.Scopes`** (`:2005`, currently `code.AuthorizationRequest.Scope()`).

Leave `IssueOfflineGrantRefreshTokenOptions.Scopes` (`:1898`, currently the local `scopes` from `:1800`) **unfiltered**. That is the grant, not the token. If the drop were applied there, a scope revoked for one hour would be gone from the offline grant permanently, and re-enabling the key would silently fail to restore it — a data-loss bug that only shows up long after the change that caused it. Keep the two values distinct rather than reusing one local; a reviewer should be able to see at a glance which one is narrowed.

**`issueTokensForRefreshToken`** — after the existing check (`:2066-2070`). Filter the value that reaches `PrepareUserAccessGrantOptions.Scopes` (`:2106`, currently `offlineGrantSession.Scopes`). `offlineGrantSession.Scopes` itself is untouched, for the same reason, so the next refresh re-evaluates from the full set.

Both paths run inside the token endpoint's transaction (§2.1), so this is one extra indexed read per access token issuance on a resource-bound grant only. Requests with no `resource` return at the `resourceURI == ""` guard without touching the database.

### 4.4 The client has to be told what it got

`IssueAccessGrantResult.WriteTo` (`pkg/lib/oauth/grant_access_service.go:40-46`) writes `token_type`, `access_token` and `expires_in` — **not `scope`**. The only `resp.Scope(...)` in non-test code is the `client_credentials` path (`service_token.go:447`).

So as things stand, §4.3 hands back a token with fewer scopes than the client asked for and no indication that anything was dropped — the client keeps calling an API it no longer has permission for and gets 403s it cannot explain.

**Always return `scope` on the user-grant path.** Add `Scopes []string` to `IssueAccessGrantResult`, populated from the scopes actually issued, and emit it in `WriteTo`:

```go
func (r *IssueAccessGrantResult) WriteTo(resp protocol.TokenResponse) {
	if r != nil && resp != nil {
		resp.TokenType(r.TokenType)
		resp.AccessToken(r.Token)
		resp.ExpiresIn(r.ExpiresIn)
		resp.Scope(strings.Join(r.Scopes, " "))
	}
}
```

Always rather than only-when-different: RFC 6749 §5.1 makes `scope` OPTIONAL when it matches the request and REQUIRED when it does not, so unconditional is correct in both cases, and it is one less branch whose condition can be wrong. It also matches what `client_credentials` already does, so the two grant families stop disagreeing about whether a token response says what it granted.

Three call sites inherit this, all of them OAuth token responses: the token endpoint (`handler_token.go:345`), anonymous user signup (`pkg/auth/handler/api/anonymous_user_signup.go:146`), and the Admin API's session-token facade (`pkg/admin/facade/oauth.go:135`). Adding a standard field to each is additive.

This also closes the pre-existing gap where the user-grant response omitted `scope` even when nothing was dropped.

## 5. Test plan

Unit (Convey, extending the existing suites):

- `pkg/lib/oauth/scope_test.go` — `IsResourceScope` is `false` for every member of `AllowedScopes` and `true` for `read:orders`.
- `pkg/lib/oauth/handler/handler_authz_test.go` — `validateResource` for each of the four categories against a Resource that allows it, and against one that allows a *different* category (→ `invalid_target`); `m2m` → `unauthorized_client`; a resource URI prefixed by the project endpoint → `invalid_target`; no `resource` → nil, no service call. The cross-category cases are the ones that matter: a transposed arm in `AllowsClient` shows up as a client of one category being admitted by another's key, and only a matrix catches that. Part 1 §7 tests the mapping directly; this tests that the handler feeds it the right pair.
- `pkg/lib/oauth/handler/handler_token_test.go` — **refresh with the Resource key cleared → `invalid_target`, and the offline grant is not modified**; refresh with a Scope key cleared → token issued, the resource scope absent from the access grant, `offlineGrantSession.Scopes` unchanged, and the response `scope` reflects what was issued; refresh with the key re-enabled afterwards → the scope is back (this is the assertion that pins §4.3's "grant, not token" split); refresh with everything still allowed → byte-identical to today.
- The same three for `doIssueTokensForAuthorizationCode`.

e2e — extend `e2e/tests/dcr/resource_indicator.test.yaml` for the dynamic-third-party regression, and add `e2e/tests/resource_scope/resource_indicator_categories.test.yaml`:

1. Static first-party (`spa`) + Resource/Scope granting `allowStaticFirstPartyClientAccess` → full authorization-code flow succeeds; the access token is a JWT whose `aud` is `[<resource_uri>]`, not the project endpoint.
2. The same client and Resource with the key cleared → `/oauth2/authorize` redirects with `error=invalid_target`.
3. The same client against a Resource that grants only `allowDynamicThirdPartyClientAccess` → `invalid_target`. (Cross-category, at the HTTP boundary.)
4. Static third-party (`third_party_app`) + `allowStaticThirdPartyClientAccess` → succeeds, consent screen shown.
5. `m2m` at `/oauth2/authorize` with `resource=` → `unauthorized_client`, unchanged.
6. `m2m` at `client_credentials` against a Resource with **every** `access_policy` key `true` but **no** association → still `invalid_target`. This is the partition, tested from the direction that would break first if the two mechanisms were ever OR-ed.
7. Revocation, Resource level: authorize with `resource=`, clear the key via the Admin API, refresh → `invalid_target`.
8. Revocation, Scope level: same setup, clear only the Scope's key, refresh → 200, the resource scope gone from the response `scope` and from the token's `scope` claim, OIDC scopes intact.
9. Re-enable the Scope key, refresh again → the scope is back.
10. `/resolve` with a static first-party client's resource-bound JWT → valid (§3.2); `/resolve` with a static third-party client's resource-bound JWT → invalid.

## 6. No `docs/BREAKING-CHANGES.md` entry

Nothing here breaks a normal deployment.

§3 refuses nothing it did not refuse before, and permits nothing until an admin sets a key that exists on no current row.

§4 makes a cleared key take effect on the next refresh instead of at token expiry. An admin who cleared a key was asking for exactly that; the old behaviour was the bug. Depending on it would mean depending on a revocation not happening, which is not normal usage.

§4.4 adds a standard field to a token response.

Cover §4's change of timing in the PR description, since it is the one an operator would want to know about.

## 7. Commit plan

1. **`Share one access policy read path between the authorize and token handlers`** — §2 (interface rename, `TokenHandler` field, `allowedResourceScopes`, `errResourceNotAvailable`, the `wire.Bind`, regenerated mocks). Behaviour-identical: `validateResource` still reaches it only for a dynamic third-party client.
2. **`Allow every client category to request a resource`** — §3, including dropping Part 1 §6.1's placeholder pair, plus its unit tests. The widening, on its own, so it can be reverted without taking the revocation fix with it.
3. **`Return the granted scope in the token response`** — §4.4, on its own: it touches three grant families and is worth reverting independently if anything downstream turns out to parse the response strictly.
4. **`Re-read the access policy on every access token issuance`** — §4.1-§4.3 and its unit tests. Lands after 3 so the scope-drop assertions can check the response, not just the persisted grant.
5. **`Add e2e tests for access policy categories and revocation`** — §5.
