# Sessions

Sessions in Authgear are stateful. The user can manage sessions. The developer can configure session characteristics such as lifetime and idle timeout.

  * [Session Attributes](#session-attributes)
  * [Session Management](#session-management)
  * [Session token](#session-token)
  * [Session Types](#session-types)
    * [IdP Session](#idp-session)
    * [Offline Grant](#offline-grant)
  * [Concurrent Sessions](#concurrent-sessions)
    * [Project-wise concurrent sessions](#future-works-project-wise-concurrent-sessions)
    * [Client-wise concurrent sessions](#client-wise-concurrent-sessions)

## Session Attributes

Session has the following attributes:

- ID
- User ID
- AMR
- Creation Time
- Last Access Time
- Creation IP
- Last Access IP
- User Agent

In particular, session does not have reference to involved identity and authenticators in the authentication. Removal of identity and authenticators does not invalidate session.

## Session Management

The user manages their sessions in the settings page. They can list the sessions and revoke them.

TODO(session): support session name. Default session name should be device name.

## Session token

Session token must be treated as opaque string.

## Session Types

### IdP Session

When the user authenticates successfully, an IdP session is created.

Idp session has a configurable lifetime. IdP session may optionally have idle timeout. The session must be accessed before the timeout, or the session is expired.

IdP session token is stored in the user agent cookie storage. The cookie domain attribute is configurable. The default value is eTLD + 1. As long as the web application is under the same domain with Authgear, the IdP session is shared across between the two. The cookie is a persistent cookie by default. The cookie is http-only and is not configurable. The cookie is SameSite=lax and is not configurable. The cookie is secure by default.

The IdP session configuration is global.

### Offline Grant

Each OAuth client has its own configuration of offline grant.

Offline grant consists of a refresh token and access token. As long as the refresh token remains valid, access tokens can be refreshed with the refresh token independent of the IdP session. Offline grant is intended for use in native application.

If the refresh token is used when the OAuth client is deleted, the refresh token is considered invalid.
The SDK will log the user out in that case.
However, if the OAuth client is restored, subsequent usage of refresh token will NOT be affected.

Since offline grant must always reference to its OAuth client, changes in the configuration of the OAuth client will take effect immediately.

Access token has a configurable lifetime.

Refresh token has a configurable lifetime. Refresh token may optionally have idle timeout. The session must be accessed before the timeout, or the session is expired.
The old access token is invalidated during refresh. At any time there is at most one valid access token.

The lifetime of offline grant is the lifetime of its refresh token.

## Concurrent Sessions

### (Future Works) Project-wise concurrent sessions

The developer can specify the maximum number of concurrent sessions per project.

When it is specified, sessions will be considered **in terms of SSO group**. Excessive session
groups will be terminated when the user authenticates.

#### Use case 1: When the maximum number of concurrent sessions per project is 1 and SSO is enabled

1. The end-user logs in to multiple apps and websites on Device A.
1. The end-user logs in to a website on Device B.
1. The maximum number of concurrent sessions per project is 1, so the session group on device A will be revoked. All apps on device A will be logged out.
1. The end-user logs in an app on Device B, the website on Device B will remain logged in since they are in the same SSO group.

#### Use case 2: When the maximum number of concurrent sessions per project is 1 and SSO is disabled
1. The end-user logs in to the app on Device A.
1. The end-user logs in to the website on Device A. Since SSO is disabled, those sessions are considered in different groups. The app will be logged out.

### Client-wise concurrent sessions

The developer can specify the maximum number of concurrent sessions per OAuth client.

When it is specified, sessions will be considered **in terms of refresh token**. Excessive refresh tokens will be terminated when the user authenticates.

*Client-wise concurrent sessions* cannot restrict the number of sessions for the cookies-based website.

Configurations see [Custom Client Metadata](./oidc.md#custom-client-metadata)

#### Use case 1: When the maximum number of concurrent sessions per OAuth client is 1 and SSO is enabled

1. The end-user logs in to multiple apps on Device A
    - Log in to App A with OAuth Client A, IdP session is generated on Device A
    - Log in to App B with OAuth Client B by clicking continue on the continue screen
    - Log in to Token-based website C with OAuth Client C by clicking continue on the continue screen
1. The end-user logs in on Device B
    - Log in to App A with OAuth Client A, IdP session is generated on Device B
1. The active sessions of the end-user at this point:
    - IdP session on Device A
    - IdP session on Device B
    - App A refresh token on Device B (App A refresh token on Device A is terminated due to the concurrent session limit)
    - App B refresh token on Device A
    - Website C refresh token on Device A
1. If the end-user wants to use App A on Device A, they can call authenticate again. The continue screen will be shown, and App A on Device B will be logged out if they choose to continue.

#### Use case 2: When the maximum number of concurrent sessions per OAuth client is 1 and SSO is disabled

1. The end-user logs in to multiple apps on Device A, they will need to input credentials for every login. No IdP session will be generated.
    - Log in to App A with OAuth Client A
    - Log in to App B with OAuth Client B
    - Log in to Token-based website C with OAuth Client C
1. The end-user logs in on Device B
    - Log in to App A with OAuth Client A
1. The active sessions of the end-user at this point:
    - App A refresh token on Device B (App A refresh token on Device A is terminated due to the concurrent session limit)
    - App B refresh token on Device A
    - Website C refresh token on Device A
