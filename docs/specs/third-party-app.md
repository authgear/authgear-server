# Third-Party App (`third_party_app`)

A `third_party_app` is an OAuth client operated by an external party — a developer who is not a collaborator of the Authgear project. It uses the Authorization Code flow to authenticate users and must obtain explicit user consent before accessing protected resources.

Depending on the chosen `token_endpoint_auth_method`, a `third_party_app` may be either confidential (with a `client_secret`) or public (PKCE-only, no `client_secret`).

## Table of Contents

- [Overview](#overview)
- [Differences from First-Party Clients](#differences-from-first-party-clients)
- [Registration](#registration)
  - [Static registration](#static-registration)
  - [Dynamic registration (DCR)](#dynamic-registration-dcr)
- [Supported Flows](#supported-flows)
- [Consent Screen](#consent-screen)
- [Token Endpoint Authentication](#token-endpoint-authentication)
- [Scopes](#scopes)
- [PII in ID Token](#pii-in-id-token)
- [Sessions and Authorization Management](#sessions-and-authorization-management)
- [Client Configuration](#client-configuration)

## Overview

| Property                   | Value                                                            |
| -------------------------- | ---------------------------------------------------------------- |
| `x_application_type`       | `third_party_app`                                                |
| Party                      | Third-party                                                      |
| Confidentiality            | Confidential or public (depends on `token_endpoint_auth_method`) |
| Consent screen             | Required                                                         |
| PII in ID token            | Yes                                                              |
| `full-access` scope        | Not allowed                                                      |
| `full-userinfo` scope      | Not allowed                                                      |
| `client_credentials` grant | Not allowed                                                      |
| Privileged user operations | Not allowed                                                      |

## Differences from First-Party Clients

**Trust model.** First-party clients are created by project collaborators and are trusted implicitly — the consent screen is skipped. Third-party apps are created by external developers and are not trusted — users must grant explicit consent.

**Scope access.** Third-party apps may only request a subset of the available scopes. See [Scopes](#scopes) for the full list.

**PII in ID token.** Because third-party apps cannot trigger reauthentication (which would embed the ID token in a URL), it is safe to include PII (email, phone, profile claims) directly in the ID token. See [PII in ID Token](#pii-in-id-token).

**Session management.** First-party client sessions appear in the Sessions list; revoking one terminates that specific session. Third-party app authorizations are tracked separately as per-client grants — revoking an authorization removes all refresh tokens issued to that client for the user, but does not affect the user's login session with Authgear.

## Registration

### Static registration

Static clients are configured in `authgear.yaml` by project collaborators. A third-party app requires:

- `x_application_type: third_party_app`
- `client_name`: the human-readable name shown on the consent screen (required for `third_party_app`)
- `redirect_uris`: at least one redirect URI
- `client_id`: assigned by Authgear

For confidential clients (default), a `client_secret` is also required and stored in `authgear.secrets.yaml`. Public clients (`token_endpoint_auth_method: none`) do not use a `client_secret`.

Example `authgear.yaml` entry:

```yaml
oauth:
  clients:
    - name: "My App"
      client_id: "my-third-party-app"
      x_application_type: third_party_app
      client_name: "My App"
      client_uri: "https://myapp.example.com"
      redirect_uris:
        - "https://myapp.example.com/callback"
      policy_uri: "https://myapp.example.com/privacy"
      tos_uri: "https://myapp.example.com/terms"
      logo_uri: "https://myapp.example.com/logo.png"
```

The corresponding `authgear.secrets.yaml` entry (confidential clients only):

```yaml
- data:
    items:
      - client_id: my-third-party-app
        keys:
          - created_at: 1136171045
            k: <base64url-encoded-secret>
            kid: <key-id>
            kty: oct
  key: oauth.client_secrets
```

### Dynamic registration (DCR)

Third-party apps are the default client type for Dynamic Client Registration. When DCR is enabled with `initial_access_token_required: false`, only `third_party_app` may be registered without an Initial Access Token. See [DCR spec](./dcr.md) for full details.

## Supported Flows

| Grant type           | Supported                                      |
| -------------------- | ---------------------------------------------- |
| `authorization_code` | Yes                                            |
| `refresh_token`      | Yes (when `offline_access` scope is requested) |
| `client_credentials` | No                                             |

Third-party apps use the standard Authorization Code flow with PKCE (RFC 7636).

## Consent Screen

The consent screen is shown when a user authorizes a third-party app for the first time, or after the user revokes the previous authorization. It displays the `client_name`, the requested permissions, and optionally `policy_uri` and `tos_uri`.

The consent screen is **not** shown if there is an existing valid authorization record for the same user, client, and scope set.

Example (when `offline_access` is requested):

```
Authorize <client_name>

<client_name> wants to access your account.

- Allows <client_name> to access your information after login.

[Cancel] [Authorize]
```

The permission descriptions shown depend on the requested scopes:

| Scope            | Consent description                           |
| ---------------- | --------------------------------------------- |
| `offline_access` | Allows access to your information after login |

Standard OIDC scopes (`profile`, `email`, `phone`, `address`) do not generate individual permission descriptions on the consent screen.

## Token Endpoint Authentication

The `token_endpoint_auth_method` determines whether the client is confidential or public.

| `token_endpoint_auth_method` | Supported | Notes                                                       |
| ---------------------------- | --------- | ----------------------------------------------------------- |
| `client_secret_post`         | Yes       | Default. `client_secret` is sent as a POST body parameter   |
| `none`                       | Yes       | Public client — no `client_secret` issued; PKCE is required |
| `client_secret_basic`        | No        |                                                             |

When `none` is used, the client relies on PKCE for proof of possession. This is suitable for native or mobile third-party apps that cannot safely store a client secret.

## Scopes

| Scope                                               | Allowed | Notes                                    |
| --------------------------------------------------- | ------- | ---------------------------------------- |
| `openid`                                            | Yes     | Required                                 |
| `profile`                                           | Yes     |                                          |
| `email`                                             | Yes     |                                          |
| `phone`                                             | Yes     |                                          |
| `address`                                           | Yes     |                                          |
| `offline_access`                                    | Yes     | Issues a refresh token                   |
| `https://authgear.com/scopes/full-userinfo`         | **No**  | Exposes internal identities and authenticators; standard OIDC scopes are sufficient |
| `https://authgear.com/scopes/full-access`           | **No**  | Restricted to first-party public clients |
| `device_sso`                                        | **No**  | Restricted to clients with pre-authenticated URL enabled |
| `https://authgear.com/scopes/pre-authenticated-url` | **No**  | Restricted to clients with pre-authenticated URL enabled |

## PII in ID Token

Third-party apps receive PII (personally identifiable information) in their ID tokens, unlike first-party public clients.

**Rationale:** First-party public clients have access to voluntary reauthentication, which passes an `id_token_hint` in the URL. Including PII in that ID token would expose it in the URL and browser history. Third-party apps have no access to reauthentication, so it is safe to include PII in their ID tokens.

The claims included in the ID token depend on the granted scopes:

- `email` → `email`, `email_verified`
- `phone` → `phone_number`, `phone_number_verified`
- `profile` → standard profile claims (name, given_name, family_name, etc.)

## Sessions and Authorization Management

> **Not yet implemented.**

The **Sessions** page (`/settings/sessions`) lists IdP sessions and first-party client sessions only. Third-party app authorizations are intentionally excluded because revoking a first-party session terminates the session itself, whereas revoking a third-party authorization only removes that app's access tokens — the user's login session with Authgear is unaffected.

The **Authorized Apps** page (`/settings/authorized-apps`) lists per-client authorizations for third-party apps. Each entry shows the client name and the scopes granted. Revoking an authorization deletes all refresh tokens issued to that client for the current user.

## Client Configuration

`third_party_app` supports a subset of the OAuth client config fields available to first-party clients. See [Custom Client Metadata](./oidc.md#custom-client-metadata) for the full reference.

Supported fields:

- `client_id`: Required. Unique identifier assigned to the client.
- `name`: Required. Internal display name shown in the portal.
- `client_name`: Required. Human-readable name shown to end-users on the consent screen (OIDC standard field).
- `redirect_uris`: Required. One or more allowed redirect URIs.
- `x_application_type`: Required. Must be `third_party_app`.

- `client_uri`: Optional. URI of the client's home page, shown on the consent screen.
- `logo_uri`: Optional. URI of the client's logo, shown on the consent screen. Must be a public HTTPS URL.
- `policy_uri`: Optional. URI of the client's privacy policy, shown on the consent screen.
- `tos_uri`: Optional. URI of the client's terms of service, shown on the consent screen.

- `access_token_lifetime_seconds`: Optional. Duration in seconds. Minimum 300.
- `refresh_token_lifetime_seconds`: Optional. Duration in seconds.
- `refresh_token_idle_timeout_enabled`: Optional. Boolean.
- `refresh_token_idle_timeout_seconds`: Optional. Duration in seconds.
- `issue_jwt_access_token`: Effective value is always `true`; any configured value is ignored. Access tokens are always issued as signed JWTs (RFC 9068) so that third-party resource servers can validate them independently.

- `grant_types`: Optional. Defaults to `["authorization_code", "refresh_token"]`.
- `response_types`: Optional. Defaults to `["code"]`.
- `post_logout_redirect_uris`: Optional. Allowed URIs for post-logout redirect.
