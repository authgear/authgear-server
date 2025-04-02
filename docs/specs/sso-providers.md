* [SSO Providers](#sso-providers)
  * [OAuth User Profile](#oauth-user-profile)
  * [Identity Attributes](#identity-attributes)
  * [Algorithm for converting the OAuth User Profile to the identity attributes](#algorithm-for-converting-the-oauth-user-profile-to-the-identity-attributes)
  * [Supported Providers](#supported-providers)
    * [ADFS](#adfs)
      * [Constructing the OAuth User Profile](#constructing-the-oauth-user-profile)
    * [Apple](#apple)
      * [Constructing the OAuth User Profile](#constructing-the-oauth-user-profile-1)
    * [Azure B2C](#azure-b2c)
      * [Constructing the OAuth User Profile](#constructing-the-oauth-user-profile-2)
    * [Azure AD](#azure-ad)
      * [Constructing the OAuth User Profile](#constructing-the-oauth-user-profile-3)
    * [Facebook](#facebook)
      * [Constructing the OAuth User Profile](#constructing-the-oauth-user-profile-4)
    * [Github](#github)
      * [Constructing the OAuth User Profile](#constructing-the-oauth-user-profile-5)
    * [Linkedin](#linkedin)
      * [Constructing the OAuth User Profile](#constructing-the-oauth-user-profile-6)
    * [Google](#google)
      * [Constructing the OAuth User Profile](#constructing-the-oauth-user-profile-7)
    * [Wechat](#wechat)
      * [Constructing the OAuth User Profile](#constructing-the-oauth-user-profile-8)

# SSO Providers

This documents listed out the current supported providers, and their corresponding behaviors.

- [ADFS](#adfs)
- [Apple](#apple)
- [Azure B2C](#azure-b2c)
- [Azure AD](#azure-ad)
- [Facebook](#facebook)
- [Github](#github)
- [Linkedin](#linkedin)
- [Google](#google)
- [Wechat](wechat)

## OAuth User Profile

OAuth User Profile is JSON object representing an OAuth account from an OAuth provider.
The actual shape varies with the type of the OAuth provider.

## Identity Attributes

See [Identity Attributes](./glossary.md.md#identity-attributes)

## Algorithm for converting the OAuth User Profile to the identity attributes

It extracts the following fields as identity attributes if it exist in the OAuth User Profile:

- `name`: string
- `given_name`: string
- `family_name`: string
- `middle_name`: string
- `nickname`: string
- `preferred_username`: string
- `profile`: url
- `picture`: url
- `website`: url
- `email`: string
- `email_verified`: boolean
- `gender`: string
- `birthdate`: string
- `zoneinfo`: string
- `locale`: string
- `phone_number`: string
- `phone_number_verified`: boolean
- `address`: object
  - `address.formatted`: string
  - `address.street_address`: string
  - `address.locality`: string
  - `address.region`: string
  - `address.postal_code`: string
  - `address.country`: string

Then, the extracted attributes will be normalized according to the following steps.

1. Normalize the `email` if it exist, respecting the current project settings on email normalization.
1. Normalize the `phone_number` if it exist, to E.164 format.
1. For all string fields, delete the key if it is a empty string, or it is not a string.
1. For all booleans, delete the key if is not a boolean.
1. For all urls, delete the key if it is not a valid uri.
1. Ensure `birthdate` is a valid date string, else remove the key.
1. Ensure `zoneinfo` is in valid timezone format, else remove the key.
1. Ensure `locale` is a valid locale, else remove the key.

## Supported Providers

### ADFS

#### Constructing the OAuth User Profile

The OAuth User Profile is obtained by decoding the id token returned from the token endpoint.
The token endpoint is obtained using `discovery_document_endpoint`.

For available claims, please read:
https://learn.microsoft.com/en-us/windows-server/identity/ad-fs/development/ad-fs-openid-connect-oauth-concepts#claims

Note that, the claims inside the id token is customizable and could be different for each ADFS setup.

### Apple

#### Constructing the OAuth User Profile

The OAuth User Profile is obtained by decoding the id token returned from the token endpoint.

For available claims, please read:
https://developer.apple.com/documentation/sign_in_with_apple/sign_in_with_apple_rest_api/authenticating_users_with_sign_in_with_apple#3383773

### Azure B2C

#### Constructing the OAuth User Profile

The OAuth User Profile is obtained by decoding the id token returned from the token endpoint.
The token endpoint is obtained using the discovery endpoint `"https://{tenant}.b2clogin.com/{tenant}.onmicrosoft.com/{policy}/v2.0/.well-known/openid-configuration"`. Where `tenant` and `policy` are available configs of this provider.

Please read this documents for details:
https://learn.microsoft.com/en-us/azure/active-directory-b2c/identity-provider-generic-openid-connect?pivots=b2c-user-flow

Note that, the claims inside the id token is customizable and could be different for each Azure B2C setup.

### Azure AD

#### Constructing the OAuth User Profile

The OAuth User Profile is obtained by decoding the id token returned from the token endpoint.
The token endpoint is obtained using the discovery endpoint `"https://login.microsoftonline.com/{tenant}/v2.0/.well-known/openid-configuration`. Where `tenant` is an available config of this provider.

Please read this documents for details:
https://learn.microsoft.com/en-us/entra/identity-platform/v2-protocols-oidc

And for the available claims, read:
https://learn.microsoft.com/en-us/entra/identity-platform/id-token-claims-reference

### Facebook

#### Constructing the OAuth User Profile

The OAuth User Profile is obtained by this api:
https://graph.facebook.com/v11.0/me?fields=id,email,first_name,last_name,middle_name,name,name_format,picture,short_name

Then, we will construct the OAuth User Profile from the response by the following process:

1. Map the following fields from the response to the OAuth User Profile.

- From `email` to `email`
- From `first_name` to `given_name`
- From `last_name` to `family_name`
- From `name` to `name`
- From `short_name` to `nickname`
- From `picture.data.url` to `picture`

### Github

#### Constructing the OAuth User Profile

The OAuth User Profile is obtained by this api:
https://api.github.com/user

Then, we will construct the OAuth User Profile from the response by the following process:

1. Map the following fields from the response to the OAuth User Profile.

- From `email` to `email`
- From `login` to `name`
- From `login` to `given_name`
- From `avatar_url` to `picture`
- From `html_url` to `profile`

### Linkedin

#### Constructing the OAuth User Profile

The OAuth User Profile is constructed using the responses from these APIs:

- The me api: https://api.linkedin.com/v2/me?projection=(id,localizedFirstName,localizedLastName,profilePicture(displayImage~digitalmediaAsset:playableStreams))
- The contact api: https://api.linkedin.com/v2/clientAwareMemberHandles?q=members&projection=(elements*(primary,type,handle~))

The following steps are used to construct the OAuth User Profile:

1. Extract `email` from the response of the contact api, by finding one item inside the `elements` field, which:

- `primary` is `true`.
- `type` is `EMAIL`.
- If the above 2 conditions are met, read `handle~.emailAddress` and use it as the `email` field of the user profile.

1. Extract `localizedFirstName` as `given_name`.
1. Extract `localizedLastName` as `family_name`.
1. Extract `profilePicture.displayImage.elements[last_item].identifiers[first_item].identifier` as `picture`.

### Google

#### Constructing the OAuth User Profile

The OAuth User Profile is obtained by decoding the id token returned from the token endpoint.

For available claims, please read:
https://developers.google.com/identity/openid-connect/openid-connect#an-id-tokens-payload

### Wechat

#### Constructing the OAuth User Profile

The OAuth User Profile is constructed from the response of this API:
https://api.weixin.qq.com/sns/userinfo

The following steps are used to construct the OAuth User Profile:

1. If `sex` is `1`, set `gender` to `male`, else if it is `2`, set gender to `female`.
1. Extract `nickname` as `name`.
1. Copy `name` to `given_name`.
1. Extract `language` as `locale`.
