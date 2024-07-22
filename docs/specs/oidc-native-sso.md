- [OIDC Native SSO](#oidc-native-sso)
  * [Key points to note](#key-points-to-note)
  * [Changes on Client](#changes-on-client)
  * [Changes on OfflineGrant](#changes-on-offlinegrant)
  * [Changes on Authorization Endpoint](#changes-on-authorization-endpoint)
  * [Changes on Token Endpoint](#changes-on-token-endpoint)
    + [`grant_type=authorization_code`](#grant_typeauthorization_code)
    + [`grant_type=urn:authgear:params:oauth:grant-type:biometric-request`](#grant_typeurnauthgearparamsoauthgrant-typebiometric-request)
    + [`grant_type=refresh_token`](#grant_typerefresh_token)
    + [`grant_type=urn:authgear:params:oauth:grant-type:app2app`](#grant_typeurnauthgearparamsoauthgrant-typeapp2app)
    + [`grant_type=urn:ietf:params:oauth:grant-type:token-exchange`](#grant_typeurnietfparamsoauthgrant-typetoken-exchange)
    + [When issuing ID tokens](#when-issuing-id-tokens)
  * [Changes on Admin API](#changes-on-admin-api)
  * [Changes on SDK](#changes-on-sdk)
    + [Recipe: Two applications written by the same vendor](#recipe-two-applications-written-by-the-same-vendor)
    + [Recipe: A application opening webapps with custom webview](#recipe-a-application-opening-webapps-with-custom-webview)
  * [Caveats](#caveats)

# OIDC Native SSO

This document specifies the implementation of [OIDC Native SSO](https://openid.net/specs/openid-connect-native-sso-1_0.html).

## Key points to note

- A OfflineGrant used to has only one set of client-specific information, like `client_id` and `refresh_token`. Now, a OfflineGrant can have multiple refresh tokens.
- Since there is only one OfflineGrant, the apps sharing user authentication with Native SSO does not share refresh token. But the underlying session is shared. Thus signing out in one app will sign out all apps. This is by design. See [App2app](./app2app.md) if you want the apps have independent sessions.
- Native SSO is done through the Token Endpoint, thus it requires no user interaction.
- Existing OfflineGrant cannot perform Native SSO. It is because `device_sso` is not in the `scope`. Sign in again to obtain a Native SSO OfflineGrant to perform Native SSO.

## Changes on Client

- Add `x_device_sso_enabled: boolean`.
- `scope=device_sso` is allowed if `x_device_sso_enabled=true`.

> Do we need to add `x_device_sso_key: string` to designate which group of clients can perform Native SSO?
> Only clients with `x_device_sso_enabled=true` AND the same value of `x_device_sso_key` can perform Native SSO with each other.
> This seems very advanced to me.

## Changes on OfflineGrant

- The following fields become client-specific
  - `ClientID`
  - `AuthorizationID`
  - `CreatedAt`
  - `Scopes`
  - `TokenHash`
- Add a new field `RefreshTokens`. It will store all of the above imformation per refresh token.
- Add a new field `DeviceSecretHash`. It is the hex of SHA256 of `device_secret`.
- All new offline grants will use the new `RefreshTokens` to store information about refresh tokens.

## Changes on Authorization Endpoint

- Allow `scope=device_sso` if the client has `x_device_sso_enabled=true`.

## Changes on Token Endpoint

### `grant_type=authorization_code`

- If `authorization_code.scope=device_sso` and `device_secret` is present and it is valid, a new `refresh_token` is added to Native SSO offline grant.
- If `authorization_code.scope=device_sso` and `device_secret` is absent or it is invalid, a new Native SSO offline grant is created.

### `grant_type=refresh_token`

- If `refresh_token.scope=device_sso` and `device_secret` is present and it is valid, nothing to do.
- If `refresh_token.scope=device_sso` and `device_secret` is absent or it is invalid, a new `device_secret` is generated. `DeviceSecretHash` is updated.

### `grant_type=urn:authgear:params:oauth:grant-type:biometric-request`

- Support a new parameter `scope`.
- Allow `scope=device_sso` if the client has `x_device_sso_enabled=true`.
- If `scope=device_sso` and `device_secret` is present and it is valid, a new `refresh_token` is added to Native SSO offline grant.
- If `scope=device_sso` and `device_secret` is absent or it is invalid, a new Native SSO offline grant is created.

### `grant_type=urn:authgear:params:oauth:grant-type:app2app`

Native SSO has no direct impact on app2app.

### `grant_type=urn:ietf:params:oauth:grant-type:token-exchange`

- Validate `scope=device_sso`. Return `error=invalid_request` otherwise. This is because this is the only Token Change flow we support.
- Validate `audience` is the origin of the endpoint.
- Validate `subject_token` is a valid ID token issued to the first app. An expired ID token is still valid. (4.3 Point 2)
- Validate `subject_token_type` is `urn:ietf:params:oauth:token-type:id_token`.
- Validate `actor_token` is a valid `device_secret`. (4.3 Point 1)
- Validate `actor_token_type` is `urn:x-oath:params:oauth:token-type:device-secret`.
- Validate `requested_token_type` is absent.
- Validate `subject_token.ds_hash` is the hex of SHA256 of `actor_token`. (4.3 Point 3)
- Validate `subject_token.sid` is pointing to a valid session. (4.3 Point 4)
- Validate `client_id` and `subject_token.aud` are allowed to perform Native SSO. (4.3 Point 5)
- Validate `scope` is equal or a subset of the `scope` in all refresh tokens of the Native SSO offline grant. (4.3 Point 6)

### When issuing ID tokens

- If the offline grant has `DeviceSecretHash`, set `ds_hash` in the ID token.
- Keep setting `sid`.

## Changes on Admin API

- Add `clientIDs` to `Session`.
- `Session.clientID` is the first client ID of a Native SSO offline grant.

## Changes on SDK

- Rename `isSSOEnabled` to `isBrowserSSOEnabled` in `ConfigureOptions`.
- Add `isDeviceSSOEnabled` to `ConfigureOptions`.
- If `isDeviceSSOEnabled` is true, then `device_sso` is included in authorization requests.
- Add `deviceSecretStore`. It is responsible for storing `device_secret` and `id_token` in Token Response.
- If `device_secret` is found in `deviceSecretStore` and `isDeviceSSOEnabled` is true, then `device_secret` is included in Token Request.
- If `id_token` is present in Token Response, it is persisted into `deviceSecretStore`.
- If `device_secret` is present in Token Response, it is persisted into `deviceSecretStore`.
- `logout()` clears `deviceSecretStore`.
- Add `checkDeviceSSOPossible(): Promise<void>`. It throws error if either `device_secret` or `id_token` is not found in `deviceSecretStore`.
- Add `authenticateDeviceSSO(): Promise<UserInfo>`.
- Expose `refreshToken` on Authgear.

Future works
- Add `IOSAppGroupDeviceSecretStorage`.
- Add `AndroidAccountManagerDeviceSecretStorage`.

### Recipe: Two applications written by the same vendor

> Recipe requires Future works to be done first.

1. Configure both apps to use `IOSAppGroupDeviceSecretStorage` and `AndroidAccountManagerDeviceSecretStorage`.
2. Configure both apps to set `isDeviceSSOEnabled` to true.
3. Sign in normally in App 1.
4. In App 2, call `checkDeviceSSOPossible()`. It returns normally.
5. In App 2, if `checkDeviceSSOPossible()` returns normally, call `authenticateDeviceSSO()`.
6. In App 2, the end-user is authenticated. No user interaction is involved.

### Recipe: A application opening webapps with custom webview

1. Configure the app to set `isDeviceSSOEnabled` to true.
2. Sign in normally.
3. To open a webapp, do the following
4. Construct a new Container with `tokenStorage` set to TransientStorage, and `isDeviceSSOEnabled` to true. The purpose of this is to prevent this Container from messing with the original Container. The new Container is now unauthenticated (due to TransientStorage) but has `device_secret` (due to `isDeviceSSOEnabled` being true).
5. Call `authenticateDeviceSSO()`
6. Inject `refreshToken` of the new Container into the custom webview. This requires knowledge on the implementation details of the Web SDK.
7. The Container of the Web SDK considered itself as authenticated due to the injected refresh token.

## Caveats

- The iOS keychain will not be cleaned up even all apps in the app group were removed.
- Developer is required to provide the `accountType` to initialize the store. Applications must belong to the given app group or the SDK might malfunction. Developer must also define a `<account-authenticator>` resource with the same `accountType`, and a `<service>` which uses the defined account authenticator in all apps that is sharing the authentication session. For details read [this document](https://developer.android.com/reference/android/accounts/AbstractAccountAuthenticator).
- The first installed app will be the "authenticator" app in the android, which in fact owns the accounts. Once the app was removed, the accounts will be removed together with the app.
