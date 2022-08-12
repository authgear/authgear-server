package identity

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type OAuthSpec struct {
	ProviderID     config.ProviderID      `json:"provider_id"`
	SubjectID      string                 `json:"subject_id"`
	RawProfile     map[string]interface{} `json:"raw_profile,omitempty"`
	StandardClaims map[string]interface{} `json:"standard_claims,omitempty"`
}
