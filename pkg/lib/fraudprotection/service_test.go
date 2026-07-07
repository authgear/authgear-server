package fraudprotection

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

// --- stub implementations for testing ---

type stubMetrics struct {
	rollingMax                       int64
	country24h                       int64
	country1h                        int64
	ip24h                            int64
	recordErr                        error
	recordUnverifiedDrainErr         error
	getErr                           error
	calls                            *[]string
	lastRecordedUnverifiedDrainCount int
}

func (s *stubMetrics) RecordVerified(_ context.Context, _, _ string) error {
	if s.calls != nil {
		*s.calls = append(*s.calls, "metrics.record_verified")
	}
	return s.recordErr
}
func (s *stubMetrics) RecordUnverifiedSMSOTPCountDrained(_ context.Context, _, _ string, count int) error {
	if s.calls != nil {
		*s.calls = append(*s.calls, "metrics.record_unverified_sms_otp_count_drained")
	}
	s.lastRecordedUnverifiedDrainCount = count
	return s.recordUnverifiedDrainErr
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
	triggered                 LeakyBucketTriggered
	levels                    LeakyBucketLevels
	sentErr                   error
	drainErr                  error
	verifyCountryErr          error
	recordVerifiedCountryCall int
	recordVerifiedDrainCall   int
	lastDrainCount            int
	calls                     *[]string
}

func (s *stubLeakyBucket) RecordUnverifiedSMSOTPSent(_ context.Context, _, _ string, _ LeakyBucketThresholds) (LeakyBucketTriggered, LeakyBucketLevels, error) {
	return s.triggered, s.levels, s.sentErr
}
func (s *stubLeakyBucket) DrainUnverifiedSMSOTPSent(_ context.Context, _, _ string, _ LeakyBucketThresholds, count int) error {
	s.recordVerifiedDrainCall++
	s.lastDrainCount = count
	if s.calls != nil {
		*s.calls = append(*s.calls, "leaky_bucket.drain_unverified_sms_otp_sent")
	}
	return s.drainErr
}
func (s *stubLeakyBucket) RecordSMSOTPVerifiedCountry(_ context.Context, _, _ string) error {
	s.recordVerifiedCountryCall++
	if s.calls != nil {
		*s.calls = append(*s.calls, "leaky_bucket.record_verified_country")
	}
	return s.verifyCountryErr
}

type stubEventService struct{}

func (s *stubEventService) DispatchEventImmediately(_ context.Context, _ event.NonBlockingPayload) error {
	return nil
}

type stubVerifiedClaimChecker struct {
	exists bool
	err    error
}

func (s *stubVerifiedClaimChecker) ExistsByClaimNameAndValue(_ context.Context, _, _ string) (bool, error) {
	return s.exists, s.err
}

type stubDatabaseHandle struct{}

func (s *stubDatabaseHandle) IsInTx(_ context.Context) bool { return false }
func (s *stubDatabaseHandle) ReadOnly(_ context.Context, do func(context.Context) error) error {
	return do(context.Background())
}

// --- helpers ---

func defaultCfg() *config.FraudProtectionConfig {
	c := &config.FraudProtectionConfig{}
	config.SetFieldDefaults(c)
	return c
}

// --- tests ---

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

