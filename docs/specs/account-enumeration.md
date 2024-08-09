- [Account Enumeration](#account-enumeration)
  * [The goals of Account Enumeration Prevention](#the-goals-of-account-enumeration-prevention)
  * [Configuration](#configuration)
    + [Addition 1: Add `account_enumeration_prevention` to SignupFlow and LoginFlow, and `type: check_duplicate_account` , `type: check_if_account_exists`](#addition-1-add-account_enumeration_prevention-to-signupflow-and-loginflow-and-type-check_duplicate_account--type-check_if_account_exists)
      - [Signup flow behavior](#signup-flow-behavior)
      - [Login flow behavior](#login-flow-behavior)
    + [Addition 2: Add `authentication.account_enumeration_prevention`](#addition-2-add-authenticationaccount_enumeration_prevention)
  * [Reference](#reference)
    + [Auth0](#auth0)
    + [Okta](#okta)
    + [Keycloak](#keycloak)
    + [Amazon Cognito](#amazon-cognito)
    + [Stytch](#stytch)
    + [FusionAuth](#fusionauth)
    + [AWS console](#aws-console)

# Account Enumeration

## The goals of Account Enumeration Prevention

OWASP has published [an article](https://owasp.org/www-project-web-security-testing-guide/latest/4-Web_Application_Security_Testing/03-Identity_Management_Testing/04-Testing_for_Account_Enumeration_and_Guessable_User_Account) on this subject.

To prevent account enumeration, we have to

1. Report an identical error message when the user is non-existent, or the password is incorrect.
2. Report a message that an email, a SMS, or a Whatsapp message has been sent, regardless of whether the account exists.
3. Minimize the difference in the processing time between a non-existent account and an existing account.

**At the current stage, we address the first 2 goals only.**

## Configuration

### Addition 1: Add `account_enumeration_prevention` to SignupFlow and LoginFlow, and `type: check_duplicate_account` , `type: check_if_account_exists`

Example
```yaml
signup_flows:
- name: default
  account_enumeration_prevention:
    enabled: true
  steps:
  - type: identify
    name: identify
    one_of:
    - identification: email
  - type: verify
    target_step: identify
  - type: check_duplicate_account # This step will report error
  - type: create_authenticator
    one_of:
    - authentication: primary_password
  - type: create_authenticator
    one_of:
    - authentication: secondary_totp

login_flows:
- name: default
  account_enumeration_prevention:
    enabled: true
  steps:
  - type: identify
    name: identify
    one_of:
    - identification: email
  - type: authenticate
    one_of:
    - authentication: primary_password
  - type: authenticate
    one_of:
    - authentication: secondary_totp
  - type: check_if_account_exists # This step will report error
  - type: check_account_status
```

#### Signup flow behavior
When a signup flow has account enumeration prevention enabled,
1. `type: identify` does not report duplication error. It will proceed to the next step.
2. `type: verify` will still send the verification code. It will proceed to the next step if the code is correct.
3. `type: create_authenticator` will be skipped if duplication error was detected.
4. `type: fill_in_user_profile` will be skipped if duplication error was detected.
5. `type: view_recovery_code` will be skipped if duplication error was detected.
6. `type: prompt_create_passkey` will be skipped if duplication error was detected.
7. `type: check_duplicate_account` is the new step. It will report invalid credentials error if duplication error was detected.

It is **RECOMMENDED** that `type: check_duplicate_account` is placed before the steps that will be skipped.

#### Login flow behavior
When a login flow has account enumeration prevention enabled,
1. `type: identify` does not report user not found error. It will proceed to the next step.
2. `type: authenticate` will still send OTP or verify password. For OTP, it will proceed to the next step if the OTP is correct. For password, it will always proceed.
3. `type: check_account_status` will be skipped if user not found.
4. `type: terminate_other_sessions` will be skipped if user not found.
5. `type: change_password` will be skipped if user not found.
6. `type: prompt_create_passkey` will be skipped if user not found.
7. `type: check_if_account_exists` is the new step. It will report invalid credentials error if user not found error was detected.

It is **RECOMMENDED** that `type: check_if_account_exists` is placed before the steps that will be skipped.

### Addition 2: Add `authentication.account_enumeration_prevention`

```yaml
authentication:
  account_enumeration_prevention:
    enabled: true
```

When `authentication.account_enumeration_prevention.enabled=true`, the default signup flow and the default login flow will have `account_enumeration_prevention.enabled=true`, and `type: check_duplicate_account` and `type: check_if_account_exists` will be inserted before the skippable steps.

## Reference

### Auth0

With email + password on the same page, so yes.

With email + passwordless + Universal Login, the flow becomes signup/login flow.
Even I have configured "Disable signup" in my database connection, the copywriting in the page still show "Enter your email to sign in or create account". So it does not support account enumeration prevention in this case.

In summary, it does not have support account enumeration prevention in all cases.

### Okta

Okta has account enumeration prevention. With it is enabled, Okta will not report whether the account exists or not, it will show authenticator verification error instead.
However, when MFA is enabled, the doc does not say whether the authenticator verification error happens in the first factor, or in the second factor.

https://support.okta.com/help/s/article/How-Does-User-Enumeration-Prevention-Work?language=en_US

### Keycloak

Keycloak does not support account enumeration prevention at the moment. https://github.com/keycloak/keycloak/issues/17629

### Amazon Cognito

Amazon Cognito supports account enumeration prevention in login. https://docs.aws.amazon.com/cognito/latest/developerguide/cognito-user-pool-managing-errors.html

Interestingly, Amazon Cognito also supports account enumeration prevention in signup. For this feature to work, the first step of sign up must be email verification of phone number verification. The code is delivered to the recipient. If the end-user does not have the ownership of the email or the phone number, the flow stops there. If the end-user does have access to the email or phone number, Amazon Cognito reports an error when the end-user provides a correct code. https://docs.aws.amazon.com/cognito/latest/developerguide/cognito-user-pool-managing-errors.html#cognito-user-pool-managing-errors-prevent-userexistence-errors

### Stytch

With username + password on the same page, so yes.

With email + passwordless, Stytch requires switching to a signup/login flow. It does not support account enumeration prevention like Amazon Cognito. https://stytch.com/docs/b2b/resources/platform/account-enumeration

### FusionAuth

With email + password on the same page, so yes.
In signup, it reports email registered error, so no.

### AWS console

When the root account has enabled MFA

1. Correct username, incorrect password => Prompt for MFA TOTP code, always fail at the end.
2. Correct username, correct password => Prompt for MFA TOTP code.
3. Incorrect username => Show invalid username immediately.
