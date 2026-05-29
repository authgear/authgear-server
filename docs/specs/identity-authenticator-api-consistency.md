# Identity & Authenticator Data Model — Cross-Surface Consistency Analysis

## Background

The identity and authenticator payload shapes differ across Authgear's API surfaces. This document maps what each surface currently exposes, identifies the remaining inconsistencies, and proposes a canonical data model to align them.

## Surfaces Covered

| Surface | Entry point |
|---|---|
| OIDC Userinfo | `https://authgear.com/claims/user/identities`, `https://authgear.com/claims/user/authenticators` |
| Admin GraphQL | `nodeIdentity`, `nodeAuthenticator` |
| Auth Flow API | `IdentificationOption` in flow state response |
| User Import API | `POST /_api/admin/users/import` |
| Webhook Events | `identity.oauth.connected`, `identity.oauth.disconnected`, etc. |

---

## What Each Surface Exposes

### Surface 1 — OIDC Userinfo

Source: `pkg/api/model/userinfo.go`, `pkg/lib/userinfo/userinfo.go`

```json
// Identity element (https://authgear.com/claims/user/identities)
{
  "type": "login_id",          // snake_case enum
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "login_id_key": "email",          // login_id only — configured key name
  "login_id_type": "email",         // login_id only — resolved type
  "oauth_provider_type": "google",  // oauth only — provider kind
  "oauth_provider_alias": "google"  // oauth only — configured alias
}

// Authenticator element (https://authgear.com/claims/user/authenticators)
{
  "type": "password",          // snake_case
  "kind": "primary",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

Typed fields are only present for `login_id` and `oauth` identity types. All other identity types (`passkey`, `biometric`, `anonymous`, `siwe`, `ldap`) appear in the array with only `type`, `created_at`, and `updated_at` — no type-specific fields.

Notable omissions compared to the canonical model:
- `oauth_subject_id` is not exposed (privacy consideration — not needed by typical OIDC clients).
- `login_id_value` is intentionally not exposed (PII; available via Admin API only).
- Authenticator type-specific fields (`is_default`, `totp_display_name`, `oob_otp_email`, etc.) are not exposed — authenticator detail is not the purpose of this endpoint.

---

### Surface 2 — Admin GraphQL

Source: `pkg/admin/graphql/identity.go`, `pkg/admin/graphql/authenticator.go`

```graphql
type Identity {
  id: ID!
  createdAt: DateTime!           # camelCase
  updatedAt: DateTime!
  type: IdentityType!            # SCREAMING_SNAKE_CASE: LOGIN_ID, OAUTH, ...
  claims: IdentityClaims!        # raw map with fully-namespaced keys
}

