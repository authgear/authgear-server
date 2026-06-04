# Third-Party Clients

A third-party client is an OAuth client operated by an external developer — someone who is not a collaborator of the Authgear project. Because the operator is not trusted by the project, users must grant explicit consent before the client may access their account.

> **Note on `x_application_type: third_party_app`:** Statically configuring an OAuth client with `x_application_type: third_party_app` in `authgear.yaml` is deprecated. Existing OAuth clients of this type remain functional but are not recommended for new integrations. New third-party clients must be created via Dynamic Client Registration (DCR). See [DCR spec](./dcr.md).

## Table of Contents

- [First-Party vs Third-Party Clients](#first-party-vs-third-party-clients)
- [Registration](#registration)
- [Consent Screen](#consent-screen)
- [Supported Flows](#supported-flows)
- [Scopes](#scopes)
- [Token Behavior](#token-behavior)
- [Sessions and Authorization Management](#sessions-and-authorization-management)

## First-Party vs Third-Party Clients

| Property | First-Party Client | Third-Party Client |
|---|---|---|
| Operated by | Project collaborator | External developer |
| Consent screen | Not shown | Required |
| Access token without resource indicator | JWT (`aud = [<project_endpoint>]`) | Opaque |
| `client_credentials` grant | Allowed (M2M clients) | Not allowed |
| `https://authgear.com/scopes/full-access` | Allowed (public clients only) | Not allowed |
| `https://authgear.com/scopes/full-userinfo` | Allowed | Not allowed |
| PII in ID token | No (public clients) / Yes (confidential clients) | Yes |
| Session management | Sessions page | Authorized Apps page |
| Registration | `authgear.yaml` or DCR | DCR only |

**Public vs confidential clients.** A first-party client is *public* if it has no `client_secret` (e.g. `spa`, `native` application types) and *confidential* if it does (e.g. `confidential` type). The `https://authgear.com/scopes/full-access` scope grants access to voluntary reauthentication and app session token exchange (`/oauth2/app-session-token`); it is restricted to public clients because they are the only client type that supports these operations.

**Trust model.** First-party clients are created by project collaborators and are implicitly trusted — the consent screen is skipped. Third-party clients are created by external developers and are not trusted — users must explicitly grant access on the consent screen.

**PII in ID token.** First-party public clients have access to voluntary reauthentication, which passes an `id_token_hint` in the URL; including PII there would expose it in browser history. Third-party clients and first-party confidential clients have no access to reauthentication, so it is safe to include PII in their ID tokens.

## Registration

Third-party clients can only be created via Dynamic Client Registration (DCR). When DCR is enabled with `initial_access_token_required: false`, any caller may register a `third_party_app` client without presenting an Initial Access Token.

See [DCR spec](./dcr.md) for the full registration flow.

## Consent Screen

The consent screen is shown when a user authorizes a third-party client for the first time, or after the user has revoked a previous authorization. It displays the client name, the requested permissions, and optionally a privacy policy and terms of service link.

The consent screen is **not** shown if there is an existing valid authorization record for the same user, client, and scope set.

## Supported Flows

| Grant type | Supported |
|---|---|
| `authorization_code` | Yes |
| `refresh_token` | Yes (when `offline_access` is granted) |
| `client_credentials` | No |

## Scopes

| Scope | Allowed | Notes |
|---|---|---|
| `openid` | Yes | Required |
| `profile` | Yes | |
| `email` | Yes | |
| `phone` | Yes | |
| `address` | Yes | |
| `offline_access` | Yes | Issues a refresh token |
| `https://authgear.com/scopes/full-userinfo` | **No** | Exposes internal identities and authenticators; standard OIDC scopes are sufficient |
| `https://authgear.com/scopes/full-access` | **No** | Grants privileged user operations (e.g. app session token); restricted to first-party public clients |
| `device_sso` | **No** | Restricted to clients with pre-authenticated URL enabled |
| `https://authgear.com/scopes/pre-authenticated-url` | **No** | Restricted to clients with pre-authenticated URL enabled |

## Token Behavior

By default (no `resource` parameter), a third-party client receives an **opaque** access token. The opaque token can only be used with the userinfo endpoint and cannot be validated independently by a resource server.

When the `resource` parameter is specified and the referenced Resource permits access, a **JWT** access token is issued with `aud` set to the resource URI only.

See [Access Token Audience Binding](./access-token-audience-binding.md) for the full specification.

## Sessions and Authorization Management

Third-party client authorizations are tracked separately from IdP sessions.

The **Sessions** page (`/settings/sessions`) lists IdP sessions and first-party client sessions only. Third-party client authorizations are excluded because revoking a first-party session terminates the session itself, whereas revoking a third-party authorization only removes that client's access tokens — the user's login session with Authgear is unaffected.

The **Authorized Apps** page (`/settings/authorized-apps`) lists per-client authorizations for third-party clients. Each entry shows the client name and the granted scopes. Revoking an authorization deletes all refresh tokens issued to that client for the current user.
