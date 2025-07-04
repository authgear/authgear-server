This file documents common models in event payloads.

Although the models appears in the event payload as a json object, this document uses Typescript for stricter type documentations.

## Table of Contents

* [User](#user)
* [Identity](#identity)
* [Authenticator](#authenticator)
* [AMR](#amr)
* [AuthenticationContext](#authenticationcontext)
* [AuthenticationFlow](#authenticationflow)
* [Authentication](#authentication)
* [Identification](#identification)

## User

```typescript
interface User {
  id: string;
  is_anonymized: boolean;
  is_anonymous: boolean;
  is_deactivated: boolean;
  is_disabled: boolean;
  is_verified: boolean;
  last_login_at: string; // ISO 8601
  roles: string[];
  groups: string[];
  standard_attributes: { [key: string]: any };
  custom_attributes: { [key: string]: any };
  created_at: string; // ISO 8601
  updated_at: string; // ISO 8601
}
```

## Identity

```typescript
interface Identity {
  id: string;
  created_at: string; // ISO 8601
  updated_at: string; // ISO 8601
  type: "login_id" | "oauth" | "anonymous" | "biometric" | "passkey" | "ldap";
  claims: { [key: string]: any };
}
```

## Authenticator

```typescript
interface Authenticator {
  id: string;
  created_at: string; // ISO 8601
  updated_at: string; // ISO 8601
  user_id: string;
  type: "password" | "passkey" | "totp" | "oob_otp_email" | "oob_otp_sms";
  is_default: boolean;
  kind: "primary" | "secondary";
}
```

## AMR

```typescript
type AMR =
  | "pwd"
  | "otp"
  | "sms"
  | "mfa"
  | "x_biometric"
  | "x_primary_password"
  | "x_primary_oob_otp_email"
  | "x_primary_oob_otp_sms"
  | "x_primary_passkey"
  | "x_secondary_password"
  | "x_secondary_oob_otp_email"
  | "x_secondary_oob_otp_sms"
  | "x_secondary_totp"
  | "x_recovery_code"
  | "x_device_token";
```

## AuthenticationContext

Details about the current authentication.

```typescript
interface AuthenticationContext {
  user: User | null; // null if user is not known
  asserted_authentications: Authentication[];
  asserted_identifications: Identification[];
  amr: AMR[];
  authentication_flow: AuthenticationFlow | null; // null if the event is not triggered from authenfication flow
}
```

## AuthenticationFlow

```typescript
interface AuthenticationFlow {
  type: "signup" | "promote" | "login" | "signup_login" | "reauth" | "account_recovery";
  name: string;
}
```

## Authentication

```typescript
interface Authentication {
  authentication:
    | "primary_password"
    | "primary_passkey"
    | "primary_oob_otp_email"
    | "primary_oob_otp_sms"
    | "secondary_password"
    | "secondary_totp"
    | "secondary_oob_otp_email"
    | "secondary_oob_otp_sms"
    | "recovery_code"
    | "device_token";
  authenticator: Authenticator | null; // Non-null if authentication is primary_password, primary_passkey, primary_oob_otp_email, primary_oob_otp_sms, secondary_password, secondary_totp, secondary_oob_otp_email, or secondary_oob_otp_sms.
}
```

## Identification

```typescript
interface Identification {
  identification:
    | "email"
    | "phone"
    | "username"
    | "oauth"
    | "passkey"
    | "id_token"
    | "ldap";
  identity: Identity | null; // Non-null if identification is email, phone, username, oauth, passkey, or ldap.
  id_token: string | null; // Non-null if identification is id_token
}
```
