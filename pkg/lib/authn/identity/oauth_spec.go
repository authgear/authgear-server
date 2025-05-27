package identity

import (
	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type OAuthSpec struct {
	// DoNotStoreIdentityAttributes is not data, but a piece of information to be
	// used by the identity service when creating a new identity.Info.
	DoNotStoreIdentityAttributes bool `json:"do_not_store_identity_attributes"`

	ProviderID     oauthrelyingparty.ProviderID `json:"provider_id"`
	SubjectID      string                       `json:"subject_id"`
	RawProfile     map[string]interface{}       `json:"raw_profile,omitempty"`
	StandardClaims map[string]interface{}       `json:"standard_claims,omitempty"`
}

func NewIncomingOAuthSpec(providerConfig oauthrelyingparty.ProviderConfig, userProfile oauthrelyingparty.UserProfile) *OAuthSpec {
	return &OAuthSpec{
		DoNotStoreIdentityAttributes: config.OAuthSSOProviderConfig(providerConfig).DoNotStoreIdentityAttributes(),
		ProviderID:                   providerConfig.ProviderID(),
		SubjectID:                    userProfile.ProviderUserID,
		RawProfile:                   userProfile.ProviderRawProfile,
		StandardClaims:               userProfile.StandardAttributes,
	}
}
