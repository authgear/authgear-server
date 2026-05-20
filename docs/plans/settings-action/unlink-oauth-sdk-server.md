# Settings Action: Unlink OAuth — Server Implementation Plan

> Source repo: `authgear-server`. Companion plan: `unlink-oauth-sdk-js-web.md`.

## Goal

Introduce a new settings action `unlink_oauth` that lets a client app delegate "unlink an OAuth
provider" to the Authgear settings page. The action is **pure unlink-only**: it shows only the
specified provider and has no connect capability.

`oauthProviderAlias` is **required**. Calling `unlink_oauth` without specifying an alias is an error.

Unlike `link_oauth`, there is no OAuth provider round-trip. The user sees the filtered list
immediately and clicks disconnect to finish.

---

## End-to-end call flow

```
App → startUnlinkOAuth({ redirectURI, oauthProviderAlias: "google" })
  → OIDC /authorize (response_type=settings-action, x_settings_action=unlink_oauth,
                      x_oauth_provider_alias=google)
  → BuildSettingsActionURL → validates alias is present, then:
      /settings/identity/oauth?x_provider_alias=google
                               &x_settings_action=unlink_oauth
                               &x_settings_action_id=...
  → Session middleware: creates webapp session, 302 to clean URL
  → GET /settings/identity/oauth?x_provider_alias=google
                                  &x_settings_action=unlink_oauth
                                  &x_settings_action_id=...

    Branch D — provider IS linked:
      → Render filtered list (only "google"), show disconnect button only

    Branch D — provider NOT linked:
      → Render filtered list (only "google"), show "provider not linked" error, no buttons

    Branch D — provider alias not configured:
      → Render empty list, show "provider not configured" error banner (IsUnknownProvider)

  → User clicks Disconnect
  → POST /settings/identity/oauth (action=remove, q_identity_id=<id>, x_settings_action_id=...)
      → DeleteIdentityOAuth
      → FinishSettingsActionWithResult → 302 to app's redirectURI

App → finishUnlinkOAuth()
```

---

## File-by-file changes

### 1. `pkg/lib/settingsaction/actions.go`

Add after `SettingsActionLinkOAuth`:

```go
SettingsActionUnlinkOAuth SettingsAction = "unlink_oauth"
```

---

### 2. `pkg/lib/oauth/oidc/ui.go` — add `unlink_oauth` case

Add after the `SettingsActionLinkOAuth` case in `BuildSettingsActionURL`:

```go
case settingsaction.SettingsActionUnlinkOAuth:
    alias := r.OAuthProviderAlias()
    if alias == "" {
        return nil, NewErrInvalidSettingsAction("oauthProviderAlias is required for unlink_oauth")
    }
    endpoint = b.Endpoints.SettingsIdentityOAuthURL()
    b.addToEndpoint(endpoint, r, e)
    q := endpoint.Query()
    q.Set("x_provider_alias", alias)
    q.Set("x_settings_action", string(settingsaction.SettingsActionUnlinkOAuth))
    endpoint.RawQuery = q.Encode()
    return endpoint, nil
```

`x_settings_action` uses the `x_` prefix so `PreserveQuery` retains it across page interactions.
The handler uses this value to distinguish unlink mode from link mode.

---

### 3. `pkg/auth/handler/webapp/authflowv2/settings_identity_list_oauth.go`

#### 3a. View model — add unlink-specific flags

`IsUnknownProvider` is already present on the view model (added for `link_oauth`). Add the two
new unlink-specific fields:

```go
type AuthflowV2SettingsIdentityListOAuthViewModel struct {
    OAuthCandidates []identity.Candidate
    OAuthIdentities []*identity.OAuth
    Verifications   map[string][]verification.ClaimStatus
    IdentityCount   int
    CreateDisabled  bool

    // Settings action state
    IsInSettingsAction     bool  // true when inside any settings action
    IsAlreadyLinked        bool  // link_oauth: provider is already linked (error)
    IsUnknownProvider      bool  // any mode: alias not in app config (error) — already added
    IsUnlinkSettingsAction bool  // true when inside unlink_oauth settings action
    IsNotLinked            bool  // unlink_oauth: provider is not linked (error)
}
```

#### 3b. `handleGet` — unlink mode detection and Branch D

At the top of `handleGet`, read the action type from the URL:

```go
settingsActionParam := r.URL.Query().Get("x_settings_action")
isUnlinkMode := ctrl.IsInSettingsAction(s, webappSession) &&
    settingsActionParam == string(settingsaction.SettingsActionUnlinkOAuth)
```

Guard Branch A (auto-trigger link) and Branch B (finish after link) to skip when in unlink mode:

