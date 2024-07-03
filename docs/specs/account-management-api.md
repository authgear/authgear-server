# Account Management API

Account Management API is an API for account management on behalf of the end user. Account Management API requires a authenticated session, either by cookie or the HTTP header Authorization.

Account Management API supports the following operations:

- [Manage identifications](#manage-identifications)
  - [List identifications](#list-identifications)
  - [Add a new email, with verification](#add-a-new-email-with-verification)
  - [Add a new phone number, with verification](#add-a-new-phone-number-with-verification)
  - [Add a new username](#add-a-new-username)
  - [Add an OAuth provider account, with authorization code flow](#add-an-oauth-provider-account-with-authorization-code-flow)
  - [Add a biometric](#add-a-biometric)
  - Add a passkey
  - [Remove an email](#remove-an-email)
  - [Remove a phone number](#remove-a-phone-number)
  - [Remove a username](#remove-a-username)
  - [Remove an OAuth provider account](#remove-an-oauth-provider-account)
  - [Remove a biometric](#remove-a-biometric)
  - [Remove a passkey](#remove-a-passkey)
  - [Update an email, with verification](#update-an-email-with-verification)
  - [Update a phone number, with verification](#update-a-phone-number-with-verification)
  - [Update a username](#update-a-username)
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
- [Auxiliary operations](#auxiliary-operations)
  - [Verify OTP](#verify-otp)
  - [Resend OTP](#resend-otp)

## Manage identifications

### List identifications

`GET /api/v1/account/identification`

Response

```json
{
  "result": {
    "identifications": [
      {
        "identification": "email",
        "login_id": "user@example.com",
        "claims": {
          "email": "user@example.com"
        },
        "created_at": "2006-01-02T03:04:05Z",
        "updated_at": "2006-01-02T03:04:05Z"
      },
      {
        "identification": "username",
        "login_id": "johndoe",
        "claims": {
          "preferred_username": "johndoe"
        },
        "created_at": "2006-01-02T03:04:05Z",
        "updated_at": "2006-01-02T03:04:05Z"
      },
      {
        "identification": "phone",
        "login_id": "+85251000001",
        "claims": {
          "phone_number": "+85251000001"
        },
        "created_at": "2006-01-02T03:04:05Z",
        "updated_at": "2006-01-02T03:04:05Z"
      },
      {
        "identification": "oauth",
        "provider_type": "google",
        "provider_user_id": "USER_ID_AT_GOOGLE",
        "alias": "google",
        "claims": {
          "email": "user@gmail.com"
        },
        "created_at": "2006-01-02T03:04:05Z",
        "updated_at": "2006-01-02T03:04:05Z"
      },
      {
        "identification": "passkey",
        "credential_id": "hlNmaS6DQjV4voxP8SPJDzDG-j79nWL8r4OTgcPizi0"
        "created_at": "2006-01-02T03:04:05Z",
        "updated_at": "2006-01-02T03:04:05Z"
      },
      {
        "identification": "biometric",
        "key_id": "KEY_ID",
        "display_name": "iPhone 12 mini",
        "created_at": "2006-01-02T03:04:05Z",
        "updated_at": "2006-01-02T03:04:05Z"
      }
    ]
  }
}
```

### Add a new email, with verification

`POST /api/v1/account/identification`

Request

```json
{
  "identification": "email",
  "login_id": "user@example.com"
}
```

Response

If verification is not required, then the email is added immediately.

```json
{
  "result": {
    "identification_method": {
      "identification": "email",
      "login_id": "user@example.com",
      "claims": {
        "email": "user@example.com"
      },
      "created_at": "2006-01-02T03:04:05Z",
      "updated_at": "2006-01-02T03:04:05Z"
    }
  }
}
```

If verification is required, you need to perform verification.

```json
{
  "result": {
    "verification": {
      "token": "verificationtoken_blahblahblah",
      "channel": "email",
      "otp_form": "code",
      "code_length": 6,
      "can_resend_at": "2023-09-21T00:00:00+08:00",
      "can_check": false,
      "failed_attempt_rate_limit_exceeded": false
    }
  }
}
```

Use [Verify OTP](#verify-otp) and [Resend OTP](#resend-otp) to continue.

If the email is already taken by another account, then you will receive the following error.

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

### Add a new phone number, with verification

`POST /api/v1/account/identification`

Request

```json
{
  "identification": "phone",
  "login_id": "+85251000001"
}
```

Response

If verification is not required, then the phone number is added immediately.

```json
{
  "result": {
    "identification_method": {
      "identification": "phone",
      "login_id": "+85251000001",
      "claims": {
        "phone_number": "+85251000001"
      },
      "created_at": "2006-01-02T03:04:05Z",
      "updated_at": "2006-01-02T03:04:05Z"
    }
  }
}
```

If verification is required, you need to perform verification.

```json
{
  "result": {
    "verification": {
      "token": "verificationtoken_blahblahblah",
      "channel": "sms",
      "otp_form": "code",
      "code_length": 6,
      "can_resend_at": "2023-09-21T00:00:00+08:00",
      "can_check": false,
      "failed_attempt_rate_limit_exceeded": false
    }
  }
}
```

Use [Verify OTP](#verify-otp) and [Resend OTP](#resend-otp) to continue.

If the phone number is already taken by another account, then you will receive the following error.

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

### Add a new username

`POST /api/v1/account/identification`

Request

```json
{
  "identification": "username",
  "login_id": "johndoe"
}
```

Response

```json
{
  "result": {
    "identification_method": {
      "identification": "username",
      "login_id": "johndoe",
      "claims": {
        "preferred_username": "johndoe"
      },
      "created_at": "2006-01-02T03:04:05Z",
      "updated_at": "2006-01-02T03:04:05Z"
    }
  }
}
```

If the username is already taken by another account, then you will receive the following error.

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

### Add a biometric

Please use the existing SDK method `enableBiometric()` to do so.

### Remove an email

`DELETE /api/v1/account/identification`

Request

```json
{
  "identification": "email",
  "login_id": "user@example.com"
}
```

Response

```json
{
  "result": {}
}
```

If it is disallowed to delete an email, you will receive the following error.

```json
{
  "error": {
    "name": "Invalid",
    "reason": "InvariantViolated",
    "message": "identity modification disabled",
    "code": 400,
    "info": {
      "cause": {
        "kind": "IdentityModifyDisabled"
      }
    }
  }
}
```

### Remove a phone number

`DELETE /api/v1/account/identification`

Request

```json
{
  "identification": "phone",
  "login_id": "+85251000001"
}
```

Response

```json
{
  "result": {}
}
```

If it is disallowed to delete a phone number, you will receive the following error.

```json
{
  "error": {
    "name": "Invalid",
    "reason": "InvariantViolated",
    "message": "identity modification disabled",
    "code": 400,
    "info": {
      "cause": {
        "kind": "IdentityModifyDisabled"
      }
    }
  }
}
```

### Remove a username

`DELETE /api/v1/account/identification`

Request

```json
{
  "identification": "username",
  "login_id": "johndoe"
}
```

Response

```json
{
  "result": {}
}
```

If it is disallowed to delete a username, you will receive the following error.

```json
{
  "error": {
    "name": "Invalid",
    "reason": "InvariantViolated",
    "message": "identity modification disabled",
    "code": 400,
    "info": {
      "cause": {
        "kind": "IdentityModifyDisabled"
      }
    }
  }
}
```

### Remove an OAuth provider account

`DELETE /api/v1/account/identification`

Request

```json
{
  "identification": "oauth",
  "alias": "google",
  "provider_user_id": "USER_ID_AT_GOOGLE"
}
```

Response

```json
{
  "result": {}
}
```

If it is disallowed to delete an OAuth provider account, you will receive the following error.

```json
{
  "error": {
    "name": "Invalid",
    "reason": "InvariantViolated",
    "message": "identity modification disabled",
    "code": 400,
    "info": {
      "cause": {
        "kind": "IdentityModifyDisabled"
      }
    }
  }
}
```

### Remove a biometric

`DELETE /api/v1/account/identification`

Request

```json
{
  "identification": "biometric",
  "key_id": "KEY_ID"
}
```

Response

```json
{
  "result": {}
}
```

### Remove a passkey

`DELETE /api/v1/account/identification`

Request

```json
{
  "identification": "biometric",
  "credential_id": "hlNmaS6DQjV4voxP8SPJDzDG-j79nWL8r4OTgcPizi0"
}
```

Response

```json
{
  "result": {}
}
```

### Update an email, with verification

`PUT /api/v1/account/identification`

Request

```json
{
  "identification": "email",
  "old_login_id": "user@example.com",
  "new_login_id": "user1@example.com"
}
```

Response

If verification is not required, then the email is updated immediately.

```json
{
  "result": {
    "identification_method": {
      "identification": "email",
      "login_id": "user1@example.com",
      "claims": {
        "email": "user1@example.com"
      },
      "created_at": "2006-01-02T03:04:05Z",
      "updated_at": "2006-01-02T03:04:05Z"
    }
  }
}
```

If verification is required, you need to perform verification.

```json
{
  "result": {
    "verification": {
      "token": "verificationtoken_blahblahblah",
      "channel": "email",
      "otp_form": "code",
      "code_length": 6,
      "can_resend_at": "2023-09-21T00:00:00+08:00",
      "can_check": false,
      "failed_attempt_rate_limit_exceeded": false
    }
  }
}
```

Use [Verify OTP](#verify-otp) and [Resend OTP](#resend-otp) to continue.

If the email is already taken by another account, then you will receive the following error.

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

### Update a phone number, with verification

`PUT /api/v1/account/identification`

Request

```json
{
  "identification": "phone",
  "old_login_id": "+85251000001",
  "new_login_id": "+85251000002"
}
```

Response

If verification is not required, then the phone number is updated immediately.

```json
{
  "result": {
    "identification_method": {
      "identification": "phone",
      "login_id": "+85251000002",
      "claims": {
        "phone_number": "+85251000002"
      },
      "created_at": "2006-01-02T03:04:05Z",
      "updated_at": "2006-01-02T03:04:05Z"
    }
  }
}
```

If verification is required, you need to perform verification.

```json
{
  "result": {
    "verification": {
      "token": "verificationtoken_blahblahblah",
      "channel": "sms",
      "otp_form": "code",
      "code_length": 6,
      "can_resend_at": "2023-09-21T00:00:00+08:00",
      "can_check": false,
      "failed_attempt_rate_limit_exceeded": false
    }
  }
}
```

Use [Verify OTP](#verify-otp) and [Resend OTP](#resend-otp) to continue.

If the phone number is already taken by another account, then you will receive the following error.

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

### Update a username

`PUT /api/v1/account/identification`

Request

```json
{
  "identification": "username",
  "old_login_id": "johndoe",
  "new_login_id": "janedoe"
}
```

Response

```json
{
  "result": {
    "identification_method": {
      "identification": "username",
      "login_id": "janedoe",
      "claims": {
        "preferred_username": "janedoe"
      },
      "created_at": "2006-01-02T03:04:05Z",
      "updated_at": "2006-01-02T03:04:05Z"
    }
  }
}
```

If the username is already taken by another account, then you will receive the following error.

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

## Auxiliary operations

### Verify OTP

`POST /api/v1/account/otp/verify`

Request

```json
{
  "token": "verificationtoken_blahblahblah",
  "code": "123456"
}
```

Response

If successful, then the response is different depending on what operation you are originally performing.

If the OTP is incorrect, you will receive the following error

```json
{
  "error": {
    "name": "Forbidden",
    "reason": "InvalidOTPCode",
    "message": "invalid otp code",
    "code": 403,
    "info": {
      "cause": {
        "kind": "InvalidCode"
      }
    }
  }
}
```

### Resend OTP

`POST /api/v1/account/otp/resend`

Request

```json
{
  "token": "verificationtoken_blahblahblah"
}
```

Response

The `can_resend_at` is updated.

```json
{
  "result": {
    "verification": {
      "token": "verificationtoken_blahblahblah",
      "channel": "email",
      "otp_form": "code",
      "code_length": 6,
      "can_resend_at": "2023-09-21T00:00:00+08:00",
      "can_check": false,
      "failed_attempt_rate_limit_exceeded": false
    }
  }
}
```
