package custom

import (
	"encoding/json"

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

// See docs/specs/sms_gateway.md
type ResponseBody struct {
	Code        string                 `json:"code"`
	Description string                 `json:"description,omitempty"`
	Info        map[string]interface{} `json:"info,omitempty"`
}

func ParseResponseBody(jsonData []byte) (*ResponseBody, error) {
	var response ResponseBody
	err := json.Unmarshal(jsonData, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
