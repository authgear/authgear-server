package template

const templateForgotPasswordEmailTxt = `Dear {{ user.email }},

You received this email because someone tries to reset your account password on {{ appname }}. To reset your account password, click this link:

{{ link }}

If you did not request to reset your account password, Please ignore this email.

Thanks.`
