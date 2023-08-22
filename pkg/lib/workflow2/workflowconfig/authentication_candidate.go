package workflowconfig

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

type PasswordPolicy struct {
	MinimumLength      *int `json:"minimum_length,omitempty"`
	UppercaseRequired  bool `json:"uppercase_required,omitempty"`
	LowercaseRequired  bool `json:"lowercase_required,omitempty"`
	AlphabetRequired   bool `json:"alphabet_required,omitempty"`
	DigitRequired      bool `json:"digit_required,omitempty"`
	SymbolRequired     bool `json:"symbol_required,omitempty"`
	MinimumZxcvbnScore *int `json:"minimum_zxcvbn_score,omitempty"`
}

func NewPasswordPolicy(c *config.PasswordPolicyConfig) *PasswordPolicy {
	policy := &PasswordPolicy{
		MinimumLength:     c.MinLength,
		UppercaseRequired: c.UppercaseRequired,
		LowercaseRequired: c.LowercaseRequired,
		AlphabetRequired:  c.AlphabetRequired,
		DigitRequired:     c.DigitRequired,
		SymbolRequired:    c.SymbolRequired,
	}
	if c.MinimumGuessableLevel > 0 {
		score := c.MinimumGuessableLevel - 1
		policy.MinimumZxcvbnScore = &score
	}
	return policy
}

type UseAuthenticationCandidate struct {
	AuthenticationMethod config.WorkflowAuthenticationMethod `json:"authentication_method"`
	MaskedDisplayName    string                              `json:"masked_display_name,omitempty"`
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
			AuthenticationMethod: m,
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
				AuthenticationMethod: config.WorkflowAuthenticationMethodPrimaryPassword,
			}
		case model.AuthenticatorTypePasskey:
			return UseAuthenticationCandidate{
				AuthenticationMethod: config.WorkflowAuthenticationMethodPrimaryPasskey,
			}
		case model.AuthenticatorTypeOOBEmail:
			return UseAuthenticationCandidate{
				AuthenticationMethod: config.WorkflowAuthenticationMethodPrimaryOOBOTPEmail,
				MaskedDisplayName:    mail.MaskAddress(i.OOBOTP.Email),
				AuthenticatorID:      i.ID,
			}
		case model.AuthenticatorTypeOOBSMS:
			return UseAuthenticationCandidate{
				AuthenticationMethod: config.WorkflowAuthenticationMethodPrimaryOOBOTPSMS,
				MaskedDisplayName:    phone.Mask(i.OOBOTP.Phone),
				AuthenticatorID:      i.ID,
			}
		}
	case model.AuthenticatorKindSecondary:
		switch i.Type {
		case model.AuthenticatorTypePassword:
			return UseAuthenticationCandidate{
				AuthenticationMethod: config.WorkflowAuthenticationMethodSecondaryPassword,
			}
		case model.AuthenticatorTypeOOBEmail:
			return UseAuthenticationCandidate{
				AuthenticationMethod: config.WorkflowAuthenticationMethodSecondaryOOBOTPEmail,
				MaskedDisplayName:    mail.MaskAddress(i.OOBOTP.Email),
				AuthenticatorID:      i.ID,
			}
		case model.AuthenticatorTypeOOBSMS:
			return UseAuthenticationCandidate{
				AuthenticationMethod: config.WorkflowAuthenticationMethodSecondaryOOBOTPSMS,
				MaskedDisplayName:    phone.Mask(i.OOBOTP.Phone),
				AuthenticatorID:      i.ID,
			}
		case model.AuthenticatorTypeTOTP:
			return UseAuthenticationCandidate{
				AuthenticationMethod: config.WorkflowAuthenticationMethodSecondaryTOTP,
			}
		}
	}

	panic(fmt.Errorf("unknown authentication method: %v %v", i.Kind, i.Type))
}

func KeepAuthenticationMethod(ams ...config.WorkflowAuthenticationMethod) authenticator.Filter {
	return authenticator.FilterFunc(func(ai *authenticator.Info) bool {
		am := NewUseAuthenticationCandidateFromInfo(ai).AuthenticationMethod
		for _, t := range ams {
			if t == am {
				return true
			}
		}
		return false
	})
}
