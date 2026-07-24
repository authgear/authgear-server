# Implementation Plan: `id_token_hint` support for RP-Initiated Logout

Spec reference: [docs/specs/oidc.md § RP-Initiated Logout](../../specs/oidc.md#rp-initiated-logout).

## 1. Goal / scope

The end session endpoint (`<endpoint>/oauth2/end_session`) already implements the
OIDC RP-Initiated Logout redirect mechanics (`post_logout_redirect_uri` validation,
`state` round-trip, the confirmation page at `/logout`). It does **not** implement
the spec's `id_token_hint` rule:

> If `id_token_hint` is present, and its `sid` matches the current logged in IdP
> session, AND the client is a first-party client, the session is logged out
> directly, without asking the end-user to confirm the logout action. In all
> other cases ... the end-user is shown a confirmation page.

Today, `pkg/lib/oauth/oidc/handler/handler_end_session.go` decides whether to log
out directly using only a `SameSiteStrict` companion cookie check
(`session.CookieDef.SameSiteStrictDef`). That check is a distinct, pre-existing
CSRF safeguard, not the id_token_hint rule, and in practice it almost never
succeeds for a genuine RP-initiated logout call, because `SameSite=Strict`
cookies are omitted on any cross-site navigation (which is what an RP calling
the end session endpoint always is), regardless of HTTP method.

This plan adds the missing `id_token_hint`-based direct-logout path, and fixes a
related session-cookie-visibility gap that a naive implementation would hit:

- The IdP session cookie is `SameSite=Lax`
  (`session.NewSessionCookieDef` in `pkg/lib/session/cookie.go`). A **cross-site
  top-level GET** navigation to the end session endpoint carries the Lax cookie
  (this is the entire point of Lax mode), so `session.GetSession(ctx)` already
  resolves correctly for `GET` calls today.
- A **cross-site POST** (the method RPs are expected to use specifically to keep
  `id_token_hint` out of the URL, per the spec's own "PII in the URL" section)
  does **not** reliably carry the Lax session cookie in current browsers. Today,
  this means a POST-based logout call sees `s == nil` regardless of whether the
  end-user actually has a session, and the handler silently falls through to
  the final redirect **without logging anything out**. This is a real gap
  independent of `id_token_hint`.
- The fix: when the request is a POST, redirect the browser to the same
  endpoint via a plain `302` (so the browser performs a fresh **top-level GET**
  navigation, which does carry the Lax cookie), before making any
  session/id_token_hint decision.
- Because the caller used POST specifically to avoid putting `id_token_hint` in
  a URL, the redirect-to-self must not put it there either. The request is
  sealed (AES-256-GCM, random per-request key) into an opaque blob carried in
  the redirect's query string; the key is carried in a short-lived, path-scoped
  cookie set on the same `302` response. Only the resumed GET, which holds both
  the cookie and the query blob, can open it. This is a **stateless** mechanism
  (no Redis / server-side session store): see [§3](#3-stash-mechanism-poststateless-seal).

## 2. Current runtime flow (baseline)

`pkg/auth/handler/oauth/end_session.go` (`EndSessionHandler.ServeHTTP`):
parses `r.Form` (query for GET, body for POST) into a
`protocol.EndSessionRequest` (`map[string]string`), opens a DB transaction, reads
`session.GetSession(ctx)`, and calls
`ProtocolEndSessionHandler.Handle(ctx, sess, req, r, rw)`.

`pkg/lib/oauth/oidc/handler/handler_end_session.go` (`EndSessionHandler.Handle`):

1. Reads the `SameSiteStrict` cookie. If a session is present and the cookie
   reads `"true"`, calls `SessionManager.Logout` and sets `s = nil`.
2. If `s != nil` (still), builds `endSessionURL` (original request forwarded as
   query params) and redirects to `/logout?redirect_uri=<endSessionURL>`
   (`URLs.LogoutURL`), which renders/handles the manual confirmation page
   (`pkg/auth/handler/webapp/logout.go`).
3. If `s == nil`, validates `post_logout_redirect_uri`, appends `state`, and
   responds via `oauth.WriteResponse`.

## 3. Stash mechanism (POST → stateless seal → redirect-to-self)

New file: `pkg/lib/oauth/oidc/handler/end_session_stash.go` (package `handler`).

```go
package handler

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc/protocol"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

// endSessionRefQueryParam carries the sealed, opaque request blob across the
// POST -> redirect-to-self -> GET round trip. It is safe to appear in the URL
// because it contains no plaintext PII; only the holder of the matching
// EndSessionRefKeyCookieDef cookie (set on the same origin, same response) can
// decrypt it.
const endSessionRefQueryParam = "x_end_session_ref"

const endSessionRefKeyCookieMaxAge = 300 // 5 minutes; only needs to survive one redirect round trip.

// EndSessionRefKeyCookieDef's NameSuffix mirrors endSessionRefQueryParam
// ("x_end_session_ref" -> "end_session_ref_key"), so the two are
// recognizable as a pair. **Renamed after review** from the original
// "end_session_stash": the cookie doesn't hold "the stash" (that's the
// sealed blob in the query string) — it holds the key that decrypts it, and
// its name should say so and read as belonging with the query param.
var EndSessionRefKeyCookieDef = &httputil.CookieDef{
	NameSuffix: "end_session_ref_key",
	Path:       "/oauth2/end_session",
	SameSite:   http.SameSiteLaxMode,
	MaxAge:     endSessionRefKeyCookieMaxAgePtr(),
}

func endSessionRefKeyCookieMaxAgePtr() *int {
	v := endSessionRefKeyCookieMaxAge
	return &v
}

// ErrEndSessionStashInvalid is returned when a resumed request carries a
// x_end_session_ref query value that cannot be opened: the cookie is missing (expired,
// blocked, or a different browser/tab), the key/ciphertext pairing doesn't
// authenticate (tampered or mismatched), or the payload doesn't decode. The
// caller (pkg/auth/handler/oauth/end_session.go) maps this to a 400 response.
var ErrEndSessionStashInvalid = errors.New("end_session: invalid or expired stash")

// sealEndSessionRequest encrypts req under a freshly generated random 256-bit
// key using AES-GCM. It returns the key (to be stored in
// EndSessionRefKeyCookieDef) and the sealed blob nonce||ciphertext||tag,
// both base64url-encoded (to be carried in endSessionRefQueryParam).
func sealEndSessionRequest(req protocol.EndSessionRequest) (key string, sealed string, err error) {
	keyBytes := make([]byte, 32)
	if _, err = rand.Read(keyBytes); err != nil {
		return "", "", err
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return "", "", err
	}

	plaintext, err := json.Marshal(req)
	if err != nil {
		return "", "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	key = base64.RawURLEncoding.EncodeToString(keyBytes)
	sealed = base64.RawURLEncoding.EncodeToString(ciphertext)
	return key, sealed, nil
}

// openEndSessionRequest reverses sealEndSessionRequest. Any failure (bad
// base64, bad key length, GCM authentication failure, bad JSON) collapses to
// ErrEndSessionStashInvalid; callers must not distinguish further, since the
// distinction is not actionable by the caller.
func openEndSessionRequest(key string, sealed string) (protocol.EndSessionRequest, error) {
	keyBytes, err := base64.RawURLEncoding.DecodeString(key)
	if err != nil {
		return nil, ErrEndSessionStashInvalid
	}
	ciphertext, err := base64.RawURLEncoding.DecodeString(sealed)
	if err != nil {
		return nil, ErrEndSessionStashInvalid
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, ErrEndSessionStashInvalid
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrEndSessionStashInvalid
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrEndSessionStashInvalid
	}
	nonce, ct := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, ErrEndSessionStashInvalid
	}

	var req protocol.EndSessionRequest
	if err := json.Unmarshal(plaintext, &req); err != nil {
		return nil, ErrEndSessionStashInvalid
	}
	return req, nil
}
```

Notes:

- `keyBytes` (32 bytes) selects AES-256. A fresh key is generated per POST, so
  compromise of one sealed blob (e.g. via a leaked `Location` header in a
  proxy log) does not help decrypt any other request.
- The cookie is scoped to `Path: "/oauth2/end_session"` (not `/`), minimizing
  where the browser will attach it.
- `MaxAge: 300` is generous for a same-machine redirect round trip that
  normally completes in well under a second; it bounds exposure if the
  redirect is never followed (e.g. user closes the tab).
- No Redis / `oauthsession` store is used. This was considered (the
  `oauthsession.StoreRedis` pattern already used by the authorize endpoint for
  the exact same "stash a request object across a redirect" problem) and
  rejected in favor of the stateless seal, per explicit direction: avoid adding
  a Redis dependency for a single-round-trip artifact.

## 4. `id_token_hint` → session/client resolution

**Revised after implementation.** The first version of this plan compared
`id_token_hint`'s `sid` claim against `oauth.EncodeSID(s)` for exact string
equality. This is wrong for the actual, spec-intended use case: a first-party
client that requested `offline_access` (the normal way a real client gets a
persistent login) receives an `id_token` whose `sid` is bound to *its own
offline grant* (`oauth.EncodeSID(offlineGrant)`), never directly to the
browser's IDP session cookie — confirmed by reading
`pkg/lib/oauth/handler/handler_token.go`'s authorization_code grant handling:
`accessTokenSessionID`/`Kind` are only ever assigned to the bare IDP session
type when the *original* `/oauth2/authorize` call itself already carried an
`id_token_hint` (the reauthentication case), which doesn't apply to a first
login. Exact `sid` string equality would therefore never match for the one
case the spec is actually about.

What *does* relate the offline grant to the login's IDP session is the SSO
group: `IssueOfflineGrantOptions.SSOEnabled` (set from the original
authorize request's `x_sso_enabled`) is stored on the grant
(`OfflineGrant.SSOEnabled`, `OfflineGrant.IDPSessionID`), and
`session.SessionBase.SSOGroupIDPSessionID()` / `ListableSession.IsSameSSOGroup`
(`pkg/lib/session/session.go`, implemented in
`pkg/lib/session/idpsession/session.go` and `pkg/lib/oauth/grant_offline.go`)
already exist precisely to answer "is this offline grant/session part of the
same login as that other session" — this is the same mechanism
`session.Manager.invalidate` (`pkg/lib/oauth/session_manager.go`) already uses
to decide which sessions get invalidated together on logout. So "matches the
current logged in IdP session" means `sidSession.IsSameSSOGroup(s)`, not
`sid == oauth.EncodeSID(s)`.

**Revised again, after a live debugging session with real HAR traces.** Pure
`IsSameSSOGroup` turned out to be too strict for this specific decision. Real
scenario: an ordinary OIDC client (a generic `openid-client`-based RP, not
using Authgear's own SDK) logs in — a real IDP session is created (per
`AuthorizationRequest.SuppressIDPSessionCookie`'s backward-compatibility
default of *not* suppressing the cookie when `x_sso_enabled` is absent
entirely) — then exchanges its code for an offline grant with
`IDPSessionID` set to that exact session, but `SSOEnabled: false` (per
`AuthorizationRequest.SSOEnabled()`'s strict `== "true"` check, which
defaults to `false` when the param is simply never sent — a client that has
never heard of Authgear's `x_sso_enabled` extension can't have sent it).
When that client then calls `end_session` with its own `id_token_hint`,
`IsSameSSOGroup` returns `false` (it requires `SSOEnabled`), so the
confirmation page is shown — even though the offline grant demonstrably came
from, and only from, this exact browser session. Investigated and confirmed
via `redis-cli GET app:<app>:offline-grant:<id>` in the live environment
that `idp_session_id` matched the session cookie's ID exactly, with no
`sso_enabled` key present in the stored JSON at all (`omitempty` on `false`).

**Three designs were tried for this, in order, before settling on the third:**

1. **A separate `IsCreatedFromSession` check, used only by `end_session`'s
   own decision.** `IsSameSSOGroup` seemed deliberately strict because it's
   also the mechanism behind two effects reaching beyond this one decision:
   `session.Manager.invalidate`'s cascade to *other* sessions/grants sharing
   the same IDP session, and `OfflineGrantService.GetOfflineGrant` tying a
   grant's own validity to its IDP session staying alive
   (`pkg/lib/oauth/grant_offline_service.go:103`). So a new,
   narrower method was added instead, and `Handle` called both
   `SessionManager.Logout` on the session *and*, when the match came only
   from this narrower check, a second explicit `Logout` call on the grant
   itself (since `invalidate`'s own cascade, keyed on `IsSameSSOGroup`,
   would never reach it). This worked, but needed real care to avoid a
   latent panic: calling `Logout` twice on a session already deleted by the
   first call's cascade hits `Manager.invalidate`'s `provider` staying a nil
   interface (since `Manager.List` no longer returns it), and
   `provider.ClearCookie()` on a nil interface panics. Avoiding this required
   computing, and returning, whether the match came from `IsSameSSOGroup`
   (cascade already covers it) or only from the new check (cascade won't),
   and calling the second `Logout` only in the latter case.
2. **Deriving `SSOEnabled` from `IDPSessionID` at issuance time**, in
   `doIssueTokensForAuthorizationCode` (`pkg/lib/oauth/handler/handler_token.go`):
   `SSOEnabled: offlineGrantIDPSessionID != ""` instead of
   `code.AuthorizationRequest.SSOEnabled()`. This fixes the data at the
   source instead of special-casing the check, and — since `IsSameSSOGroup`
   itself is unchanged — automatically makes `session.Manager.invalidate`'s
   cascade revoke the grant too, for free. Rejected in favor of design 3:
   changing token issuance is a much larger, harder-to-reason-about surface
   (every `authorization_code` exchange in the system, not just
   `end_session`), for no benefit over fixing the comparison directly.
3. **Changing `OfflineGrant.IsSameSSOGroup` itself** (the design implemented
   — see below). Simpler than both of the above: no new method, no
   `Handle`-level double-`Logout` bookkeeping, no token-issuance change. But
   this method has a *second* consumer — `pkg/lib/sessionlisting/listing.go`
   uses it to decide whether an offline-grant row should be flagged as the
   end-user's current device on the settings page — and changing it directly
   changes that consumer's behavior too. Confirmed this is the desired
   behavior (not a regression) before implementing: an unrelated existing
   unit test (`pkg/lib/sessionlisting/listing_test.go`) had a fixture
   constructing exactly this same "`IDPSessionID` matches, `SSOEnabled:
   false`" combination and asserting it should *not* flag as current/be
   grouped; per the same reasoning as this fix, it now should, and the test
   was updated to match (§16).

**Final design:** `pkg/lib/oauth/grant_offline.go`'s `IsSameSSOGroup` no
longer gates its own (`g`, the receiver's) side of the comparison on
`g.SSOEnabled` — only the *argument*'s side keeps using
`ss.SSOGroupIDPSessionID()` (still gated on the argument's own `SSOEnabled`
where applicable, e.g. when comparing two offline grants against each other):

```go
// IsSameSSOGroup returns true when the session argument
// - is the same offline grant
// - is the idp session that this offline grant's IDPSessionID names directly
//   (regardless of this offline grant's own SSOEnabled: creating an offline
//   grant from an existing/created IDP session is itself what makes it part
//   of that session's group, whether or not the client that requested it
//   also asked for cross-client SSO sharing)
// - is another offline grant in the same sso group (that other grant needs
//   to be sso enabled, via SSOGroupIDPSessionID, for grant-to-grant sharing
//   to apply)
func (g *OfflineGrant) IsSameSSOGroup(ss session.SessionBase) bool {
	if g.EqualSession(ss) {
		return true
	}

	if g.IDPSessionID == "" {
		return false
	}

	return g.IDPSessionID == ss.SSOGroupIDPSessionID()
}
```

`OfflineGrant.SSOGroupIDPSessionID()` itself is untouched (still gated on
`SSOEnabled`) — it has its own, correctly-still-strict consumer in
`pkg/lib/sessionlisting/listing.go` (deciding whether a grant should be
*combined/folded* into its IDP session's display row at all, a separate
concern from "is this row the current device"). `IDPSession.IsSameSSOGroup`
(`pkg/lib/session/idpsession/session.go`) is also untouched: the asymmetry is
intentional — revoking an IDP session now sweeps every grant created from it
regardless of that grant's own `SSOEnabled` (§16), but revoking a
`SSOEnabled: false` grant *directly* still does not sweep its IDP session,
matching the pre-existing, still-correct semantics for that direction.

`Handle`'s Step 4 (§5) goes back to exactly `sidSession.IsSameSSOGroup(s)` —
no new method, no second `Logout` call, no wrapper function. Verified live
(§13.3): `http_status: 303` for the `end_session` call, then `http_status:
400`/`invalid_grant` (not `200`) when presenting that same refresh token
afterward — the grant is genuinely revoked now, via `session.Manager.invalidate`'s
own cascade, automatically.

This also means the hint's `sid` claim alone (a string) isn't enough:
`IsSameSSOGroup` is a method on the *resolved* session/offline-grant object
(it needs to read `SSOEnabled`/`IDPSessionID`, which aren't encoded in the sid
string itself), so the sid must be decoded (`oauth.DecodeSID`) and the actual
object fetched — exactly what the existing
`oidc.IDTokenHintResolver.ResolveIDTokenHint` (in `pkg/lib/oauth/oidc/id_token.go`)
already does for the authorization endpoint's reauthentication flow. That
exact method still cannot be reused directly here: its signature takes
`protocol.AuthorizationRequest` (`pkg/lib/oauth/protocol`), whose
`IDTokenHint() (string, bool)` shape doesn't match
`oidc/protocol.EndSessionRequest.IDTokenHint() string`, so `EndSessionRequest`
doesn't satisfy that interface. A dedicated method is added instead, mirroring
`ResolveIDTokenHint`'s decode-and-fetch logic exactly (same two dependencies,
`IDTokenHintResolverSessionProvider`/`IDTokenHintResolverOfflineGrantService`,
just declared as this package's own interface types so wire can bind them
independently).

Add to `pkg/lib/oauth/oidc/handler/handler_end_session.go`:

```go
type IDTokenVerifier interface {
	VerifyIDToken(idTokenHint string) (idToken jwt.Token, err error)
}

// IDTokenHintSessionProvider resolves id_token_hint's sid when it names an
// IDP session (session.TypeIdentityProvider).
type IDTokenHintSessionProvider interface {
	Get(ctx context.Context, id string) (*idpsession.IDPSession, error)
}

// IDTokenHintOfflineGrantService resolves id_token_hint's sid when it names
// an offline grant (session.TypeOfflineGrant) — the common case for a
// first-party client that requested offline_access, whose id_token is bound
// to its own refresh token session, not directly to the browser's IDP
// session cookie.
type IDTokenHintOfflineGrantService interface {
	GetOfflineGrant(ctx context.Context, id string) (*oauth.OfflineGrant, error)
}
```

Add fields `IDTokenVerifier IDTokenVerifier`, `Sessions IDTokenHintSessionProvider`,
`OfflineGrants IDTokenHintOfflineGrantService` to `EndSessionHandler`.

Add method:

```go
// resolveIDTokenHintSession verifies idTokenHint's signature, extracts its
// aud (client_id) claim, and resolves its sid claim to the actual session or
// offline grant it names. ok is false if the token doesn't verify, is
// missing either claim, names a client that no longer exists, or names a
// session/offline grant that no longer exists — in all such cases the caller
// must treat it exactly like "no id_token_hint" (fall through to the
// confirmation page), not as an error: an unrecognized or malformed hint is
// not proof of anything, but it is also not a protocol violation by itself.
func (h *EndSessionHandler) resolveIDTokenHintSession(ctx context.Context, idTokenHint string) (client *config.OAuthClientConfig, sidSession session.ListableSession, ok bool) {
	idToken, err := h.IDTokenVerifier.VerifyIDToken(idTokenHint)
	if err != nil {
		return nil, nil, false
	}

	sidInterface, hasSID := idToken.Get(string(model.ClaimSID))
	sidStr, isString := sidInterface.(string)
	if !hasSID || !isString || sidStr == "" {
		return nil, nil, false
	}

	aud := idToken.Audience()
	if len(aud) != 1 {
		return nil, nil, false
	}

	client, ok = h.Config.GetClient(aud[0])
	if !ok {
		return nil, nil, false
	}

	typ, sessionID, ok := oauth.DecodeSID(sidStr)
	if !ok {
		return nil, nil, false
	}

	switch typ {
	case session.TypeIdentityProvider:
		sess, err := h.Sessions.Get(ctx, sessionID)
		if err != nil {
			return nil, nil, false
		}
		sidSession = sess
	case session.TypeOfflineGrant:
		grant, err := h.OfflineGrants.GetOfflineGrant(ctx, sessionID)
		if err != nil {
			return nil, nil, false
		}
		sidSession = grant
	default:
		return nil, nil, false
	}

	return client, sidSession, true
}
```

`VerifyIDToken` (in `pkg/lib/oauth/oidc/id_token.go`) intentionally does not
enforce `exp`, which already matches the spec's implicit requirement that
`id_token_hint` accept expired ID tokens (see the `IDTokenValidDuration` comment
in that file).

Wire binds needed in `pkg/lib/deps/deps_common.go` (alongside the existing
binds for `oidc.IDTokenHintResolverSessionProvider`/`OfflineGrantService`,
reusing the same concrete types):

```go
wire.Bind(new(oidchandler.IDTokenHintSessionProvider), new(*idpsession.Provider)),
wire.Bind(new(oidchandler.IDTokenHintOfflineGrantService), new(*oauth.OfflineGrantService)),
```

## 5. Revised `Handle` call flow

`pkg/lib/oauth/oidc/handler/handler_end_session.go`:

```go
func (h *EndSessionHandler) Handle(ctx context.Context, s session.ResolvedSession, req protocol.EndSessionRequest, r *http.Request, rw http.ResponseWriter) error {
	// Step 1: resume from a POST that was stashed on a previous pass through
	// this handler. Must run before anything else touches req or s. If the
	// stash cannot be opened (cookie missing, or doesn't match the sealed
	// value — expected in normal use, e.g. a stale browser-history link after
	// the short-lived stash cookie already expired), resumeFromStash logs it
	// and returns an empty request rather than an error: the rest of Handle
	// already knows how to handle a parameterless request.
	if sealed, hasStash := req[endSessionRefQueryParam]; hasStash {
		req = h.resumeFromStash(ctx, r, rw, sealed)

	// Step 2: a POST that hasn't been through the stash round trip yet. The
	// Lax session cookie may not be visible on this request even if the
	// end-user has a session (cross-site POST). Stash and force a same-origin
	// top-level GET so the Lax cookie becomes visible on the next pass.
	} else if r.Method == http.MethodPost {
		key, sealed, err := sealEndSessionRequest(req)
		if err != nil {
			return err
		}
		httputil.UpdateCookie(rw, h.Cookies.ValueCookie(EndSessionRefKeyCookieDef, key))
		selfURL := urlutil.WithQueryParamsAdded(
			h.Endpoints.EndSessionEndpointURL(),
			map[string]string{endSessionRefQueryParam: sealed},
		)
		httputil.Redirect(ctx, rw, r, selfURL.String(), http.StatusFound)
		return nil
	}

	idTokenHint := req.IDTokenHint()

	if idTokenHint == "" {
		// Step 3: existing SameSiteStrict fast path (unrelated CSRF
		// safeguard, preserved as-is; covers same-site navigations, e.g. a
		// link from Authgear's own settings page). Only consulted when no
		// id_token_hint was given at all — see the revised note below.
		sameSiteStrict, err := h.Cookies.GetCookie(r, h.SessionCookieDef.SameSiteStrictDef)
		if s != nil && err == nil && sameSiteStrict.Value == "true" {
			_, err := h.SessionManager.Logout(ctx, s, rw)
			if err != nil {
				return err
			}
			s = nil
		}
	} else if s != nil {
		// Step 4: id_token_hint fast path (spec: sid matches the current
		// logged in IdP session, and first-party client => direct logout, no
		// confirmation). "Matches" means the same SSO group, not sid string
		// equality: a first-party client that requested offline_access (the
		// normal case) gets an id_token bound to its own offline grant, not
		// directly to the browser's IDP session cookie, but that offline
		// grant and the IDP session cookie both trace back to the same login
		// via IDPSessionID, which is exactly what IsSameSSOGroup checks — for
		// an offline grant, regardless of its own SSOEnabled. See §4.
		if client, sidSession, ok := h.resolveIDTokenHintSession(ctx, idTokenHint); ok &&
			client.IsFirstParty() && sidSession.IsSameSSOGroup(s) {
			_, err := h.SessionManager.Logout(ctx, s, rw)
			if err != nil {
				return err
			}
			s = nil
		}
	}

	// Step 5: neither fast path fired; show the confirmation page. Strip
	// id_token_hint before forwarding: the /logout confirmation flow never
	// needs it (the direct-logout decision is already final by this point),
	// and forwarding it would defeat the POST-based / stash mechanism above
	// by re-exposing it in the /logout?redirect_uri=<...> URL.
	if s != nil {
		endSessionURL := urlutil.WithQueryParamsAdded(
			h.Endpoints.EndSessionEndpointURL(),
			req.WithoutIDTokenHint(),
		)
		logoutURL := h.URLs.LogoutURL(endSessionURL)
		httputil.Redirect(ctx, rw, r, logoutURL.String(), http.StatusFound)
		return nil
	}

	// Step 6: unchanged — no session, validate post_logout_redirect_uri and respond.
	redirectURI := req.PostLogoutRedirectURI()
	valid, client := h.validateRedirectURI(redirectURI)
	if !valid {
		if client != nil && client.ClientURI != "" {
			redirectURI = client.ClientURI
		} else {
			redirectURI = h.URLs.SettingsURL().String()
		}
		http.Redirect(rw, r, redirectURI, http.StatusFound)
		return nil
	}

	if state := req.State(); state != "" {
		uri, err := url.Parse(redirectURI)
		if err != nil {
			return err
		}
		redirectURI = urlutil.WithQueryParamsAdded(uri, map[string]string{"state": state}).String()
	}

	redirectURIURL, err := url.Parse(redirectURI)
	if err != nil {
		panic(err)
	}

	writeResponseOptions := oauth.WriteResponseOptions{
		RedirectURI:  redirectURIURL,
		ResponseMode: "query",
		UseHTTP200:   client.UseHTTP200(),
		Response:     make(map[string]string),
	}
	oauth.WriteResponse(rw, r, writeResponseOptions)
	return nil
}
```

Notes on ordering:

- Step 1/2 run before Step 3/4 unconditionally, so `s` (the session resolved
  from `session.GetSession(ctx)` before `Handle` was even called, by upstream
  middleware) is only trusted once we know the Lax cookie had a chance to be
  visible on this exact request. On the very first POST pass, `s` may be
  wrong (nil when a session exists); that's fine, because Step 2 returns
  immediately without reading `s`.
- No infinite loop: the resumed request is always a `GET` (an HTTP redirect
  always issues a GET, and `httputil.Redirect` here always uses
  `http.StatusFound`), so it can never re-enter the Step 2 branch (which
  requires `r.Method == http.MethodPost`).
- **Revised after implementation.** Step 3 and Step 4 are now mutually
  exclusive on whether `idTokenHint == ""`, not "Step 4 runs only if Step 3
  didn't already clear `s`" as originally planned. The original design ("keep
  both, id_token_hint as fallback" — SameSiteStrict checked unconditionally
  first, id_token_hint checked only if a session remained) meant a same-site
  `SameSiteStrict` cookie could log out a session even when a genuine
  `id_token_hint` was present but didn't actually resolve to the same SSO
  group — i.e. the hint's own (negative) verdict could be silently overridden
  by an unrelated cookie. The revised rule: once the caller supplies
  `id_token_hint` at all, its resolution is the sole authority on the
  direct-logout decision; `SameSiteStrict` is only consulted when no hint was
  given. This also removes the e2e test harness's need to strip the
  `SameSiteStrict` companion cookie before every hint-bearing call (§13.1
  point 6) for any call that carries `id_token_hint` — only calls with no
  hint at all (e.g. §13.3 case 4's second `end_session` call) still need it.

## 6. New / changed accessor

`pkg/lib/oauth/oidc/protocol/end_session.go`: add

```go
func (r EndSessionRequest) WithoutIDTokenHint() EndSessionRequest {
	out := EndSessionRequest{}
	for k, v := range r {
		if k == "id_token_hint" {
			continue
		}
		out[k] = v
	}
	return out
}
```

## 7. `CookieManager` interface widening

`pkg/lib/oauth/oidc/handler/handler_end_session.go` currently declares:

```go
type CookieManager interface {
	GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error)
}
```

Widen to:

```go
type CookieManager interface {
	GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error)
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
	ClearCookie(def *httputil.CookieDef) *http.Cookie
}
```

No wiring change is needed for this: `pkg/lib/deps/deps_utils.go:29` already
declares `wire.Bind(new(oidchandler.CookieManager), new(*httputil.CookieManager))`,
and `*httputil.CookieManager` (`pkg/util/httputil/cookie.go`) already implements
`GetCookie`, `ValueCookie`, and `ClearCookie`. Widening the local interface only
requires the already-bound concrete type to keep satisfying it, which it does.
`go build` / `make generate` will confirm this; no `wire_gen.go` change is
expected from this specific interface widening.

## 8. New dependencies: `IDTokenVerifier`, `Sessions`, `OfflineGrants`

Three new fields on `EndSessionHandler`
(`pkg/lib/oauth/oidc/handler/handler_end_session.go`), which is constructed via
`wire.Struct(new(EndSessionHandler), "*")`
(`pkg/lib/oauth/oidc/handler/deps.go`). None of these interface types exist yet
in the wire graph, so binds must be added:

`pkg/lib/deps/deps_common.go`:

```go
// alongside the existing oidc.IDTokenIssuer-related binds (~line 505):
wire.Bind(new(oidchandler.IDTokenVerifier), new(*oidc.IDTokenIssuer)),
// alongside the existing oidc.IDTokenHintResolverSessionProvider bind (~line 236):
wire.Bind(new(oidchandler.IDTokenHintSessionProvider), new(*idpsession.Provider)),
// alongside the existing oidc.IDTokenHintResolverOfflineGrantService bind (~line 492):
wire.Bind(new(oidchandler.IDTokenHintOfflineGrantService), new(*oauth.OfflineGrantService)),
```

(`oidchandler` is the existing import alias for `pkg/lib/oauth/oidc/handler` in
this file, already used for `oidchandler.LogoutSessionManager`.) All three
concrete types (`*oidc.IDTokenIssuer`, `*idpsession.Provider`,
`*oauth.OfflineGrantService`) already implement the exact method signatures
these interfaces declare — they're the same concrete types already bound to
the structurally-identical `oidc.IDTokenHintResolver*` interfaces used by the
authorization endpoint's reauthentication flow (§4) — so no changes are needed
to any of them.

These binds change the wire graph, so `wire_gen.go` (`pkg/auth/wire_gen.go` and
any other generated injector that constructs `oidchandler.EndSessionHandler`)
must be regenerated in the same commit via `make generate`. The resulting diff
is small and mechanical: one new `*oauth.OfflineGrantService` construction
(reusing providers already in scope for that injector) and two new struct
field assignments on `EndSessionHandler`.

## 9. Handling an invalid/expired POST logout stash

**Revised twice after implementation.** The first version had `Handle` return
`ErrEndSessionStashInvalid` up to `pkg/auth/handler/oauth/end_session.go`,
which mapped it to a `400`. On reflection, an invalid/expired stash is
*expected* to happen in normal use — the canonical case is the end-user
revisiting a stale `end_session` link from browser history after the
short-lived stash cookie has already expired — so surfacing any kind of error
response (even a `400`, rather than a `500`) is the wrong UX: it's not
exceptional, so it shouldn't look like a failure to the end-user.

The final design handles this entirely inside `Handle`
(`pkg/lib/oauth/oidc/handler/handler_end_session.go`), via a
`resumeFromStash` helper: if the stash cookie is missing, or doesn't match the
sealed value in the URL, it logs a `Warn` (so a spike is still visible/
alertable — e.g. a buggy RP integration retrying stale POSTs, or cookie-
blocking client software affecting a meaningful fraction of logout attempts)
and returns an **empty** `protocol.EndSessionRequest{}`, i.e. treats the
request exactly as if it had arrived with no parameters at all. The rest of
`Handle` already knows what to do with that (confirmation page if a session is
present, otherwise straight through) — no new response-writing code path is
needed, and `pkg/auth/handler/oauth/end_session.go`'s `ServeHTTP` is
completely unchanged from before this feature (its `err != nil` branch is
never reached by this case, since `Handle` returns `nil`).

```go
// resumeFromStash opens the request stashed by an earlier POST pass through
// this handler (see end_session_stash.go). If the stash cannot be opened,
// this is logged and treated as an end_session request with no parameters
// at all: the rest of Handle already knows how to handle that.
func (h *EndSessionHandler) resumeFromStash(ctx context.Context, r *http.Request, rw http.ResponseWriter, sealed string) protocol.EndSessionRequest {
	// A closure, not a bare `defer httputil.UpdateCookie(rw,
	// h.Cookies.ClearCookie(...))`: defer only postpones the outer call, not
	// evaluation of its arguments, so ClearCookie() would otherwise run
	// immediately, before GetCookie() below gets a chance to read the
	// still-live cookie. (This exact bug shipped in an earlier draft and was
	// only caught because the unit tests below assert on the *result* of the
	// round trip, not just that no error was returned.)
	defer func() {
		httputil.UpdateCookie(rw, h.Cookies.ClearCookie(EndSessionRefKeyCookieDef))
	}()

	cookie, err := h.Cookies.GetCookie(r, EndSessionRefKeyCookieDef)
	if err != nil {
		h.logInvalidStash(ctx, err)
		return protocol.EndSessionRequest{}
	}

	opened, err := openEndSessionRequest(cookie.Value, sealed)
	if err != nil {
		h.logInvalidStash(ctx, err)
		return protocol.EndSessionRequest{}
	}

	return opened
}

func (h *EndSessionHandler) logInvalidStash(ctx context.Context, cause error) {
	logger := EndSessionHandlerLogger.GetLogger(ctx)
	logger.WithError(cause).Warn(ctx, "end_session: invalid or expired logout stash")
}
```

`EndSessionHandlerLogger = slogutil.NewLogger("oidc-end-session")` is a new
package-level logger in `handler_end_session.go`, mirroring the exact pattern
already used by sibling packages (e.g. `pkg/lib/oauth/handler/handler_authz.go`'s
`AuthorizationHandlerLogger`). It logs via the `ctx` parameter `Handle` already
receives, not `r.Context()`, so this needs no `.vettedpositions` entry (the
`requestcontext` lint rule only fires on `r.Context()` call sites).

`ErrEndSessionStashInvalid` (`end_session_stash.go`) still exists and is still
returned by `openEndSessionRequest` — `resumeFromStash` is simply its only
caller now, and it never lets the error escape `Handle`.

## 10. Compatibility and deployment behavior

- No config schema changes. No new `x_application_type` values, no new client
  metadata.
- No persisted/storage state at all (the stash is a stateless sealed cookie +
  query pair); there is nothing to migrate, backfill, or dual-write, and no
  rollout ordering constraint between the app server and any datastore.
- Existing `GET` callers with no `id_token_hint`: unaffected. They still hit
  the `SameSiteStrict` check (almost always false for genuine cross-site
  calls) and then the confirmation page, exactly as today.
- Existing `GET` callers with `id_token_hint` from a first-party client whose
  `sid` matches: today they always see the confirmation page; after this
  change they log out directly. This is the intended, spec-mandated behavior
  change.
- Existing `POST` callers: today they silently no-op (see §1); after this
  change they get one extra `302` round trip (transparent to a browser
  following redirects normally) before reaching the same decision tree GET
  callers already go through. This is a behavior *fix*, not just an addition.
- Third-party clients: never get the `id_token_hint` fast path
  (`client.IsFirstParty()` gates it), matching the spec unchanged.
- Mixed-version rollout (rolling deploy): the stash cookie and its sealed
  payload are only ever produced and consumed by the same binary version
  within a single request's redirect round trip (sub-second in practice); an
  old-binary POST simply keeps today's (already broken) no-op behavior until
  its pod is replaced. No cross-version compatibility surface is introduced.

## 11. File-level change plan

| File | Change |
|---|---|
| `pkg/lib/oauth/oidc/handler/end_session_stash.go` | **New.** `EndSessionRefKeyCookieDef`, `ErrEndSessionStashInvalid`, `sealEndSessionRequest`, `openEndSessionRequest`, `endSessionRefQueryParam`. |
| `pkg/lib/oauth/oidc/handler/handler_end_session.go` | Add `IDTokenVerifier`, `IDTokenHintSessionProvider`, `IDTokenHintOfflineGrantService` interfaces + fields (`IDTokenVerifier`, `Sessions`, `OfflineGrants`); widen `CookieManager` interface; add `resolveIDTokenHintSession` (resolves the hint's `sid` to a session/offline-grant object and checks `IsSameSSOGroup`, not `sid` string equality — see §4); add `resumeFromStash`/`logInvalidStash` for graceful stash-invalid handling (see §9); rewrite `Handle` per §5. |
| `pkg/lib/oauth/oidc/protocol/end_session.go` | Add `WithoutIDTokenHint()`. |
| `pkg/lib/deps/deps_common.go` | Add `wire.Bind(new(oidchandler.IDTokenVerifier), new(*oidc.IDTokenIssuer))`, `wire.Bind(new(oidchandler.IDTokenHintSessionProvider), new(*idpsession.Provider))`, `wire.Bind(new(oidchandler.IDTokenHintOfflineGrantService), new(*oauth.OfflineGrantService))`. |
| `pkg/auth/wire_gen.go` (and any other generated injector wiring `oidchandler.EndSessionHandler`) | Regenerate via `make generate`. |
| `pkg/auth/handler/oauth/end_session.go` | **Unchanged** from before this feature — `resumeFromStash` handles the invalid-stash case entirely inside `Handle` (see §9), so `Handle` never returns `ErrEndSessionStashInvalid` and `ServeHTTP`'s `err != nil` branch is never reached by that case. (An earlier draft mapped this error to `400` here; superseded — see §9.) |
| `pkg/lib/oauth/oidc/handler/end_session_stash_test.go` | **New.** Seal/open round-trip unit tests. |
| `pkg/lib/oauth/oidc/handler/handler_end_session_test.go` | **New.** `Handle` behavior tests (see §12). |
| `pkg/lib/oauth/oidc/handler/handler_end_session_mock_test.go` | **New, generated.** `mockgen` output for `LogoutSessionManager`, `CookieManager`, `IDTokenVerifier`, `IDTokenHintSessionProvider`, `IDTokenHintOfflineGrantService`. |
| `pkg/lib/oauth/oidc/handler/handler_end_session.go` | Add `//go:generate go tool mockgen -source=handler_end_session.go -destination=handler_end_session_mock_test.go -package handler_test` (matching `pkg/lib/oauth/handler/handler_authz.go`'s convention), then run `go generate`. |
| `e2e/pkg/e2eclient/client.go` | Add `OAuthExchangeCodeResult.RawIDToken`; replace `SetupOAuth()` with `SetupOAuth(SetupOAuthOptions)`; add `ClientID`/`ClientSecret` to `OAuthExchangeCodeOptions`; add `ApproveConsent(redirectURI string)`; add `ClearCookies(names ...string)`. See §13.2. |
| `e2e/pkg/testrunner/models.go` | Add `Step.OAuthSetupClientID`, `OAuthSetupScope`, `OAuthSetupSSOEnabled`, `OAuthApproveConsentRedirectURI`, `OAuthExchangeCodeClientID`, `OAuthExchangeCodeClientSecret`, `ClearCookiesNames`; add `StepActionOAuthApproveConsent`, `StepActionClearCookies`; add `HTTPOutput.LocationNotContains`. See §13.2. |
| `e2e/pkg/testrunner/testcase.go` | Wire the new `Step` fields into `StepActionOAuthSetup`/`StepActionOAuthExchangeCode`; add `case StepActionOAuthApproveConsent`/`StepActionClearCookies`; add the `LocationNotContains` check to `validateHTTPOutput`. See §13.2. |
| `e2e/var/authgear.yaml`, `e2e/var/authgear.secrets.yaml` | Add shared `e2ethirdparty` (`x_application_type: third_party_app`) fixture client + secret. See §13.2. |
| `e2e/tests/oidc/end_session_id_token_hint.test.yaml` | **New.** E2E cases per §13.3. Uses only the standard `offline_access` + `oauth_exchange_code` flow to obtain `id_token_hint` — no `urn:authgear:params:oauth:grant-type:id-token` grant, per §4's SSO-group redesign. |

No changes to `docs/specs/oidc.md` are required by this plan — the spec already
documents the target behavior. (A short spec note about the internal
POST/stash mechanism could be added as a follow-up doc commit, matching the two
most recent commits on this branch, but it is not required for correctness and
is left out of this plan's atomic commits below.)

## 12. Unit test plan

Test style: this package (`pkg/lib/oauth/oidc/handler`) currently has no
`*_test.go` files. The sibling package `pkg/lib/oauth/handler` (same directory
depth, same kind of handler, e.g. `handler_authz_test.go`) uses Convey BDD style
(`. "github.com/smartystreets/goconvey/convey"`) with `gomock`-generated mocks
for interfaces and small hand-written mock structs for simple ones, in an
external `package handler_test`. Match that style exactly.

`pkg/lib/oauth/oidc/handler/end_session_stash_test.go` (pure functions, no
mocks needed):

1. Seal then open round-trips to the original `protocol.EndSessionRequest`.
2. Open with a key that doesn't match the sealed blob's key → `ErrEndSessionStashInvalid`.
3. Open with a truncated/corrupted sealed value → `ErrEndSessionStashInvalid`.
4. Open with a malformed base64 key or sealed value → `ErrEndSessionStashInvalid`.
5. Two `sealEndSessionRequest` calls for the same input produce different keys
   and different sealed values (confirms fresh randomness per call, i.e. no
   accidental key/nonce reuse).

`pkg/lib/oauth/oidc/handler/handler_end_session_test.go` (`Handle`, via
`gomock` for `LogoutSessionManager`/`IDTokenVerifier`/`IDTokenHintSessionProvider`/
`IDTokenHintOfflineGrantService` and a hand-written fake `CookieManager` backed
by an in-memory map, following the `mockCookieManager` pattern in
`handler_authz_test.go`). The "current session" fixture (`s`) is a real
`&idpsession.IDPSession{ID: "session-id"}` value, not the shared
`sessiontest.MockSession` helper: `MockSession.SSOGroupIDPSessionID()` is
hardcoded to return `""`, which can never satisfy a genuinely-matching
`IsSameSSOGroup` comparison, and widening that shared mock's behavior would
have unpredictable blast radius on other packages' tests that use it.

**Revised after implementation**: cases 1–3 below now hinge on SSO-group
membership (`IsSameSSOGroup`), not `sid` string equality — see §4. An offline
grant fixture (`&oauth.OfflineGrant{IDPSessionID: sess.ID, SSOEnabled: true}`)
is used to model the realistic case of a first-party client's `offline_access`
grant sharing a login with the session cookie; a bare IDP-session-typed `sid`
is also covered directly.

1. **GET, `id_token_hint`'s `sid` names an offline grant in the same SSO group
   as the current session, first-party client** → `SessionManager.Logout`
   called once; final response is the `post_logout_redirect_uri` response (not
   a redirect to `/logout`).
2. **GET, `id_token_hint`'s `sid` names the current IDP session directly**
   (`session.TypeIdentityProvider`, matching `s.ID`) → same silent-logout
   assertions as case 1.
3. **GET, matching SSO group, third-party client** → `SessionManager.Logout`
   not called; response is a redirect to `/logout?redirect_uri=...`; assert
   the forwarded URL's query does **not** contain `id_token_hint`. Proves
   `client.IsFirstParty()` gates the fast path independently of the SSO-group
   check.
4. **GET, `id_token_hint`'s `sid` names an offline grant from a *different*
   login** (`IDPSessionID` set to some other session's ID) → confirmation-page
   redirect (same assertions as case 3).
5. **Added after the `IsSameSSOGroup` fix (§4/§14). GET, `id_token_hint`'s
   `sid` names an offline grant with `SSOEnabled: false` but `IDPSessionID`
   matching the current session** → **silent logout**: `IsSameSSOGroup` no
   longer gates its own (the offline grant's) side of the comparison on
   `SSOEnabled` — only raw `IDPSessionID` equality is checked — this is the
   common case for any ordinary OIDC client unaware of `x_sso_enabled`.
6. **GET, `id_token_hint`'s `sid` names an offline grant with `SSOEnabled:
   false` *and* `IDPSessionID` naming a different session** →
   confirmation-page redirect: `IDPSessionID` mismatch alone is enough to
   fail the match, regardless of `SSOEnabled`; this is a genuinely unrelated
   grant, not just one that opted out of SSO sharing.
7. **GET, no `id_token_hint`, `SameSiteStrict` cookie `"true"`, session
   present** → `SessionManager.Logout` called (existing behavior preserved).
8. **GET, `id_token_hint` given but unresolvable (bad signature),
   `SameSiteStrict` cookie `"true"`, session present** → confirmation-page
   redirect, `SessionManager.Logout` **not** called. Proves `SameSiteStrict`
   is skipped entirely once `id_token_hint` is present at all, even though
   the same cookie would have triggered an unconditional silent logout in
   case 7 above. This is the case that would have failed under the original
   "SameSiteStrict checked first" ordering.
9. **GET, malformed/unverifiable `id_token_hint` (bad signature), no
   `SameSiteStrict` cookie** → treated as no hint: confirmation-page redirect,
   not an error.
10. **GET, no session (`s == nil`)** → unchanged existing behavior: straight
    to `post_logout_redirect_uri` / settings redirect, no confirmation, no
    `Logout` call.
11. **POST, `id_token_hint` naming a same-SSO-group offline grant, first-party**
    → first response is a `302` to `<end_session_endpoint>?x_end_session_ref=...`
    with a `Set-Cookie` for `EndSessionRefKeyCookieDef`; assert the `Location`
    header does **not** contain `id_token_hint` in any form. Feed the recorded
    `Set-Cookie` and the `x_end_session_ref` query value into a second `Handle`
    call (as a fresh `GET`, with `s` now populated, simulating the browser
    reattaching the Lax session cookie on the resumed top-level navigation) →
    `SessionManager.Logout` called once; final response is the
    `post_logout_redirect_uri` response. **Also asserts the stash cookie is
    actually cleared** on this resumed response (`assertStashCookieCleared`:
    the fake cookie manager's backing store no longer holds it, and the
    response carries a `Set-Cookie` with a negative `MaxAge`) — a stale
    `x_end_session_ref` revisited later (e.g. from browser history) must not
    be able to pair with a cookie that was never actually deleted.
12. **POST with no `id_token_hint` at all, session present** → same two-step
    round trip as case 11, ending at the confirmation-page redirect (proves
    the stash round trip runs unconditionally for POST, not only when
    `id_token_hint` is present).
13. **Resumed GET with `x_end_session_ref` set but the stash cookie missing,
    session present** → `resumeFromStash` logs a warning and treats the
    request as parameterless; since `s != nil` and there's no
    `post_logout_redirect_uri` left to honor, this resolves to the
    confirmation-page redirect, not an error. (Revised from an earlier draft
    that asserted `ErrEndSessionStashInvalid` — see §9.) Also asserts the
    stash cookie is cleared, same as case 11.
14. **Resumed GET with `x_end_session_ref` set and a stash cookie present but
    not matching (wrong key), session present** → same graceful fallback and
    assertions as case 13, including that the (non-matching) stash cookie
    the request actually carried is still cleared regardless of why opening
    it failed.

`pkg/auth/handler/oauth/end_session.go`'s `ServeHTTP` needs no dedicated unit
test for the invalid-stash case: it never sees `ErrEndSessionStashInvalid` at
all (`resumeFromStash` handles it entirely inside `Handle`, which always
returns `nil` for this case), so `ServeHTTP` is unchanged from its pre-feature
form and has no new behavior to cover. The graceful fallback itself is
exercised end-to-end by the e2e suite (§13.3, case 7).

## 13. E2E test plan

### 13.1 Harness gaps found in `e2e/pkg/e2eclient` and `e2e/pkg/testrunner`

Investigated the existing OAuth/OIDC e2e coverage
(`e2e/tests/oidc/userinfo.test.yaml`, `e2e/tests/saml/slo.test.yaml`,
`e2e/tests/m2m/token.test.yaml`) and the underlying client
(`e2e/pkg/e2eclient/client.go`) and step runner
(`e2e/pkg/testrunner/testcase.go`, `models.go`), then implemented the plan
below and ran it against a live server (Postgres/Redis/Deno via
`podman compose`, per `e2e/run.sh`) to find what static schema validation
alone could not catch. Several gaps and one wrong assumption surfaced only
by running:

1. `OAuthExchangeCodeResult` (`client.go:281`) only exposed decoded ID token
   claims (`IDToken map[string]any`) and `AccessToken`; the raw signed JWT
   string was parsed and then discarded. `id_token_hint` needs the raw
   string, not its decoded claims. Fixed by adding `RawIDToken string`.
2. `SetupOAuth()`/`OAuthExchangeCode()` hardcoded `client_id: "e2e"`,
   `redirect_uri: http://localhost:4000`, and (for setup)
   `scope: openid offline_access https://authgear.com/scopes/full-access`,
   with no way to drive the same login for a second client. Fixed by adding
   `ClientID`/`Scope`/`ClientSecret` options (empty defaults to the historical
   hardcoded values, so every existing e2e test using these methods is
   unaffected — confirmed by re-running `oidc/userinfo.test.yaml` and all of
   `saml/*.test.yaml` after the change).
3. No e2e action existed for approving the OAuth consent screen. Fixed by
   adding `Client.ApproveConsent(redirectURI string)` and an
   `oauth_approve_consent` step. `redirectURI` must be the login flow's
   `finish_redirect_uri`: reaching the consent page is not a fixed,
   parameter-free URL — it requires following the same-origin redirect chain
   from `finish_redirect_uri` through `/oauth2/authorize` (which decides
   consent is required) down to `/oauth2/consent`, which carries the oauth
   session reference in its own query/cookie state. An earlier version of
   this method did a bare `GET /oauth2/consent` with no such state and was
   caught immediately on the first live run (the third-party test case's
   `authenticate` step came back with `{"action": {"type": "finished"}}` and
   no `data` at all, because the login flow's own redirect_uri had silently
   gone nowhere).
4. No `x_application_type: third_party_app` client existed in the shared e2e
   fixture config (`e2e/var/authgear.yaml`); the closest fixture,
   `e2econfidential`, is `x_application_type: confidential`, which is
   first-party per the spec's client table. Added `e2ethirdparty` (`redirect_uris`/
   `post_logout_redirect_uris: [http://localhost:4000]` — **not**
   `http://localhost` as first drafted; that mismatch against
   `SetupOAuth`/`OAuthExchangeCode`'s hardcoded `http://localhost:4000` made
   every third-party `/oauth2/authorize` call return `400` on the first live
   run) and a matching `oauth.client_secrets` entry (`e2esecret`, same
   plaintext already used by `e2econfidential`/`e2em2mclient`).
5. `SetupOAuth()` hardcoded `x_sso_enabled=false`. This **suppresses the IDP
   session cookie entirely** (`AuthorizationRequest.SuppressIDPSessionCookie()`,
   `pkg/lib/oauth/protocol/authz.go:79-96`: true whenever `x_sso_enabled !=
   "true"`) — the client is expected to rely purely on tokens, which is fine
   for every *other* e2e test, but this feature is specifically about
   session-cookie-driven logout, so tests need a real cookie. Added
   `SetupOAuthOptions.SSOEnabled bool` (default `false`, preserving existing
   behavior for every other caller) and set it to `true` in every case in
   §13.3 below. This was only discovered by directly inspecting Set-Cookie
   response headers during a live run — the first attempt at cases 1/2
   "passed" outright, but for the wrong reason (see point 7).
6. This Go-based test client's cookie jar does not model `SameSite` policy at
   all (unlike a real browser — `net/http/cookiejar` only filters by domain,
   path, and expiry). The interactive login's session creation always sets
   the `SameSiteStrictDef` companion cookie (`session.CookieDef.SameSiteStrictDef`)
   alongside the real session cookie, and since the jar attaches every stored
   cookie regardless of `SameSite`, this companion cookie leaks into every
   later request in the same test case and makes the handler's pre-existing
   `SameSiteStrict` fast path fire unconditionally — masking the
   `id_token_hint` decision path entirely, independent of what the test
   actually sends. Added `Client.ClearCookies(names ...string)`: with no
   arguments it replaces the whole jar (used nowhere in the final test file,
   kept for general reuse); given names, it expires just those cookies,
   leaving the real session cookie (and everything else) intact. Wired as a
   `clear_cookies` step with an optional `clear_cookies_names` field.
   **Revised after §5's "SameSiteStrict only applies when no id_token_hint is
   given" fix**: every case's `end_session` call that carries `id_token_hint`
   (even a malformed one) no longer needs `clear_cookies` at all, since the
   handler now skips `SameSiteStrict` unconditionally whenever a hint is
   present. Only §13.3 case 4's *second* `end_session` call — which
   deliberately carries no `id_token_hint`, to prove the session survived the
   first call — still needs it, and it is called just before that one call
   instead of unconditionally near the top of the test case.
7. **Revised after further review.** A fresh `authorization_code` exchange
   without `offline_access` scope cannot issue *any* access token for a public
   client at all (`pkg/lib/oauth/handler/handler_token.go`'s `cannot issue
   access token` gate: `accessTokenSessionID`/`Kind` are only ever populated
   when `offline_access` was requested, when `client.IsConfidential()`, or
   when the original `/oauth2/authorize` call itself carried `id_token_hint`)
   — and *with* `offline_access`, the resulting `sid` is offline-grant-based
   (`oauth.EncodeSID(offlineGrant)`), not the IDP session's directly. An
   earlier draft of this test worked around this by minting the hint via the
   `urn:authgear:params:oauth:grant-type:id-token` grant instead (which reads
   the session cookie directly and mints `SID: oauth.EncodeSID(s)`) — but that
   grant is not how any real client obtains an `id_token_hint` in practice; the
   feature this test exists to cover is specifically "a client that used
   `offline_access`, got an `id_token`, and passes it to `end_session`", so
   exercising a different, unrelated grant type in its place would validate
   the wrong thing. The actual fix, per §4's SSO-group redesign: the
   production comparison itself was wrong, not the test. An offline grant
   issued `SSOEnabled: true` during a login shares that login's
   `IDPSessionID`, so `offlineGrant.IsSameSSOGroup(idpSession)` is exactly the
   right check — "matches the current logged in IdP session" means same SSO
   group, not identical `sid`. The e2e test therefore uses a plain
   `oauth_exchange_code` call (the normal, `offline_access`-scoped flow,
   already `SetupOAuth`'s default) with no grant-type workaround at all; see
   §13.3.
8. Dropped a planned `userinfo`-based post-logout check (assert `401` on the
   access token obtained before logout): the `e2e` fixture client has
   `issue_jwt_access_token: true`, so its access tokens are self-contained
   JWTs that remain valid until their own `exp`, independent of whether the
   underlying session/offline grant was revoked. That assertion was checking
   token statelessness, not this feature, and was removed once it failed for
   a reason unrelated to `id_token_hint`.
9. **Added after review: every case asserting a silent logout or its absence
   only checked the `end_session` response's status/redirect target, never
   that a session or offline grant was actually deleted from storage** — a
   handler bug that returned the right status code without actually calling
   `SessionManager.Logout` (or that called it when it shouldn't have) would
   have passed unnoticed. Fixed by adding `OAuthExchangeCodeResult.RefreshToken`
   (`json:"refresh_token"`, read straight off the token response alongside the
   existing `access_token`/`id_token` fields) and, after each silent-logout or
   no-op assertion, presenting that refresh token to an independent code path
   — a plain `grant_type=refresh_token` call to `/oauth2/token`, whose
   `ParseRefreshToken` looks the offline grant up from storage rather than
   trusting any in-request state — expecting `invalid_grant` (`400`) after a
   real logout, or `200` after a confirmed no-op. This directly exercises the
   SSO-group design (§4): the offline grant this refresh token belongs to is
   only revoked because it's in the same SSO group as the IDP session
   `end_session` logged out, not because it was the literal session named by
   the request.
10. **Found via point 9's new check, in the SSO-group-mismatch case:** logging
    in as user 2 while the client's cookie jar still carried user 1's session
    cookie caused the server to delete user 1's IDP session *and* offline
    grant outright (visible in the server log as `"delete IDP session"` /
    `"delete offline grant"` fired synchronously during user 2's login) —
    a "one login per browser" platform behavior entirely unrelated to
    `end_session`. This meant the mismatch case was accidentally passing for
    the wrong reason: `id_token_hint` was failing to resolve at all (the
    grant it named no longer existed), not resolving successfully to a
    genuinely different SSO group as the test intended. Fixed by calling
    `clear_cookies` (no `clear_cookies_names`, i.e. a full jar reset — a
    fresh "browser") between user 1's and user 2's logins, so the server
    never sees user 1's cookie again and user 1's session/grant survive,
    independently verifiable by point 9's refresh-token check.
11. **No case simulated a human actually confirming logout on the
    confirmation page.** Every case that reached `/logout` only checked that
    it was reached (`http_status: 200`, `redirect_path: /logout`); the
    confirmation form's own `PostAction("logout", ...)`
    (`pkg/auth/handler/webapp/logout.go`) — the code path that actually runs
    when an end-user clicks "Log out" — had no e2e coverage at all. Fixed by
    adding case 4 (§13.3): capture the confirmation page's exact URL via the
    `http_request` action's existing `result.http_final_url` field (already
    present in `NewResultHTTPResponse`, `e2e/pkg/testrunner/models.go`; no
    harness change needed), `POST` `x_action: logout` to it (the rendered
    form's only field — confirmed live that no CSRF token field is required,
    matching `ApproveConsent`'s identical GET-then-POST-to-final-URL
    pattern), and follow the resulting redirect chain back through
    `end_session` to the final OIDC response, with the same refresh-token
    revocation check as the silent-logout cases.
12. **No case modeled an OIDC client that never sends `x_sso_enabled` at
    all** — every existing case used `oauth_setup_sso_enabled: true` or
    (elsewhere) relied on the historical `false` default, but neither models
    the real-world client that motivated the `IsSameSSOGroup` fix
    (§4/§14): one that doesn't know the parameter exists. `SetupOAuth`
    always sent `x_sso_enabled` explicitly (`"true"` or `"false"`), so there
    was no way to construct this case at all. Added
    `SetupOAuthOptions.SSOEnabledOmitted bool` (`e2e/pkg/e2eclient/client.go`):
    when true, `x_sso_enabled` is omitted from the `/oauth2/authorize`
    request entirely, regardless of `SSOEnabled`'s value. Wired as
    `oauth_setup_sso_enabled_omitted` (`e2e/pkg/testrunner/models.go`,
    `testcase.go`), purely additive like the other `SetupOAuthOptions`
    fields.

### 13.2 Harness changes (implemented)

1. `e2e/pkg/e2eclient/client.go`:
   - `OAuthExchangeCodeResult.RawIDToken string` (`json:"raw_id_token"`), set
     from the already-parsed `idTokenStr` before it was discarded.
   - `OAuthExchangeCodeResult.RefreshToken string` (`json:"refresh_token"`),
     read straight off the token response — see §13.1 point 9 for why this
     was needed (independently verifying that a logout actually revoked the
     offline grant, not just that `end_session` returned the expected status).
   - `SetupOAuthOptions{ClientID string; Scope []string; SSOEnabled bool; SSOEnabledOmitted bool}`;
     empty `ClientID` defaults to `"e2e"`, empty `Scope` defaults to the
     historical hardcoded list, `SSOEnabled` defaults to `false`
     (`x_sso_enabled=false`, unchanged for every existing caller).
     `SSOEnabledOmitted: true` omits `x_sso_enabled` from the request
     entirely instead — see §13.1 point 12.
   - `OAuthExchangeCodeOptions.ClientID`/`ClientSecret string` (empty
     `ClientID` defaults to `"e2e"`; empty `ClientSecret` omits the token
     request's `client_secret` field exactly as before).
   - `func (c *Client) ApproveConsent(redirectURI string) (output map[string]any, err error)`:
     `GET redirectURI` via `c.HTTPClient` (follows the same-origin redirect
     chain down to the rendered consent page), reads the followed request's
     final URL, `POST`s that URL with `{"x_action": "consent"}` via
     `c.NoRedirectClient` (the post-consent redirect target — the client's own
     `redirect_uri` — is not a real, fetchable address in this harness, so it
     must not be followed), and returns `{"redirect_uri": <Location header>}`.
   - `func (c *Client) ClearCookies(names ...string)`: no-args replaces the
     jar (`cookiejar.New(nil)` wrapped in a fresh `HostAwareCookieJar`,
     re-pointed into `CookieJar`/`HTTPClient.Jar`/`NoRedirectClient.Jar`);
     given names, sets an already-expired (`MaxAge: -1`) cookie for each name
     via the existing jar instead, leaving other cookies untouched.
2. `e2e/pkg/testrunner/models.go` + `testcase.go`:
   - `Step.OAuthSetupClientID/OAuthSetupScope/OAuthSetupSSOEnabled`,
     `OAuthApproveConsentRedirectURI`, `ClearCookiesNames []string`,
     `OAuthExchangeCodeClientID/OAuthExchangeCodeClientSecret`, each with a
     matching JSON Schema property (`"oauth_setup_sso_enabled": {"type": "boolean"}`,
     `"clear_cookies_names": {"type": "array", "items": {"type": "string"}}`,
     etc.) and, for `oauth_approve_consent`, a `required` conditional on
     `oauth_approve_consent_redirect_uri`.
   - `StepActionOAuthApproveConsent = "oauth_approve_consent"` and
     `StepActionClearCookies = "clear_cookies"`, both added to the `action`
     enum, with matching `case` branches in the step switch.
   - `HTTPOutput.LocationNotContains []string` (`json:"location_not_contains"`)
     and a check in `validateHTTPOutput` that fails the step if
     `response.Header.Get("Location")` contains any listed substring.
3. `e2e/var/authgear.yaml` / `authgear.secrets.yaml`: shared `e2ethirdparty`
   fixture client (`x_application_type: third_party_app`,
   `redirect_uris`/`post_logout_redirect_uris: [http://localhost:4000, http://localhost:4000/after-logout]`)
   and a matching `oauth.client_secrets` entry, additive alongside
   `e2econfidential`/`e2em2mclient`.

Verified backward compatible: the full e2e suite (`go test ./pkg/testrunner/
-count 1 -timeout 10m`, fresh `podman compose` environment) passes with no
regressions, including `oidc/userinfo.test.yaml` and every `saml/*.test.yaml`
case, which exercise the same shared `SetupOAuth`/`OAuthExchangeCode`/
`InjectSession` code paths this change touched.

### 13.3 Test file (implemented and passing)

**Revised after the SSO-group redesign (§4).** All cases below now mint
`id_token_hint` via the normal `offline_access` + `oauth_exchange_code` flow
(`SetupOAuth`'s default scope already includes `offline_access`), never via
the `urn:authgear:params:oauth:grant-type:id-token` grant — see §13.1 point 7.

`e2e/tests/oidc/end_session_id_token_hint.test.yaml` (8 cases, sharing
`e2e/tests/oidc/end_session_id_token_hint_users.json`, two users:
`e2e_esh_user1`/`e2e_esh_user2`, same bcrypt fixture hash as
`oidc/users.json`, password `password`). **Case names and comments use "SSO
group", not "sid"**, to match §4's design — an earlier draft's case titles
("...sid matches current session...", "sid mismatch...") predated that
redesign and were corrected once the SSO-group check landed, since `sid` the
claim and "same SSO group" the actual check are no longer the same thing:

1. **First-party client, matching `id_token_hint` (same SSO group) → silent
   logout.** `oauth_setup` (`oauth_setup_sso_enabled: true`, default client
   `e2e`) → interactive login (`create`/`input` identify+authenticate) →
   `oauth_exchange_code` (code verifier + `finish_redirect_uri` from the login
   flow) to obtain `raw_id_token` — its `sid` is bound to the offline grant
   that exchange created, which was issued `SSOEnabled` from this same login,
   so it is in the same SSO group as the session cookie `end_session` will
   resolve → `GET /oauth2/end_session?id_token_hint=...&post_logout_redirect_uri=http://localhost:4000/after-logout`,
   `http_request_follow_redirects: false` (the final redirect target,
   `localhost:4000`, resolves to the *same* running server in this
   environment, so following it would 404 on a path it doesn't route — the
   point being tested is the redirect status/target, not that page). Assert
   `http_status: 303`. **No `clear_cookies` step is needed**: since
   `id_token_hint` is present, the handler skips the `SameSiteStrict` fast
   path regardless of that cookie's value (§5) — see point 6 above.
   **Then, independently verify the logout was real (§13.1 point 9)**:
   present `exchange_code`'s `refresh_token` to `grant_type=refresh_token`
   (`client_id: e2e`) and assert `http_status: 400`,
   `json_body: {"error": "invalid_grant"}` — proving the offline grant
   (same SSO group as the session cookie) was actually deleted from storage,
   not just that this one request returned `303`.

2. **Added after live debugging with real HAR traces (§13.1 point 12).
   First-party client that never sends `x_sso_enabled` at all → still gets
   silent logout, and its own grant is revoked too.** `oauth_setup_sso_enabled_omitted:
   true` (not `oauth_setup_sso_enabled: false` — this omits the parameter
   entirely, modeling a generic OIDC client unaware of Authgear's
   `x_sso_enabled` extension) → login → `oauth_exchange_code` → the
   resulting offline grant has `IDPSessionID` set (a real IDP session was
   still created, per `SuppressIDPSessionCookie`'s backward-compatibility
   default) but `SSOEnabled: false` → `GET /oauth2/end_session` with that
   `raw_id_token`, `follow_redirects: false` → assert `http_status: 303`
   (silent logout: `IsSameSSOGroup` matches on raw `IDPSessionID` equality
   alone, regardless of `SSOEnabled` — this is exactly the case that used to
   incorrectly show the confirmation page). Then, present the same
   `refresh_token` to `grant_type=refresh_token` (`client_id: e2e`) and
   assert `http_status: 400`, `json_body: {"error": "invalid_grant"}` — the
   grant is genuinely revoked too, via `session.Manager.invalidate`'s own
   cascade (which reuses this same, now-fixed `IsSameSSOGroup`), with no
   special-casing needed in `end_session`'s own handler code.

3. **First-party client, `POST` with the same kind of `id_token_hint` →
   stash round trip then silent logout.** Same login/`exchange_code` setup.
   `POST /oauth2/end_session` (form body, `follow_redirects: false`) → assert
   `http_status: 302` and `location_not_contains: [id_token_hint]`. Then `GET`
   the captured `{{ .steps.post_end_session.result.http_response_headers.location }}`
   (note: `http_response_headers` keys are lowercased by
   `NewResultHTTPResponse`) with `follow_redirects: false` → assert
   `http_status: 303`. Same refresh-token revocation check as case 1
   afterward.

4. **Third-party client, matching `id_token_hint` → confirmation page, not
   silent logout.** `oauth_setup` with `oauth_setup_sso_enabled: true`,
   `oauth_setup_client_id: e2ethirdparty`, `oauth_setup_scope: [openid, offline_access]`
   (third-party clients cannot request `full-access`) → login →
   `oauth_approve_consent` (`redirect_uri` = `finish_redirect_uri`) →
   `oauth_exchange_code` (`client_id`/`client_secret: e2ethirdparty`/`e2esecret`)
   → `GET /oauth2/end_session` with that `raw_id_token`, default
   `follow_redirects: true` (the confirmation page, `/logout`, is same-origin
   and real, so following it is safe) → assert `http_status: 200`,
   `redirect_path: /logout`. This client's offline grant is genuinely in the
   same SSO group as the session cookie (it came from the same login), so
   this case proves `client.IsFirstParty()` alone already forces the
   confirmation page, independent of the SSO-group check succeeding.
   **Then, the mirror check of case 1's**: present the same `refresh_token`
   to `grant_type=refresh_token` (`client_id`/`client_secret:
   e2ethirdparty`/`e2esecret`) and assert `http_status: 200` — proving the
   confirmation-page path is a strict no-op, not a delayed or partial logout.

5. **User confirms logout on the confirmation page → session and offline
   grant are revoked.** **New — closes a real gap**: every other case that
   reaches the confirmation page only ever checked that it was *reached*
   (`http_status: 200`, `redirect_path: /logout`); none of them simulated a
   human actually clicking "Log out" there, so the confirmation flow's own
   `PostAction("logout", ...)` (`pkg/auth/handler/webapp/logout.go`) had no
   e2e coverage at all. Same login/`exchange_code` setup as case 1, but no
   `id_token_hint` at all this time (so the confirmation page is reached via
   the "no hint" path, after `clear_cookies_names: [same_site_strict]` to
   stop that leaked cookie from short-circuiting straight to a silent
   logout) → `GET /oauth2/end_session?post_logout_redirect_uri=...` → assert
   `http_status: 200`, `redirect_path: /logout`, and capture
   `result.http_final_url` (the rendered confirmation page's exact URL,
   carrying the original request minus `id_token_hint` in its own
   `redirect_uri` query param — see `WithoutIDTokenHint`/`LogoutURL`). Then
   `POST` that same URL with `x_action: logout` (the confirmation form's only
   field, per `resources/authgear/templates/en/web/authflowv2/logout.html` —
   no CSRF token field is needed; `ApproveConsent`'s identical
   GET-then-POST-to-final-URL pattern already established this works),
   `follow_redirects: false` → assert `http_status: 302` (the handler logs
   the session out, then redirects back to that `redirect_uri` — the
   original `end_session` call, replayed now that the session is gone).
   `GET` that `Location` → assert `http_status: 303` (the same no-session
   OIDC response the silent-logout cases reach). Finally, the same
   refresh-token revocation check as case 1, proving the offline grant was
   actually deleted, not just that the confirmation page's own redirect
   chain looked right.

6. **SSO group mismatch → confirmation page even for a first-party client.**
   User 1 logs in and exchanges a code for an `id_token_hint` bound to their
   own offline grant (`exchange_code_1`). **`clear_cookies` (full jar reset,
   no `clear_cookies_names`) before user 2 logs in** — see §13.1 point 10:
   without this, the server would delete user 1's IDP session *and* offline
   grant outright as soon as user 2 logs in with user 1's cookie still in the
   jar (a "one login per browser" behavior unrelated to `end_session`),
   which would make this case pass for the wrong reason (`id_token_hint`
   failing to resolve at all, not resolving to a different SSO group). User 2
   then logs in as a fresh, unrelated browser session and exchanges their own
   code (`exchange_code_2`, used only to establish session 2's login).
   `GET /oauth2/end_session` with **user 1's** hint while the active session
   is user 2's → assert `http_status: 200`, `redirect_path: /logout` (user
   1's offline grant is not in user 2's SSO group). Then, **because this next
   call carries no `id_token_hint` at all**, `clear_cookies_names:
   [same_site_strict]` first (otherwise the leaked companion cookie from
   user 2's login would trigger the `SameSiteStrict` fast path and silently
   log the session out, defeating the point of this check — see point 6
   above), and call `end_session` again with **no** `id_token_hint`,
   asserting `http_status: 200`, `redirect_path: /logout` once more —
   proving user 2's session cookie is still alive (a logged-out session
   would instead fall through to the no-session `post_logout_redirect_uri`
   path). Finally, independently verify **both** users' offline grants
   survived the mismatch: present `exchange_code_1`'s and `exchange_code_2`'s
   `refresh_token`s to `grant_type=refresh_token` (`client_id: e2e`) and
   assert `http_status: 200` for both.

7. **Malformed `id_token_hint` (not a JWT at all) → treated as absent,
   confirmation page, not an error.** Login (no need to mint any real hint)
   → `GET /oauth2/end_session?id_token_hint=not-a-valid-jwt&...` → assert
   `http_status: 200`, `redirect_path: /logout`. No `clear_cookies` step
   needed: `id_token_hint` is present (even though malformed), which is
   enough to skip the `SameSiteStrict` fast path.

8. **No session at all with a valid-looking `id_token_hint` → straight to
   `post_logout_redirect_uri`, no confirmation, no error.** No login at all
   in this case. `GET /oauth2/end_session?id_token_hint=not-a-valid-jwt&...`,
   `follow_redirects: false` → assert `http_status: 303`.

9. **Invalid/expired stash → graceful fallback, not an error.** **Revised**:
   an earlier draft asserted `http_status: 400` here (see §9's first
   revision). The final design treats an unopenable stash as a parameterless
   `end_session` request, not an error. `GET /oauth2/end_session?x_end_session_ref=not-a-real-sealed-value`
   with no session and no stash cookie ever set, `follow_redirects: false` →
   `resumeFromStash` logs a warning and returns an empty request; with no
   session and no `post_logout_redirect_uri` left to honor, this falls
   through to the invalid-redirect-URI branch, which redirects to the
   settings page → assert `http_status: 302`.

All 9 cases pass against a live server; the exact status codes above
(`303`/`302`/`200`) are the real, observed values — the client's
`UseHTTP200()` is `false` for both `e2e` and `e2ethirdparty` (neither sets
`CustomUIURI`), so the direct-logout path is `oauth.WriteResponseOptions`'s
default `303` (`HTTP303HTMLRedirect`), and the confirmation-page redirect
(`httputil.Redirect(..., http.StatusFound)`, `302`) is followed through by
default (`http_request_follow_redirects` defaults to `true`) to the
confirmation page itself (`200` HTML), which is what `redirect_path: /logout`
and `http_status: 200` together assert.
## 14. Fixed behavioral decisions

(Decided; not open questions.)

- **Revised after implementation.** `SameSiteStrict`-cookie direct logout only
  applies when the request carries **no** `id_token_hint` at all; the moment
  an `id_token_hint` is given, its own resolution is the sole authority on
  whether to log out directly, and the `SameSiteStrict` cookie is not
  consulted. The original design checked `SameSiteStrict` unconditionally
  first, with `id_token_hint` only as a fallback if a session remained
  afterward — but that let an unrelated same-site cookie override a hint that
  had already (correctly) failed to match, and it forced e2e tests to strip
  the `SameSiteStrict` cookie before every hint-bearing call to observe the
  hint logic at all (see §13.1 point 6 / the note in §5).
- **"Matches the current logged in IdP session" means `sidSession.IsSameSSOGroup(s)`,
  not `sid` string equality — and `OfflineGrant.IsSameSSOGroup` itself no
  longer requires `SSOEnabled` on its own (receiver) side.** Revised after
  implementation, then revised twice more after live debugging — see §4 for
  the full progression (a separate `IsCreatedFromSession` check with a
  `Handle`-level double-`Logout` call; deriving `SSOEnabled` from
  `IDPSessionID` at token-issuance time; settling on changing
  `IsSameSSOGroup` itself, the simplest of the three). A first-party
  client's normal, `offline_access`-based `id_token` is bound to its own
  offline grant, not directly to the IDP session cookie; comparing raw `sid`
  strings would therefore never match for the spec's actual target case.
  The original `IsSameSSOGroup` (requiring `SSOEnabled: true` on the offline
  grant) was *also* insufficient: any ordinary OIDC client that never sends
  `x_sso_enabled` (ordinary in practice — that's an Authgear-specific
  extension) still gets a real IDP session, but its offline grant's
  `SSOEnabled` defaults to `false`, so `IsSameSSOGroup` would always deny it
  — even though `IDPSessionID` demonstrably names that exact session. Fixed
  by dropping that gate on the offline grant's own side of the comparison
  (raw `IDPSessionID` equality now suffices); the argument's side
  (`ss.SSOGroupIDPSessionID()`) is untouched, so grant-to-grant comparisons
  still correctly require both sides to have opted in. This also means
  `session.Manager.invalidate`'s cascade (which reuses this exact method)
  now revokes such a grant when its own IDP session logs out — a
  deliberate, confirmed widening (§16), not an accidental side effect; see
  §4 for why this reaches `pkg/lib/sessionlisting/listing.go`'s "current
  device" flag too, and why that's also correct.
- An unresolvable `id_token_hint` (bad signature, missing `sid`/`aud` claim,
  unknown client, or a `sid` naming a session/offline grant that no longer
  exists) is treated identically to "no `id_token_hint` given" — falls through
  to the confirmation page. It is never an error.
- An unresolvable **stash** (`x_end_session_ref` present but cookie missing, or
  cookie/query pairing fails to authenticate) is logged (`Warn`) and treated
  as an `end_session` request with **no parameters at all**, not as an error.
  **Revised twice** — see §9: the first draft returned `ErrEndSessionStashInvalid`
  up to `pkg/auth/handler/oauth/end_session.go`, mapped to `400`; on reflection
  this is expected to happen in normal use (e.g. a stale browser-history link
  after the short-lived stash cookie expired), so surfacing any error response
  is the wrong UX. The rest of `Handle` already knows what to do with a
  parameterless request.
- The self-redirect-and-stash applies to **every** POST request, regardless of
  whether `id_token_hint` is present, because the underlying problem (Lax
  session cookie invisibility on cross-site POST) affects the session-presence
  check itself, not just the `id_token_hint` check.
- `id_token_hint` is stripped from the request before it is ever forwarded to
  the `/logout` confirmation page, regardless of whether the original request
  was `GET` or `POST`, since the confirmation flow never needs it and
  forwarding it would re-expose it in a new URL.
- No Redis/`oauthsession` storage is used for the stash; it is a stateless
  AES-256-GCM seal with a random per-request key carried in a short-lived,
  path-scoped cookie.

## 15. Implementation order

1. `end_session_stash.go` + `end_session_stash_test.go` (pure, no dependencies
   on the rest of the change; can be fully verified in isolation).
2. `protocol/end_session.go`: `WithoutIDTokenHint()`.
3. `handler_end_session.go`: `IDTokenVerifier` interface/field,
   `resolveIDTokenHintClient` (original name; renamed `resolveIDTokenHintSession`
   in step 12 below), widened `CookieManager`, rewritten `Handle`.
4. `deps_common.go` wire bind + `make generate` (regenerate `wire_gen.go`).
5. `pkg/auth/handler/oauth/end_session.go`: `ErrEndSessionStashInvalid` → 400
   (superseded by step 11 below).
6. `handler_end_session_test.go` + generated mocks (`go generate`).
7. `make update-vettedpositions` if `make lint`'s goanalysis output shifts any
   existing line-number-keyed entries as a result of the above.
8. E2E harness changes (§13.2): `e2eclient/client.go`, `testrunner/models.go`,
   `testrunner/testcase.go`, fixture config in `e2e/var/authgear.yaml` /
   `authgear.secrets.yaml`. Done after 1–7 so the harness's raw-ID-token and
   consent-approval additions can be validated against the real, already-built
   feature rather than against speculative behavior.
9. `e2e/tests/oidc/end_session_id_token_hint.test.yaml` (§13.3), run via
   `cd e2e && make teardown && make setup` then `go test ./pkg/testrunner/
   -count 1 -v -timeout 10m -run "TestAuthflow/oidc/end_session_id_token_hint"`,
   fixed up until green, per the `write-e2e-test` skill's standard loop.
10. `resumeFromStash`/`logInvalidStash`: replace step 5's `400` mapping with
    graceful in-`Handle` fallback (§9, atomic commit 9).
11. `resolveIDTokenHintSession`: replace step 3's `sid`-equality comparison
    with the SSO-group check (§4, atomic commit 10), regenerate mocks and
    `wire_gen.go` again.
12. Fix the e2e test file to use `oauth_exchange_code` instead of the
    `id-token` grant workaround, and update case 7's expectation to match step
    10's graceful fallback (atomic commit 11); re-verify live per step 9's
    loop.

## 16. Atomic commit plan

1. **`Add stateless seal/open helpers for end_session POST stash`**
   - Files: `pkg/lib/oauth/oidc/handler/end_session_stash.go`,
     `pkg/lib/oauth/oidc/handler/end_session_stash_test.go`.
   - Pure addition, no existing behavior touched, fully covered by its own
     unit tests.

2. **`Add WithoutIDTokenHint to EndSessionRequest`**
   - Files: `pkg/lib/oauth/oidc/protocol/end_session.go`.
   - Pure addition.

3. **`Implement id_token_hint direct logout and fix POST session-cookie visibility for end_session`**
   - Files: `pkg/lib/oauth/oidc/handler/handler_end_session.go`,
     `pkg/lib/oauth/oidc/handler/handler_end_session_mock_test.go` (generated),
     `pkg/lib/oauth/oidc/handler/handler_end_session_test.go`,
     `pkg/lib/deps/deps_common.go`, `pkg/auth/wire_gen.go`.
   - Must land together: the new `IDTokenVerifier` field requires the wire
     bind, which requires `wire_gen.go` regeneration, which the tests exercise
     indirectly by constructing `EndSessionHandler` directly (tests don't run
     through wire, but a broken wire graph would fail `make generate`/build).
   - Run `go generate ./pkg/lib/oauth/oidc/handler/...` before committing the
     mock file.

4. **`Return 400 for invalid or expired end_session logout stash`**
   - Files: `pkg/auth/handler/oauth/end_session.go`, and its test file (new or
     existing, per whatever is found at implementation time).
   - Small, independently reviewable; only meaningful after commit 3 exists
     (introduces `ErrEndSessionStashInvalid`), so it must land after it, but
     is kept separate for bisectability of the 400-vs-500 behavior change.

5. **`chore: Update .vettedpositions`** (only if step 15.7 produced changes).

6. **`Extend e2e harness for raw ID tokens, alternate OAuth clients, and consent approval`**
   - Files: `e2e/pkg/e2eclient/client.go`, `e2e/pkg/testrunner/models.go`,
     `e2e/pkg/testrunner/testcase.go`, `e2e/var/authgear.yaml`,
     `e2e/var/authgear.secrets.yaml`.
   - Purely additive/backward-compatible: existing `oauth_setup` /
     `oauth_exchange_code` call sites keep working via the new options'
     zero-value defaults (`ClientID` defaults to `"e2e"`, etc.); the new
     `e2ethirdparty` fixture client and `location_not_contains` assertion
     don't affect any existing test. Kept separate from commit 7 so a harness
     regression is bisectable independently of the feature-specific test
     content.
   - Verify with `cd e2e && make teardown && make setup` followed by the full
     existing e2e suite (or at minimum `TestAuthflow/oidc/...` and
     `TestAuthflow/saml/...`, which exercise `SetupOAuth`/`OAuthExchangeCode`
     and `http_request_session_cookie` respectively) to confirm no regression
     before adding the new test file.

7. **`Add e2e tests for end_session id_token_hint`**
   - Files: `e2e/tests/oidc/end_session_id_token_hint.test.yaml`.
   - Depends on commit 6. Run per §15 step 9 until green; fix the test file
     (not the feature code, unless the run surfaces a genuine bug) if
     assertions don't match observed behavior.

8. **`Fix e2e harness and tests after running against a live server`**
   - Files: `e2e/pkg/e2eclient/client.go` (`SetupOAuthOptions.SSOEnabled`,
     `Client.ClearCookies`), `e2e/pkg/testrunner/models.go`/`testcase.go`
     (`oauth_setup_sso_enabled`, `clear_cookies`/`clear_cookies_names`),
     `e2e/tests/oidc/end_session_id_token_hint.test.yaml`,
     `e2e/var/authgear.yaml` (`e2ethirdparty`'s `redirect_uris` fixed to
     `http://localhost:4000`).
   - Commits 6–7 were written and schema-validated but never run against a
     live server. Doing so (§13.1, points 5–8) found that cases 1/2/4 were
     passing for the wrong reason (no session cookie at all, due to
     `x_sso_enabled=false`, masking whether `id_token_hint` was evaluated at
     all) and that case 3 was failing outright (a `redirect_uris` mismatch
     that a schema check can't catch, since it's a cross-file config/hardcoded-
     value consistency issue, not a shape violation). This commit is kept
     separate from 6–7 specifically because it was authored only after live
     verification, unlike them.

9. **`Handle invalid/expired end_session logout stash gracefully`** (landed)
   - Files: `pkg/lib/oauth/oidc/handler/handler_end_session.go` (adds
     `resumeFromStash`/`logInvalidStash`, `EndSessionHandlerLogger`),
     `pkg/lib/oauth/oidc/handler/handler_end_session_test.go`,
     `pkg/lib/oauth/oidc/handler/end_session_stash.go` (doc comment update),
     `pkg/auth/handler/oauth/end_session.go` (reverted to its pre-feature
     form), `.vettedpositions` (reverted to its pre-commit-4 entries).
   - Supersedes commit 4's `400` mapping per §9/§14: an invalid/expired stash
     is expected in normal use (stale browser-history link), so it is logged
     and treated as a parameterless request instead of surfaced as an error.

10. **`Compare id_token_hint by SSO group, not sid equality`**
    - Files: `pkg/lib/oauth/oidc/handler/handler_end_session.go` (new
      `IDTokenHintSessionProvider`/`IDTokenHintOfflineGrantService` interfaces
      and `Sessions`/`OfflineGrants` fields, `resolveIDTokenHintSession`
      replacing the old `sid`-equality comparison, updated `Handle`),
      `pkg/lib/oauth/oidc/handler/handler_end_session_test.go` (rewritten
      per §12, real `*idpsession.IDPSession` fixture instead of
      `sessiontest.MockSession`), `pkg/lib/oauth/oidc/handler/handler_end_session_mock_test.go`
      (regenerated via `go generate ./pkg/lib/oauth/oidc/handler/...`),
      `pkg/lib/deps/deps_common.go` (two new wire binds), `pkg/auth/wire_gen.go`
      (regenerated via `make generate`).
    - Must land together: the new fields require the wire binds, which
      require `wire_gen.go` regeneration.
    - Fixes the fundamental design error in commit 3: exact `sid` string
      equality can never match for a first-party client's normal
      `offline_access`-based `id_token` (bound to an offline grant, not the
      IDP session cookie directly). See §4/§14.

11. **`Fix e2e test to use offline_access instead of an id-token grant workaround`**
    - Files: `e2e/tests/oidc/end_session_id_token_hint.test.yaml`.
    - Removes every `id_token_grant`/`id_token_grant_1`/`id_token_grant_2`
      step (the `urn:authgear:params:oauth:grant-type:id-token` workaround) and
      replaces them with the standard `oauth_exchange_code` flow (already
      `SetupOAuth`'s default `offline_access` scope); updates case 7's
      expectation from `400` to `302` (graceful fallback per commit 9, not an
      error); depends on commit 10 (the SSO-group comparison must exist for
      the offline-grant-based hints this test now sends to actually match).
    - Verified against a live server: `cd e2e && ./run.sh setup` then
      `go test ./pkg/testrunner/ -count 1 -v -timeout 10m -run
      "TestAuthflow/oidc/end_session_id_token_hint"` (all 7 cases pass) and
      `-run "TestAuthflow/oidc"` (no regressions), then `./run.sh teardown`.

12. **`Only check SameSiteStrict when no id_token_hint is given`**
    - Files: `pkg/lib/oauth/oidc/handler/handler_end_session.go` (Steps 3/4
      restructured into an `if idTokenHint == "" { ... } else if s != nil {
      ... }` branch instead of two independent, sequential checks),
      `pkg/lib/oauth/oidc/handler/handler_end_session_test.go` (new case: hint
      given but unresolvable + `SameSiteStrict` cookie `"true"` → confirmation
      page, `SameSiteStrict` not consulted), `e2e/tests/oidc/end_session_id_token_hint.test.yaml`
      (drops `clear_cookies`/`remove_same_site_strict` from every case whose
      `end_session` call carries `id_token_hint`, keeping it only for case 4's
      second, hint-less call).
    - Behavior change: previously `SameSiteStrict` was checked unconditionally
      first, so a same-site `SameSiteStrict` cookie could silently override a
      negative `id_token_hint` verdict. Now, once `id_token_hint` is present
      at all, its resolution is the sole authority; `SameSiteStrict` only
      applies when no hint was given. See §5/§14.
    - Verified: unit tests (`go test ./pkg/lib/oauth/oidc/handler/...`) and
      the full e2e run (`cd e2e && ./run.sh setup`, `go test
      ./pkg/testrunner/ -count 1 -v -timeout 10m -run
      "TestAuthflow/oidc/end_session_id_token_hint"` and `-run
      "TestAuthflow/oidc"`, then `./run.sh teardown`) both pass.

13. **Plan doc update** (this document): §4, §5, §8, §9, §11, §12, §13.1,
    §13.3, §14, §16 updated to describe the SSO-group redesign, the e2e
    proper-scopes fix, and the `SameSiteStrict`-priority fix, replacing the
    stale `sid`-equality design and the `id_token_grant`-workaround test plan.

14. **`Verify end_session actually revokes the underlying offline grant in e2e tests`**
    - Files: `e2e/pkg/e2eclient/client.go` (`OAuthExchangeCodeResult.RefreshToken`),
      `e2e/tests/oidc/end_session_id_token_hint.test.yaml` (refresh-token
      revocation/no-op checks added to cases 1–4; case 4's setup fixed to
      reset the cookie jar between logins so its liveness checks test a
      genuine SSO-group mismatch rather than an already-deleted grant).
    - Every prior version of this test only asserted `end_session`'s own
      response status/redirect target, never that a session or offline grant
      was actually deleted from (or preserved in) storage — a real gap: a
      handler bug that returned the right status without actually revoking
      anything (or that revoked when it shouldn't have) would have passed
      unnoticed. See §13.1 points 9–10 for what running this live surfaced,
      including a platform "one login per browser" behavior that had been
      accidentally masking case 4's intended assertion.
    - Verified against a live server: `cd e2e && ./run.sh setup` then
      `go test ./pkg/testrunner/ -count 1 -v -timeout 10m -run
      "TestAuthflow/oidc/end_session_id_token_hint"` and the full suite
      (`go test ./pkg/testrunner/ -count 1 -timeout 10m`), then
      `./run.sh teardown`.

15. **`Rename stale sid-based test names to SSO group, and cover interactive logout confirmation`**
    - Files: `e2e/tests/oidc/end_session_id_token_hint.test.yaml` (case
      titles/comments referring to "sid matches"/"sid mismatch" renamed to
      "SSO group", per §4's redesign; new case 4 added — a human confirming
      logout on the `/logout` page, verified end-to-end including offline
      grant revocation).
    - Two review findings, both real gaps: (1) case titles and comments still
      described the pre-redesign `sid`-equality check even though the
      production comparison had moved to SSO-group membership; (2) no case
      exercised `PostAction("logout", ...)` (`pkg/auth/handler/webapp/logout.go`)
      — the code path that runs when an end-user actually clicks "Log out" on
      the confirmation page — every other case only checked that the page was
      *reached*. See §13.1 points 10 (renamed) and 11 (new case).
    - Verified against a live server: all 8 cases pass, plus the full e2e
      suite for regression.

16. **`Assert the end_session stash cookie is actually cleared after being consumed`**
    - Files: `pkg/lib/oauth/oidc/handler/handler_end_session_test.go` (new
      `assertStashCookieCleared` helper; applied to the successful POST round
      trip and both graceful-fallback "resumed GET" cases — see §12 cases
      10, 12, 13).
    - Both properties this checks — `EndSessionRefKeyCookieDef.MaxAge = 300`
      (§3) and `resumeFromStash`'s unconditional `defer`-based clear (§9) —
      already existed in the implementation; this commit only makes them
      test-verified rather than merely implicit in the response code
      assertions.

17. **`Rename end_session_stash cookie to end_session_ref_key`**
    - Files: `pkg/lib/oauth/oidc/handler/end_session_stash.go`
      (`EndSessionStashCookieDef` → `EndSessionRefKeyCookieDef`,
      `NameSuffix: "end_session_stash"` → `"end_session_ref_key"`, and the
      matching `endSessionStashCookieMaxAge*` identifiers),
      `pkg/lib/oauth/oidc/handler/handler_end_session.go` and
      `handler_end_session_test.go` (references updated to match).
    - The actual cookie name a browser sees didn't read as related to
      `x_end_session_ref`, the query param it pairs with. Renamed so the two
      are recognizable together: `end_session_ref` (query) /
      `end_session_ref_key` (cookie holding the key that decrypts it). The
      `end_session_stash.go` filename, `sealEndSessionRequest`/
      `resumeFromStash`/`ErrEndSessionStashInvalid` and other "stash"-named
      identifiers for the round-trip mechanism itself are unchanged — only
      the cookie's identifier and on-wire name changed.
    - Verified: unit tests (`go test ./pkg/lib/oauth/oidc/handler/...`) and
      the full e2e run (`cd e2e && ./run.sh setup`, `go test
      ./pkg/testrunner/ -count 1 -v -timeout 10m -run
      "TestAuthflow/oidc/end_session_id_token_hint"`, then
      `./run.sh teardown`) both pass.

18. **`Match id_token_hint on its own creating session, not SSO group alone`**
    - Files: `pkg/lib/oauth/grant_offline.go` (new
      `OfflineGrant.IsCreatedFromSession(ss session.SessionBase) bool`,
      alongside `IsSameSSOGroup`), `pkg/lib/oauth/oidc/handler/handler_end_session.go`
      (new `matchesCurrentSession` helper combining `IsSameSSOGroup` and
      `IsCreatedFromSession`, replacing the bare `sidSession.IsSameSSOGroup(s)`
      check in `Handle`), `pkg/lib/oauth/oidc/handler/handler_end_session_test.go`
      (the "not SSO enabled" case rewritten to expect silent logout; a new
      case added for the genuinely-unrelated-grant-and-not-SSO-enabled
      combination — see §12 cases 5–6).
    - Fixes a real bug found via a live HAR trace: an ordinary OIDC client
      that never sends `x_sso_enabled` (the common case — that's an
      Authgear-specific extension) still authenticates through a real IDP
      session, but its offline grant's `SSOEnabled` defaults to `false`, so
      `IsSameSSOGroup` alone always denied it the silent-logout fast path,
      even though `IDPSessionID` demonstrably named that exact session.
      Deliberately scoped to this one decision — `session.Manager.invalidate`'s
      revocation cascade is untouched, so such a grant still isn't revoked
      when its own session logs out this way. See §4/§14 for the full
      reasoning on why widening `IsSameSSOGroup` itself was rejected.
    - Verified: unit tests (`go test ./pkg/lib/oauth/... ./pkg/lib/oauth/oidc/handler/...`).

19. **`Add e2e coverage for a client that never sends x_sso_enabled`**
    - Files: `e2e/pkg/e2eclient/client.go` (`SetupOAuthOptions.SSOEnabledOmitted`),
      `e2e/pkg/testrunner/models.go`/`testcase.go` (`oauth_setup_sso_enabled_omitted`),
      `e2e/tests/oidc/end_session_id_token_hint.test.yaml` (new case 2 —
      see §13.3).
    - `SetupOAuth` always sent `x_sso_enabled` explicitly before this, so
      there was no way to construct the exact scenario that motivated commit
      18. Verifies both halves of that fix's scope: the browser session is
      logged out silently, but the client's own offline grant/refresh token
      is not revoked by that logout (`session.Manager.invalidate`'s cascade
      is unchanged).
    - Verified against a live server: `cd e2e && ./run.sh setup` then
      `go test ./pkg/testrunner/ -count 1 -v -timeout 10m -run
      "TestAuthflow/oidc/end_session_id_token_hint"` (all 9 cases pass) and
      the full suite (`go test ./pkg/testrunner/ -count 1 -timeout 10m`),
      then `./run.sh teardown`.

20. **Plan doc update** (this document): §4, §5, §12, §13.1–13.3, §14, §16
    updated to describe the `matchesCurrentSession`/`IsCreatedFromSession`
    fix and its new e2e coverage.

21. **`spec: Document id_token_hint's session-matching rules for RP-Initiated Logout`**
    - Files: `docs/specs/oidc.md` (`### id_token_hint` under
      `## RP-Initiated Logout`).
    - The spec's original wording ("its `sid` matches the current logged in
      IdP session") predated both the SSO-group redesign (§4) and the
      `IsCreatedFromSession` fix (commit 18); it no longer described what
      the endpoint actually does. Rewritten to spell out the three matching
      cases (same session directly, created-from-this-session regardless of
      `x_sso_enabled`, or same SSO group), why the second case exists (most
      clients never send `x_sso_enabled`), and that a match through the
      second case alone does not get the grant itself revoked — only the
      IdP session.
    - **Superseded by commit 24 below** once the design changed again: the
      grant *is* now revoked too, and the "three cases" collapsed to two.

22. **`Revert IsCreatedFromSession/matchesCurrentSession in favor of fixing IsSameSSOGroup directly`**
    - Files: `pkg/lib/oauth/grant_offline.go` (removes
      `IsCreatedFromSession`; changes `IsSameSSOGroup` itself — see §4's
      "Final design"), `pkg/lib/oauth/oidc/handler/handler_end_session.go`
      (removes `matchesCurrentSession`; `Handle`'s Step 4 goes back to the
      bare `sidSession.IsSameSSOGroup(s)` check),
      `pkg/lib/oauth/oidc/handler/handler_end_session_test.go` (the "same
      login but not SSO enabled" case reverts to a single `Logout` call,
      since `invalidate`'s own cascade now covers the grant automatically).
    - Prompted by two further rounds of user feedback after commit 18
      landed: first, that the underlying grant should actually be revoked
      too (not just the browser session), which commit 18's design
      deliberately did *not* do; then, having considered deriving
      `SSOEnabled` from `IDPSessionID` at token-issuance time instead (a
      valid alternative, discussed and rejected — see §4) in favor of the
      simplest option: fix `IsSameSSOGroup` itself, and let its two existing
      consumers (`session.Manager.invalidate`, `pkg/lib/sessionlisting`)
      both benefit for free.
    - Must land together: reverting the handler and grant_offline.go changes
      independently would leave `Handle` calling a helper that no longer
      exists.
    - Verified: unit tests (`go test ./pkg/lib/oauth/... ./pkg/lib/oauth/oidc/handler/...`)
      and a live e2e run (all 9 `end_session_id_token_hint` cases plus the
      full suite) both pass; confirmed the earlier
      `verify_offline_grant_not_revoked` (`200`) e2e assertion now correctly
      needs to flip to `invalid_grant` (`400`) — see commit 23.

23. **`Update e2e test for the grant now being revoked too`**
    - Files: `e2e/tests/oidc/end_session_id_token_hint.test.yaml` (case 2:
      renamed, comments updated, `verify_offline_grant_not_revoked` (`200`)
      → `verify_offline_grant_revoked` (`400`/`invalid_grant`)).
    - Depends on commit 22. Verified live: `cd e2e && ./run.sh setup`, `go
      test ./pkg/testrunner/ -count 1 -v -timeout 10m -run
      "TestAuthflow/oidc/end_session_id_token_hint"` (all 9 cases pass) and
      the full suite, then `./run.sh teardown`.

24. **`Update sessionlisting_test.go for the IsSameSSOGroup fix`**
    - Files: `pkg/lib/sessionlisting/listing_test.go`.
    - Running the full `pkg/...` test suite after commit 22 surfaced a real,
      concrete consequence: `listing.go` reuses `IsSameSSOGroup` to decide
      whether an offline-grant row on the settings page should be flagged
      as the end-user's current device. An existing fixture
      (`makeOfflineGrant("3", ..., idpSession.ID, "spa-client-id", false)`)
      constructed exactly the "`IDPSessionID` matches, `SSOEnabled: false`"
      combination this fix targets, and asserted it should *not* be flagged
      current — confirmed with the user that, per the same reasoning as the
      rest of this fix, it now correctly *should* be, and updated the two
      affected `Convey` blocks' expected output accordingly. The other two
      `Convey` blocks in the same group (checking the reverse direction —
      an `IDPSession`'s own `IsSameSSOGroup`, unchanged) needed no changes,
      confirming the fix's intended asymmetry.
    - Verified: `go test ./pkg/lib/sessionlisting/...` and the full `go test
      ./pkg/...` suite, both green.

25. **`spec: Simplify id_token_hint matching rules and document the IsSameSSOGroup fix`**
    - Files: `docs/specs/oidc.md` (`### id_token_hint`: collapses the three
      matching cases from commit 21 into two, since — for this endpoint,
      always comparing against the current IdP session — the "created from
      this session" and "same SSO group" cases turned out to be identical;
      removes the now-inaccurate "grant survives" caveat), `docs/specs/oidc-sso-browser.md`
      (`Logout` → `Revoke IdP session`: now revokes any refresh token whose
      `idp_session_id` names the session, regardless of that token's own
      `sso_enabled`, not only `sso_enabled=true` ones; `Session listing` →
      adds a note that an `sso_enabled=false` grant can still be flagged as
      the current device while remaining its own separate, uncombined
      entry).

26. **Plan doc update** (this document): §4, §5, §12, §13.1–13.3, §14, §16
    updated again to describe the final `IsSameSSOGroup` fix, replacing the
    superseded `matchesCurrentSession`/`IsCreatedFromSession` design,
    including the `sessionlisting` consequence and the spec re-simplification.
