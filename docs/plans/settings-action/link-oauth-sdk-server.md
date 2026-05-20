# Settings Action: Link OAuth — Server Implementation Plan

> Source repo: `authgear-server`. Companion plans: `connect-oauth-sdk-js-web.md`, `connect-oauth-sdk-js-rn.md`.

## Goal

Introduce a new settings action `link_oauth` that lets a client app delegate "link an OAuth provider"
to the Authgear settings page. The action is **pure link-only**: it shows only the specified provider
and has no disconnect capability. A separate `unlinkOAuth` action will be designed later.

`oauthProviderAlias` is **required**. Calling `link_oauth` without specifying an alias is an error.

---

## End-to-end call flow

```
App → startLinkOAuth({ redirectURI, oauthProviderAlias: "google" })
  → OIDC /authorize (response_type=settings-action, x_settings_action=link_oauth, x_oauth_provider_alias=google)
  → BuildSettingsActionURL → validates alias is present, then:
      /settings/identity/oauth?x_provider_alias=google&x_settings_action_id=...
  → Session middleware: creates webapp session, strips x_ref, 302 to clean URL
  → GET /settings/identity/oauth?x_provider_alias=google&x_settings_action_id=...     [Branch A or C]

    Branch A (not yet linked):
      → StartAddIdentityOAuth → AccountManagementToken
      → GenerateState (stores SettingsActionID + ProviderAlias in WebappOAuthState)
      → 302 to Google OAuth

    Branch C (already linked):
      → Render filtered list (only "google"), show "already linked" error, hide buttons

  → Google callback → /sso/oauth2/callback
      → FinishAddingIdentityOAuth
      → Success: 302 to /settings/identity/oauth?x_provider_alias=google&x_settings_action_id=...&x_oauth_linked=1
      → Error:   302 to /settings/identity/oauth?x_provider_alias=google&x_settings_action_id=... (+ error cookie)

  → GET /settings/identity/oauth?...&x_oauth_linked=1                                 [Branch B]
      → FinishSettingsActionWithResult → 302 to app's redirectURI
App → finishLinkOAuth()

  → GET /settings/identity/oauth?x_provider_alias=google&x_settings_action_id=...     [Branch C, SSO error]
      → Render filtered list (only "google"), show SSO error from cookie, show connect button
```

---

## File-by-file changes

### 1. `pkg/lib/settingsaction/actions.go` ✅ done

`SettingsActionLinkOAuth SettingsAction = "link_oauth"` already added.

---

### 2. `pkg/lib/endpoints/endpoints.go` ✅ done

`SettingsIdentityOAuthURL()` already added.

---

### 3. `pkg/lib/oauth/oidc/ui.go` — validate alias is required

Update the `link_oauth` case in `BuildSettingsActionURL` to return an error when `oauthProviderAlias` is
absent:

```go
case settingsaction.SettingsActionLinkOAuth:
    alias := r.OAuthProviderAlias()
    if alias == "" {
        return nil, errors.New("oauthProviderAlias is required for link_oauth settings action")
    }
    endpoint = b.Endpoints.SettingsIdentityOAuthURL()
    b.addToEndpoint(endpoint, r, e)
    q := endpoint.Query()
    q.Set("x_provider_alias", alias)
    endpoint.RawQuery = q.Encode()
    return endpoint, nil
```

---

### 4. `pkg/lib/webappoauth/state.go` ✅ done

`SettingsActionID string` already added to `WebappOAuthState`.

---

### 5. `pkg/auth/handler/webapp/sso_callback.go` — preserve x_provider_alias in all redirects

`x_settings_action_id` is already preserved. Additionally preserve `x_provider_alias` (from
`state.ProviderAlias`) in both the success and error redirect URLs:

```go
redirectURL, _ := url.Parse("/settings/identity/oauth")

// Preserve state params in both error and success paths.
{
    q := redirectURL.Query()
    if state.SettingsActionID != "" {
        q.Set(settingsaction.QUERY_SETTINGS_ACTION_ID, state.SettingsActionID)
    }
    if state.ProviderAlias != "" {
        q.Set("x_provider_alias", state.ProviderAlias)
    }
    redirectURL.RawQuery = q.Encode()
}

_, err = h.AccountManagement.FinishAddingIdentityOAuth(...)
if err != nil {
    h.ErrorRenderer.MakeAuthflowErrorResult(r.Context(), w, r, *redirectURL, err).WriteResponse(w, r)
    return
}

q := redirectURL.Query()
q.Set("x_oauth_linked", "1")
redirectURL.RawQuery = q.Encode()
http.Redirect(w, r, redirectURL.String(), http.StatusFound)
```

