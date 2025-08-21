# Use case

This document outlines a list of use cases.
Each use case may contribute to the design of Organization.

## Use case 1: How to enable Organization?

I am a developer who just signed up portal.authgear.com, ready to create my first project.

### Use case 1.1: Organization must be enabled during project creation

During the project creation, the first question I am asked is to choose whether Organization is enabled or not.

> [!NOTE]
> Stytch is a competitor doing this.
> When you create a new project at Stytch, they ask you whether you want Consumer Authentication or B2B Authentication.


### Use case 1.2: Organization is opt-in at anytime

During the project creation, Organization is not mentioned at all.

> [!NOTE]
> Auth0 follows this style.
> There is no a single switch to turn the project to fully Organizational.
> Instead, the developer has to read through the docs,
> and makes the project configuration to enable Organization.

### Use case 1: Design decision

Prefer [Use case 1.2](#use-case-12-organization-is-opt-in-at-anytime)

> [!IMPORTANT]
> I tend to keep the project creation simple.
> At the end of the project creation,
> we can add a link to redirect the developer to enable Organization if
> that is what he is looking for.

## Use case 2: Can an end-user exist without being a member of an Organization?

### Use case 2.1: No, every end-user must be a member of an Organization

> [!NOTE]
> Stytch B2B Authentication follows this style.
> Under Stytch B2B Authentication, all end-user must be a member of an Organization.

It follows naturally that **End-users can create Organization in a self-serve fashion**.

For example, this is how this works in Stytch B2B Authentication.

When `louischan@oursky.com` signs up,
and there is no Organization claiming the `oursky.com` domain,
`louischan@oursky.com` is forced to create an Organization for `oursky.com`.
He will become the admin of the created Organization.

### Use case 2.2: No. But Organization needs approval.

Basically this is the same as the previous use case.
But the created Organization is not approved.
The project admin has to approve the Organization creation.

The end-user would just see `Your organization is being reviewed. Please come back later`.

> [!NOTE]
> Actually I doubt the usefulness of this use case.
> Stytch B2B Authentication is clearly designed for developers who are developing SaaS.
> In that approval is not needed.
> Just like you do not need approval in project creation.

### Use case 2.3: Yes, end-user can belong to no Organizations.

When an end-user signs up via the project sign up URL `https://auth.myproject.com/signup`.
He ends up with a User without belonging to any Organizations.

Given that the signed-in Organization is reported in the ID token,
the developer detect this situation, and do whatever he wants.
He can

- Allow the end-user to continue using the app, as long as the business requirements allow.
- Force the end-user to create an Organization. The developer would use the Admin API to create Organization and add the end-user as member.

Or, if the end-user signs up with an email address,
which happens to be an auto-membership domain of an Organization,
then the end-user becomes member of that Organization.

> [!NOTE]
> Auth0 follows this style.
> Organizations can only be created at Dashboard or via the Management API.
> See https://auth0.com/docs/manage-users/organizations/configure-organizations/create-organizations

### Use case 2: Design decision

Prefer [Use case 2.3](#use-case-23-yes-end-user-can-belong-to-no-organizations)

Use case 2.3 is easier to implement,
and it matches our existing behavior more closely.

The advantage of supporting Use case 2.1 is that the developer can have a zero-code solution,
if he is building SaaS.

Supporting Use case 2.1 and Use case 2.2 requires us to support creating Organization in Auth UI,
which requires non-trivial amount of effort.

## Use case 3: What are common between User and Member?

To discuss this, we first need to know what a User owns

- `disabled`, `disabled_reason`, `is_deactivated`
- `standard_attributes` and `custom_attributes`
- `delete_at`
- `is_anonymized`, `anonymize_at` and `anonymized_at`
- `mfa_grace_period`

### Use case 3.1: Disable a Member in an Organization

Instead of disabling the User, making he unable to sign in at all,
it may be favorable to only disable the Member to sign in to an Organization.
The User can still sign in other Organizations.

### Use case 3.2: Member-specific Standard Attributes and Custom Attributes

Suppose the project has a custom attribute `job_title`.
It is likely that the end-user will have different `job_title` in different Organizations.

Instead of defining how User-specific and Member-specific Standard Attributes and Custom Attributes,
it is better make Member-specific Standard Attributes and Custom Attributes a feature to be turned on.
This feature is project-wise.

When Member-specific Standard Attributes and Custom Attributes are turned on:

- Standard Attributes are populated when the User becomes a Member.
  - The `email` attribute is populated if the Organization does not define a list of allowed domains.
  - If the Organization defines a list of allowed domains, and the User has a verified email in one of the allowed domain, the `email` is also populated.
  - Other Standard Attributes are copied from the User.
- Custom Attributes **ARE NOT** populated. The developer has to update Custom Attributes themselves.

> [!WARNING]
> There may be a few caveats, but I can think of one at the moment.
> The blocking hook mutations mutate User, not Member.
> It is confusing that it sometimes mutates User, sometimes mutates Member, depending on the feature is turned on.

### Use case 3.3: Scheduled removal of Member from Organization

Suppose Users and Organizations are used to model employees and their companies.

When an employee resigns, usually they will not leave the company immediately (unless they got fired).
So it is nice to have scheduled removal of Member from Organization.

### Use case 3.4: Anonymization

Anonymization in User is a way to keep the User ID while removing all identities and authenticators.

As long as we **NEVER expose Member ID**, we do not need Anonymization in Member at all.

Simply deleting the Member is enough.
In case the User rejoins, create a new Member.

### Use case 3.5: MFA grace period

Given that we adopt the Member model,
the User share the same set of authenticators across all Organizations he is a member of.

It really depends on whether Organizations are allowed to have different MFA grace period settings.

Personally I do not see a need for that.

### Use case 3: Design Decision

As explained in Use case 3.4 and Use case 3.5, they are dropped.

Use case 3.3 is nice-to-have, it is not a blocker.

Use case 3.1 is relatively easy to implement, but in MVP, but I think it is not a must for MVP.

If Use case 3.2 is unavailable, a workaround is possible if it is acceptable to have multiple Users.

For example, instead of having

```
Organization (org_id=oursky)
  Oursky (@oursky.com)

Organization (org_id=yoursky)
  Yoursky (@yoursky.com)

User (user_id=louischan)
  Google Workspace (louischan@oursky.com)
  Google Workspace (louischan@yoursky.com)

Member (user_id=louischan, org_id=oursky)
  Name: Louis Chan
  email: louischan@oursky.com
  job_title: Software Engineer

Member (user_id=louischan, org_id=yoursky)
  Name: Louis Chan
  email: louischan@yoursky.com
  job_title: Software Consultant
```

We can have multiple Users.

```
Organization (org_id=oursky)
  Oursky (@oursky.com)

Organization (org_id=yoursky)
  Yoursky (@yoursky.com)

User (user_id=louischanoursky)
  Google Workspace (louischan@oursky.com)
  Name: Louis Chan
  email: louischan@oursky.com
  job_title: Software Engineer

User (user_id=louischanyoursky)
  Google Workspace (louischan@yoursky.com)
  Name: Louis Chan
  email: louischan@yoursky.com
  job_title: Software Consultant

Member (user_id=louischanoursky, org_id=oursky)

Member (user_id=louischanyoursky, org_id=yoursky)
```

## Use case 4: How can a User becomes Member of Organization?

### Use case 4.1: Add a User as a Member of Organization via Admin API

This is trivial.

### Use case 4.2: Direct invitation URL

This invitation URL is intended for a single email address only.

The lifetime of the invitation URL is configurable.
The default is 3 days.

This use case comes in 3 flavors:

1. An OIDC client application is selected **AND** the OIDC client application must implement `initiate_login_uri`.
2. No client application is selected.
3. A SAML client application is selected **AND** the SAML client application must support IdP-initiated SSO.

> [!WARNING]
> Flavor 1 requires us to implement https://openid.net/specs/openid-connect-core-1_0.html#ThirdPartyInitiatedLogin first

> [!WARNING]
> Flavor 3 requires us to implement SAML IdP-initiated SSO first.
> See https://groups.oasis-open.org/higherlogic/ws/public/download/56782/sstc-saml-profiles-errata-2.0-wd-07.pdf Section 4.1.3 and section 4.1.3.5 for details.

In Flavor 1, when the URL is visited, Authgear redirects to `initiate_login_uri` with `iss` and `login_hint`.
The client application uses Authgear SDK to `authenticate()`, passing `login_hint` down the flow.

The `login_hint` encapsulates:

- The Organization, so a Organization-specific Auth UI is shown
- The intended email address, so authenticated as another email is forbidden.
- A one-time use invitation code.

If the authentication is successful, the User becomes Member of the Organization.

In Flavor 2, when the URL Is visited, the end-user is shown a Organization-specific Auth UI.
Similar to Flavor 1, the invitation is one-time and email address is checked.
The end-user is shown a page `You are now member of Organization. You may close this page now.`.

In Flavor 3, it is the same as Flavor 2, except that at the end,
The end-user is redirected to the SAML SP ASC URL with the SAML Response.

> [!NOTE]
> In Direct Invitation URL, no approval is required.
> This is different from what Ben told me initially.

### Use case 4.3: Public Invitation URL

Public Invitation URL does not have intended email address.

The lifetime of Public Invitation URL can be:

- Never expire. An Organization can at most have 1 Public Invitation URL that never expire.
- Expire in a specific duration, expressed in days. An Organization can at most have 10 Public Invitation URL that expires.

The number of use of Public Invitation URL can be:

- One-time use. That is, the URL expires after a successful use.
- N time use, where N is configurable number at the time of creation.
- Unlimited.

The lifetime and the number of use work independently of each other.
Public Invitation URL expires whichever the condition becomes true first.

Public Invitation URL by default does not make the User a Member of the Organization.
Instead, the User must be approved first.
Public Invitation URL can be configured to skip approval.

Otherwise, Public Invitation URL is similar to Direct Invitation URL.

### Use case 4.4: Auto-membership

When the end-user signs up or signs in via a URL with Organization pre-determined by the developer:

- The Organization is known at the beginning.
- The Login Methods of the Organization is shown.
- Suppose the end-user authenticates with `johndoe@example.com`, but `@example.com` is not an auto-membership domain, an error is shown.
- Otherwise the User becomes a Member.

When the end-user signs up or signs in via the project-specific URL:

- The Login Methods is shown. They may be a superset of that of an Organization.
- Suppose the end-user authenticates with `johndoe@example.com`.
- For each Organization, if `example.com` is an auto-membership domain, the User becomes Member.
- The end-user is shown a list of Organizations he is Member of.

> [!IMPORTANT]
> The number of organizations can be large.
> We need to store the auto-membership domain in a way that allow fast lookup.

> [!NOTE]
> The list of auto-membership domain is independent of Login ID email domain allowlist and blocklist.
> That is, if the domain blocked by the Login ID email domain allowlist or blocklist,
> The sign up is blocked before auto-membership has a chance to take effect.

> [!NOTE]
> Auto-membership only works one way.
> It does not do auto-remove membership.

> [!NOTE]
> In some very sensitive software like 1Password,
> Auto-membership requires approval.
> This could be an Organization-specific configuration,
> and be handled like Use case 4.3.

The auto-membership domains are stored in the following database table.

```
CREATE TABLE _auth_organization_auto_membership_domain (
  id PRIMARY KEY text,
  app_id NOT NULL text,
  org_id NOT NULL text,
  created_at NOT NULL timestamp,
  updated_at NOT NULL timestamp,
  domain NOT NULL text
)
CREATE UNIQUE INDEX _auth_organization_auto_membership_domain_uniq _auth_organization_auto_membership_domain (org_id, domain)
```

### Use Case 4: Design Decision

> [!NOTE]
> Fung suggests that we target Use case 4.1 and Use case 4.4 for MVP

## Use case 5: How to dissociate a User from an Organization?

### Use case 5.1: Remove a User from Member of Organization via Admin API

This is trivial.

### Use case 5.2: A User can leave an Organization in the settings page

This is a rabbit hole because there could be many use cases that
the developer may allow a User to leave an Organization.

Let me name a few here.

- Members in an Organization could have Roles. Depending on the Roles a Member has, he may or may not leave as he wishes.
- The developer simply do not want Members to leave in a self-serve way.
- The developer allows Members to leave freely.

> [!NOTE]
> Auth0 support removing members from organizations via the dashboard or the Management API.
> See https://auth0.com/docs/manage-users/organizations/configure-organizations/remove-members

### Use case 5: Design Decision

We implement Use case 5.1 for MVP.

## Use case 6: How do we maintain the validity of the membership?

In Use case 4, we discussed the ways that a User can become Member of Organization.

In this use case, we discuss how do we maintain the validity of the membership.

When we read through Use case 4, we know that if a User has an allowed email address,
then the User is entitled to be a Member of an Organization.

In other words, as long as the User owns the email address, the User can be a Member.

It implies that updating the email address or removing the email address may not be desirable.

### Use case 6.1: Disallow Member to add, remove, or update Identities

> [!NOTE]
> TODO
> 2025-08-06: Tung asked how does Account Linking interact with this use case.
> I need to think about that.

> [!NOTE]
> TODO
> 2025-08-06: Ben mentioned that if the developer's application is GitHub-style, it is weird to disallow adding, removing, or updating emails.
> I need further study on that.

> [!NOTE]
> In Stytch B2B Authentication, it is forbidden to add, remove, or update own email address.

> [!NOTE]
> In Kinde, Users with Enterprise Identities cannot have other identities.
> See https://docs.kinde.com/authenticate/enterprise-connections/about-enterprise-connections/#how-identities-are-handled-in-enterprise-connections

When a User is a Member of an Organization,
the User is not allowed to add, remove, or update the following Identities:

- Email Login ID Identity
- Phone Login ID Identity
- Username Login ID Identity
- OAuth Identity
- LDAP Identity

The following Identities can still be added and removed:

- Biometric Identity
- Passkey Identity

#### Use case 6.1.1: Organization with Federated Login

Refer to Use case 11 to learn what Federated Login is.

- The end-user is not expected to add Identities. Federated Login means the end-user must sign in via the external Identity Provider. Adding Identities (and thus adding more login methods) are not useful at all.
- The end-user is not expected to update Identities. As long as the subject ID stays the same, Standard Attributes `email`, `phone_number` and `preferred_username` are updated automatically on login. The end-user need not update them manually.
- The end-user is not expected to delete Identities. In normal case, the User has only one Identity, which cannot be deleted anyway.

#### Use case 6.1.2: Organization with auto-membership

Since auto-membership works one way, allowing the user to update or remove Identities may contradict with the result of auto-membership.

Adding Identities generally does not contradict with auto-membership though.

### Use case 6: Design Decision

We implement User case 6.1 for MVP.

## Use case 7: UI/UX of Organizational signup and login

### Use case 7.1: Organizational signup and login in OIDC

In Auth0, depending on the selected "Login Experience", the developer can further configure "Login Flow"

- Login Experience - Individuals
- Login Experience - Business Users
  - Login Flow - Prompt for Credentials
  - Login Flow - Prompt for Organization
  - Login Flow - No Prompt
- Login Experience - Both
  - Login Flow - Prompt for Credentials
  - Login Flow - No Prompt

> [!NOTE]
> "Login Flow - No Prompt" means the developer has to specify the Organization in the authentication request.

If we try to encode the 2 enums into 1 enum, we have:

> [!NOTE]
> TODO
> 2025-08-06: Tung confused that `prompt_end_user_for_organization_first` means
> we construct a list for the end-user to choose,
> while it means prompting the end-user to enter Organization Slug.
> I need a rename is needed.

| enum                                                                                        | Auth0 equivalent                                                         | Description                                                                                                                                                                                              |
| ---                                                                                         | ---                                                                      | ---                                                                                                                                                                                                      |
| `x_organization_behavior=only_non_member`                                                   | Login Experience - Individuals                                           | This is the default value because it is back compatible with the pre-organization era. No prompts on Organization. The end-user signs in without Organization.                                           |
| `x_organization_behavior=only_member:prompt_end_user_for_organization_last`                 | Login Experience - Business Users + Login Flow - Prompt for Credentials  | For signups, it is expected that the signed up User will be made Members of some Organizations via auto-membership, otherwise the end-user will be shown an error screen as a dead end.                  |
| `x_organization_behavior=only_member:prompt_end_user_for_organization_first`                | Login Experience - Business Users + Login Flow - Prompt for Organization | The end-user is expected to know the Organization Slug. Like Auth0, if the end-user enters an invalid Organization Slug, an error is shown immediately.                                                  |
| `x_organization_behavior=only_member:developer_specified_organization`                      | Login Experience - Business Users + Login Flow - No Prompt               | The developer **MUST** specifies which organization to sign in. It is an OAuth Error or SAML error if organization is unspecified by the developer.                                                      |
| `x_organization_behavior=either_member_or_non_member:prompt_end_user_for_organization_last` | Login Experience - Both + Login Flow - Prompt for Credentials            | The end-user is prompted to select "No Organization" and the Organizations he is a member of. The "No Organization" option always exist.                                                                 |
| `x_organization_behavior=either_member_or_non_member:developer_specified_organization`      | Login Experience - Both + Login Flow - No Prompt                         | The developer **OPTIONALLY** specifies which organization to sign in. If unspecified, it behaves the same as `x_organization_behavior=either_member_or_non_member:prompt_end_user_for_organization_last` |

In Auth0, the query parameter `organization` can be used by the developer to specify the Organization.
See https://auth0.com/docs/manage-users/organizations/using-tokens#authenticate-users-through-an-organization
I propose we support an equivalent with a different name `x_org_slug`.
It starts with `x_` like all existing proprietary query parameters like `x_sso_enabled`.
And it has `_slug` in it, it is clear to the developer that they should provide an Organization Slug.

When Organization is not known at the beginning, Organization-specific configuration **IS NOT** applied.
This implies the authentication **COULD** be invalidated by the choice of Organization.

For example, the end-user signs in with Email Login ID and password, select an Organization with Federated Login enabled.
Then the end-user has to restart the authentication from the beginning.

For example, the end-user signs in with Email Login ID and password, select an Organization that requires MFA.
In this particular case, it makes more sense to only require the end-user to do MFA only, rather than restarting the authentication from the beginning.

For example, the end-user signs in with Email Login ID and password, select an Organization that has a strict password policy that the current password does not meet.
In this particular case, the end-user has to change the password in order to complete the login.

`prompt_end_user_for_organization_last`, in many ways, does not work well as Authentication Flow.

> [!IMPORTANT]
> Discussion point: How to fit `prompt_end_user_for_organization_last` with Authentication Flow?
>
> 1. How do we model `prompt_end_user_for_organization_last`? As a new step in authflow?
> 2. Once organization is known, the generated authflow may change. Do we compute a diff between the executed authflow with the newly generated authflow? If yes, do we execute the diff?

### Use case 7.2: Organizational signup and login in SAML

In Auth0, to provide developer-provided organization, a query parameter `organization` can be added to the SAML Login URL.
See https://auth0.com/docs/authenticate/single-sign-on/outbound-single-sign-on/configure-auth0-saml-identity-provider#configure-saml-sso-on-the-service-provider

Alternatively, we can make use of `<Extensions>`, as specified in Section 3.2.1 in https://groups.oasis-open.org/higherlogic/ws/public/download/56776/sstc-saml-core-errata-2.0-wd-07.pdf
to allow specify the intended Organization Slug in `<AuthnRequest>`.

For consistency, I propose we support query parameter `x_org_slug` in the SAML Login URL, like what we support in Use case 7.1.

### Use case 7.3: Home Realm Discovery

The discussion cannot begin unless we have a clear definition of what Home Realm Discovery.

> [!IMPORTANT]
> This feature is called Email Domain Discovery in Fung's proposal.
> However, competitors like Auth0 and Kinde calls it Home Realm Discovery,
> and the mechanism of how this works is quite different.

> [!IMPORTANT]
> This is the definition of Home Realm Discovery.
>
> The following conditions must hold:
>
> - The developer does not specify `x_org_slug`.
> - The URL does not include an Organization Slug.
> - The end-user lands on the signup / login page of Auth UI.
> - The end-user is shown an email input field.
> - The end-user enters an email address.
> - The system **DOES NOT** authenticate the end-user, and redirects the end-user to an external Identity Provider.

#### Use case 7.3.1: Auth0 Home Realm Discovery

The official documentation is at https://auth0.com/docs/authenticate/login/auth0-universal-login/identifier-first

This is how it is set up:

- Create an Organization
- Add an Enterprise Connection to the Organization, such as Microsoft Entra ID.
- In the Enterprise Connection, configure a domain that will redirect any Users with that domain to this connection.

This is how it works:

- The end-user lands on the Universal Login page.
- The end-user enters an email in the input field.
- If the domain of the email matches the domain of the Enterprise Connection, the user is redirected to the Enterprise Connection to sign in.

This is how it **DOES NOT WORK**:

- https://community.auth0.com/t/hrd-redirect-affecting-all-organizations-with-same-email-domain-expected-behavior-or-misconfiguration/138495

#### Use case 7.3.2: Stytch Discovery Email OTP

> [!NOTE]
> Stytch does not support Home Realm Discovery natively.
> We, however, discuss it here
> because readers may wonder what it is as the feature is called "Discovery Flow".

In Stytch, it supports either [Discovery Flow](https://stytch.com/docs/b2b/guides/ui-components/discovery-flow) or [Organization Flow](https://stytch.com/docs/b2b/guides/ui-components/organization-flow)

Discovery Flow means the end-user lands on a generic signup / login page, authenticates themselves, and select Organization in the last step.

Organization Flow means the end-user lands on Organization-specific page, authenticates themselves to that Organization directly.

#### Use case 7.3.3: Kinde Home Realm Discovery

The official documentation is at https://docs.kinde.com/authenticate/enterprise-connections/home-realm-discovery/

Some sensible restrictions imposed by Kinde:

- When you configure a Microsoft Entra ID connection, you can set up a Home Realm domain for it.
- All Home Realm domains must be unique. They cannot repeat in different connections.

Some magical default behaviors by Kinde:

- If all Enterprise connections has Home Realm domain configured, then the authentication page shows a email input field only. The end-user is expected to enter an email manually, and then based on the email, be redirected to a corresponding Identity Provider.

#### Use case 7.3.4: How do we shape Home Realm Discovery in Authgear

> [!NOTE]
> TODO
> 2025-08-06: Ben challenged that
>
> - What exactly is Home Realm Discovery in Authgear?
> - It can exist without Organization.
>
> He listed a number of use cases in logins:
>
> 1. Google Workspace is the sole IdP.
> 2. Any @oursky.com can login.
> 3. Enter email, pass email OTP, show which Organization to sign in (Stytch-style).
> 4. Enter email, primary authenticate, show which Organization to sign in (Auth0-style).
> 5. (Variant of 4) Enter email, redirect to the sole external OAuth Provider.
>
> I need to think deeper what HRD is, and what use cases does it fulfill.

1. From use case 7.3.1 and use case 7.3.3, we realize that Home Realm Domain is associated with a Connection, rather than an Organization.
2. From use case 7.3.3, we realize that Home Realm Domains are unique across connections, not across Organizations.
3. Home Realm Discovery conflicts with `developer_specified_organization`. `developer_specified_organization` has a high precedence.
4. Home Realm Discovery conflicts with `prompt_end_user_for_organization_first`. `prompt_end_user_for_organization_first` has a high precedence.
5. When Home Realm Discovery is enabled, the email field is always shown, even Email Login ID is disabled.

However, we do not have Connection in Authgear, so the closest design we can make is **associate a Home Realm Domain with an Organization**.

This is how it works

- The developer enables Home Realm Discovery in the project configuration. This is required because we need to know whether we have to always show the email input field in the general authentication page.
- The developer creates an Organization.
- The developer associates a Home Realm Domain with the Organization. This information has to be stored in a database table for fast lookup. The Home Realm Domain **MUST** associate with an OAuth Provider in the Organization-specific configuration.
- The end-user lands on the general authentication page.
- The end-user enters an email address.
- If the email address matches one of the Home Realm Domains,
  - Then the end-user is redirected to the Organization-specific authentication page.
  - Else if Email Login ID is enabled,
    - Then the end-user proceeds to sign up with Email Login ID.
    - Else an error is shown. This implies Home Realm Discovery is inherently prone to account enumeration.

### Use case 7: Design Decision

For MVP, we can implement parts of Use case 7.1:

- `x_organization_behavior=only_non_member`
- `x_organization_behavior=only_member:prompt_end_user_for_organization_first`
- `x_organization_behavior=only_member:developer_specified_organization`

Other variants requires `prompt_end_user_for_organization_last` to be sorted out first.

## Use case 8: Developer-facing Authentication Output

This use case discusses various developer-facing Authentication Output.

### Use case 8.1: `org_slug` in ID Token and JWT Access Token

> [!NOTE]
> TODO
> 2025-08-06: Tung suggested that we should either:
> - Rename `org_slug` to `x_org_slug` to stay consistent with the query parameter.
> - Rename `org_slug` to URL-style, to stay consistent with the rest of our custom claims.

> [!NOTE]
> Auth0 by default includes `org_id` in ID token.
> It also offers a tenant-wise configuration to include `org_name` in ID token.
> See https://auth0.com/docs/manage-users/organizations/using-tokens#authenticate-users-through-an-organization and
> https://auth0.com/docs/manage-users/organizations/configure-organizations/use-org-name-authentication-api

When authentication is done via OIDC (The is the most common case), the developer wants to know which Organization the Member has signed in.

Organization has a developer-provided unique Slug across the Project.
We include the Organization Slug in the ID token instead of the generated ID of the Organization.

This is an example of ID token:
```json
{
  "iss": "https://auth.myapp.com",
  "aud": ["mobileapp"],
  "sub": "johndoe",
  "scope": "openid offline_access",
  "org_slug": "myorg"
}
```

This ID token implies:

- The User `johndoe` is a Member of `myorg`.
- The User `johndoe` signed in `myorg`.

> [!NOTE]
> This proposal is different from what Fung proposed in Linear.

### Use case 8.2: `member_orgs` in UserInfo

> [!NOTE]
> Auth0 does not support this out-of-the-box.
> The developer is expected to make use of the Management API to find out what Organizations the User is Member of.

We first need to answer this question:

> Is it appropriate to disclose what Organization the User is Member of,
> given that the User may be a Member of an Organization that has a more strict password policy, MFA requirement, or Federated Login.

> [!NOTE]
> TODO
> 2025-08-06: Tung disagreed with my answer to this question.
> He said it is inappropriate to disclose `member_orgs` via UserInfo endpoint,
> as it is accessible by the authorized client,
> while the Auth UI is accessible by the end-user only.
>
> Maybe First Party client (which has full-access) has access.
> We need to introduce an Authgear scope for Third Party client.

Actually this question also applies to `prompt_end_user_for_organization_last` in Use case 7.1.
If we agree that it is appropriate to have Use case 7.1, then the answer to this question is "Yes, it is appropriate to disclose such information, as long as the User is identified and authenticated".

Another potential problem is that the User could be a Member of many Organizations.

> [!IMPORTANT]
> Should we limit the maximum number of Organizations a User can be a member of? Is 100 a reasonable hard limit?

This use case could be essential for implementing Organization Switching.

In Organization Switching we commonly see in online services,
the following information are shown

- A display name of the Organization. This is different from the Organization Slug.
- An icon of the Organization, if the Organization has one.
- The Organization Slug, either as a standalone string, or be interpolated as a URL.

So the developer wants information like

```json
[
  {
    "org_display_name": "Oursky",
    "org_slug": "oursky",
    "org_icon_url": "https://oursky.com/favicon.png"
  },
  {
    "org_display_name": "FormX",
    "org_slug": "formx",
    "org_icon_url": "https://formx.ai/favicon.png"
  }
]
```

Given that this information is not bound to session (unlike Use case 8.1), maybe we should include this information in UserInfo instead.
Not including this information in tokens can ensure the size of the tokens is small, and does not grow proportional to the number of Members.
Access Token is usually included in HTTP header, and HTTP headers commonly has a practical size limit of 4000 bytes to 8000 bytes.

The final proposal is to include `member_orgs` in UserInfo.

Here is an example:

```json
{
  "sub": "johndoe",
  "member_orgs": [
    {
      "org_display_name": "Oursky",
      "org_slug": "oursky",
      "org_icon_url": "https://oursky.com/favicon.png"
    },
    {
      "org_display_name": "FormX",
      "org_slug": "formx",
      "org_icon_url": "https://formx.ai/favicon.png"
    }
  ],
  "email": "louischan@oursky.com"
}
```

### Use case 8.3: Authentication Output in client applications using cookies

When the client application uses cookies (Likely the applications are SSR),
the developer resolve the cookie into a number of headers with the resolve endpoint. This feature is specified in [../api-resolver.md](../api-resolver.md)

Similar to Use case 8.1, we need to report to the developer which Organization the User has signed in to.

I propose we add this new header:

> x-authgear-session-org-slug
>
> If this header is absent, the User did not signed in to an Organization.
> If this header is present, the value indicates the Organization Slug the User signed in to.
>
> Example:
> x-authgear-session-org-slug: myorg

To support a similar use case like Use case 8.2,
it is **not recommended** to use the cookie to impersonate the end-user to access the UserInfo endpoint directly.

Instead, the client application (which should be a backend service) is recommended to query the Members of the User via the Admin API.

### Use case 8.4: Authentication Output in SAML

> [!NOTE]
> TODO
> 2025-08-07: Tung said that I should study other specified values of NameID.
> He said there may exist a NameID that allow encoding of Organization.
> He also said the existing feature of attribute mapping maps OIDC claims to AttributeStatement,
> so maybe we just by default map `org_slug` there.

The below XML snippet lists out a list of possible places we can put the Organization the User signed in to.

```
<Assertion>
  <AuthnStatement>
    ...
    <AuthnContext>
      <AuthnContextClassRef></AuthnContextClassRef>        <!-- Optional -->
      <AuthnContextDecl></AuthnContextDecl>                <!-- Optional -->
      <AuthnContextDeclRef></AuthnContextDeclRef>          <!-- Optional -->
      <AuthenticatingAuthority></AuthenticatingAuthority>  <!-- Zero or More -->
    </AuthnContext>
  </AuthnStatement>
  <AttributeStatement>
    ...
  </AttributeStatement>
</Assertion>
```

- `AuthnContextClassRef`: Example values include `urn:oasis:names:tc:SAML:2.0:ac:classes:Password`. Irrelevant.
- `AuthnContextDecl` and `AuthnContextDeclRef`: A document of `<AuthnContext>` or a URI to it. Irrelevant.
- `AuthenticatingAuthority`: Example values include `https://sso.company.com`. Irrelevant.

So the remaining candidate is `AttributeStatement`.
The official description of it is:

> The <AttributeStatement> element describes a statement by the SAML authority asserting that the
> assertion subject is associated with the specified attributes.

However, the signed Organization is associated with the particular SAML AuthnRequest, not associated with the subject in general.

> [!IMPORTANT]
> Discussion point: How to report signed in Organization in SAML Assertion?

Similar to Use case 8.3, the SAML SP is recommended to query the Members of the User via the Admin API.

### Use case 8: Design Decision

For MVP, implement
- Use case 8.1
- Use case 8.2
- Use case 8.3

## Use case 9: What information does an Organization contain?

> [!NOTE]
> In Auth0, an Organization contains:
> - Name - An unique identifier that can be updated.
> - Display Name - A name that will be displayed to end users.
> - Logo URL - An external URL to an image. Does not support upload image. (Tenant Logo works like this as well.)
> - Primary Color - Change the color of CTAs.
> - Page background color - Change the page background color.
> - Metadata - Advanced feature hidden behind a UI toggle.
>
> The metadata is not exposed by default in anyway.
> It is expected to include it by post-login action, like in
> https://community.auth0.com/t/add-organization-metadata-in-token/90610

> [!NOTE]
> In Fung's proposal, the customizable UI items are:
> - Logo
> - Primary color
> - Page background color

### Use case 9.1: Organization Slug

- Organization is identified by `slug` within a Project.
- Organization Slug is required and cannot be empty.
- An Organization Slug, as the name suggests, appears in a URL.
- Only unreserved characters defined in [RFC3986 Section 2.3](https://datatracker.ietf.org/doc/html/rfc3986#section-2.3) can be used as Organization Slug.
- This regex describes the syntax of Organization Slug: `[-a-zA-Z0-9._~]+`.
- Organization Slug cannot be updated.

### Use case 9.2: Organization Name

- Organization can have an optional name.
- Organization Name will be displayed to end-users.
- Organization Name is not localizable.

### Use case 9.3: Create a new Organization

When create a new Organization in the portal, the developer is required to fill in:
- Organization Slug
- Organization Name (Even it is optional).

### Use case 9.4: Custom theme in Organization

> [!WARNING]
> This section mentions implementation details.
> It is expected that you do not understand some of the terms used here.

A new Organization Layer in resource filesystem is introduced to support custom theme.

The following information is stored in the layer:

- Organization Logo
- Organization Primary color
- Organization Page Background Color

> [!WARNING]
> What happen if the Project defines Page Background Image while an Organization defines Page Background Color?

### Use case 9: Design Decision

For MVP, implement

- Use case 9.1
- Use case 9.2
- Use case 9.3

## Use case 10: What Organization-specific configurations are available?

### Use case 10.1: authgear.organization.yaml

> [!IMPORTANT]
> Discussion point: What is the actual shape of the configuration?
>
> The customizable Organization-specific configuration is meant to be complete,
> but the actual shape is subject to discussion.

The following fields from `authgear.yaml` are Organization-specific:

- `authentication.identities` - Organization-specific identification methods.
- `authentication.primary_authenticators` - Organization-specific primary authenticators.
- `authentication.secondary_authenticators` - Organization-specific secondary authenticators.
- `authentication.secondary_authentication_mode` - Organization-specific MFA requirement.
- `identity.ldap` - Organization-specific LDAP servers.
- `identity.login_id.keys.key` - Only a list of keys is allowed. It is intended that the available keys are defined at Project level.
- `identity.oauth.providers` - Organization specific OAuth Providers.
- `authenticator.password.policy` and `authenticator.password.force_change` - Organization specific password policy.
- `authenticator.password.expiry` - Organization specific password expiry.
- `authenticator.oob_otp.sms.phone_otp_mode` - Organization specific OTP mode (Whatsapp, SMS)
- `authenticator.oob_otp.email.email_otp_mode` - Organization specific email OTP mode (Code, Login Link)

### Use case 10.2: The storage of Organization-specific configurations

> [!WARNING]
> This section mentions implementation details.
> It is expected that you do not understand some of the terms used here.

All Organization data are stored in the Organization table.
In particular, Use case 9.4 and Use case 10.1 are stored in resource filesystem style.

### Use case 10: Design Decision

The use case is an essential building block of Organization, so it must be implemented.

## Use case 11: Federated Login

> [!WARNING]
> The term "Federated Login" is not a common term for this feature.
> For example, if you search `competitor name` + "Federated Login",
> you will likely find unrelated results.
>
> It seems like we are misusing a very common term for a specific feature.

According to Fung's proposal, "Federated Login" is about:

- Configure an Organization-specific OAuth Provider. This is covered in Use case 10 already.
- If the Organization has one and only one applicable Identity Provider (OAuth or SAML), the end-user is redirected to it without seeing the Auth UI. This is just a QoL improvement of the Auth UI, and can be implemented independently of Organizations.

> [!NOTE]
> Auth0 does not support this feature out-of-the-box.
> The best you can do is to provide the `connection` parameter in the authorization endpoint.
> See https://auth0.com/docs/api/authentication/login/enterprise-saml-and-others#remarks
> That does not always work because you need to know in advance what Organization the end-user is going to sign in to.

### Use case 11.1: Federated Login in OIDC

We already support a proprietary authorization endpoint parameter named `x_provider_alias`.
If it is provided, then Auth UI will be skipped, the end-user will be redirected to that OAuth Provider directly.

We can support Federated Login in OIDC like this:

- When the Organization has specific configuration, and it has only one OAuth Provider, the default value of `x_provider_alias` is the alias of that OAuth Provider.
- If `x_provider_alias` is specified in the authentication request, it is respected.

So if the developer does not want the end-user to see the Auth UI,
they configure one and only one Organization-specific OAuth Provider.

This also means Federated Login is **implicit**.
By configuring one and only one Organization-specific OAuth Provider, Federated Login is enabled for the Organization.

### Use case 11.2: Federated Login in SAML

Since this is a feature of Auth UI, no special support is needed.

We may need to consider support `x_provider_alias` in the SAML Login URL.

### Use case 11: Design Decision

Implement Use case 11.1 for MVP.

## Use case 12: Built-in Organization Switching

Before we discuss Organization Switching,
we first have to acknowledge that all competitors associate a Session with at most 1 Organization.

For example, in Use case 7.1, regardless of `x_organization_behavior`,
the User signs in to 1 Organization, or no Organization.

For example, in Use case 8.1, `org_slug` is a string, not an array of strings.

### Use case 12.1: Organization Switching in OIDC

> [!WARNING]
> This use case requires Use case 8.2.

Among of the competitors,
only Stytch provides enough support for the developer to implement Organization Switching.
It support Organization Switching by [Exchange Session](https://stytch.com/docs/b2b/sdks/session-management/exchange-session#exchange-session)

> From the developer's point of view, what do they expect to see in the Session list of a User?

- User A signs in Organization A via Client A: See Session A in the listing, # Sessions = 1
- User A switches to Organization B in Session A: See Session A in the listing, # Sessions = 1

**The developer expects to revoke Session A to make User A completely signs out of Client A.**

> [!NOTE]
> TODO
> 2025-08-06: Tung raised a question.
> Can an admin of Organization A terminates a Session of a Member,
> but the Session actually is for Organization B?

> From the developer's point of view, do they expect to see `user.authenticated` fire again when switching Organizations?

My gut feeling is No.

> From the end-user's point of view, do they expect to be authenticated again if switching to an Organization that has equivalent configuration?

No.

> From the end-user's point of view, do they expect to be authenticated again if switching to an Organization that has Federated Login?

Yes, if the end-user has not done that Federated Login within the same Session before.

> From the end-user's point of view, do they expect to be authenticated again if switching back to an Organization that has no Federated Login?

No.

> From the end-user's point of view, do they expect to be authenticated again if switching back to the Organization that has Federated Login?

No, since they have done that already.

> From the end-user's point of view, do they expect to be authenticated again if switching to an Organization that requires MFA?

Yes, they expect MFA.

> From the end-user's point of view, do they expect to be authenticated again if switching back to another Organization, and then switching to the Organization that requires MFA?

No, since they have done that already.

---

The following desirable behaviors can be derived from the above Q&A.

- The end-user may expect no user interaction is involved if switching to an already authenticated Organization. This implies a Session should remember what Organizations the User has authenticated to.
- The developer may expect 2 conflicting behaviors regarding MFA
  - Behavior 1: Organizations have disjoint sets of secondary authenticators. The developer **DOES NOT** care about the actual method. As long as the end-user has done MFA once, the end-user will not be prompted again for MFA when switching Organizations.
  - Behavior 2: Organizations have disjoint sets of secondary authenticators. The developer **DOES** care about the actual method. The end-user **MUST** have the allowed MFA method done in order to switch to the Organization.
- If the switching involves additional authentication, it has to be done through the authentication endpoint.
- To indicate an existing Session, we can follow how we implemented reauthentication. We provide `id_token_hint` which embeds `sid` (Session ID).
- The developer has no idea whether the switching can result in additional authentication. Therefore, the developer expects a single entrypoint. It implies Organization switching always begins in the authentication endpoint. If no authentication is needed, the authentication endpoint can issue authorization code and redirect immediately.
- Since it always begins with the authentication endpoint and ends with an authorization code, the frontend can treat it like an ordinary authentication.
- Like reauthentication, no new Session is created. Instead, refresh token and access token are issued like they always do.

The following table summarizes the behavior in the authentication endpoint.

| Scenario                                                             | `login_hint`                                    | `id_token_hint` | `max_age`   | `prompt`                           | `x_org_slug`      |
| ---                                                                  | ---                                             | ---             | ---         | ---                                | ---               |
| Signup / Login                                                       |                                                 |                 |             | `prompt=login`                     |                   |
| Login with a specific email                                          | Specify the email address (not implemented yet) |                 |             | `prompt=login`                     |                   |
| Signup / Login with developer-provided Organization Slug             |                                                 |                 |             | `prompt=login`                     | `x_org_slug=SLUG` |
| Reauthenticate                                                       |                                                 | Must specify    | `max_age=0` | `max_age=0` implies `prompt=login` |                   |
| Organization switching with least user interaction                   |                                                 | Must specify    |             |                                    | `x_org_slug=SLUG` |
| (Subject to discussion) Organization switching with reauthentication |                                                 | Must specify    | `max_age=0` | `max_age=0` implies `prompt=login` | `x_org_slug=SLUG` |

> [!IMPORTANT]
> Discussion point: Do we support Organization switching with reauthentication? How do we do it?

### Use case 12.2: Organization Switching in SAML

In SAML, only a SAML assertion is returned to the SP.
The bearer of the SAML assertion **CANNOT** access Use case 8.2.

Therefore, Organization Switching is **IMPOSSIBLE** in SAML.

### Use case 12: Design Decision

I think this is an advanced use case that is optional in the MVP.

## Use case 13: UX of the settings page

> [!NOTE]
> TODO
> 2025-08-06: Ben mentioned the Settings page for Organization would be a caveat.

## Use case 14: Do we have Member ID?

> [!NOTE]
> TODO
> This is something I missed in the original draft.
> There is no public facing Member ID.
> We do not expose Member ID in any API.
