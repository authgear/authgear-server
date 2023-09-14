package tester

import "github.com/authgear/authgear-server/pkg/lib/config"

const (
	ClientIDTester string = "tester"
)

func NewTesterClient(testerEndpoint string) *config.OAuthClientConfig {
	return &config.OAuthClientConfig{
		ClientID:        ClientIDTester,
		Name:            "Tester",
		ApplicationType: config.OAuthClientApplicationTypeSPA,
		RedirectURIs:    []string{testerEndpoint},
		GrantTypes:      []string{"authorization_code", "refresh_token"},
	}
}
