# Settings Action: Connect OAuth — JS Web SDK Implementation Plan

> Source repo: `authgear-sdk-js`. Companion plan: `connect-oauth-server.md`.
> Platform: **authgear-web**.

## Goal

Expose `startLinkOAuth(options)` / `finishLinkOAuth()` on the public `WebContainer`, mirroring the pattern of every other settings-action pair (e.g. `startAddEmail` / `finishAddEmail`).

## Public API (target)

`oauthProviderAlias` is **required**. The server rejects requests without it.

```ts
// Start — redirects the browser to Authgear, which redirects to the OAuth provider
await authgear.startLinkOAuth({
  redirectURI: "https://myapp.com/oauth-callback",
  oauthProviderAlias: "google",   // required
});

// Finish — called on the redirectURI page after the round-trip
await authgear.finishLinkOAuth();
```

---

## File-by-file changes

### 1. `packages/authgear-core/src/types.ts`

**1a. Extend the `SettingsAction` enum.** Add a new member after `ChangeUsername` (the current last entry):

```ts
/**
 * Connect an OAuth provider in Authgear settings page.
 */
LinkOAuth = "link_oauth",
```

**1b. Extend the `xSettingsAction` union in `_OIDCAuthenticationRequest`.** Append `"link_oauth"`:

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
  | "link_oauth";
```

`oauthProviderAlias?: string` is **already** on `_OIDCAuthenticationRequest` — no new field needed. `startSettingsAction` spreads `...options` into `authorizeEndpoint`, so adding `oauthProviderAlias` to `_InternalSettingsActionOptions` (step 2) is sufficient to forward it.

---

### 2. `packages/authgear-web/src/types.ts`

**2a. Add `LinkOAuthOptions`.** Place after `ChangeUsernameOptions`:

```ts
/**
 * Options for connecting an OAuth provider via settings action.
 * @public
 */
export interface LinkOAuthOptions extends SettingsActionOptions {
  /**
   * The alias of the OAuth provider to link,
   * as configured in Authgear Portal under Social / Enterprise Login.
   * This field is required.
   */
  oauthProviderAlias: string;
}
```

**2b. Add `oauthProviderAlias?` to `_InternalSettingsActionOptions`:**

```ts
export interface _InternalSettingsActionOptions extends SettingsActionOptions {
  qLoginID?: string;
  oauthProviderAlias?: string;
}
```

---

### 3. `packages/authgear-web/src/container.ts`

No change to `startSettingsAction` — it already spreads `...options` into `authorizeEndpoint`, so `oauthProviderAlias` flows through automatically once it is on `_InternalSettingsActionOptions`.

**3a. Add `startLinkOAuth`.** Place after `startChangeUsername`:

```ts
/**
 * Start settings action "link_oauth" by redirecting to the authorization endpoint.
 * @public
 */
async startLinkOAuth(options: LinkOAuthOptions): Promise<void> {
  await this.startSettingsAction(SettingsAction.LinkOAuth, options);
}
```

**3b. Add `finishLinkOAuth`.** Place after `finishChangeUsername`:

```ts
/**
 * Finish settings action "link_oauth".
 * @public
 */
async finishLinkOAuth(): Promise<void> {
  return this.finishSettingsAction();
}
```

`finishSettingsAction` delegates to `this.baseContainer._finishSettingsAction(window.location.href)`, identical to all other finish methods except `finishDeleteAccount`.

---

### 4. Public exports

Add `LinkOAuthOptions` to `packages/authgear-web/src/index.ts` if types are explicitly re-exported there. `WebContainer` is already re-exported so the new methods come along automatically.

---

## Verification

- **Type-check:** `npx tsc --noEmit` in `packages/authgear-core` and `packages/authgear-web`.
- **Build:** `npm run build` in `packages/authgear-web`.
- **API Extractor:** new `startLinkOAuth`, `finishLinkOAuth`, and `LinkOAuthOptions` should appear in the public API report.
- **Manual:** complete the round trip against a local `authgear-server` with the server-side changes applied.

---

## Checklist

- [ ] `SettingsAction.LinkOAuth` in `authgear-core/src/types.ts`
- [ ] `"link_oauth"` in the `xSettingsAction` union
- [ ] `LinkOAuthOptions` in `authgear-web/src/types.ts`
- [ ] `oauthProviderAlias?` on `_InternalSettingsActionOptions`
- [ ] `startLinkOAuth` / `finishLinkOAuth` on `WebContainer`
- [ ] `LinkOAuthOptions` exported from the package entry point
- [ ] TypeScript build / typecheck pass
- [ ] API Extractor report regenerated and reviewed
