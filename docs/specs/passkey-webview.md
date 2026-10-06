- [Passkey in a WebView](#passkey-in-a-webview)
  * [Use cases](#use-cases)
  * [Configuration](#configuration)
  * [Association files](#association-files)
  * [SDK](#sdk)
  * [Limitations](#limitations)
  * [Future works](#future-works)

# Passkey in a WebView

This document specifies passkeys in AuthUI when a mobile app shows AuthUI in its own WebView instead of the system browser. Passkey concepts and the web flows are specified in [webauthn.md](./webauthn.md).

A WebView does not prove its page's origin to the operating system the way a browser does. The operating system instead allows a ceremony only when the app is listed in an association file served at the Relying Party ID. This document specifies how a project lists its apps, and the files Authgear serves from that list.

The Relying Party ID is unchanged: it is the host of `http.public_origin`.

## Use cases

| Use case | Supported |
|---|---|
| Sign up, sign in, and add a passkey in AuthUI shown in the app's WebView | Yes |
| The same, in a Custom UI shown in the app's WebView | Yes |
| Passkey ceremonies performed by the app natively, without a WebView | No, see [Future works](#future-works) |

## Configuration

```yaml
oauth:
  clients:
  - client_id: example-app
    x_application_type: native
    x_native_apps:
    - platform: ios
      team_id: ABCDE12345
      bundle_id: com.example.app
    - platform: android
      package_name: com.example.app
      sha256_cert_fingerprints:
      - "8B:BF:39:60:61:89:30:A4:45:F3:D7:09:1E:7B:1B:05:0F:8A:FD:AF:24:EB:F1:EB:2E:3D:13:88:09:FC:79:59"
```

`x_native_apps` lists a client's mobile apps that are allowed to perform passkey ceremonies for the project. When no client has it, no association file is served.

| Field | Meaning | Required |
|---|---|---|
| `platform` | `ios` or `android` | Yes |
| `team_id` | The Apple team ID, ten uppercase letters and digits | For `ios` |
| `bundle_id` | The bundle ID | For `ios` |
| `package_name` | The Android package name | For `android` |
| `sha256_cert_fingerprints` | The SHA-256 fingerprint of every certificate the app is signed with, as 32 colon-separated hex bytes, compared case-insensitively | For `android`, at least one |

- An entry takes only the fields of its platform.
- An app may be listed only once across all clients: the same `bundle_id` for `ios`, or the same `package_name` for `android`, may not appear twice.
- List every certificate the app is distributed with: the release certificate, and the Play App Signing certificate, which is the one on a user's device when the app is installed from Google Play. A build whose certificate is absent cannot sign in with a passkey.
- List a debug certificate only in a project used for development. A debug keystore is stored unencrypted with a well-known password, so anyone who obtains it can sign an app that passes as the listed one.
- The association files merge the entries of every client. Listing an app on a client does not restrict the app to that client: a listed app may perform ceremonies whichever OAuth client it signs in with.

Listing an app trusts it with the project's sign-ins. An app controls whatever it shows in its WebView, including AuthUI, and a listed app can also request the user's passkeys. List only apps the project controls.

`x_native_apps` has no effect while `passkey` is not in `authentication.identities`: no association file is served, and no app origin is accepted.

## Association files

Authgear serves both files at the host of `http.public_origin`, generated from every client's `x_native_apps`.

| Path | Served when | Content |
|---|---|---|
| `/.well-known/apple-app-site-association` | at least one `ios` entry | `webcredentials.apps`, one `<team_id>.<bundle_id>` per `ios` entry |
| `/.well-known/assetlinks.json` | at least one `android` entry | one statement per `android` entry, with the relations `delegate_permission/common.handle_all_urls` and `delegate_permission/common.get_login_creds` |

Both are served over HTTPS with `200` and `Content-Type: application/json`, without a redirect, and without authentication. A path that is not served returns `404`.

- Android requires both relations for a passkey sign-in; with `get_login_creds` alone, sign-in is refused on the device.
- `handle_all_urls` also lets a listed Android app become the handler for URLs on the host and its subdomains, if it declares a verified intent filter for them. These include AuthUI, the authorization endpoint, links in emails, and a Custom UI hosted under the host. An app must not declare one. Check the app's merged manifest, because a library can add an intent filter.

The origin an Android app reports is accepted as specified in [webauthn.md](./webauthn.md#where-a-ceremony-may-run).

## SDK

| Platform | UI implementation | What the SDK does | What the app does |
|---|---|---|---|
| iOS | `WKWebViewUIImplementation` | Nothing new | Declares `webcredentials:<host of http.public_origin>` in its Associated Domains entitlement |
| Android | `WebKitWebViewUIImplementation` | Enables WebAuthn in its WebView for the app's associated domains, when the device's WebView supports it | Nothing |

AuthUI offers no passkey autofill in an Android WebView, because the WebView does not support it; the passkey button is used instead.

In an Android WebView, AuthUI hides every passkey option when `PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable()` resolves to `false`, which is what the WebView reports when it cannot perform a ceremony, for example without the required Google Play services. The options are signing in with a passkey, signing up with a passkey, adding a passkey in settings, and the prompt to create a passkey after signing in, which is skipped. Other browsers keep today's behaviour, because a `false` there still allows a security key or a passkey on another device.

A Custom UI in a WebView follows [Where a ceremony may run](./webauthn.md#where-a-ceremony-may-run), as it does in a browser, and needs no listing beyond `x_native_apps`: the operating system checks the app against the Relying Party ID, not against the Custom UI's host. The rule above for hiding passkey options applies to AuthUI only; a Custom UI shown in an Android WebView should apply the same check.

The SDK does not switch to a Custom Tab when passkeys are unavailable. The user signs in with another method, and an app that prefers Custom Tabs chooses `CustomTabsUIImplementation`.

## Limitations

- iOS requires iOS 16.
- Android requires an Android System WebView that supports WebAuthn. It updates through Google Play.
- Android requires Google Play services 24.07 or newer, on every Android version, because the WebView checks for it before any ceremony ([Chromium source](https://chromium.googlesource.com/chromium/src/+/HEAD/components/webauthn/android/java/src/org/chromium/components/webauthn/GmsCoreUtils.java)). On a device without it, passkeys are unavailable in the WebView; other sign-in methods are unaffected.
- Changing `http.public_origin` changes the Relying Party ID, which makes existing passkeys unusable. This is existing behaviour. The association files follow the new public origin without further configuration.
- Removing an app from `x_native_apps` takes effect at once on Android, because its origin is no longer accepted. On iOS the server cannot tell which app performed a ceremony, so the removal takes effect only when devices fetch the updated `apple-app-site-association`.
- Apple fetches `apple-app-site-association` through its own CDN, so a change can take time to reach devices. During development, the app can append `?mode=developer` to the entitlement to fetch it directly.

## Future works

**Native passkey.** Ceremonies performed by the app through the platform passkey APIs, without a WebView. It uses the same `x_native_apps`, and adds a Relying Party ID other than the host of `http.public_origin`.