type Authenticator {
  id: ID!
  createdAt: DateTime!
  updatedAt: DateTime!
  type: AuthenticatorType!       # SCREAMING_SNAKE_CASE: PASSWORD, OOB_OTP_EMAIL, ...
  kind: AuthenticatorKind!       # PRIMARY, SECONDARY
  isDefault: Boolean!
  expireAfter: DateTime          # password only
  claims: AuthenticatorClaims!   # raw map
}
```

The `claims` scalar exposes the raw internal claims map defined in `pkg/lib/authn/identity/claim_key.go`. Example keys for `login_id` and `oauth`:

```
https://authgear.com/claims/login_id/type           → login ID type ("email" | "phone" | "username")
https://authgear.com/claims/login_id/key            → configured key name
https://authgear.com/claims/login_id/value          → actual login ID value
https://authgear.com/claims/login_id/original_value → pre-normalisation value
https://authgear.com/claims/oauth/provider_type     → OAuth provider kind ("google", "facebook", …)
https://authgear.com/claims/oauth/provider_alias    → configured alias
https://authgear.com/claims/oauth/subject_id        → provider subject ID
https://authgear.com/claims/oauth/profile           → raw provider profile
```

All identity types (`login_id`, `oauth`, `passkey`, `biometric`, `anonymous`, `siwe`, `ldap`) have their type-specific data in the same opaque map. Callers must know internal claim URIs to read anything useful — there are no typed fields for `loginIDKey`, `oauthProviderType`, `oauthProviderAlias`, etc.

---

### Surface 3 — Auth Flow API

Source: `pkg/lib/authenticationflow/declarative/data_identification.go`

This surface exposes *available identification options* during the login/signup step, not the identities already linked to a user.

```json
{
  "identification": "email",    // login ID type, NOT identity type ("email"|"phone"|"username"|"oauth"|"passkey"|"ldap")
  "provider_type": "google",    // oauth: provider kind — should be oauth_provider_type
  "alias": "google",            // oauth: configured alias — should be oauth_provider_alias
}
```

Notable inconsistencies with the canonical model:
- `provider_type` → canonical name is `oauth_provider_type`.
- `alias` → canonical name is `oauth_provider_alias`.
- The `identification` field conflates login-ID-type with identity-type (`"email"` means "identify by email login ID", not `IdentityType.LoginID`).

---

### Surface 4 — Webhook Events

Source: `pkg/lib/hook/`, `pkg/api/event/`, `docs/specs/event.md`

Identity data in webhook payloads (e.g. `identity.oauth.connected`, `identity.oauth.disconnected`) uses the same raw claims map as Admin GraphQL:

```json
{
  "identity": {
    "id": "...",
    "type": "oauth",
    "claims": {
      "https://authgear.com/claims/oauth/provider_type": "google",
      "https://authgear.com/claims/oauth/provider_alias": "google",
      "https://authgear.com/claims/oauth/subject_id": "1234567",
      "https://authgear.com/claims/oauth/profile": {}
    },
    "created_at": "...",
    "updated_at": "..."
  }
}
```

Like Admin GraphQL, callers must know internal claim URIs to extract `oauth_provider_type` and `oauth_provider_alias`. There are no top-level typed fields.

---

### Surface 5 — User Import API

Source: `pkg/lib/userimport/model.go`, docs: `POST /_api/admin/users/import`

Login IDs are represented as flat top-level standard attribute fields — there is no structured identity array:

```json
{
  "email": "user@example.com",
  "phone_number": "+85200000000",
  "preferred_username": "alice",
  "email_verified": true,
  "phone_number_verified": true,
  "password": { "type": "bcrypt", "password_hash": "..." },
  "mfa": {
    "email": "mfa@example.com",
    "phone_number": "+85200000001",
    "totp": { "secret": "..." }
  }
}
```

This API is intentionally limited:
- Only `login_id` identities are supported (via standard attribute fields).
- Only `password` and `totp` authenticators are importable.
- All other identity types (`oauth`, `passkey`, `biometric`, `anonymous`, `siwe`, `ldap`) are not importable.
- There is no structured `identities` array — login IDs are inferred from standard attributes.

---

## Inconsistency Map

### Identity fields

| Concept | OIDC Userinfo | Admin GraphQL | Webhook Events | Auth Flow API | Internal claim key |
|---|---|---|---|---|---|
| Identity type enum | `login_id`, `oauth` (snake_case) | `LOGIN_ID`, `OAUTH` (SCREAMING_SNAKE_CASE) | `login_id`, `oauth` (snake_case) | `email`, `phone`, `username`, `oauth` (mixed — login ID type conflated) | N/A |
| OAuth provider kind | `oauth_provider_type` | `claims["…/oauth/provider_type"]` | `claims["…/oauth/provider_type"]` | **`provider_type`** (→ `oauth_provider_type`) | `https://authgear.com/claims/oauth/provider_type` |
| OAuth provider alias | `oauth_provider_alias` | `claims["…/oauth/provider_alias"]` | `claims["…/oauth/provider_alias"]` | **`alias`** ← different name | `https://authgear.com/claims/oauth/provider_alias` |
| Login ID key | `login_id_key` | `claims["…/login_id/key"]` | `claims["…/login_id/key"]` | not returned | `https://authgear.com/claims/login_id/key` |
| Login ID type | `login_id_type` | `claims["…/login_id/type"]` | `claims["…/login_id/type"]` | conflated with `identification` | `https://authgear.com/claims/login_id/type` |
| Login ID value | not exposed (by design) | `claims["…/login_id/value"]` | `claims["…/login_id/value"]` | request input only | `https://authgear.com/claims/login_id/value` |
| `created_at`/`updated_at` | snake_case | camelCase | snake_case | N/A | N/A |

### Authenticator fields

| Concept | OIDC Userinfo | Admin GraphQL | Internal |
|---|---|---|---|
| Type enum | `password`, `oob_otp_email` (snake_case) | `PASSWORD`, `OOB_OTP_EMAIL` (SCREAMING_SNAKE_CASE) | snake_case |
| Kind enum | `primary`, `secondary` (snake_case) | `PRIMARY`, `SECONDARY` (SCREAMING_SNAKE_CASE) | snake_case |
| `isDefault` | not exposed | `isDefault` | stored in DB |
| `expireAfter` | not exposed | `expireAfter` | stored in DB |
| OOB email address | not exposed | `claims["…"]` | stored in DB |
| OOB phone number | not exposed | `claims["…"]` | stored in DB |

### Remaining inconsistencies

1. **`alias` vs `oauth_provider_alias`** — Auth Flow uses `alias`; OIDC and the internal model use `provider_alias` (canonical: `oauth_provider_alias`). Fix: add `oauth_provider_alias` alongside `alias` (non-breaking).
2. **`provider_type` vs `oauth_provider_type`** — Auth Flow uses the unqualified `provider_type`; the canonical name is `oauth_provider_type` to match the `login_id_*` / `oauth_*` prefix pattern. Fix: add `oauth_provider_type` alongside `provider_type` (non-breaking).
3. **Admin GraphQL `claims` is an opaque blob** — typed fields exist for `type`, `kind`, and timestamps, but all identity-type-specific data requires knowing internal claim URIs.
4. **Auth Flow `identification` conflates concepts** — `"email"` means "identify by email login ID", not `IdentityType.LoginID`. This is confusing when building a mental model across surfaces.
5. **Enum casing differs by surface** — OIDC and Auth Flow API use snake_case; Admin GraphQL uses SCREAMING_SNAKE_CASE. This is expected: each surface follows its own paradigm convention (GraphQL spec vs JSON/REST). No fix needed.

