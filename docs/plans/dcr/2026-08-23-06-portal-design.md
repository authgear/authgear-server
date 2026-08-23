# DCR Part 6 — Portal Configuration (Design)

Spec: [docs/specs/dcr.md](../../specs/dcr.md). Backend: PR #5870 (parts 1–5). Ticket: DEV-3801.

Companion implementation plan: [2026-08-23-06-portal-plan.md](./2026-08-23-06-portal-plan.md).

## 1. Problem Statement

PR #5870 implements DCR end to end on the backend, but every admin-facing control
is out of reach of the portal:

- Enabling DCR and choosing between IAT-required and open registration means
  hand-editing `oauth.dynamic_client_registration` in `authgear.yaml` (or the raw
  config editor).
- Issuing and revoking Initial Access Tokens requires calling the Admin GraphQL
  API directly (`createInitialAccessToken`, `revokeInitialAccessToken`).
- Registered dynamic clients are invisible: there is no way to see what has
  registered against a project, and deleting one (`deleteDynamicClient`) again
  requires raw GraphQL.
- Granting dynamic clients access to API Resources and Scopes requires calling
  `createResource` / `updateScope` with `accessPolicy` by hand — the portal's
  API Resources screens neither show nor edit
  `allow_dynamic_third_party_client_access`.

The PR's own "how to test" instructions demonstrate the friction: GraphiQL
mutations plus YAML edits before a single MCP client can connect. The target
user — a project admin setting up MCP-server auth or CI ephemeral clients —
should be able to complete the whole setup in the portal.

