# Identity Provider Account Linking

## Abstract

A single user could have accounts in multiple identity providers, such as Google, Facebook, Github. We want to provide a way to identify and link accounts from different identity providers into a single authgear account.

An account linking config section is added to configuration of each oauth provider config for this purpose. The developer must specify the account matching logic for the specific identity provider if they wish to turn on account linking feature for that identity provider.

## Configuration

Here is an example of the configuration:

```yaml
identity:
  on_conflict:
    signup: "login_and_link"
  oauth:
    providers:
      - alias: google
        client_id: exampleclientid
        type: google
        link_by:
          pointer: "/email"
      - alias: azureadv2
        client_id: exampleclientid
        type: azureadv2
        link_by:
          pointer: "/preferred_username"
      
```

- The `link_by` object was added to provider config.
- `link_by.pointer`: The pointer to the claim which used to match the oauth account to an authgear user. For example `"/email"`. It could be a standard claim, or a custom claim. Sensible options will be displayed in portal, but no valiation will be done at server because we would like to keep the flexibility to use any claim for any provider. The default value is `pointer: "/email"`, except that for wechat, it is default `null`.

- The `identity.on_conflict.signup` config was added to specify the behavior if an conflict occurred during signup. The possible values are:
  - `"error"`: Reject the signup with an error. This is the default.
  - `"login"`: Switch to login flow of the existing account.
  - `"login_and_link"`: Switch to login flow of the existing account. After that, add the new identity which triggered the conflict to the logged in account.
  - `"create_new_account"`: Create a new account with this new identity, ignoring the conflict.
  - `"hook"`: Use a hook to decide the behavior.

At this stage, we will only implement `error`, `login` and `login_and_link`.

## Trigger conditions and actions

- Whenever an oauth identity is used, and there is no existing user for that identity.

  1. If `link_by` is null, do nothing. Else,
  2. Get the value of claim from the idp claim using `link_by`. If the value is empty, do nothing.
  3. Query from existing identities using the claim key specified by `link_by`, where the claim value is matching the value in 2.
  4. Do:

  - If there is at least one match,
    - If `identity.on_conflict.signup=error`, return an error and terminate the signup flow.
    - If `identity.on_conflict.signup=login`, trigger login flow for that existing user, and discard the new identity.
    - If `identity.on_conflict.signup=login_and_link`, trigger login flow for that existing user, and add the new oauth identity to that user after the flow was completed sucessfully.
    - If `identity.on_conflict.signup=create_new_account`, create a new user with that oauth identity.
    - If `identity.on_conflict.signup=hook`, trigger a hook and use the result to determine what to do.
  - If there is no match, create a new user.

- Whenever a new user creates an new identity during signup.
  1. For every `identity.providers` with non-null `link_by`, use `link_by` to obtain a value from the new identity. If the value is empty, do nothing.
  2. Use the value we get in 1 to query for existing oauth identities with the same claim specified in `link_by`.
  3. If there is at least one identify found in step 2,
    - If `identity.on_conflict.signup=error`, return an error and terminate the signup flow.
    - If `identity.on_conflict.signup=login`, trigger login flow for that existing user, and discard the new identity.
    - If `identity.on_conflict.signup=login_and_link`, trigger login flow for that existing user, and add the new identity to that user after the flow was completed sucessfully.
    - If `identity.on_conflict.signup=create_new_account`, create a new user with the new identity.
    - If `identity.on_conflict.signup=hook`, trigger a hook and use the result to determine what to do.

## Future enhancements

- Currently, the developer may have no idea how the claims are generated from the external idp user claims if they did not look at the source code. In the future, we should support customizing the transformation logic by providing custom script or hook so that the developer has more control on the linking behavior.

- Only `error` `login` `login_and_link` will be implemented for `identity.on_conflict.signup` at the moment. `create_new_account` and `hook` could be added in a later stage.

## Caveats

- Wechat
  - Account linking is probably not usable with wechat because the wechat profile doesn't provide any information about email or phone number.

- For `identity.on_conflict.signup=login_and_link`, the user might add an email identity to an existing account with oauth login. This could be an unexpected side effect because the user probably just forgot he registered with oauth intead of email.

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
