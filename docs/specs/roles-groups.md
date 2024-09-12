- [Roles and Groups](#roles-and-groups)
- [Role key and Group key](#role-key-and-group-key)
- [The effective roles of a user](#the-effective-roles-of-a-user)
- [CRUD of Roles and Groups](#crud-of-roles-and-groups)
- [The database schema of Roles and Groups](#the-database-schema-of-roles-and-groups)
- [The omission of the support of roles / groups assignment at Admin API user creation](#the-omission-of-the-support-of-roles--groups-assignment-at-admin-api-user-creation)
- [Changes in JWT access token](#changes-in-jwt-access-token)
- [Changes in User Info](#changes-in-user-info)
- [Changes in the headers of resolver](#changes-in-the-headers-of-resolver)
- [Changes in blocking event hooks](#changes-in-blocking-event-hooks)
- [Changes in user listing](#changes-in-user-listing)
- [Changes in account deletion and account anonymization](#changes-in-account-deletion-and-account-anonymization)

# Roles and Groups

- A role has a key, an optional name, and an optional description.
- A group has a key, an optional name, and an optional description.
- Roles and groups have a M-to-N relationship.
- Roles and users have a M-to-N relationship.
- Groups and users have a M-to-N relationship.

# Role key and Group key

- The key must be nonempty.
- The key must be between 1 and 40 characters long.
- The key must satisfy this regex `^[a-zA-Z_][a-zA-Z0-9:_]*$`. That is, the valid characters of the key are alphanumeric, a colon, or an underscore. The first character must be an alphabet or an underscore.
- The prefix `authgear:` is reserved for future use.

Here are some example of valid keys:

- `reader`
- `app:editor`
- `store_manager`

Here are some example of invalid keys:

- `` (because it is empty.)
- `store-manager` (because it contains `-` which is not an allowed character.)
- `authgear:admin` (because it starts with a reserved prefix `authgear:`.)

# The effective roles of a user

The effective roles of a user is the union of

- The roles added directly to the user.
- The roles of the groups which the user is a member.

For example, suppose there are 3 roles:

- `store_manager`
- `salesperson`

And there are 2 groups:

- `newcomer`
- `manager`

The group `newcomer` has these roles:

- `salesperson`

The group `manager` has these roles:

- `store_manager`
- `salesperson`

User John has the role `salesperson` and is in the group `newcomer`. The effective roles of John is `salesperson`.
User Jane has no roles, and is in the group `manager`. The effective roles of Jane is `store_manager` and `salesperson`.

The order of the effective roles of a user is unspecified.

# CRUD of Roles and Groups

The CRUD of Roles and Groups DO NOT generate events.

The following schema snippets describes the addition to the Admin API GraphQL schema.

```graphql
type Query {
  # Roles can be searched by the prefix of name or key.
  roles(searchKeyword: String, after: String, before: String, first: Int, last: Int): RoleConnection
  # Groups can be searched by the prefix of name or key.
  groups(searchKeyword: String, after: String, before: String, first: Int, last: Int): GroupConnection
}

type User {
  effectiveRoles(after: String, before: String, first: Int, last: Int): RoleConnection
  roles(after: String, before: String, first: Int, last: Int): RoleConnection
  groups(after: String, before: String, first: Int, last: Int): GroupConnection
}

type Mutation {
  createRole(input: CreateRoleInput!): CreateRolePayload!
  updateRole(input: UpdateRoleInput!): UpdateRolePayload!
  deleteRole(input: DeleteRoleInput!): DeleteRolePayload!

  createGroup(input: CreateGroupInput!): CreateGroupPayload!
  updateGroup(input: UpdateGroupInput!): UpdateGroupPayload!
  deleteGroup(input: DeleteGroupInput!): DeleteGroupPayload!

  """
  We have 3 types forming M-to-N relationship with each other.
  The pattern of the mutations is
  - add[N]To[Ms]
  - remove[N]From[Ms]

  So there 2 (two patterns) * 3 (N) * 2 (Ms) = 12 mutations in total.
  """

  """Subject is role"""
  addRoleToUsers(input: AddRoleToUsersInput!): AddRoleToUsersPayload!
  removeRoleFromUsers(input: RemoveRoleFromUsersInput!): RemoveRoleFromUsersPayload!
  addRoleToGroups(input: AddRoleToGroupsInput!): AddRoleToGroupsPayload!
  removeRoleFromGroups(input: RemoveRoleFromGroupsInput!): RemoveRoleFromGroupsPayload!

  """Subject is group"""
  addGroupToUsers(input: AddGroupToUsersInput!): AddGroupToUsersPayload!
  removeGroupFromUsers(input: RemoveGroupFromUsersInput!): RemoveGroupFromUsersPayload!
  addGroupToRoles(input: AddGroupToRolesInput!): AddGroupToRolesPayload!
  removeGroupFromRoles(input: RemoveGroupFromRolesInput!): RemoveGroupFromRolesPayload!

  """Subject is user"""
  addUserToRoles(input: AddUserToRolesInput!): AddUserToRolesPayload!
  removeUserFromRoles(input: RemoveUserFromRolesInput!): RemoveUserFromRolesPayload!
  addUserToGroups(input: AddUserToGroupsInput!): AddUserToGroupsPayload!
  removeUserFromGroups(input: RemoveUserFromGroupsInput!): RemoveUserFromGroupsPayload!
}

type RoleConnection {
  edges: [RoleEdge]
  pageInfo: PageInfo!
  totalCount: Int
}

type RoleEdge {
  cursor: String!
  role: Role
}

type Role implements Entity & Node {
  id: ID!
  createdAt: DateTime!
  updatedAt: DateTime!
  key: String!
  name: String
  description: String
  """The list of users who has this role"""
  users(after: String, before: String, first: Int, last: Int): UserConnection
  """The list of groups which has this role"""
  groups(after: String, before: String, first: Int, last: Int): GroupConnection
}

type GroupConnection {
  edges: [GroupEdge]
  pageInfo: PageInfo!
  totalCount: Int
}

type GroupEdge {
  cursor: String!
  group: Group
}

type Group implements Entity & Node {
  id: ID!
  createdAt: DateTime!
  updatedAt: DateTime!
  key: String!
  name: String
  description: String
  """The list of users who belong to this group"""
  users(after: String, before: String, first: Int, last: Int): UserConnection
  """The list of roles this group has"""
  roles(after: String, before: String, first: Int, last: Int): RoleConnection
}

input CreateRoleInput {
  key: String!
  name: String
  description: String
}

type CreateRolePayload {
  role: Role!
}

input UpdateRoleInput {
  id: ID!
  """The new key"""
  key: String
  """The new name"""
  name: String
  """The new description"""
  description: String
}

type UpdateRolePayload {
  role: Role!
}

input DeleteRoleInput {
  id: ID!
}

type DeleteRolePayload {
  ok: Boolean
}

input CreateGroupInput {
  key: String!
  name: String
  description: String
}

type CreateGroupPayload {
  group: Group!
}

input UpdateGroupInput {
  id: ID!
  """The new key"""
  key: String
  """The new name"""
  name: String
  """The new description"""
  description: String
}

type UpdateGroupPayload {
  group: Group!
}

input DeleteGroupInput {
  id: ID!
}

type DeleteGroupPayload {
  ok: Boolean
}

input AddRoleToUsersInput {
  roleKey: String!
  userIDs: [ID!]!
}

type AddRoleToUsersPayload {
  role: Role!
}

input RemoveRoleFromUsersInput {
  roleKey: String!
  userIDs: [ID!]!
}

type RemoveRoleFromUsersPayload {
  role: Role!
}

input AddRoleToGroupsInput {
  roleKey: String!
  groupKeys: [String!]!
}

type AddRoleToGroupsPayload {
  role: Role!
}

input RemoveRoleFromGroupsInput {
  roleKey: String!
  groupKeys: [String!]!
}

type RemoveRoleFromGroupsPayload {
  role: Role!
}

input AddGroupToUsersInput {
  groupKey: String!
  userIDs: [ID!]!
}

type AddGroupToUsersPayload {
  group: Group!
}

input RemoveGroupFromUsersInput {
  groupKey: String!
  userIDs: [ID!]!
}

type RemoveGroupFromUsersPayload {
  group: Group!
}

input AddGroupToRolesInput {
  groupKey: String!
  roleKeys: [String!]!
}

type AddGroupToRolesPayload {
  group: Group!
}

input RemoveGroupFromRolesInput {
  groupKey: String!
  roleKeys: [String!]!
}

type RemoveGroupFromRolesPayload {
  group: Group!
}

input AddUserToRolesInput {
  userID: ID!
  roleKeys: [String!]!
}

type AddUserToRolesPayload {
  user: User!
}

input RemoveUserFromRolesInput {
  userID: ID!
  roleKeys: [String!]!
}

type RemoveUserFromRolesPayload {
  user: User!
}

input AddUserToGroupsInput {
  userID: ID!
  groupKeys: [String!]!
}

type AddUserToGroupsPayload {
  user: User!
}

input RemoveUserFromGroupsInput {
  userID: ID!
  groupKeys: [String!]!
}

type RemoveUserFromGroupsPayload {
  user: User!
}
```

# The database schema of Roles and Groups

```sql
-- Roles
CREATE TABLE _auth_role (
  id text PRIMARY KEY,
  app_id text NOT NULL,
  created_at timestamp without time zone NOT NULL,
  updated_at timestamp without time zone NOT NULL,
  key text NOT NULL,
  name text,
  description text
);
-- Each project has its own set of roles. The role keys are unique within a project.
CREATE UNIQUE INDEX _auth_role_key_unique ON _auth_role USING btree (app_id, key);
-- This index supports typeahead search for roles within a project.
CREATE INDEX _auth_role_key_typeahead ON _auth_role USING btree (app_id, key text_pattern_ops);
CREATE INDEX _auth_role_name_typeahead ON _auth_role USING btree (app_id, name text_pattern_ops);

-- Groups
CREATE TABLE _auth_group (
  id text PRIMARY KEY,
  app_id text NOT NULL,
  created_at timestamp without time zone NOT NULL,
  updated_at timestamp without time zone NOT NULL,
  key text NOT NULL,
  name text,
  description text
);
-- Each project has its own set of groups. The group keys are unique within a project.
CREATE UNIQUE INDEX _auth_group_key_unique ON _auth_group USING btree (app_id, key);
-- This index supports typeahead search for groups within a project.
CREATE INDEX _auth_group_key_typeahead ON _auth_group USING btree (app_id, key text_pattern_ops);
CREATE INDEX _auth_group_name_typeahead ON _auth_group USING btree (app_id, name text_pattern_ops);

-- Roles and Groups
CREATE TABLE _auth_group_role (
  id text PRIMARY KEY,
  app_id text NOT NULL,
  created_at timestamp without time zone NOT NULL,
  updated_at timestamp without time zone NOT NULL,
  group_id text NOT NULL REFERENCES _auth_group(id),
  role_id text NOT NULL REFERENCES _auth_role(id)
);
-- A role and a group can only be associated at most once.
CREATE UNIQUE INDEX _auth_group_role_unique ON _auth_group_role USING btree (app_id, group_id, role_id);
-- This index supports joining from Group.
CREATE INDEX _auth_group_role_group ON _auth_group_role USING btree (app_id, group_id);
-- This index supports joining from Role.
CREATE INDEX _auth_group_role_role ON _auth_group_role USING btree (app_id, role_id);

-- Roles and Users
CREATE TABLE _auth_user_role (
  id text PRIMARY KEY,
  app_id text NOT NULL,
  created_at timestamp without time zone NOT NULL,
  updated_at timestamp without time zone NOT NULL,
  user_id text NOT NULL REFERENCES _auth_user(id),
  role_id text NOT NULL REFERENCES _auth_role(id)
);
-- A role and a user can only be associated at most once.
CREATE UNIQUE INDEX _auth_user_role_unique ON _auth_user_role USING btree (app_id, user_id, role_id);
-- This index supports joining from User.
CREATE INDEX _auth_user_role_user ON _auth_user_role USING btree (app_id, user_id);
-- This index supports joining from Role.
CREATE INDEX _auth_user_role_role ON _auth_user_role USING btree (app_id, role_id);

-- Groups and Users
CREATE TABLE _auth_user_group (
  id text PRIMARY KEY,
  app_id text NOT NULL,
  created_at timestamp without time zone NOT NULL,
  updated_at timestamp without time zone NOT NULL,
  user_id text NOT NULL REFERENCES _auth_user(id),
  group_id text NOT NULL REFERENCES _auth_group(id)
);
-- A group and a user can only be associated at most once.
CREATE UNIQUE INDEX _auth_user_group_unique ON _auth_user_group USING btree(app_id, user_id, group_id);
-- This index supports joining from User.
CREATE INDEX _auth_user_group_user ON _auth_user_group USING btree (app_id, user_id);
-- This index supports joining from Role.
CREATE INDEX _auth_user_group_group ON _auth_user_group USING btree (app_id, group_id);
```

# The omission of the support of roles / groups assignment at Admin API user creation

Due to the fact that Admin API user creation is currently implemented by interaction,
we cannot add the support of roles / groups now.
We need to switch the implementation to direct creation of users, identities and authenticators,
and then we can add support for roles / groups assignment.

# Changes in JWT access token

The effective role keys of a user will appear as `https://authgear.com/claims/user/roles` in the JWT access token.
It is an array of strings, for example, `["store_manager", "salesperson"]`.
The order is unspecified.

# Changes in User Info

The effective role keys of a user will appear as `https://authgear.com/claims/user/roles` in the User Info.
It is an array of strings, for example, `["store_manager", "salesperson"]`.
The order is unspecified.

# Changes in the headers of resolver

See [./api-resolver.md#x-authgear-user-roles](./api-resolver.md#x-authgear-user-roles).

# Changes in blocking event hooks

Some blocking event hooks can mutate the `roles` and `groups` of a user.
See [./hook.md#blocking-event-mutations](./hook.md#blocking-event-mutations).

# Changes in user listing

In user listing, it is possible to filter users by group keys.

```graphql
type Query {
  # Other arguments are omitted for brevity.
  users(groupKeys: [String!]): UserConnection
}
```

> Filtering users by role keys is not implemented because role keys should be computed from the developer's point of view.
> If we implemented a filtering by non-effective role keys, the behavior is confusing.

Group keys are NOT indexed in Elasticsearch. So any changes on group and role assignments DO NOT trigger reindexing.
Searching with keyword is mutually exclusive with filtering with group keys.

> (Future works)
> When reindexing is done with queue, we can index the group keys in Elasticsearch.
> So searching with keyword will work with filtering with group keys.

# Changes in account deletion and account anonymization

The roles and the groups of the users are removed at account deletion and account anonymization.
