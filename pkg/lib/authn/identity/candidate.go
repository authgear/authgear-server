package identity

import (
	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/wechat"
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

	CandidateKeyCreateDisabled = "create_disabled"
	CandidateKeyUpdateDisabled = "update_disabled"
	CandidateKeyDeleteDisabled = "delete_disabled"
)

func NewOAuthCandidate(cfg oauthrelyingparty.ProviderConfig) Candidate {
	return Candidate{
		CandidateKeyIdentityID:        "",
		CandidateKeyType:              string(model.IdentityTypeOAuth),
		CandidateKeyProviderType:      string(cfg.Type()),
		CandidateKeyProviderAlias:     cfg.Alias(),
		CandidateKeyProviderSubjectID: "",
		CandidateKeyProviderAppType:   string(wechat.ProviderConfig(cfg).AppType()),
		CandidateKeyDisplayID:         "",
		CandidateKeyCreateDisabled:    cfg.ModifyDisabled(),
		CandidateKeyUpdateDisabled:    cfg.ModifyDisabled(),
		CandidateKeyDeleteDisabled:    cfg.ModifyDisabled(),
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
		CandidateKeyCreateDisabled: *c.CreateDisabled,
		CandidateKeyUpdateDisabled: *c.UpdateDisabled,
		CandidateKeyDeleteDisabled: *c.DeleteDisabled,
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
