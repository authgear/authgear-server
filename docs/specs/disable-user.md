# Disable User

  * [Disabling user](#disabling-user)
  * [Disabled user session](#disabled-user-session)
  * [Disabled status](#disabled-status)
  * [Clarification](#clarification)
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

## Clarification

This section aims to clarify some common concepts related to this feature.

### Apple

- [Apple ID can be disabled by the platform.](https://support.apple.com/en-hk/HT204106)
- [Apple ID can be deactivated and reactivated by the end-user.](https://support.apple.com/en-us/HT208503)

### Facebook

- [Facebook account can be disabled by the platform.](https://www.facebook.com/help/103873106370583)
- [Facebook account can be deactivated and reactivated by the end-user.](https://www.facebook.com/help/250563911970368)

### Google

- [Google account can be disabled by the platform.](https://support.google.com/accounts/answer/40695)

### Twitter

- [Twitter account can be suspended by the platform.](https://help.twitter.com/en/managing-your-account/suspended-twitter-accounts)
- [Twitter account can be deactivated and reactivated by the end-user.](https://help.twitter.com/en/managing-your-account/how-to-deactivate-twitter-account)
- [Twitter account can be locked / limited / restricted by the platform.](https://help.twitter.com/en/managing-your-account/locked-and-limited-accounts)

### Auth0

- [Auth0 account can be blocked and unblocked by the developer.](https://auth0.com/docs/manage-users/user-accounts/block-and-unblock-users)

### Summary

- "disable" and "re-enable" are performed by the platform.
- "deactivate" and "reactivate" are performed by the end-user.
- "disable" and "suspend" means the same thing.
- The result of "disable" and "deactivate" is the same. The user cannot log in their account. The cause is very different though.
- Account deletion is usually preceded by account deactivation.
- Locked / limited / restricted account has application-specific meanings, so it is irrelevant to Authgear.

> Discussion: Currently we have a column `is_disabled` of type BOOL on the user table.
> Should we differentiate Disabled Account and Deactivated Account by introducing `is_deactivated`?

> Discussion: Apple ID can also be disabled for security reason like too many failed login attempts.
> Should Authgear consider this use case as disable as well?
> IMO, this use case should be called protected account.

## Future works

- Web-hooks
- Automatic expiry of disable status.
