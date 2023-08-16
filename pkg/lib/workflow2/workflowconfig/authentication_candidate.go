package workflowconfig

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

type AuthenticationCandidate struct {
	AuthenticationMethod config.WorkflowAuthenticationMethod `json:"authentication_method"`
	MaskedDisplayName    string                              `json:"masked_display_name,omitempty"`
	// AuthenticatorID is omitted from the output.
	// The caller must use index to select a candidate.
	AuthenticatorID string `json:"-"`
}

func NewAuthenticationCandidateFromMethod(m config.WorkflowAuthenticationMethod) AuthenticationCandidate {
	return AuthenticationCandidate{
		AuthenticationMethod: m,
	}

}

func NewAuthenticationCandidateRecoveryCode() AuthenticationCandidate {
	return NewAuthenticationCandidateFromMethod(config.WorkflowAuthenticationMethodRecoveryCode)
}

func NewAuthenticationCandidateFromInfo(i *authenticator.Info) AuthenticationCandidate {
	switch i.Kind {
	case model.AuthenticatorKindPrimary:
		switch i.Type {
		case model.AuthenticatorTypePassword:
			return AuthenticationCandidate{
				AuthenticationMethod: config.WorkflowAuthenticationMethodPrimaryPassword,
			}
		case model.AuthenticatorTypePasskey:
			return AuthenticationCandidate{
				AuthenticationMethod: config.WorkflowAuthenticationMethodPrimaryPasskey,
			}
		case model.AuthenticatorTypeOOBEmail:
			return AuthenticationCandidate{
				AuthenticationMethod: config.WorkflowAuthenticationMethodPrimaryOOBOTPEmail,
				MaskedDisplayName:    mail.MaskAddress(i.OOBOTP.Email),
				AuthenticatorID:      i.ID,
			}
		case model.AuthenticatorTypeOOBSMS:
			return AuthenticationCandidate{
				AuthenticationMethod: config.WorkflowAuthenticationMethodPrimaryOOBOTPSMS,
				MaskedDisplayName:    phone.Mask(i.OOBOTP.Phone),
				AuthenticatorID:      i.ID,
			}
		}
	case model.AuthenticatorKindSecondary:
		switch i.Type {
		case model.AuthenticatorTypePassword:
			return AuthenticationCandidate{
				AuthenticationMethod: config.WorkflowAuthenticationMethodSecondaryPassword,
			}
		case model.AuthenticatorTypeOOBEmail:
			return AuthenticationCandidate{
				AuthenticationMethod: config.WorkflowAuthenticationMethodSecondaryOOBOTPEmail,
				MaskedDisplayName:    mail.MaskAddress(i.OOBOTP.Email),
				AuthenticatorID:      i.ID,
			}
		case model.AuthenticatorTypeOOBSMS:
			return AuthenticationCandidate{
				AuthenticationMethod: config.WorkflowAuthenticationMethodSecondaryOOBOTPSMS,
				MaskedDisplayName:    phone.Mask(i.OOBOTP.Phone),
				AuthenticatorID:      i.ID,
			}
		case model.AuthenticatorTypeTOTP:
			return AuthenticationCandidate{
				AuthenticationMethod: config.WorkflowAuthenticationMethodSecondaryTOTP,
			}
		}
	}

	panic(fmt.Errorf("unknown authentication method: %v %v", i.Kind, i.Type))
}

func KeepAuthenticationMethod(ams ...config.WorkflowAuthenticationMethod) authenticator.Filter {
	return authenticator.FilterFunc(func(ai *authenticator.Info) bool {
		am := NewAuthenticationCandidateFromInfo(ai).AuthenticationMethod
		for _, t := range ams {
			if t == am {
				return true
			}
		}
		return false
	})
}
