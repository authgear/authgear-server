# A brief research on how Organization works in Stytch

## "Authentication Type" in Stytch

When you go to stytch.com and sign in for an account,
it will ask you what "Authentication Type" you want.
It is either "B2B Authentication" or "Consumer Authentication".

Once "Authentication Type" is chosen, it cannot be changed.
You must create a new project if you want another "Authentication Type".

Stytch also makes these two "Authentication Type" very distinct.
For example, they have their own docs home page.

- Consumer Authentication: https://stytch.com/docs/guides
- B2B Authentication: https://stytch.com/docs/b2b/guides

## Concepts in Stytch B2B

Read this https://stytch.com/docs/b2b/guides/what-is-stytch-b2b-auth

### Project

- Stytch B2B is equivalent to a project with "Authentication Type" being "B2B Authentication"
- A project has many organizations.

### Organization

- An organization controls its own authentication settings like allowed authentication methods (email magic link, email OTP, password, Google, etc.)

### Member

- A member MUST belong to an organization.
- A member is identified by an email address. Identification by username or phone number is not supported.
- Different members from different organization can share the same email address. 

## Concepts in Stytch Consumer

This is no official introduction document to explain Consumer Authentication.
I guess it is because it should be self-explanatory.

### Project

- Stytch Consumer is equivalent to a project with "Authentication Type" being "Consumer Authentication"
- Organization just does not exist in this project.
- Member just does not exist in this project.
- I happened to use the B2B example code to connect to a Consumer project.
  I discovered that the B2B example code connects to b2b specific endpoints (with /b2b/ in the path).
  The endpoint detects that the project is a Consumer project and prompts me I am doing something wrong.

### User

- Email addresses are unique within the project.

## Notable behaviors in Stytch

### No mixed "Authentication Type"

You must either choose "B2B Authentication" or "Consumer Authentication", not both.
It implies you CANNOT build a GitHub-style service with Stytch.

### Self-serve organization creation

If your project is just created and no organization exists, the end-user is prompted to create one.
There is no option to turn off self-serve organization creation.

When the end-user signs in with `louischan@oursky.com`, and there is no corresponding org for `oursky.com`.
The end-user is prompted to create one. The created organization will be named `Oursky`.
`louischan@oursky.com` of course become the admin of `Oursky`.

### Just-in-time (JIT) Provisioning / Discovery

It just means member can join an organization when they sign up.
JIT Provisioning is a settings of an organization.

### No invitation in the portal

Invitation is only available in the SDK.
In the portal, the member is always created immediately.

## How to do usecase X

This section explains how to do a certain usercase X in Stytch, with an example oriented around a company called FormX having a tenant (also called Formx) on Stytch.

FormX is a SaaS doing form extraction with self-built machine learning models, as well as LLM-powered models.
It serves both casual end-users like those with a `@gmail.com` email, and enterprise end-users.

### Usecase 1: Serving both casual end-users and enterprise end-users

You have to create two separate Stylch projects.
One Consumer project for casual end-users.
One B2B project for enterprise end-users.

And then you can create a separate organization to have different password policy.

> [!WARNING]
> Having two projects implies the following undesirable consequences (not intended to be exhaustive):
> - You can no longer view ALL users in the user management page.
>   You must switch the project.
> - You have to configure your API gateway / relevant services to accept two JWKs.
>   For example,
>   - https://test.stytch.com/v1/b2b/sessions/jwks/project-test-7acf993c-1688-4bc8-bd05-7110db1eda23
>   - https://test.stytch.com/v1/sessions/jwks/project-test-e19ead0f-8ad6-437f-8034-6c35c32e00df

### Usecase 2: Different password policies

Read https://stytch.com/docs/b2b/guides/passwords/overview

In short, you have to choose between "Cross-organization Password" and "Organization-Scoped Password".
This settings can only be changed when the project has no members.

### Usecase 3: User Isolation by Organization

In Stytch, a member belongs to one and only one organization, so User Isolation by Organization is supported out-of-the-box.

### Usecase 4: Enable MFA for some organization

In Stytch, each organization has its own authentication settings, so it is supported out-of-the-box.

### Usecase 5: Invitation

In Stytch, invitation is not supported in the portal.
You either have to use the API, and implement yourselves.
See https://stytch.com/docs/b2b/api/send-invite-email

Note that the project MUST have a configured `invite_redirect_url`.
See https://stytch.com/docs/workspace-management/redirect-urls

The invitation email will contain `invite_redirect_url` as the CTA.
The `invite_redirect_url` SHOULD BE an URL to your backend server,
and it is the backend server responsibility to drive the invitation flow to finish, by using Stytch SDK or Stytch API.

### Usecase 6: Email discovery

Read https://stytch.com/docs/b2b/guides/what-is-stytch-b2b-auth#core-flows

- Discovery Authentication
  - The end-user can create an organization in a self-served fashion.
- Organization-Specific Login
  - The organization is specific, the sign-in must result in the specific organization.

### Usecase 7: Organization switcher

Read the following
- https://stytch.com/docs/b2b/guides/what-is-stytch-b2b-auth#organization-switching
- https://stytch.com/docs/b2b/api/list-discovered-organizations
- https://stytch.com/docs/b2b/api/exchange-session

In Stytch, the member belongs to one and only one organization.
The session is bound to the member.
Switching organization is the same as "Exchange Session"
To query possible organizations that a member can switch to, Stytch offers "List Discovered Organizations"
Once you have a list of discovered organization, you can then use "Exchange Session".
Stytch will then perform necessary checking to determine whether the switch can be done right away,
or you have to go through additional steps to sign the end-user in.
