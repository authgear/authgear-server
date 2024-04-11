# Identity Provider Account Linking

## Introduction

A single user could have accounts in multiple identity providers, such as Google, Facebook, Github. We want to provide a way to identify and link accounts from different identity providers into a single authgear account.

This spec documents a feature that allows users to link a oauth account to an existing authgear account during the signup flow.

## Configuration

Here is an example of the account linking configuration:

```yaml
identity:
  oauth:
    providers:
      - alias: google
        client_id: exampleclientid
        type: google
        link_by:
          identity: "email"
      - alias: azureadv2
        client_id: exampleclientid
        type: azureadv2
        link_by:
          identity: "username"
          idp_claim: "/preferred_username"
```

- The `link_by` object was added to provider config. This section defines how an account in an external idp can be linked (i.e. matched) to an existing authgear user. It contains the following fields. Nullable. When null, no linking is available (i.e. Conflicts will never occur).
  - `link_by.identity`: The identity used to match an account in authgear. Possible values are `email`, `phone`, `username`. The default is `email`.
  - `link_by.idp_claim`: A key of the user info from the external idp. The key will be used to extract a value from the user info json, and match with any authgear user by looking at `link_by.identity`. If it is not specified, authgear will automatically choose a suitable claim to match with the selected `link_by.identity`. If authgear is unable to choose a suitable claim, it is an error.
    - For example, `/email` will be choosed for a google account if `link_by.identity` is set to `email`. However, if `link_by.identity` is `phone`, the user must specify `link_by.idp_claim` because authgear have no idea which claim from the google user info can be used to match with a phone identity.
    - If neccessary, the user can use a hook to perform transformation on the idp user info. This is to handle use cases like multiple claims in the idp user info could be used. For example, an idp might separate the country code and the national number into two claims. In this case, the developer must use a hook to transform the related claims into a single claim, and match with that claim instead. This will not be implemented at current stage.

