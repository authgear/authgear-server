# SDK - Settings Actions

This document specifies the API design of settings actions.

- [Add / Change / Remove Email](#add--change--remove-email)
- [Add / Change / Remove Phone](#add--change--remove-phone)
- [Add / Change / Remove Username](#add--change--remove-username)
- [Link / Unlink OAuth](#link--unlink-oauth)
- [Setup / Change Password](#setup--change-password)
- [Manage MFA](#manage-mfa)
- [Setup / Change / Remove MFA Phone](#setup--change--remove-MFA-phone)
- [Setup / Change / Remove MFA Email](#setup--change--remove-MFA-email)
- [Setup / Change / Remove MFA Password](#setup--change--remove-MFA-password)
- [Setup / Manage MFA TOTP](#setup--manage-mfa-totp)
- [Setup / View Recovery Code](#setup--view-recovery-code)
- [Display Linked OAuth Providers](#display-linked-oauth-providers)

---

- [Full UserInfo Design](#full-userinfo-design)
- [Security Considerations](#security-considerations)

## Add / Change / Remove Email

### Intention

App developers might want custom button to trigger the UI in native App which manage the user's Email.

### SDK Design

- Add Email

```typescript
await authgear.addEmail({ redirectURI: "com.example://complete" });
```

- Change Email

```typescript
await authgear.changeEmail("user@example.com", {
  redirectURI: "com.example://complete",
});
```

- Display Email

```typescript
const userInfo = await authgear.fetchUserInfo();
const email = userInfo.email;
```

## Add / Change / Remove Phone

### Intention

App developers might want custom button to trigger the UI in native App which manage the user's Phone.

### SDK Design

- Add Phone

```typescript
await authgear.addPhone({ redirectURI: "com.example://complete" });
```

- Change Phone

```typescript
await authgear.changePhone("+85212341234", {
  redirectURI: "com.example://complete",
});
```

- Display Phone

```typescript
const userInfo = await authgear.fetchUserInfo();
const phone = userInfo.phoneNumber;
```

## Add / Change / Remove Username

### Intention

App developers might want custom button to trigger the UI in native App which manage the user's Username.

### SDK Design

- Add Username

```typescript
await authgear.addUsername("example", {
  redirectURI: "com.example://complete",
});
```

- Change Username

```typescript
await authgear.changeUsername("example", {
  redirectURI: "com.example://complete",
});
```

- Display Username

```typescript
const userInfo = await authgear.fetchUserInfo();
const username = userInfo.preferredUsername;
```

## Link / Unlink OAuth

### Intention

App developers might want to offer custom buttons to trigger the UI in native App which manages the user's OAuth connections (e.g., Sign in with Google).

`oauthProviderAlias` is the alias of the OAuth provider as configured in Authgear Portal under Social / Enterprise Login. It is **required** for both link and unlink.

### SDK Design

- Link OAuth

```typescript
await authgear.linkOAuth({
  oauthProviderAlias: "google",
  redirectURI: "com.example://complete",
});
```

- Unlink OAuth

```typescript
await authgear.unlinkOAuth({
  oauthProviderAlias: "google",
  redirectURI: "com.example://complete",
});
```

## Setup / Change Password

### Intention

App developers might want a custom button to trigger the UI in native App which manages the user's password.

### SDK Design

- Display Password Status

```typescript
const userInfo = await authgear.fetchUserInfo();
const passwordEnabled = userInfo.authenticators.some(
  (a) => a.kind === "primary" && a.type === "password"
);
```

- Setup Password

```typescript
await authgear.setupPassword({ redirectURI: "com.example://complete" });
```

- Change Password

```typescript
await authgear.changePassword({ redirectURI: "com.example://complete" });
```

- Remove Password

This is not supported.

## Manage MFA

### Intention

App developers might want custom button to trigger the UI in native App which manages the user's MFAs.

### SDK Design

- Display MFA Status

```typescript
const userInfo = await authgear.fetchUserInfo();
const isMFAEnabled = userInfo.authenticators.some(
  (a) => a.kind === "secondary"
);
```

- Manage MFA

```typescript
await authgear.manageMFA({ redirectURI: "com.example://complete" });
```

## Setup / Change / Remove MFA Phone

### Intention

App developers might want to display user's MFA phone number, or add button to trigger the UI in native App which manages the user's MFA phone number.

### SDK Design

- Setup MFA Phone

```typescript
await authgear.setupMFAPhone({ redirectURI: "com.example://complete" });
```

- Change MFA Phone

```typescript
const userInfo = await authgear.fetchUserInfo();
const phone = userInfo.authenticators.find(
  (a) => a.kind === "secondary" && a.type === "oob_otp_sms"
)?.phoneNumber;
await authgear.changeMFAPhone(phone!, { redirectURI: "com.example://complete" });
```

- Remove MFA Phone

```typescript
const userInfo = await authgear.fetchUserInfo();
const phone = userInfo.authenticators.find(
  (a) => a.kind === "secondary" && a.type === "oob_otp_sms"
)?.phoneNumber;
await authgear.removeMFAPhone(phone!, { redirectURI: "com.example://complete" });
```

## Setup / Change / Remove MFA Email

### Intention

App developers might want to display user's MFA email address, or add button to trigger the UI in native App which manages the user's MFA email address.

### SDK Design

- Setup MFA Email

```typescript
await authgear.setupMFAEmail({ redirectURI: "com.example://complete" });
```

- Change MFA Email

```typescript
const userInfo = await authgear.fetchUserInfo();
const email = userInfo.authenticators.find(
  (a) => a.kind === "secondary" && a.type === "oob_otp_email"
)?.email;
await authgear.changeMFAEmail(email!, { redirectURI: "com.example://complete" });
```

- Remove MFA Email

```typescript
const userInfo = await authgear.fetchUserInfo();
const email = userInfo.authenticators.find(
  (a) => a.kind === "secondary" && a.type === "oob_otp_email"
)?.email;
await authgear.removeMFAEmail(email!, { redirectURI: "com.example://complete" });
```

## Setup / Change / Remove MFA Password

### Intention

App developers might want custom button to trigger the UI in native App which manages the user's MFA password.

### SDK Design

- Display MFA Password Status

```typescript
const userInfo = await authgear.fetchUserInfo();
const isMFAPasswordEnabled = userInfo.authenticators.some(
  (a) => a.kind === "secondary" && a.type === "password"
);
```

- Setup MFA Password

```typescript
await authgear.setupMFAPassword({ redirectURI: "com.example://complete" });
```

- Change MFA Password

```typescript
await authgear.changeMFAPassword({ redirectURI: "com.example://complete" });
```

- Remove MFA Password

```typescript
await authgear.removeMFAPassword({ redirectURI: "com.example://complete" });
```

## Setup / Manage MFA TOTP

### Intention

App developers might want to offer custom buttons to trigger the UI in native App which manages the user's MFA TOTPs.

### SDK Design

- Display MFA TOTP Status

```typescript
const userInfo = await authgear.fetchUserInfo();
const totpEnabled = userInfo.authenticators.some(
  (a) => a.type === "totp"
);
```

- Setup MFA TOTP

```typescript
await authgear.setupMFATOTP({ redirectURI: "com.example://complete" });
```

- Manage MFA TOTP

```typescript
await authgear.manageMFATOTP({ redirectURI: "com.example://complete" });
```

## Setup / View Recovery Code

### Intention

App developers might want to offer custom buttons to trigger the UI in native App which manages the user's recovery code.

### SDK Design

- Display Recovery Code Status

```typescript
const userInfo = await authgear.fetchUserInfo();
const recoveryCodeEnabled = userInfo.authenticators.some(
  (a) => a.type === "recovery_code"
);
```

- Setup Recovery Code

```typescript
await authgear.setupRecoveryCode({ redirectURI: "com.example://complete" });
```

- View existing Recovery Codes

```typescript
await authgear.viewRecoveryCode({ redirectURI: "com.example://complete" });
```

## Display Linked OAuth Providers

### Intention

App developers might want to know which OAuth providers the user has linked to their account, for example to show a "Connected accounts" screen or gate features behind a specific provider being linked.

### SDK Design

- Display Linked OAuth Providers

```typescript
const userInfo = await authgear.fetchUserInfo();
const linkedOAuthProviders = userInfo.identities
  .filter((i) => i.type === "oauth")
  .map((i) => i.providerAlias);
```

- Check whether a specific provider is linked

```typescript
const userInfo = await authgear.fetchUserInfo();
const isGoogleLinked = userInfo.identities.some(
  (i) => i.type === "oauth" && i.providerAlias === "google"
);
```

## Full UserInfo Design

- SDK Object

```typescript
interface Authenticator {
  kind: "primary" | "secondary";
  type: "password" | "passkey" | "totp" | "oob_otp_email" | "oob_otp_sms";
  createdAt: Date;
  updatedAt: Date;
  email?: string;
  phoneNumber?: string;
}

interface Identity {
  type: "login_id" | "oauth" | "anonymous" | "biometric" | "passkey" | "siwe" | "ldap";
  createdAt: Date;
  updatedAt: Date;
  loginIDKey?: string; // Present when type is "login_id", e.g. "email", "phone", "username"
  loginIDType?: "email" | "phone" | "username"; // Present when type is "login_id"
  providerType?: string; // Present when type is "oauth", e.g. "google", "facebook"
  providerAlias?: string; // Present when type is "oauth"
}

interface UserInfo {
  sub: string;
  email: string;
  phoneNumber: string;
  preferredUsername: string;
  authenticators: []Authenticator;
  identities: []Identity;
  recoveryCodeEnabled: boolean;
}
```

- OIDC userinfo endpoint

```jsonc
{
  "sub": "00000000-0000-0000-0000-000000000000",
  "email": "user@example.com",
  "phone_number": "+85211111111",
  "preferred_username": "example",
  "https://authgear.com/claims/user/authenticators": [
    {
      "kind": "primary",
      "type": "password",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    },
    {
      "kind": "secondary",
      "type": "oob_otp_email",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z",
      "email": "oob_otp_email@example.com"
    },
    {
      "kind": "secondary",
      "type": "oob_otp_sms",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z",
      "phone_number": "+85212345678"
    },
    {
      "kind": "secondary",
      "type": "totp",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "https://authgear.com/claims/user/identities": [
    {
      "type": "oauth",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z",
      "provider_type": "google",
      "provider_alias": "google"
    },
    {
      "type": "login_id",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z",
      "login_id_key": "email",
      "login_id_type": "email"
    }
  ],
  "https://authgear.com/claims/user/recovery_code_enabled": true
}
```

## Security Considerations

- We will expose user's password status together with MFA emails and phone numbers in userinfo endpoint, therefore client apps will be able to know them. If the client app is malicious, they may use the information to attack an authgear user.
- We expose which identity types and OAuth provider aliases are linked via `https://authgear.com/claims/user/identities`. A malicious client app could use this to infer which providers a user has accounts with.
- We can hide these fields in Third-Party Clients (by checking scope `https://authgear.com/scopes/full-access`) to mitigate the risk.
