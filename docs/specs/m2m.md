# M2M authentication and authorization

## Table of Contents

- [Use-cases](#use-cases)
  - [Sign in `johndoe` in `mobileapp` without `resource`](#use-cases-sign-in-johndoe-in-mobileapp-without-resource)
  - [Sign in `johndoe` in `mobileapp` with single `resource`](#use-cases-sign-in-johndoe-in-mobileapp-with-single-resource)
  - [Sign in `johndoe` in `mobileapp` with multiple `resource`](#use-cases-sign-in-johndoe-in-mobileapp-with-multiple-resource)
  - [M2M of `inventory` to `https://onlinestore.myapp.com`](#use-cases-m2m-of-inventory-to-httpsonlinestoremyappcom)
  - [Handling of `access_token` in Resources](#use-cases-handling-of-access_token-in-resources)
- [Discussions](#discussions)
  - [M2M in other protocol, like SAML 2.0](#discussion-m2m-in-other-protocol-like-saml-20)
  - [About `scope` and `access_token`](#discussion-about-scope-and-access_token)
  - [Resource and Scope](#discussion-resource-and-scope)
  - [Allowed values of Scope](#discussion-allowed-values-of-scope)
  - [Resource, Scope, Client, and downscoping](#discussion-resource-scope-client-and-downscoping)
  - [Client Credentials Grant and downscoping](#discussion-client-credentials-grant-and-downscoping)
  - [Resource, Scope, Client, consent](#discussion-resource-scope-client-consent)
  - [Signify the intention of granting all scopes to a client](#discussion-signify-the-intention-of-granting-all-scopes-to-a-client)
  - [(Auth0) General facts on Auth0 M2M](#discussion-auth0-general-facts-on-auth0-m2m)
  - [(Auth0) Example of JWT access token of Client Credentials Grant](#discussion-auth0-example-of-jwt-access-token-of-client-credentials-grant)
  - [(Auth0) Example of JWT access token of Authorization Code Grant](#discussion-auth0-example-of-jwt-access-token-of-authorization-code-grant)
  - [(Auth0) Third party public client and Organization](#discussion-auth0-third-party-public-client-and-organization)
  - [(Auth0) Third party public client and Connection](#discussion-auth0-third-party-public-client-and-connection)
  - [(Auth0) Third party public client and consent](#discussion-auth0-third-party-public-client-and-consent)
  - [(Auth0) Multiple audience](#discussion-auth0-multiple-audience)
  - [(Auth0) The `permissions` claim](#discussion-auth0-the-permissions-claim)
  - [(Authgear) Proposed Token Exchange behavior](#discussion-authgear-proposed-token-exchange-behavior)
  - [(Auth0) Rich Authorization Requests](#discussion-auth0-rich-authorization-requests)
  - [(Future work) Restricting access to Admin API](#discussion-future-work-restricting-access-to-admin-api)
- [MVP](#mvp)
  - [Changes in data models](#changes-in-data-models)
  - [Changes in Admin API](#changes-in-admin-api)
  - [Changes in OAuth 2.0 implementation](#changes-in-oauth-20-implementation)
    - [The request](#the-request)
    - [The successful response](#the-successful-response)
    - [The access token](#the-access-token)
    - [The error response](#the-error-response)
  - [Changes in SDKs](#changes-in-sdks)
  - [Changes in Documentation](#changes-in-documentation)

M2M authentication and authorization is based on the following specifications:

- [RFC6749: The OAuth 2.0 Authorization Framework Authorization Code Grant](https://datatracker.ietf.org/doc/html/rfc6749#section-4.1)
  - We will refer it as `RFC6749 section-4.1` in the below text.
- [OpenID Connect Core 1.0 incorporating errata set 2](https://openid.net/specs/openid-connect-core-1_0.html)
  - We will refer it as `OIDC-Core` in the below text.
- [RFC6749: The OAuth 2.0 Authorization Framework Client Credentials Grant](https://datatracker.ietf.org/doc/html/rfc6749#section-4.4)
  - We will refer it as `RFC6749 section-4.4` in the below text.
- [RFC9068: JSON Web Token (JWT) Profile for OAuth 2.0 Access Tokens](https://datatracker.ietf.org/doc/html/rfc9068)
  - We will refer it as `RFC9068` in the below text.
- [RFC8707: Resource Indicators for OAuth 2.0](https://datatracker.ietf.org/doc/html/rfc8707)
  - We will refer it as `RFC8707` in the below text.

It is assumed that the reader has read the above RFCs, or at least have an idea of what they are.

## Use-cases

This section lists out how each use-case looks like.

To begin the story, let us assume we have

- A company owning the domain `myapp.com`.
- Authgear is `https://auth.myapp.com`.
- A public client with `client_id=mobileapp`.
- A confidential client with `client_id=inventory`.
- A Resource `https://onlinestore.myapp.com` with the following Scopes:
  - `read:orders`
  - `write:orders`
  - `delete:orders`
- A Resource `https://inventory.myapp.com` with the following Scopes:
  - `read:orders`
  - `write:orders`
  - `delete:orders`
- `https://onlinestore.myapp.com` and `https://inventory.myapp.com` are 2 separate systems. Although their scopes share the same name, the scopes are not the same.
- A Role `onlinestore:admin`. By definition, it has
  - `read:orders` of `https://onlinestore.myapp.com`.
  - `write:orders` of `https://onlinestore.myapp.com`.
  - `delete:orders` of `https://onlinestore.myapp.com`.
- A Role `inventory:admin`. By definition, it has
  - `read:orders` of `https://inventory.myapp.com`.
  - `write:orders` of `https://inventory.myapp.com`.
  - `delete:orders` of `https://inventory.myapp.com`.
- An existing User with ID `johndoe`.
  - `johndoe` is of Role `onlinestore:admin`.
  - `johndoe` is assigned `read:orders` of `https://inventory.myapp.com`.

### Use-cases: Sign in `johndoe` in `mobileapp` without `resource`

The mobile app `mobileapp` utilizes `OIDC-Core`, which in turn is based on `RFC6749 section-4.1` to authenticate end-users.
When `johndoe` signs in, `mobileapp` will perform `OIDC-Core` and receive an ID token.
The ID token looks like:

```json
{
  "iss": "https://auth.myapp.com",
  "sub": "johndoe",
  "aud": ["mobileapp"],
  "https://authgear.com/claims/user/roles": ["onlinestore:admin"],
  "scope": "openid offline_access profile email",
  "https://authgear.com/claims/scope_by_aud": [
    {
      "aud": "mobileapp",
      "scope": "openid offline_access profile email"
    }
  ]
}
```

The ID token is an assertion of who has been authenticated.
It cannot be used to access Resources.
For actual access, the `access_token` has to be used.

To ease integration between systems, `https://auth.myapp.com` implements `RFC9068`.
The `access_token` in the previous authentication of `johndoe` looks like

```json
{
  "iss": "https://auth.myapp.com",
  "sub": "johndoe",
  "aud": ["https://auth.myapp.com"],
  "client_id": "mobileapp",
  "https://authgear.com/claims/user/roles": ["onlinestore:admin"],
  "scope": "openid offline_access",
  "https://authgear.com/claims/scope_by_aud": [
    {
      "aud": "mobileapp",
      "scope": "openid offline_access"
    }
  ]
}
```

- The `aud` of this `access_token` does not include `https://onlinestore.myapp.com`.
- The `scope` of this `access_token` does not include `read:orders write:orders delete:orders`.

Before the introduction of Resource and Scope, `https://onlinestore.myapp.com` has to be programmed specifically to accept this `access_token`:

- Check if `access_token` is a JWT.
- Check if `iss` is `https://auth.myapp.com`.
- Fetch `jwks_uri` from `https://auth.myapp.com/.well-known/openid-configuration`.
- Verify if the `access_token` is signed by one of the JWK in `jwks_uri`.
- DO NOT check `aud` is `https://onlinestore.myapp.com`. Instead, check if it includes `https://auth.myapp.com`.
- Check if `roles` includes `onlinestore:admin`.

### Use-cases: Sign in `johndoe` in `mobileapp` with single `resource`

Suppose

- `mobileapp` is associated with `https://onlinestore.myapp.com`, and all the scopes of it.
- During the Authorization Code Grant, `resource=https://onlinestore.myapp.com`.

This `access_token` is obtained:

```json
{
  "iss": "https://auth.myapp.com",
  "sub": "johndoe",
  "aud": ["https://auth.myapp.com", "https://onlinestore.myapp.com"],
  "client_id": "mobileapp",
  "https://authgear.com/claims/user/roles": ["onlinestore:admin"],
  "scope": "openid offline_access read:orders write:orders delete:orders",
  "https://authgear.com/claims/scope_by_aud": [
    {
      "aud": "https://auth.myapp.com",
      "scope": "openid offline_access"
    },
    {
      "aud": "https://onlinestore.myapp.com",
      "scope": "read:orders write:orders delete:orders"
    }
  ]
}
```

The Resource `https://onlinestore.myapp.com` can now validate the `access_token` with these rules:

- Check if `access_token` is a JWT.
- Check if `iss` is `https://auth.myapp.com`.
- Fetch `jwks_uri` from `https://auth.myapp.com/.well-known/openid-configuration`.
- Verify if the `access_token` is signed by one of the JWK in `jwks_uri`.
- Check `aud` includes `https:/onlinestore.myapp.com`.
- Check `scope` includes the necessary scopes.

### Use-cases: Sign in `johndoe` in `mobileapp` with multiple `resource`

Suppose

- `mobileapp` is associated with `https://onlinestore.myapp.com`, and all the scopes of it.
- `mobileapp` is associated with `https://inventory.myapp.com`, and all the scopes of it.
- During the Authorization Code Grant, `resource=https://onlinestore.myapp.com&resource=https://inventory.myapp.com`.

This `access_token` is obtained:

```json
{
  "iss": "https://auth.myapp.com",
  "sub": "johndoe",
  "aud": ["https://auth.myapp.com", "https://onlinestore.myapp.com", "https://inventory.myapp.com"],
  "client_id": "mobileapp",
  "https://authgear.com/claims/user/roles": ["onlinestore:admin"],
  "scope": "openid offline_access read:orders",
  "https://authgear.com/claims/scope_by_aud": [
    {
      "aud": "https://auth.myapp.com",
      "scope": "openid offline_access"
    },
    {
      "aud": "https://onlinestore.myapp.com",
      "scope": "read:orders write:orders delete:orders"
    },
    {
      "aud": "https://inventory.myapp.com",
      "scope": "read:orders"
    }
  ]
}
```

Note that the `access_token` is downscoped. See [this for details](#discussion-resource-scope-client-and-downscoping)

### Use-cases: M2M of `inventory` to `https://onlinestore.myapp.com`

In `https://inventory.myapp.com`, there is a daemon process that needs access to `https://onlinestore.myapp.com`.
In particular, it needs to `read:orders`.

`https://inventory.myapp.com` use `RFC6749 section-4.4` to obtain an `access_token` to gain access.
In particular, it sends this Token request to `https://auth.myapp.com/oauth2/token`.

```
POST /oauth2/token HTTP/1.1
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials
client_id=inventory
client_secret=THE_CLIENT_SECRET
resource=https://onlinestore.myapp.com
scope=read:orders
```

`https://auth.myapp.com` process the Token request with these rules:

- Check if `grant_type` is supported, and it is `client_credentials`.
- Authenticate the client by checking `client_id` and `client_secret`.
- Validate `resource` refers to an existing Resource, and the Resource is associated with `client_id`.
- Validate `scope` and check if it refers to the valid values as defined in `resource`, and is associated with `client_id`.

`https://auth.myapp.com` will return an `access_token` like this:

```json
{
  "iss": "https://auth.myapp.com",
  "sub": "client_id_inventory",
  "aud": ["https://onlinestore.myapp.com"],
  "client_id": "inventory",
  "scope": "read:orders",
  "https://authgear.com/claims/scope_by_aud": [
    {
      "aud": "http://onlinestore.myapp.com",
      "scope": "read:orders"
    }
  ]
}
```

If you compare this `access_token` to the `access_token` of the end-user, you will notice:

|             | Description                      | `access_token` of end-user                                     | `access_token` of client                                                             |
| ---         | ---                              | ---                                                            | ---                                                                                  |
| `iss`       | Same value                       |                                                                |                                                                                      |
| `sub`       | Different value                  | The user ID (`johndoe`)                                        | The string concatenation of `client_id_` and the `client_id` (`client_id_inventory`) |
| `aud`       | The target audience of the token | The authentication server plus `resource`                      | The `resource` parameter (`https://onlinestore.myapp.com`)                           |
| `client_id` | Different value                  | The client acting on behalf of an end-user (`mobileapp`)       | The client acting on behalf of itself (`inventory`)                                  |
| `scope`     | Same meaning                     | The access of the `client_id` to `aud` on `sub`. See Remarks 1 | The access of the `client_id` to `aud` on behalf of `sub`. See Remarks 2             |
| `roles`     | Present and Absent               | The roles of `sub`                                             | Absent because Role-based access control (RBAC) does not apply to clients            |

Remarks

1. It means that `mobileapp` (`client_id`) has access to the ID token (`openid`), the `refresh_token` (`offline_access`), the information returned by the UserInfo endpoint (`profile email`) of `johndoe` (`sub`), plus all `scopes` granted by the `sub` (`johndoe`).
2. It means that `inventory` (`client_id`) has access to `https://onlinestore.myapp.com` (`aud`) limited to `read:orders` (`scope`).

At `https://onlinestore.myapp.com`, it validates this `access_token` with these rules:

- Check if `access_token` is a JWT.
- Check if `iss` is `https://auth.myapp.com`.
- Fetch `jwks_uri` from `https://auth.myapp.com/.well-known/openid-configuration`.
- Verify if the `access_token` is signed by one of the JWK in `jwks_uri`.
- Check if `aud` is `https://onlinestore.myapp.com`.
- Check if `scope` is sufficient.

### Use-cases: Handling of `access_token` in Resources

With the introduction of Resource, Scope, and their association with Client and User,

the handling of `access_token` is now consistent at the Resource.

The Resource **SHOULD** process the `access_token` with these rules:

- Check if `access_token` is a JWT.
- Check if `iss` is `https://auth.myapp.com`.
- Fetch `jwks_uri` from `https://auth.myapp.com/.well-known/openid-configuration`.
- Verify if the `access_token` is signed by one of the JWK in `jwks_uri`.
- Check `aud` includes `https://onlinestore.myapp.com`.
- Check `scope` is sufficient.

## Discussions

### Discussion: M2M in other protocol, like SAML 2.0

There is no M2M authentication and authorization related profiles in SAML 2.0.
This statement is made after reading https://groups.oasis-open.org/higherlogic/ws/public/download/56782/sstc-saml-profiles-errata-2.0-wd-07.pdf
No profiles mentioned there is relevant to M2M authentication and authorization.

When you search M2M authentication and authorization on the web, the most popular protocol is OAuth 2.0.

Therefore, this document solely discuss M2M authentication and authorization in the context of OAuth 2.0.

### Discussion: About `scope` and `access_token`

In `RFC6749 section-1.4`, `access_token` is

> Access tokens are credentials used to access protected resources.
> An access token is a string representing an authorization issued to the client.

In `RFC6749 section-3.3`, `scope` is

> The authorization and token endpoints allow the client to specify the scope of the access request using the "scope" request parameter.
> In turn, the authorization server uses the "scope" response parameter to inform the client of the scope of the access token issued.

In my own interpretation:

- **`scope` means the access of `client_id` to `aud` acting on behalf of `sub`**.
- The access of `sub` on `aud` explicitly granted by `sub` to `client_id` can also appear in `scope`.

### Discussion: Resource and Scope

To support M2M, we need to introduce Resource and its associated Scope.

Per `RFC8707`, resource has to be identified with a URI.
Therefore, it follows naturally that we mandate a Resource identified by a non-modifiable URI.
Both Auth0 and Kinde disallow changing this URI identifier, so it should be a sane design decision.

The URI of a Resource must satisfy the following requirements:

- It is a URI as defined in [RFC3986](https://datatracker.ietf.org/doc/html/rfc3986).
- It must be unique within a Project.
- It must be of `https:` scheme.
- It must not be a subdomain of the default domains of Authgear. For example, if Authgear has default domains `authgearapps.com` and `authgear.cloud`, then its domain must not be those, of subdomains of those.
- It can optionally have a path component. For example, both `https://api.myapp.com` and `https://api.myapp.com/` are valid. They are treated as different Resources though. No path normalization is taken.
- It must not have a query component. For example, `https://api.myapp.com?a=b` is NOT a valid URI of a Resource.
- It must not have a fragment component. For example, `https://api.myapp.com#a` is NOT a valid URI of a Resource.
- It must not have a userinfo component. For example, `https://username:password@api.myapp.com` is NOT a valid URI of a Resource.

To maximize the compatibility of M2M with a wide range of software,
we make Scope local to a specific Resource.
This means the `read:orders` of `https://onlinestore.myapp.com` is different from `https://inventory.myapp.com`.

### Discussion: Allowed values of Scope

It is observed that `OIDC-Core` defined scopes **CANNOT** be used to create Permission of API in Auth0.

A list of well-known scopes:

- [openid](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest)
- [profile email address phone](https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims)
- [offline_access](https://openid.net/specs/openid-connect-core-1_0.html#OfflineAccess)
- [device_sso](https://openid.net/specs/openid-connect-native-sso-1_0.html#section-3.1)

The Scope of a Resource must satisfy the following requirements:

- It must not be one of the following: `openid`, `profile`, `email`, `address`, `phone`, `offline_access`, or `device_sso`.
- It must not start with `https://authgear.com`.
- It must be valid for the grammar defined in [RFC6749 section-3.3](https://datatracker.ietf.org/doc/html/rfc6749#section-3.3)

### Discussion: Resource, Scope, Client, and downscoping

To strictly follow [About scope and access_token](#discussion-about-scope-and-accesstoken),
**explicit** association between Resource and Client are **required**,
regardless of whether the client is **public** or **confidential**, **first-party** or **third-party**.

For the following examples, assume

- `johndoe` is `onlinestore:admin` **but not** `inventory:admin`.
- `johndoe` is granted `read:orders` of `https://inventory.myapp.com`.

---

If the developer wants to obtain an `access_token` of `johndoe` like the following:

```json
{
  "iss": "https://auth.myapp.com",
  "sub": "johndoe",
  "aud": ["https://auth.myapp.com", "https://onlinestore.myapp.com"],
  "client_id": "mobileapp",
  "https://authgear.com/claims/user/roles": ["onlinestore:admin"],
  "scope": "openid offline_access profile email read:orders write:orders delete:orders",
  "https://authgear.com/claims/scope_by_aud": [
    {
      "aud": "https://auth.myapp.com",
      "scope": "openid offline_access profile email"
    },
    {
      "aud": "https://onlinestore.myapp.com",
      "scope": "read:orders write:orders delete:orders"
    }
  ]
}
```

The developer has to:

- Assign `read:orders`, `write:orders`, `delete:orders` of `https://onlinestore.myapp.com` to `mobileapp`, even `mobileapp` is a public client, not a confidential client.
  - This is different from Auth0. Auth0 only allows assigning Permissions to confidential clients.
  - Note that `mobileapp` is a public client. Even it has been assigned the scope, it cannot obtain an `access_token` with Client Credentials Grant due to the fact that it lacks `client_secret`.
- Specify `resource=https://onlinestore.myapp.com` in the authentication request.

---

If the developer wants to obtain an `access_token` of `johndoe` like the following by doing

- Assign `read:orders`, `write:orders`, `delete:orders` of `https://onlinestore.myapp.com` to `mobileapp`, even `mobileapp` is a public client, not a confidential client.
- Specify `resource=https://onlinestore.myapp.com` in the authentication request.
- Specify `resource=https://inventory.myapp.com` in the authentication request.

```json
{
  "iss": "https://auth.myapp.com",
  "sub": "johndoe",
  "aud": ["https://auth.myapp.com", "https://onlinestore.myapp.com", "https://inventory.myapp.com"],
  "client_id": "mobileapp",
  "https://authgear.com/claims/user/roles": ["onlinestore:admin"],
  "scope": "openid offline_access profile email read:orders write:orders delete:orders"
}
```

It is **IMPOSSIBLE** because [RFC8707 section-2.2](https://datatracker.ietf.org/doc/html/rfc8707#section-2.2) specifies that `aud` and `scope` are "the cartesian product of all the scopes at all the target services".

Instead, the resulting `access_token` will be "downscoped" (downscope is a term in RFC8707 section-2.2) to:

```json
{
  "iss": "https://auth.myapp.com",
  "sub": "johndoe",
  "aud": ["https://auth.myapp.com", "https://onlinestore.myapp.com", "https://inventory.myapp.com"],
  "client_id": "mobileapp",
  "https://authgear.com/claims/user/roles": ["onlinestore:admin"],
  "scope": "openid offline_access profile email read:orders",
  "https://authgear.com/claims/scope_by_aud": [
    {
      "aud": "https://auth.myapp.com",
      "scope": "openid offline_access"
    },
    {
      "aud": "https://onlinestore.myapp.com",
      "scope": "read:orders write:orders delete:orders"
    },
    {
      "aud": "https://inventory.myapp.com",
      "scope": "read:orders"
    }
  ]
}
```

If downscoping is unwanted, the developer **MUST**:

- Read `https://authgear.com/claims/scope_by_aud` instead. It is a JSON array of objects. Each object has `aud` and `scope`, which is the `scope` of `client_id` to `aud` on behalf of `sub`.
- Use unique scope, for example, prepend the Resource URI before the scope, like [what Google does](https://developers.google.com/identity/protocols/oauth2/scopes).
- Do not mix `resource` that could lead to downscoping. Instead, use Token Exchange to obtain non-ambiguous `access_token`. Or simply use one `access_token` per `resource`.

### Discussion: Client Credentials Grant and downscoping

```json
{
  "iss": "https://auth.myapp.com",
  "sub": "client_id_inventory",
  "aud": ["https://onlinestore.myapp.com"],
  "client_id": "inventory",
  "scope": "read:orders write:orders",
  "https://authgear.com/claims/scope_by_aud": [
    {
      "aud": "https://onlinestore.myapp.com",
      "scope": "read:orders write:orders",
    }
  ]
}
```

The above `access_token` can `read:orders` **AND** `write:orders` on `https://onlinestore.myapp.com`.

---

```json
{
  "iss": "https://auth.myapp.com",
  "sub": "client_id_inventory",
  "aud": ["https://inventory.myapp.com"],
  "client_id": "inventory",
  "scope": "read:orders",
  "https://authgear.com/claims/scope_by_aud": [
    {
      "aud": "https://inventory.myapp.com",
      "scope": "read:orders",
    }
  ]
}
```

The above `access_token` can `read:orders` on `https://inventory.myapp.com`.

---

```json
{
  "iss": "https://auth.myapp.com",
  "sub": "client_id_inventory",
  "aud": ["https://onlinestore.myapp.com", "https://inventory.myapp.com"],
  "client_id": "inventory",
  "scope": "read:orders",
  "https://authgear.com/claims/scope_by_aud": [
    {
      "aud": "https://onlinestore.myapp.com",
      "scope": "read:orders write:orders",
    },
    {
      "aud": "https://inventory.myapp.com",
      "scope": "read:orders",
    }
  ]
}
```

The above `access_token` is downscoped to `read:orders`.
The developer **MUST** read `https://authgear.com/claims/scope_by_aud` to get the `scope` by `aud`.

### Discussion: Resource, Scope, Client, consent

Originally we have a table `_auth_oauth_authorization` with unique index on `(app_id, user_id, client_id)`.This table models the authorization of `aud` being the authorization server itself.

Now that we have introduced Resource, we need a new table `_auth_oauth_authorization_resource` with unique index on `(app_id, user_id, client_id, resource_id)`.

In first-party client, consent is not asked and authorization is implicitly granted.

In third-party client, authorization is computed and compared to existing authorization.
If authorization does not exist at all, or scope has changes, consent is prompted.

For example, in order to issue the following `access_token`:

```json
{
  "iss": "https://auth.myapp.com",
  "sub": "johndoe",
  "aud": ["https://auth.myapp.com", "https://onlinestore.myapp.com", "https://inventory.myapp.com"],
  "client_id": "thirdpartyclient",
  "scope": "openid offline_access profile email read:orders write:orders delete:orders"
}
```

The following authorizations are computed:

- `(johndoe, thirdpartyclient, https://onlinestore.myapp.com)`: `read:orders write:orders delete:orders`.
- `(johndoe, thirdpartyclient, https://inventory.myapp.com)`: `read:orders write:orders delete:orders`.

Suppose the existing authorizations are:

- `(johndoe, thirdpartyclient, https://onlinestore.myapp.com)`: `read:orders`.

Then, the following consent screen is shown to `johndoe`:

```
The application `thirdpartyclient` requests:

- `https://onlinestore.myapp.com`:
  - `write:orders`
  - `delete:orders`
- `https://inventory.myapp.com`:
  - `read:orders`
  - `write:orders`
  - `delete:orders`

[Allow] [Reject]
```

Allowing the end-user to allow a subset of the requested scope **IS NOT** supported.

### Discussion: Signify the intention of granting all scopes to a client

In Auth0, even you have granted all existing Permissions to a client,
when a new permission is added to an API,
the newly added permission **IS NOT** automatically granted to the client.
You have to go through all clients and grant the newly added permission for each client.

If this becomes tedious, we can introduce a new boolean column `all_scopes_are_granted` in `_auth_client_resource`.
When this column is NULL or false, the source of truth of granted scopes is `_auth_client_resource_scope`.
When this column is true, all of the scopes of the resource is automatically granted to the client.

Since this is a new column, we can introduce this in the future, no need to include this in the MVP.

### Discussion: (Auth0) General facts on Auth0 M2M

In Auth0 M2M authentication and authorization, the following concepts must be known first:

- There are API, Permission, Application, User, Role, Organization.
- An API has 0 or more Permissions.
- Each Permission is associated with exactly one API.
- Permission belonging to different API **CAN** share the same name. They ARE NOT considered equal.
- A M2M Application **MUST** have at least one API.
- A M2M Application **MAY** have no granted Permissions from any APIs associated with it.
- A Role can be associated with a Permission of an API **globally**.
- A User can be associated with a Permission of an API **directly** **globally**.
- A User can be associated with a Role **organizationally**. In this case, the User inherit the Permissions the Role has **organizationally**.
- A M2M Application can further has its access limited **organizationally** if you are on a paid plan. See https://auth0.com/docs/manage-users/organizations/organizations-for-m2m-applications/configure-your-application-for-m2m-access#define-organization-behavior

Other facts:

- In Authorization Code Flow, if `audience` is unspecified, the returned `access_token` is **NOT** a JWT. See https://community.auth0.com/t/opaque-versus-jwt-access-token/31028
- The `permissions` claim in JWT `access_token` is present only if the API has "Enable RBAC" **AND** "Add Permissions in the Access Token".
- In Authorization Code Grant, Permissions appear as `permissions`.
- In Client Credentials Grant, Permissions appear as `scope`.
- In Client Credentials Grant, `audience` is required and can only appear once. It replaces `resource`.
  - The implies the JWT `access_token` associates with one and only one `audience`.
  - Since `audience` is singleton, `scope` is never ambiguous.

Facts on Organizations:

- The **globally** assigned Roles and Permissions to a user is ignored. `permissions` is always an empty array. See https://community.auth0.com/t/organization-permissions-claim-empty/99135

References:

- https://auth0.com/docs/get-started/authentication-and-authorization-flow/client-credentials-flow/call-your-api-using-the-client-credentials-flow#request-tokens
- https://community.auth0.com/t/opaque-versus-jwt-access-token/31028
- https://community.auth0.com/t/organization-permissions-claim-empty/99135
- https://auth0.com/docs/manage-users/organizations/organizations-for-m2m-applications/configure-your-application-for-m2m-access#define-organization-behavior

### Discussion: (Auth0) Example of JWT access token of Client Credentials Grant

```json
{
  "iss": "https://dev-fnc259uk.auth0.com/",
  "sub": "62Amt65MWTRwkoySX65O8JVT735c8UI2@clients",
  "aud": "myapi",
  "iat": 1751011899,
  "exp": 1751098299,
  "scope": "write:users",
  "jti": "jsBXsU6pXyB29Bo2RVG3Ki",
  "client_id": "62Amt65MWTRwkoySX65O8JVT735c8UI2",
  "permissions": [
    "write:users"
  ]
}
```

### Discussion: (Auth0) Example of JWT access token of Authorization Code Grant

```json
{
  "iss": "https://dev-fnc259uk.auth0.com/",
  "sub": "auth0|683d622f1a6ac540d22d1409",
  "aud": [
    "myapi",
    "https://dev-fnc259uk.auth0.com/userinfo"
  ],
  "iat": 1750926072,
  "exp": 1751012472,
  "scope": "openid profile email",
  "org_id": "org_zsW1uJZxKUryA8kB",
  "jti": "8VJ3PDNU93FFJNkySb26Wh",
  "client_id": "STtVMnNqdKcO7GzO8mYvpvkKgOucKFVo",
  "permissions": [
    "write:users"
  ]
}
```

### Discussion: (Auth0) Third party public client and Organization

Third party public client and organization are mutually exclusive.
For example, if you try to set `is_first_party=false` of an organization-aware client via the Management API,
you will get this error.

```json
{
  "statusCode": 403,
  "error": "Forbidden",
  "message": "Organizations are only available to first party clients on user-based flows. Properties organization_usage and organization_require_behavior must be unset for third party clients.",
  "errorCode": "invalid_body"
}
```

**This implies Third party client and Organization are mutually exclusive features.**

### Discussion: (Auth0) Third party public client and Connection

- Third party public client cannot be associated to any connections **explicitly**.
- Instead, third party public client are associated **implicitly** to all `is_domain_connection=true` connections.
- To set `is_domain_connection=true` for a given connection, you do it via the Management API.

See https://community.auth0.com/t/error-enabling-domain-connection-for-a-third-party-application/188320

### Discussion: (Auth0) Third party public client and consent

Given that

- The API has "Enable RBAC" **AND** "Add Permissions in the Access Token".
- The User has been assigned Permissions of the API.
- Note that public client **CANNOT** be assigned Permissions of the API. Only confidential client can be assigned Permissions of the API.
- `audience` is specified in the authorization request.

We have:

- The consent screen **WILL NOT** ask for the consent of the API.
- In other words, the consent screen only asks for `scope`.
- The `access_token` will have `permissions` populated according to what the User has on the `audience`.

This implies:

- `permissions` is not treated as `scope`.
- Consent applies to `scope` only.
- Auth0 does not really support consent on the Permissions assigned to the User.

### Discussion: (Auth0) Multiple audience

As mentioned in [General Facts](#discussion-auth0-general-facts-on-auth0-m2m), `audience` is single,
it is impossible to generate an `access_token` that can be used in multiple APIs.

To facilitate the multiple-audience use case, Auth0 instead offers a Enterprise-only implementation of [RFC8693: OAuth 2.0 Token Exchange](https://datatracker.ietf.org/doc/html/rfc8693).
This allows the developer to [get access token for another audience](https://auth0.com/docs/authenticate/custom-token-exchange#use-case-get-auth0-tokens-for-another-audience)

### Discussion: (Auth0) The `permissions` claim

Given that `permissions` is turned on according to [General Facts](#discussion-auth0-general-facts-on-auth0-m2m):

- The API can simply rely on `permissions` to determine access.
- The API **MUST** handle `sub` conditionally, by checking whether `sub` ends with `@clients` or not.

### Discussion: (Authgear) Proposed Token Exchange behavior

The `access_token` of `grant_type=authorization_code`, that is, the `subject_token`:

```json
{
  "iss": "https://auth.myapp.com",
  "sub": "johndoe",
  "aud": ["https://auth.myapp.com"],
  "client_id": "mobileapp",
  "https://authgear.com/claims/user/roles": ["onlinestore:admin"],
  "scope": "openid offline_access"
}
```

This `access_token` can be submitted to resources which are programmed to accept `aud=https://auth.myapp.com`, AND read `https://authgear.com/claims/user/roles` to determine the access.

For resources that always validate `aud`, a [RFC8693: OAuth 2.0 Token Exchange](https://datatracker.ietf.org/doc/html/rfc8693) is required to obtain an `access_token` with the intended `aud`.

Suppose the `access_token` is submitted to `https://inventory.myapp.com`, which has specially cased to handle `aud=https://auth.myapp.com`. Now `https://inventory.myapp.com` wants to access `https://onlinestore.myapp.com` on behalf of `johndoe`, it needs to perform a Token Exchange.

To perform a Token Exchange, `https://inventory.myapp.com` has to obtain an `access_token` representing itself first. This is essentially the `access_token` obtained with Client Credentials Grant.

The request to `https://auth.myapp.com` to obtain the `actor_token`.

```
POST /oauth2/token HTTP/1.1
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials
client_id=inventory
client_secret=THE_CLIENT_SECRET
resource=https://onlinestore.myapp.com
```

The `actor_token`:

```json
{
  "iss": "https://auth.myapp.com",
  "sub": "client_id_inventory",
  "aud": ["https://onlinestore.myapp.com"],
  "client_id": "inventory",
  "scope": "read:orders"
}
```

The Token Exchange request:

```
POST /oauth2/token HTTP/1.1
Content-Type: application/x-www-form-urlencoded

grant_type=urn:ietf:params:oauth:grant-type:token-exchange
resource=https://onlinestore.myapp.com
requested_token_type=urn:ietf:params:oauth:token-type:access_token
subject_token=SUBJECT_TOKEN
subject_token_type=urn:ietf:params:oauth:token-type:access_token
actor_token=ACTOR_TOKEN
actor_token_type=urn:ietf:params:oauth:token-type:access_token
```

The exchanged token:

```json
{
  "iss": "https://auth.myapp.com",
  "sub": "johndoe",
  "aud": ["https://onlinestore.myapp.com"],
  "client_id": "inventory",
  "https://authgear.com/claims/user/roles": ["onlinestore:admin"],
  "scope": "read:orders write:orders delete:orders"
}
```

Note that

- The `scope` of `subject_token` is `openid offline_access`.
- The `aud` of `subject_token` is `https://auth.myapp.com`, which is not acceptable by `https://onlinestore.myapp.com`.
- The `scope` of `actor_token` is `read:orders`, so `https://inventory.myapp.com` does not have admin access on its own.
- The `scope` of exchanged token is `read:orders write:orders delete:orders`. It means that when `https://inventory.myapp.com` acting on behalf of `johndoe`, who is `onlinestore:admin`, has additional `scope` inherit from `sub` (`johndoe`).

### Discussion: (Auth0) Rich Authorization Requests

[RFC9396](https://datatracker.ietf.org/doc/html/rfc9396) introduces `authorization_details`,
which is an enhancement over `scope` and `resource` to specify authorization details in a structured way in form of JSON.

This is supported by Auth0 as an Add-on. See https://auth0.com/docs/get-started/apis/configure-rich-authorization-requests

### Discussion: (Future work) Restricting access to Admin API

In Auth0, `RFC6749 section-4.4` is also used to create an `access_token` that can be used to access the Management API.

We can implement the same for our Admin API.
Specifically, we need to introduce an artificial Resource with URI `https://auth.myapp.com/_api/admin`, and define a comprehensive list of scopes that restrict access to the different Resources within the Admin API.

## MVP

In the MVP, we are going to implement M2M only.
This means

- The table `__auth_oauth_authorization_resource` **IS NOT** added.
- The `resource` parameter in the Authorization endpoint **IS NOT** implemented.
- The grant_type `client_credentials` **IS** supported in the Token endpoint.
- The `resource` parameter and `scope` parameter **ARE** supported when `grant_type=client_credentials`.
- The relevant CRUD of Resource, Scope, and Client **ARE** supported.

### Changes in data models

Here are the schema of the changes:

```sql
CREATE TABLE _auth_resource (
  id text PRIMARY KEY,
  app_id text NOT NULL,
  created_at timestamp without time zone NOT NULL,
  updated_at timestamp without time zone NOT NULL,
  uri text NOT NULL,
  name text,
  metadata jsonb
);
-- Each project has its own set of Resources. The URI must be unique within a project.
CREATE UNIQUE INDEX _auth_resource_uri_unique ON _auth_resource USING btree (app_id, uri);
-- Support typeahead search
CREATE INDEX _auth_resource_uri_typeahead ON _auth_resource USING btree (app_id, uri text_pattern_ops);
CREATE INDEX _auth_resource_name_typeahead ON _auth_resource USING btree (app_id, name text_pattern_ops);

CREATE TABLE _auth_resource_scope (
  id text PRIMARY KEY,
  app_id text NOT NULL,
  created_at timestamp without time zone NOT NULL,
  updated_at timestamp without time zone NOT NULL,
  resource_id text NOT NULL REFERENCES _auth_resource(id),
  scope text NOT NULL,
  description text,
  metadata jsonb
);
-- Each Resource has its own set of Scopes. The scope must be unique within a Resource.
CREATE UNIQUE INDEX _auth_resource_scope_unique ON _auth_resource_scope USING btree (app_id, resource_id, scope);
-- Support typeahead search
CREATE INDEX _auth_resource_scope_scope_typeahead ON _auth_resource_scope USING btree (app_id, resource_id, scope text_pattern_ops);

CREATE TABLE _auth_client_resource (
  id text PRIMARY KEY,
  app_id text NOT NULL,
  created_at timestamp without time zone NOT NULL,
  updated_at timestamp without time zone NOT NULL,
  -- Since client is not stored in the database, it is not a foreign key.
  client_id text NOT NULL,
  resource_id text NOT NULL REFERENCES _auth_resource(id),
);
-- Each Client can only associate with a Resource once.
CREATE UNIQUE INDEX _auth_client_resource_unique ON _auth_client_resource USING btree (app_id, client_id, resource_id);

CREATE TABLE _auth_client_resource_scope (
  id text PRIMARY KEY,
  app_id text NOT NULL,
  created_at timestamp without time zone NOT NULL,
  updated_at timestamp without time zone NOT NULL,
  -- Since client is not stored in the database, it is not a foreign key.
  client_id text NOT NULL,
  resource_id text NOT NULL REFERENCES _auth_resource(id),
  scope_id text NOT NULL REFERENCES _auth_resource_scope(id)
);
-- Each Client can only associate with a Resource scope once.
CREATE UNIQUE INDEX _auth_client_resource_scope_unique ON _auth_client_resource USING btree (app_id, client_id, resource_id, scope_id);

-- A sibling table of _auth_oauth_authorization, that takes resource_id into account.
CREATE TABLE _auth_oauth_authorization_resource (
  id text PRIMARY KEY,
  app_id text NOT NULL,
  client_id text NOT NULL,
  user_id text NOT NULL REFERENCES _auth_user(id),
  resource_id text NOT NULL REFERENCES _auth_resource(id),
  scope_id text NOT NULL REFERENCES _auth_resource_scope(id)
)
CREATE UNIQUE INDEX _auth_oauth_authorization_resource_unique ON _auth_oauth_authorization_resource USING btree (app_id, client_id, user_id, resource_id, scope_id);
```

### Changes in Admin API

The changes are mainly the CRUD of Resources and Scopes.

The CRUD of Resources and Scopes DO NOT generate events.

The following GraphQL schema snippet describe the changes to the Admin API GraphQL schema.

```graphql
type Query {
  """If clientID is null, then all resources are returned in a paginated fashion."""
  """If clientID is specified, then all resources associated with the clientID are returned in a paginated fashion."""
  """If searchKeyword is non-null, a prefix search of uri or name is performed."""
  """If both clientID and searchKeyword are specified, they are AND-ed."""
  resources(clientID: String, searchKeyword: String, after: String, before: String, first: Int, last: Int): ResourceConnection
}

type Mutation {
  createResource(input: CreateResourceInput!): CreateResourcePayload!
  updateResource(input: UpdateResourceInput!): UpdateResourcePayload!
  deleteResource(input: DeleteResourceInput!): DeleteResourcePayload!

  createScope(input: CreateScopeInput!): CreateScopePayload!
  updateScope(input: UpdateScopeInput!): UpdateScopePayload!
  deleteScope(input: DeleteScopeInput!): DeleteScopePayload!

  addResourceToClientID(input: AddResourceToClientIDInput!): AddResourceToClientIDPayload!
  removeResourceFromClientID(input: RemoveResourceFromClientIDInput!): RemoveResourceFromClientIDPayload!
  addScopesToClientID(input: AddScopesToClientIDInput!): AddScopesToClientIDPayload!
  removeScopesFromClientID(input: RemoveScopesFromClientIDInput!): RemoveScopesFromClientIDPayload!
  replaceScopesOfClientID(input: ReplaceScopesOfClientIDInput!): ReplaceScopesOfClientIDPayload!
}

type Resource implements Entity & Node {
  id: ID!
  createdAt: DateTime!
  updatedAt: DateTime!
  uri: String!
  name: String
  """If clientID is null, then all scopes of this Resource is returned."""
  """If clientID is specified, then only scopes that are associated with clientID is returned."""
  """If searchKeyword is non-null, a prefix search of scope is performed."""
  """If both clientID and searchKeyword are specified, they are AND-ed."""
  scopes(clientID: String, searchKeyword: String, after: String, before: string, first: Int, last: Int): ScopeConnection
  """The list of client IDs associated with this Resource."""
  clientIDs: [String!]!
}

type Scope implements Entity & Node {
  id: ID!
  createdAt: DateTime!
  updatedAt: DateTime!
  resourceID: ID!
  scope: String!
  description: String
}

type ResourceEdge {
  cursor: String!
  resource: Resource
}

type ResourceConnection {
  edges: [ResourceEdge]
  pageInfo: PageInfo!
  totalCount: Int
}

type ScopeEdge {
  cursor: String!
  scope: Scope
}

type ScopeConnection {
  edges: [ScopeEdge]
  pageInfo: PageInfo!
  totalCount: Int
}

input CreateResourceInput {
  uri: String!
  name: String
}

type CreateResourcePayload {
  resource: Resource!
}

input UpdateResourceInput {
  resourceURI: String!
  """The new name"""
  name: String
}

type UpdateResourcePayload {
  resource: Resource!
}

input DeleteResourceInput {
  resourceURI: String!
}

type DeleteResourcePayload {
  ok: Boolean
}

input CreateScopeInput {
  resourceURI: String!
  scope: String!
  description: String
}

type CreateScopePayload {
  scope: Scope!
}

input UpdateScopeInput {
  resourceURI: String!
  scope: String!
  """The new description"""
  description: String
}

type UpdateScopePayload {
  scope: Scope!
}

input DeleteScopeInput {
  resourceURI: String!
  scope: String!
}

type DeleteScopePayload {
  ok: Boolean
}

input AddResourceToClientIDInput {
  resourceURI: String!
  clientID: String!
}

type AddResourceToClientIDPayload {
  resource: Resource!
}

input RemoveResourceFromClientIDInput {
  resourceURI: String!
  clientID: String!
}

type RemoveResourceFromClientIDPayload {
  resource: Resource!
}

input AddScopesToClientIDInput {
  resourceURI: String!
  scopes: [String!]!
  clientID: String!
}

type AddScopesToClientIDPayload {
  scopes: [Scope!]!
}

input RemoveScopesFromClientIDInput {
  resourceURI: String!
  scopes: [String!]!
  clientID: String!
}

type RemoveScopesFromClientIDPayload {
  scopes: [Scope!]!
}

input ReplaceScopesOfClientIDInput {
  resourceURI: String!
  clientID: String!
  scopes: [String!]!
}

type ReplaceScopesOfClientIDPayload {
  scopes: [Scope!]!
}
```

### Changes in OAuth 2.0 implementation

This section describes the protocol-level changes.
You do not need to read this if you are not interested.

#### The request

Per `RFC6749 section-4.4`, the request is sent to the Token endpoint.

```
POST /oauth2/token HTTP/1.1
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials
client_id=THE_CLIENT_ID
client_secret=THE_CLIENT_SECRET
resource=https://api.myapp.com
scope=scopea+scopeb+scopec
```

- `grant_type`: (Required) It must be `client_credentials`.
- `client_id`: (Required) The `client_id`, as defined by `RFC6749 section-2.3.1`.
- `client_secret`: (Required) The `client_secret`, as defined by `RFC6749 section-2.3.1`. This is consistent with `token_endpoint_auth_methods_supported=[none,client_secret_post]`.
- `resource`: (Required) The `resource` parameter defined by `RFC8707`. It must appear once and only once.
- `scope`: (Optional) The `scope` parameter defined by `RFC6749 section-4.4`. It is optional. When it is present, validation is done to ensure it is valid with respect to `resource`.

#### The successful response

```json
{
  "access_token": "A_JWT",
  "token_type": "bearer",
  "expires_in": 3600,
  "scope": "scopea+scopeb+scopec"
}
```

- `access_token`: A JWT access token conforming to `RFC9068` if `issue_jwt_access_token=true`, otherwise an opaque access token. Note that an opaque access token is not very useful because Authgear does not support [RFC7662 Token Introspection](https://datatracker.ietf.org/doc/html/rfc7662). The developer **SHOULD** turn on `issue_jwt_access_token`.
- `token_type`: The only value is `bearer` at the moment.
- `expires_in`: The lifetime in seconds of `access_token`. It follows the configuration of the client.
- `scope`: Per `RFC6749 section3.3`, it is the actual scope granted to the client. Before, M2M, it never appears in the Token Response. After the introduction of M2M, it always appear, even if the actual scope is the same as the requested scope. When the original request does not specify `scope`, the value is all scopes granted to the client. Note that `scope` also appears when `grant_type=authorization_code`.

#### The access token

Header
```json
{
  "alg": "RS256",
  "typ": "at+jwt",
  "kid": "blahblahblah"
}
```

- `alg`: It is `RS256` at the moment.
- `typ`: Per `RFC9068`, it is `at+jwt`.
- `kid`: The `kid` of the JWK key that was used to sign this JWT. The JWK key is retrieved from `jwks_uri`, as defined by `RFC8414`.

Payload
```json
{
  "jti": "AN_OPAQUE_STRING"
  "iss": "https://myapp.authgear.cloud",
  "aud": ["https://api.myapp.com"],
  "sub": "client_id_CLIENT_ID",
  "iat": 1750919927,
  "exp": 1751006327,
  "client_id": "CLIENT_ID",
  "scope": "scopea+scopeb"
}
```

- `jti`: A unique identifier for this JWT, as defined by `RFC7519 section-4.1.7`.
- `iss`: The endpoint of Authgear.
- `aud`: An array of string containing the `resource` parameter in the request.
- `sub`: The string concatenation of `client_id_` and the `client_id` parameter in the request.
- `iat`: The issue time of this JWT, as defined by `RFC7519 section-4.1.6`.
- `exp`: The expiration time of this JWT, as defined by `RFC7519 section-4.1.4`.
- `client_id`: The `client_id` parameter in the request, as defined by `RFC8693 section-4.3`.
- `scope`: Same as the `scope` in the token response.

> [!NOTE]
> The value of the `sub` is a string concatenation to mitigate the risk mentioned in https://datatracker.ietf.org/doc/html/rfc9068#section-5

#### The error response

- `invalid_grant`: When `grant_type` is invalid. Actually if the grant_type is, for example, `refresh_token`, the entire different flow is run.
- `invalid_client`: When `client_id` is invalid, or `client_secret` is invalid.
- `invalid_resource`: When `resource` is invalid, or it is not granted to `client_id`.
- `invalid_scope`: When `scope` is invalid with respect to the combination of `client_id` and `resource`. That is, the requested scope of `resource` is not granted to `client_id`.

### Changes in SDKs

There is no changes in SDKs.
The changes are in the Token endpoint,
which are supposed to be integrated by the developer using HTTP directly.

### Changes in Documentation

We will have the following documentation changes:

- Document Resource and its Scopes.
  - What is the motivation and the use-cases?
  - What is Resource and what is Scope?
  - How can I create them on the portal?
  - How can I create an `access_token` in one of my backend server?
  - How can I consume the `access_token` in another backend server?
