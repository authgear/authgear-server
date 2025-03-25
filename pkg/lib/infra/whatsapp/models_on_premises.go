package whatsapp

import (
	"time"
)

type onPremisesSendTemplateRequest struct {
	RecipientType string              `json:"recipient_type"`
	To            string              `json:"to"`
	Type          string              `json:"type"`
	Template      *onPremisesTemplate `json:"template"`
}

type onPremisesTemplate struct {
	Name       string                        `json:"name"`
	Language   *onPremisesTemplateLanguage   `json:"language"`
	Components []onPremisesTemplateComponent `json:"components"`
	Namespace  *string                       `json:"namespace,omitempty"`
}

type onPremisesTemplateLanguage struct {
	Policy string `json:"policy"`
	Code   string `json:"code"`
}

type onPremisesTemplateComponentType string

const (
	onPremisesTemplateComponentTypeHeader onPremisesTemplateComponentType = "header"
	onPremisesTemplateComponentTypeBody   onPremisesTemplateComponentType = "body"
	onPremisesTemplateComponentTypeButton onPremisesTemplateComponentType = "button"
)

type onPremisesTemplateComponentSubType string

const (
	onPremisesTemplateComponentSubTypeURL onPremisesTemplateComponentSubType = "url"
)

type onPremisesTemplateComponent struct {
	Type       onPremisesTemplateComponentType        `json:"type"`
	SubType    *onPremisesTemplateComponentSubType    `json:"sub_type,omitempty"`
	Index      *int                                   `json:"index,omitempty"`
	Parameters []onPremisesTemplateComponentParameter `json:"parameters"`
}

type onPremisesTemplateComponentParameterType string

const (
	onPremisesTemplateComponentParameterTypeText onPremisesTemplateComponentParameterType = "text"
)

type onPremisesTemplateComponentParameter struct {
	Type onPremisesTemplateComponentParameterType `json:"type"`
	Text string                                   `json:"text"`
}

type onPremisesLoginResponse struct {
	Users []onPremisesLoginResponseUser `json:"users"`
}

type onPremisesLoginResponseUser struct {
	Token        string                                 `json:"token"`
	ExpiresAfter onPremisesloginResponseUserExpiresTime `json:"expires_after"`
}

type onPremisesUserToken struct {
	Endpoint string    `json:"endpoint"`
	Username string    `json:"username"`
	Token    string    `json:"token"`
	ExpireAt time.Time `json:"expire_at"`
}

type onPremisesloginResponseUserExpiresTime time.Time

const (
	onPremisesLoginResponseUserExpiresTimeLayout = "2006-01-02 15:04:05-07:00"
)

// Implement Marshaler and Unmarshaler interface
func (j *onPremisesloginResponseUserExpiresTime) UnmarshalText(textb []byte) error {
	t, err := time.Parse(onPremisesLoginResponseUserExpiresTimeLayout, string(textb))
	if err != nil {
		return err
	}
	*j = onPremisesloginResponseUserExpiresTime(t)
	return nil
}

func (j onPremisesloginResponseUserExpiresTime) MarshalText() ([]byte, error) {
	return []byte(time.Time(j).Format(onPremisesLoginResponseUserExpiresTimeLayout)), nil
}

func onPremisesNewTemplateComponent(typ onPremisesTemplateComponentType) *onPremisesTemplateComponent {
	return &onPremisesTemplateComponent{
		Type:       typ,
		Parameters: []onPremisesTemplateComponentParameter{},
	}
}

func onPremisesNewTemplateButtonComponent(subtyp onPremisesTemplateComponentSubType, index int) *onPremisesTemplateComponent {
	return &onPremisesTemplateComponent{
		Type:       onPremisesTemplateComponentTypeButton,
		SubType:    &subtyp,
		Index:      &index,
		Parameters: []onPremisesTemplateComponentParameter{},
	}
}

func onPremisesNewTemplateComponentTextParameter(text string) *onPremisesTemplateComponentParameter {
	return &onPremisesTemplateComponentParameter{
		Type: onPremisesTemplateComponentParameterTypeText,
		Text: text,
	}
}
