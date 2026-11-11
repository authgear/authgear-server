# Passkey check in the SDK WebView — part 4: Android SDK

Specs: [docs/specs/webview-bridge.md](../../specs/webview-bridge.md), [docs/specs/passkey-webview.md](../../specs/passkey-webview.md). Repository: `authgear-sdk-android`.

Parts: [1 bridge](./2026-11-10-01-webview-check-bridge.md) · [2 AuthUI](./2026-11-10-02-webview-check-authui.md) · [3 iOS SDK](./2026-11-10-03-webview-check-ios-sdk.md) · **4 Android SDK**

## 1. Goal and scope

`WebKitWebViewUIImplementation` provides the bridge of [part 1](./2026-11-10-01-webview-check-bridge.md) where the device's WebView supports web message listeners, and answers `credentials.get` with Credential Manager.

Files:

- `sdk/build.gradle.kts`
- `sdk/src/main/java/com/oursky/authgear/WebViewBridge.kt` (new)
- `sdk/src/main/java/com/oursky/authgear/WebKitWebViewActivity.kt`
- `sdk/src/test/java/com/oursky/authgear/WebViewBridgeTest.kt` (new)
- `CHANGELOG.md`

## 2. Runtime

### 2.1 Dependencies: `sdk/build.gradle.kts`

```kotlin
implementation("androidx.credentials:credentials:1.5.0")
implementation("androidx.credentials:credentials-play-services-auth:1.5.0")
```

- `credentials-play-services-auth` is what makes Credential Manager work on Android 13 and lower.
- Use 1.5.0. 1.6.0 needs Kotlin 2.1, newer than the SDK's compiler.
- These are new dependencies, so they are downloaded on the first build after this change.
- `androidx.credentials` 1.5.0 requires API level 21, the same as `androidx.webkit` 1.12.1, which the SDK already depends on. So it adds no new minimum.

### 2.2 New file `sdk/src/main/java/com/oursky/authgear/WebViewBridge.kt`

An `internal` class that holds everything about the bridge, so that the activity only wires it up.

```kotlin
internal class WebViewBridge(private val activity: ComponentActivity) {
    fun install(webView: WebView, allowedOrigins: List<Uri>)
    fun cancelPending()
}
```

Pure, unit-testable helpers in the same file. JVM unit tests in this SDK run against the Android stubs (no Robolectric), where `org.json`, `android.net.Uri` and `Bundle` don't work. So the helpers take `java.net.URI` (converted with `URI(uri.toString())`) and use `kotlinx.serialization` (`JsonObject`), which the SDK already depends on:

- `allowedOriginRule(url: URI)`: `"<scheme>://<host>"`, plus `":<port>"` when the URL has an explicit port. Don't use the authority; it can carry user info.
- `sameOrigin(a: URI, b: URI)`: scheme, host and port equal, with the scheme's default port filled in.
- `parseRequest(text)`: returns the `id`, the `type` and the JSON object, or null when the text is not a JSON object with a string `id`.
- `reply(id, status, ...)`: builds the reply text.

`install`:

1. Return if `!WebViewFeature.isFeatureSupported(WebViewFeature.WEB_MESSAGE_LISTENER)`. AuthUI then finds no bridge.
2. `WebViewCompat.addWebMessageListener(webView, "authgearBridge", allowedOrigins.map { allowedOriginRule(URI(it.toString())) }.toSet(), listener)`. In this release `allowedOrigins` holds only the authorization URL; it is a list so that allowing a Custom UI later (future works) only adds entries. Call it before `loadUrl`, so the object exists when the page's scripts run.
3. In the listener: ignore the message when `!isMainFrame` or `message.data` is null or `parseRequest` returns null. Otherwise handle it in `activity.lifecycleScope.launch { ... }` and send exactly one reply with `replyProxy.postMessage(text)`. Keep the `Job` of a pending `credentials.get` and the listener's `sourceOrigin`.
4. Switch on `type`: `"bridge.capabilities"` replies `ok` with `capabilities`, a constant of the bridge: `["credentials.get", "credentials.get:publicKey", "credentials.get:immediate"]` (`[]` until `credentials.get` is added in commit 2); `"credentials.get"` below; anything else replies `unsupported`.
5. Before sending a `credentials.get` reply, check that the WebView's current URL (`webView.url`) has the same origin as `sourceOrigin`. If not, drop the reply. This backs up `cancelPending()`, because `onPageStarted` is not called for every navigation (see the comment in `MyWebViewClient.onPageStarted`).

`cancelPending()` cancels the pending `Job`, which also cancels Credential Manager's request, and sends no reply: the page that asked is gone.

`credentials.get`:

1. `mediation` must be `"immediate"`, else `unsupported`. `publicKey` must be present, else `unsupported` (it is the only member this version supports; other members, such as `password`, are ignored).
2. If a `credentials.get` is pending, reply `error` with message `"a passkey request is in progress"`.
3. Take `publicKey` (a JSON object, else `error`). Remove `extensions` from a copy: they are meant for browsers.
4. Call:

   ```kotlin
   val request = GetCredentialRequest(
       credentialOptions = listOf(GetPublicKeyCredentialOption(publicKey.toString())),
       preferImmediatelyAvailableCredentials = true,
   )
   val result = CredentialManager.create(activity).getCredential(activity, request)
   ```

   Pass the activity as the context, so the sheet is shown over it.

Results:

