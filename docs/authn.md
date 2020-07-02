# Authentication & OIDC mechanism

## Introduction

Auth Gear act as an IdP. End-user can log in to Auth Gear using
different means:
- Using Auth UI directly.
- Using app with Auth API.
- Using first-party app through OIDC framework.
- Using third-party app through OIDC framework.

After logging in to Auth Gear, an IdP session is created.
This session is produced and consumed by Auth Gear and other components,
including other gears, gateway, and developer apps, are not supposed to
consume it directly.

Auth Gear also implements OIDC framework. Other components and obtain access
token and ID token (and optionally refresh token) of a user through OIDC
authorization process. In the process, Access Grant (and optionally Offline
Grant) is created.

## Session

A session is represented in code as `core/authn.Session`. It represents
a stateful session of user interaction. In Auth Gear, there are three
implementation of session:
- `auth/session.IdPSession`: represents an IdP session.
- `auth/oauth.OfflineGrant`: represents an OAuth refresh token.
- `core/authn.Info`: a representation of one of the above implementation,
                     for inter-service communication.

Sessions have a set of common attributes (`core/authn.Attrs`). They implements
`core/authn.Attributer` to provide the attributes. The attributes includes:
- User ID
- Principal data
- Authenticator data

Additionally, both real session implementation (`IdPSession` and
`OfflineGrant`) conforms to `auth/auth.AuthSession` to provide additional
functions used in Auth Gear:
- API models conversion
- Read/Update access event records
- etc.

### authn.Info

`core/authn.Info` is a bundle of session information, representing a real
session implementation. It is used to convey session information across service
boundary.

In additional common session attributes, it also contains useful user info for
consumption by services.

### Context

`core/authn` package provides a session context. The session contexts contains
user session info (`core/authn.Session`), a user info
(`core/authn.UserInfo`), and validity of session (a boolean).

The session context can be in three kind of state:
- No session is provided: session and user is `nil`, valid.
- A session is resolved: session and user is not `nil`, valid.
- A session cannot be resolved: session and user is `nil`, invalid.

A session context can be invalid if the session token/access token cannot be
resolved/expired. This is different from not providing any tokens.

### Middleware

`auth/auth` package provides a session middleware. This middleware is used in
Auth Gear only. It performs two process:
- Resolve session token/access token to a session implementation and
  populate the session context.
- Update session access event records.

For other components (e.g. other gears), the middleware in
`core/gears/middleware` package is used instead. It parses a `core/authn.Info`
from incoming request headers and populate the session context.

### Session resolution

When the gateway received a request, it will forward the request headers to
resolver endpoint of Auth Gear (`/_auth/session/resolve`). The endpoint would
read the session context (populated by middleware), convert it to
`core/authn.Info`, and serialize the info bundle into response headers.

Then, gateway would replace the Authgear headers of the incoming requests with
the resolver endpoint response header, and forward it to other gear/services.
Note that requests to Auth Gear is not using this flow, and is passed directly
to Auth Gear.

Other gears/services can then deserialize a `core/authn.Info` from the request
headers into the session context using the gear middleware.

## OAuth/OIDC

To model OAuth/OIDC, two concepts is introduced:
- Authorization: represents a consent of a user given a specific OAuth client,
                 to access resources within specific scopes.
- Grants: a credential that allow access to resources according to an
          Authorization.

### Authentication

OAuth/OIDC framework assumes user is authenticated. Therefore, if user is
unauthenticated, user is redirected to Auth UI to perform authentication before
returning to OAuth flow. OAuth module and Auth UI module has no special
knowledge about each other: they integrate using redirect URI.

### Authorization

When an OAuth client request authorization from a user, a consent page
is shown if:
- user had not given consent to the client previously; or
- new scopes is requested, not presented in existing authorization; or
- a refresh token is requested.

The authorization is persisted in database. It can be revoked (NYI), and
all associated grants would be invalidated.

### Grants

After authorization, grants may be requested. Most commonly, an authorization
code (Code Grant) is requested as stepping stone. Then, the real grant is
requested:
- Access Grant: represents an access token.
- Offline Grant: represents a refresh token.

Offline Grant is created only if offline_access scope is requested, and the
client is allowed to use refresh tokens. It behaves very similar to an IdP
session, but is associated with an authorization: it would be invalidated
upon revocation.

Access Grant is associated with a session:
- If refresh token is requested, it is assoicated with the created offline
  grant.
- Otherwise, it is associated with the IdP session of user.

Access Grant is invalidated when the associated session is invalidated.
Therefore, an access token would be invalidated if user logged out from
the IdP, if refresh token is not requested.

## Auth API

Auth API is integrated with the session & OIDC mechanism. However, it must
be explicitly enabled in configuration.

A OAuth client ID must be provided for authentication as a 'API Key'.

If the client is configured to use cookie in Auth API, an IdP session would
be created, and session token would be written to cookie.

If the client is configured to use tokens in Auth API, an Offline Grant and
Access Grant would be created, and the tokens would be returned in response
body.

The tokens created by Auth API and Auth UI is fully compatible, since the
underlying authentication mechanism is the same.

## Authentication Session

Authentication Session (AS) represents an intermediate state between not
authenticated and authenticated: user is required to perform additional action
before considered as fully authenticated. It is currently used for various
purposes:

- Perform additional authentication (e.g. MFA)
- Setup required additional authentication
- etc.

An AS consists of a series of steps and common session attributes. User is
required to step through the steps to populate the session attributes. Upon
completion of the steps, a real session is created from the contained
information.

The AS is represented by `auth/authn.Session`. All session creation should
go through AS process to ensure consistent behavior.

Some Auth API accepts both a user session and authentication session.
For example, MFA authenticate API behaves differently:
- For user session, it step up the session using the new authenticator.
- For authentication session, it populates the authenticator info.

For convenience of implementing these APIs, the AS type also conforms to
`core/authn.Attributer` interface to enable easy read/update of sesssion
attributes.
