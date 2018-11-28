package config

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTenantConfig(t *testing.T) {
	Convey("Test TenantConfiguration", t, func() {
		Convey("should load tenant config correctly from env", func() {
			os.Clearenv()
			os.Setenv("API_KEY", "testapikey")
			os.Setenv("MASTER_KEY", "testmasterkey")
			os.Setenv("APP_NAME", "testappname")
			os.Setenv("WELCOME_EMAIL_ENABLED", "true")

			config, err := NewTenantConfigurationFromEnv(nil)

			So(err, ShouldBeNil)
			So(config.APIKey, ShouldEqual, "testapikey")
			So(config.MasterKey, ShouldEqual, "testmasterkey")
			So(config.AppName, ShouldEqual, "testappname")
			So(config.WelcomeEmail.Enabled, ShouldEqual, true)
		})

		Convey("should have default value when load from env", func() {
			os.Clearenv()
			os.Setenv("API_KEY", "testapikey")
			os.Setenv("MASTER_KEY", "testmasterkey")
			os.Setenv("APP_NAME", "testappname")

			config, err := NewTenantConfigurationFromEnv(nil)

			So(err, ShouldBeNil)
			So(config.CORSHost, ShouldEqual, "*")
			So(config.SMTP.Port, ShouldEqual, 25)
		})

		Convey("should use master key as token secret if it is not provided in env", func() {
			os.Clearenv()
			os.Setenv("API_KEY", "testapikey")
			os.Setenv("MASTER_KEY", "testmasterkey")
			os.Setenv("APP_NAME", "testappname")

			config, err := NewTenantConfigurationFromEnv(nil)

			So(err, ShouldBeNil)
			So(config.TokenStore.Secret, ShouldEqual, "testmasterkey")
		})

		Convey("should validate when load from env", func() {
			os.Clearenv()

			_, err := NewTenantConfigurationFromEnv(nil)

			So(err, ShouldBeError)
		})

		Convey("should have default value when load from JSON", func() {
			b := []byte(`{
				"API_KEY": "testapikey",
				"APP_NAME": "testappname",
				"MASTER_KEY": "testmasterkey",
				"WELCOME_EMAIL": {
					"ENABLED": true
				}
			}`)

			config := NewTenantConfiguration()
			err := json.Unmarshal(b, &config)

			So(err, ShouldBeNil)
			So(config.APIKey, ShouldEqual, "testapikey")
			So(config.MasterKey, ShouldEqual, "testmasterkey")
			So(config.AppName, ShouldEqual, "testappname")
			So(config.WelcomeEmail.Enabled, ShouldEqual, true)
		})

		Convey("should load tenant config correctly from JSON", func() {
			b := []byte(`{
				"API_KEY": "testapikey",
				"APP_NAME": "testappname",
				"MASTER_KEY": "testmasterkey"
			}`)

			config := NewTenantConfiguration()
			err := json.Unmarshal(b, &config)

			So(err, ShouldBeNil)
			So(config.CORSHost, ShouldEqual, "*")
			So(config.SMTP.Port, ShouldEqual, 25)
		})

		Convey("should use master key as token secret if it is not provided in JSON", func() {
			b := []byte(`{
				"API_KEY": "testapikey",
				"APP_NAME": "testappname",
				"MASTER_KEY": "testmasterkey"
			}`)

			config := NewTenantConfiguration()
			err := json.Unmarshal(b, &config)

			So(err, ShouldBeNil)
			So(config.TokenStore.Secret, ShouldEqual, "testmasterkey")
		})

		Convey("should validate when load from JSON", func() {
			b := []byte(`{}`)

			config := NewTenantConfiguration()
			err := json.Unmarshal(b, &config)

			So(err, ShouldBeError)
		})

		Convey("should be the same config after set and get from header", func() {
			config := NewTenantConfiguration()
			config.APIKey = "testapikey"
			config.MasterKey = "testmasterkey"
			config.AppName = "testappname"

			req, _ := http.NewRequest("POST", "", nil)

			SetTenantConfig(req, config)
			newConfig := GetTenantConfig(req)
			So(newConfig, ShouldResemble, config)
		})
	})
}
