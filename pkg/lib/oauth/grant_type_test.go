package oauth

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestGetAllowedGrantTypes(t *testing.T) {
	Convey("GetAllowedGrantTypes", t, func() {
		test := func(clientGrantTypes []string, expected []string) {
			c := &config.OAuthClientConfig{
				GrantTypes_do_not_use_directly: clientGrantTypes,
			}
			actual := GetAllowedGrantTypes(c)
			So(actual, ShouldResemble, expected)
		}

		test(nil, []string{
			"authorization_code",
			"refresh_token",
			"urn:ietf:params:oauth:grant-type:token-exchange",
			"urn:authgear:params:oauth:grant-type:anonymous-request",
			"urn:authgear:params:oauth:grant-type:biometric-request",
			"urn:authgear:params:oauth:grant-type:app2app-request",
			"urn:authgear:params:oauth:grant-type:id-token",
			"urn:authgear:params:oauth:grant-type:settings-action",
		})

		test([]string{}, []string{
			"authorization_code",
			"refresh_token",
			"urn:ietf:params:oauth:grant-type:token-exchange",
			"urn:authgear:params:oauth:grant-type:anonymous-request",
			"urn:authgear:params:oauth:grant-type:biometric-request",
			"urn:authgear:params:oauth:grant-type:app2app-request",
			"urn:authgear:params:oauth:grant-type:id-token",
			"urn:authgear:params:oauth:grant-type:settings-action",
		})

		test([]string{"authorization_code"}, []string{
			"authorization_code",
			"refresh_token",
			"urn:ietf:params:oauth:grant-type:token-exchange",
			"urn:authgear:params:oauth:grant-type:anonymous-request",
			"urn:authgear:params:oauth:grant-type:biometric-request",
			"urn:authgear:params:oauth:grant-type:app2app-request",
			"urn:authgear:params:oauth:grant-type:id-token",
			"urn:authgear:params:oauth:grant-type:settings-action",
		})

		test([]string{"authorization_code", "refresh_token"}, []string{
			"authorization_code",
			"refresh_token",
			"urn:ietf:params:oauth:grant-type:token-exchange",
			"urn:authgear:params:oauth:grant-type:anonymous-request",
			"urn:authgear:params:oauth:grant-type:biometric-request",
			"urn:authgear:params:oauth:grant-type:app2app-request",
			"urn:authgear:params:oauth:grant-type:id-token",
			"urn:authgear:params:oauth:grant-type:settings-action",
		})

		test([]string{"refresh_token", "authorization_code"}, []string{
			"refresh_token",
			"authorization_code",
			"urn:ietf:params:oauth:grant-type:token-exchange",
			"urn:authgear:params:oauth:grant-type:anonymous-request",
			"urn:authgear:params:oauth:grant-type:biometric-request",
			"urn:authgear:params:oauth:grant-type:app2app-request",
			"urn:authgear:params:oauth:grant-type:id-token",
			"urn:authgear:params:oauth:grant-type:settings-action",
		})
	})
}
