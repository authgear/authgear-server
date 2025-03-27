# Account Linking

TODO(ldap): Specify how an LDAP identity take part in Account Linking.

- [Introduction](#introduction)
- [Configuration](#configuration)
  - [Defining how the linking occurs and the corresponding action](#defining-how-the-linking-occurs-and-the-corresponding-action)
    - [Defining Linkings](#defining-linkings)
    - [Linking Actions](#linking-actions)
  - [Customize the config in specific step](#customize-the-config-in-specific-step)
  - [Default Behaviors](#default-behaviors)
    - [The Current Defaults](#the-current-defaults)
    - [The default linking of different provider types](#the-default-linking-of-different-provider-types)
- [Identity Attributes](#identity-attributes)
  - [The built-in oauth identity standard attributes](#the-built-in-oauth-identity-standard-attributes)
  - [Customizing the oauth identity attributes](#customizing-the-oauth-identity-attributes)
- [Login and Link Flow](#login-and-link-flow)
- [Account Linking by Login IDs](#account-linking-by-login-ids)
  - [Defaults of Account Linking of Login IDs](#defaults-of-account-linking-of-Login-ids)
- [Account Linking in Promote Flow](#account-linking-in-promote-flow)
- [Q&A](#qa)
  - [Why we need to login the user before linking the account?](#why-we-need-to-login-the-user-before-linking-the-account)
  - [Why we need to continue the original signup flow instead of simply adding the oauth identity to the user?](#why-we-need-to-continue-the-original-signup-flow-instead-of-simply-adding-the-oauth-identity-to-the-user)
- [Usecases](#usecases)
- [References](#references)

## Introduction

A single user could have accounts in multiple identity providers, such as Google, Facebook, Github. We want to provide a way to identify and link accounts from different identity providers into a single authgear account.

This spec documents a feature that allows users to link a oauth account to an existing authgear account during the signup flow.

## Configuration

### Defining how the linking occurs and the corresponding action

To use account linking, you must define two things:

1. How the linking occurs. That is, by what condition, authgear should try to link your oauth account to an existing account.
2. What should be done when the linking occurs.

Let's explain with the following example:

```yaml
account_linking:
  oauth:
    - alias: adfs
      oauth_claim:
        pointer: "/preferred_username"
      user_profile:
        pointer: "/preferred_username"
      action: login_and_link
authentication_flow:
  signup_flows:
    - name: default
      steps:
        - name: identify
          type: identify
          one_of:
            - identification: email
              steps:
                - name: authenticate_primary_email
                  type: create_authenticator
                  one_of:
                    - authentication: primary_password
            - identification: oauth
```

The `account_linking` section defined the account linking behavior of any authentication flow. It have the following meanings:

- When user is trying to sign up with an `oauth` identity, which belongs to the oauth provider with alias `adfs`, account linking may occurs. The related provider is specified by the `alias` field, which should match one of the providers specified in the `identity.oauth.providers` config.
- Linking should occur if the new oauth identity is having a `"email"` claim, which the value is equal to the `"email"` attribute of any existing user's identity. For details, please read the [Defining Linkings](#defining-linkings) section.
- When the linking occurs, it should trigger a login using the `"default"` login flow. After the login, the new oauth identity will be linked to the account logged in. For details, please read the [Linking Actions](#linking-actions) section.

The configs under `account_linking.oauth` only controls account linking behavior when the new identity is an `oauth` identity. As a result, only the `identification: oauth` step may trigger account linking.

```yaml
name: default
steps:
  - name: identify
    type: identify
    one_of:
      - identification: email
        steps:
          - name: authenticate_primary_email
            type: create_authenticator
            one_of:
              - authentication: primary_password
      - identification: oauth # <-- Only this step may trigger account linking
```

Currently, only `oauth` is supported in account linking. However, account linking of `login_id` may also be supported in the future.

#### Defining Linkings

We define linkings between the new oauth identity and any existing identities using the `oauth_claim` and `user_profile` fields.

- `oauth_claim`: An object containing a json pointer, specified in `oauth_claim.pointer`, pointing to a claim of the incoming oauth user profile. Note that, for oidc compatible providers, this pointer is used to access value from the oidc claims, which is from the user info endpoint. For non-oidc compatible providers, please read the [SSO Providers](/docs/specs/sso-providers.md) document for the corresponing logics authgear implemented to obtain a user profile from the provider.

- `user_profile`: An object containing a json pointer, specified in `user_profile.pointer`, pointing to a value of the user profile of an existing authgear user, or the attribute of an existing authgear identity. For the meaning of attribute of authgear identity, please read the [Identity Attribute](#identity-attribute) section.

Whenever the value pointed by `oauth_claim.pointer` of the new oauth identity matches the value pointed by `user_profile.pointer` of any existing authgear identity or authgear user profile, account linking will be triggered.

For what should happen on linking, please read the following [Linking Actions](#linking-actions) section.

#### Linking Actions

We define the action to link the new oauth identity with the existing identity's owner user account using the `action` field.

- `action`: Defines the desire action if this linking was triggered.
  The possible values are:
  - `error`: Reject the signup with an error.
  - `login`: Switch to login flow of the existing account.
    - This will not be implemented as it seems a duplicate of `"error"`.
  - `login_and_link`: Switch to login flow of the existing account. After user completing the login flow, add the new oauth identity to the logged in account. When `login_and_link` is choosed, read the [Login and link flow](#login-and-link-flow) section for the detailed behavior of this option.
  - `always_link_without_login`: This is similar to `login_and_link`, but no login is required. Caution: This could become a risk that someone will be able to takeover some authgear accounts using identities from the oauth provider. Only use this option if you trust the oauth provider and knows the linking logics works as you expected.
  - `link_without_login_when_verified`: This is similar to `login_and_link`, but no login is required only if the oauth provider claims that the email in the user's claim is verified, by using the `email_verified` claim.
  - `create_new_account`: Create a new account with this new identity, ignoring the link.
  - `create_new_account_or_link`: Allow the user to choose between behavior of `login_and_link` and `create_new_account`.
  - `hook`: Use a hook to decide the behavior.

Currently, only `error` and `login_and_link` will be implemented.

### Customize the config in specific step

We provide a way to customize account linking configs in a specific step. This is because there are valid use cases that each step should behave differently. For example, you want the `default` signup flow to do account linking with the `default` signup flow, but in other side, you want another signup flow `signup_flow_1` to do account linking with another `login_flow_1` login_flow. Use the following config to achieve this behavior:

```yaml
account_linking:
  oauth:
    - alias: adfs
      oauth_claim:
        pointer: "/preferred_username"
      user_profile:
        pointer: "/preferred_username"
      action: login_and_link
    - name: adfs_link_by_email # name must be defined if overriding this config in specific flow is needed
      alias: adfs
      oauth_claim:
        pointer: "/email"
      user_profile:
        pointer: "/email"
      action: login_and_link
authentication_flow:
  signup_flows:
    - name: signup_flow_1
      steps:
        - name: identify
          type: identify
          one_of:
            - identification: email
              steps:
                - name: authenticate_primary_email
                  type: create_authenticator
                  one_of:
                    - authentication: primary_password
            - name: identify_by_oauth
              identification: oauth
              account_linking:
                oauth:
                  - name: adfs_link_by_email # This object overrides account linking configs with the same name
                    action: link_without_login_when_verified
                    login_flow: login_flow_1
```

In the above example, an `account_linking` object was added inside an `identify` step of the signup flow `signup_flow_1`, and it overrides some configs in the `authentication_flow.account_linking`.

You must give a `name` to the object you want to customize. In the above example, `name: adfs_link_by_email` was added to an object inside `account_linking.oauth`. `name` can be any string, it is used to reference this object in the later configs.

Then, we add one item in `account_linking.oauth` inside step `identify_by_oauth` of the signup flow `signup_flow_1`, which has `name: adfs_link_by_email`. This means the config inside the new item will override any config in `account_linking.oauth` with `name` equal to `adfs_link_by_email`. In this example, `action` has been changed to `link_without_login_when_verified` if account linking was triggered at the step.

Not all configs are overridable inside a flow. We only support overriding the following configs:

- `action`

Read the above [Defining how the linking occurs and the corresponding action](#defining-how-the-linking-occurs-and-the-corresponding-action) section for meaning of each fields.

`oauth_claim` and `user_profile` is not supported in the config overrides because we think there is no valid usecase for it. You propably always want a consistent linking logics between flows.

Additionally, `login_flow` is available as a configurable parameter of account linking. If a login flow is triggered during account linking, and `login_flow` is provided, the configured `login_flow` will be used as the name of the login flow which used to perform account linking. If not provided, the current flow name will be used. For example, if you triggered account linking in the `default` signup flow, then `default` login flow will be used if `login_flow` is not specified.

### Default Behaviors

The account linking config will be read according to the below precedence:

1. If exist, always use the configurations in `account_linking` inside the current flow object in `signup_flows`.
2. Else, use the `authentication_flow.account_linking` confuguration, if exist.
3. Else, it is the built in default behavior. Please read the following sections for detail.

#### The Current Defaults

The current defaults are identical to the following config:

```yaml
oauth_claim:
  pointer: "/email"
user_profile:
  pointer: "/email"
action: error
```

These defaults will be applied to all oauth providers if account linking is not configured for that provider.

Please refer to [Defining Linkings](#defining-linkings) and [Linking Actions](#linking-actions) sections for the meanings of the configs.

#### The default linking of different provider types

Different default account linking configs could be provided for different type of providers. However, this is not implemented at the moment.

## Identity Attributes

In authgear, each identity contributes to some standard attributes of the authgear user profile. Please see the following table for the standard attributes contributed by the email, username and phone identities.

| Identity Type | Value            | Standard Attributes                   |
| ------------- | ---------------- | ------------------------------------- |
| Email         | `id@example.com` | `{ "email": "id@example.com" }`       |
| Username      | `example`        | `{ "preferred_username": "example" }` |
| Phone         | `+8520001`       | `{ "phone_number": "+8520001" }`      |

OAuth identities also contributes attributes of the user profile, but it is more complicated than the above three identity types. The following sections will discuss the related behavior and configs.

### The built-in oauth identity standard attributes

For each supported oauth provider types, authgear has implemented a built-in standard attribute mappings. You could find the mappings of each provider in the [SSO Providers](/docs/specs/sso-providers.md) document.

### Customizing the oauth identity attributes

As the built-in mappings may not be able to handle all use cases, we support configuring custom mappings on oauth identity attributes. Let's start with the following example config:

```yaml
identity:
  oauth:
    providers:
      - alias: adfs
        client_id: exampleclientid
        type: adfs
        user_profile_mapping:
          - oauth_claim:
              pointer: "/primary_phone"
            user_profile:
              pointer: "/phone_number"
```

The above config means:

For any oauth identity of `adfs`, we will read a value from the `"primary_phone"` claim of the provider user profile, and write that value into the `"phone_number"` attribute of that identity. Note that, it is not writing directly to the user's attribute, but the attributes that belongs to this identity. The user can later select this identity in the portal to populate these attributes into the authgear user profile.

And the meaning of each configs are:

- `user_profile_mapping`: It is an array, which specifies a mapping. From one claim of the provider user profile, to one attribute of the authgear user identity attribute.
  - `user_profile_mapping.oauth_claim`: An object, which only has one field `pointer`. The `pointer` is the JSON pointer pointing to the claim value of the oauth provider user profile.
  - `user_profile_mapping.user_profile`: An object, which only has one field `pointer`. The `pointer` is the JSON pointer pointing to the attribute of the authgear identity attribute.
    - We only support standard attributes at the moment, but custom attributes may also be supported in the future.

## Login and Link Flow

During a signup, when a linking is occurred, and `action` is set to `login_and_link`, the user will enter a login and link flow. Please see the following example to understand the actual flow:

Assume we have the following authflow config:

```yaml
account_linking:
  oauth:
    - alias: google
      oauth_claim:
        pointer: "/email"
      user_profile:
        pointer: "/email"
      action: login_and_link
signup_flows:
  - name: default
    steps:
      - name: identify
        type: identify
        one_of:
          - identification: email
            steps:
              - name: email_setup_primary_email
                type: create_authenticator
                one_of:
                  - authentication: primary_password
          - identification: oauth
            steps:
              - name: oauth_setup_primary_email
                type: create_authenticator
                one_of:
                  - authentication: primary_password
                    steps:
                      - type: create_authenticator
                        one_of:
                          - authentication: secondary_totp
                            steps:
                              - type: view_recovery_code

login_flows:
  - name: default
    steps:
      - name: identify
        type: identify
        one_of:
          - identification: oauth
          - identification: email
            steps:
              - name: email_authenticate_password
                type: authenticate
                one_of:
                  - authentication: primary_password
```

Assume now there is an existing authgear user with the following identities and authenticators:

- User A
  - Email Identity: a@example.com
  - Primary Password Authenticator

And now, the user tries to sign up with a new google account, which has an email `a@example.com` in the google user profile. And authgear matched that oauth identity to the existing login ID `a@example.com`.

1. The user should first select a matched identity, in this example, there is only one matched identity `Email: a@example.com`. So this identity will be selected by the user.
2. After `Email: a@example.com` being selected, the user will switch to the `default` login flow. The name of the login flow used, unless specified, will be the same of the current signup flow. For details about how to use a different login flow, please read [Customize the config in specific step](#customize-the-config-in-specific-step).
3. The login flow will be executed inside the signup flow, step by step.
4. The selected identity `Email: a@example.com` will be automatically used to pass the first identification step:

   ```yaml
   - identification: email
     steps:
       - name: authenticate_primary_email
         type: authenticate
         one_of:
           - authentication: primary_password
   ```

5. Then, the login flow will continue. The user has to enter primary password to pass the authentication.
6. After user entered password, the login flow was completed. Now, the original signup flow will be continued:

   ```yaml
   - identification: email
     steps:
       - name: email_setup_primary_email
         type: create_authenticator
         one_of:
           - authentication: primary_password
   - identification: oauth # <-- Resume here
     steps:
       - name: oauth_setup_primary_email
         type: create_authenticator
         one_of:
           - authentication: primary_password
             steps:
               - type: create_authenticator
                 one_of:
                   - authentication: secondary_totp
                     steps:
                       - type: view_recovery_code
   ```

7. As the next step of the original flow is to create primary password authenticator, the user will need to create primary password authenticator if he doesn't have one. As the original user already has a primary password authenticator, the step will be skipped. The priciple is, the resumed signup flow will try to skip all identification or authentication steps that the user already has one related identity or authenticator.
8. And the next step will be creating `secondary_totp`. As the user does not have a secondary totp, the user will create a totp in this step.
9. Finally, all created identities and authenticators will be added to the existing user, together with the new oauth identity.

Resulting user:

- User A
  - Email Identity: a@example.com
  - OAuth Identity: Google (email:a@example.com)
  - Primary Password Authenticator
  - Secondary TOTP Authenticator

## Q&A

### Why we need to login the user before linking the account?

This is to prevent the user account being taken over using a oauth idenity. For example, if an idp allows registering an user account without verifying the email, that idp can be used to create accounts to take over authgear accounts if we link the account without running the login flow. Therefore passing the login flow before linking the account is neccessary.

### Why we need to continue the original signup flow instead of simply adding the oauth identity to the user?

Consider the following signup flow config:

```yaml
signup_flows:
  - name: default
    steps:
      - name: identify
        type: identify
        one_of:
          - identification: email
            steps:
              - name: email_setup_primary_email
                type: create_authenticator
                one_of:
                  - authentication: primary_password
          - identification: oauth
            steps:
              - type: create_authenticator
                one_of:
                  - authentication: secondary_totp
```

If we add the oauth identity to the user without completing the whole signup flow, the step that create `secondary_totp` would be skipped. Which may break the assumption that all users created by signup flow with oauth identity will have `secondary_totp` setup. Therefore we should continue the signup flow. However, we should skip unncessary steps to prevent duplicated authenticators of the same type being added.

## Account Linking by Login IDs

The config of account linking by login ids is defined by an object inside `account_linking.login_id`:

```yaml
account_linking:
  login_id:
    - key: phone
      user_profile:
        pointer: "/phone_number"
      action: "login_and_link"
    - key: username
      user_profile:
        pointer: "/preferred_username"
      action: "error"
```

The above configs defined the account linking behavior of two login id types:

- For `phone`, link with any existing user or identity by looking at the `"phone_number"` value in the user profile or identity attribute. When linking occurs, use `login_and_link` as the action. Read the [Linking Actions](#linking-actions) section for the exact meaning of `login_and_link`.
- For `username`, link with any existing user or identity by looking at the `"preferred_username"` value in the user profile or identity attribute. When linking occurs, returns an error and stop the signup flow.

### Defaults of Account Linking of Login IDs

The default values of each login id types are different. Please see the below table.

| Login ID Type | Defaults                                                                                               |
| ------------- | ------------------------------------------------------------------------------------------------------ |
| `email`       | <pre><code>user_profile:<br/>&nbsp;&nbsp;pointer: "/email"<br/>action: error</code></pre>              |
| `phone`       | <pre><code>user_profile:<br/>&nbsp;&nbsp;pointer: "/phone_number"<br/>action: error</code></pre>       |
| `username`    | <pre><code>user_profile:<br/>&nbsp;&nbsp;pointer: "/preferred_username"<br/>action: error</code></pre> |

Therefore when not specified, all attempts to link a new login id to existing users or idenitities will result in error.

## Account Linking in Promote Flow

Account linking can occur in promote flow, however, only `error` is allowed to be the action of account linking.

This is because currently prmote flow does not support logging in to existing user in promote flow. So actions such as `login_and_link` which actually logged in to an existing user is also not possible at the moment. However, this restriction could be relaxed after we supported logging in to an existing user in promote flow.

Any action set in `authentication_flow.account_linking` will be treated as error. And `action` in `signup_flows.account_linking` only allows `error` as value.

## Usecases

### Common Cases

1. User already owns an account in authgear, but logged in with an oauth account that is not known to authgear.
   Expect one of the following results:

- Created a new account with no relationship with the existing account.
- Tell the user he has an existing account, and he should login to that account instead.
- Allow the user to continue login with the oauth account and finally he result in logged in the existing account, and the oauth account can also be used to login to this account in the future.

  ```yaml
  account_linking:
    oauth:
      - alias: google
        oauth_claim:
          pointer: "/email"
        user_profile:
          pointer: "/email"
        action: login_and_link
  ```

2. User already owns an account in authgear, but tried to signup with a new email / phone number / username.
   Expect one of the following results:

- Created a new account with no relationship with the existing account.
- Tell the user he has an existing account, and he should login to that account instead.
- Redirect the user to login to the existing account. And in future, that email / phone number / username can be used to login to same account.

  ```yaml
  account_linking:
    login_id:
      - key: phone
        user_profile:
          pointer: "/phone_number"
        action: "login_and_link"
      - key: email
        user_profile:
          pointer: "/email"
        action: "login_and_link"
      - key: username
        user_profile:
          pointer: "/username"
        action: "login_and_link"
  ```

### Edge Cases

1. Say a project has two clients, A and B. And he has two different signup flows and two login flows for the two clients. Each client should always use the flow belongs to that client. And the project want to enable account linking.

   - Signup flow of A should use login flow of A during account linking. And signup flow of B should use login flow of B during account linking.
   - Say if only signup flow of B allows signing up with email, while both flows support signing up with oauth. Then signup flow of B should not be able to trigger account linking by email while signing up with oauth.

   The solution is to override the login flow name per flow.

   ```yaml
   account_linking:
     oauth:
       - name: adfs_linking
         alias: adfs
         oauth_claim:
           pointer: "/preferred_username"
         user_profile:
           pointer: "/preferred_username"
         action: login_and_link
   authentication_flow:
     signup_flows:
       - name: flow_1
         steps:
           - name: identify
             type: identify
             one_of:
               - identification: email
               - identification: oauth
                 account_linking:
                   oauth:
                     - name: adfs_linking
                       login_flow: flow_1
   ```

2. Say a project has pre-imported users. And we know that these users may have a adfs account. We probably already know the user id of the adfs account already. So what we want is:

   - The pre-imported users will know about their adfs account username. So whenever a new adfs account is trying to login, we can map it to a pre-imported user. And trigger login for that user instead.

   The solution is to link by a custom attribute, which was set during the user was imported.

   ```yaml
   account_linking:
     oauth:
       - name: adfs_linking
         alias: adfs
         oauth_claim:
           pointer: "/preferred_username"
         user_profile:
           pointer: "/x_adfs_username"
         action: login_and_link
   authentication_flow:
     signup_flows:
       - name: default
         steps:
           - name: identify
             type: identify
             one_of:
               - identification: email
               - identification: oauth
   ```

3. Consider a signup flow which accepts using google account OR phone number, but phone number wil still be collected if user signup with google. And the followings are wanted:

   - The user can login with the google account.
   - The user cannot login with the phone number.
   - No one can signup a new account with that phone number.

   The solution is to link by a custom attribute, which was set after the user signed up by google. The phone number collected after google signup flow should be stored into a custom attribute `/custom_profile_phone_number`.

   ```yaml
   account_linking:
     login_id:
       - key: phone
         user_profile:
           pointer: "/custom_profile_phone_number"
         action: "error"
   authentication_flow:
     signup_flows:
       - name: default
         steps:
           - name: identify
             type: identify
             one_of:
               - identification: phone
               - identification: oauth
   ```

4. Assume google oauth signup and email signup are both available for an app. Now, you want users of the app to be able to login to his own existing account even if he originally signed up by google, but entered the same email during signup. The verification step is not necessary in the signup flow. So you want:

   - If the user has verified his email during signup, re-login of existing account is not needed, because you believe the owner of a same email should also own that google account.
   - Else, he has to re-login to his existing account to prove his ownership.

   It can be done by the following config:

   ```yaml
   account_linking:
     login_id:
       - name: email_linking
         key: email
         user_profile:
           pointer: "/email"
         action: login_and_link
   authentication_flow:
     signup_flows:
       - name: default
         steps:
           - name: identify
             type: identify
             one_of:
               - identification: email # This branch require email verification
                 account_linking:
                   login_id:
                     - name: email_linking
                       action: link_without_login_when_verified # Override action in this branch so user do not need to login again for account linking
                 steps:
                   - target_step: identify
                     type: verify
               - identification: username # Another branch that doesn't require email verification
                 steps:
                   - type: identity
                     one_of:
                       - identificaiton: email # No override, therefore login is still required for account linking
   ```

## References

We designed the feature based on the following references:

- Auth0: https://auth0.com/docs/manage-users/user-accounts/user-account-linking/link-user-accounts

  Auth0 doesn't have auto merge, they merge (link) accounts by user link apis.

  A primary user & secondary user must be specified, and the user profile of secondary user will be merged into primary user's profile, but it will not override existing values of primary profile.

- Okta: https://developer.okta.com/docs/concepts/identity-providers/#account-linking

  When Account Link Policy is "Automatic", there are two options:

  Match Against: The field / property to match in the okta account, e.g. Okta Username

  IdP Username: The field used to match with "Match Against", in the user profile obtained from the idp.

  Manual linking is also possible: https://developer.okta.com/docs/reference/api/idps/#link-a-user-to-a-social-provider-without-a-transaction

- AWS Cognito: https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_AdminLinkProviderForUser.html

  Similar to Okta's automatic account link, you need to specify "ProviderAttributeName" and "ProviderAttributeValue", so that AWS Cognito knows how to match and link the external identity to an existing account.
