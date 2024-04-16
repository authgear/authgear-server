# Account Linking

- [Introduction](#introduction)
- [Configuration](#configuration)
  - [Defining how the linking occurs and the corresponding action](#defining-how-the-linking-occurs-and-the-corresponding-action)
    - [Defining Linkings](#defining-linkings)
    - [Linking Actions](#linking-actions)
  - [Defining a default for all flows](#defining-a-default-for-all-flows)
    - [Default Behaviors](#default-behaviors)
      - [The Current Defaults](#the-current-defaults)
      - [The default linking of different provider types](#the-default-linking-of-different-provider-types)
- [Identity Attributes](#identity-attributes)
  - [The built-in oauth identity standard attributes](#the-built-in-oauth-identity-standard-attributes)
  - [Customizing the oauth identity attributes](#customizing-the-oauth-identity-attributes)
- [Login and Link Flow](#login-and-link-flow)
- [Q&A](#qa)
  - [Why we need to login the user before linking the account?](#why-we-need-to-login-the-user-before-linking-the-account)
  - [Why we need to continue the original signup flow instead of simply adding the oauth identity to the user?](#why-we-need-to-continue-the-original-signup-flow-instead-of-simply-adding-the-oauth-identity-to-the-user)
  - [Why only `oauth` identities supports `account_linking`?](#why-only-oauth-identities-supports-account_linking)
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
authentication_flow:
  signup_flows:
    - name: default
      account_linking:
        oauth:
          - alias: adfs
            incoming_claim:
              pointer: "/preferred_username"
            existing_attribute:
              pointer: "/preferred_username"
            action: login_and_link
            login_flow: default
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

The `account_linking` section inside `signup_flows` defined the account linking behavior of the `default` signup flow. It have the following meanings:

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

We define linkings between the new oauth identity and any existing identities using the `incoming_claim` and `existing_attribute` fields.

- `incoming_claim`: An object containing a json pointer, specified in `incoming_claim.pointer`, pointing to a claim of the incoming oauth user profile. Note that, for oidc compatible providers, this pointer is used to access value from the oidc claims, which is from the user info endpoint. For non-oidc compatible providers, please read the [SSO Providers](#todo) document for the corresponing logics authgear implemented to obtain a user profile from the provider.

- `existing_attribute`: An object containing a json pointer, specified in `existing_attribute.pointer`, pointing to an attribute of an existing authgear identity. For the meaning of attribute of authgear identity, please read the [Identity Attribute](#identity-attribute) section.

Whenever the value pointed by `incoming_claim.pointer` of the new oauth identity matches the value pointed by `existing_attribute.pointer` of any existing authgear identity, account linking will be triggered by this linking.

For what should happen on linking, please read the following [Linking Actions](#linking-actions) section.

#### Linking Actions

We define the action to link the new oauth identity with the existing identity's owner user account using the `action` field.

- `action`: Defines the desire action if this linking was triggered.
  The possible values are:
  - `error`: Reject the signup with an error.
  - `login`: Switch to login flow of the existing account.
    - This will not be implemented as it seems a duplicate of `"error"`.
  - `login_and_link`: Switch to login flow of the existing account. After user completing the login flow, add the new oauth identity to the logged in account. When `login_and_link` is choosed, the following additional configs are available:
    - `login_flow`: The login flow name to switch to when an linking occurs. The selected login flow must start with a `identity` step. This field can be omitted. If omitted, the default value will be the same flow name
    - Read the [Login and link flow](#login-and-link-flow) section for the detailed behavior of this option.
  - `always_link_without_login`: This is similar to `login_and_link`, but no login is required. Caution: This could become a risk that someone will be able to takeover some authgear accounts using identities from the oauth provider. Only use this option if you trust the oauth provider and knows the linking logics works as you expected.
  - `link_without_login_when_verified`: This is similar to `login_and_link`, but no login is required only if the oauth provider claims that the email in the user's claim is verified, by using the `email_verified` claim.
  - `create_new_account`: Create a new account with this new identity, ignoring the link.
  - `create_new_account_or_link`: Allow the user to choose between behavior of `login_and_link` and `create_new_account`.
  - `hook`: Use a hook to decide the behavior.

Currently, only `error` and `login_and_link` will be implemented.

### Defining a default for all flows

All config mentioned above was defined inside a single flow object of a `signup_flows`. However, you may simply want a config applied to all signup flows. Therefore, we have the `default_account_linking` section for this purpose. Here is an example:

```yaml
authentication_flow:
  default_account_linking:
    oauth:
      - alias: google
        incoming_claim:
          pointer: "/email"
        existing_attribute:
          pointer: "/email"
        action: login_and_link
        login_flow: default
```

It supports all configs as mentioned in the above [Defining how the linking occurs and the corresponding action](#defining-how-the-linking-occurs-and-the-corresponding-action).

If `authentication_flow.default_account_linking` is specified, it will be applied to all signup flows. If any signup flows does not have the `account_linking` config specified, it will be treated as having same configs inside `authentication_flow.default_account_linking`.

For detail about default behaviors, please read the [Default Behaviors](#default-behaviors) section.

### Default Behaviors

The account linking config will be read according to the below precedence:

1. If exist, always use the configurations in `account_linking` inside the current flow object in `signup_flows`.
2. Else, use the `authentication_flow.default_account_linking` confuguration, if exist.
3. Else, it is the built in default behavior. Please read the following sections for detail.

#### The Current Defaults

The current defaults are identical to the following config:

```yaml
incoming_claim:
  pointer: "/email"
existing_attribute:
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

For each supporteed oauth provider types, authgear has implemented a built-in standard attribute mappings. You could find the mappings of each provider in the [SSO Providers](#todo) document.

### Customizing the oauth identity attributes

As the built-in mappings may not be able to handle all use cases, we support configuring custom mappings on oauth identity attributes. Let's start with the following example config:

```yaml
identity:
  oauth:
    providers:
      - alias: adfs
        client_id: exampleclientid
        type: adfs
        attribute_mappings:
          - from_claim:
              pointer: "/primary_phone"
            to_attribute:
              pointer: "/phone_number"
```

The above config means:

For any oauth identity of `adfs`, we will read a value from the `"primary_phone"` claim of the provider user profile, and write that value into the `"phone_number"` attribute of that identity. Note that, it is not writing directly to the user's attribute, but the attributes that belongs to this identity. The user can later select this identity in the portal to populate these attributes into the authgear user profile.

And the meaning of each configs are:

- `attribute_mappings`: It is an array, which specifies a mapping. From one claim of the provider user profile, to one attribute of the authgear user identity attribute.
  - `attribute_mappings.from_claim`: An object, which only has one field `pointer`. The `pointer` is the JSON pointer pointing to the claim value of the oauth provider user profile.
  - `attribute_mappings.to_attribute`: An object, which only has one field `pointer`. The `pointer` is the JSON pointer pointing to the attribute of the authgear identity attribute.
    - We only support standard attributes at the moment, but custom attributes may also be supported in the future.

## Login and Link Flow

During a signup, when a linking is occurred, and `action` is set to `login_and_link`, the user will enter a login and link flow. Please see the following example to understand the actual flow:

Assume we have the following authflow config:

```yaml
signup_flows:
  - name: default
    account_linking:
      oauth:
        - alias: google
          incoming_claim:
            pointer: "/email"
          existing_attribute:
            pointer: "/email"
          action: login_and_link
          login_flow: default
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
2. After `Email: a@example.com` being selected, the user will switch to the `default` login flow. Which was specified by `login_flow` in the config.
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

### Why only `oauth` identities supports `account_linking`?

We think that the common use case is to link an oauth account to an existing login id, but not the reverse. So triggering account linking by other types of identities are at the moment. However, theoretically it is possible to support account linking of other `identification` methods too. This could be added in the future.

The proposed config of account linking by login ids could be:

```yaml
account_linking:
  login_id:
    - key: phone
      existing_pointer: "/phone_number"
      action: "error"
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
