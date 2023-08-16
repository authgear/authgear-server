package workflowconfig

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type IdentificationCandidateKey string

type IdentificationCandidate map[IdentificationCandidateKey]interface{}

const (
	IdentificationCandidateKeyIdentificationMethod IdentificationCandidateKey = "identification_method"
)

func NewIdentificationCandidateFromMethod(m config.WorkflowIdentificationMethod) IdentificationCandidate {
	return IdentificationCandidate{
		IdentificationCandidateKeyIdentificationMethod: m,
	}
}

func (m IdentificationCandidate) IdentificationMethod() config.WorkflowIdentificationMethod {
	return m[IdentificationCandidateKeyIdentificationMethod].(config.WorkflowIdentificationMethod)
}
