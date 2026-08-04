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

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestParseFeatureConfig(t *testing.T) {
	Convey("default feature config", t, func() {
		ctx := context.Background()
		cfg, err := config.ParseFeatureConfig(ctx, []byte(`{}`))
		So(err, ShouldBeNil)

		data, err := os.ReadFile("testdata/default_feature.yaml")
		So(err, ShouldBeNil)

		var defaultCfg config.FeatureConfig
		err = yaml.Unmarshal(data, &defaultCfg)
		So(err, ShouldBeNil)

		So(cfg, ShouldResemble, &defaultCfg)
	})

	Convey("merge feature config", t, func() {
		ctx := context.Background()
		type TestCase struct {
			Configs []any `yaml:"configs"`
			Result  any   `yaml:"result"`
		}

		f, err := os.Open("testdata/merge_feature.yaml")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		decoder := goyaml.NewDecoder(f)
		var testCase TestCase
		err = decoder.Decode(&testCase)
		if err != nil {
			panic(err)
		}

		resultData, err := goyaml.Marshal(testCase.Result)
		if err != nil {
			panic(err)
		}

		expected, err := config.ParseFeatureConfig(ctx, resultData)
		So(err, ShouldBeNil)

		mergedConfig := &config.FeatureConfig{}
		for _, cfgData := range testCase.Configs {
			data, err := goyaml.Marshal(cfgData)
			if err != nil {
				panic(err)
			}

			cfg, err := config.ParseFeatureConfigWithoutDefaults(ctx, data)
			So(err, ShouldBeNil)

			mergedConfig = mergedConfig.Merge(cfg)
		}
		config.SetFieldDefaults(mergedConfig)
		mergedConfig = mergedConfig.Migrate()

		So(mergedConfig, ShouldResemble, expected)
	})

	Convey("ParseFeatureConfig", t, func() {
		f, err := os.Open("testdata/parse_feature_tests.yaml")
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
				ctx := context.Background()
				data, err := goyaml.Marshal(testCase.Config)
				if err != nil {
					panic(err)
				}

				_, err = config.ParseFeatureConfig(ctx, data)
				if testCase.Error != nil {
					So(err, ShouldBeError, *testCase.Error)
				} else {
					So(err, ShouldBeNil)
				}
			})
		}
	})
}

func TestFeatureConfigMigrate(t *testing.T) {
	Convey("Migrate converts deprecated usage limits when unified usage config is absent", t, func() {
		enabled := true
		quota := 24
		cfg := (&config.FeatureConfig{
			Messaging: &config.MessagingFeatureConfig{
				SMSUsage: &config.Deprecated_UsageLimitConfig{
					Enabled: &enabled,
					Period:  config.Deprecated_UsageLimitPeriodMonth,
					Quota:   &quota,
				},
			},
		}).Migrate()

		So(cfg.Usage, ShouldNotBeNil)
		So(cfg.Usage.Limits, ShouldNotBeNil)
		So(cfg.Usage.Limits.SMS, ShouldResemble, []config.FeatureUsageLimitConfig{{
			Quota:  24,
			Period: model.UsageLimitPeriodMonth,
			Action: model.UsageLimitActionBlock,
		}})
	})

	Convey("Migrate keeps unified usage limits when deprecated usage config also exists", t, func() {
		enabled := true
		quota := 24
		cfg := (&config.FeatureConfig{
			Messaging: &config.MessagingFeatureConfig{
				SMSUsage: &config.Deprecated_UsageLimitConfig{
					Enabled: &enabled,
					Period:  config.Deprecated_UsageLimitPeriodMonth,
					Quota:   &quota,
				},
			},
			Usage: &config.FeatureUsageConfig{
				Limits: &config.FeatureUsageLimitsConfig{
					SMS: []config.FeatureUsageLimitConfig{{
						Quota:  3,
						Period: model.UsageLimitPeriodDay,
						Action: model.UsageLimitActionAlert,
					}},
				},
			},
		}).Migrate()

		So(cfg.Usage, ShouldNotBeNil)
		So(cfg.Usage.Limits, ShouldNotBeNil)
		So(cfg.Usage.Limits.SMS, ShouldResemble, []config.FeatureUsageLimitConfig{{
			Quota:  3,
			Period: model.UsageLimitPeriodDay,
			Action: model.UsageLimitActionAlert,
		}})
	})
}

