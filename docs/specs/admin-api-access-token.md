# Admin API Access Token

- [Overview](#overview)
- [Use Cases](#use-cases)
  - [UC1: Customer support tool integration](#uc1-customer-support-tool-integration)
  - [UC2: HR system user provisioning](#uc2-hr-system-user-provisioning)
  - [UC3: Automated fraud response](#uc3-automated-fraud-response)
  - [UC4: Compliance audit log archival](#uc4-compliance-audit-log-archival)
  - [UC5: Data warehouse sync](#uc5-data-warehouse-sync)
- [Built-in Resource](#built-in-resource)
- [Scopes](#scopes)
- [Scope Permissions](#scope-permissions)
- [Obtaining a Token](#obtaining-a-token)
- [Token Validation](#token-validation)
- [Backward Compatibility](#backward-compatibility)
- [Examples](#examples)

## Overview

The Admin API supports two authentication methods:

| Method                                                 | Behaviour                                                         |
| ------------------------------------------------------ | ----------------------------------------------------------------- |
| Keypair JWT (`typ: JWT`) signed by `admin-api.auth`    | Full Admin API access. No scope enforcement. Legacy method.       |
| OAuth access token issued via Client Credentials Grant | Scope-controlled. Only scopes granted to the client are enforced. |

This document describes the OAuth-based scoped access token method. The legacy keypair method is unchanged and remains supported for backward compatibility.

## Use Cases

### UC1: Customer support tool integration

A company integrates their support tool (e.g. Zendesk, Intercom) with Authgear so that support agents can look up accounts and perform account operations directly from the support interface — disabling compromised accounts, terminating active sessions, and removing authenticators when a user loses access to their device.

**Setup:**

1. In the portal, create an M2M client named `support-tool`.
2. Associate the client with the Admin API resource (`https://auth.myapp.com/_api/admin`).
3. Grant the following scopes to the client: `user:read user:write`.
4. Copy the `client_id` and `client_secret` into the support tool's backend configuration.

**Token request (called by the support tool backend at startup or on expiry):**

```
POST /oauth2/token
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials
client_id=support-tool
client_secret=THE_CLIENT_SECRET
resource=https://auth.myapp.com/_api/admin
scope=user:read user:write
```

The returned access token is then attached as a `Bearer` token on all subsequent Admin API calls made by the support tool backend.

---

### UC2: HR system user provisioning

A company uses an HR system (e.g. Workday, BambooHR) as the source of truth for employee accounts. A backend sync service creates Authgear accounts when employees join, updates profiles when details change, disables accounts when employees leave, and manages group membership according to department structure.

**Setup:**

1. In the portal, create an M2M client named `hr-sync`.
2. Associate the client with the Admin API resource (`https://auth.myapp.com/_api/admin`).
3. Grant the following scopes to the client: `user:read user:write group:read group:write`.
4. Copy the `client_id` and `client_secret` into the HR sync service configuration.

**Token request:**

```
POST /oauth2/token
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials
client_id=hr-sync
client_secret=THE_CLIENT_SECRET
resource=https://auth.myapp.com/_api/admin
scope=user:read user:write group:read group:write
```

---

### UC3: Automated fraud response

A risk engine monitors user behaviour and flags suspicious accounts. When the engine records a fraud decision, an automation service reads the decision and immediately disables the flagged account. The service only acts on fraud signals and has no reason to access user profiles or audit logs.

**Setup:**

1. In the portal, create an M2M client named `fraud-responder`.
2. Associate the client with the Admin API resource (`https://auth.myapp.com/_api/admin`).
3. Grant the following scopes to the client: `fraud-protection:read user:write`.
4. Copy the `client_id` and `client_secret` into the fraud response service configuration.

**Token request:**

```
POST /oauth2/token
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials
client_id=fraud-responder
client_secret=THE_CLIENT_SECRET
resource=https://auth.myapp.com/_api/admin
scope=fraud-protection:read user:write
```

---

### UC4: Compliance audit log archival

A company is required to retain audit logs for a fixed period under SOC2 or GDPR obligations. An automated pipeline reads audit log entries and archives them to cold storage (e.g. S3, GCS). The pipeline has no reason to access user data.

**Setup:**

1. In the portal, create an M2M client named `audit-archiver`.
2. Associate the client with the Admin API resource (`https://auth.myapp.com/_api/admin`).
3. Grant the following scope to the client: `audit-log:read`.
4. Copy the `client_id` and `client_secret` into the archival pipeline configuration.

**Token request:**

```
POST /oauth2/token
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials
client_id=audit-archiver
client_secret=THE_CLIENT_SECRET
resource=https://auth.myapp.com/_api/admin
scope=audit-log:read
```

---

### UC5: Role sync from enterprise directory

An enterprise uses Active Directory or Okta as the source of truth for employee permissions. A sync service maps directory group membership to Authgear roles — when an employee is promoted or changes teams, their Authgear role assignments are updated automatically. The service needs to read users to resolve identities, and read and write roles to manage assignments.

**Setup:**

1. In the portal, create an M2M client named `role-sync`.
2. Associate the client with the Admin API resource (`https://auth.myapp.com/_api/admin`).
3. Grant the following scopes to the client: `user:read role:read role:write`.
4. Copy the `client_id` and `client_secret` into the role sync service configuration.

**Token request:**

```
POST /oauth2/token
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials
client_id=role-sync
client_secret=THE_CLIENT_SECRET
resource=https://auth.myapp.com/_api/admin
scope=user:read role:read role:write
```

---

## Built-in Resource

The Admin API is represented as a built-in resource with the following URI:

```
https://{authgear_endpoint}/_api/admin
```

For example, if the Authgear endpoint is `https://auth.myapp.com`, the resource URI is `https://auth.myapp.com/_api/admin`.

This resource is built-in and not stored in the `_auth_resource` table. Its URI is derived from the Authgear public endpoint at runtime.

**Shadowing:** If a user-defined resource is created with a URI that matches the built-in resource URI, the built-in resource takes precedence. The user-defined resource is effectively ignored for that URI.

**Client eligibility:** Only M2M (confidential) clients may be associated with the built-in Admin API resource and granted admin scopes. This is enforced when assigning scopes to a client in the portal or Admin API.

**Visibility:** The built-in Admin API resource is returned by the `resources` query and its scopes are returned by scope queries. It can be assigned to or removed from M2M clients using `resource:write` mutations (`addResourceToClientID`, `removeResourceFromClientID`, `addScopesToClientID`, `removeScopesFromClientID`, `replaceScopesOfClientID`). However, the built-in resource and its scopes cannot be created, updated, or deleted — `createResource`, `updateResource`, `deleteResource`, `createScope`, `updateScope`, and `deleteScope` do not apply to them.

## Scopes

The following scopes are defined on the built-in Admin API resource.

| Scope                   | Grants access to                                                                                                                   |
| ----------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| `user:read`             | Read users, identities, authenticators, sessions, authorizations                                                                   |
| `user:write`            | Mutate users, identities, authenticators, sessions, authorizations; image upload                                                   |
| `role:read`             | Read roles                                                                                                                         |
| `role:write`            | Mutate roles; add/remove role–user and role–group assignments                                                                      |
| `group:read`            | Read groups                                                                                                                        |
| `group:write`           | Mutate groups; add/remove group–user and group–role assignments                                                                    |
| `resource:read`         | Read resources and scopes                                                                                                          |
| `resource:write`        | Mutate resources, scopes, and client assignments                                                                                   |
| `fraud-protection:read` | Read fraud protection overview and decision records                                                                                |
| `audit-log:read`        | Read audit logs                                                                                                                    |
| `user:import`           | Bulk user import                                                                                                                   |
| `user:export`           | Bulk user export                                                                                                                   |
| `pii:read`              | Bypass `pii.masking` — all PII fields returned in cleartext; also grants PII-based search (see [pii-masking.md](./pii-masking.md)) |
| `pii:search`            | Use PII fields as search/filter criteria without bypassing response masking (see [pii-masking.md](./pii-masking.md))               |

## Scope Permissions

This section lists the exact queries, mutations, and HTTP endpoints permitted by each scope.

> **Note on mutation responses:** `user:write` without `user:read` permits calling write mutations but not Query operations. Mutation responses that include a `user` object (e.g. `revokeSession`, `deleteIdentity`) are still returned in full — this is a natural side effect of the mutation, not a query.

### `user:read`

**Queries:**

- `users`
- `getUsersByStandardAttribute`
- `getUserByLoginID`
- `getUserByOAuth`
- `node` / `nodes` (for `User`, `Identity`, `Authenticator`, `Session`, `Authorization` node types)

> **Note on PII-based lookups:** When `pii.masking` is enabled, `getUsersByStandardAttribute`
> (attributeName = `email` / `phone_number` / `preferred_username`), `getUserByLoginID`, and
> `users` with a PII-containing `searchKeyword` additionally require `pii:search` or `pii:read`.
> See [pii-masking.md](./pii-masking.md).

### `user:write`

**HTTP endpoints:**

- `POST /_api/admin/images/upload`

**Mutations:**

- `createUser`
- `updateUser`
- `deleteUser`
- `resetPassword`
- `setPasswordExpired`
- `setMFAGracePeriod`
- `removeMFAGracePeriod`
- `sendResetPasswordMessage`
- `generateOOBOTPCode`
- `setVerifiedStatus`
- `setDisabledStatus`
- `setAccountValidFrom`
- `setAccountValidUntil`
- `setAccountValidPeriod`
- `scheduleAccountDeletion`
- `unscheduleAccountDeletion`
- `scheduleAccountAnonymization`
- `unscheduleAccountAnonymization`
- `anonymizeUser`
- `createIdentity`
- `updateIdentity`
- `deleteIdentity`
- `createAuthenticator`
- `deleteAuthenticator`
- `createSession`
- `revokeSession`
- `revokeAllSessions`
- `deleteAuthorization`

### `role:read`

**Queries:**

- `roles`
- `node` / `nodes` (for `Role` node type)

### `role:write`

**Mutations:**

- `createRole`
- `updateRole`
- `deleteRole`
- `addRoleToUsers`
- `removeRoleFromUsers`
- `addUserToRoles`
- `removeUserFromRoles`
- `addRoleToGroups`
- `removeRoleFromGroups`

### `group:read`

**Queries:**

- `groups`
- `node` / `nodes` (for `Group` node type)

### `group:write`

**Mutations:**

- `createGroup`
- `updateGroup`
- `deleteGroup`
- `addGroupToUsers`
- `removeGroupFromUsers`
- `addUserToGroups`
- `removeUserFromGroups`
- `addGroupToRoles`
- `removeGroupFromRoles`

### `resource:read`

**Queries:**

- `resources`
- `node` / `nodes` (for `Resource`, `Scope` node types)

> **Note:** The built-in Admin API resource and its scopes are included in these results. Its scopes can be assigned to clients via `resource:write` mutations, but the resource and its scopes cannot be created, updated, or deleted. See [Built-in Resource](#built-in-resource).

### `resource:write`

**Mutations:**

- `createResource`
- `updateResource`
- `deleteResource`
- `createScope`
- `updateScope`
- `deleteScope`
- `addResourceToClientID`
- `removeResourceFromClientID`
- `addScopesToClientID`
- `removeScopesFromClientID`
- `replaceScopesOfClientID`

### `fraud-protection:read`

**Queries:**

- `fraudProtectionOverview`
- `fraudProtectionLogs`
- `node` / `nodes` (for `FraudProtectionDecisionRecord` node type)

### `audit-log:read`

**Queries:**

- `auditLogs`
- `node` / `nodes` (for `AuditLog` node type)

> **Note on PII filters:** When `pii.masking` is enabled, using the `emailAddresses` or
> `phoneNumbers` filter arguments on `auditLogs` additionally requires `pii:search` or `pii:read`.
> See [pii-masking.md](./pii-masking.md).

### `user:import`

**HTTP endpoints:**

- `POST /_api/admin/users/import`
- `GET /_api/admin/users/import/{id}`

### `user:export`

**HTTP endpoints:**

- `POST /_api/admin/users/export`
- `GET /_api/admin/users/export/{id}`

## Obtaining a Token

Use the OAuth 2.0 Client Credentials Grant with the built-in Admin API resource URI. See [M2M — The request](./m2m.md#the-request) for the full flow description.

```
POST /oauth2/token HTTP/1.1
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials
client_id=my-backend
client_secret=THE_CLIENT_SECRET
resource=https://auth.myapp.com/_api/admin
scope=user:read pii:read
```

The returned access token is a JWT conforming to [RFC9068](https://datatracker.ietf.org/doc/html/rfc9068):

```json
{
  "iss": "https://auth.myapp.com",
  "sub": "client_id_my-backend",
  "aud": ["https://auth.myapp.com/_api/admin"],
  "client_id": "my-backend",
  "scope": "user:read pii:read",
  "exp": 1234567890
}
```

## Token Validation

The Admin API accepts a Bearer token in the `Authorization` header and determines the access method by inspecting the token:

1. **Keypair JWT** — If the token is signed by the `admin-api.auth` keypair, full access is granted (legacy path, no scope enforcement).
2. **OAuth access token** — If the token is a standard JWT whose `aud` includes the built-in resource URI, scope-controlled access is granted based on the `scope` claim.

## Backward Compatibility

Keypair JWTs signed by `admin-api.auth` retain full Admin API access unchanged. Existing tokens and integrations are unaffected.

## Examples

Read-only integration (no PII):

```
POST /oauth2/token
grant_type=client_credentials&client_id=my-backend&client_secret=...&resource=https://auth.myapp.com/_api/admin&scope=user%3Aread
```

Data migration script (needs raw email addresses):

```
POST /oauth2/token
grant_type=client_credentials&client_id=my-backend&client_secret=...&resource=https://auth.myapp.com/_api/admin&scope=user%3Aread+pii%3Aread
```
