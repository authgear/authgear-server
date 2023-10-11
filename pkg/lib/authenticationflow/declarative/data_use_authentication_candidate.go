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

type UseAuthenticationOption struct {
	Authentication config.AuthenticationFlowAuthentication `json:"authentication"`

	// OTPForm is specific to OOBOTP.
	OTPForm otp.Form `json:"otp_form,omitempty"`
	// MaskedDisplayName is specific to OOBOTP.
	MaskedDisplayName string `json:"masked_display_name,omitempty"`
	// Channels is specific to OOBOTP.
	Channels []model.AuthenticatorOOBChannel `json:"channels,omitempty"`

	// Count is specific to password.
	Count *int `json:"count,omitempty"`

	// WebAuthnRequestOptions is specific to Passkey.
	RequestOptions *model.WebAuthnRequestOptions `json:"request_options,omitempty"`

	// AuthenticatorID is omitted from the output.
	// The caller must use index to select a option.
	AuthenticatorID string `json:"-"`
}

func NewUseAuthenticationOptionRecoveryCode() UseAuthenticationOption {
	return UseAuthenticationOption{
		Authentication: config.AuthenticationFlowAuthenticationRecoveryCode,
	}
}

func NewUseAuthenticationOptionPassword(am config.AuthenticationFlowAuthentication, count int) UseAuthenticationOption {
	return UseAuthenticationOption{
		Authentication: am,
		Count:          &count,
	}
}

func NewUseAuthenticationOptionPasskey(requestOptions *model.WebAuthnRequestOptions) UseAuthenticationOption {
	return UseAuthenticationOption{
		Authentication: config.AuthenticationFlowAuthenticationPrimaryPasskey,
		RequestOptions: requestOptions,
	}
}

func NewUseAuthenticationOptionTOTP() UseAuthenticationOption {
	return UseAuthenticationOption{
		Authentication: config.AuthenticationFlowAuthenticationSecondaryTOTP,
	}
}

func NewUseAuthenticationOptionOOBOTP(oobConfig *config.AuthenticatorOOBConfig, i *authenticator.Info) UseAuthenticationOption {
	am := AuthenticationFromAuthenticator(i)
	switch am {
	case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
		fallthrough
	case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
		channels := getChannels(model.ClaimEmail, oobConfig)
		otpForm := getOTPForm(model.ClaimEmail, oobConfig.Email)
		return UseAuthenticationOption{
			Authentication:    am,
			OTPForm:           otpForm,
			Channels:          channels,
			MaskedDisplayName: mail.MaskAddress(i.OOBOTP.Email),
			AuthenticatorID:   i.ID,
		}
	case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
		fallthrough
	case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
		channels := getChannels(model.ClaimPhoneNumber, oobConfig)
		otpForm := getOTPForm(model.ClaimPhoneNumber, oobConfig.Email)
		return UseAuthenticationOption{
			Authentication:    am,
			OTPForm:           otpForm,
			Channels:          channels,
			MaskedDisplayName: phone.Mask(i.OOBOTP.Phone),
			AuthenticatorID:   i.ID,
		}
	}

	panic(fmt.Errorf("NewUseAuthenticationOptionOOBOTP: unexpected authentication method: %v %v", i.Kind, i.Type))
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
