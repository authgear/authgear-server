package fraudprotection

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

// --- stub implementations for testing ---

type stubMetrics struct {
	rollingMax     int64
	country24h     int64
	country1h      int64
	ip24h          int64
	recordErr      error
	getErr         error
}

func (s *stubMetrics) RecordVerified(_ context.Context, _, _ string) error {
	return s.recordErr
}
func (s *stubMetrics) GetVerifiedByCountry24h(_ context.Context, _ string) (int64, error) {
	return s.country24h, s.getErr
}
func (s *stubMetrics) GetVerifiedByCountry1h(_ context.Context, _ string) (int64, error) {
	return s.country1h, s.getErr
}
func (s *stubMetrics) GetVerifiedByIP24h(_ context.Context, _ string) (int64, error) {
	return s.ip24h, s.getErr
}
func (s *stubMetrics) GetVerifiedByCountryPast14DaysRollingMax(_ context.Context, _ string) (int64, error) {
	return s.rollingMax, s.getErr
}

type stubLeakyBucket struct {
	triggered LeakyBucketTriggered
	sentErr   error
	drainErr  error
}

func (s *stubLeakyBucket) RecordSMSOTPSent(_ context.Context, _, _ string, _ LeakyBucketThresholds) (LeakyBucketTriggered, error) {
	return s.triggered, s.sentErr
}
func (s *stubLeakyBucket) RecordSMSOTPVerified(_ context.Context, _, _ string, _ LeakyBucketThresholds, _ int) error {
	return s.drainErr
}

// --- helpers ---

func newBoolPtr(b bool) *bool { return &b }

func defaultCfg() *config.FraudProtectionConfig {
	c := &config.FraudProtectionConfig{}
	config.SetFieldDefaults(c)
	return c
}

func defaultFeatureCfg(modifiable bool) *config.FraudProtectionFeatureConfig {
	b := modifiable
	return &config.FraudProtectionFeatureConfig{IsModifiable: &b}
}

// --- tests ---

func TestEffectiveFraudProtectionConfig(t *testing.T) {
	Convey("effectiveFraudProtectionConfig", t, func() {
		Convey("returns hardcoded default when IsModifiable is false", func() {
			app := &config.FraudProtectionConfig{
				Enabled:  newBoolPtr(false), // intentionally disabled by app
				Warnings: nil,
			}
			feature := defaultFeatureCfg(false)

			result := effectiveFraudProtectionConfig(app, feature)

			// Hardcoded default has Enabled=true
			So(*result.Enabled, ShouldBeTrue)
			// All 5 warning types present
			So(len(result.Warnings), ShouldEqual, 5)
		})

		Convey("returns app config when IsModifiable is true", func() {
			app := &config.FraudProtectionConfig{
				Enabled:  newBoolPtr(false),
				Warnings: nil,
			}
			feature := defaultFeatureCfg(true)

			result := effectiveFraudProtectionConfig(app, feature)

			So(*result.Enabled, ShouldBeFalse)
		})
	})
}