---

### 6. `pkg/auth/handler/webapp/authflowv2/settings_identity_list_oauth.go` — filtered view + pure link

#### 6a. View model — add IsAlreadyLinked and IsInSettingsAction flags

Add to `AuthflowV2SettingsIdentityListOAuthViewModel`:

```go
type AuthflowV2SettingsIdentityListOAuthViewModel struct {
    OAuthCandidates     []identity.Candidate
    OAuthIdentities     []*identity.OAuth
    Verifications       map[string][]verification.ClaimStatus
    IdentityCount       int
    CreateDisabled      bool
    // New fields for settings-action filtered view:
    IsInSettingsAction  bool   // true when rendered inside a link_oauth settings action
    IsAlreadyLinked     bool   // true when the specified provider is already linked
}
```

#### 6b. GET handler — Branch A / B / C

**Branch A** (auto-trigger link, not yet linked):
- Condition: `IsInSettingsAction && x_provider_alias present && x_oauth_linked absent`
- Look up candidate by alias. If `identity_id` is non-empty → provider already linked → fall through to Branch C.
- Otherwise start OAuth flow as before (store `SettingsActionID` + `ProviderAlias` in `WebappOAuthState`).

**Branch B** (finish after successful link):
- Condition: `IsInSettingsAction && x_oauth_linked == "1"`
- Call `ctrl.FinishSettingsActionWithResult` → write response.

**Branch C** (filtered list render):
- Reached when: already linked (from Branch A), or after SSO error redirect, or any direct GET
  in settings-action mode without `x_oauth_linked`.
- When `IsInSettingsAction`:
  - Filter `OAuthCandidates` to only the candidate whose alias matches `x_provider_alias`.
  - Set `vm.IsInSettingsAction = true`.
  - Set `vm.IsAlreadyLinked = true` if the candidate's `identity_id` is non-empty.
- When not in a settings action: render the full list unchanged (existing behaviour).

#### 6c. POST "add" handler

No change to the existing logic (already stores `SettingsActionID` in `WebappOAuthState`).
This path is only reachable when Branch C renders with a connect button (i.e. not already linked).

#### 6d. POST "remove" handler — revert to plain PostAction

The "remove" handler must be a plain `ctrl.PostAction` (not `PostActionWithSettingsActionWebSession`).
`linkOAuth` is pure link-only; the disconnect button is hidden by the template in settings-action mode,
so this code path is unreachable in a settings action. Keeping it plain avoids accidental
`FinishSettingsAction` from a disconnect.

---

### 7. `resources/authgear/templates/en/web/authflowv2/settings_identity_list_oauth.html` — filtered template

When `$.IsInSettingsAction` is true:

1. **Show only the single provider** (the `range .OAuthCandidates` list will already be filtered to one
   entry by the handler — no template change needed for filtering).
2. **Already-linked state** — when `$.IsAlreadyLinked`:
   - Show an error banner: *"This account is already linked."*
   - Hide the connect button **and** the disconnect button (render neither form).
3. **SSO error state** — when not `$.IsAlreadyLinked`:
   - Show connect button as normal (the recoverable error cookie is already rendered by the base
     template's `RawError` mechanism).

---

## Checklist

- [x] `SettingsActionLinkOAuth` constant added
- [x] `SettingsIdentityOAuthURL` endpoint method added and exposed on interface
- [x] `BuildSettingsActionURL` case added
- [ ] `BuildSettingsActionURL` — return error when alias is absent
- [x] `WebappOAuthState.SettingsActionID` added
- [x] `sso_callback.go` preserves `x_settings_action_id` in all redirects
- [ ] `sso_callback.go` also preserves `x_provider_alias` in all redirects
- [x] GET handler — Branch A / B / C wired up
- [ ] Branch A — remove the "already linked → render list" inline code; instead fall through to Branch C
- [ ] Branch C — filter candidates to `x_provider_alias` when in settings action
- [ ] Branch C — set `IsInSettingsAction` / `IsAlreadyLinked` flags on view model
- [ ] POST "remove" — revert to plain `ctrl.PostAction` (no settings-action finish)
- [ ] Template — already-linked error banner + hide buttons when `IsAlreadyLinked`
- [ ] Template — hide disconnect button when `IsInSettingsAction` (redundant with above, but defensive)
- [ ] `make generate` clean
- [ ] `make test` passes
