# M2M authentication and authorization

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

This section is an overview of how each use-case relevant to M2M looks like.

To begin the story, let us assume we have

- A company owning the domain `myapp.com`.
- The Authentication and Authorization server `https://auth.myapp.com`.
- A pre-registered client with `client_id=mobileapp`.
- A pre-registered client with `client_id=inventory`.
- A Role `onlinestore:admin`.
- A Role `inventory:admin`.
- An existing User with ID `johndoe`.
  - `johndoe` is of Role `onlinestore:admin`.
- A Resource `https://onlinestore.myapp.com` with the following Scopes:
  - `read:orders`
  - `write:orders`
  - `delete:orders`
- A Resource `https://inventory.myapp.com` with the following Scopes:
  - `read:orders`
  - `write:orders`
  - `delete:orders`
- `https://onlinestore.myapp.com` and `https://inventory.myapp.com` are 2 separate systems. Although their scopes share the same name, the scopes are not the same.

### Use-cases: Authentication of end-user, and RBAC authorization of end-user

The mobile app `mobileapp` utilizes `OIDC-Core`, which in turn is based on `RFC6749 section-4.1` to authenticate end-users.
When `johndoe` signs in, `mobileapp` will perform `OIDC-Core` and receive an ID token.
The ID token looks like:

```json
{
  "iss": "https://auth.myapp.com",
  "sub": "johndoe",
  "aud": ["mobileapp"],
  "https://authgear.com/claims/user/roles": ["onlinestore:admin"],
  "scope": "openid offline_access profile email"
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
  "scope": "openid offline_access"
}
```

You should notice that the `aud` is different.
For ID token, the `aud` is `mobileapp`.
For the access token, the `aud` is `https://auth.myapp.com`.
The access token can be used to access `https://auth.myapp.com/oauth2/userinfo`,
thus it follows logically that `aud` is `https://auth.myapp.com`.

The mobile app `mobileapp` allows the authenticated user to manage online store orders.
This means the `access_token` will be sent to `https://onlinestore.myapp.com`.

At `https://onlinestore.myapp.com`, it validates `access_token` with these rules:

- Check if `access_token` is a JWT.
- Check if `iss` is `https://auth.myapp.com`.
- Fetch `jwks_uri` from `https://auth.myapp.com/.well-known/openid-configuration`.
- Verify if the `access_token` is signed by one of the JWK in `jwks_uri`.
- DO NOT check `aud` is `https://onlinestore.myapp.com`. Instead, check if it includes `https://auth.myapp.com`.
- Check if `roles` includes `onlinestore:admin`.

### Use-cases: Authentication of client, and OAuth 2.0 scope-based authorization of client

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
- Validate `resource` appear exactly once, and it refers to an existing Resource.
- Validate `scope` and check if it refers to the valid values as defined in `resource`.
- Determine whether `client_id` is authorized to access `scope` of `resource`, using its prior knowledge.

`https://auth.myapp.com` will return an `access_token` like this:

```json
{
  "iss": "https://auth.myapp.com",
  "sub": "client_id_inventory",
  "aud": ["https://onlinestore.myapp.com"],
  "client_id": "inventory",
  "scope": "read:orders"
}
```

If you compare this `access_token` to the `access_token` of the end-user, you will notice:

|             | Description        | `access_token` of end-user                                     | `access_token` of client                                                             |
| ---         | ---                | ---                                                            | ---                                                                                  |
| `iss`       | Same               |                                                                |                                                                                      |
| `sub`       | Different          | The user ID (`johndoe`)                                        | The string concatenation of `client_id_` and the `client_id` (`client_id_inventory`) |
| `aud`       | Different          | The authentication server (`https://auth.myapp.com`)           | The `resource` parameter (`https://onlinestore.myapp.com`)                           |
| `client_id` | Different          | The client acting on behalf of an end-user (`mobileapp`)       | The client acting on behalf of itself (`inventory`)                                  |
| `scope`     | Different          | The access of the `client_id` to `aud` on `sub`. See Remarks 1 | The access of the `client_id` to `aud`. See Remarks 2                                |
| `roles`     | Present and Absent | The roles of `sub`                                             | Absent because Role-based access control (RBAC) does not apply to clients            |

Remarks

1. It means that `mobileapp` (`client_id`) has access to the ID token (`openid`), the `refresh_token` (`offline_access`), the information returned by the UserInfo endpoint (`profile email`) of `johndoe` (`sub`).
2. It means that `inventory` (`client_id`) has access to `https://onlinestore.myapp.com` (`aud`) limited to `read:orders` (`scope`).

