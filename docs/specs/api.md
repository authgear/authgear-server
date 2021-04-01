# API

Here are the endpoints that are not part of OIDC, but essential to provide some features of Authgear.

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

### Response Schema

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
