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
	Authentication    config.AuthenticationFlowAuthentication `json:"authentication"`
	MaskedDisplayName string                                  `json:"masked_display_name,omitempty"`
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

func NewUseAuthenticationCandidateFromInfo(i *authenticator.Info) UseAuthenticationCandidate {
	switch i.Kind {
	case model.AuthenticatorKindPrimary:
		switch i.Type {
		case model.AuthenticatorTypePassword:
			return UseAuthenticationCandidate{
				Authentication: config.AuthenticationFlowAuthenticationPrimaryPassword,
			}
		case model.AuthenticatorTypePasskey:
			return UseAuthenticationCandidate{
				Authentication: config.AuthenticationFlowAuthenticationPrimaryPasskey,
			}
		case model.AuthenticatorTypeOOBEmail:
			return UseAuthenticationCandidate{
				Authentication:    config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail,
				MaskedDisplayName: mail.MaskAddress(i.OOBOTP.Email),
				AuthenticatorID:   i.ID,
			}
		case model.AuthenticatorTypeOOBSMS:
			return UseAuthenticationCandidate{
				Authentication:    config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS,
				MaskedDisplayName: phone.Mask(i.OOBOTP.Phone),
				AuthenticatorID:   i.ID,
			}
		}
	case model.AuthenticatorKindSecondary:
		switch i.Type {
		case model.AuthenticatorTypePassword:
			return UseAuthenticationCandidate{
				Authentication: config.AuthenticationFlowAuthenticationSecondaryPassword,
			}
		case model.AuthenticatorTypeOOBEmail:
			return UseAuthenticationCandidate{
				Authentication:    config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail,
				MaskedDisplayName: mail.MaskAddress(i.OOBOTP.Email),
				AuthenticatorID:   i.ID,
			}
		case model.AuthenticatorTypeOOBSMS:
			return UseAuthenticationCandidate{
				Authentication:    config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS,
				MaskedDisplayName: phone.Mask(i.OOBOTP.Phone),
				AuthenticatorID:   i.ID,
			}
		case model.AuthenticatorTypeTOTP:
			return UseAuthenticationCandidate{
				Authentication: config.AuthenticationFlowAuthenticationSecondaryTOTP,
			}
		}
	}

	panic(fmt.Errorf("unknown authentication method: %v %v", i.Kind, i.Type))
}

func KeepAuthenticationMethod(ams ...config.AuthenticationFlowAuthentication) authenticator.Filter {
	return authenticator.FilterFunc(func(ai *authenticator.Info) bool {
		am := NewUseAuthenticationCandidateFromInfo(ai).Authentication
		for _, t := range ams {
			if t == am {
				return true
			}
		}
		return false
	})
}
