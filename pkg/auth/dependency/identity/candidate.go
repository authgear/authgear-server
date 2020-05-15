package identity

import (
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type Candidate map[string]interface{}

const (
	CandidateKeyType = "type"

	CandidateKeyEmail = "email"

	CandidateKeyProviderType      = "provider_type"
	CandidateKeyProviderAlias     = "provider_alias"
	CandidateKeyProviderSubjectID = "provider_subject_id"

	CandidateKeyLoginIDType  = "login_id_type"
	CandidateKeyLoginIDKey   = "login_id_key"
	CandidateKeyLoginIDValue = "login_id_value"
)

func NewOAuthCandidate(c *config.OAuthProviderConfiguration) Candidate {
	return Candidate{
		CandidateKeyType:              string(authn.IdentityTypeOAuth),
		CandidateKeyEmail:             "",
		CandidateKeyProviderType:      string(c.Type),
		CandidateKeyProviderAlias:     string(c.ID),
		CandidateKeyProviderSubjectID: "",
	}
}

func NewLoginIDCandidate(c *config.LoginIDKeyConfiguration) Candidate {
	return Candidate{
		CandidateKeyType:         string(authn.IdentityTypeLoginID),
		CandidateKeyEmail:        "",
		CandidateKeyLoginIDType:  string(c.Type),
		CandidateKeyLoginIDKey:   string(c.Key),
		CandidateKeyLoginIDValue: "",
	}
}
