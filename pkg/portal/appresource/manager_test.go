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

// testCertificatePEM is an arbitrary, valid, self-signed certificate used
// wherever a test needs a well-formed X509Certificate PEM and does not care
// about its content.
const testCertificatePEM = "-----BEGIN CERTIFICATE-----\n" +
	"MIIDejCCAmKgAwIBAgIgLKKTB6GZMFHZVUiFIq8LcNIr0p8HFHwKM6r5/BQ/un4w\n" +
	"DQYJKoZIhvcNAQEFBQAwUDEJMAcGA1UEBhMAMQkwBwYDVQQKDAAxCTAHBgNVBAsM\n" +
	"ADENMAsGA1UEAwwEdGVzdDEPMA0GCSqGSIb3DQEJARYAMQ0wCwYDVQQDDAR0ZXN0\n" +
	"MB4XDTI0MDgwODA2NTY0OFoXDTM0MDgwOTA2NTY0OFowQTEJMAcGA1UEBhMAMQkw\n" +
	"BwYDVQQKDAAxCTAHBgNVBAsMADENMAsGA1UEAwwEdGVzdDEPMA0GCSqGSIb3DQEJ\n" +
	"ARYAMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA5zRfTtkaa7cIsQS+\n" +
	"F1Dg25wPEvcjHsHcq598n+RzRJzfSLRtYwgEfs0VhyjHfo2O7KhNFh5cqdkEfzwA\n" +
	"bfxtgVLvy3yUjTMFO0FnJqrO3dkGiOAl654XUlXb4rF8DF1sPnUdd9QEZaZHGV/8\n" +
	"YuVOc3RV15jsr2jB9rra9//guAQ0CSP4XLJ5m9vf9nJILAHLryFIzDSgOVmhi4Ig\n" +
	"o59e9n3Hemavrta2C5Zj4cP6RNwuCV/i5lQOkzJIgksH9/EZCsR93DMEgkBS5oQQ\n" +
	"rt9Bzlr03TNGW4n/CYKNULK/osqJd5r5g3zUaQZY2KAan+oSsEXvBjzYtrehN1dm\n" +
	"dfbUEQIDAQABo08wTTAdBgNVHQ4EFgQUiXG6MG9PSB/clTIuzm8rW+8xLWkwHwYD\n" +
	"VR0jBBgwFoAUiXG6MG9PSB/clTIuzm8rW+8xLWkwCwYDVR0RBAQwAoIAMA0GCSqG\n" +
	"SIb3DQEBBQUAA4IBAQBTjdS9po3eEXukksMK6xBL3kQF1MEFUaWcgoN+h497lS9J\n" +
	"Xe1rmWpdZ1Aehp21GQmniRKU8uPLPRQKoX8Mhc/d3fHyv9u0YPns/2Wm8TBzxwHY\n" +
	"V2KdXZfpBdN+Z5bBRbgtKxx1z2GBfB39S2WCakS9xK8f7fuQPLIZz8eq7so5T8Hm\n" +
	"TU95acndEpnA0u6/MjbvXtZesTRZCewQw4CkcSLTCzB8dLG55UXHytnISWlCpuAx\n" +
	"8svq/ryZIi5vhBQFO/hG9s2Q32VvfKt2ZW8qA+gvOxEVDfAEFekKokP0Taiz77Q2\n" +
	"AVZxEXeABxJGtiMunQTr2q1tCrJQN0d08xlA5jXl\n" +
	"-----END CERTIFICATE-----\n"