For related steps in authflow, please see the below [Action on conflict](#action-on-conflict) section.

```yaml
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
            on_conflict: # This section was added
              action: "login_and_link"
              login_flow: "default"
              target_step: "step_1"

login_flows:
  - name: default
    steps:
      - name: step_1
        type: identify
        one_of:
          - identification: oauth
          - identification: email
            steps:
              - name: authenticate_primary_email
                type: authenticate
                one_of:
                  - authentication: primary_password
      - type: check_account_status
      - type: terminate_other_sessions
```

- The `on_conflict` section was added to the identify step of the signup flow to specify the behavior if an conflict occurred during signup. Currently, it can only be configured with `identification: oauth`. The possible values of `on_conflict.action` are:
  - `"error"`: Reject the signup with an error. This is the default.
  - `"login"`: Switch to login flow of the existing account. `login_flow` must be specified when using this option.
    - This will not be implemented as it seems a duplicate of `"error"`.
  - `"login_and_link"`: Switch to login flow of the existing account. After that, add the new identity which triggered the conflict to the logged in account. When `login_and_link` is choosed, the following fields must be specified:
    - `on_conflict.login_flow`: The login flow to switch to when an conflict occurs.
    - `on_conflict.target_step`: The step inside the login flow which we will start from. The selected step must satisfy the below conditions:
      - It must be a `identify` step.
      - It must be the first `identify` step.
      - It must be a step at the root level. i.e., It cannot be a nested step inside any `one_of` branches.
    - Read the [Login and link flow](#login-and-link-flow) for the detailed behavior.
  - `"create_new_account"`: Create a new account with this new identity, ignoring the conflict.
    - This will not be implemented at the moment.
  - `"create_new_account_or_link"`: Allow the user to choose between behavior of `login_and_link` and `create_new_account`.
    - This will not be implemented at the moment.
  - `"hook"`: Use a hook to decide the behavior.
    - This will not be implemented at the moment.

## Action on conflict

- Whenever an oauth identity is used, and there is no existing user for that identity.

  1. If `link_by` is null, do nothing. Else,
  2. Get the value of claim from the idp claim using `link_by`. If the value is empty, do nothing.
  3. Query from existing identities using the claim key specified by `link_by`, where the claim value is matching the value in 2.
  4. Do:

  - If there is at least one match,
    - If `on_conflict=error`, return an error and terminate the signup flow.
    - If `on_conflict=login`, trigger login flow for that existing user, and discard the new identity.
    - If `on_conflict=login_and_link`, trigger login flow for that existing user, and link the new oauth identity to that user after the flow was completed sucessfully. Read the [Login and link flow](#login-and-link-flow) for the detailed behavior.
    - If `on_conflict=create_new_account`, create a new user with that oauth identity.
    - If `on_conflict=hook`, trigger a hook and use the result to determine what to do.
  - If there is no match, create a new user.

- Whenever a new user creates an new login id during signup.

  1. For every `identity.providers` with non-null `link_by`, use `link_by` to obtain a value from the new identity. If the value is empty, do nothing.
  2. Use the value we get in 1 to query for existing oauth identities with the same claim specified in `link_by`.
  3. If there is at least one identify found in step 2, return an error.

## Login and Link Flow

During a signup, when a conflict is occurred, and `on_conflict` is set to `login_and_link`, the user will enter a login and link flow. Please see the following example to understand the actual flow:

Assume we have the following authflow config:

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
            on_conflict:
              action: login_and_link
              login_flow: default
              target_step: step_1
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
      - name: step_1
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

and we have the following provider config:

```yaml
identity:
  oauth:
    providers:
      - alias: google
        client_id: exampleclientid
        type: google
        link_by:
          identity: "email"
          idp_claim: "/email"
```

Assume now there is an existing authgear user with the following identities and authenticators:

- User A
  - Email Identity: a@example.com
  - Primary Password Authenticator

And now, the user tries to sign up with a new google account, which has an email `a@example.com` in the google user info. And authgear matched that oauth identity to the existing login ID `a@example.com`.

1. The user should first select a conflicted identity, in this example, there is only one conflicted identity `Email: a@example.com`. So this identity should be selected by the user.
2. After `Email: a@example.com` being selected, the user will switch to a login flow. Which was specified by `on_conflict.login_flow` And `on_conflict.target_step`:

```yaml
on_conflict:
  action: login_and_link
  login_flow: default
  target_step: step_1
```

3. The login flow `default` will be executed, and starting from the specified step `step_1`.
4. The selected identity `Email: a@example.com` will be automatically used to pass the first identity step.

```yaml
- identification: email
  steps:
    - name: authenticate_primary_email
      type: authenticate
      one_of:
        - authentication: primary_password
```

5. Then, the login flow will continue. The user has to enter primary password to pass the authentication.
6. After user entered password, the login flow was completed. Now, the original signup flow will be continued.

```yaml
- identification: email
  steps:
    - name: email_setup_primary_email
      type: create_authenticator
      one_of:
        - authentication: primary_password
- identification: oauth # <-- Resume here
  on_conflict:
    action: login_and_link
    login_flow: default
    target_step: step_1
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

7. As the next step of the original flow is create primary password authenticator, the user will need to created primary authenticator.
8. As the original user already has a primary password authenticator, the step will be skipped.
9. And the next step will be create `secondary_totp`. As the user don't have a secondary totp, the user should create a totp in this step.
10. Finally, all created identities and authenticators will be added to the existing user, together with the new oauth identity.

Resulting user:

- User A
  - Email Identity: a@example.com
  - OAuth Identity: Google (email:a@example.com)
  - Primary Password Authenticator
  - Secondary TOTP Authenticator

## Q&A

- Why we need to login the user before linking the account?

  - This is to prevent the user account being taken over using a oauth idenity. For example, if an idp allows registering an user account without verifying the email, that idp can be used to create accounts to take over authgear accounts if we link the account without running the login flow. Therefore passing the login flow before linking the account is neccessary.

- Why we need to continue the original signup flow instead of simply adding the oauth identity to the user?

  - Consider the following:
    - Assume we have the following signup flow:
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
                on_conflict:
                  action: login_and_link
                  login_flow: default
                  target_step: step_1
                steps:
                  - type: create_authenticator
                    one_of:
                      - authentication: secondary_totp
    ```
  - If we add the oauth identity to the user without completing the whole signup flow, the step that create `secondary_totp` would be skipped. Which may break the assumption that all users created by signup flow with oauth identity will have `secondary_totp` setup. Therefore we should continue the signup flow. However, we should skip unncessary steps to prevent duplicated authenticators of the same type being added.

- Why `on_conflict` is only available on `identification: oauth` but not other login ids such as `identification: email`?
  - We think that the common use case is to link an oauth account to an existing login id, but not the reverse. So it is not supportted at the moment. However, theoretically it is possible to support `on_conflict` of other `identification` methods too. This could be added in the future.

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
