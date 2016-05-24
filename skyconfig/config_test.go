package skyconfig

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRequiredConfig(t *testing.T) {
	Convey("Config", t, func() {
		Convey("NewConfiguration will provide default", func() {
			config := NewConfiguration()
			So(config.HTTP.Host, ShouldEqual, ":3000")
			So(config.App.Name, ShouldEqual, "myapp")
		})

		Convey("Required APP_NAME, API_KEY, MASTER_KEY, DATABASE_URL all present will pass", func() {
			config := Configuration{}
			err := config.Validate()
			So(err, ShouldNotBeNil)
			os.Setenv("APP_NAME", "nonempty")
			os.Setenv("API_KEY", "so-secret")
			os.Setenv("MASTER_KEY", "ging-secret")
			os.Setenv("DATABASE_URL", "postgres://postgres:@localhost/postgres?sslmode=disable")

			config.ReadFromEnv()
			err = config.Validate()
			So(err, ShouldBeNil)

			// Clean up
			os.Setenv("APP_NAME", "")
			os.Setenv("API_KEY", "")
			os.Setenv("MASTER_KEY", "")
			os.Setenv("DATABASE_URL", "")
		})

		Convey("NewConfigurationWithKeys is ready to use", func() {
			config := NewConfigurationWithKeys()
			So(config.Validate(), ShouldBeNil)
		})

		Convey("Validate the APP_NAME", func() {
			config := NewConfigurationWithKeys()
			os.Setenv("APP_NAME", "nonempty")
			config.ReadFromEnv()
			So(config.Validate(), ShouldBeNil)

			os.Setenv("APP_NAME", "YES_itcan")
			config.ReadFromEnv()
			So(config.Validate(), ShouldBeNil)

			os.Setenv("APP_NAME", "No-itcannot")
			config.ReadFromEnv()
			So(config.Validate(), ShouldNotBeNil)

			os.Setenv("APP_NAME", "No!")
			config.ReadFromEnv()
			So(config.Validate(), ShouldNotBeNil)

			os.Setenv("APP_NAME", "")
			config.ReadFromEnv()
			So(config.Validate(), ShouldNotBeNil)

			// Clean up
			os.Setenv("APP_NAME", "")
		})

		Convey("Validate the APNS_ENV", func() {
			config := NewConfigurationWithKeys()
			os.Setenv("APNS_ENABLE", "YES")
			So(config.Validate(), ShouldBeNil)

			os.Setenv("APNS_ENV", "sandbox")
			config.ReadFromEnv()
			So(config.Validate(), ShouldBeNil)

			os.Setenv("APNS_ENV", "production")
			config.ReadFromEnv()
			So(config.Validate(), ShouldBeNil)

			os.Setenv("APNS_ENV", "apple.com")
			config.ReadFromEnv()
			So(config.Validate(), ShouldNotBeNil)

			os.Setenv("APNS_ENABLE", "")
		})
	})
}
