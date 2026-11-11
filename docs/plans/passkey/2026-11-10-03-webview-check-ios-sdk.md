# Passkey check in the SDK WebView — part 3: iOS SDK

Specs: [docs/specs/webview-bridge.md](../../specs/webview-bridge.md), [docs/specs/passkey-webview.md](../../specs/passkey-webview.md). Repository: `authgear-sdk-ios`.

Parts: [1 bridge](./2026-11-10-01-webview-check-bridge.md) · [2 AuthUI](./2026-11-10-02-webview-check-authui.md) · **3 iOS SDK** · [4 Android SDK](./2026-11-10-04-webview-check-android-sdk.md)

## 1. Goal and scope

`WKWebViewUIImplementation` provides the bridge of [part 1](./2026-11-10-01-webview-check-bridge.md) on iOS 16+, and answers `credentials.get` with the platform passkey API.

Files:

- `Sources/AGWKWebViewBridge.swift` (new)
- `Sources/AGWKWebViewController.swift`
- `Authgear.xcodeproj/project.pbxproj`
- `Tests/AuthgearTests/`
- `CHANGELOG.md`

`Sources/Base64.swift` already has the base64url helpers (`base64urlEncodedString()`, `base64urlToBase64(base64url:)`).

## 2. Runtime

### 2.1 New file `Sources/AGWKWebViewBridge.swift`

`@available(iOS 16.0, *) class AGWKWebViewBridge: NSObject, WKScriptMessageHandlerWithReply, ASAuthorizationControllerDelegate, ASAuthorizationControllerPresentationContextProviding`

- `static let name = "authgearBridge"`.
- `init(allowedOrigins: [URL])`: in this release, only the authorization URL the controller was created with. Only the scheme, host and port of each URL are used. It is a list so that allowing a Custom UI later (future works) only adds entries.
- `var presentationAnchorProvider: (() -> ASPresentationAnchor?)?`, set by the view controller to return `self?.view.window`.
- `func invalidate()`: cancels a pending request (see "Cleanup").

`userContentController(_:didReceive:replyHandler:)`:

