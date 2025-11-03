* [User profile](#user-profile)
  * [Standard attributes](#standard-attributes)
  * [Standard attributes population](#standard-attributes-population)
  * [List of standard attributes subject to population](#list-of-standard-attributes-subject-to-population)
    * [Single\-line string standard attributes](#single-line-string-standard-attributes)
    * [Multi\-line string standard attributes](#multi-line-string-standard-attributes)
    * [URL standard attributes](#url-standard-attributes)
    * [gender](#gender)
    * [zoneinfo](#zoneinfo)
    * [locale](#locale)
    * [birthdate](#birthdate)
  * [List of standard attributes that are coupled with identities](#list-of-standard-attributes-that-are-coupled-with-identities)
    * [email](#email)
    * [email\_verified](#email_verified)
    * [phone\_number](#phone_number)
    * [phone\_number\_verified](#phone_number_verified)
    * [preferred\_username](#preferred_username)
  * [Custom attributes](#custom-attributes)
    * [Defining custom attributes](#defining-custom-attributes)
      * [Custom Attribute type string](#custom-attribute-type-string)
      * [Custom Attribute type integer](#custom-attribute-type-integer)
      * [Custom Attribute type number](#custom-attribute-type-number)
      * [Custom Attribute type enum](#custom-attribute-type-enum)
      * [Custom Attribute type phone\_number](#custom-attribute-type-phone_number)
      * [Custom Attribute type email](#custom-attribute-type-email)
      * [Custom Attribute type url](#custom-attribute-type-url)
      * [Custom Attribute type alpha2](#custom-attribute-type-alpha2)
  * [Roles](#roles)
  * [Access control](#access-control)
    * [The end\-user](#the-end-user)
    * [The session bearer](#the-session-bearer)
    * [The portal](#the-portal)
    * [The Admin API](#the-admin-api)
    * [Available access control levels](#available-access-control-levels)
    * [Access control configuration](#access-control-configuration)
    * [Exhaustive list of access control combination](#exhaustive-list-of-access-control-combination)
  * [ID Token](#id-token)
  * [User Info endpoint](#user-info-endpoint)
    * [Special Claims](#special-claims)
  * [Webhook](#webhook)
  * [Use case examples](#use-case-examples)
    * [Using Authgear just for authentication\.](#using-authgear-just-for-authentication)
    * [Using Authgear for storing additional user information](#using-authgear-for-storing-additional-user-information)

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

The custom attributes are defined with a configuration.

Here is an example of the configuration.

```
user_profile:
  custom_attributes:
    attributes:
    - id: "0001"
      pointer: /x_phone_number
      type: phone_number
    - id: "0002"
      pointer: /x_email
      type: email
```

Each custom attribute MUST have an unique ID, an unique pointer, and of one of the defined types.

Internally, all custom attributes are stored in a JSON object using the ID as the key.
Given the above configuration, the custom attributes in storage form may appear as
```json
{
  "0001": "+85298765432",
  "0002": "user@example.com"
}
```

The pointer of custom attribute MUST have exactly ONE level, as in `/this_is_one_level`.
The pointer MUST also be non-empty, so `/` is not a valid pointer.

Only `a-z`, `A-Z`, `0-9` and `_` character are allowed in the pointer of custom attribute.

The pointer MUST NOT conflict with the pointer of any standard attributes,
so the developer CANNOT define a custom attribute with pointer `/email`.

Once a custom attribute is defined, it CANNOT be removed.
The `id` and the `type` CANNOT be changed.

> If custom attribute were allowed to be removed, or its type were allowed to be changed, the developer can remove it and define
> a new custom attribute with the same ID but a incompatible type.
> This will result in a situation that is very complicated to handle.
> Given that it is extremely easy to define custom attribute with just a few clicks,
> a few clicks should never lead to such a complicated scenario.

The `pointer` can be changed as long as the developer is aware of the consequence of the rename.
They have to be prepared for receiving custom attributes shown in the new pointer.

The developer is allowed to freely change other supplementary configuration of a particular type of custom attribute.
However, it is the developer's responsible to make sure they can handle that situation.

The label of the custom attribute can be localized with the translation key `custom-attribute-label-{pointer}`.
For example, if the custom attribute is `/x_email`, then the translation key is `custom-attribute-label-/x_email`.

If the label of the custom attribute is not localized, a default label is generated based on the pointer.
`_` are replaced by space character, and the first character of each word is capitalized.
For example, the default label of `/job_title` is `Job Title`.

#### Custom Attribute type `string`

The custom attribute is of type string.
The UI control of it is a text field.

#### Custom Attribute type `integer`

The custom attribute is of type integer.

Optionally, the developer can define the minimum and the maximum allowed value.
For example, if the developer wants to define a custom attribute of non-negative integer, they write

```
- id: "0000"
  pointer: /x_age
  type: integer
  minimum: 0
  maximum: 200
```

If the valid range is narrowed, future write of the value will fail validation.

The UI control of it is a text field restricted to integers.

#### Custom Attribute type `number`

The custom attribute is of type number.

Optionally, the developer can define the minimum and the maximum allowed value.
For example, if the developer wants to define a custom attribute of non-negative number, they write

```
- id: "0000"
  pointer: /hourly_wage
  type: number
  minimum: 0.0
  maximum: 100.0
```

If the valid range is narrowed, future write of the value will fail validation.

The UI control of it is a text field restricted to numbers.

#### Custom Attribute type `enum`

`enum` is a string, but can only take one of the values defined by the developer.

For example,

```
- id: "0000"
  pointer: /x_rank
  type: enum
  enum: ["junior", "senior", "staff"]
```

If any of the value is removed, the value can still be displayed.
But future write will reject that removed value.

The UI control of it is a dropdown.

The label of the values can be localized with the translation key `custom-attribute-enum-label-{pointer}-{value}`.

For example, to localize the label of `junior`, provide the translation key `custom-attribute-enum-label-/x_rank-junior`.

If localization is not provided, the value itself is used as label.

#### Custom Attribute type `phone_number`

`phone_number` is a string, and the value must be in the format of E.164.

The UI control of it is a text field with validation.

#### Custom Attribute type `email`

`email` is a string, and the value must be an email address without name.

The UI control of it is a text field with validation.

#### Custom Attribute type `url`

`url` is a string, and the value must be an URL of any scheme.

The UI control of it is a text field with validation.

#### Custom Attribute type `alpha2`

`alpha2` is a string, and the value must be one of the ISO 3166-1 alpha-2 codes.

The UI control of it is a dropdown.

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

### The portal

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

- The default access control level for `portal_ui` is `readwrite`.
- The default access control level is `bearer` is `readonly`.
- The default access control level of standard attribute for `end_user` is `readwrite`.
- The default access control level of custom attribute for `end_user` is `hidden`.

Here is an example with default values shown.

```yaml
user_profile:
  standard_attributes:
    access_control:
    - pointer: /given_name
      access_control:
        end_user: readwrite
        bearer: readwrite
        portal_ui: readwrite
  custom_attributes:
    attributes:
    - id: "0000"
      pointer: /hobby
      type: string
      access_control:
        end_user: hidden
        bearer: readwrite
        portal_ui: readwrite
```

### Exhaustive list of access control combination

|end\_user|bearer|portal\_ui|
|---------|------|-----------|
|hidden|hidden|hidden|
|hidden|hidden|readonly|
|hidden|hidden|readwrite|
|hidden|readonly|readonly|
|hidden|readonly|readwrite|
|readonly|readonly|readonly|
|readonly|readonly|readwrite|
|readwrite|readonly|readwrite|

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
  "https://authgear.com/claims/user/roles": ["manager"]
}
```

### Special Claims

Below is a list of special claims returned in the OIDC User Info endpoint:

#### https://authgear.com/claims/user/roles

An array of effective roles of the user, including roles assigned directly and roles assigned through a group. Read [Roles and Groups](../roles-groups.md) for details.

#### https://authgear.com/claims/user/is_anonymous

A boolean. True if the user is a anonymous user. Read [Anonymous User](../anonymous-user.md) for details.

#### https://authgear.com/claims/user/is_verified

A boolean. True if the user is a verified user. i.e. Have at least one verified identity.

#### https://authgear.com/claims/user/can_reauthenticate

A boolean. True if the user can perform reauthentication.


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
        portal_ui: hidden
    - pointer: /phone_number
      access_control:
        end_user: hidden
        bearer: hidden
        portal_ui: hidden
    - pointer: /preferred_username
      access_control:
        end_user: hidden
        bearer: hidden
        portal_ui: hidden
    - pointer: /given_name
      access_control:
        end_user: hidden
        bearer: hidden
        portal_ui: hidden
    - pointer: /family_name
      access_control:
        end_user: hidden
        bearer: hidden
        portal_ui: hidden
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
