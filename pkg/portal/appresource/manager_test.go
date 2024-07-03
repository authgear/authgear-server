package appresource_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"

	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	configtest "github.com/authgear/authgear-server/pkg/lib/config/test"
	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/portal/appresource"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

func TestManager(t *testing.T) {
	Convey("ApplyUpdates", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

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
		tutorialService := NewMockTutorialService(ctrl)
		tutorialService.EXPECT().OnUpdateResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
		denoClient := NewMockDenoClient(ctrl)
		denoClient.EXPECT().Check(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
		domainService := NewMockDomainService(ctrl)
		domainService.EXPECT().ListDomains(gomock.Any()).AnyTimes().Return([]*apimodel.Domain{
			{ID: "domain-id", AppID: "app-id", Domain: "test"},
		}, nil)

		portalResMgr := &appresource.Manager{
			Context:            context.Background(),
			AppResourceManager: resMgr,
			AppFS:              appResourceFs,
			AppFeatureConfig:   cfg.FeatureConfig,
			AppHostSuffixes:    &config.AppHostSuffixes{},
			Tutorials:          tutorialService,
			DomainService:      domainService,
			DenoClient:         denoClient,
			Clock:              clock.NewMockClock(),
		}

		applyUpdates := func(updates []appresource.Update) ([]*resource.ResourceFile, error) {
			return portalResMgr.ApplyUpdates(appID, updates)
		}

		func() {
			appConfigYAML, _ := yaml.Marshal(cfg.AppConfig)
			secretConfigYAML, _ := yaml.Marshal(cfg.SecretConfig)
			_ = afero.WriteFile(appFs, "authgear.yaml", appConfigYAML, 0666)
			_ = afero.WriteFile(appFs, "authgear.secrets.yaml", secretConfigYAML, 0666)

			resource.RegisterResource(web.ImageDescriptor{
				Name:      "myimage",
				SizeLimit: 100 * 1024,
			})
		}()

		Convey("validate new config without crash", func() {
			// We do not use updates to create new config.
			_, err := applyUpdates(nil)
			So(err, ShouldBeNil)
		})

		Convey("validate file size", func() {
			Convey("validate file with default size limit", func() {
				_, err := applyUpdates([]appresource.Update{{
					Path: "authgear.yaml",
					Data: []byte("id: " + string(make([]byte, 1024*1024))),
				}})
				So(err, ShouldBeError, `invalid resource 'authgear.yaml': too large (1048580 > 102400)`)
			})

			Convey("validate file with specified size limit", func() {
				_, err := applyUpdates([]appresource.Update{{
					Path: "static/en/myimage.png",
					Data: make([]byte, 500*1024),
				}})
				So(err, ShouldBeError, `invalid resource 'static/en/myimage.png': too large (512000 > 102400)`)
			})
		})

		Convey("validate configuration YAML", func() {
			_, err := applyUpdates([]appresource.Update{{
				Path: "authgear.yaml",
				Data: []byte("{}"),
			}})
			So(err, ShouldBeError, `cannot parse incoming app config: invalid configuration:
<root>: required
  map[actual:<nil> expected:[http id] missing:[http id]]`)

			_, err = applyUpdates([]appresource.Update{{
				Path: "authgear.yaml",
				Data: []byte("id: test\nhttp:\n  public_origin: \"http://test\""),
			}})
			So(err, ShouldBeError, `invalid resource 'authgear.yaml': incorrect app ID`)

		})

		Convey("validate configuration YAML with plan", func() {
			applyUpdatesWithPlan := func(planName configtest.FixturePlanName, updates []appresource.Update) error {
				fc := configtest.FixtureFeatureConfig(planName)
				config.PopulateFeatureConfigDefaultValues(fc)
				portalResMgr.AppFeatureConfig = fc
				_, err := portalResMgr.ApplyUpdates(appID, updates)
				return err
			}

			var err error
			err = applyUpdatesWithPlan(configtest.FixtureLimitedPlanName, []appresource.Update{{
				Path: "authgear.yaml",
				Data: []byte("id: app-id\nhttp:\n  public_origin: http://test\noauth:\n  clients:\n    - name: Test Client\n      client_id: test-client\n      redirect_uris:\n        - \"https://example.com\"\n    - name: Test Client2\n      client_id: test-client2\n      redirect_uris:\n        - \"https://example2.com\""),
			}})
			So(err, ShouldBeError, `invalid authgear.yaml:
/oauth/clients: exceed the maximum number of oauth clients, actual: 2, expected: 1`)
		})

		Convey("allow updating secrets", func() {
			updateSecretConfigInstructions := configtest.FixtureUpdateSecretConfigUpdateInstruction()
			bytes, err := json.Marshal(updateSecretConfigInstructions)
			So(err, ShouldBeNil)

			_, err = applyUpdates([]appresource.Update{{
				Path: "authgear.secrets.yaml",
				Data: bytes,
			}})
			So(err, ShouldBeNil)
		})

		Convey("forbid deleting configuration YAML", func() {
			_, err := applyUpdates([]appresource.Update{{
				Path: "authgear.yaml",
				Data: nil,
			}})
			So(err, ShouldBeError, "cannot delete 'authgear.yaml'")

			_, err = applyUpdates([]appresource.Update{{
				Path: "authgear.secrets.yaml",
				Data: nil,
			}})
			So(err, ShouldBeError, "cannot delete 'authgear.secrets.yaml'")
		})

		Convey("forbid unknown resource files", func() {
			_, err := applyUpdates([]appresource.Update{{
				Path: "unknown.txt",
				Data: nil,
			}})
			So(err, ShouldBeError, `invalid resource 'unknown.txt': unknown resource path`)
		})

		Convey("clean up orphaned resources files", func() {
			_ = appFs.MkdirAll("deno", 0777)
			_ = afero.WriteFile(appFs, "deno/a.ts", []byte("a.ts"), 0666)
			appConfigYAML, _ := yaml.Marshal(cfg.AppConfig)

			files, err := applyUpdates([]appresource.Update{{
				Path: "authgear.yaml",
				Data: appConfigYAML,
			}})
			So(err, ShouldBeNil)
			So(len(files), ShouldEqual, 2)
			So(files[1].Location.Fs.GetFsLevel(), ShouldEqual, resource.FsLevelApp)
			So(files[1].Location.Path, ShouldEqual, "deno/a.ts")
			So(files[1].Data, ShouldEqual, []uint8(nil))
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
