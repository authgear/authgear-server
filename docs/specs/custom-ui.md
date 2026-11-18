- [Custom UI](#custom-ui)
  - [Configuration](#configuration)
  - [Pages](#pages)
  - [Resolution](#resolution)
  - [Origins](#origins)

# Custom UI

A Custom UI is a UI hosted by the developer that replaces pages of Auth UI, by using the [Authentication Flow API](./authentication-flow-api-reference.md).

## Configuration

A project sets the pages of its Custom UI, and a static client can override them key by key.

```yaml
ui:
  custom_ui_endpoints:
    authentication: "https://custom.app/auth"
    login_link: "https://custom.app/login-link"
    account_recovery: "https://custom.app/reset"

oauth:
  clients:
  - client_id: client_b
    x_custom_ui_endpoints:
      authentication: "https://ui-b.example.com/auth"
```

Every key is optional, and its value is an absolute URI. A key that is not set falls back as described in [Resolution](#resolution).

`x_custom_ui_uri` is deprecated. It means the same as `x_custom_ui_endpoints.authentication`, which takes precedence when both are set.

## Pages

| Key | Authgear sends the end-user there when | Query parameters added | The page |
|---|---|---|---|
| `authentication` | An [authentication request](./oidc.md) is received. | Parameters of the authentication request needed to create a flow. | Creates an Authentication Flow. |
| `login_link` | The end-user opens a [login link](./login-link.md). | `code` | Is the [approval page](./login-link.md#approval-page). |
| `account_recovery` | The end-user opens an account recovery link. | `code` | Passes `code` as [`account_recovery_code`](./authentication-flow-api-reference.md#type-account_recovery-actiontype-verify_account_recovery_code) without a `state_token`. |

## Resolution

For each page, Authgear uses the first of these that is set:

1. The client's `x_custom_ui_endpoints` key.
2. The client's `x_custom_ui_uri`, for `authentication` only.
3. The project's `ui.custom_ui_endpoints` key.
4. The page of Auth UI.

The client is:

| Page | Client |
|---|---|
| `authentication` | The client of the authentication request. |
| `login_link`, `account_recovery` | The client of the Authentication Flow that sends the link, given by `client_id` in `url_query` when the flow is created. |

When there is no client, steps 1 and 2 are skipped. This includes a flow created without `client_id`, and the reset password message sent by the Admin API.

The project's keys apply to every client, including DCR and CIMD clients, which cannot set `x_custom_ui_endpoints`.

## Origins

- CORS: the origin of every URI in `ui.custom_ui_endpoints`, every client's `x_custom_ui_endpoints`, and every client's `x_custom_ui_uri` is allowed.
- Passkey ceremonies and `frame-ancestors`: the origin of every URI that can resolve to `authentication` is trusted, that is the project's `authentication`, and every client's `authentication` and `x_custom_ui_uri`.