This matches the expectations listed in DEV-3801
([comment](https://linear.app/authgear/issue/DEV-3801/update-portal-for-dcr-configs#comment-427995ff)):
(1) issue/revoke IAT, (2) list/delete dynamic clients, (3) enable/disable DCR
with token lifetime configs, (4) grant dynamic clients access to API
Resources + Scopes.

## 2. Goals

- A project admin can enable DCR, choose IAT-required vs open registration, and
  set default token lifetimes for DCR clients — entirely in the portal.
- A project admin can issue and revoke IATs, with the one-time token value
  surfaced exactly once, and the elevated privilege of first-party IATs made
  visually obvious.
- A project admin can see all DCR-registered clients, inspect their metadata,
  delete them, and see usage against the plan's dynamic-client cap.
- A project admin can allow dynamic third-party clients to access an API
  Resource and select which of its Scopes they may request.
- The complete MCP setup flow (spec UC2) and CI ephemeral-client flow (spec
  UC1) are achievable without leaving the portal.

## 3. Non-Goals

- Rate-limit editing UI (`rate_limits.per_ip` / `.per_project`) — defaults are
  sensible; the raw config editor remains the escape hatch.
- Per-client config overrides, RFC 7592 self-service management, per-IAT config
  — all "Future Works" in dcr.md.
- CIMD UI. The `dynamicClients` query already returns CIMD clients and the list
  renders `source` generically, but no CIMD-specific configuration surface is
  designed here.
- Registering clients *from* the portal. DCR clients register themselves via
  `POST /oauth2/register`; static clients keep the existing Create Application
  flow.
- Backend changes of any kind. Parts 1–5 already expose everything this design
  consumes.

## 4. Information Architecture

Everything lives under the existing **Client Applications** area, because
DCR-registered clients are OAuth clients and that is where admins look for
them.

```
Client Applications (nav item, unchanged)
└── Applications screen  (/project/:appID/configuration/apps)
    ├── Tab: Applications        — existing static client list, unchanged
    └── Tab: Dynamic clients     — NEW: DB-backed paginated list
        └── "Registration settings" button
            └── DCR settings screen  (/project/:appID/configuration/apps/dcr)
                ├── Section: Enable registration
                ├── Section: Registration security (IAT-required + IAT management)
                └── Section: Default client configuration

API Resources (existing area)
├── Resource details › Applications tab — NEW pinned "Dynamic clients" toggle row
└── Scope editor — NEW "allow dynamic third-party clients" checkbox
    (+ "Dynamic" indicator in the Scopes tab list)
```

No new top-level navigation item. Rationale (decided during design review):
a dedicated nav item adds clutter for a feature most projects won't enable,
and splitting the pieces across Advanced/Admin API would fragment one story.

## 5. Surface Designs

### 5.1 Dynamic clients tab (Applications screen)

- Paginated table over the Admin API `dynamicClients` connection. Columns:
  client name, `client_id` (`dcrc_…`, copyable), kind (First-party /
  Third-party badge), registered at. Source (DCR/CIMD) is shown in the detail
  view only, until CIMD ships.
- Header line above the table:
  - Registration endpoint as a copy field: `<public_origin>/oauth2/register`.
  - When the plan configures an `oauth_client_dcr` usage limit with
    `action: block`: "N of Q dynamic clients used" (N = connection
    `totalCount`, Q = smallest `block` quota). Hidden when no limit exists.
  - "Registration settings" button → DCR settings screen.
- Row click opens a read-only details dialog: name, client ID (copyable),
  kind, source, registered at, redirect URIs, grant types, response types,
  application type, logo/client/ToS/policy URIs. Footer: Delete button.
- Delete (from dialog or row action) opens a confirmation dialog that states
  the current backend behavior honestly: deletion frees a limit slot and
  blocks new authorizations, but **already-issued tokens stay valid until they
  expire** (dcr.md `deleteDynamicClient` implementation status).
- Empty states:
  - DCR disabled: explains the feature, "Enable registration" CTA → settings
    screen.
  - DCR enabled, no clients yet: explains that clients self-register, shows
    the registration endpoint copy field.

### 5.2 DCR settings screen

One screen, three sections. The config sections save through the standard
app-config save bar; IAT actions (create/revoke) apply immediately, independent
of unsaved config edits.

**Enable registration** — toggle bound to
`oauth.dynamic_client_registration.enabled`.

**Registration security** —

- "Require initial access token" toggle bound to
  `initial_access_token_required` (absent = on, the spec default). Turning it
  **off** opens a confirmation dialog warning that anyone will be able to
  register clients, and that a standing client cap, once reached, stays
  exhausted until an admin deletes clients.
- IAT list (active tokens only, as returned by `initialAccessTokens`): type
  badge (Third-party neutral; First-party in warning styling), created at,
  expires at, per-row Revoke with confirmation.
- "Create token" button → dialog: type radio (default Third-party; selecting
  First-party reveals an inline warning to treat the token like the Admin API
  private key) + expiry select (1 hour / 24 hours / 7 days / 30 days).
  On success a second dialog shows the plaintext token once, in a copy field,
  with a "you won't be able to see this token again" notice.
- The IAT section stays visible even under open registration: an IAT still
  has meaning there (a first-party IAT registers first-party clients).

**Default client configuration** — the four fields of
`default_client_config`, mirroring the existing per-client token settings UI:
access token lifetime, refresh token lifetime, idle timeout enabled, idle
timeout. Placeholder values show the effective defaults (1800 / 2592000 / on /
1209600).

### 5.3 Resource / scope access policy

- **Resource › Applications tab**: a pinned "Dynamic clients" row above the
  per-client m2m grant list — same place admins already answer "who can access
  this resource". Its toggle writes
  `accessPolicy.allowDynamicThirdPartyClientAccess` via `updateResource`. The
  row is collective ("all dynamically registered clients"), visually distinct
  from individual application rows.
- **Scope editor** (create and edit): checkbox "Allow dynamic third-party
  clients to request this scope" bound to the scope's
  `accessPolicy.allowDynamicThirdPartyClientAccess`. When the resource-level
  toggle is off, the checkbox stays editable but a note explains that
  resource-level access is currently off and gates overall access.
- **Scopes tab list**: a small "Dynamic" badge on scopes that allow dynamic
  access.

## 6. User Flows

### UF1 — Enable open registration for an MCP server (spec UC2)

1. Client Applications → Dynamic clients tab → empty state → **Enable
   registration**.
2. Settings screen: turn **Enable registration** on; turn **Require initial
   access token** off → warning dialog → confirm → Save.
3. API Resources → create resource for `https://mcp-server.example.com` with
   scopes `read:tools`, `execute:tools`.
4. Resource → Applications tab → toggle **Dynamic clients** on.
5. Scopes tab → each scope → tick "Allow dynamic third-party clients" → Save.
6. Done — MCP clients self-register and appear in the Dynamic clients tab.

### UF2 — CI ephemeral first-party clients (spec UC1)

1. Settings screen: turn **Enable registration** on; leave IAT required. Save.
2. Registration security → **Create token** → type First-party (warning
   shown) → expiry → Create.
3. Copy the `iat_fp_…` token from the one-time reveal dialog into CI secrets.
4. CI calls `POST /oauth2/register` with the token; clients appear in the
   Dynamic clients tab; CI deletes them via Admin API when PRs close.

### UF3 — Monitor and clean up

1. Dynamic clients tab: watch "N of Q dynamic clients used".
2. Click a suspicious client → inspect metadata → **Delete** → confirmation
   (notes tokens remain valid until expiry) → confirm.

### UF4 — Revoke an IAT

1. Settings screen → Registration security → IAT row → **Revoke** → confirm.
   The token can no longer be used to register clients.
