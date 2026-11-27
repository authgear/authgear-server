- [SDK Authflow API](#sdk-authflow-api)
  * [Methods](#methods)
  * [Options](#options)
  * [AuthflowState](#authflowstate)
    + [Branching](#branching)
    + [Errors](#errors)
    + [Finishing](#finishing)
    + [Saving and restoring](#saving-and-restoring)
  * [Steps the app handles](#steps-the-app-handles)
  * [Use cases](#use-cases)
    + [UC1. Sign in with an email address and a password, then a TOTP](#uc1-sign-in-with-an-email-address-and-a-password-then-a-totp)
    + [UC2. Sign in or sign up with an email OTP](#uc2-sign-in-or-sign-up-with-an-email-otp)
    + [UC3. Sign in with Google](#uc3-sign-in-with-google)
    + [UC4. Sign in with an email address and a passkey](#uc4-sign-in-with-an-email-address-and-a-passkey)
    + [UC5. Going back](#uc5-going-back)
    + [UC6. Restoring after the app was ended](#uc6-restoring-after-the-app-was-ended)
    + [UC7. Reauthenticate before a sensitive operation](#uc7-reauthenticate-before-a-sensitive-operation)
    + [UC8. Promote an anonymous user](#uc8-promote-an-anonymous-user)
    + [UC9. Forgot password](#uc9-forgot-password)
    + [UC10. Reset password from a link](#uc10-reset-password-from-a-link)
    + [UC11. Offer the passkey on this device](#uc11-offer-the-passkey-on-this-device)
  * [Limitations](#limitations)
  * [Future works](#future-works)

# SDK Authflow API

The Authflow API of the SDK lets an app sign the user in with its own native UI, without a browser. The app renders each step of an [Authentication Flow](./authentication-flow.md) and passes the user's input. The SDK calls the [Authentication Flow API](./authentication-flow-api-reference.md), and when the app finishes the flow, obtains tokens as the browser-based methods in [SDK](./sdk.md#authentication) do.

Except for account recovery, which issues no tokens, the SDK sends the authorization request to the [Pushed Authorization Request endpoint](./oidc.md#pushed-authorization-request), then creates the flow [with the returned `request_uri`](./authentication-flow-api-reference.md#create-with-an-authorization-request). So the client must be one that endpoint accepts. The SDK creates an account recovery flow with [`url_query`](./authentication-flow-api-reference.md#create-an-authentication-flow) carrying `client_id` and `ui_locales`.

## Methods

`authgear.authflow` has a method to start each flow type, and a method to finish each flow type that issues tokens. Method names are the flow types, in the casing of each language.

| Flow type | Start | Finish | Browser-based counterpart |
|---|---|---|---|
| `login` | `login(options)` | `finishLogin(state)` | `authenticate()` with `page: login` |
| `signup` | `signup(options)` | `finishSignup(state)` | `authenticate()` with `page: signup` |
| `signup_login` | `signupLogin(options)` | `finishSignupLogin(state)` | `authenticate()` |
| `reauth` | `reauth(options)` | `finishReauth(state)` | `reauthenticate()` |
| `promote` | `promote(options)` | `finishPromote(state)` | `promoteAnonymousUser()` |
| `account_recovery` | `accountRecovery(options)` | none | none |

- `reauth` fails unless the user is signed in and the SDK has an ID token. Call `refreshIDToken()` first.
- `promote` fails unless an anonymous user is signed in on this device.
- `accountRecoveryWithLink(link)` continues the account recovery flow that sent a link, and returns the state after the link's code is verified. `link` is the URL the email or SMS links to. The SDK reads the code from its `code` query parameter, and fails when there is none.

```typescript
interface AuthflowOptions {
  name?: string;
  uiLocales?: string[];
}

interface AuthflowAuthorizationOptions extends AuthflowOptions {
  resource?: string;
  scopes?: string[];
}

function login(options?: AuthflowAuthorizationOptions): Promise<AuthflowState>;
function signup(options?: AuthflowAuthorizationOptions): Promise<AuthflowState>;
function signupLogin(options?: AuthflowAuthorizationOptions): Promise<AuthflowState>;
function reauth(options?: AuthflowAuthorizationOptions): Promise<AuthflowState>;
function promote(options?: AuthflowAuthorizationOptions): Promise<AuthflowState>;
function accountRecovery(options?: AuthflowOptions): Promise<AuthflowState>;
function accountRecoveryWithLink(link: string): Promise<AuthflowState>;

function finishLogin(state: AuthflowState): Promise<UserInfo>;
function finishSignup(state: AuthflowState): Promise<UserInfo>;
function finishSignupLogin(state: AuthflowState): Promise<UserInfo>;
function finishReauth(state: AuthflowState): Promise<UserInfo>;
function finishPromote(state: AuthflowState): Promise<UserInfo>;

function restore(serialized: string): AuthflowState;
```

## Options

| Option | Methods | Meaning |
|---|---|---|
| `name` | All | The name of the flow, as in [Create an authentication flow](./authentication-flow-api-reference.md#create-an-authentication-flow). Default `"default"`. |
| `uiLocales` | All | As in `authenticate()`. |
| `resource` | All except `accountRecovery` | As in `authenticate()`. |
| `scopes` | All except `accountRecovery` | As in `authenticate()`. |

## AuthflowState

```typescript
interface AuthflowState {
  readonly stateToken: string;
  readonly type: AuthflowType;
  readonly name: string;
  readonly action: AuthflowAction;
  // True when action.type is "finished".
  readonly isFinished: boolean;

  input(input: object): Promise<AuthflowState>;
  batchInput(inputs: object[]): Promise<AuthflowState>;
  // Retrieves this state again.
  refresh(): Promise<AuthflowState>;
  serialize(): string;
}

interface AuthflowAction {
  type: AuthflowActionType;
  identification?: AuthflowIdentification;
  authentication?: AuthflowAuthentication;
  data: AuthflowData;
  rawData: object;
}
```

The fields and inputs are those of the [Authentication Flow API](./authentication-flow-api-reference.md#reference-on-input-and-output), in the casing of each language.

| Type | Cases |
|---|---|
| `AuthflowType` | One per flow type in [Methods](#methods) |
| `AuthflowActionType` | One per `action.type` in [Reference on input and output](./authentication-flow-api-reference.md#reference-on-input-and-output), and `finished` |
| `AuthflowIdentification` | One per `identification` in [Reference on input and output](./authentication-flow-api-reference.md#reference-on-input-and-output) |
| `AuthflowAuthentication` | One per `authentication` in [Reference on input and output](./authentication-flow-api-reference.md#reference-on-input-and-output) |
| `AuthflowData` | One per data type in [Reference on action data](./authentication-flow-api-reference.md#reference-on-action-data), chosen by `data.type`, with every field of that data typed |

Each type also has an unknown case that keeps the raw value, for a value the server added after the SDK was released. Decoding never fails on an unknown value, so an app keeps working until it updates the SDK. `rawData` is `action.data` as JSON, including fields the SDK does not model yet.

`input` takes a JSON object, as described in the API reference.

`refresh()` is for a state whose data changes without input, such as waiting for a login link to be approved on another device.

Intended usage:

```typescript
let state = await authgear.authflow.signupLogin();
while (!state.isFinished) {
  const input = await showScreenFor(state.action);
  try {
    state = await state.input(input);
  } catch (e) {
    showError(e);
  }
}
const userInfo = await authgear.authflow.finishSignupLogin(state);
```

### Branching

A state never changes. Each `input()` returns a new state, and the state it was called on stays usable. Passing input to an earlier state takes another branch, as in [State and Branching](./authentication-flow-api-reference.md#state-and-branching), so going back is reusing the earlier state.

### Errors

When the server rejects an input, `input()` fails with an error carrying the server's `reason` and `info`. The state it was called on stays usable.

When the flow has expired, `input()`, `batchInput()` and `refresh()` fail. The app starts a new flow.

### Finishing

A finish method exchanges the `code` of a finished state, as specified in [Code from an Authentication Flow](./oidc.md#code-from-an-authentication-flow), and changes the SDK as its browser-based counterpart does:

| Method | Changes in the SDK |
|---|---|
| `finishLogin`, `finishSignup`, `finishSignupLogin` | Saves the new session, replacing any current one. The session state becomes `authenticated`, and the app is notified of the change. Biometric login is disabled |
| `finishPromote` | As `finishLogin`, and the anonymous user's key is removed from the device |
| `finishReauth` | Replaces the ID token, which then has a new `auth_time`. Nothing else changes, and the app is not notified |

Each returns the user's `UserInfo`. It fails when:

- `isFinished` is false.
- The state was not started, or restored, from the matching start method. A `signup_login` flow becomes `login` or `signup` as it runs, so `type` does not decide this.
- The state belongs to another container.
- The state was already finished. Its `code` can be used once.

An account recovery flow has no finish method. It is complete when `isFinished` is true.

### Saving and restoring

`serialize()` returns a string from which `authgear.authflow.restore()` rebuilds the state, for example after the operating system ends the app in the background.

The string contains the PKCE code verifier. Keep it with the app's own saved state, and never log it or send it anywhere.

## Steps the app handles

| Step | What the app does |
|---|---|
| [`identification: oauth`](./authentication-flow-api-reference.md#identification-oauth) | Opens `oauth_authorization_url` in a browser session, receives the callback at the `redirect_uri` it passed, and passes the callback's query as input |
| [Bot protection](./authentication-flow-api-reference.md#bot-protection) | Obtains the token from the provider's SDK and passes it as input |
| [`identification: passkey`](./authentication-flow-api-reference.md#identification-passkey), [`authentication: primary_passkey`](./authentication-flow-api-reference.md#authentication-primary_passkey), [`prompt_create_passkey`](./authentication-flow-api-reference.md#type-signup-actiontype-prompt_create_passkey) | Passes `request_options` or `creation_options` to the platform passkey API, and passes the result as `assertion_response` or `creation_response`. The app must be listed in [`x_native_apps`](./passkey-webview.md#configuration), and `http.public_origin` must be on a custom domain, as in [Association files](./passkey-webview.md#association-files) |

## Use cases

An app that renders its own UI is built for its project's flows, so it knows the steps of each flow. Each example assumes a flow with the steps it shows, and passes inputs as in the [API reference](./authentication-flow-api-reference.md#reference-on-input-and-output).

The examples are React Native screens. Each shows what the screen displays from `state.action.data`, and what it does when the user presses a button. Error handling is left out: when `input()` fails, the screen shows the error and keeps its state, so the user can try again.

### UC1. Sign in with an email address and a password, then a TOTP

```tsx
// Home screen
<Button title="Sign in" onPress={onSignIn} />

const onSignIn = async () => {
  const state = await authgear.authflow.login();
  navigation.navigate("Email", { state });
};

// Email screen: identification_data
<TextInput value={email} onChangeText={setEmail} placeholder="Email" />
<Button title="Continue" onPress={onContinue} />

const onContinue = async () => {
  const nextState = await state.input({ identification: "email", login_id: email });
  navigation.navigate("Password", { state: nextState });
};

// Password screen: authentication_data
<TextInput value={password} onChangeText={setPassword} secureTextEntry placeholder="Password" />
<Button title="Sign in" onPress={onSignIn} />

const onSignIn = async () => {
  const nextState = await state.input({ authentication: "primary_password", password });
  navigation.navigate("TOTP", { state: nextState });
};

// TOTP screen: authentication_data, of the secondary authenticators
<TextInput value={code} onChangeText={setCode} placeholder="Code from your authenticator app" />
<Button title="Verify" onPress={onVerify} />

const onVerify = async () => {
  const finishedState = await state.input({ authentication: "secondary_totp", code });
  const userInfo = await authgear.authflow.finishLogin(finishedState);
  navigation.popToTop();
};
```

### UC2. Sign in or sign up with an email OTP

The flow is started with `signupLogin()`. The email screen also chooses the OTP option, by its index.

```tsx
// Email screen: identification_data, then authentication_data
const onContinue = async () => {
  const identifiedState = await state.input({ identification: "email", login_id: email });
  const { options } = identifiedState.action.data as AuthenticationData;
  const index = options.findIndex((o) => o.authentication === "primary_oob_otp_email");
  const nextState = await identifiedState.input({ authentication: "primary_oob_otp_email", index });
  navigation.navigate("OTP", { state: nextState });
};

// OTP screen: verify_oob_otp_data
const data = state.action.data as VerifyOOBOTPData;

<Text>Enter the {data.codeLength}-digit code sent to {data.maskedClaimValue}</Text>
<TextInput value={code} onChangeText={setCode} maxLength={data.codeLength} />
<Button title="Verify" onPress={onVerify} />
<Button title="Resend" onPress={onResend} disabled={now < data.canResendAt} />

const onResend = async () => {
  const nextState = await state.input({ resend: true });
  // The screen shows the new state, with a new canResendAt.
  navigation.setParams({ state: nextState });
};

const onVerify = async () => {
  const finishedState = await state.input({ code });
  const userInfo = await authgear.authflow.finishSignupLogin(finishedState);
  navigation.popToTop();
};
```

### UC3. Sign in with Google

```tsx
// Sign-in screen: identification_data
<Button title="Continue with Google" onPress={onSignIn} />

const onSignIn = async () => {
  const redirectURI = "com.example.app://oauth-callback";
  const nextState = await state.input({ identification: "oauth", alias: "google", redirect_uri: redirectURI });
  // oauth_data
  const { oauthAuthorizationURL } = nextState.action.data as OAuthData;
  // openAuthSession opens oauthAuthorizationURL in a browser, and returns the URL redirected to.
  const callbackURL = await openAuthSession(oauthAuthorizationURL, redirectURI);
  const finishedState = await nextState.input({ query: new URL(callbackURL).search.slice(1) });
  const userInfo = await authgear.authflow.finishLogin(finishedState);
  navigation.popToTop();
};
```

The `redirect_uri` is registered with the provider, as in [Any other OAuth providers](./authentication-flow.md#any-other-oauth-providers-supported-by-authgear).

### UC4. Sign in with an email address and a passkey

The email screen is as in UC1, and navigates to this screen.

```tsx
// Passkey screen: authentication_data
<Button title="Sign in with a passkey" onPress={onSignIn} />

const onSignIn = async () => {
  const { options } = state.action.data as AuthenticationData;
  const option = options.find((o) => o.authentication === "primary_passkey");
  // getPasskey shows the system passkey sheet with requestOptions, and returns the assertion as WebAuthn JSON.
  const assertion = await getPasskey(option.requestOptions);
  const finishedState = await state.input({ authentication: "primary_passkey", assertion_response: assertion });
  const userInfo = await authgear.authflow.finishLogin(finishedState);
  navigation.popToTop();
};
```

### UC5. Going back

In UC1, the back button returns from the password screen to the email screen. The email screen still has its own state, so pressing Continue again passes the new input to that state, as in [Branching](#branching).

### UC6. Restoring after the app was ended

Each screen keeps its state in its route params as `serialize()` output, and restores it with `authgear.authflow.restore()`. The app then persists its navigation state as usual, and the flow is restored with its screens.

```tsx
navigation.navigate("Password", { state: nextState.serialize() });

// Password screen
const state = authgear.authflow.restore(route.params.state);
```

The persisted navigation state contains the PKCE code verifier, as in [Saving and restoring](#saving-and-restoring). When the flow has expired meanwhile, the next `input()` fails.

### UC7. Reauthenticate before a sensitive operation

```tsx
// Settings screen
<Button title="Delete account" onPress={onDelete} />

const onDelete = async () => {
  await authgear.refreshIDToken();
  if (!authgear.canReauthenticate()) {
    // The user has nothing to reauthenticate with. As with reauthenticate(), the server decides.
    await deleteAccount(authgear.getIDTokenHint());
    return;
  }
  const state = await authgear.authflow.reauth();
  navigation.navigate("Reauth", { state });
};

// Reauth screen: authentication_data. The user is identified by the ID token.
<TextInput value={password} onChangeText={setPassword} secureTextEntry placeholder="Password" />
<Button title="Confirm" onPress={onConfirm} />

const onConfirm = async () => {
  const finishedState = await state.input({ authentication: "primary_password", password });
  await authgear.authflow.finishReauth(finishedState);
  await deleteAccount(authgear.getIDTokenHint());
};
```

### UC8. Promote an anonymous user

The flow is started with `promote()`. The email and OTP screens are as in UC1 and UC2, and the OTP screen navigates to this screen instead of finishing.

```tsx
// New password screen: create_authenticator_data
const { options } = state.action.data as CreateAuthenticatorData;
const { passwordPolicy } = options.find((o) => o.authentication === "primary_password");

<TextInput value={password} onChangeText={setPassword} secureTextEntry placeholder="New password" />
{/* For example, "At least 8 characters" from passwordPolicy.minimumLength. */}
<PasswordRequirements policy={passwordPolicy} />
<Button title="Save" onPress={onSave} />

const onSave = async () => {
  const finishedState = await state.input({ authentication: "primary_password", new_password: password });
  // The same user, no longer anonymous.
  const userInfo = await authgear.authflow.finishPromote(finishedState);
  navigation.popToTop();
};
```

### UC9. Forgot password

The user starts from the password screen of UC1, which already has the email address.

```tsx
// Password screen of UC1
<Button title="Forgot password?" onPress={onForgotPassword} />

const onForgotPassword = async () => {
  const recoveryState = await authgear.authflow.accountRecovery();
  // account_recovery_identification_data. The email address the user entered in UC1.
  const nextState = await recoveryState.input({ identification: "email", login_id: email });
  navigation.navigate("Destination", { state: nextState });
};

// Destination screen: account_recovery_select_destination_data
const { options } = state.action.data as AccountRecoverySelectDestinationData;

{options.map((o, index) => (
  <Button key={index} title={`Send a code to ${o.maskedDisplayName}`} onPress={() => onChoose(index)} />
))}

const onChoose = async (index: number) => {
  const nextState = await state.input({ index });
  navigation.navigate("RecoveryCode", { state: nextState });
};

// Recovery code screen: account_recovery_verify_code_data
const data = state.action.data as AccountRecoveryVerifyCodeData;

<Text>Enter the {data.codeLength}-digit code sent to {data.maskedDisplayName}</Text>
<TextInput value={code} onChangeText={setCode} maxLength={data.codeLength} />
<Button title="Continue" onPress={onContinue} />

const onContinue = async () => {
  const nextState = await state.input({ account_recovery_code: code });
  navigation.navigate("ResetPassword", { state: nextState });
};

// Reset password screen: new_password_data
const { passwordPolicy } = state.action.data as NewPasswordData;

<TextInput value={password} onChangeText={setPassword} secureTextEntry placeholder="New password" />
<PasswordRequirements policy={passwordPolicy} />
<Button title="Reset password" onPress={onReset} />

const onReset = async () => {
  await state.input({ new_password: password });
  // The flow is finished. The password screen still has its login state,
  // so the user signs in there with the new password.
  navigation.navigate("Password");
};
```

### UC10. Reset password from a link

The destination the user chose in UC9 sends a link instead of a code. The app handles the link as a universal link or App Link.

The app opens a screen with the link, whether the link starts the app or reaches it while running, for example with React Navigation's linking configuration.

```tsx
// Recovery link screen, opened with the link as route.params.link
<ActivityIndicator />

useEffect(() => {
  (async () => {
    const state = await authgear.authflow.accountRecoveryWithLink(route.params.link);
    // new_password_data. The reset password screen of UC9.
    navigation.replace("ResetPassword", { state });
  })();
}, []);
```

When the link has expired or was already used, the screen shows the error.

### UC11. Offer the passkey on this device

When the email screen of UC1 appears, it asks for a passkey that is already on this device. If there is one, the system passkey sheet appears without a tap, and choosing it signs the user in. The passkey identifies and authenticates the user, so the flow finishes after it. If there is none, nothing appears, and the user continues with an email address.

```tsx
// Email screen of UC1: identification_data
useEffect(() => {
  (async () => {
    const { options } = state.action.data as IdentificationData;
    const option = options.find((o) => o.identification === "passkey");
    try {
      // getPasskey is as in UC4. onlyOnThisDevice shows the sheet only when this device has a passkey.
      const assertion = await getPasskey(option.requestOptions, { onlyOnThisDevice: true });
      const finishedState = await state.input({ identification: "passkey", assertion_response: assertion });
      const userInfo = await authgear.authflow.finishLogin(finishedState);
      navigation.popToTop();
    } catch (e) {
      if (isNoPasskeyOrCancelled(e)) {
        // No passkey on this device, or the user closed the sheet. The email screen stays.
        return;
      }
      setError(e);
    }
  })();
}, []);
```

The screen asks once. Coming back to it, as in UC5, does not ask again.

## Limitations

- **No SSO.** The flow creates no IdP session, so it shares no session with sign-ins in a browser.
- **No "remember this device".** The device token is set as a cookie, which the SDK does not keep. An app should not offer the option. Passing it has no effect on later sign-ins.
- **Account recovery by link** completes in the app only when the app handles the link as a universal link or App Link, as in [UC10](#uc10-reset-password-from-a-link). Otherwise it completes in the browser that opens the link.

## Future works

- Typed inputs.
- Methods that perform the platform part of a step and then pass the input: OAuth providers and passkeys.
- "Remember this device".
