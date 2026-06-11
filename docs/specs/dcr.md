# Dynamic Client Registration (DCR)

Authgear supports Dynamic Client Registration as defined by:

- [RFC 7591 — OAuth 2.0 Dynamic Client Registration Protocol](https://www.rfc-editor.org/rfc/rfc7591)
- [OpenID Connect Dynamic Client Registration 1.0](https://openid.net/specs/openid-connect-registration-1_0.html)

## Table of Contents

- [Glossary](#glossary)
- [Use Cases](#use-cases)
- [Configuration](#configuration)
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
    expiresAt
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

> Client deletion is not supported in this version. See [Future Works](#future-works).

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
3. On the Resource and on each scope that MCP clients should be able to request, set `access_policy.allow_third_party_client_access: true`.

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
- `oauth.dynamic_client_registration.initial_access_token_required`: Optional. Boolean. Default `true`. When `true`, registration requires a valid IAT in the `Authorization: Bearer` header; all `application_type` values are accepted. When `false`, open registration is permitted but only `application_type: web` and `application_type: native` are accepted.

- `oauth.dynamic_client_registration.default_client_config`: Optional. Object. The default client config applied to all DCR-registered clients. Useful when stricter settings are needed for the DCR cohort. Per-client overrides are not yet supported; see [Future Works](#future-works). Supports a subset of the fields defined in [Custom Client Metadata](./oidc.md#custom-client-metadata): `access_token_lifetime_seconds`, `refresh_token_lifetime_seconds`, `refresh_token_idle_timeout_enabled`, `refresh_token_idle_timeout_seconds`.

> **Note:** Resource access for third-party clients is configured via the portal, not `authgear.yaml`. Resources and Scopes with `access_policy.allow_third_party_client_access: true` are accessible to all third-party clients, including DCR-registered ones. See [API Resources and Scopes](./api-resource.md#access-policy).

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

An IAT is an **opaque** token issued by the Admin API (see [Admin API — IAT mutation](#new-mutation-createinitialaccesstoken)). It is passed as `Authorization: Bearer <iat>` to the registration endpoint.

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

**201 Created** on success.

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

- `client_secret` is not issued in this version. Confidential clients are not supported via DCR.
- `client_secret_expires_at: 0` means non-expiring (per RFC 7591 §3.2.1).
- `client_secret` is returned **once only** and is not recoverable afterwards. The caller must store it securely.

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
| `access_denied` | 403 | Registration is not permitted (e.g. DCR is disabled, or a first-party IAT is required but a third-party IAT or no IAT was presented) |

**`invalid_client_metadata` causes:**

| Condition | Example |
|---|---|
| `redirect_uris` is missing | omitted from request body |
| `redirect_uris` contains a URI with a fragment component | `https://example.com/callback#section` |
| `token_endpoint_auth_method` is provided (field not accepted) | `token_endpoint_auth_method=client_secret_post` |
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

DCR-registered clients, being third-party clients, support resource indicators via API Resources registered in the portal. Only Resources with `access_policy.allow_third_party_client_access: true` are accessible to third-party clients, and only Scopes with `access_policy.allow_third_party_client_access: true` may be requested. All other project resources and scopes remain inaccessible, preventing audience confusion against first-party clients.

The admin configures the access policy once per Resource/Scope in the portal. Individual DCR clients then autonomously use `resource=<uri>` in their authorization requests without any further admin action per client. See [API Resources and Scopes](./api-resource.md#access-policy) and [Access Token Audience Binding](./access-token-audience-binding.md) for the full design.

## Admin API

The portal displays registered clients by querying the Admin GraphQL API. Client creation is done by calling `POST /oauth2/register` directly with an IAT (when required); client management (read, update, delete) is deferred to RFC 7592.

### IAT management

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

### New GraphQL type

```graphql
enum OAuthClientKind {
  FIRST_PARTY
  THIRD_PARTY
}

type OAuthClient implements Node {
  id: ID!
  clientID: String!
  clientName: String!
  """The OIDC application_type value: "web" or "native"."""
  applicationType: String!
  """Whether this client is first-party or third-party."""
  kind: OAuthClientKind!

  # ISO 8601 timestamp of when the client was registered via DCR.
  # Null for statically configured clients.
  registeredAt: DateTime

  redirectURIs: [String!]!
  grantTypes: [String!]!
  responseTypes: [String!]!
}
```

### New query

```graphql
extend type Query {
  # Returns all clients: both static (authgear.yaml) and DCR-registered.
  oauthClients: [OAuthClient!]!
}
```

## Future Works

### Per client config update

Currently `default_client_config` applies a single set of token lifetimes to all DCR clients. Providers such as Keycloak support per-client config overrides configured by an admin after registration. This will be supported via the Admin API or portal once per-client management of DCR clients is implemented.

### Client management (RFC 7592)

DCR clients cannot currently be read, updated, or deleted after registration. RFC 7592 is the planned mechanism for all post-registration client management — see below.

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

