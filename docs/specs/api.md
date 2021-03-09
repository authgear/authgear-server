# API

Here are the endpoints that are not part of OIDC, but essential to provide some features of Authgear.

## /api/identity/biometric/remove

This endpoint is for removing Biometric identity.
The user still can remove biometric identity in the settings page.

### Request Schema

```json
{
  "type": "object",
  "additionalProperties": false,
  "properties": {
    "kid": { "type": "string" }
  },
  "required": ["kid"]
}
```

### Response Schema

No response is necessary.

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
