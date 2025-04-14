* [Convention](#convention)
  * [User Profile Pointer](#user-profile-pointer)
    * [Known violation of User Profile Pointer: User profile definition](#known-violation-of-user-profile-pointer-user-profile-definition)
    * [Known violation of User Profile Pointer: SAML NameID attribute](#known-violation-of-user-profile-pointer-saml-nameid-attribute)
    * [Known violation of User Profile Pointer: Authentication flow user\_profile step](#known-violation-of-user-profile-pointer-authentication-flow-user_profile-step)
  * [Identity Attributes Pointer](#identity-attributes-pointer)
    * [Known violation of Identity Attributes Pointer: Account linking identity attributes mapping](#known-violation-of-identity-attributes-pointer-account-linking-identity-attributes-mapping)

# Convention

This text documents the convention used in Authgear configuration files.

## User Profile Pointer

When a configuration needs to refer to an attribute of the [User Info](./glossary.md#user-info), the configuration MUST look like:

```yaml
user_profile:
  pointer: /email
```

Additional fields can be added to the object, for example:

```yaml
user_profile:
  pointer: /email
  required: true
```

When a configuration needs to refer to a list of attributes of the [User Info](./glossary.md#user-info), the configuration MUST look like:

```yaml
a_good_name:
- user_profile:
    pointer: /email
    required: true
- user_profile:
    pointer: /phone_number
    required: true
```

### Known violation of User Profile Pointer: User profile definition

It is defined as:

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

Technically it is not a violation since it is defining the attribute, not refering to the attribute.

### Known violation of User Profile Pointer: SAML NameID attribute

It is defined as:

```yaml
saml:
  service_providers:
  - client_id: EXAMPLE_ID
    nameid_format: urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified
    nameid_attribute_pointer: /email
```

If it followed the convention, it would be:

```yaml
saml:
  service_providers:
  - client_id: EXAMPLE_ID
    nameid_format: urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified
    nameid_attribute:
      user_profile:
        pointer: /email
```

### Known violation of User Profile Pointer: Authentication flow user_profile step

It is defined as:

```yaml
authentication_flow:
  signup_flows:
  - name: default
    steps:
    - type: user_profile
      user_profile:
      - pointer: /x_age
        required: true
      - pointer: /x_hobby
        required: true
```

If it followed the convention, it would be:

```yaml
authentication_flow:
  signup_flows:
  - name: default
    steps:
    - type: user_profile
      user_profile:
      - user_profile:
          pointer: /x_age
          required: true
      - user_profile:
          pointer: /x_hobby
          required: true
```

## Identity Attributes Pointer

When a configuration needs to refer to an attribute of [Identity Attributes](./glossary.md#identity-attributes), the configuration MUST look like:

```yaml
identity_attributes:
  pointer: /email
```

### Known violation of Identity Attributes Pointer: Account linking identity attributes mapping

It is defined as:

```yaml
identity:
  oauth:
    providers:
    - alias: adfs
      client_id: exampleclientid
      type: adfs
      user_profile_mapping:
      - oauth_claim:
          pointer: "/primary_phone"
        user_profile:
          pointer: "/phone_number"
```

It it followed the convention: it would be:

```yaml
identity:
  oauth:
    providers:
    - alias: adfs
      client_id: exampleclientid
      type: adfs
      identity_attributes_mapping:
      - oauth_claim:
          pointer: "/primary_phone"
        identity_attributes:
          pointer: "/phone_number"
```
