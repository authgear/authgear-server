package workflowconfig

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

type UseAuthenticationCandidate struct {
	Authentication    config.WorkflowAuthenticationMethod `json:"authentication"`
	MaskedDisplayName string                              `json:"masked_display_name,omitempty"`
	// AuthenticatorID is omitted from the output.
	// The caller must use index to select a candidate.
	AuthenticatorID string `json:"-"`
}

// NewUseAuthenticationCandidateFromMethod is not a total function.
// It will panic for invalid input.
func NewUseAuthenticationCandidateFromMethod(m config.WorkflowAuthenticationMethod) UseAuthenticationCandidate {
	switch m {
	case config.WorkflowAuthenticationMethodPrimaryPassword:
		fallthrough
	case config.WorkflowAuthenticationMethodPrimaryPasskey:
		fallthrough
	case config.WorkflowAuthenticationMethodSecondaryPassword:
		fallthrough
	case config.WorkflowAuthenticationMethodSecondaryTOTP:
		fallthrough
	case config.WorkflowAuthenticationMethodRecoveryCode:
		return UseAuthenticationCandidate{
			Authentication: m,
		}
	case config.WorkflowAuthenticationMethodPrimaryOOBOTPEmail:
		fallthrough
	case config.WorkflowAuthenticationMethodPrimaryOOBOTPSMS:
		fallthrough
	case config.WorkflowAuthenticationMethodSecondaryOOBOTPEmail:
		fallthrough
	case config.WorkflowAuthenticationMethodSecondaryOOBOTPSMS:
		fallthrough
	case config.WorkflowAuthenticationMethodDeviceToken:
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
				Authentication: config.WorkflowAuthenticationMethodPrimaryPassword,
			}
		case model.AuthenticatorTypePasskey:
			return UseAuthenticationCandidate{
				Authentication: config.WorkflowAuthenticationMethodPrimaryPasskey,
			}
		case model.AuthenticatorTypeOOBEmail:
			return UseAuthenticationCandidate{
				Authentication:    config.WorkflowAuthenticationMethodPrimaryOOBOTPEmail,
				MaskedDisplayName: mail.MaskAddress(i.OOBOTP.Email),
				AuthenticatorID:   i.ID,
			}
		case model.AuthenticatorTypeOOBSMS:
			return UseAuthenticationCandidate{
				Authentication:    config.WorkflowAuthenticationMethodPrimaryOOBOTPSMS,
				MaskedDisplayName: phone.Mask(i.OOBOTP.Phone),
				AuthenticatorID:   i.ID,
			}
		}
	case model.AuthenticatorKindSecondary:
		switch i.Type {
		case model.AuthenticatorTypePassword:
			return UseAuthenticationCandidate{
				Authentication: config.WorkflowAuthenticationMethodSecondaryPassword,
			}
		case model.AuthenticatorTypeOOBEmail:
			return UseAuthenticationCandidate{
				Authentication:    config.WorkflowAuthenticationMethodSecondaryOOBOTPEmail,
				MaskedDisplayName: mail.MaskAddress(i.OOBOTP.Email),
				AuthenticatorID:   i.ID,
			}
		case model.AuthenticatorTypeOOBSMS:
			return UseAuthenticationCandidate{
				Authentication:    config.WorkflowAuthenticationMethodSecondaryOOBOTPSMS,
				MaskedDisplayName: phone.Mask(i.OOBOTP.Phone),
				AuthenticatorID:   i.ID,
			}
		case model.AuthenticatorTypeTOTP:
			return UseAuthenticationCandidate{
				Authentication: config.WorkflowAuthenticationMethodSecondaryTOTP,
			}
		}
	}

	panic(fmt.Errorf("unknown authentication method: %v %v", i.Kind, i.Type))
}

func KeepAuthenticationMethod(ams ...config.WorkflowAuthenticationMethod) authenticator.Filter {
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
