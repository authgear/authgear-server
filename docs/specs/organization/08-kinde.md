# Kinde

## Concepts

### Business

- A business is the top level of everything.
- A business contains multiple environments.

### Environment

- An environment contains authentication settings, like login methods, password policies, and MFA policies.
- An environment contains organizations, users, and applications.

### Organization

- An organization can override authentication settings. But once you do that, nothing is inherited from the environment.
  See https://docs.kinde.com/authenticate/manage-authentication/organization-auth-experience/ This is a paid feature.
- The ID token can be configured to exposes what organizations the user belongs to. See https://docs.kinde.com/build/organizations/orgs-for-developers/#using-the-id_token-to-get-a-list-of-organization-ids and https://docs.kinde.com/authenticate/manage-authentication/navigate-between-organizations/

### User

- Roles and permissions are set on the level of organization.
  This implies a user must belong to an organization before they can have roles and permissions assigned.
- Email addresses are globally unique.

### Application

- An application is a OAuth 2.0 client application.

## How to do usecase X

### Usecase 1: Different password policies

By using the paid feature https://docs.kinde.com/authenticate/manage-authentication/organization-auth-experience/

### Usecase 2: User Isolation by Organization

Not supported. Email addresses are globally unique.

### Usecase 3: Enable MFA for some organization

By using the paid feature https://docs.kinde.com/authenticate/manage-authentication/organization-auth-experience/

### Usecase 4: Invitation

It is not supported. You can only create the user immediately and assign it to an organization.
See https://docs.kinde.com/build/organizations/allow-user-signup-org/

### Usecase 5: Email discovery

This is supported out-of-the-box. See https://docs.kinde.com/build/organizations/orgs-for-developers/#signing-users-into-an-existing-organization

### Usecase 6: Organization switcher

See https://docs.kinde.com/authenticate/manage-authentication/navigate-between-organizations/#step-2-build-the-switcher

The ID token can be configured to exposes what organizations the user belongs to.
It is an application-specific settings.

Once turned on, the ID token contain the necessary information for you to build the switcher yourself.

Note that it is still necessary to sign in again to switch organization.
