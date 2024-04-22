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

## Supported Providers

### ADFS

#### Constructing the User Profile

The ADFS user profile is obtained by decoding the id token returned from the token endpoint.
The token endpoint is obtained using `discovery_document_endpoint`.

For available claims, please read:
https://learn.microsoft.com/en-us/windows-server/identity/ad-fs/development/ad-fs-openid-connect-oauth-concepts#claims

Note that, the claims inside the id token is customizable and could be different for each ADFS setup.

Then, the claims are passed into the [The generic standard attribute extraction algorithm](#the-generic-standard-attribute-extraction-algorithm).

The resulting json object will be stored as the authgear user profile of that identity.

### Apple

#### Constructing the User Profile

The apple user profile is obtained by decoding the id token returned from the token endpoint.

For available claims, please read:
https://developer.apple.com/documentation/sign_in_with_apple/sign_in_with_apple_rest_api/authenticating_users_with_sign_in_with_apple#3383773

Then, the claims are passed into the [The generic standard attribute extraction algorithm](#the-generic-standard-attribute-extraction-algorithm).

The resulting json object will be stored as the authgear user profile of that identity.

### Azure B2C

#### Constructing the User Profile

The Azure B2C user profile is obtained by decoding the id token returned from the token endpoint.
The token endpoint is obtained using the discovery endpoint `"https://{tenant}.b2clogin.com/{tenant}.onmicrosoft.com/{policy}/v2.0/.well-known/openid-configuration"`. Where `tenant` and `policy` are available configs of this provider.

Please read this documents for details:
https://learn.microsoft.com/en-us/azure/active-directory-b2c/identity-provider-generic-openid-connect?pivots=b2c-user-flow

Note that, the claims inside the id token is customizable and could be different for each Azure B2C setup.

Then, the claims are passed into the [The generic standard attribute extraction algorithm](#the-generic-standard-attribute-extraction-algorithm).

The resulting json object will be stored as the authgear user profile of that identity.

### Azure AD

#### Constructing the User Profile

The Azure AD user profile is obtained by decoding the id token returned from the token endpoint.
The token endpoint is obtained using the discovery endpoint `"https://login.microsoftonline.com/{tenant}/v2.0/.well-known/openid-configuration`. Where `tenant` is an available config of this provider.

Please read this documents for details:
https://learn.microsoft.com/en-us/entra/identity-platform/v2-protocols-oidc

And for the available claims, read:
https://learn.microsoft.com/en-us/entra/identity-platform/id-token-claims-reference

Then, the claims are passed into the [The generic standard attribute extraction algorithm](#the-generic-standard-attribute-extraction-algorithm).

The resulting json object will be stored as the authgear user profile of that identity.

### Faceboook

#### Constructing the User Profile

The facebook user profile is obtained by this api:
https://graph.facebook.com/v11.0/me?fields=id,email,first_name,last_name,middle_name,name,name_format,picture,short_name

Then, we will construct a user profile from the response by the following process:

1. Map the following fields from the response into the user profile.

- From `email` to `email`
- From `first_name` to `given_name`
- From `last_name` to `family_name`
- From `name` to `name`
- From `short_name` to `nickname`
- From `picture.data.url` to `picture`

1. The resulting user profile will be passed into the [The generic standard attribute extraction algorithm](#the-generic-standard-attribute-extraction-algorithm).

The resulting json object will be stored as the authgear user profile of that identity.

### Github

#### Constructing the User Profile

The github user profile is obtained by this api:
https://api.github.com/user

Then, we will construct a user profile from the response by the following process:

1. Map the following fields from the response into the user profile.

- From `email` to `email`
- From `login` to `name`
- From `login` to `given_name`
- From `avatar_url` to `picture`
- From `html_url` to `profile`

1. The resulting user profile will be passed into the [The generic standard attribute extraction algorithm](#the-generic-standard-attribute-extraction-algorithm).

The resulting json object will be stored as the authgear user profile of that identity.

### Linkedin

#### Constructing the User Profile

We construct the user using responses from two apis:

- The me api: https://api.linkedin.com/v2/me?projection=(id,localizedFirstName,localizedLastName,profilePicture(displayImage~digitalmediaAsset:playableStreams))
- The contact api: https://api.linkedin.com/v2/clientAwareMemberHandles?q=members&projection=(elements*(primary,type,handle~))

The following steps are used to construct the user profile:

1. Extract `email` from the response of the contact api, by finding one item inside the `elements` field, which:

- `primary` is `true`.
- `type` is `EMAIL`.
- If the above 2 conditions are met, read `handle~.emailAddress` and use it as the `email` field of the user profile.

1. Extract `localizedFirstName` as `given_name`.
1. Extract `localizedLastName` as `family_name`.
1. Extract `profilePicture.displayImage.elements[last_item].identifiers[first_item].identifier` as `picture`.
1. Pass the constructed object into the [The generic standard attribute extraction algorithm](#the-generic-standard-attribute-extraction-algorithm).

The resulting json object will be stored as the authgear user profile of that identity.

### Google

#### Constructing the User Profile

The google user profile is obtained by decoding the id token returned from the token endpoint.

For available claims, please read:
https://developers.google.com/identity/openid-connect/openid-connect#an-id-tokens-payload

Then, the claims are passed into the [The generic standard attribute extraction algorithm](#the-generic-standard-attribute-extraction-algorithm).

The resulting json object will be stored as the authgear user profile of that identity.

### Wechat

#### Constructing the User Profile

The wechat user profile is constructed from the response of this api:
https://api.weixin.qq.com/sns/userinfo

The following steps are used to construct the user profile:

1. If `sex` is `1`, set `gender` to `male`, else if it is `2`, set gender to `female`.
1. Extract `nickname` as `name`.
1. Copy `name` to `given_name`.
1. Extract `language` as `locale`.
1. Pass the constructed object into the [The generic standard attribute extraction algorithm](#the-generic-standard-attribute-extraction-algorithm).

The resulting json object will be stored as the authgear user profile of that identity.

## The generic standard attribute extraction algorithm

Authgear implements a algorithm to extract standard attributes from a oauth provider user profile.

It extracts the following fields as standard attributes if it exist in the oauth provider user profile:

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
