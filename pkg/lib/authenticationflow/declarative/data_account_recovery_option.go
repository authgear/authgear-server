package declarative

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
)

type AccountRecoveryIdentificationOption struct {
	Identification config.AuthenticationFlowAccountRecoveryIdentification `json:"identification"`
	BotProtection  *BotProtectionData                                     `json:"bot_protection,omitempty"`
}

func (i *AccountRecoveryIdentificationOption) isBotProtectionRequired() bool {
	if i.BotProtection == nil {
		return false
	}
	if i.BotProtection.Enabled != nil && *i.BotProtection.Enabled && i.BotProtection.Provider != nil && i.BotProtection.Provider.Type != "" {
		return true
	}

	return false
}

type AccountRecoveryChannel string

const (
	AccountRecoveryChannelEmail    AccountRecoveryChannel = AccountRecoveryChannel(config.AccountRecoveryCodeChannelEmail)
	AccountRecoveryChannelSMS      AccountRecoveryChannel = AccountRecoveryChannel(config.AccountRecoveryCodeChannelSMS)
	AccountRecoveryChannelWhatsapp AccountRecoveryChannel = AccountRecoveryChannel(config.AccountRecoveryCodeChannelWhatsapp)
)

type AccountRecoveryOTPForm string

const (
	AccountRecoveryOTPFormLink AccountRecoveryOTPForm = AccountRecoveryOTPForm(config.AccountRecoveryCodeFormLink)
	AccountRecoveryOTPFormCode AccountRecoveryOTPForm = AccountRecoveryOTPForm(config.AccountRecoveryCodeFormCode)
)

type AccountRecoveryDestinationOption struct {
	MaskedDisplayName string                 `json:"masked_display_name"`
	Channel           AccountRecoveryChannel `json:"channel"`
	OTPForm           AccountRecoveryOTPForm `json:"otp_form"`
}

type AccountRecoveryDestinationOptionInternal struct {
	AccountRecoveryDestinationOption
	TargetLoginID string `json:"target_login_id"`
}

func (o *AccountRecoveryDestinationOptionInternal) ForgotPasswordCodeKind() forgotpassword.CodeKind {
	switch o.OTPForm {
	case AccountRecoveryOTPFormCode:
		return forgotpassword.CodeKindShortCode
	case AccountRecoveryOTPFormLink:
		return forgotpassword.CodeKindLink
	}
	panic(fmt.Sprintf("account recovery: unknown otp form %s", o.OTPForm))
}

func (o *AccountRecoveryDestinationOptionInternal) ForgotPasswordCodeChannel() forgotpassword.CodeChannel {
	switch o.Channel {
	case AccountRecoveryChannelWhatsapp:
		return forgotpassword.CodeChannelWhatsapp
	case AccountRecoveryChannelSMS:
		return forgotpassword.CodeChannelSMS
	case AccountRecoveryChannelEmail:
		return forgotpassword.CodeChannelEmail
	default:
		return forgotpassword.CodeChannelUnknown
	}
}

type AccountRecoveryIdentity struct {
	Identification config.AuthenticationFlowAccountRecoveryIdentification
	IdentitySpec   *identity.Spec
	MaybeIdentity  *identity.Info
}
