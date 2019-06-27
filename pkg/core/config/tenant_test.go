package config

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
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
			c, _ := NewTenantConfigurationFromScratch(FromScratchOptions{
				AppName:     "myapp",
				DatabaseURL: "postgres://",
				APIKey:      "apikey",
				MasterKey:   "masterkey",
			})
			base64msgpack, err := c.StdBase64Msgpack()
			So(err, ShouldBeNil)

			cc, err := NewTenantConfigurationFromStdBase64Msgpack(base64msgpack)
			So(err, ShouldBeNil)
			So(c, ShouldResemble, cc)
		})
		Convey("should be losslessly converted between Go and JSON", func() {
			c, _ := NewTenantConfigurationFromScratch(FromScratchOptions{
				AppName:     "myapp",
				DatabaseURL: "postgres://",
				APIKey:      "apikey",
				MasterKey:   "masterkey",
			})
			b, err := json.Marshal(c)
			So(err, ShouldBeNil)

			cc, err := NewTenantConfigurationFromJSON(bytes.NewReader(b))
			So(err, ShouldBeNil)
			So(c, ShouldResemble, cc)
		})
		Convey("should be losslessly converted between Go and YAML", func() {
			c, _ := NewTenantConfigurationFromScratch(FromScratchOptions{
				AppName:     "myapp",
				DatabaseURL: "postgres://",
				APIKey:      "apikey",
				MasterKey:   "masterkey",
			})
			b, err := yaml.Marshal(c)
			So(err, ShouldBeNil)

			cc, err := NewTenantConfigurationFromYAML(bytes.NewReader(b))
			So(err, ShouldBeNil)
			So(c, ShouldNonRecursiveDataDeepEqual, cc)
		})
		// DeploymentRoutes
		Convey("should serialize deployment routes", func() {
			c, _ := NewTenantConfigurationFromScratch(FromScratchOptions{
				AppName:     "myapp",
				DatabaseURL: "postgres://",
				APIKey:      "apikey",
				MasterKey:   "masterkey",
			})
			c.DeploymentRoutes = []DeploymentRoute{
				DeploymentRoute{
					Version: "a",
					Path:    "/",
					Type:    "http-service",
					TypeConfig: map[string]interface{}{
						"backend_url": "http://localhost:3000",
					},
				},
				DeploymentRoute{
					Version: "a",
					Path:    "/api",
					Type:    "http-service",
					TypeConfig: map[string]interface{}{
						"backend_url": "http://localhost:3001",
					},
				},
			}
			b, err := yaml.Marshal(c)
			So(err, ShouldBeNil)

			cc, err := NewTenantConfigurationFromYAML(bytes.NewReader(b))
			So(err, ShouldBeNil)
			So(c, ShouldNonRecursiveDataDeepEqual, cc)
		})
		// Env
		Convey("should load tenant config from env", func() {
			os.Clearenv()
			_, err := NewTenantConfigurationFromEnv()
			So(err, ShouldBeError, "DATABASE_URL is not set")

			os.Setenv("DATABASE_URL", "postgres://")
			os.Setenv("APP_NAME", "myapp")
			os.Setenv("API_KEY", "api_key")
			os.Setenv("MASTER_KEY", "master_key")
			c, err := NewTenantConfigurationFromEnv()

			So(err, ShouldBeNil)
			So(c.AppName, ShouldEqual, "myapp")
			So(c.AppConfig.DatabaseURL, ShouldEqual, "postgres://")
			So(c.UserConfig.APIKey, ShouldEqual, "api_key")
			So(c.UserConfig.MasterKey, ShouldEqual, "master_key")
		})
		Convey("should load tenant config from yaml and env", func() {
			os.Clearenv()

			os.Setenv("DATABASE_URL", "postgres://remote")
			os.Setenv("APP_NAME", "yourapp")
			os.Setenv("API_KEY", "your_api_key")
			os.Setenv("MASTER_KEY", "your_master_key")

			c, err := NewTenantConfigurationFromYAMLAndEnv(func() (io.Reader, error) {
				return strings.NewReader(inputMinimalYAML), nil
			})

			So(err, ShouldBeNil)
			So(c.AppName, ShouldEqual, "yourapp")
			So(c.AppConfig.DatabaseURL, ShouldEqual, "postgres://remote")
			So(c.UserConfig.APIKey, ShouldEqual, "your_api_key")
			So(c.UserConfig.MasterKey, ShouldEqual, "your_master_key")
		})
	})
}
