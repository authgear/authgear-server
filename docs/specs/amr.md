# Supported Authentication Method Reference

This document outlines the supported Authentication Method Reference (AMR) values and their corresponding authenticators.

To understand AMR, read https://www.rfc-editor.org/rfc/rfc8176.html.

## Supported AMR Values

The following AMR values are supported by Authgear:

| AMR Value                   | Meaning                                                                                                                                 |
| --------------------------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| `pwd`                       | Password-based authentication.                                                                                                          |
| `otp`                       | One-time password (OTP) authentication.                                                                                                 |
| `sms`                       | SMS-based authentication.                                                                                                               |
| `mfa`                       | Multi-factor authentication; Added when multiple authenticators are used in a single flow, OR one authenticator with one recovery code. |
| `x_biometric`               | Biometric authentication.                                                                                                               |
| `x_passkey`                 | Indicates passkey authentication.                                                                                                       |
| `x_primary_password`        | Indicates primary password authentication.                                                                                              |
| `x_primary_oob_otp_email`   | Indicates primary one-time password (OTP) authentication via email.                                                                     |
| `x_primary_oob_otp_sms`     | Indicates primary one-time password (OTP) authentication via SMS.                                                                       |
| `x_primary_passkey`         | Indicates passkey authentication.                                                                                                       |
| `x_secondary_password`      | Indicates secondary password authentication.                                                                                            |
| `x_secondary_oob_otp_email` | Indicates secondary one-time password (OTP) authentication via email.                                                                   |
| `x_secondary_oob_otp_sms`   | Indicates secondary one-time password (OTP) authentication via SMS.                                                                     |
| `x_secondary_totp`          | Indicates secondary Time-based One-time Password (TOTP) authentication.                                                                 |
| `x_recovery_code`           | Indicates authentication with a recovery code.                                                                                          |
| `x_device_token`            | Indicates authentication with a device token.                                                                                           |

## Relationship of Authentication and AMR Values

This table documents the relationship between a `authentication` option and AMR values.

| Authentication            | AMR Values                              | Description                                                                           |
| ------------------------- | --------------------------------------- | ------------------------------------------------------------------------------------- |
| `primary_password`        | `pwd`, `x_primary_password`             | User authenticates with their primary password                                        |
| `primary_oob_otp_email`   | `otp`, `x_primary_oob_otp_email`        | User authenticates with a one-time password sent to their email identity              |
| `primary_oob_otp_sms`     | `otp`, `sms`,`x_primary_oob_otp_sms`    | User authenticates with a one-time password sent to their phone identity via SMS      |
| `primary_passkey`         | `x_passkey`, `x_primary_passkey`        | User authenticates with a passkey                                                     |
| `secondary_password`      | `pwd`, `x_secondary_password`           | User authenticates with their secondary password                                      |
| `secondary_oob_otp_email` | `otp`, `x_secondary_oob_otp_email`      | User authenticates with a one-time password sent to a configured email address        |
| `secondary_oob_otp_sms`   | `otp`, `sms`, `x_secondary_oob_otp_sms` | User authenticates with a one-time password sent via SMS to a configured phone number |
| `secondary_totp`          | `otp`, `x_secondary_totp`               | User authenticates with a TOTP                                                        |
| `recovery_code`           | `x_recovery_code`                       | User authenticates with a Recovery Code                                               |
| `device_token`            | `x_device_token`                        | User authenticates with a Device Token                                                |

## Notes

- The AMR values are included in the ID token claims to indicate which authentication methods were used during the authentication process.
- Auth0 uses `phr` (phishing-resistant) to represent passkey authentication. https://auth0.com/blog/all-you-need-to-know-about-passkeys-at-auth0/. In Authgear, however, we do not use this value for passkey authentication.
