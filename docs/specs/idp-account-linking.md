# Identity Provider Account Linking

## Abstract

A single user could have accounts in multiple identity providers, such as Google, Facebook, Github. We want to provide a way to identify and link accounts from different identity providers into a single authgear account.

An account linking config section is added to configuration of each oauth provider config for this purpose. The developer must specify the account matching logic for the specific identity provider if they wish to turn on account linking feature for that identity provider.

## Configuration

Here is an example of the configuration:

```yaml
identity:
  providers:
    - alias: google
      client_id: exampleclientid
      type: google
      account_linking:
        enabled: true
        idp_claim_key: "email"
        match_against_claim_key: "email"
```

- The `account_linking` object was added.
- `account_linking.enabled`: A boolean indicates if account linking is enabled for this provider.
- `account_linking.idp_claim_key`: The key which used to read a claim value from from userinfo of the idp.
- `account_linking.match_against_claim_key`: The key which used to read a claim value from from userinfo of authgear. If the value matched with the value read using `account_linking.idp_claim_key` in the idp userinfo, an account linking will be performed.

## Trigger conditions and actions

- Whenever an oauth identity is used, and there is no existing user for that identity.

  1. If `account_linking.enabled` is not `true`, do nothing. Else,
  2. Get the value of claim from the idp profile using `account_linking.idp_claim_key`. If the value is empty, do nothing.
  3. Query from existing users using `account_linking.match_against_claim_key`, where the claim value is matching the value in 2.
  4. If there is exactly one match, add the oauth identity to the existing user as a new identity. Else, create a new user.

  - If there are more than one match, throw an error and reject the login.

- Whenever a new user create an identity during signup.
  1. For every `identity.providers` with `account_linking.enabled` equal to `true`, do:
  - 1.1. Use `account_linking.match_against_claim_key` to obtain a value from the new user. If the value is empty, do nothing.
  - 1.2. Use `account_linking.idp_claim_key` to query users with an oauth identity of the current provider matching the value in 1.1.
  2. (TBC) If there is at least one user found in 1, block the login and ask user to signin with with the found identity.

## Caveats

- Linkedin

  - The current integration called the me and contact apis, and combined them into one object as the profile. As the generated profile is not directly equals to the response of apis, user could have no idea what to set in `account_linking.idp_claim_key`.
  - Maybe we should re-implement the integration using the v2 api which is oidc compatible. https://learn.microsoft.com/en-us/linkedin/consumer/integrations/self-serve/sign-in-with-linkedin-v2

- Wechat
  - The user profile returned is not using oidc standard claims, but it shouldn't cause any problem as `account_linking.idp_claim_key` can be any keys.

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
