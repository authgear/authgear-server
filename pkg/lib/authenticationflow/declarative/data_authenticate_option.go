package declarative

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

type AuthenticateOptionForOutput struct {
	Authentication config.AuthenticationFlowAuthentication `json:"authentication"`

	BotProtection *BotProtectionData `json:"bot_protection,omitempty"`
	// OTPForm is specific to OOBOTP.
	OTPForm otp.Form `json:"otp_form,omitempty"`
	// MaskedDisplayName is specific to OOBOTP.
	MaskedDisplayName string `json:"masked_display_name,omitempty"`
	// Channels is specific to OOBOTP.
	Channels []model.AuthenticatorOOBChannel `json:"channels,omitempty"`

	// WebAuthnRequestOptions is specific to Passkey.
	RequestOptions *model.WebAuthnRequestOptions `json:"request_options,omitempty"`
}

type AuthenticateOption struct {
	Authentication config.AuthenticationFlowAuthentication `json:"authentication"`

	BotProtection *BotProtectionData `json:"bot_protection,omitempty"`
	// OTPForm is specific to OOBOTP.
	OTPForm otp.Form `json:"otp_form,omitempty"`
	// MaskedDisplayName is specific to OOBOTP.
	MaskedDisplayName string `json:"masked_display_name,omitempty"`
	// Channels is specific to OOBOTP.
	Channels []model.AuthenticatorOOBChannel `json:"channels,omitempty"`

	// WebAuthnRequestOptions is specific to Passkey.
	RequestOptions *model.WebAuthnRequestOptions `json:"request_options,omitempty"`

	AuthenticatorID string `json:"authenticator_id,omitempty"`

	IdentityID string `json:"identity_id,omitempty"`
}

func (o *AuthenticateOption) ToOutput() AuthenticateOptionForOutput {
	return AuthenticateOptionForOutput{
		Authentication:    o.Authentication,
		OTPForm:           o.OTPForm,
		BotProtection:     o.BotProtection,
		MaskedDisplayName: o.MaskedDisplayName,
		Channels:          o.Channels,
		RequestOptions:    o.RequestOptions,
	}
}

func (o *AuthenticateOption) isBotProtectionRequired() bool {
	if o.BotProtection == nil {
		return false
	}
	if o.BotProtection.Enabled != nil && *o.BotProtection.Enabled && o.BotProtection.Provider != nil && o.BotProtection.Provider.Type != "" {
		return true
	}
	return false
}

func NewAuthenticateOptionRecoveryCode(authflowBotProtectionCfg *config.AuthenticationFlowBotProtection, appBotProtectionConfig *config.BotProtectionConfig) AuthenticateOption {
	return AuthenticateOption{
		Authentication: config.AuthenticationFlowAuthenticationRecoveryCode,
		BotProtection:  GetBotProtectionData(authflowBotProtectionCfg, appBotProtectionConfig),
	}
}

func NewAuthenticateOptionPassword(am config.AuthenticationFlowAuthentication, authflowBotProtectionCfg *config.AuthenticationFlowBotProtection, appBotProtectionConfig *config.BotProtectionConfig) AuthenticateOption {
	return AuthenticateOption{
		Authentication: am,
		BotProtection:  GetBotProtectionData(authflowBotProtectionCfg, appBotProtectionConfig),
	}
}

func NewAuthenticateOptionPasskey(requestOptions *model.WebAuthnRequestOptions, authflowBotProtectionCfg *config.AuthenticationFlowBotProtection, appBotProtectionConfig *config.BotProtectionConfig) AuthenticateOption {
	return AuthenticateOption{
		Authentication: config.AuthenticationFlowAuthenticationPrimaryPasskey,
		RequestOptions: requestOptions,
		BotProtection:  GetBotProtectionData(authflowBotProtectionCfg, appBotProtectionConfig),
	}
}

func NewAuthenticateOptionTOTP(authflowBotProtectionCfg *config.AuthenticationFlowBotProtection, appBotProtectionConfig *config.BotProtectionConfig) AuthenticateOption {
	return AuthenticateOption{
		Authentication: config.AuthenticationFlowAuthenticationSecondaryTOTP,
		BotProtection:  GetBotProtectionData(authflowBotProtectionCfg, appBotProtectionConfig),
	}
}

