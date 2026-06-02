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
- [Future Works](#future-works)

## Glossary

**Dynamic Client Registration (DCR)** — the process by which an OAuth client registers itself programmatically with an Authorization Server at runtime, rather than being statically configured in `authgear.yaml`.

**Initial Access Token (IAT)** — a short-lived token presented to the registration endpoint that authorizes the caller to register a new client. Required when `initial_access_token_required: true` (the default). An IAT is a JWT signed with the Admin API auth key.

## Use Cases

### UC1. Ephemeral clients for CI / pull-request preview environments

A CI system holds the Admin API private key for a project. For each pull request, the CI registers a new first-party client scoped to that PR's redirect URI.

An IAT is required because `spa`, `native`, and `traditional_webapp` are first-party client types — they bypass the consent screen and must only be registered by an authorized administrator.

**Required configuration:**

```yaml
oauth:
  dynamic_client_registration:
    enabled: true
    initial_access_token_required: true   # default; explicitly set for clarity
```

No `allowed_resources` or `default_client_config` override is needed — CI clients use the project-level token lifetimes and do not require resource indicator support.

**Step 1 — Mint an IAT**

Sign a JWT with the Admin API private key:

```json
{
  "iss": "my-ci-pipeline",
  "aud": "https://myapp.authgear.cloud/oauth2/register",
  "iat": 1700000000,
  "exp": 1700003600,
  "scope": "dcr"
}
```

**Step 2 — Register the client**

```
POST /oauth2/register HTTP/1.1
Host: myapp.authgear.cloud
Content-Type: application/json
Authorization: Bearer <iat>

{
  "client_name": "PR #123 preview",
  "redirect_uris": ["https://pr-123.preview.example.com/callback"],
  "application_type": "spa"
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
  "application_type": "spa",
  "token_endpoint_auth_method": "none"
}
```

**Step 3 — Use the client in the authorization code flow**

The PR preview app uses `client_id=dcrc_Xf2kLmNpQrStUvWx` as a normal SPA client for the lifetime of the PR.

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
    allowed_resources:
      - uri: "https://mcp-server.example.com"
        scopes:
          - "read:tools"
          - "execute:tools"
```

**Admin setup (once)**

Apply the configuration above. No further per-client admin action is required — any MCP client can self-register and immediately use the declared resources.

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
  "redirect_uris": ["https://mcp-client.example.com/callback"],
  "token_endpoint_auth_method": "none"
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
  "application_type": "third_party_app",
  "token_endpoint_auth_method": "none"
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

The issued access token has `aud: ["https://myapp.authgear.cloud", "https://mcp-server.example.com"]`. The MCP server validates `aud` contains its own URI.

## Configuration

```yaml
oauth:
  dynamic_client_registration:
    enabled: true
    initial_access_token_required: true
    allowed_resources:
      - uri: "https://api.example.com"
        scopes:
          - "read:data"
          - "write:data"
    default_client_config:
      access_token_lifetime_seconds: 1800
      refresh_token_lifetime_seconds: 2592000
      refresh_token_idle_timeout_enabled: true
      refresh_token_idle_timeout_seconds: 1209600
```

- `oauth.dynamic_client_registration.enabled`: Optional. Boolean. Default `false`. Enables `POST /oauth2/register`.
- `oauth.dynamic_client_registration.initial_access_token_required`: Optional. Boolean. Default `true`. When `true`, registration requires a valid IAT in the `Authorization: Bearer` header; all `application_type` values are accepted. When `false`, open registration is permitted but only `application_type: third_party_app` is accepted.

- `oauth.dynamic_client_registration.allowed_resources`: Optional. List of objects. Default empty. Allow-list of resource server URIs that DCR clients may request via the `resource` parameter (RFC 8707). When empty, DCR clients cannot use resource indicators.
- `oauth.dynamic_client_registration.allowed_resources[].uri`: Required. String. Absolute `https://` URI of the resource server. Must not be prefixed by the Authgear project endpoint.
- `oauth.dynamic_client_registration.allowed_resources[].scopes`: Optional. List of strings. Default empty. Scopes DCR clients may request for this resource. Requesting an unlisted scope returns `invalid_scope`.

- `oauth.dynamic_client_registration.default_client_config`: Optional. Object. The default client config applied to all DCR-registered clients. Useful when stricter settings are needed for the DCR cohort. Per-client overrides are not yet supported; see [Future Works](#future-works). Supports a subset of the fields defined in [Custom Client Metadata](./oidc.md#custom-client-metadata): `access_token_lifetime_seconds`, `refresh_token_lifetime_seconds`, `refresh_token_idle_timeout_enabled`, `refresh_token_idle_timeout_seconds`.

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

An IAT is a JWT signed with the private key material corresponding to the `admin-api.auth` JWK set in `authgear.secrets.yaml` (same mechanism already used for Admin API JWT authentication). RS256 and ES256 are both accepted.

### Required JWT claims

| Claim | Value |
|---|---|
| `iss` | Any non-empty string identifying the issuer |
| `aud` | Exact registration endpoint URL: `<authgear_endpoint>/oauth2/register` |
| `iat` | Unix timestamp of issuance |
| `exp` | Unix timestamp of expiry (must be in the future; recommended ≤ 1 hour from `iat`) |
| `scope` | Must contain the value `dcr` (space-separated if multiple scopes) |

Authgear validates:

1. Signature against the `admin-api.auth` public key set.
2. `aud` equals `<authgear_endpoint>/oauth2/register` exactly.
3. `exp` has not elapsed.
4. `scope` contains `dcr`.

An IAT is single-use. Authgear rejects a replayed IAT whose `jti` (if present) has already been seen, or one that was issued before a key rotation.

### Minting an IAT (example)

```json
{
  "iss": "my-ci-pipeline",
  "aud": "https://myapp.authgear.cloud/oauth2/register",
  "iat": 1700000000,
  "exp": 1700003600,
  "scope": "dcr"
}
```

Sign this JWT with the Admin API private key (RS256 or ES256) and pass it as `Authorization: Bearer <iat>` to the registration endpoint.

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
  "client_secret": "s3cr3t...",
  "client_id_issued_at": 1700000000,
  "client_secret_expires_at": 0,
  "client_name": "PR #123 preview",
  "redirect_uris": ["https://pr-123.preview.example.com/callback"],
  "grant_types": ["authorization_code", "refresh_token"],
  "response_types": ["code"],
  "application_type": "third_party_app",
  "token_endpoint_auth_method": "client_secret_post"
}
```

- `client_secret` is present only when the effective auth method is `client_secret_post`. It is absent when `token_endpoint_auth_method` is `none` or the type is inherently public (`spa`, `native`).
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
| `invalid_initial_access_token` | 401 | IAT is missing, expired, has wrong `aud`, or has wrong `scope` |
| `access_denied` | 403 | Registration is not permitted (e.g. DCR is disabled, or open registration attempted with a non-`third_party_app` type) |

**`invalid_client_metadata` causes:**

| Condition | Example |
|---|---|
| `redirect_uris` is missing | omitted from request body |
| `redirect_uris` contains a URI with a fragment component | `https://example.com/callback#section` |
| `token_endpoint_auth_method` conflicts with `application_type` | `application_type=spa` + `token_endpoint_auth_method=client_secret_post` |
| `token_endpoint_auth_method` is an unsupported value | `token_endpoint_auth_method=client_secret_basic` |
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

Controls the security profile applied to the client. The OIDC Dynamic Registration spec defines `web` and `native`; Authgear extends this with additional values:

| Value | First/Third party | Consent screen | Default `token_endpoint_auth_method` | `token_endpoint_auth_method` overridable |
|---|---|---|---|---|
| `spa` | First-party | No | `none` | No |
| `native` | First-party | No | `none` | No |
| `traditional_webapp` | First-party | No | `client_secret_post` | No |
| `confidential` | First-party | No | `client_secret_post` | No |
| `third_party_app` | Third-party | Yes | `client_secret_post` | Yes (`none` or `client_secret_post`) |

Default: `third_party_app`.

**IAT requirement by type:** `spa`, `native`, `traditional_webapp`, and `confidential` are first-party types — they bypass the consent screen and may only be registered with a valid IAT. When `initial_access_token_required: false` (open registration), only `third_party_app` is accepted.

Stored internally as `x_application_type` in the client configuration.

### `token_endpoint_auth_method` (optional)

Declares how the client authenticates at the token endpoint.

| Value | Meaning |
|---|---|
| `none` | Public client — no `client_secret` issued; PKCE is required |
| `client_secret_post` | Confidential client — `client_secret` issued, sent as a POST body parameter |

For most `application_type` values this field is fixed (see table above). Only `third_party_app` allows a choice. Passing a value that conflicts with the `application_type` returns `invalid_client_metadata`.

When omitted, the default for the given `application_type` applies.

`client_secret_basic` (the RFC 7591 default) is intentionally not supported. Modern OAuth 2.0 encourages PKCE with `token_endpoint_auth_method: none`, making `client_secret_basic` unnecessary for the DCR use cases Authgear targets.

### `logo_uri` (optional)

URL of the client's logo image, shown on the consent screen. Must be an `https://` URL.

### `client_uri` (optional)

URL of the client's home page. Must be an `https://` URL.

### `tos_uri` (optional)

URL of the client's Terms of Service page, shown on the consent screen. Must be an `https://` URL.

### `policy_uri` (optional)

URL of the client's Privacy Policy page, shown on the consent screen. Must be an `https://` URL.

## Client ID Format

DCR-registered clients use the prefixed ID format:

| Field | Prefix | Entropy | Example |
|---|---|---|---|
| `client_id` | `dcrc_` | 22 chars URL-safe base64 (16 bytes) | `dcrc_Xf2kLmNpQrStUvWx` |

`dcrc` = **D**ynamic **C**lient **R**egistration **C**lient. The prefix distinguishes DCR clients from statically configured clients in `authgear.yaml`.

## Storage Architecture

DCR-registered clients are stored in the **database**, not in `authgear.yaml`. Authgear loads both static clients (from `authgear.yaml`) and DCR clients (from the database) at request time, merging them into a unified client list.

The runtime behavior of a DCR client (authorization code flow, token endpoint, consent screen, etc.) is identical to that of a static client with the same `application_type` and metadata.

DCR client secrets are stored hashed in the database.

## Security Considerations

### Access Token Audience Binding

By default, all Authgear access tokens share `aud = [<project_endpoint>]`. A resource server that only validates `aud` cannot distinguish tokens intended for different services — this is the **audience confusion** risk.

Authgear mitigates this via RFC 8707 resource indicators. Resource owners pre-register their API as a Resource in the portal and associate it with allowed clients. When a client requests a token with `resource=<uri>`, the issued access token includes that URI in `aud`, and the resource server can enforce `aud` contains its own URI.

DCR-registered clients support resource indicators via the `allowed_resources` list in `authgear.yaml`. Only resources explicitly listed there — with their permitted scopes — are accessible to DCR clients. All other project resources remain inaccessible, preventing audience confusion against first-party APIs.

The admin configures `allowed_resources` once (e.g. for the MCP server). Individual DCR clients then autonomously use `resource=<uri>` in their authorization requests without any further admin action. See [Access Token Audience Binding](./access-token-audience-binding.md) for the full design.

## Admin API

The portal displays registered clients by querying the Admin GraphQL API. Client creation is done by calling `POST /oauth2/register` directly with an IAT; client management (read, update, delete) is deferred to RFC 7592.

### New GraphQL type

```graphql
type OAuthClient implements Node {
  id: ID!
  clientID: String!
  clientName: String!
  applicationType: String!

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

