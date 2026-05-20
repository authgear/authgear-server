# Settings Action: Unlink OAuth — JS Web SDK Implementation Plan

> Source repo: `authgear-sdk-js`. Companion plan: `unlink-oauth-sdk-server.md`.
> Platform: **authgear-web**.

## Goal

Expose `startUnlinkOAuth(options)` / `finishUnlinkOAuth()` on the public `WebContainer`, mirroring
the pattern of every other settings-action pair (e.g. `startLinkOAuth` / `finishLinkOAuth`).

## Public API (target)

`oauthProviderAlias` is **required**. The server rejects requests without it.

```ts
// Start — redirects the browser to Authgear settings page
await authgear.startUnlinkOAuth({
  redirectURI: "https://myapp.com/oauth-callback",
  oauthProviderAlias: "google",   // required
});

// Finish — called on the redirectURI page after the user disconnects
await authgear.finishUnlinkOAuth();
```

---

## File-by-file changes

### 1. `packages/authgear-core/src/types.ts`

**1a. Extend the `SettingsAction` enum.** Add after `LinkOAuth`:

```ts
/**
 * Unlink an OAuth provider in Authgear settings page.
 */
UnlinkOAuth = "unlink_oauth",
```

**1b. Extend the `xSettingsAction` union in `_OIDCAuthenticationRequest`.** Append `"unlink_oauth"`:

```ts
xSettingsAction?:
  | "change_password"
  | "delete_account"
  | "add_email"
  | "add_phone"
  | "add_username"
  | "change_email"
  | "change_phone"
  | "change_username"
  | "link_oauth"
  | "unlink_oauth";
```

`oauthProviderAlias?: string` is already on `_OIDCAuthenticationRequest` — no new field needed.

---

### 2. `packages/authgear-web/src/types.ts`

**2a. Add `UnlinkOAuthOptions`.** Place after `LinkOAuthOptions`:

```ts
/**
 * Options for disconnecting an OAuth provider via settings action.
 * @public
 */
export interface UnlinkOAuthOptions extends SettingsActionOptions {
  /**
   * The alias of the OAuth provider to unlink,
   * as configured in Authgear Portal under Social / Enterprise Login.
   * This field is required.
   */
  oauthProviderAlias: string;
}
```

`_InternalSettingsActionOptions` already has `oauthProviderAlias?: string` — no change needed.

---

### 3. `packages/authgear-web/src/container.ts`

No change to `startSettingsAction` — it already spreads `...options` into `authorizeEndpoint`,
so `oauthProviderAlias` flows through automatically.

**3a. Add `startUnlinkOAuth`.** Place after `startLinkOAuth`:

```ts
/**
 * Start settings action "unlink_oauth" by redirecting to the settings page.
 * @public
 */
async startUnlinkOAuth(options: UnlinkOAuthOptions): Promise<void> {
  await this.startSettingsAction(SettingsAction.UnlinkOAuth, options);
}
```

**3b. Add `finishUnlinkOAuth`.** Place after `finishLinkOAuth`:

```ts
/**
 * Finish settings action "unlink_oauth".
 * @public
 */
async finishUnlinkOAuth(): Promise<void> {
  return this.finishSettingsAction();
}
```

`finishSettingsAction` delegates to `this.baseContainer._finishSettingsAction(window.location.href)`,
identical to all other finish methods except `finishDeleteAccount`.

---

### 4. Public exports

Add `UnlinkOAuthOptions` to `packages/authgear-web/src/index.ts` alongside `LinkOAuthOptions`.
`WebContainer` is already re-exported so the new methods come along automatically.

---

## Verification

- **Type-check:** `npx tsc --noEmit` in `packages/authgear-core` and `packages/authgear-web`.
- **Build:** `npm run build` in `packages/authgear-web`.
- **API Extractor:** new `startUnlinkOAuth`, `finishUnlinkOAuth`, and `UnlinkOAuthOptions` should
  appear in the public API report.
- **Manual:** complete the round trip against a local `authgear-server` with the server-side
  changes applied.

---

## Checklist

- [ ] `SettingsAction.UnlinkOAuth` in `authgear-core/src/types.ts`
- [ ] `"unlink_oauth"` in the `xSettingsAction` union
- [ ] `UnlinkOAuthOptions` in `authgear-web/src/types.ts`
- [ ] `startUnlinkOAuth` / `finishUnlinkOAuth` on `WebContainer`
- [ ] `UnlinkOAuthOptions` exported from the package entry point
- [ ] TypeScript build / typecheck pass
- [ ] API Extractor report regenerated and reviewed
