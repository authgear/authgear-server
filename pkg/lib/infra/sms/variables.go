package sms

import (
	"github.com/authgear/authgear-server/pkg/lib/translation"
)

type TemplateVariables struct {
	AppName     string `json:"app_name,omitempty"`
	ClientID    string `json:"client_id,omitempty"`
	Code        string `json:"code,omitempty"`
	Email       string `json:"email,omitempty"`
	HasPassword bool   `json:"has_password,omitempty"`
	Host        string `json:"host,omitempty"`
	Link        string `json:"link,omitempty"`
	Password    string `json:"password,omitempty"`
	Phone       string `json:"phone,omitempty"`
	State       string `json:"state,omitempty"`
	UILocales   string `json:"ui_locales,omitempty"`
	URL         string `json:"url,omitempty"`
	XState      string `json:"x_state,omitempty"`
}

func NewTemplateVariablesFromPreparedTemplateVariables(v *translation.PreparedTemplateVariables) *TemplateVariables {
	return &TemplateVariables{
		AppName:     v.AppName,
		ClientID:    v.ClientID,
		Code:        v.Code,
		Email:       v.Email,
		HasPassword: v.HasPassword,
		Host:        v.Host,
		Link:        v.Link,
		Password:    v.Password,
		Phone:       v.Phone,
		State:       v.State,
		UILocales:   v.UILocales,
		URL:         v.URL,
		XState:      v.XState,
	}
}
