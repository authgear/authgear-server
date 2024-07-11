# Account Management API

Account Management API is an API for account management on behalf of the end user. Account Management API requires a authenticated session, either by cookie or the HTTP header Authorization.

Account Management API supports the following operations:

- [Manage identifications](#manage-identifications)
  - List identifications
  - Add a new email, with verification
  - Add a new phone number, with verification
  - Add a new username
  - [Add an OAuth provider account, with authorization code flow](#add-an-oauth-provider-account-with-authorization-code-flow)
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

### Add an OAuth provider account, with authorization code flow

> This endpoint is considered as advanced. It is assumed that you are capable of
> receiving OAuth callback via `redirect_uri`.

`POST /api/v2/account/identification`

Request

```json
{
  "identification": "oauth",
  "alias": "google",
  "redirect_uri": "https://myapp.authgear.cloud/sso/oauth2/callback/google"
}
```

- `redirect_uri`: The settings page of Authgear uses `<origin>/sso/oauth2/callback/{alias}` to receive the OAuth callback. You have to specify your own redirect URI to your app or your website.

Response

```json
{
  "result": {
    "oauth": {
      "token": "oauthtoken_blahblahblah",
      "authorization_url": "https://www.google.com?client_id=client_id&rredirect_uri=redirect_uri"
    }
  }
}
```

- `oauth.authorization_url`: You MUST redirect the end-user to this URL to continue the authorization code flow. You can add `state` to the URL to help you maintain state and do CSRF protection.

Finally, the OAuth provider will call your redirect URI with `state` (if you have provided), and other query parameters. You then call the following endpoint.

`POST /api/v2/account/identification/oauth`

Request

```json
{
  "token": "oauthtoken_blahblahblah",
  "query": "code=code"
}
```

- `token`: `oauth.token` in the previous response.
- `query`: The query of the redirect URI.

Response

If successful, then the OAuth provider account is added.

```json
{
  "result": {
    "identification_method": {
      "identification": "oauth",
      "provider_type": "google",
      "provider_user_id": "USER_ID_AT_GOOGLE",
      "alias": "google",
      "claims": {
        "email": "user@gmail.com"
      },
      "created_at": "2006-01-02T03:04:05Z",
      "updated_at": "2006-01-02T03:04:05Z"
    }
  }
}
```

If the OAuth provider account is already taken by another account, then you will receive the following error.

```json
{
  "error": {
    "name": "Invalid",
    "reason": "InvariantViolated",
    "message": "identity already exists",
    "code": 400,
    "info": {
      "cause": {
        "kind": "DuplicatedIdentity"
      }
    }
  }
}
```
