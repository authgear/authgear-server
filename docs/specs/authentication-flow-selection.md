- [Authentication Flow Selection](#authentication-flow-selection)
  * [Introduction](#introduction)
  * [Concepts](#concepts)
    + [Flow Group](#flow-group)
    + [Flow Allowlist](#flow-allowlist)
  * [Flow Selection in Different UIs](#flow-selection-in-different-uis)
    + [Auth UI](#auth-ui)
    + [Custom UI](#custom-ui)
  * [Configuration](#configuration)
  * [Usage](#usage)
    + [Auth UI](#auth-ui-1)
    + [Custom UI / Authflow API](#custom-ui--authflow-api)

# Authentication Flow Selection

## Introduction

A project can have multiple variants for each authentication flow for different clients. For example, a project can have a flow for OAuth login only in public app, and another flow for email/password login with 2FA in internal app.

This spec documents the behaviour of Authentication Flow Selection in different UIs using Flow Group and Flow Allowlist.

## Concepts

### Flow Group

Flow groups can be defined in the configuration to group multiple flow types together and thus allow initiating customized flows.

For example, a project can have multiple login flow and signup flows. After defining the flow groups and allowlist, the client app can specify which flow group to use using `x_authentication_flow_group` in the [authentication request](/docs/specs/oidc.md#x_authentication_flow_group).

### Flow Allowlist

Flow group only instructs Auth UI to use a specific group of flows. To allow them to be used in custom UI, the flow group must be defined in the allowlist configuration.

## Flow Selection in Different UIs

There are two types of UIs:
- **Auth UI**: The UI provided by Authgear.
- **Custom UI**: A custom UI that uses Authgear's Authflow API to initiate a flow.

They have different behaviours when selecting authentication flows.

### Auth UI

Auth UI uses flow group for deriving allowed flows.

Following table explains how authentication flow is selected in Auth UI:

| Flow allowlist | Authorization Request | Behaviour |
| --- | --- | --- |
| Defined | `x_authentication_flow_group` is present | Selected flow group is checked against the allowlist and used if valid. |
| Defined | `x_authentication_flow_group` is **NOT** present | First flow group in the allowlist is used. |
| Not defined | `x_authentication_flow_group` is present | Selected flow group is used. |
| Not defined | `x_authentication_flow_group` is **NOT** present | Default flow group is used. |

### Custom UI
Unlike Auth UI, Custom UI creates flows using Authflow API. `x_authentication_flow_allowlist` is not used to restrict the flows created by Custom UI.

Instead, it computes the effective flow allowlist using union of group and individual flows in the allowlist to decide which flows can be created.

## Configuration

Each client app may have an allowlist of flow groups defined with `x_authentication_flow_allowlist`. Once specified, the authorization request will be rejected if specified flow group is not in the allowlist.

For example, given following flow groups defined:

```yaml
ui:
  authentication_flow:
    groups:
    - name: oauth_only
      flows:
      - type: login
        name: oauth_only
      - type: signup
        name: oauth_only
      - type: signup
        name: email_password_2fa
      - type: signup_login
        name: oauth_only
      - type: reauth
        name: password
      - type: promote
        name: password
      - type: account_recovery
        name: email
      - type: login
        name: email_password_2fa
      - type: signup
        name: email_password_2fa
      - type: signup_login
        name: email_password_2fa
      - type: reauth
        name: sms_code
      - type: promote
        name: sms_code
      - type: account_recovery
        name: sms
```

You can specify allowed flow groups and individual flows for each client app:

```yaml
oauth:
  clients:
  - client_id: public_app
    x_authentication_flow_allowlist:
      # Only allow OAuth login for public_app
      groups:
      - name: oauth_only
  - client_id: internal
    x_authentication_flow_allowlist:
      # Allow both OAuth login and email_password_2fa for internal
      groups:
      - name: oauth_only
      - name: email_password_2fa
      # Allow additional individual flows
      flows:
      - type: reauth
        name: email_password_2fa
  - client_id: custom_app
    x_custom_ui_uri: "https://custom.app"
    x_authentication_flow_allowlist:
      flows:
      - type: login
        name: oauth_only
```

## Usage

### Auth UI

To authorize using a flow group, include `x_authentication_flow_group` and `x_page` in the [authentication request](/docs/specs/oidc.md#x_authentication_flow_group)

In the example, client app `public_app` is only allowed to use `oauth_only` flow group:

```jsonc
{
  // ... other fields of the authorization request
  "x_page": "login",
  "x_authentication_flow_group": "oauth_only"
}
```

### Custom UI / Authflow API

When Authentication Flow API receives create request from `custom_app`, it will only allow the flows specified in the allowlist.

For example to create a login flow for `custom_app`:

```jsonc
// POST /api/v1/authentication_flows

{
    "type": "login",
    "name": "oauth_only",
    "url_query": "client_id=custom_app"
}
```
