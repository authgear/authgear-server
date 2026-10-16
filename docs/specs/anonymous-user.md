# Anonymous User

In this page we are going to explain how the SDKs authenticate and promote an anonymous user in different environments.

## Mobile SDK

### Authentication

1. Mobile SDK gets or creates the keypair, and store it into the native encrypted store.
1. Mobile SDK calls the OIDC token endpoint with [`grant_type`](./oidc.md#grant_type) and [Anonymous Identity JWT](./user-model.md#anonymous-identity-jwt), server will check and create the user if it is necessary. Tokens will be issued directly.

### Promotion

1. Mobile SDK obtains the keypair from the native encrypted store.
1. Mobile SDK opens the authorization endpoint with [`login_hint`](./oidc.md#login_hint), the login_hint should be a URL with query parameters `type=anonymous` and `jwt`.

## Web SDK

### Authentication

1. Web SDK calls the [/api/anonymous_user/signup](./api.md/#apianonymous_usersignup) api to signup an new user.
1. The app should call `authenticateAnonymously` only if the user has not logged in.
1. If the app calls `authenticateAnonymously` if the user has logged in.
    1. If the logged in user is a anonymous user, current user info will be returned.
    2. If the logged in user is a normal user, error will be returned.

### Promotion

1. Web SDK calls the [/api/anonymous_user/promotion_code](./api.md/#apianonymous_userpromotion_code) api to request a promotion code.
1. Web SDK opens the authorization endpoint with [`login_hint`](./oidc.md#login_hint), the login_hint should be a URL with query parameters `type=anonymous` and `promotion_code`.

## Resource Indicators

Signing up or logging in an anonymous user is a [new grant](./access-token-audience-binding.md#new-grant). Both `grant_type=urn:authgear:params:oauth:grant-type:anonymous-request` and [/api/anonymous_user/signup](./api.md#apianonymous_usersignup) with `session_type=refresh_token` take an optional `resource` and `scope`:

- `scope` omitted: `openid offline_access https://authgear.com/scopes/full-access`.
- `scope` given: it must include `offline_access` and `https://authgear.com/scopes/full-access`, otherwise `invalid_scope`.

When `/api/anonymous_user/signup` is given a `refresh_token`, no grant is created; `resource` and `scope` follow the [`refresh_token` grant](./access-token-audience-binding.md#refresh_token-grant). With `session_type=cookie`, which issues no token, either parameter fails the request.

Promotion goes through the authorization endpoint, which takes `resource` as usual.
