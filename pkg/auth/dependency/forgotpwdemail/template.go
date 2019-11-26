package forgotpwdemail

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

const (
	TemplateItemTypeForgotPasswordEmailTXT    config.TemplateItemType = "forgot_password_email.txt"
	TemplateItemTypeForgotPasswordEmailHTML   config.TemplateItemType = "forgot_password_email.html"
	TemplateItemTypeForgotPasswordResetHTML   config.TemplateItemType = "forgot_password_reset.html"
	TemplateItemTypeForgotPasswordSuccessHTML config.TemplateItemType = "forgot_password_success.html"
	TemplateItemTypeForgotPasswordErrorHTML   config.TemplateItemType = "forgot_password_error.html"
)

var TemplatePasswordEmailTXT = template.Spec{
	Type: TemplateItemTypeForgotPasswordEmailTXT,
	Default: `Dear {{ .email }},

You received this email because someone tries to reset your account password on {{ .appname }}. To reset your account password, click this link:

{{ .link }}

If you did not request to reset your account password, Please ignore this email.

Thanks.`,
}

var TemplatePasswordEmailHTML = template.Spec{
	Type:   TemplateItemTypeForgotPasswordEmailHTML,
	IsHTML: true,
}

var TemplateForgotPasswordResetHTML = template.Spec{
	Type:   TemplateItemTypeForgotPasswordResetHTML,
	IsHTML: true,
	Default: `<!DOCTYPE html>
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
</form>`,
}

var TemplateForgotPasswordSuccessHTML = template.Spec{
	Type:   TemplateItemTypeForgotPasswordSuccessHTML,
	IsHTML: true,
	Default: `<!DOCTYPE html>
<head>
<meta name="viewport" content="width=device-width, initial-scale=1">
</head>
<p>Your password is reset successfully.</p>`,
}

var TemplateForgotPasswordErrorHTML = template.Spec{
	Type:   TemplateItemTypeForgotPasswordErrorHTML,
	IsHTML: true,
	Default: `<!DOCTYPE html>
<head>
<meta name="viewport" content="width=device-width, initial-scale=1">
</head>
<p>{{ .error.Message }}</p>`,
}