func TestEvaluateWarnings(t *testing.T) {
	Convey("evaluateWarnings", t, func() {
		svc := &Service{}

		allWarningsCfg := &config.FraudProtectionConfig{
			Warnings: []*config.FraudProtectionWarning{
				{Type: config.FraudProtectionWarningTypeSMSPhoneCountriesByIPDaily},
				{Type: config.FraudProtectionWarningTypeSMSUnverifiedOTPsByPhoneCountryDaily},
				{Type: config.FraudProtectionWarningTypeSMSUnverifiedOTPsByPhoneCountryHourly},
				{Type: config.FraudProtectionWarningTypeSMSUnverifiedOTPsByIPDaily},
				{Type: config.FraudProtectionWarningTypeSMSUnverifiedOTPsByIPHourly},
			},
		}

		Convey("no warnings when nothing is triggered", func() {
			warnings := svc.evaluateWarnings(allWarningsCfg, LeakyBucketTriggered{})
			So(warnings, ShouldBeEmpty)
		})

		Convey("all 5 warnings when all buckets are triggered and all are configured", func() {
			triggered := LeakyBucketTriggered{
				IPCountriesDaily: true,
				CountryDaily:     true,
				CountryHourly:    true,
				IPDaily:          true,
				IPHourly:         true,
			}
			warnings := svc.evaluateWarnings(allWarningsCfg, triggered)
			So(len(warnings), ShouldEqual, 5)
		})

		Convey("only configured warnings are returned even if all buckets trigger", func() {
			limitedCfg := &config.FraudProtectionConfig{
				Warnings: []*config.FraudProtectionWarning{
					{Type: config.FraudProtectionWarningTypeSMSPhoneCountriesByIPDaily},
				},
			}
			triggered := LeakyBucketTriggered{
				IPCountriesDaily: true,
				CountryDaily:     true,
				IPHourly:         true,
			}
			warnings := svc.evaluateWarnings(limitedCfg, triggered)
			So(len(warnings), ShouldEqual, 1)
			So(warnings[0], ShouldEqual, config.FraudProtectionWarningTypeSMSPhoneCountriesByIPDaily)
		})

		Convey("IPCountriesDaily trigger maps to correct warning type", func() {
			warnings := svc.evaluateWarnings(allWarningsCfg, LeakyBucketTriggered{IPCountriesDaily: true})
			So(len(warnings), ShouldEqual, 1)
			So(warnings[0], ShouldEqual, config.FraudProtectionWarningTypeSMSPhoneCountriesByIPDaily)
		})

		Convey("CountryDaily trigger maps to correct warning type", func() {
			warnings := svc.evaluateWarnings(allWarningsCfg, LeakyBucketTriggered{CountryDaily: true})
			So(len(warnings), ShouldEqual, 1)
			So(warnings[0], ShouldEqual, config.FraudProtectionWarningTypeSMSUnverifiedOTPsByPhoneCountryDaily)
		})

		Convey("IPHourly trigger maps to correct warning type", func() {
			warnings := svc.evaluateWarnings(allWarningsCfg, LeakyBucketTriggered{IPHourly: true})
			So(len(warnings), ShouldEqual, 1)
			So(warnings[0], ShouldEqual, config.FraudProtectionWarningTypeSMSUnverifiedOTPsByIPHourly)
		})
	})
}

func TestIsAlwaysAllowed(t *testing.T) {
	Convey("isAlwaysAllowed", t, func() {
		svc := &Service{}

		Convey("returns false when AlwaysAllow is nil", func() {
			cfg := &config.FraudProtectionConfig{
				Decision: &config.FraudProtectionDecision{AlwaysAllow: nil},
			}
			So(svc.isAlwaysAllowed(cfg, "1.2.3.4", "+6591234567", "SG"), ShouldBeFalse)
		})

		Convey("returns false when Decision is nil", func() {
			cfg := &config.FraudProtectionConfig{Decision: nil}
			So(svc.isAlwaysAllowed(cfg, "1.2.3.4", "+6591234567", "SG"), ShouldBeFalse)
		})

		Convey("IP CIDR allowlist", func() {
			cfg := &config.FraudProtectionConfig{
				Decision: &config.FraudProtectionDecision{
					AlwaysAllow: &config.FraudProtectionAlwaysAllow{
						IPAddress: &config.FraudProtectionIPAlwaysAllow{
							CIDRs: []string{"10.0.0.0/8"},
						},
					},
				},
			}

			Convey("allows IP inside CIDR", func() {
				So(svc.isAlwaysAllowed(cfg, "10.1.2.3", "+6591234567", "SG"), ShouldBeTrue)
			})

			Convey("does not allow IP outside CIDR", func() {
				So(svc.isAlwaysAllowed(cfg, "11.0.0.1", "+6591234567", "SG"), ShouldBeFalse)
			})

			Convey("skips invalid CIDR gracefully", func() {
				cfg.Decision.AlwaysAllow.IPAddress.CIDRs = []string{"not-a-cidr", "10.0.0.0/8"}
				So(svc.isAlwaysAllowed(cfg, "10.5.5.5", "+6591234567", "SG"), ShouldBeTrue)
			})
		})

		Convey("phone country allowlist", func() {
			cfg := &config.FraudProtectionConfig{
				Decision: &config.FraudProtectionDecision{
					AlwaysAllow: &config.FraudProtectionAlwaysAllow{
						PhoneNumber: &config.FraudProtectionPhoneNumberAlwaysAllow{
							GeoLocationCodes: []string{"SG", "MY"},
						},
					},
				},
			}

			Convey("allows phone from allowed country", func() {
				So(svc.isAlwaysAllowed(cfg, "1.2.3.4", "+6591234567", "SG"), ShouldBeTrue)
			})

			Convey("does not allow phone from non-allowed country", func() {
				So(svc.isAlwaysAllowed(cfg, "1.2.3.4", "+12125550000", "US"), ShouldBeFalse)
			})
		})

		Convey("phone regex allowlist", func() {
			cfg := &config.FraudProtectionConfig{
				Decision: &config.FraudProtectionDecision{
					AlwaysAllow: &config.FraudProtectionAlwaysAllow{
						PhoneNumber: &config.FraudProtectionPhoneNumberAlwaysAllow{
							Regex: []string{`^\+6591\d{6}$`},
						},
					},
				},
			}

			Convey("allows phone matching regex", func() {
				So(svc.isAlwaysAllowed(cfg, "1.2.3.4", "+6591234567", "SG"), ShouldBeTrue)
			})

			Convey("does not allow phone not matching regex", func() {
				So(svc.isAlwaysAllowed(cfg, "1.2.3.4", "+6592345678", "SG"), ShouldBeFalse)
			})

			Convey("skips invalid regex gracefully", func() {
				cfg.Decision.AlwaysAllow.PhoneNumber.Regex = []string{`[invalid`, `^\+6591\d{6}$`}
				So(svc.isAlwaysAllowed(cfg, "1.2.3.4", "+6591234567", "SG"), ShouldBeTrue)
			})
		})
	})
}