func NewAuthenticateOptionOOBOTPFromAuthenticator(oobConfig *config.AuthenticatorOOBConfig, i *authenticator.Info, authflowBotProtectionCfg *config.AuthenticationFlowBotProtection, appBotProtectionConfig *config.BotProtectionConfig) (*AuthenticateOption, bool) {
	am := AuthenticationFromAuthenticator(i)
	switch am {
	case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
		fallthrough
	case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
		purpose := otp.PurposeOOBOTP
		channels := getChannels(model.ClaimEmail, oobConfig)
		otpForm := getOTPForm(purpose, model.ClaimEmail, oobConfig.Email)
		return &AuthenticateOption{
			Authentication:    am,
			OTPForm:           otpForm,
			Channels:          channels,
			MaskedDisplayName: mail.MaskAddress(i.OOBOTP.Email),
			AuthenticatorID:   i.ID,
			BotProtection:     GetBotProtectionData(authflowBotProtectionCfg, appBotProtectionConfig),
		}, true
	case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
		fallthrough
	case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
		purpose := otp.PurposeOOBOTP
		channels := getChannels(model.ClaimPhoneNumber, oobConfig)
		otpForm := getOTPForm(purpose, model.ClaimPhoneNumber, oobConfig.Email)
		return &AuthenticateOption{
			Authentication:    am,
			OTPForm:           otpForm,
			Channels:          channels,
			MaskedDisplayName: phone.Mask(i.OOBOTP.Phone),
			AuthenticatorID:   i.ID,
			BotProtection:     GetBotProtectionData(authflowBotProtectionCfg, appBotProtectionConfig),
		}, true
	default:
		return nil, false
	}
}

func NewAuthenticateOptionOOBOTPFromIdentity(oobConfig *config.AuthenticatorOOBConfig, i *identity.Info, authflowBotProtectionCfg *config.AuthenticationFlowBotProtection, appBotProtectionConfig *config.BotProtectionConfig) (*AuthenticateOption, bool) {
	switch i.Type {
	case model.IdentityTypeLoginID:
		switch i.LoginID.LoginIDType {
		case model.LoginIDKeyTypeEmail:
			purpose := otp.PurposeOOBOTP
			channels := getChannels(model.ClaimEmail, oobConfig)
			otpForm := getOTPForm(purpose, model.ClaimEmail, oobConfig.Email)
			return &AuthenticateOption{
				Authentication:    config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail,
				OTPForm:           otpForm,
				Channels:          channels,
				MaskedDisplayName: mail.MaskAddress(i.LoginID.LoginID),
				IdentityID:        i.ID,
				BotProtection:     GetBotProtectionData(authflowBotProtectionCfg, appBotProtectionConfig),
			}, true
		case model.LoginIDKeyTypePhone:
			purpose := otp.PurposeOOBOTP
			channels := getChannels(model.ClaimPhoneNumber, oobConfig)
			otpForm := getOTPForm(purpose, model.ClaimPhoneNumber, oobConfig.Email)
			return &AuthenticateOption{
				Authentication:    config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS,
				OTPForm:           otpForm,
				Channels:          channels,
				MaskedDisplayName: phone.Mask(i.LoginID.LoginID),
				IdentityID:        i.ID,
				BotProtection:     GetBotProtectionData(authflowBotProtectionCfg, appBotProtectionConfig),
			}, true
		default:
			return nil, false
		}
	default:
		return nil, false
	}
}

func AuthenticationFromAuthenticator(i *authenticator.Info) config.AuthenticationFlowAuthentication {
	switch i.Kind {
	case model.AuthenticatorKindPrimary:
		switch i.Type {
		case model.AuthenticatorTypePassword:
			return config.AuthenticationFlowAuthenticationPrimaryPassword
		case model.AuthenticatorTypePasskey:
			return config.AuthenticationFlowAuthenticationPrimaryPasskey
		case model.AuthenticatorTypeOOBEmail:
			return config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail
		case model.AuthenticatorTypeOOBSMS:
			return config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS
		}
	case model.AuthenticatorKindSecondary:
		switch i.Type {
		case model.AuthenticatorTypePassword:
			return config.AuthenticationFlowAuthenticationSecondaryPassword
		case model.AuthenticatorTypeOOBEmail:
			return config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail
		case model.AuthenticatorTypeOOBSMS:
			return config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS
		case model.AuthenticatorTypeTOTP:
			return config.AuthenticationFlowAuthenticationSecondaryTOTP
		}
	}

	panic(fmt.Errorf("unknown authentication method: %v %v", i.Kind, i.Type))
}

func KeepAuthenticationMethod(ams ...config.AuthenticationFlowAuthentication) authenticator.Filter {
	return authenticator.FilterFunc(func(ai *authenticator.Info) bool {
		am := AuthenticationFromAuthenticator(ai)
		for _, t := range ams {
			if t == am {
				return true
			}
		}
		return false
	})
}

func IsDependentOf(info *identity.Info) authenticator.Filter {
	return authenticator.FilterFunc(func(ai *authenticator.Info) bool {
		return ai.IsDependentOf(info)
	})
}
