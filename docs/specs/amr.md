# Supported Authentication Method Reference

This document documents the supported Authentication Method Reference (AMR) values, and the corresponding authenticators.

To understand AMR, read https://www.rfc-editor.org/rfc/rfc8176.html.

## Supported AMR Values

The following AMR values are supported:

| AMR Value     | Description                      | Authenticator Type                      |
| ------------- | -------------------------------- | --------------------------------------- |
| `pwd`         | Password-based authentication    | `password`                              |
| `otp`         | One-time password authentication | `totp`, `oob_email`, `oob_sms`          |
| `sms`         | SMS-based authentication         | `oob_sms`                               |
| `mfa`         | Multi-factor authentication      | Added when multiple authenticators used |
| `x_biometric` | Biometric authentication         | `biometric`                             |
| `x_passkey`   | Passkey authentication           | `passkey`                               |

## Notes

- The `mfa` AMR value is automatically added when multiple authenticators are used in a single authentication flow.
- The `x_biometric` and `x_passkey` values are custom AMR values that are not defined in RFC 8176, but are used to represent biometric and passkey authentication methods respectively.
- The AMR values are included in the ID token claims to indicate which authentication methods were used during the authentication process.
- Auth0 is using `phr` to represent passkey authentication. https://auth0.com/blog/all-you-need-to-know-about-passkeys-at-auth0/
