package config_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	goyaml "go.yaml.in/yaml/v2"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/adfs"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/apple"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/azureadb2c"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/azureadv2"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/facebook"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/google"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/linkedin"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/wechat"
)

func TestApplyFeatureConfigConstraints(t *testing.T) {
	Convey("ApplyFeatureConfigConstraints", t, func() {
		Convey("phone input allowlist is intersected with feature allowlist", func() {
			appConfig := &config.AppConfig{
				UI: &config.UIConfig{
					PhoneInput: &config.PhoneInputConfig{
						AllowList: []string{"SG", "MY", "TH"},
					},
				},
			}
			featureConfig := &config.FeatureConfig{
				UI: &config.UIFeatureConfig{
					PhoneInput: &config.PhoneInputFeatureConfig{
						AllowList: []string{"SG", "MY"},
					},
				},
			}
			config.ApplyFeatureConfigConstraints(appConfig, featureConfig)
			So(appConfig.UI.PhoneInput.AllowList, ShouldResemble, []string{"SG", "MY"})
		})

		Convey("phone input pinned list is intersected with feature allowlist", func() {
			appConfig := &config.AppConfig{
				UI: &config.UIConfig{
					PhoneInput: &config.PhoneInputConfig{
						PinnedList: []string{"SG", "MY"},
					},
				},
			}
			featureConfig := &config.FeatureConfig{
				UI: &config.UIFeatureConfig{
					PhoneInput: &config.PhoneInputFeatureConfig{
						AllowList: []string{"SG"},
					},
				},
			}
			config.ApplyFeatureConfigConstraints(appConfig, featureConfig)
			So(appConfig.UI.PhoneInput.PinnedList, ShouldResemble, []string{"SG"})
		})

		Convey("nil phone input allowlist is left untouched", func() {
			appConfig := &config.AppConfig{
				UI: &config.UIConfig{
					PhoneInput: &config.PhoneInputConfig{
						AllowList: nil,
					},
				},
			}
			featureConfig := &config.FeatureConfig{
				UI: &config.UIFeatureConfig{
					PhoneInput: &config.PhoneInputFeatureConfig{
						AllowList: []string{"SG"},
					},
				},
			}
			config.ApplyFeatureConfigConstraints(appConfig, featureConfig)
			So(appConfig.UI.PhoneInput.AllowList, ShouldBeNil)
		})

		Convey("does not panic when phone input config is absent", func() {
			So(func() {
				config.ApplyFeatureConfigConstraints(&config.AppConfig{}, &config.FeatureConfig{})
			}, ShouldNotPanic)
		})

		Convey("fraud protection is reset to defaults when IsModifiable is false", func() {
			customFP := &config.FraudProtectionConfig{
				Enabled:  new(false),
				Warnings: nil,
			}
			appConfig := &config.AppConfig{FraudProtection: customFP}
			featureConfig := &config.FeatureConfig{
				FraudProtection: &config.FraudProtectionFeatureConfig{
					IsModifiable: new(false),
				},
			}

			config.ApplyFeatureConfigConstraints(appConfig, featureConfig)

			// Reset to defaults: Enabled=true, all 5 warning types.
			So(*appConfig.FraudProtection.Enabled, ShouldBeTrue)
			So(len(appConfig.FraudProtection.Warnings), ShouldEqual, 5)
			So(*appConfig.FraudProtection.SMS.UnverifiedOTPBudget.DailyRatio, ShouldEqual, 0.3)
			So(*appConfig.FraudProtection.SMS.UnverifiedOTPBudget.HourlyRatio, ShouldEqual, 0.2)
		})

		Convey("fraud protection is left unchanged when IsModifiable is true", func() {
			customFP := &config.FraudProtectionConfig{
				Enabled:  new(false),
				Warnings: nil,
			}
			appConfig := &config.AppConfig{FraudProtection: customFP}
			featureConfig := &config.FeatureConfig{
				FraudProtection: &config.FraudProtectionFeatureConfig{
					IsModifiable: new(true),
				},
			}

			config.ApplyFeatureConfigConstraints(appConfig, featureConfig)

			So(*appConfig.FraudProtection.Enabled, ShouldBeFalse)
		})

		Convey("fraud protection is left unchanged when feature config has no fraud protection entry", func() {
			customFP := &config.FraudProtectionConfig{Enabled: new(false)}
			appConfig := &config.AppConfig{FraudProtection: customFP}

			config.ApplyFeatureConfigConstraints(appConfig, &config.FeatureConfig{})

			So(*appConfig.FraudProtection.Enabled, ShouldBeFalse)
		})
	})
}

