package otp

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity/loginid"
	"github.com/authgear/authgear-server/pkg/template"
)

type OOBOperationType string

const (
	OOBOperationTypeSetup        OOBOperationType = "setup"
	OOBOperationTypeAuthenticate OOBOperationType = "authenticate"
)

type OOBAuthenticationStage string

const (
	OOBAuthenticationStagePrimary   OOBAuthenticationStage = "primary_auth"
	OOBAuthenticationStageSecondary OOBAuthenticationStage = "secondary_auth"
)

type MessageTemplateContext struct {
	AppName              string
	Email                string
	Phone                string
	LoginID              *loginid.LoginID
	Code                 string
	Host                 string
	Operation            OOBOperationType
	Stage                OOBAuthenticationStage
	StaticAssetURLPrefix string
}

const (
	TemplateItemTypeOTPMessageSMSTXT    config.TemplateItemType = "otp_message_sms.txt"
	TemplateItemTypeOTPMessageEmailTXT  config.TemplateItemType = "otp_message_email.txt"
	TemplateItemTypeOTPMessageEmailHTML config.TemplateItemType = "otp_message_email.html"
)

var TemplateOTPMessageSMSTXT = template.Spec{
	Type: TemplateItemTypeOTPMessageSMSTXT,
}

var TemplateOTPMessageEmailTXT = template.Spec{
	Type: TemplateItemTypeOTPMessageEmailTXT,
}

var TemplateOTPMessageEmailHTML = template.Spec{
	Type:   TemplateItemTypeOTPMessageEmailHTML,
	IsHTML: true,
}
