# SDK - Settings Actions

> [!WARNING]
> **ARCHIVED DOCUMENT: For Historical Record Only.** This design was **not** adopted. Do **not** follow these specifications for current work. The **approved design** is documented in [here](./sdk-settings-actions.md).

This document specifies the API design of settings actions.

- [Add / Change / Remove Email](#add--change--remove-email)
- [Add / Change / Remove Phone](#add--change--remove-phone)
- [Add / Change / Remove Username](#add--change--remove-username)
- [Setup / Change Password](#setup--change-password)
- [Manage MFA](#manage-mfa)
- [Setup / Change / Remove MFA Phone](#setup--change--remove-MFA-phone)
- [Setup / Change / Remove MFA Email](#setup--change--remove-MFA-email)
- [Setup / Change / Remove MFA Password](#setup--change--remove-MFA-password)
- [Setup / Manage MFA TOTP](#setup--manage-mfa-totp)
- [Setup / View Recovery Code](#setup--view-recovery-code)

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

## Setup / Change Password

### Intention

App developers might want a custom button to trigger the UI in native App which manages the user's password.

### SDK Design

- Display Password Status

```typescript
const userInfo = await authgear.fetchUserInfo();
const passwordEnabled = userInfo.passwordEnabled;
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
const isMFAEnabled = userInfo.mfa.enabled;
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
const phone = userInfo.mfa.phoneNumbers[0];
await authgear.changeMFAPhone(phone, { redirectURI: "com.example://complete" });
```

- Remove MFA Phone

```typescript
const userInfo = await authgear.fetchUserInfo();
const phone = userInfo.mfa.phoneNumbers[0];
await authgear.removeMFAPhone(phone, { redirectURI: "com.example://complete" });
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
const email = userInfo.mfa.emails[0];
await authgear.changeMFAEmail(email, { redirectURI: "com.example://complete" });
```

- Remove MFA Email

```typescript
const userInfo = await authgear.fetchUserInfo();
const email = userInfo.mfa.emails[0];
await authgear.removeMFAEmail(email, { redirectURI: "com.example://complete" });
```

## Setup / Change / Remove MFA Password

### Intention

App developers might want custom button to trigger the UI in native App which manages the user's MFA password.

### SDK Design

- Display MFA Password Status

```typescript
const userInfo = await authgear.fetchUserInfo();
const isMFAPasswordEnabled = userInfo.mfa.passwordEnabled;
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
const totpEnabled = userInfo.mfa.totpEnabled;
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
const recoveryCodeEnabled = userInfo.mfa.recoveryCodeEnabled;
```

- Setup MFA Recovery Code

```typescript
await authgear.setupRecoveryCode({ redirectURI: "com.example://complete" });
```

- View existing Recovery Codes

```typescript
await authgear.viewRecoveryCode({ redirectURI: "com.example://complete" });
```

## Full UserInfo Design

- SDK Object

```typescript
interface UserInfo {
  sub: string;
  email: string;
  phoneNumber: string;
  preferredUsername: string;
  passwordEnabled: boolean;
  mfa: {
    enabled: boolean;
    emails: []string;
    phoneNumbers: []string;
    passwordEnabled: boolean;
    totpEnabled: boolean;
    recoveryCodeEnabled: boolean;
  }
}
```

- OIDC userinfo endpoint

```jsonc
{
  "sub": "00000000-0000-0000-0000-000000000000",
  "email": "user@example.com",
  "phone_number": "+85211111111",
  "preferred_username": "example",
  "https://authgear.com/claims/user/password_enabled": true,
  "https://authgear.com/claims/user/mfa_enabled": true,
  "https://authgear.com/claims/user/mfa_emails": ["mfa@example.com"],
  "https://authgear.com/claims/user/mfa_phone_numbers": ["+85212222222"],
  "https://authgear.com/claims/user/mfa_password_enabled": true,
  "https://authgear.com/claims/user/mfa_totp_enabled": true,
  "https://authgear.com/claims/user/recovery_code_enabled": true
}
```

## Security Considerations

- We will expose user's password status together with MFA emails and phone numbers in userinfo endpoint, therefore client apps will be able to know them. If the client app is malicious, they may use the information to attack an authgear user.
- We can hide these fields in Third-Party Clients (by checking the scope `https://authgear.com/scopes/full-access`) to mitigate the risk.
