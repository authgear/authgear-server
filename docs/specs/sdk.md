# SDK

This document specifies the user experience of the SDK.

## Container, container name and storage

An instance of Container is the entrypoint of the SDK.
Unless otherwise specified, `authgear` is an instance of Container.
`authgear.foobar()` means calling the `foobar()` method of the container.

Every container has a name.
If no name is given, the name is `"default"`.

Everything stored by a container is scoped by its name.
Therefore, if the developer wants to implement multiple account,
they must create different container instances with distinct names.

## The use of persistent storage

The SDK by default will store refresh tokens in persistent storage.
In some applications such as banking applications,
It is more preferable to have ephemeral refresh token that are only stored in memory.

The SDK can be configured to store refresh token in memory only.

## Logout

Logout is provided by the following API

```typescript
/**
 * This function can be called when the user has logged in.
 *
 * Depending on whether biometric authentication is enabled or not,
 * this function behaves differently.
 *
 * When biometric authentication is disabled,
 * calling this function will allow logging in as a different user.
 *
 * When biometric authentication is enabled,
 * calling this function will ONLY terminate the current session.
 * The last logged in user will NOT be forgotten.
 * Logging in a different user will be disallowed in the web UI.
 * If switching different user is desired,
 * disableBiometric() must be called.
 *
 */
function logout(): Promise<void>;
```

## Biometric authentication

Biometric authentication is provided by the following API

```typescript
/**
 * This function can be called at anytime.
 * This function reports whether biometric authentication is available.
 */
function isBiometricAvailable(): Promise<boolean>;

/**
 * This function can be called at anytime.
 * This function returns the user ID of last logged in user.
 * The data is stored with kSecAttrAccessibleWhenUnlockedThisDeviceOnly,
 * therefore the data is not included in backup, nor transfered to new device.
 */
function getLastUserID(): Promise<string | null>;

/**
 * This function can be called at anytime.
 * This function tells if the last logged in user has enabled biometric authentication.
 * The data is stored with kSecAttrAccessibleWhenUnlockedThisDeviceOnly,
 * therefore the data is not included in backup, nor transfered to new device.
 */
function isBiometricEnabled(): Promise<boolean>;

/**
 * This function can be called when the user has logged in.
 * This function WILL trigger biometric authentication if necessary.
 *
 * 1. An UUID is generated as kid.
 * 2. A key pair is generated and stored in the keychain.
 * 3. A challenge is requested from the server.
 * 4. A Biometric identity is added to the user.
 *
 * The key pair is stored with kSecAttrAccessibleWhenPasscodeSetThisDeviceOnly and kSecAccessControlBiometryAny,
 * therefore the data is not included in backup, nor transfered to new device,
 * biometric authentication or password code is prompted if necessary.
 */
function enableBiometric(): Promise<void>;

/**
 * This function can be called when the user has logged in, or at anytime if an option is specified.
 * This function WILL trigger biometric authentication if necessary.
 *
 * The following steps are performed:
 *
 * 1. Call /api/identity/biometric/remove
 * 2. If the error is identity not found, the error is always ignored.
 * 3. If `force` is specified, the error is ignored, otherwise the error is thrown.
 * 4. The key pair is deleted from the keychain.
 *
 * The key pair is stored with kSecAttrAccessibleWhenPasscodeSetThisDeviceOnly and kSecAccessControlBiometryAny,
 * therefore the data is not included in backup, nor transfered to new device,
 * biometric authentication or password code is prompted if necessary.
 */
function disableBiometric(options?: { force?: boolean }): Promise<void>;

/**
 * This function can be called at anytime.
 * This function WILL trigger biometric authentication if necessary.
 *
 * 1. The key pair is retrieved from the keychain.
 * 2. A challenge is requested from the server.
 * 3. Call /oauth2/token to request refresh token and access token.
 * 4. If the identity is not found, do the following steps.
 *   4.1. Remove the biometric user ID so that isBiometricEnabled() returns false.
 */
function authenticateBiometric(): Promise<AuthorizeResult>;
```

### Intended usage

```typescript
// Authenticate the user via non-biometric means.
await authgear.authorize({ ... });


// Check if the device supports biometric authentication.
const available = await authgear.isBiometricAvailable();
if (available) {
  // Show a screen to ask the user if they want to enable biometric authentication.
  await authgear.enableBiometric();
}


// When the session has expired.
const enabled = await authgear.isBiometricEnabled();

if (enabled) {
  // The user MAY have biometric set up previously.
  // Show a screen to ask if they want to authenticate with biometric.
  if (userHasAgreedToUseBiometric) {
    try {
      await authgear.authenticateBiometric();
    } catch(e) {
      // Inspect the error to see if it is identity not found.
      // Fallback to non-biometric login screen.
      // This is necessary because the biometric identity can be removed from the settings page.
    }
  }
}


// When the user want to "logout".
//
// Logout can mean two things in application with biometric authentication.
//
// 1. Terminate the current session. Retain the biometric authentication. Signing in the same user is allowed.
// 2. Perform a whole cleanup so that switching user is possible.
//
// For case 1, just call logout().
// For case 2, call disableBiometric() before calling logout().

const onClickTerminateSession = () => {
  await authgear.logout();
};

const onClickSwitchAccount = () => {
  const enabled = await authgear.isBiometricEnabled();
  if (enabled) {
    prompt("Switching account requires disabling biometric authentication first. You have to setup again later. Continue?")
      .then(() => {
        await authgear.disableBiometric();
        await authgear.logout();
      });
  }
};
```
