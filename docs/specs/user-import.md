# User Import

User Import allows the developer to bulk import users from an existing system to Authgear.

## About User Import

- It is not a synchronous operation. The import is created and runs in the background. The developer can query the status of it.
- It DOES NOT fire existing hooks, namely `user.pre_create` and `user.created`.

## POST /api/users/import

- The endpoint requires Admin API JWT token to access.
- This endpoint is added to Admin API server.
- It is not part of the GraphQL API.
- The request body has a limit of 500KB.

Use this endpoint to initiate an import.

## GET /api/users/import/ID

- The endpoint requires Admin API JWT token to access.
- This endpoint is added to Admin API server.
- It is not part of the GraphQL API.

Use this endpoint to query the status of an import.

## The input format

The input is a JSON document. Here is an example.

```
{
  "upsert": true,
  "identifier": "email",
  "users": [
    {
      "preferred_username": "louischan",
      "email": "louischan@oursky.com",
      "phone_number": "+85298765432",

      "email_verified": true,
      "phone_number_verified": true,

      "name": "Louis Chan",
      "given_name": "Louis",
      "family_name": "Chan",
      "middle_name": "",
      "nickname": "Lou",
      "profile": "https://example.com",
      "picture": "https://example.com",
      "website": "https://example.com",
      "gender": "male",
      "birthdate": "1990-01-01",
      "zoneinfo": "Asia/Hong_Kong",
      "locale": "zh-Hant-HK",
      "address": {
        "formatted": "1 Unnamed Road, Central, Hong Kong Island, HK",
        "street_address": "1 Unnamed Road",
        "locality": "Central",
        "region": "Hong Kong",
        "postal_code": "N/A",
        "country": "HK"
      },

      "custom_attributes": {
        "member_id": "123456789"
      },

      "roles": ["role_a", "role_b"],
      "groups": ["group_a"],

      "disabled": false,

      "password": {
        "type": "plain",
        "plain_password": "secret"
      },

      "mfa": {
        "email": "louischan@oursky.com",
        "phone_number": "+85251388325",
        "password": {
          "type": "plain",
          "plain_password": "secret"
        },
        "totp": {
          "secret": "secret"
        }
      }
    }
  ]
}
```

- `upsert` is an optional boolean. It is false by default. If it is true, then the user is updated. The [update behavior](#update-behavior) of each attribute will be explained below. If it is false, then the user is skipped when it exists already.
- `identifier` is **required**. Valid values are `preferred_username`, `email`, and `phone_number`. It tells Authgear which attribute to use in the input to identify an existing user.
- `preferred_username`, `email`, `phone_number`: If it is not `identifier`, then the update behavior is **UPDATED_IF_PRESENT_AND_REMOVED_IF_NULL**. The corresponding Login ID will be created, updated or removed as needed.
- `email_verified`, `phone_number_verified`: The update behavior is **UPDATED_IF_PRESENT**. For example, in the first import, if `email_verified` is absent, then email is marked as unverified.
- The update behavior of other standard attributes is **UPDATED_IF_PRESENT_AND_REMOVED_IF_NULL**. In particular, `address` IS NOT merged with the existing value, but REPLACES the existing `address` value.
- `custom_attributes.*`: For each attribute in `custom_attributes`, the update behavior is **UPDATED_IF_PRESENT_AND_REMOVED_IF_NULL**.
- `roles` and `groups`: The update behavior is **UPDATED_IF_PRESENT`.
- `disabled`: The update behavior is **UPDATED_IF_PRESENT**. So re-importing a user without specifying `disabled` WILL NOT accidentally alter the disabled state previously set by other means.
- `password`: The update behavior is **IGNORED**. If `password` was not provided when the user was first imported, subsequent import CANNOT add password.
- `mfa.email`: The update behavior is **UPDATED_IF_PRESENT_AND_REMOVED_IF_NULL**. The user can perform 2FA with email OTP.
- `mfa.phone_number`: The update behavior is **UPDATED_IF_PRESENT_AND_REMOVED_IF_NULL**. The user can perform 1FA with phone OTP.
- `mfa.password`: The update behavior is **IGNORED**. If `mfa.password` was not provided when the user was first imported, subsequent import CANNOT add additional password.
- `mfa.totp`: The update behavior is **IGNORED**. If `mfa.totp` was not provided when the user was first imported, subsequent import CANNOT add TOTP.

## Update behavior

- **UPDATED_IF_PRESENT_AND_REMOVED_IF_NULL**: If the field is present and the value is not null, then it is updated. If the field is present and the value is null, then it is removed. If the field is absent, no operation is done.
- **UPDATED_IF_PRESENT**: If the field is present, then it is updated. If the field is absent, no operation is done.
- **IGNORED**: If the user exists already, then the field is ignored. If the field is absent, no operation is done.

## Supported password format

### Plain password

```
{
  "type": "plain",
  "plain_password": "secret"
}
```

### Bcrypt password

```
{
  "type": "bcrypt",
  "password_hash": "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
}
```

## The response

You will receive a response similar to the following when you just initiated an import.

```
{
  "id": "some_opaque_string",
  "created_at": "2024-01-01T00:00:00.000Z"
  "status": "pending",
}
```

You will receive a response similar to the following when the import completed (with or without errors).
Sensitive values are redacted in the response.

```
{
  "id": "some_opaque_string",
  "created_at": "2024-01-01T00:00:00.000Z"
  "status": "completed",
  "summary": {
    "total": 100,
    "inserted": 50,
    "updated": 49,
    "failed": 1
  },
  "errors": [
    {
      "user": {
        "preferred_username": "louischan",
        "email": "louischan@oursky.com"
      },
      "error": {
        "reason": "DuplicatedIdentity",
        "message": "identity already exists"
      }
    }
  ]
}
```

After the import has completed, the information of the import will be deleted automatically 24 hours after the completion.

## Known issues

It is impossible to swap `identifier`. The developer must somehow free one of them first.

## Use cases

### One-off import

I want to import users from my old system into Authgear. This is a one-off operation. I need to prepare an input that looks like

```
{
  "identifier": "email",
  "users": [
    {
      "email": "user@example.com",
      "email_verified": true
      "password": {
        "type": "bcrypt",
        "password_hash": "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
      }
    }
  ]
}
```

### Correct a previous one-off import

Some user data was incorrect in the previous one-off import. I have corrected the input. I need to add `"upsert": true` to the input.

```
{
  "upsert": true,
  "identifier": "email",
  "users": [
    {
      "email": "user@example.com",
      "email_verified": true
      "name": "John Doe",
      "password": {
        "type": "bcrypt",
        "password_hash": "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
      }
    }
  ]
}
```

### Syncing with my HR system

I have an existing HR system that defines the set of accounts that can sign in. On every night, I will
- Query my HR system to see what accounts were deleted.
- Use bulk delete users API of Authgear to delete those users, freeing the email addresses and the phone numbers.
- Prepare the input of User Import.
  - Set `upsert` to `true` because my HR system is the source of truth.
  - Set `identifier` to `preferred_username`. `preferred_username` is the unique ID in my HR system.
