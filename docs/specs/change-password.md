# Change Password

- [Introduction](#introduction)
- [Use Cases](#use-cases)
- [Design](#design)
  - [Change Password](#change-password-1)
    - [Error Response](#error-response)
  - [Force Change Password on Next Login](#force-change-password-on-next-login)
    - [Error Response](#error-response-1)

## Introduction

This document describes the feature for administrators to change the password of a user through the Portal Admin API.

## Use Cases

- An administrator needs to change the password of an existing user.
- An administrator needs to force a user to change their password.
- An administrator needs to reset the password of a user who has forgotten their password.

## Design

### Change Password

`resetPassword` mutation is an existing mutation in the Portal Admin API. It will be extended to support following:
- Send a notification to user to inform them of the password change.
- Force the user to change their password on next login.
- Generate a new password automatically. In this case, the new password will be sent to the user through email.

```graphql
type Mutation {
  # other root fields...

  resetPassword(
    input: ResetPasswordInput!
  ): ResetPasswordPayload!
}

input ResetPasswordInput {
  # other fields...

  # Email will be sent to user with new password
  sendPassword: Boolean
  # User will be forced to change password on next login
  forceChangeOnLogin: Boolean
  # Generate password and send it to user through email
  generateAndSendPassword: Boolean
}
```

#### Error Response

|Description|Name|Reason|Info|
|---|---|---|---|
|Email identity not found, required if `sendPassword` or `generateAndSendPassword` is set to `true`.|`NotFound`|`EmailIdentityNotFound`|-|

### Force Change Password on Next Login

A `forceChangePassword` and `cancelForceChangePassword` mutation will be added to the Portal Admin API for forcing a user to change their password on next login.

```graphql
type Mutation {
  # other root fields...

  forceChangePassword(
    input: ForceChangePasswordInput!
  ): ForceChangePasswordPayload!

  cancelForceChangePassword(
    input: CancelForceChangePasswordInput!
  ): CancelForceChangePasswordPayload!
}

input ForceChangePasswordInput {
  userID: ID!
}

input CancelForceChangePasswordInput {
  userID: ID!
}
```

#### Error Response

|Description|Name|Reason|Info|
|---|---|---|---|
|User Not Found|`NotFound`|`UserNotFound`|-|
|Password Authenticator Not Found|`NotFound`|`PasswordAuthenticatorNotFound`|-|
