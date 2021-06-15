- [Event](#event)
  * [Event Definition](#event-definition)
  * [Event Shape](#event-shape)
    + [Event Shape Versioning](#event-shape-versioning)
    + [Event Context](#event-context)
  * [Event List](#event-list)
    + [Blocking Events](#blocking-events)
      - [user.pre_create](#userpre-create)
    + [Non-blocking Events](#non-blocking-events)
      - [user.created](#usercreated)
      - [user.authenticated](#userauthenticated)
      - [user.signed_out](#usersigned-out)
      - [user.anonymous.promoted](#useranonymouspromoted)
      - [authentication.identity.login_id.failed](#authenticationidentitylogin-idfailed)
      - [authentication.identity.anonymous.failed](#authenticationidentityanonymousfailed)
      - [authentication.identity.biometric.failed](#authenticationidentitybiometricfailed)
      - [authentication.primary.password.failed](#authenticationprimarypasswordfailed)
      - [authentication.primary.oob_otp_email.failed](#authenticationprimaryoob-otp-emailfailed)
      - [authentication.primary.oob_otp_sms.failed](#authenticationprimaryoob-otp-smsfailed)
      - [authentication.secondary.password.failed](#authenticationsecondarypasswordfailed)
      - [authentication.secondary.totp.failed](#authenticationsecondarytotpfailed)
      - [authentication.secondary.oob_otp_email.failed](#authenticationsecondaryoob-otp-emailfailed)
      - [authentication.secondary.oob_otp_sms.failed](#authenticationsecondaryoob-otp-smsfailed)
      - [authentication.secondary.recovery_code.failed](#authenticationsecondaryrecovery-codefailed)
      - [identity.email.added](#identityemailadded)
      - [identity.email.removed](#identityemailremoved)
      - [identity.email.updated](#identityemailupdated)
      - [identity.phone.added](#identityphoneadded)
      - [identity.phone.removed](#identityphoneremoved)
      - [identity.phone.updated](#identityphoneupdated)
      - [identity.username.added](#identityusernameadded)
      - [identity.username.removed](#identityusernameremoved)
      - [identity.username.updated](#identityusernameupdated)
      - [identity.oauth.connected](#identityoauthconnected)
      - [identity.oauth.disconnected](#identityoauthdisconnected)

# Event

## Event Definition

Events are triggered when an operation is performed.

Events have two kinds, namely Blocking and Non-blocking.

Blocking event is triggered before the operation is performed. The operation can be aborted by webhook handler.

Non-blocking event is triggered after the operation is performed.

## Event Shape

All events have the following shape:

```json5
{
  "id": "0E1E9537-DF4F-4AF6-8B48-3DB4574D4F24",
  "seq": 435,
  "type": "user.created",
  "payload": { /* ... */ },
  "context": { /* ... */ }
}
```

- `id`: The ID of the event.
- `seq`: A monotonically increasing signed 64-bit integer.
- `type`: The type of the event.
- `payload`: The payload of the event, varies with type.
- `context`: The context of the event.

### Event Shape Versioning

All fields are guaranteed that only backward-compatible changes would be made.

- Existing fields would not be removed or changed in meaning.
- New fields may be added.

### Event Context

- `timestamp`: signed 64-bit UNIX timestamp of when this event is generated. Retried deliveries do not affect this field.
- `user_id`: The ID of the user associated with the event. It may be absent. For example, the user has not authenticated yet.
- `preferred_languages`: User preferred languages which are inferred from the request. Return values of `ui_locales` query if it is provided in auth ui, otherwise return languages in `Accept-Language` request header.
- `language`: User locale which is derived based on user's preferred languages and app's languages config.
- `triggered_by`: Triggered by indicates who triggered the events, values can be `user` or `admin_api`. `user` means it is triggered by user in auth ui. `admin_api` means it is triggered by admin api or admin portal.

## Event List

### Blocking Events

- [user.pre_create](#userpre_create)

#### user.pre_create

Occurs right before the user creation. User can be created by user signup, user signup as anonymous user, admin api or admin portal create an user.

```json5
{
  "context": {
    ....
    "oauth": {
      "state": "the-value-of-state-if-provided"
    }
  },
  "payload": {
    "user": { /* ... */ },
    "identities": [ { /* ... */ } ]
  }
}
```

- `oauth.state`: OAuth state if the signup is triggered through authorize endpoint with state parameter.

### Non-blocking Events

- [user.created](#usercreated)
- [user.authenticated](#userauthenticated)
- [user.signed_out](#usersigned-out)
- [user.anonymous.promoted](#useranonymouspromoted)
- [authentication.identity.login_id.failed](#authenticationidentitylogin-idfailed)
- [authentication.identity.anonymous.failed](#authenticationidentityanonymousfailed)
- [authentication.identity.biometric.failed](#authenticationidentitybiometricfailed)
- [authentication.primary.password.failed](#authenticationprimarypasswordfailed)
- [authentication.primary.oob_otp_email.failed](#authenticationprimaryoob-otp-emailfailed)
- [authentication.primary.oob_otp_sms.failed](#authenticationprimaryoob-otp-smsfailed)
- [authentication.secondary.password.failed](#authenticationsecondarypasswordfailed)
- [authentication.secondary.totp.failed](#authenticationsecondarytotpfailed)
- [authentication.secondary.oob_otp_email.failed](#authenticationsecondaryoob-otp-emailfailed)
- [authentication.secondary.oob_otp_sms.failed](#authenticationsecondaryoob-otp-smsfailed)
- [authentication.secondary.recovery_code.failed](#authenticationsecondaryrecovery-codefailed)
- [identity.email.added](#identityemailadded)
- [identity.email.removed](#identityemailremoved)
- [identity.email.updated](#identityemailupdated)
- [identity.phone.added](#identityphoneadded)
- [identity.phone.removed](#identityphoneremoved)
- [identity.phone.updated](#identityphoneupdated)
- [identity.username.added](#identityusernameadded)
- [identity.username.removed](#identityusernameremoved)
- [identity.username.updated](#identityusernameupdated)
- [identity.oauth.connected](#identityoauthconnected)
- [identity.oauth.disconnected](#identityoauthdisconnected)

#### user.created

Occurs after a new user is created. User can be created by user signup, user signup as anonymous user, admin api or admin portal create an user.

```json5
{
  "payload": {
    "user": { /* ... */ },
    "identities": [ { /* ... */ } ]
  }
}
```

#### user.authenticated

Occurs after user logged in.

```json5
{
  "payload": {
    "user": { /* ... */ },
    "session": { /* ... */ }
  }
}
```

#### user.signed_out

Occurs after the user signed out, or revoked their session.
Note that there is no event when the session expires normally.

```json5
{
  "payload": {
    "user": { /* ... */ },
    "session": { /* ... */ }
  }
}
```

#### user.anonymous.promoted

Occurs whenever an anonymous user is promoted to normal user.

```json5
{
  "payload": {
    "anonymous_user": { /* ... */ },
    "user": { /* ... */ },
    "identities": [{ /* ... */ }]
  }
}
```

#### authentication.identity.login_id.failed

Occurs after a Email / Phone / Username Login ID was attempted to log in but it does not exist.

```json5
{
  "payload": {
    "login_id": "..."
  }
}
```

#### authentication.identity.anonymous.failed

Occurs after an anonymous user attempted to log in but failed to do so.

```json5
{
  "payload": {
    "user": { /* ... */ }
  }
}
```

#### authentication.identity.biometric.failed

Occurs after an user attempted to log in with biometric but failed to do so.

```json5
{
  "payload": {
    "user": { /* ... */ }
  }
}
```

#### authentication.primary.password.failed

Occurs after the user failed to input their primary password.

```json5
{
  "payload": {
    "user": { /* ... */ }
  }
}
```

#### authentication.primary.oob_otp_email.failed

Occurs after the user failed to input the OTP delivered to their email address.

```json5
{
  "payload": {
    "user": { /* ... */ }
  }
}
```

#### authentication.primary.oob_otp_sms.failed

Occurs after the user failed to input the OTP delivered to their phone.

```json5
{
  "payload": {
    "user": { /* ... */ }
  }
}
```

#### authentication.secondary.password.failed

Occurs after the user failed to input their secondary password.

```json5
{
  "payload": {
    "user": { /* ... */ }
  }
}
```

#### authentication.secondary.totp.failed

Occurs after the user failed to input the time-based one-time password.

```json5
{
  "payload": {
    "user": { /* ... */ }
  }
}
```

#### authentication.secondary.oob_otp_email.failed

Occurs after the user failed to input the OTP delivered to their email address.

```json5
{
  "payload": {
    "user": { /* ... */ }
  }
}
```

#### authentication.secondary.oob_otp_sms.failed

Occurs after the user failed to input the OTP delivered to their phone.

```json5
{
  "payload": {
    "user": { /* ... */ }
  }
}
```

#### authentication.secondary.recovery_code.failed

Occurs after the user failed to input the recovery code.

```json5
{
  "payload": {
    "user": { /* ... */ }
  }
}
```

#### identity.email.added

Occurs when a new email is added to existing user. Email can be added by user in setting page, added by admin through admin api or portal.

```json5
{
  "payload": {
    "user": { /* ... */ },
    "identity": { /* ... */ }
  }
}
```


#### identity.email.removed

Occurs when email is removed from existing user. Email can be removed by user in setting page, removed by admin through admin api or portal.

```json5
{
  "payload": {
    "user": { /* ... */ },
    "identity": { /* ... */ }
  }
}
```


#### identity.email.updated

Occurs when email is updated. Email can be updated by user in setting page.

```json5
{
  "payload": {
    "user": { /* ... */ },
    "old_identity": { /* ... */ },
    "new_identity": { /* ... */ }
  }
}
```

#### identity.phone.added

Occurs when a new phone number is added to existing user. Phone number can be added by user in setting page, added by admin through admin api or portal.

```json5
{
  "payload": {
    "user": { /* ... */ },
    "identity": { /* ... */ }
  }
}
```


#### identity.phone.removed

Occurs when phone number is removed from existing user. Phone number can be removed by user in setting page, removed by admin through admin api or portal.

```json5
{
  "payload": {
    "user": { /* ... */ },
    "identity": { /* ... */ }
  }
}
```


#### identity.phone.updated

Occurs when phone number is updated. Phone number can be updated by user in setting page.


```json5
{
  "payload": {
    "user": { /* ... */ },
    "old_identity": { /* ... */ },
    "new_identity": { /* ... */ }
  }
}
```

#### identity.username.added

Occurs when a new username is added to existing user. Username can be added by user in setting page, added by admin through admin api or portal.

```json5
{
  "payload": {
    "user": { /* ... */ },
    "identity": { /* ... */ }
  }
}
```


#### identity.username.removed

Occurs when username is removed from existing user. Username can be removed by user in setting page, removed by admin through admin api or portal.

```json5
{
  "payload": {
    "user": { /* ... */ },
    "identity": { /* ... */ }
  }
}
```


#### identity.username.updated

Occurs when username is updated. Username can be updated by user in setting page.

```json5
{
  "payload": {
    "user": { /* ... */ },
    "old_identity": { /* ... */ },
    "new_identity": { /* ... */ }
  }
}
```

#### identity.oauth.connected

Occurs when user connected to a new OAuth provider.

```json5
{
  "payload": {
    "user": { /* ... */ },
    "identity": { /* ... */ }
  }
}
```

#### identity.oauth.disconnected

Occurs when user disconnected from an OAuth provider. It can be done by user disconnected OAuth provider in the setting page, or admin removed OAuth identity through admin api or portal.

```json5
{
  "payload": {
    "user": { /* ... */ },
    "identity": { /* ... */ }
  }
}
```
