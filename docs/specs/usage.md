# Overview

Authgear acts as a OIDC provider. Its has several intended usages for different scenarios.

# Scenario 1: Server-side rendered (SSR) website, on the same eTLD+1 of Authgear

In this scenario, the [IdP session](./sessions.md#idp-session) stored in the user agent cookie storage is shared by the website and Authgear.

It is supposed that a gateway in front of the website will initiate subrequest to the resolve endpoint to resolve the session in the HTTP request cookies.

# Scenario 2: Server-side rendered (SSR) website, on different eTLD+1 of Authgear

In this scenario, the website cannot share the cookie with Authgear. Therefore the website must integrate with Authgear by acting as a OIDC RP. Being as a RP, the website must manage its own session. The session management in Authgear web UI settings and Authgear portal have no effect on the sessions of the website.

> TODO: The above limitation may be solved by https://github.com/authgear/authgear-server/issues/315

# Scenario 3: Generic OAuth client

In this scenario, the application is acting as a generic OAuth client. The application performs OIDC Authorizatin code flow with PKCE to obtain a refresh token and an access token.

It is supposed that a gateway in front of the application will initiate subrequest to the resolve endpoint to resolve the session in the HTTP header `Authorization:`.
