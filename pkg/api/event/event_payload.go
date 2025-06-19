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
	AuthenticationFlow      *AuthenticationFlowContext `json:"authentication_flow,omitzero"`
	User                    *model.User                `json:"user,omitzero"`
	AssertedIdentifications []model.Identification     `json:"asserted_identifications,omitzero"`
	AssertedAuthenticators  []model.Authenticator      `json:"asserted_authenticators,omitzero"`
	AMR                     []string                   `json:"amr,omitzero"`
}

func (c *AuthenticationContext) AddAssertedIdentification(iden model.Identification) bool {
	if iden.Identity != nil {
		for _, existingID := range c.AssertedIdentifications {
			if existingID.Identity == nil {
				continue
			}
			if existingID.Identity.ID == iden.Identity.ID {
				return false
			}
		}
		c.AssertedIdentifications = append(c.AssertedIdentifications, iden)
		return true
	}
	if iden.IDToken != nil {
		for _, existingID := range c.AssertedIdentifications {
			if existingID.IDToken == nil {
				continue
			}
			if *existingID.IDToken == *iden.IDToken {
				return false
			}
		}
		c.AssertedIdentifications = append(c.AssertedIdentifications, iden)
		return true
	}
	return false
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
