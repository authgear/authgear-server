package service

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	configtest "github.com/authgear/authgear-server/pkg/lib/config/test"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
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

		Convey("validate new config without crash", func() {
			appConfigYAML, err := yaml.Marshal(cfg.AppConfig)
			So(err, ShouldBeNil)
			secretConfigYAML, err := yaml.Marshal(cfg.SecretConfig)
			So(err, ShouldBeNil)

			err = ValidateConfig(appID, config.Config{}, []*model.AppConfigFile{
				{Path: "/" + configsource.AuthgearYAML, Content: string(appConfigYAML)},
				{Path: "/" + configsource.AuthgearSecretYAML, Content: string(secretConfigYAML)},
			}, nil)
			So(err, ShouldBeNil)
		})

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

func TestRedaction(t *testing.T) {
	Convey("redactSecrets add back removed secrets", t, func() {
		svc := &AppService{
			AppConfig: &portalconfig.AppConfig{
				Secret: portalconfig.AppSecretConfig{
					DatabaseURL:      "secret",
					DatabaseSchema:   "secret",
					RedisURL:         "secret",
					SMTPHost:         "secret",
					SMTPPort:         1,
					SMTPMode:         "secret",
					SMTPUsername:     "secret",
					SMTPPassword:     "secret",
					TwilioAccountSID: "secret",
					TwilioAuthToken:  "secret",
					NexmoAPIKey:      "",
					NexmoAPISecret:   "",
				},
			},
		}
		cfg := &config.SecretConfig{}

		err := svc.redactSecrets(cfg)
		So(err, ShouldBeNil)

		bytes, err := yaml.Marshal(cfg)
		So(err, ShouldBeNil)

		So(string(bytes), ShouldEqual, `secrets:
- data:
    database_schema: <REDACTED>
    database_url: <REDACTED>
  key: db
- data:
    redis_url: <REDACTED>
  key: redis
- data:
    host: <REDACTED>
    password: <REDACTED>
    username: <REDACTED>
  key: mail.smtp
- data:
    account_sid: <REDACTED>
    auth_token: <REDACTED>
  key: sms.twilio
`)
	})
}
