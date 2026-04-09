package server

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
)

// validBase returns a Config with the fields that are always required set.
func validBase() *Config {
	return &Config{
		ConfigSource: &configsource.Config{Type: configsource.TypeDatabase},
		EnvironmentConfig: &config.EnvironmentConfig{
			GlobalDatabase: config.GlobalDatabaseCredentialsEnvironmentConfig{
				DatabaseURL: "postgres://localhost/test",
			},
		},
	}
}

func TestConfigValidate(t *testing.T) {
	Convey("Config.Validate", t, func() {
		Convey("always requires DATABASE_URL", func() {
			cfg := validBase()
			cfg.EnvironmentConfig.GlobalDatabase.DatabaseURL = ""
			err := cfg.Validate(LoadConfigOptions{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "DATABASE_URL")
		})

		Convey("portal mode", func() {
			Convey("requires AUTHGEAR_CLIENT_ID and AUTHGEAR_ENDPOINT", func() {
				cfg := validBase()
				err := cfg.Validate(LoadConfigOptions{ServePortal: true})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "AUTHGEAR_CLIENT_ID")
				So(err.Error(), ShouldContainSubstring, "AUTHGEAR_ENDPOINT")
			})

			Convey("passes when portal fields are set", func() {
				cfg := validBase()
				cfg.Authgear = portalconfig.AuthgearConfig{
					ClientID: "client-id",
					Endpoint: "https://auth.example.com",
				}
				err := cfg.Validate(LoadConfigOptions{ServePortal: true})
				So(err, ShouldBeNil)
			})
		})

		Convey("siteadmin mode", func() {
			Convey("requires SITEADMIN_AUTHGEAR_APP_ID and SITEADMIN_AUTHGEAR_ENDPOINT", func() {
				cfg := validBase()
				err := cfg.Validate(LoadConfigOptions{ServeSiteadmin: true})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "SITEADMIN_AUTHGEAR_APP_ID")
				So(err.Error(), ShouldContainSubstring, "SITEADMIN_AUTHGEAR_ENDPOINT")
			})

			Convey("passes when siteadmin fields are set", func() {
				cfg := validBase()
				cfg.SiteadminAuthgear = portalconfig.AuthgearConfig{
					AppID:    "app-id",
					Endpoint: "https://auth.example.com",
				}
				err := cfg.Validate(LoadConfigOptions{ServeSiteadmin: true})
				So(err, ShouldBeNil)
			})
		})

		Convey("portal+siteadmin mode validates both sets of fields", func() {
			cfg := validBase()
			err := cfg.Validate(LoadConfigOptions{ServePortal: true, ServeSiteadmin: true})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "AUTHGEAR_CLIENT_ID")
			So(err.Error(), ShouldContainSubstring, "SITEADMIN_AUTHGEAR_APP_ID")
		})
	})
}
