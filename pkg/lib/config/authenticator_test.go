package config_test

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestValidPeriodDeprecateFlow(t *testing.T) {
	Convey("Email Code Mode", t, func() {
		test := func(inputYAML string, mode string) {
			cfg, err := config.Parse([]byte(inputYAML))
			So(err, ShouldBeNil)

      var data []byte
      switch mode {
      case "code":
        data, err = os.ReadFile("testdata/code_validperiod_test.yaml")
        if err != nil {
          panic(err)
        }
      case "login_link":
        data, err = os.ReadFile("testdata/link_validperiod_test.yaml")
        if err != nil {
          panic(err)
        }
      default:
        data, err = os.ReadFile("testdata/default_validperiod_test.yaml")
        if err != nil {
          panic(err)
        }
      }

			defaultConfig, err := config.Parse(data)
			So(err, ShouldBeNil)

			So(cfg, ShouldResemble, defaultConfig)
		}

		// Both deprecated with email_otp_mode: code
		test(`
id: test
http:
  public_origin: http://test
authenticator:
  oob_otp:
    email:
      email_otp_mode: code
      maximum: 99
      code_valid_period: 321s
    sms:
      maximum: 99
      phone_otp_mode: whatsapp_sms
      code_valid_period: 321s
verification:
  code_valid_period: 321s
`, "code")

		// Both updated valid_periods with email_otp_mode: code
		test(`
id: test
http:
  public_origin: http://test
authenticator:
  oob_otp:
    email:
      email_otp_mode: code
      maximum: 99
      valid_periods:
        code: 321s
    sms:
      maximum: 99
      phone_otp_mode: whatsapp_sms
      valid_periods:
        code: 321s
verification:
  code_valid_period: 321s
`, "code")

		// One deprecated & One new with email_otp_mode: code
		test(`
id: test
http:
  public_origin: http://test
authenticator:
  oob_otp:
    email:
      email_otp_mode: code
      maximum: 99
      code_valid_period: 321s
    sms:
      maximum: 99
      phone_otp_mode: whatsapp_sms
      valid_periods:
        code: 321s
verification:
  code_valid_period: 321s
`, "code")

		// One deprecated & One new with email_otp_mode: code
		test(`
id: test
http:
  public_origin: http://test
authenticator:
  oob_otp:
    email:
      email_otp_mode: code
      maximum: 99
      valid_periods:
        code: 321s
    sms:
      maximum: 99
      phone_otp_mode: whatsapp_sms
      code_valid_period: 321s
verification:
  code_valid_period: 321s
`, "code")

		// Both deprecated with email_otp_mode: login_link
		test(`
id: test
http:
  public_origin: http://test
authenticator:
  oob_otp:
    email:
      email_otp_mode: login_link
      maximum: 99
      code_valid_period: 322s
    sms:
      maximum: 99
      phone_otp_mode: whatsapp_sms
      code_valid_period: 321s
verification:
  code_valid_period: 321s
`, "login_link")

		// Both updated valid_periods with email_otp_mode: login_link
		test(`
id: test
http:
  public_origin: http://test
authenticator:
  oob_otp:
    email:
      email_otp_mode: login_link
      maximum: 99
      valid_periods:
        link: 322s
    sms:
      maximum: 99
      phone_otp_mode: whatsapp_sms
      valid_periods:
        code: 321s
verification:
  code_valid_period: 321s
`, "login_link")

		// One deprecated One updated with email_otp_mode: login_link
		test(`
id: test
http:
  public_origin: http://test
authenticator:
  oob_otp:
    email:
      email_otp_mode: login_link
      maximum: 99
      code_valid_period: 322s
    sms:
      maximum: 99
      phone_otp_mode: whatsapp_sms
      valid_periods:
        code: 321s
verification:
  code_valid_period: 321s
`, "login_link")

		// One deprecated One updated with email_otp_mode: login_link
		test(`
id: test
http:
  public_origin: http://test
authenticator:
  oob_otp:
    email:
      email_otp_mode: login_link
      maximum: 99
      valid_periods:
        link: 322s
    sms:
      maximum: 99
      phone_otp_mode: whatsapp_sms
      code_valid_period: 321s
verification:
  code_valid_period: 321s
`, "login_link")

		test(`
id: test
http:
  public_origin: http://test
authenticator:
  oob_otp:
    email:
      email_otp_mode: code
      maximum: 99
    sms:
      maximum: 99
      phone_otp_mode: whatsapp_sms
verification:
  code_valid_period: 300s
`, "default")
	})

}
