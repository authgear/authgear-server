package template

var templateVerifyEmailTxt = `Dear {{ login_id }},

You received this email because {{ appname }} would like to verify your email address. If you have recently signed up for this app or if you have recently made changes to your account, click the following link:

{{ link }}

If you are unsure why you received this email, please ignore this email and you do not need to take any action.

Thanks.`
