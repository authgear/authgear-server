# Event Use Cases

This document documents the expected use cases of some events.

## Table of Contents

### Simple Use Cases

- [Blocking / Allowing user signup or login according to geo location](#blocking--allowing-user-signup-or-login-according-to-geo-location)
- [Blocking / Allowing user login according to user roles for a certain application](#blocking--allowing-user-login-according-to-user-roles-for-a-certain-application)
- [Ensure a user login to a specific App with a specific flow](#ensure-a-user-login-to-a-specific-app-with-a-specific-flow)
- [Email domain whitelist](#email-domain-whitelist)
- [Block login during weekends](#block-login-during-weekends)
- [Require MFA only for users with high risk (Adaptive MFA)](#require-mfa-only-for-users-with-high-risk-adaptive-mfa)
- [Block user login according to AMR](#block-user-login-according-to-amr)

### Advanced Use Cases

- [Applying stricter rate limits in an authentication flow](#applying-stricter-rate-limits-in-an-authentication-flow)
- [Enable bot protection under specific conditions](#enable-bot-protection-under-specific-conditions)

## Simple Use Cases

### Blocking / Allowing user signup or login according to geo location

You can use `authentication.post_identified` to block or allow users from a certain location.

For example, if you want to allow only users in Hong Kong to use your app:

```typescript
export default async function (
  e: EventAuthenticationPostIdentified
): Promise<EventAuthenticationPostIdentifiedResponse> {
  if (e.context.geo_location_code === "HK") {
    return {
      is_allowed: true,
    };
  } else {
    return {
      is_allowed: false,
    };
  }
}
```

### Blocking / Allowing user login according to user roles for a certain application

You can use `authentication.post_identified` to block or allow user authenticate in a certain app according to their roles, standard or custom attributes.

For example, if you want to allow only users with role `sales` to access the app `crm-system` with client id is `c8da9b322e1f494e`:

```typescript
export default async function (
  e: EventAuthenticationPostIdentified
): Promise<EventAuthenticationPostIdentifiedResponse> {
  if (
    e.context.client_id === "c8da9b322e1f494e" &&
    ["login", "signup"].contains(e.context.authentication_flow?.type)
  ) {
    if (e.user.roles.contains("sales")) {
      return {
        is_allowed: true,
      };
    } else if (e.user.custom_attributes.can_access_crm === "true") {
      // Alternatively, use custom_attributes to determine if the user is allowed to access the app
      return {
        is_allowed: true,
      };
    } else {
      return {
        is_allowed: false,
      };
    }
  }
  // Allow login or signups of other clients
  return {
    is_allowed: true,
  };
}
```

### Ensure a user login to a specific App with a specific flow

You can use `authentication.post_identified` to block user from logging into a specific app without using a correct flow.

For example, if you want to enforce 2FA in the app `hr-system` with client ID `c8da9b322e1f494e`, and you have an authentication flow named `2fa_required_login` which require 2FA during user login:

```typescript
export default async function (
  e: EventAuthenticationPostIdentified
): Promise<EventAuthenticationPostIdentifiedResponse> {
  if (
    e.context.client_id === "c8da9b322e1f494e" &&
    e.context.authentication_flow?.type === "login"
  ) {
    if (e.context.authentication_flow.name !== "2fa_required_login") {
      return {
        is_allowed: false,
      };
    } else {
      return {
        is_allowed: true,
      };
    }
  }
  // Allow login of other clients
  return {
    is_allowed: true,
  };
}
```

### Email domain whitelist

You can use `authentication.post_identified` to block user from signing up in your system if they are not signing up with a specific email domain.

For example, you only want user with email domain `@authgear.com` to be able to signup:

```typescript
export default async function (
  e: EventAuthenticationPostIdentified
): Promise<EventAuthenticationPostIdentifiedResponse> {
  if (e.identity.claims?.email?.endsWith("@authgear.com")) {
    return {
      is_allowed: true,
    };
  }
  // Block signup of all other emails
  return {
    is_allowed: false,
  };
}
```

### Block login during weekends

You can use `authentication.post_identified` to block user from logging in during weekends.

For example, if your business only operate during weekdays, therefore you do not want any user login during weekends:

```typescript
export default async function (
  e: EventAuthenticationPostIdentified
): Promise<EventAuthenticationPostIdentifiedResponse> {
  const today = new Date();
  // 0 is sunday, and 6 is saturday
  if (today.getDay() === 0 || today.getDay() === 6) {
    return {
      is_allowed: false,
    };
  }
  return {
    is_allowed: true,
  };
}
```

Note: Even login is blocked during weekends, refresh tokens, access tokens and IDP sessions issued during weekdays will not be invalidated.

### Require MFA only for users with high risk (Adaptive MFA)

You can use `authentication.pre_authenticated` to implement Adaptive MFA.

For example, you consider logins from outside `HK` is at a higher risk, therefore MFA should be required:

```typescript
export default async function (
  e: EventAuthenticationPreAuthenticated
): Promise<EventAuthenticationPreAuthenticatedResponse> {
  if (e.context.geo_location_code !== "HK") {
    return {
      // Allow the login with a mfa contraint
      is_allowed: true,
      contraints: {
        amr: ["mfa"],
      },
    };
  }
  // Else, simply allow the login
  return {
    is_allowed: true,
  };
}
```

If `contraints.amr` with value `["mfa"]` is returned in the response, depending on where the authentication is triggered, the following behavior applies:

- Authentication Flow / Auth UI:
  - Signup / Promote: If the user does not have any secondary authenticator setup during the flow, a step will be added at the end of the flow to force user to setup an secondary authenticator. Available authenticators are the enabled authenticators of the project.
  - Login / Reauth: If the user does not use any secondary authenticator during the flow, a step will be added at the end of the flow to authenticate the user with any secondary authenticator available to the user. If no secondary authenticator is available, the flow fail because it runs into dead end, with reason `NoAuthenticator`.
  - Account Recovery: No effect, because account recovery does not support 2FA at the moment.
- Interaction: Fail immediately, because interaction does not support Adaptive MFA.
- Workflow / Latte: Fail immediately, because workflow does not support Adaptive MFA.

Note, you have no control on the MFA steps in the Authentication Flow when using `authentication.pre_authenticated`. If you need full control on when the MFA steps being inserted, see [Adaptive MFA with customized Authflow](#adaptive-mfa-with-customized-authflow) below.

"amr" stands for (Authentication Methods References)[https://www.rfc-editor.org/rfc/rfc8176.html], it is a claim used with OpenID Connect for storing details about how the authentication was performed. Only "mfa" is supported at the moment. Any other values will have no effect.

### Block user login according to AMR

You can use `authentication.pre_authenticated` to block user login according to the AMR (Authentication Methods References) used during authentication.

For example, if you want to block users who did not use `mfa` (multi-factor authentication) during the authentication:

```typescript
export default async function (
  e: EventAuthenticationPreAuthenticated
): Promise<EventAuthenticationPreAuthenticatedResponse> {
  if (!e.amr.includes("mfa")) {
    return {
      is_allowed: false,
    };
  }
  return {
    is_allowed: true,
  };
}
```

## Advanced Use Cases

### Applying stricter rate limits in an authentication flow

You can apply a stricter rate limit in an authentication flow using hooks.

For example, you want to allow 5 attempts of account enumeration per minute in Hong Kong. And 10 attepts per minute in any other places outside Hong Kong.

Firstly, you will have the following rate limit config:

```yaml
authentication:
  rate_limits:
    account_enumeration:
      per_ip:
        enabled: true
        period: 1m
        burst: 10
```

This sets the rate limit of account enumeration to 10/minute.

Then, you can write the following hook:

```typescript
export default async function (
  e: EventAuthenticationPreInitialize
): Promise<EventAuthenticationPreInitializeResponse> {
  if (e.context.geo_location_code === "HK") {
    return {
      is_allowed: true,
      rate_limits: {
        "authentication.account_enumeration": {
          weight: 2
        }
      },
    };
  } else {
    return {
      is_allowed: true
    };
  }
}
```

By setting `"rate_limits.authentication.account_enumeration.weight"` to 2, it means any attempt of account enumeration (Such as identify steps) will contribute `2` attempts to the rate limit. Therefore, 5 attempts are only allowed in 1 minute. (10 / 2 = 5)

`authentication.account_enumeration` is the corresponding rate limit name. See the [rate limit spec](./rate-limit.md) for details.

`weight` can also be lower than 1. When set to `0`, the rate limit will never be hit.


Note that only rate limits checked after the hook is triggered are affected. For example, setting the `weight` of `authentication.account_enumeration` in an `authentication.post_identified` hook will likely be ineffective. This is because the `authentication.account_enumeration` rate limit is checked during the identify step, which runs before the `authentication.post_identified` hook. However, adjusting the weight for `authentication.password` in the same hook would still be effective, as the user has probably not authenticated with a password yet.

If you want to make sure all rate limits are affected, it is suggested to use the `authentication.pre_initialize` event.

### Enable bot protection under specific conditions

Use `authentication.pre_initialize` to enable bot protection under specific conditions.

For example, if you want to display captcha only for user outside Hong Kong:

Firstly, enable bot protection with mode `never`.

```yaml
bot_protection:
  enabled: true
  requirements:
    signup_or_login:
      mode: "never"
```

When `mode` is `never`, bot protection will not be enabled by default.

Then, return `bot_protection.mode` in your `authentication.pre_initialize` hook:

```typescript
export default async function (
  e: EventAuthenticationPreInitialize
): Promise<EventAuthenticationPreInitializeResponse> {
  if (e.context.geo_location_code !== "HK") {
    return {
      is_allowed: true,
      bot_protection: {
        mode: "always",
      },
    };
  }
  // Else, simply allow the login
  return {
    is_allowed: true,
  };
}
```

This overrides the original `mode` of bot_protection in your config. Therefore, bot_protection will be turned on if the user is trying to signup or login outside Hong Kong.
