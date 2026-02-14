# Biometric authentication

This document explains the design of biometric authentication.

## Prior Art

### Auth0

Auth0 [supports](https://auth0.com/docs/libraries/auth0-swift/auth0-swift-touchid-faceid) biometric authentication by storing [Credentials](https://github.com/auth0/Auth0.swift/blob/master/Auth0/Credentials.swift).

This approach effectively stores the refresh token in the keychain on the device.
Biometric authentication is required to retrieve the refresh token.
Therefore, there is only one long session.
The user can STILL see the session in the settings page even if they are not actively using the application.

On the other hand, if the user logs out, the refresh token is forgotten.
Biometric authentication no longer works.

[This approach also does NOT support biometric authentication with passwordless login or OAuth.](https://community.auth0.com/t/biometrics-with-sso/41969/6)

### HSBC mobile application

HSBC is a proprietary banking application so we can only guess its implementation from an end user's point of view.

Apparently, each user can setup an additional password called Mobile Security Key.
The Mobile Security Key is stored in the keychain on the device.
Biometric authentication is required to retrieve the Mobile Security Key.

The HSBC app offers a logout button in the app.
If the user logs out, the current session is terminated, but biometric authentication is still available.
It is suspected the HSBC app remembers the last login user.
Next time the app is launched, the user can just use biometric authentication to login.

It makes sense for a banking application to keep session lifetime short.

Once a user has logged in the HSBC app, the app remembers the user forever.
It is impossible to switch user unless the app is uninstalled and installed again.
The uninstallation wipes the keychain.

## Design Overview

Storing the refresh token implies long session lifetime.
Storing the password poses a hard restriction on the authenticator the user can use.
In order to have short session lifetime, as well as support for passwordless login, OAuth or any other login means,
we must store something unique.

Borrowing the design of anonymous user, a keypair is stored in the keychain.
Biometric authentication is required to retrieve the private key.
The keypair is used to both identify and authenticate the user.

## Security Concerns

The common user experience of biometric authentication is by scanning fingerprint or face to
log in an account which was signed in previously with non-biometric means.

This means enabling biometric authentication binds biometric authentication to the application authentication.
If the application is not a single-user application like the HSBC app,
then we must ensure when User B logs in, User B can never access the keypair of User A.

The SDK stores the keypair in the keychain of the device per container, requiring biometric authentication to access.
Under normal condition, when biometric authentication is enabled, the developer should let the user to use biometric authentication.
Every time non-biometric authentication is used, the previous keypair will be deleted from the keychain.
The user has to set up biometric authentication again.

## Configuration

Biometric identity is not shown along with Login ID identity and OAuth identity.
Biometric identity is shown separately in the settings page.
By default, biometric identity is hidden.

```yaml
identity:
  biometric:
    list_enabled: false
```

## Errors

### Invalid Account Status

If a user with one of the following status:

- Disabled
- Deactivated
- Scheduled deletion by admin
- Scheduled deletion by end-user
- Scheduled anonymization by admin

is trying to use Biometric login, the login attempt will be rejected, and an error will be returned by the token endpoint.

The error format follows [rfc6749 section 5.2](https://www.rfc-editor.org/rfc/rfc6749.html#section-5.2), with `error=invalid_account_status`. The following is an example of error when a disbled user is trying to use biometric login:

```
error=invalid_account_status
error_description=user is disabled
```

The SDK should simply throw an `OAuthError` containing the same information as the error returned from the token endpoint, and stop the login process. The biometric key should be kept.

The error object thrown by the SDK should looks like:

```swift
OAuthError(
  error="invalid_account_status",
  error_description="user is disabled"
)
```
