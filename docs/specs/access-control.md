# Access Control

## Information

Information is the information of the user.

Currently there are

- Standard Attributes
- Custom Attributes
- Claims

## Party

Party is someone that have access to Information.

Currently there are

- OAuth Client
- The User on web UI
- The Caller of Admin API

## Overview of Access Control between Information and Party

|                       |Standard Attributes|Custom Attributes|Claims                               |
|-----------------------|-------------------|-----------------|-------------------------------------|
|OAuth Client           |No access          |No access        |Read access controlled by OAuth scope|
|The User on web UI     |Read and Write     |No access (yet)  |No access                            |
|The Caller of Admin API|Read and Write     |Read and Write   |Read                                 |

## Access Control of OAuth Client

Standard Attributes and Custom Attributes are invisible to OAuth Client because they are not part of the OAuth standard.
OAuth Client can only see claims that is authorized by the user.

> TODO: As of 2020-11-20, the only supported scope is "https://authgear.com/scopes/full-access", which is the greatest scope.
> In the future, we will introduce much smaller scope such as "email" and "phone_number".
> The returned claims must respect the scope authorized to the client.

> This could be implemented by adding `required_scope` to [claims mapping](./user-model.md#claims-mapping-json-schema).

## Access Control of the User on web UI

The user on web UI can read and write Standard Attributes on the settings page, such as changing the email.

The user on web UI has no access to Custom Attributes yet.
In the future we could let the developer to declare which custom attributes are readable / writable by the user on web UI.
However, this access control is more like a visibility configuration.
It is because it could happen that the custom attribute is not readable by the user on the web UI,
but the attribute is mapped to claims. So the attribute is indirectly readable.

> This could be implemented by letting the developer to declare access control in the [Custom Attributes JSON schema](./user-model.md#custom-attributes-validation).

The user on web UI has no access to Claims. It is because web UI does not rely on Claims.

## Access Control of the Caller of Admin API

The Caller of Admin API has full access to everything.
