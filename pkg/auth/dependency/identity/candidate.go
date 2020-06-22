package identity

import (
	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/core/authn"
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

func NewOAuthCandidate(c *config.OAuthSSOProviderConfig) Candidate {
	return Candidate{
		CandidateKeyType:              string(authn.IdentityTypeOAuth),
		CandidateKeyEmail:             "",
		CandidateKeyProviderType:      string(c.Type),
		CandidateKeyProviderAlias:     c.Alias,
		CandidateKeyProviderSubjectID: "",
	}
}

func NewLoginIDCandidate(c *config.LoginIDKeyConfig) Candidate {
	return Candidate{
		CandidateKeyType:         string(authn.IdentityTypeLoginID),
		CandidateKeyEmail:        "",
		CandidateKeyLoginIDType:  string(c.Type),
		CandidateKeyLoginIDKey:   c.Key,
		CandidateKeyLoginIDValue: "",
	}
}
