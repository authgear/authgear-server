package declarative

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type IdentificationCandidate struct {
	Identification config.AuthenticationFlowIdentification `json:"identification"`

	// Aliases is specific to OAuth.
	Aliases []string `json:"alias,omitempty"`
}

func NewIdentificationCandidates(oauthConfig *config.OAuthSSOConfig, identifications []config.AuthenticationFlowIdentification) []IdentificationCandidate {
	output := []IdentificationCandidate{}
	for _, identification := range identifications {
		switch identification {
		case config.AuthenticationFlowIdentificationEmail:
			fallthrough
		case config.AuthenticationFlowIdentificationPhone:
			fallthrough
		case config.AuthenticationFlowIdentificationUsername:
			output = append(output, IdentificationCandidate{
				Identification: identification,
			})
		case config.AuthenticationFlowIdentificationOAuth:
			if len(oauthConfig.Providers) > 0 {
				aliases := []string{}
				for _, p := range oauthConfig.Providers {
					aliases = append(aliases, p.Alias)
				}
				output = append(output, IdentificationCandidate{
					Identification: identification,
					Aliases:        aliases,
				})
			}
		}
	}
	return output
}
