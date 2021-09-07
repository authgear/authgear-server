# User profile

User profile consists of standard attributes and custom attributes.

## Standard attributes

Standard attributes is a subset of [OIDC standard claims](https://openid.net/specs/openid-connect-core-1_0.html#StandardClaims).
All standard attributes are optional.
All standard attributes are readable and writable via the Admin API.
All standard attributes are readable and writable in the portal.

All standard attributes are by default readable and writable by the end-user in the settings page.
The developer can restrict access to individual standard attribute.

```yaml
ui:
  standard_attributes:
  - pointer: /given_name
    access_control: hidden
  - pointer: /family_name
    access_control: hidden
```

### email

The primary email address of the end-user.
The value always comes from one of the identity the end-user has.
Therefore, the input control is a dropdown.
It is subject to [automatic population](#automatic-population).

### email\_verified

Whether the primary email address is verified.
This attribute is present when `email` is present.
It is NOT writable.

### phone\_number

The primary phone number of the end-user.
The value always from from one of the identity the end-user has.
Therefore, the input control is a dropdown.
It is subject to [automatic population](#automatic-population).

### phone\_number\_verified

Whether the primary phone number is verified.
This attribute is present when `phone_number` is present.
It is NOT writable.

### preferred\_username

The primary username of the end-user.
The value always comes from one of the identity the end-user has.
Therefore, the input control is a dropdown.
It is subject to [automatic population](#automatic-population).

### zoneinfo

The preferred tz database name, e.g. `Asia/Hong_Kong`.
The list of tz database names are constants, so the input control is a dropdown.

### locale

The preferred BCP47 tag, e.g. `zh-HK`.
The input control is a dropdown, and the options are the supported languages of the project.

### birthdate

The supported format is `YYYY-MM-DD`, representing the birthdate, e.g. 1 Jan 1992.

> The OIDC spec also allows `0000-MM-DD` for birthday and `YYYY` for year of birth.
> We only support `YYYY-MM-DD` for simplicity.

### given\_name

The given name of the first name of the end-user.
The input control is a simple text input without validation.
It is subject to [automatic population](#automatic-population).

### family\_name

The family name or the last name of the end-user.
The input control is a simple text input without validation.
It is subject to [automatic population](#automatic-population).

## Automatic population

> Automatic population CANNOT be turned off!

Some of the standard attributes are subject to automatic population.

The automatic population runs when one of the following situation happens:

- When the end-user signs up.
- When the end-user adds an identity.
- When the end-user updates an identity.
- When the end-user removes an identity.

The steps of the automatic population are:

1. Generate a list of email address candidates from the identities, candidates from newer identities are ordered first.
1. Generate a list of phone number candidates from the identities, candidates from newer identities are ordered first.
1. Generate a list of username candidates from the identities, candidates from newer identities are ordered first.
1. Generate a list of given name and family name candidates from the identities, candidates from newer identities are ordered first.
1. If the `email` standard attribute does not refer to a candidate in the list, clear it.
1. If the `phone_number` standard attribute does not refer to a candidate in the list, clear it.
1. If the `preferred_username` standard attribute does not refer to a candidate in the list, clear it.
1. If the `email` standard attribute is absent and the list of candidate is non-empty, set it to the first candidate in the list.
1. If the `phone_number` standard attribute is absent and the list of candidate is non-empty, set it to the first candidate in the list.
1. If the `preferred_username` standard attribute is absent and the list of candidate is non-empty, set it to the first candidate in the list.
1. If both the `given_name` and the `family_name` standard attribute are absent and the list of candidate is non-empty, set it to the first candidate in the list.

## Custom attributes

In addition to standard attributes, the developer can define their own custom attributes.

### Defining custom attributes

The custom attributes are defined as a single JSON schema written against a subset of JSON schema 2019-09.

Here is an example of the schema of the custom attributes.

```JSON
{
  "properties": {
    "app_user_role": {
      "type": "string",
      "enum": ["owner", "editor", "viewer"]
    },
    "stripe_customer_id": {
      "type": "string"
    }
  }
}
```

All changes made to custom attributes must be valid against the schema.

#### Supported subset of JSON schema 2019-09

- `type`: `boolean`, `string`, `number`, `integer`
- `enum`
- `multipleOf`
- `maximum`
- `exclusiveMaximum`
- `minimum`
- `exclusiveMinimum`
- `maxLength`
- `minLength`
- `properties`

### Custom attributes and the resolver

The resolver originally can tell whether the request is authenticated.
If the developer has defined a custom attribute to store the role of the user,
the developer will want to know the role of the user as well.
Then the backend server can do authentication and authorization by forwarding a subrequest to the resolver, without the overhead of calling the Admin API.

The resolver includes the custom attributes of the end-user as
a base64URL encoded JSON under the header `x-authgear-user-custom-attributes`.

### Access control on custom attributes

#### Access control on custom attributes for Admin API and the portal

The Admin API and the portal always have full access to custom attributes.

#### Access control on custom attributes for the resolver

The resolver first checks if the session has client ID.
If a client ID is present, then the access control of the client application is applied.
If a client ID is absent, then the full custom attributes are included in the response header.

#### Access control on custom attributes for client application

Client application can access the custom attributes via the User Info endpoint.
Client application does NOT have write access to custom attributes because the User Info endpoint is not a mutation.
Client application by default has read access to all custom attributes via the User Info endpoint.
The developer can restrict access by declaring which custom attribute the client application has access to.

```yaml
oauth:
  clients:
  - client_id: my_client
    allowed_custom_attributes:
    # Other custom attributes not listed here are not visible to this client.
    # Each entry is a JSON pointer in JSON string representation.
    - /app_user_role
```

#### Access control on custom attributes for end-user

The end-user can interact with custom attributes via 2 means.

Depending on which client application is being used, the end-user can see different set of custom attributes.
Thus the access control for client application also affects the end-user.

Custom attributes can also be configured to be read-only or read-write in the settings page.
Undeclared custom attributes are hidden in the settings page.

```yaml
ui:
  custom_attributes:
  - pointer: /app_user_role
    access_control: readonly
  - pointer: /hobby
    access_control: readwrite
```

### Editing custom attributes in the settings page and in the portal

- Custom attributes without `type` are ignored.
- Custom attributes of `type` `boolean` is a `<input type="checkbox">`
- Custom attributes of `type` `string` is a `<input type="text">`
- Custom attributes of `type` `number` is a `<input type="text">`
- Custom attributes of `type` `integer` is a `<input type="text">`

## ID Token

The ID token never contain standard attributes nor custom attributes.
This is because ID token can be used as `id_token_hint` and can appear in URL query.

## User Info endpoint

The standard attributes appear in the root of the user info response.
The custom attributes appear under the key `custom_attributes`.

Here is an example of the response.

```json
{
  "sub": "user_id",
  "email": "user@example.com",
  "email_verified": true,
  "given_name": "John",
  "family_name": "Doe",
  "custom_attributes": {
    "hobby": "reading"
  }
}
```

## Synchronization of user profile between Authgear and the backend server

There is NO blocking nor non-blocking webhook when standard attributes and custom attributes
are changed by the end-user.

If the developer needs to share some of the attributes,
they should store the attributes themselves and only synchronize some essential attributes back to Authgear.

Custom attributes are opaque to Authgear so the developer never need to synchronize custom attributes back to Authgear.

Essential standard attributes are `zoneinfo` and `locale`.
When the developer is aware of localization, then they will want
to display a consistent UI in the SAME language to the end-user.
In this case, the developer should store `zoneinfo` and `locale` in their backend,
and when `zoneinfo` and `locale` changes, update them with the Admin API.

## Use case examples

### Using Authgear just for authentication.

The developer does not want Authgear to manage the user profile for them.
The developer can just ignore custom attributes.
The developer have to manually opt-out standard attributes by hiding them.

```yaml
ui:
  standard_attributes:
  - pointer: /email
    access_control: hidden
  - pointer: /phone_number
    access_control: hidden
  - pointer: /preferred_username
    access_control: hidden
  - pointer: /given_name
    access_control: hidden
  - pointer: /family_name
    access_control: hidden
  - pointer: /zoneinfo
    access_control: hidden
  - pointer: /locale
    access_control: hidden
  - pointer: /birthdate
    access_control: hidden
```

The developer can still update `zoneinfo` and `locale` via the Admin API
so that Authgear can present localized content to the end-user.

### Using Authgear for authentication and authorization

The developer can define custom attributes for authorization.

```JSON
{
  "properties": {
    "app_user_role": {
      "type": "string"
    }
  }
}
```

Later on the developer can set `app_user_role` via the Admin API.
The resolver will include `app_user_role` in the response header.
The backend server then can use `app_user_role` to do authorization.

### Using Authgear for an very simple demo application

The developer can define custom attributes for storing user profile.

```JSON
{
  "properties": {
    "hobby": {
      "type": "string"
    }
  }
}
```

The developer directs the end-user to the settings page to edit
standard attributes, as well as the custom attributes.

The developer calls the User Info endpoint to retrieve the standard attributes,
and the custom attributes.

Finally, the developer can display the attributes in their application.
