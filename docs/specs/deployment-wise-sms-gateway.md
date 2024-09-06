# Deployment wise SMS Gateway

This document describes configuration of sms gateway.

- [New environment variables](#new-environment-variables)
- [Addition to authgear.yaml](#addition-to-authgearyaml)
- [Behavior explained](#behavior-explained)

## New environment variables

The following environment variables are added

```
SMS_GATEWAY_TWILIO_ACCOUNT_SID=
SMS_GATEWAY_TWILIO_AUTH_TOKEN=
SMS_GATEWAY_TWILIO_MESSAGING_SERVICE_SID=

SMS_GATEWAY_NEXMO_API_KEY=
SMS_GATEWAY_NEXMO_API_SECRET=

SMS_GATEWAY_CUSTOM_URL=
SMS_GATEWAY_CUSTOM_TIMEOUT=

SMS_GATEWAY_DEFAULT_USE_CONFIG_FROM=environment_variable|authgear.secrets.yaml
SMS_GATEWAY_DEFAULT_PROVIDER=twilio|nexmo|custom
```

- `SMS_GATEWAY_{PROVIDER}_*`
  - The provider specific configs and credentials.
- `SMS_GATEWAY_DEFAULT_USE_CONFIG_FROM`
  - The default sms gateway config source. Can be `environment_variable` or `authgear.secrets.yaml`.
- `SMS_GATEWAY_DEFAULT_PROVIDER`
  - The default sms provide. Can be `twillio | nexmo | custom`.

## Addition to authgear.yaml

```diff
 messaging:
   sms_provider: twillio | nexmo | custom
+  sms_gateway:
+    use_config_from: environment_variable | authgear.secrets.yaml
+    provider: twillio | nexmo | custom
```

- `messaging.sms_gateway`
  - The new sms gateway config, capable to specify reading config from environment_variable or project config.
- `messaging.sms_gateway.use_config_from`
  - Required when `messaging.sms_gateway` presents.
  - Can be from `environment_variable` / `authgear.secrets.yaml`.
- `messaging.sms_gateway.provider`
  - Can be `twillio | nexmo | custom`.
  - When `messaging.sms_gateway.use_config_from` is `environment_variable`, it is optional.
  - When `messaging.sms_gateway.use_config_from` is `authgear.secrets.yaml`, it is required.

## Behavior explained

### Table 1

| `sms_provider` | `sms_gateway` | Behavior                                                                                                                                |
| -------------- | ------------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| absent         | absent        | See table 2.                                                                                                                            |
| absent         | **present**   | Use sms gateway config. See Table 3                                                                                                     |
| **present**    | absent        | Use `messaging.sms_provider` from `authgear.secrets.yaml`. Read config from `sms.{messaging.sms_provider}` from `authgear.secrets.yaml` |
| **present**    | **present**   | `messaging.sms_provider` will be ignored. Use sms gateway config, See Table 3                                                           |

### Table 2

When `sms_provider` and `sms_gateway` are absent, the provider config is termined by environment variables.

| `DEFAULT_USE_CONFIG_FROM` | `DEFAULT_PROVIDER` | Behavior                                                                                                                                   |
| ------------------------- | ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------ |
| absent                    | NC                 | `provider` will be determined from application logic. Read config from `sms.{provider}` from `authgear.secrets.yaml`                       |
| `environment_variable`    | absent             | `provider` will be determined from application logic. Read config from `SMS_GATEWAY_{provider}_*` from environment variables               |
| `environment_variable`    | **present**        | Use `SMS_GATEWAY_DEFAULT_PROVIDER` as provider. Will read config from `SMS_GATEWAY_{SMS_GATEWAY_DEFAULT_PROVIDER}_*` environment variables |
| `authgear.secrets.yaml`   | NC                 | `provider` will be determined from application logic. Read config from `sms.{provider}` from `authgear.secrets.yaml`                       |

### Table 3

When `sms_gateway` is provided,

Note: Since `sms_gateway.use_config_from` is required when `sms_gateway` is presented in authgear.yaml, it should always override `DEFAULT_USE_CONFIG_FROM` in environment variable. The following table would not consider the value of `DEFAULT_USE_CONFIG_FROM`.

| `sms_gateway.use_config_from` | `sms_gateway.provider` | `DEFAULT_PROVIDER` | Behaviour                                                                                                                                  |
| ----------------------------- | ---------------------- | ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------ |
| `environment_variable`        | absent                 | absent             | `provider` will be determined from application logic. Read config from `SMS_GATEWAY_{provider}_*` from environment variables               |
| `environment_variable`        | absent                 | **present**        | Use `SMS_GATEWAY_DEFAULT_PROVIDER` as provider. Will read config from `SMS_GATEWAY_{SMS_GATEWAY_DEFAULT_PROVIDER}_*` environment variables |
| `environment_variable`        | **present**            | NC                 | Use `sms_gateway.provider` as provider. Will read config from `SMS_GATEWAY_{sms_gateway.provider}_*` environment variables                 |
| `authgear.secrets.yaml`       | absent                 | NC                 | Error as `sms_gateway.provider` is required in this circumstance.                                                                          |
| `authgear.secrets.yaml`       | **present**            | NC                 | Use provider configs from `authgear.secrets.yaml`. Will read config from `sms.{sms_gateway.provider}` from `authgear.secrets.yaml`         |
