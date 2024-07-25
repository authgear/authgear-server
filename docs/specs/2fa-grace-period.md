# 2FA Grace Period

- [Abstract](#abstract)
- [Configuration](#configuration)
  - [Global Grace Period](#global-grace-period)
  - [Per-user Grace Period](#per-user-grace-period)
  - [Customized Authentication Flow](#customized-authentication-flow)

## Abstract

In order to enforce 2FA, we will provide a grace period for users to enroll in 2FA after they have signed up.

After the grace period, the user will be required to enroll in 2FA before they can sign in.

## Configuration

By default, no grace period is provided. User must enroll in 2FA immediately after [2FA mode](./user-model.md#secondary-authenticator) is set to `required`.

For new users, the grace period starts from the time the user is created, whereas for existing users, the grace period starts from when the grace period is enabled.

### Global Grace Period

After 2FA mode is set to `required`, a global grace period can be configured as follow:

```yaml
authentication:
  secondary_authentication_mode: "required"
  secondary_authentication_rollout:
    global_grace_period_enabled_since: "2021-01-01T00:00:00Z"
    global_grace_period_days: 30
```

### Per-user Grace Period

Regardless of the global grace period, a per-user grace period can be granted for 10 days by default, or a custom duration can be specified:

```yaml
authentication:
  secondary_authentication_mode: "required"
  secondary_authentication_rollout:
    per_user_grace_period_days: 10
```

It is possible to extend the grace period of a user multiple times or cancel immediately through Admin Portal / GraphQL API.

### Customized Authentication Flow

By default all signup / login flows respect the global 2FA mode configuration. However, you can also override it in the flow configuration:

```yaml
authentication_flow:
  login_flows:
    - name: internal_user
      secondary_authentication_mode: "required"
      steps:
      # ....
    - name: public_user
      secondary_authentication_mode: "if_exists"
      steps:
      # ...

  signup_flows:
    - name: default
      secondary_authentication_mode: "disabled"
      steps:
      # ...

  signup_login_flows:
    - name: default
      secondary_authentication_mode: "required"
      steps:
      # ...

  reauth_flows:
    - name: default
      secondary_authentication_mode: "required"
      steps:
      # ...
```

When 2FA mode has been set to `required` and first 2FA step is met, user will be required to either use existing 2FA or create a new one.
