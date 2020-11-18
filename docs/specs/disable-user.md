# Disable User

  * [Disabling user](#disabling-user)
  * [Disabled user session](#disabled-user-session)
  * [Disabled status](#disabled-status)
  * [Future works](#future-works)

## Disabling user

A user can be disabled by admins. A disabled user cannot login, and appropriate
error message would be shown when login is attempted.

Admin may optionally provide a reason when disabling user. This reason would be
shown when user attempted to login.

When a disabled user attempts to log in, the user would be informed of disabled
status only after performing all authentication process, including MFA if
required.

## Disabled user session

When a user is disabled, the user can no longer create new sessions. However,
there may be existing logged in sessions.

The authentication tokens (i.e. sessions, access token, refresh tokens, etc.)
would be treated as invalidated. However, these tokens may still be manipulated
through admin API, and would become valid again if user is re-enabled.

## Disabled status

The disabled status of a user can be accessed or changed through admin API.
However, it would not be visible in OIDC ID token, since OAuth flow cannot be
performed by a disabled user.

Resolver would also not return the disabled status of user in HTTP headers;
instead, the sessions of disabled users would be reported as invalid.

## Future works

- Web-hooks
- Automatic expiry of disable status.