func TestManager(t *testing.T) {
	ctx := context.Background()
	Convey("ApplyUpdates0", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		appID := "app-id"
		// This config is supposed to be not effective because it is a fixture
		cfg := &config.Config{
			AppConfig:     configtest.FixtureAppConfig("app-id"),
			SecretConfig:  configtest.FixtureSecretConfig(0),
			FeatureConfig: configtest.FixtureFeatureConfig(ctx, configtest.FixtureLimitedPlanName),
		}

		baseFs := afero.NewMemMapFs()
		appFs := afero.NewMemMapFs()
		baseResourceFs := &resource.LeveledAferoFs{Fs: baseFs, FsLevel: resource.FsLevelBuiltin}
		appResourceFs := &resource.LeveledAferoFs{Fs: appFs, FsLevel: resource.FsLevelApp}
		resMgr := resource.NewManager(resource.DefaultRegistry, []resource.Fs{
			baseResourceFs,
			appResourceFs,
		})
		tutorialService := NewMockTutorialService(ctrl)
		tutorialService.EXPECT().OnUpdateResource0(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
		denoClient := NewMockDenoClient(ctrl)
		denoClient.EXPECT().Check(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
		domainService := NewMockDomainService(ctrl)
		domainService.EXPECT().ListDomains(gomock.Any(), gomock.Any()).AnyTimes().Return([]*apimodel.Domain{
			{ID: "domain-id", AppID: "app-id", Domain: "test"},
		}, nil)

		portalResMgr := &appresource.Manager{
			AppResourceManager: resMgr,
			AppFS:              appResourceFs,
			AppFeatureConfig:   cfg.FeatureConfig,
			AppHostSuffixes:    &config.AppHostSuffixes{},
			Tutorials:          tutorialService,
			DomainService:      domainService,
			DenoClient:         denoClient,
			Clock:              clock.NewMockClock(),
		}

		applyUpdates := func(ctx context.Context, updates []appresource.Update) ([]*resource.ResourceFile, error) {
			return portalResMgr.ApplyUpdates0(ctx, appID, updates)
		}

		func() {
			appConfigYAML, _ := yaml.Marshal(cfg.AppConfig)
			secretConfigYAML, _ := yaml.Marshal(cfg.SecretConfig)
			_ = afero.WriteFile(appFs, "authgear.yaml", appConfigYAML, 0666)
			_ = afero.WriteFile(appFs, "authgear.secrets.yaml", secretConfigYAML, 0666)

			resource.RegisterResource(web.LocaleAwareImageDescriptor{
				Name:      "myimage",
				SizeLimit: 100 * 1024,
			})
		}()

		Convey("validate new config without crash", func() {
			// We do not use updates to create new config.
			ctx := context.Background()
			_, err := applyUpdates(ctx, nil)
			So(err, ShouldBeNil)
		})

		Convey("validate file size", func() {
			Convey("validate file with default size limit", func() {
				ctx := context.Background()
				_, err := applyUpdates(ctx, []appresource.Update{{
					Path: "authgear.yaml",
					Data: []byte("id: " + string(make([]byte, 1024*1024))),
				}})
				So(err, ShouldBeError, `invalid resource 'authgear.yaml': too large (1048580 > 102400)`)
			})

			Convey("validate file with specified size limit", func() {
				ctx := context.Background()
				_, err := applyUpdates(ctx, []appresource.Update{{
					Path: "static/en/myimage.png",
					Data: make([]byte, 500*1024),
				}})
				So(err, ShouldBeError, `invalid resource 'static/en/myimage.png': too large (512000 > 102400)`)
			})
		})

		Convey("validate configuration YAML", func() {
			ctx := context.Background()
			_, err := applyUpdates(ctx, []appresource.Update{{
				Path: "authgear.yaml",
				Data: []byte("{}"),
			}})
			So(err, ShouldBeError, `cannot parse incoming app config: invalid configuration:
<root>: required
  map[actual:<nil> expected:[http id] missing:[http id]]`)

			_, err = applyUpdates(ctx, []appresource.Update{{
				Path: "authgear.yaml",
				Data: []byte("id: test\nhttp:\n  public_origin: \"http://test\""),
			}})
			So(err, ShouldBeError, `invalid resource 'authgear.yaml': incorrect app ID`)

		})

		Convey("validate configuration YAML with plan", func() {
			applyUpdatesWithPlan := func(planName configtest.FixturePlanName, updates []appresource.Update) error {
				fc := configtest.FixtureFeatureConfig(ctx, planName)
				config.PopulateFeatureConfigDefaultValues(fc)
				portalResMgr.AppFeatureConfig = fc
				ctx := context.Background()
				_, err := portalResMgr.ApplyUpdates0(ctx, appID, updates)
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

			ctx := context.Background()
			_, err = applyUpdates(ctx, []appresource.Update{{
				Path: "authgear.secrets.yaml",
				Data: bytes,
			}})
			So(err, ShouldBeNil)
		})

		Convey("forbid deleting configuration YAML", func() {
			ctx := context.Background()
			_, err := applyUpdates(ctx, []appresource.Update{{
				Path: "authgear.yaml",
				Data: nil,
			}})
			So(err, ShouldBeError, "cannot delete 'authgear.yaml'")

			_, err = applyUpdates(ctx, []appresource.Update{{
				Path: "authgear.secrets.yaml",
				Data: nil,
			}})
			So(err, ShouldBeError, "cannot delete 'authgear.secrets.yaml'")
		})

		Convey("forbid unknown resource files", func() {
			ctx := context.Background()
			_, err := applyUpdates(ctx, []appresource.Update{{
				Path: "unknown.txt",
				Data: nil,
			}})
			So(err, ShouldBeError, `invalid resource 'unknown.txt': unknown resource path`)
		})

		Convey("clean up orphaned resources files", func() {
			_ = appFs.MkdirAll("deno", 0777)
			_ = afero.WriteFile(appFs, "deno/a.ts", []byte("a.ts"), 0666)
			appConfigYAML, _ := yaml.Marshal(cfg.AppConfig)

			ctx := context.Background()
			files, err := applyUpdates(ctx, []appresource.Update{{
				Path: "authgear.yaml",
				Data: appConfigYAML,
			}})
			So(err, ShouldBeNil)
			So(len(files), ShouldEqual, 2)
			So(files[1].Location.Fs.GetFsLevel(), ShouldEqual, resource.FsLevelApp)
			So(files[1].Location.Path, ShouldEqual, "deno/a.ts")
			So(files[1].Data, ShouldEqual, []uint8(nil))
		})

		Convey("clean up orphaned audit log stream tls secrets", func() {
			telemetryStream := func(name string, tlsEnabled bool) *config.TelemetryAuditLogStreamConfig {
				return &config.TelemetryAuditLogStreamConfig{
					Name:      name,
					Type:      config.TelemetryAuditLogStreamTypeSyslog,
					Transport: config.TelemetryAuditLogStreamTransportTCP,
					TCP: &config.TelemetryAuditLogStreamTCPConfig{
						Address: "collector.internal:5140",
						TLS:     &config.TelemetryAuditLogStreamTCPTLSConfig{Enabled: tlsEnabled},
					},
					Syslog: &config.TelemetryAuditLogStreamSyslogConfig{
						Format:  config.SyslogFormatRFC5424,
						Framing: config.SyslogFramingNewline,
					},
				}
			}

			// appConfigYAMLWithStreams marshals cfg.AppConfig with the given
			// streams (or none, when streams is nil) as the "incoming"
			// authgear.yaml of an update.
			appConfigYAMLWithStreams := func(streams []*config.TelemetryAuditLogStreamConfig) []byte {
				appConfig := *cfg.AppConfig
				if streams != nil {
					appConfig.Telemetry = &config.TelemetryConfig{
						AuditLogs: &config.TelemetryAuditLogsConfig{Streams: streams},
					}
				}
				data, err := yaml.Marshal(&appConfig)
				So(err, ShouldBeNil)
				return data
			}

			// writeSecretsWithMaterials writes the current authgear.secrets.yaml
			// on disk, with a telemetry.audit_logs.streams.tls item per name.
			writeSecretsWithMaterials := func(streamNames ...string) {
				var materials config.TelemetryAuditLogStreamTLSMaterials
				for _, name := range streamNames {
					materials = append(materials, config.TelemetryAuditLogStreamTLSMaterialsItem{
						StreamName: name,
						CertificateAuthority: &config.X509Certificate{
							Pem: config.X509CertificatePem(testCertificatePEM),
						},
					})
				}
				materialsJSON, err := json.Marshal(materials)
				So(err, ShouldBeNil)

				secretConfig := *cfg.SecretConfig
				secretConfig.Secrets = append(append([]config.SecretItem{}, secretConfig.Secrets...), config.SecretItem{
					Key:     config.TelemetryAuditLogStreamTLSMaterialsKey,
					RawData: json.RawMessage(materialsJSON),
				})
				data, err := yaml.Marshal(&secretConfig)
				So(err, ShouldBeNil)
				_ = afero.WriteFile(appFs, "authgear.secrets.yaml", data, 0666)
			}

			readBackMaterials := func(data []byte) config.TelemetryAuditLogStreamTLSMaterials {
				secretConfig, err := config.ParsePartialSecret(context.Background(), data)
				So(err, ShouldBeNil)
				materials, ok := secretConfig.LookupData(config.TelemetryAuditLogStreamTLSMaterialsKey).(*config.TelemetryAuditLogStreamTLSMaterials)
				if !ok {
					return nil
				}
				return *materials
			}

			Convey("prunes the item of a removed stream and leaves other secrets untouched", func() {
				writeSecretsWithMaterials("kept", "removed")

				ctx := context.Background()
				files, err := applyUpdates(ctx, []appresource.Update{{
					Path: "authgear.yaml",
					Data: appConfigYAMLWithStreams([]*config.TelemetryAuditLogStreamConfig{telemetryStream("kept", true)}),
				}})
				So(err, ShouldBeNil)
				So(len(files), ShouldEqual, 2)
				So(files[1].Location.Path, ShouldEqual, "authgear.secrets.yaml")

				materials := readBackMaterials(files[1].Data)
				So(materials, ShouldHaveLength, 1)
				So(materials[0].StreamName, ShouldEqual, "kept")

				// Every other secret key survives untouched.
				secretConfig, err := config.ParsePartialSecret(context.Background(), files[1].Data)
				So(err, ShouldBeNil)
				_, _, ok := secretConfig.Lookup(config.DatabaseCredentialsKey)
				So(ok, ShouldBeTrue)
			})

			Convey("keeps the item of a stream that still exists with tls disabled", func() {
				writeSecretsWithMaterials("kept")

				ctx := context.Background()
				files, err := applyUpdates(ctx, []appresource.Update{{
					Path: "authgear.yaml",
					Data: appConfigYAMLWithStreams([]*config.TelemetryAuditLogStreamConfig{telemetryStream("kept", false)}),
				}})
				So(err, ShouldBeNil)
				// Nothing was orphaned, so no secrets file is rewritten.
				So(len(files), ShouldEqual, 1)
			})

			Convey("removes the whole secret entry when the last stream is removed", func() {
				writeSecretsWithMaterials("removed")

				ctx := context.Background()
				files, err := applyUpdates(ctx, []appresource.Update{{
					Path: "authgear.yaml",
					Data: appConfigYAMLWithStreams(nil),
				}})
				So(err, ShouldBeNil)
				So(len(files), ShouldEqual, 2)

				secretConfig, err := config.ParsePartialSecret(context.Background(), files[1].Data)
				So(err, ShouldBeNil)
				_, _, found := secretConfig.Lookup(config.TelemetryAuditLogStreamTLSMaterialsKey)
				So(found, ShouldBeFalse)
			})

			Convey("an update set without an authgear.yaml change prunes nothing", func() {
				writeSecretsWithMaterials("removed")

				updateSecretConfigInstructions := configtest.FixtureUpdateSecretConfigUpdateInstruction()
				instructionBytes, err := json.Marshal(updateSecretConfigInstructions)
				So(err, ShouldBeNil)

				ctx := context.Background()
				files, err := applyUpdates(ctx, []appresource.Update{{
					Path: "authgear.secrets.yaml",
					Data: instructionBytes,
				}})
				So(err, ShouldBeNil)
				So(len(files), ShouldEqual, 1)

				materials := readBackMaterials(files[0].Data)
				So(materials, ShouldHaveLength, 1)
				So(materials[0].StreamName, ShouldEqual, "removed")
			})

			Convey("an authgear.yaml update with no orphans returns no secrets file", func() {
				writeSecretsWithMaterials("kept")

				ctx := context.Background()
				files, err := applyUpdates(ctx, []appresource.Update{{
					Path: "authgear.yaml",
					Data: appConfigYAMLWithStreams([]*config.TelemetryAuditLogStreamConfig{telemetryStream("kept", true)}),
				}})
				So(err, ShouldBeNil)
				So(len(files), ShouldEqual, 1)
			})
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
