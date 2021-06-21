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

## Authentication

Authentication is provided by the following API

```typescript
interface AuthorizeOptions {
  // Required standard parameters.
  redirectURI: string;

  // Optional standard parameters.
  prompt?: "login";
  state?: string;
  uiLocales?: string[];

  // Authgear specific parameters.
  page?: "login" | "signup";
}

interface ReauthenticateOptions {
  // Required standard parameters.
  redirectURI: string;

  // Optional standard parameters.
  state?: string;
  uiLocales?: string[];

  // The default value is 0.
  maxAge?: number;

  // If this is true, then biometric is not used for reauthentication.
  skipUsingBiometric?: boolean;
}

function authorize(options: AuthorizeOptions): Promise<{ userInfo: UserInfo, state?: string }>;

// Reauthenticate the current user.
// If biometric is enabled, biometric is used to reauthenticate the user, unless skipUsingBiometric is true.
// Otherwise, it behaves like authorize().
function reauthenticate(options: ReauthenticateOptions): Promise<{ userInfo: UserInfo, state?: string }>;

// canReauthenticate returns true if the user has logged in and they have authenticator to reauthenticate themselves.
// Note that there are users who CANNOT reauthenticate themselves.
// For example, a user who signed up with Google does not have any authenticator thus cannot reauthenticate themselves.
// Before you want to trigger reauthentication, you must first call this function to see if reauthentication is possible.
// Otherwise, your user see will an error message if you trigger reauthentication.
function canReauthenticate(): boolean;

// Always authenticate the end-user fully.
// This is the suitable user experience when the user taps "Sign in" or "Sign up"
const options: AuthorizeOptions = {
  redirectURI: "myapp://host/path",
  prompt: "login",
};

// Authenticate the end-user, allow reusing existing session.
const options: AuthorizeOptions = {
  redirectURI: "myapp://host/path",
};

// Always reauthenticate the current user.
// You MUST first decode the ID token and check if the claim `https://authgear.com/user/can_reauthenticate` is true,
// otherwise your user WILL see an error.
const options: AuthorizeOptions = {
  redirectURI: "myapp://host/path",
  prompt: "login",
  maxAge: 0,
};

// Reauthenticate the current user if last authentication is not within 5 minutes.
// You MUST first decode the ID token and check if the claim `https://authgear.com/user/can_reauthenticate` is true,
// otherwise your user WILL see an error.
const options: AuthorizeOptions = {
  redirectURI: "myapp://host/path",
  prompt: "login",
  maxAge: 60 * 5,
};

// After reauthentication, get the ID token from the SDK and send it along with your sensitive request.
// Your server is responsible for verifying the signature of the ID token and validating the `auth_time`.
authgear.fetch("https://api.myserver.com/v1/edit-profile", {
  method: "POST",
  headers: {
    "content-type": "application/json",
  },
  body: JSON.stringify({
    profile: { ... },
    id_token: authgear.getIDTokenHint(),
  }}),
});

// Example of how to trigger reauthentication before performing sensitive operation.
async function onClickSave() {
  const canReauthenticate = authgear.canReauthenticate();

  // The user cannot be reauthenticated.
  // However, we do not want to block them from performing sensitive operation.
  // So we still allow them to proceed.
  if (!canReauthenticate) {
    // You must send the ID token to your server.
    // Your server must verify the signature and check the claims inside.
    // Your server should allow user to skip reauthentication if it is impossible.
    await performSensitiveOperation(authgear.getIDTokenHint());
    return;
  }

  const options: ReauthenticateOptions = {
    redirectURI: "myapp://host/path",
    maxAge: 0,
  };

  await authgear.reauthenticate(options);

  // If we reach here, the ID token is newly issued.
  // Send it along with the request to the server
  await performSensitiveOperation(authgear.getIDTokenHint());
  return;
}
```

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
