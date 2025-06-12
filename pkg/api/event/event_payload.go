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

func (c *AuthenticationContext) AddAssertedIdentity(iden model.Identity) bool {
	for _, existingID := range c.AssertedIdentities {
		if existingID.ID == iden.ID {
			return false
		}
	}
	c.AssertedIdentities = append(c.AssertedIdentities, iden)
	return true
}

func (c *AuthenticationContext) AddAssertedAuthenticator(authn model.Authenticator) bool {
	for _, existingAuth := range c.AssertedAuthenticators {
		if existingAuth.ID == authn.ID {
			return false
		}
	}
	c.AssertedAuthenticators = append(c.AssertedAuthenticators, authn)
	return true
}
