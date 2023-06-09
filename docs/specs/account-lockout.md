# Account Lockout

## Abstract

In order to increase resistance to brute forcing, we will lock the account for a period of time after a several times of authentication failure of a single account.

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

## References

We designed the feature based on the following references:

- [Azure AD smart lockout](https://learn.microsoft.com/en-us/azure/active-directory/authentication/howto-password-smart-lockout#how-smart-lockout-works)

  - The account locks again after each subsequent failed sign-in attempt, for one minute at first and longer in subsequent attempts. We referenced such behavior in authgear account lockout.

- iOS

  - The device will be locked for 1 minutes after the 6th failed attempts, and the duration is increasing of each of the subsequent failed attempts. The function to calculate the lockout interval is unknown. For simplicity we introduced `backoff_factor` to control the incrementation in authgear.

- [Kratos](https://github.com/ory/kratos/issues/3037)

  - `consecutive_login_interval` defined the duration of login histories participated in the account lockout. `history_duration` in our configuration serves a similar purpose in authgear.
