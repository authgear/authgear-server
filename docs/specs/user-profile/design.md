# User profile

User profile consists of standard attributes and custom attributes.

## Standard attributes

Standard attributes are [OIDC standard claims](https://openid.net/specs/openid-connect-core-1_0.html#StandardClaims).
All standard attributes are optional.

## Access control on standard attributes

The Admin API and the portal always have full access to all standard attributes.

```yaml
user_profile:
  standard_attributes:
    access_control:
    - pointer: /given_name
      access_control: hidden
    - pointer: /zoneinfo
      access_control: readonly
```

Possible values of `access_control` are:

- `hidden`: The standard attribute is hidden from the end-user, the User Info endpoint and the resolver.
- `internal`: The standard attribute is hidden in the settings page. But it is visible to the User Info endpoint and the resolver.
- `readonly`: The standard attribute is visible to the end-user, the User Info endpoint and the resolver. The end-user can view but cannot edit it in the settings page.
- `readwrite`: The standard attribute is visible to the end-user, the User Info endpoint and the resolver. The end-user can view and edit it in the settings page.

The default value of `access_control` of all standard attributes is `readwrite`.

## Standard attributes population

Most of the standard attributes are subject to population.

```yaml
user_profile:
  standard_attributes:
    population:
      strategy: on_signup
```

Possible values of `strategy` are

- `none`: No population.
- `on_signup`: Populate from the identity being used in sign up.

## List of standard attributes subject to population

- `name`
- `given_name`
- `family_name`
- `middle_name`
- `nickname`
- `profile`
- `picture`
- `website`
- `gender`
- `birthdate`
- `zoneinfo`
- `locale`
- `address.formatted`
- `address.street_address`
- `address.locality`
- `address.region`
- `address.postal_code`
- `address.country`

### Single-line string standard attributes

- `name`
- `given_name`
- `family_name`
- `middle_name`
- `nickname`
- `address.locality`
- `address.region`
- `address.postal_code`

### Multi-line string standard attributes

- `address.formatted`
- `address.street_address`

### URL standard attributes

- `profile`
- `picture`
- `website`

### gender

The predefined values are `male` and `female`. It is possible to use other values.

### zoneinfo

The preferred tz database name, e.g. `Asia/Hong_Kong`.
The list of tz database names are constants, so the input control is a dropdown.

### locale

The preferred BCP47 tag, e.g. `zh-HK`.
The input control is a dropdown, and the options are the supported languages of the project.
When it is populated, the supported languages are taken into account.

### birthdate

The supported format is `YYYY-MM-DD`, representing the birthdate, e.g. 1 Jan 1992.

> The OIDC spec also allows `0000-MM-DD` for birthday and `YYYY` for year of birth.
> We only support `YYYY-MM-DD` for simplicity.

## List of standard attributes that are coupled with identities

- `email`
- `email_verified`
- `phone_number`
- `phone_number_verified`
- `preferred_username`

When the identities of the end-user have been changed, the following steps are taken to compute the above attributes:

1. Generate a list of email address candidates from the identities, candidates from newer identities are ordered first.
1. Generate a list of phone number candidates from the identities, candidates from newer identities are ordered first.
1. Generate a list of username candidates from the identities, candidates from newer identities are ordered first.
1. If the `email` standard attribute does not refer to a candidate in the list, clear it.
1. If the `phone_number` standard attribute does not refer to a candidate in the list, clear it.
1. If the `preferred_username` standard attribute does not refer to a candidate in the list, clear it.
1. If the `email` standard attribute is absent and the list of candidate is non-empty, set it to the first candidate in the list.
1. If the `phone_number` standard attribute is absent and the list of candidate is non-empty, set it to the first candidate in the list.
1. If the `preferred_username` standard attribute is absent and the list of candidate is non-empty, set it to the first candidate in the list.

### email

The primary email address of the end-user.
The value always comes from one of the identity the end-user has.
Therefore, the input control is a dropdown.

### email\_verified

Whether the primary email address is verified.
This attribute is present when `email` is present.
This attribute is read-only.

### phone\_number

The primary phone number of the end-user.
The value always from from one of the identity the end-user has.
Therefore, the input control is a dropdown.

### phone\_number\_verified

Whether the primary phone number is verified.
This attribute is present when `phone_number` is present.
This attribute is read-only.

### preferred\_username

The primary username of the end-user.
The value always comes from one of the identity the end-user has.
Therefore, the input control is a dropdown.

## Custom attributes

In addition to standard attributes, the developer can define their own custom attributes.

> Role based access control should be done with the roles feature!
> Custom attributes should not be used!

### Defining custom attributes

The custom attributes are defined as a single JSON schema written against a subset of JSON schema 2019-09.

Here is an example of the schema of the custom attributes.

```yaml
user_profile:
  custom_attributes:
    schema:
      properties:
        stripe_customer_id:
          type: string
        hobby:
          type: string
```

All changes made to custom attributes must be valid against the schema.

#### Supported subset of JSON schema 2019-09

- `type`: `boolean`, `string`, `number`, `integer`
- `format`: `email`, `phone`, `uri`, `date-time`
- `enum`
- `multipleOf`
- `maximum`
- `exclusiveMaximum`
- `minimum`
- `exclusiveMinimum`
- `maxLength`
- `minLength`
- `properties`

### Access control on custom attributes

The Admin API and the portal always have full access to all custom attributes.

```yaml
user_profile:
  custom_attributes:
    access_control:
    - pointer: /app_user_role
      access_control: internal
    - pointer: /hobby
      access_control: readwrite
```

Possible values of `access_control` are:

- `hidden`: The custom attribute is hidden from the end-user, the User Info endpoint and the resolver.
- `internal`: The custom attribute is hidden from the settings page. But it is visible to the User Info endpoint and the resolver. This value is for custom attributes that are for internal use, like role.
- `readonly`: The custom attribute is visible to the end-user, the User Info endpoint and the resolver. The end-user can view but cannot edit it in the settings page.
- `readwrite`: The custom attribute is visible to the end-user, the User Info endpoint and the resolver. The end-user can view and edit it in the settings page.

The default value of `access_control` of all custom attributes is `internal`.

### Editing custom attributes in the settings page and in the portal

The UI control of the custom attribute is determined by the `type` and the `format`.

- `type: string` is `<input type="text">`.
- `type: boolean` is `<input type="checkbox">`.
- `type: number` is `<input type="text">` but restricted to be a number.
- `type: integer` is `<input type="text">` but restricted to be an integer.
- `type: string` and `format: email` is `<input type="email">`.
- `type: string` and `format: phone` is rendered using a phone input library.
- `type: string` and `format: uri` is `<input type="text">` with validation.
- `type: string` and `format: date-time` is rendered using a library.

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

See [user.profile.updated](../event.md#userprofileupdated).

## Rationale on access control

The User Info endpoint is visible to all clients.
The resolver endpoint is public, as long as someone has a valid access token or IDP session, they can call the resolver endpoint.
The end-user and the client can access the User Info endpoint with an access token.
The end-user and the client can also access the User Info endpoint with an IDP session.

The User Info endpoint and the resolver endpoint should share the same level of access control.
Imagine if the resolver endpoint has more privilege than the User Info endpoint,
there is a loophole to call the resolver endpoint instead of the User Info endpoint to access more information.

## Use case examples

### Using Authgear just for authentication.

The developer does not want Authgear to manage the user profile for them.
The developer can just ignore custom attributes.
The developer have to manually opt-out standard attributes by hiding them.

```yaml
user_profile:
  standard_attributes:
    access_control:
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

### Using Authgear for storing additional user information

The developer can define custom attributes for storing additional user information.

```JSON
{
  "properties": {
    "hobby": {
      "type": "string"
    }
  }
}
```

The developer directs the end-user to the settings page to edit standard attributes, as well as the custom attributes.

The developer calls the User Info endpoint to retrieve the standard attributes, and the custom attributes.

Finally, the developer can display the attributes in their application.
