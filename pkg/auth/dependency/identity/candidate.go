package identity

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

type Candidate map[string]interface{}

const (
	CandidateKeyIdentityID = "identity_id"
	CandidateKeyType       = "type"

	CandidateKeyProviderType      = "provider_type"
	CandidateKeyProviderAlias     = "provider_alias"
	CandidateKeyProviderSubjectID = "provider_subject_id"

	CandidateKeyLoginIDType  = "login_id_type"
	CandidateKeyLoginIDKey   = "login_id_key"
	CandidateKeyLoginIDValue = "login_id_value"

	CandidateKeyDisplayID = "display_id"
)

func NewOAuthCandidate(c *config.OAuthSSOProviderConfig) Candidate {
	return Candidate{
		CandidateKeyIdentityID:        "",
		CandidateKeyType:              string(authn.IdentityTypeOAuth),
		CandidateKeyProviderType:      string(c.Type),
		CandidateKeyProviderAlias:     c.Alias,
		CandidateKeyProviderSubjectID: "",
		CandidateKeyDisplayID:         "",
	}
}

func NewLoginIDCandidate(c *config.LoginIDKeyConfig) Candidate {
	return Candidate{
		CandidateKeyIdentityID:   "",
		CandidateKeyType:         string(authn.IdentityTypeLoginID),
		CandidateKeyLoginIDType:  string(c.Type),
		CandidateKeyLoginIDKey:   c.Key,
		CandidateKeyLoginIDValue: "",
		CandidateKeyDisplayID:    "",
	}
}
