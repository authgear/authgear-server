# CIMD Part 6 — Consent Screen Phishing Mitigation & Third-Party Authorization Display

Spec: [docs/specs/cimd.md — Phishing Mitigation](../../specs/cimd.md#phishing-mitigation), [Reading a CIMD Client's Stored Config](../../specs/cimd.md#reading-a-cimd-clients-stored-config).

Depends on [Part 1](2026-08-28-01-config-and-client-id.md) (`config.OAuthClientConfig.DynamicSource` / `IsCIMDClient()`) and [Part 3](2026-08-28-03-authorize-time-resolution.md) (a resolvable CIMD client actually exists).

## 1. Goal / Scope

Two independent pieces of user-facing work:

1. **The consent screen shows the `client_id` URL's hostname**, prominently, alongside `client_name` — required by draft §8.5, the one genuinely new UI requirement CIMD introduces. Plus `logo_uri` rendering, which no client of any source gets today.
2. **Third-party authorizations for dynamically-resolved clients stop being invisible.** `oauth.KeepThirdPartyAuthorizationFilter` is built from `authgear.yaml` only, so a DCR or CIMD third-party client's authorization is filtered out of `user.authorizations` in the Admin API. That is a pre-existing DCR bug; it is fixed here for both sources.

A third collection has the same static-config-only defect — `SessionListingService`'s third-party client-id set — and is **deliberately not** changed. §1.1 explains why touching it would be a regression rather than a fix.

### 1.1 Finding: there is no "Authorized Apps settings page"

Spec § Reading a CIMD Client's Stored Config says "The Authorized Apps settings page reads the same persisted record for display (`client_name`, `logo_uri`), so no separate end-user mechanism is needed either." Spec § Where resolution happens likewise lists "The Authorized Apps settings page" as a reader of the persisted record.

**That page does not exist.** Verified:

- `AuthflowV2SettingsSessionsHandler` (`pkg/auth/handler/webapp/authflowv2/settings_sessions.go:90-116`) does compute `SettingsSessionsViewModel.Authorizations` and embed it into the template data;
- **no template in `resources/authgear/templates/` references `Authorizations`** (`grep -rn "Authorization" resources/` returns only `wechat.html`'s unrelated `AuthorizationURL`);
- there is no route other than `/settings/sessions`, and no revoke-authorization handler in `pkg/auth`.

So `Authorizations` is computed on every `/settings/sessions` render and thrown away. The end-user-facing surface the spec describes is dead code.

**The exclusion it was meant to pair with is live, and it is inconsistent — but do not "fix" it.** `SessionListingService.FilterForDisplay` (`pkg/lib/sessionlisting/listing.go:59-72`) builds `thirdPartyClientIDs` from `s.OAuthConfig.Clients` and then skips any offline grant where `IsOnlyUsedInClientIDs(thirdPartyClientIDs)` is true:

```go
thirdPartyClientIDs := []string{}
for _, c := range s.OAuthConfig.Clients {
	if c.IsThirdParty() {
		thirdPartyClientIDs = append(thirdPartyClientIDs, c.ClientID)
	}
}
...
for _, offlineGrant := range offlineGrants {
	if offlineGrant.IsOnlyUsedInClientIDs(thirdPartyClientIDs) {
		continue
	}
```

That establishes the intended design: **"Signed in Devices" was meant to be first-party sessions only, with third-party apps listed separately** — which is why `Authorizations` is computed alongside it. And because the set is static-config-only, a DCR or CIMD third-party client's offline grant is not recognised as third-party, so it *is* listed as a device today, while a statically-configured third-party client's grant is not.

That inconsistency is real: whether a third-party client's session is visible depends on nothing more principled than where the client was configured. **But an earlier draft of this plan resolved it by excluding dynamic clients too, and that was wrong.** Three facts, all verified, say the current inclusion is the safer behavior and should be left alone:

- **The row is imprecise, not incorrect.** An offline grant for a third-party client is a genuine live session holding a refresh token. `DisplayName` comes from `DeviceInfo`'s device model, falling back to the formatted User-Agent (`pkg/lib/oauth/grant_offline.go:184-190`) — uninformative for an MCP CLI client, but the user really did authorize it from a device.
- **Terminating it works, and is the user's only control.** The page's `revoke` action calls `Sessions.RevokeWithEvent` on whatever is listed (`pkg/auth/handler/webapp/authflowv2/settings_sessions.go:147-159`). There is no revoke-*authorization* path anywhere in the codebase (§1.2), and no Authorized Apps surface. So excluding these grants would remove the only way an end user can cut off a self-asserted third-party app.
- **CIMD makes that worse, not better.** A static `third_party_app` client was deliberately configured by a project admin; a CIMD client is self-asserted by whoever can host a document. Hiding the riskier one while showing the safer one inverts the sensible priority.

So: **leave `FilterForDisplay` unchanged in this plan set.** The inconsistency is a known issue to resolve together with the Authorized Apps surface — excluding third-party grants from the device list is only safe once there is somewhere else that shows and revokes them, and at that point the resolution should probably be to exclude *both* sources, not to exclude neither.

**Scope decision.** Building the end-user Authorized Apps surface — list, per-app detail, revoke — is a feature in its own right, outside anything cimd.md specifies, and it needs a design decision this plan set should not make unilaterally (below). It is **out of scope.** In scope: making the data behind it correct, and fixing §3.5's live bug.

This needs saying in the PR description and correcting in the spec (§6): a reader of cimd.md today would reasonably conclude the end-user surface is already handled.

### 1.2 If it is built: a section on the sessions page, not a device row

Recording the answer here so the decision is cheap when someone takes it.

**How first-party sessions are displayed today.** `/settings/sessions` is titled "Signed in Devices" (`v2.page.settings-sessions.default.title`) and renders one row per session from `SessionListingService.FilterForDisplay`, showing `DisplayName` (device/browser info), last-accessed IP with country, last-accessed time, a "current session" marker, a per-row terminate action and a "Terminate all other sessions" action. Both IDP sessions and offline grants appear, collapsed by SSO group.

**Third-party apps should not be shown the same way**, and the existing code already agrees — §1.1's exclusion exists precisely to keep them out of that list. They are a different kind of object:

| | Session | Authorization |
|---|---|---|
| Means | a device or browser you are signed in on | an app you granted access to your account |
| Identified by | device/user-agent | client name, and for CIMD the `client_id` hostname |
| Useful columns | IP, country, last activity, is-current | scopes granted, when granted, whether it has full user-info access |
| The action | terminate this session | revoke this app's access |

Rendering an app as a device row would put a device name where an app name belongs and offer "terminate" where "revoke access" belongs. Google and GitHub both separate these ("Your devices" vs "Third-party apps with account access"; "Sessions" vs "Authorized OAuth Apps").

**Recommendation: a second section on the same `/settings/sessions` page**, not a new route. It needs no new route, no navigation change, and no new handler — `SettingsSessionsViewModel.Authorizations` is already populated there, and after §3 it is populated *correctly*. A separate page would mean a route, a nav entry, and a settings-page registration for a list most users will find empty.

**The one decision that has to be made first, and why this is not just UI work:** what "revoke" does. `deleteDynamicClient` deletes the client row and does **not** revoke authorizations or offline grants (verified: `oauthclient.Commands.DeleteClient` only deletes and invalidates the cache; DCR Part 2 §6.4 flagged the missing revocation as a known gap). So there is currently *no* code path that revokes a third-party app's access — an end-user revoke button would be the first one, and it has to decide whether it deletes the `_auth_oauth_authorization` row, revokes the offline grants for that client, or both. That is a real behavioral design question, and it is the reason this is out of scope rather than a template change.

## 2. Consent screen — the hostname

### 2.1 Why the hostname, and why it must not be conflated with `client_name`

This is a spec requirement, not an Authgear invention. draft-ietf-oauth-client-id-metadata-document-02 §8.5 (titled **"OAuth Phishing Attacks"** in the draft; cimd.md calls it "Phishing Mitigation") says, verbatim:

> The authorization server SHOULD display the hostname of the `client_id` on the authorization interface, in addition to displaying the fetched client information if any.

and, in the same section:

> If fetching the Client ID Metadata Document fails for any reason, the `client_id` URL is the only piece of information the user has as an indication of which application they are authorizing.

Core OAuth says nothing about consent-screen content — RFC 6749 has no such requirement — so §8.5 is the whole normative basis, at SHOULD strength. It is also established practice in the deployed systems that use URL-based client IDs: AT Protocol/Bluesky OAuth and Solid-OIDC both use a client metadata document at a URL as the `client_id` and both display its origin. The underlying principle is the same one browsers apply by showing the origin in the address bar rather than the page title: display the identifier the other party cannot forge.

The second quote does not create work here. Authgear never renders a consent screen with no fetched information: a first-time fetch failure makes the client unresolvable (Part 3 §6.3), and a *re*fetch failure serves the last known-good record (Part 3 §5). So `client_name` is always present by the time this screen renders — the hostname is additional, never a fallback.

That parenthetical is the entire design constraint. `client_name`, `logo_uri`, `client_uri`, `tos_uri` and `policy_uri` are all strings the client wrote about itself; a malicious client can set `client_name` to "Google" and `logo_uri` to Google's logo. The hostname is the only field an attacker cannot forge without controlling that domain, because Authgear fetched the document from it over verified TLS. So it must be rendered as a distinct, visibly-different piece of information — not merged into the title, not used as a fallback for a missing `client_name`.

### 2.2 `ConsentViewModel` — three additions and one bug fix

`pkg/auth/handler/oauth/consent.go:61-69`:

```go
type ConsentViewModel struct {
	ClientName      string
	ClientPolicyURI string
	ClientTOSURI    string
	// ClientLogoURI is the client's self-asserted logo. Empty unless the
	// client declared one. Rendered directly in Part 6; replaced by a
	// server-side proxy URL in Part 7 (see docs/plans/cimd/2026-08-28-07-logo-proxy.md).
	ClientLogoURI string
	// ClientIDHostname is the hostname of a CIMD client's client_id URL, and
	// EMPTY for every other client source. It is the only client-identifying
	// value on this screen that the client did not assert about itself: it is
	// the host Authgear actually fetched the metadata document from, over
	// verified TLS. See docs/specs/cimd.md § Phishing Mitigation.
	ClientIDHostname string
	Scopes              []string
	CustomScopes        []ConsentScope
	IdentityDisplayName string
	UserProfile         webapp.UserProfile
}
```

`renderConsentPage` (`consent.go:161-171`):

```go
	// BUG FIX: this was consentRequired.Client.ClientName, the RAW
	// client_name, which is "" whenever a client omits it -- and the template
	// then renders the literal string "null" via `or $.ClientName "null"`.
	// Client.Name is the resolved display name: client_name for a static
	// client, and DisplayName() ("client_name, or 'Client <clientID>'") for a
	// dynamic one (pkg/lib/oauthclient/client.go:66-71). Every DCR client
	// registered without a client_name hits this today; a CIMD client will
	// hit it far more often, since client_name is optional in the document.
	viewModel.ClientName = consentRequired.Client.Name
	viewModel.ClientPolicyURI = consentRequired.Client.PolicyURI
	viewModel.ClientTOSURI = consentRequired.Client.TOSURI
	viewModel.ClientLogoURI = consentRequired.Client.LogoURI
	if consentRequired.Client.IsCIMDClient() {
		if u, err := url.Parse(consentRequired.Client.ClientID); err == nil {
			viewModel.ClientIDHostname = u.Hostname()
		}
	}
```

`Client.Name` is populated for a dynamic client by `ToClientConfig` (`client_config.go:133`: `Name: c.DisplayName()`), and for a static client by config. So this fix is source-agnostic and improves DCR too.

The `url.Parse` cannot realistically fail — `ParseCIMDClientID` already parsed the same string — but it is a `*config.OAuthClientConfig` field by the time it gets here, with no parse result carried along, so re-parsing is simpler than threading one through. On the impossible error path the hostname is simply omitted, which is fail-safe (the screen shows less, never something wrong).

`IsCIMDClient()` rather than a string prefix check is what keeps a *static* client whose `client_id` happens to be an `https://` URL — spec § Client ID Format's supported pre-registration pattern — from picking up a CIMD-specific hostname banner it has not earned. That client's `DynamicSource` is `""` (Part 1 §5.1).

### 2.3 Template — `resources/authgear/templates/en/web/authflowv2/consent.html`

Two changes. First, the existing `$clientName` line:

```gotemplate
{{ $appName := (translate "app.name" nil) }}
{{ $clientName := or $.ClientName "null" }}
```

The `"null"` fallback becomes unreachable once §2.2's fix lands (`Client.Name` is never empty for any resolvable client), but leave the `or` in place as a belt — a literal `"null"` on screen is ugly but not dangerous, and removing the guard makes an empty title possible.

Second, insert the hostname and logo. Place the hostname **directly under the title**, before the scope list, so it is read as part of "who is asking":

```gotemplate
  <div class="screen-title-description">
    {{ if $.ClientLogoURI }}
      <img
        class="mx-auto h-16 w-16 rounded-lg object-contain"
        src="{{ $.ClientLogoURI }}"
        alt=""
        referrerpolicy="no-referrer"
        loading="lazy"
      >
    {{ end }}
    <h1 class="screen-title">
      {{ include "v2.page.consent.default.title" (dict "ClientName" $clientName) }}
    </h1>
    {{ if $.ClientIDHostname }}
      <p class="text-center text-sm text-[var(--color-neutral-600)]">
        {{ include "v2.page.consent.default.client-hostname-desc" (dict "hostname" $.ClientIDHostname) }}
      </p>
    {{ end }}
```

New translation key in `resources/authgear/templates/en/translation.json` (alphabetically ordered — the file is sorted, keep it that way):

```json
  "v2.page.consent.default.client-hostname-desc": "This app is published at <b>{hostname}</b>. Authgear has not verified its identity beyond this domain.",
```

Wording notes for review: it names the domain as the verified fact and explicitly disclaims verification of anything else, which is the accurate description of what CIMD establishes. Keep `{hostname}` inside `<b>` so it is the visually dominant part of the sentence — the whole point is that a user comparing "Continue to Google" against "published at **mcp-client.evil.example**" notices the mismatch. Do not translate the hostname or truncate it.

`translation.json` is hand-edited, not generated: `scripts/npm/export-v2-translations.mjs` **reads** `resources/authgear/templates/<locale>/translation.json` and writes a CSV for translators, so the English file is the source of truth. Add the key to `en/` only; `zh-HK`/`zh-TW` come back through the translation pipeline. (Confirm against the `update-email-templates` skill's conventions for translation-file edits — that skill covers email templates specifically, but its rule about never hand-editing non-`en` locales applies here too.)

### 2.4 `logo_uri` — rendered directly, with the privacy consequence stated

Rendering `src="{{ $.ClientLogoURI }}"` means **the end user's browser fetches the image directly from the client's own server**, which leaks that user's IP address, `User-Agent` and TLS fingerprint to a third party as a side effect of merely viewing a consent screen. That is exactly the tracking-pixel risk spec § Privacy Considerations §9.2 describes — and the spec claims the design already avoids it, which is not true of this part.

This is a deliberate, sequenced choice: ship the visible feature first, then remove the leak in [Part 7](2026-08-28-07-logo-proxy.md), which replaces `ClientLogoURI` with an Authgear-hosted proxy URL and changes nothing else about the template. Mitigations applied here in the meantime:

- `referrerpolicy="no-referrer"` so the client's server does not learn *which* authorization request or project the user is in the middle of. This does not stop the IP leak; nothing in a direct `<img>` can.
- `loading="lazy"` — marginal, but the image is above the fold so this mostly does not help. Include it anyway; it costs nothing.
- `alt=""` so a broken or blocked image renders as nothing rather than as attacker-supplied alt text next to the app name.

**The CSP does not block this, and must not be changed to.** `pkg/util/httputil/csp.go:339-348` shows the policy sets no `default-src` and deliberately omits `img-src` ("img-src is not needed when there is no default-src", with the directive constant commented out). So a cross-origin `<img>` already renders, and the leak in Part 6 is real rather than theoretical. Part 7 makes the image same-origin, which needs no CSP change either. **Do not add an `img-src` directive** in either part: adding one to constrain the logo would have to enumerate every other image source the AuthUI uses, and getting that list wrong breaks the login page.

**Do not fetch `logo_uri` server-side here.** That is Part 7's whole content.

### 2.5 Where the consent screen is and is not shown

Confirm during implementation, and state in the PR: the consent screen is only reached for a client that requires consent — third-party clients. Every CIMD client is third-party (spec § Mapping), so every CIMD authorization goes through it. A first-party static client skips it entirely, so `ClientIDHostname` being empty for those clients is not merely correct but unreachable.

## 3. Third-party authorization filtering — fixed for DCR and CIMD

### 3.1 The bug

```go
// pkg/lib/oauth/authz_filters.go:37-47 (today)
func NewKeepThirdPartyAuthorizationFilter(oauthConfig *config.OAuthConfig) *KeepThirdPartyAuthorizationFilter {
	s := make(setutil.Set[string])
	for _, c := range oauthConfig.Clients {   // <-- authgear.yaml only
		if c.IsThirdParty() {
			s[c.ClientID] = struct{}{}
		}
	}
	return &KeepThirdPartyAuthorizationFilter{ThirdPartyClientIDSet: s}
}
```

A DCR or CIMD client is not in `oauthConfig.Clients`, so its `client_id` is not in the set, so `Keep` returns false and the authorization is dropped. Consequences today:

- **Admin API `User.authorizations`** (`pkg/admin/graphql/user.go:365`) omits every dynamic third-party client's authorization. Live and portal-visible: an operator looking at a user who has authorized an MCP client sees nothing.
- **`settings_sessions.go`'s view model** omits them too — currently invisible (§1.1) but wrong.

### 3.2 `AuthorizationFilter.Keep` gains a `context.Context`

Resolving a dynamic client requires a context (Redis/Postgres). `Keep` has none. Change the interface:

```go
// pkg/lib/oauth/authz_filters.go

type AuthorizationFilter interface {
	Keep(ctx context.Context, authz *Authorization) bool
}

type AuthorizationFilterFunc func(ctx context.Context, a *Authorization) bool

func (f AuthorizationFilterFunc) Keep(ctx context.Context, a *Authorization) bool {
	return f(ctx, a)
}
```

Complete list of things this touches (`grep -rn "AuthorizationFilter\|\.Keep(" --include=*.go pkg/`):

| Site | Change |
|---|---|
| `pkg/lib/oauth/authz_filters.go` — `AuthorizationFilter`, `AuthorizationFilterFunc`, `ApplyAuthorizationFilters`, `KeepThirdPartyAuthorizationFilter.Keep` | signature |
| `pkg/lib/oauth/authz_service.go:31-45` — `AuthorizationService.ListByUser`'s filter loop | pass `ctx` |
| `pkg/admin/graphql/user.go:365` — construction | new constructor args |
| `pkg/auth/handler/webapp/authflowv2/settings_sessions.go:95` — construction | new constructor args |

`ApplyAuthorizationFilters` (`authz_filters.go:18-31`) has **no callers** — `ListByUser` inlines the same loop instead. Update its signature for consistency rather than leaving a second, now-divergent filter runner; mention in the PR that it is dead and could be deleted, but do not delete it as a drive-by (it is a public API of the package).

Passing `ctx` as the first parameter, not storing it on the filter struct, is the point: a captured `context.Context` on a struct field is a cancellation and tracing bug waiting to happen, and this repo consistently threads `ctx` explicitly (`ResolveClient(ctx, clientID)` throughout).

### 3.3 The resolver-backed filter

```go
type KeepThirdPartyAuthorizationFilterClientResolver interface {
	ResolveClient(ctx context.Context, clientID string) *config.OAuthClientConfig
}

// KeepThirdPartyAuthorizationFilter keeps only authorizations belonging to a
// third-party client, whatever its source: a static third_party_app client,
// or a DCR/CIMD-resolved client (which is always
// OAuthClientApplicationTypeDynamicThirdParty, i.e. IsThirdParty()).
//
// The static-client-id set is GONE. It was not a cache -- it was the only
// lookup, which is why dynamic clients were dropped. Resolver.ResolveClient
// already checks authgear.yaml first and returns the static config without
// touching Redis or Postgres (pkg/lib/oauthclient/resolver.go:249-259), so a
// static client costs exactly what it cost before: one linear scan of
// oauth.clients. A dynamic client costs one Redis GET, cached for 5 minutes.
type KeepThirdPartyAuthorizationFilter struct {
	Resolver KeepThirdPartyAuthorizationFilterClientResolver
}

func NewKeepThirdPartyAuthorizationFilter(resolver KeepThirdPartyAuthorizationFilterClientResolver) *KeepThirdPartyAuthorizationFilter {
	return &KeepThirdPartyAuthorizationFilter{Resolver: resolver}
}

func (f *KeepThirdPartyAuthorizationFilter) Keep(ctx context.Context, authz *Authorization) bool {
	client := f.Resolver.ResolveClient(ctx, authz.ClientID)
	if client == nil {
		// An unresolvable client_id: a static client removed from
		// authgear.yaml, a deleted dynamic client, or a CIMD client_id whose
		// domain is no longer allowlisted. Dropped, which preserves today's
		// exact behavior for the removed-static-client case.
		//
		// This does mean a user's grant to a since-deleted third-party client
		// is not listed anywhere -- and therefore cannot be found and
		// revoked through this surface. That gap exists today for removed
		// static clients and is not made worse here; revocation is still
		// available through session revocation and the Admin API by
		// authorization id. Keeping such an authorization and rendering the
		// bare client_id was considered and rejected: it would newly surface
		// removed FIRST-party clients (whose grants are deliberately not
		// shown) as unlabelled entries, a bigger behavior change than the
		// gap it closes. Revisit if and when the Authorized Apps page in
		// §1.1 is built, where "revoke a grant to a client that no longer
		// exists" is a real user need.
		return false
	}
	return client.IsThirdParty()
}
```

Performance note for the PR: `ListByUser` now performs one `ResolveClient` per authorization instead of one map lookup. A user has a handful of authorizations, and the per-call cost is either an in-memory scan (static) or a cached Redis GET (dynamic). No batching is warranted. If a future caller lists authorizations for many users at once, revisit — that caller does not exist.

### 3.4 Call site updates

**`pkg/admin/graphql/user.go:365`** needs a resolver on the GraphQL context. `Context` (`pkg/admin/graphql/context.go:268-284`) has `OAuthConfig` but no resolver; `facade.OAuthFacade` has one (`pkg/admin/facade/oauth.go:45-57`) but reaching through the facade for an unrelated concern is worse than adding the field. Add:

```go
// pkg/admin/graphql/context.go
type Context struct {
	// ... existing ...
	OAuthClientResolver OAuthClientResolver
}

type OAuthClientResolver interface {
	ResolveClient(ctx context.Context, clientID string) *config.OAuthClientConfig
}
```

wired from the same provider `facade.OAuthFacade.OAuthClientResolver` binds to (`pkg/lib/deps/deps_common.go:689-695`). Then:

```go
	filter := oauth.NewKeepThirdPartyAuthorizationFilter(gqlCtx.OAuthClientResolver)
```

`gqlCtx.OAuthConfig` stays — it is used elsewhere in that file.

**`pkg/auth/handler/webapp/authflowv2/settings_sessions.go`** — replace the static name map and the filter construction:

```go
type AuthflowV2SettingsSessionsHandler struct {
	// ... existing ...
	OAuthConfig         *config.OAuthConfig
	OAuthClientResolver SettingsSessionsClientResolver // new
}

	// Get third party app authorization
	filter := oauth.NewKeepThirdPartyAuthorizationFilter(h.OAuthClientResolver)
	authorizations, err := h.Authorizations.ListByUser(ctx, *userID, filter)
	if err != nil {
		return nil, err
	}
	authzs := []Authorization{}
	for _, authz := range authorizations {
		// One resolve per authorization, same as the filter just did --
		// ResolveClient is cached, and the alternative (threading the
		// resolved config out of the filter) would couple the filter to this
		// caller's rendering needs.
		//
		// Client.Name, not Client.ClientName: the display-name fallback, so a
		// dynamic client with no client_name shows "Client <clientID>" rather
		// than blank. Same fix as §2.2.
		clientName := authz.ClientID
		logoURI := ""
		if c := h.OAuthClientResolver.ResolveClient(ctx, authz.ClientID); c != nil {
			clientName = c.Name
			logoURI = c.LogoURI
		}
		authzs = append(authzs, Authorization{
			ID:                    authz.ID,
			ClientID:              authz.ClientID,
			ClientName:            clientName,
			ClientLogoURI:         logoURI, // new field; see §1.1 -- nothing renders it yet
			Scope:                 authz.Scopes,
			CreatedAt:             authz.CreatedAt,
			HasFullUserInfoAccess: authz.IsAuthorized([]string{oauth.FullUserInfoScope}),
		})
	}
```

`h.OAuthConfig` may become unused in this handler once the name map is gone — check and remove the field if so, rather than leaving a dead dependency. Verify with `make lint`.

`ClientLogoURI` is added to the `Authorization` struct because spec § Reading a CIMD Client's Stored Config names `logo_uri` as something that surface displays. It is populated and unrendered, exactly like the rest of `Authorizations` (§1.1). Populating it now means the field is correct whenever the page is built; if the reviewer prefers not to add unrendered fields, dropping it is fine — say which was chosen.

### 3.5 Wiring

`make generate` for the admin GraphQL context provider and `settings_sessions`'s wire struct. No GraphQL **schema** change: `Authorization` in the Admin API exposes only `clientID` and `scopes` (`pkg/api/model/authorization.go`, `pkg/admin/graphql/authorization.go:25`), so the fix changes *which* authorizations are returned, not their shape. `make export-schemas` should produce no diff — verify.

## 4. Test Plan

**Unit — `pkg/lib/oauth/authz_filters_test.go` (new — there is no test for this file today)**

With a stub resolver:

| Authorization's client | Expect |
|---|---|
| static `third_party_app` | kept |
| static `spa` / `native` / `traditional_webapp` / `confidential` | dropped |
| static `m2m` | dropped |
| dynamic `x_dynamic_third_party` (DCR or CIMD) | **kept** — the regression this part fixes |
| dynamic first-party (`spa` via `Kind: FIRST_PARTY`) | dropped |
| unresolvable `client_id` | dropped |

Plus: `ctx` is threaded through to `ResolveClient` (assert the stub receives the same context value).

**Unit — `pkg/lib/oauth/authz_service_test.go` (extend or new)**

`ListByUser` with two filters where the first drops: the second is not consulted for that authorization (short-circuit preserved).

**Unit — `pkg/auth/handler/oauth/consent_test.go` (new if absent)**

`renderConsentPage`'s view model, with a fake `ConsentRequired`:

- static client with `client_name: "Foo"` → `ClientName == "Foo"`, `ClientIDHostname == ""`;
- static client with **no** `client_name` → `ClientName` is the config's `Name`, **not** `""` (the §2.2 bug fix);
- CIMD client `https://mcp-client.example.com/oauth/client-metadata.json` with `client_name: "Example MCP Client"` → `ClientName == "Example MCP Client"`, `ClientIDHostname == "mcp-client.example.com"`;
- CIMD client with **no** `client_name` → `ClientName == "Client https://…"` (the `DisplayName()` fallback), `ClientIDHostname` still set;
- **static** client whose `client_id` is `https://pinned.example.com/x` → `ClientIDHostname == ""`. This is the pre-registration pattern, and it is the case a naive `strings.HasPrefix(clientID, "https://")` implementation gets wrong;
- CIMD client with `logo_uri` → `ClientLogoURI` set; without → empty.

**Golden template test**

Check whether the repo has a template-render golden test harness (`resources/authgear/templates` + `pkg/util/template`) — several AuthUI templates have snapshot-ish coverage. If one exists, add a consent-screen case asserting the hostname `<p>` appears with the right hostname and is absent for a static client. If not, do not build a harness for this; rely on the view-model tests plus the e2e below.

**e2e — `e2e/tests/cimd/consent.test.yaml` (new)**

Using the Part 3 §7 document-server fixture and its permissive feature-config override: complete an authorization for a CIMD client and assert the consent page HTML contains the `client_id`'s hostname and the document's `client_name`. Then a second app with a static third-party client, asserting the hostname line is absent. Follow the `write-e2e-test` skill for how existing tests assert on rendered HTML — if the harness cannot assert on page content, assert instead that the consent page renders 200 and leave content assertions to the unit tests, and say so.

**e2e — `e2e/tests/cimd/authorizations.test.yaml` (new)**

Complete an authorization for a CIMD client, then query the Admin API `user { authorizations { edges { node { clientID } } } }` and assert the CIMD `client_id` appears. Add the DCR equivalent in the same file — it is the same fix and it is worth having a test that would have caught the original bug. Model on `e2e/tests/admin_api/`.

**Commands to run**

```
go test ./pkg/lib/oauth/... ./pkg/auth/... ./pkg/admin/...
make generate && git status --porcelain        # must be empty
make export-schemas && git status --porcelain  # must be empty
make lint
make -C e2e run
```

Frontend: no `authui/src` or `portal/src` change — the consent screen is a server-rendered Go template and the Admin API shape is unchanged. So no `npm run typecheck` is required. Confirm by checking that nothing in `authui/src` references the consent page's DOM beyond generic components.

## 5. Fixed Behavioral Decisions

- **D1. The hostname is rendered as a separate line, not merged into the title or used as a `client_name` fallback.** It is the only non-self-asserted identifier on the screen and must be distinguishable from the ones the client wrote itself.
- **D2. The hostname is shown only for CIMD clients**, gated on `IsCIMDClient()`, not on a `client_id` prefix check. A static client with a URL-shaped `client_id` gets no banner.
- **D3. `ConsentViewModel.ClientName` is fed from `Client.Name`, not `Client.ClientName`.** Fixes the literal `"null"` title for any client without a `client_name`; benefits static, DCR and CIMD clients alike.
- **D4. `logo_uri` is rendered directly in this part, accepting the §9.2 privacy leak**, with `referrerpolicy="no-referrer"` as partial mitigation. Part 7 removes the leak. The CSP already permits cross-origin images (no `default-src`, no `img-src`), so this works as written; **no `img-src` directive is added** in either part.
- **D5. `AuthorizationFilter.Keep` takes a `context.Context`.** Four call sites; `ctx` is threaded, never captured on a struct.
- **D6. The static-client-id set is removed from `KeepThirdPartyAuthorizationFilter`.** It was the bug, not an optimisation; `ResolveClient` already fast-paths static clients.
- **D7. An unresolvable `client_id`'s authorization is dropped**, preserving today's behavior for removed static clients, at the cost of a since-deleted dynamic client's grant not being listable. Rationale and the rejected alternative are in §3.3.
- **D8. No Authorized Apps end-user surface is built.** §1.1. The data behind it is corrected; the surface needs its own design, and specifically a decision about what "revoke" does — there is no revocation code path today (§1.2).
- **D9. `SessionListingService.FilterForDisplay` is deliberately left unchanged**, even though its third-party client-id set is static-config-only like the two collections §3 does fix. §1.1: the resulting inclusion of dynamic third-party grants in the device list is the user's only means of cutting off a self-asserted third-party app, since no revoke-authorization path exists. Making it "consistent" with the static case would be a security regression, not a fix. Revisit only alongside the Authorized Apps surface.

## 6. Spec Updates

`doc:` commit against `docs/specs/cimd.md`:

1. **§ Reading a CIMD Client's Stored Config** and **§ Where resolution happens**: the "Authorized Apps settings page" does not exist in AuthUI. Replace those references with an accurate statement: the persisted record is what the Admin API's `user.authorizations` and any future end-user surface read, and no end-user page exists today. Do not silently drop the sentences — a reader needs to know the end-user surface is a gap, not a delivered feature.
2. **§ Privacy Considerations §9.2**: the claim that "the prefetch-and-cache handling of `logo_uri` already described in SSRF Protection … avoids" the browser-side leak is not true as of Part 6. Either defer this edit to Part 7 (which makes it true) or add an explicit "as of vN" note. Prefer deferring — Part 7 §6 owns this edit.
3. **§ Phishing Mitigation**: add the actual copy, or at least state that the screen presents the domain as the only verified fact, so the wording is reviewable as a security control rather than UI polish. Also note that the draft's own section title is **"OAuth Phishing Attacks"**, not "Phishing Mitigation" — cimd.md's heading and its §8.5 link text disagree with the draft, which makes the citation hard to follow.
4. **§ Phishing Mitigation**: the requirement is a **SHOULD**, not a MUST (draft §8.5). cimd.md states it as flat fact; saying which it is matters if anyone later wants to make the hostname conditional.

## 7. Atomic Commit Plan

1. `Fix the consent screen showing "null" when a client has no client_name` — the §2.2 `Client.Name` change alone, with its test. A real bug fix affecting static and DCR clients today; it should not be buried in a CIMD commit.
2. `[CIMD] Show the client_id hostname on the consent screen` — §2.2 (`ClientIDHostname`), §2.3 (template + translation key), consent view-model tests.
3. `[CIMD] Show the client logo on the consent screen` — §2.4. Separate commit so it can be reverted independently of the hostname work, and so Part 7's replacement of it is a legible diff.
4. `Fix third-party authorizations of dynamic clients being filtered out` — §3.2, §3.3, §3.4, §3.5, `authz_filters_test.go`. Titled as a fix, not `[CIMD]`: it repairs DCR too.
5. `[CIMD] Add e2e tests for the consent screen and third-party authorizations` — §4 e2e.
6. `doc: Correct cimd.md on the Authorized Apps surface and consent copy` — §6.

Body of each: `ref DEV-XXXX`. Run `make update-vettedpositions` after commit 4 (`authz_filters.go`, `authz_service.go`, `user.go` line numbers move).
