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
