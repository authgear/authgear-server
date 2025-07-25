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
	AuthenticationFlow      *AuthenticationFlowContext `json:"authentication_flow"`
	User                    *model.User                `json:"user"`
	AssertedIdentifications []model.Identification     `json:"asserted_identifications"`
	AssertedAuthentications []model.Authentication     `json:"asserted_authentications"`
	AMR                     []string                   `json:"amr"`
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

func (c *AuthenticationContext) AddAssertedAuthentication(authn model.Authentication) bool {
	if authn.Authenticator == nil {
		c.AssertedAuthentications = append(c.AssertedAuthentications, authn)
		return true
	}
	for _, existingAuth := range c.AssertedAuthentications {
		if existingAuth.Authenticator == nil {
			continue
		}
		if existingAuth.Authenticator.ID == authn.Authenticator.ID {
			return false
		}
	}
	c.AssertedAuthentications = append(c.AssertedAuthentications, authn)
	return true
}
