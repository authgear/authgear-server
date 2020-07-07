# UI

The UI creates new users and authenticates existing ones. The user can manage their account in the settings page.

  * [Theming](#theming)
  * [The phone input widget](#the-phone-input-widget)
  * [frame-ancestors](#frame-ancestors)
  * [The login page](#the-login-page)
  * [The enter password page](#the-enter-password-page)
  * [The signup page](#the-signup-page)
  * [The create password page](#the-create-password-page)
  * [The forgot password page](#the-forgot-password-page)
  * [The reset password page](#the-reset-password-page)
  * [The OOB OTP page](#the-oob-otp-page)
  * [The identity page](#the-identity-page)

## Theming

The developer can provide a CSS stylesheet to customize the theme of the UI. Or they override the templates with their own ones.

## The phone input widget

The developer can customize the list of country calling code and the default country calling code of the phone input widget via configuration. By default the list includes all country calling codes globally. The default value is the first one in the list.

## frame-ancestors

The `frame-ancestors` directive of the HTTP header `Content-Security-Policy:` is derived from the `redirect_uris` of all clients. If the `redirect_uri` is of scheme `https`, the host is added to to frame-ancestors. If the `redirect_uri` is `http` and the host is loopback address or the domain ends with `.localhost`, the host is also added to frame-ancestors.

## The login page

The login page authenticates the user. It lists out the configured IdPs. It shows a text field for login ID. The login ID field is either a plain text input or a phone number input, depending on the type of the first login ID key. Link to the forgot password page is shown if Password Authenticator is enabled.

```
|---------------------------|
| Login with Google         |
|---------------------------|
| Login with Facebook       |
|---------------------------|

              Or

|--------------------------------------|  |----------|
| Enter in your email or username here |  | Continue |
|--------------------------------------|  |----------|

Login with a phone number instead.
Forgot password?
```

## The enter password page

The enter password page displays a visibility toggleable password field.

```
|-----------------------|  |----------|
| Enter a password here |  | Continue |
|-----------------------|  |----------|
```

## The signup page

The signup page creates new user. It looks like the login page. It displays the first login ID key by default. Other login ID keys are available to choose from.

```
Sign up with email
|--------------------------|  |----------|
| Enter in your email here |  | Continue |
|--------------------------|  |----------|

Sign up with phone instead.
Sign up with username instead.
```

## The create password page

The create password page displays a visibility toggleable password field with password requirements.

```
|------------------------|  |----------|
| Create a password here |  | Continue |
|------------------------|  |----------|

- [ ] At least one digit
- [ ] At least one uppercase English character
- [ ] At least one lowercase English character
- [ ] At least one symbols ~`!@#$%^&*()-_=+[{]}\|;:'",<.>/?
- [ ] At least 8 characters long
```

## The forgot password page

The forgot password page displays an email text field. When the user enter a valid Email Login ID, a reset password link to sent to that email address.

## The reset password page

The reset password page looks like the create password page.

## The OOB OTP page

The OOB OTP page lets the user to input OOB OTP. A resend button with cooldown is shown as well.

## The identity page

The identity page lists out the candidates of identity and the status.

```
|---------------------------------------|
| Google                        Connect |
|---------------------------------------|
| Email                                 |
| user@example.com               Change |
|---------------------------------------|
| Phone                             Add |
|---------------------------------------|
```
