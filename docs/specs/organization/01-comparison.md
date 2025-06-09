# A brief comparison of organization-related feature support

> [!IMPORTANT]
> FushionAuth is not included in this table because it does not natively support organization at all.

> [!IMPORTANT]
> SuperTokens is not included in this table because it does not natively support organization at all.

| Item                                  | Auth0                                   | Stytch                                                                       | Zitadel                                          | Clerk | Kinde                               |
| ---                                   | ---                                     | ---                                                                          | ---                                              | ---   | ---                                 |
| Organization is optional              | Yes                                     | Authentication Type is determined at project creation. Unmodifiable.         | No                                               | Yes   | Yes                                 |
| Support building GitHub-style service | Yes                                     | Have to emulate with 2 projects                                              | Yes                                              | Yes   | Yes                                 |
| Different password policies           | Yes                                     | Either Cross-organization or Organization-scoped. Once chosen, no going back | Yes                                              | No    | Paid feature                        |
| Different MFA policies                | Use post-login action to customize      | Yes                                                                          | Yes                                              | No    | Paid feature                        |
| IAM                                   | Yes                                     | Yes                                                                          | Yes                                              | No    | No                                  |
| Invitation                            | Yes. Can return to specific application | Only supported at API level. Can return to specific URL to your backend      | Yes when V2 API is enabled. Cannot return to URL | Yes   | No                                  |
| Email discovery                       | Yes                                     | Yes                                                                          | No                                               | No    | Yes                                 |
| Organization switcher                 | No. Session bound to single org         | Provide API for self-implementation                                          | No                                               | Yes   | Provide API for self-implementation |

## Implications

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

### IAM

In all competitors, the IAM use case is trivial to implement.
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
