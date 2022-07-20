package config_test

import (
	"testing"

	"github.com/authgear/authgear-server/pkg/lib/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestOAuthClientConfig(t *testing.T) {
	Convey("OAuthClientConfig", t, func() {
		c := &config.OAuthClientConfig{
			RedirectURIs: []string{
				"https://example.com",
				"https://app.example.com/",
				"https://app.example.com/redirect",
				"myapp://hostname/path",
				"http://localhost:3000",
				"http://accounts.localhost:3100",
			},
		}

		So(c.RedirectURIHosts(), ShouldResemble, []string{
			"example.com",
			"app.example.com",
			"app.example.com",
			"localhost:3000",
			"accounts.localhost:3100",
		})
	})
}
