# Passkey check in the SDK WebView — part 1: the bridge

Specs: [docs/specs/webview-bridge.md](../../specs/webview-bridge.md) (the bridge), [docs/specs/passkey-webview.md](../../specs/passkey-webview.md#passkey-on-page-load) (the passkey on page load).

Parts: **1 bridge** · [2 AuthUI](./2026-11-10-02-webview-check-authui.md) · [3 iOS SDK](./2026-11-10-03-webview-check-ios-sdk.md) · [4 Android SDK](./2026-11-10-04-webview-check-android-sdk.md)

The bridge's contract is `webview-bridge.md`. This part records why it is designed this way, and what AuthUI does with it. Parts 2 to 4 implement it. Where this part and the spec differ, the spec wins.

## 1. Goal

When the login page loads inside the SDK's own WebView, the app checks this device for a passkey with the OS's "only passkeys already on this device" request.

- The device has a passkey for the site: the OS passkey sheet appears with no tap. Choosing a passkey signs the user in.
- The device has none: nothing appears. No QR code, no security key dialog.

A web page can't do this itself: `navigator.credentials.get` either needs a tap or waits for the user to open the keyboard. So AuthUI asks the app through the bridge, and the app answers with the passkey assertion. AuthUI then submits the assertion as if it came from the browser, so the sign-in is still driven by the page.

## 2. Design rules

1. **The SDK is a thin mechanism; the page holds the policy.** Apps keep an SDK version for years, and the server is deployed weekly. The SDK only answers messages. When to ask, how often, and what to do with the answer are decided by the page.
2. **One general bridge, typed messages.** The bridge is not passkey-only: later features add a message type, or a feature of an existing type, never another bridge. Possible later types: `credentials.get` with a `password` member (passkey or password in one sheet), `credentials.create` (native passkey creation, automatic upgrade), `credentials.store` (save a password), `otp.receive` (SMS code on Android), `webview.close`.
3. **Every request gets exactly one reply**, including messages the SDK doesn't support (`unsupported`). This is what lets page and SDK versions mix freely.
4. **Capabilities are a flat list of strings** (`<type>` or `<type>:<feature>`), answered by every SDK from the first version. A page checks before using a type or feature, because some requests can't be tried without showing UI. Designs compared: Telegram Mini Apps (one host version: unusable with several independently released SDKs), Microsoft Teams JS (`isSupported()` per capability; do not gate on host name), WebAuthn `getClientCapabilities()` (flat names), LSP and MCP (capability objects at start, unknown keys ignored), EIP-5792 `wallet_getCapabilities`. The flat list is the simplest to implement in every SDK and needs no restructuring as features are added.
5. **`credentials.get` follows `navigator.credentials.get()`:** each kind of credential is a member of one request, and `credential.type` tells which kind came back. A combined passkey and password sheet is then one request, as it is natively on both platforms.
6. **No SDK name or version in replies.** Each wrapper SDK (Flutter, React Native, Capacitor) has its own WebView code and no version constant, and pages must not decide anything from them.

## 3. What AuthUI does with it

1. On the login page (and the signup page when it is shown as the combined signup-login page), if the bridge exists, AuthUI asks `bridge.capabilities`. If it lists `credentials.get`, `credentials.get:publicKey` and `credentials.get:immediate`, AuthUI sends one `credentials.get` with `mediation: "immediate"` and the page's `publicKey`.
2. `ok` with a `public-key` credential: submit the assertion. Anything else: do nothing visible.
3. **Once per WebView.** AuthUI records in `sessionStorage` that it asked, before sending. Later loads of the login page in the same WebView (back navigation, an error after a rejected passkey, reload) don't ask again. A new WebView has a new `sessionStorage`, so the next sign-in asks again.
4. **No passkey autofill when the bridge exists.** On iOS, autofill in a `WKWebView` keeps a passkey request pending for the app. iOS allows one passkey request per app, and that pending request was seen to outlive the WebView, after which every native request failed at once with `ASAuthorizationError` 1004 "Request already in progress" until the app was force-quit (device test, 2026-11-02). Turning autofill off in the SDK WebView removed the errors. Android WebView has no autofill anyway. The check is the presence of the bridge object, so that AutoFill can start without waiting for a reply.
5. The "Continue with passkey" button is unchanged. It still uses `navigator.credentials.get` in the WebView.

## 4. Security

- The page that may use the bridge is fixed by origin and main-frame checks in the SDK, not by the page's claims. Without them, any page shown in the WebView, such as a social login provider's page or one reached through an open redirect, could ask for an assertion for the project's domain and relay it to sign in as the user.
- Only the endpoint's origin is allowed in this release. Allowing a Custom UI (`bridgeAllowedOrigins`, in the spec's future works) is left out because nothing in this release needs it, it would be public API on every SDK before a Custom UI has tried it, and a wrongly listed origin would be its main risk. Adding it later changes no message.
- When the top-level page starts loading another page, the SDK cancels pending requests. iOS allows one passkey request per app, so a request left over from a page that is gone would block the next page's request, and the sheet would stay over a page that never asked. It also makes sure a late reply never reaches the new page.
- The SDK doesn't check `rpId` itself. The OS only lets an app request passkeys of a domain whose association file lists the app. That is the existing `x_native_apps` setup, so the server needs no new configuration.
- Server-side verification is unchanged: iOS reports the origin `https://<rpId>`, which is the public origin; Android reports `android:apk-key-hash:<hash>` and `androidPackageName`, which the server already accepts for apps in `x_native_apps`.

## 5. Compatibility

| SDK | Server | Result |
|---|---|---|
| Old (no bridge) | New | As today: no check on load; autofill as in a browser |
| New | Old (AuthUI never sends) | As today |
| New | New | Check on load; autofill off in the SDK WebView |

Old SDK versions with autofill in `WKWebView` were tested on 2026-11-11: "Continue with passkey" kept opening after the WebView was opened and closed repeatedly.

Other SDKs with their own WebView (Flutter, React Native, Capacitor) implement the same bridge later. They are out of scope here.

## 6. Requirements that already exist

- The project is on a custom domain and lists the app in `x_native_apps`.
- iOS app: Associated Domains has `webcredentials:<host>`. iOS 16+ for `credentials.get`.
- Android app: nothing in the app; Google Play services on the device.

Where these are missing, the native request fails, the SDK replies `error` or `unsupported`, and the user sees nothing.

## 7. Fixed decisions

- The bridge is general (`authgearBridge` with typed messages), specified on its own in `webview-bridge.md`.
- Capabilities are a flat list of strings; no SDK name or version.
- The passkey request is `credentials.get` with a `publicKey` member, not a passkey-only type.
- The first SDK versions support `bridge.capabilities` and `credentials.get` with `publicKey` and `immediate` only, for the endpoint's origin only.
- AuthUI enforces "once per WebView" with `sessionStorage`. The SDK does not refuse a second request.
- AutoFill is off in a WebView that has the bridge.
- On iOS, a closed sheet and "no passkey" are both `no_credential`, because iOS reports them the same way.

## 8. Not in scope

- Other message types and `credentials.get` members, such as `password`.
- ASWebAuthenticationSession and Custom Tabs: they have no bridge.
- A Custom UI in the SDK's WebView: allowing its origin (`bridgeAllowedOrigins`) is future work.
