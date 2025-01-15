package tester

import "github.com/authgear/authgear-server/pkg/lib/config"

const (
	ClientIDTester string = "tester"
)

func NewTesterClient(testerEndpoint string) *config.OAuthClientConfig {
	c := &config.OAuthClientConfig{
		ClientID:                       ClientIDTester,
		Name:                           "Tester",
		ApplicationType:                config.OAuthClientApplicationTypeSPA,
		RedirectURIs:                   []string{testerEndpoint},
		GrantTypes_do_not_use_directly: []string{"authorization_code", "refresh_token"},
	}
	c.SetDefaults()
	return c
}