func TestComputeThresholds(t *testing.T) {
	Convey("ComputeThresholds", t, func() {
		ctx := context.Background()

		newSvc := func(rollingMax, country24h, country1h, ip24h int64) *Service {
			return &Service{
				Metrics: &stubMetrics{
					rollingMax: rollingMax,
					country24h: country24h,
					country1h:  country1h,
					ip24h:      ip24h,
				},
				LeakyBucket: &stubLeakyBucket{},
			}
		}

		Convey("uses minimums when historical counts are low", func() {
			svc := newSvc(0, 0, 0, 0)
			thresholds, err := svc.ComputeThresholds(ctx, "1.2.3.4", "SG")
			So(err, ShouldBeNil)
			// countryDaily = max(20, max(0*0.2, 0*0.2)) = 20
			So(thresholds.CountryDaily, ShouldEqual, 20)
			// countryHourly = max(3, max(20/6, 0*0.2)) ≈ max(3, 3.33) = 3.33
			So(thresholds.CountryHourly, ShouldAlmostEqual, 20.0/6, 0.001)
			// ipDaily = max(10, 0*0.2) = 10
			So(thresholds.IPDaily, ShouldEqual, 10)
			// ipHourly = max(5, 0*0.2/6) = 5
			So(thresholds.IPHourly, ShouldEqual, 5)
		})

		Convey("scales with large historical counts", func() {
			// 1000 verified in past 24h from country → daily threshold driven by count
			svc := newSvc(500, 1000, 200, 600)
			thresholds, err := svc.ComputeThresholds(ctx, "1.2.3.4", "SG")
			So(err, ShouldBeNil)
			// countryDaily = max(20, max(500*0.2=100, 1000*0.2=200)) = 200
			So(thresholds.CountryDaily, ShouldEqual, 200)
			// countryHourly = max(3, max(200/6≈33.3, 200*0.2=40)) = 40
			So(thresholds.CountryHourly, ShouldEqual, 40)
			// ipDaily = max(10, 600*0.2=120) = 120
			So(thresholds.IPDaily, ShouldEqual, 120)
			// ipHourly = max(5, 600*0.2/6=20) = 20
			So(thresholds.IPHourly, ShouldEqual, 20)
		})

		Convey("returns error when metrics query fails", func() {
			svc := &Service{
				Metrics:     &stubMetrics{getErr: errMetricsFailure},
				LeakyBucket: &stubLeakyBucket{},
			}
			_, err := svc.ComputeThresholds(ctx, "1.2.3.4", "SG")
			So(err, ShouldNotBeNil)
		})
	})
}

var errMetricsFailure = &testError{"metrics failure"}

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }

