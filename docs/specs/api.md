# API

Here are the endpoints that are not part of OIDC, but essential to provide some features of Authgear.

If the api request is success, the response of the API will be a JSON body with key `result`.

```json
{
  "result": { /* the result object */ }
}
```

If the api request is failed, the response of the API will be a JSON body with key `error`.

```json
{
  "error": {
    "name": "Invalid",
    "reason": "ValidationFailed",
    "message": "invalid request body",
    "code": 400,
    "info": { /* the error info */ }
  }
}
```

## /oauth2/challenge

This endpoint is for requesting a one-time use, short-lived challenge.
The challenge is used in anonymous user and biometric authentication.

## Request Schema

```json
{
  "type": "object",
  "additionalProperties": false,
  "properties": {
    "purpose": { "type": "string" }
  },
  "required": ["purpose"]
}
```

### Response Result Object Schema

```json
{
  "type": "object",
  "properties": {
    "token": { "type": "string" },
    "expire_at": { "type": "string" }
  },
  "required": ["token", "expire_at"]
}
```

## /api/anonymous_user/signup

This api is for signing up an new anonymous user in the Web SDK.

For the `cookie` app,

1. The api will check the cookie.
1. If there is no logged in session, new user will be created. The server will return the `Set-Cookie` header and the result object will be an empty object.
1. If the logged in user is an anonymous user, a success response will be returned. No new cookie header will be return and the result object will be an empty object.
1. If the logged in user is a normal user, an error will be returned.

For the `refresh_token` app,

1. If the user has logged in, `refresh_token` will be included in the request.
1. If there is no logged in session, new user will be created. `refresh_token` and `access_token` will be issued in the result object.
1. If the logged in user is an anonymous user, a success response will be returned. New `access_token` will be issued in the result object.
1. If the logged in user is a normal user, an error will be returned.

### Request Schema

```json
{
  "type": "object",
  "additionalProperties": false,
  "properties": {
    "client_id": { "type": "string" },
    "refresh_token": { "type": "string" },
    "session_type": {
      "type": "string",
      "enum": ["cookie", "refresh_token"]
    }
  },
  "required": ["client_id", "session_type"]
}
```

### Response Result Object Schema

When the `session_type` is `cookie`, the server will return the `Set-Cookie` header and the result object will be an empty object.

When the `session_type` is `refresh_token`, the tokens will be issued directly.

```json
{
  "type": "object",
  "properties": {
    "token_type": { "type": "string" },
    "access_token": { "type": "string" },
    "refresh_token": { "type": "string" },
    "expires_in": { "type": "integer" }
  },
  "required": ["token_type", "access_token", "expires_in"]
}
```
## /api/anonymous_user/promotion_code

### Request Schema

This api is for requesting a promotion code for promoting an anonymous user.

When the `session_type` is `cookie`, the session information should be provided through the `Cookie` header. The request body should be empty.

When the `session_type` is `refresh_token`, the refresh token should be provided through the request body with the following schema:

```json
{
  "type": "object",
  "additionalProperties": false,
  "properties": {
    "refresh_token": { "type": "string" }
  },
  "required": ["refresh_token"]
}
```
### Response Result Object Schema

```json
{
  "type": "object",
  "properties": {
    "promotion_code": { "type": "string" },
    "expire_at": { "type": "string" }
  },
  "required": ["promotion_code", "expire_at"]
}
```

The `promotion_code` should be used in the authorization endpoint, see [login_hint](/oidc.md#login_hint).

