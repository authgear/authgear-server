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

var TemplateMFAOOBCodeSMSTXT = template.T{
	Type:    TemplateItemTypeMFAOOBCodeSMSTXT,
	Default: `Your MFA code is: {{ .code }}`,
}

var TemplateMFAOOBCodeEmailTXT = template.T{
	Type:    TemplateItemTypeMFAOOBCodeEmailTXT,
	Default: `Your MFA code is: {{ .code }}`,
}
