# Authentication Flow Selection

- [Introduction](#introduction)
- [Concepts](#concepts)
  - [Flow Group](#flow-group)
  - [Flow Allowlist](#flow-allowlist)
- [Flow Selection in Different UIs](#flow-selection-in-different-uis)
  - [Default UI](#default-ui)
  - [Custom UI](#custom-ui)
- [Configuration](#configuration)
  - [Flow Group Allowlist](#flow-group-allowlist)
  - [Flow Allowlist](#flow-allowlist-1)

## Introduction

A project can have multiple variants for each authentication flow for different clients. For example, a project can have a flow for OAuth login only in public app, and another flow for email/password login with 2FA in internal app.

This spec documents the behaviour of Authentication Flow Selection in different UIs using Flow Group and Flow Allowlist.

## Concepts

### Flow Group

By default configuration, flows with name `default` are used. Flow groups can be defined in the configuration to group multiple flow types together and thus allow initiating customized flows.

For example, a project can have multiple login flow and signup flows. After defining the flow groups and allowlist, the client app can specify which flow group to use using `x_authentication_flow_group` in the [authentication request](/docs/specs/oidc.md#x_authentication_flow_group).

### Flow Allowlist

In cases where flow group is not used, the allowlist of flows can be defined in the configuration per client. It restricts what flows can be initiated by the client app.

If flow group is also defined, the final allow list is computed with union of the flow group allowlist and flow allowlist.

## Flow Selection in Different UIs

There are two types of UIs:
- **Default UI**: The default UI provided by Authgear.
- **Custom UI**: A custom UI that uses Authgear's Authflow API to initiate a flow.

They have different behaviours when selecting authentication flows.

### Default UI

Default UI uses flow group for deriving allowed flows.

Following table explains how authentication flow is selected in the default UI:

| Flow Group Allowlist | Authorization Request | Behaviour |
| --- | --- | --- |
| Defined | `x_authentication_flow_group` is present  | Selected flow group is checked against the allowlist and used if valid. |
| Defined | `x_authentication_flow_group` is **NOT** present  | First flow group in the allowlist is used. |
| Not defined | `x_authentication_flow_group` is present  | Selected flow group is used. |
| Not defined | `x_authentication_flow_group` is **NOT** present  | Default flow group is used. |

### Custom UI
Unlike Default UI, Custom UI can create flows using Authflow API. The flow group allowlist cannot be used to restrict the flows created by Custom UI.

Instead, it computes the effective flow allowlist using union of `x_authentication_flow_group_allowlist` and `x_authentication_flow_allowlist`. `x_authentication_flow_group` of [authentication request](/docs/specs/oidc.md#x_authentication_flow_group) is Custom UI.

Following table explains how authentication flow is selected in the custom UI:

| Flow Group Allowlist | Flow Allowlist | Behaviour |
| --- | --- | --- |
| Defined | Defined | All flow names in flow groups allowlist and flow allowlist are checked against. |
| Defined | Not defined | All flow names allowed flow groups are checked against. |
| Not defined | Defined | Flow allowlist is used. |
| Not defined | Not defined | All flows are allowed. |

## Configuration

### Flow Group Allowlist

Each client app may have an allowlist of flow groups defined with `x_authentication_flow_group_allowlist`. Once specified, the authorization request will be rejected if specified flow group is not in the allowlist.

For example, given following configuration:

```yaml
ui:
  authentication_flow:
    groups:
    - name: oauth_only
      login_flow: oauth_only
      signup_flow: oauth_only
      signup_login_flow: oauth_only
      reauth_flow: password
      promote_flow: password
      account_recovery_flow: email
    - name: email_password_2fa
      login_flow: email_password_2fa
      signup_flow: email_password_2fa
      signup_login_flow: email_password_2fa
      reauth_flow: sms_code
      promote_flow: sms_code
      account_recovery_flow: sms

oauth:
  clients:
    - client_id: public_app
      # Only allow OAuth login for public_app
      x_authentication_flow_group_allowlist:
      - oauth_only
    - client_id: internal
      # Allow both OAuth login and email_password_2fa for internal
      x_authentication_flow_group_allowlist:
      - oauth_only
      - email_password_2fa
```

To authorize using a flow group, include `x_authentication_flow_group` and `x_page` in the [authentication request](/docs/specs/oidc.md#x_authentication_flow_group)

In the example, client app `public_app` is only allowed to use `oauth_only` flow group:

```jsonc
{
  // ... other fields of the authorization request
  "x_page": "login",
  "x_authentication_flow_group": "oauth_only"
}
```

### Flow Allowlist

Each client app may have an allowlist of flows defined with `x_authentication_flow_allowlist`. The effective allowlist is computed by taking union of `x_authentication_flow_group_allowlist` and `x_authentication_flow_allowlist`. Once specified, the Authflow API will only allow the flows specified in the allowlist.

For example, given following configuration:

```yaml
oauth:
  clients:
    - client_id: custom_app
      x_custom_ui_url: "https://example.com"
      x_authentication_flow_allowlist:
        login_flows:
        - oauth_only
        signup_flows:
        - oauth_only
        reauth_flows:
        - password
        promote_flows:
        - password
        account_recovery_flows:
        - email
```

When Authentication Flow API receives create flow request from `custom_app`, it will only allow the flows specified in the allowlist.

For example to create a login flow for `custom_app`:

```jsonc
// POST /api/v1/authentication_flows

{
    "type": "login",
    "name": "oauth_only",
    "url_query": "client_id=custom_app"
}
```
