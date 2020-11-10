# Delete User

  * [Deleting user](#deleting-user)
  * [Cached data](#cached-data)
  * [Future works](#future-works)

## Deleting user

A user can be deleted by admins. A deleted user would have all its associated
data (exposed from API) erased from database (hard-delete), including
identities, authenticators, password data, sessions, etc. 

## Cached data

Some internal data may be present in cache (Redis), such as OAuth states,
MFA device tokens, rate limit counter. These data would remain in the cache
until its natural expiry.

## Future works

- Web-hooks
- Reserve deleted login ID
- Self-service deletion
