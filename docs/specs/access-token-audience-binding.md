# Access Token Audience Binding

Access token audience binding is the mechanism by which Authgear binds the `aud` claim of a JWT access token to one or more specific resource server URIs, preventing a token issued for one resource from being accepted by another.

This is implemented via [RFC 8707 — Resource Indicators for OAuth 2.0](https://www.rfc-editor.org/rfc/rfc8707).

## Table of Contents

- [Glossary](#glossary)
- [Background](#background)
- [Default Audience and Audience Confusion Risk](#default-audience-and-audience-confusion-risk)
- [How It Works](#how-it-works)
  - [Without Resource Indicator](#without-resource-indicator)
  - [With Resource Indicator](#with-resource-indicator)
- [Authorization Endpoint](#authorization-endpoint)
- [Token Endpoint](#token-endpoint)
  - [authorization_code grant](#authorization_code-grant)
  - [refresh_token grant](#refresh_token-grant)
- [Access Token Claims](#access-token-claims)
- [Error Cases](#error-cases)
- [Backward Compatibility](#backward-compatibility)
- [Relationship to M2M](#relationship-to-m2m)

## Glossary

**Resource** — a protected API or service identified by an `https://` URI (e.g. `https://api.example.com/orders`). Resources are pre-registered in the portal and optionally marked with `allow_any_client_access` to permit DCR client access. See [API Resources and Scopes](./api-resource.md).

**Resource-specific Scope** — a scope value (e.g. `read:orders`) that is defined on a Resource and only meaningful when the corresponding Resource is included in the `resource` parameter.

**Resource Indicator** — the `resource` request parameter defined by RFC 8707, used by clients to declare which resource(s) they want a token to be bound to.

**Access Token Audience Binding** — the act of including one or more resource URIs in the `aud` claim of an access token, so that each resource server can validate that the token was intended for it.

## Background

Without access token audience binding, all Authgear access tokens share `aud = [<project_endpoint>]`. A resource server that only validates `aud` cannot distinguish tokens intended for different services — a token issued to a third-party client would be structurally accepted by a first-party client on the same project. This is the **audience confusion** risk.

The standard solution is RFC 8707 resource indicators: clients declare their target resource at request time, and Authgear binds the `aud` of the issued token to that resource URI. Resource servers can then enforce `aud` contains their own URI.

Authgear previously supported resource indicators only for `m2m` clients using the `client_credentials` grant. This spec extends support to all client types using the `authorization_code` and `refresh_token` grants.

## Default Audience and Audience Confusion Risk

### The problem with `aud = [<project_endpoint>]`

Without any resource binding, all JWT access tokens issued by a project share `aud = [<project_endpoint>]`. This means a token issued to client A is structurally accepted by any resource server that validates against the same project endpoint — including APIs that were never intended to accept tokens from client A. The audience confusion risk is especially acute for third-party clients, which are operated by external developers.

### Competitor analysis

We reviewed how other providers handle this:

| Provider | Default `aud` without explicit audience config | Out-of-box isolation |
|---|---|---|
| Auth0 | Issues an **opaque** (non-JWT) token scoped only to userinfo | **Enforced by design.** Without specifying `audience=` (a pre-registered API identifier), callers cannot obtain a JWT at all — forcing developers to consciously bind every token to a resource. |
| Keycloak | No meaningful resource server audience | **None by default.** Keycloak provides "Audience Mapper" configuration: admins create a Client Scope, attach an Audience Mapper with the resource server URI, and assign that scope to specific clients. This works when configured, but requires deliberate per-resource setup. Deployments that skip this configuration remain fully exposed. |
| Okta | Fixed audience set at the authorization server level (e.g. `api://default`) | **Partial, coarse-grained.** All tokens from one authorization server share a fixed `aud`. Isolation between different resource servers requires deploying separate authorization servers — impractical for most projects. |

### Authgear's decision

Authgear takes a different approach for first-party and third-party clients:

**First-party clients:**

The JWT access token retains the existing default:

```
aud = ["<project_endpoint>"]
```

This preserves backward compatibility for existing first-party deployments.

**Third-party clients:**

An **opaque** access token is issued instead of a JWT. The opaque token:

- Can be presented to the userinfo endpoint (`/oauth2/userinfo`) to retrieve user information.
- Cannot be used with the `/resolve` endpoint.
- Has no `aud` claim and cannot be validated by a resource server independently.

This solves the audience confusion problem for third-party clients by design: without specifying a `resource`, a third-party client can only access userinfo and nothing else.

**Both client types (with `resource` parameter):**

A JWT access token is issued with:

```
aud = ["<resource_uri>"]
```

The project endpoint is **not** included. See [How It Works](#how-it-works) for the access precondition.

## How It Works

### Without Resource Indicator

| Client type | Token type | `aud` |
|---|---|---|
| First-party | JWT | `[<project_endpoint>]` |
| Third-party | Opaque | N/A |

### With Resource Indicator

When `resource` is specified, Authgear checks whether the client is permitted to access that resource using the following logic:

1. If the Resource has `allow_any_client_access: true` **and** the requested Scope(s) have `allow_any_client_access: true` — any client (first-party or third-party) is allowed.
2. Otherwise, an explicit Client-Resource Association is required. Currently only M2M clients support explicit associations (see [API Resources and Scopes](./api-resource.md#client-resource-association)). Third-party clients without `allow_any_client_access` on the resource cannot use it.

When access is permitted, a JWT access token is issued with `aud = [<resource_uri>]`. The project endpoint is **not** included in `aud`.

See [API Resources and Scopes](./api-resource.md) for how to register Resources and configure access.

## Authorization Endpoint

```
GET /oauth2/authorize
  ?client_id=<client_id>
  &response_type=code
  &scope=openid offline_access read:orders
  &redirect_uri=<redirect_uri>
  &code_challenge=<challenge>
  &code_challenge_method=S256
  &resource=https://api.example.com/orders       ← optional, repeatable
  &resource=https://api.example.com/inventory    ← multiple resources allowed
```

**Rules:**

- `resource` is optional.
  - First-party client, omitted: issues a JWT with `aud = [<project_endpoint>]`.
  - Third-party client, omitted: issues an opaque access token.
- Each `resource` value must refer to a Resource the client is permitted to access: either the Resource and requested Scopes have `allow_any_client_access: true`, or the client is an M2M client with an explicit Client-Resource Association for that Resource. Otherwise `invalid_target` is returned.
- Resource URIs must not be prefixed by the Authgear project endpoint.
- The granted resources are bound to the authorization code and stored server-side.

## Token Endpoint

### `authorization_code` grant

```
POST /oauth2/token
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code
&code=<code>
&code_verifier=<verifier>
&client_id=<client_id>
&redirect_uri=<redirect_uri>
&resource=https://api.example.com/orders    ← optional
```

**Rules:**

- `resource` is optional at this step.
- If provided, it must be a subset of the resources bound to the authorization code. Requesting a resource outside the bound set returns `invalid_target`.
- If omitted:
  - If resources were bound to the authorization code, the token is issued as a JWT with `aud` containing those resource URIs.
  - If no resources were bound (first-party client only): JWT with `aud = [<project_endpoint>]`.
  - If no resources were bound (third-party client): opaque access token.

### `refresh_token` grant

```
POST /oauth2/token
Content-Type: application/x-www-form-urlencoded

grant_type=refresh_token
&refresh_token=<token>
&client_id=<client_id>
&resource=https://api.example.com/orders    ← optional, downscoping allowed
```

**Rules:**

- `resource` is optional.
- If provided, it must be a subset of the resources originally authorized (downscoping is allowed; upscoping is not).
- If omitted, the new access token is issued for the same resources as the previous access token in this session.
- Requesting a resource not in the original grant returns `invalid_target`.

## Access Token Claims

### With Resource Indicator

When `resource` is specified, `aud` contains **only** the requested resource URI(s). The Authgear project endpoint is not included. The `scope_by_aud` claim maps which scopes apply to which resource. OIDC scopes (e.g. `openid`, `offline_access`) that were granted appear in the top-level `scope` field even though there is no corresponding `aud` entry for the project endpoint.

```json
{
  "iss": "https://myapp.authgear.cloud",
  "sub": "user-id",
  "aud": ["https://api.example.com/orders"],
  "client_id": "dcrc_Xf2kLmNpQrStUvWx",
  "scope": "openid offline_access read:orders",
  "https://authgear.com/claims/scope_by_aud": [
    {
      "aud": "https://api.example.com/orders",
      "scope": "read:orders"
    }
  ]
}
```

The userinfo endpoint accepts tokens where `scope` contains OIDC scopes (e.g. `openid`, `profile`, `email`), regardless of the `aud` claim. Resource servers should validate `aud` contains their own URI and `scope` contains the required resource-specific scopes.

### Default — first-party client

A JWT is issued with `aud` set to the project endpoint:

```json
{
  "iss": "https://myapp.authgear.cloud",
  "sub": "user-id",
  "aud": ["https://myapp.authgear.cloud"],
  "client_id": "spa-client-id",
  "scope": "openid offline_access"
}
```

### Default — third-party client

An opaque access token is issued. It has no `aud` claim and cannot be decoded by the caller. It is only accepted by the userinfo endpoint.

### Resource server validation

A resource server at `https://api.example.com/orders` should validate:

1. `access_token` is a valid JWT signed by the Authgear project key (via `jwks_uri`).
2. `iss` matches the expected Authgear project endpoint.
3. `aud` includes `https://api.example.com/orders`.
4. `scope` (or `scope_by_aud` for the resource's entry) contains the required scopes.

## Error Cases

Error response format differs by endpoint:

- **Authorization endpoint** — errors are returned as a redirect to `redirect_uri` with `error` and `error_description` query parameters (per RFC 6749 §4.1.2.1). There is no direct HTTP error response.
- **Token endpoint** — errors are returned as a JSON body with HTTP 400 (per RFC 6749 §5.2).

### Authorization endpoint errors

| Condition | `error` |
|---|---|
| `resource` URI is not a pre-registered Resource | `invalid_target` |
| `resource` URI is prefixed by the Authgear project endpoint | `invalid_target` |
| Client is third-party and the Resource does not have `allow_any_client_access: true` | `invalid_target` |
| Client is an M2M client, Resource does not have `allow_any_client_access: true`, and no explicit Client-Resource Association exists | `invalid_target` |
| `scope` includes a resource-specific scope but no matching `resource` was requested | `invalid_scope` |
| Requested scope is not permitted for the client on that resource | `invalid_scope` |

### Token endpoint errors

| Condition | `error` | HTTP status |
|---|---|---|
| `resource` URI at token exchange (`authorization_code` grant) is not a subset of what was authorized | `invalid_target` | 400 |
| `resource` URI at refresh (`refresh_token` grant) is not a subset of the original grant | `invalid_target` | 400 |

## Backward Compatibility

### First-party clients

Unchanged. JWT with `aud = [<project_endpoint>]`. Existing resource servers that validate `aud` contains `<project_endpoint>` continue to work without modification.

### Third-party clients

Third-party clients (dynamically registered via DCR) are new. No existing behavior is affected.

### `aud` when `resource` is specified

When `resource` is specified, `aud` contains **only** the resource URI(s). This is new behavior — `resource` support for `authorization_code` and `refresh_token` grants did not previously exist.

## Relationship to M2M

The `m2m` client type (`client_credentials` grant) already supports resource indicators as described in `docs/specs/m2m.md`. This spec extends the same mechanism — the same pre-registered Resources, the same client-resource association model, and the same `scope_by_aud` claim — to the `authorization_code` and `refresh_token` grants for all client types.

The key difference is that for `client_credentials`, `resource` is **required** (per existing implementation). For `authorization_code` and `refresh_token`, `resource` is **optional** to preserve backward compatibility.
