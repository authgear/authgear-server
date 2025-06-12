# Browser Single Sign On

This is a cookie based SSO mechanism. User must be authenticated through a same browser in all apps.

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

- [https://docs.authgear.com/authentication-and-access/single-sign-on](https://docs.authgear.com/authentication-and-access/single-sign-on)
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
- Refresh tokens that are generated by biometrics or anonymous login cannot be grouped, even if `sso_enabled` is `true`.

## Implementation details

### Server-side

- Authorization Endpoint
    - New query parameter `x_sso_enabled`
        - When `x_sso_enabled=false`
            - Equal to `x_suppress_idp_session_cookie=true`.
            - The existing IdP session in cookie is ignored, and login will NOT create IdP Session.
            - Mark refresh token's `sso_enabled` to `false`.
        - When `x_sso_enabled=true`
            - Equal to `x_suppress_idp_session_cookie=false`.
            - Mark refresh token's `sso_enabled` to `true`.
            - SSO enabled refresh token will be valid only if its IdP session is valid
            - When updating the refresh token's last access time, its IdP session's last access time will also be updated. So that the IdP session won't be invalidated due to inactivity.
    - `x_suppress_idp_session_cookie` remains unchanged for backward compatibility, but the new SDKs won't send this parameter anymore. It is replaced by `x_sso_enabled`.
    - When the developer uses `x_sso_enabled=true` + `prompt=login`, the enduser will need to log in again. The new refresh token will be linked to the new IdP session, and the old IdP session will be overwritten and revoked.

- Preserve `client_id` and `redirect_uri` query during the whole login flow.
    - After authentication, redirect the enduser based on following whitelisted
        1. Redirect URL in query param (must be whitelisted) 
        2. Default redirect URL of the client
        3. Post-login URL
        4. `/settings`
    - Default redirect URL of the client
        - In the config, the first item of **redirect_uris** is the default implicitly
        - Update the Portal UI to allow setting the default

- Logout
    - Revoke IdP session
        - Check if any associate refresh tokens that have `sso_enabled=true`, revoke them
    - Revoke refresh token
        - If the refresh token's `sso_enabled` is `true`, invalidates its IdP session and all its siblings with `sso_enabled=true`
        - If the refresh token's `sso_enabled` is `false` (default), revoke the refresh token only

- "Continue as" screen
    - When the enduser chooses **Login with another account**, the existing IdP session (including its refresh token with `sso_enabled=true`) will be revoked after the new authentication is completed.

- Update application config
    - For `x_application_type=spa`, change `post_logout_uris` to optional. Change the label in the Portal to **Post Logout Redirect URI (Legacy)**. See [Web SDK](#web-sdk) for details. Show a description to mention that this field is not needed after which SDK version.

- Session listing
    - Combining Sessions
        - Both the settings page, admin API and Portal combine sessions in the same way
        - For refresh tokens that have `sso_enabled=true`, its IdP session and siblings will be combined into a single entry. Since revoking one of them will also revoke the others. The entry will be shown as a single entry without grouping.
        - The sessions that cannot be combined will be listed separately without grouping. (Refresh tokens with `sso_enabled=false` and IdP sessions don't have `sso_enabled` refresh tokens)
    - Settings page
        - Sessions will be shown as a list without grouping
    - Admin API and Portal
        - The "sessions" API will be updated. For the combined refresh token and IdP sessions, only the IdP session will be added to the list.

### Web SDK

- Add `ssoEnabled` to `authgear.configure`, it is useful only when `sessionType=refresh_token`. Default `ssoEnabled` is `false`.
- When `sessionType=refresh_token`
    - Update authz endpoint to remove `x_suppress_idp_session_cookie` and add `x_sso_enabled`
    - Update `authgear.logout` to revoke the refresh token only, don't need to redirect to `logout`. (Revoking `sso_enabled` refresh token will also revoke its IdP session and siblings).
- When `sessionType=cookie`
    - no change

### Mobile SDK

- Replace `shareSessionWithSystemBrowser` with `ssoEnabled`. Default `ssoEnabled=false`.

## Backward compatibility

- If the developer updates the server and the apps are still using the old version SDK.
    - The app's behavior should remind unchanged. The updated server handles `x_suppress_idp_session_cookie` for backward compatibility.
- If the developer updates the SDK, but the server is not updated. (Don't recommend)
    - Web SDK (`sessionType=cookie`)
        - The app's behavior should remind unchanged.
    - Web SDK (`sessionType=refresh_token`)
        - Login
            - The new SDK will send both `x_suppress_idp_session_cookie=true` and `x_sso_enabled=false` to the server when `ssoEnabled` is `false` (default). The old version will read `x_suppress_idp_session_cookie=true` and no IdP session will be generated, so the behavior reminds unchanged.
            - If developers set `ssoEnabled` to `true`. The new SDK will send `x_sso_enabled=true`, but the old server won't be able to handle it. Both the Idp session and refresh token will be generated, but they will not be in the same SSO group.
        - Logout
            - The old SDK will redirect the user to the Authgear logout page after revoking the refresh token. But after updating the SDK, only the current refresh token will be revoked. So if there is any IdP session/refresh token of other apps, they will not be revoked.
    - Mobile SDK
        - If the developer specifies `shareSessionWithSystemBrowser` when calling `authgear.configure`, there will be compilation error.
        - If the developer doesn't specify either `shareSessionWithSystemBrowser` or `ssoEnabled` (this is the case we want to handle). The new SDK will send both `x_suppress_idp_session_cookie=true` and `x_sso_enabled=false` to the server when `ssoEnabled` is `false` (default). Only refresh token will be generated and no Idp session will be created, which matches the original default behavior.
        - If developers set `ssoEnabled` to `true`. The new SDK will send `x_sso_enabled=true`, but the old server won't be able to handle it. Both the Idp session and refresh token will be generated, but they will not be in the same SSO group.
