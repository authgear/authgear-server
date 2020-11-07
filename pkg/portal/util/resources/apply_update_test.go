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

func TestApplyUpdates(t *testing.T) {
	Convey("ApplyUpdates", t, func() {
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
		allowlist := []string{
			"admin-api.auth",
			"webhook",
			"sso.oauth.client",
		}
		applyUpdates := func(updates []resources.Update) error {
			_, err := resources.ApplyUpdates(appID, appResourceFs, resMgr, allowlist, updates)
			return err
		}

		func() {
			appConfigYAML, _ := yaml.Marshal(cfg.AppConfig)
			secretConfigYAML, _ := yaml.Marshal(cfg.SecretConfig)
			_ = afero.WriteFile(appFs, "authgear.yaml", appConfigYAML, 0666)
			_ = afero.WriteFile(appFs, "authgear.secrets.yaml", secretConfigYAML, 0666)
		}()

		Convey("validate new config without crash", func() {
			// We do not use updates to create new config.
			err := applyUpdates(nil)
			So(err, ShouldBeNil)
		})

		Convey("validate file size", func() {
			err := applyUpdates([]resources.Update{{
				Path: "authgear.yaml",
				Data: []byte("id: " + string(make([]byte, 1024*1024))),
			}})
			So(err, ShouldBeError, `invalid resource 'authgear.yaml': too large (1048580 > 102400)`)
		})

		Convey("validate configuration YAML", func() {
			err := applyUpdates([]resources.Update{{
				Path: "authgear.yaml",
				Data: []byte("{}"),
			}})
			So(err, ShouldBeError, `invalid resource: cannot parse app config: invalid configuration:
<root>: required
  map[actual:<nil> expected:[http id] missing:[http id]]`)

			err = applyUpdates([]resources.Update{{
				Path: "authgear.yaml",
				Data: []byte("id: test\nhttp:\n  public_origin: \"http://test\""),
			}})
			So(err, ShouldBeError, `invalid resource 'authgear.yaml': incorrect app ID`)

		})

		Convey("forbid deleting required items in secrets", func() {
			err := applyUpdates([]resources.Update{{
				Path: "authgear.secrets.yaml",
				Data: []byte("{}"),
			}})
			So(err, ShouldBeError, `invalid secret config: invalid secrets:
<root>: admin API auth key materials (secret 'admin-api.auth') is required`)

		})

		Convey("forbid updating secrets no in the allowlist", func() {
			newSecretConfig := configtest.FixtureSecretConfig(1)
			bytes, err := yaml.Marshal(newSecretConfig)
			So(err, ShouldBeNil)

			err = applyUpdates([]resources.Update{{
				Path: "authgear.secrets.yaml",
				Data: bytes,
			}})
			So(err, ShouldBeError, "'db' in secret config is not allowed")
		})

		Convey("allow updating secrets", func() {
			newSecretConfig := configtest.FixtureSecretConfig(1)

			// Remove keys that are not in the allowlist
			allowmap := make(map[string]struct{})
			for _, key := range allowlist {
				allowmap[key] = struct{}{}
			}
			var secrets []config.SecretItem
			for _, secretItem := range newSecretConfig.Secrets {
				_, allowed := allowmap[string(secretItem.Key)]
				if allowed {
					secrets = append(secrets, secretItem)
				}
			}
			newSecretConfig.Secrets = secrets

			bytes, err := yaml.Marshal(newSecretConfig)
			So(err, ShouldBeNil)

			err = applyUpdates([]resources.Update{{
				Path: "authgear.secrets.yaml",
				Data: bytes,
			}})
			So(err, ShouldBeNil)
		})

		Convey("forbid deleting configuration YAML", func() {
			err := applyUpdates([]resources.Update{{
				Path: "authgear.yaml",
				Data: nil,
			}})
			So(err, ShouldBeError, "cannot delete 'authgear.yaml'")

			err = applyUpdates([]resources.Update{{
				Path: "authgear.secrets.yaml",
				Data: nil,
			}})
			So(err, ShouldBeError, "cannot delete 'authgear.secrets.yaml'")
		})

		Convey("forbid unknown resource files", func() {
			err := applyUpdates([]resources.Update{{
				Path: "unknown.txt",
				Data: nil,
			}})
			So(err, ShouldBeError, `invalid resource 'unknown.txt': unknown resource path`)
		})
	})
}
