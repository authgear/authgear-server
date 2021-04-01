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
 * If refresh token is used, then the refresh token is deleted.
 * If cookie is used, then the SDK redirects to the end session endpoint.
 *
 */
function logout(): Promise<void>;
```

## Biometric authentication

Biometric authentication is provided by the following API

```typescript
/**
 * This function can be called at anytime.
 *
 * If this function does not throw, then biometric authentication is supported on the current device.
 *
 * Otherwise the developer should inspect the thrown error to inform the user what is missing to enable biometric authentication.
 */
function isBiometricSupported(): Promise<void>;

/**
 * This function can be called at anytime.
 * It tells whether the container has a keypair stored in the keychain, thus biometric authentication is possible.
 */
function isBiometricEnabled(): Promise<boolean>;

/**
 * This function can be called when the user has logged in.
 * This function WILL trigger biometric authentication.
 *
 * 1. An UUID is generated as kid.
 * 2. A key pair is generated and stored in the keychain.
 * 3. A challenge is requested from the server.
 * 4. A Biometric identity is added to the user.
 */
function enableBiometric(): Promise<void>;

/**
 * This function can be called at anytime.
 *
 * The following steps are performed:
 *
 * 1. The key pair is deleted from the keychain.
 */
function disableBiometric(): Promise<void>;

/**
 * This function can be called at anytime.
 * This function WILL trigger biometric authentication.
 *
 * 1. The key pair is retrieved from the keychain.
 * 2. A challenge is requested from the server.
 * 3. Call /oauth2/token to request refresh token and access token.
 *
 * If the biometric identity is detected to be removed from the server, the keypair is removed.

 * If the biometric info is detected to be changed, then the keypair is forgotten.
 * This is the case when the user has removed, added, or changed Face ID / Touch ID on iOS.
 */
function authenticateBiometric(): Promise<AuthorizeResult>;
```

### Intended usage

```typescript
// Authenticate the user via non-biometric means.
await authgear.authorize({ ... });



// Check if the device supports biometric authentication.
try {
  await authgear.isBiometricSupported();
} catch (e) {
 // Handle the error to properly.
 return;
}
try {
  await promptIfTheUserWantsToEnableBiometric();
} catch (e) {
  // The user declined.
  return;
}
await authgear.enableBiometric();



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



// Authenticate the user via non-biometric means.
await authgear.authorize({ ... });
// biometric will be disabled after successful non-biometric authentication.
const enabled = await authgear.isBiometricEnabled();
assert(enabled === false);
```
