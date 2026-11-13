# Passkey check in the SDK WebView — part 2: AuthUI

Specs: [docs/specs/webview-bridge.md](../../specs/webview-bridge.md) (the bridge), [docs/specs/passkey-webview.md](../../specs/passkey-webview.md#passkey-on-page-load) (the passkey on page load).

Parts: [1 bridge](./2026-11-10-01-webview-check-bridge.md) · **2 AuthUI** · [3 iOS SDK](./2026-11-10-03-webview-check-ios-sdk.md) · [4 Android SDK](./2026-11-10-04-webview-check-android-sdk.md)

## 1. Goal and scope

AuthUI uses the bridge of [part 1](./2026-11-10-01-webview-check-bridge.md): it asks the SDK's WebView for a passkey when the login page loads, and turns AutoFill off there.

Repository: `authgear-server`. This builds on passkey AutoFill ([plan](./2026-11-04-autofill-authui-v2.md)).

Files:

- `authui/src/webviewBridge.ts`, `authui/src/webviewBridge.test.ts` (new)
- `authui/src/passkey.ts`, `authui/src/passkey.test.ts`
- `authui/src/authflowv2.ts`
- `resources/authgear/templates/en/web/authflowv2/login.html`, `signup.html`
- `pkg/auth/handler/webapp/authflowv2/passkey_autofill_test.go`

## 2. Runtime

No Go code changes apart from the template test. The server already accepts what the apps send (see [part 1, Security](./2026-11-10-01-webview-check-bridge.md#4-security)).

### 2.1 Bridge client: new file `authui/src/webviewBridge.ts`

A small module with no Stimulus and no passkey knowledge.

```ts
// Returns null when the page is not in an Authgear SDK WebView with the bridge.
export function getWebViewBridge(): WebViewBridge | null;

export interface WebViewBridge {
  // Sends {id, type, ...fields} and resolves with the reply object.
  // Rejects when the transport fails or the reply is not a JSON object.
  request(type: string, fields: Record<string, unknown>): Promise<WebViewBridgeReply>;
}

export interface WebViewBridgeReply {
  status: string;
  [key: string]: unknown;
}

// Resolves with the reply's capabilities, or [] when the request is rejected
// (iOS, origin not allowed) or the reply is not ok.
export async function getWebViewBridgeCapabilities(bridge: WebViewBridge): Promise<string[]>;
```

- iOS: `window.webkit.messageHandlers.authgearBridge.postMessage(text)` returns a promise of the reply text.
- Android: `window.authgearBridge.postMessage(text)`; replies arrive at `window.authgearBridge.onmessage`. Set `onmessage` once, keep a `Map` from `id` to resolver, and drop replies with an unknown `id`.
- `id`: a counter in the module, as a string.
- Check iOS first, then Android.

### 2.2 New controller `PasskeyWebViewCheckController` in `authui/src/passkey.ts`

Stimulus identifier `passkey-webview-check`. Targets `input`, `submit`. Value `options` (string). Register it in `authflowv2.ts` only (the pages are authflowv2 pages).

On `connect()`:

1. `const bridge = getWebViewBridge()`. If null, return.
2. If the "asked" mark is set, return. Otherwise set it **before** sending.
   - The mark is `sessionStorage` key `authgear.passkey.webview_check`. Wrap reads and writes in `try/catch`. If storage throws, fall back to a module-level boolean (it survives Turbo visits, which is enough to avoid asking twice in a row).
3. `const capabilities = await getWebViewBridgeCapabilities(bridge)`. Return unless it includes `credentials.get`, `credentials.get:publicKey` and `credentials.get:immediate`.
4. Send `credentials.get` with `mediation: "immediate"` and `publicKey: JSON.parse(this.optionsValue).publicKey`. Pass the JSON as it is: do **not** run it through `deserializeRequestOptions`, which turns base64url into bytes for the browser API.
5. Reply `status === "ok"` with a `credential` object whose `type` is `"public-key"`: `this.inputTarget.value = JSON.stringify(reply.credential); this.submitTarget.click();`.
6. Anything else: nothing visible. `console.warn` for `error` only.
7. If the controller was disconnected while waiting (the page changed), ignore the reply.

It does not listen to `pageshow`: a page restored from the cache must not ask again.

### 2.3 Autofill is off when the bridge exists

In `PasskeyAutofillController.start()`, return early when `getWebViewBridge() != null`. Add a short comment with the reason (one passkey request per app on iOS; see part 1, section 3, rule 4). This also covers the `pageshow` restart, because it goes through `start()`.

### 2.4 Templates

Attach the new controller where the page already has a passkey form with the options: the container of the "Continue with passkey" button (`data-controller="authflow-passkey-request"`, `login.html` near line 321, `signup.html` near line 375). Stimulus allows two controllers on one element:

```html
data-controller="authflow-passkey-request passkey-webview-check"
data-passkey-webview-check-options-value="{{ $.PasskeyRequestOptionsJSON }}"
...
<input ... data-authflow-passkey-request-target="input" data-passkey-webview-check-target="input">
<button ... data-authflow-passkey-request-target="submit" data-passkey-webview-check-target="submit"></button>
```

Conditions:

- Never in inline preview (`$.InlinePreview`).
- `login.html`: wherever the passkey button is rendered.
- `signup.html`: only when `$.UIVariant` is `signup_login`. The plain signup page must not ask.

Don't use the autofill block for this: it is only rendered when the page has a login ID input, and the check must also run on a login page without one.

Check that the button block's form posts the same fields as the autofill block (`x_assertion_response`, `x_action=passkey`).

## 3. Test plan

`authui/src/passkey.test.ts` already mocks `navigator.credentials.get` and has `render()`, `flush()`, `showPage()`. Add a way to install a fake bridge on `window` (both shapes) and remove it in `afterEach`, and clear `sessionStorage` in `afterEach`.

New file `authui/src/webviewBridge.test.ts`:

- returns null when neither object exists;
- iOS shape: sends `{id, type, ...}` as text and resolves with the parsed reply;
- Android shape: resolves the right promise when two requests are answered out of order; ignores a reply with an unknown `id`;
- rejects on a reply that is not a JSON object;
- `getWebViewBridgeCapabilities`: returns the list on `ok`; returns `[]` on a rejected request, on `unsupported`, and when `capabilities` is not an array.

In `passkey.test.ts`:

- with a bridge: `bridge.capabilities` is sent, then one `credentials.get` with `mediation: "immediate"` and the page's `publicKey` unchanged (the challenge is still the base64url string);
- capabilities missing any of the three: no `credentials.get` is sent; a rejected `bridge.capabilities`: no `credentials.get`;
- `ok` with a `credential.type` other than `public-key`: no submit;
- `ok`: the input holds the credential JSON and the submit button is clicked once;
- `no_credential`, `cancelled`, `unsupported`, `error`, and a rejected promise: no submit;
- asked once: render, clear the body, render again, and no second message is sent; also not after `showPage(true)`;
- with a bridge, autofill makes no `navigator.credentials.get` call, also after `showPage(true)`;
- without a bridge, everything behaves as before (existing tests stay green and unchanged);
- the button still calls `navigator.credentials.get` when a bridge exists.

Go: extend `passkey_autofill_test.go` to assert that the new attributes are rendered on the login page and the signup-login page, and not on the plain signup page or in inline preview.

## 4. Checks

- `npm --prefix authui run typecheck`
- `npm --prefix authui run eslint`
- `npm --prefix authui run prettier`
- `cd authui && npx jest`
- `go test ./pkg/auth/handler/webapp/authflowv2/...`
- `go vet ./pkg/auth/handler/webapp/authflowv2/...`
- `make authui`

## 5. Manual test

On physical devices. Needs parts 3 and 4, a project on a custom domain, and the example apps listed in `x_native_apps`.

1. Device with a passkey: open sign-in in the SDK WebView. The sheet appears with no tap; choosing a passkey signs in.
2. Close the sheet: stays on the login page. Go to another page and back, or reload: no second sheet.
3. Close the WebView and open sign-in again: the sheet appears again. Repeat five times without killing the app; it must keep appearing (this is the 1004 regression check on iOS).
4. Device without a passkey: nothing appears.
5. After closing the sheet, tap "Continue with passkey": the dialog opens.
6. A passkey that was deleted on the server: an error is shown on the login page, and the sheet does not reappear.
7. Safari, Chrome, ASWebAuthenticationSession and Custom Tabs: unchanged, autofill still offered.
8. An older SDK version without the bridge, in `WKWebView`: AutoFill is offered as in a browser. Open and close the WebView several times and check that "Continue with passkey" still opens its dialog.

## 6. Compatibility

- Without the bridge, AuthUI behaves as before, including AutoFill.
- The `sessionStorage` key `authgear.passkey.webview_check` is new and is only read by this controller.
- A native request on iOS reports the origin `https://<rpId>`. It equals the public origin when the public origin uses the default port.

## 7. Fixed decisions

- The check is attached to the passkey button's container, not the AutoFill block.
- The button's `_use()` flow is not changed, and nothing is awaited between the tap and its `navigator.credentials.get`.
- The button is not disabled while the check is pending. The OS sheet covers the page while it is shown.

## 8. Atomic commits

The spec is already committed (`doc: Specify the WebView bridge and the passkey on page load`).

1. **Add the AuthUI WebView bridge client.** `authui/src/webviewBridge.ts`, `authui/src/webviewBridge.test.ts`.
2. **Turn off passkey autofill in the SDK WebView.** `authui/src/passkey.ts`, `authui/src/passkey.test.ts`.
3. **Ask the SDK WebView for a passkey when the login page loads.** `authui/src/passkey.ts`, `authui/src/passkey.test.ts`, `authui/src/authflowv2.ts`, `login.html`, `signup.html`, `passkey_autofill_test.go`.
