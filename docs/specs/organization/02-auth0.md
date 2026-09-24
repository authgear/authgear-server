# A brief research on how Organization works in Auth0

## Concepts

This section introduces a few concepts you must know in order to read this document.

### Tenant

- A tenant is like an Authgear project.
- It can have many connections, applications, and organizations.

### Connection

- An isolated user pool.
- Each connection presents a way to identity an user, and the means to authenticate them, for example, email-password, email-OTP, username-password, or Social Login Provider like Google Login.
- The same email address in 2 different email connections represent different users.
- For connection like email-password and username-password, each connection has its own password policy separate from each other.

### User

- A user belongs to one and only one connection.

### Application

- An application is an OIDC client application.
- Applications and connections form a N-to-M relationship. This relationship controls what users can sign in to a particular application.
- The above relationship does not hold always. It depends on the "Login Experience" option of your application.
  - When "Login Experience" is "Individuals", the application is NOT organization-aware, the application-connection relationship control what users can sign in.
  - When "Login Experience" is "Business Users", the application-connection relationship is ignored. Any users from any organizations can sign in.
    The "Login Flow" option further determines the behavior.
    - When "Login Flow" is "Prompt for Credentials", the end user enters email, authenticate with password, and finally got asked which organization to sign in to.
    - When "Login Flow" is "Prompt for Organization", the end user enters the organization slug first, and then enter email, authenticate with password.
    - When "Login Flow" is "No Prompt", you need to specify which organization to sign in to in the SDK integration code.

### Organization

- An organization is a way to organize users.
- Organizations and users form a N-to-M relationship through what Auth0 calls Membership.
- Organizations and connections form a N-to-M relationship. This relationship controls what users can belong to a particular organization.
- There is no relationship between organizations and applications. See above to understand how applications and organizations work together.

## How to do usecase X

This section explains how to do a certain usercase X in Auth0, with an example oriented around a company called FormX having a tenant (also called Formx) on Auth0.

FormX is a SaaS doing form extraction with self-built machine learning models, as well as LLM-powered models.
It serves both casual end-users like those with a `@gmail.com` email, and enterprise end-users.

### Usecase 1: Different password policies

For casual end-users, FormX does not want to impose a very restrictive password policy on them.
FormX just wants their casual end-users to have a password of at least 8 characters long.

FormX also serves an enterprise called GreatMall, which require a more restrictive password policy of 16 characters long.

To fulfil this usecase, FormX does

- Create a connection `casual-email-password` to store casual end-users. The password policy is 8 characters long.
- Create a connection `greatmall-email-password` to store end-users from GreatMall. The password policy is 16 characters long.
- Create an organization `greatmall`.
  Assign `greatmall-email-password` as the only connection to organization `greatmall`.
  Restrict the email domain of organization `greatmall` to `@greatmall.com`.
  Turn on auto-membership.

### Usecase 2: User Isolation by Organization

Most cloud providers offer User Isolation by Organization to allow their customers to manage User Isolation by Organization users within their account.

In FormX, GreatMall also wants User Isolation by Organization.
Since in Auth0, connection is already a isolated user pool. This requirement is automatically fuifilled.

### Usecase 3: Enable MFA for some organization

FormX does not want to force their casual end-users to enroll MFA mandatorily.
However, GreatMall requires all of its employee to enable MFA when its employee sign in subprocessor services like FormX.

In Auth0, MFA is a tenant-wise option which cannot be turned on and off for a particular connection, application, nor organization.
To implement this requirement, FormX has to use Auth0 post-login action to dynamically require MFA when the end-user is signing in to GreatMALL.

### Usecase 4: Invitation

One of the common usecase of organization is to invite someone in the organization to join it.
Often, the invitation is in form of an email with a link to sign in.
When the end-user signs in with the link, the end-user joins the organization automatically, and be redirected to a application to continue their journey.

In Auth0, to create an invitation to an organization, one must select an application.
If the application does not have `initiate_login_uri` configured, the invitation CANNOT be created.
This means FormX must first implement `initiate_login_uri` before they can use the invitation feature.
The details are documented in https://auth0.com/docs/authenticate/login/auth0-universal-login/configure-default-login-routes and https://auth0.com/docs/authenticate/login/auth0-universal-login/configure-default-login-routes#invite-organization-members.

> [!WARNING]
> Implication on Authgear
> This means https://openid.net/specs/openid-connect-core-1_0.html#ThirdPartyInitiatedLogin is a pre-requisite of organization. We must first implement that before we work on supporting organization.

### Usecase 5: Email discovery

FormX has a lot of enterprise customers, not just GreatMall.
FormX wants the sign-in page to show a single email input,
and depending on what the end-user enters, select the organization,
and possibly redirect the end-user to the configured enterprise connection, like Azure Entra ID.

This can be implemented with [Identifier First Authentication](https://auth0.com/docs/authenticate/login/auth0-universal-login/identifier-first) and [Home Realm Discovery](https://auth0.com/docs/authenticate/login/auth0-universal-login/identifier-first#define-home-realm-discovery-identity-providers)

### Usecase 6: Organization switcher

FormX has another enterprise customer called GreatProperty, who is onf of the leading property developer in the city.
GreatProperty has many contractors. GreatProperty and its contractors deal with a lot of textual documents.

An employee of GreatProperty `johndoe@greatproperty.com` of course belong to the organization `greatproperty`.
He also belongs one of the contractor GreatContractor, represented by the organization `greatcontractor`.

In the daily work of `johndoe@greatproperty.com`, he needs to work on both the documents of GreatProperty and GreatContractor.
So he needs to be able to switch between the two organizations.

In Auth0, switching between organizations is not supported natively. During sign-in, the organization is provided via
- With "Login Flow" being "Prompt for Credentials", Auth0 asks the end-user just before the authentication finishes.
- With "Login Flow" being "Prompt for Organization",  the end-user provides it.
- With "Login Flow" being "No prompt", the application developer provide it in the integration code.

In either case, the sign-in is specific to the 1 organization only.

To support organization switcher, the application developer has to implement themselves.

> [!NOTE]
> I didn't dig into how to do this exactly. But I believe this is do-able.

## Caveats

This section documents the known caveats around organization in Auth0.

### Caveat 1: Resolution to ambiguous account

This caveat is caused by an organization can have more than 1 connections, and email addresses are NOT unique across different connections.

It is better to explain this with a concrete example so

- Connection `email-password-1`
  - User `johndoe@example.com`
- Connection `email-password-2`
  - user `johndoe@example.com`
- Org `org1`
  - Connection `email-password-1`
  - Members
    - `johndoe@example.com` from `email-password-1`.
- Org `org2`
  - Connection `email-password-2`
  - Members
    - `johndoe@example.com` from `email-password-2`.
- Application `myspa`
  - "Login Experience" set to `Business Users` with "Login Flow" being `Prompt for Credentials`.

The end-user (owner of the email `johndoe@example.com`) enters his email in the Auth0 sign-in page.
He will either sign in `johndoe@example.com` from `email-password-1` or `johndoe@example.com` from `email-password-2`.
Which one will be used is not documented by Auth0.
But by observation, the resolution is consistent.
In my own testing, Auth0 always resolves to `johndoe@example.com` from `email-password-1`.
So if I enter the password of `johndoe@example.com` from `email-password-2`, Auth0 will tell me the password is incorrect.

To avoid this, it is your own responsibility to make sure the end-users identifier within an organization NEVER overlap.
