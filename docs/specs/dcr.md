# Dynamic Client Registration (DCR)

Authgear supports Dynamic Client Registration as defined by:

- [RFC 7591 — OAuth 2.0 Dynamic Client Registration Protocol](https://www.rfc-editor.org/rfc/rfc7591)
- [OpenID Connect Dynamic Client Registration 1.0](https://openid.net/specs/openid-connect-registration-1_0.html)

> See also [Client ID Metadata Documents (CIMD)](./cimd.md) — a proposed, registration-free alternative for the same "unregistered client" problem, likely to supersede DCR's open-registration mode for the MCP use case below.

## Table of Contents

- [Glossary](#glossary)
- [Use Cases](#use-cases)
- [Configuration](#configuration)
  - [Client Limit](#client-limit)
  - [Rate Limits](#rate-limits)
- [OIDC Discovery Metadata](#oidc-discovery-metadata)
- [Initial Access Token](#initial-access-token)
- [Registration Endpoint](#registration-endpoint)
  - [Request](#request)
  - [Response](#response)
  - [Errors](#errors)
- [Accepted Client Metadata](#accepted-client-metadata)
- [Client ID Format](#client-id-format)
- [Storage Architecture](#storage-architecture)
- [Security Considerations](#security-considerations)
  - [Access Token Audience Binding](#access-token-audience-binding)
- [Admin API](#admin-api)
  - [IAT management](#iat-management)
- [Audit Log](#audit-log)
- [Future Works](#future-works)

## Glossary

**Dynamic Client Registration (DCR)** — the process by which an OAuth client registers itself programmatically with an Authorization Server at runtime, rather than being statically configured in `authgear.yaml`.

**Initial Access Token (IAT)** — an opaque token issued by the Admin API and presented to the registration endpoint. Two types exist, with distinct token prefixes that make their privilege level immediately visible:

- **Third-party IAT** (prefix `iat_tp_`) — allows registration of `web` and `native` clients as third-party clients (consent screen shown). Lower privilege; safe to distribute to developers building integrations.
- **First-party IAT** (prefix `iat_fp_`) — allows registration of `web` and `native` clients as first-party clients (consent screen bypassed). High privilege — treat with the same care as the Admin API private key.

When `initial_access_token_required: false` (open registration), no IAT is required and only third-party clients may be registered.

## Use Cases

### UC1. Ephemeral clients for CI / pull-request preview environments

A CI system holds the Admin API private key for a project. For each pull request, the CI registers a new first-party client scoped to that PR's redirect URI.

A first-party IAT (`iat_fp_`) is required because first-party clients bypass the consent screen and must only be created by an authorized administrator.

**Required configuration:**

```yaml
oauth:
  dynamic_client_registration:
    enabled: true
    initial_access_token_required: true   # default; explicitly set for clarity
```

No `default_client_config` override is needed — CI clients use the project-level token lifetimes and do not require resource indicator support.

**Step 1 — Create a first-party IAT via the Admin API**

Call the `createInitialAccessToken` Admin API mutation (see [Admin API](#admin-api)):

```graphql
mutation {
  createInitialAccessToken(input: { type: FIRST_PARTY, expiresIn: 3600 }) {
    token      # iat_fp_Xf2kLmNpQrStUvWx
    initialAccessToken {
      expiresAt
    }
  }
}
```

Store the returned `token` value securely — it is returned once only.

**Step 2 — Register the client**

```
POST /oauth2/register HTTP/1.1
Host: myapp.authgear.cloud
Content-Type: application/json
Authorization: Bearer <iat>

{
  "client_name": "PR #123 preview",
  "redirect_uris": ["https://pr-123.preview.example.com/callback"],
  "application_type": "web"
}
```

Response:

```json
{
  "client_id": "dcrc_Xf2kLmNpQrStUvWx",
  "client_id_issued_at": 1700000000,
  "client_name": "PR #123 preview",
  "redirect_uris": ["https://pr-123.preview.example.com/callback"],
  "grant_types": ["authorization_code", "refresh_token"],
  "response_types": ["code"],
  "application_type": "web"
}
```

**Step 3 — Use the client in the authorization code flow**

The PR preview app uses `client_id=dcrc_Xf2kLmNpQrStUvWx` as a normal SPA client for the lifetime of the PR.

**Step 4 — Backend validates the access token**

Because no `resource` parameter is used, the issued access token is a JWT with the default audience:

```json
{
  "iss": "https://myapp.authgear.cloud",
  "sub": "<user-id>",
  "aud": ["https://myapp.authgear.cloud"],
  "client_id": "dcrc_Xf2kLmNpQrStUvWx",
  "scope": "openid"
}
```

The PR preview backend validates the token as follows:

1. Confirm the token is a JWT.
2. Fetch `jwks_uri` from `https://myapp.authgear.cloud/.well-known/openid-configuration` and verify the JWT signature.
3. Check `iss` equals `https://myapp.authgear.cloud`.
4. Check `aud` includes `https://myapp.authgear.cloud`.
5. Check `exp` has not elapsed.

> When the PR is closed, the CI can remove the client with the [`deleteDynamicClient`](#new-mutation) Admin API mutation, which also frees a slot against the project's [client limit](#client-limit). Client *self*-service management — a client reading, updating or deleting its own registration — is deferred to RFC 7592; see [Future Works](#future-works).

---

### UC2. MCP (Model Context Protocol) clients

Per the [MCP Authorization specification](https://modelcontextprotocol.io/specification/2025-11-25/basic/authorization), each MCP client registers itself with the Authorization Server at first use. With open registration enabled, MCP clients self-register without any admin involvement per client.

**Required configuration:**

```yaml
oauth:
  dynamic_client_registration:
    enabled: true
    initial_access_token_required: false   # open registration — no IAT needed
```

**Admin setup (once)**

1. Enable open registration as shown above.
2. In the portal, create an API Resource for `https://mcp-server.example.com` with scopes `read:tools` and `execute:tools`.
3. On the Resource and on each scope that MCP clients should be able to request, set `access_policy.allow_dynamic_third_party_client_access: true`.

No further per-client admin action is required — any MCP client can self-register and immediately use the declared resources.

**Step 1 — Discover the authorization server**

```
GET /.well-known/oauth-authorization-server HTTP/1.1
Host: myapp.authgear.cloud
MCP-Protocol-Version: 2025-11-25
```

Response includes `registration_endpoint`.

**Step 2 — Register the client**

```
POST /oauth2/register HTTP/1.1
Host: myapp.authgear.cloud
Content-Type: application/json

{
  "redirect_uris": ["https://mcp-client.example.com/callback"]
}
```

Response:

```json
{
  "client_id": "dcrc_AbCdEfGhIjKlMnOpQr",
  "client_id_issued_at": 1700000000,
  "client_name": "Client dcrc_AbCdEfGhIjKlMnOpQr",
  "redirect_uris": ["https://mcp-client.example.com/callback"],
  "grant_types": ["authorization_code", "refresh_token"],
  "response_types": ["code"],
  "application_type": "web"
}
```

**Step 3 — Authorization code flow with resource indicator**

```
GET /oauth2/authorize
  ?client_id=dcrc_AbCdEfGhIjKlMnOpQr
  &response_type=code
  &scope=openid+read:tools
  &redirect_uri=https://mcp-client.example.com/callback
  &code_challenge=<challenge>
  &code_challenge_method=S256
  &resource=https://mcp-server.example.com HTTP/1.1
Host: myapp.authgear.cloud
```

The user sees a consent screen and authorizes the MCP client.

**Step 4 — Exchange code for tokens**

```
POST /oauth2/token HTTP/1.1
Host: myapp.authgear.cloud
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code
&code=<code>
&code_verifier=<verifier>
&client_id=dcrc_AbCdEfGhIjKlMnOpQr
&redirect_uri=https://mcp-client.example.com/callback
&resource=https://mcp-server.example.com
```

The issued access token has `aud: ["https://mcp-server.example.com"]` (the resource URI only; the project endpoint is not included). The MCP server validates `aud` contains its own URI. If `openid` or other OIDC scopes were also requested and granted, the userinfo endpoint remains accessible via that token.

## Configuration

```yaml
oauth:
  dynamic_client_registration:
    enabled: true
    initial_access_token_required: true
    default_client_config:
      access_token_lifetime_seconds: 1800
      refresh_token_lifetime_seconds: 2592000
      refresh_token_idle_timeout_enabled: true
      refresh_token_idle_timeout_seconds: 1209600
```

- `oauth.dynamic_client_registration.enabled`: Optional. Boolean. Default `false`. Enables `POST /oauth2/register`.
- `oauth.dynamic_client_registration.initial_access_token_required`: Optional. Boolean. Default `true`. When `true`, registration requires a valid IAT in the `Authorization: Bearer` header. When `false`, open registration is permitted and every client is registered as third-party. The set of accepted `application_type` values does not depend on this setting: it is always `web` and `native` only — see [Accepted Client Metadata](#accepted-client-metadata). What the IAT determines is whether the registered client is first-party or third-party, not which `application_type` values are allowed.

- `oauth.dynamic_client_registration.default_client_config`: Optional. Object. The default client config applied to all DCR-registered clients. Useful when stricter settings are needed for the DCR cohort. Per-client overrides are not yet supported; see [Future Works](#future-works). Supports a subset of the fields defined in [Custom Client Metadata](./oidc.md#custom-client-metadata): `access_token_lifetime_seconds`, `refresh_token_lifetime_seconds`, `refresh_token_idle_timeout_enabled`, `refresh_token_idle_timeout_seconds`.

> **Note:** Resource access for third-party clients is configured via the portal, not `authgear.yaml`. Resources and Scopes with `access_policy.allow_dynamic_third_party_client_access: true` are accessible to dynamic third-party clients — DCR-registered today, CIMD-resolved later — not to a static `third_party_app` client declared in `authgear.yaml`, which has no mechanism to be granted resource access for these grants. See [API Resources and Scopes](./api-resource.md#access-policy).

### Client Limit

A DCR client is created by anyone holding a valid IAT (or, under open registration, by anyone at all) — no per-client admin action is required, unlike a static client. Left uncapped, this lets the project's client population grow without bound.

The limit is configured as a [usage limit](./usage.md) under the `oauth_client_dcr` usage name:

**authgear.features.yaml**

```yaml
usage:
  limits:
    oauth_client_dcr:
      - quota: 20
        action: block
```

- `usage.limits.oauth_client_dcr`: Optional. Default absent (no limit). The maximum number of DCR-registered clients the project may have at once, checked against the current count of `OAuthClient` records with `source: DCR` — a [standing usage name](./usage.md#supported-usage-names). Once at `quota`, `POST /oauth2/register` is rejected with `access_denied` (see [Errors](#errors)) regardless of IAT validity, via the matching entry's `action: block`.

This is a plan-tier limit, set in `authgear.features.yaml`'s feature-config hierarchy, not something a project admin edits directly — distinct in both file and purpose from `oauth.dynamic_client_registration.*` in `authgear.yaml` above.

Note that a standing limit does not recover on its own. A rate-limit bucket refills and a periodic usage limit resets each period, but once `oauth_client_dcr` is at `quota` it stays there until an admin frees a slot with [`deleteDynamicClient`](#new-mutation). Under open registration this means anyone can render further registration impossible until an admin intervenes; a project that cannot tolerate that should keep `initial_access_token_required: true` (the default). Configuring an additional lower-`quota` entry with `action: alert` is recommended so the admin is notified before the cap is reached.

### Rate Limits

`POST /oauth2/register` is rate limited by two project-configurable limits:

| Name | Scope | Default Rate |
|---|---|---|
| `oauth.register.per_ip` | Per requesting IP, per project | 10 / minute |
| `oauth.register.per_project` | Per project | 1000 / hour |

Configured under `oauth.dynamic_client_registration.rate_limits` in `authgear.yaml`:

```yaml
oauth:
  dynamic_client_registration:
    rate_limits:
      per_ip:
        enabled: true
        period: 1m
        burst: 10
      per_project:
        enabled: true
        period: 1h
        burst: 1000
```

- `oauth.dynamic_client_registration.rate_limits.per_ip` / `.per_project`: Optional. Each defaults to the rate above when omitted. A project expecting many independent, legitimate first-time registrations in a short window — e.g. an MCP-style integration where each new user's client self-registers once on first connection — may need a higher `per_project` allowance than the default; size it relative to the project's own `oauth_client_dcr` quota (above), not the default shown here.

Both are consumed by every attempt, successful or not, and both are evaluated before the `Authorization` header is parsed — so an invalid IAT cannot be used to probe the endpoint more cheaply. Exceeding either returns `x_rate_limited` with HTTP 429 (see [Errors](#errors)).

These limits bound request volume against an endpoint that, under open registration, is unauthenticated and creates a database record per call. They are not a substitute for the [client limit](#client-limit) above, and they do not prevent a caller from reaching that limit. See [rate-limit.md](./rate-limit.md).

## OIDC Discovery Metadata

When DCR is enabled, `registration_endpoint` is added to the discovery documents at:

- `<endpoint>/.well-known/openid-configuration`
- `<endpoint>/.well-known/oauth-authorization-server`

Full example of `/.well-known/openid-configuration` with DCR enabled (fields taken from the actual Authgear implementation):

```jsonc
{
  "issuer": "https://myapp.authgear.cloud",
  "authorization_endpoint": "https://myapp.authgear.cloud/oauth2/authorize",
  "token_endpoint": "https://myapp.authgear.cloud/oauth2/token",
  "userinfo_endpoint": "https://myapp.authgear.cloud/oauth2/userinfo",
  "end_session_endpoint": "https://myapp.authgear.cloud/oauth2/logout",
  "revocation_endpoint": "https://myapp.authgear.cloud/oauth2/revoke",
  "jwks_uri": "https://myapp.authgear.cloud/oauth2/jwks",
  "registration_endpoint": "https://myapp.authgear.cloud/oauth2/register", // Added
  // ...
}
```

## Initial Access Token

An IAT is an **opaque** token issued by the Admin API (see [IAT management](#iat-management)). It is passed as `Authorization: Bearer <iat>` to the registration endpoint.

An IAT authorizes the bearer to register a new OAuth client. The key behavioral rules are:

- **With a first-party IAT** (`iat_fp_`) — `web` and `native` clients are registered as first-party (consent screen bypassed).
- **With a third-party IAT** (`iat_tp_`) — `web` and `native` clients are registered as third-party (consent screen shown).
- **Without an IAT** (open registration, `initial_access_token_required: false`) — `web` and `native` clients are registered as third-party.

### Per-IAT configuration

The Admin API may attach per-token configuration when creating an IAT. The exact set of supported config options is not yet defined and will be extended over time. The current behavior (IAT presence grants first-party registration) requires no additional config.

### IAT storage

IATs are stored hashed in the database. The plaintext value is returned exactly once at creation time and is not recoverable afterwards.

```sql
CREATE TABLE _auth_oauth_initial_access_token (
  id text PRIMARY KEY,
  app_id text NOT NULL,
  created_at timestamp without time zone NOT NULL,
  expires_at timestamp without time zone NOT NULL,
  -- 'THIRD_PARTY' | 'FIRST_PARTY', matching InitialAccessTokenType verbatim.
  -- Stored explicitly because the type cannot be recovered from token_hash:
  -- the iat_tp_ / iat_fp_ prefix is part of the hashed plaintext.
  token_type text NOT NULL,
  token_hash text NOT NULL
);
CREATE UNIQUE INDEX _auth_oauth_initial_access_token_hash_unique ON _auth_oauth_initial_access_token USING btree (app_id, token_hash);
```

## Registration Endpoint

```
POST /oauth2/register
```

### Request

```
POST /oauth2/register HTTP/1.1
Host: myapp.authgear.cloud
Content-Type: application/json
Authorization: Bearer <initial_access_token>   (omit when initial_access_token_required: false)
```

See [Accepted Client Metadata](#accepted-client-metadata) for the full list of request body fields.

### Response

**201 Created** on success. A successful registration is recorded in the [audit log](#audit-log); a failed one is not.

```json
{
  "client_id": "dcrc_Xf2kLmNpQrStUvWx",
  "client_id_issued_at": 1700000000,
  "client_name": "PR #123 preview",
  "redirect_uris": ["https://pr-123.preview.example.com/callback"],
  "grant_types": ["authorization_code", "refresh_token"],
  "response_types": ["code"],
  "application_type": "web"
}
```

- `client_secret` is not issued in this version, and neither is `client_secret_expires_at`. Confidential clients are not supported via DCR — every DCR-registered client is public and uses PKCE. Should confidential clients ever be supported, the once-only-return and `client_secret_expires_at: 0` semantics of RFC 7591 §3.2.1 would apply then; they do not apply to anything in this version.

### Errors

Error responses follow [RFC 7591 §3.2.2](https://www.rfc-editor.org/rfc/rfc7591#section-3.2.2):

```json
{
  "error": "invalid_client_metadata",
  "error_description": "redirect_uris must use HTTPS. See https://docs.authgear.com/..."
}
```

| `error` value | HTTP status | Meaning |
|---|---|---|
| `invalid_redirect_uri` | 400 | One or more `redirect_uris` are invalid (e.g. plain `http://` for non-localhost) |
| `invalid_client_metadata` | 400 | Other metadata validation failure — see table below |
| `invalid_initial_access_token` | 401 | IAT is missing, expired, or not recognized |
| `access_denied` | 403 | Registration is not permitted (e.g. DCR is disabled, a first-party IAT is required but a third-party IAT or no IAT was presented, or the project's [client limit](#client-limit) has been reached) |
| `x_rate_limited` | 429 | Too many registration attempts from this IP, or too many for this project — see [Rate Limits](#rate-limits) |

**`invalid_client_metadata` causes:**

| Condition | Example |
|---|---|
| `redirect_uris` is missing | omitted from request body |
| `redirect_uris` contains a URI with a fragment component | `https://example.com/callback#section` |
| `token_endpoint_auth_method` is provided and is not `none` | `token_endpoint_auth_method=client_secret_post` |
| `grant_types` contains an unsupported value | `grant_types=["implicit"]` |
| `response_types` contains an unsupported value | `response_types=["token"]` |
| `response_types` is inconsistent with `grant_types` | `grant_types=["refresh_token"]` + `response_types=["code"]` without `authorization_code` |
| `logo_uri`, `client_uri`, `tos_uri`, or `policy_uri` is not `https://` | `logo_uri=http://example.com/logo.png` |

## Accepted Client Metadata

The following fields are accepted in the registration request body. All other client configuration fields require direct admin access through the portal.

### `client_name` (optional)

Human-readable name for the client, displayed on the consent screen and in the portal. When omitted, Authgear generates a default name from the `client_id` (e.g. `Client dcrc_Xf2kLmNpQrStUvWx`).

### `redirect_uris` (required)

Array of redirect URIs the client will use in authorization code flows. Each URI must be:

- An `https://` URI, **or**
- A custom URI scheme (e.g., `com.example.app://callback`) for native apps.

Plain `http://` URIs are rejected except for `http://localhost` (loopback), which is allowed for native app development.

Each URI must be an absolute URI (per RFC 3986 §4.3) and must not contain a fragment component (`#`).

If `redirect_uris` is omitted, the server returns `invalid_client_metadata`.

### `grant_types` (optional)

Array of grant types the client is allowed to use. Accepted values:

| Value | Meaning |
|---|---|
| `authorization_code` | Standard OAuth 2.0 authorization code flow |
| `refresh_token` | Allows the client to exchange a refresh token for new access tokens |

Default: `["authorization_code", "refresh_token"]`.

### `response_types` (optional)

Array of response types. Must be consistent with `grant_types`. The only accepted value is `code`, which must be paired with the `authorization_code` grant type. Requesting `response_types=["code"]` without `authorization_code` in `grant_types`, or vice versa, returns `invalid_client_metadata`.

Default: `["code"]`.

### `application_type` (optional)

Controls the client's technical profile (redirect URI rules, PKCE requirements). Authgear accepts the two standard OIDC DCR values:

| Value | IAT type required | Consent screen | `kind` | Redirect URI validation |
|---|---|---|---|---|
| `web` (default) | none or `iat_tp_` | Yes | `THIRD_PARTY` | Must use `https://`; `localhost` not allowed |
| `native` | none or `iat_tp_` | Yes | `THIRD_PARTY` | Custom URI scheme or `http://localhost` |
| `web` | `iat_fp_` | No | `FIRST_PARTY` | Must use `https://`; `localhost` not allowed |
| `native` | `iat_fp_` | No | `FIRST_PARTY` | Custom URI scheme or `http://localhost` |

Default: `web`.

The IAT type — not `application_type` — determines whether the registered client is first-party or third-party. `application_type` describes only the technical profile (redirect URI rules, etc.).

### `token_endpoint_auth_method` (optional)

The only accepted value is `none`. Every DCR-registered client is public and uses PKCE — Authgear never issues a `client_secret` via DCR — so `none` is simply a client explicitly stating what's already true, and is accepted. Any other value (e.g. `client_secret_post`, `client_secret_basic`) returns `invalid_client_metadata`, since Authgear has no client secret to authenticate with. Omitting the field entirely is equivalent to sending `none`.

### `logo_uri` (optional)

URL of the client's logo image, shown on the consent screen. Must be an `https://` URL.

### `client_uri` (optional)

URL of the client's home page. Must be an `https://` URL.

### `tos_uri` (optional)

URL of the client's Terms of Service page, shown on the consent screen. Must be an `https://` URL.

### `policy_uri` (optional)

URL of the client's Privacy Policy page, shown on the consent screen. Must be an `https://` URL.

## Client ID Format

DCR-registered clients and IATs use the following prefixed formats:

| Token | Prefix | Entropy | Example |
|---|---|---|---|
| `client_id` | `dcrc_` | 22 chars URL-safe base64 (16 bytes) | `dcrc_Xf2kLmNpQrStUvWx` |
| Third-party IAT | `iat_tp_` | 22 chars URL-safe base64 (16 bytes) | `iat_tp_Xf2kLmNpQrStUvWx` |
| First-party IAT | `iat_fp_` | 22 chars URL-safe base64 (16 bytes) | `iat_fp_Xf2kLmNpQrStUvWx` |

`dcrc` = **D**ynamic **C**lient **R**egistration **C**lient. The prefix distinguishes DCR clients from statically configured clients in `authgear.yaml`. The `iat_tp_` / `iat_fp_` prefixes make the privilege level of an IAT immediately visible — a leaked `iat_fp_` token has significantly higher blast radius than a leaked `iat_tp_` token.

## Storage Architecture

DCR-registered clients are stored in the **database**, not in `authgear.yaml`. Authgear loads both static clients (from `authgear.yaml`) and DCR clients (from the database) at request time, merging them into a unified client list.

The runtime behavior of a DCR client (authorization code flow, token endpoint, consent screen, etc.) is identical to that of a static client with the same `kind` and `application_type`.

DCR client secrets are stored hashed in the database.

## Security Considerations

### Access Token Audience Binding

By default, all Authgear access tokens share `aud = [<project_endpoint>]`. A resource server that only validates `aud` cannot distinguish tokens intended for different services — this is the **audience confusion** risk.

Authgear mitigates this via RFC 8707 resource indicators. Resource owners pre-register their API as a Resource in the portal and associate it with allowed clients. When a client requests a token with `resource=<uri>`, the issued access token includes that URI in `aud`, and the resource server can enforce `aud` contains its own URI.

DCR-registered clients, being dynamic third-party clients, support resource indicators via API Resources registered in the portal. Only Resources with `access_policy.allow_dynamic_third_party_client_access: true` are accessible, and only Scopes with `access_policy.allow_dynamic_third_party_client_access: true` may be requested — this policy is dynamic-only by design, so a static `third_party_app` client cannot use it even if the flag is set. All other project resources and scopes remain inaccessible, preventing audience confusion against first-party clients.

A DCR client that requests no `resource` parameter receives an opaque access token, scoped to the userinfo endpoint only — never a JWT with the project endpoint as `aud`, which is reserved for first-party clients. See [Access Token Audience Binding — How It Works](./access-token-audience-binding.md#how-it-works).

The admin configures the access policy once per Resource/Scope in the portal. Individual DCR clients then autonomously use `resource=<uri>` in their authorization requests without any further admin action per client. See [API Resources and Scopes](./api-resource.md#access-policy) and [Access Token Audience Binding](./access-token-audience-binding.md) for the full design.

## Admin API

The portal displays registered clients by querying the Admin GraphQL API. Client creation is done by calling `POST /oauth2/register` directly with an IAT (when required). Client self-service management (the client reading/updating/deleting its own registration via a Registration Access Token) is deferred to RFC 7592 — see [Future Works](#future-works). Admin-initiated deletion is a separate, already-defined capability — see [New mutation](#new-mutation) below.

### IAT management

Creating and revoking an IAT are both recorded in the [audit log](#audit-log). The plaintext token is never recorded.

```graphql
type Query {
  """Returns all active (non-expired) Initial Access Tokens for the project."""
  initialAccessTokens: [InitialAccessToken!]!
}

type Mutation {
  """Creates an opaque Initial Access Token for use with POST /oauth2/register."""
  createInitialAccessToken(input: CreateInitialAccessTokenInput!): CreateInitialAccessTokenPayload!

  """Revokes an Initial Access Token so it can no longer be used for registration."""
  revokeInitialAccessToken(input: RevokeInitialAccessTokenInput!): RevokeInitialAccessTokenPayload!
}

enum InitialAccessTokenType {
  """
  Can register web and native clients as third-party (consent screen shown).
  Token prefix: iat_tp_
  """
  THIRD_PARTY

  """
  Can register web and native clients as first-party (consent screen bypassed).
  Token prefix: iat_fp_
  High privilege — protect this token like the Admin API private key.
  """
  FIRST_PARTY
}

type InitialAccessToken implements Node {
  id: ID!
  createdAt: DateTime!
  expiresAt: DateTime!
  type: InitialAccessTokenType!
}

input CreateInitialAccessTokenInput {
  """
  Token lifetime in seconds. If omitted, a server default is used (e.g. 3600).
  """
  expiresIn: Int
  """
  Defaults to THIRD_PARTY. Specify FIRST_PARTY only when registering
  first-party clients is required (e.g. CI/CD pipelines). The issued token
  will carry the iat_fp_ prefix as a visible indicator of its elevated privilege.
  """
  type: InitialAccessTokenType
}

type CreateInitialAccessTokenPayload {
  """
  The opaque IAT value. Returned ONCE only — not recoverable after this response.
  Store it securely and pass it as Authorization: Bearer <token> to POST /oauth2/register.
  """
  token: String!
  initialAccessToken: InitialAccessToken!
}

input RevokeInitialAccessTokenInput {
  id: ID!
}

type RevokeInitialAccessTokenPayload {
  ok: Boolean
}
```

### Client model

DCR-registered clients are represented using the unified `OAuthClient` model defined in [Client Model](./client.md). See that document for the full type definition and the mapping from DCR registration fields to model fields.

### New query

```graphql
extend type Query {
  """
  Returns clients that exist outside authgear.yaml: DCR-registered clients and,
  when enabled, CIMD-resolved clients (see cimd.md). Distinguish the two via
  `source` on OAuthClient (DCR vs CIMD). Static clients are managed via
  authgear.yaml and are not returned here.
  """
  dynamicClients(
    first: Int
    after: String
    last: Int
    before: String
  ): OAuthClientConnection!
}

type OAuthClientConnection {
  edges: [OAuthClientEdge]
  pageInfo: PageInfo!
  totalCount: Int
}

type OAuthClientEdge {
  node: OAuthClient
  cursor: String!
}
```

The Connection and Edge types are named after the node type (`OAuthClient`), and their nullability matches every other Connection in the Admin API, because they are produced by the same shared relay helper rather than hand-written for this query. "Dynamic" describes which clients the query returns, not a distinct GraphQL type.

### New mutation

```graphql
extend type Mutation {
  """
  Deletes a DCR-registered or CIMD-resolved client: removes its persisted
  record and revokes all outstanding authorizations and tokens issued to it,
  for every user. Frees one slot against the client's corresponding limit
  (see Client Limit in dcr.md / cimd.md). Not applicable to static clients —
  manage those via authgear.yaml instead.
  """
  deleteDynamicClient(input: DeleteDynamicClientInput!): DeleteDynamicClientPayload!
}

input DeleteDynamicClientInput {
  clientID: String!
}

type DeleteDynamicClientPayload {
  ok: Boolean
}
```

> **Implementation status:** the token and authorization revocation described above is not yet implemented. The first implementation of this mutation removes the persisted client record only, which frees the client-limit slot and stops any new authorization, but leaves already-issued access tokens valid until they expire and already-issued refresh tokens usable. Until revocation lands, deleting a DCR client is not a way to cut off a client that is actively misbehaving.

For a DCR client, deletion is permanent: the same `client_id` never reappears unless a new `POST /oauth2/register` call creates it again. For a CIMD client, deletion only evicts the current persisted record — the same `client_id` URL can produce a new record on its very next successful resolution, since nothing prevents a caller from presenting that URL again. This mutation frees a slot immediately in both cases, but for CIMD it is not a durable ban; see [cimd.md — Domain Trust](./cimd.md#domain-trust) for the closest thing to one.

Deleting a client is recorded in the [audit log](#audit-log). Because the client record itself is removed, that audit entry is the only remaining record of what was deleted.

## Audit Log

Every DCR lifecycle action is recorded in the [audit log](./audit-log.md).

A successful registration emits [`oauth.client.registered`](./event.md#oauthclientregistered). A rejected registration attempt emits nothing: under open registration `POST /oauth2/register` is unauthenticated, so logging failures would let any caller write audit entries at will. Attempts rejected by the [rate limits](#rate-limits) are already covered by `rate_limit.blocked`.

Creating and revoking an IAT, and deleting a dynamic client, are recorded the same way as every other Admin API mutation.

**No token value is ever written to the audit log** — not the plaintext IAT, not a client secret. An IAT is returned exactly once at creation and stored only as a hash ([IAT storage](#iat-storage)); an audit entry containing it would be a second, admin-readable copy, retained for the audit log's retention period, of a credential this design treats as unrecoverable. The audit entry identifies an IAT by its ID and type only.

## Future Works

### Per client config update

Currently `default_client_config` applies a single set of token lifetimes to all DCR clients. Providers such as Keycloak support per-client config overrides configured by an admin after registration. This will be supported via the Admin API or portal once per-client management of DCR clients is implemented.

### Client management (RFC 7592)

DCR clients cannot currently be read or updated after registration, and cannot delete themselves (admin-initiated deletion is already covered by [`deleteDynamicClient`](#new-mutation)). RFC 7592 is the planned mechanism for client self-service management — see below.

### RFC 7592 — Client Registration Management

[RFC 7592](https://www.rfc-editor.org/rfc/rfc7592) defines three endpoints for managing a registered client after initial registration, each protected by a per-client **Registration Access Token (RAT)**:

- `GET /oauth2/register/{client_id}` — read current client metadata
- `PUT /oauth2/register/{client_id}` — replace mutable metadata fields
- `DELETE /oauth2/register/{client_id}` — delete the client and revoke all its tokens

When RFC 7592 is implemented, the registration response (`POST /oauth2/register`) will also include:

```json
{
  "registration_access_token": "rat_Yz9mAbCdEfGhIjKlMnOpQrStUvWxYz",
  "registration_client_uri": "https://myapp.authgear.cloud/oauth2/register/dcrc_Xf2kLmNpQrStUvWx"
}
```

The RAT will use the prefix `rat_` (32 chars URL-safe base64, 24 bytes entropy) and be stored hashed in the database. It will be issued once and not recoverable if lost.