func TestAppConfig(t *testing.T) {
	ctx := context.Background()
	Convey("AppConfig", t, func() {
		fixture := `id: test
http:
  public_origin: http://test
identity:
  oauth:
    providers:
    - type: google
      alias: google
      client_id: a
    - type: facebook
      alias: facebook
      client_id: a
    - type: linkedin
      alias: linkedin
      client_id: a
    - type: azureadv2
      alias: azureadv2
      client_id: a
      tenant: a
    - type: azureadb2c
      alias: azureadb2c
      client_id: a
      tenant: a
      policy: a
    - type: adfs
      alias: adfs
      client_id: a
      discovery_document_endpoint: http://test
    - type: apple
      alias: apple
      client_id: a
      key_id: a
      team_id: a
    - type: wechat
      alias: wechat
      client_id: a
      app_type: web
      account_id: gh_
`

		Convey("populate default values", func() {
			cfg, err := config.Parse(ctx, []byte(fixture))
			So(err, ShouldBeNil)

			data, err := os.ReadFile("testdata/default_config.yaml")
			if err != nil {
				panic(err)
			}

			_, err = config.Parse(ctx, data)
			So(err, ShouldBeNil)

			var defaultCfg config.AppConfig
			err = yaml.Unmarshal(data, &defaultCfg)
			if err != nil {
				panic(err)
			}

			So(cfg, ShouldResemble, &defaultCfg)
		})

		Convey("round-trip default configuration", func() {
			cfg, err := config.Parse(ctx, []byte(fixture))
			So(err, ShouldBeNil)

			data, err := yaml.Marshal(cfg)
			So(err, ShouldBeNil)

			cfg2, err := config.Parse(ctx, data)
			So(err, ShouldBeNil)
			So(cfg, ShouldResemble, cfg2)
			So(*cfg2.FraudProtection.SMS.UnverifiedOTPBudget.DailyRatio, ShouldEqual, 0.3)
			So(*cfg2.FraudProtection.SMS.UnverifiedOTPBudget.HourlyRatio, ShouldEqual, 0.2)
		})

		// Regression test: empty arrays in required fields must survive JSON→YAML round-trip.
		// Previously, omitempty caused empty slices to be dropped, turning e.g.
		// source:{cidrs:[],geo_location_codes:[]} into source:{}, which then
		// disappeared in YAML serialization and broke the "required" schema check
		// on all subsequent saves from unrelated pages.
		Convey("IP filter rule with empty source arrays is valid after JSON-YAML round-trip", func() {
			// Simulate portal saving an enabled IP blocklist with no entries filled in.
			inputJSON := `{
				"id": "test",
				"http": {"public_origin": "http://test"},
				"network_protection": {
					"ip_filter": {
						"default_action": "allow",
						"rules": [{"name": "__portal", "action": "deny", "source": {"cidrs": [], "geo_location_codes": []}}]
					}
				}
			}`

			inputYAML, err := yaml.JSONToYAML([]byte(inputJSON))
			So(err, ShouldBeNil)
			cfg, err := config.Parse(ctx, inputYAML)
			So(err, ShouldBeNil)

			// Simulate the server serialising the stored config back to JSON
			// (as it does when responding to a portal query), then the portal
			// re-saving from an unrelated page — which sends that JSON verbatim.
			roundTrippedJSON, err := json.Marshal(cfg)
			So(err, ShouldBeNil)
			roundTrippedYAML, err := yaml.JSONToYAML(roundTrippedJSON)
			So(err, ShouldBeNil)
			_, err = config.Parse(ctx, roundTrippedYAML)
			So(err, ShouldBeNil)
		})

		Convey("fraud protection by-phone-country with empty geo_location_codes is valid after JSON-YAML round-trip", func() {
			inputJSON := `{
				"id": "test",
				"http": {"public_origin": "http://test"},
				"fraud_protection": {
					"sms": {
						"unverified_otp_budget": {
							"by_phone_country": [{"geo_location_codes": []}]
						}
					}
				}
			}`

			inputYAML, err := yaml.JSONToYAML([]byte(inputJSON))
			So(err, ShouldBeNil)
			cfg, err := config.Parse(ctx, inputYAML)
			So(err, ShouldBeNil)

			roundTrippedJSON, err := json.Marshal(cfg)
			So(err, ShouldBeNil)
			roundTrippedYAML, err := yaml.JSONToYAML(roundTrippedJSON)
			So(err, ShouldBeNil)
			_, err = config.Parse(ctx, roundTrippedYAML)
			So(err, ShouldBeNil)
		})

		Convey("parse validation", func() {
			f, err := os.Open("testdata/config_tests.yaml")
			if err != nil {
				panic(err)
			}
			defer f.Close()

			type TestCase struct {
				Name   string  `yaml:"name"`
				Error  *string `yaml:"error"`
				Config any     `yaml:"config"`
			}

			decoder := goyaml.NewDecoder(f)
			for {
				var testCase TestCase
				err := decoder.Decode(&testCase)
				if errors.Is(err, io.EOF) {
					break
				} else if err != nil {
					panic(err)
				}

				Convey(testCase.Name, func() {
					data, err := goyaml.Marshal(testCase.Config)
					if err != nil {
						panic(err)
					}

					_, err = config.Parse(ctx, data)
					if testCase.Error != nil {
						So(err, ShouldBeError, *testCase.Error)
					} else {
						So(err, ShouldBeNil)
					}
				})
			}
		})
	})
}