```go
// Branch A: link_oauth auto-trigger (skipped in unlink_oauth mode)
ssoError := r.URL.Query().Get("q_sso_error")
if ctrl.IsInSettingsAction(s, webappSession) && !isUnlinkMode &&
    providerAlias != "" && oauthConnected == "" && ssoError == "" {
    if handled, err := h.autoTriggerOAuth(ctx, w, r, s, providerAlias); err != nil || handled {
        return err
    }
}

// Branch B: finish after link (skipped in unlink_oauth mode)
if ctrl.IsInSettingsAction(s, webappSession) && !isUnlinkMode && oauthConnected == "1" {
    settingsActionResult, err := ctrl.FinishSettingsActionWithResult(ctx, s, webappSession)
    if err != nil {
        return err
    }
    settingsActionResult.WriteResponse(w, r)
    return nil
}
```

In Branch C (filtered list rendering), extend the settings-action block to handle unlink mode:

```go
if ctrl.IsInSettingsAction(s, webappSession) && providerAlias != "" {
    filtered := []identity.Candidate{}
    if c := findCandidate(vm.OAuthCandidates, providerAlias); c != nil {
        filtered = append(filtered, c)
        identityID, _ := c[identity.CandidateKeyIdentityID].(string)
        if isUnlinkMode {
            vm.IsUnlinkSettingsAction = true
            vm.IsNotLinked = identityID == ""
        } else {
            vm.IsAlreadyLinked = identityID != ""
        }
    } else {
        vm.IsUnknownProvider = true
    }
    vm.OAuthCandidates = filtered
    vm.IsInSettingsAction = true
}
```

When `findCandidate` returns nil (alias not in app config), `IsUnknownProvider` is set for both
link and unlink modes. The existing `IsUnknownProvider` error banner in the template and the
`[To Developer]`-prefixed translation key already cover this case — no additional template or
translation changes are needed.

#### 3c. POST "remove" — finish settings action after delete

Change `ctrl.PostAction("remove", ...)` to `ctrl.PostActionWithSettingsActionWebSession` so the
handler receives `webappSession` and can call `FinishSettingsActionWithResult`:

```go
ctrl.PostActionWithSettingsActionWebSession("remove", r, func(ctx context.Context, webappSession *webapp.Session) error {
    s := session.GetSession(ctx)

    identityID := r.Form.Get("q_identity_id")

    _, err := h.AccountManagement.DeleteIdentityOAuth(ctx, s, &accountmanagement.DeleteIdentityOAuthInput{
        IdentityID: identityID,
    })
    if err != nil {
        return err
    }

    if ctrl.IsInSettingsAction(s, webappSession) {
        settingsActionResult, err := ctrl.FinishSettingsActionWithResult(ctx, s, webappSession)
        if err != nil {
            return err
        }
        settingsActionResult.WriteResponse(w, r)
        return nil
    }

    redirectURI := httputil.HostRelative(r.URL).String()
    result := webapp.Result{RedirectURI: redirectURI}
    result.WriteResponse(w, r)
    return nil
})
```

When not in a settings action, `webappSession` is nil, `IsInSettingsAction` returns false, and
the existing redirect-back-to-page behaviour is preserved.

---

### 4. `resources/authgear/templates/en/web/authflowv2/settings_identity_list_oauth.html`

#### 4a. Not-linked error banner

Add after the `IsAlreadyLinked` banner block:

```html
{{ if $.IsNotLinked }}
  {{ template "authflowv2/__alert_message.html"
    (dict
      "Type" "error"
      "Message" (translate "v2.page.settings-identity-oauth.unlink-oauth.not-linked-error" nil)
      "Classname" "widget-content__alert--settings"
    )
  }}
{{ end }}
```

#### 4b. Action button logic in `__settings_action_item_connect.html`

Replace the existing conditional chain with four clearly separated cases:

