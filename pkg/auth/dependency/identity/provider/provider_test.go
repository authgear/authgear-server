package provider

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func TestProviderListCandidates(t *testing.T) {
	Convey("Provider ListCandidates", t, func() {
		p := &Provider{
			Authentication: &config.AuthenticationConfiguration{},
			Identity: &config.IdentityConfiguration{
				OAuth:   &config.OAuthConfiguration{},
				LoginID: &config.LoginIDConfiguration{},
			},
		}

		Convey("no candidates", func() {
			actual := p.ListCandidates()
			So(actual, ShouldBeEmpty)
		})

		Convey("oauth", func() {
			p.Authentication.Identities = []string{"oauth"}
			p.Identity.OAuth.Providers = []config.OAuthProviderConfiguration{
				{
					ID:   "google",
					Type: "google",
				},
			}

			actual := p.ListCandidates()
			So(actual, ShouldResemble, []identity.Candidate{
				{
					"type":           "oauth",
					"provider_type":  "google",
					"provider_alias": "google",
				},
			})
		})

		Convey("loginid", func() {
			p.Authentication.Identities = []string{"login_id"}
			p.Identity.LoginID.Keys = []config.LoginIDKeyConfiguration{
				{
					Type: "email",
					Key:  "email",
				},
			}

			actual := p.ListCandidates()
			So(actual, ShouldResemble, []identity.Candidate{
				{
					"type":          "login_id",
					"login_id_type": "email",
					"login_id_key":  "email",
				},
			})
		})

		Convey("respect authentication", func() {
			p.Identity.OAuth.Providers = []config.OAuthProviderConfiguration{
				{
					ID:   "google",
					Type: "google",
				},
			}
			p.Identity.LoginID.Keys = []config.LoginIDKeyConfiguration{
				{
					Type: "email",
					Key:  "email",
				},
			}

			actual := p.ListCandidates()
			So(actual, ShouldBeEmpty)
		})
	})
}
