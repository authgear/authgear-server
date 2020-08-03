package otp

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/template"
)

type MessageType string

const (
	MessageTypeVerification             MessageType = "verification"
	MessageTypeSetupPrimaryOOB          MessageType = "setup-primary-oob"
	MessageTypeSetupSecondaryOOB        MessageType = "setup-secondary-oob"
	MessageTypeAuthenticatePrimaryOOB   MessageType = "authenticate-primary-oob"
	MessageTypeAuthenticateSecondaryOOB MessageType = "authenticate-secondary-oob"
)

type MessageTemplateContext struct {
	AppName              string
	Email                string
	Phone                string
	Code                 string
	Host                 string
	StaticAssetURLPrefix string
}

const (
	TemplateItemTypeVerificationSMSTXT    config.TemplateItemType = "verification_sms.txt"
	TemplateItemTypeVerificationEmailTXT  config.TemplateItemType = "verification_email.txt"
	TemplateItemTypeVerificationEmailHTML config.TemplateItemType = "verification_email.html"

	TemplateItemTypeSetupPrimaryOOBSMSTXT    config.TemplateItemType = "setup_primary_oob_sms.txt"
	TemplateItemTypeSetupPrimaryOOBEmailTXT  config.TemplateItemType = "setup_primary_oob_email.txt"
	TemplateItemTypeSetupPrimaryOOBEmailHTML config.TemplateItemType = "setup_primary_oob_email.html"

	TemplateItemTypeSetupSecondaryOOBSMSTXT    config.TemplateItemType = "setup_secondary_oob_sms.txt"
	TemplateItemTypeSetupSecondaryOOBEmailTXT  config.TemplateItemType = "setup_secondary_oob_email.txt"
	TemplateItemTypeSetupSecondaryOOBEmailHTML config.TemplateItemType = "setup_secondary_oob_email.html"

	TemplateItemTypeAuthenticatePrimaryOOBSMSTXT    config.TemplateItemType = "authenticate_primary_oob_sms.txt"
	TemplateItemTypeAuthenticatePrimaryOOBEmailTXT  config.TemplateItemType = "authenticate_primary_oob_email.txt"
	TemplateItemTypeAuthenticatePrimaryOOBEmailHTML config.TemplateItemType = "authenticate_primary_oob_email.html"

	TemplateItemTypeAuthenticateSecondaryOOBSMSTXT    config.TemplateItemType = "authenticate_secondary_oob_sms.txt"
	TemplateItemTypeAuthenticateSecondaryOOBEmailTXT  config.TemplateItemType = "authenticate_secondary_oob_email.txt"
	TemplateItemTypeAuthenticateSecondaryOOBEmailHTML config.TemplateItemType = "authenticate_secondary_oob_email.html"
)

var (
	TemplateVerificationSMSTXT    = template.Spec{Type: TemplateItemTypeVerificationSMSTXT}
	TemplateVerificationEmailTXT  = template.Spec{Type: TemplateItemTypeVerificationEmailTXT}
	TemplateVerificationEmailHTML = template.Spec{Type: TemplateItemTypeVerificationEmailHTML, IsHTML: true}

	TemplateSetupPrimaryOOBSMSTXT    = template.Spec{Type: TemplateItemTypeSetupPrimaryOOBSMSTXT}
	TemplateSetupPrimaryOOBEmailTXT  = template.Spec{Type: TemplateItemTypeSetupPrimaryOOBEmailTXT}
	TemplateSetupPrimaryOOBEmailHTML = template.Spec{Type: TemplateItemTypeSetupPrimaryOOBEmailHTML, IsHTML: true}

	TemplateSetupSecondaryOOBSMSTXT    = template.Spec{Type: TemplateItemTypeSetupSecondaryOOBSMSTXT}
	TemplateSetupSecondaryOOBEmailTXT  = template.Spec{Type: TemplateItemTypeSetupSecondaryOOBEmailTXT}
	TemplateSetupSecondaryOOBEmailHTML = template.Spec{Type: TemplateItemTypeSetupSecondaryOOBEmailHTML, IsHTML: true}

	TemplateAuthenticatePrimaryOOBSMSTXT    = template.Spec{Type: TemplateItemTypeAuthenticatePrimaryOOBSMSTXT}
	TemplateAuthenticatePrimaryOOBEmailTXT  = template.Spec{Type: TemplateItemTypeAuthenticatePrimaryOOBEmailTXT}
	TemplateAuthenticatePrimaryOOBEmailHTML = template.Spec{Type: TemplateItemTypeAuthenticatePrimaryOOBEmailHTML, IsHTML: true}

	TemplateAuthenticateSecondaryOOBSMSTXT    = template.Spec{Type: TemplateItemTypeAuthenticateSecondaryOOBSMSTXT}
	TemplateAuthenticateSecondaryOOBEmailTXT  = template.Spec{Type: TemplateItemTypeAuthenticateSecondaryOOBEmailTXT}
	TemplateAuthenticateSecondaryOOBEmailHTML = template.Spec{Type: TemplateItemTypeAuthenticateSecondaryOOBEmailHTML, IsHTML: true}
)
