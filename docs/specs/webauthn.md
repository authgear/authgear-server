- [WebAuthn](#webauthn)
  * [User Interface Guidelines from Apple](#user-interface-guidelines-from-apple)
    + [AutoFill for passkeys](#autofill-for-passkeys)
  * [Use cases](#use-cases)
    + [Use case 1: Sign up with a passkey](#use-case-1-sign-up-with-a-passkey)
    + [Use case 2: Sign in with a passkey](#use-case-2-sign-in-with-a-passkey)
    + [Use case 3: Add a passkey to the account](#use-case-3-add-a-passkey-to-the-account)
  * [Design decisions](#design-decisions)
    + [WebAuthn credential is both an Identity and a Primary Authenticator](#webauthn-credential-is-both-an-identity-and-a-primary-authenticator)
    + [Passkey is unsupported by Authgear on platform without discoverable credentials](#passkey-is-unsupported-by-authgear-on-platform-without-discoverable-credentials)
    + [The user either set up a password or a passkey during sign-up](#the-user-either-set-up-a-password-or-a-passkey-during-sign-up)
    + [The user can add a password by going through the reset password flow](#the-user-can-add-a-password-by-going-through-the-reset-password-flow)
  * [Future works](#future-works)
    + [Allow the user to remove their password](#allow-the-user-to-remove-their-password)
    + [Ensure passkey support for syncing cross-platform authenticator](#ensure-passkey-support-for-syncing-cross-platform-authenticator)
  * [Configuration](#configuration)
  * [Implementation Details](#implementation-details)
    + [Credential ID](#credential-id)
    + [PublicKeyCredentialCreationOptions](#publickeycredentialcreationoptions)
    + [PublicKeyCredentialRequestOptions](#publickeycredentialrequestoptions)
    + [AuthenticatorAttestationResponse](#authenticatorattestationresponse)
    + [signCount](#signcount)
    + [Cancellation error](#cancellation-error)
    + [Timeout error](#timeout-error)
    + [Error triggered by excludeCredentials](#error-triggered-by-excludecredentials)
    + [User gesture bug of Fetch API on Safari on iOS 15.5](#user-gesture-bug-of-fetch-api-on-safari-on-ios-155)
  * [Appendix](#appendix)
    + [AuthenticatorAttachment behavior on various platforms](#authenticatorattachment-behavior-on-various-platforms)
    + [signCount support](#signcount-support)
    + [Consequence of deleting the credential on the server side](#consequence-of-deleting-the-credential-on-the-server-side)
    + [Consequence of deleting the credential on the authenticator](#consequence-of-deleting-the-credential-on-the-authenticator)

# WebAuthn

This document describes how Authgear makes use of WebAuthn to provide the next-generation authentication experience.

We start with talking about the user interface guidelines from Apple.
With the guidelines in our head, we shape our use cases.
Finally we talk about the design decisions and the implementation details.

## User Interface Guidelines from Apple

The following guidelines are extracted from the [WWDC2022 Meet passkeys](https://developer.apple.com/videos/play/wwdc2022/10092/) video.

- The user-visible term for WebAuthn is "passkey".
- Passkey is a countable common noun.
- The [SF Symbols](https://developer.apple.com/design/human-interface-guidelines/foundations/sf-symbols) for passkey are `person.key.badge` and `person.key.badge.fill`.

### AutoFill for passkeys

For supported platforms, we have to provide AutoFill for passkeys.

- Include [webauthn](https://html.spec.whatwg.org/#attr-fe-autocomplete-webauthn) in the `autocomplete` attribute of a `<input>` element.
- Detect the feature support of [conditional mediation](https://w3c.github.io/webappsec-credential-management/#dom-credentialmediationrequirement-conditional) via [PublicKeyCredential.isConditionalMediationAvailable()](https://w3c.github.io/webappsec-credential-management/#dom-credential-isconditionalmediationavailable).

This implies passkeys are replacement for passwords.
Passkeys DO NOT replace username.
The end-users still have their username for their account.

## Use cases

### Use case 1: Sign up with a passkey

1. The user lands on the sign-up page.
1. The user sees a Login ID input field.
1. The user enters a Login ID, typically an email address.
1. The user is prompted to set up a passkey.
1. The name of passkey is the Login ID. It is not editable by the user.
1. The user goes through the procedure of creating a passkey, as guided by the platform.
1. The user does not need to set up MFA even MFA is required. It is because passkeys are strong enough.
1. The user finishes the sign up. They end up with a Email Login ID identity, a WebAuthn identity, and a WebAuthn primary authenticator.

If the user refuses to set up a passkey, they set up a password instead. This is the same as before.

### Use case 2: Sign in with a passkey

1. The user lands on the sign-in page.
1. The user sees a Login ID input field.
1. Authgear tries to perform AutoFill for passkeys.
1. The user is prompted to choose a passkey to use.
1. The user chooses a passkey.
1. The user goes through the procedure of authenticating the user of the passkey, as guided by the platform.
1. The user finishes the sign in. The user DOES NOT need to perform MFA. It is because passkeys are strong enough.

If AutoFill is unavailable, the following flow is expected.

1. The user lands on the sign-in page.
1. The user sees a Login ID input field.
1. The user enters their Login ID.
1. Authgear knows the account has passkeys and prompt the user to select a passkey to use.
1. The user chooses a passkey.
1. The user goes through the procedure of authenticating the user of the passkey, as guided by the platform.
1. The user finishes the sign in. The user DOES NOT need to perform MFA. It is because passkeys are strong enough.

If the user lost their passkey because their passkey is created on Safari < 16 or other platforms, the following flow is expected.

1. The user lands on the sign-in page.
1. The user sees a Login ID input field.
1. The user enters their Login ID.
1. Authgear knows the account has passkeys and prompt the user to select a passkey to use.
1. The user CANNOT choose a passkey because they have lost access to it.
1. The user cancels the modal.
1. The user sees the option "Lost access to your passkey? Sign in with password instead"
1. The user clicks on the option and see the usual enter-password page.
1. The user did a few attempts to sign in with password. They all result in failure because the user DOES NOT have a password in the first place.
1. The user sees the option "Forgot your password? Reset password"
1. The user goes through the reset password process.
1. The user ends up with "adding" a password to their account.
1. The user recovers their account by themselves.

### Use case 3: Add a passkey to the account

1. The user lands on the settings page.
1. The user enters the passkey settings page.
1. The user click the add button.
1. The name of the passkey is populated by the end-user identifier, and not editable by the user. This is the same experience as Use case 1.
1. Authgear tells the platform all the passkeys the user already has, to prevent creating duplicate passkey.
1. The user ends up with a new passkey.

## Design decisions

### WebAuthn credential is both an Identity and a Primary Authenticator

The first reason is that if we ever need to support a use case of NOT showing the Login ID field for AutoFill,
we can still look up a user when given a passkey because a passkey is an Authgear Identity.

The second reason is to path the way for future support of using a passkey as a new 2FA factor.
When a passkey is used as a 2FA factor, it is a WebAuthn Secondary Authenticator.
Basically it is the same as WebAuthn Primary Authenticator except that its kind is secondary, thus
can only be used in secondary authentication.

### Passkey is unsupported by Authgear on platform without discoverable credentials

Due to the decision that we treat every passkey as both Authgear Identity and Authgear Authenticator,
we DO NOT support passkey on platform without [discoverable credentials](https://www.w3.org/TR/webauthn-2/#discoverable-credential).
This simplifies our implementation because we can assume every passkey is discoverable.

The notable platform that does not support discoverable credentials is [Chrome on Android](#authenticatorattachment-behavior-on-various-platforms).

### The user either set up a password or a passkey during sign-up

This signifies a passkey is a replacement of a password.
Instead of always requiring the user to set up a password,
the user can just set up a passkey only.
This improves security by eliminating a weaker credential from the account.

### The user can add a password by going through the reset password flow

This helps the user who is not using Safari 16 to benefit from using passkeys,
without the downside of permanent loss access to their account.

## Future works

### Allow the user to remove their password

If the user has at least one passkey in their account, the user can remove their password.

### Ensure passkey support for syncing cross-platform authenticator

1Password has expressed their intended support for passkey in their blogpost https://blog.1password.com/1password-is-joining-the-fido-alliance/
As of 2022-07-20, the build 80800215 of 1Password 8 still has no support for passkey that I can test with.
From the video in the blogpost, the authenticator provided by 1Password is a cross-platform authenticator.
It seems that 1Password will take care of syncing the passkeys across devices in their own way, just as iCloud Keychain does.

We should keep an eye of the progress of 1Password passkey support to ensure it works seamlessly with Authgear.

## Configuration

The following configuration are added.

```yaml
authentication:
    identities:
    - login_id
    - oauth
    - webauthn
    primary_authenticators:
    - password
    - webauthn
```

- `authentication.identities` and `authentication.primary_authenticators` : `webauthn` is added. They have to be present or absent at the same time. If `webauthn` comes before other primary authenticators, the user is prompted to set up a passkey first.

## Implementation Details

This section is intended for Authgear implementers.
Casual readers can skip this section.

### Credential ID

The [Credential ID](https://www.w3.org/TR/webauthn-2/#credential-id) is stored in its [Base64url](https://www.w3.org/TR/webauthn-2/#base64url-encoding) encoded form.
It is primarily used in [allowCredentials](https://www.w3.org/TR/webauthn-2/#dom-publickeycredentialrequestoptions-allowcredentials) and [excludeCredentials](https://www.w3.org/TR/webauthn-2/#dom-publickeycredentialcreationoptions-excludecredentials).

In Authgear, credential ID is used to look up an Authgear Identity or an Authgear Authenticator.

### PublicKeyCredentialCreationOptions

[PublicKeyCredentialCreationOptions](https://www.w3.org/TR/webauthn-2/#dictionary-makecredentialoptions) allows us to configure the desired
characteristics of the public key credential we are going to create.

This is an example of the create options we generate

```json
{
    "rp": {
        "id": "myapp.authgearapps.com",
        "name": "Myapp"
    },
    "user": {
        "id": "...",
        "name": "user@example.com",
        "displayName": "user@example.com"
    },
    "challenge": "...",
    "pubKeyCredParams": [...],
    "timeout": 300000,
    "excludeCredentials": [...],
    "authenticatorSelection": {
        "residentKey": "required",
        "requireResidentKey": true,
        "userVerification": "preferred"
    },
    "attestation": "direct",
    "extensions": {
        "uvm": true,
        "credProps": true
    }
}
```

- `rp.id`: The host of Authgear as required by [the spec](https://www.w3.org/TR/webauthn-2/#rp-id).
- `rp.name`: The name of the project.
- `user.name` and `user.displayName`: Pre-filled by the end user identifier. Not editable.
- `authenticatorSelection.authenticatorAttachment`: It is kept [unset](https://www.w3.org/TR/webauthn-2/#dom-authenticatorselectioncriteria-authenticatorattachment) so that the authenticator can be platform or cross-platform.
- `authenticatorSelection.residentKey`: Set to `required` so that the created credential is discoverable.
- `authenticatorSelection.requireResidentKey`: [Deprecated field](https://www.w3.org/TR/webauthn-2/#dom-authenticatorselectioncriteria-requireresidentkey).
- `authenticatorSelection.userVerification`: Apple WWDC video suggests setting to `preferred` for good UX on device without biometric.
- `attestation`: Set to [direct](https://www.w3.org/TR/webauthn-2/#attestation-conveyance) so that we can attest the authenticator.
- `extensions.uvm`: Set to [true](https://www.w3.org/TR/webauthn-2/#sctn-uvm-extension) so that we can know how the user was verified.
- `extensions.credProps`: Set to [true](https://www.w3.org/TR/webauthn-2/#sctn-authenticator-credential-properties-extension) so that we can know the credential properties.

### PublicKeyCredentialRequestOptions

[PublicKeyCredentialRequestOptions](https://www.w3.org/TR/webauthn-2/#dictionary-assertion-options) configures the prompt displayed the platform when we ask for credentials.

This is an example of the get options we generate

```json
{
    "challenge": "...",
    "timeout": 300000,
    "rpId": "myapp.authgearapps.com",
    "allowCredentials": [...],
    "userVerification": "preferred",
    "extensions": {
        "uvm": true
    }
}
```

- `rpId`: The same value we used in the create options.
- `allowCredentials`: For Use Case 1, it is an empty array because we have no idea who is signing in. For Use Case 2, it is an array of credentials owned by the identified user.
- `userVerification`: Apple WWDC video suggests setting to `preferred` for good UX on device without biometric.
- `extensions.uvm`: Set to [true](https://www.w3.org/TR/webauthn-2/#sctn-uvm-extension) so that we can know how the user was verified.

### AuthenticatorAttestationResponse

[AuthenticatorAttestationResponse](https://www.w3.org/TR/webauthn-2/#iface-authenticatorattestationresponse) is the return value of [navigator.credentials.create](https://w3c.github.io/webappsec-credential-management/#dom-credentialscontainer-create).
It contains the public key credential we need to verify signature later.

### signCount

[signCount](https://www.w3.org/TR/webauthn-2/#sctn-sign-counter) is for authenticator cloning detection.
If a signCount mismatch is detected, the authentication is disallowed.
The signCount is updated when authentication succeeds.
See [signCount support](#signcount-support) for details.

In Authgear, signCount is stored in an Authgear Authenticator.
It is not stored in an Authgear Identity.

### Cancellation error

When navigator.credentials.create or navigator.credentials.get is canceled by the end-user, the following error are observed.

||Error being thrown|
|---|---|
|Chrome|`DOMException(code=0, name="NotAllowedError", message="The operation either timed out or was not allowed. See: https://www.w3.org/TR/webauthn-2/#sctn-privacy-considerations-client.")`|
|Safari|`DOMException(code=0, name="NotAllowedError", message="This request has been cancelled by the user.")`|

### Timeout error

When navigator.credentials.create or navigator.credentials.get timed out, the following error are observed.

||Error being thrown|
|---|---|
|Chrome|`DOMException(code=0, name="NotAllowedError", message="The operation either timed out or was not allowed. See: https://www.w3.org/TR/webauthn-2/#sctn-privacy-considerations-client.")`|
|Safari|`DOMException(code=3, name="HierarchyRequestError", message="The operation would yield an incorrect node tree.")`|

### Error triggered by excludeCredentials

When the platform can detect duplicate credential, the following error are observed.

||Error being thrown|
|---|---|
|Safari|`DOMException(code=11, name="InvalidStateError", message="At least one credential matches an entry of the excludeCredentials list in the platform attached authenticator.")`|

### User gesture bug of Fetch API on Safari on iOS 15.5

Safari on iOS 15.5 does not propagate user gesture for Fetch API. Therefore, it is impossible to listen for click and then fetch challenge, finally call navigator.credentials.create. See https://bugs.webkit.org/show_bug.cgi?id=213595 The workaround is to use XMLHttpRequest instead.

## Appendix

### AuthenticatorAttachment behavior on various platforms

||AuthenticatorAttachment: platform|AuthenticatorAttachment: cross-platform|
|---|---|---|
|Chrome 103 on macOS 12.4|The user can choose "This Device". Credential can be clear with `Settings -> Privacy and security -> Clear browsing data -> Advanced -> Check password and other sign-in data`|Must use security key|
|Safari 15.5 on macOS 12.4|The user can continue with Touch ID. Credential can be clear with `History -> Clear History ... -> Clear All History`.|Must use security key|
|Edge 103 on macOS 12.4|The user can choose "This Device". Credential can be clear with `Settings -> Privacy, Search and Services -> Clear browsing data -> Check Passwords`.|Must use security key|
|Safari on iOS 15.5|The user can continue with Face ID or Touch ID. Credential can be clear with `Settings -> Safari -> Clear History and Website Data`|Must use security key|
|Safari on iOS 16 Beta 3|The platform authenticator becomes iCloud Keychain only. The credential can be found in `Settings -> Passwords`. The credential can be viewed and deleted individually|Must use security key|
|ASWebAuthenticationSession on iOS 15.5|Same as Safari|Same as Safari|
|Chrome 103 on Android 12|See below|See below|

Chrome 103 on Android 12 does not support [residentKey=required](https://www.w3.org/TR/webauthn-2/#dom-authenticatorselectioncriteria-residentkey) in `navigator.credentials.create`.
It throws this error `DOMException(code=9, name="NotSupportedError", message="Either the device has received unexpected request parameters, or the device cannot support this request.")`.
Chrome 103 on Android 12 also does not support empty [allowCredentials](https://www.w3.org/TR/webauthn-2/#dom-publickeycredentialrequestoptions-allowcredentials) in `navigator.credentials.get`.
It throws this error `DOMException(code=9, name="NotSupportedError", message="Use of an empty `allowCredentials` list is not supported on this device.")`.
See https://github.com/w3c/webauthn/issues/1457#issuecomment-692701186

### signCount support

Platform authenticators of Chrome and Safari always return a signCount of 0.
Yubikey supports signCount.

### Consequence of deleting the credential on the server side

|navigator.credentials.create|navigator.credentials.get|
|---|---|
|Create another key|Send an unknown credential to server to trigger server error|

### Consequence of deleting the credential on the authenticator

|navigator.credentials.create|navigator.credentials.get|
|---|---|
|Create a new key|Nothing to select|
