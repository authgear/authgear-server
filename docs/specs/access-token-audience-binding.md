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

- **Multiple `resource` values in one request.** Exactly one `resource` value is accepted; the `aud` claim therefore always contains a single URI. The multi-resource behaviour and the intersection-downscoping rule described below apply only once first-party multi-resource support is built.
- **The `scope_by_aud` claim.** Not implemented. A resource-bound access token carries `aud` and the (unfiltered) `scope` claim only; there is no claim mapping individual scopes to individual audiences. This is safe today because at most one resource URI can ever appear in `aud` (see the point above) — every granted scope already applies to that single audience, so a per-audience breakdown carries no extra information. `scope_by_aud` becomes necessary once multiple resources can appear in one token; it is deferred alongside multi-resource support.

Everything else in this document is implemented as written.

## Glossary

**Resource** — a protected API or service identified by an `https://` URI (e.g. `https://api.example.com/orders`). Resources are pre-registered in the portal, each carrying an `access_policy` object with one key per client category that declares which categories may request it. See [API Resources and Scopes](./api-resource.md).

**Client category** — one of static first-party, static third-party, dynamic first-party, dynamic third-party. Every client belongs to exactly one, and each has its own `access_policy` key. See [API Resources and Scopes — Client categories](./api-resource.md#client-categories).

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

**`/resolve` accepts an access token only when it belongs to a first-party client**, static or dynamic. Resource binding makes no difference either way.

`/resolve` is meant for a client deployed on the same site as Authgear, behind the same reverse proxy (an nginx `auth_request`) — a first-party client can be deployed that way however it was registered. A third-party client is by definition not trusted by the project and not same-site; `/resolve` also has no notion of "resource" and never exposes the token's `aud` for a caller to check independently (see [api-resolver.md](./api-resolver.md)), so accepting one would invite audience confusion.

See [client.md — Access Token Behavior by Client Kind](./client.md#access-token-behavior-by-client-kind) for the per-client-kind breakdown.

**Both client types (with `resource` parameter):**

A JWT access token is issued with:

```
aud = ["<resource_uri>"]
```

The project endpoint is **not** included. See [How It Works](#how-it-works) for the access precondition. Such a JWT is accepted by `/resolve` when its client is first-party, and never when it is third-party — as above, resource binding does not change that.

## How It Works

### Without Resource Indicator

| Client type | Token type | `aud` |
|---|---|---|
| First-party | JWT | `[<project_endpoint>]` |
| Third-party | Opaque | N/A |

### With Resource Indicator

When `resource` is specified at `/oauth2/authorize`, Authgear determines the client's [category](./api-resource.md#client-categories) and permits access only if both the Resource and every requested Scope have that category's `access_policy` key set to `true`. Otherwise `invalid_target`.

The check is per-category and literal: a static third-party client is not covered by `allow_dynamic_third_party_client_access`, and a dynamic first-party client is not covered by `allow_static_first_party_client_access`.

The policy is re-read on every access token issuance, not only here, so a grant records what was authorized rather than whether it still is — see [API Resources and Scopes — Revocation](./api-resource.md#revocation).

Client-Resource Associations play no part in any of this — they are consulted only by `client_credentials`, and the two mechanisms are partitioned by grant rather than OR-ed. See [Relationship to Client-Resource Association](./api-resource.md#relationship-to-client-resource-association).

When access is permitted, a JWT access token is issued with `aud = [<resource_uri>]`. The project endpoint is **not** included in `aud`.

Authgear-specific scopes (`https://authgear.com/scopes/full-access`, `https://authgear.com/scopes/full-userinfo`, `https://authgear.com/scopes/pre-authenticated-url`) and `device_sso` may be requested alongside `resource`; its presence neither grants nor withdraws them. A resource-bound token can therefore carry project-level privileges even though its `aud` names an external resource server. This is deliberate.

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
- Each `resource` value must refer to a Resource the client is permitted to access: the Resource and every requested Scope must both set the `access_policy` key for the client's [category](./api-resource.md#client-categories). Otherwise `invalid_target` is returned.
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
- The Resource's and Scopes' `access_policy` is re-read before the token is issued — see [API Resources and Scopes — Revocation](./api-resource.md#revocation).

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
- The Resource's and Scopes' `access_policy` is re-read before the token is issued — see [API Resources and Scopes — Revocation](./api-resource.md#revocation).

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

The userinfo endpoint accepts tokens where `scope` contains OIDC scopes (e.g. `openid`, `profile`, `email`), regardless of the `aud` claim. Resource servers should validate `aud` contains their own URI and `scope` contains the required resource-specific scopes. A resource-bound token issued to a **third-party** client is additionally never accepted by the `/resolve` endpoint (see [Authgear's decision](#authgears-decision)) — for such a client, only the resource server it names in `aud` should accept it.

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
| The Resource does not set the `access_policy` key for the client's category | `invalid_target` |
| Client is an M2M client (M2M clients cannot use `/oauth2/authorize` at all) | `unauthorized_client` |
| `scope` includes a resource-specific scope but no matching `resource` was requested | `invalid_scope` |
| Requested scope is not permitted for the client on that resource | `invalid_scope` |

### Token endpoint errors

| Condition | `error` | HTTP status |
|---|---|---|
| `resource` URI at token exchange (`authorization_code` grant) is not a subset of what was authorized | `invalid_target` | 400 |
| `resource` URI at refresh (`refresh_token` grant) is not a subset of the original grant | `invalid_target` | 400 |
| The bound Resource no longer allows the client's category, or has been deleted (either grant) | `invalid_target` | 400 |

## Backward Compatibility

### First-party clients

Unchanged when `resource` is omitted: JWT with `aud = [<project_endpoint>]`. Existing resource servers that validate `aud` contains `<project_endpoint>` continue to work without modification.

Extending the `access_policy` to first-party clients does not change any existing deployment, because every key defaults to `false` (see [api-resource.md — Defaults](./api-resource.md#defaults)): a first-party client that sends `resource` keeps receiving `invalid_target` until an admin turns the relevant key on, and no Resource that exists today becomes reachable by a client that could not reach it before.

Once a key is on, note that the resulting token's `aud` **replaces** the project endpoint rather than adding to it — a client that starts sending `resource` loses the audience its existing resource servers validate against. Sending `resource` is therefore a per-request decision, not a project-wide switch.

### Third-party clients

Third-party clients are new. No existing behavior is affected.

### `aud` when `resource` is specified

When `resource` is specified, `aud` contains **only** the resource URI(s). This is new behavior — `resource` support for `authorization_code` and `refresh_token` grants did not previously exist.

## Relationship to M2M

The `m2m` client type (`client_credentials` grant) already supports resource indicators as described in `docs/specs/m2m.md`. This spec extends resource indicators — the same pre-registered Resources and Scopes — to the `authorization_code` and `refresh_token` grants. `docs/specs/m2m.md` documents a `scope_by_aud` claim for multi-resource `client_credentials` tokens; that claim is not implemented for either grant family today (see [Implementation Status](#implementation-status)).

Two things do **not** carry across:

- **`resource` is required** for `client_credentials` (per existing implementation), and **optional** for `authorization_code` and `refresh_token`, to preserve backward compatibility.
- **The authorization mechanism is different.** `client_credentials` uses an explicit Client-Resource Association and never consults `access_policy`; `authorization_code`/`refresh_token` use `access_policy` and never consult an association. The two are partitioned by grant, not layered — see [api-resource.md — Relationship to Client-Resource Association](./api-resource.md#relationship-to-client-resource-association) and [m2m.md](./m2m.md#discussion-resource-scope-client-and-downscoping).
