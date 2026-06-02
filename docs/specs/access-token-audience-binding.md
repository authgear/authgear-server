# Access Token Audience Binding

Access token audience binding is the mechanism by which Authgear binds the `aud` claim of a JWT access token to one or more specific resource server URIs, preventing a token issued for one resource from being accepted by another.

This is implemented via [RFC 8707 — Resource Indicators for OAuth 2.0](https://www.rfc-editor.org/rfc/rfc8707).

## Table of Contents

- [Glossary](#glossary)
- [Background](#background)
- [Default Audience and Audience Confusion Risk](#default-audience-and-audience-confusion-risk)
- [How It Works](#how-it-works)
  - [Statically configured clients](#statically-configured-clients)
  - [DCR-registered clients](#dcr-registered-clients)
- [Authorization Endpoint](#authorization-endpoint)
- [Token Endpoint](#token-endpoint)
  - [authorization_code grant](#authorization_code-grant)
  - [refresh_token grant](#refresh_token-grant)
- [Access Token Claims](#access-token-claims)
- [Error Cases](#error-cases)
- [Backward Compatibility](#backward-compatibility)
- [Relationship to M2M](#relationship-to-m2m)

## Glossary

**Resource** — a protected API or service identified by an `https://` URI (e.g. `https://api.example.com/orders`). For statically configured clients, Resources are pre-registered in the portal. For DCR clients, Resources are declared in the `allowed_resources` list in `authgear.yaml`.

**Resource-specific Scope** — a scope value (e.g. `read:orders`) that is defined on a Resource and only meaningful when the corresponding Resource is included in the `resource` parameter.

**Resource Indicator** — the `resource` request parameter defined by RFC 8707, used by clients to declare which resource(s) they want a token to be bound to.

**Access Token Audience Binding** — the act of including one or more resource URIs in the `aud` claim of an access token, so that each resource server can validate that the token was intended for it.

## Background

Without access token audience binding, all Authgear access tokens share `aud = [<project_endpoint>]`. A resource server that only validates `aud` cannot distinguish tokens intended for different services — a token issued to a third-party app would be structurally accepted by a first-party API on the same project. This is the **audience confusion** risk.

The standard solution is RFC 8707 resource indicators: clients declare their target resource at request time, and Authgear binds the `aud` of the issued token to that resource URI. Resource servers can then enforce `aud` contains their own URI.

Authgear previously supported resource indicators only for `m2m` clients using the `client_credentials` grant. This spec extends support to all client types using the `authorization_code` and `refresh_token` grants.

## Default Audience and Audience Confusion Risk

### The problem with `aud = [<project_endpoint>]`

Without any resource binding, all JWT access tokens issued by a project share `aud = [<project_endpoint>]`. This means a token issued to client A is structurally accepted by any resource server that validates against the same project endpoint — including APIs that were never intended to accept tokens from client A. This is the **audience confusion** risk.

Leaving the default as `aud = [<project_endpoint>]` is therefore not acceptable: every deployment is silently vulnerable unless developers explicitly use the `resource` parameter.

### Competitor analysis

We reviewed how other providers handle this:

| Provider | Default `aud` without explicit audience config | Out-of-box isolation |
|---|---|---|
| Auth0 | Issues an **opaque** (non-JWT) token scoped only to userinfo | **Enforced by design.** Without specifying `audience=` (a pre-registered API identifier), callers cannot obtain a JWT at all — forcing developers to consciously bind every token to a resource. |
| Keycloak | No meaningful resource server audience | **None by default.** Keycloak provides "Audience Mapper" configuration: admins create a Client Scope, attach an Audience Mapper with the resource server URI, and assign that scope to specific clients. This works when configured, but requires deliberate per-resource setup. Deployments that skip this configuration remain fully exposed. |
| Okta | Fixed audience set at the authorization server level (e.g. `api://default`) | **Partial, coarse-grained.** All tokens from one authorization server share a fixed `aud`. Isolation between different resource servers requires deploying separate authorization servers — impractical for most projects. |

RFC 9068 §3 requires that when no `resource` parameter is present, the authorization server MUST use a default resource indicator in the `aud` claim. It does not prescribe what that default should be.

Auth0's approach (opaque token by default) is the most principled but is a breaking change. Keycloak and Okta require explicit admin configuration with no safe default.

### Authgear's decision: per-client default audience

Authgear uses the client's own URI as the default audience, providing per-client token isolation without requiring any resource indicator configuration.

**When no `resource` parameter is specified**, the issued JWT access token includes:

```
aud = ["<project_endpoint>", "<project_endpoint>/clients/<client_id>"]
```

For example, for client `my-spa` on project `https://myapp.authgear.cloud`:

```
aud = ["https://myapp.authgear.cloud", "https://myapp.authgear.cloud/clients/my-spa"]
```

**Rationale:**
- Tokens from different clients are no longer interchangeable. A resource server can restrict access to tokens from specific clients by validating `aud` contains `<project_endpoint>/clients/<expected-client-id>`.
- Adding `<project_endpoint>/clients/<client_id>` to `aud` is additive. Existing resource servers that validate `aud` contains `<project_endpoint>` continue to work without changes.
- The `<project_endpoint>` entry is retained for backward compatibility so existing deployments are not broken.
- This is consistent with RFC 9068's requirement to use a meaningful default, and mirrors how OIDC ID tokens use the requesting client's `client_id` as `aud`.

**Resource servers that want per-client isolation** should validate that `aud` includes `<project_endpoint>/clients/<their-expected-client-id>`, not just `<project_endpoint>`.

## How It Works

If a client never specifies `resource`, tokens are issued with `aud = [<project_endpoint>, <project_endpoint>/clients/<client_id>]` (see [Default Audience and Audience Confusion Risk](#default-audience-and-audience-confusion-risk) above).

### Statically configured clients

> **Not yet supported.** Resource association for statically configured clients is planned but not implemented. This section describes the intended design.

1. An administrator pre-registers Resources in the portal, each with a URI and a set of scopes.
2. The administrator associates a client with one or more Resources (and which scopes on each resource the client may request).
3. At authorization time, the client includes `resource=<uri>` in its request. Multiple resources may be requested by repeating the parameter.
4. Authgear validates the requested resources against the client's allowed list, binds them to the authorization code, and issues an access token with those URIs added to `aud`.

### DCR-registered clients

DCR clients use a separate allow-list defined in `authgear.yaml` rather than the per-client portal associations. This prevents DCR clients from accessing resources not explicitly approved for DCR use.

```yaml
oauth:
  dynamic_client_registration:
    allowed_resources:
      - uri: "https://mcp-server.example.com"
        scopes:
          - "read:tools"
          - "execute:tools"
```

- Only URIs listed in `allowed_resources` are valid `resource` values for DCR clients. Requesting any other URI returns `invalid_target`.
- Only scopes listed under that URI are valid. Requesting an unlisted scope returns `invalid_scope`.
- The `allowed_resources` list is a project-wide policy shared by all DCR clients. It is configured once by the admin; individual DCR clients need no further admin action to use it.
- Resources in `allowed_resources` do not need to be separately registered in the portal. The URI and scopes are fully defined inline.

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

- `resource` is optional. Omitting it produces a token with `aud = [<project_endpoint>, <project_endpoint>/clients/<client_id>]`.
- Each `resource` value must be a pre-registered Resource URI associated with the requesting client. Unrecognized or unassociated resources return `invalid_target`.
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
- If omitted, the token is issued with `aud` containing all resources bound to the authorization code plus the per-client default entries (`<project_endpoint>` and `<project_endpoint>/clients/<client_id>`).

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

### With `resource`

The `aud` claim expands to include each requested resource URI alongside the project endpoint. The `scope_by_aud` claim maps which scopes apply to which audience:

```json
{
  "iss": "https://myapp.authgear.cloud",
  "sub": "user-id",
  "aud": ["https://myapp.authgear.cloud", "https://api.example.com/orders"],
  "client_id": "dcrc_Xf2kLmNpQrStUvWx",
  "scope": "openid offline_access read:orders",
  "https://authgear.com/claims/scope_by_aud": [
    {
      "aud": "https://myapp.authgear.cloud",
      "scope": "openid offline_access"
    },
    {
      "aud": "https://api.example.com/orders",
      "scope": "read:orders"
    }
  ]
}
```

### Without `resource`

The `aud` claim includes the project endpoint and the per-client URI:

```json
{
  "iss": "https://myapp.authgear.cloud",
  "sub": "user-id",
  "aud": ["https://myapp.authgear.cloud", "https://myapp.authgear.cloud/clients/spa-client-id"],
  "client_id": "spa-client-id",
  "scope": "openid offline_access"
}
```

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
| `resource` URI is not a pre-registered Resource (static client) | `invalid_target` |
| `resource` URI is not associated with this client (static client) | `invalid_target` |
| `resource` URI is not in `allowed_resources` (DCR client) | `invalid_target` |
| `resource` URI is prefixed by the Authgear project endpoint | `invalid_target` |
| `scope` includes a resource-specific scope but no matching `resource` was requested | `invalid_scope` |

### Token endpoint errors

| Condition | `error` | HTTP status |
|---|---|---|
| `scope` is not in the allowed scopes for that resource (DCR client) | `invalid_scope` | 400 |
| `resource` URI at token exchange (`authorization_code` grant) is not a subset of what was authorized | `invalid_target` | 400 |
| `resource` URI at refresh (`refresh_token` grant) is not a subset of the original grant | `invalid_target` | 400 |

## Backward Compatibility

The default `aud` now includes `<project_endpoint>/clients/<client_id>` in addition to `<project_endpoint>`. This is an additive change: existing resource servers that validate `aud` contains `<project_endpoint>` continue to work without modification.

Resource-specific audience binding via the `resource` parameter is opt-in. Clients that do not specify `resource` receive the per-client default `aud` described above.

## Relationship to M2M

The `m2m` client type (`client_credentials` grant) already supports resource indicators as described in `docs/specs/m2m.md`. This spec extends the same mechanism — the same pre-registered Resources, the same client-resource association model, and the same `scope_by_aud` claim — to the `authorization_code` and `refresh_token` grants for all client types.

The key difference is that for `client_credentials`, `resource` is **required** (per existing implementation). For `authorization_code` and `refresh_token`, `resource` is **optional** to preserve backward compatibility.
