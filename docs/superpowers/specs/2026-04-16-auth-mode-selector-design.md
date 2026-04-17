# Design: Authentication Mode Selector for Traditional Web Applications

**Date:** 2026-04-16
**Status:** Implemented (PR #5663)
**Related:** DEV-3469 — Update application quickstart for NextJS app

---

## Overview

Traditional Web Application clients now support two distinct authentication patterns, introduced by the release of `authgear-sdk-nextjs`. The portal application detail page needs to let users declare which pattern they are using so the correct configuration settings are surfaced.

---

## Background

### Two auth patterns under one application type

| Pattern | How it works | Session lifetime controlled by |
|---|---|---|
| **Cookie Session** | Authgear sets cookies on the project domain | Cookie Session Settings (`/advanced/session`) |
| **Server-side SDK** | The SDK (e.g. `authgear-sdk-nextjs`) manages tokens server-side via encrypted HTTP-only cookies | Access token & refresh token lifetime fields on the application page |

---

## Design

### New UI element: Authentication Mode radio group

A clearly labelled **"Authentication Mode"** section is added to the application detail page for Traditional Web Application clients. It sits between **Basic Info** and **URIs**.

The section contains two radio options (Server-side SDK listed first as the recommended option):

**Option 1 — Server-side SDK** *(default for new apps)*
> The SDK manages tokens server-side. Session lifetime is set by access token & refresh token settings below. e.g. authgear-sdk-nextjs for Next.js.

**Option 2 — Cookie Session**
> Authgear manages cookies on your project domain. Session lifetime is configured in Cookie Session Settings.

---

### Section visibility rules

| Mode selected | Sections shown | Sections hidden |
|---|---|---|
| Server-side SDK | Refresh Token fields, Access Token fields | Cookie-based authentication section |
| Cookie Session | Cookie-based authentication (link to session settings) | Refresh Token fields, Access Token fields |

The URIs section, Basic Info, and Custom UI sections are unaffected by mode selection.

---

### Default value

- **Existing applications** (created before this feature ships): default to **Cookie Session**. No user action required; existing behaviour is preserved.
- **New applications** (created after this feature ships): default to **Server-side SDK**.

---

### Mode switching behaviour

When a user switches from **Server-side SDK → Cookie Session** within a single editing session (before saving), an inline note appears inside the Authentication Mode box:

> ℹ️ Token lifetime settings are not used in Cookie Session mode.

The note is visible only while the user has an unsaved mode change (SDK → Cookie Session) in the current session. It disappears if the user switches back to Server-side SDK. After saving with Cookie Session selected, the note does not appear on subsequent page loads.

The token lifetime values themselves are preserved in the database — they are simply hidden from the UI. Switching back to Server-side SDK (even in a future session) restores the fields with their previously saved values.

No confirmation dialog is shown — the action is fully reversible.

---

### Data model

Field on the OAuth client config: `x_traditional_webapp_session_type`

Values: `access_token` | `cookie`

This field is only meaningful for `application_type = traditional_webapp`. It has no effect on other application types.

The fallback for existing apps (no value stored) is `cookie` to preserve backward compatibility.

---

## Application creation wizard

### New wizard step: Select Authentication Mode

When the user selects **Traditional Web Application** in the creation wizard (`/configuration/apps/add`) and clicks Next/Save, a new step — **Select Authentication Mode** — is inserted before the final save.

**Wizard step order:**
- `traditional_webapp`: SelectType → **SelectAuthMode** (new) → Save
- `m2m`: SelectType → AuthorizeResource → Save
- All other types: SelectType → Save

### Step UI

The step presents the same two radio options as on the detail page, with Server-side SDK pre-selected:

**Server-side SDK** *(pre-selected)*
> The SDK manages tokens server-side. Session lifetime is set by access token & refresh token settings. e.g. authgear-sdk-nextjs for Next.js.

**Cookie Session**
> Authgear manages cookies on your project domain. Session lifetime is configured in Cookie Session Settings.

Buttons: **Save** (primary) | **Back** (navigates back to SelectType).

---

## Type selector description

The Traditional Web Application type description in the creation wizard reads:

> Websites with server-side rendering. Supports server-side SDK (e.g. Next.js) or cookie session.

---

## Quickstart guide

The quickstart guide (both the post-creation `?quickstart=true` screen and the sidebar widget) always shows two items for Traditional Web Application clients, regardless of auth mode:

1. **Next.js** — links to `https://docs.authgear.com/get-started/regular-web-app/nextjs`
2. **Other Framework** — links to `https://docs.authgear.com/get-started/start-building`

---

## Scope

**In scope:**
- Authentication Mode radio group on `EditOAuthClientForm` for `traditional_webapp` type (Server-side SDK first)
- Authentication Mode step in the application creation wizard for `traditional_webapp` type (Server-side SDK pre-selected)
- Conditional rendering of Cookie Session section vs. Token Lifetime sections based on selected mode
- Inline note on mode switch (SDK → Cookie Session)
- Default to Server-side SDK for new Traditional Web Application clients; Cookie Session preserved for existing apps
- New config field `x_traditional_webapp_session_type` to persist the mode selection
- Updated quickstart: always show Next.js + Other Framework items

**Out of scope:**
- Changes to other application types (SPA, Native, M2M, OIDC/SAML)
- Backend enforcement or validation of token settings based on mode

---

## Files affected

| File | Change |
|---|---|
| `pkg/lib/config/oauth.go` | Add `x_traditional_webapp_session_type` field (`cookie` \| `access_token`) to schema and Go struct |
| `portal/src/types.ts` | Add `x_traditional_webapp_session_type` to `OAuthClientConfig` TypeScript type |
| `portal/src/graphql/portal/EditOAuthClientForm.tsx` | Add Authentication Mode radio group; conditionally show/hide sections; Server-side SDK first |
| `portal/src/graphql/portal/CreateOAuthClientScreen.tsx` | Add `SelectAuthMode` step; default to `access_token`; Server-side SDK first in options |
| `portal/src/graphql/portal/EditOAuthClientScreen.tsx` | Quickstart always shows Next.js + Other Framework regardless of session type |
| `portal/src/locale-data/en.json` | Translation keys for auth mode selector, type description, and quickstart items |

---

## Acceptance criteria

- [x] Authentication Mode radio group appears on the application detail page for Traditional Web Application clients only
- [x] Server-side SDK is listed first in the radio group
- [x] Cookie Session mode: shows Cookie-based authentication section, hides token lifetime fields
- [x] Server-side SDK mode: shows Refresh Token and Access Token fields, hides Cookie-based authentication section
- [x] Existing apps default to Cookie Session with no user action required
- [x] New apps default to Server-side SDK
- [x] Switching from SDK → Cookie Session shows the inline note; switching back removes it
- [x] Token lifetime values are preserved (not deleted) when switching to Cookie Session mode
- [x] Authentication Mode step added to creation wizard; appears only when Traditional Web Application is selected
- [x] Server-side SDK is pre-selected in the creation wizard
- [x] Mode selected in wizard is saved into the OAuth client config (`x_traditional_webapp_session_type`)
- [x] Quickstart always shows Next.js (first) and Other Framework items for Traditional Web Application
- [x] `npm run typecheck`, `npm run eslint`, `npm run prettier`, and `make check-tidy` all pass clean
