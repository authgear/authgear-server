# Implementation Plan: `https://authgear.com/claims/user/identities` — JS SDK

## 1. Goal / Scope

Add the `identities` field to the `UserInfo` type in the JS SDK so that callers of `fetchUserInfo()` can read which identities are linked to the user, for login ID identities which key was used, and for OAuth identities which provider alias they used.

This mirrors how `authenticators` was added. All changes are in `packages/authgear-core/src/types.ts` and its test file. The three other packages (`authgear-web`, `authgear-react-native`, `authgear-capacitor`) import `UserInfo` from core and require no changes.

Spec: `docs/specs/sdk-settings-actions.md` (Full UserInfo Design, Display Linked OAuth Providers).

---

## 2. Type Changes

### File: `packages/authgear-core/src/types.ts`

#### Add `IdentityType` enum

Place after the existing `AuthenticatorKind` enum (before `Authenticator` interface):

```typescript
/**
 * @public
 */
export enum IdentityType {
  LoginID = "login_id",
  OAuth = "oauth",
  Anonymous = "anonymous",
  Biometric = "biometric",
  Passkey = "passkey",
  SIWE = "siwe",
  LDAP = "ldap",
  Unknown = "unknown",
}
```

`Unknown` follows the same defensive pattern used by `AuthenticatorType` — unknown values from the server decode gracefully.

#### Add `LoginIDType` enum

Place after `IdentityType`:

```typescript
/**
 * @public
 */
export enum LoginIDType {
  Email = "email",
  Phone = "phone",
  Username = "username",
  Unknown = "unknown",
}
```

`Unknown` follows the same pattern — if the server introduces a new login ID type, the SDK decodes it gracefully rather than silently dropping or corrupting the value.

#### Add `Identity` interface

Place after the `LoginIDType` enum:

```typescript
/**
 * @public
 */
export interface Identity {
  type: IdentityType;
  createdAt: Date;
  updatedAt: Date;
  loginIDKey?: string;
  loginIDType?: LoginIDType;
  providerType?: string;
  providerAlias?: string;
}
```

`createdAt` and `updatedAt` are always present, decoded from the `"created_at"` and `"updated_at"` RFC 3339 strings in the JSON response — matching the pattern used by `Authenticator`.

`loginIDKey` is present only when `type` is `IdentityType.LoginID` (e.g. `"email"`, `"phone"`, `"username"`).

`loginIDType` is present only when `type` is `IdentityType.LoginID`. It is one of `LoginIDType.Email`, `LoginIDType.Phone`, `LoginIDType.Username`, or `LoginIDType.Unknown`.

`providerType` is present only when `type` is `IdentityType.OAuth` (e.g. `"google"`, `"facebook"`).

`providerAlias` is present only when `type` is `IdentityType.OAuth`.

#### Extend `UserInfo` interface

Add `identities` after `authenticators`:

```typescript
  authenticators?: Authenticator[];
  identities?: Identity[];
```

#### Add `parseIdentityType` and `parseLoginIDType` functions

Place after the existing `parseAuthenticatorKind` function:

```typescript
/**
 * @internal
 */
export function parseIdentityType(value: string): IdentityType {
  switch (value) {
    case "login_id":
      return IdentityType.LoginID;
    case "oauth":
      return IdentityType.OAuth;
    case "anonymous":
      return IdentityType.Anonymous;
    case "biometric":
      return IdentityType.Biometric;
    case "passkey":
      return IdentityType.Passkey;
    case "siwe":
      return IdentityType.SIWE;
    case "ldap":
      return IdentityType.LDAP;
    default:
      return IdentityType.Unknown;
  }
}

/**
 * @internal
 */
export function parseLoginIDType(value: string): LoginIDType {
  switch (value) {
    case "email":
      return LoginIDType.Email;
    case "phone":
      return LoginIDType.Phone;
    case "username":
      return LoginIDType.Username;
    default:
      return LoginIDType.Unknown;
  }
}
```

#### Add `_decodeIdentities` function

Place after `_decodeAuthenticators`:

