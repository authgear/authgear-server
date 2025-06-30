package ratelimit

import (
	"context"
	"testing"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRateLimits(t *testing.T) {
	Convey("RateLimit", t, func() {
		ctx := context.Background()
		userID := "testuserid"
		ipAddress := "1.2.3.4"
		target := "test@example.com"
		phone := "+1555000123"
		purpose := "testpurpose"
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
					Name:      VerifyPasswordPerIP,
					Arguments: []string{ipAddress},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("1m").Duration(),
					Burst:     1,
				}, {
					Name:      VerifyPasswordPerUserPerIP,
					Arguments: []string{userID, ipAddress},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("2m").Duration(),
					Burst:     2,
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
					Name:      VerifyPasswordPerIP,
					Arguments: []string{ipAddress},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("3m").Duration(),
					Burst:     3,
				}, {
					Name:      VerifyPasswordPerUserPerIP,
					Arguments: []string{userID, ipAddress},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("4m").Duration(),
					Burst:     4,
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
					Name:      OOBOTPTriggerEmailPerIP,
					Arguments: []string{purpose, ipAddress},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("1m").Duration(),
					Burst:     1,
				}, {
					Name:      OOBOTPTriggerEmailPerUser,
					Arguments: []string{userID, purpose},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("2m").Duration(),
					Burst:     2,
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
					Name:      OOBOTPValidateEmailPerIP,
					Arguments: []string{purpose, ipAddress},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("1m").Duration(),
					Burst:     1,
				}, {
					Name:      OOBOTPValidateEmailPerUserPerIP,
					Arguments: []string{userID, ipAddress, purpose},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("2m").Duration(),
					Burst:     2,
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
					Name:      OOBOTPValidateEmailPerIP,
					Arguments: []string{purpose, ipAddress},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("3m").Duration(),
					Burst:     3,
				}, {
					Name:      OOBOTPValidateEmailPerUserPerIP,
					Arguments: []string{userID, ipAddress, purpose},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("4m").Duration(),
					Burst:     4,
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
					Name:      OOBOTPTriggerSMSPerIP,
					Arguments: []string{purpose, ipAddress},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("1m").Duration(),
					Burst:     1,
				}, {
					Name:      OOBOTPTriggerSMSPerUser,
					Arguments: []string{userID, purpose},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("2m").Duration(),
					Burst:     2,
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
					Name:      OOBOTPValidateSMSPerIP,
					Arguments: []string{purpose, ipAddress},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("1m").Duration(),
					Burst:     1,
				}, {
					Name:      OOBOTPValidateSMSPerUserPerIP,
					Arguments: []string{userID, ipAddress, purpose},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("2m").Duration(),
					Burst:     2,
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
					Name:      OOBOTPValidateSMSPerIP,
					Arguments: []string{purpose, ipAddress},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("3m").Duration(),
					Burst:     3,
				}, {
					Name:      OOBOTPValidateSMSPerUserPerIP,
					Arguments: []string{userID, ipAddress, purpose},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("4m").Duration(),
					Burst:     4,
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
					Name:      VerifyTOTPPerIP,
					Arguments: []string{ipAddress},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("1m").Duration(),
					Burst:     1,
				}, {
					Name:      VerifyTOTPPerUserPerIP,
					Arguments: []string{userID, ipAddress},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("2m").Duration(),
					Burst:     2,
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
					Name:      VerifyTOTPPerIP,
					Arguments: []string{ipAddress},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("3m").Duration(),
					Burst:     3,
				}, {
					Name:      VerifyTOTPPerUserPerIP,
					Arguments: []string{userID, ipAddress},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("4m").Duration(),
					Burst:     4,
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
					Name:      VerifyRecoveryCodePerIP,
					Arguments: []string{ipAddress},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("1m").Duration(),
					Burst:     1,
				}, {
					Name:      VerifyRecoveryCodePerUserPerIP,
					Arguments: []string{userID, ipAddress},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("2m").Duration(),
					Burst:     2,
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
					Name:      VerifyRecoveryCodePerIP,
					Arguments: []string{ipAddress},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("3m").Duration(),
					Burst:     3,
				}, {
					Name:      VerifyRecoveryCodePerUserPerIP,
					Arguments: []string{userID, ipAddress},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("4m").Duration(),
					Burst:     4,
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
					Name:      VerifyDeviceTokenPerIP,
					Arguments: []string{ipAddress},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("1m").Duration(),
					Burst:     1,
				}, {
					Name:      VerifyDeviceTokenPerUserPerIP,
					Arguments: []string{userID, ipAddress},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("2m").Duration(),
					Burst:     2,
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
					Name:      VerifyDeviceTokenPerIP,
					Arguments: []string{ipAddress},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("3m").Duration(),
					Burst:     3,
				}, {
					Name:      VerifyDeviceTokenPerUserPerIP,
					Arguments: []string{userID, ipAddress},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("4m").Duration(),
					Burst:     4,
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
					Name:      VerifyPasskeyPerIP,
					Arguments: []string{ipAddress},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("1m").Duration(),
					Burst:     1,
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
					Name:      VerifyPasskeyPerIP,
					Arguments: []string{ipAddress},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("3m").Duration(),
					Burst:     3,
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
					Name:      VerifySIWEPerIP,
					Arguments: []string{ipAddress},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("1m").Duration(),
					Burst:     1,
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
					Name:      VerifySIWEPerIP,
					Arguments: []string{ipAddress},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("3m").Duration(),
					Burst:     3,
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
					Name:      SignupPerIP,
					Arguments: []string{ipAddress},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("1m").Duration(),
					Burst:     1,
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
					Name:      SignupAnonymousPerIP,
					Arguments: []string{ipAddress},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("1m").Duration(),
					Burst:     1,
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
					Name:      AccountEnumerationPerIP,
					Arguments: []string{ipAddress},
					IsGlobal:  false,
					Enabled:   true,
					Period:    config.DurationString("1m").Duration(),
					Burst:     1,
				}})
			})
		})
	})
}