| Outcome | Reply |
|---|---|
| `result.credential` is `PublicKeyCredential` | `status: "ok"`, `credential` = `credential.authenticationResponseJson` parsed as a JSON object. Credential Manager already returns `AuthenticationResponseJSON`; pass it through unchanged |
| Another credential type | `error` |
| `NoCredentialException` | `no_credential` |
| `GetCredentialCancellationException` | `cancelled` |
| `GetCredentialProviderConfigurationException`, `GetCredentialUnsupportedException` | `unsupported` |
| Any other `GetCredentialException` | `error`, message `"<type>: <message>"` |
| Any other exception | `error` |

Catch the specific exceptions before `GetCredentialException`. Let `CancellationException` from the coroutine propagate without a reply: the activity is going away and so is the page.

### 2.3 `WebKitWebViewActivity`

Add a field, `internal var mBridge: WebViewBridge? = null`. It is `internal`, not `private`, because `MyWebViewClient` reads it.

In `onCreate`, after the WebAuthn setting and before the first `loadUrl`:

```kotlin
val options = this.getOptions()
this.mBridge = WebViewBridge(this).also {
    it.install(this.mWebView, listOf(options.url))
}
```

In `MyWebViewClient.onPageStarted`, call `activity.mBridge?.cancelPending()`. It is called for the top-level page only, when it starts loading another page.

`lifecycleScope` also cancels a pending request when the activity is destroyed.

### 2.4 Public API and docs

- No public API change. The bridge is on whenever `WebKitWebViewUIImplementation` is used and the device's WebView supports web message listeners, for the endpoint's origin.
- Add a line to `CHANGELOG.md` in the file's existing style.
- If the SDK ships consumer ProGuard rules, check that `androidx.credentials` needs none beyond its own.

## 3. Test plan

JVM unit tests in `WebViewBridgeTest.kt` for the helpers:

- `allowedOriginRule`: no port; explicit port; user info is dropped; path and query are dropped;
- `sameOrigin`: equal; different host; different scheme; explicit default port equals no port;
- the `bridge.capabilities` reply lists exactly the three capabilities;
- `parseRequest`: valid; not JSON; a JSON array; missing `id`; non-string `id`;
- reply building: each status; `ok` carries the credential object unchanged.

The Credential Manager path and `cancelPending()` have no unit test; they are covered by the device test.

## 4. Checks

- `./gradlew :sdk:assembleDebug :javasample:assembleDebug`.
- `./gradlew :sdk:testDebugUnitTest`.

## 5. Manual test

On a physical device, with part 2 deployed on a project on a custom domain that lists the example app (`javasample`) and its signing certificate in `x_native_apps`, with **Use WebKit WebView** on.

1. Device with a passkey for the site: open sign-in. The bottom sheet appears with no tap; choosing a passkey signs in. The server accepts the `android:apk-key-hash:` origin.
2. Close the sheet: stays on the login page; the reply is `cancelled`. Go to another page and back: no second sheet.
3. Close the WebView and open sign-in again: the sheet appears again.
4. Device without a passkey: nothing appears; the reply is `no_credential`.
5. After closing the sheet, tap "Continue with passkey": the WebView's own passkey dialog opens.
6. Rotate the device while the sheet is open: one sheet stays; choosing a passkey signs in. The activity handles rotation itself (`android:configChanges`), so the page and the pending request survive it. Without that, the activity was recreated: the page reloaded with a new `sessionStorage` and asked again, and the old request's sheet stayed open and did nothing when used. Also check the toolbar height in both orientations.
7. Navigation while pending: open sign-in, and while the sheet is showing, run `location.reload()` in Chrome DevTools (`chrome://inspect`). The sheet closes; on the reloaded page, which doesn't check again, "Continue with passkey" opens its dialog.
8. App not listed in `x_native_apps`: nothing appears; the reply is `error` or `unsupported`.
9. Repeat 1 and 4 on Android 13 or lower and on Android 14+. On Android 14+, also with a third-party passkey provider enabled and holding no passkey for the site: nothing appears.

Chrome DevTools (`chrome://inspect`) shows the page's console for the replies.

## 6. Compatibility

- **Public API:** no change.
- **A WebView without `WEB_MESSAGE_LISTENER`:** no bridge, so AuthUI behaves as before.
- **A server without part 2** never sends a message, so the bridge is unused.
- **An app not listed in `x_native_apps`** gets `error` or `unsupported` replies, and the user sees nothing.

## 7. Atomic commits

1. **Add the AuthUI bridge to WebKitWebViewUIImplementation.** `WebViewBridge.kt` with the envelope, the origin rule, `bridge.capabilities` with an empty list, and `unsupported` for every other type; the call in `WebKitWebViewActivity.kt`; `WebViewBridgeTest.kt`.
2. **Get a passkey for AuthUI in WebKitWebViewUIImplementation.** `sdk/build.gradle.kts`; `credentials.get` and its three capabilities, `cancelPending()` and the reply origin check in `WebViewBridge.kt`; the call in `MyWebViewClient.onPageStarted`; tests; `CHANGELOG.md`.

## 8. Future works

**Allow a Custom UI's origin.** `WebKitWebViewUIImplementation` gets `var bridgeAllowedOrigins: List<Uri> = emptyList()`, declared in the class body so that the constructor Java callers use does not change. It is copied into `WebKitWebViewActivity.Options` (kept in its `Bundle` with `putParcelableArrayList`), and appended to the list passed to `install`. Entries are filtered with `isAcceptableAllowedOrigin(url: URI)` (true for `https`, and for `http` with the host `localhost`), and each ignored entry is logged with `Log.w`. Tests: the filter as a JVM unit test; the `Bundle` round trip on a device, by rotating while a Custom UI is shown. It is one commit on top of this plan, and changes no message. See the spec's future works.
