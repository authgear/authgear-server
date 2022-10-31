# Single Sign On with Authgear

## What is Single-sign-on

- Single-sign-on (SSO) is defined as login once, logged in all apps.
- The end-user is not required to enter their authentication credentials again.
- The end-users expect to logout from all the apps by clicking 1 button too. (logout from this computer)

## Summary

Let’s say oursky.com is a big company with multiple products that live in different URLs, domains and platforms. Their apps can be either mobile, SPA (token-based) or Web-App (cookie-based) depending on the projects. On Authgear, they create an "IdP" project oursky.authgearapps.com

|App Type|Login|Logout|Redirect behavior when invalid auth code|
|--------|-----|------|----------------------------------------|
|Web-App|Session shared among `*.oursky.apps`, no need to go to `/login` again|1. Clear cookie<br/>2. Invalidate all refresh tokens that are associated with the same IdP session.|No effect, the session is still set in cookie|
|SPA - with SSO|If IdP session exists, login with "Continue As…" <br/>If not, login by entering credentials, a IdP session cookie is set in the browser. <br/>The refresh tokens generated from different clients are linked together by the same IdP session.|1. Clear refresh token <br/>2. Clear cookie <br/>3. Invalidate all other refresh tokens that are associated with the same IdP session.|Need to click "Continue As…" again|
|SPA - without SSO|Must enter the credentials to login. <br/>The existing IdP session in cookie is ignored, and login will NOT set a cookie in the browser.|Clear refresh token for that particular SPA.|Enter credentials again|
|Mobile - with SSO (shared with browser)|`shareSessionWithSystemBrowser: true`<br/> IdP cookie is set or continue with existing session from the browser.|1. Clear token in apps. <br/>2. Clear cookie in browser. <br/>3. Invalidate all other refresh tokens that are associated with the same IdP session.|Not Applicable|
|Mobile - without SSO (session only inside the app, not in browser)|Must enter the credentials to login. <br/>IdP session cookie is not read nor set. |Clear refresh token in the app.|Not Applicable|
|OIDC - with SSO|3rd party use-case, working like "Login with Google"|Logout is not handled via Authgear|Goes back to the app pre-login (e.g. Quora)|


## Web-App (Cookie-based)

This requires them to create a custom domain, and therefore they added `auth.oursky.apps`

By doing so, all apps under the `*.oursky.apps` domain are sharing the login session. e.g. `mail.oursky.apps`, `note.oursky.apps`

SSO is on by nature. There is no way to opt out if the apps live under the same domain. A user is automatically login.

Logout in Cookie-based means "logout from this computer", all refresh tokens that are associated with the same IdP session will also be invalidated.

### Multiple domains (Currently not supported)

While it is not supported yet, it is possible to add more custom domains so there will be multiple login sessions for the different apps. e.g. `auth.skymakers.co.uk` for apps in `*.skymakers.co.uk` which shares the same user database.

## SPA (token-based)

A SPA can be enabled with SSO corresponding to a domain.

**If SSO is on**, IdP session cookie is used to share the session between domains and SPAs. When logging in, the cookie is checked first. If session is valid, the enduser can just click continue as that user. If no session (or `prompt: login`), the enduser need to enter their credentials. The IdP session cookie is set in their browser during the process.

When clicking "logout", the end-user intends to log out all apps of Oursky. Authgear should:

1. Clear refresh token
2. Clear IdP session cookie
3. Invalidate all other refresh tokens that are associated with the same IdP session.

This is particularly useful when Oursky provides a suite of apps. The apps are interconnected to form a unified brand experience.

**If SSO is off**, the IdP session cookie is irrelevant and suppressed. Which means, the cookie is ignored and will not be set upon login.

This is useful when the app is standalone and we don’t want the end-user to associate this app to other products of Oursky.

## Login endpoint

When the end-user login via `auth.oursky.apps/login`, they are redirected to the default **Post-login URL** (and `/settings` if unspecified). 

If the end-user login via the SDK, the client ID and redirect URL is embedded to the login URL in query params. After login, they will be redirected to the allowed Redirect URL. If the redirect URL is missing, they will be redirected to the default redirect URL of that client.

Ordered in priority:

1. Redirect URL in query param (must be whitelisted) 
2. Default redirect URL of the client
3. Post-login URL
4. `/settings`

For SPA, the authentication is completed with a [code challenge](https://www.rfc-editor.org/rfc/rfc7636#section-4.2). The [code exchange](https://www.oauth.com/oauth2-servers/pkce/authorization-code-exchange/) is impossible upon returning to the SPA if the login link is not generated from the SDK in the same session. The `finishAuthentication()` function will return the error. If using the SDK to startAuthentication at this point, depends on SSO is on or not:

- On -> The end-user will need to click "Continue As…"
- Off -> The end-user will need to enter their credentials again

|SPA Login link invalid|Logged in on domain|Logged out on domain|
|----------------------|-------------------|--------------------|
|SSO On|Continue As… x 2|Enter credential -> Continue As…|
|SSO Off|Enter credential -> enter credential|Enter credential -> enter credential|

The **Post-login URL** is set per project basis. The developer should put a link to the portal of their apps. e.g. `https://www.oursky.apps/all` 

If the portal is a web-app, it is immediately logged in; if the portal is an SPA, it will be the same as the code exchange problem mentioned above.

## Mobile apps

These control the SSO behavior:

- Between browser and apps:
    - `shareSessionWithSystemBrowser: true`
- Between apps:
    - relevant settings in `AppGroup` on iOS and `AccountManager API` on Android.
    - In the future, we can introduce `AppGroupPersistentTokenStorage`and `AccountManagerPersistentTokenStorage` options to `tokenStorage`
    - If SSO between apps is enabled, the account is removed from the device; If not, apps can be logout individually.

Reference:

- [https://docs.authgear.com/integrate/single-sign-on](https://docs.authgear.com/integrate/single-sign-on)
- https://github.com/authgear/authgear-server/issues/1391

Invalid auth code is not a problem because it is unlikely to happen.

## OIDC-Compatible client

### 3rd party

Authgear is used in the context of an SSO provider, working just like "Login with Google". When login, user may enter credentials or click "Continue as". Logout is independent to the Authgear session. `*.oursky.apps` is also logged in after the log in.

## Known issue

- Cannot link tokens if not "continue as…"
    - "Login with another account" > "Logout and then continue with another account"
    - User clears the browser’s cookie
- Cookie and refresh tokens' expiry time may be mismatched
    - refresh token created by continuing an IdP session should not exceed the IdP session lifetime.

