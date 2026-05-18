# Settings Action: Link OAuth — Server Implementation Plan

> Source repo: `authgear-server`. Companion plans: `connect-oauth-sdk-js-web.md`, `connect-oauth-sdk-js-rn.md`.

## Goal

Introduce a new settings action `link_oauth` that lets a client app delegate "link an OAuth provider" to the Authgear settings page, with no UI shown between the app's call and the OAuth provider's authorization page.

## End-to-end call flow

```
App → startLinkOAuth({ redirectURI, oauthProviderAlias: "google" })
  → OIDC /authorize (response_type=settings-action, x_settings_action=link_oauth, x_oauth_provider_alias=google)
  → BuildSettingsActionURL → /settings/identity/oauth?x_provider_alias=google (+ x_ref, x_settings_action_id, …)
  → Session middleware: creates webapp session (OAuthSessionID, SettingsActionID, RedirectURI), strips x_ref, 302 to clean URL
  → GET /settings/identity/oauth?x_provider_alias=google                       [Branch A]
       → StartAddIdentityOAuth → AccountManagementToken
       → GenerateState → state token
       → 302 to Google OAuth
  → Google callback → /sso/oauth2/callback
       → FinishAddingIdentityOAuth
       → 302 to /settings/identity/oauth?x_oauth_linked=1
  → GET /settings/identity/oauth?x_oauth_linked=1                              [Branch B]
       → FinishSettingsActionWithResult
       → 302 to app's redirectURI
App → finishLinkOAuth()
```

---

## File-by-file changes

### 1. `pkg/lib/settingsaction/actions.go`

Add a new `SettingsAction` constant alongside the existing ones (`SettingsActionChangePassword`, `SettingsActionDeleteAccount`, etc.):

```go
SettingsActionLinkOAuth SettingsAction = "link_oauth"
```

---

### 2. `pkg/lib/endpoints/endpoints.go`

Add a URL builder for the OAuth identity settings page:

```go
func (e *Endpoints) SettingsIdentityOAuthURL() *url.URL {
    return e.urlOf("settings/identity/oauth")
}
```

---

### 3. `pkg/lib/oauth/oidc/ui.go`

**3a. Extend `UIURLBuilderAuthUIEndpointsProvider` interface:**

```go
SettingsIdentityOAuthURL() *url.URL
```

**3b. Add case in `BuildSettingsActionURL`:**

```go
case settingsaction.SettingsActionLinkOAuth:
    endpoint = b.Endpoints.SettingsIdentityOAuthURL()
    b.addToEndpoint(endpoint, r, e)
    if alias := r.OAuthProviderAlias(); alias != "" {
        q := endpoint.Query()
        q.Set("x_provider_alias", alias)
        endpoint.RawQuery = q.Encode()
    }
    return endpoint, nil
```

`r.OAuthProviderAlias()` reads `x_oauth_provider_alias` from the OIDC authorization request (already implemented in `protocol/authz.go`).

---

### 4. `pkg/auth/handler/webapp/sso_callback.go`

After `FinishAddingIdentityOAuth` succeeds on the account-management path, append `x_oauth_linked=1` to the redirect URL:

```go
redirectURL, _ := url.Parse("/settings/identity/oauth")
q := redirectURL.Query()
q.Set("x_oauth_linked", "1")
redirectURL.RawQuery = q.Encode()
```

---

### 5. `pkg/auth/handler/webapp/authflowv2/settings_identity_list_oauth.go`

Switch GET from `ctrl.Get` to `ctrl.GetWithSettingsActionWebSession`. Add two branches before the normal render:

#### Branch A — auto-trigger link (first GET)

**Condition:** `IsInSettingsAction && x_provider_alias present && x_oauth_linked absent`

1. Look up the OAuth candidate matching `x_provider_alias`.
2. Determine the redirect URI (shared or alias-specific).
3. Call `StartAddIdentityOAuth` → `AccountManagementToken` + provider auth URL.
4. Build `WebappOAuthState{AccountManagementToken: token}`.
5. Call `GenerateState` → state token.
6. Call `generateAuthorizationURLWithState` → full provider URL.
7. `http.Redirect` to the provider. No template rendered.

#### Branch B — finish action (post-link GET)

**Condition:** `IsInSettingsAction && x_oauth_linked == "1"`

1. Call `ctrl.FinishSettingsActionWithResult`.
2. Write the response (302 to app's `redirect_uri`).

#### Branch C — normal GET (unchanged)

Fall through to the existing render path.

---

### After all code changes

```
make generate
```

---

## Checklist

- [x] `SettingsActionLinkOAuth` constant added
- [x] `SettingsIdentityOAuthURL` endpoint method added and exposed on interface
- [x] `BuildSettingsActionURL` case added
- [x] `sso_callback.go` appends `x_oauth_linked=1` on the account-management path
- [x] GET handler switched to `GetWithSettingsActionWebSession` with Branch A / B / C
- [x] `make generate` clean
- [ ] `make test` passes; relevant package tests added