func TestCheckAndRecord(t *testing.T) {
	Convey("CheckAndRecord", t, func() {
		ctx := context.Background()

		enabledCfg := defaultCfg()
		enabledCfg.Decision = &config.FraudProtectionDecision{
			Action: config.FraudProtectionDecisionActionDenyIfAnyWarning,
		}
		config.SetFieldDefaults(enabledCfg.Decision)

		disabledCfg := &config.FraudProtectionConfig{Enabled: newBoolPtr(false)}

		Convey("returns nil immediately when disabled", func() {
			svc := &Service{
				Config:        disabledCfg,
				FeatureConfig: defaultFeatureCfg(true),
				Metrics:       &stubMetrics{},
				LeakyBucket:   &stubLeakyBucket{},
			}
			err := svc.CheckAndRecord(ctx, "+6591234567", "otp")
			So(err, ShouldBeNil)
		})

		Convey("returns nil for unparseable phone number", func() {
			svc := &Service{
				Config:        enabledCfg,
				FeatureConfig: defaultFeatureCfg(true),
				RemoteIP:      httputil.RemoteIP("1.2.3.4"),
				Metrics:       &stubMetrics{},
				LeakyBucket:   &stubLeakyBucket{},
			}
			err := svc.CheckAndRecord(ctx, "not-a-phone", "otp")
			So(err, ShouldBeNil)
		})

		Convey("returns nil when no warnings triggered (record_only)", func() {
			recordOnlyCfg := defaultCfg()
			recordOnlyCfg.Decision = &config.FraudProtectionDecision{
				Action: config.FraudProtectionDecisionActionRecordOnly,
			}
			svc := &Service{
				Config:        recordOnlyCfg,
				FeatureConfig: defaultFeatureCfg(true),
				RemoteIP:      httputil.RemoteIP("1.2.3.4"),
				Metrics:       &stubMetrics{},
				LeakyBucket:   &stubLeakyBucket{triggered: LeakyBucketTriggered{CountryDaily: true}},
			}
			err := svc.CheckAndRecord(ctx, "+6591234567", "otp")
			So(err, ShouldBeNil)
		})

		Convey("returns ErrBlockedByFraudProtection when warning triggered and action is deny", func() {
			svc := &Service{
				Config:        enabledCfg,
				FeatureConfig: defaultFeatureCfg(true),
				RemoteIP:      httputil.RemoteIP("1.2.3.4"),
				Metrics:       &stubMetrics{},
				LeakyBucket:   &stubLeakyBucket{triggered: LeakyBucketTriggered{CountryDaily: true}},
			}
			err := svc.CheckAndRecord(ctx, "+6591234567", "otp")
			So(err, ShouldEqual, ErrBlockedByFraudProtection)
		})

		Convey("returns nil (non-fatal) when leaky bucket fails", func() {
			import_err := &testError{"redis error"}
			svc := &Service{
				Config:        enabledCfg,
				FeatureConfig: defaultFeatureCfg(true),
				RemoteIP:      httputil.RemoteIP("1.2.3.4"),
				Metrics:       &stubMetrics{},
				LeakyBucket:   &stubLeakyBucket{sentErr: import_err},
			}
			err := svc.CheckAndRecord(ctx, "+6591234567", "otp")
			So(err, ShouldBeNil)
		})

		Convey("skips check for allowlisted IP CIDR", func() {
			cfgWithAllowlist := defaultCfg()
			cfgWithAllowlist.Decision = &config.FraudProtectionDecision{
				Action: config.FraudProtectionDecisionActionDenyIfAnyWarning,
				AlwaysAllow: &config.FraudProtectionAlwaysAllow{
					IPAddress: &config.FraudProtectionIPAlwaysAllow{
						CIDRs: []string{"10.0.0.0/8"},
					},
				},
			}
			// Even with a triggered bucket, the allowlist should bypass the block.
			svc := &Service{
				Config:        cfgWithAllowlist,
				FeatureConfig: defaultFeatureCfg(true),
				RemoteIP:      httputil.RemoteIP("10.1.2.3"),
				Metrics:       &stubMetrics{},
				LeakyBucket:   &stubLeakyBucket{triggered: LeakyBucketTriggered{CountryDaily: true}},
			}
			err := svc.CheckAndRecord(ctx, "+6591234567", "otp")
			So(err, ShouldBeNil)
		})
	})
}