```html
{{ define "__settings_action_item_connect.html" }}

{{ if .Ctx.IsAlreadyLinked }}
  {{/* Case 1: link_oauth error — provider already linked. No action. */}}

{{ else if .Ctx.IsNotLinked }}
  {{/* Case 2: unlink_oauth error — provider not linked. No action. */}}

{{ else if not .Verified }}
  {{/* Case 3: provider not yet linked (or claims unverified).
       Show connect button, except in unlink_oauth mode where connecting is not the intent. */}}
  {{ if not .Ctx.IsUnlinkSettingsAction }}
  <form method="post" novalidate>
    <input type="hidden" name="x_provider_alias" value="{{ .ProviderAlias }}">
    <button
      class="settings-link-btn"
      type="submit"
      name="x_action"
      value="add"
    >
      {{ translate "v2.page.settings-identity-oauth.default.create-oauth-button-label" nil }}
    </button>
  </form>
  {{ end }}

{{ else }}
  {{/* Case 4: provider is linked and verified.
       Show disconnect button when:
         - normal mode (not inside any settings action), OR
         - unlink_oauth settings action mode.
       Hidden in link_oauth settings action mode (IsInSettingsAction=true, IsUnlinkSettingsAction=false).
       Always guarded by IdentityCount > 1 to prevent removing the last identity. */}}
  {{ if and (or (not .Ctx.IsInSettingsAction) .Ctx.IsUnlinkSettingsAction) (gt .Ctx.IdentityCount 1) }}
  {{ if not .DeleteDisabled }}
  <button
    class="settings-link-btn--destructive"
    data-controller="dialog"
    data-action="click->dialog#open"
    id="{{ .ProviderAlias }}"
  >
    {{ translate "v2.page.settings-identity-oauth.default.remove-oauth-button-label" nil }}
  </button>

  {{ $provider_name := (translate (printf "v2.page.settings-identity-oauth.default.provider.%s" .ProviderType) nil) }}

  {{ template "authflowv2/__settings_dialog.html"
    (dict
      "Ctx" .Ctx
      "DialogID" .ProviderAlias
      "Title" (include "v2.page.settings-oauth.default.remove-oauth-dialog-title" (dict "ProviderName" $provider_name))
      "Description" (include "v2.page.settings-oauth.default.remove-oauth-dialog-description" (dict "ProviderName" $provider_name))
      "FormContent" (include "__settings_oauth_dialog_remove_input.html" (dict "ProviderAlias" .ProviderAlias "IdentityID" .IdentityID))
      "Buttons"
        (list
          (dict
            "Type" "Destructive"
            "Label" (include "v2.component.button.default.label-remove" nil)
            "Value" "remove"
            "Event" "authgear.button.remove_oauth"
          )
          (dict
            "Type" "Cancel"
            "Label" (include "v2.component.button.default.label-cancel" nil)
          )
        )
  )}}
  {{ end }}
  {{ end }}

{{ end }}
{{ end }}
```

---

### 5. Translation keys

Add to `resources/authgear/templates/en/translation.json`:

```json
"v2.page.settings-identity-oauth.unlink-oauth.not-linked-error": "This account is not linked to {{ .ProviderName }}."
```

Regenerate translations with `make generate-translations`.

---

### 6. `.make-lint-translation-keys-expect`

After editing the template, the template linter reports line numbers for known-missing translation
keys (e.g. `v2.page.settings-identity-oauth.default.provider.%s`). Adding lines shifts those
numbers. Regenerate the expect file so `make lint` passes:

```sh
go run ./devtools/gotemplatelinter \
  --ignore-rule indentation \
  --ignore-rule eol-at-eof \
  ./resources/authgear/templates/en/web/authflowv2 \
  > .make-lint-translation-keys-expect 2>&1
```

Commit the updated `.make-lint-translation-keys-expect` alongside the template changes.

---

## Checklist

- [ ] `SettingsActionUnlinkOAuth` constant added
- [ ] `BuildSettingsActionURL` — `unlink_oauth` case with alias validation and `x_settings_action` URL param
- [ ] View model — `IsUnlinkSettingsAction` and `IsNotLinked` flags added (`IsUnknownProvider` already present)
- [ ] `handleGet` — Branch A and B guarded with `!isUnlinkMode`
- [ ] `handleGet` — Branch C sets `IsUnlinkSettingsAction` / `IsNotLinked` in unlink mode; sets `IsUnknownProvider` when alias not found (both modes)
- [ ] POST "remove" — changed to `PostActionWithSettingsActionWebSession`; calls `FinishSettingsActionWithResult` when in settings action
- [ ] Template — not-linked error banner added
- [ ] Template — connect button hidden in unlink mode
- [ ] Template — disconnect button visible in unlink mode (with IdentityCount > 1 guard)
- [ ] Translation key added and translations regenerated
- [ ] `.make-lint-translation-keys-expect` regenerated after template changes
- [ ] `make generate` clean
- [ ] `make check-tidy` passes
- [ ] `make lint` passes
- [ ] `make test` passes

---

## Atomic commit plan

1. **`pkg/lib/settingsaction/actions.go`** — add `SettingsActionUnlinkOAuth` constant.

2. **`pkg/lib/oauth/oidc/ui.go`** — add `unlink_oauth` case in `BuildSettingsActionURL` with alias
   validation and `x_settings_action` URL param.

3. **`pkg/auth/handler/webapp/authflowv2/settings_identity_list_oauth.go`** — add view model
   flags (`IsUnlinkSettingsAction`, `IsNotLinked`), gate Branch A/B on `!isUnlinkMode`, populate
   unlink flags in Branch C (including `IsUnknownProvider` when alias not found), change POST
   "remove" to `PostActionWithSettingsActionWebSession`.

4. **`resources/authgear/templates/en/web/authflowv2/settings_identity_list_oauth.html`**,
   **`resources/authgear/templates/en/translation.json`**, and
   **`.make-lint-translation-keys-expect`** — not-linked error banner, updated button logic, new
   translation key. Run `make generate-translations` then regenerate
   `.make-lint-translation-keys-expect` (see §6) in the same commit.

5. **`.vettedpositions`** — update for line-number drift if any `requestcontext` positions shift.
