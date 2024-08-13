# 2FA Grace Period

- [Abstract](#abstract)
- [Configuration and the global grace period](#configuration-and-the-global-grace-period)
- [Per-user Grace Period](#per-user-grace-period)
  - [Changes on the Admin GraphQL API](#changes-on-the-admin-graphql-api)
    - [Error Response](#error-response)
  - [Changes on database schema](#changes-on-database-schema)

## Abstract

In order to enforce 2FA, we will provide a grace period for users to enroll in 2FA after they have signed up.

After the grace period, the user will be required to enroll in 2FA before they can sign in.

## Configuration and the global grace period

Global grace period can be enabled for forcing users to enroll in 2FA upon login instead of blocking them.

```yaml
authentication:
  secondary_authentication_mode: "required"
  secondary_authentication_grace_period:
    enabled: true
    end_at: "2021-01-01T00:00:00Z"
```

By default, no grace period is provided. User without 2FA must contact admin to login after [2FA mode](./user-model.md#secondary-authenticator) is set to `required`.

## Per-user Grace Period

Regardless of global grace period, specific users can be granted grace period through Admin Portal / GraphQL API.

### Changes on the Admin GraphQL API

```gql
type User {
  id: ID!
  # ...

  mfaGracePeriodEndAt: DateTime
}

type SetMFAGracePeriodInput {
  userID: ID!
  endAt: DateTime!
}

type SetMFAGracePeriodPayload {
  user: User!
}

type RemoveMFAGracePeriodInput {
  userID: ID!
}

type RemoveMFAGracePeriodPayload {
  user: User!
}

type Mutation {
  setMFAGracePeriod(
    input: SetMFAGracePeriodInput!
  ): SetMFAGracePeriodPayload!
  removeMFAGracePeriod(
    input: RemoveMFAGracePeriodInput!
  ): RemoveMFAGracePeriodPayload!
}
```

Both `setMFAGracePeriod` and `removeMFAGracePeriod` mutations are indempotent, users can have grace period granted/revoked multiple times even with same value.

#### Error Response

|Description|Name|Reason|Info|
|---|---|---|---|
|Invalid Grace Period<br />i.e. in the past or too far in the future (> 90 days)|`Invalid`|`MFAGracePeriodInvalid`|-|

### Changes on database schema

Per-user grace period granted through Admin API is stored in the user model as `mfa_grace_period_end_at`.

```sql
ALTER TABLE _authgear_user ADD COLUMN mfa_grace_period_end_at TIMESTAMP WITHOUT TIME ZONE;
```
