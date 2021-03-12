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
 * This function WILL trigger biometric authentication if necessary.
 *
 * This function will perform cleanup so that logging in a different user is possible.
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
function enabledBiometric(): Promise<void>;

/**
 * This function can be called when the user has logged in.
 * This function WILL trigger biometric authentication if necessary.
 *
 * 1. The key pair is deleted from the keychain.
 * 2. If the key pair is not found, save the error as e.
 * 3. Call /api/identity/biometric/remove
 * 4. If the identity is not found, throw the error immediately.
 * 5. If identity was removed and e is non-null, set e to null.
 * 6. Throw e if it is non-null.
 *
 * The key pair is stored with kSecAttrAccessibleWhenPasscodeSetThisDeviceOnly and kSecAccessControlBiometryAny,
 * therefore the data is not included in backup, nor transfered to new device,
 * biometric authentication or password code is prompted if necessary.
 */
function disableBiometric(): Promise<void>;

/**
 * This function can be called at anytime.
 * This function WILL trigger biometric authentication if necessary.
 *
 * 1. The key pair is retrieved from the keychain.
 * 2. A challenge is requested from the server.
 * 3. Call /oauth2/token to request refresh token and access token.
 * 4. If the identity is not found, do the following steps.
 *   4.1. Remove the kid from the list of keys so that listBiometricKeys no longer includes that kid.
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
  await authgear.enabledBiometric();
}


// When the session has expired.
const userID = await authgear.getLastUserID();
const enabled = await authgear.isBiometricEnabled();

if (userID != null && enabled) {
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
// The meaning of "logout" varies between applications.
// In app like HSBC, logout means terminating the current session but
// logging in another account is NOT allowed.
//
// If your application interprets logout as allowing logging into different account,
// then your should prompt to the user to disable biometric first.
//
// If you fail to do so, logout() will call disableBiometric() for you.
// However, this behavior may surprise the user.
// Disabling biometric authentication requires biometric authentication
// because access to the keypair (in this case, deletion) requires biometric authentication.

const onClickLogout = () => {
  const userID = await authgear.getLastUserID();
  const enabled = await authgear.isBiometricEnabled();
  if (userID != null && enabled) {
    // Prompt the user we have to disable biometric first.
    // So they will not be scared by the biometric authentication UI.
  }
};
```
