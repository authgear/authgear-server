# Delete User

  * [Deleting user](#deleting-user)
  * [Scheduled deletion](#scheduled-deletion)
  * [Cached data](#cached-data)
  * [Future works](#future-works)

## Deleting user

A deleted user would have all its associated data (exposed from API) erased
from database (hard-delete), including identities, authenticators, password data, sessions, etc.

### Deleting user via Admin API

The developer can delete user via the Admin API.

### Deleting user on the portal

The admin can delete user on the portal.

## Scheduled deletion

Instead of deleting a user immediately, a deletion can be scheduled.

The schedule is measured in terms of days. The default value is 30 days. Valid values are [1, 180].

When deletion is scheduled by admin API or portal admin, the user is **disabled**.
To cancel the scheduled deletion, re-enable the user.

When deletion is scheduled by the end-user, the user is **deactivated**.
To cancel the scheduled deletion, the end-user has to reactivate their account.
The end-user cannot reactivate their account by themselves at the moment.
They have to contact support.
It is possible to cancel the scheduled deletion on behalf of the end-user by re-enabling the user.

Behind the scene, re-enable or reactivate user always remove scheduled deletion, if any.

## Cached data

Some internal data may be present in cache (Redis), such as OAuth states,
MFA device tokens, rate limit counter. These data would remain in the cache
until its natural expiry.

## Future works

- Web-hooks
- Reserve deleted login ID
- Soft-delete
