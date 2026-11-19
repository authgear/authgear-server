- [WebView bridge](#webview-bridge)
  * [Availability](#availability)
  * [Allowed pages](#allowed-pages)
  * [Messages](#messages)
  * [Capabilities](#capabilities)
  * [Message types](#message-types)
    + [bridge.capabilities](#bridgecapabilities)
    + [credentials.get](#credentialsget)
  * [Using the bridge in a page](#using-the-bridge-in-a-page)
  * [Adding a message type](#adding-a-message-type)
  * [Future works](#future-works)

# WebView bridge

When a mobile app shows its sign-in in the SDK's own WebView, the WebView provides `authgearBridge`. Through it, the page asks the app to do something only the app can do natively, such as showing the operating system's passkey sheet. AuthUI uses the bridge.

The bridge is a set of messages. Each message type is a separate feature: an SDK supports some of them, and a page asks which before using one.

## Availability

| Platform | UI implementation | Requires | The page sends | The reply arrives |
|---|---|---|---|---|
| iOS | `WKWebViewUIImplementation` | iOS 14 | `window.webkit.messageHandlers.authgearBridge.postMessage(text)` | As the value of the promise it returns |
| Android | `WebKitWebViewUIImplementation` | A WebView that supports web message listeners | `window.authgearBridge.postMessage(text)` | At `window.authgearBridge.onmessage`, as the event's `data` |

Each message type has its own requirements, which may be stricter. An SDK provides the bridge only where it supports at least one message type other than `bridge.capabilities`.

## Allowed pages

The bridge is available only to the top-level page, never to a frame, and only to a page whose origin is the origin of the SDK's `endpoint`. Origins are compared exactly; a differing port is a differing origin.

On Android, any other page has no `window.authgearBridge`. On iOS, `window.webkit.messageHandlers.authgearBridge` exists on every page, and a message from any other page is rejected. A page treats a rejected message as no bridge.

When the top-level page starts loading another page, every pending request is cancelled, and its reply, if any, is never delivered to the new page.

## Messages

Each message is a JSON object encoded as a string, in both directions. Every request gets exactly one reply.

```json
{ "id": "1", "type": "bridge.capabilities" }
```

```json
{ "id": "1", "status": "ok" }
```

| Field | In | Meaning |
|---|---|---|
| `id` | Request, reply | A string chosen by the page, unique within the page. The reply has the same `id` |
| `type` | Request | The message type |
| `status` | Reply | `ok`, `unsupported` or `error` for every type. A type may define more |
| `message` | Reply | With `error`: a description for logs. It is not meant for the user, and its text may change |

- A message of a type the app does not support is answered with `status: "unsupported"`.
- Unknown fields are ignored, in requests and replies.
- A message that is not a JSON object with a string `id` is answered with `status: "error"` on iOS, and is dropped on Android.
- There is no version number. The bridge changes only by adding message types, fields, and field values.

## Capabilities

A page finds out what the app supports with [`bridge.capabilities`](#bridgecapabilities). Its reply lists capabilities, each a string:

- a message type, such as `credentials.get`, meaning the type is supported;
- a message type, a colon, and a feature that type defines, such as `credentials.get:publicKey`, meaning the feature is supported.

A feature is named after what it enables: a member of the request (`publicKey`), a value of a field (`immediate`, a value of `mediation`), or a field of a member, as a path (`publicKey.hints`).

A listed capability is supported; one that is not listed is not. A page ignores capabilities it does not know. A listed capability means a request is worth trying, not that it will succeed: the reply to the request is the final answer, because a device can still lack what the request needs.

Every capability defined so far:

| Capability | Meaning |
|---|---|
| `credentials.get` | [`credentials.get`](#credentialsget) is supported |
| `credentials.get:publicKey` | Its `publicKey` member: a passkey |
| `credentials.get:immediate` | Its `mediation: "immediate"` |

[`bridge.capabilities`](#bridgecapabilities) is not a capability: every SDK that provides the bridge supports it.

## Message types

### bridge.capabilities

Lists the app's capabilities. Every SDK that provides the bridge supports it, so it is not listed itself.

```json
{ "id": "1", "type": "bridge.capabilities" }
```

```json
{
  "id": "1",
  "status": "ok",
  "capabilities": ["credentials.get", "credentials.get:publicKey", "credentials.get:immediate"]
}
```

The list does not change while the page is shown.

### credentials.get

Asks the app for a credential with the platform's credential API, in the shape of `navigator.credentials.get()`: each kind of credential is a member of the request. Its capabilities are in [Capabilities](#capabilities).

Requires iOS 16 on iOS. On Android, it requires a device whose Credential Manager supports passkeys.

```json
{
  "id": "2",
  "type": "credentials.get",
  "mediation": "immediate",
  "publicKey": {
    "challenge": "3PzOb9VvB54BIdrOC5b88ewjYt1wEOmKbCd0IM8FQSA",
    "rpId": "auth.example.com",
    "userVerification": "preferred",
    "allowCredentials": []
  }
}
```

| Field | Meaning |
|---|---|
| `mediation` | Required. `immediate`: show the sheet only if this device has a credential for the request right now; otherwise end at once without showing anything. A value the app does not support is answered with `unsupported` |
| `publicKey` | `PublicKeyCredentialRequestOptionsJSON` as defined by WebAuthn Level 3, with binary values as base64url without padding. The app uses `challenge`, `rpId`, `userVerification`, and `allowCredentials` when it is not empty; other fields are ignored |

A request has at least one supported credential member, or it is answered with `unsupported`. Members the app does not support are ignored.

| `status` | Meaning |
|---|---|
| `ok` | The user chose a credential, in `credential` |
| `no_credential` | The request ended without a credential. On iOS this also covers the user closing the sheet, because iOS reports both the same way |
| `cancelled` | The user closed the sheet. Sent only where the platform tells this apart from `no_credential` |
| `unsupported` | This app or device cannot perform the request |
| `error` | Anything else, including a `credentials.get` sent while another is pending |

`credential.type` tells the kind of credential. For `public-key`, `credential` is the `AuthenticationResponseJSON` defined by WebAuthn Level 3:

```json
{
  "id": "2",
  "status": "ok",
  "credential": {
    "id": "dFcL6B0cTujk-mONTRqsP4TXVrLWWvzWfa7oG_b36T8",
    "rawId": "dFcL6B0cTujk-mONTRqsP4TXVrLWWvzWfa7oG_b36T8",
    "type": "public-key",
    "authenticatorAttachment": "platform",
    "response": {
      "clientDataJSON": "...",
      "authenticatorData": "...",
      "signature": "...",
      "userHandle": "..."
    },
    "clientExtensionResults": {}
  }
}
```

The app performs one `credentials.get` at a time. The operating system lets the app use only the passkeys of a domain whose association file lists the app, and the server verifies an assertion from the bridge as it verifies one from the WebView.

## Using the bridge in a page

A page finds the bridge, asks for its capabilities, and falls back to its usual behaviour whenever something is missing or a reply is not `ok`.

```js
function getAuthgearBridge() {
  const ios = window.webkit?.messageHandlers?.authgearBridge;
  const android = window.authgearBridge;
  let send;
  if (ios) {
    send = (text) => ios.postMessage(text);
  } else if (android && typeof android.postMessage === "function") {
    const pending = new Map();
    android.onmessage = (event) => {
      let reply;
      try { reply = JSON.parse(event.data); } catch { return; }
      pending.get(reply.id)?.(event.data);
      pending.delete(reply.id);
    };
    send = (text) => new Promise((resolve) => {
      pending.set(JSON.parse(text).id, resolve);
      android.postMessage(text);
    });
  } else {
    return null;
  }
  let nextId = 1;
  return {
    async request(type, fields = {}) {
      const id = String(nextId++);
      return JSON.parse(await send(JSON.stringify({ id, type, ...fields })));
    },
  };
}

async function getCapabilities(bridge) {
  if (bridge == null) return [];
  try {
    const reply = await bridge.request("bridge.capabilities");
    return reply.status === "ok" ? reply.capabilities : [];
  } catch {
    return []; // iOS rejects a page whose origin is not allowed.
  }
}
```

Keep the page's usual behaviour, such as its passkey button: the bridge is an addition, not a replacement.

Do not offer passkey AutoFill (`mediation: "conditional"`) on a page that has the bridge. On iOS, AutoFill keeps a passkey request pending for the app, and the operating system allows one passkey request per app at a time, so the bridge's `credentials.get` would fail.

## Adding a message type

A new message type defines:

- its name, `<area>.<action>`, such as `credentials.get`;
- its capabilities: the type itself, and a `<type>:<feature>` for each optional part a later SDK may add, added to the table in [Capabilities](#capabilities);
- its requirements on each platform;
- its fields, and any statuses beyond `ok`, `unsupported` and `error`.

A later addition to an existing type, such as a new member of `credentials.get`, adds a field and a capability, and changes nothing that exists. Every SDK with its own WebView implements each type it lists.

## Future works

**Custom UI.** A Custom UI shown in the SDK's WebView is on its own origin, so it has no bridge. The UI implementation would get `bridgeAllowedOrigins`, a list of URLs set by the app, whose origins are allowed besides the endpoint's. Only the scheme, host and port of each URL would be used, so a Custom UI's `x_custom_ui_uri` can be listed as it is, and an entry would be ignored unless its scheme is `https`, or it is `http` with the host `localhost`. A listed page can use every message type, such as showing the passkey sheet inside the app, where the user is likely to trust it, so only origins the project controls should be listed. Adding it changes no message: a page written for the bridge works once its origin is listed.

A Custom UI that offers a passkey of this device on its login page, with the helpers of [Using the bridge in a page](#using-the-bridge-in-a-page):

```js
const bridge = getAuthgearBridge();
const capabilities = await getCapabilities(bridge);
const option = flow.action.data.options.find((o) => o.identification === "passkey");

if (
  option != null &&
  ["credentials.get", "credentials.get:publicKey", "credentials.get:immediate"]
    .every((c) => capabilities.includes(c))
) {
  const reply = await bridge.request("credentials.get", {
    mediation: "immediate",
    publicKey: option.request_options.publicKey, // As it is, still base64url.
  }).catch(() => null);

  if (reply?.status === "ok" && reply.credential.type === "public-key") {
    await fetch("/api/v1/authentication_flows/states/input", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify({
        state_token: flow.state_token,
        input: { identification: "passkey", assertion_response: reply.credential },
      }),
    });
  }
}
```
