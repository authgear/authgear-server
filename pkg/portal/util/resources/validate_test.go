package resources_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
	configtest "github.com/authgear/authgear-server/pkg/lib/config/test"
	"github.com/authgear/authgear-server/pkg/portal/util/resources"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

func TestValidate(t *testing.T) {
	Convey("Validate", t, func() {
		appID := "app-id"
		cfg := &config.Config{
			AppConfig:    configtest.FixtureAppConfig("app-id"),
			SecretConfig: configtest.FixtureSecretConfig(0),
		}
		config.PopulateDefaultValues(cfg.AppConfig)

		baseFs := afero.NewMemMapFs()
		appFs := afero.NewMemMapFs()
		baseResourceFs := &resource.AferoFs{Fs: baseFs}
		appResourceFs := &resource.AferoFs{Fs: appFs}
		resMgr := resource.NewManager(resource.DefaultRegistry, []resource.Fs{
			baseResourceFs,
			appResourceFs,
		})
		validate := func(updates []resources.Update) error {
			return resources.Validate(appID, appResourceFs, resMgr, updates)
		}

		func() {
			appConfigYAML, _ := yaml.Marshal(cfg.AppConfig)
			secretConfigYAML, _ := yaml.Marshal(cfg.SecretConfig)
			_ = afero.WriteFile(appFs, "authgear.yaml", appConfigYAML, 0666)
			_ = afero.WriteFile(appFs, "authgear.secrets.yaml", secretConfigYAML, 0666)
		}()

		Convey("validate new config without crash", func() {
			appConfigYAML, err := yaml.Marshal(cfg.AppConfig)
			So(err, ShouldBeNil)
			secretConfigYAML, err := yaml.Marshal(cfg.SecretConfig)
			So(err, ShouldBeNil)

			err = validate([]resources.Update{
				{Path: "authgear.yaml", Data: appConfigYAML},
				{Path: "authgear.secrets.yaml", Data: secretConfigYAML},
			})
			So(err, ShouldBeNil)
		})

		Convey("accept empty updates", func() {
			err := validate(nil)
			So(err, ShouldBeNil)
		})

		Convey("validate file size", func() {
			err := validate([]resources.Update{{
				Path: "authgear.yaml",
				Data: []byte("id: " + string(make([]byte, 1024*1024))),
			}})
			So(err, ShouldBeError, `invalid resource 'authgear.yaml': too large (1048580 > 102400)`)
		})

		Convey("validate configuration YAML", func() {
			err := validate([]resources.Update{{
				Path: "authgear.yaml",
				Data: []byte("{}"),
			}})
			So(err, ShouldBeError, `invalid resource 'authgear.yaml': cannot parse app config: invalid configuration:
<root>: required
  map[actual:<nil> expected:[http id] missing:[http id]]`)

			err = validate([]resources.Update{{
				Path: "authgear.yaml",
				Data: []byte("id: test\nhttp:\n  public_origin: \"http://test\""),
			}})
			So(err, ShouldBeError, `invalid resource 'authgear.yaml': incorrect app ID`)

			err = validate([]resources.Update{{
				Path: "authgear.secrets.yaml",
				Data: []byte("{}"),
			}})
			So(err, ShouldBeError, `invalid resource 'authgear.secrets.yaml': cannot parse secret config: invalid secrets:
<root>: required
  map[actual:<nil> expected:[secrets] missing:[secrets]]`)
		})

		Convey("forbid deleting configuration YAML", func() {
			err := validate([]resources.Update{{
				Path: "authgear.yaml",
				Data: nil,
			}})
			So(err, ShouldBeError, "missing 'authgear.yaml': specified resource is not configured")

			err = validate([]resources.Update{{
				Path: "authgear.secrets.yaml",
				Data: nil,
			}})
			So(err, ShouldBeError, "missing 'authgear.secrets.yaml': specified resource is not configured")
		})

		Convey("forbid unknown resource files", func() {
			err := validate([]resources.Update{{
				Path: "unknown.txt",
				Data: nil,
			}})
			So(err, ShouldBeError, `invalid resource 'unknown.txt': unknown resource path`)
		})

		Convey("forbid overriding base secrets", func() {
			secretConfigYAML, _ := yaml.Marshal(cfg.SecretConfig)
			_ = afero.WriteFile(baseFs, "authgear.secrets.yaml", secretConfigYAML, 0666)

			err := validate([]resources.Update{{
				Path: "authgear.secrets.yaml",
				Data: []byte("secrets: [{data: {redis_url: redis://localhost}, key: redis}]"),
			}})
			So(err, ShouldBeError, `invalid resource 'authgear.secrets.yaml': cannot override secret 'redis' defined in base config`)
		})
	})
}
