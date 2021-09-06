# Literature review on user profile

This document reviews the design of user profile in various competitors.

## Auth0

The user profile is a JSON object.
The user profile contains root fields, `user_metadata` and `app_metadata`.
Root fields are similar to OIDC standard claims.

`user_metadata` is a JSON object intended for storing additional information, such as hobby. The end-user can read or write this.

`app_metadata` is a JSON object intended for storing authentication and authorization related information. The end-user cannot read nor write this.

`user_metadata` and `app_metadata` together share a maximum size of 16MiB.

By default some root fields are NOT editable because
a connection by default is configured to `on_each_login`,
meaning that on each login, the root fields of a user are updated from
the current user info returned by the identity provider.

A connection can be configured to `on_first_login`,
meaning that if the corresponding root fields are absent, they are populated
from the current user info returned by the identity provider.

When all connections are `on_first_login`, the root fields are editable by the admin.

Auth0 does NOT offer a settings page for the end-user to read or write user profile.

### Auth0 references

- https://auth0.com/docs/users/user-profiles
- https://auth0.com/docs/users/user-profiles/user-profile-structure
- https://auth0.com/docs/users/user-profiles/configure-connection-sync-with-auth0
- https://auth0.com/docs/users/metadata
- https://auth0.com/docs/users/metadata/metadata-fields-data#limitations-and-restrictions

## Okta

Okta user profile is the user profile of an end-user.
App user profile is the user profile of an external identity provider.

The developer can configure attribute mapping to map App user profile to Okta user profile.
The attribute mapping supports a subset of the Spring Expression Language (SpEL) functions.

An attribute can be marked as sensitive so that the end-user cannot read nor write it.

### Okta references

- https://help.okta.com/en/prod/Content/Topics/users-groups-profiles/usgp-about-attribute-mappings.htm
- http://developer.okta.com/docs/getting_started/okta_expression_lang.html
- https://help.okta.com/en/prod/Content/Topics/users-groups-profiles/usgp-hide-sensitive-attributes.htm

## Azure B2C

Azure B2C supports builtin attributes and custom attributes.
The developer has to specify the name, the datatype and the description when defining a custom attribute.

The interaction between Azure B2C and the end-user is called user flow.
The developer can customize the user flow to ask Azure B2C to collect
a builtin attribute or a custom attribute during the signup user flow.

While Azure B2C does not offer a full-feature settings page,
it supports a profile editing user flow.
The developer can customize the user flow to let the user to edit the builtin attributes and custom attributes.

The developer can also read and write builtin attributes and custom attributes with the Microsoft Graph API.

### Azure B2C references

- https://docs.microsoft.com/en-us/azure/active-directory-b2c/user-profile-attributes
- https://docs.microsoft.com/en-us/azure/active-directory-b2c/configure-user-input?pivots=b2c-custom-policy
- https://docs.microsoft.com/en-us/azure/active-directory-b2c/user-flow-custom-attributes?pivots=b2c-user-flow
- https://docs.microsoft.com/en-us/azure/active-directory/external-identities/user-flow-add-custom-attributes
- https://docs.microsoft.com/en-us/azure/active-directory-b2c/add-profile-editing-policy?pivots=b2c-custom-policy

## Amazon Cognito

Amazon Cognito supports standard attributes and custom attributes.

Up to 25 custom attributes can be defined in a user pool.
A custom attribute can have minimum length and maximum length.
A custom attribute can either be a string or a number.
A custom attribute is always optional.
A custom attribute can be immutable or mutable.
Immutable custom attribute cannot be changed after the end-user has provided the value.
The developer can configure the read-write access of a custom attribute per client.

### Amazon Cognito references

- https://docs.aws.amazon.com/cognito/latest/developerguide/user-pool-settings-attributes.html

## Summary

|   |Auth0|Okta|Azure B2C|Amazon Cognito|
|---|-----|----|---------|--------------|
|Standard attributes|Yes|No|Yes|Yes|
|Standard attributes population|Yes|No|No|No|
|Custom attributes|Yes|Yes|Yes|Yes|
|Custom attributes schema|No|Common datatype only|Common datatype only|string or number only|
|Expose to end-user via Web UI|No|No|Yes, via profile editing user flow|No|
|Expose to end-user via User Info endpoint|Yes|Yes|Yes|Yes|
|Custom attributes access control for end-user|No|Yes|Yes, by customizing the user flow|No|
|Custom attributes access control for clients|No|No|No|Yes|
|Attribute mapping|No|Yes|No|No|
