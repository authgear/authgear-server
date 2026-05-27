# Implementation Plan: `https://authgear.com/claims/user/identities` — JS SDK

## 1. Goal / Scope

Add the `identities` field to the `UserInfo` type in the JS SDK so that callers of `fetchUserInfo()` can read which identities are linked to the user and, for OAuth identities, which provider alias they used.

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

#### Add `Identity` interface

Place after the `IdentityType` enum:

```typescript
/**
 * @public
 */
export interface Identity {
  type: IdentityType;
  providerAlias?: string;
}
```

`providerAlias` is present only when `type` is `IdentityType.OAuth`.

#### Extend `UserInfo` interface

Add `identities` after `authenticators`:

```typescript
  authenticators?: Authenticator[];
  identities?: Identity[];
```

#### Add `parseIdentityType` function

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
    };
    if (identity.type === IdentityType.OAuth) {
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
} from "./types";
```

#### Extend `USER_INFO` fixture

Add `"https://authgear.com/claims/user/identities"` after the authenticators array in the JSON string:

```json
"https://authgear.com/claims/user/identities": [
  {
    "type": "oauth",
    "provider_alias": "google"
  },
  {
    "type": "login_id"
  },
  {
    "type": "unknown_future_type"
  }
],
```

#### Extend `expected` object in the test

Add `identities` after `authenticators`:

```typescript
      identities: [
        {
          type: IdentityType.OAuth,
          providerAlias: "google",
        },
        {
          type: IdentityType.LoginID,
        },
        {
          type: IdentityType.Unknown,
        },
      ],
```

The third case (`"unknown_future_type"`) verifies the `Unknown` fallback.

#### Extend `raw` in the `expected` object

Add the raw key to match the fixture (the `raw` field in `UserInfo` is the unmodified input):

```typescript
        "https://authgear.com/claims/user/identities": [
          { type: "oauth", provider_alias: "google" },
          { type: "login_id" },
          { type: "unknown_future_type" },
        ],
```

---

## 4. File-Level Change Summary

| File | Change |
|---|---|
| `packages/authgear-core/src/types.ts` | Add `IdentityType` enum, `Identity` interface, `parseIdentityType`, `_decodeIdentities`; extend `UserInfo` and `_decodeUserInfo` |
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