func TestAlwaysAllowReason(t *testing.T) {
	Convey("alwaysAllowReason", t, func() {
		svc := &Service{}

		Convey("returns empty when AlwaysAllow is nil", func() {
			cfg := &config.FraudProtectionConfig{
				Decision: &config.FraudProtectionDecision{AlwaysAllow: nil},
			}
			So(svc.alwaysAllowReason(cfg, "1.2.3.4", "+6591234567", "SG"), ShouldEqual, model.FraudProtectionAllowReason(""))
		})

		Convey("returns empty when Decision is nil", func() {
			cfg := &config.FraudProtectionConfig{Decision: nil}
			So(svc.alwaysAllowReason(cfg, "1.2.3.4", "+6591234567", "SG"), ShouldEqual, model.FraudProtectionAllowReason(""))
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

			Convey("returns always_allow_ip for IP inside CIDR", func() {
				So(svc.alwaysAllowReason(cfg, "10.1.2.3", "+6591234567", "SG"), ShouldEqual, model.FraudProtectionAllowReasonAlwaysAllowIP)
			})

			Convey("returns empty for IP outside CIDR", func() {
				So(svc.alwaysAllowReason(cfg, "11.0.0.1", "+6591234567", "SG"), ShouldEqual, model.FraudProtectionAllowReason(""))
			})

			Convey("skips invalid CIDR gracefully", func() {
				cfg.Decision.AlwaysAllow.IPAddress.CIDRs = []string{"not-a-cidr", "10.0.0.0/8"}
				So(svc.alwaysAllowReason(cfg, "10.5.5.5", "+6591234567", "SG"), ShouldEqual, model.FraudProtectionAllowReasonAlwaysAllowIP)
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

			Convey("returns always_allow_phone for phone from allowed country", func() {
				So(svc.alwaysAllowReason(cfg, "1.2.3.4", "+6591234567", "SG"), ShouldEqual, model.FraudProtectionAllowReasonAlwaysAllowPhone)
			})

			Convey("returns empty for phone from non-allowed country", func() {
				So(svc.alwaysAllowReason(cfg, "1.2.3.4", "+12125550000", "US"), ShouldEqual, model.FraudProtectionAllowReason(""))
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

			Convey("returns always_allow_phone for phone matching regex", func() {
				So(svc.alwaysAllowReason(cfg, "1.2.3.4", "+6591234567", "SG"), ShouldEqual, model.FraudProtectionAllowReasonAlwaysAllowPhone)
			})

			Convey("returns empty for phone not matching regex", func() {
				So(svc.alwaysAllowReason(cfg, "1.2.3.4", "+6592345678", "SG"), ShouldEqual, model.FraudProtectionAllowReason(""))
			})

			Convey("skips invalid regex gracefully", func() {
				cfg.Decision.AlwaysAllow.PhoneNumber.Regex = []string{`[invalid`, `^\+6591\d{6}$`}
				So(svc.alwaysAllowReason(cfg, "1.2.3.4", "+6591234567", "SG"), ShouldEqual, model.FraudProtectionAllowReasonAlwaysAllowPhone)
			})
		})
	})
}

func TestComputeThresholds(t *testing.T) {
	Convey("ComputeThresholds", t, func() {
		ctx := context.Background()

		newSvc := func(rollingMax, country24h, country1h, ip24h int64) *Service {
			return &Service{
				Config: defaultCfg(),
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
			// countryDaily = max(20, max(0*0.3, 0*0.3)) = 20
			So(thresholds.CountryDaily, ShouldEqual, 20)
			// countryHourly = max(3, max(0/6*0.2, 0*0.2)) = 3
			So(thresholds.CountryHourly, ShouldEqual, 3)
			// ipDaily = max(10, 0*0.3) = 10
			So(thresholds.IPDaily, ShouldEqual, 10)
			// ipHourly = max(5, 0/6*0.2) = 5
			So(thresholds.IPHourly, ShouldEqual, 5)
		})

		Convey("scales with large historical counts", func() {
			svc := newSvc(500, 1000, 200, 600)
			thresholds, err := svc.ComputeThresholds(ctx, "1.2.3.4", "SG")
			So(err, ShouldBeNil)
			// countryDaily = max(20, max(500*0.3=150, 1000*0.3=300)) = 300
			So(thresholds.CountryDaily, ShouldEqual, 300)
			// countryHourly = max(3, max(500/6*0.2≈16.7, 200*0.2=40)) = 40
			So(thresholds.CountryHourly, ShouldEqual, 40)
			// ipDaily = max(10, 600*0.3=180) = 180
			So(thresholds.IPDaily, ShouldEqual, 180)
			// ipHourly = max(5, 600/6*0.2=20) = 20
			So(thresholds.IPHourly, ShouldEqual, 20)
		})

		Convey("uses the rolling-max-based hourly formula instead of daily threshold / 6", func() {
			svc := newSvc(600, 1000, 10, 0)
			thresholds, err := svc.ComputeThresholds(ctx, "1.2.3.4", "SG")
			So(err, ShouldBeNil)
			// countryDaily = max(20, max(600*0.3=180, 1000*0.3=300)) = 300
			So(thresholds.CountryDaily, ShouldEqual, 300)
			// countryHourly = max(3, max(600/6*0.2=20, 10*0.2=2)) = 20
			So(thresholds.CountryHourly, ShouldEqual, 20)
		})

		Convey("uses the first matching country override and falls back independently by dimension", func() {
			cfg := defaultCfg()
			cfg.SMS.UnverifiedOTPBudget.ByPhoneCountry = []*config.FraudProtectionSMSUnverifiedOTPBudgetByPhoneCountryConfig{
				{
					GeoLocationCodes: []string{"SG", "HK"},
					DailyRatio:       new(0.15),
				},
				{
					GeoLocationCodes: []string{"SG"},
					DailyRatio:       new(0.9),
					HourlyRatio:      new(0.1),
				},
			}
			svc := newSvc(600, 1000, 10, 600)
			svc.Config = cfg

			thresholds, err := svc.ComputeThresholds(ctx, "1.2.3.4", "SG")
			So(err, ShouldBeNil)
			// countryDaily = max(20, max(600*0.15=90, 1000*0.15=150)) = 150
			So(thresholds.CountryDaily, ShouldEqual, 150)
			// countryHourly uses the first matching override's hourly_ratio if present;
			// here it is missing, so it falls back to the global hourly_ratio = 0.2.
			// countryHourly = max(3, max(600/6*0.2=20, 10*0.2=2)) = 20
			So(thresholds.CountryHourly, ShouldEqual, 20)
			// ipDaily = max(10, 600*0.3=180) = 180
			So(thresholds.IPDaily, ShouldEqual, 180)
			// ipHourly = max(5, 600/6*0.2=20) = 20
			So(thresholds.IPHourly, ShouldEqual, 20)
		})

		Convey("returns error when metrics query fails", func() {
			svc := &Service{
				Config:      defaultCfg(),
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

		disabledCfg := &config.FraudProtectionConfig{Enabled: new(false)}

		Convey("returns nil immediately when disabled", func() {
			svc := &Service{
				Config:      disabledCfg,
				Metrics:     &stubMetrics{},
				LeakyBucket: &stubLeakyBucket{},
			}
			err := svc.CheckAndRecord(ctx, "+6591234567", "otp")
			So(err, ShouldBeNil)
		})

		Convey("returns nil for unparseable phone number", func() {
			svc := &Service{
				Config:      enabledCfg,
				RemoteIP:    httputil.RemoteIP("1.2.3.4"),
				Metrics:     &stubMetrics{},
				LeakyBucket: &stubLeakyBucket{},
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
				Config:         recordOnlyCfg,
				RemoteIP:       httputil.RemoteIP("1.2.3.4"),
				Metrics:        &stubMetrics{},
				LeakyBucket:    &stubLeakyBucket{triggered: LeakyBucketTriggered{CountryDaily: true}},
				Clock:          clock.NewMockClock(),
				Database:       &stubDatabaseHandle{},
				EventService:   &stubEventService{},
				VerifiedClaims: &stubVerifiedClaimChecker{},
			}
			err := svc.CheckAndRecord(ctx, "+6591234567", "otp")
			So(err, ShouldBeNil)
		})

		Convey("returns ErrBlockedByFraudProtection when warning triggered and action is deny", func() {
			svc := &Service{
				Config:         enabledCfg,
				RemoteIP:       httputil.RemoteIP("1.2.3.4"),
				Metrics:        &stubMetrics{},
				LeakyBucket:    &stubLeakyBucket{triggered: LeakyBucketTriggered{CountryDaily: true}},
				Clock:          clock.NewMockClock(),
				Database:       &stubDatabaseHandle{},
				EventService:   &stubEventService{},
				VerifiedClaims: &stubVerifiedClaimChecker{},
			}
			err := svc.CheckAndRecord(ctx, "+6591234567", "otp")
			So(err, ShouldEqual, ErrBlockedByFraudProtection)
		})

		Convey("returns nil when phone is a verified claim even if warning triggered", func() {
			svc := &Service{
				Config:         enabledCfg,
				RemoteIP:       httputil.RemoteIP("1.2.3.4"),
				Metrics:        &stubMetrics{},
				LeakyBucket:    &stubLeakyBucket{triggered: LeakyBucketTriggered{CountryDaily: true}},
				Clock:          clock.NewMockClock(),
				Database:       &stubDatabaseHandle{},
				EventService:   &stubEventService{},
				VerifiedClaims: &stubVerifiedClaimChecker{exists: true},
			}
			err := svc.CheckAndRecord(ctx, "+6591234567", "otp")
			So(err, ShouldBeNil)
		})

		Convey("returns error when leaky bucket fails", func() {
			import_err := &testError{"redis error"}
			svc := &Service{
				Config:      enabledCfg,
				RemoteIP:    httputil.RemoteIP("1.2.3.4"),
				Metrics:     &stubMetrics{},
				LeakyBucket: &stubLeakyBucket{sentErr: import_err},
			}
			err := svc.CheckAndRecord(ctx, "+6591234567", "otp")
			So(err, ShouldEqual, import_err)
		})

		Convey("returns error when verified claim check fails", func() {
			import_err := &testError{"db error"}
			svc := &Service{
				Config:         enabledCfg,
				RemoteIP:       httputil.RemoteIP("1.2.3.4"),
				Metrics:        &stubMetrics{},
				LeakyBucket:    &stubLeakyBucket{triggered: LeakyBucketTriggered{CountryDaily: true}},
				VerifiedClaims: &stubVerifiedClaimChecker{err: import_err},
			}
			err := svc.CheckAndRecord(ctx, "+6591234567", "otp")
			So(err, ShouldEqual, import_err)
		})

		Convey("allowlisted IP CIDR still fills buckets but is never blocked", func() {
			cfgWithAllowlist := defaultCfg()
			cfgWithAllowlist.Decision = &config.FraudProtectionDecision{
				Action: config.FraudProtectionDecisionActionDenyIfAnyWarning,
				AlwaysAllow: &config.FraudProtectionAlwaysAllow{
					IPAddress: &config.FraudProtectionIPAlwaysAllow{
						CIDRs: []string{"10.0.0.0/8"},
					},
				},
			}
			leakyBucket := &stubLeakyBucket{triggered: LeakyBucketTriggered{CountryDaily: true}}
			svc := &Service{
				Config:         cfgWithAllowlist,
				RemoteIP:       httputil.RemoteIP("10.1.2.3"),
				Metrics:        &stubMetrics{},
				LeakyBucket:    leakyBucket,
				Clock:          clock.NewMockClock(),
				Database:       &stubDatabaseHandle{},
				EventService:   &stubEventService{},
				VerifiedClaims: &stubVerifiedClaimChecker{},
			}
			err := svc.CheckAndRecord(ctx, "+6591234567", "otp")
			So(err, ShouldBeNil)
			// bucket was filled despite the allowlist match
			So(leakyBucket.triggered.CountryDaily, ShouldBeTrue)
		})
	})
}

func TestRecordSMSOTPVerified(t *testing.T) {
	Convey("RecordSMSOTPVerified", t, func() {
		ctx := context.Background()
		cfg := defaultCfg()

		Convey("records metrics, marks the verified country, and then drains buckets", func() {
			calls := []string{}
			leakyBucket := &stubLeakyBucket{calls: &calls}
			svc := &Service{
				Config:      cfg,
				RemoteIP:    httputil.RemoteIP("1.2.3.4"),
				Metrics:     &stubMetrics{calls: &calls},
				LeakyBucket: leakyBucket,
			}

			err := svc.RecordSMSOTPVerified(ctx, "+6591234567")
			So(err, ShouldBeNil)
			So(calls, ShouldResemble, []string{
				"metrics.record_verified",
				"leaky_bucket.record_verified_country",
				"leaky_bucket.drain_unverified_sms_otp_sent",
				"metrics.record_unverified_sms_otp_count_drained",
			})
			So(leakyBucket.recordVerifiedCountryCall, ShouldEqual, 1)
			So(leakyBucket.recordVerifiedDrainCall, ShouldEqual, 1)
			So(leakyBucket.lastDrainCount, ShouldEqual, 1)
		})

		Convey("returns verified-country errors before draining buckets", func() {
			calls := []string{}
			expectedErr := &testError{"verified country write failure"}
			leakyBucket := &stubLeakyBucket{
				calls:            &calls,
				verifyCountryErr: expectedErr,
			}
			svc := &Service{
				Config:      cfg,
				RemoteIP:    httputil.RemoteIP("1.2.3.4"),
				Metrics:     &stubMetrics{calls: &calls},
				LeakyBucket: leakyBucket,
			}

			err := svc.RecordSMSOTPVerified(ctx, "+6591234567")
			So(err, ShouldEqual, expectedErr)
			So(calls, ShouldResemble, []string{
				"metrics.record_verified",
				"leaky_bucket.record_verified_country",
			})
			So(leakyBucket.recordVerifiedDrainCall, ShouldEqual, 0)
		})
	})
}

func TestRevertSMSOTPSent(t *testing.T) {
	Convey("RevertSMSOTPSent", t, func() {
		Convey("drains and records unverified-drain metric", func() {
			ctx := context.Background()
			cfg := defaultCfg()
			leakyBucket := &stubLeakyBucket{}
			metrics := &stubMetrics{}
			svc := &Service{
				Config:      cfg,
				RemoteIP:    httputil.RemoteIP("1.2.3.4"),
				Metrics:     metrics,
				LeakyBucket: leakyBucket,
			}

			err := svc.RevertSMSOTPSent(ctx, "+6591234567", 2)
			So(err, ShouldBeNil)
			So(leakyBucket.recordVerifiedCountryCall, ShouldEqual, 0)
			So(leakyBucket.recordVerifiedDrainCall, ShouldEqual, 1)
			So(leakyBucket.lastDrainCount, ShouldEqual, 2)
			So(metrics.lastRecordedUnverifiedDrainCount, ShouldEqual, 2)
		})

		Convey("does not record drain metric if draining fails", func() {
			ctx := context.Background()
			cfg := defaultCfg()
			leakyBucket := &stubLeakyBucket{drainErr: &testError{"drain failed"}}
			metrics := &stubMetrics{}
			svc := &Service{
				Config:      cfg,
				RemoteIP:    httputil.RemoteIP("1.2.3.4"),
				Metrics:     metrics,
				LeakyBucket: leakyBucket,
			}

			err := svc.RevertSMSOTPSent(ctx, "+6591234567", 2)
			So(err, ShouldNotBeNil)
			So(metrics.lastRecordedUnverifiedDrainCount, ShouldEqual, 0)
		})
	})
}
