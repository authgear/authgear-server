# User profile

User profile consists of standard attributes, custom attributes and roles.

## Standard attributes

Standard attributes are [OIDC standard claims](https://openid.net/specs/openid-connect-core-1_0.html#StandardClaims).
All standard attributes are optional.

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

## Roles

A role is an opaque string defined by the developer.

The Admin API is the only way to manipulate roles.
The portal uses the Admin API to offer a GUI to manage roles.

The developer can:

- Create a role
- Delete a role
- Rename a role
- Add a role to an end-user
- Remove a role from an end-user

A role internally is referenced by an ID, however, it is referenced by its name externally.

An end-user has zero or more roles.

The name of a role must only consist of alphanumeric characters, hyphen, dot and underscore.
The name of a role must be non-empty.

## Access control

Access control on standard attributes and custom attributes are defined per party.
There are 4 parties.

Roles are always publicly readable.

### The end-user

The end-user can view or edit the standard attributes and the custom attributes via the settings page.

### The session bearer

The session bearer is someone who has a valid IDP session cookie or a valid access token.
The session bearer can then access the User Info endpoint and the resolver endpoint.

Behind the session bearer, it is the end-user, the mobile app or the website.

### The admin user

The admin user can view or edit the standard attributes and the custom attributes via the portal.
Behind the scene, the portal uses the Admin API.

### The Admin API

The Admin API allows the developer to view or edit the standard attributes and the custom attributes.
The admin API always have full access to all standard attributes and the custom attributes.

### Available access control levels

The access control levels are defined as follows.
They are ordered by in increasing order.

- `hidden`: Hidden from the party.
- `readonly`: Read-only to the party.
- `readwrite`: Read-write to the party.

### Access control configuration

- The default access control level for `admin_user` is `readwrite`.
- The default access control level is `bearer` is `readwrite`.
- The access control level `readwrite` is equivalent to `readonly` for `bearer`.
- The default access control level of standard attribute for `end_user` is `readwrite`.
- The default access control level of custom attribute for `end_user` is `hidden`.
- The access control level for `end_user` must be equal or less than that that for `bearer`.
- The access control level for `bearer` must be equal or less than that for `admin_user`.

Here is an example with default values shown.

```yaml
user_profile:
  standard_attributes:
    access_control:
    - pointer: /given_name
      access_control:
        end_user: readwrite
        bearer: readwrite
        admin_user: readwrite
  custom_attributes:
    access_control:
    - pointer: /hobby
      access_control:
        end_user: hidden
        bearer: readwrite
        admin_user: readwrite
```

### Exhaustive list of access control combination

|end\_user|bearer|admin\_user|
|---------|------|-----------|
|hidden|hidden|hidden|
|hidden|hidden|readonly|
|hidden|hidden|readwrite|
|hidden|readonly|readonly|
|hidden|readonly|readwrite|
|hidden|readwrite|readwrite|
|readonly|readonly|readonly|
|readonly|readonly|readwrite|
|readonly|readwrite|readwrite|
|readwrite|readwrite|readwrite|

## ID Token

The ID token never contain standard attributes nor custom attributes.
This is because ID token can be used as `id_token_hint` and can appear in URL query.

## User Info endpoint

The standard attributes appear in the root of the user info response.
The custom attributes appear under the key `custom_attributes`.
The roles appear under the key `roles`.

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
  },
  "roles": ["manager"]
}
```

## Webhook

The blocking event [user.profile.pre_update](../event.md#userprofilepre-update) fires before an update on the user profile.

The non-blocking event [user.profile.updated](../event.md#userprofileupdated) notifies an update on the user profile.

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
      access_control:
        end_user: hidden
        bearer: hidden
        admin_user: hidden
    - pointer: /phone_number
      access_control:
        end_user: hidden
        bearer: hidden
        admin_user: hidden
    - pointer: /preferred_username
      access_control:
        end_user: hidden
        bearer: hidden
        admin_user: hidden
    - pointer: /given_name
      access_control:
        end_user: hidden
        bearer: hidden
        admin_user: hidden
    - pointer: /family_name
      access_control:
        end_user: hidden
        bearer: hidden
        admin_user: hidden
    # List out other standard attribute you want to hide.
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
