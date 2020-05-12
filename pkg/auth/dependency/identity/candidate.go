package identity

import (
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type Candidate map[string]interface{}

const (
	CandidateKeyType = "type"

	CandidateKeyProviderType  = "provider_type"
	CandidateKeyProviderAlias = "provider_alias"

	CandidateKeyLoginIDType = "login_id_type"
	CandidateKeyLoginIDKey  = "login_id_key"
)

func NewOAuthCandidate(c *config.OAuthProviderConfiguration) Candidate {
	return Candidate{
		CandidateKeyType:          string(authn.IdentityTypeOAuth),
		CandidateKeyProviderType:  string(c.Type),
		CandidateKeyProviderAlias: string(c.ID),
	}
}

func NewLoginIDCandidate(c *config.LoginIDKeyConfiguration) Candidate {
	return Candidate{
		CandidateKeyType:        string(authn.IdentityTypeLoginID),
		CandidateKeyLoginIDType: string(c.Type),
		CandidateKeyLoginIDKey:  string(c.Key),
	}
}
