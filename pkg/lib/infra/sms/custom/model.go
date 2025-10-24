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

// See https://github.com/authgear/authgear-sms-gateway/blob/main/pkg/handler/api.go
type ResponseBody struct {
	Code              string `json:"code"`
	ProviderName      string `json:"provider_name,omitempty"`
	ProviderType      string `json:"provider_type,omitempty"`
	ProviderErrorCode string `json:"provider_error_code,omitempty"`
	GoError           string `json:"go_error,omitempty"`
	DumpedResponse    []byte `json:"dumped_response,omitempty"`
}

func ParseResponseBody(jsonData []byte) (*ResponseBody, error) {
	var response ResponseBody
	err := json.Unmarshal(jsonData, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
