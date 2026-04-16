# Design: Authentication Mode Selector for Traditional Web Applications

**Date:** 2026-04-16
**Status:** Approved
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

Currently the application detail page shows the "Cookie-based authentication" section for all Traditional Web Application clients, with no way to indicate SDK usage or surface token lifetime settings for that type.

---

## Design

### New UI element: Authentication Mode radio group

A clearly labelled **"Authentication Mode"** section is added to the application detail page for Traditional Web Application clients. It sits between **Basic Info** and **URIs**.

The section contains two radio options:

**Option 1 — Cookie Session**
> Authgear manages cookies on your project domain. Session lifetime is configured in Cookie Session Settings.

**Option 2 — Server-side SDK**
> The SDK manages tokens server-side. Session lifetime is set by access token & refresh token settings below. e.g. authgear-sdk-nextjs for Next.js.

---

### Section visibility rules

| Mode selected | Sections shown | Sections hidden |
|---|---|---|
| Cookie Session | Cookie-based authentication (link to session settings) | Refresh Token fields, Access Token fields |
| Server-side SDK | Refresh Token fields, Access Token fields | Cookie-based authentication section |

The URIs section, Basic Info, and Custom UI sections are unaffected by mode selection.

---

### Default value

- **Existing applications** (created before this feature ships): default to **Cookie Session**. No user action required; existing behaviour is preserved.
- **New applications** (created after this feature ships): default to **Cookie Session**.

---

### Mode switching behaviour

When a user switches from **Server-side SDK → Cookie Session** within a single editing session (before saving), an inline note appears inside the Authentication Mode box:

> ℹ️ Token lifetime settings are not used in Cookie Session mode.

The note is visible only while the user has an unsaved mode change (SDK → Cookie Session) in the current session. It disappears if the user switches back to Server-side SDK. After saving with Cookie Session selected, the note does not appear on subsequent page loads — Cookie Session is now the persisted mode.

The token lifetime values themselves are preserved in the database — they are simply hidden from the UI. Switching back to Server-side SDK (even in a future session) restores the fields with their previously saved values.

No confirmation dialog is shown — the action is fully reversible.

---

### Data model

A new field is required on the OAuth client config to persist the user's mode selection. The exact field name should follow the existing naming convention in `pkg/lib/config/` — suggested: `x_traditional_webapp_auth_mode`. Values: `cookie_session` (default) | `server_side_sdk`.

This field is only meaningful for `application_type = traditional_webapp`. It has no effect on other application types.

---

## Scope

**In scope:**
- Authentication Mode radio group on `EditOAuthClientForm` for `traditional_webapp` type
- Conditional rendering of Cookie Session section vs. Token Lifetime sections based on selected mode
- Inline note on mode switch
- Default to Cookie Session for all existing and new Traditional Web Application clients
- New config field to persist the mode selection

**Out of scope:**
- Changes to other application types (SPA, Native, M2M, OIDC/SAML)
- Proposed change 1 (rename "Traditional Web Application" label) — separate task
- Proposed change 3 (Next.js quickstart guide) — separate task
- Backend enforcement or validation of token settings based on mode

---

## Files likely affected

| File | Change |
|---|---|
| `portal/src/graphql/portal/EditOAuthClientForm.tsx` | Add Authentication Mode radio group; conditionally show/hide sections |
| `portal/src/locale-data/en.json` | Add translation keys for new labels and inline note |
| `pkg/lib/config/` | Add new field to OAuth client config schema |
| Portal GraphQL mutation | Include new field in update mutation |

---

## Acceptance criteria

- [ ] Authentication Mode radio group appears on the application detail page for Traditional Web Application clients only
- [ ] Cookie Session mode: shows Cookie-based authentication section, hides token lifetime fields
- [ ] Server-side SDK mode: shows Refresh Token and Access Token fields, hides Cookie-based authentication section
- [ ] Existing apps default to Cookie Session with no user action required
- [ ] Switching from SDK → Cookie Session shows the inline note; switching back removes it
- [ ] Token lifetime values are preserved (not deleted) when switching to Cookie Session mode
- [ ] `npm run typecheck` passes clean
