# API Resources and Scopes

API Resources represent protected external services identified by HTTPS URIs. Together with their Scopes, they are the mechanism by which Authgear binds access tokens to specific audiences and controls what permissions a client may request.

Resources are shared across multiple features:

- **M2M** (`client_credentials` grant) — confidential clients request tokens bound to a specific Resource. See [M2M spec](./m2m.md).
- **Third-party clients** (dynamically registered via DCR) — clients can request tokens for Resources marked as `allow_any_client_access`. See [DCR spec](./dcr.md) and [Third-Party Client spec](./third-party-client.md).
- **First-party clients** (all grant types) — first-party clients can request resource-bound tokens if they are explicitly associated with the Resource.

## Table of Contents

- [Glossary](#glossary)
- [Resource URI Requirements](#resource-uri-requirements)
- [Scope Requirements](#scope-requirements)
- [Access Without Per-Client Association](#access-without-per-client-association)
- [Client-Resource Association](#client-resource-association)
- [Access Token Behavior](#access-token-behavior)
- [Data Model](#data-model)
- [Admin API](#admin-api)

## Glossary

**Resource** — a protected external API or service, uniquely identified within a project by an HTTPS URI.

**Scope** — a permission value defined on a Resource (e.g. `read:orders`). Scopes are local to their Resource; `read:orders` on `https://onlinestore.myapp.com` is a different permission from `read:orders` on `https://inventory.myapp.com`.

**Client-Resource Association** — an explicit link between an OAuth client and a Resource, together with a subset of the Resource's Scopes that the client may request. Required for first-party M2M clients.

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

## Access Without Per-Client Association

By default, Resources and Scopes are only accessible to clients with an explicit [Client-Resource Association](#client-resource-association). The `allow_any_client_access` flag removes this requirement, allowing any client to access the Resource or Scope without a per-client association.

The primary use case is dynamically registered third-party clients (DCR clients), which cannot be given explicit per-client associations by design. However, the flag applies to all client types — any client that requests a Resource with this flag set may receive a token for it.

The `allow_any_client_access` flag works as follows:

- **Resource level** — when `true` on a Resource, any client may include that Resource URI in the `resource` parameter of their authorization requests, without requiring an explicit association.
- **Scope level** — when `true` on a Scope, that scope is requestable by any client without explicit association. When `false` (the default), the scope requires explicit client association even if the parent Resource has the flag set.
- Both the Resource and the individual Scope must have `allow_any_client_access: true` for a client to successfully request that scope without association.

This allows fine-grained control: for example, a Resource may expose `read:orders` to any client but keep `delete:orders` restricted to explicitly associated clients.

> **Rationale:** Third-party clients registered via DCR are not created by project collaborators and cannot be individually trusted with per-client associations. The `allow_any_client_access` flag lets admins declare, once per Resource/Scope, which permissions are safe to expose to any client. Any subsequently registered third-party client (or other client) may then access those resources without further admin action per client.

## Client-Resource Association

First-party M2M clients (using `client_credentials`) require explicit associations:

1. The admin associates a client with a Resource in the portal.
2. The admin grants specific Scopes from that Resource to the client.
3. The client may then request tokens using `resource=<uri>` and (optionally) `scope=<scopes>`.

If a client requests a Resource it is not associated with, the server returns `invalid_resource`. If a client requests a Scope not in its grant, the server returns `invalid_scope`. If no `scope` is specified, all scopes in the client's association are granted.

Clients accessing a Resource that has `allow_any_client_access: true` do not require a per-client association.

## Access Token Behavior

When a client requests a token with `resource=<uri>`, the issued access token has:

- `aud = [<resource_uri>]` — the audience is set to the requested resource URI only. The Authgear project endpoint is **not** included.
- `scope` — includes both resource-specific scopes and any OIDC scopes (e.g. `openid`, `profile`, `email`) that were requested and granted.

The userinfo endpoint accepts tokens where `scope` contains OIDC scopes (e.g. `openid`, `profile`, `email`), regardless of whether the Authgear project endpoint is present in `aud`. This allows clients that specify `resource` to still call userinfo if the token contains the appropriate OIDC scopes.

When multiple `resource` values are requested (first-party clients only), `aud` includes all requested resource URIs and a `scope_by_aud` claim maps which scopes apply to which audience. When scopes are ambiguous across resources, the token is downscoped to the intersection. See [M2M spec](./m2m.md) for details on downscoping.

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
  allow_any_client_access boolean NOT NULL DEFAULT false
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
  allow_any_client_access boolean NOT NULL DEFAULT false
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

type Resource implements Entity & Node {
  id: ID!
  createdAt: DateTime!
  updatedAt: DateTime!
  resourceURI: String!
  name: String
  """Whether any client may access this resource without a per-client association."""
  allowAnyClientAccess: Boolean!
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
  """Whether any client may request this scope without a per-client association."""
  allowAnyClientAccess: Boolean!
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
  """Default false."""
  allowAnyClientAccess: Boolean
}

type CreateResourcePayload {
  resource: Resource!
}

input UpdateResourceInput {
  resourceURI: String!
  """The new name."""
  name: String
  allowAnyClientAccess: Boolean
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
  """Default false."""
  allowAnyClientAccess: Boolean
}

type CreateScopePayload {
  scope: Scope!
}

input UpdateScopeInput {
  resourceURI: String!
  scope: String!
  """The new description."""
  description: String
  allowAnyClientAccess: Boolean
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
