package identity

import (
	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type OAuthSpec struct {
	// DoNotStoreIdentityAttributes is not part of the data of identity.Info, but a piece of information to be
	// used by the identity service when creating a new identity.Info.
	DoNotStoreIdentityAttributes bool `json:"do_not_store_identity_attributes"`

	// IncludeIdentityAttributesInIDToken is read by authflow service.
	IncludeIdentityAttributesInIDToken bool `json:"include_identity_attributes_in_id_token"`

	// ProviderAlias is transient data, its sole purpose is to aid in creating
	// the info object.
	ProviderAlias string `json:"provider_alias"`

	ProviderID     oauthrelyingparty.ProviderID `json:"provider_id"`
	SubjectID      string                       `json:"subject_id"`
	RawProfile     map[string]interface{}       `json:"raw_profile,omitempty"`
	StandardClaims map[string]interface{}       `json:"standard_claims,omitempty"`
}

func NewIncomingOAuthSpec(providerConfig oauthrelyingparty.ProviderConfig, userProfile oauthrelyingparty.UserProfile) *OAuthSpec {
	return &OAuthSpec{
		DoNotStoreIdentityAttributes:       config.OAuthSSOProviderConfig(providerConfig).DoNotStoreIdentityAttributes(),
		IncludeIdentityAttributesInIDToken: config.OAuthSSOProviderConfig(providerConfig).IncludeIdentityAttributesInIDToken(),
		ProviderAlias:                      config.OAuthSSOProviderConfig(providerConfig).Alias(),
		ProviderID:                         providerConfig.ProviderID(),
		SubjectID:                          userProfile.ProviderUserID,
		RawProfile:                         userProfile.ProviderRawProfile,
		StandardClaims:                     userProfile.StandardAttributes,
	}
}

func (s OAuthSpec) ToClaimsForIDToken() map[string]any {
	claims := make(map[string]any)
	claims[IdentityClaimOAuthProviderAlias] = s.ProviderAlias
	claims[IdentityClaimOAuthProviderType] = s.ProviderID.Type
	claims[IdentityClaimOAuthSubjectID] = s.SubjectID

	if s.RawProfile != nil {
		claims[IdentityClaimOAuthProfile] = s.RawProfile
	} else {
		claims[IdentityClaimOAuthProfile] = make(map[string]any)
	}

	return claims
}
