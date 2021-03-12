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

## Biometric authentication

Biometric authentication is provided by the following API

```typescript
interface BiometricKey {
  userID: string;
  kid: string;
  createdAt: Date;
}

/**
 * This function can be called at anytime.
 * This function access the keychain to find out biometric key IDs on this device.
 * The list of the keys is stored with kSecAttrAccessibleWhenUnlockedThisDeviceOnly,
 * therefore the data is not included in backup, nor transfered to new device.
 */
function listBiometricKeys(): Promise<BiometricKey[]>;

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
function addBiometric(): Promise<BiometricKey>;

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
function removeBiometric(kid: string): Promise<void>;

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
function authenticateBiometric(kid: string): Promise<AuthorizeResult>;
```

### Open questions

#### Should we allow only 1 private key per container name?

If yes, consider the following scenario:

1. User A logs in and sets up biometric.
1. User A logs out.
1. User B logs in and sets up biometric. User A's kid is forgotten. Note that the identity is still stored in the server so the identity is still listed in the settings page.
1. User B logs out.
1. User A logs in and they HAVE TO set up biometric AGAIN.

If no, we have to introduce `kid` as argument in some functions.
However, the above scenario would become:

1. User A logs in and sets up biometric.
1. User A logs out.
1. User B logs in and sets up biometric. 2 keys is known to the SDK.
1. User B logs out.
1. It is possible to show a screen to let the user to log in either User A or User B with biometric only. However, in normal condition, the app should remember the last login user via non-biometric means. Biometric screen is only shown when there is such user. It is the application logic to make this happen.

### Intended usage

```typescript
// Authenticate the user via non-biometric means.
await authgear.authorize({ ... });

// Store the last login user via non-biometric means.
myapp.setUserID(userID);


// When the session has expired.
let kid;
const userID = await myapp.getUserID();
if (userID != null) {
  const keys = await authgear.listBiometricKeys();
  for (const key of keys) {
    if (key.userID === userID) {
      // The user MAY have biometric set up previously.
      // Show a screen to ask if they want to authenticate with biometric.
      kid = key.kid;
    }
  }
}

if (userHasAgreedToUseBiometric && kid != null) {
  try {
    await authgear.authenticateBiometric(kid);
  } catch(e) {
    // Inspect the error to see if it is identity not found.
    // Fallback to non-biometric login screen.
    // This is necessary because the biometric identity can be removed from the settings page.
  }
}


// When the user disables biometric.
authgear.removeBiometric(kid);
```
