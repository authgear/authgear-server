package otp

import (
	"github.com/authgear/authgear-server/pkg/util/template"
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
	URL                  string
	Host                 string
	StaticAssetURLPrefix string
}

const (
	TemplateItemTypeVerificationSMSTXT    string = "verification_sms.txt"
	TemplateItemTypeVerificationEmailTXT  string = "verification_email.txt"
	TemplateItemTypeVerificationEmailHTML string = "verification_email.html"

	TemplateItemTypeSetupPrimaryOOBSMSTXT    string = "setup_primary_oob_sms.txt"
	TemplateItemTypeSetupPrimaryOOBEmailTXT  string = "setup_primary_oob_email.txt"
	TemplateItemTypeSetupPrimaryOOBEmailHTML string = "setup_primary_oob_email.html"

	TemplateItemTypeSetupSecondaryOOBSMSTXT    string = "setup_secondary_oob_sms.txt"
	TemplateItemTypeSetupSecondaryOOBEmailTXT  string = "setup_secondary_oob_email.txt"
	TemplateItemTypeSetupSecondaryOOBEmailHTML string = "setup_secondary_oob_email.html"

	TemplateItemTypeAuthenticatePrimaryOOBSMSTXT    string = "authenticate_primary_oob_sms.txt"
	TemplateItemTypeAuthenticatePrimaryOOBEmailTXT  string = "authenticate_primary_oob_email.txt"
	TemplateItemTypeAuthenticatePrimaryOOBEmailHTML string = "authenticate_primary_oob_email.html"

	TemplateItemTypeAuthenticateSecondaryOOBSMSTXT    string = "authenticate_secondary_oob_sms.txt"
	TemplateItemTypeAuthenticateSecondaryOOBEmailTXT  string = "authenticate_secondary_oob_email.txt"
	TemplateItemTypeAuthenticateSecondaryOOBEmailHTML string = "authenticate_secondary_oob_email.html"
)

var (
	TemplateVerificationSMSTXT    = template.T{Type: TemplateItemTypeVerificationSMSTXT}
	TemplateVerificationEmailTXT  = template.T{Type: TemplateItemTypeVerificationEmailTXT}
	TemplateVerificationEmailHTML = template.T{Type: TemplateItemTypeVerificationEmailHTML, IsHTML: true}

	TemplateSetupPrimaryOOBSMSTXT    = template.T{Type: TemplateItemTypeSetupPrimaryOOBSMSTXT}
	TemplateSetupPrimaryOOBEmailTXT  = template.T{Type: TemplateItemTypeSetupPrimaryOOBEmailTXT}
	TemplateSetupPrimaryOOBEmailHTML = template.T{Type: TemplateItemTypeSetupPrimaryOOBEmailHTML, IsHTML: true}

	TemplateSetupSecondaryOOBSMSTXT    = template.T{Type: TemplateItemTypeSetupSecondaryOOBSMSTXT}
	TemplateSetupSecondaryOOBEmailTXT  = template.T{Type: TemplateItemTypeSetupSecondaryOOBEmailTXT}
	TemplateSetupSecondaryOOBEmailHTML = template.T{Type: TemplateItemTypeSetupSecondaryOOBEmailHTML, IsHTML: true}

	TemplateAuthenticatePrimaryOOBSMSTXT    = template.T{Type: TemplateItemTypeAuthenticatePrimaryOOBSMSTXT}
	TemplateAuthenticatePrimaryOOBEmailTXT  = template.T{Type: TemplateItemTypeAuthenticatePrimaryOOBEmailTXT}
	TemplateAuthenticatePrimaryOOBEmailHTML = template.T{Type: TemplateItemTypeAuthenticatePrimaryOOBEmailHTML, IsHTML: true}

	TemplateAuthenticateSecondaryOOBSMSTXT    = template.T{Type: TemplateItemTypeAuthenticateSecondaryOOBSMSTXT}
	TemplateAuthenticateSecondaryOOBEmailTXT  = template.T{Type: TemplateItemTypeAuthenticateSecondaryOOBEmailTXT}
	TemplateAuthenticateSecondaryOOBEmailHTML = template.T{Type: TemplateItemTypeAuthenticateSecondaryOOBEmailHTML, IsHTML: true}
)