At `https://onlinestore.myapp.com`, it validates this `access_token` with these rules:

- Check if `access_token` is a JWT.
- Check if `iss` is `https://auth.myapp.com`.
- Fetch `jwks_uri` from `https://auth.myapp.com/.well-known/openid-configuration`.
- Verify if the `access_token` is signed by one of the JWK in `jwks_uri`.
- Check if `aud` is `https://onlinestore.myapp.com`.
- Check if `scope` is sufficient to access the Resource the `access_token` is trying to access.

### Use-cases: Handling of `access_token` in Resources

If we combine handling of `access_token` of end-user, and `access_token` of client, we end up with these rules:

- Check if `access_token` is a JWT.
- Check if `iss` is `https://auth.myapp.com`.
- Fetch `jwks_uri` from `https://auth.myapp.com/.well-known/openid-configuration`.
- Verify if the `access_token` is signed by one of the JWK in `jwks_uri`.
- Check `aud`
  - If `aud` includes `https://auth.myapp.com`, then use RBAC to authorize the access.
    - This means checking whether `roles` is sufficient.
  - If `aud` includes `https://onlinestore.myapp.com`, then use `scope` to authorize the access.
    - This means checking whether `scope` is sufficient.

Now we see clearly that when `RFC6749 section-4.4` is introduced to the system,
Additional handling has to be added to the Resources.
We also see that the subtle difference in `aud` has a great impact on the interpretation of the `access_token`.

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

In my own interpretation, **`scope` means the access of `client_id` to `aud` acting on behalf of `sub`**. The access of `sub` **SHOULD NOT** be included in `scope`.

This is coherent with the practice of Auth0 that they use a separate `permissions` to represent the Permission of the User.

### Discussion: About Resource and Scope

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
To enforce this constraint, it follows naturally that **the `aud` claim never contain include more than 1 Resource**.
It follows naturally that **we have to imply a more strict rule on processing the `resource` parameter in `RFC8707`: `resource` must be specified once and only once.**

With this constraint, we ensure that the following two `access_token` have no ambiguity:

---

```json
{
  "iss": "https://auth.myapp.com",
  "sub": "client_id_inventory",
  "aud": ["https://onlinestore.myapp.com"],
  "client_id": "inventory",
  "scope": "read:orders"
}
```

The above `access_token` can `read:orders` on `https://onlinestore.myapp.com`.

---

```json
{
  "iss": "https://auth.myapp.com",
  "sub": "client_id_inventory",
  "aud": ["https://inventory.myapp.com"],
  "client_id": "inventory",
  "scope": "read:orders"
}
```

The above `access_token` can `read:orders` on `https://inventory.myapp.com`.

---

`access_token` like this

```json
{
  "iss": "https://auth.myapp.com",
  "sub": "client_id_inventory",
  "aud": ["https://onlinestore.myapp.com", "https://inventory.myapp.com"],
  "client_id": "inventory",
  "scope": "read:orders"
}
```

will never be issued.
It is ambiguous that whether `inventory` can `read:orders` on `https://onlinestore.myapp.com` or `https://inventory.myapp.com`, or both.

---

What if a client really wants to access more than 1 Resources?
The client should request separate `access_token` from `https://auth.myapp.com`.

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

### Discussion: (Future work) User / Role / Group and Scope

Auth0 allows Permission (their term for Scope) to be associated with User and Role.
For this to work, the developer **MUST**

