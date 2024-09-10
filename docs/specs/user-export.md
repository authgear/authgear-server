# User Export

User Export allows the developer to bulk export all users from Authgear to a file.

## About User Export

- It is not a synchronous operation. The export is created and runs in the background. The developer can query the status of it.

## Create an export

The endpoint is `POST /_api/admin/users/export`.

- The endpoint requires Admin API JWT token to access.
- This endpoint is added to Admin API server.
- It is not part of the GraphQL API.
- At most 1 running export at any given moment for a project.
- The rate limit of creating a export is 24 per 24h. This means failed requests do not count. This rate limit is hard-coded at the moment. It could be configurable in the future.
- A pending export lasts for 24h before it expires.

### The request body of Create an export

```
{
  "format": "ndjson",
  "csv": {
    "fields": [
      {
        "pointer": "/sub",
        "field_name": "user_id"
      }
    ]
  }
}
```

- `format`: Required. It must be `ndjson` or `csv`.
  - `ndjson`: The output is a ndjson file. See https://github.com/ndjson/ndjson-spec
  - `csv`: The output is a CSV file. See https://datatracker.ietf.org/doc/html/rfc4180
- `csv.fields`: Required when `format` is `csv`. It must be an non-empty array.
  - `csv.fields.pointer`: Required. Select which field in the record to output. It must be a JSON pointer of at least one reference token. Each reference token must be non-empty. See https://datatracker.ietf.org/doc/html/rfc6901 and [The record format](#the-record-format)
  - `csv.fields.field_name`: See [The field name](#the-field-name)

#### The field name

- `field_name` is optional.
- If `field_name` is given, then it is used as is.
- If `field_name` is not given, then it is derived from `pointer` with the following rules.
  - Let `parts` be the list of the reference tokens in `pointer`.
  - Join `parts` with the character `.`.

For example, given `pointer` is `/address/formatted`,

- Then `parts` is `["address", "formatted"]`.
- The join result is `address.formatted`.
- The field name is the join result.

For example, given `pointer` is `/roles/0`,

- Then `parts` is `["roles", "0"]`.
- The join result is `roles.0`.
- The field name is the join result.

Regardless of whether the field names are given or derived, they must be unique.
If the field names are not unique, it is an error when the export is created.
An error is immediately returned in this case, the import is not created.
See [The error response of Create an export](#the-error-response-of-create-an-export)

### The response body of Create an export

See [The response body](#the-response-body).

### The error response of Create an export

> we have a global middleware of handling Admin API authentication,
> that middleware returns 403 without a JSON body when authentication failed.

|Description|Name|Reason|Info|
|---|---|---|---|
|When the rate limit exceeded|`TooManyRequest`|`RateLimited`|{"bucket_name": "UserExport"}|
|When there is a running export|`TooManyRequest`|`MaximumConcurrentJobLimitExceeded`|-|
|When the input fails the validation|`Invalid`|`ValidationFailed`|The info should contain the JSON schema validation output|
|When the field names are not unique|`Invalid`|`UserExportNonUniqueFieldNames`|Output the full list of "field_names". Like `{"field_names": ["sub", "a", "b", "a"]}`|

## Get the status of an export

The endpoint is `GET /_api/admin/users/export/{ID}`

- The endpoint requires Admin API JWT token to access.
- This endpoint is added to Admin API server.
- It is not part of the GraphQL API.
- The result of a completed export lasts for 24h before it expires.
- When the export is completed, the response body includes a freshly signed URL to the export file. The signed URL is valid for 60s.

### The response body of Get the status of an export

See [The response body](#the-response-body).

### The error response of Get the status of an export

> we have a global middleware of handling Admin API authentication,
> that middleware returns 403 without a JSON body when authentication failed.

|Description|Name|Reason|Info|
|---|---|---|---|
|When the given ID does not refer to an export|`NotFound`|`TaskNotFound`|-|

## The response body

The response body of a just created export looks like

```
{
  "id": "some_opaque_string",
  "status": "pending",
  "created_at": "2024-01-01T00:00:00.000Z",
  "request": {
    "format": "ndjson"
  }
}
```

The response body of a completed export looks like

```
{
  "id": "some_opaque_string",
  "status": "completed",
  "created_at": "2024-01-01T00:00:00.000Z",
  "completed_at": "2024-01-01T00:01:00.000Z",
  "request": {
    "format": "ndjson"
  },
  "download_url": "https://some-signed-url?with-a=signature"
}
```

The response body of a failed export looks like

```
{
  "id": "some_opaque_string",
  "status": "completed",
  "created_at": "2024-01-01T00:00:00.000Z",
  "failed_at": "2024-01-01T00:01:00.000Z",
  "request": {
    "format": "ndjson"
  },
  "error": {
    "message": "blahblahblah",
    "reason": "SomeReason"
  }
}
```

- `error`: The API error object we have been using in all other API.

## The export file

### The name of the export file

The name of the export file is `{{ .AppID }}-{{ .TaskID }}-{{ .TaskCompletedAtInISO9601BasicFormat}}.{ndjson|csv}`.

For example, given

- The project is `myapp`.
- The id of the task is `userexport_deadbeef`.
- The completion time of the task is `2024-09-09T10:46:51.275Z`.
- The requested `format` is `ndjson`.

The name of the export file is `myapp-userexport_deadbeef-20240909104651Z.ndjson`.

### The metadata of the export file

- `Content-Disposition`: `attachment; filename=FILENAME`
- `Content-Type`: `application/x-ndjson` for `ndjson`, `text/csv` for `csv`.

Since the export file is accessed with a signed URL of a short validity, setting `Cache-Control` is not really helpful.

### The housekeep of the export file

- GCS: https://cloud.google.com/storage/docs/lifecycle
- S3: https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lifecycle-mgmt.html
- Azure: https://learn.microsoft.com/en-us/azure/storage/blobs/lifecycle-management-overview?tabs=azure-portal

The above cloud storage can be configured to delete objects of a certain age.
So there is no need to housekeep manually.

### The content of the export file

- If `format` is `ndjson`, then the file contains a record (See [The record format](#the-record-format)) per line.
  - Each line is terminated by a `\n` (The newline character).
  - The number of lines correspond to the number of exported records.
  - Exporting in a project without any user will in a file of zero length.

> Implementation note: Please make sure each line is ended by a `\n`!

- If `format` is `csv`, then the file starts with a header, followed by records.
  - The header correspond to the `csv.fields`.
  - The order of the field in the header correspond to the order in `csv.fields`.
  - The name of the field is taken from `csv.fields.field_name` if it is specified, or derived from `csv.fields.pointer` if `field_name` is absent.
  - The CSV follows the format documented in RFC4180. Internally, it is handled with https://pkg.go.dev/encoding/csv.
  - If `pointer` resolves to
    - JSON string, then the string is written directly.
    - JSON number, then the number is written directly.
    - JSON boolean, then `true` or `false` is written directly.
    - JSON null, then an empty string is written.
    - non-existing value, then an empty string is written.
    - JSON array, then the compact JSON serialization is written.
    - JSON object, then the compact JSON serialization is written.

For example, given the request

```
{
  "format": "csv",
  "csv": {
    "fields": [
      {
        "pointer": "/sub"
      },
      {
        "pointer": "/roles"
      },
      {
        "pointer": "/address"
      },
      {
        "pointer": "/address/formatted",
        "field_name": "address_formatted"
      }
    ]
  }
}
```

, and a record

```
{
  "sub": "opaque_user_id",

  "address": {
    "formatted": "1 Unnamed Road, Central, Hong Kong Island, HK",
    "street_address": "1 Unnamed Road",
    "locality": "Central",
    "region": "Hong Kong",
    "postal_code": "N/A",
    "country": "HK"
  },

  "roles": ["role_a", "role_b"]
}
```

The content of the file is

```
sub,roles,address,address_formatted
opaque_user_id,"[""role_a"",""role_b""]","{""formatted"":""1 Unnamed Road, Central, Hong Kong Island, HK"",""street_address"":""1 Unnamed Road"",""locality"":""Central"",""region"":""Hong Kong"",""postal_code"":""N/A"",""country"":""HK""}","1 Unnamed Road, Central, Hong Kong Island, HK"
```

### The record format

Here is an example of the record

```
{
  "sub": "opaque_user_id",

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

  "disabled": false
}
```

- `sub`: The Authgear user ID.
- `preferred_username`: The primary username of the user.
- `email`: The primary email of the user.
  - `email_verified`: Whether the email is verified.
- `phone_number`: The primary phone number of the user.
  - `phone_number_verified`: Whether the phone number is verified.
- `name`, `given_name`, `family_name`, `middle_name`, `nickname`: The OIDC standard attributes about names. See https://openid.net/specs/openid-connect-core-1_0.html#StandardClaims
- `profile`, `picture`, `website`, `gender`, `birthdate`, `zoneinfo`, `locale`, `address`: Other OIDC standard attributes. See https://openid.net/specs/openid-connect-core-1_0.html#StandardClaims
- `custom_attributes`: All custom_attributes of the user.
- `roles`: The role keys of all roles directly assigned to the user. This DOES NOT include roles implied by the groups.
- `groups`: The group keys of all groups assigned to the user.
- `disabled`: Whether the user is disabled.

> TODO: Support exporting `identities`

> Future work: Support exporting the password hash.

## Ceveats

Since we do not export the password hash and MFA information.
The exported JSON record cannot be imported into another Authgear project directly.
