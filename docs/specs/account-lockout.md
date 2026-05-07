# Account Lockout

## Table of Contents

- [Motivation](#motivation)
- [Configuration](#configuration)
- [Algorithm](#algorithm)
  - [Failed attempts](#failed-attempts)
  - [Lockout duration](#lockout-duration)
- [Admin API](#admin-api)
  - [User.accountLockout Query](#useraccountlockout-query)
  - [unlockUser Mutation](#unlockuser-mutation)
- [Usecases](#usecases)
  - [1. Customer Support – Emergency Account Access](#1-customer-support--emergency-account-access-per_user-lockout)
  - [2. Diagnosing Lockout](#2-diagnosing-lockout-per_user_per_ip-lockout)
  - [3. Failed Brute Force Attempt Recovery](#3-failed-brute-force-attempt-recovery-both-lockout-types)
  - [4. Brute Force Attack Investigation](#4-brute-force-attack-investigation-per_user_per_ip-lockout)
  - [5. Account Recovery After Password Reset](#5-account-recovery-after-password-reset-both-lockout-types)
  - [6. Testing & Development](#6-testing--development)
- [References](#references)

## Motivation

Account lockout is a defensive mechanism that:
- **Stops brute force attacks** by temporarily disabling login after repeated failed attempts
- **Allows legitimate users to regain access** after waiting or with admin help

## Configuration

Account lockout can be configured as follow:

```yaml
authentication:
  lockout:
    max_attempts: 6 # Maximum attempts to login before the account was locked
    history_duration: "1h" # The duration of recorded attempts persist after the last update
    minimum_duration: "1m" # The initial lockout duration of the account
    maximum_duration: "5m" # The maximum lockout duration of the account after multipled by backoff_factor
    backoff_factor: 2 # The factor to be multiplied to calculate lockout duration in subsequent login failures
    lockout_type: "per_user" # "per_user" or "per_user_per_ip", read "Algorithm" section for details
    # The followings define which authenticator types participate in account lockout.
    # If enabled is true, fail attempts of that authenticator type will be counted,
    # and contribute to the lockout time.
    password:
      enabled: true
    totp:
      enabled: true
    oob_otp:
      enabled: true
    recovery_code:
      enabled: true
```

## Algorithm

### Failed attempts

We record each failed authentication attempts, with the IP address that made the attempt.

When calculating the total failed attempts:

- If `lockout_type` is `per_user`, attempts from all IP address will be accounted.
- If `lockout_type` is `per_user_per_ip`, only attempts from the actor's IP address will be accounted.

Once the user successfully passed the authentication process using an authenticator participated in lockout, the number of attempts associated with the actor's IP address will be reset to 0.

Here are some examples:

Case 1, assumes:

- `lockout_type` is `per_user`
- only `password` is enabled in lockout
- `max_attempts` is `3`

1. Actor A with IP address 127.0.0.1 has made 2 failed password attempts. Now the total attempt count is 2.
2. Actor B with IP address 127.0.0.2 has made 1 failed password attempts. Now the total attempt count is 3. The user is locked.
3. Actor A passed the authentication process with a correct password after the lockout period. Now the attempts made by Actor A was set to 0. Therefore now the total attempts is 1.
4. Actor B with IP address 127.0.0.2 has made 1 failed password attempts. Now the total attempt count is 2. The user is not locked.

Case 2, assumes:

- `lockout_type` is `per_user_per_ip`
- only `password` is enabled in lockout
- `max_attempts` is `3`

1. Actor A with IP address 127.0.0.1 has made 2 failed password attempts. Now the total attempt count of 127.0.0.1 is 2.
2. Actor B with IP address 127.0.0.2 has made 1 failed password attempts. Now the total attempt count of 127.0.0.2 is 1. The user is not locked.
3. Actor A with IP address 127.0.0.1 has made 1 failed password attempts. Now the total attempt count of 127.0.0.1 is 3. The user is locked in the perspective of A.
4. Actor B with IP address 127.0.0.2 has made 1 failed password attempts. Now the total attempt count of 127.0.0.2 is 2. The user is not locked in the perspective of B.
5. Actor B with IP address 127.0.0.2 has made 1 failed password attempts. Now the total attempt count of 127.0.0.2 is 3. The user is locked in the perspective of B.
6. Actor A passed the authentication process with a correct password after the lockout period. Now the attempts made by Actor A was set to 0. However, the user is still locked in the perspective of B.
7. Actor B with IP address 127.0.0.2 has made 1 failed password attempts after the lockout period. Now the total attempt count of 127.0.0.2 is 4. The user is locked again and the lockout duraction will be increased according to `backoff_factor`.

### Lockout duration

The locking duration is calculated by the following equation:

```
min(minimum_duration * backoff_factor^(failedAttemps - max_attempts), maximum_duration)
```

Once the user was locked, the locking period will not be updated. Any attempts, no matter success or failures, within the locked period will not affect the lock. The server will reject any user input in the authentication process during the locked period.

## Admin API

### User.accountLockout Query

Tenant admins can query the current lockout state of a user via the Admin GraphQL API:

```graphql
query {
  node(id: "<USER_NODE_ID>") {
    ... on User {
      id
      accountLockout {
        lockoutType: "per_user" | "per_user_per_ip"
        isLocked: Boolean!
        lockedUntil: DateTime
        lockedIPs: [String!]
      }
    }
  }
}
```

**For `per_user` lockout type:**
- `lockoutType`: "per_user"
- `isLocked`: true if user is locked globally, false otherwise
- `lockedUntil`: DateTime in UTC when the global lock expires (null if not locked)
- `lockedIPs`: Empty array (not applicable for this type)

**For `per_user_per_ip` lockout type:**
- `lockoutType`: "per_user_per_ip"
- `isLocked`: true if user is locked from ANY IP, false if no IPs have locks
- `lockedUntil`: null (not applicable; different IPs have different expiration times)
- `lockedIPs`: Array of IP addresses currently locked (e.g., ["192.168.1.1", "203.0.113.45"]).

### unlockUser Mutation

Tenant admins can unlock a user by clearing all lockout state:

```graphql
mutation {
  unlockUser(input: { userID: "<USER_NODE_ID>" }) {
    user {
      id
      accountLockout {
        isLocked
      }
    }
  }
}
```

Behavior:
- Clears lockout state for all authenticator types (password, TOTP, OOB OTP, recovery code)
- **For `per_user` lockout type**: Clears the global per-user lock
- **For `per_user_per_ip` lockout type**: Clears all IP-specific locks for this user across all IPs
- User can immediately authenticate again from any IP without waiting for the lockout period to elapse
- If lockout is not configured or enabled, the mutation succeeds with no effect

## Usecases

### 1. Customer Support – Emergency Account Access (per_user lockout)

**Scenario**: Using `per_user` lockout. A legitimate user reports they are locked out after entering wrong credentials multiple times. They have been locked for 30+ minutes but need immediate access (e.g., to complete a critical transaction).

**Solution**:
- Support team queries `User.accountLockout` to check:
  - `isLocked: true` → confirms account is locked
  - `lockedUntil: 2024-05-07T14:30:00Z` → when access restores automatically
- Calls `unlockUser` to immediately restore access without waiting for the automatic unlock period

### 2. Diagnosing Lockout (per_user_per_ip lockout)

**Scenario**: Using `per_user_per_ip` lockout. User reports authentication failures. Admin needs to determine if the issue is lockout-related and from which IP(s).

**Solution**:
- Support queries `User.accountLockout` to see:
  - `isLocked: true` → user is locked from some IP(s)
  - `lockedIPs: ["192.168.1.100", "203.0.113.45"]` → exactly which IPs have locks
- Now admin can inform user: "Your account is locked from IPs X and Y. Try logging in from a different IP, or we can unlock it immediately."

### 3. Failed Brute Force Attempt Recovery (both lockout types)

**Scenario**: An attacker attempts a brute force attack against a user account, triggering lockout after many failed attempts. The legitimate user then tries to log in and is blocked by the lockout.

**Solution**:
- Admin queries `User.accountLockout` to confirm lockout state
- For per_user_per_ip: `lockedIPs` shows which attacker IPs triggered the locks
- Calls `unlockUser` to clear all lockout state, allowing the legitimate user to log in immediately

### 4. Brute Force Attack Investigation (per_user_per_ip lockout)

**Scenario**: Using `per_user_per_ip` lockout. Admin detects a brute force attack and wants to understand which IPs are attacking which users.

**Solution**:
- Admin queries multiple affected users' `accountLockout.lockedIPs` fields
- Sees patterns like: "User A locked from IPs 1,2,3; User B locked from IPs 2,3,4"
- Identifies the attacking IP range (2,3,4)
- Decides whether to unlock users and monitor, or allow lockout to expire naturally

### 5. Account Recovery After Password Reset (both lockout types)

**Scenario**: User forgets password and resets it through email verification. However, the repeated failed login attempts before the reset still have the account locked. User tries to log in with the new password but is still locked out.

**Solution**:
- Admin queries `User.accountLockout` to confirm the account is locked due to previous attempts
- Calls `unlockUser` to clear the old lockout state
- User can now immediately log in with their new password (no waiting required)
- Improves user experience by allowing immediate access after password reset

### 6. Testing & Development

**Scenario**: During development or testing, engineers need to simulate lockout scenarios and then reset account state without waiting for natural timeout.

**Solution**: Test automation calls `unlockUser` to reset account state between test runs, enabling faster test cycles (works for both lockout types).

## References

We designed the feature based on the following references:

- [Azure AD smart lockout](https://learn.microsoft.com/en-us/azure/active-directory/authentication/howto-password-smart-lockout#how-smart-lockout-works)

  - The account locks again after each subsequent failed sign-in attempt, for one minute at first and longer in subsequent attempts. We referenced such behavior in authgear account lockout.

- iOS

  - The device will be locked for 1 minutes after the 6th failed attempts, and the duration is increasing of each of the subsequent failed attempts. The function to calculate the lockout interval is unknown. For simplicity we introduced `backoff_factor` to control the incrementation in authgear.

- [Kratos](https://github.com/ory/kratos/issues/3037)

  - `consecutive_login_interval` defined the duration of login histories participated in the account lockout. `history_duration` in our configuration serves a similar purpose in authgear.
