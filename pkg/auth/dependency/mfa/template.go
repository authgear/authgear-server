package mfa

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

const (
	TemplateItemTypeMFAOOBCodeSMSTXT    config.TemplateItemType = "mfa_oob_code_sms.txt"
	TemplateItemTypeMFAOOBCodeEmailTXT  config.TemplateItemType = "mfa_oob_code_email.txt"
	TemplateItemTypeMFAOOBCodeEmailHTML config.TemplateItemType = "mfa_oob_code_email.html"
)

var TemplateMFAOOBCodeSMSTXT = template.Spec{
	Type: TemplateItemTypeMFAOOBCodeSMSTXT,
	Default: `Your {{ .appname }} Two Factor Auth Verification code is: {{ .code }}

Please enter the Verification Code on the Sign-in screen.

Please ignore this code if this Sign-in was not initiated by you.
`,
}

var TemplateMFAOOBCodeEmailTXT = template.Spec{
	Type: TemplateItemTypeMFAOOBCodeEmailTXT,
	Default: `Your {{ .appname }} Two Factor Auth Verification code is: {{ .code }}

Please enter the Verification Code on the Sign-in screen.

Please ignore this code if this Sign-in was not initiated by you.
`,
}

var TemplateMFAOOBCodeEmailHTML = template.Spec{
	Type:   TemplateItemTypeMFAOOBCodeEmailHTML,
	IsHTML: true,
	Default: `<!DOCTYPE html>
<html>
<body>
<p>Your {{ .appname }} Two Factor Auth Verification code is: {{ .code }}</p>
<p>Please enter the Verification Code on the Sign-in screen.</p>
<p>Please ignore this code if this Sign-in was not initiated by you.</p>
</body>
</html>
`,
}
