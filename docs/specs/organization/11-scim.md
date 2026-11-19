In the documentation of SCIM on Auth0, it mentions organization in this section

- https://auth0.com/docs/authenticate/protocols/scim/configure-inbound-scim#organizations

> For SCIM-provisioned users to become members of an Organization,
> the connection must be configured to Enable Auto-Membership as described in Grant Just-In-Time Membership to an Organization Connection.

A quick google performed on 2025-07-10 led me to this question on SCIM and Organization on Auth0

- https://community.auth0.com/t/add-user-on-scim-import/157136

From the answer, SCIM and Organization is not tightly coupled.
If auto-membership does not work for the developer, he has to resort to Auth0 Management API to do whatever he wants.
