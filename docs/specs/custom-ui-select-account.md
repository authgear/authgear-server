# Custom UI: Select Account

- [Overview](#overview)
- [Goals](#goals)
- [Use Cases](#use-cases)
  - [UC1. Returning to a second app under the same project (browser SSO continuation)](#uc1-returning-to-a-second-app-under-the-same-project-browser-sso-continuation)
  - [UC2. Signing up for a new account while a different session is active](#uc2-signing-up-for-a-new-account-while-a-different-session-is-active)
  - [UC3. Step-up assurance before reusing a session](#uc3-step-up-assurance-before-reusing-a-session)
  - [UC4. Enforcing a single-active-session policy on every login, including continuation](#uc4-enforcing-a-single-active-session-policy-on-every-login-including-continuation)
- [Design](#design)
  - [Config changes](#config-changes)
  - [HTTP API changes](#http-api-changes)
  - [CORS](#cors)
  - [Session and account resolution](#session-and-account-resolution)
  - [Completing identification with the existing session](#completing-identification-with-the-existing-session)
- [End-to-end sequence](#end-to-end-sequence)
- [Security analysis](#security-analysis)
- [Backward compatibility](#backward-compatibility)
- [Edge cases](#edge-cases)
---

## Overview

This document specifies how a Custom UI presents an account-selection screen when the end-user already has an active Authgear session, letting them continue as that account without re-entering credentials — implemented as a **new identification option, `select_account`**, inside the `identify` step of the `login` and `signup_login` flows (not `signup`, `reauth`, or `account_recovery` — see [Edge cases](#edge-cases)). Like `oauth`/`passkey`, a project can configure it to complete a login without a further `authenticate` step — not because the engine treats any of these specially, but because nothing forces a flow to route a `one_of` entry into an `authenticate` step it wasn't given (see [Config changes](#config-changes)).

Continuation happens through the same `POST /api/v1/authentication_flows` / `.../states/input` calls a Custom UI already makes — no new token, no separate endpoint, no dedicated "decline" input (declining is just choosing a different option).

This works because a credentialed `fetch()` from a Custom UI hosted same-site with Authgear is a same-site request: `SameSite=Lax`/`Strict` only block *cross-site* requests, so the session cookie is sent as long as the browser includes credentials and the server reflects CORS for that origin. A cross-site Custom UI gets none of this — its origin is never CORS-allow-listed, so `select_account` simply never appears.

---

## Goals

- Allow a Custom UI hosted same-site with Authgear to detect that the end-user has an existing Authgear session.
- Allow the Custom UI to show user account information (display name) for the account-selection screen.
- Allow the end-user to continue with the existing session without re-entering credentials.
- Preserve security: an attacker who captures flow state must not be able to complete authentication on behalf of the victim.
- Do this using the existing Authentication Flow API surface — no new token store, no new endpoints, no new step type, no dedicated "decline" input.

---

## Use Cases

### UC1. Returning to a second app under the same project (browser SSO continuation)

Two OAuth clients, `client_a` and `client_b`, belong to the same Authgear project and each have a same-site Custom UI (`ui-a.example.com`, `ui-b.example.com`, both under the same registrable domain as Authgear, `auth.example.com`). The end-user has already signed in through `client_a`'s Custom UI. They now open `client_b`'s app for the first time in this browser. Because the Authgear session cookie is shared across the whole registrable domain, `client_b`'s Custom UI can offer to continue as the same account without the end-user re-entering credentials.

**Required configuration** — `client_b` needs `x_custom_ui_uri` set, and its `login` flow needs `select_account` added to `identify`:

```yaml
oauth:
  clients:
  - client_id: client_b
    x_custom_ui_uri: "https://ui-b.example.com/auth"

authentication_flow:
  login_flows:
  - name: default
    steps:
    - type: identify
      one_of:
      - identification: select_account
      - identification: email
        steps:
        - type: authenticate
          one_of:
          - authentication: primary_password
```

**Step 1 — App B redirects the browser to Authgear**

```
GET /oauth2/authorize?client_id=client_b&redirect_uri=https://app-b.example.com/callback&response_type=code&code_challenge=...&code_challenge_method=S256 HTTP/1.1
Host: auth.example.com
Cookie: session=... (still valid from signing in through client_a earlier)
```

**Step 2 — Authgear redirects to `client_b`'s Custom UI**

```
HTTP/1.1 302 Found
Location: https://ui-b.example.com/auth?x_ref=oauthsession_abc123&client_id=client_b&redirect_uri=https://app-b.example.com/callback
```

**Step 3 — Custom UI creates a `login` flow, forwarding `x_ref`**

```
POST /api/v1/authentication_flows HTTP/1.1
Host: auth.example.com
Origin: https://ui-b.example.com
Content-Type: application/json

{ "type": "login", "name": "default", "url_query": "client_id=client_b&x_ref=oauthsession_abc123" }
```

Because the request is same-site with credentials included, the session cookie is read and `select_account` appears:

```json
{
  "result": {
    "state_token": "authflowstate_xyz",
    "type": "login",
    "name": "default",
    "action": {
      "type": "identify",
      "data": {
        "type": "identification_data",
        "options": [
          { "identification": "select_account", "display_name": "user@example.com" },
          { "identification": "email" }
        ]
      }
    }
  }
}
```

**Step 4 — End-user clicks "Continue as user@example.com"**

```
POST /api/v1/authentication_flows/states/input HTTP/1.1
Host: auth.example.com
Origin: https://ui-b.example.com
Content-Type: application/json

{ "state_token": "authflowstate_xyz", "input": { "identification": "select_account", "index": 0 } }
```

```json
{
  "result": {
    "state_token": "authflowstate_xyz2",
    "type": "login",
    "name": "default",
    "action": {
      "type": "finished",
      "data": { "finish_redirect_uri": "https://auth.example.com/oauth2/consent?..." }
    }
  }
}
```

**Step 5 — Custom UI navigates to `finish_redirect_uri`; Authgear resolves consent and redirects to the app**

```
HTTP/1.1 302 Found
Location: https://app-b.example.com/callback?code=authcode_xyz&state=...
```

**Step 6 — App B exchanges the code for tokens**

```
POST /oauth2/token HTTP/1.1
Host: auth.example.com
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code&code=authcode_xyz&code_verifier=...&client_id=client_b&redirect_uri=https://app-b.example.com/callback
```

> No credentials were re-entered anywhere in this sequence — the end-user only clicked "Continue as user@example.com".

---

### UC2. Signing up for a new account while a different session is active

A project uses a single combined "Continue" entry point (`signup_login`) rather than separate sign-in/sign-up screens. An end-user who is already signed in as `existing@example.com` opens the app in the same browser and wants to register a *second*, unrelated account rather than continue as the one they're signed in as.

**Required configuration** — `select_account` declares `login_flow` only; `email` declares both, as usual for `signup_login`:

```yaml
authentication_flow:
  signup_login_flows:
  - name: default
    steps:
    - type: identify
      one_of:
      - identification: select_account
        login_flow: default_login
      - identification: email
        signup_flow: default_signup
        login_flow: default_login
```

**Step 1 — Custom UI creates the `signup_login` flow**

```
POST /api/v1/authentication_flows HTTP/1.1
Host: auth.example.com
Origin: https://ui.example.com

{ "type": "signup_login", "name": "default", "url_query": "client_id=client_a&x_ref=oauthsession_def456" }
```

```json
{
  "result": {
    "state_token": "authflowstate_abc",
    "type": "signup_login",
    "name": "default",
    "action": {
      "type": "identify",
      "data": {
        "type": "identification_data",
        "options": [
          { "identification": "select_account", "display_name": "existing@example.com" },
          { "identification": "email" }
        ]
      }
    }
  }
}
```

**Step 2 — End-user ignores "Continue as existing@example.com" and enters a new, not-yet-registered email instead**

```
POST /api/v1/authentication_flows/states/input HTTP/1.1
Host: auth.example.com

{ "state_token": "authflowstate_abc", "input": { "identification": "email", "login_id": "new-account@example.com" } }
```

Since `new-account@example.com` has no existing identity, the server switches this `signup_login` flow into `default_signup` — the response now reflects the signup flow's own next step (e.g. `create_authenticator`), with `result.type` changed to `"signup"`:

```json
{
  "result": {
    "state_token": "authflowstate_ghi",
    "type": "signup",
    "name": "default_signup",
    "action": { "type": "create_authenticator", "data": { "type": "create_authenticator_data", "options": [ { "authentication": "primary_password" } ] } }
  }
}
```

> No `select_account`-specific handling was needed here — declining is just choosing `email` like any other option, and the existing `signup_login` new-vs-existing resolution takes over from there.

---

### UC3. Step-up assurance before reusing a session

A higher-security project is comfortable letting end-users skip re-entering their password when their session is still valid, but wants a fresh TOTP code specifically when continuing via `select_account` — without imposing that extra step on brand-new logins, which already get 2FA enforced through their own `authenticate` step.

**Required configuration** — `select_account` gets its own nested `authenticate` step. There's deliberately no shared top-level `authenticate` step here: each `one_of` entry owns exactly the authentication it needs, nested as deep as necessary — `email` needs `primary_password` and then, nested one level further under that, `secondary_totp`; `select_account` needs only `secondary_totp`. This avoids the alternative of a shared step after `identify` that both `one_of` entries would fall through to, which would ask `select_account` for TOTP twice — once via its own nested step, once via the shared one:

```yaml
authentication_flow:
  login_flows:
  - name: default
    steps:
    - type: identify
      one_of:
      - identification: select_account
        steps:
        - type: authenticate
          one_of:
          - authentication: secondary_totp
      - identification: email
        steps:
        - type: authenticate
          one_of:
          - authentication: primary_password
            steps:
            - type: authenticate
              one_of:
              - authentication: secondary_totp
```

**Step 1 — End-user picks `select_account`; response asks for a TOTP code instead of finishing immediately**

```
POST /api/v1/authentication_flows/states/input HTTP/1.1
Host: auth.example.com

{ "state_token": "authflowstate_xyz", "input": { "identification": "select_account", "index": 0 } }
```

```json
{
  "result": {
    "state_token": "authflowstate_xyz3",
    "type": "login",
    "name": "default",
    "action": {
      "type": "authenticate",
      "authentication": "secondary_totp",
      "data": { "type": "authentication_data", "options": [ { "authentication": "secondary_totp" } ] }
    }
  }
}
```

**Step 2 — End-user submits the TOTP code; flow completes**

```
POST /api/v1/authentication_flows/states/input HTTP/1.1
Host: auth.example.com

{ "state_token": "authflowstate_xyz3", "input": { "authentication": "secondary_totp", "code": "123456" } }
```

> A normal `email` login on this same flow config still needs both `primary_password` and, nested under it, `secondary_totp` — two steps. `select_account` needs only its own nested `secondary_totp` — one step, entered once — since the session already satisfies the primary factor and there is no shared step left for it to also pass through.

---

### UC4. Enforcing a single-active-session policy on every login, including continuation

A banking app enforces one active session per account: signing in anywhere terminates that account's other sessions, via a `terminate_other_sessions` step placed after `authenticate` in its login flow. This must hold regardless of how the login happened — a user continuing via `select_account` must trigger the same termination as one who just typed a password, not a quieter path that skips it.

**Required configuration** — this only matters for `signup_login`, since `select_account` there switches into the *named* login flow rather than completing directly, so that flow's later steps still run. `default_login` needs its own `select_account` `one_of` entry (with no nested `steps`) for the switch to land anywhere without asking for a password:

```yaml
authentication_flow:
  signup_login_flows:
  - name: default
    steps:
    - type: identify
      one_of:
      - identification: select_account
        login_flow: default_login
      - identification: email
        signup_flow: default_signup
        login_flow: default_login

  login_flows:
  - name: default_login
    steps:
    - type: identify
      one_of:
      - identification: select_account
      - identification: email
        steps:
        - type: authenticate
          one_of:
          - authentication: primary_password
    - type: terminate_other_sessions
```

**Step 1 — End-user picks `select_account` in the `signup_login` flow; this switches into `default_login`, replaying the input into its `select_account` entry — which has no nested `steps`, so `identify` completes immediately, landing directly on `terminate_other_sessions`**

```
POST /api/v1/authentication_flows/states/input HTTP/1.1
Host: auth.example.com

{ "state_token": "authflowstate_abc", "input": { "identification": "select_account", "index": 0 } }
```

```json
{
  "result": {
    "state_token": "authflowstate_jkl",
    "type": "signup_login",
    "name": "default",
    "action": { "type": "terminate_other_sessions", "data": {} }
  }
}
```

> The same `terminate_other_sessions` step would appear at this point for a normal `email`+`primary_password` login too — continuation via `select_account` doesn't skip it.

---

## Design

### Config changes

Add `select_account` as a new allowed `identification` value inside an `identify` step's `one_of`, in both the `login` and `signup_login` flows. Not added to `signup` (see [Edge cases](#edge-cases)).

What happens after it's chosen is controlled by that entry's own optional nested `steps` — same as any other `identification` option. Omit `steps` for immediate completion. There's deliberately no shared top-level `authenticate` step here: any step placed after `identify` is reached regardless of *which* `one_of` entry was chosen, `oauth`/`select_account` included, and whether it then prompts the user depends only on whether they have a matching enrolled authenticator for it — not on how they identified. Giving `email` its own nested `authenticate` step instead avoids routing `oauth`/`select_account` into one they don't need:

```yaml
authentication_flow:
  login_flows:
  - name: default
    steps:
    - type: identify
      one_of:
      - identification: select_account
      - identification: email
        steps:
        - type: authenticate
          one_of:
          - authentication: primary_password
      - identification: oauth
        alias: google
```

Or require step-up 2FA specifically on this entry:

```yaml
      - identification: select_account
        steps:
        - type: authenticate
          one_of:
          - authentication: secondary_totp
```

**`signup_login`:** `select_account` declares `login_flow` only, never `signup_flow` — it can only ever continue an existing login. Choosing it switches into the named `login_flow`, replaying the same `identify` input into that flow — exactly like `email`/`oauth` replay their `identify` input into whichever flow they switch into. **The referenced `login_flow` must itself declare a matching `select_account` `one_of` entry** for this to go anywhere; the switch only replays the `identify` input, not proof of authentication — whether the user is then asked to authenticate again is entirely up to that target entry's own nested `steps`, the same rule as everywhere else in this document:

```yaml
authentication_flow:
  signup_login_flows:
  - name: default
    steps:
    - type: identify
      one_of:
      - identification: select_account
        login_flow: default_login
      - identification: email
        signup_flow: default_signup
        login_flow: default_login

  login_flows:
  - name: default_login
    steps:
    - type: identify
      one_of:
      - identification: select_account
      - identification: email
        steps:
        - type: authenticate
          one_of:
          - authentication: primary_password
```

Here, `default_login`'s own `select_account` entry has no nested `steps`, so the switched-into flow completes immediately — no different from reaching that same entry by creating `default_login` directly. Anything configured after `identify` in `default_login` (e.g. `terminate_other_sessions`) still runs normally, same as for any other entry.

---

### HTTP API changes

No new endpoints, step types, or `action.type`. The `identify` step's response gains a new possible `options` entry:

```json
{
  "result": {
    "state_token": "authflowstate_xyz",
    "type": "login",
    "name": "default",
    "action": {
      "type": "identify",
      "data": {
        "type": "identification_data",
        "options": [
          { "identification": "select_account", "display_name": "user@example.com" },
          { "identification": "email" }
        ]
      }
    }
  }
}
```

`display_name` is the only `select_account`-specific field. No user identifier is included — the server resolves identity internally from the input's `index` (below). Absent when there's no eligible session — the response then looks exactly as it did before this feature.

Unlike `masked_display_name` elsewhere in this API, `display_name` here is returned **unmasked**: it identifies the account already bound to the caller's own session cookie, not an as-yet-unauthenticated identity, so there's nothing to mask.

Selection is by `index` in the *input* (an existing pattern in this API, also used by e.g. `primary_oob_otp_email`'s options) — the option itself carries no index:

```json
{ "state_token": "authflowstate_xyz", "input": { "identification": "select_account", "index": 0 } }
```

`index` is this entry's position in the full `options` array. There's at most one `select_account` entry today; the field exists so the shape doesn't need to change if multiple concurrent accounts are supported later. `select_account`'s position in `options` follows its position in the `one_of` config, same as every other entry, and is stable across calls — a Custom UI can rely on it not moving around between requests to the same flow config.

Declining needs no special input — submit any other option's input instead (e.g. `{"identification": "email", "login_id": "..."}`).

---

### CORS

For the session cookie to reach `POST /api/v1/authentication_flows`/`.../states/input` on a cross-origin `fetch()`, responses must carry:

```
Access-Control-Allow-Origin: https://ui.example.com
Access-Control-Allow-Credentials: true
```

reflecting the caller's origin (never a wildcard), whenever it matches **any** OAuth client's `x_custom_ui_uri` in the project — the check is "is this origin a known Custom UI at all", not "is this origin the Custom UI of the specific `client_id` this request happens to target". Matching against a project-wide set rather than a single client is deliberate: origin and `client_id` are independent pieces of information (an origin either is or isn't a registered Custom UI, regardless of which client's flow it's about to call), and this project already checks it that way for every other endpoint using this same allow-list.

This must also apply to the `OPTIONS` preflight `fetch()` triggers first, which despite having no body isn't a problem: a preflight always carries the `Origin` header, and matching only ever depends on that header against the allow-list above — never on `client_id` or anything else the body would carry. This project already has this infrastructure, including for preflight, on other endpoints.

---

### Session and account resolution

The `select_account` option is derived entirely from the request's session cookie — nothing client-supplied influences it. It's omitted, using the same rules as the built-in Auth UI's existing account-selection screen, when:

- No session is present.
- The session was established with "do not persist" semantics (`x_suppress_idp_session_cookie`).
- `prompt=login` is present.
- `login_hint` is present and identifies a different user than the session.

(`prompt=none` is decided earlier, before any flow exists — not applicable here.)

---

### Completing identification with the existing session

- **`login`:** whether anything further is asked is controlled by `select_account`'s own nested `steps` (see [Config changes](#config-changes)) — omitted, the flow completes immediately; with a nested `authenticate` step, that must be satisfied first.
- **`signup_login`:** switches into the declared `login_flow`, replaying the same `identify` input into that flow — the switch itself carries no proof of authentication, only the input. Whether anything further is asked is controlled by the *target* flow's own `select_account` `one_of` entry and its nested `steps`, exactly as in the `login` case above; any steps configured after `identify` in that flow still run regardless.

Before completing, two checks must pass:

- `index` must be in bounds and point to a `select_account` entry — enforced the same way as any other option's `index`, by standard input schema validation, no `select_account`-specific error.
- The session cookie must still resolve to the same user recorded when the option was computed — guards against the session changing between the two calls — else:
  ```json
  { "error": { "name": "Unauthorized", "reason": "SelectAccountSessionChanged", "message": "session no longer matches the selected account", "code": 401 } }
  ```

The existing session itself is reused as-is — not rotated or renewed — matching the built-in Auth UI's current account-selection screen.

---

## End-to-end sequence

This shows the whole OAuth lifecycle a Custom UI sits inside, not just the `identify` exchange — including where the Custom UI's flow-creation call fits between `/oauth2/authorize` and the final redirect back to the app.

One constraint worth calling out before the diagram: the Custom UI's **first** flow-creation call of this sequence must be `type: "login"` or `type: "signup_login"`, never `type: "signup"` directly — since `select_account` only exists in those two flows, that's the only way to learn whether a session is eligible. A combined-entry-point Custom UI already does this by default; one with separate "Sign in"/"Sign up" screens must route its first call through `login`/`signup_login` regardless of which screen the user is on, only creating a genuinely separate `signup` flow afterward if the user has no eligible session (or declines it) and wants a new account.

```
App
  │
  ├─▶ GET /oauth2/authorize?client_id=...&redirect_uri=...&code_challenge=...
  │       Authgear sees x_custom_ui_uri configured for this client
  ├─◀ 302 → https://ui.example.com/auth?x_ref=oauthsession_abc&client_id=...&redirect_uri=...
  │
Custom UI (same-site with Authgear)
  │
  ├─▶ POST /api/v1/authentication_flows   (credentials: 'include')
  │       { type: "login", name: "default", url_query: "client_id=...&x_ref=oauthsession_abc" }
  ├─◀ 200 { state_token, action: { type: "identify", data: { options: [ ... ] } } }
  │
  ├── Case A: a `select_account` entry is present
  │     │   [Display "Continue as user@example.com" alongside the usual options]
  │     │
  │     └─▶ User continues:
  │           POST /api/v1/authentication_flows/states/input
  │             { state_token, input: { identification: "select_account", index: 0 } }
  │
  │       (declining instead: submit a normal identify input — in `signup_login` the
  │        server resolves it into `signup`/`login` as usual; in `login`, the Custom UI
  │        creates a separate `signup` flow if the user wants a brand new account)
  │
  └── Case B: no `select_account` entry (no eligible session) — response is
        unchanged from before this feature; Custom UI proceeds as today
  │
  ├─◀ 200 { state_token, action: { type: "finished", data: { finish_redirect_uri: "https://auth.example.com/oauth2/consent?..." } } }
  │
  ├─▶ Top-level browser navigation (NOT a fetch call) to `finish_redirect_uri`,
  │       i.e. GET /oauth2/consent?... — this hands control back to Authgear
  │
Authgear
  │
  ├─▶ /oauth2/consent — may redirect immediately, or require the user to approve first
  ├─◀ 302 → https://app.example.com/callback?code=authcode_xyz&state=...
  │
App
  │
  └─▶ POST /oauth2/token  (code + code_verifier) → access/refresh tokens
```

---

## Security analysis

| Threat | Mitigation |
|---|---|
| Forged `POST .../states/input` from an attacker's page, using a captured `state_token` | Attacker's origin isn't CORS-allow-listed, so its `fetch()` can't carry the session cookie and can't read a response either way. |
| Compromised sibling subdomain (same registrable domain) | Out of scope — already able to impersonate the Custom UI generally. |
| A Custom UI origin registered for one client calling a different client's flow | Intentional, not a gap: the allow-list check is "is this origin *some* project client's registered Custom UI?", not tied to the `client_id` in the request (see [CORS](#cors)) — and the caller is still bounded by the same session-cookie validation as every other case. |
| Session changes between the option being shown and the input being submitted | Bounds + user-match check on submission (see [Completing identification](#completing-identification-with-the-existing-session)). |
| Forged `select_account` data | Derived entirely server-side from the resolved session. |

---

## Backward compatibility

`select_account` is a new value; an unmodified `login`/`signup_login` `identify` config doesn't list it, so nothing changes for those projects.

Opting in requires updating the Custom UI to recognize the new `identification` value, same as adding any other new identification method — plus, for a Custom UI with separate sign-in/sign-up screens, routing its first call through `login`/`signup_login` rather than `signup` (see [End-to-end sequence](#end-to-end-sequence)) — a real behavior change, not just a config one.

---

## Edge cases

- **`prompt=login`**: option omitted.
- **`prompt=none`**: not applicable — decided before any flow exists.
- **`login_hint` present**: option omitted, regardless of match — the caller already knows which identity it wants.
- **Multiple active accounts**: not supported; at most one entry today.
- **`signup`**: not added — completing it via `select_account` would create no new user, contradicting what `type: "signup"` represents.
- **`reauth`**: not in scope — targets a specific known user, no account choice to make.
- **`account_recovery`**: not in scope — exists for when the user has no usable session at all.
