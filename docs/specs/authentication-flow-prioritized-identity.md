# Authentication Flow - Prioritized Identity

- [Introduction](#introduction)
- [Usecases](#usecases)
  - [Force users to use a safer OAuth Provider (such as Google Login)](#force-users-to-use-a-safer-oauth-provider-such-as-google-login)
  - [Internal ADFS always have priority over username login for security purpose](#internal-adfs-always-have-priority-over-username-login-for-security-purpose)
- [Behaviors](#behaviors)
  - [When there are multiple levels of priority](#when-there-are-multiple-levels-of-priority)
  - [When the user has only some of the prioritized identities](#when-the-user-has-only-some-of-the-prioritized-identities)

## Introduction

It is common that some identities are considered more secure, or more preferable comparing with other identities of the same user. Apps might want to have a way to specify the priority of different types of identities in a login flow, so when a user has another identity which is more preferable, it will be used for identification inside a login flow.

## Usecases

### Force users to use a safer OAuth Provider (such as Google Login)

Assume you have a portal app which supports two signup methods: Email with Password, and Google oauth login.

User of the portal can signup to the portal by one of the two methods, and login with that method afterwards.

If the user has signed up with email with password, the user can connect to a Google account at any time after he logged in to the portal (For example, through [Account Linking](./account-linking.md)). In your perspective, Google oauth login is the preferred login method because it is more secure. Therefore once user connected their google account, the portal should only accept google oauth login for the same account, and do not accept email with password logins.

Use the following config to acheive the goal:

```yaml
authentication_flow:
  login_flows:
    - name: default
      steps:
        - name: identify
          type: identify
          one_of:
            - identification: oauth
              priority: 1 # <- Configure priority here
            - identification: email
              steps:
                - type: authenticate
                  one_of:
                    - authentication: primary_password
```

In the above config, the `default` login flow has an `identify` step as the first step. And this `identify` step defined two possible identifications: `oauth` and `email`.

We added the `priority` field to the object of `identification: oauth`, with value of `1`. The `priority` field defines the level of preferability of this identification option, the higher the value, the higher the preferability. If not given, it is the default value of `0`.

Whenever an identification method was chosen in an `identify` step, the `priority` of all identification methods will be considered:

1. The `priority` of the chosen identification method will be compared to all other identification methods. If there are no identification methods with `priority` higher than the selected identification method, continue the login flow. Else, proceed to step 2.
2. For each of the identification method with `priority` higher than the selected identification method, check if the logging in user has at least one identity that could be used in this identification method. If no, continue the login flow. Else, proceed to step 3.
3. Returns an error, which tell the user to use another identity to pass this step:

   ```json
   {
     "name": "Invalid",
     "reason": "PrioritizedIdentityRequired",
     "message": "please use another identification method",
     "code": 400,
     "info": {
       "PreferredIdentitifications": [
         {
           "identification": "oauth",
           "provider_type": "google",
           "alias": "google"
         }
       ]
     }
   }
   ```

   In the above error, `info.PreferredIdentitifications` is an array containing all identitifications having `priority` higher than the current selected identity. The format of the objects is as same as the items inside output data of the `identify` step of login flow in `options`. For details, please read [the authentication flow api reference](./authentication-flow-api-reference.md##type-login-steptype-identify)

Now consider a practical example:

Assume Alice is a user of the above mentioned portal app. She signed up by Email `alice@example.com` with Primary Password.

And recently, she connected her Google account to the existing account using [account linking](./account-linking.md).

So now her account has two identities:

1. The email identity `alice@example.com`
2. The oauth identity from Google

And now, Alice try to login using the `default` login flow.

She first pass an input to the authentication flow api, which trys to use `alice@example.com` as the identification method:

```jsonc
{
  "identification": "email",
  "login_id": "alice@example.com"
}
```

Look at the authentication flow config again:

```yaml
authentication_flow:
  login_flows:
    - name: default
      steps:
        - name: identify
          type: identify
          one_of:
            - identification: oauth
              priority: 1
            - identification: email # <- This identification method was chosen
              steps:
                - type: authenticate
                  one_of:
                    - authentication: primary_password
```

`identification: email` was chosen according to the input. This option has a `priority` value of `0`, which is the default value because it is not given.

Looking at other existing options, we found there is another option `identification: oauth`. And the `priority` value is `1` for this option. `1` > `0`, so this option has a higher priority value than the chosen identification method.

Looking at the existing identities of Alice, she has a Google oauth identity, which can be used to pass the `identification: oauth`. As a result, an error will be returned by the authentication flow API:

```jsonc
{
  "name": "Invalid",
  "reason": "PrioritizedIdentityRequired",
  "message": "please use another identification method",
  "code": 400,
  "info": {
    "PreferredIdentitifications": [
      {
        "identification": "oauth",
        "provider_type": "google",
        "alias": "google"
      }
    ]
  }
}
```

When Alice saw this error, she knows she should use `oauth` identification method to login. So passing in a new input:

```jsonc
{
  "identification": "oauth",
  "alias": "google",
  "redirect_uri": "http://example.com/sso/oauth2/callback/google"
}
```

This time, `identification: oauth` was chosen according to the input. This option has a `priority` value of `1`.

Looking at other existing options, there is no other options with priority higher than `1`, therefore she is able to pass the step and continue the login flow.

### Internal ADFS always have priority over username login for security purpose

Assume you have a HR management system which does not support public signup.

All users of the system were pre-created with an username with a primary password.

The users must connect to a ADFS account after the first login, and use that adfs account to login afterwards.

This case is basically same as the above [Force users to use a safer OAuth Provider (such as Google Login)](#force-users-to-use-a-safer-oauth-provider-such-as-google-login) usecase, we can use the following config to acheive the expected behavior:

```yaml
authentication_flow:
  login_flows:
    - name: default
      steps:
        - name: identify
          type: identify
          one_of:
            - identification: oauth
              priority: 1
            - identification: username
              steps:
                - type: authenticate
                  one_of:
                    - authentication: primary_password
```

Similar to the previous case, when `username` is used to login into an account with `adfs` oauth account connected, an error will be thrown:

```jsonc
{
  "name": "Invalid",
  "reason": "PrioritizedIdentityRequired",
  "message": "please use another identification method",
  "code": 400,
  "info": {
    "PreferredIdentitifications": [
      {
        "identification": "oauth",
        "provider_type": "adfs",
        "alias": "adfs"
      }
    ]
  }
}
```

## Behaviors

This section specifies the behavior of `priority` under different cases.

### When there are multiple levels of priority

Assume we have the following config:

```yaml
authentication_flow:
  login_flows:
    - name: default
      steps:
        - name: identify
          type: identify
          one_of:
            - identification: oauth
              priority: 3
            - identification: phone
              priority: 2
            - identification: email
              priority: 1
            - identification: username
              priority: 0
```

In this example, we have three levels of `priority`. `0`, `1` and `2`.

When the following input was passed in:

```json
{
  "identification": "email",
  "login_id": "alice@example.com"
}
```

`identification: email` will be selected as the identification method, which has `priority` of `1`. `phone` and `oauth` which has priority higher than the current selected identification method `email` will be included in the `PreferredIdentitifications` array, while `username` will be excluded.

```jsonc
{
  "name": "Invalid",
  "reason": "PrioritizedIdentityRequired",
  "message": "please use another identification method",
  "code": 400,
  "info": {
    "PreferredIdentitifications": [
      {
        "identification": "oauth",
        "provider_type": "google",
        "alias": "google"
      },
      {
        "identification": "phone"
      }
    ]
  }
}
```

### When the user has only some of the prioritized identities

Assume we have the following config:

```yaml
authentication_flow:
  login_flows:
    - name: default
      steps:
        - name: identify
          type: identify
          one_of:
            - identification: oauth
              priority: 1
            - identification: phone
              priority: 1
            - identification: email
              priority: 1
            - identification: username
              priority: 0
```

And assume Alice only has `email` and `oauth` identities.

When the following input was passed in:

```json
{
  "identification": "email",
  "login_id": "alice@example.com"
}
```

The error is:

```jsonc
{
  "name": "Invalid",
  "reason": "PrioritizedIdentityRequired",
  "message": "please use another identification method",
  "code": 400,
  "info": {
    "PreferredIdentitifications": [
      {
        "identification": "oauth",
        "provider_type": "google",
        "alias": "google"
      },
      {
        "identification": "phone"
      }
    ]
  }
}
```

`phone` will still appear as one item in `PreferredIdentitifications`, even Alice actually does not have a `phone` identity.