// This mirrors configsource/resources.go's viewEffectiveResource exactly
// (fold layers via Merge, then yaml.Marshal, then re-parse) rather than the
// "merge feature config" Convey block above, which stops at
// SetFieldDefaults/Migrate on the in-memory merged struct and so never
// exercises the yaml.Marshal step -- the one place a field tagged
// `omitempty` (as opposed to `omitzero`) can silently turn an explicit
// empty slice back into an absent/nil field once re-parsed.
func TestFeatureConfigEffectiveResourceRoundTrip(t *testing.T) {
	Convey("an explicit empty phone_input.allowlist override survives the merge fold's marshal-and-reparse round trip", t, func() {
		ctx := context.Background()

		planYAML := []byte(`
ui:
  phone_input:
    allowlist:
      - US
      - GB
`)
		appYAML := []byte(`
ui:
  phone_input:
    allowlist: []
`)

		planCfg, err := config.ParseFeatureConfigWithoutDefaults(ctx, planYAML)
		So(err, ShouldBeNil)
		appCfg, err := config.ParseFeatureConfigWithoutDefaults(ctx, appYAML)
		So(err, ShouldBeNil)

		mergedConfig := &config.FeatureConfig{}
		mergedConfig = mergedConfig.Merge(planCfg)
		mergedConfig = mergedConfig.Merge(appCfg)

		mergedYAML, err := yaml.Marshal(mergedConfig)
		So(err, ShouldBeNil)

		effective, err := config.ParseFeatureConfig(ctx, mergedYAML)
		So(err, ShouldBeNil)

		So(effective.UI, ShouldNotBeNil)
		So(effective.UI.PhoneInput, ShouldNotBeNil)
		So(effective.UI.PhoneInput.AllowList, ShouldResemble, []string{})
	})

	Convey("an app override that doesn't touch phone_input.allowlist inherits the plan's list", t, func() {
		ctx := context.Background()

		planYAML := []byte(`
ui:
  phone_input:
    allowlist:
      - US
      - GB
`)
		// The app layer is present and non-empty, but never mentions
		// phone_input -- this is the "not set" case, distinct from the
		// explicit-empty case above, and must inherit the plan's list.
		appYAML := []byte(`
collaborator:
  maximum: 5
`)

		planCfg, err := config.ParseFeatureConfigWithoutDefaults(ctx, planYAML)
		So(err, ShouldBeNil)
		appCfg, err := config.ParseFeatureConfigWithoutDefaults(ctx, appYAML)
		So(err, ShouldBeNil)

		mergedConfig := &config.FeatureConfig{}
		mergedConfig = mergedConfig.Merge(planCfg)
		mergedConfig = mergedConfig.Merge(appCfg)

		mergedYAML, err := yaml.Marshal(mergedConfig)
		So(err, ShouldBeNil)

		effective, err := config.ParseFeatureConfig(ctx, mergedYAML)
		So(err, ShouldBeNil)

		So(effective.UI, ShouldNotBeNil)
		So(effective.UI.PhoneInput, ShouldNotBeNil)
		So(effective.UI.PhoneInput.AllowList, ShouldResemble, []string{"US", "GB"})
	})
}

// These single-field sections have `false` as their real, correct default
// (not disabled) -- `omitempty` hid that from JSON output, making a
// fully-resolved section marshal as an empty object indistinguishable from
// one with no fields at all. Sections stay non-nil with a `false` value
// regardless, from SetFieldDefaults's generic pointer-to-struct
// initialization (default.go) -- this test is specifically about the
// *serialized shape*, which is what the Site Admin API's feature-config UI
// (and any other JSON consumer of FeatureConfig) actually sees.
func TestFeatureConfigDisabledFieldsSerializeExplicitly(t *testing.T) {
	Convey("false-valued Disabled fields marshal explicitly, not as an empty object", t, func() {
		ctx := context.Background()
		cfg, err := config.ParseFeatureConfig(ctx, []byte(`{}`))
		So(err, ShouldBeNil)

		data, err := json.Marshal(cfg)
		So(err, ShouldBeNil)

		var raw map[string]any
		err = json.Unmarshal(data, &raw)
		So(err, ShouldBeNil)

		So(raw["custom_domain"], ShouldResemble, map[string]any{"disabled": false})
		So(raw["google_tag_manager"], ShouldResemble, map[string]any{"disabled": false})
		So(raw["rate_limits"], ShouldResemble, map[string]any{"disabled": false})

		ui, _ := raw["ui"].(map[string]any)
		So(ui["white_labeling"], ShouldResemble, map[string]any{"disabled": false})

		authentication, _ := raw["authentication"].(map[string]any)
		secondaryAuthenticators, _ := authentication["secondary_authenticators"].(map[string]any)
		So(secondaryAuthenticators["oob_otp_sms"], ShouldResemble, map[string]any{"disabled": false})

		identity, _ := raw["identity"].(map[string]any)
		loginID, _ := identity["login_id"].(map[string]any)
		types, _ := loginID["types"].(map[string]any)
		So(types["phone"], ShouldResemble, map[string]any{"disabled": false})

		oauth, _ := identity["oauth"].(map[string]any)
		providers, _ := oauth["providers"].(map[string]any)
		So(providers["google"], ShouldResemble, map[string]any{"disabled": false})
	})
}
