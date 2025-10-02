package declarative

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

type AuthenticateOptionForOutput struct {
	Authentication model.AuthenticationFlowAuthentication `json:"authentication"`

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
	Authentication model.AuthenticationFlowAuthentication `json:"authentication"`

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

	AMR []string `json:"amr,omitempty"`
}

var _ AMROption = AuthenticateOption{}

func (o AuthenticateOption) GetAMR() []string {
	return o.AMR
}

func (o *AuthenticateOption) ToOutput(ctx context.Context) AuthenticateOptionForOutput {
	shdBypassBotProtection := ShouldExistingResultBypassBotProtectionRequirement(ctx)
	if shdBypassBotProtection {
		o.BotProtection = nil
	}
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

func NewAuthenticateOptionRecoveryCode(flows authflow.Flows, authflowBotProtectionCfg *config.AuthenticationFlowBotProtection, appBotProtectionConfig *config.BotProtectionConfig) AuthenticateOption {
	return AuthenticateOption{
		Authentication: model.AuthenticationFlowAuthenticationRecoveryCode,
		BotProtection:  GetBotProtectionData(flows, authflowBotProtectionCfg, appBotProtectionConfig),
		AMR:            model.AuthenticationFlowAuthenticationRecoveryCode.AMR(),
	}
}

func NewAuthenticateOptionPassword(flows authflow.Flows, am model.AuthenticationFlowAuthentication, authflowBotProtectionCfg *config.AuthenticationFlowBotProtection, appBotProtectionConfig *config.BotProtectionConfig) AuthenticateOption {
	return AuthenticateOption{
		Authentication: am,
		BotProtection:  GetBotProtectionData(flows, authflowBotProtectionCfg, appBotProtectionConfig),
		AMR:            am.AMR(),
	}
}

func NewAuthenticateOptionPasskey(flows authflow.Flows, requestOptions *model.WebAuthnRequestOptions, authflowBotProtectionCfg *config.AuthenticationFlowBotProtection, appBotProtectionConfig *config.BotProtectionConfig) AuthenticateOption {
	return AuthenticateOption{
		Authentication: model.AuthenticationFlowAuthenticationPrimaryPasskey,
		RequestOptions: requestOptions,
		BotProtection:  GetBotProtectionData(flows, authflowBotProtectionCfg, appBotProtectionConfig),
		AMR:            model.AuthenticationFlowAuthenticationPrimaryPasskey.AMR(),
	}
}

func NewAuthenticateOptionTOTP(flows authflow.Flows, authflowBotProtectionCfg *config.AuthenticationFlowBotProtection, appBotProtectionConfig *config.BotProtectionConfig) AuthenticateOption {
	return AuthenticateOption{
		Authentication: model.AuthenticationFlowAuthenticationSecondaryTOTP,
		BotProtection:  GetBotProtectionData(flows, authflowBotProtectionCfg, appBotProtectionConfig),
		AMR:            model.AuthenticationFlowAuthenticationSecondaryTOTP.AMR(),
	}
}

func NewAuthenticateOptionOOBOTPFromAuthenticator(flows authflow.Flows, oobConfig *config.AuthenticatorOOBConfig, i *authenticator.Info, authflowBotProtectionCfg *config.AuthenticationFlowBotProtection, appBotProtectionConfig *config.BotProtectionConfig) (*AuthenticateOption, bool) {
	am := AuthenticationFromAuthenticator(i)
	switch am {
	case model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
		fallthrough
	case model.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
		purpose := otp.PurposeOOBOTP
		channels := getChannels(model.ClaimEmail, oobConfig)
		otpForm := getOTPForm(purpose, model.ClaimEmail, oobConfig.Email)
		return &AuthenticateOption{
			Authentication:    am,
			OTPForm:           otpForm,
			Channels:          channels,
			MaskedDisplayName: mail.MaskAddress(i.OOBOTP.Email),
			AuthenticatorID:   i.ID,
			BotProtection:     GetBotProtectionData(flows, authflowBotProtectionCfg, appBotProtectionConfig),
			AMR:               am.AMR(),
		}, true
	case model.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
		fallthrough
	case model.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
		purpose := otp.PurposeOOBOTP
		channels := getChannels(model.ClaimPhoneNumber, oobConfig)
		otpForm := getOTPForm(purpose, model.ClaimPhoneNumber, oobConfig.Email)
		return &AuthenticateOption{
			Authentication:    am,
			OTPForm:           otpForm,
			Channels:          channels,
			MaskedDisplayName: phone.Mask(i.OOBOTP.Phone),
			AuthenticatorID:   i.ID,
			BotProtection:     GetBotProtectionData(flows, authflowBotProtectionCfg, appBotProtectionConfig),
			AMR:               am.AMR(),
		}, true
	default:
		return nil, false
	}
}

func NewAuthenticateOptionOOBOTPFromIdentity(
	ctx context.Context, flows authflow.Flows, deps *authflow.Dependencies,
	i *identity.Info,
	authflowBotProtectionCfg *config.AuthenticationFlowBotProtection,
	appBotProtectionConfig *config.BotProtectionConfig,
) (*AuthenticateOption, bool, error) {
	oobConfig := deps.Config.Authenticator.OOB
	switch i.Type {
	case model.IdentityTypeLoginID:
		identityID := i.ID
		authnID := ""
		target := i.LoginID.LoginID
		var authentication model.AuthenticationFlowAuthentication
		switch i.LoginID.LoginIDType {
		case model.LoginIDKeyTypeEmail:
			authentication = model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail
		case model.LoginIDKeyTypePhone:
			authentication = model.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS
		default:
			return nil, false, nil
		}
		expectedAuthenticatorInfo, err := createAuthenticator(ctx, deps, i.UserID, authentication, target)
		if err != nil {
			return nil, false, err
		}
		allAuthenticators, err := deps.Authenticators.List(ctx, i.UserID)
		if err != nil {
			return nil, false, err
		}

		for _, authenticator := range allAuthenticators {
			if authenticator.Equal(expectedAuthenticatorInfo) {
				// An existing authenticator is found, use it instead of identity ID
				authnID = authenticator.ID
				identityID = ""
				break
			}
		}
		switch i.LoginID.LoginIDType {
		case model.LoginIDKeyTypeEmail:
			purpose := otp.PurposeOOBOTP
			channels := getChannels(model.ClaimEmail, oobConfig)
			otpForm := getOTPForm(purpose, model.ClaimEmail, oobConfig.Email)
			return &AuthenticateOption{
				Authentication:    model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail,
				OTPForm:           otpForm,
				Channels:          channels,
				MaskedDisplayName: mail.MaskAddress(i.LoginID.LoginID),
				IdentityID:        identityID,
				AuthenticatorID:   authnID,
				BotProtection:     GetBotProtectionData(flows, authflowBotProtectionCfg, appBotProtectionConfig),
				AMR:               model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail.AMR(),
			}, true, nil
		case model.LoginIDKeyTypePhone:
			purpose := otp.PurposeOOBOTP
			channels := getChannels(model.ClaimPhoneNumber, oobConfig)
			otpForm := getOTPForm(purpose, model.ClaimPhoneNumber, oobConfig.Email)
			return &AuthenticateOption{
				Authentication:    model.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS,
				OTPForm:           otpForm,
				Channels:          channels,
				MaskedDisplayName: phone.Mask(i.LoginID.LoginID),
				IdentityID:        identityID,
				AuthenticatorID:   authnID,
				BotProtection:     GetBotProtectionData(flows, authflowBotProtectionCfg, appBotProtectionConfig),
				AMR:               model.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS.AMR(),
			}, true, nil
		default:
			return nil, false, nil
		}
	default:
		return nil, false, nil
	}
}

func AuthenticationFromAuthenticator(i *authenticator.Info) model.AuthenticationFlowAuthentication {
	switch i.Kind {
	case model.AuthenticatorKindPrimary:
		switch i.Type {
		case model.AuthenticatorTypePassword:
			return model.AuthenticationFlowAuthenticationPrimaryPassword
		case model.AuthenticatorTypePasskey:
			return model.AuthenticationFlowAuthenticationPrimaryPasskey
		case model.AuthenticatorTypeOOBEmail:
			return model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail
		case model.AuthenticatorTypeOOBSMS:
			return model.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS
		}
	case model.AuthenticatorKindSecondary:
		switch i.Type {
		case model.AuthenticatorTypePassword:
			return model.AuthenticationFlowAuthenticationSecondaryPassword
		case model.AuthenticatorTypeOOBEmail:
			return model.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail
		case model.AuthenticatorTypeOOBSMS:
			return model.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS
		case model.AuthenticatorTypeTOTP:
			return model.AuthenticationFlowAuthenticationSecondaryTOTP
		}
	}

	panic(fmt.Errorf("unknown authentication method: %v %v", i.Kind, i.Type))
}

func KeepAuthenticationMethod(ams ...model.AuthenticationFlowAuthentication) authenticator.Filter {
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