- [Enable Role-Based Access Control for APIs](https://auth0.com/docs/get-started/apis/enable-role-based-access-control-for-apis).
- Pass `audience` when they initiate an Authorization Code Grant. This also means that `permissions` normally does not appear at all.

When that setting is enabled, Auth0 always include `permissions` (an array of string) in all `access_token`.

- When it is Authorization Code Grant, the `permissions` is the effective Permissions of `sub` to `aud`.
- When it is Client Credentials Grant, the `permissions` is the effective Permissions of `sub` to `aud`.

Since in Auth0, Permission is local to API, `audience` (i.e. `resource`) can appear once and only once.
This implies it is impossible to generate an `access_token` that can be used to multiple APIs.

As of 2025-06-27, Auth0 offers a Enterprise-only implementation of [RFC8693: OAuth 2.0 Token Exchange](https://datatracker.ietf.org/doc/html/rfc8693),
to allow the developer to [get a `access_token` with another `audience`](https://auth0.com/docs/authenticate/custom-token-exchange#use-case-get-auth0-tokens-for-another-audience).

In essence, Auth0 offers

- The association between Permission and User / Role. APIs can simply rely on `permissions` to determine access. No more checking for `aud` and act differently.
- APIs are still required to handle `sub` conditionally, by checking whether `sub` ends with `@clients` or not.
- Work around the constraint that `audience` is once and only once, by allowing the developer to exchange token with another `audience`.

To adopt a similar functionality, I propose:

- Model Resource and Scope as database tables, so that Role and User can have relationship with Scope.
- Allow relationship between User and Scope (direct association).
- Allow relationship between Role and Scope.
- No relationship between Group and Scope to keep things simple. After all, a Group is just a collection of Roles.

In contrast to Auth0, `resource` **IS NOT** allowed when `grant_type=authorization_code`.
By disallowing `resource` in `grant_type=authorization_code`,
the `aud` in the issued `access_token` is always a singleton array containing `https://auth.myapp.com`.
Thus, `scope` always mean the access of `client_id` to `aud` acting on behalf of `sub`.

With this constraint, the `access_token` of `grant_type=authorization_code` always look like:

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

This `access_token` can be submitted to resources which are programmed to accept `aud=https://auth.myapp.com`, AND read `roles` to determine the access.

For resources that always validate `aud`, a [RFC8693: OAuth 2.0 Token Exchange](https://datatracker.ietf.org/doc/html/rfc8693) is required to obtain an `access_token` with the intended `aud`.

Suppose the `access_token` is submitted to `https://inventory.myapp.com`, which has specially cased to handle `aud=https://auth.myapp.com`. Now `https://inventory.myapp.com` wants to access `https://onlinestore.myapp.com` on behalf of `johndoe`, it needs to perform a Token Exchange.

In this context, the `access_token` with `sub=johndoe` is the `subject_token`.

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
  "scope": "read:orders write:orders list:orders"
}
```

Note that

- The `scope` of `subject_token` is `openid offline_access`.
- The `aud` of `subject_token` is `https://auth.myapp.com`, which is not acceptable by `https://onlinestore.myapp.com`.
- The `scope` of `actor_token` is `read:orders`, so `https://inventory.myapp.com` does not have admin access on its own.
- The `scope` of exchanged token is `read:orders write:orders list:orders`. It means that when `https://inventory.myapp.com` acting on behalf of `johndoe`, who is `onlinestore:admin`, has additional `scope` inherit from `sub` (`johndoe`).

### Discussion: (Future work) Restricting access to Admin API

In Auth0, `RFC6749 section-4.4` is also used to create an `access_token` that can be used to access the Management API.

We can implement the same for our Admin API.
Specifically, we need to introduce an artificial Resource with URI `https://auth.myapp.com/_api/admin`, and define a comprehensive list of scopes that restrict access to the different Resources within the Admin API.

### Discussion: (Future work) Rich Authorization Requests

[RFC9396](https://datatracker.ietf.org/doc/html/rfc9396) introduces `authorization_details`,
which is an enhancement over `scope` and `resource` to specify authorization details in a structured way in form of JSON.

This is supported by Auth0 as an addon. See https://auth0.com/docs/get-started/apis/configure-rich-authorization-requests

## Changes in data models

In order to be forward-compatible with [User / Role / Group and Scope](#discussion-future-work-user-role-group-and-scope), Resource and Scope will be stored in the database, while Client is still stored in project configuration.

Here are the schema of the changes:

```sql
CREATE TABLE _auth_resource (
  id text PRIMARY KEY,
  app_id text NOT NULL,
  created_at timestamp without time zone NOT NULL,
  updated_at timestamp without time zone NOT NULL,
  uri text NOT NULL,
  name text,
  data jsonb
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
  data jsonb
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
```

## Changes in Admin API

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
  addScopeToClientID(input: AddScopeToClientIDInput!): AddScopeToClientIDPayload!
  removeScopeFromClientID(input: RemoveScopeFromClientIDInput!): RemoveScopeFromClientIDPayload!
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
  resource: Resource!
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
  id: ID!
  """The new name"""
  name: String
}

type UpdateResourcePayload {
  resource: Resource!
}

input DeleteResourceInput {
  id: ID!
}

type DeleteResourcePayload {
  ok: Boolean
}

input CreateScopeInput {
  resourceID: ID!
  scope: String!
  description: String
}

type CreateScopePayload {
  scope: Scope!
}

input UpdateScopeInput {
  id: ID!
  """The new description"""
  description: String
}

type UpdateScopePayload {
  scope: Scope!
}

input DeleteScopeInput {
  id: ID!
}

type DeleteScopePayload {
  ok: Boolean
}

input AddResourceToClientIDInput {
  resourceID: ID!
  clientID: String!
}

type AddResourceToClientIDPayload {
  resource: Resource!
}

input RemoveResourceFromClientIDInput {
  resourceID: ID!
  clientID: String!
}

type RemoveResourceFromClientIDPayload {
  resource: Resource!
}

input AddScopeToClientIDInput {
  scopeID: ID!
  clientID: String!
}

type AddScopeToClientIDPayload {
  scope: Scope!
}

input RemoveScopeFromClientIDInput {
  scopeID: ID!
  clientID: String!
}

type RemoveScopeFromClientIDPayload {
  scope: Scope!
}
```

## Changes in OAuth 2.0 implementation

This section describes the protocol-level changes.
You do not need to read this if you are not interested.

### The request

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

### The successful response

```json
{
  "access_token": "A_JWT",
  "token_type": "bearer",
  "expires_in": 3600,
  "scope": "scopea+scopeb+scopec"
}
```

- `access_token`: a JWT access token conforming to `RFC9068`.
- `token_type`: The only value is `bearer` at the moment.
- `expires_in`: The lifetime in seconds of `access_token`. It follows the configuration of the client.
- `scope`: The effective scope. It always appear even the original request does not specify `scope`. When the original request does not specify `scope`, the value is all scopes granted to the client.

### The access token

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

### The error response

- `invalid_grant`: When `grant_type` is invalid. Actually if the grant_type is, for example, `refresh_token`, the entire different flow is run.
- `invalid_client`: When `client_id` is invalid, or `client_secret` is invalid.
- `invalid_resource`: When `resource` is invalid, or it is not granted to `client_id`.
- `invalid_scope`: When `scope` is invalid with respect to the combination of `client_id` and `resource`. That is, the requested scope of `resource` is not granted to `client_id`.

## Changes in SDKs

There is no changes in SDKs.
The changes are in the Token endpoint,
which are supposed to be integrated by the developer using HTTP directly.

## Changes in Documentation

We will have the following documentation changes:

- Document Resource and its Scopes.
  - What is the motivation and the use-cases?
  - What is Resource and what is Scope?
  - How can I create them on the portal?
  - How can I create an `access_token` in one of my backend server?
  - How can I consume the `access_token` in another backend server?

## Prior implementations

This section includes reviews on prior implementations by competitors.

### Auth0

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
- The `permissions` claim is only available if the API has "Enable RBAC" **AND** "Add Permissions in the Access Token".
- In Authorization Code Flow, Permission is `permissions`.
- In Client Credentials Flow, Permission is `scope`.
- When Organization is involved, the globally assigned Roles and Permissions to a User is ignored. `permissions` will always be an empty array. See https://community.auth0.com/t/organization-permissions-claim-empty/99135
- In Client Credentials Flow, `audience` is required instead of `resource`. And `audience` can only appear once.
  - This means the JWT access token is always associated with **ONE** `audience`.
  - Since the JWT access token is always associated with one `audience`, `scope` does not have ambiguity even Permission belonging to different API can share the same name.

An example of the JWT returned by Auth0 when running the Client Credentials Flow

Payload
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

An example of the JWT returned by Auth0 when running the Authorization Code Flow with Organization involved

Payload
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

References:

- https://auth0.com/docs/get-started/authentication-and-authorization-flow/client-credentials-flow/call-your-api-using-the-client-credentials-flow#request-tokens
- https://community.auth0.com/t/opaque-versus-jwt-access-token/31028
- https://community.auth0.com/t/organization-permissions-claim-empty/99135
- https://auth0.com/docs/manage-users/organizations/organizations-for-m2m-applications/configure-your-application-for-m2m-access#define-organization-behavior

### Kinde

Basically it is the same as Auth0. For example

- It copies Auth0 `gty` claim.
- It uses `azp` claim to output the `client_id`.
- It requires the `audience` parameter when `grant_type=client_credentials`.

However,

- It does not include the `sub` claim.

Example header
```json
{
  "alg": "RS256",
  "kid": "59:f7:c0:42:03:a2:f4:07:1f:38:e1:d9:75:d9:c5:43",
  "typ": "JWT"
}
```

Example payload
```json
{
  "aud": [
    "https://myapi.com"
  ],
  "azp": "e4eec6f1389e4118b466fd7d19d7d37e",
  "exp": 1751020780,
  "gty": [
    "client_credentials"
  ],
  "iat": 1750934380,
  "iss": "https://louischanoursky.kinde.com",
  "jti": "4ffa128e-205f-471e-9452-dfb4ed3f57d7",
  "scope": "",
  "scp": [],
  "v": "2"
}
```

References:

- https://docs.kinde.com/developer-tools/kinde-api/access-token-for-api/#method-2-perform-a-post-request-to-get-an-access-token
