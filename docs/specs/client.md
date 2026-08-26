# Client Model

An Authgear **client** is any OAuth 2.0 / OIDC application that interacts with the authorization server. Clients may originate from three sources:

1. **Static clients** — declared in `authgear.yaml` under `oauth.clients`. Changes require a configuration deploy.
2. **DCR clients** — registered at runtime via [Dynamic Client Registration (DCR)](./dcr.md). Stored in the database.
3. **Client ID Metadata Document (CIMD) clients** *(proposed, see [cimd.md](./cimd.md))* — resolved by fetching a JSON document from a `client_id` URL, then persisted as a single shared record per `client_id`, refreshed on each refetch. Not registered like a DCR client; there is no registration event, only resolution.

Sources 2 and 3 are collectively referred to as **dynamic clients**: clients that exist outside `authgear.yaml` and require no per-client admin action. That is the sense in which [`dynamicClients`](./dcr.md#new-query) and [`deleteDynamicClient`](./dcr.md#new-mutation) use the word — both cover DCR and CIMD clients, and neither covers static ones.

This document defines the unified client model and presents it as a GraphQL type. It then describes how each source maps into the model.

## Table of Contents

- [GraphQL Type](#graphql-type)
- [Mapping from Static Config](#mapping-from-static-config)
- [Mapping from DCR](#mapping-from-dcr)
- [Access Token Behavior by Client Kind](#access-token-behavior-by-client-kind)

## GraphQL Type

```graphql
type OAuthClient {
  """Unique, immutable client identifier."""
  clientID: String!

  """Which of the three sources this client came from."""
  source: OAuthClientSource!

  """
  Whether the client is first-party or third-party. Third-party clients always
  show a consent screen before issuing tokens; first-party clients skip it.
  """
  kind: OAuthClientKind!

  """
  Whether the client is confidential. Confidential clients authenticate at the
  token endpoint using a client_secret. Public clients use PKCE instead.
  """
  isConfidential: Boolean!

  """
  Whether the client is a service client (M2M). When true, the client acts as
  its own principal — API resource assignments grant access to the client itself.
  When false, the client acts on behalf of a user — API resource assignments
  grant user-delegated access.
  """
  isServiceClient: Boolean!

  """
  OIDC application_type of this client. Only set for DCR clients; null for
  static config clients. Stored to support future RFC 7592 management — when a
  DCR client updates its redirect_uris, this value is used to re-validate them.
  "web": redirect URIs must use https://; localhost is not allowed.
  "native": custom URI schemes and http://localhost are allowed.
  """
  applicationType: String

  """
  Human-readable display name shown in the portal.
  For static clients this is the `name` field.
  For DCR clients this is client_name (auto-generated as "Client <clientID>" if omitted).
  """
  name: String!

  """
  OIDC client_name presented on the consent screen.
  Null for static clients that do not set client_name (spa, traditional_webapp, native, m2m),
  and for DCR/CIMD clients that omit client_name at registration -- the generated
  "Client <clientID>" fallback is never persisted as client_name itself, only
  reflected in `name` above.
  """
  clientName: String

  """URI of the client's home page. https:// only."""
  clientURI: String

  """URI of the client's logo image. Shown on the consent screen. https:// only."""
  logoURI: String

  """URI of the client's Terms of Service page. https:// only."""
  tosURI: String

  """URI of the client's Privacy Policy page. https:// only."""
  policyURI: String

  """Redirect URIs the client is allowed to use. Empty for M2M clients."""
  redirectURIs: [String!]!

  """
  Post-logout redirect URIs. Non-empty when the client requires back-channel
  logout notification (i.e. x_application_type: traditional_webapp in config).
  Always empty for DCR clients.
  """
  postLogoutRedirectURIs: [String!]!

  """Grant types the client is permitted to use."""
  grantTypes: [String!]!

  """Response types the client is permitted to request."""
  responseTypes: [String!]!

  """Access token lifetime in seconds."""
  accessTokenLifetimeSeconds: Int!

  """Refresh token lifetime in seconds."""
  refreshTokenLifetimeSeconds: Int!

  """Whether the refresh token idle timeout is active."""
  refreshTokenIdleTimeoutEnabled: Boolean!

  """Idle timeout for refresh tokens in seconds."""
  refreshTokenIdleTimeoutSeconds: Int!

  """Whether refresh token rotation is enabled."""
  refreshTokenRotationEnabled: Boolean!

  """
  Whether the server issues JWT access tokens instead of opaque tokens.
  Always true for M2M clients.
  """
  issueJWTAccessToken: Boolean!

  """
  Maximum number of concurrent sessions. 0 = unlimited, 1 = at most one.
  Always 0 for DCR clients.
  """
  maxConcurrentSession: Int!

  """
  URI of a custom auth UI. When set, the auth server responds HTTP 200 instead
  of redirecting. Static clients only; always null for DCR clients.
  """
  customUIURI: String

  """Whether App2App is enabled. Static clients only; always false for DCR clients."""
  app2appEnabled: Boolean!

  """Static clients only; always false for DCR clients."""
  app2appInsecureDeviceKeyBindingEnabled: Boolean!

  """Whether DPoP sender-constraint is disabled. Static clients only; always false for DCR clients."""
  dpopDisabled: Boolean!

  """Allowed authentication flows. Static clients only; always null for DCR clients."""
  authenticationFlowAllowlist: AuthenticationFlowAllowlist

  """Whether the pre-authenticated URL feature is enabled. Static clients only; always false for DCR clients."""
  preAuthenticatedURLEnabled: Boolean!

  """Allowed origins for pre-authenticated URL. Static clients only; always empty for DCR clients."""
  preAuthenticatedURLAllowedOrigins: [String!]!

  """
  When true, the project logo is replaced with logoURI on the auth UI.
  Static clients only; always false for DCR clients.
  """
  replaceProjectLogoWithLogoURI: Boolean!

  """RFC 3339 timestamp of DCR registration. Null for static and CIMD clients."""
  registeredAt: DateTime

  """
  RFC 3339 timestamp of the most recent successful fetch of the client's
  Client ID Metadata Document, refreshed on each refetch (see cimd.md). Null
  for static and DCR clients, which are never fetched from an external URL.
  """
  lastFetchedAt: DateTime
}

type AuthenticationFlowAllowlist {
  groups: [AuthenticationFlowAllowlistGroup!]!
  flows:  [AuthenticationFlowAllowlistFlow!]!
}

type AuthenticationFlowAllowlistGroup {
  name: String!
}

type AuthenticationFlowAllowlistFlow {
  """One of: signup, promote, login, signup_login, reauth, account_recovery."""
  type: String!
  name: String!
}

enum OAuthClientKind {
  """Operated by a project collaborator. Consent screen is not shown."""
  FIRST_PARTY

  """Operated by an external developer. Consent screen is always shown."""
  THIRD_PARTY
}

enum OAuthClientSource {
  """Declared in authgear.yaml."""
  STATIC

  """Registered via Dynamic Client Registration. See dcr.md."""
  DCR

  """Resolved from a Client ID Metadata Document. See cimd.md."""
  CIMD
}
```

## Mapping from Static Config

`source` is always `STATIC` for every client in this section. The config field `x_application_type` is a shorthand that encodes `isThirdParty`, `isConfidential`, and `applicationType` together. The table below shows the decomposition. The `spa` and `traditional_webapp` values both map to the same three fields; the distinction between them is preserved in `postLogoutRedirectURIs` (non-empty for `traditional_webapp`).

| `x_application_type` | `kind` | `isConfidential` | `isServiceClient` | `applicationType` |
|---|---|---|---|---|
| `spa` | `FIRST_PARTY` | `false` | `false` | `null` |
| `traditional_webapp` | `FIRST_PARTY` | `false` | `false` | `null` |
| `native` | `FIRST_PARTY` | `false` | `false` | `null` |
| `confidential` | `FIRST_PARTY` | `true` | `false` | `null` |
| `third_party_app` *(deprecated)* | `THIRD_PARTY` | `true` | `false` | `null` |
| `m2m` | `FIRST_PARTY` | `true` | `true` | `null` |

All other fields map directly by name. Given this `authgear.yaml` entry:

```yaml
oauth:
  clients:
    - client_id: myapp
      name: My SPA
      x_application_type: spa
      redirect_uris:
        - https://myapp.example.com/callback
      access_token_lifetime_seconds: 1800
      refresh_token_lifetime_seconds: 86400
      refresh_token_idle_timeout_enabled: true
      refresh_token_idle_timeout_seconds: 3600
      issue_jwt_access_token: true
```

The resulting `OAuthClient` object is:

```json
{
  "clientID": "myapp",
  "source": "STATIC",
  "kind": "FIRST_PARTY",
  "isConfidential": false,
  "isServiceClient": false,
  "applicationType": null,
  "name": "My SPA",
  "clientName": null,
  "clientURI": null,
  "logoURI": null,
  "tosURI": null,
  "policyURI": null,
  "redirectURIs": ["https://myapp.example.com/callback"],
  "postLogoutRedirectURIs": [],
  "grantTypes": ["authorization_code", "refresh_token"],
  "responseTypes": ["code"],
  "accessTokenLifetimeSeconds": 1800,
  "refreshTokenLifetimeSeconds": 86400,
  "refreshTokenIdleTimeoutEnabled": true,
  "refreshTokenIdleTimeoutSeconds": 3600,
  "refreshTokenRotationEnabled": false,
  "issueJWTAccessToken": true,
  "maxConcurrentSession": 0,
  "customUIURI": null,
  "app2appEnabled": false,
  "app2appInsecureDeviceKeyBindingEnabled": false,
  "dpopDisabled": false,
  "authenticationFlowAllowlist": null,
  "preAuthenticatedURLEnabled": false,
  "preAuthenticatedURLAllowedOrigins": [],
  "replaceProjectLogoWithLogoURI": false,
  "registeredAt": null,
  "lastFetchedAt": null
}
```

Fields absent from `authgear.yaml` resolve to their defaults via `OAuthClientConfig.SetDefaults()`. Extension fields not present in the config resolve to `false` / `null` / `[]`.

## Mapping from DCR

`source` is always `DCR` for every client in this section. When a client is registered via `POST /oauth2/register`, `kind` is determined by the IAT type and `applicationType` comes directly from the OIDC `application_type` field in the request body. DCR clients are always public, so `isConfidential` is always `false`.

| DCR `application_type` | IAT type | `kind` | `isConfidential` | `isServiceClient` | `applicationType` |
|---|---|---|---|---|---|
| `web` (or omitted) | First-party (`iat_fp_`) | `FIRST_PARTY` | `false` | `false` | `"web"` |
| `native` | First-party (`iat_fp_`) | `FIRST_PARTY` | `false` | `false` | `"native"` |
| `web` (or omitted) | Third-party (`iat_tp_`) or none | `THIRD_PARTY` | `false` | `false` | `"web"` |
| `native` | Third-party (`iat_tp_`) or none | `THIRD_PARTY` | `false` | `false` | `"native"` |

Given this DCR request (with a first-party IAT):

```http
POST /oauth2/register
Authorization: Bearer iat_fp_Xf2kLmNpQrStUvWx
Content-Type: application/json

{
  "client_name": "PR #123 preview",
  "redirect_uris": ["https://pr-123.preview.example.com/callback"],
  "application_type": "web"
}
```

The resulting `OAuthClient` object is:

```json
{
  "clientID": "dcrc_Xf2kLmNpQrStUvWx",
  "source": "DCR",
  "kind": "FIRST_PARTY",
  "isConfidential": false,
  "isServiceClient": false,
  "applicationType": "web",
  "name": "PR #123 preview",
  "clientName": "PR #123 preview",
  "clientURI": null,
  "logoURI": null,
  "tosURI": null,
  "policyURI": null,
  "redirectURIs": ["https://pr-123.preview.example.com/callback"],
  "postLogoutRedirectURIs": [],
  "grantTypes": ["authorization_code", "refresh_token"],
  "responseTypes": ["code"],
  "accessTokenLifetimeSeconds": 1800,
  "refreshTokenLifetimeSeconds": 2592000,
  "refreshTokenIdleTimeoutEnabled": true,
  "refreshTokenIdleTimeoutSeconds": 1209600,
  "refreshTokenRotationEnabled": false,
  "issueJWTAccessToken": false,
  "maxConcurrentSession": 0,
  "customUIURI": null,
  "app2appEnabled": false,
  "app2appInsecureDeviceKeyBindingEnabled": false,
  "dpopDisabled": false,
  "authenticationFlowAllowlist": null,
  "preAuthenticatedURLEnabled": false,
  "preAuthenticatedURLAllowedOrigins": [],
  "replaceProjectLogoWithLogoURI": false,
  "registeredAt": "2024-11-15T00:00:00Z",
  "lastFetchedAt": null
}
```

Token lifetime fields are populated from `oauth.dynamic_client_registration.default_client_config` when set. There is no project-level token lifetime configuration to fall back on — token lifetimes exist only per client — so when a field is unset, a DCR client receives the same built-in default a static client that omits the field receives:

| Field | Default |
|---|---|
| `accessTokenLifetimeSeconds` | `1800` (30 minutes) |
| `refreshTokenLifetimeSeconds` | `31449600` (52 weeks) |
| `refreshTokenIdleTimeoutEnabled` | `true` |
| `refreshTokenIdleTimeoutSeconds` | `2592000` (30 days) |

Each field falls back independently, so a `default_client_config` that sets only `access_token_lifetime_seconds` leaves the other three at the defaults above. Note that the example object shown earlier in this section reflects the `default_client_config` in [dcr.md's Configuration](./dcr.md#configuration), not these defaults.

All Authgear extension fields are fixed at their zero values for DCR clients and cannot be changed at registration time.

Static clients are implicitly bounded — each one requires a project admin to edit and deploy `authgear.yaml`. DCR-registered and CIMD-resolved clients are not, since neither requires a per-client admin action; each source defines its own `authgear.features.yaml` limit — see [dcr.md — Client Limit](./dcr.md#client-limit) and [cimd.md — Client Limit](./cimd.md#client-limit).

## Access Token Behavior by Client Kind

For a client using the `authorization_code` / `refresh_token` grants, this table summarizes whether its access token is a JWT or opaque, whether it can be used at the [session resolver endpoint](./api-resolver.md) (`/resolve`) and at `/oauth2/userinfo`, and whether the consent screen is shown. The authoritative rules live in [access-token-audience-binding.md](./access-token-audience-binding.md) (token type, `aud`, resolver/userinfo usability) and [dcr.md](./dcr.md) (consent screen, via IAT type) — this table exists so the whole picture is visible in one place, not to redefine any rule.

| Client | `resource=` | Access token | `aud` | `/resolve` usable? | `/oauth2/userinfo` usable? | Consent screen |
|---|---|---|---|---|---|---|
| Static first-party (`spa` / `traditional_webapp` / `native` / `confidential`), `issue_jwt_access_token: false` (the default when unset) | Not usable* | Opaque | — | Yes | Yes | Bypassed |
| Static first-party, `issue_jwt_access_token: true` | Not usable* | JWT | `[<project_endpoint>]` | Yes | Yes | Bypassed |
| Static third-party (`third_party_app`, deprecated) | Not usable* | Opaque | — | **No†** | Yes | Shown |
| DCR first-party (registered with a first-party IAT) | Not usable* | Opaque (`issue_jwt_access_token` is fixed `false` for every DCR client — see [Mapping from DCR](#mapping-from-dcr)) | — | Yes | Yes | Bypassed |
| DCR third-party, `resource` omitted | Omitted | Opaque | — | **No†** | Yes | Shown |
| DCR third-party, `resource` provided | Provided, and the Resource/Scope grant `allowDynamicThirdPartyClientAccess` | JWT | `[<resource_uri>]` | **No†** | Yes | Shown |

\* `resource=` is, today, usable **only** by a dynamic third-party client (DCR-registered now, CIMD-resolved once built) — every other row above gets `invalid_target` outright, regardless of any Resource's `access_policy`. First-party support (static or dynamic) is planned separately, and a static third-party client has no mechanism to be associated with a Resource for these grants at all. See access-token-audience-binding.md's Implementation Status.

† **`/resolve` never accepts a third-party client's access token, of any shape** — opaque, or a resource-bound JWT — while a first-party client's access token (opaque or JWT) always works there. `/resolve` has no notion of "resource" and never exposes a token's `aud` to its caller (see [api-resolver.md](./api-resolver.md)), so this is a deliberate decision to avoid audience confusion, not a side effect of the opaque-token default.

Not covered above:

- **M2M clients** (`client_credentials` grant) follow an entirely different flow: no `/oauth2/authorize`, no consent screen, `resource` is *required* rather than optional, and token type is controlled purely by `issue_jwt_access_token` with no third-party override. See [m2m.md](./m2m.md).
- **CIMD clients** *(proposed, not yet implemented)* are dynamic third-party clients like DCR ones — once built, they are expected to follow the same two DCR-third-party rows above. See [cimd.md](./cimd.md).
- `/oauth2/userinfo` usability assumes the granted `scope` includes the relevant OIDC scope (e.g. `openid`) for the claim being requested; it is evaluated against `scope`, not `aud`, so it is unaffected by resource binding either way.
