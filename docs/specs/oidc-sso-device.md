# Device Single Sign On

This is a SSO mechanism to share authentication session between applications on a single device without using browser cookies.

## Abstract

This is a implementation based on the [OIDC Native SSO Specification](https://openid.net/specs/openid-connect-native-sso-1_0.html).

It is expected to be used by mobile apps to share the authentication session with another app installed in the same device.

The authgear server will generate a device token with an id token for a single device. Applications on the device are able to share authentication session using that device token and id token.

Applications which are able to share authentication session are expected to be signed by the same vendor certificate, and are expected share the device token and id token by a shared storage which only accessible by applications signed by the same vendor certificate.

## Implementation details

### Server-side

- Token Endpoint

  - Support a new scope `device_sso`
    - If `device_sso` was included:
      - If no valid `device_secret` was provided from the client:
        - Create a new device grant, which represents a new shared session on the device.
        - Generate `device_secret` from the device grant. The token is a random string. Use the token hash to map it to the device grant.
      - If a valid `device_secret` was provided from the client:
        - Reuse the device grant id from the provided `device_secret`.
        - A new `device_secret` was generated which shares the same device grant id.
      - The device grant id will be included in the generated `refresh_token`.
      - The lifetime of the device grant is the longest lifetime of its associated offline grants, therefore it will be extended when new refresh tokens are generated.
      - Mark refresh token's `device_sso_enabled` to `true`.
      - Include `device_secret` in the token response.
      - Include `ds_hash` in `id_token` in the token response. `ds_hash` is the hash of `device_secret`.
      - `x_suppress_idp_session_cookie` will be implied as `true`, because device sso cannot be used together with browser sso. If user specifies `x_sso_enabled=true` at the same time, it will be an error.
  - If `grant_type=urn:authgear:params:oauth:grant-type:id-token`:
    - Include `device_secret` in the token response if the scopes include `device_sso`. Generate one if one was not provided by client.

- Logout

  - Revoke refresh token
    - If `device_sso_enabled` of the refresh token is `true`:
      - Revoke the device grant
      - Revoke all refresh tokens associated with the same device grant id

- Configuration

  - `x_device_sso_group` was added to `oauth.clients`
  - When `x_device_sso_group` is not empty, the client can participate in device SSO.
  - Only clients with the same value in `x_device_sso_group` can share authentication session.

- Session listing
  - Combining Sessions
    - Both the settings page, admin API and Portal combine sessions in the same way
    - Refresh tokens associated to the same device grant id will be combined into a single entry. Since revoking one of them will also revoke the others. The entry will be shown as a single entry without grouping.
    - The sessions that cannot be combined will be listed separately without grouping. (Refresh tokens with `device_sso_enabled=false`)

### Web

Device SSO is not supported in web application.

### iOS

The iOS SDK accepts the following parameters during initialization.

- `isDeviceSSOEnabled`

  - If `true`, `device_sso` will be included in `scope` of authorization requests.
  - On everytime the sdk called the token endpoint, the sdk will update `device_secret` and `id_token` in the `deviceTokenStore` if provided.

- `deviceTokenStore`
  - A shared store between all applications which are expected to share authentication session. Stores `device_sercret` and `id_token`.
  - An implementation using iOS Keychain with AppGroup is included in the SDK. Developer is required to provide the access group id to initialize the store. Applications must belong to the given app group or the SDK might malfunction. For details read [this document](https://developer.apple.com/documentation/security/keychain_services/keychain_items/sharing_access_to_keychain_items_among_a_collection_of_apps).

#### Known issues

- The iOS keychain will not be cleaned up even all apps in the app group were removed.

### Android

The Android SDK accepts the following parameters during initialization.

- `isDeviceSSOEnabled`

  - If `true`, `device_sso` will be included in `scope` of authorization requests.
  - On everytime the sdk called the token endpoint, the sdk will update `device_secret` and `id_token` in the `deviceTokenStore` if provided.

- `deviceTokenStore`
  - A shared store between all applications which are expected to share authentication session. Stores `device_sercret` and `id_token`.
  - An implementation using android account manager is included in the SDK. Developer is required to provide the `accountType` to initialize the store. Applications must belong to the given app group or the SDK might malfunction. Developer must also define a `<account-authenticator>` resource with the same `accountType`, and a `<service>` which uses the defined account authenticator in all apps that is sharing the authentication session. For details read [this document](https://developer.android.com/reference/android/accounts/AbstractAccountAuthenticator).

#### Known issues

- The first installed app will be the "authenticator" app in the android, which in fact owns the accounts. Once the app was removed, the accounts will be removed together with the app.
