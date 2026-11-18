package config

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCustomUIEndpointsResolution(t *testing.T) {
	Convey("Custom UI endpoints resolution", t, func() {
		project := &UIConfig{
			CustomUIEndpoints: &CustomUIEndpointsConfig{
				Authentication:  "https://project.example.com/auth",
				LoginLink:       "https://project.example.com/login-link",
				AccountRecovery: "https://project.example.com/reset",
			},
		}

		Convey("x_custom_ui_uri overrides the project authentication page only", func() {
			client := &OAuthClientConfig{
				CustomUIURI: "https://client.example.com/auth",
			}
			So(project.ResolveCustomUIAuthenticationURI(client), ShouldEqual, "https://client.example.com/auth")
			So(project.ResolveCustomUILoginLinkURI(), ShouldEqual, "https://project.example.com/login-link")
			So(project.ResolveCustomUIAccountRecoveryURI(), ShouldEqual, "https://project.example.com/reset")
		})

		Convey("falls back to the project authentication page", func() {
			So(project.ResolveCustomUIAuthenticationURI(nil), ShouldEqual, "https://project.example.com/auth")
			So(project.ResolveCustomUIAuthenticationURI(&OAuthClientConfig{}), ShouldEqual, "https://project.example.com/auth")
		})

		Convey("returns empty when nothing is set", func() {
			var nilUI *UIConfig
			So(nilUI.ResolveCustomUIAuthenticationURI(nil), ShouldEqual, "")
			So((&UIConfig{}).ResolveCustomUILoginLinkURI(), ShouldEqual, "")
			So((&UIConfig{CustomUIEndpoints: &CustomUIEndpointsConfig{}}).ResolveCustomUIAccountRecoveryURI(), ShouldEqual, "")
		})

		Convey("lists URIs", func() {
			oauth := &OAuthConfig{
				Clients: []OAuthClientConfig{
					{CustomUIURI: "https://client.example.com/auth"},
					{},
				},
			}

			So(CustomUIAllURIs(project, oauth), ShouldResemble, []string{
				"https://project.example.com/auth",
				"https://project.example.com/login-link",
				"https://project.example.com/reset",
				"https://client.example.com/auth",
			})

			So(CustomUIAuthenticationURIs(project, oauth), ShouldResemble, []string{
				"https://project.example.com/auth",
				"https://client.example.com/auth",
			})

			So(CustomUIAllURIs(nil, nil), ShouldBeEmpty)
		})
	})
}
