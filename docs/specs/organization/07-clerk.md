# A brief research on how Organization works in Clerk

This document is valid as of 2025-06-09.

## Concepts

Clerk does not build on top of OIDC, so it does not have things like OIDC client applications.

### Application

- An application has its own domain.

### Organization

- Organization is off by default.
- Once it is turned on, every user is allowed to create organization by default.
- Organization does not have specific settings, e.g. password policy.

### User

- A user can belong to no organization.
- A user can belong to multiple organization.

## How to do usecase X

### Usecase 1: Different password policies

Not possible. It is because organization does not have specific settings.
At least I did not see it in the dashboard.
Checking the API documentation does not reveal something similar.
See https://clerk.com/docs/references/backend/organization/update-organization

### Usecase 2: User Isolation by Organization

Not possible. It is because email address has to be globally unique.
It is not possible to create an organization user that shares the same address with a user in the global user pool.

### Usecase 3: Different MFA policies

Not possible. The reason is the same as that of different password policies.

### Usecase 4: Invitation

See https://clerk.com/docs/organizations/invitations

The invitation must target an email address.
The invitation can optionally take `redirect_uri` which allows to redirect the user.

### Usecase 5: Email discovery

The prebuilt <SignIn> component does not offer a way to select organization on sign in.
The active organization is taken from the last active one.
The last active one is the one selected by the end-user in the <OrganizationSwitcher> component.

In this sense, it does not really support the typical email discovery.

### Usecase 6: Organization Switcher

It supports a prebuilt <OrganizationSwitcher> component in the NextJS SDK.

However, due to the fact that it does not support organization specific settings,
switching organization is just about switching the last active org ID.
It does not do things like authenticate again if the specific organization requires MFA.
