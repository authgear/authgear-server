package identity

import (
	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
)

type OAuthSpec struct {
	ProviderID     oauthrelyingparty.ProviderID `json:"provider_id"`
	SubjectID      string                       `json:"subject_id"`
	RawProfile     map[string]interface{}       `json:"raw_profile,omitempty"`
	StandardClaims map[string]interface{}       `json:"standard_claims,omitempty"`
}
