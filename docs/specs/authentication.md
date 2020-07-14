# Authentication Flow

  * [Authentication](#authentication)
  * [Interaction](#interaction)
    * [Login intent](#login-intent)
    * [Signup intent](#signup-intent)
    * [Add Identity intent](#add-identity-intent)

## Authentication

- The developer can configure enabled identity types. By default, all supported identity types are enabled.
- The developer can configure enabled primary authenticators. By default, Password Authenticator is enabled.
- The developer can configure enabled secondary authenticators. By default, TOTP, OOB-OTP and Bearer Token Authenticator are enabled.
- The developer can configure whether secondary authentication is necessary.
  - `required`: secondary authentication is required. Every user must have at least one secondary authenticator.
  - `if_exists`: secondary authentication is opt-in. If the user has at least one secondary authenticator, then the user must perform secondary authentication.
  - `if_requested`: secondary authentication is purely optional even the user has at least one secondary authenticator.

## Interaction

Manipulation of user, identities and authenticators are driven by interaction. An interaction starts with an intent and has various steps. When all required steps have been gone through, the interaction is committed to the database.

### Login intent

The login intent authenticate existing user. It involves the following steps:

- Select identity
- Authenticate with primary authenticator
- Authenticate with secondary authenticator / Setup secondary authenticator

For example,

Login with login ID and password

- Select identity by providing a login ID
- Authenticate with password

### Signup intent

The signup intent creates a new user. It involves the following steps:

- Create identity
- Setup primary authenticator
- Setup secondary authenticator

For example,

Login in with Google

- Create identity by perform OIDC authorization code flow with Google
- No primary authenticator is required

### Add Identity intent

The add identity intent adds a new identity to a user. It involves the following steps:

- Create identity
- Setup primary authenticator

For example,

Add Email login ID to a user with 1 OAuth Identity

- The user provides the email address
- Setup OOB-OTP authenticator of the given email address
