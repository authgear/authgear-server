# Resource Indicator Part 7 — Anonymous user

Spec: [docs/specs/anonymous-user.md — Resource Indicators](../../specs/anonymous-user.md#resource-indicators), [docs/specs/api.md — /api/anonymous_user/signup](../../specs/api.md#apianonymous_usersignup), [Access Token Audience Binding — New grant](../../specs/access-token-audience-binding.md#new-grant).

Depends on Part 5 (`resourceScopesForIssuance`) and Part 6 (`newGrantScopes`, `checkOfflineFullAccessScopes`).

## 1. Goal and scope

| Entry point | Change |
|---|---|
| `/oauth2/token`, `grant_type=urn:authgear:params:oauth:grant-type:anonymous-request` (mobile SDK) | New optional `resource` and `scope` |
| `/api/anonymous_user/signup`, `session_type: refresh_token`, new user (Web SDK) | New optional `resource` and `scope` |
| `/api/anonymous_user/signup`, `session_type: refresh_token`, with `refresh_token` | `resource` follows the `refresh_token` grant; `scope` is ignored, as in that grant |
| `/api/anonymous_user/signup`, `session_type: cookie` | `resource` and `scope` rejected by the request schema |

`scope` defaults to `openid offline_access https://authgear.com/scopes/full-access`, and must include `offline_access` and `https://authgear.com/scopes/full-access` when given.

Promotion goes through `/oauth2/authorize`, which already supports `resource`; no change.

## 2. Token grant

`TokenHandler.handleAnonymousRequest` (`pkg/lib/oauth/handler/handler_token.go`), in order:

1. Full-access client check, `r.DeviceInfo()` (unchanged).
2. New, **before `Graphs.DryRun`**, because `Graphs.Run` creates the user:
   ```go
   resourceID, scopes, err := newGrantScopes(ctx, h.ResourceAccessPolicyService, h.IDTokenIssuer.Iss(), client, r.Resource(), r.Scope(),
       []string{oauth.ScopeOpenID, oauth.OfflineAccess, oauth.FullAccessScope})
   ```
   then `checkOfflineFullAccessScopes(scopes)`.
3. Remove `// TODO(oauth): allow specifying scopes` and the hard-coded `scopes`.
4. `CheckAndGrant(ctx, client.ClientID, info.UserID, resourceID, scopes)`.
5. `IssueOfflineGrantOptions.ResourceURI = r.Resource()`, `PrepareUserAccessGrantOptions.ResourceURI = r.Resource()`.

## 3. Web SDK API

### 3.1 HTTP layer

`pkg/auth/handler/api/anonymous_user_signup.go`:

- `AnonymousUserSignupAPIRequestSchema`: add `"resource": { "type": "string" }` and `"scope": { "type": "string" }`, and append to `allOf`:
  ```json
  {
    "if": { "properties": { "session_type": { "const": "cookie" } } },
    "then": { "properties": { "resource": false, "scope": false } }
  }
  ```
- `AnonymousUserSignupAPIRequest`: add `Resource string \`json:"resource"\`` and `Scope string \`json:"scope"\``.
- The `AnonymousUserHandler` interface in this file: `SignupAnonymousUser` gains `resource string, scopes []string`.
- `ServeHTTP` passes `payload.Resource` and `strings.Fields(payload.Scope)`. The token endpoint's splitter (`parseSpaceDelimitedString`) is unexported in `protocol`.

### 3.2 Handler

`pkg/lib/oauth/handler/handler_anonymous_user.go`:

- `AnonymousUserHandler` gains `IDTokenIssuer IDTokenIssuer` and `ResourceAccessPolicyService ResourceAccessPolicyService`. Both interfaces are already bound in `pkg/lib/deps/deps_common.go`; regenerate wire.
- `SignupAnonymousUser(ctx, req, clientID string, sessionType WebSessionType, refreshToken string, resource string, scopes []string)`. The cookie branch does not take them.
- `signupAnonymousUserWithRefreshTokenSessionType(ctx, req, clientID, refreshToken, resource string, scopes []string)`:

**With `refreshToken`** (follows the `refresh_token` grant):

1. Parse the refresh token, check the user is anonymous, `grant.ToSession(hash)` (unchanged).
2. `resource != "" && resource != session.ResourceURI` → `apierrors.NewInvalid("resource must be a subset of the resource originally authorized")`.
3. `issued, err := resourceScopesForIssuance(ctx, h.ResourceAccessPolicyService, client, session.ResourceURI, session.Scopes)`; on error return `oauthErrorToAPIError(err)`.
4. `PrepareUserAccessGrantOptions{Scopes: issued, ResourceURI: session.ResourceURI, ...}`. This replaces the hard-coded `scopes`; existing sessions hold that same default list.

**New user**:

1. **Before `runSignupAnonymousUserGraph`**: `resourceID, scopes, err := newGrantScopes(ctx, h.ResourceAccessPolicyService, h.IDTokenIssuer.Iss(), client, resource, scopes, defaults)`, then `checkOfflineFullAccessScopes(scopes)`. Both errors go through `oauthErrorToAPIError`.
2. `CheckAndGrant(ctx, client.ClientID, info.UserID, resourceID, scopes)`.
3. `IssueOfflineGrantOptions.ResourceURI = resource`, `PrepareUserAccessGrantOptions.ResourceURI = resource`.

New helper in the same file:

```go
// oauthErrorToAPIError reports an OAuth error from the shared resource
// checks in this API's error format.
func oauthErrorToAPIError(err error) error
```

If `errors.As(err, &oauthErr)` with `oauthErr *protocol.OAuthProtocolError`, return `apierrors.NewInvalid(oauthErr.Error())`; otherwise return `err` unchanged. Without it `ServeHTTP` treats the error as unexpected, logs it at error level, and answers 500.

## 4. Compatibility

- Requests without `resource` and `scope` behave as today.
- The token grant used to ignore `scope`. A client that sends one now has it validated, and fails with `invalid_scope` if it lacks `offline_access` or full-access.
- `/api/anonymous_user/signup` used `additionalProperties: false`, so no existing caller sends the new fields.

## 5. Tests

E2E helper: new `e2e/pkg/testrunner/anonymous.go`, modelled on `app2app.go`, registered in `makeTemplateFuncMap`:

- `generateAnonymousRequestJWT(challenge string)`: header `typ: vnd.authgear.anonymous-request`, `alg: ES256`, `kid`, `jwk`; payload `challenge`, `action: auth`.

Challenges come from `POST /oauth2/challenge` with `purpose: anonymous_request`. Config: `authentication.identities: [login_id, anonymous]`; Resource R allowing static first-party with Scope `read:orders`. "Bound to R" is checked by a `refresh_token` grant with `resource=R` returning 200.

`e2e/tests/resource_indicator/anonymous.test.yaml` (new):

| Step | Expect |
|---|---|
| `anonymous-request` with `resource=R`, `scope=openid offline_access https://authgear.com/scopes/full-access read:orders` | 200, `scope` contains `read:orders`; refresh with `resource=R` → 200 |
| `anonymous-request` with `scope=openid https://authgear.com/scopes/full-access` | `invalid_scope` |
| `anonymous-request` with neither | 200, `scope` = the three defaults |

`e2e/tests/resource_indicator/anonymous_web.test.yaml` (new):

| Step | Expect |
|---|---|
| `POST /api/anonymous_user/signup` `{client_id, session_type: refresh_token, resource: R, scope: …read:orders}` | tokens; refresh with `resource=R` → 200 |
| Same with the returned `refresh_token` and an unrelated `resource` | error |
| `{session_type: cookie, resource: R}` | validation error |

## 6. Commits

1. **`Accept resource and scope on the anonymous-request grant`**: `handler_token.go`, `e2e/pkg/testrunner/anonymous.go`, `testcase.go`, `anonymous.test.yaml`.
2. **`Accept resource and scope on anonymous user signup`**: `anonymous_user_signup.go`, `handler_anonymous_user.go`, regenerated `wire_gen.go` files, `anonymous_web.test.yaml`.
