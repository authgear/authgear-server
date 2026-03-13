# Zitadel

## Concepts

### Team

- When you first sign up at zitadel.com, you end up creating a new team.

### Instance

- An instance belongs to one and only one team.
- An instance must have at least one organization.
- When you create an instance, you must create an admin user in the instance.
  This implies the admin users and the end-users are managed within the same instance.

### Organization

- An organization belongs to one and only one instance.
- An organization contains all kinds of settings, like authentication methods, MFA, branding, etc.
- An organization has its own domain.

### Project

- A project belongs to one and only one organization.
- A project is a collection of applications.
- A project CAN BE granted to another organization to grant access.
- When you first create an instance, a project named ZITADEL is automatically added to your instance.
  The ZITADEL project contains applications
  - Management-API
  - Admin-API
  - Auth-API
  - Console
- You are strongly suggested NOT to modify the ZITADEL project.
- You have to create a new project to start integration.

### User

- A user belongs to one and only one organization.
- A user must have a loginname.
- The loginname must be unique within the instance, or in other words, across all organizations.
  See https://zitadel.com/docs/concepts/structure/users#uniqueness-of-users
- Email address != loginname, email address need not be unique.

https://zitadel.com/docs/guides/solution-scenarios/introduction
https://zitadel.com/docs/guides/solution-scenarios/b2c
https://zitadel.com/docs/guides/solution-scenarios/b2b
https://zitadel.com/docs/guides/solution-scenarios/saas
https://zitadel.com/docs/guides/solution-scenarios/domain-discovery
https://zitadel.com/docs/guides/manage/console/organizations#default-organization
https://zitadel.com/docs/guides/solution-scenarios/domain-discovery#enable-domain-discovery
https://zitadel.com/docs/guides/integrate/onboarding/b2b#invite-team-members (This is no invite, it is just create user)

## How to do usecase X

### Usecase 1: Serving both casual end-users and enterprise end-users

- Create an organization for casual end-users.
  You have to do this anyway because Zitadel force you to have at least one organization in an instance.
- Create as many organizations as you need to represent an enterprise.

### Usecase 2: Different password policies

This is trivial in Zitadel because each organization has its own settings.

### Usecase 3: User Isolation by Organization

This is trivial in Zitadel because each user belongs to one and only one organization.

### Usecase 4: Enable MFA for some organization

This is trivial in Zitadel because each organization has its own settings.

### Usecase 5: Invitation

> [!WARNING]
> Invitation is supported in the console if you have enabled "Use V2 Api in Console for User creation" in Default settings.
> In my experience, even the instance is newly created, the flag is still off by default.

With the V2 API, invitation is possible.
See https://zitadel.com/docs/guides/manage/console/users#create-user

However, sign in to a particular application is not supported.

### Usecase 6: Email discovery

I follow the official guide https://zitadel.com/docs/guides/solution-scenarios/domain-discovery
and set up this scenario

- Create an organization `louischantest`
- Create an organization `louischantest1`
- Create a user in `louischantest` with email `louischan@oursky.com`
- Create a user in `louischantest1` with email `louischan@oursky.com`

And then I try to sign in with the example app.
Zitadel returns the error `Multiple users found`.

### Usecase 7: Organization switcher

Not supported at all.
