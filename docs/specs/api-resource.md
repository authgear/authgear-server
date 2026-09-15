# API Resources and Scopes

API Resources represent protected external services identified by HTTPS URIs. Together with their Scopes, they are the mechanism by which Authgear binds access tokens to specific audiences and controls what permissions a client may request.

Resources are shared across multiple features:

Which mechanism grants a Resource depends on the **grant**, not on the client — a `confidential` client that uses both is subject to both:

- **`client_credentials`** — the client requests tokens bound to a specific Resource through an explicit Client-Resource Association. The `access_policy` described below does **not** govern this grant. See [M2M spec](./m2m.md).
- **`authorization_code` and `refresh_token`** — a client may request a token for a Resource when that Resource's `access_policy` allows the client's *category*. There are four categories and one policy key per category; see [Access Policy](#access-policy), [DCR spec](./dcr.md) and [Third-Party Client spec](./third-party-client.md).

## Table of Contents

- [Glossary](#glossary)
- [Resource URI Requirements](#resource-uri-requirements)
- [Scope Requirements](#scope-requirements)
- [Access Policy](#access-policy)
- [Client-Resource Association](#client-resource-association)
- [Access Token Behavior](#access-token-behavior)
- [Data Model](#data-model)
- [Admin API](#admin-api)

## Glossary

**Resource** — a protected external API or service, uniquely identified within a project by an HTTPS URI.

**Scope** — a permission value defined on a Resource (e.g. `read:orders`). Scopes are local to their Resource; `read:orders` on `https://onlinestore.myapp.com` is a different permission from `read:orders` on `https://inventory.myapp.com`.

**Client-Resource Association** — an explicit link between an OAuth client and a Resource, together with a subset of the Resource's Scopes that the client may request. Required by, and available only to, clients using the `client_credentials` grant.

## Resource URI Requirements

The URI of a Resource must satisfy the following:

- It is a URI as defined in [RFC 3986](https://datatracker.ietf.org/doc/html/rfc3986).
- It must be unique within a Project.
- It must use the `https:` scheme.
- It must not be a domain or subdomain of Authgear's default domains (e.g. `authgearapps.com`, `authgear.cloud`).
- It may have a path component. `https://api.myapp.com` and `https://api.myapp.com/` are both valid but are treated as different Resources.
- It must not have a query component.
- It must not have a fragment component.
- It must not have a userinfo component.

## Scope Requirements

Scopes are defined per-Resource. A scope value must:

- Not be any of the following reserved values: `openid`, `profile`, `email`, `address`, `phone`, `offline_access`, `device_sso`.
- Not start with `https://authgear.com`.
- Conform to the `scope-token` grammar defined in [RFC 6749 §3.3](https://datatracker.ietf.org/doc/html/rfc6749#section-3.3).

## Access Policy

Each Resource and Scope has an `access_policy` JSON object that declares which **categories** of client may access it without a per-client association. All policy keys default to `false`; an absent key is `false`.

### Client categories

A client's category is the pair *(how it was registered, first-party or third-party)*. Every client falls into exactly one category, and each category has exactly one policy key:

| Category | Policy key | Which clients |
|---|---|---|
| Static first-party | `allow_static_first_party_client_access` | `spa`, `traditional_webapp`, `native` and `confidential` clients declared in `authgear.yaml` |
| Static third-party | `allow_static_third_party_client_access` | `third_party_app` clients declared in `authgear.yaml` (a deprecated client type — see [Third-Party Client spec](./third-party-client.md)) |
| Dynamic first-party | `allow_dynamic_first_party_client_access` | Clients DCR-registered with a first-party Initial Access Token, and CIMD-resolved clients whose kind is first-party |
| Dynamic third-party | `allow_dynamic_third_party_client_access` | Clients DCR-registered with a third-party Initial Access Token, and CIMD-resolved third-party clients |

Each key is literal: `allow_dynamic_third_party_client_access` never covers a static third-party client, and `allow_static_first_party_client_access` never covers a dynamic one. The keys are independent — setting one has no effect on the others.

The category is determined by *how the client was registered*, not by *which application type it declares*. A dynamic first-party client presents itself as an ordinary `native` or `spa` application type and is indistinguishable from a static client of the same type by application type alone; it is the registration source that places it in the dynamic bucket.

`m2m` clients are in **no** category. An `m2m` client can only use `client_credentials`, which this policy does not govern (below), so no key here ever applies to one.

### Defaults

Every key defaults to `false`, at both the Resource and the Scope level. A key that is not set means that category is not allowed.

This holds for `allow_static_first_party_client_access` too, even though a static first-party client is declared by a project collaborator. "First-party" is not a synonym for "trusted with every Resource": `spa` and `native` clients are **public**, and their tokens live on end-user devices. A Resource that exists today is reachable only by the clients an admin explicitly associated with it; a `true` default would, on upgrade, let every public client in the project mint a token whose `aud` names that Resource, with no admin action. A resource server that authorizes on `aud` and user identity rather than on scopes would accept those tokens.

The convenience a `true` default reaches for belongs in the portal instead: **the create-Resource form pre-selects "static first-party clients"**, so a Resource created through the portal allows them from the start. Existing Resources are untouched, and an admin who wants a new Resource closed to those clients clears the selection before creating it.

There is no such pre-selection at the Scope level — a new Scope starts with every key `false`. Allowing a category on a Resource grants no resource-specific scope on its own.

### Grants governed by the access policy

`access_policy` governs the `authorization_code` and `refresh_token` grants only. It does **not** govern `client_credentials`, which always requires an explicit Client-Resource Association regardless of every key in this object.

A `confidential` client is the only client type subject to both rules, because it is the only one that is in a category *and* can use `client_credentials`: `allow_static_first_party_client_access: true` lets it send `resource` at `/oauth2/authorize`, and does **not** let it skip the association at `client_credentials`.

> **Rationale:** the `client_credentials` grant has no user and no consent screen — the client *is* the principal. Its least-privilege model depends on an admin naming each client and each scope explicitly, so a blanket category-level grant would silently widen it. See [M2M spec](./m2m.md).

### Two-level check

A client is permitted to request a scope only when **the same key** is `true` on both the Resource and the Scope:

- **Resource level** — when `true` for the client's category, a client of that category may name the Resource URI in the `resource` parameter.
- **Scope level** — when `true` for the client's category, a client of that category may request that scope. When `false` (the default), the scope is unrequestable even if the parent Resource allows the category.

This allows fine-grained control: a Resource may expose `read:orders` to dynamic third-party clients while keeping `delete:orders` restricted, and may expose a further scope to first-party clients only.

A Scope that allows a category its Resource does not is unreachable — the Resource-level check fails first. This is a valid state, not an error; the portal should surface it as a warning rather than reject it.

### Revocation

Both levels are re-read every time an access token is issued, not only when the grant is created. Clearing a key therefore cuts off the affected clients on their next refresh, with no need to hunt down existing grants:

- **Resource level** — the refresh fails with `invalid_target`. Authgear does not fall back to an unbound token, which would hand the client an audience it never authorized; the grant stays bound to that Resource, so the client must re-authorize.
- **Scope level** — the scope is dropped from the issued token, which is still issued with whatever remains.

Deleting a Resource or Scope has the same effect as clearing its keys. OIDC scopes are never affected. Already-issued access tokens are self-contained JWTs that Authgear is not consulted on, so revocation is bounded by the access token lifetime rather than immediate.

### Relationship to Client-Resource Association

There are two ways a client can be granted a Resource, and **the grant decides which one applies** — they are not alternatives evaluated for the same request:

| Grant | Mechanism | The other mechanism |
|---|---|---|
| `client_credentials` | An explicit [Client-Resource Association](#client-resource-association) | `access_policy` is not consulted |
| `authorization_code`, `refresh_token` | `access_policy`, at both the Resource and Scope level | An association is not consulted |

So neither mechanism can override the other: a request only ever goes through one of them.

`confidential` is the only client type that can hold an association *and* use `authorization_code`/`refresh_token`. Even for it the split holds per request: its association is what `client_credentials` reads, and `access_policy` is what `/oauth2/authorize` reads. An association never widens what it can do at the authorization endpoint, and `allow_static_first_party_client_access` never widens what it can do at `client_credentials`.

There is deliberately **no** key meaning "only explicitly selected clients may access this". For `authorization_code`/`refresh_token` that is what an `access_policy` with all keys `false` already means, and a key named `allow_*` that restricted rather than granted would invert the meaning of every other key in the object.

> **Rationale:** dynamic clients (DCR/CIMD) are not created by project collaborators and cannot be enumerated by an admin in advance, so a category-level declaration is the only mechanism available for them. Static clients *can* be enumerated, so a per-client association would be more precise — but associations do not reach these grants today, which is why the category keys cover static clients too.

The partition is a consequence of associations not yet covering user-delegated grants — see [Future work: associations for user-delegated grants](#future-work-associations-for-user-delegated-grants).

> **Extensibility:** `access_policy` is a JSON object rather than a set of boolean columns so that new policy dimensions can be added without a schema migration. New keys default to `false` when absent, preserving the behavior of existing records.

## Client-Resource Association

Clients using `client_credentials` require explicit associations:

1. The admin associates a client with a Resource in the portal.
2. The admin grants specific Scopes from that Resource to the client.
3. The client may then request tokens using `resource=<uri>` and (optionally) `scope=<scopes>`.

If a client requests a Resource it is not associated with, the server returns `invalid_target`. If a client requests a Scope not in its grant, the server returns `invalid_scope`. If no `scope` is specified, all scopes in the client's association are granted.

An association is consulted only by `client_credentials`. The `authorization_code` and `refresh_token` grants are governed by [`access_policy`](#access-policy) alone — see [Relationship to Client-Resource Association](#relationship-to-client-resource-association).

### Which clients may be associated

Only a client whose application type permits the `client_credentials` grant may be associated with a Resource. The Admin API rejects an attempt to associate any other client, separately from the error it already returns for a client ID that does not exist.

An association on a client that cannot use `client_credentials` has no effect at all — the `authorization_code` and `refresh_token` grants are governed solely by `access_policy` and never consult one. Such an association would show in the portal and the Admin API as a grant that is silently doing nothing.

Associations of this shape that already exist are not removed; only new ones are rejected.

### Future work: associations for user-delegated grants

An association today expresses only *"this client may access this Resource as itself"*, which is why it is confined to `client_credentials`, where the client is its own principal. Extending associations to the grants where a client acts **on behalf of a user** — `authorization_code` and `refresh_token` — is planned but not built. It would change three things:

- A Resource could be opened to a category *and* to named individual clients, so the two mechanisms would genuinely OR for one request instead of being [partitioned by grant](#relationship-to-client-resource-association).
- [Which clients may be associated](#which-clients-may-be-associated) would relax to admit those clients.
- A "selected clients only" policy key would become meaningful. Until then there is nothing for such a key to select, which is why none exists.

An association would say which Resource a client may *request* while acting for a user; it would not grant the user permissions they do not already have.

## Access Token Behavior

When a client requests a token with `resource=<uri>`, the issued access token has:

- `aud = [<resource_uri>]` — the audience is set to the requested resource URI only. The Authgear project endpoint is **not** included.
- `scope` — includes both resource-specific scopes and any OIDC scopes (e.g. `openid`, `profile`, `email`) that were requested and granted.

The userinfo endpoint accepts tokens where `scope` contains OIDC scopes (e.g. `openid`, `profile`, `email`), regardless of whether the Authgear project endpoint is present in `aud`. This allows clients that specify `resource` to still call userinfo if the token contains the appropriate OIDC scopes.

When multiple `resource` values are requested (first-party clients only; **not yet implemented**, see [Access Token Audience Binding — Implementation Status](./access-token-audience-binding.md#implementation-status)), `aud` includes all requested resource URIs and a `scope_by_aud` claim maps which scopes apply to which audience. When scopes are ambiguous across resources, the token is downscoped to the intersection. See [M2M spec](./m2m.md) for details on downscoping.

See [Access Token Audience Binding](./access-token-audience-binding.md) for the full specification of audience behavior.

## Data Model

```sql
CREATE TABLE _auth_resource (
  id text PRIMARY KEY,
  app_id text NOT NULL,
  created_at timestamp without time zone NOT NULL,
  updated_at timestamp without time zone NOT NULL,
  uri text NOT NULL,
  name text,
  metadata jsonb,
  -- Access policy JSON object. Missing keys default to false.
  -- One boolean key per client category:
  --   allow_static_first_party_client_access
  --   allow_static_third_party_client_access
  --   allow_dynamic_first_party_client_access
  --   allow_dynamic_third_party_client_access
  access_policy jsonb NOT NULL DEFAULT '{}'
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
  metadata jsonb,
  -- Access policy JSON object. Missing keys default to false.
  -- Same four keys as _auth_resource; allowing a category on a Resource
  -- never grants its scopes.
  access_policy jsonb NOT NULL DEFAULT '{}'
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
  resource_id text NOT NULL REFERENCES _auth_resource(id)
);
-- Each Client can only be associated with a Resource once.
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
-- Each Client can only be associated with a Resource Scope once.
CREATE UNIQUE INDEX _auth_client_resource_scope_unique ON _auth_client_resource_scope USING btree (app_id, client_id, resource_id, scope_id);
```

## Admin API

The following GraphQL schema changes support managing Resources and Scopes via the Admin API.

Resource and Scope CRUD operations do **not** generate events.

```graphql
type Query {
  """If clientID is null, then all resources are returned in a paginated fashion."""
  """If clientID is specified, then all resources associated with the clientID are returned in a paginated fashion."""
  """If searchKeyword is non-null, a prefix search of resourceURI or name is performed."""
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

"""
Access policy for a Resource or Scope. Controls which categories of client
may access it without a per-client association, for the authorization_code
and refresh_token grants only — client_credentials always requires an
explicit association. All fields default to false when absent from the
underlying JSON storage.
"""
type AccessPolicy {
  """
  When true, any client declared in authgear.yaml with a first-party
  application type (spa, traditional_webapp, native, confidential) may access
  this Resource or Scope without a per-client association. Does not cover m2m
  clients, and has no effect on the client_credentials grant.
  """
  allowStaticFirstPartyClientAccess: Boolean!

  """
  When true, any third_party_app client declared in authgear.yaml may access
  this Resource or Scope without a per-client association.
  """
  allowStaticThirdPartyClientAccess: Boolean!

  """
  When true, any dynamically registered first-party client (DCR-registered
  with a first-party Initial Access Token, or CIMD-resolved as first-party)
  may access this Resource or Scope without a per-client association.
  """
  allowDynamicFirstPartyClientAccess: Boolean!

  """
  When true, any dynamically registered third-party client (DCR/CIMD) may
  access this Resource or Scope without a per-client association. Static
  third-party clients are never covered by this flag.
  """
  allowDynamicThirdPartyClientAccess: Boolean!
}

"""
Input for setting an access policy. Fields are merged into the stored
policy: a field present sets that key, a field omitted leaves it unchanged.
Merging, not replacing, so a caller that knows only some of the keys cannot
silently clear the rest. Omitting accessPolicy itself leaves the whole
policy unchanged on update, and defaults every key to false on create.
"""
input AccessPolicyInput {
  """Default false on create; unchanged on update."""
  allowStaticFirstPartyClientAccess: Boolean
  """Default false on create; unchanged on update."""
  allowStaticThirdPartyClientAccess: Boolean
  """Default false on create; unchanged on update."""
  allowDynamicFirstPartyClientAccess: Boolean
  """Default false on create; unchanged on update."""
  allowDynamicThirdPartyClientAccess: Boolean
}

type Resource implements Entity & Node {
  id: ID!
  createdAt: DateTime!
  updatedAt: DateTime!
  resourceURI: String!
  name: String
  accessPolicy: AccessPolicy!
  """If clientID is null, then all scopes of this Resource is returned."""
  """If clientID is specified, then only scopes that are associated with clientID is returned."""
  """If searchKeyword is non-null, a prefix search of scope is performed."""
  """If both clientID and searchKeyword are specified, they are AND-ed."""
  scopes(clientID: String, searchKeyword: String, after: String, before: String, first: Int, last: Int): ScopeConnection
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
  accessPolicy: AccessPolicy!
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
  resourceURI: String!
  name: String
  """If omitted, all access policy fields default to false."""
  accessPolicy: AccessPolicyInput
}

type CreateResourcePayload {
  resource: Resource!
}

input UpdateResourceInput {
  resourceURI: String!
  """The new name."""
  name: String
  """If omitted, the existing access policy is unchanged."""
  accessPolicy: AccessPolicyInput
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
  """If omitted, all access policy fields default to false."""
  accessPolicy: AccessPolicyInput
}

type CreateScopePayload {
  scope: Scope!
}

input UpdateScopeInput {
  resourceURI: String!
  scope: String!
  """The new description."""
  description: String
  """If omitted, the existing access policy is unchanged."""
  accessPolicy: AccessPolicyInput
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
