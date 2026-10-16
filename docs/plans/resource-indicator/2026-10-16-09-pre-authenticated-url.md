# Resource Indicator Part 9 — Pre-Authenticated URL

Spec: [docs/specs/pre-authenticated-url.md — Resource Indicators](../../specs/pre-authenticated-url.md#resource-indicators), [Access Token Audience Binding — Derived grant](../../specs/access-token-audience-binding.md#derived-grant).

Depends on Part 5 (`PreAuthenticatedURLTokenServiceImpl.ResourceAccessPolicy`, issuance-time stripping) and Part 8 (`derivedGrantResourceScopes`).

## 1. Goal and scope

The token exchange (`grant_type=urn:ietf:params:oauth:grant-type:token-exchange`, `requested_token_type=urn:authgear:params:oauth:token-type:pre-authenticated-url-token`) gains an optional `resource`. The browser session's tokens, issued later at `/oauth2/authorize`, are bound to it.

- **Original refresh tokens:** the refresh tokens of the session's initial client (`OfflineGrant.InitialClientID`) in the offline grant named by the `subject_token`'s `sid`.
- **New client:** the token exchange `client_id` (the web app).

`/oauth2/authorize` takes no `resource` for this flow.

## 2. Model and storage

`pkg/lib/oauth/pre_authenticated_url_token.go`, `PreAuthenticatedURLToken`:

```go
ResourceURI string `json:"resource_uri,omitempty"`
```

Stored as JSON in Redis by `pkg/lib/oauth/redis/store.go` (`CreatePreAuthenticatedURLToken`, `ConsumePreAuthenticatedURLToken`); no store change.

`pkg/lib/oauth/grant_offline.go`, next to `getScopesSet`:

```go
// ScopesByResourceURI groups the scopes of clientID's refresh tokens by the
// Resource each is bound to, "" for unbound.
func (g *OfflineGrant) ScopesByResourceURI(clientID string) map[string][]string
```

- If `g.InitialClientID == clientID`, add `g.Deprecated_Scopes` under `""` (legacy grants have no `RefreshTokens` and are unbound).
- For each `token` in `g.RefreshTokens` with `token.ClientID == clientID`, add `token.Scopes` under `token.ResourceURI`.
- Deduplicate within each key.

## 3. Runtime changes

### 3.1 Token exchange

`service_preauthenticated_url.go`: `IssuePreAuthenticatedURLTokenOptions` gains `ResourceURI string`; `IssuePreAuthenticatedURLToken` copies it into the model.

`TokenHandler.handlePreAuthenticatedURLToken` (`handler_token.go`), after `verifyIDTokenDeviceSecretHash`. This replaces the `requestedScopes` block and the `CheckAndGrant` call:

1. Only when `r.Resource() != ""`: `originalClient := h.ClientResolver.ResolveClient(ctx, offlineGrant.InitialClientID)`; `nil` → `protocol.NewError("server_error", "cannot find original client for pre-authenticated url")`. A request without `resource` does not depend on the original client's config, as before.
2. `originals := offlineGrant.ScopesByResourceURI(offlineGrant.InitialClientID)`.
3. `base := scopes`, which is today's `offlineGrant.GetScopes(offlineGrant.InitialClientID)`.
4. If `len(r.Scope()) > 0`:
   1. `offlineGrant.HasAllScopes(...)` → `invalid_scope` (unchanged).
   2. `var known []string`; if `r.Resource() != ""`, `known = slice.Filter(originals[r.Resource()], oauth.IsResourceScope)`.
   3. `oauth.ValidateScopesByClientConfig(r.Scope(), known)` → return its error. A Resource Scope the session holds only on another Resource is unknown here.
   4. `base = r.Scope()`.
5. `projectScopes := oauth.GrantedScopes(client, base, nil)` when `scope` was requested (as today); otherwise `slice.Filter(base, func(s string) bool { return !oauth.IsResourceScope(s) })`.
6. `resourceID, resourceScopes, err := derivedGrantResourceScopes(ctx, h.ResourceAccessPolicyService, originalClient, client, originals, r.Resource())`. When `scope` was requested, keep only those in `r.Scope()`.
7. `scopes = append(projectScopes, resourceScopes...)`.
8. `CheckAndGrant(ctx, client.ClientID, offlineGrant.GetUserID(), resourceID, scopes)` (was `""`).
9. `IssuePreAuthenticatedURLTokenOptions{..., Scopes: scopes, ResourceURI: r.Resource()}`.

### 3.2 Browser session issuance

`PreAuthenticatedURLTokenServiceImpl.ExchangeForAccessToken`: Part 5 calls `resourceScopesForIssuance(ctx, s.ResourceAccessPolicy, client, "", tokenModel.Scopes)`. Replace `""` with `tokenModel.ResourceURI`, and set `ResourceURI: tokenModel.ResourceURI` in both `CreateNewRefreshTokenOptions` and `PrepareUserAccessGrantOptions`. The policy is re-read here for the web app, per [Revocation](../../specs/api-resource.md#revocation).

## 4. Compatibility and deployment

| Case | Behaviour |
|---|---|
| Token written by the old version, consumed by the new | No `resource_uri` → unbound; Resource Scopes stripped (Part 5) |
| Token written by the new version, consumed by an old pod during rollout | `resource_uri` ignored → unbound, carrying the Resource's Scopes (pre-Part 5 behaviour) |

The window is the token lifetime (`oauth.PreAuthenticatedURLTokenLifetime`, `duration.Short`). No dual-write or backfill.

Without `resource`, behaviour is as after Part 5.

## 5. Tests

Unit (Convey), `pkg/lib/oauth/grant_offline_test.go`, for `ScopesByResourceURI`:

- legacy grant (`Deprecated_Scopes` only) → under `""`
- two refresh tokens of the client bound to different Resources → two keys
- another client's refresh tokens are excluded

E2E: `e2e/tests/oidc/pre_authenticated_url_resource.test.yaml`. The browser's cookie token is not observable in e2e; its scopes and binding are covered by the unit tests of `derivedGrantResourceScopes` (Part 8) and `resourceScopesForIssuance` (Part 5).

| Step | Expect |
|---|---|
| Token exchange with `resource=R` | 200; `/oauth2/authorize` with the token → 303 without `error=` |
| Token exchange with an unrelated Resource | `invalid_target` |
| Token exchange with `scope` containing `read:orders` and no `resource` | `invalid_scope` |
| Authorization query | `read:orders` recorded on R for the web client only after the `resource=R` exchange |

## 6. Commit

**`Accept resource on the pre-authenticated URL token exchange`**: `pre_authenticated_url_token.go`, `grant_offline.go`, `grant_offline_test.go`, `service_preauthenticated_url.go`, `handler_token.go`, `pre_authenticated_url_resource.test.yaml`. No wiring changes (the dependency was added in Part 5).
