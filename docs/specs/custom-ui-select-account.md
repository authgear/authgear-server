# Custom UI: Select Account

- [Overview](#overview)
- [Background](#background)
- [Goals](#goals)
- [Design](#design)
  - [Security principle](#security-principle)
  - [Phase 1: Authorization endpoint — detect existing session](#phase-1-authorization-endpoint--detect-existing-session)
  - [Phase 2: Custom UI — display the account picker](#phase-2-custom-ui--display-the-account-picker)
  - [Phase 3a: User continues with existing account](#phase-3a-user-continues-with-existing-account)
  - [Phase 3b: User switches to a different account](#phase-3b-user-switches-to-a-different-account)
- [HTTP API](#http-api)
  - [GET /api/v1/accounts_hint/{token}](#get-apiv1accounts_hinttoken)
  - [GET /api/v1/select_account](#get-apiv1select_account)
- [End-to-end flow diagram](#end-to-end-flow-diagram)
- [Edge cases](#edge-cases)
- [Security analysis](#security-analysis)
- [Backward compatibility](#backward-compatibility)
- [Reference: x_ref](#reference-x_ref)
---

## Overview

This document specifies how a Custom UI can present an account-selection screen when the end-user already has an active Authgear session, allowing them to continue as their existing account without re-entering credentials.

---

## Background

When the built-in Auth UI handles a returning user, it routes the browser through `/authflow/v2/select_account`. This handler runs on Authgear's own domain, reads the session cookie, and—if the user clicks "Continue"—completes the OAuth authorization flow without creating an authentication flow at all.

Custom UI is hosted on a different domain and communicates with Authgear via the [Authentication Flow HTTP API](./authentication-flow-api-reference.md). Two constraints make a direct port of the built-in behavior impossible:

1. **Cross-domain cookies**: API calls from the Custom UI are cross-origin; the browser does not send Authgear's session cookie with them. The Custom UI cannot detect an existing session by calling the API.
2. **Backward compatibility**: The Authentication Flow API must not inject new action types into existing flows, as that would break Custom UI implementations that do not know how to handle them.

---

## Goals

- Allow a Custom UI to detect that the end-user has an existing Authgear session.
- Allow the Custom UI to show user account information (display name) for the account-selection screen.
- Allow the end-user to continue with the existing session without re-entering credentials.
- Preserve security: an attacker who captures the redirect URL must not be able to complete authentication on behalf of the victim.
- Keep existing Custom UI integrations working without modification.

---

## Design

### Security principle

The session cookie is the proof of identity for the "continue" path. It can only be read during a same-origin browser navigation to Authgear's domain. Therefore:

> **The account continuation step MUST be a browser navigation to Authgear, not a JSON API call from the Custom UI.**

This mirrors exactly what the built-in select account handler does today.

---

### Phase 1: Authorization endpoint — detect existing session

At `GET /oauth2/authorize`, when **all** of the following conditions hold:

1. The requesting OAuth client has `x_custom_ui_uri` configured.
2. A valid session exists in the browser (readable via cookie at this same-origin navigation).
3. The request does not include `prompt=login`.
4. The request does not include `prompt=none`.
5. `login_hint` is not present in the authorization request.

Authgear MUST:

1. Enumerate all logged-in accounts (via session cookie). Record the ordered list—`[{user_id, display_name}, …]`—associated with `x_ref` (server-side). The order is stable and defines the `x_account_index` used at continuation.
2. Generate a random, cryptographically secure **accounts hint token** (32 bytes, URL-safe base64-encoded).
3. Store the token with a TTL of **10 minutes**, associated with:
   - The same ordered list of eligible accounts: `[{user_id, display_name}, …]`
   - `x_ref` (to prevent use across different authorization requests)
4. Append `x_accounts_hint=<token>` to the Custom UI redirect URL.

The token MUST NOT contain any PII or user-identifiable information. It is an opaque random identifier only.

**Example redirect to Custom UI:**

```
https://custom.example.com/auth?x_ref=oauthsession_abc123&client_id=my_app&redirect_uri=https%3A%2F%2Fapp.example.com%2Fcallback&x_accounts_hint=Rn4xT7...
```

---

### Phase 2: Custom UI — display the account picker

When the Custom UI receives `x_accounts_hint` in its URL parameters, it MUST call `GET /api/v1/accounts_hint/{x_accounts_hint}` to retrieve account display names, then present an account-selection screen showing the logged-in accounts. See [HTTP API](#get-apiv1accounts_hinttoken) for the response format.

If `x_accounts_hint` is absent from the Custom UI URL, the Custom UI MUST proceed with a normal authentication flow as if no existing session exists (see [Phase 3b](#phase-3b-user-switches-to-a-different-account)).

---

### Phase 3a: User continues with existing account

When the user selects an existing account, the Custom UI MUST perform a **top-level GET redirect** to Authgear's account continuation endpoint:

```js
window.location.href =
  authgearEndpoint + '/api/v1/select_account'
  + '?x_ref=' + encodeURIComponent(xRef)
  + '&x_account_index=' + selectedAccountIndex;
```

A GET redirect is required for a specific reason: the session cookie has `SameSite=Lax`. Under this policy, the browser sends the cookie on cross-site requests **only** when they are top-level navigations using a safe method (GET/HEAD). A cross-site form POST would not include the cookie, so Authgear would never see the session. A GET redirect satisfies both requirements and makes the cookie available to Authgear on this same-origin request.

The `x_account_index` parameter is the 0-based position of the selected account in the array returned by `GET /api/v1/accounts_hint/{token}`. If omitted, it defaults to `0`.

Using an index rather than a user ID ensures that no user identifier appears in the URL.

Authgear then:

1. Reads the `x_account_index` query parameter (default: `0`).
2. Reads the `x_ref` query parameter.
3. Looks up the OAuth session by `x_ref` and retrieves the stored eligible accounts list.
4. Validates that `x_account_index` is within the bounds of the eligible accounts list. If not, respond with an error and abort.
5. Resolves `user_id = eligible_accounts[x_account_index].user_id` **server-side only**.
6. Reads the session cookie from the browser request.
7. Validates that the session cookie matches the resolved user. If not, redirect to the Custom UI with `error=account_changed`.
8. Completes the OAuth authorization using the existing session and resolves the final redirect URI.
9. Redirects the browser to `redirect_uri?code=…` (same as completing any authorization flow).

See [GET /api/v1/select_account](#get-apiv1select_account) for the full endpoint spec.

---

### Phase 3b: User switches to a different account

When the user chooses to sign in with a different account, the Custom UI creates a normal authentication flow:

```
POST /api/v1/authentication_flows
{
  "type": "login",
  "name": "default",
  "url_query": "client_id=...&x_ref=..."
}
```

This is identical to the current Custom UI flow. The `x_accounts_hint` is simply ignored. The user proceeds through `identify` → `authenticate` as normal.

---

## HTTP API

This feature introduces two new endpoints, both under `/api/v1/` (the namespace for Custom UI integration). They differ in how they must be called:

| Endpoint | Call method | Response type | Cookie required |
|---|---|---|---|
| `GET /api/v1/accounts_hint/{token}` | XHR / fetch (cross-origin) | JSON | No |
| `GET /api/v1/select_account` | Top-level browser navigation (`window.location.href`) | HTTP 302 redirect | Yes (session cookie) |

`/authflow/v2/` is the internal prefix used by Authgear's built-in Auth UI and is not part of the Custom UI integration API. Both custom UI endpoints are under `/api/v1/`.

---

### GET /api/v1/accounts_hint/{token}

Retrieves account display information for the accounts hint token. This is a read-only, unauthenticated endpoint. Its result is informational only and does not grant any authentication.

**Request:**

```
GET /api/v1/accounts_hint/Rn4xT7... HTTP/1.1
```

**Successful response (200):**

```json
{
  "result": {
    "accounts": [
      {
        "display_name": "user@example.com"
      },
      {
        "display_name": "another@example.com"
      }
    ]
  }
}
```

Each entry corresponds to one eligible account. The position in the array is the `x_account_index` the Custom UI MUST submit to the continuation endpoint. No user identifier is included in the response; the server resolves the identity internally from the index.

`display_name` is the primary identity display name of the account (email address, phone number, or username depending on the project configuration). It is returned unmasked.

**Token not found or expired (404):**

```json
{
  "error": {
    "name": "NotFound",
    "reason": "AccountsHintNotFound",
    "message": "account hint not found or expired",
    "code": 404
  }
}
```

When the Custom UI receives a 404, it MUST fall back to Phase 3b (normal authflow).

The token is NOT consumed by this endpoint. It may be called multiple times within the TTL. The token is invalidated once `GET /api/v1/select_account` completes the authorization successfully.

---

### GET /api/v1/select_account

Completes the OAuth authorization using the end-user's existing session. This is a browser-navigation endpoint (not a JSON API). It MUST be reached via a top-level GET redirect so that the browser includes the Authgear session cookie (`SameSite=Lax`). Do NOT call this via XHR or fetch — the browser will not send the cookie.

**Request:**

```
GET /api/v1/select_account?x_ref=oauthsession_abc123&x_account_index=0 HTTP/1.1
```

| Parameter | Required | Description |
|---|---|---|
| `x_ref` | Yes | The OAuth session ID passed to the Custom UI. |
| `x_account_index` | No | 0-based index of the selected account from the `GET /api/v1/accounts_hint/{token}` response. Defaults to `0`. |

**Validation:**

The server validates all of the following. If any check fails, the behavior depends on the nature of the failure:

| Failure | Behavior |
|---|---|
| `x_ref` is invalid or expired | Return HTTP 400 |
| `x_account_index` is out of bounds for the eligible accounts list | Return HTTP 400 |
| No session cookie present | Redirect to the Custom UI URL with `error=login_required` |
| Session cookie does not match the resolved user at `x_account_index` | Redirect to the Custom UI URL with `error=account_changed` |

**Error redirect format:**

When a session-related check fails, Authgear redirects the browser back to the Custom UI URL (the original `x_custom_ui_uri` with `x_ref` preserved), appending OAuth-style error parameters:

```
https://custom.example.com/auth?x_ref=...&error=login_required&error_description=No+active+session+found
```

| Error code | Meaning | Recommended Custom UI behavior |
|---|---|---|
| `login_required` | No active session found | Proceed with normal authflow (Phase 3b) |
| `account_changed` | Session exists but is for a different account than selected | Show a message that the session has changed, then proceed with normal authflow (Phase 3b) |

**Success:**

The server completes the OAuth authorization (identical to what the built-in select account handler does on "continue"), then issues a browser redirect to the app's `redirect_uri`:

```
HTTP/1.1 302 Found
Location: https://app.example.com/callback?code=authcode_xyz&state=...
```

The app then exchanges the `code` for tokens at `POST /oauth2/token` using its PKCE `code_verifier`, exactly as in any other authorization flow.

---

## End-to-end flow diagram

```
App
  │
  ├─▶ GET /oauth2/authorize?client_id=...&code_challenge=...
  │       Authgear reads session cookie ✓
  │       Stores eligible user_ids in OAuth session
  │       Generates x_accounts_hint (random opaque token)
  │       ↓
  ├─◀ 302 → https://custom.example.com?x_ref=...&x_accounts_hint=...
  │
Custom UI
  │
  ├─▶ GET /api/v1/accounts_hint/{x_accounts_hint}
  │       ↓
  ├─◀ { accounts: [{ display_name }, …] }
  │
  │   [Show "Continue as user@example.com / Use different account"]
  │
  │   User clicks "Continue" (selects account at index N)
  │       ↓
  ├─▶ GET /api/v1/select_account?x_ref=...&x_account_index=N
  │       (top-level GET redirect — SameSite=Lax cookie is sent ✓)
  │       Authgear reads session cookie ✓
  │       Resolves user_id = eligible_accounts[N].user_id (server-side)
  │       Validates cookie user == resolved user_id ✓
  │       Completes OAuth authorization
  │       ↓
  ├─◀ 302 → https://app.example.com/callback?code=...
  │
App
  │
  └─▶ POST /oauth2/token (code + code_verifier)
         ↓
      Access token + Refresh token
```

---

## Edge cases

### `x_accounts_hint` expires before the user acts

The token has a 10-minute TTL. If the display info call returns 404 (token expired before the Custom UI loaded), the Custom UI MUST fall back to Phase 3b (normal authflow).

If the Custom UI already fetched and cached the accounts list before expiry, it MAY still navigate to the account continuation endpoint — it does not require `x_accounts_hint` and is unaffected by its expiry. The only requirement for continuation is a valid session cookie.

### No session at continuation time

If the session expired or was revoked between authorization start and continuation, the cookie check fails. The server MUST redirect back to the Custom UI with `error=login_required`, preserving `x_ref` so the authflow can complete the same authorization request.

### `prompt=login`

When the authorization request includes `prompt=login`, Authgear MUST NOT generate `x_accounts_hint`. The user is required to re-authenticate. The Custom UI receives no account-selection signal.

### `prompt=none`

When the authorization request includes `prompt=none`, Authgear either completes authentication silently (if a valid session exists) or returns a `login_required` error — in neither case is the Custom UI involved, so `x_accounts_hint` is never generated.

### Multiple active accounts (Not implemented)

Multiple active accounts are not supported at this time. The eligible accounts list always contains exactly one entry. This section is included for future reference: when multiple accounts are supported, the eligible accounts list will contain one entry per active account, the Custom UI will display all accounts, and the user will select one by its `x_account_index`. The continuation endpoint will resolve the selected user server-side from the index and validate the session cookie against it.

### `login_hint` present

When the authorization request includes `login_hint`, it targets a specific user and `x_accounts_hint` MUST NOT be generated.

### CSRF

An attacker who captures the victim's `x_ref` can trick the victim's browser into navigating to `GET /api/v1/select_account?x_ref=<victim_x_ref>&x_account_index=0`. The session cookie check passes because the victim's browser carries the victim's cookie, so the victim is authenticated as themselves and the authorization code is issued to the registered `redirect_uri`.

The attacker gains nothing from this: the victim authenticates as themselves (not the attacker), and the code goes to a registered `redirect_uri` the attacker cannot observe. This is a force-login — a known weak property of OAuth redirect-based flows — not an account takeover.

No additional CSRF protection is required.

---

## Security analysis

| Threat | Mitigation |
|---|---|
| Attacker captures the Custom UI redirect URL (contains `x_ref` and `x_accounts_hint`) | Continuing requires the victim's session cookie in the attacker's browser. The attacker's browser does not have it. |
| Attacker calls the display info endpoint with a captured `x_accounts_hint` | Learns the account display name only (not credentials). The display name is not sufficient for authentication. Token TTL limits the exposure window. |
| Forged `x_accounts_hint` | The token is a cryptographically random server-generated value. An attacker cannot forge a valid token. |

---

## Backward compatibility

The `x_accounts_hint` parameter is additive. Custom UI implementations that do not recognize it simply ignore it. They receive `x_ref` and other existing parameters as before, create a normal authentication flow, and proceed through `identify` → `authenticate` unchanged.

The Authentication Flow API (`POST /api/v1/authentication_flows` and `POST /api/v1/authentication_flows/states/input`) is not modified. No new action types are added to existing flows.

---

## Reference: x_ref

`x_ref` is an opaque identifier for the pending OAuth authorization request. When the app initiates an authorization, Authgear redirects the browser to the Custom UI and appends `x_ref` as a query parameter. The Custom UI includes `x_ref` in all subsequent interactions with Authgear — when creating an authentication flow and when navigating back to Authgear on completion — so that Authgear can associate those interactions with the correct authorization request.

`x_ref` is not a new concept introduced by this spec; it is part of the existing Custom UI integration. This spec reuses it as a parameter to `GET /api/v1/select_account` for the same reason: to identify which authorization request the continuation belongs to.

---


