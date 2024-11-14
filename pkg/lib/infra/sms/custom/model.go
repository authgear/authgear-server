package custom

import (
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsapi"
)

type SendOptions struct {
	To                string                    `json:"to"`
	Body              string                    `json:"body"`
	AppID             string                    `json:"app_id"`
	TemplateName      string                    `json:"template_name"`
	LanguageTag       string                    `json:"language_tag"`
	TemplateVariables *smsapi.TemplateVariables `json:"template_variables"`
}
