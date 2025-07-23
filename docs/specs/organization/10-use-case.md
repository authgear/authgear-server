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

### Use case 3: What are common between User and Member?

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

This is the simplest case.
No invitation is involved.
The Organization domain settings **IS RESPECTED**.

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
