# A brief comparison of organization-related feature support

> [!IMPORTANT]
> FushionAuth is not included in this table because it does not natively support organization at all.

> [!IMPORTANT]
> SuperTokens is not included in this table because it does not natively support organization at all.

| Item                                  | Auth0   | Stytch  | Zitadel | Clerk | Kinde   |
| ---                                   | ---     | ---     | ---     | ---   | ---     |
| Organization is optional              | Yes     | Yes[^3] | No      | Yes   | Yes     |
| Support building GitHub-style service | Yes     | Yes[^2] | Yes     | Yes   | Yes     |
| Different password policies           | Yes     | Yes[^4] | Yes     | No    | Yes[^1] |
| Different MFA policies                | Yes[^5] | Yes     | Yes     | No    | Yes[^1] |
| User Isolation by Organization        | Yes     | Yes     | Yes     | No    | No      |
| Invitation to specific application    | Yes     | No      | No      | Yes   | No      |
| Email discovery                       | Yes     | No      | No      | No    | Yes     |
| Organization switcher                 | No      | Yes[^6] | No      | Yes   | Yes[^6] |

> [!WARNING]
> Those "Yes" with footnotes usually mean there is some caveats. Please read the footnotes!

## Observations on each item

### Sign-in session is bound to a single organization

In all competitors, a sign-in session is bound to a single organization only.

In Stytch, organization switching is done with token exchange.

In Clerk, the end-user can just switch organization without signing-in again.

### GitHub-style service

IMO, Auth0 is the most easiest to work with.
It does not enforce that an organization must exist, and does not enforce that users must belong to one and only one organization.
It just models GitHub-Style service naturally.

Clerk also models this easily.

### Different password policies and different MFA policies

In Auth0, password policies to tied to the connection, while MFA is a project-wide setting.
In competitors where organization is mandatory and user belonging to a single organization, all these authentication settings are tied to the organization.

IMO, authentication settings should be organization-overridable.

Notably, Clerk does not support this.

In Kinde, these are paid features.

### User Isolation by Organization

In all competitors, User Isolation by Organization is trivial to implement.
We should consider that in our design.

Notably, Clerk and Kinde do not support this.

### Invitation

It seems that invitation is not very well implemented among competitors.
Auth0 does the best in this area.

Clerk supports this quite well, given that it is not OIDC-based, and it literally just allows you to specify `redirect_uri`.

Kinde does not support this.

### Email discovery

Again Auth0 does the best here.
Its dashboard explains this very well.

In other competitors, you have to look up the lengthy documentation and find no answsers.
At the end you have to try out the example app to test it out yourselves.

Kinde supports this out-of-the-box.

### Organization switcher

In particular, Auth0 does not report to the client application that how organization the user belong to.

In other competitors where a user belongs to one and only one organization, you have to do it yourselves.

Only Clerk supports this out-of-box, but given its lack of support of different password policies and different MFA policies,
it is not very useful.


[^1]: Paid feature.
[^2]: Emulate with 2 projects. 1 Stytch B2B, 1 Stytch Consumer.
[^3]: Actually In Stytch B2B, organization is mandatory. In Stytch Consumer, organization is unsupported.
[^4]: Either cross-organization or organization-scoped. Once chosen, no going back.
[^5]: Use post-login action to customize.
[^6]: Not a builtin feature. Have to use the API to implement yourselves.
