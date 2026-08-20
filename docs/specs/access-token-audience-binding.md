# Access Token Audience Binding

Access token audience binding is the mechanism by which Authgear binds the `aud` claim of a JWT access token to one or more specific resource server URIs, preventing a token issued for one resource from being accepted by another.

This is implemented via [RFC 8707 — Resource Indicators for OAuth 2.0](https://www.rfc-editor.org/rfc/rfc8707).

## Table of Contents

- [Implementation Status](#implementation-status)
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

## Implementation Status

Two parts of this specification are described in full below but are **not implemented**:

- **First-party clients using `resource`.** Outside the `client_credentials` grant, a first-party client that sends `resource` receives `invalid_target`. Only **dynamic** third-party clients — DCR-registered today, CIMD-resolved later — can currently use the parameter with `authorization_code` / `refresh_token`. A static `third_party_app` client also receives `invalid_target`: it has no mechanism to be associated with a Resource for these grants (unlike an M2M client's explicit Client-Resource Association for `client_credentials`), so this is not a "not yet implemented" gap to close later — `allow_dynamic_third_party_client_access` is dynamic-only by design, per its name. First-party support is planned separately.
- **Multiple `resource` values in one request.** Exactly one `resource` value is accepted; the `aud` claim therefore always contains a single URI. The multi-resource behaviour and the intersection-downscoping rule described below apply only once first-party multi-resource support is built.
- **The `scope_by_aud` claim.** Not implemented. A resource-bound access token carries `aud` and the (unfiltered) `scope` claim only; there is no claim mapping individual scopes to individual audiences. This is safe today because at most one resource URI can ever appear in `aud` (see the point above) — every granted scope already applies to that single audience, so a per-audience breakdown carries no extra information. `scope_by_aud` becomes necessary once multiple resources can appear in one token; it is deferred alongside multi-resource support.

Everything else in this document is implemented as written.

## Glossary

**Resource** — a protected API or service identified by an `https://` URI (e.g. `https://api.example.com/orders`). Resources are pre-registered in the portal and optionally configured with `access_policy.allow_dynamic_third_party_client_access: true` to permit dynamic (DCR/CIMD) third-party client access. See [API Resources and Scopes](./api-resource.md).

**Resource-specific Scope** — a scope value (e.g. `read:orders`) that is defined on a Resource and only meaningful when the corresponding Resource is included in the `resource` parameter.

**Resource Indicator** — the `resource` request parameter defined by RFC 8707, used by clients to declare which resource(s) they want a token to be bound to.

**Access Token Audience Binding** — the act of including one or more resource URIs in the `aud` claim of an access token, so that each resource server can validate that the token was intended for it.

## Background

Without access token audience binding, all Authgear access tokens share `aud = [<project_endpoint>]`. A resource server that only validates `aud` cannot distinguish tokens intended for different services — a token issued to a third-party client would be structurally accepted by a first-party client on the same project. This is the **audience confusion** risk.

The standard solution is RFC 8707 resource indicators: clients declare their target resource at request time, and Authgear binds the `aud` of the issued token to that resource URI. Resource servers can then enforce `aud` contains their own URI.

Authgear previously supported resource indicators only for `m2m` clients using the `client_credentials` grant. This spec extends support to the `authorization_code` and `refresh_token` grants. See [Implementation Status](#implementation-status) for which client types that currently covers.

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

1. If the Resource has `access_policy.allow_dynamic_third_party_client_access: true` **and** the requested Scope(s) have `access_policy.allow_dynamic_third_party_client_access: true` — any **dynamic** third-party client is allowed. A static `third_party_app` client is never allowed by this policy, regardless of the flag — see [Implementation Status](#implementation-status).
2. Otherwise, an explicit Client-Resource Association is required. Currently only M2M clients support explicit associations (see [API Resources and Scopes](./api-resource.md#client-resource-association)). Dynamic third-party clients without the access policy set on the resource cannot use it, and static third-party clients cannot use it at all.

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
  &resource=https://api.example.com/orders       ← optional; single value only, see Implementation Status
```

**Rules:**

- `resource` is optional.
  - First-party client, omitted: issues a JWT with `aud = [<project_endpoint>]`.
  - Third-party client, omitted: issues an opaque access token.
- Each `resource` value must refer to a Resource the client is permitted to access: either the Resource and requested Scopes have `access_policy.allow_dynamic_third_party_client_access: true` (for **dynamic** third-party clients only — a static `third_party_app` client is never permitted, see [Implementation Status](#implementation-status)), or the client is an M2M client with an explicit Client-Resource Association for that Resource. Otherwise `invalid_target` is returned.
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

When `resource` is specified, `aud` contains **only** the requested resource URI. The Authgear project endpoint is not included. All granted scopes — including OIDC scopes (e.g. `openid`, `offline_access`) that have no relationship to the resource — appear together in the top-level `scope` field; there is no `scope_by_aud` claim breaking them down per audience (see [Implementation Status](#implementation-status)).

```json
{
  "iss": "https://myapp.authgear.cloud",
  "sub": "user-id",
  "aud": ["https://api.example.com/orders"],
  "client_id": "dcrc_Xf2kLmNpQrStUvWx",
  "scope": "openid offline_access read:orders"
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
4. `scope` contains the required scopes.

## Error Cases

Error response format differs by endpoint:

- **Authorization endpoint** — errors are returned as a redirect to `redirect_uri` with `error` and `error_description` query parameters (per RFC 6749 §4.1.2.1). There is no direct HTTP error response.
- **Token endpoint** — errors are returned as a JSON body with HTTP 400 (per RFC 6749 §5.2).

### Authorization endpoint errors

| Condition | `error` |
|---|---|
| `resource` URI is not a pre-registered Resource | `invalid_target` |
| `resource` URI is prefixed by the Authgear project endpoint | `invalid_target` |
| Client is a dynamic third-party client and the Resource does not have `access_policy.allow_dynamic_third_party_client_access: true` | `invalid_target` |
| Client is a static `third_party_app` client, or any first-party client (static or dynamic) | `invalid_target` |
| Client is an M2M client and no explicit Client-Resource Association exists | `invalid_target` |
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

Third-party clients are new. No existing behavior is affected.

### `aud` when `resource` is specified

When `resource` is specified, `aud` contains **only** the resource URI(s). This is new behavior — `resource` support for `authorization_code` and `refresh_token` grants did not previously exist.

## Relationship to M2M

The `m2m` client type (`client_credentials` grant) already supports resource indicators as described in `docs/specs/m2m.md`. This spec extends the same mechanism — the same pre-registered Resources and the same client-resource association model — to the `authorization_code` and `refresh_token` grants for all client types. `docs/specs/m2m.md` documents a `scope_by_aud` claim for multi-resource `client_credentials` tokens; that claim is not implemented for either grant family today (see [Implementation Status](#implementation-status)).

The key difference is that for `client_credentials`, `resource` is **required** (per existing implementation). For `authorization_code` and `refresh_token`, `resource` is **optional** to preserve backward compatibility.
