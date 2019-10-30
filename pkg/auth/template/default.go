package template

/* #nosec */
const DefaultForgotPasswordEmailTXT = `Dear {{ .email }},

You received this email because someone tries to reset your account password on {{ .appname }}. To reset your account password, click this link:

{{ .link }}

If you did not request to reset your account password, Please ignore this email.

Thanks.`

/* #nosec */
const DefaultForgotPasswordResetHTML = `<!DOCTYPE html>

<head>
  <meta name="viewport" content="width=device-width, initial-scale=1">
</head>
{{ if .error }}
<p>{{ .error.Message }}</p>
{{ end }}
<form method="POST" action="{{ .action_url }}">
  <label for="password">New Password</label>
  <input type="password" name="password"><br>
  <label for="confirm">Confirm Password</label>
  <input type="password" name="confirm"><br>
  <input type="hidden" name="code" value="{{ .code }}">
  <input type="hidden" name="user_id" value="{{ .user_id }}">
  <input type="hidden" name="expire_at" value="{{ .expire_at }}">
  <input type="submit" value="Submit">
</form>`

/* #nosec */
const DefaultForgotPasswordSuccessHTML = `<!DOCTYPE html>

<head>
  <meta name="viewport" content="width=device-width, initial-scale=1">
</head>
<p>Your password is reset successfully.</p>`

/* #nosec */
const DefaultWelcomeEmailTXT = `Hello {{ .email }},

Welcome to Skygear.

Thanks.`

/* #nosec */
const DefaultUserVerificationSMSTXT = `Your {{ .appname }} Verification Code is: {{ .code }}`

/* #nosec */
const DefaultUserVerificationEmailTXT = `Dear {{ .login_id }},

You received this email because {{ .appname }} would like to verify your email address. If you have recently signed up for this app or if you have recently made changes to your account, click the following link:

{{ .link }}

If you are unsure why you received this email, please ignore this email and you do not need to take any action.

Thanks.`

/* #nosec */
const DefaultUserVerificationSuccessHTML = `<!DOCTYPE html>
<head>
  <meta name="viewport" content="width=device-width, initial-scale=1">
</head>
<p>Your account information is verified successfully.</p>`

/* #nosec */
const DefaultMFAOOBCodeSMSTXT = `Your MFA code is: {{ .code }}`

/* #nosec */
const DefaultMFAOOBCodeEmailTXT = `Your MFA code is: {{ .code }}`

/* #nosec */
const DefaultErrorHTML = `<!DOCTYPE html>
<head>
  <meta name="viewport" content="width=device-width, initial-scale=1">
</head>
<p>{{ .error.Message }}</p>`
