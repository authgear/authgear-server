# Account Lockout

## Abstract

In order to increase resistance to brute forcing, we will lock the account for a period of time after a several times of authentication failure of a single account.

## Configuration

Account lockout can be configured as follow:

```yaml
authentication:
  lockout:
    max_attempts: 6 # Maximum attempts to login before the account was locked
    history_duration: "10m" # The duration of login histories participated in  max_attempts
    minimum_duration: "1m" # The initial lockout duration of the account
    maximum_duration: "5m" # The maximum lockout duration of the account after multipled by backoff_factor
    backoff_factor: 2 # The factor to be multiplied to calculate lockout duration in subsequent login failures
```

## Algorithm

We define `lastLoginAt` as the last login attempt timestamp.

When an user failed an login attemmpt:

- If `lastLoginAt < (NOW - history_duration)`:
  - the account is not locked.
  - Set `accumulatedLoginFailure` to 1
- Else:
  - Increment `accumulatedLoginFailure`.
  - If `accumulatedLoginFailure > max_attempts`:
    - Lock the account. The lockout duration is calculated by:
      ```
      min(minimum_duration * backoff_factor^(accumulatedLoginFailure - max_attempts - 1), maximum_duration)
      ```

## References

We designed the feature based on the following references:

- [Azure AD smart lockout](https://learn.microsoft.com/en-us/azure/active-directory/authentication/howto-password-smart-lockout#how-smart-lockout-works)

  - The account locks again after each subsequent failed sign-in attempt, for one minute at first and longer in subsequent attempts. We referenced such behavior in authgear account lockout.

- iOS

  - The device will be locked for 1 minutes after the 6th failed attempts, and the duration is increasing of each of the subsequent failed attempts. The function to calculate the lockout interval is unknown. For simplicity we introduced `backoff_factor` to control the incrementation in authgear.

- [Kratos](https://github.com/ory/kratos/issues/3037)

  - `consecutive_login_interval` defined the duration of login histories participated in the account lockout. `history_duration` in our configuration serves a similar purpose in authgear.
