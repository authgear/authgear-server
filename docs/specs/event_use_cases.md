# Event Use Cases

This document documents the expected use cases of some events.

## Simple Use Cases

### Blocking / Allowing user signup or login according to geo location

You can use `user.auth.identified` to block or allow users from a certain location.

For example, if you want to allow only users in Hong Kong to use your app:

```typescript
export default async function (
  e: EventUserAuthIdentified
): Promise<EventUserAuthIdentifiedResponse> {
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

You can use `user.auth.identified` to block or allow user authenticate in a certain app according to their roles, standard or custom attributes.

For example, if you want to allow only users with role `sales` to access the app `crm-system` with client id is `c8da9b322e1f494e`:

```typescript
export default async function (
  e: EventUserAuthIdentified
): Promise<EventUserAuthIdentifiedResponse> {
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

You can use `user.auth.identified` to block user from logging into a specific app without using a correct flow.

For example, if you want to enforce 2FA in the app `hr-system` with client ID `c8da9b322e1f494e`, and you have an authentication flow named `2fa_required_login` which require 2FA during user login:

```typescript
export default async function (
  e: EventUserAuthIdentified
): Promise<EventUserAuthIdentifiedResponse> {
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

You can use `user.auth.identified` to block user from signing up in your system if they are not signing up with a specific email domain.

For example, you only want user with email domain `@authgear.com` to be able to signup:

```typescript
export default async function (
  e: EventUserAuthIdentified
): Promise<EventUserAuthIdentifiedResponse> {
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

You can use `user.auth.identified` to block user from logging in during weekends.

For example, if your business only operate during weekdays, therefore you do not want any user login during weekends:

```typescript
export default async function (
  e: EventUserAuthIdentified
): Promise<EventUserAuthIdentifiedResponse> {
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

You can use `user.auth.adaptive_control` to implement Adaptive MFA.

For example, you consider logins from outside `HK` is at a higher risk, therefore MFA should be required:

```typescript
export default async function (
  e: EventUserAuthAdaptiveControl
): Promise<EventUserAuthAdaptiveControlResponse> {
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

TODO(tung): Document `contraints`, and possible values of `amr`.

If `contraints.amr` with value `["mfa"]` is returned in the response, depending on where the authentication is triggered, the following behavior applies:

- Authentication Flow / Auth UI:
  - Signup / Promote: If the user does not have any secondary authenticator setup during the flow, a step will be added at the end of the flow to force user to setup an secondary authenticator. Available authenticators are the enabled authenticators of the project.
  - Login / Reauth: If the user does not use any secondary authenticator during the flow, a step will be added at the end of the flow to authenticate the user with any secondary authenticator available to the user. If no secondary authenticator is available, the flow fail because it runs into dead end, with reason `NoAuthenticator`.
  - Account Recovery: No effect, because account recovery does not support 2FA at the moment.
- Interaction: Fail immediately, because interaction does not support Adaptive MFA.
- Workflow / Latte: Fail immediately, because workflow does not support Adaptive MFA.

Note, you have no control on the MFA steps in the Authentication Flow when using `user.auth.adaptive_control`. If you need full control on when the MFA steps being inserted, see Advanced Use Cases [TODO(tung): Insert the link] below.

## Advanced Use Cases

### Applying stricter rate limits for account enumeration according to geo location

You can use `user.auth.identified` to apply a stricter rate limit for account enumeration based on geo location.

For example, you want to allow 10 attempts of account enumeration per minute in Hong Kong. And 5 attepts per minute in any other places outside Hong Kong.

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
  e: EventUserAuthIdentified
): Promise<EventUserAuthIdentifiedResponse> {
  if (e.context.geo_location_code === "HK") {
    return {
      is_allowed: true,
    };
  } else {
    return {
      is_allowed: true,
      overrides: {
        rate_limit: {
          weight: 2,
        },
      },
    };
  }
}
```

By setting `overrides.rate_limit.weight` to 2, it means this attempt of identification will contribute `2` counts to the rate limit. Therefore, 5 attempts are only allowed in 1 minute. (10 / 2 = 5)

`weight` can also be lower than 1. When set to `0`, this attempt will never hit rate limit.

TODO(tung): Document `overrides`. The only supported property is `rate_limit` at the moment.
