package identity

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/authn"
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

func NewOAuthCandidate(c *config.OAuthSSOProviderConfig) Candidate {
	return Candidate{
		CandidateKeyIdentityID:        "",
		CandidateKeyType:              string(authn.IdentityTypeOAuth),
		CandidateKeyProviderType:      string(c.Type),
		CandidateKeyProviderAlias:     c.Alias,
		CandidateKeyProviderSubjectID: "",
		CandidateKeyProviderAppType:   string(c.AppType),
		CandidateKeyDisplayID:         "",
		CandidateKeyModifyDisabled:    *c.ModifyDisabled,
	}
}

func NewLoginIDCandidate(c *config.LoginIDKeyConfig) Candidate {
	return Candidate{
		CandidateKeyIdentityID:     "",
		CandidateKeyType:           string(authn.IdentityTypeLoginID),
		CandidateKeyLoginIDType:    string(c.Type),
		CandidateKeyLoginIDKey:     c.Key,
		CandidateKeyLoginIDValue:   "",
		CandidateKeyDisplayID:      "",
		CandidateKeyModifyDisabled: *c.ModifyDisabled,
	}
}

func IsOAuthSSOProviderTypeDisabled(typ config.OAuthSSOProviderType, featureConfig *config.OAuthSSOProvidersFeatureConfig) bool {
	switch typ {
	case config.OAuthSSOProviderTypeGoogle:
		return featureConfig.Google.Disabled
	case config.OAuthSSOProviderTypeFacebook:
		return featureConfig.Facebook.Disabled
	case config.OAuthSSOProviderTypeLinkedIn:
		return featureConfig.LinkedIn.Disabled
	case config.OAuthSSOProviderTypeAzureADv2:
		return featureConfig.Azureadv2.Disabled
	case config.OAuthSSOProviderTypeADFS:
		return featureConfig.ADFS.Disabled
	case config.OAuthSSOProviderTypeApple:
		return featureConfig.Apple.Disabled
	case config.OAuthSSOProviderTypeWechat:
		return featureConfig.Wechat.Disabled
	default:
		panic(fmt.Sprintf("node: unknown oauth sso type: %T", typ))
	}
}
