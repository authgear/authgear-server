# A brief research on how Organization works in SuperTokens

As of 2025-06, SuperTokens does not natively support organization.
The official documentation suggests you model organization with multi-tenancy feature.
See https://supertokens.com/docs/authentication/enterprise/introduction

Due to the nature of multi-tenancy, the following use cases are supported out-of-the box

- Different password policies
- User Isolation by Organization
- Different MFA policies

Advanced use cases like the following have to be built-by-yourselves.

- Invitation
- Email discovery
- Organization switcher
