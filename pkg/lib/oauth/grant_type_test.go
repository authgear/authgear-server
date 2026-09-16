package oauth

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestGetAllowedGrantTypes(t *testing.T) {
	Convey("GetAllowedGrantTypes", t, func() {
		test := func(name string, c *config.OAuthClientConfig, expected []string) {
			Convey(name, func() {
				actual := GetAllowedGrantTypes(c)
				So(actual, ShouldResemble, expected)
			})
		}

		test("nil clientGrantTypes", &config.OAuthClientConfig{
			GrantTypes_do_not_use_directly: nil,
		}, []string{
			"authorization_code",
			"refresh_token",
			"urn:ietf:params:oauth:grant-type:token-exchange",
			"urn:authgear:params:oauth:grant-type:anonymous-request",
			"urn:authgear:params:oauth:grant-type:biometric-request",
			"urn:authgear:params:oauth:grant-type:app2app-request",
			"urn:authgear:params:oauth:grant-type:id-token",
			"urn:authgear:params:oauth:grant-type:settings-action",
		})

		test("empty clientGrantTypes", &config.OAuthClientConfig{
			GrantTypes_do_not_use_directly: []string{},
		}, []string{
			"authorization_code",
			"refresh_token",
			"urn:ietf:params:oauth:grant-type:token-exchange",
			"urn:authgear:params:oauth:grant-type:anonymous-request",
			"urn:authgear:params:oauth:grant-type:biometric-request",
			"urn:authgear:params:oauth:grant-type:app2app-request",
			"urn:authgear:params:oauth:grant-type:id-token",
			"urn:authgear:params:oauth:grant-type:settings-action",
		})

		test("authorization_code only", &config.OAuthClientConfig{
			GrantTypes_do_not_use_directly: []string{"authorization_code"},
		}, []string{
			"authorization_code",
			"refresh_token",
			"urn:ietf:params:oauth:grant-type:token-exchange",
			"urn:authgear:params:oauth:grant-type:anonymous-request",
			"urn:authgear:params:oauth:grant-type:biometric-request",
			"urn:authgear:params:oauth:grant-type:app2app-request",
			"urn:authgear:params:oauth:grant-type:id-token",
			"urn:authgear:params:oauth:grant-type:settings-action",
		})

		test("authorization_code and refresh_token", &config.OAuthClientConfig{
			GrantTypes_do_not_use_directly: []string{"authorization_code", "refresh_token"},
		}, []string{
			"authorization_code",
			"refresh_token",
			"urn:ietf:params:oauth:grant-type:token-exchange",
			"urn:authgear:params:oauth:grant-type:anonymous-request",
			"urn:authgear:params:oauth:grant-type:biometric-request",
			"urn:authgear:params:oauth:grant-type:app2app-request",
			"urn:authgear:params:oauth:grant-type:id-token",
			"urn:authgear:params:oauth:grant-type:settings-action",
		})

		test("refresh_token and authorization_code (reversed)", &config.OAuthClientConfig{
			GrantTypes_do_not_use_directly: []string{"refresh_token", "authorization_code"},
		}, []string{
			"refresh_token",
			"authorization_code",
			"urn:ietf:params:oauth:grant-type:token-exchange",
			"urn:authgear:params:oauth:grant-type:anonymous-request",
			"urn:authgear:params:oauth:grant-type:biometric-request",
			"urn:authgear:params:oauth:grant-type:app2app-request",
			"urn:authgear:params:oauth:grant-type:id-token",
			"urn:authgear:params:oauth:grant-type:settings-action",
		})

		test("confidential client does not get client_credentials", &config.OAuthClientConfig{
			ApplicationType:                config.OAuthClientApplicationTypeConfidential,
			GrantTypes_do_not_use_directly: []string{},
		}, []string{
			"authorization_code",
			"refresh_token",
			"urn:ietf:params:oauth:grant-type:token-exchange",
			"urn:authgear:params:oauth:grant-type:anonymous-request",
			"urn:authgear:params:oauth:grant-type:biometric-request",
			"urn:authgear:params:oauth:grant-type:app2app-request",
			"urn:authgear:params:oauth:grant-type:id-token",
			"urn:authgear:params:oauth:grant-type:settings-action",
		})

		// Regression guard: client_credentials is no longer the default for
		// confidential clients, but nothing validates GrantTypes_do_not_use_directly
		// against ApplicationType, so a config left over from before that
		// change (or one hand-edited to add it) must not re-grant it.
		test("confidential client cannot re-grant client_credentials via raw config", &config.OAuthClientConfig{
			ApplicationType:                config.OAuthClientApplicationTypeConfidential,
			GrantTypes_do_not_use_directly: []string{"client_credentials"},
		}, []string{
			"authorization_code",
			"refresh_token",
			"urn:ietf:params:oauth:grant-type:token-exchange",
			"urn:authgear:params:oauth:grant-type:anonymous-request",
			"urn:authgear:params:oauth:grant-type:biometric-request",
			"urn:authgear:params:oauth:grant-type:app2app-request",
			"urn:authgear:params:oauth:grant-type:id-token",
			"urn:authgear:params:oauth:grant-type:settings-action",
		})

		test("m2m client keeps client_credentials", &config.OAuthClientConfig{
			ApplicationType: config.OAuthClientApplicationTypeM2M,
		}, []string{
			"client_credentials",
		})
	})
}
