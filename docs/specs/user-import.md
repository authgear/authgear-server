- [User Import](#user-import)
  * [About User Import](#about-user-import)
  * [The usage limit](#the-usage-limit)
  * [POST /_api/admin/users/import](#post-_apiadminusersimport)
  * [GET /_api/admin/users/import/ID](#get-_apiadminusersimportid)
  * [The input format](#the-input-format)
    + [Update behavior](#update-behavior)
    + [Update behavior of each field](#update-behavior-of-each-field)
  * [Supported password format](#supported-password-format)
    + [Bcrypt password](#bcrypt-password)
  * [The response](#the-response)
  * [Known issues](#known-issues)
  * [Use cases](#use-cases)
    + [One-off import](#one-off-import)
    + [Correct a previous one-off import](#correct-a-previous-one-off-import)
    + [Syncing with my HR system](#syncing-with-my-hr-system)

# User Import

User Import allows the developer to bulk import users from an existing system to Authgear.

## About User Import

- It is not a synchronous operation. The import is created and runs in the background. The developer can query the status of it.
- It DOES NOT fire existing hooks, namely `user.pre_create` and `user.created`.
- There is a usage limit of user import. See [the usage limit](#the-usage-limit).

## The usage limit

The usage limit is specified in `authgear.features.yaml`.
The following example shows the default usage limit.

```yaml
admin_api:
  user_import_usage:
    enabled: true
    period: day
    quota: 10000
```

## POST /_api/admin/users/import

- The endpoint requires Admin API JWT token to access.
- This endpoint is added to Admin API server.
- It is not part of the GraphQL API.
- The request body has a limit of 500KB.

Use this endpoint to initiate an import.

## GET /_api/admin/users/import/ID

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
  "records": [
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
        "type": "bcrypt",
        "password_hash": "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
      },

      "mfa": {
        "email": "louischan@oursky.com",
        "phone_number": "+85251388325",
        "password": {
          "type": "bcrypt",
          "password_hash": "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
        },
        "totp": {
          "secret": "secret"
        }
      }
    }
  ]
}
```

- `upsert` is an optional boolean. It is false by default. If it is true, then the user is updated. The [update behavior](#update-behavior) of each attribute will be explained below. If it is false, then the record is skipped when it exists already.
- `identifier` is **required**. Valid values are `preferred_username`, `email`, and `phone_number`. It tells Authgear which attribute to use in the input to identify an existing user.

### Update behavior

- **UPDATED_IF_PRESENT_AND_REMOVED_IF_NULL**: If the field is present and the value is not null, then it is updated. If the field is present and the value is null, then it is removed. If the field is absent, no operation is done.
- **UPDATED_IF_PRESENT**: If the field is present, then it is updated. If the field is absent, no operation is done.
- **IGNORED**: If the user exists already, then the field is ignored. If the field is absent, no operation is done.

### Update behavior of each field

|Item|Update Behavior|Description|
|---|---|---|
|`preferred_username`, `email`, `phone_number`|**UPDATED_IF_PRESENT_AND_REMOVED_IF_NULL**|If it is not `identifier`, then the update behavior applies. The corresponding Login ID will be created, updated or removed as needed.|
|`email_verified`, `phone_number_verified`|**UPDATED_IF_PRESENT**|For example, in the first import, if `email_verified` is absent, then email is marked as unverified.|
|All other standard attributes|**UPDATED_IF_PRESENT_AND_REMOVED_IF_NULL**|In particular, `address` IS NOT merged with the existing value, but REPLACES the existing `address` value.|
|`custom_attributes.*`|**UPDATED_IF_PRESENT_AND_REMOVED_IF_NULL**|For each attribute in `custom_attributes`, the update behavior applies individually. So an absent custom attribute in an upsert does not change the existing value.|
|`roles`, `groups`|**UPDATED_IF_PRESENT**|If present, the roles and groups of the user will match the value. For example, supposed the user originally has `["role_a", "role_b"]`. `roles` is `["role_a", "role_c"]`. `role_b` is removed and `role_c` is added.|
|`disabled`|**UPDATED_IF_PRESENT**|Re-importing a record without specifying `disabled` WILL NOT accidentally alter the disabled state previously set by other means.|
|`password`|**IGNORED**|If it was not provided when the record was first imported, subsequent import CANNOT add it back.|
|`mfa.email`|**UPDATED_IF_PRESENT_AND_REMOVED_IF_NULL**|If provided, the user can perform 2FA with email OTP.|
|`mfa.phone_number`|**UPDATED_IF_PRESENT_AND_REMOVED_IF_NULL**|If provided, the user can perform 2FA with phone OTP.|
|`mfa.password`|**IGNORED**|If it was not provided when the record was first imported, subsequent import CANNOT add it back.|
|`mfa.totp`|**IGNORED**|If it was not provided when the record was first imported, subsequent import CANNOT add it back.|

## Supported password format

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

- `details[*].index`: The index of the record in the request.
- `details[*].outcome`: The outcome of the import. Valid values are `inserted`, `updated`, `skipped`, and `failed`.
- `details[*].user_id`: The user ID of the record. It is absent when the import would result in insert but the outcome is `failed`.
- `details[*].record`: The redacted record. In particular, `password_hash` and `secret` are redacted.
- `details[*].warnings`: The warnings encountered in the import. Note that records with warnings are still inserted or updated.
- `details[*].errors`: The errors encountered in the import. Note that records with errors are not inserted nor updated.

```
{
  "id": "some_opaque_string",
  "created_at": "2024-01-01T00:00:00.000Z"
  "status": "completed",
  "summary": {
    "total": 100,
    "inserted": 50,
    "updated": 49,
    "skipped": 0,
    "failed": 1
  },
  "details": [
    {
      "index": 0,
      "outcome": "updated",
      "user_id": "some_opaque_id",
      "record": {
        "preferred_username": "johndoe",
        "email": "johndoe@example.com"
        "password": {
          "type": "bcrypt",
          "password_hash": "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
        }
      },
      "warnings": [
        {
          "message": "password is ignored because the user exists already."
        }
      ]
    },
    {
      "index": 1,
      "outcome": "inserted",
      "user_id": "some_opaque_id",
      "record": {
        "email": "janedoe@example.com"
      }
    },
    {
      "index": 2,
      "outcome": "failed",
      "record": {
        "preferred_username": "louischan",
        "email": "louischan@oursky.com"
      },
      "errors": [
        {
          "reason": "DuplicatedIdentity",
          "message": "identity already exists"
        }
      ]
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
      "email_verified": true,
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
      "email_verified": true,
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
