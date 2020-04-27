package interaction

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

const (
	TemplateItemTypeOOBCodeSMSTXT    config.TemplateItemType = "oob_code_sms.txt"
	TemplateItemTypeOOBCodeEmailTXT  config.TemplateItemType = "oob_code_email.txt"
	TemplateItemTypeOOBCodeEmailHTML config.TemplateItemType = "oob_code_email.html"
)

var TemplateOOBCodeSMSTXT = template.Spec{
	Type: TemplateItemTypeOOBCodeSMSTXT,
	Default: `{{ .code }} is your {{ .appname }} verification code.

Please ignore if you didn't sign in or sign up.

@{{ .host }} #{{ .code }}
`,
}

var TemplateOOBCodeEmailTXT = template.Spec{
	Type: TemplateItemTypeOOBCodeEmailTXT,
	// TODO(interaction): update OOB code email template
	Default: `{{ .code }} is your {{ .appname }} verification code.

Please ignore if you didn't sign in or sign up.
`,
}

var TemplateOOBCodeEmailHTML = template.Spec{
	Type:   TemplateItemTypeOOBCodeEmailHTML,
	IsHTML: true,
	// TODO(interaction): update OOB code email template
	Default: `<!DOCTYPE html>
<html>
<body>
<p>{{ .code }} is your {{ .appname }} verification code.</p>
<p>Please ignore if you didn't sign in or sign up.</p>
</body>
</html>
`,
}
