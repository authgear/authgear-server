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

### Request Schema

```json
{
  "type": "object",
  "additionalProperties": false,
  "properties": {
    "client_id": { "type": "string" },
    "session_type": {
      "type": "string",
      "enum": ["cookie", "refresh_token"]
    }
  },
  "required": ["client_id", "session_type"]
}
```

### Response Result Object Schema

When the `session_type` is `cookie`, the server will return the `Set-Cookie` header and the response body will be empty.

When the `session_type` is `refresh_token`, the tokens will be issued directly.

```json
{
  "type": "object",
  "properties": {
    "token_type": { "type": "string" },
    "access_token": { "type": "string" },
    "refresh_token": { "type": "string" },
    "expires_in": { "type": "string" }
  },
  "required": ["token_type", "access_token", "refresh_token", "expires_in"]
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

