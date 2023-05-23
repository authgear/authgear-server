package whatsapp

import "time"

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

type TemplateComponentParameter struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type LoginResponse struct {
	Users []LoginResponseUser `json:"users"`
}

type LoginResponseUser struct {
	Token        string    `json:"token"`
	ExpiresAfter time.Time `json:"expires_after"`
}
