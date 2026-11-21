- [Login Link](#login-link)
  - [Configuration](#configuration)
  - [Lifecycle](#lifecycle)
  - [Approval page](#approval-page)
  - [API](#api)
    - [POST /api/v1/login_link/inspect](#post-apiv1login_linkinspect)
    - [POST /api/v1/login_link/approve](#post-apiv1login_linkapprove)
  - [Custom UI](#custom-ui)

# Login Link

A login link is an email [OOB-OTP](./user-model.md#oob-otp-authenticator) delivered as a URL instead of a code. The end-user opens the link, on any device, and approves it. The [Authentication Flow](./authentication-flow.md) that sent the link then continues on the device where it was started.

Login link is used when authenticating with, or setting up, an email OOB-OTP authenticator. Setting one up with a link also verifies the email address. The `verify` step of signup and promote flows always sends a code, never a link.

## Configuration

```yaml
authenticator:
  oob_otp:
    email:
      email_otp_mode: login_link # default "code"
      valid_periods:
        link: 20m # default "20m"
```

The message is rendered by the [`login_link` templates](./templates.md#login_link).

## Lifecycle

1. The flow sends the link to the email address. The step returns `otp_form: link`. See [Login link in the Authentication Flow API](./authentication-flow-api-reference.md#login-link) for the step's input and output.
2. The end-user opens the link. It leads to an [approval page](#approval-page).
3. The end-user approves the link on the approval page. Every listener on the flow's websocket receives `{"kind":"refresh"}`.
4. The device driving the flow completes the step. The flow consumes the code and continues.

A code is invalid once it has expired or has been consumed in step 4. Approving does not consume it.

## Approval page

The approval page receives the code in the link.

1. On load, it calls [inspect](#post-apiv1login_linkinspect). If the code cannot be approved, it tells the end-user that the link is invalid, used or expired, and offers no way to approve.
2. It calls [approve](#post-apiv1login_linkapprove) only when the end-user explicitly confirms, such as by tapping a button. It MUST NOT approve on page load: email security scanners open links automatically, and would otherwise approve every link they scan.

Auth UI provides an approval page. A [Custom UI](#custom-ui) can provide its own.

## API

The API takes the code from the link, not a flow state token, because the approval page may be opened on a device that has never seen the flow. It follows the [API response conventions](./api.md), and the same CORS policy and rate limit as the Authentication Flow API.

Both endpoints take the same request:

```json
{
  "code": "M6CGA4WV6M9XTXNWFYFHRQDWF6VFR7K4"
}
```

On success, both respond:

```json
{
  "result": {}
}
```

Both return these errors:

| `reason` | `code` | When |
|---|---|---|
| `InvalidOTPCode` | 403 | The code does not exist, has expired, has been consumed, or has been replaced by a resent link. |
| `RateLimited` | 429 | Too many verification attempts from the IP address, or too many failed attempts on the link. |
| `TooManyRequest` | 429 | Too many Authentication Flow API requests from the IP address. |

### POST /api/v1/login_link/inspect

Succeeds when the code can be approved. It does not approve the code, and does not notify the flow.

### POST /api/v1/login_link/approve

Approves a code and notifies the flow that sent it. Approving an already approved code succeeds again.

## Custom UI

To approve links in a Custom UI, host an [approval page](#approval-page) that calls the [API](#api), and set it as the [`login_link` page](./custom-ui.md#pages) of the Custom UI. The link then points to that page.
