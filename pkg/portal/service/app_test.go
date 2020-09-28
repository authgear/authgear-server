package service

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	configtest "github.com/authgear/authgear-server/pkg/lib/config/test"
	"github.com/authgear/authgear-server/pkg/portal/model"
)

func TestValidateConfig(t *testing.T) {
	Convey("ValidateConfig", t, func() {
		appID := "app-id"
		cfg := &config.Config{
			AppConfig:    configtest.FixtureAppConfig("app-id"),
			SecretConfig: configtest.FixtureSecretConfig(0),
		}
		config.PopulateDefaultValues(cfg.AppConfig)

		Convey("accept empty updates", func() {
			err := ValidateConfig(appID, *cfg, nil, nil)
			So(err, ShouldBeNil)
		})

		Convey("validate file size", func() {
			updateFiles := []*model.AppConfigFile{{
				Path:    "/authgear.yaml",
				Content: "id: " + string(make([]byte, 1024*1024)),
			}}
			err := ValidateConfig(appID, *cfg, updateFiles, nil)
			So(err, ShouldBeError, `/authgear.yaml is too large: 1048580 > 102400`)
		})

		Convey("validate configuration YAML", func() {
			updateFiles := []*model.AppConfigFile{{
				Path:    "/authgear.yaml",
				Content: `{}`,
			}}
			err := ValidateConfig(appID, *cfg, updateFiles, nil)
			So(err, ShouldBeError, `/authgear.yaml is invalid: invalid configuration:
<root>: required
  map[actual:<nil> expected:[id] missing:[id]]`)

			updateFiles = []*model.AppConfigFile{{
				Path:    "/authgear.yaml",
				Content: `id: test`,
			}}
			err = ValidateConfig(appID, *cfg, updateFiles, nil)
			So(err, ShouldBeError, `/authgear.yaml is invalid: invalid app ID`)

			updateFiles = []*model.AppConfigFile{{
				Path:    "/authgear.secrets.yaml",
				Content: `{}`,
			}}
			err = ValidateConfig(appID, *cfg, updateFiles, nil)
			So(err, ShouldBeError, `/authgear.secrets.yaml is invalid: invalid secrets:
<root>: required
  map[actual:<nil> expected:[secrets] missing:[secrets]]`)

			updateFiles = []*model.AppConfigFile{{
				Path:    "/authgear.secrets.yaml",
				Content: `secrets: []`,
			}}
			err = ValidateConfig(appID, *cfg, updateFiles, nil)
			So(err.Error(), ShouldStartWith, `invalid configuration: invalid secrets`)
		})

		Convey("forbid deleting configuration YAML", func() {
			deleteFiles := []string{"/authgear.yaml"}
			err := ValidateConfig(appID, *cfg, nil, deleteFiles)
			So(err, ShouldBeError, "cannot delete main configuration YAML files")

			deleteFiles = []string{"/authgear.secrets.yaml"}
			err = ValidateConfig(appID, *cfg, nil, deleteFiles)
			So(err, ShouldBeError, "cannot delete main configuration YAML files")
		})

		Convey("allow mutating template files", func() {
			cfg.AppConfig.Template.Items = []config.TemplateItem{{
				Type: "login",
				URI:  "file:///templates/login.html",
			}}

			deleteFiles := []string{"/templates/login.html"}
			err := ValidateConfig(appID, *cfg, nil, deleteFiles)
			So(err, ShouldBeNil)

			updateFiles := []*model.AppConfigFile{{
				Path:    "/templates/login.html",
				Content: "Login",
			}}
			err = ValidateConfig(appID, *cfg, updateFiles, nil)
			So(err, ShouldBeNil)
		})

		Convey("forbid mutating irrelevant files", func() {
			deleteFiles := []string{"/foobar"}
			err := ValidateConfig(appID, *cfg, nil, deleteFiles)
			So(err, ShouldBeError, "invalid file '/foobar': file is not referenced from configuration")

			updateFiles := []*model.AppConfigFile{{
				Path:    "/foobar",
				Content: "what",
			}}
			err = ValidateConfig(appID, *cfg, updateFiles, nil)
			So(err, ShouldBeError, "invalid file '/foobar': file is not referenced from configuration")
		})
	})
}
