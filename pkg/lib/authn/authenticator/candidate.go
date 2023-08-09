package authenticator

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type Candidate map[string]interface{}

const (
	CandidateKeyAuthenticationMethod = "authentication_method"
	CandidateKeyAuthenticatorID      = "authenticator_id"
	CandidateKeyAuthenticatorType    = "authenticator_type"
	CandidateKeyAuthenticatorKind    = "authenticator_kind"
	CandidateKeyMaskedDisplayID      = "masked_display_id"
)

func NewCandidateRecoveryCode() Candidate {
	return Candidate{
		CandidateKeyAuthenticationMethod: config.WorkflowAuthenticationMethodRecoveryCode,
	}
}
