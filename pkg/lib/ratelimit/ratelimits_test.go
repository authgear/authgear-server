package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"

	. "github.com/smartystreets/goconvey/convey"
)

func boolPtr(b bool) *bool {
	return &b
}

func TestRateLimits(t *testing.T) {
	Convey("RateLimit.ResolveBucketSpecs", t, func() {
		ctx := context.Background()
		userID := "testuserid"
		ipAddress := "1.2.3.4"
		target := "test@example.com"
		phone := "+1555000123"
		purpose := "testpurpose"
		clientID := "testclient"
		Convey("authentication.password", func() {
			rl := RateLimitAuthenticationPassword
			Convey("set in config", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          authentication:
            rate_limits:
              password:
                per_ip:
                  burst: 1
                  enabled: true
                  period: 1m
                per_user_per_ip:
                  burst: 2
                  enabled: true
                  period: 2m
              general:
                per_ip:
                  burst: 3
                  enabled: true
                  period: 3m
                per_user_per_ip:
                  burst: 4
                  enabled: true
                  period: 4m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					UserID:    userID,
					IPAddress: ipAddress,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitAuthenticationPasswordPerIP,
					RateLimitGroup: rl,
					Name:           VerifyPasswordPerIP,
					Arguments:      []string{ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("1m").Duration(),
					Burst:          1,
				}, {
					RateLimitName:  RateLimitAuthenticationPasswordPerUserPerIP,
					RateLimitGroup: rl,
					Name:           VerifyPasswordPerUserPerIP,
					Arguments:      []string{userID, ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("2m").Duration(),
					Burst:          2,
				}})
			})

			Convey("fallback to authentication.general when not set", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          authentication:
            rate_limits:
              general:
                per_ip:
                  burst: 3
                  enabled: true
                  period: 3m
                per_user_per_ip:
                  burst: 4
                  enabled: true
                  period: 4m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					UserID:    userID,
					IPAddress: ipAddress,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitAuthenticationPasswordPerIP,
					RateLimitGroup: rl,
					Name:           VerifyPasswordPerIP,
					Arguments:      []string{ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("3m").Duration(),
					Burst:          3,
				}, {
					RateLimitName:  RateLimitAuthenticationPasswordPerUserPerIP,
					RateLimitGroup: rl,
					Name:           VerifyPasswordPerUserPerIP,
					Arguments:      []string{userID, ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("4m").Duration(),
					Burst:          4,
				}})
			})
			Convey("disabled in config should not fallback", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          authentication:
            rate_limits:
              password:
                per_ip:
                  enabled: false
                per_user_per_ip:
                  enabled: false
              general:
                per_ip:
                  burst: 3
                  enabled: true
                  period: 3m
                per_user_per_ip:
                  burst: 4
                  enabled: true
                  period: 4m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					UserID:    userID,
					IPAddress: ipAddress,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitAuthenticationPasswordPerIP,
					RateLimitGroup: rl,
					Name:           VerifyPasswordPerIP,
					Arguments:      []string{ipAddress},
					IsGlobal:       false,
					Enabled:        false,
					Period:         time.Duration(0),
					Burst:          0,
				}, {
					RateLimitName:  RateLimitAuthenticationPasswordPerUserPerIP,
					RateLimitGroup: rl,
					Name:           VerifyPasswordPerUserPerIP,
					Arguments:      []string{userID, ipAddress},
					IsGlobal:       false,
					Enabled:        false,
					Period:         time.Duration(0),
					Burst:          0,
				}})
			})
		})

		Convey("authentication.oob_otp.email.trigger", func() {
			rl := RateLimitAuthenticationOOBOTPEmailTrigger
			Convey("set in config", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          authentication:
            rate_limits:
              oob_otp:
                email:
                  trigger_per_ip:
                    burst: 1
                    enabled: true
                    period: 1m
                  trigger_per_user:
                    burst: 2
                    enabled: true
                    period: 2m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					UserID:    userID,
					IPAddress: ipAddress,
					Target:    target,
					Purpose:   purpose,
					Channel:   model.AuthenticatorOOBChannelEmail,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitAuthenticationOOBOTPEmailTriggerPerIP,
					RateLimitGroup: rl,
					Name:           OOBOTPTriggerEmailPerIP,
					Arguments:      []string{purpose, ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("1m").Duration(),
					Burst:          1,
				}, {
					RateLimitName:  RateLimitAuthenticationOOBOTPEmailTriggerPerUser,
					RateLimitGroup: rl,
					Name:           OOBOTPTriggerEmailPerUser,
					Arguments:      []string{userID, purpose},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("2m").Duration(),
					Burst:          2,
				}})
			})
		})

		Convey("authentication.oob_otp.email.validate", func() {
			rl := RateLimitAuthenticationOOBOTPEmailValidate
			Convey("set in config", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          authentication:
            rate_limits:
              oob_otp:
                email:
                  validate_per_ip:
                    burst: 1
                    enabled: true
                    period: 1m
                  validate_per_user_per_ip:
                    burst: 2
                    enabled: true
                    period: 2m
              general:
                per_ip:
                  burst: 3
                  enabled: true
                  period: 3m
                per_user_per_ip:
                  burst: 4
                  enabled: true
                  period: 4m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					UserID:    userID,
					IPAddress: ipAddress,
					Target:    target,
					Purpose:   purpose,
					Channel:   model.AuthenticatorOOBChannelEmail,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitAuthenticationOOBOTPEmailValidatePerIP,
					RateLimitGroup: rl,
					Name:           OOBOTPValidateEmailPerIP,
					Arguments:      []string{purpose, ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("1m").Duration(),
					Burst:          1,
				}, {
					RateLimitName:  RateLimitAuthenticationOOBOTPEmailValidatePerUserPerIP,
					RateLimitGroup: rl,
					Name:           OOBOTPValidateEmailPerUserPerIP,
					Arguments:      []string{userID, ipAddress, purpose},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("2m").Duration(),
					Burst:          2,
				}})
			})

			Convey("fallback to authentication.general when not set", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          authentication:
            rate_limits:
              general:
                per_ip:
                  burst: 3
                  enabled: true
                  period: 3m
                per_user_per_ip:
                  burst: 4
                  enabled: true
                  period: 4m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					UserID:    userID,
					IPAddress: ipAddress,
					Target:    target,
					Purpose:   purpose,
					Channel:   model.AuthenticatorOOBChannelEmail,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitAuthenticationOOBOTPEmailValidatePerIP,
					RateLimitGroup: rl,
					Name:           OOBOTPValidateEmailPerIP,
					Arguments:      []string{purpose, ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("3m").Duration(),
					Burst:          3,
				}, {
					RateLimitName:  RateLimitAuthenticationOOBOTPEmailValidatePerUserPerIP,
					RateLimitGroup: rl,
					Name:           OOBOTPValidateEmailPerUserPerIP,
					Arguments:      []string{userID, ipAddress, purpose},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("4m").Duration(),
					Burst:          4,
				}})
			})
			Convey("disabled in config should not fallback", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          authentication:
            rate_limits:
              oob_otp:
                email:
                  validate_per_ip:
                    enabled: false
                  validate_per_user_per_ip:
                    enabled: false
              general:
                per_ip:
                  burst: 3
                  enabled: true
                  period: 3m
                per_user_per_ip:
                  burst: 4
                  enabled: true
                  period: 4m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					UserID:    userID,
					IPAddress: ipAddress,
					Target:    target,
					Purpose:   purpose,
					Channel:   model.AuthenticatorOOBChannelEmail,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitAuthenticationOOBOTPEmailValidatePerIP,
					RateLimitGroup: rl,
					Name:           OOBOTPValidateEmailPerIP,
					Arguments:      []string{purpose, ipAddress},
					IsGlobal:       false,
					Enabled:        false,
					Period:         time.Duration(0),
					Burst:          0,
				}, {
					RateLimitName:  RateLimitAuthenticationOOBOTPEmailValidatePerUserPerIP,
					RateLimitGroup: rl,
					Name:           OOBOTPValidateEmailPerUserPerIP,
					Arguments:      []string{userID, ipAddress, purpose},
					IsGlobal:       false,
					Enabled:        false,
					Period:         time.Duration(0),
					Burst:          0,
				}})
			})
		})

		Convey("authentication.oob_otp.sms.trigger", func() {
			rl := RateLimitAuthenticationOOBOTPSMSTrigger
			Convey("set in config", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          authentication:
            rate_limits:
              oob_otp:
                sms:
                  trigger_per_ip:
                    burst: 1
                    enabled: true
                    period: 1m
                  trigger_per_user:
                    burst: 2
                    enabled: true
                    period: 2m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					UserID:    userID,
					IPAddress: ipAddress,
					Target:    phone,
					Purpose:   purpose,
					Channel:   model.AuthenticatorOOBChannelSMS,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitAuthenticationOOBOTPSMSTriggerPerIP,
					RateLimitGroup: rl,
					Name:           OOBOTPTriggerSMSPerIP,
					Arguments:      []string{purpose, ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("1m").Duration(),
					Burst:          1,
				}, {
					RateLimitName:  RateLimitAuthenticationOOBOTPSMSTriggerPerUser,
					RateLimitGroup: rl,
					Name:           OOBOTPTriggerSMSPerUser,
					Arguments:      []string{userID, purpose},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("2m").Duration(),
					Burst:          2,
				}})
			})
		})

		Convey("authentication.oob_otp.sms.validate", func() {
			rl := RateLimitAuthenticationOOBOTPSMSValidate
			Convey("set in config", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          authentication:
            rate_limits:
              oob_otp:
                sms:
                  validate_per_ip:
                    burst: 1
                    enabled: true
                    period: 1m
                  validate_per_user_per_ip:
                    burst: 2
                    enabled: true
                    period: 2m
              general:
                per_ip:
                  burst: 3
                  enabled: true
                  period: 3m
                per_user_per_ip:
                  burst: 4
                  enabled: true
                  period: 4m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					UserID:    userID,
					IPAddress: ipAddress,
					Target:    phone,
					Purpose:   purpose,
					Channel:   model.AuthenticatorOOBChannelSMS,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitAuthenticationOOBOTPSMSValidatePerIP,
					RateLimitGroup: rl,
					Name:           OOBOTPValidateSMSPerIP,
					Arguments:      []string{purpose, ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("1m").Duration(),
					Burst:          1,
				}, {
					RateLimitName:  RateLimitAuthenticationOOBOTPSMSValidatePerUserPerIP,
					RateLimitGroup: rl,
					Name:           OOBOTPValidateSMSPerUserPerIP,
					Arguments:      []string{userID, ipAddress, purpose},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("2m").Duration(),
					Burst:          2,
				}})
			})

			Convey("fallback to authentication.general when not set", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          authentication:
            rate_limits:
              general:
                per_ip:
                  burst: 3
                  enabled: true
                  period: 3m
                per_user_per_ip:
                  burst: 4
                  enabled: true
                  period: 4m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					UserID:    userID,
					IPAddress: ipAddress,
					Target:    phone,
					Purpose:   purpose,
					Channel:   model.AuthenticatorOOBChannelSMS,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitAuthenticationOOBOTPSMSValidatePerIP,
					RateLimitGroup: rl,
					Name:           OOBOTPValidateSMSPerIP,
					Arguments:      []string{purpose, ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("3m").Duration(),
					Burst:          3,
				}, {
					RateLimitName:  RateLimitAuthenticationOOBOTPSMSValidatePerUserPerIP,
					RateLimitGroup: rl,
					Name:           OOBOTPValidateSMSPerUserPerIP,
					Arguments:      []string{userID, ipAddress, purpose},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("4m").Duration(),
					Burst:          4,
				}})
			})
			Convey("disabled in config should not fallback", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          authentication:
            rate_limits:
              oob_otp:
                sms:
                  validate_per_ip:
                    enabled: false
                  validate_per_user_per_ip:
                    enabled: false
              general:
                per_ip:
                  burst: 3
                  enabled: true
                  period: 3m
                per_user_per_ip:
                  burst: 4
                  enabled: true
                  period: 4m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					UserID:    userID,
					IPAddress: ipAddress,
					Target:    phone,
					Purpose:   purpose,
					Channel:   model.AuthenticatorOOBChannelSMS,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitAuthenticationOOBOTPSMSValidatePerIP,
					RateLimitGroup: rl,
					Name:           OOBOTPValidateSMSPerIP,
					Arguments:      []string{purpose, ipAddress},
					IsGlobal:       false,
					Enabled:        false,
					Period:         time.Duration(0),
					Burst:          0,
				}, {
					RateLimitName:  RateLimitAuthenticationOOBOTPSMSValidatePerUserPerIP,
					RateLimitGroup: rl,
					Name:           OOBOTPValidateSMSPerUserPerIP,
					Arguments:      []string{userID, ipAddress, purpose},
					IsGlobal:       false,
					Enabled:        false,
					Period:         time.Duration(0),
					Burst:          0,
				}})
			})
		})

		Convey("authentication.totp", func() {
			rl := RateLimitAuthenticationTOTP
			Convey("set in config", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          authentication:
            rate_limits:
              totp:
                per_ip:
                  burst: 1
                  enabled: true
                  period: 1m
                per_user_per_ip:
                  burst: 2
                  enabled: true
                  period: 2m
              general:
                per_ip:
                  burst: 3
                  enabled: true
                  period: 3m
                per_user_per_ip:
                  burst: 4
                  enabled: true
                  period: 4m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					UserID:    userID,
					IPAddress: ipAddress,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitAuthenticationTOTPPerIP,
					RateLimitGroup: rl,
					Name:           VerifyTOTPPerIP,
					Arguments:      []string{ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("1m").Duration(),
					Burst:          1,
				}, {
					RateLimitName:  RateLimitAuthenticationTOTPPerUserPerIP,
					RateLimitGroup: rl,
					Name:           VerifyTOTPPerUserPerIP,
					Arguments:      []string{userID, ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("2m").Duration(),
					Burst:          2,
				}})
			})

			Convey("fallback to authentication.general when not set", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          authentication:
            rate_limits:
              general:
                per_ip:
                  burst: 3
                  enabled: true
                  period: 3m
                per_user_per_ip:
                  burst: 4
                  enabled: true
                  period: 4m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					UserID:    userID,
					IPAddress: ipAddress,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitAuthenticationTOTPPerIP,
					RateLimitGroup: rl,
					Name:           VerifyTOTPPerIP,
					Arguments:      []string{ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("3m").Duration(),
					Burst:          3,
				}, {
					RateLimitName:  RateLimitAuthenticationTOTPPerUserPerIP,
					RateLimitGroup: rl,
					Name:           VerifyTOTPPerUserPerIP,
					Arguments:      []string{userID, ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("4m").Duration(),
					Burst:          4,
				}})
			})
			Convey("disabled in config should not fallback", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          authentication:
            rate_limits:
              totp:
                per_ip:
                  enabled: false
                per_user_per_ip:
                  enabled: false
              general:
                per_ip:
                  burst: 3
                  enabled: true
                  period: 3m
                per_user_per_ip:
                  burst: 4
                  enabled: true
                  period: 4m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					UserID:    userID,
					IPAddress: ipAddress,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitAuthenticationTOTPPerIP,
					RateLimitGroup: rl,
					Name:           VerifyTOTPPerIP,
					Arguments:      []string{ipAddress},
					IsGlobal:       false,
					Enabled:        false,
					Period:         time.Duration(0),
					Burst:          0,
				}, {
					RateLimitName:  RateLimitAuthenticationTOTPPerUserPerIP,
					RateLimitGroup: rl,
					Name:           VerifyTOTPPerUserPerIP,
					Arguments:      []string{userID, ipAddress},
					IsGlobal:       false,
					Enabled:        false,
					Period:         time.Duration(0),
					Burst:          0,
				}})
			})
		})

		Convey("authentication.recovery_code", func() {
			rl := RateLimitAuthenticationRecoveryCode
			Convey("set in config", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          authentication:
            rate_limits:
              recovery_code:
                per_ip:
                  burst: 1
                  enabled: true
                  period: 1m
                per_user_per_ip:
                  burst: 2
                  enabled: true
                  period: 2m
              general:
                per_ip:
                  burst: 3
                  enabled: true
                  period: 3m
                per_user_per_ip:
                  burst: 4
                  enabled: true
                  period: 4m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					UserID:    userID,
					IPAddress: ipAddress,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitAuthenticationRecoveryCodePerIP,
					RateLimitGroup: rl,
					Name:           VerifyRecoveryCodePerIP,
					Arguments:      []string{ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("1m").Duration(),
					Burst:          1,
				}, {
					RateLimitName:  RateLimitAuthenticationRecoveryCodePerUserPerIP,
					RateLimitGroup: rl,
					Name:           VerifyRecoveryCodePerUserPerIP,
					Arguments:      []string{userID, ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("2m").Duration(),
					Burst:          2,
				}})
			})

			Convey("fallback to authentication.general when not set", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          authentication:
            rate_limits:
              general:
                per_ip:
                  burst: 3
                  enabled: true
                  period: 3m
                per_user_per_ip:
                  burst: 4
                  enabled: true
                  period: 4m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					UserID:    userID,
					IPAddress: ipAddress,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitAuthenticationRecoveryCodePerIP,
					RateLimitGroup: rl,
					Name:           VerifyRecoveryCodePerIP,
					Arguments:      []string{ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("3m").Duration(),
					Burst:          3,
				}, {
					RateLimitName:  RateLimitAuthenticationRecoveryCodePerUserPerIP,
					RateLimitGroup: rl,
					Name:           VerifyRecoveryCodePerUserPerIP,
					Arguments:      []string{userID, ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("4m").Duration(),
					Burst:          4,
				}})
			})
			Convey("disabled in config should not fallback", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          authentication:
            rate_limits:
              recovery_code:
                per_ip:
                  enabled: false
                per_user_per_ip:
                  enabled: false
              general:
                per_ip:
                  burst: 3
                  enabled: true
                  period: 3m
                per_user_per_ip:
                  burst: 4
                  enabled: true
                  period: 4m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					UserID:    userID,
					IPAddress: ipAddress,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitAuthenticationRecoveryCodePerIP,
					RateLimitGroup: rl,
					Name:           VerifyRecoveryCodePerIP,
					Arguments:      []string{ipAddress},
					IsGlobal:       false,
					Enabled:        false,
					Period:         time.Duration(0),
					Burst:          0,
				}, {
					RateLimitName:  RateLimitAuthenticationRecoveryCodePerUserPerIP,
					RateLimitGroup: rl,
					Name:           VerifyRecoveryCodePerUserPerIP,
					Arguments:      []string{userID, ipAddress},
					IsGlobal:       false,
					Enabled:        false,
					Period:         time.Duration(0),
					Burst:          0,
				}})
			})
		})

		Convey("authentication.device_token", func() {
			rl := RateLimitAuthenticationDeviceToken
			Convey("set in config", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          authentication:
            rate_limits:
              device_token:
                per_ip:
                  burst: 1
                  enabled: true
                  period: 1m
                per_user_per_ip:
                  burst: 2
                  enabled: true
                  period: 2m
              general:
                per_ip:
                  burst: 3
                  enabled: true
                  period: 3m
                per_user_per_ip:
                  burst: 4
                  enabled: true
                  period: 4m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					UserID:    userID,
					IPAddress: ipAddress,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitAuthenticationDeviceTokenPerIP,
					RateLimitGroup: rl,
					Name:           VerifyDeviceTokenPerIP,
					Arguments:      []string{ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("1m").Duration(),
					Burst:          1,
				}, {
					RateLimitName:  RateLimitAuthenticationDeviceTokenPerUserPerIP,
					RateLimitGroup: rl,
					Name:           VerifyDeviceTokenPerUserPerIP,
					Arguments:      []string{userID, ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("2m").Duration(),
					Burst:          2,
				}})
			})

			Convey("fallback to authentication.general when not set", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          authentication:
            rate_limits:
              general:
                per_ip:
                  burst: 3
                  enabled: true
                  period: 3m
                per_user_per_ip:
                  burst: 4
                  enabled: true
                  period: 4m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					UserID:    userID,
					IPAddress: ipAddress,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitAuthenticationDeviceTokenPerIP,
					RateLimitGroup: rl,
					Name:           VerifyDeviceTokenPerIP,
					Arguments:      []string{ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("3m").Duration(),
					Burst:          3,
				}, {
					RateLimitName:  RateLimitAuthenticationDeviceTokenPerUserPerIP,
					RateLimitGroup: rl,
					Name:           VerifyDeviceTokenPerUserPerIP,
					Arguments:      []string{userID, ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("4m").Duration(),
					Burst:          4,
				}})
			})
			Convey("disabled in config should not fallback", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          authentication:
            rate_limits:
              device_token:
                per_ip:
                  enabled: false
                per_user_per_ip:
                  enabled: false
              general:
                per_ip:
                  burst: 3
                  enabled: true
                  period: 3m
                per_user_per_ip:
                  burst: 4
                  enabled: true
                  period: 4m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					UserID:    userID,
					IPAddress: ipAddress,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitAuthenticationDeviceTokenPerIP,
					RateLimitGroup: rl,
					Name:           VerifyDeviceTokenPerIP,
					Arguments:      []string{ipAddress},
					IsGlobal:       false,
					Enabled:        false,
					Period:         time.Duration(0),
					Burst:          0,
				}, {
					RateLimitName:  RateLimitAuthenticationDeviceTokenPerUserPerIP,
					RateLimitGroup: rl,
					Name:           VerifyDeviceTokenPerUserPerIP,
					Arguments:      []string{userID, ipAddress},
					IsGlobal:       false,
					Enabled:        false,
					Period:         time.Duration(0),
					Burst:          0,
				}})
			})
		})

		Convey("authentication.passkey", func() {
			rl := RateLimitAuthenticationPasskey
			Convey("set in config", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          authentication:
            rate_limits:
              passkey:
                per_ip:
                  burst: 1
                  enabled: true
                  period: 1m
              general:
                per_ip:
                  burst: 3
                  enabled: true
                  period: 3m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					IPAddress: ipAddress,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitAuthenticationPasskeyPerIP,
					RateLimitGroup: rl,
					Name:           VerifyPasskeyPerIP,
					Arguments:      []string{ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("1m").Duration(),
					Burst:          1,
				}})
			})

			Convey("fallback to authentication.general when not set", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          authentication:
            rate_limits:
              general:
                per_ip:
                  burst: 3
                  enabled: true
                  period: 3m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					IPAddress: ipAddress,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitAuthenticationPasskeyPerIP,
					RateLimitGroup: rl,
					Name:           VerifyPasskeyPerIP,
					Arguments:      []string{ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("3m").Duration(),
					Burst:          3,
				}})
			})
			Convey("disabled in config should not fallback", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          authentication:
            rate_limits:
              passkey:
                per_ip:
                  enabled: false
              general:
                per_ip:
                  burst: 3
                  enabled: true
                  period: 3m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					IPAddress: ipAddress,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitAuthenticationPasskeyPerIP,
					RateLimitGroup: rl,
					Name:           VerifyPasskeyPerIP,
					Arguments:      []string{ipAddress},
					IsGlobal:       false,
					Enabled:        false,
					Period:         time.Duration(0),
					Burst:          0,
				}})
			})
		})

		Convey("authentication.siwe", func() {
			rl := RateLimitAuthenticationSIWE
			Convey("set in config", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          authentication:
            rate_limits:
              siwe:
                per_ip:
                  burst: 1
                  enabled: true
                  period: 1m
              general:
                per_ip:
                  burst: 3
                  enabled: true
                  period: 3m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					IPAddress: ipAddress,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitAuthenticationSIWEPerIP,
					RateLimitGroup: rl,
					Name:           VerifySIWEPerIP,
					Arguments:      []string{ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("1m").Duration(),
					Burst:          1,
				}})
			})

			Convey("fallback to authentication.general when not set", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          authentication:
            rate_limits:
              general:
                per_ip:
                  burst: 3
                  enabled: true
                  period: 3m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					IPAddress: ipAddress,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitAuthenticationSIWEPerIP,
					RateLimitGroup: rl,
					Name:           VerifySIWEPerIP,
					Arguments:      []string{ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("3m").Duration(),
					Burst:          3,
				}})
			})
			Convey("disabled in config should not fallback", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          authentication:
            rate_limits:
              siwe:
                per_ip:
                  enabled: false
              general:
                per_ip:
                  burst: 3
                  enabled: true
                  period: 3m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					IPAddress: ipAddress,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitAuthenticationSIWEPerIP,
					RateLimitGroup: rl,
					Name:           VerifySIWEPerIP,
					Arguments:      []string{ipAddress},
					IsGlobal:       false,
					Enabled:        false,
					Period:         time.Duration(0),
					Burst:          0,
				}})
			})
		})

		Convey("authentication.signup", func() {
			rl := RateLimitAuthenticationSignup
			Convey("set in config", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          authentication:
            rate_limits:
              signup:
                per_ip:
                  burst: 1
                  enabled: true
                  period: 1m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					IPAddress: ipAddress,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitAuthenticationSignupPerIP,
					RateLimitGroup: rl,
					Name:           SignupPerIP,
					Arguments:      []string{ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("1m").Duration(),
					Burst:          1,
				}})
			})
		})

		Convey("authentication.signup_anonymous", func() {
			rl := RateLimitAuthenticationSignupAnonymous
			Convey("set in config", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          authentication:
            rate_limits:
              signup_anonymous:
                per_ip:
                  burst: 1
                  enabled: true
                  period: 1m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					IPAddress: ipAddress,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitAuthenticationSignupAnonymousPerIP,
					RateLimitGroup: rl,
					Name:           SignupAnonymousPerIP,
					Arguments:      []string{ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("1m").Duration(),
					Burst:          1,
				}})
			})
		})

		Convey("authentication.account_enumeration", func() {
			rl := RateLimitAuthenticationAccountEnumeration
			Convey("set in config", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          authentication:
            rate_limits:
              account_enumeration:
                per_ip:
                  burst: 1
                  enabled: true
                  period: 1m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					IPAddress: ipAddress,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitAuthenticationAccountEnumerationPerIP,
					RateLimitGroup: rl,
					Name:           AccountEnumerationPerIP,
					Arguments:      []string{ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("1m").Duration(),
					Burst:          1,
				}})
			})
		})

		Convey("verification.email.trigger", func() {
			rl := RateLimitVerificationEmailTrigger
			Convey("set in config", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          verification:
            rate_limits:
              email:
                trigger_per_ip:
                  burst: 1
                  enabled: true
                  period: 1m
                trigger_per_user:
                  burst: 2
                  enabled: true
                  period: 2m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					UserID:    userID,
					IPAddress: ipAddress,
					Channel:   model.AuthenticatorOOBChannelEmail,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitVerificationEmailTriggerPerIP,
					RateLimitGroup: rl,
					Name:           VerificationTriggerEmailPerIP,
					Arguments:      []string{ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("1m").Duration(),
					Burst:          1,
				}, {
					RateLimitName:  RateLimitVerificationEmailTriggerPerUser,
					RateLimitGroup: rl,
					Name:           VerificationTriggerEmailPerUser,
					Arguments:      []string{userID},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("2m").Duration(),
					Burst:          2,
				}})
			})
		})

		Convey("verification.email.validate", func() {
			rl := RateLimitVerificationEmailValidate
			Convey("set in config", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          verification:
            rate_limits:
              email:
                validate_per_ip:
                  burst: 1
                  enabled: true
                  period: 1m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					IPAddress: ipAddress,
					Channel:   model.AuthenticatorOOBChannelEmail,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitVerificationEmailValidatePerIP,
					RateLimitGroup: rl,
					Name:           VerificationValidateEmailPerIP,
					Arguments:      []string{ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("1m").Duration(),
					Burst:          1,
				}})
			})
		})

		Convey("verification.sms.trigger", func() {
			rl := RateLimitVerificationSMSTrigger
			Convey("set in config", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          verification:
            rate_limits:
              sms:
                trigger_per_ip:
                  burst: 1
                  enabled: true
                  period: 1m
                trigger_per_user:
                  burst: 2
                  enabled: true
                  period: 2m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					UserID:    userID,
					IPAddress: ipAddress,
					Channel:   model.AuthenticatorOOBChannelSMS,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitVerificationSMSTriggerPerIP,
					RateLimitGroup: rl,
					Name:           VerificationTriggerSMSPerIP,
					Arguments:      []string{ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("1m").Duration(),
					Burst:          1,
				}, {
					RateLimitName:  RateLimitVerificationSMSTriggerPerUser,
					RateLimitGroup: rl,
					Name:           VerificationTriggerSMSPerUser,
					Arguments:      []string{userID},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("2m").Duration(),
					Burst:          2,
				}})
			})
		})

		Convey("verification.sms.validate", func() {
			rl := RateLimitVerificationSMSValidate
			Convey("set in config", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          verification:
            rate_limits:
              sms:
                validate_per_ip:
                  burst: 1
                  enabled: true
                  period: 1m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					IPAddress: ipAddress,
					Channel:   model.AuthenticatorOOBChannelSMS,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitVerificationSMSValidatePerIP,
					RateLimitGroup: rl,
					Name:           VerificationValidateSMSPerIP,
					Arguments:      []string{ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("1m").Duration(),
					Burst:          1,
				}})
			})
		})

		Convey("forgot_password.email.trigger", func() {
			rl := RateLimitForgotPasswordEmailTrigger
			Convey("set in config", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          forgot_password:
            rate_limits:
              email:
                trigger_per_ip:
                  burst: 1
                  enabled: true
                  period: 1m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					IPAddress: ipAddress,
					Channel:   model.AuthenticatorOOBChannelEmail,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitForgotPasswordEmailTriggerPerIP,
					RateLimitGroup: rl,
					Name:           ForgotPasswordTriggerEmailPerIP,
					Arguments:      []string{ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("1m").Duration(),
					Burst:          1,
				}})
			})
		})

		Convey("forgot_password.email.validate", func() {
			rl := RateLimitForgotPasswordEmailValidate
			Convey("set in config", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          forgot_password:
            rate_limits:
              email:
                validate_per_ip:
                  burst: 1
                  enabled: true
                  period: 1m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					IPAddress: ipAddress,
					Channel:   model.AuthenticatorOOBChannelEmail,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitForgotPasswordEmailValidatePerIP,
					RateLimitGroup: rl,
					Name:           ForgotPasswordValidateEmailPerIP,
					Arguments:      []string{ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("1m").Duration(),
					Burst:          1,
				}})
			})
		})

		Convey("forgot_password.sms.trigger", func() {
			rl := RateLimitForgotPasswordSMSTrigger
			Convey("set in config", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          forgot_password:
            rate_limits:
              sms:
                trigger_per_ip:
                  burst: 1
                  enabled: true
                  period: 1m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, nil, &ResolveBucketSpecOptions{
					IPAddress: ipAddress,
					Channel:   model.AuthenticatorOOBChannelSMS,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitForgotPasswordSMSTriggerPerIP,
					RateLimitGroup: rl,
					Name:           ForgotPasswordTriggerSMSPerIP,
					Arguments:      []string{ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("1m").Duration(),
					Burst:          1,
				}})
			})
		})

		Convey("forgot_password.sms.validate", func() {
			rl := RateLimitForgotPasswordSMSValidate
			Convey("set in config", func() {
				cfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          forgot_password:
            rate_limits:
              sms:
                validate_per_ip:
                  burst: 1
                  enabled: true
                  period: 1m
        `))
				So(err, ShouldBeNil)
				specs := rl.ResolveBucketSpecs(cfg, nil, &config.RateLimitsEnvironmentConfig{}, &ResolveBucketSpecOptions{
					IPAddress: ipAddress,
					Channel:   model.AuthenticatorOOBChannelSMS,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitForgotPasswordSMSValidatePerIP,
					RateLimitGroup: rl,
					Name:           ForgotPasswordValidateSMSPerIP,
					Arguments:      []string{ipAddress},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("1m").Duration(),
					Burst:          1,
				}})
			})
		})

		Convey("messaging.sms", func() {
			rl := RateLimitMessagingSMS
			Convey("set in all configs", func() {
				appCfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          messaging:
            rate_limits:
              sms:
                burst: 10
                enabled: true
                period: 1m
              sms_per_ip:
                burst: 20
                enabled: true
                period: 1m
              sms_per_target:
                burst: 30
                enabled: true
                period: 1m
        `))
				So(err, ShouldBeNil)

				// featureCfg rate is smaller, so it will be used
				featureCfg := &config.FeatureConfig{Messaging: &config.MessagingFeatureConfig{
					RateLimits: &config.MessagingRateLimitsFeatureConfig{
						SMS:          &config.RateLimitConfig{Enabled: boolPtr(true), Period: "1m", Burst: 1}, // rate 1/m < 10/m
						SMSPerIP:     &config.RateLimitConfig{Enabled: boolPtr(true), Period: "1m", Burst: 2}, // rate 2/m < 20/m
						SMSPerTarget: &config.RateLimitConfig{Enabled: boolPtr(true), Period: "1m", Burst: 3}, // rate 3/m < 30/m
					},
				}}

				envCfg := &config.RateLimitsEnvironmentConfig{
					SMS:          config.RateLimitsEnvironmentConfigEntry{Enabled: true, Burst: 4, Period: config.DurationString("1m").Duration()},
					SMSPerIP:     config.RateLimitsEnvironmentConfigEntry{Enabled: true, Burst: 5, Period: config.DurationString("1m").Duration()},
					SMSPerTarget: config.RateLimitsEnvironmentConfigEntry{Enabled: true, Burst: 6, Period: config.DurationString("1m").Duration()},
				}

				specs := rl.ResolveBucketSpecs(appCfg, featureCfg, envCfg, &ResolveBucketSpecOptions{
					IPAddress: ipAddress,
					Target:    phone,
				})

				So(specs, ShouldResemble, []*BucketSpec{
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingSMS,
						Arguments:      nil,
						IsGlobal:       false,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          1,
					},
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingSMS,
						Arguments:      nil,
						IsGlobal:       true,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          4,
					},
					{
						RateLimitName:  RateLimitMessagingSMSPerIP,
						RateLimitGroup: rl,
						Name:           MessagingSMSPerIP,
						Arguments:      []string{ipAddress},
						IsGlobal:       false,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          2,
					},
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingSMSPerIP,
						Arguments:      []string{ipAddress},
						IsGlobal:       true,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          5,
					},
					{
						RateLimitName:  RateLimitMessagingSMSPerTarget,
						RateLimitGroup: rl,
						Name:           MessagingSMSPerTarget,
						Arguments:      []string{phone},
						IsGlobal:       false,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          3,
					},
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingSMSPerTarget,
						Arguments:      []string{phone},
						IsGlobal:       true,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          6,
					},
				})
			})
			Convey("feature config rate is larger", func() {
				appCfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          messaging:
            rate_limits:
              sms:
                burst: 10
                enabled: true
                period: 1m
              sms_per_ip:
                burst: 20
                enabled: true
                period: 1m
              sms_per_target:
                burst: 30
                enabled: true
                period: 1m
        `))
				So(err, ShouldBeNil)

				// featureCfg rate is larger, so it will NOT be used
				featureCfg := &config.FeatureConfig{Messaging: &config.MessagingFeatureConfig{
					RateLimits: &config.MessagingRateLimitsFeatureConfig{
						SMS:          &config.RateLimitConfig{Enabled: boolPtr(true), Period: "1m", Burst: 100}, // rate 100/m > 10/m
						SMSPerIP:     &config.RateLimitConfig{Enabled: boolPtr(true), Period: "1m", Burst: 200}, // rate 200/m > 20/m
						SMSPerTarget: &config.RateLimitConfig{Enabled: boolPtr(true), Period: "1m", Burst: 300}, // rate 300/m > 30/m
					},
				}}

				envCfg := &config.RateLimitsEnvironmentConfig{
					SMS:          config.RateLimitsEnvironmentConfigEntry{Enabled: true, Burst: 4, Period: config.DurationString("1m").Duration()},
					SMSPerIP:     config.RateLimitsEnvironmentConfigEntry{Enabled: true, Burst: 5, Period: config.DurationString("1m").Duration()},
					SMSPerTarget: config.RateLimitsEnvironmentConfigEntry{Enabled: true, Burst: 6, Period: config.DurationString("1m").Duration()},
				}

				specs := rl.ResolveBucketSpecs(appCfg, featureCfg, envCfg, &ResolveBucketSpecOptions{
					IPAddress: ipAddress,
					Target:    phone,
				})

				So(specs, ShouldResemble, []*BucketSpec{
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingSMS,
						Arguments:      nil,
						IsGlobal:       false,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          10,
					},
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingSMS,
						Arguments:      nil,
						IsGlobal:       true,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          4,
					},
					{
						RateLimitName:  RateLimitMessagingSMSPerIP,
						RateLimitGroup: rl,
						Name:           MessagingSMSPerIP,
						Arguments:      []string{ipAddress},
						IsGlobal:       false,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          20,
					},
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingSMSPerIP,
						Arguments:      []string{ipAddress},
						IsGlobal:       true,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          5,
					},
					{
						RateLimitName:  RateLimitMessagingSMSPerTarget,
						RateLimitGroup: rl,
						Name:           MessagingSMSPerTarget,
						Arguments:      []string{phone},
						IsGlobal:       false,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          30,
					},
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingSMSPerTarget,
						Arguments:      []string{phone},
						IsGlobal:       true,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          6,
					},
				})
			})
			Convey("app config enabled, feature config disabled", func() {
				appCfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          messaging:
            rate_limits:
              sms:
                burst: 10
                enabled: true
                period: 1m
              sms_per_ip:
                burst: 20
                enabled: true
                period: 1m
              sms_per_target:
                burst: 30
                enabled: true
                period: 1m
        `))
				So(err, ShouldBeNil)

				featureCfg := &config.FeatureConfig{Messaging: &config.MessagingFeatureConfig{
					RateLimits: &config.MessagingRateLimitsFeatureConfig{
						SMS:          &config.RateLimitConfig{Enabled: boolPtr(false)},
						SMSPerIP:     &config.RateLimitConfig{Enabled: boolPtr(false)},
						SMSPerTarget: &config.RateLimitConfig{Enabled: boolPtr(false)},
					},
				}}

				envCfg := &config.RateLimitsEnvironmentConfig{
					SMS:          config.RateLimitsEnvironmentConfigEntry{Enabled: true, Burst: 4, Period: config.DurationString("1m").Duration()},
					SMSPerIP:     config.RateLimitsEnvironmentConfigEntry{Enabled: true, Burst: 5, Period: config.DurationString("1m").Duration()},
					SMSPerTarget: config.RateLimitsEnvironmentConfigEntry{Enabled: true, Burst: 6, Period: config.DurationString("1m").Duration()},
				}

				specs := rl.ResolveBucketSpecs(appCfg, featureCfg, envCfg, &ResolveBucketSpecOptions{
					IPAddress: ipAddress,
					Target:    phone,
				})

				So(specs, ShouldResemble, []*BucketSpec{
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingSMS,
						Arguments:      nil,
						IsGlobal:       false,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          10,
					},
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingSMS,
						Arguments:      nil,
						IsGlobal:       true,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          4,
					},
					{
						RateLimitName:  RateLimitMessagingSMSPerIP,
						RateLimitGroup: rl,
						Name:           MessagingSMSPerIP,
						Arguments:      []string{ipAddress},
						IsGlobal:       false,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          20,
					},
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingSMSPerIP,
						Arguments:      []string{ipAddress},
						IsGlobal:       true,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          5,
					},
					{
						RateLimitName:  RateLimitMessagingSMSPerTarget,
						RateLimitGroup: rl,
						Name:           MessagingSMSPerTarget,
						Arguments:      []string{phone},
						IsGlobal:       false,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          30,
					},
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingSMSPerTarget,
						Arguments:      []string{phone},
						IsGlobal:       true,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          6,
					},
				})
			})
			Convey("feature config enabled, app config disabled", func() {
				appCfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          messaging:
            rate_limits:
              sms:
                enabled: false
              sms_per_ip:
                enabled: false
              sms_per_target:
                enabled: false
        `))
				So(err, ShouldBeNil)

				featureCfg := &config.FeatureConfig{Messaging: &config.MessagingFeatureConfig{
					RateLimits: &config.MessagingRateLimitsFeatureConfig{
						SMS:          &config.RateLimitConfig{Enabled: boolPtr(true), Period: "1m", Burst: 1},
						SMSPerIP:     &config.RateLimitConfig{Enabled: boolPtr(true), Period: "1m", Burst: 2},
						SMSPerTarget: &config.RateLimitConfig{Enabled: boolPtr(true), Period: "1m", Burst: 3},
					},
				}}

				envCfg := &config.RateLimitsEnvironmentConfig{
					SMS:          config.RateLimitsEnvironmentConfigEntry{Enabled: true, Burst: 4, Period: config.DurationString("1m").Duration()},
					SMSPerIP:     config.RateLimitsEnvironmentConfigEntry{Enabled: true, Burst: 5, Period: config.DurationString("1m").Duration()},
					SMSPerTarget: config.RateLimitsEnvironmentConfigEntry{Enabled: true, Burst: 6, Period: config.DurationString("1m").Duration()},
				}

				specs := rl.ResolveBucketSpecs(appCfg, featureCfg, envCfg, &ResolveBucketSpecOptions{
					IPAddress: ipAddress,
					Target:    phone,
				})

				So(specs, ShouldResemble, []*BucketSpec{
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingSMS,
						Arguments:      nil,
						IsGlobal:       false,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          1,
					},
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingSMS,
						Arguments:      nil,
						IsGlobal:       true,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          4,
					},
					{
						RateLimitName:  RateLimitMessagingSMSPerIP,
						RateLimitGroup: rl,
						Name:           MessagingSMSPerIP,
						Arguments:      []string{ipAddress},
						IsGlobal:       false,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          2,
					},
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingSMSPerIP,
						Arguments:      []string{ipAddress},
						IsGlobal:       true,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          5,
					},
					{
						RateLimitName:  RateLimitMessagingSMSPerTarget,
						RateLimitGroup: rl,
						Name:           MessagingSMSPerTarget,
						Arguments:      []string{phone},
						IsGlobal:       false,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          3,
					},
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingSMSPerTarget,
						Arguments:      []string{phone},
						IsGlobal:       true,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          6,
					},
				})
			})
		})

		Convey("messaging.email", func() {
			rl := RateLimitMessagingEmail
			Convey("set in all configs", func() {
				appCfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          messaging:
            rate_limits:
              email:
                burst: 10
                enabled: true
                period: 1m
              email_per_ip:
                burst: 20
                enabled: true
                period: 1m
              email_per_target:
                burst: 30
                enabled: true
                period: 1m
        `))
				So(err, ShouldBeNil)

				// featureCfg rate is smaller, so it will be used
				featureCfg := &config.FeatureConfig{Messaging: &config.MessagingFeatureConfig{
					RateLimits: &config.MessagingRateLimitsFeatureConfig{
						Email:          &config.RateLimitConfig{Enabled: boolPtr(true), Period: "1m", Burst: 1}, // rate 1/m < 10/m
						EmailPerIP:     &config.RateLimitConfig{Enabled: boolPtr(true), Period: "1m", Burst: 2}, // rate 2/m < 20/m
						EmailPerTarget: &config.RateLimitConfig{Enabled: boolPtr(true), Period: "1m", Burst: 3}, // rate 3/m < 30/m
					},
				}}

				envCfg := &config.RateLimitsEnvironmentConfig{
					Email:          config.RateLimitsEnvironmentConfigEntry{Enabled: true, Burst: 4, Period: config.DurationString("1m").Duration()},
					EmailPerIP:     config.RateLimitsEnvironmentConfigEntry{Enabled: true, Burst: 5, Period: config.DurationString("1m").Duration()},
					EmailPerTarget: config.RateLimitsEnvironmentConfigEntry{Enabled: true, Burst: 6, Period: config.DurationString("1m").Duration()},
				}

				specs := rl.ResolveBucketSpecs(appCfg, featureCfg, envCfg, &ResolveBucketSpecOptions{
					IPAddress: ipAddress,
					Target:    target,
				})

				So(specs, ShouldResemble, []*BucketSpec{
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingEmail,
						Arguments:      nil,
						IsGlobal:       false,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          1,
					},
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingEmail,
						Arguments:      nil,
						IsGlobal:       true,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          4,
					},
					{
						RateLimitName:  RateLimitMessagingEmailPerIP,
						RateLimitGroup: rl,
						Name:           MessagingEmailPerIP,
						Arguments:      []string{ipAddress},
						IsGlobal:       false,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          2,
					},
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingEmailPerIP,
						Arguments:      []string{ipAddress},
						IsGlobal:       true,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          5,
					},
					{
						RateLimitName:  RateLimitMessagingEmailPerTarget,
						RateLimitGroup: rl,
						Name:           MessagingEmailPerTarget,
						Arguments:      []string{target},
						IsGlobal:       false,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          3,
					},
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingEmailPerTarget,
						Arguments:      []string{target},
						IsGlobal:       true,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          6,
					},
				})
			})
			Convey("feature config rate is larger", func() {
				appCfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          messaging:
            rate_limits:
              email:
                burst: 10
                enabled: true
                period: 1m
              email_per_ip:
                burst: 20
                enabled: true
                period: 1m
              email_per_target:
                burst: 30
                enabled: true
                period: 1m
        `))
				So(err, ShouldBeNil)

				// featureCfg rate is larger, so it will NOT be used
				featureCfg := &config.FeatureConfig{Messaging: &config.MessagingFeatureConfig{
					RateLimits: &config.MessagingRateLimitsFeatureConfig{
						Email:          &config.RateLimitConfig{Enabled: boolPtr(true), Period: "1m", Burst: 100}, // rate 100/m > 10/m
						EmailPerIP:     &config.RateLimitConfig{Enabled: boolPtr(true), Period: "1m", Burst: 200}, // rate 200/m > 20/m
						EmailPerTarget: &config.RateLimitConfig{Enabled: boolPtr(true), Period: "1m", Burst: 300}, // rate 300/m > 30/m
					},
				}}

				envCfg := &config.RateLimitsEnvironmentConfig{
					Email:          config.RateLimitsEnvironmentConfigEntry{Enabled: true, Burst: 4, Period: config.DurationString("1m").Duration()},
					EmailPerIP:     config.RateLimitsEnvironmentConfigEntry{Enabled: true, Burst: 5, Period: config.DurationString("1m").Duration()},
					EmailPerTarget: config.RateLimitsEnvironmentConfigEntry{Enabled: true, Burst: 6, Period: config.DurationString("1m").Duration()},
				}

				specs := rl.ResolveBucketSpecs(appCfg, featureCfg, envCfg, &ResolveBucketSpecOptions{
					IPAddress: ipAddress,
					Target:    target,
				})

				So(specs, ShouldResemble, []*BucketSpec{
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingEmail,
						Arguments:      nil,
						IsGlobal:       false,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          10,
					},
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingEmail,
						Arguments:      nil,
						IsGlobal:       true,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          4,
					},
					{
						RateLimitName:  RateLimitMessagingEmailPerIP,
						RateLimitGroup: rl,
						Name:           MessagingEmailPerIP,
						Arguments:      []string{ipAddress},
						IsGlobal:       false,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          20,
					},
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingEmailPerIP,
						Arguments:      []string{ipAddress},
						IsGlobal:       true,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          5,
					},
					{
						RateLimitName:  RateLimitMessagingEmailPerTarget,
						RateLimitGroup: rl,
						Name:           MessagingEmailPerTarget,
						Arguments:      []string{target},
						IsGlobal:       false,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          30,
					},
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingEmailPerTarget,
						Arguments:      []string{target},
						IsGlobal:       true,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          6,
					},
				})
			})
			Convey("app config enabled, feature config disabled", func() {
				appCfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          messaging:
            rate_limits:
              email:
                burst: 10
                enabled: true
                period: 1m
              email_per_ip:
                burst: 20
                enabled: true
                period: 1m
              email_per_target:
                burst: 30
                enabled: true
                period: 1m
        `))
				So(err, ShouldBeNil)

				featureCfg := &config.FeatureConfig{Messaging: &config.MessagingFeatureConfig{
					RateLimits: &config.MessagingRateLimitsFeatureConfig{
						Email:          &config.RateLimitConfig{Enabled: boolPtr(false)},
						EmailPerIP:     &config.RateLimitConfig{Enabled: boolPtr(false)},
						EmailPerTarget: &config.RateLimitConfig{Enabled: boolPtr(false)},
					},
				}}

				envCfg := &config.RateLimitsEnvironmentConfig{
					Email:          config.RateLimitsEnvironmentConfigEntry{Enabled: true, Burst: 4, Period: config.DurationString("1m").Duration()},
					EmailPerIP:     config.RateLimitsEnvironmentConfigEntry{Enabled: true, Burst: 5, Period: config.DurationString("1m").Duration()},
					EmailPerTarget: config.RateLimitsEnvironmentConfigEntry{Enabled: true, Burst: 6, Period: config.DurationString("1m").Duration()},
				}

				specs := rl.ResolveBucketSpecs(appCfg, featureCfg, envCfg, &ResolveBucketSpecOptions{
					IPAddress: ipAddress,
					Target:    target,
				})

				So(specs, ShouldResemble, []*BucketSpec{
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingEmail,
						Arguments:      nil,
						IsGlobal:       false,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          10,
					},
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingEmail,
						Arguments:      nil,
						IsGlobal:       true,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          4,
					},
					{
						RateLimitName:  RateLimitMessagingEmailPerIP,
						RateLimitGroup: rl,
						Name:           MessagingEmailPerIP,
						Arguments:      []string{ipAddress},
						IsGlobal:       false,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          20,
					},
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingEmailPerIP,
						Arguments:      []string{ipAddress},
						IsGlobal:       true,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          5,
					},
					{
						RateLimitName:  RateLimitMessagingEmailPerTarget,
						RateLimitGroup: rl,
						Name:           MessagingEmailPerTarget,
						Arguments:      []string{target},
						IsGlobal:       false,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          30,
					},
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingEmailPerTarget,
						Arguments:      []string{target},
						IsGlobal:       true,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          6,
					},
				})
			})
			Convey("feature config enabled, app config disabled", func() {
				appCfg, err := config.Parse(ctx, []byte(`
          id: test
          http:
            public_origin: http://test
          messaging:
            rate_limits:
              email:
                enabled: false
              email_per_ip:
                enabled: false
              email_per_target:
                enabled: false
        `))
				So(err, ShouldBeNil)

				featureCfg := &config.FeatureConfig{Messaging: &config.MessagingFeatureConfig{
					RateLimits: &config.MessagingRateLimitsFeatureConfig{
						Email:          &config.RateLimitConfig{Enabled: boolPtr(true), Period: "1m", Burst: 1},
						EmailPerIP:     &config.RateLimitConfig{Enabled: boolPtr(true), Period: "1m", Burst: 2},
						EmailPerTarget: &config.RateLimitConfig{Enabled: boolPtr(true), Period: "1m", Burst: 3},
					},
				}}

				envCfg := &config.RateLimitsEnvironmentConfig{
					Email:          config.RateLimitsEnvironmentConfigEntry{Enabled: true, Burst: 4, Period: config.DurationString("1m").Duration()},
					EmailPerIP:     config.RateLimitsEnvironmentConfigEntry{Enabled: true, Burst: 5, Period: config.DurationString("1m").Duration()},
					EmailPerTarget: config.RateLimitsEnvironmentConfigEntry{Enabled: true, Burst: 6, Period: config.DurationString("1m").Duration()},
				}

				specs := rl.ResolveBucketSpecs(appCfg, featureCfg, envCfg, &ResolveBucketSpecOptions{
					IPAddress: ipAddress,
					Target:    target,
				})

				So(specs, ShouldResemble, []*BucketSpec{
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingEmail,
						Arguments:      nil,
						IsGlobal:       false,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          1,
					},
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingEmail,
						Arguments:      nil,
						IsGlobal:       true,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          4,
					},
					{
						RateLimitName:  RateLimitMessagingEmailPerIP,
						RateLimitGroup: rl,
						Name:           MessagingEmailPerIP,
						Arguments:      []string{ipAddress},
						IsGlobal:       false,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          2,
					},
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingEmailPerIP,
						Arguments:      []string{ipAddress},
						IsGlobal:       true,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          5,
					},
					{
						RateLimitName:  RateLimitMessagingEmailPerTarget,
						RateLimitGroup: rl,
						Name:           MessagingEmailPerTarget,
						Arguments:      []string{target},
						IsGlobal:       false,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          3,
					},
					{
						RateLimitName:  "",
						RateLimitGroup: rl,
						Name:           MessagingEmailPerTarget,
						Arguments:      []string{target},
						IsGlobal:       true,
						Enabled:        true,
						Period:         config.DurationString("1m").Duration(),
						Burst:          6,
					},
				})
			})
		})

		Convey("oauth.token.client_credentials", func() {
			rl := RateLimitOAuthTokenClientCredentials
			Convey("is enabled", func() {
				specs := rl.ResolveBucketSpecs(nil, nil, nil, &ResolveBucketSpecOptions{
					ClientID: clientID,
				})
				So(specs, ShouldResemble, []*BucketSpec{{
					RateLimitName:  RateLimitOAuthTokenClientCredentialsPerClient,
					RateLimitGroup: rl,
					Name:           OAuthTokenClientCredentialsPerClient,
					Arguments:      []string{clientID},
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("1m").Duration(),
					Burst:          5,
				}, {
					RateLimitName:  RateLimitOAuthTokenClientCredentialsPerProject,
					RateLimitGroup: rl,
					Name:           OAuthTokenClientCredentialsPerProject,
					Arguments:      nil,
					IsGlobal:       false,
					Enabled:        true,
					Period:         config.DurationString("1m").Duration(),
					Burst:          20,
				}})
			})
		})
	})

	Convey("RateLimit.ResolveWeight", t, func() {
		Convey("should return default weight when no weights are in context", func() {
			ctx := context.Background()
			weight := RateLimitAuthenticationPassword.ResolveWeight(ctx)
			So(weight, ShouldEqual, 1.0)
		})

		Convey("should return default weight for unnamed rate limit", func() {
			ctx := WithRateLimitWeights(context.Background())
			weights := map[RateLimitGroup]float64{
				RateLimitAuthenticationPassword: 2.0,
			}
			SetRateLimitWeights(ctx, weights)
			var r RateLimitGroup = ""
			weight := r.ResolveWeight(ctx)
			So(weight, ShouldEqual, 1.0)
		})

		Convey("should return specific weight if defined", func() {
			ctx := WithRateLimitWeights(context.Background())
			weights := map[RateLimitGroup]float64{
				RateLimitAuthenticationPassword: 2.0,
			}
			SetRateLimitWeights(ctx, weights)
			weight := RateLimitAuthenticationPassword.ResolveWeight(ctx)
			So(weight, ShouldEqual, 2.0)
		})

		Convey("should return fallback weight if specific not defined", func() {
			ctx := WithRateLimitWeights(context.Background())
			weights := map[RateLimitGroup]float64{
				RateLimitAuthenticationGeneral: 3.0,
			}
			SetRateLimitWeights(ctx, weights)
			// RateLimitAuthenticationPassword falls back to RateLimitAuthenticationGeneral
			weight := RateLimitAuthenticationPassword.ResolveWeight(ctx)
			So(weight, ShouldEqual, 3.0)
		})

		Convey("should return specific weight even if fallback is defined", func() {
			ctx := WithRateLimitWeights(context.Background())
			weights := map[RateLimitGroup]float64{
				RateLimitAuthenticationGeneral:  3.0,
				RateLimitAuthenticationPassword: 2.0,
			}
			SetRateLimitWeights(ctx, weights)
			weight := RateLimitAuthenticationPassword.ResolveWeight(ctx)
			So(weight, ShouldEqual, 2.0)
		})

		Convey("should return default weight without specific weight", func() {
			ctx := WithRateLimitWeights(context.Background())
			weights := map[RateLimitGroup]float64{
				RateLimitAuthenticationGeneral: 3.0,
			}
			SetRateLimitWeights(ctx, weights)
			// This one has no fallback
			weight := RateLimitVerificationEmailTrigger.ResolveWeight(ctx)
			So(weight, ShouldEqual, 1.0)
		})

		Convey("should return specific weight for rate limits without fallback", func() {
			ctx := WithRateLimitWeights(context.Background())
			weights := map[RateLimitGroup]float64{
				RateLimitVerificationEmailTrigger: 5.0,
			}
			SetRateLimitWeights(ctx, weights)
			weight := RateLimitVerificationEmailTrigger.ResolveWeight(ctx)
			So(weight, ShouldEqual, 5.0)
		})

		Convey("should return 0 for negative weight", func() {
			ctx := WithRateLimitWeights(context.Background())
			weights := map[RateLimitGroup]float64{
				RateLimitAuthenticationPassword: -2.0,
			}
			SetRateLimitWeights(ctx, weights)
			weight := RateLimitAuthenticationPassword.ResolveWeight(ctx)
			So(weight, ShouldEqual, 0)
		})

		Convey("should return 0 for negative fallback weight", func() {
			ctx := WithRateLimitWeights(context.Background())
			weights := map[RateLimitGroup]float64{
				RateLimitAuthenticationGeneral: -3.0,
			}
			SetRateLimitWeights(ctx, weights)
			weight := RateLimitAuthenticationPassword.ResolveWeight(ctx)
			So(weight, ShouldEqual, 0)
		})

		Convey("should return fallback weight for all applicable rate limits", func() {
			limitsWithFallback := []RateLimitGroup{
				RateLimitAuthenticationPassword,
				RateLimitAuthenticationOOBOTPEmailValidate,
				RateLimitAuthenticationOOBOTPSMSValidate,
				RateLimitAuthenticationTOTP,
				RateLimitAuthenticationRecoveryCode,
				RateLimitAuthenticationDeviceToken,
				RateLimitAuthenticationPasskey,
				RateLimitAuthenticationSIWE,
			}

			for _, limit := range limitsWithFallback {
				limit := limit
				Convey(string(limit), func() {
					ctx := WithRateLimitWeights(context.Background())
					weights := map[RateLimitGroup]float64{
						RateLimitAuthenticationGeneral: 3.0,
					}
					SetRateLimitWeights(ctx, weights)
					weight := limit.ResolveWeight(ctx)
					So(weight, ShouldEqual, 3.0)
				})
			}
		})
	})
}
