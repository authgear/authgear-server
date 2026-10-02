# Resource Indicator Part 4 — Dropping scopes the client is not granted

Spec: [docs/specs/oidc.md — Scope Validation](../../specs/oidc.md#scope-validation).

Depends on Part 3 (the access-policy checks at `/oauth2/authorize` and `/oauth2/token` that this part relaxes from rejecting to dropping).

## 1. Goal and scope

Today every scope a client may not have fails the request with `invalid_scope`. After this change:

| Requested scope | Before | After |
|---|---|---|
| Unknown: neither in `oauth.AllowedScopes` nor a Scope of the requested Resource | `invalid_scope` | `invalid_scope` (unchanged) |
| Built-in scope the client is not granted | `invalid_scope` (except `full-userinfo`, which was granted to everyone) | Dropped |
| Scope of the requested Resource, not opened by the access policy / not in the association | `invalid_scope` | Dropped |
| `openid` missing; `pre-authenticated-url` without `device_sso`; scope outside the exchanged grant | `invalid_scope` | `invalid_scope` (unchanged) |

Entry points in scope:

- `/oauth2/authorize`: `AuthorizationHandler.doHandleRequestWithTx` and `AuthorizationHandler.doHandleConsentRequest`
- `/oauth2/token`: `TokenHandler.handlePreAuthenticatedURLToken`, `TokenHandler.handleBiometricAuthenticate`, `TokenHandler.handleClientCredentials`

Out of scope: the `refresh_token` grant takes no `scope` parameter, and its Scope-level re-check in `TokenHandler.resourceScopesForIssuance` already drops. Grants that already exist are not re-filtered (see §5).

## 2. Config

No config change.

## 3. Runtime

### 3.1 `pkg/lib/oauth/scope.go`

Replace the current permission checks with one validator that knows nothing about the client, plus one filter.

```go
// ValidateScopes returns invalid_scope for the first scope not in knownScopes.
func ValidateScopes(scopes []string, knownScopes []string) error
```

The signature is unchanged. The error description changes from `specified scope is not allowed: %s` to `unknown scope: %s`.

```go
// ValidateScopesByClientConfig rejects unknown scopes and malformed scope
// combinations. Whether the client is granted each scope is GrantedScopes.
func ValidateScopesByClientConfig(scopes []string, knownResourceScopes []string) error
```

- The `client` parameter is removed, because no check needs it anymore.
- Known set: `oauth.AllowedScopes` plus `knownResourceScopes`. If it's empty, use `oauth.AllowedScopes` alone, as today.
- Checks, in order:
  1. `ValidateScopes(scopes, known)`
  2. `PreAuthenticatedURLScope` present without `DeviceSSOScope` → `invalid_scope`, `device_sso must be requested when using pre-authenticated url`
  3. `ScopeOpenID` absent → `invalid_scope`, `must request 'openid' scope`
- Removed: the `offline access is not allowed`, `full access is not allowed`, `device_sso is not allowed` and `pre-authenticated url is not allowed` errors. `GrantedScopes` handles these cases now.

```go
// GrantedScopes returns scopes without those client is not granted.
// See docs/specs/oidc.md § Scope Validation.
func GrantedScopes(client *config.OAuthClientConfig, scopes []string, grantedResourceScopes []string) []string

func isBuiltinScopeGranted(client *config.OAuthClientConfig, scope string) bool
```

`GrantedScopes` keeps scope `s`, in request order, when one of these holds:
- `slices.Contains(AllowedScopes, s)` and `isBuiltinScopeGranted(client, s)`, or
- `!slices.Contains(AllowedScopes, s)` and `slices.Contains(grantedResourceScopes, s)`.

`isBuiltinScopeGranted` returns:

| Condition | Result |
|---|---|
| `client.ApplicationType == config.OAuthClientApplicationTypeM2M` | `false` |
| `OfflineAccess` | `slices.Contains(GetAllowedGrantTypes(client), RefreshTokenGrantType)`, always `true` here because `refresh_token` is allowed for every non-`m2m` client |
| `FullAccessScope` | `client.HasFullAccessScope()`, also `true` for a client without `x_application_type` |
| `FullUserInfoScope` | `client.IsFirstParty()` (new: third-party clients were not filtered before) |
| `DeviceSSOScope`, `PreAuthenticatedURLScope` | `client.PreAuthenticatedURLEnabled` |
| any other built-in scope | `true` |

### 3.2 `pkg/lib/oauth/protocol/authz.go`

```go
func (r AuthorizationRequest) SetScope(scopes []string) { r["scope"] = strings.Join(scopes, " ") }
```

The authorize path narrows the request in place, not by passing a separate scope list. The OAuth session entry, the consent screen (`ConsentRequired.Scopes`), `Authorizations.CheckAndGrant`, the code grant and `doIssueTokensForAuthorizationCode` all read `AuthorizationRequest.Scope()`. Narrowing the request once means they all see the same granted set.

### 3.3 Resource scopes: `pkg/lib/resourcescope`

`access_policy_service.go`:

```go
// ListAllAndAllowedScopesByResourceID returns every Scope of resourceID,
// and the subset its access policy opens to client.
func (s *AccessPolicyService) ListAllAndAllowedScopesByResourceID(ctx context.Context, resourceID string, client model.ClientCategoryClassifier) (all []*Scope, allowed []*Scope, err error)
```

`ListScopesByResourceID` is reimplemented as a call to this method that returns `allowed`, so the access-policy filter lives in one place. It is still used by `AuthorizationHandler.resourceScopeDisplayNames`.

`client_resource_service.go`:

```go
// ListResourceScopes returns every Scope of resourceID, associated with the client or not.
func (s *ClientResourceScopeService) ListResourceScopes(ctx context.Context, resourceID string) ([]*Scope, error)
```

This calls `s.Store.ListScopesByResourceID(ctx, resourceID)`.

Both reads are one `SELECT` on `_auth_resource_scope` by `resource_id`. This is the same query `ListScopesByResourceID` already runs, so the authorize path costs no extra query.

### 3.4 `pkg/lib/oauth/handler/resource_access_policy.go`

- Add `ListAllAndAllowedScopesByResourceID` to the `ResourceAccessPolicyService` interface.
- Rename `allowedResourceScopes` to `resourceScopes`:

```go
// resourceScopes returns every Scope name of resourceURI and the subset its
// access policy opens to client, or errResourceNotAvailable if the Resource
// itself does not.
func resourceScopes(ctx context.Context, svc ResourceAccessPolicyService, client *config.OAuthClientConfig, resourceURI string) (resource *resourcescope.Resource, known []string, allowed []string, err error)
```

- `TokenHandler.resourceScopesForIssuance` calls `resourceScopes` and ignores `known`. Its behavior is unchanged.

### 3.5 `pkg/lib/oauth/handler/handler_authz.go`

`validateResource` now returns `(resourceID string, knownScopes []string, grantedScopes []string, err error)` from `resourceScopes`. When `resource` is absent it returns `"", nil, nil, nil`.

`doHandleRequestWithTx`:

1. `resourceID, knownResourceScopes, grantedResourceScopes, err := h.validateResource(ctx, client, r)`. Errors return as today (`invalid_target`).
2. `oauth.ValidateScopesByClientConfig(r.Scope(), knownResourceScopes)`. Errors return `invalid_scope`.
3. `r.SetScope(oauth.GrantedScopes(client, r.Scope(), grantedResourceScopes))`
4. The rest is unchanged. The settings-action and normal session entries are saved after step 3, so both store the narrowed request.

`doHandleConsentRequest`: same three steps, on `opts.ConsentRequest.OAuthSessionEntry.T.AuthorizationRequest`. The access policy may have changed since `/oauth2/authorize`, so the stored request is narrowed again. `doHandleConsent` builds `ConsentRequired.Scopes` from the same map after this call, so the consent screen lists the narrowed set. The entry is not re-saved: the next consent request runs the same steps.

### 3.6 `pkg/lib/oauth/handler/handler_token.go`

Add `ListResourceScopes(ctx context.Context, resourceID string) ([]*resourcescope.Scope, error)` to the `TokenHandlerClientResourceScopeService` interface.

`handlePreAuthenticatedURLToken` (around line 971):
1. The `offlineGrant.HasAllScopes` check stays first and still returns `invalid_scope` (RFC 6749 §6).
2. `oauth.ValidateScopesByClientConfig(requestedScopes, nil)`
3. `scopes = oauth.GrantedScopes(client, requestedScopes, nil)`

`handleBiometricAuthenticate` (around line 1394):
1. `oauth.ValidateScopesByClientConfig(requestedScopes, nil)`
2. `scopes = oauth.GrantedScopes(client, requestedScopes, nil)`
3. The existing `ContainsAllScopes(scopes, {offline_access, full-access})` check then fails with its current error when either scope was dropped. This matches the spec: "same error as when it is not requested".

`handleClientCredentials` (around line 2361), only when `len(r.Scope()) > 0`:
1. `resourceScopeList, err := h.ClientResourceScopeService.ListResourceScopes(ctx, resource.ID)`
2. `known := append(slices.Clone(oauth.AllowedScopes), <names of resourceScopeList>...)`
3. `oauth.ValidateScopes(r.Scope(), known)`. An unknown scope returns `invalid_scope`.
4. `scopes = slice.Filter(r.Scope(), func(s string) bool { return slices.Contains(allowedScopeStrs, s) })`

Built-in scopes are never in an association, so step 4 drops them, which matches "`m2m` is granted no built-in scope". If every requested scope is dropped, the token is issued with an empty `scope`. When `scope` is omitted, the behavior is unchanged: the token gets every associated scope.

### 3.7 Token response

`IssueAccessGrantResult.WriteTo` and `TokenService.IssueClientCredentialsAccessToken` already write `scope` from the issued scopes. No change.

## 4. Storage and migration

No schema, Redis key or payload change. OAuth session entries and code grants still store `AuthorizationRequest`. They now hold the narrowed `scope`.

## 5. Compatibility and deployment

- **Requests that succeed today are unaffected,** with one exception: a third-party client requesting `https://authgear.com/scopes/full-userinfo` now has it dropped. That is the bug fix, and no `docs/BREAKING-CHANGES.md` entry is needed.
- **Requests that fail today with a permission `invalid_scope` now succeed** with fewer scopes. Per the decision recorded in §8, there is no `docs/BREAKING-CHANGES.md` entry.
- **In-flight OAuth sessions and codes during deploy** were created under the stricter rule, so they contain no scope the client lacks. No dual handling is needed.
- **Existing authorizations and offline grants** keep the scopes they already store, including `full-userinfo` on a third-party grant. They are not re-filtered.

## 6. File-level change plan

| File | Change |
|---|---|
| `pkg/lib/oauth/scope.go` | `ValidateScopes` message; new `ValidateScopesByClientConfig` signature and checks; add `GrantedScopes`, `isBuiltinScopeGranted` |
| `pkg/lib/oauth/protocol/authz.go` | Add `AuthorizationRequest.SetScope` |
| `pkg/lib/resourcescope/access_policy_service.go` | Add `ListAllAndAllowedScopesByResourceID`; `ListScopesByResourceID` delegates to it |
| `pkg/lib/resourcescope/client_resource_service.go` | Add `ListResourceScopes` |
| `pkg/lib/oauth/handler/resource_access_policy.go` | Interface method; `allowedResourceScopes` → `resourceScopes` |
| `pkg/lib/oauth/handler/handler_authz.go` | `validateResource` return values; narrowing in `doHandleRequestWithTx` and `doHandleConsentRequest` |
| `pkg/lib/oauth/handler/handler_token.go` | Interface method; three call sites in §3.6; `resourceScopesForIssuance` uses `resourceScopes` |
| `pkg/lib/oauth/handler/resource_access_policy_mock_test.go`, `handler_token_mock_test.go` | Regenerated by `go generate ./pkg/lib/oauth/handler/...` |

## 7. Test plan

Unit tests use Convey, matching the existing files.

`pkg/lib/oauth/scope_test.go`:
- Rewrite `TestValidateScopesByClientConfig` for the new signature:
  - unknown scope → `unknown scope: x`
  - resource scope with `knownResourceScopes` → ok
  - resource scope without it → error
  - missing `openid` → error
  - `pre-authenticated-url` without `device_sso` → error
- Add `TestGrantedScopes`:
  - `spa` keeps `full-access`; `confidential` drops it
  - a client without `refresh_token` in `grant_types` keeps `offline_access`, since that grant is always allowed
  - `third_party_app` and `x_dynamic_third_party` drop `full-userinfo`; first-party keeps it
  - `device_sso` and `pre-authenticated-url` are dropped unless `PreAuthenticatedURLEnabled`
  - `m2m` drops every built-in scope
  - a resource scope is kept only when it is in `grantedResourceScopes`
  - request order is preserved

`pkg/lib/oauth/handler/handler_authz_validate_resource_test.go`:
- Mock expectations move from `ListScopesByResourceID` to `ListAllAndAllowedScopesByResourceID`.
- Assert that `validateResource` returns both lists.

`pkg/lib/oauth/handler/handler_authz_test.go`:
- The existing `must request 'openid' scope` cases stay.
- Add: a client without `x_pre_authenticated_url_enabled` requesting `openid device_sso https://authgear.com/scopes/pre-authenticated-url` is not rejected, and the saved session entry's scope is `openid`.

`pkg/lib/oauth/handler/handler_token_test.go`:
- The client-credentials `admin` case expects `ListResourceScopes` and `unknown scope: admin`.
- Add: a Scope of the Resource that is not in the association → 200, `scope` omits it.
- Add: `openid` plus an associated scope → 200, `scope` is the associated scope only.
- Update the `resourceScopesForIssuance` mocks to the new service method.

E2E (`e2e/tests/`, YAML):
- `resource_indicator/static_clients.test.yaml`: replace "requesting a Scope without its own access policy key gets invalid_scope" with "…has the Scope dropped". Use `oauth_setup` with `openid read:partial`, run the login flow from the first test in the file, then `oauth_exchange_code` and assert `access_token_claims.scope` is `openid`. Add a case with an unknown scope (`openid read:missing` on the same Resource) → `error=invalid_scope`.
- `dcr/resource_indicator.test.yaml`: change "DCR third-party client requesting a scope without the access policy enabled gets invalid_scope" in the same way. The DCR client goes through consent, so assert on the token after the consent step that the file's success case already uses.
- `m2m/token.test.yaml`: rename "client_credentials flow with invalid scope" to "…with a built-in scope". Request `openid resource2_scope1` and assert 200 with `"scope": "resource2_scope1"`. Add "…with unknown scope" requesting `unknown_scope` → 400 `invalid_scope`.

Commands: `go test ./pkg/lib/oauth/... ./pkg/lib/resourcescope/...`, then `make -C e2e run` for the three files.

## 8. Fixed behavioral decisions

- Unknown scopes are rejected even though OIDC says SHOULD ignore. A typo fails loudly.
- Clients the Resource admits can tell from the response which scope names exist on it. This is accepted.
- No `docs/BREAKING-CHANGES.md` entry: the only newly dropped scope on a succeeding request is the `full-userinfo` bug fix, and third-party clients are not in production use.
- Existing grants are not re-filtered.

## 9. Atomic commit plan

1. **`Add AccessPolicyService.ListAllAndAllowedScopesByResourceID and ClientResourceScopeService.ListResourceScopes`**
   - Files: `pkg/lib/resourcescope/access_policy_service.go`, `pkg/lib/resourcescope/client_resource_service.go`
   - Additive, with no caller yet apart from `ListScopesByResourceID` delegating. No wiring change, because both services are already provided.
2. **`Drop scopes the client is not granted instead of rejecting them`**
   - Files: `pkg/lib/oauth/scope.go`, `pkg/lib/oauth/scope_test.go`, `pkg/lib/oauth/protocol/authz.go`, `pkg/lib/oauth/handler/resource_access_policy.go`, `pkg/lib/oauth/handler/handler_authz.go`, `pkg/lib/oauth/handler/handler_token.go`, the regenerated `*_mock_test.go`, `handler_authz_validate_resource_test.go`, `handler_authz_test.go`, `handler_token_test.go`
   - This is one commit because `ValidateScopesByClientConfig`'s signature change breaks every caller at once. The mocks are regenerated in the same commit.
3. **`Update e2e tests for dropped scopes`**
   - Files: `e2e/tests/resource_indicator/static_clients.test.yaml`, `e2e/tests/dcr/resource_indicator.test.yaml`, `e2e/tests/m2m/token.test.yaml`
   - These cases fail between commits 2 and 3 only in e2e, which isn't run per commit. If bisect safety across e2e matters, fold this commit into 2.
4. **`.vettedpositions` update**, if `make update-vettedpositions` reports moved positions.
