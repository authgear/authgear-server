package declarative

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

type UseAuthenticationCandidate struct {
	Authentication config.AuthenticationFlowAuthentication `json:"authentication"`

	// MaskedDisplayName is specific to OOBOTP.
	MaskedDisplayName string `json:"masked_display_name,omitempty"`
	// Channels is specific to OOBOTP.
	Channels []model.AuthenticatorOOBChannel `json:"channels,omitempty"`

	// Count is specific to password.
	Count *int `json:"count,omitempty"`

	// AuthenticatorID is omitted from the output.
	// The caller must use index to select a candidate.
	AuthenticatorID string `json:"-"`
}

func NewUseAuthenticationCandidateRecoveryCode() UseAuthenticationCandidate {
	return UseAuthenticationCandidate{
		Authentication: config.AuthenticationFlowAuthenticationRecoveryCode,
	}
}

func NewUseAuthenticationCandidatePassword(am config.AuthenticationFlowAuthentication, count int) UseAuthenticationCandidate {
	return UseAuthenticationCandidate{
		Authentication: am,
		Count:          &count,
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

func NewUseAuthenticationCandidateFromInfo(oobConfig *config.AuthenticatorOOBConfig, i *authenticator.Info) UseAuthenticationCandidate {
	am := AuthenticationFromAuthenticator(i)
	switch am {
	case config.AuthenticationFlowAuthenticationPrimaryPassword:
		fallthrough
	case config.AuthenticationFlowAuthenticationSecondaryPassword:
		count := 0
		return UseAuthenticationCandidate{
			Authentication: am,
			Count:          &count,
		}

	case config.AuthenticationFlowAuthenticationSecondaryTOTP:
		return UseAuthenticationCandidate{
			Authentication: am,
		}

	case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
		fallthrough
	case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
		channels := getChannels(model.ClaimEmail, oobConfig)
		return UseAuthenticationCandidate{
			Authentication:    am,
			Channels:          channels,
			MaskedDisplayName: mail.MaskAddress(i.OOBOTP.Email),
			AuthenticatorID:   i.ID,
		}
	case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
		fallthrough
	case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
		channels := getChannels(model.ClaimPhoneNumber, oobConfig)
		return UseAuthenticationCandidate{
			Authentication:    am,
			Channels:          channels,
			MaskedDisplayName: phone.Mask(i.OOBOTP.Phone),
			AuthenticatorID:   i.ID,
		}
	case config.AuthenticationFlowAuthenticationPrimaryPasskey:
		// FIXME(authflow): support passkey in login flow.
		break
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