---

## Design Rationale

Two competing designs exist in the codebase:

**Raw map with namespaced keys** (used by Admin GraphQL and Webhook Events)
```json
{ "claims": { "https://authgear.com/claims/oauth/provider_type": "google" } }
```

**Canonical typed fields** (used by OIDC Userinfo)
```json
{ "oauth_provider_type": "google" }
```

The raw map was a shortcut: the internal identity model already uses namespaced claims (inherited from OIDC/JWT conventions), so exposing it directly was cheap to implement. But it leaks an internal abstraction that API consumers should never need to know about.

Typed fields are the better design for public APIs because:
- **Discoverable** — field names appear in the GraphQL schema or OpenAPI spec; callers don't need to know internal claim URIs.
- **Type-safe** — codegen produces typed structs rather than opaque maps.
- **Consistent** — matches how every other modern API surface works.

The long-term direction is typed fields everywhere, with the raw `claims` map kept only as an escape hatch for data not yet promoted to a first-class field.

---

## Proposed Canonical Data Model

Define the canonical shape once. Each surface projects it with its own casing convention. The fields listed below are representative — not every claim key is included, but the naming pattern should be followed when adding new fields during implementation.

### Canonical Identity

```
Identity {
  id             string    // prefixed node ID (Admin API only)
  type           enum      // login_id | oauth | anonymous | biometric | passkey | siwe | ldap

  created_at     RFC3339
  updated_at     RFC3339

  // login_id fields
  login_id_key   string    // configured key name: "email" | "phone" | "username"
  login_id_type  string    // resolved type: "email" | "phone" | "username"
  login_id_value string    // actual value (Admin API only — PII, gate by scope)

  // oauth fields
  oauth_provider_type  string    // "google" | "facebook" | "github" | ...
  oauth_provider_alias string    // configured alias, e.g. "google" or "my_google"
  oauth_subject_id     string    // user's ID from the OAuth provider

  // anonymous fields
  anonymous_key_id     string    // JWK key ID of the anonymous user's key pair

  // biometric fields
  biometric_key_id              string    // JWK key ID of the biometric key pair
  biometric_formatted_device_info string  // human-readable device model (e.g. "iPhone 14 Pro")

  // passkey fields
  passkey_credential_id  string    // WebAuthn credential ID — identifies which device
  passkey_display_name   string    // human-readable name for the passkey (from WebAuthn)

  // siwe fields
  siwe_chain_id  int       // EIP-155 chain ID (e.g. 1 for Ethereum mainnet)
  siwe_address   string    // EIP-55 checksummed wallet address

  // ldap fields
  ldap_server_name      string    // configured LDAP server name
}
```

### Canonical Authenticator

```
Authenticator {
  id             string
  type           enum      // password | totp | oob_otp_email | oob_otp_sms | passkey
  kind           enum      // primary | secondary
  is_default     bool

  created_at     RFC3339
  updated_at     RFC3339

  // totp fields
  totp_display_name  string    // user-provided label when setting up TOTP

  // oob_otp_email fields
  oob_otp_email          string    // address where OTP codes are delivered

  // oob_otp_sms fields
  oob_otp_phone_number   string    // number where OTP codes are delivered

  // password fields
  password_expire_after   RFC3339

  // passkey fields
  passkey_credential_id  string    // WebAuthn credential ID — identifies which device
}
```

---

## Recommended Changes (by priority)

### P1 — Additive, no breaking change

**Add `oauth_provider_type` and `oauth_provider_alias` to Auth Flow API identification options alongside the existing fields.**
Keep `provider_type` and `alias` unchanged so existing clients are unaffected:

```json
{
  "identification": "oauth",
  "provider_type": "google",         // kept for backwards compatibility
  "oauth_provider_type": "google",   // new canonical name
  "alias": "google",                 // kept for backwards compatibility
  "oauth_provider_alias": "google"   // new canonical name
}
```

---

## Out of Scope

### Add typed fields to Admin GraphQL `Identity` and `Authenticator`, and Webhook Events

Keep `claims` as an escape hatch but add first-class fields so callers do not need to know internal claim URIs:

```graphql
type Identity {
  # ... existing fields ...
  loginIDKey:    String    # login_id only
  loginIDType:   String    # login_id only
  loginIDValue:  String    # login_id only — gate by admin scope
  oauthProviderType:  String    # oauth only
  oauthProviderAlias: String    # oauth only
  claims: IdentityClaims!  # kept as escape hatch
}
```

```graphql
type Authenticator {
  # ... existing fields ...
  email:       String    # oob_otp_email only
  phoneNumber: String    # oob_otp_sms only
}
```

For webhook events, the equivalent would be promoting `oauth_provider_type` and `oauth_provider_alias` to top-level fields alongside `claims`:

```json
{
  "identity": {
    "type": "oauth",
    "oauth_provider_type": "google",
    "oauth_provider_alias": "google",
    "claims": { "…": "…" }
  }
}
```

### Add an `identities` array to the User Import API

Allow importing OAuth identities and biometric keys, consistent with the shape used by the OIDC claim.