1. Reject unless `message.frameInfo.isMainFrame` and the origin matches one of `allowedOrigins` (`protocol`, `host`, and port; `WKSecurityOrigin.port` is 0 for the scheme's default port, so compare with the default port of the scheme filled in on both sides). Reject with `replyHandler(nil, "authgearBridge is not available to this page")`.
2. Parse `message.body` as a string holding a JSON object with a string `id`. Otherwise reply `{"status":"error","message":"invalid message"}`.
3. Switch on `type`:
   - `"bridge.capabilities"`: reply `{"id":..., "status":"ok", "capabilities": [...]}`. The list is a constant of the bridge: `["credentials.get", "credentials.get:publicKey", "credentials.get:immediate"]`. It is `[]` until `credentials.get` is added (commit 2).
   - `"credentials.get"`: see below.
   - anything else: reply `{"id":..., "status":"unsupported"}`.
4. Every path calls `replyHandler` exactly once, with a JSON string and `nil` error.

`credentials.get`:

1. `mediation` must be `"immediate"`, else `unsupported`. `publicKey` must be present, else `unsupported` (it is the only member this version supports; other members, such as `password`, are ignored).
2. If a request is pending, reply `error` with message `"a passkey request is in progress"`. Don't touch the pending one.
3. Read `publicKey`: `rpId` (string), `challenge` (base64url), optional `userVerification`, optional `allowCredentials`. Missing `rpId` or `challenge`: `error`.
4. Build the request:
   - `ASAuthorizationPlatformPublicKeyCredentialProvider(relyingPartyIdentifier: rpId).createCredentialAssertionRequest(challenge:)`;
   - `userVerificationPreference` from `userVerification` when present;
   - `allowedCredentials` from `allowCredentials[].id` (base64url) when the array is not empty. The login page sends it empty.
   - Ignore `timeout` and `extensions`.
5. If `presentationAnchorProvider` returns nil, reply `error` ("no window") instead of presenting on an empty anchor.
6. `ASAuthorizationController(authorizationRequests: [request])`, set the delegate and presentation context provider, keep the controller and the pending `id` and reply handler, then `performRequests(options: .preferImmediatelyAvailableCredentials)`.

Results:

| Callback | Reply |
|---|---|
| `didCompleteWithAuthorization` with `ASAuthorizationPlatformPublicKeyCredentialAssertion` | `status: "ok"` and `credential` (shape in `webview-bridge.md`, *credentials.get*): `id` and `rawId` from `credentialID`, `clientDataJSON` from `rawClientDataJSON`, `authenticatorData` from `rawAuthenticatorData`, `signature`, `userHandle` from `userID`; all base64url without padding; `type: "public-key"`, `authenticatorAttachment: "platform"`, `clientExtensionResults: {}` |
| Any other credential type | `error` |
| `didCompleteWithError`, code `ASAuthorizationError.canceled` (1001) | `no_credential`. With `.preferImmediatelyAvailableCredentials`, "no passkey on this device" and "the user closed the sheet" both arrive as 1001 |
| `didCompleteWithError`, any other code | `error`, message with the domain and code (for example 1004 = a request is already in progress, or the app is not associated with the domain) |

Keep the parsing and encoding as small pure functions (`parse request text`, `encode reply`, `origin check`, `assertion fields to credential dictionary`) so they can be unit tested without a `WKWebView`.

Keep the pending request (its `id`, reply handler and `ASAuthorizationController`) in a small type, `AGWKWebViewBridgePendingRequest`, with `finish(_ reply:)` and `cancel()`. Each replies at most once: `cancel()` replies `error` with message `"cancelled"` and calls `controller.cancel()`; a `finish` after `cancel()` does nothing. The delegate callbacks and `invalidate()` go through it, so the "reply once" rule can be unit tested with a fake reply handler.

### 2.2 `AGWKWebViewController`

- In `AGWKWebViewController.init`, on iOS 16+: create the bridge with `allowedOrigins: [url]` and add it with `configuration.userContentController.addScriptMessageHandler(bridge, contentWorld: .page, name: AGWKWebViewBridge.name)`. `.page` is required: AuthUI's scripts run in the page world.
- The class must still build for the SDK's deployment target (iOS 11 in the podspec), so store the bridge as `AnyObject?` and cast under `#available`.
- After `super.init`, set `presentationAnchorProvider` with `[weak self]`.
- Below iOS 16, add nothing. AuthUI then finds no bridge and behaves as in a browser.

### 2.3 Cleanup

- In `webView(_:didStartProvisionalNavigation:)`, which `AGWKWebViewController` adds to its existing `WKNavigationDelegate` methods, call `bridge.invalidate()`. WebKit calls it only for the top-level page, when it starts loading another page. Same-document changes (`pushState`, fragment) don't call it and don't need to.
- In `viewDidDisappear`, call `bridge.invalidate()` too: reply `error` to the pending handler and clear the pending state first, then call `controller.cancel()` (iOS 16+) on the `ASAuthorizationController`. `cancel()` makes the delegate report a cancellation afterwards; with the pending state cleared, that callback finds nothing to reply to and is ignored. A request must not stay pending after the WebView is gone: iOS allows one passkey request per app, and a leftover one makes every later request fail with 1004.
- No retain cycle: the user content controller holds the bridge strongly; the bridge holds the view controller only through the weak closure.

### 2.4 Adding the new file

Swift Package Manager and CocoaPods pick up `Sources/*.swift` by themselves. `Authgear.xcodeproj` lists files one by one (see how `AGWKWebViewController.swift` appears four times in `project.pbxproj`), so the new file needs entries there.

### 2.5 Public API and docs

- No public API change. The bridge is on whenever `WKWebViewUIImplementation` is used on iOS 16+, for the endpoint's origin.
- Add a line to `CHANGELOG.md` in the file's existing style.
- The example app already declares `webcredentials:` for the staging domains.

## 3. Test plan

Unit tests in `Tests/AuthgearTests/` for the pure functions:

- `bridge.capabilities`: the reply lists exactly the three capabilities;
- request parsing: valid `credentials.get`; a `credentials.get` with only a `password` member is `unsupported`; a `password` member next to `publicKey` is ignored; missing `id`; not JSON; unknown `type`; `mediation` other than `immediate`; missing `challenge` or `rpId`; base64url challenge without padding decodes;
- origin check: same origin; different host; different scheme; explicit default port equals no port; different port; the path and query of an allowed URL are ignored;
- `AGWKWebViewBridgePendingRequest`: `finish` replies once; `cancel` replies `error` once; `finish` after `cancel` replies nothing; `cancel` twice replies once;
- reply encoding: each status; the credential dictionary has the fields and base64url (no `+`, `/`, `=`).

The `ASAuthorizationController` path has no unit test; it is covered by the device test.

## 4. Checks

- `xcodebuild -scheme Authgear-iOS` builds.
- `make test`.
- `make format`.

## 5. Manual test

On a physical device with iOS 16+, with part 2 deployed on a project on a custom domain that lists the example app in `x_native_apps`, and the example app using `WKWebViewUIImplementation`.

1. Device with a passkey for the site: open sign-in. The sheet appears with no tap; choosing a passkey signs in.
2. Close the sheet, close the WebView, open sign-in again. Repeat five times without killing the app. The sheet appears every time. In Safari Web Inspector no reply has `status: "error"` (1004 would show here).
3. While the sheet is open, swipe up to the Home Screen or press the side button: the sheet closes and the app stays open (the WebView can't be touched while the sheet is open). Close the WebView and open sign-in again: the sheet appears, and no reply is `error`.
4. Device without a passkey: nothing appears; the reply is `no_credential`.
5. App without the `webcredentials:` entitlement for the domain: nothing appears; the reply is `error`.
6. iOS 15 device or simulator: no bridge (`window.webkit.messageHandlers.authgearBridge` is undefined); the page behaves as before.
7. Navigation while pending: open sign-in, and while the sheet is showing, run `location.reload()` in Safari Web Inspector. The sheet closes; on the reloaded page, which doesn't check again, "Continue with passkey" opens its dialog.
8. In every run above, the request finds a window: no reply is `error` with "no window". The controller is presented with an animation and loads the URL in `viewDidLoad`, so the page can send its message before the animation ends.

## 6. Compatibility

- **Public API:** no change.
- **Below iOS 16:** no bridge, so AuthUI behaves as in a browser.
- **A server without part 2** never sends a message, so the bridge is unused.
- **An app without the `webcredentials:` entitlement** for the domain gets `error` replies, and the user sees nothing.

## 7. Atomic commits

1. **Add the AuthUI bridge to WKWebViewUIImplementation.** `AGWKWebViewBridge.swift` with the envelope, the origin check, `bridge.capabilities` with an empty list, and `unsupported` for every other type; registration in `AGWKWebViewController.swift`; `project.pbxproj`; tests.
2. **Get a passkey for AuthUI in WKWebViewUIImplementation.** `credentials.get` and its three capabilities, `AGWKWebViewBridgePendingRequest`, and cleanup on navigation and on close, in `AGWKWebViewBridge.swift` and `AGWKWebViewController.swift`; tests; `CHANGELOG.md`.

## 8. Future works

**Allow a Custom UI's origin.** `WKWebViewUIImplementation` gets `public var bridgeAllowedOrigins: [URL] = []`, passed to `AGWKWebViewController.init` and appended to the bridge's `allowedOrigins`. Entries are filtered with `isAcceptableAllowedOrigin(_ url: URL) -> Bool` (true for `https`, and for `http` with the host `localhost`), and each ignored entry is logged with `NSLog`. Tests: the filter, and that a listed origin is allowed. It is one commit on top of this plan, and changes no message. See the spec's future works.
