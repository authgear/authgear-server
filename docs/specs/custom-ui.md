- [Custom UI](#custom-ui)
  - [Configuration](#configuration)
  - [Pages](#pages)
  - [Resolution](#resolution)
  - [Origins](#origins)

# Custom UI

A Custom UI is a UI hosted by the developer that replaces pages of Auth UI, by using the [Authentication Flow API](./authentication-flow-api-reference.md).

## Configuration

A project sets the pages of its Custom UI. A static client can set its own `authentication` page with `x_custom_ui_uri`.

```yaml
ui:
  custom_ui_endpoints:
    authentication: "https://custom.app/auth"
    login_link: "https://custom.app/login-link"
    account_recovery: "https://custom.app/reset"

oauth:
  clients:
  - client_id: client_b
    x_custom_ui_uri: "https://ui-b.example.com/auth"
```

Every key is optional, and its value is an absolute URI. A key that is not set falls back as described in [Resolution](#resolution).

## Pages

| Key | Authgear sends the end-user there when | Query parameters added | The page |
|---|---|---|---|
| `authentication` | An [authentication request](./oidc.md) is received. | Parameters of the authentication request needed to create a flow. | Creates an Authentication Flow. |
| `login_link` | The end-user opens a [login link](./login-link.md). | `code` | Is the [approval page](./login-link.md#approval-page). |
| `account_recovery` | The end-user opens an account recovery link. | `code` | Passes `code` as [`account_recovery_code`](./authentication-flow-api-reference.md#type-account_recovery-actiontype-verify_account_recovery_code) without a `state_token`. |

## Resolution

| Page | Authgear uses the first of these that is set |
|---|---|
| `authentication` | The `x_custom_ui_uri` of the client of the authentication request, then the project's key, then Auth UI. |
| `login_link`, `account_recovery` | The project's key, then Auth UI. |

The project's keys apply to every client, including DCR and CIMD clients. Links go to the project's pages even when the client sets `x_custom_ui_uri`.

## Origins

- CORS: the origin of every URI in `ui.custom_ui_endpoints`, and of every client's `x_custom_ui_uri`, is allowed.
- Passkey ceremonies and `frame-ancestors`: the origin of every URI that can resolve to `authentication` is trusted, that is the project's `authentication` and every client's `x_custom_ui_uri`.
