package config

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/yaml.v2"

	. "github.com/skygeario/skygear-server/pkg/core/skytest"
)

const inputMinimalYAML = `version: '1'
app_name: myapp
app_config:
  database_url: postgres://
user_config:
  api_key: apikey
  master_key: masterkey
`

const inputMinimalJSON = `
{
  "version": "1",
  "app_name": "myapp",
  "app_config": {
    "database_url": "postgres://"
  },
  "user_config": {
    "api_key": "apikey",
    "master_key": "masterkey",
    "welcome_email": {
      "enabled": true
    }
  }
}
`

func TestTenantConfig(t *testing.T) {
	Convey("Test TenantConfiguration", t, func() {
		// YAML
		Convey("should load tenant config from YAML", func() {
			c, err := NewTenantConfigurationFromYAML(strings.NewReader(inputMinimalYAML))
			So(err, ShouldBeNil)

			So(c.Version, ShouldEqual, "1")
			So(c.AppName, ShouldEqual, "myapp")
			So(c.AppConfig.DatabaseURL, ShouldEqual, "postgres://")
			So(c.UserConfig.APIKey, ShouldEqual, "apikey")
			So(c.UserConfig.MasterKey, ShouldEqual, "masterkey")
		})
		Convey("should have default value when load from YAML", func() {
			c, err := NewTenantConfigurationFromYAML(strings.NewReader(inputMinimalYAML))
			So(err, ShouldBeNil)
			So(c.UserConfig.CORS.Origin, ShouldEqual, "*")
			So(c.AppConfig.SMTP.Port, ShouldEqual, 25)
		})
		Convey("should use master key as token secret as fallback in YAML", func() {
			c, err := NewTenantConfigurationFromYAML(strings.NewReader(inputMinimalYAML))
			So(err, ShouldBeNil)
			So(c.UserConfig.MasterKey, ShouldEqual, "masterkey")
			So(c.UserConfig.TokenStore.Secret, ShouldEqual, c.UserConfig.MasterKey)
		})
		Convey("should validate when load from YAML", func() {
			invalidInput := `app_name: myapp
app_config:
  database_url: postgres://
user_config:
  api_key: apikey
  master_key: masterkey
`
			_, err := NewTenantConfigurationFromYAML(strings.NewReader(invalidInput))
			So(err, ShouldBeError, "Only version 1 is supported")
		})
		// JSON
		Convey("should have default value when load from JSON", func() {
			c, err := NewTenantConfigurationFromJSON(strings.NewReader(inputMinimalJSON))
			So(err, ShouldBeNil)
			So(c.Version, ShouldEqual, "1")
			So(c.AppName, ShouldEqual, "myapp")
			So(c.AppConfig.DatabaseURL, ShouldEqual, "postgres://")
			So(c.UserConfig.APIKey, ShouldEqual, "apikey")
			So(c.UserConfig.MasterKey, ShouldEqual, "masterkey")
			So(c.UserConfig.WelcomeEmail.Enabled, ShouldEqual, true)
			So(c.UserConfig.CORS.Origin, ShouldEqual, "*")
			So(c.AppConfig.SMTP.Port, ShouldEqual, 25)
		})
		Convey("should use master key as token secret as fallback in JSON", func() {
			c, err := NewTenantConfigurationFromJSON(strings.NewReader(inputMinimalJSON))
			So(err, ShouldBeNil)
			So(err, ShouldBeNil)
			So(c.UserConfig.MasterKey, ShouldEqual, "masterkey")
			So(c.UserConfig.TokenStore.Secret, ShouldEqual, c.UserConfig.MasterKey)
		})
		Convey("should validate when load from JSON", func() {
			invalidInput := `
{
  "app_name": "myapp",
  "app_config": {
    "database_url": "postgres://"
  },
  "user_config": {
    "api_key": "apikey",
    "master_key": "masterkey",
    "welcome_email": {
      "enabled": true
    }
  }
}
			`
			_, err := NewTenantConfigurationFromJSON(strings.NewReader(invalidInput))
			So(err, ShouldBeError, "Only version 1 is supported")
		})
		// Conversion
		Convey("should be losslessly converted between Go and msgpack", func() {
			c := NewTenantConfigurationFromScratch(FromScratchOptions{
				AppName:     "myapp",
				DatabaseURL: "postgres://",
				APIKey:      "apikey",
				MasterKey:   "masterkey",
			})
			base64msgpack, err := c.StdBase64Msgpack()
			So(err, ShouldBeNil)

			cc, err := NewTenantConfigurationFromStdBase64Msgpack(base64msgpack)
			So(err, ShouldBeNil)
			So(c, ShouldResemble, *cc)
		})
		Convey("should be losslessly converted between Go and JSON", func() {
			c := NewTenantConfigurationFromScratch(FromScratchOptions{
				AppName:     "myapp",
				DatabaseURL: "postgres://",
				APIKey:      "apikey",
				MasterKey:   "masterkey",
			})
			b, err := json.Marshal(c)
			So(err, ShouldBeNil)

			cc, err := NewTenantConfigurationFromJSON(bytes.NewReader(b))
			So(err, ShouldBeNil)
			So(c, ShouldResemble, *cc)
		})
		Convey("should be losslessly converted between Go and YAML", func() {
			c := NewTenantConfigurationFromScratch(FromScratchOptions{
				AppName:     "myapp",
				DatabaseURL: "postgres://",
				APIKey:      "apikey",
				MasterKey:   "masterkey",
			})
			b, err := yaml.Marshal(c)
			So(err, ShouldBeNil)

			cc, err := NewTenantConfigurationFromYAML(bytes.NewReader(b))
			So(err, ShouldBeNil)
			So(c, ShouldNonRecursiveDataDeepEqual, *cc)
		})
	})
}
