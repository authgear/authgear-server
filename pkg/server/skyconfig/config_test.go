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

		Convey("Validate the AUTH_RECORD_KEYS", func() {
			config := NewConfigurationWithKeys()
			os.Setenv("AUTH_RECORD_KEYS", "a,b,c")
			config.ReadFromEnv()
			So(config.Validate(), ShouldBeNil)
			So(config.App.AuthRecordKeys, ShouldResemble, [][]string{
				[]string{"a"},
				[]string{"b"},
				[]string{"c"},
			})

			os.Setenv("AUTH_RECORD_KEYS", "a,a,c")
			config.ReadFromEnv()
			So(config.Validate(), ShouldNotBeNil)

			os.Setenv("AUTH_RECORD_KEYS", "a,b,(abc,bbc,b,d),(bbc,d,abc,b)")
			config.ReadFromEnv()
			So(config.Validate(), ShouldNotBeNil)

			// Clean up
			os.Setenv("AUTH_RECORD_KEYS", "")
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

		Convey("Read token store config correctly", func() {
			config := NewConfigurationWithKeys()
			os.Setenv("TOKEN_STORE", "redis")
			os.Setenv("TOKEN_STORE_PATH", "redis://redis:6379")
			os.Setenv("TOKEN_STORE_PREFIX", "PREFIX")
			os.Setenv("TOKEN_STORE_EXPIRY", "60")

			config.readTokenStore()
			So(config.TokenStore.ImplName, ShouldEqual, "redis")
			So(config.TokenStore.Path, ShouldEqual, "redis://redis:6379")
			So(config.TokenStore.Prefix, ShouldEqual, "PREFIX")
			So(config.TokenStore.Expiry, ShouldEqual, 60)

			os.Setenv("TOKEN_STORE", "")
			os.Setenv("TOKEN_STORE_PATH", "")
			os.Setenv("TOKEN_STORE_PREFIX", "")
			os.Setenv("TOKEN_STORE_EXPIRY", "")
		})

		Convey("Read plugin config correctly", func() {
			config := NewConfigurationWithKeys()
			os.Setenv("PLUGINS", "CAT")
			os.Setenv("CAT_TRANSPORT", "exec")
			os.Setenv("CAT_PATH", "py-skygear")
			os.Setenv("CAT_ARGS", "chima,faseng")

			config.readPlugins()
			So(config.Plugin["CAT"], ShouldResemble, &PluginConfig{
				"exec",
				"py-skygear",
				[]string{"chima", "faseng"},
			})

			os.Setenv("PLUGINS", "")
			os.Setenv("CAT_TRANSPORT", "")
			os.Setenv("CAT_PATH", "")
			os.Setenv("CAT_ARGS", "")
		})

		Convey("Read multiple plugin config correctly", func() {
			config := NewConfigurationWithKeys()
			os.Setenv("PLUGINS", "CAT,BUG")
			os.Setenv("CAT_TRANSPORT", "exec")
			os.Setenv("CAT_PATH", "py-skygear")
			os.Setenv("CAT_ARGS", "chima,faseng")
			os.Setenv("BUG_TRANSPORT", "zmq")
			os.Setenv("BUG_PATH", "tcp://skygear:5555")

			config.readPlugins()
			So(config.Plugin["CAT"], ShouldResemble, &PluginConfig{
				"exec",
				"py-skygear",
				[]string{"chima", "faseng"},
			})

			So(config.Plugin["BUG"], ShouldResemble, &PluginConfig{
				"zmq",
				"tcp://skygear:5555",
				nil,
			})

			os.Setenv("PLUGINS", "")
			os.Setenv("CAT_TRANSPORT", "")
			os.Setenv("CAT_PATH", "")
			os.Setenv("CAT_ARGS", "")
			os.Setenv("BUG_TRANSPORT", "")
			os.Setenv("BUG_PATH", "")
		})
	})
}

func TestParseAuthRecordKeys(t *testing.T) {
	Convey("Get keys correctly", t, func() {
		result, err := parseAuthRecordKeys("a,b,c")
		So(result, ShouldResemble, [][]string{[]string{"a"}, []string{"b"}, []string{"c"}})
		So(err, ShouldBeNil)

		result, err = parseAuthRecordKeys("a, b ,c")
		So(result, ShouldResemble, [][]string{[]string{"a"}, []string{"b"}, []string{"c"}})
		So(err, ShouldBeNil)
	})

	Convey("Get key tuples correctly", t, func() {
		result, err := parseAuthRecordKeys("a,(b,c),(d,e,f),(x)")
		So(result, ShouldResemble, [][]string{[]string{"a"}, []string{"b", "c"}, []string{"d", "e", "f"}, []string{"x"}})
		So(err, ShouldBeNil)

		result, err = parseAuthRecordKeys("a,(b,c),( d, e ,f ),(x)")
		So(result, ShouldResemble, [][]string{[]string{"a"}, []string{"b", "c"}, []string{"d", "e", "f"}, []string{"x"}})
		So(err, ShouldBeNil)
	})

	Convey("Throw error for unexpected token", t, func() {
		_, err := parseAuthRecordKeys("a,(b,(c)")
		So(err, ShouldNotBeNil)

		_, err = parseAuthRecordKeys("a,(b,)c)")
		So(err, ShouldNotBeNil)
	})
}
