package event

import (
	"github.com/authgear/authgear-server/pkg/api/model"
)

// Top-level struct matching the overall JSON structure
type AuthenticationContextPayload struct {
	Authentication *AuthenticationContext `json:"authentication,omitzero"`
}

type AuthenticationFlowContext struct {
	Type string `json:"type,omitzero"`
	Name string `json:"name,omitzero"`
}

type AuthenticationContext struct {
	AuthenticationFlow     *AuthenticationFlowContext `json:"authentication_flow,omitzero"`
	User                   *model.User                `json:"user,omitzero"`
	AssertedIdentities     []model.Identity           `json:"asserted_identities,omitzero"`
	AssertedAuthenticators []model.Authenticator      `json:"asserted_authenticators,omitzero"`
	AMR                    []string                   `json:"amr,omitzero"`
}
