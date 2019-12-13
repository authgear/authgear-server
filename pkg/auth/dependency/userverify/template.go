package userverify

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

const (
	TemplateItemTypeUserVerificationSMSTXT      config.TemplateItemType = "user_verification_sms.txt"
	TemplateItemTypeUserVerificationEmailTXT    config.TemplateItemType = "user_verification_email.txt"
	TemplateItemTypeUserVerificationEmailHTML   config.TemplateItemType = "user_verification_email.html"
	TemplateItemTypeUserVerificationSuccessHTML config.TemplateItemType = "user_verification_success.html"
	TemplateItemTypeUserVerificationErrorHTML   config.TemplateItemType = "user_verification_error.html"
)

var TemplateUserVerificationSMSTXT = template.Spec{
	Type:    TemplateItemTypeUserVerificationSMSTXT,
	IsKeyed: true,
	Default: `Your {{ .appname }} Verification Code is: {{ .code }}`,
}

var TemplateUserVerificationEmailTXT = template.Spec{
	Type:    TemplateItemTypeUserVerificationEmailTXT,
	IsKeyed: true,
	Default: `Dear {{ .login_id }},

You received this email because {{ .appname }} would like to verify your email address. If you have recently signed up for this app or if you have recently made changes to your account, click the following link:

{{ .link }}

If you are unsure why you received this email, please ignore this email and you do not need to take any action.

Thanks.`,
}
var TemplateUserVerificationEmailHTML = template.Spec{
	Type:    TemplateItemTypeUserVerificationEmailHTML,
	IsKeyed: true,
	IsHTML:  true,
	Default: `<!DOCTYPE html>
<html>
<body>
<p>Dear {{ .login_id }},</p>
<p>You received this email because {{ .appname }} would like to verify your email address. If you have recently signed up for this app or if you have recently made changes to your account, click the following link:</p>
<p><a href="{{ .link }}">{{ .link }}</a></p>
<p>If you are unsure why you received this email, please ignore this email and you do not need to take any action.</p>
<p>Thanks.</p>
</body>
</html>
`,
}

var TemplateUserVerificationSuccessHTML = template.Spec{
	Type:    TemplateItemTypeUserVerificationSuccessHTML,
	IsKeyed: true,
	IsHTML:  true,
	Default: `<!DOCTYPE html>
<head>
  <meta name="viewport" content="width=device-width, initial-scale=1">
</head>
<p>Your account information is verified successfully.</p>`,
}

var TemplateUserVerificationErrorHTML = template.Spec{
	Type:    TemplateItemTypeUserVerificationErrorHTML,
	IsKeyed: true,
	IsHTML:  true,
	Default: `<!DOCTYPE html>
<head>
  <meta name="viewport" content="width=device-width, initial-scale=1">
</head>
<p>{{ .error.Message }}</p>`,
}
