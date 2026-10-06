# Native passkey

This document specifies passkey authentication performed **natively by a mobile app**, without opening a browser.

Passkey concepts, the credential model, and the web flows are specified in [webauthn.md](./webauthn.md). This document covers only what native adds: relying party domain setup, native app registration, and the SDK surface.

Open questions are marked **Q<n>** in the section they belong to, and collected in [Open questions](#open-questions).

- [What changes for the end user](#what-changes-for-the-end-user)
- [Use cases](#use-cases)
- [Relying party domain](#relying-party-domain)
- [Association files](#association-files)
- [Configuration](#configuration)
- [SDK API](#sdk-api)
- [Examples](#examples)
- [Errors](#errors)
- [Limitations](#limitations)
- [Future works](#future-works)
- [Open questions](#open-questions)
- [Remarks](#remarks)

## What changes for the end user

Today a passkey ceremony in a mobile app happens inside a browser opened by the SDK. The user leaves the app, sees a web page, and comes back.

With native passkey the ceremony happens in the app. The user taps a button and the system passkey sheet appears over the app.

**The credential is the same credential.** A passkey created in the app is the same one the user's browser offers on the same device, and the reverse, provided both use the same [relying party domain](#relying-party-domain). Native passkey does not introduce a second kind of credential, and a user who already has a passkey does not need a new one.

## Use cases

| Use case | Supported |
|---|---|
| Sign in with a passkey | Yes |
| Add a passkey to the signed-in account | Yes |
| List the account's passkeys | Yes |
| Remove a passkey from the account | Yes |
| Sign up with a passkey | No — see [Future works](#future-works) |

A user obtains their first passkey either in AuthUI, as today, or by adding one from inside the app once signed in.

## Relying party domain

Every passkey is permanently bound to one domain, the **relying party ID** (RP ID). The RP ID decides which apps and which web pages may use the credential, and it cannot be changed for a credential once created.

The rule that governs every choice below: **a passkey created for an RP ID can be used by that domain and its subdomains, and by no one else.** A sibling subdomain cannot use another sibling's credentials.

### Choosing the RP ID

Given a project whose AuthUI is at `auth.example.com`:

| RP ID | Passkey usable from | Association files hosted by |
|---|---|---|
| `auth.example.com` | the mobile apps, and AuthUI | Authgear |
| `example.com` | the mobile apps, AuthUI, and any page under `example.com` including the customer's own web app | the customer |

Choose `example.com` — the registrable parent domain — when the customer's own website will also run passkey ceremonies. Choose `auth.example.com` when all authentication happens in AuthUI and the apps; it requires no work from the customer.

A project served on the shared Authgear domain (`*.authgearapps.com`) can only use its own hostname as the RP ID, which means its passkeys are scoped to a domain the customer does not own.

### Which domains may be used

`relying_party_id` has two permitted states.

| | `http.public_origin` | Relying party ID | Association files served by | Customer's own web app can use the passkeys |
|---|---|---|---|---|
| **not set** | default domain, or a custom subdomain such as `auth.example.com` | that host | Authgear | no |
| **not set** | the apex as the custom domain, `example.com` | `example.com` | Authgear | **yes** |
| **set to the apex** | a custom subdomain such as `auth.example.com` | `example.com` | the customer | **yes** |

No other value is accepted.

Leaving it unset is the zero-setup option: Authgear routes the public origin by definition, so it can always serve the files there.

The second row is worth noticing. A project willing to make the apex its public origin gets apex scope — its own website can use the passkeys — with nothing to host and no relying party ID to set. It suits a project whose apex is not already a website.

The third row is for a project that wants apex scope while keeping its apex for its own site. Verifying a custom domain requires a TXT record on its apex, so a project with `auth.example.com` has already proven control of `example.com` — but proving control is not the same as routing it to us. The apex is verified, not served, which is why this is the one case where the customer hosts the two files themselves.

Two constraints follow:

- The apex is only accepted while a custom domain under it is active, because the public origin must be the relying party ID or a subdomain of it. `myapp.authgearapps.com` is not under `example.com`, so the combination is rejected rather than failing at ceremony time.
- The shared domain `authgearapps.com` is never permitted. A default domain is recorded as its own apex, so it never appears as the apex of anything.

Association files are needed only by native apps. A project using passkeys only on the web needs none in either state, because a browser proves its own origin.

### Surviving a domain change

Changing the relying party ID invalidates every passkey enrolled under the old value. Three ordinary operations change it when it is left to follow `http.public_origin`:

| Operation | Effect |
|---|---|
| Moving from the default domain to a custom domain | all passkeys invalidated |
| Rebranding from one custom domain to another | all passkeys invalidated |
| Deleting the custom domain the public origin points at | public origin reverts to the default domain automatically, all passkeys invalidated |

The third takes no passkey-related action at all.

Setting the relying party ID explicitly to the **apex** removes the second case and reduces the third, because a credential scoped to `example.com` remains usable from every subdomain of it. Only a change of apex invalidates it. A project that expects to move or rename its authentication hostname should prefer the apex for this reason, not only to share passkeys with its own web app.

This is the behaviour today for every project, with no way to opt out of it.

When `relying_party_id` is not set it follows the public origin's host, which is how the system behaves today: a project that switches to a custom domain loses the passkeys enrolled under the previous hostname, and its users enrol again. Setting `relying_party_id` explicitly fixes it, and the public origin is then required to be that domain or a subdomain of it, so a later move within the same apex leaves existing passkeys working.

A passkey is offered only when the relying party ID it was enrolled under matches the project's current one, compared as an exact, case-insensitive string. No change of relying party ID preserves a credential, in either direction: widening the scope from `auth.example.com` to `example.com` invalidates existing passkeys exactly as narrowing it does, because the credential is bound to the whole string and not to a range. Choosing the apex for durability is therefore worth doing before any passkey is enrolled, and costs a full re-enrolment afterwards.

Nothing is deleted when this happens, on either side. Restoring the previous relying party ID makes the credentials usable again.

Two behaviours follow:

- **A passkey the current relying party ID cannot use is shown, and marked as unusable, wherever passkeys are listed** — the account settings page, and the SDK. It remains deletable, so a user can clear entries left behind by an earlier domain. Hiding them instead would leave invisible records that accumulate each time a user re-enrols.
- **Changing a domain in a way that changes the relying party ID warns, and states how many users hold passkeys that will stop working.** This covers deleting the custom domain the public origin points at, which changes the relying party ID with no passkey-related action at all. The warning does not block the change: moving domain is a legitimate operation and passkeys do not get to veto one. It should say that restoring the previous domain restores the passkeys.

Projects that edit configuration outside the portal receive no warning; this is accepted.

## Association files

A native app has no origin, so the operating system checks that the app is authorised by the RP ID domain before allowing a ceremony. The proof is a file served from that domain.

| Platform | File | Contents |
|---|---|---|
| iOS | `https://<rp-id>/.well-known/apple-app-site-association` | the app's Apple team ID and bundle ID, under `webcredentials` |
| Android | `https://<rp-id>/.well-known/assetlinks.json` | the app's package name and SHA-256 signing certificate fingerprints, under `delegate_permission/common.get_login_creds` |

Both declarations cover **credentials for the domain**, not passkeys alone. Publishing them also lets the app offer passwords saved for that domain when the user signs in with a password. This follows from the platform declaration and is not separately configurable.

Requirements that apply to both:

- Served over HTTPS with a `200` response. A redirect fails verification, so a redirect from the RP ID domain to Authgear does not work; a reverse proxy does.
- Every signing certificate must be listed. Debug builds, release builds, and Play App Signing use different certificates, and a passkey will fail on any build whose fingerprint is absent.
- The iOS app must additionally declare the RP ID in its Associated Domains capability as `webcredentials:<rp-id>`. This is done in the app, not in Authgear.

Authgear generates and serves both files from the values in [Configuration](#configuration) when the RP ID is a domain Authgear serves. When the RP ID is a domain the customer controls, the customer must serve the files.

### The relying party domain is independent of the app's link domains

An app declares each platform capability against its own domain, and each is verified from a separate file. On iOS the Associated Domains entitlement lists `service:domain` pairs, so an app may declare `applinks:a.example.com` and `webcredentials:auth.example.com` together; iOS fetches an association file from each domain and reads only that service's key from it. On Android, App Links verification fetches `assetlinks.json` from the host in the app's verified intent filter, while a passkey ceremony fetches it from the relying party ID.

So an app's universal links and its passkeys normally live on different domains and are served by different files, with no interaction between them. In particular, app2app's files stay on the apps' own link domains and are unaffected by passkey.

### Merging, when one domain serves both

The files interact only if the relying party ID is the *same domain* that already serves an association file for another purpose. In that case the customer must merge, not replace. A domain has exactly one file per platform, shared by every service that uses that domain. Replacing it with Authgear's generated content — including by proxying the path to Authgear — removes whatever else was there, and the failure is silent: passkey works while the other feature stops.

On Android the file is an array of statements, so both relations go in one statement when they describe the same app:

```json
[{
  "relation": [
    "delegate_permission/common.handle_all_urls",
    "delegate_permission/common.get_login_creds"
  ],
  "target": {
    "namespace": "android_app",
    "package_name": "com.example.app",
    "sha256_cert_fingerprints": ["8B:BF:39:..."]
  }
}]
```

On iOS the file is an object keyed by service, so merging adds a sibling key:

```json
{
  "applinks": {
    "details": [{
      "appIDs": ["ABCDE12345.com.example.app"],
      "components": [{ "/": "/redirect*", "comment": "app2app callback" }]
    }]
  },
  "webcredentials": {
    "apps": ["ABCDE12345.com.example.app"]
  }
}
```

Choosing a relying party ID that no other service uses avoids this entirely.

Setting the relying party ID to the apex is the one configuration in which the customer must serve the association files themselves, and native passkey does not work until they do. The portal warns when that value is set, stating which two files are needed and where they must be reachable. It does not check that they are actually served: a reachability probe would have to be retried, cached and reported somewhere, and the warning already tells the customer what is required.

Declaring `webcredentials` does not let the app intercept URLs belonging to the RP ID domain; URL handling is a separate capability on both platforms. Opening AuthUI in a browser continues to work unchanged.

## Configuration

Passkey must first be enabled for the project as an identity and a primary authenticator. This is existing configuration, unchanged by native passkey:

```yaml
authentication:
  identities:
  - login_id
  - passkey
  primary_authenticators:
  - password
  - passkey
```

### Passkey configuration

```yaml
identity:
  passkey:
    relying_party_id: example.com
    native_apps:
    - platform: ios
      team_id: ABCDE12345
      bundle_id: com.example.app
    - platform: android
      package_name: com.example.app
      sha256_cert_fingerprints:
      - "8B:BF:39:60:61:89:30:A4:45:F3:D7:09:1E:7B:1B:05:0F:8A:FD:AF:24:EB:F1:EB:2E:3D:13:88:09:FC:79:59"
```

Passkey configuration sits under `identity` alongside `biometric`, `login_id`, `oauth` and `ldap`, which is where per-identity-type configuration belongs. A passkey credential is an identity as well as a primary authenticator, and both records are produced by one ceremony from one set of options.

`relying_party_id` is the domain passkeys are bound to. When absent, it is the hostname of `http.public_origin`. That default preserves today's behaviour exactly: Authgear redirects any request whose host is not `http.public_origin`, so every passkey in existence was already created under that hostname.

`native_apps` lists the mobile apps allowed to perform passkey ceremonies for this project. It is project-level rather than per-OAuth-client because the association files it produces are project-level: one file per domain, served at a fixed path with no client dimension. The Android fingerprints serve a second purpose — an Android ceremony identifies itself by signing certificate rather than by origin, and a ceremony from an unlisted certificate is rejected.

List every signing certificate an app is built with. Debug builds, release builds and Play App Signing use different certificates, and passkey fails on any build whose fingerprint is absent.

Listing an app under `identity.passkey` is what authorises it; there is no separate enable flag. Should Authgear later generate association files for another capability, a project-wide list can be introduced then and this one deprecated.

## SDK API

The iOS and Android APIs use the same method names and the same options types. They differ in one respect: Android requires an `Activity`, because the platform launches the passkey sheet from one; iOS does not.

| Operation | Method | Requires a signed-in session |
|---|---|---|
| Sign in with a passkey | `authenticatePasskey` | No |
| Add a passkey | `addPasskey` | Yes |
| List passkeys | `listPasskeys` | Yes |
| Remove a passkey | `removePasskey` | Yes |

`listPasskeys` returns, for each passkey, an identifier accepted by `removePasskey`, plus the creation time and the display name shown to the user.

Minimum versions are iOS 16 and Android 9 (API 28), imposed by the platform passkey APIs. Calling any of these below the minimum fails with a "not supported" error rather than crashing; the SDK package minimums are unchanged, so existing apps keep building.

> **Q6.** Should `authenticatePasskey` be available to third-party OAuth clients, or first-party only? Biometric authentication is first-party only today.

### iOS

```swift
@available(iOS 16.0, *)
public func authenticatePasskey(
    options: PasskeyAuthenticateOptions = .init(),
    handler: @escaping UserInfoCompletionHandler
)

@available(iOS 16.0, *)
public func addPasskey(
    options: PasskeyAddOptions = .init(),
    handler: @escaping VoidCompletionHandler
)

@available(iOS 16.0, *)
public func listPasskeys(
    handler: @escaping (Result<[PasskeyInfo], Error>) -> Void
)

@available(iOS 16.0, *)
public func removePasskey(
    identityID: String,
    handler: @escaping VoidCompletionHandler
)
```

```swift
public struct PasskeyAuthenticateOptions {
    /// Offer only passkeys already on this device, instead of also offering
    /// the cross-device flow. Defaults to false.
    public let preferImmediatelyAvailableCredentials: Bool
}

public struct PasskeyAddOptions {}
```

### Android

```kotlin
fun authenticatePasskey(options: PasskeyAuthenticateOptions, listener: OnAuthenticatePasskeyListener)
fun addPasskey(options: PasskeyAddOptions, listener: OnAddPasskeyListener)
fun listPasskeys(listener: OnListPasskeysListener)
fun removePasskey(identityID: String, listener: OnRemovePasskeyListener)

suspend fun Authgear.authenticatePasskey(options: PasskeyAuthenticateOptions): UserInfo
suspend fun Authgear.addPasskey(options: PasskeyAddOptions)
suspend fun Authgear.listPasskeys(): List<PasskeyInfo>
suspend fun Authgear.removePasskey(identityID: String)
```

```kotlin
data class PasskeyAuthenticateOptions @JvmOverloads constructor(
    var activity: Activity,
    var preferImmediatelyAvailableCredentials: Boolean = false
)

data class PasskeyAddOptions constructor(
    var activity: Activity
)
```

## Examples

### Sign in with a passkey

iOS:

```swift
authgear.authenticatePasskey { result in
    switch result {
    case let .success(userInfo):
        // signed in
    case let .failure(error):
        if case AuthgearError.passkeyNotFound = error {
            // no passkey on this device — offer another method
        }
    }
}
```

Android:

```kotlin
lifecycleScope.launch {
    try {
        val userInfo = authgear.authenticatePasskey(
            PasskeyAuthenticateOptions(activity = this@MainActivity)
        )
    } catch (e: PasskeyNotFoundException) {
        // no passkey on this device — offer another method
    }
}
```

### Add a passkey from the app's settings screen

iOS:

```swift
authgear.addPasskey { result in
    // on success, reload the list
}
```

Android:

```kotlin
authgear.addPasskey(PasskeyAddOptions(activity = this))
```

Adding requires a signed-in session. A device that already holds a passkey for this account will refuse to create a second one, surfacing as a "already exists" error.

### List and remove

```swift
authgear.listPasskeys { result in
    if case let .success(passkeys) = result {
        // show passkeys; each has an id, a display name and a creation time
    }
}
```

```kotlin
val passkeys = authgear.listPasskeys()
authgear.removePasskey(passkeys.first().id)
```

## Errors

Each SDK maps platform errors onto its own error type, so an integrator handles one error vocabulary rather than the platform's.

| Condition | Meaning for the integrator |
|---|---|
| User dismissed the system sheet | Not an error to report; the user chose to stop |
| No passkey available on the device | Offer another authentication method |
| Domain not associated | The association file or the Associated Domains capability is wrong. This is a setup mistake, not a runtime condition, and is the most common integration failure |
| Passkey already exists for this account | Only on add |
| Platform version below the minimum | Hide the passkey option |

## Limitations

- A passkey works only on devices that can reach the credential. A passkey held in iCloud Keychain is not offered on an Android device except through the cross-device flow.
- Android emulators cannot perform passkey ceremonies. Development and testing require a physical device.
- Apple caches the iOS association file for up to 48 hours. A newly published or corrected file may not take effect immediately; the app can request a developer mode that bypasses the cache during development.

## Future works

The following are out of scope for this document.

**Sign up with a passkey, natively.** Creating an account and its first passkey in one native flow. Sign-up carries verification, bot protection, and account-linking policy, which a single credential ceremony cannot express. Signing up continues to happen in AuthUI.

**Passkey as a second factor.** Passkey is a primary authenticator only.

**Passkey autofill in native apps.** Suggesting a passkey from a text field, as AuthUI does on the web.

**Automatic passkey upgrade.** Creating a passkey silently after a password sign-in ([#4308](https://github.com/authgear/authgear-server/issues/4308)). The SDK API in this document is shaped so that this can be added without changing any method signature.

## Open questions

| | Question | Section |
|---|---|---|
| Q6 | Whether native passkey sign-in is first-party only | [SDK API](#sdk-api) |

### Resolved

| | Decision |
|---|---|
| Q1 | A relying party ID change invalidates passkeys in either direction and is reversible. Unusable passkeys are shown and marked rather than hidden, and remain deletable. A domain change that changes the relying party ID warns with a count of affected users, including when a custom domain is deleted, but is never blocked. Configuration edited outside the portal is not covered |
| Q2 | Subsumed by Q1: the relying party ID stays changeable, with a warning, and old credentials keep their original relying party ID in storage — they are unusable, not erased |
| Q7 | Subsumed by Q1: already recorded, and surfaced through the domain-change warning and the unusable marking |
| Q3 | The portal warns when the relying party ID is set to the apex, naming the two files the customer must serve. It does not probe them for reachability |
| Q4 | Passkey configuration is `identity.passkey`, not a new top-level block. Every identity type is configured under `identity`, and none has a top-level block of its own |
| Q5 | `native_apps` is declared under `identity.passkey`, not on each OAuth client and not at a top-level list. The association files it produces are project-wide artefacts served at a fixed path with no client dimension, and a top-level list would need a per-entry capability declaration to say which file each app belongs in |

## Remarks

Noted while writing this document, but belonging to [webauthn.md](./webauthn.md) rather than here.

That document states that passkeys are unsupported on platforms without discoverable credentials, citing a browser limitation from 2022. No such restriction is implemented. `residentKey` is requested as `preferred` rather than `required`, so a non-discoverable credential is a legal outcome and is accepted; the `credProps` extension that reports discoverability is requested but never read; and a user is resolved from the credential ID, which every assertion carries. A non-discoverable credential therefore works whenever the user identifies first, and is simply not offered in the usernameless flow.

Correcting that statement, and the comment in the passkey configuration service that still says Android cannot create discoverable credentials, is a separate change to [webauthn.md](./webauthn.md).