```typescript
/**
 * @internal
 */
export function _decodeIdentities(r: any): Identity[] | undefined {
  if (!Array.isArray(r)) {
    return undefined;
  }
  return r.map((i) => {
    const identity: Identity = {
      type: parseIdentityType(i["type"]),
      createdAt: new Date(i["created_at"]),
      updatedAt: new Date(i["updated_at"]),
    };
    if (identity.type === IdentityType.LoginID) {
      identity.loginIDKey = i["login_id_key"];
      identity.loginIDType =
        i["login_id_type"] != null
          ? parseLoginIDType(i["login_id_type"])
          : undefined;
    }
    if (identity.type === IdentityType.OAuth) {
      identity.providerType = i["provider_type"];
      identity.providerAlias = i["provider_alias"];
    }
    return identity;
  });
}
```

#### Extend `_decodeUserInfo`

Add after the `authenticators` line:

```typescript
    authenticators: _decodeAuthenticators(
      r["https://authgear.com/claims/user/authenticators"]
    ),
    identities: _decodeIdentities(
      r["https://authgear.com/claims/user/identities"]
    ),
```

---

## 3. Test Changes

### File: `packages/authgear-core/src/types.test.ts`

#### Update import

Add `IdentityType` to the import:

```typescript
import {
  _decodeUserInfo,
  AuthenticatorType,
  AuthenticatorKind,
  IdentityType,
  LoginIDType,
} from "./types";
```

#### Extend `USER_INFO` fixture

Add `"https://authgear.com/claims/user/identities"` after the authenticators array in the JSON string:

```json
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
  },
  {
    "type": "unknown_future_type",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
],
```

#### Extend `expected` object in the test

Add `identities` after `authenticators`:

```typescript
      identities: [
        {
          type: IdentityType.OAuth,
          createdAt: new Date("2024-01-01T00:00:00Z"),
          updatedAt: new Date("2024-01-01T00:00:00Z"),
          providerType: "google",
          providerAlias: "google",
        },
        {
          type: IdentityType.LoginID,
          createdAt: new Date("2024-01-01T00:00:00Z"),
          updatedAt: new Date("2024-01-01T00:00:00Z"),
          loginIDKey: "email",
          loginIDType: LoginIDType.Email,
        },
        {
          type: IdentityType.Unknown,
          createdAt: new Date("2024-01-01T00:00:00Z"),
          updatedAt: new Date("2024-01-01T00:00:00Z"),
        },
      ],
```

The third case (`"unknown_future_type"`) verifies the `Unknown` fallback.

#### Extend `raw` in the `expected` object

Add the raw key to match the fixture (the `raw` field in `UserInfo` is the unmodified input):

```typescript
        "https://authgear.com/claims/user/identities": [
          { type: "oauth", provider_type: "google", provider_alias: "google" },
          { type: "login_id", login_id_key: "email", login_id_type: "email" },
          { type: "unknown_future_type" },
        ],
```

---

## 4. File-Level Change Summary

| File | Change |
|---|---|
| `packages/authgear-core/src/types.ts` | Add `IdentityType` enum, `LoginIDType` enum, `Identity` interface, `parseIdentityType`, `parseLoginIDType`, `_decodeIdentities`; extend `UserInfo` and `_decodeUserInfo` |
| `packages/authgear-core/src/types.test.ts` | Extend import, fixture JSON, and expected object |

No changes to `authgear-web`, `authgear-react-native`, or `authgear-capacitor` — they re-export `UserInfo` from core unchanged.

---

## 5. Verification

```
cd packages/authgear-core
npx jest src/types.test.ts
```

Also run the TypeScript check across the repo:

```
cd /path/to/authgear-sdk-js
npm run typecheck
```

---

## 6. Atomic Commit Plan

### Commit 1 — Types and decoder
**File:** `packages/authgear-core/src/types.ts`

Add `IdentityType` enum, `Identity` interface, `parseIdentityType`, `_decodeIdentities`, extend `UserInfo.identities`, extend `_decodeUserInfo`.

### Commit 2 — Tests
**File:** `packages/authgear-core/src/types.test.ts`

Extend import, fixture, and expected object. Tests must pass.
