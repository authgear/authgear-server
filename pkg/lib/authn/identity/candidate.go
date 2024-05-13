package identity

import (
	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type Candidate map[string]interface{}

const (
	CandidateKeyIdentityID = "identity_id"
	CandidateKeyType       = "type"

	CandidateKeyProviderType      = "provider_type"
	CandidateKeyProviderAlias     = "provider_alias"
	CandidateKeyProviderSubjectID = "provider_subject_id"
	CandidateKeyProviderAppType   = "provider_app_type"

	CandidateKeyLoginIDType  = "login_id_type"
	CandidateKeyLoginIDKey   = "login_id_key"
	CandidateKeyLoginIDValue = "login_id_value"

	CandidateKeyDisplayID = "display_id"

	CandidateKeyModifyDisabled = "modify_disabled"
)

func NewOAuthCandidate(cfg oauthrelyingparty.ProviderConfig) Candidate {
	// Ideally, we should import oauthrelyingparty/wechat and use ProviderConfig there.
	// But that will result in import cycle.
	app_type, _ := cfg["app_type"].(string)
	return Candidate{
		CandidateKeyIdentityID:        "",
		CandidateKeyType:              string(model.IdentityTypeOAuth),
		CandidateKeyProviderType:      string(cfg.Type()),
		CandidateKeyProviderAlias:     cfg.Alias(),
		CandidateKeyProviderSubjectID: "",
		CandidateKeyProviderAppType:   app_type,
		CandidateKeyDisplayID:         "",
		CandidateKeyModifyDisabled:    cfg.ModifyDisabled(),
	}
}

func NewLoginIDCandidate(c *config.LoginIDKeyConfig) Candidate {
	return Candidate{
		CandidateKeyIdentityID:     "",
		CandidateKeyType:           string(model.IdentityTypeLoginID),
		CandidateKeyLoginIDType:    string(c.Type),
		CandidateKeyLoginIDKey:     c.Key,
		CandidateKeyLoginIDValue:   "",
		CandidateKeyDisplayID:      "",
		CandidateKeyModifyDisabled: *c.ModifyDisabled,
	}
}

func NewSIWECandidate() Candidate {
	return Candidate{
		CandidateKeyIdentityID: "",
		CandidateKeyType:       string(model.IdentityTypeSIWE),
		CandidateKeyDisplayID:  "",
	}
}

func IsOAuthSSOProviderTypeDisabled(cfg oauthrelyingparty.ProviderConfig, featureConfig *config.OAuthSSOProvidersFeatureConfig) bool {
	return featureConfig.IsDisabled(cfg)
}
