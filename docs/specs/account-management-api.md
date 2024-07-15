# Account Management API

Account Management API is an API for account management on behalf of the end user. Account Management API requires a authenticated session, either by cookie or the HTTP header Authorization.

Account Management API supports the following operations:

- [Manage identifications](#manage-identifications)
  - List identifications
  - Add a new email, with verification
  - Add a new phone number, with verification
  - Add a new username
  - [Start adding an OAuth provider account, with authorization code flow](#start-adding-an-oauth-provider-account-with-authorization-code-flow)
  - [Finish adding an OAuth provider account, with authorization code flow](#finish-adding-an-oauth-provider-account-with-authorization-code-flow)
  - Add a biometric
  - Add a passkey
  - Remove an email
  - Remove a phone number
  - Remove a username
  - Remove an OAuth provider account
  - Remove a biometric
  - Remove a passkey
  - Update an email, with verification
  - Update a phone number, with verification
  - Update a username
- Manage authentications
  - List authentications, and recovery codes.
  - Change the primary password
  - Change the secondary password
  - Remove the secondary password
  - Add a new TOTP authenticator
  - Remove a TOTP authenticator
  - Add a new OOB-OTP authenticator, with verification.
  - Remove a OOB-OTP authenticator
  - Re-generate recovery codes
- Manage user profile
  - Update simple standard attributes
  - Update custom attributes
  - Add user profile picture
  - Replace user profile picture
  - Remove user profile picture
- Manage sessions
  - List sessions
  - Revoke a session
  - Terminate all other sessions
- Auxiliary operations
  - Verify OTP
  - Resend OTP

## Manage identifications

### Start adding an OAuth provider account, with authorization code flow

> This endpoint is considered as advanced. It is assumed that you are capable of
> receiving OAuth callback via `redirect_uri`.

`POST /api/v1/account/identification`

Request

```json
{
  "identification": "oauth",
  "alias": "google",
  "redirect_uri": "https://myapp.authgear.cloud/sso/oauth2/callback/google",
  "exclude_state_in_authorization_url": true
}
```

- `identification`: Required. It must be the value `oauth`.
- `alias`: Required. The alias of the OAuth provider you want the current account to associate with.
- `redirect_uri`: Required. You have to specify your own redirect URI to your app or your website to receive the OAuth callback.
- `exclude_state_in_authorization_url`: Optional. The default is false.
  - When it is false, the `authorization_url` has a `state` parameter included, the `token` is bound to this `state` parameter.
  - When is is true, the `authorization_url` has no `state` parameter included, the `token` is NOT bound to `state`.
  - If you wish to use your own state, you must specify `true` for this field.

Response

```json
{
  "result": {
    "token": "oauthtoken_blahblahblah",
    "authorization_url": "https://www.google.com?client_id=client_id&redirect_uri=redirect_uri"
  }
}
```

- `token`: You store this token. You need to supply it after the end-user returns to your app.
- `authorization_url`: You MUST redirect the end-user to this URL to continue the authorization code flow. If `exclude_state_in_authorization_url` is false, it has `state` parameter included.

Error response

|Description|Name|Reason|Info|
|---|---|---|---|
|If the request is not authenticated|Unauthorized|Unauthorized||

The OAuth provider ultimately will call your redirect URI with query parameters added. You continue the flow with [Finish adding an OAuth provider account, with authorization code flow](#finish-adding-an-oauth-provider-account-with-authorization-code-flow).

Here is some pseudo code that you should do

```javascript
const response = fetch("https://myapp.authgear.cloud/api/v1/account/identification", {
  method: "POST",
  headers: {
    "Authorization": `Bearer ${ACCESS_TOKEN}`,
    "Content-Type": "application/json",
  },
  body: JSON.stringify({
    "identification": "oauth",
    "alias": "google",
    "redirect_uri": "com.myapp://host/path",
    "exclude_state_in_authorization_url": true,
  }),
});
const responseJSON = await response.json();
// TODO: Add proper error handling here.
// You cannot assume you always get result.
const token = responseJSON.result.token;
const authorizationURL = new URL(responseJSON.result.authorization_url);
// Generate a random state that is NOT too long. The characters MUST be URL safe.
const state = generateAlphaNumbericRandomStringOfLength(32);
authorizationURL.searchParams.set("state", state);
// Store the state in some persistent storage
window.sessionStorage.setItem("resume", JSON.stringify({
    state,
    token,
}));
// Redirect to the OAuth provider.
window.location.href = authorizationURL.toString();
```

### Finish adding an OAuth provider account, with authorization code flow

`POST /api/v1/account/identification/oauth`

Request

```json
{
  "token": "oauthtoken_blahblahblah",
  "query": "code=code"
}
```

- `token`: The `token` you received in the response of [Start adding an OAuth provider account, with authorization code flow](#start-adding-an-oauth-provider-account-with-authorization-code-flow).
- `query`: The query of the redirect URI.

Response

If successful, then the OAuth provider account is added.

```json
{
  "result": {}
}
```

Error response

|Description|Name|Reason|Info|
|---|---|---|---|
|If the request is not authenticated|Unauthorized|Unauthorized||
|If the OAuth provider account is already taken by another account|Invalid|InvariantViolated|`{"cause": { "kind": "DuplicatedIdentity" } }`|
|If `token` is invalid|Invalid|AccountManagementOAuthTokenInvalid||
|If `exclude_state_in_authorization_url` is false, and `state` in `query` is not equal to the one bound to `token`|Invalid|AccountManagementOAuthStateNotBoundToToken||

Here is some pseudo code that you should do

```javascript
// The OAuth provider redirects the user back to us.
const url = new URL(window.location.href);
const state = url.searchParams.get("state");
if (state == null) {
  // Expected state to be present.
  return;
}
const resumeStr = window.sessionStorage.getItem("resume");
if (resumeStr == null) {
  // Expected resume to be non-null.
  return;
}
// Always remove resume.
window.sessionStorage.removeItem("resume");

const resume = JSON.parse(resumeStr);
if (state !== resume.state) {
  // Expected resume.state equals to state.
  return;
}
const token = resume.token;
const response = fetch("https://myapp.authgear.cloud/api/v1/account/identification/oauth", {
  method: "POST",
  headers: {
    "Authorization": `Bearer ${ACCESS_TOKEN}`,
    "Content-Type": "application/json",
  },
  body: JSON.stringify({
    "token": token,
    "query": url.search,
  }),
});
const responseJSON = await response.json();
// TODO: Add proper error handling here.
```
