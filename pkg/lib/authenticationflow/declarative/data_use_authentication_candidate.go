package declarative

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

type UseAuthenticationCandidate struct {
	Authentication config.AuthenticationFlowAuthentication `json:"authentication"`

	// MaskedDisplayName is specific to OOBOTP.
	MaskedDisplayName string `json:"masked_display_name,omitempty"`

	// AuthenticatorID is omitted from the output.
	// The caller must use index to select a candidate.
	AuthenticatorID string `json:"-"`
}

// NewUseAuthenticationCandidateFromMethod is not a total function.
// It will panic for invalid input.
func NewUseAuthenticationCandidateFromMethod(m config.AuthenticationFlowAuthentication) UseAuthenticationCandidate {
	switch m {
	case config.AuthenticationFlowAuthenticationPrimaryPassword:
		fallthrough
	case config.AuthenticationFlowAuthenticationPrimaryPasskey:
		fallthrough
	case config.AuthenticationFlowAuthenticationSecondaryPassword:
		fallthrough
	case config.AuthenticationFlowAuthenticationSecondaryTOTP:
		fallthrough
	case config.AuthenticationFlowAuthenticationRecoveryCode:
		return UseAuthenticationCandidate{
			Authentication: m,
		}
	case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
		fallthrough
	case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
		fallthrough
	case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
		fallthrough
	case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
		fallthrough
	case config.AuthenticationFlowAuthenticationDeviceToken:
		panic(fmt.Errorf("unexpected call to NewUseAuthenticationCandidateFromMethod: %v", m))
	default:
		panic(fmt.Errorf("unknown authentication method: %v", m))
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

func NewUseAuthenticationCandidateFromInfo(i *authenticator.Info) UseAuthenticationCandidate {
	am := AuthenticationFromAuthenticator(i)
	candidate := UseAuthenticationCandidate{
		Authentication: am,
	}
	switch am {
	case config.AuthenticationFlowAuthenticationPrimaryPassword:
		fallthrough
	case config.AuthenticationFlowAuthenticationSecondaryPassword:
		fallthrough
	case config.AuthenticationFlowAuthenticationPrimaryPasskey:
		fallthrough
	case config.AuthenticationFlowAuthenticationSecondaryTOTP:
		return candidate
	case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
		fallthrough
	case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
		candidate.MaskedDisplayName = mail.MaskAddress(i.OOBOTP.Email)
		candidate.AuthenticatorID = i.ID
		return candidate
	case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
		fallthrough
	case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
		candidate.MaskedDisplayName = phone.Mask(i.OOBOTP.Phone)
		candidate.AuthenticatorID = i.ID
		return candidate
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
