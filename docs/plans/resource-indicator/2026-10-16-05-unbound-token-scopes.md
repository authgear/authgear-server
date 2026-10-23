# Resource Indicator Part 5 — No Resource Scopes on unbound tokens

Spec: [docs/specs/access-token-audience-binding.md — Derived grant](../../specs/access-token-audience-binding.md#derived-grant) ("Omitted: … Every Scope of a Resource is dropped").

Depends on Part 3 (`resourceScopesForIssuance` and the per-Resource Authorization).

Parts 6–9 extend the remaining grants; this part is the bug fix they build on.

## 1. Goal and scope

Today app2app and Pre-Authenticated URL copy the original refresh token's scopes, Resource Scopes included, into tokens that are not bound to a Resource. Such a token has `aud = [<project endpoint>]` and a scope like `read:orders`, which only means something on its own Resource. After this part, an unbound access token never carries a Resource Scope.

| Path | Before | After |
|---|---|---|
| App2app code | Scope = `authz.AllScopes()`: the new client's whole Authorization, on every Resource | Scope = the original refresh token's scopes minus Resource Scopes |
| Pre-Authenticated URL browser tokens | Every scope of the mobile app's session | Same, minus Resource Scopes |
| Any access token issued unbound (code exchange, refresh, Pre-Authenticated URL) | Scopes issued as stored | Resource Scopes stripped at issuance |

The issuance-time strip also cleans refresh tokens that already leaked scopes: they stay unbound, and their next access token drops the Resource Scopes. Before this fix `/oauth2/authorize` could not put a Resource Scope on an unbound grant (`narrowScopes` rejects it), so no legitimate unbound grant loses a scope.

## 2. Runtime changes

### 2.1 `resourceScopesForIssuance` becomes a package function

`pkg/lib/oauth/handler/resource_access_policy.go`. Current: `func (h *TokenHandler) resourceScopesForIssuance(ctx, client, resourceURI string, granted []string) ([]string, error)`, returns `granted` unchanged when `resourceURI == ""`.

New:

```go
func resourceScopesForIssuance(
	ctx context.Context,
	svc ResourceAccessPolicyService,
	client *config.OAuthClientConfig,
	resourceURI string,
	granted []string,
) ([]string, error)
```

- `resourceURI == ""` → `slice.Filter(granted, func(s string) bool { return !oauth.IsResourceScope(s) })`.
- Otherwise unchanged: `resourceScopes(ctx, svc, client, resourceURI)`, keep project-level scopes and the Resource Scopes still in `allowed`.

It becomes a package function so `PreAuthenticatedURLTokenServiceImpl` (here) and `AnonymousUserHandler` (Part 7) can call it.

Call sites in `handler_token.go`, both passing `h.ResourceAccessPolicyService`:

- `doIssueTokensForAuthorizationCode`: `resourceScopesForIssuance(ctx, h.ResourceAccessPolicyService, client, resourceURI, code.AuthorizationRequest.Scope())`
- `issueTokensForRefreshToken`: `resourceScopesForIssuance(ctx, h.ResourceAccessPolicyService, client, resourceURI, offlineGrantSession.Scopes)`

### 2.2 App2app code scope

`TokenHandler.handleApp2AppRequest`:

1. Replace `scopes := offlineGrantSession.Scopes` with
   ```go
   // FIXME(DEV-1430): The new scopes should be validated against the new client
   scopes := slice.Filter(offlineGrantSession.Scopes, func(s string) bool { return !oauth.IsResourceScope(s) })
   ```
2. Delete the `resourceID` lookup block. `CheckAndGrant(ctx, client.ClientID, userID, "", scopes)`.
3. `artificialAuthorizationRequest["scope"] = strings.Join(scopes, " ")` (was `authz.AllScopes()`). Per [app2app — Resource Indicators](../../specs/app2app.md#resource-indicators), the code carries the original refresh token's scopes.

The `invalid_target` today's lookup returns ("Resource no longer exists or no longer admits the original client") goes away with the block. Part 8 brings the check back for requests that pass `resource`.

### 2.3 Pre-Authenticated URL

`PreAuthenticatedURLTokenServiceImpl` (`service_preauthenticated_url.go`) gains:

```go
ResourceAccessPolicy ResourceAccessPolicyService
```

`ExchangeForAccessToken`, after the client and session checks:

```go
scopes, err := resourceScopesForIssuance(ctx, s.ResourceAccessPolicy, client, "", tokenModel.Scopes)
```

Use `scopes` in both `CreateNewRefreshTokenOptions.Scopes` and `PrepareUserAccessGrantOptions.Scopes`. Part 9 replaces `""` with `tokenModel.ResourceURI`.

`ResourceAccessPolicyService` is already bound in `pkg/lib/deps/deps_common.go`. Regenerate wire.

`handlePreAuthenticatedURLToken` is unchanged here: the stored token still holds the session's scopes, and issuance strips them.

## 3. Compatibility and deployment

- No persisted format changes.
- App2app codes and Pre-Authenticated URL tokens created by the old version before the deploy are unbound, so issuance strips their Resource Scopes.
- Refresh tokens already carrying leaked Resource Scopes keep them in storage; every access token issued from them drops them.
- No `docs/BREAKING-CHANGES.md` entry: this is a bug fix, and the resource-indicator use of these flows was never supported.

## 4. Tests

Unit (Convey, package `handler`). New `pkg/lib/oauth/handler/resource_access_policy_test.go`, reusing `stubResourceScopeService` from `handler_authz_validate_resource_test.go`:

- unbound: Resource Scopes stripped, project scopes kept
- bound: a Scope the policy no longer allows is dropped
- bound: Resource no longer allows the client → `invalid_target`

E2E, updated for the fixed behaviour:

- `e2e/tests/oidc/app2app_resource.test.yaml`: rename to "app2app from a resource-bound refresh token issues no resource scopes without resource". The redeemed token's `scope` has no `read:orders`; the `app2apptarget` Authorization records no Resource Scope. Drop the second-app2app-from-an-unbound-token step, which only exercised the leak.
- `e2e/tests/oidc/pre_authenticated_url_resource.test.yaml`: update the comments to say the browser session's tokens carry no Resource Scope. Keep the token exchange and `/oauth2/authorize` steps; the Authorization still holds `read:orders` on R from the original login.

## 5. Commit

**`Fix resource scopes on unbound app2app and pre-authenticated URL tokens`**

- `pkg/lib/oauth/handler/resource_access_policy.go`, `handler_token.go`, `service_preauthenticated_url.go`
- regenerated `wire_gen.go` files
- `pkg/lib/oauth/handler/resource_access_policy_test.go`
- the two e2e files

Run `go test ./pkg/lib/oauth/...` and `make lint`; run `make update-vettedpositions` if line positions moved.
