package appresource_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
	configtest "github.com/authgear/authgear-server/pkg/lib/config/test"
	"github.com/authgear/authgear-server/pkg/portal/appresource"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

func TestManager(t *testing.T) {
	Convey("ApplyUpdates", t, func() {
		appID := "app-id"
		cfg := &config.Config{
			AppConfig:     configtest.FixtureAppConfig("app-id"),
			SecretConfig:  configtest.FixtureSecretConfig(0),
			FeatureConfig: configtest.FixtureFeatureConfig(configtest.FixtureLimitedPlanName),
		}
		config.PopulateDefaultValues(cfg.AppConfig)

		baseFs := afero.NewMemMapFs()
		appFs := afero.NewMemMapFs()
		baseResourceFs := &resource.LeveledAferoFs{Fs: baseFs, FsLevel: resource.FsLevelBuiltin}
		appResourceFs := &resource.LeveledAferoFs{Fs: appFs, FsLevel: resource.FsLevelApp}
		resMgr := resource.NewManager(resource.DefaultRegistry, []resource.Fs{
			baseResourceFs,
			appResourceFs,
		})
		allowlist := []string{
			"admin-api.auth",
			"webhook",
			"sso.oauth.client",
		}

		portalResMgr := &appresource.Manager{
			AppResourceManager: resMgr,
			AppFS:              appResourceFs,
			SecretKeyAllowlist: allowlist,
			AppFeatureConfig:   cfg.FeatureConfig,
		}

		applyUpdates := func(updates []appresource.Update) error {
			_, err := portalResMgr.ApplyUpdates(appID, updates)
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
			err := applyUpdates([]appresource.Update{{
				Path: "authgear.yaml",
				Data: []byte("id: " + string(make([]byte, 1024*1024))),
			}})
			So(err, ShouldBeError, `invalid resource 'authgear.yaml': too large (1048580 > 102400)`)
		})

		Convey("validate configuration YAML", func() {
			err := applyUpdates([]appresource.Update{{
				Path: "authgear.yaml",
				Data: []byte("{}"),
			}})
			So(err, ShouldBeError, `cannot parse incoming app config: invalid configuration:
<root>: required
  map[actual:<nil> expected:[http id] missing:[http id]]`)

			err = applyUpdates([]appresource.Update{{
				Path: "authgear.yaml",
				Data: []byte("id: test\nhttp:\n  public_origin: \"http://test\""),
			}})
			So(err, ShouldBeError, `invalid resource 'authgear.yaml': incorrect app ID`)

			err = applyUpdates([]appresource.Update{{
				Path: "authgear.yaml",
				Data: []byte("id: app-id\nhttp:\n  public_origin: http://test\noauth:\n  clients:\n    - name: Test Client\n      client_id: test-client\n      redirect_uris:\n        - \"https://example.com\"\n    - name: Test Client2\n      client_id: test-client2\n      redirect_uris:\n        - \"https://example2.com\""),
			}})
			So(err, ShouldBeError, `exceed the maximum number of oauth clients, actual: 2, expected: 1`)
		})

		Convey("forbid deleting required items in secrets", func() {
			err := applyUpdates([]appresource.Update{{
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

			err = applyUpdates([]appresource.Update{{
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

			err = applyUpdates([]appresource.Update{{
				Path: "authgear.secrets.yaml",
				Data: bytes,
			}})
			So(err, ShouldBeNil)
		})

		Convey("forbid deleting configuration YAML", func() {
			err := applyUpdates([]appresource.Update{{
				Path: "authgear.yaml",
				Data: nil,
			}})
			So(err, ShouldBeError, "cannot delete 'authgear.yaml'")

			err = applyUpdates([]appresource.Update{{
				Path: "authgear.secrets.yaml",
				Data: nil,
			}})
			So(err, ShouldBeError, "cannot delete 'authgear.secrets.yaml'")
		})

		Convey("forbid unknown resource files", func() {
			err := applyUpdates([]appresource.Update{{
				Path: "unknown.txt",
				Data: nil,
			}})
			So(err, ShouldBeError, `invalid resource 'unknown.txt': unknown resource path`)
		})
	})

	Convey("List", t, func() {
		reg := &resource.Registry{}
		fsA := afero.NewMemMapFs()
		fsB := afero.NewMemMapFs()
		res := resource.NewManager(reg, []resource.Fs{
			&resource.LeveledAferoFs{Fs: fsA, FsLevel: resource.FsLevelBuiltin},
			&resource.LeveledAferoFs{Fs: fsB, FsLevel: resource.FsLevelApp},
		})
		portalResMgr := &appresource.Manager{
			AppResourceManager: res,
			AppFS:              &resource.LeveledAferoFs{Fs: fsB, FsLevel: resource.FsLevelApp},
		}

		reg.Register(resource.SimpleDescriptor{Path: "test/a/x.txt"})
		reg.Register(resource.SimpleDescriptor{Path: "test/b/z.txt"})
		reg.Register(resource.SimpleDescriptor{Path: "test/x.txt"})
		reg.Register(resource.SimpleDescriptor{Path: "w.txt"})

		_ = fsA.MkdirAll("test/a", 0666)
		_ = fsA.MkdirAll("test/b", 0666)
		_ = fsB.MkdirAll("test/a", 0666)
		_ = afero.WriteFile(fsA, "test/a/x.txt", nil, 0666)
		_ = afero.WriteFile(fsA, "test/a/y.txt", nil, 0666)
		_ = afero.WriteFile(fsA, "test/b/z.txt", nil, 0666)
		_ = afero.WriteFile(fsB, "test/x.txt", nil, 0666)
		_ = afero.WriteFile(fsB, "test/b/z.txt", nil, 0666)
		_ = afero.WriteFile(fsB, "w.txt", nil, 0666)

		paths, err := portalResMgr.List()
		So(err, ShouldBeNil)
		So(paths, ShouldResemble, []string{
			"test/a/x.txt",
			"test/b/z.txt",
			"test/x.txt",
			"w.txt",
		})
	})

}
