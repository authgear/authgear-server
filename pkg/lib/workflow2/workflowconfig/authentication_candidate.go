package workflowconfig

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

type AuthenticationCandidateKey string

type AuthenticationCandidate map[AuthenticationCandidateKey]interface{}

const (
	AuthenticationCandidateKeyAuthenticationMethod AuthenticationCandidateKey = "authentication_method"
	AuthenticationCandidateKeyAuthenticatorID      AuthenticationCandidateKey = "authenticator_id"
	AuthenticationCandidateKeyMaskedDisplayName    AuthenticationCandidateKey = "masked_display_name"
)

func NewAuthenticationCandidateFromMethod(m config.WorkflowAuthenticationMethod) AuthenticationCandidate {
	return AuthenticationCandidate{
		AuthenticationCandidateKeyAuthenticationMethod: m,
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
				AuthenticationCandidateKeyAuthenticationMethod: config.WorkflowAuthenticationMethodPrimaryPassword,
			}
		case model.AuthenticatorTypePasskey:
			return AuthenticationCandidate{
				AuthenticationCandidateKeyAuthenticationMethod: config.WorkflowAuthenticationMethodPrimaryPasskey,
			}
		case model.AuthenticatorTypeOOBEmail:
			return AuthenticationCandidate{
				AuthenticationCandidateKeyAuthenticationMethod: config.WorkflowAuthenticationMethodPrimaryOOBOTPEmail,
				AuthenticationCandidateKeyAuthenticatorID:      i.ID,
				AuthenticationCandidateKeyMaskedDisplayName:    mail.MaskAddress(i.OOBOTP.Email),
			}
		case model.AuthenticatorTypeOOBSMS:
			return AuthenticationCandidate{
				AuthenticationCandidateKeyAuthenticationMethod: config.WorkflowAuthenticationMethodPrimaryOOBOTPSMS,
				AuthenticationCandidateKeyAuthenticatorID:      i.ID,
				AuthenticationCandidateKeyMaskedDisplayName:    phone.Mask(i.OOBOTP.Phone),
			}
		}
	case model.AuthenticatorKindSecondary:
		switch i.Type {
		case model.AuthenticatorTypePassword:
			return AuthenticationCandidate{
				AuthenticationCandidateKeyAuthenticationMethod: config.WorkflowAuthenticationMethodSecondaryPassword,
			}
		case model.AuthenticatorTypeOOBEmail:
			return AuthenticationCandidate{
				AuthenticationCandidateKeyAuthenticationMethod: config.WorkflowAuthenticationMethodSecondaryOOBOTPEmail,
				AuthenticationCandidateKeyAuthenticatorID:      i.ID,
				AuthenticationCandidateKeyMaskedDisplayName:    mail.MaskAddress(i.OOBOTP.Email),
			}
		case model.AuthenticatorTypeOOBSMS:
			return AuthenticationCandidate{
				AuthenticationCandidateKeyAuthenticationMethod: config.WorkflowAuthenticationMethodSecondaryOOBOTPSMS,
				AuthenticationCandidateKeyAuthenticatorID:      i.ID,
				AuthenticationCandidateKeyMaskedDisplayName:    phone.Mask(i.OOBOTP.Phone),
			}
		case model.AuthenticatorTypeTOTP:
			return AuthenticationCandidate{
				AuthenticationCandidateKeyAuthenticationMethod: config.WorkflowAuthenticationMethodSecondaryTOTP,
			}
		}
	}

	panic(fmt.Errorf("unknown authentication method: %v %v", i.Kind, i.Type))
}

func (c AuthenticationCandidate) AuthenticationMethod() config.WorkflowAuthenticationMethod {
	return c[AuthenticationCandidateKeyAuthenticationMethod].(config.WorkflowAuthenticationMethod)
}

func KeepAuthenticationMethod(ams ...config.WorkflowAuthenticationMethod) authenticator.Filter {
	return authenticator.FilterFunc(func(ai *authenticator.Info) bool {
		am := NewAuthenticationCandidateFromInfo(ai).AuthenticationMethod()
		for _, t := range ams {
			if t == am {
				return true
			}
		}
		return false
	})
}
