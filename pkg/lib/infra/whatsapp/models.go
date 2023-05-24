package whatsapp

import (
	"encoding/json"
	"strings"
	"time"
)

type SendTemplateRequest struct {
	RecipientType string    `json:"recipient_type"`
	To            string    `json:"to"`
	Type          string    `json:"type"`
	Template      *Template `json:"template"`
}

type Template struct {
	Name       string              `json:"name"`
	Language   *TemplateLanguage   `json:"language"`
	Components []TemplateComponent `json:"components"`
	Namespace  *string             `json:"namespace,omitempty"`
}

type TemplateLanguage struct {
	Policy string `json:"policy"`
	Code   string `json:"code"`
}

type TemplateComponentType string

const (
	TemplateComponentTypeHeader TemplateComponentType = "header"
	TemplateComponentTypeBody   TemplateComponentType = "body"
)

type TemplateComponent struct {
	Type       TemplateComponentType        `json:"type"`
	Parameters []TemplateComponentParameter `json:"parameters"`
}

func NewTemplateComponent(t TemplateComponentType) *TemplateComponent {
	return &TemplateComponent{
		Type:       t,
		Parameters: []TemplateComponentParameter{},
	}
}

type TemplateComponentParameterType string

const (
	TemplateComponentParameterTypeText TemplateComponentParameterType = "text"
)

type TemplateComponentParameter struct {
	Type TemplateComponentParameterType `json:"type"`
	Text string                         `json:"text"`
}

func NewTemplateComponentTextParameter(text string) *TemplateComponentParameter {
	return &TemplateComponentParameter{
		Type: TemplateComponentParameterTypeText,
		Text: text,
	}
}

type LoginResponse struct {
	Users []LoginResponseUser `json:"users"`
}

type LoginResponseUser struct {
	Token        string                       `json:"token"`
	ExpiresAfter LoginResponseUserExpiresTime `json:"expires_after"`
}

type UserToken struct {
	Namespace string    `json:"namespace"`
	Username  string    `json:"username"`
	Token     string    `json:"token"`
	ExpireAt  time.Time `json:"expire_at"`
}

type LoginResponseUserExpiresTime time.Time

// Implement Marshaler and Unmarshaler interface
func (j *LoginResponseUserExpiresTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse("2006-01-02 15:04:05-07:00", string(s))
	if err != nil {
		return err
	}
	*j = LoginResponseUserExpiresTime(t)
	return nil
}

func (j LoginResponseUserExpiresTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(j))
}
