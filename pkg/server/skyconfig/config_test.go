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

		Convey("User audit default values", func() {
			config := NewConfigurationWithKeys()
			config.readUserAudit()
			So(config.UserAudit.Enabled, ShouldEqual, false)
			So(config.UserAudit.TrailHandlerURL, ShouldEqual, "")
			So(config.UserAudit.PwMinLength, ShouldEqual, 0)
			So(config.UserAudit.PwUppercaseRequired, ShouldEqual, false)
			So(config.UserAudit.PwLowercaseRequired, ShouldEqual, false)
			So(config.UserAudit.PwDigitRequired, ShouldEqual, false)
			So(config.UserAudit.PwSymbolRequired, ShouldEqual, false)
			So(config.UserAudit.PwMinGuessableLevel, ShouldEqual, 0)
			So(config.UserAudit.PwExcludedKeywords, ShouldEqual, nil)
			So(config.UserAudit.PwExcludedFields, ShouldEqual, nil)
			So(config.UserAudit.PwHistorySize, ShouldEqual, 0)
			So(config.UserAudit.PwHistoryDays, ShouldEqual, 0)
			So(config.UserAudit.PwExpiryDays, ShouldEqual, 0)
		})

		Convey("Read user audit config correctly", func() {
			config := NewConfigurationWithKeys()
			handlerURLString := "file:///var/log/skygear/audit.log"
			os.Setenv("USER_AUDIT_ENABLED", "true")
			os.Setenv("USER_AUDIT_TRAIL_HANDLER_URL", handlerURLString)
			os.Setenv("USER_AUDIT_PW_MIN_LENGTH", "1")
			os.Setenv("USER_AUDIT_PW_UPPERCASE_REQUIRED", "yes")
			os.Setenv("USER_AUDIT_PW_LOWERCASE_REQUIRED", "y")
			os.Setenv("USER_AUDIT_PW_DIGIT_REQUIRED", "Yes")
			os.Setenv("USER_AUDIT_PW_SYMBOL_REQUIRED", "y")
			os.Setenv("USER_AUDIT_PW_MIN_GUESSABLE_LEVEL", "5")
			os.Setenv("USER_AUDIT_PW_EXCLUDED_KEYWORDS", "a,b")
			os.Setenv("USER_AUDIT_PW_EXCLUDED_FIELDS", "f")
			os.Setenv("USER_AUDIT_PW_HISTORY_SIZE", "2")
			os.Setenv("USER_AUDIT_PW_HISTORY_DAYS", "3")
			os.Setenv("USER_AUDIT_PW_EXPIRY_DAYS", "4")

			config.readUserAudit()
			So(config.UserAudit.Enabled, ShouldEqual, true)
			So(config.UserAudit.TrailHandlerURL, ShouldEqual, handlerURLString)
			So(config.UserAudit.PwMinLength, ShouldEqual, 1)
			So(config.UserAudit.PwUppercaseRequired, ShouldEqual, true)
			So(config.UserAudit.PwLowercaseRequired, ShouldEqual, true)
			So(config.UserAudit.PwDigitRequired, ShouldEqual, true)
			So(config.UserAudit.PwSymbolRequired, ShouldEqual, true)
			So(config.UserAudit.PwMinGuessableLevel, ShouldEqual, 5)
			So(config.UserAudit.PwExcludedKeywords, ShouldResemble, []string{
				"a",
				"b",
			})
			So(config.UserAudit.PwExcludedFields, ShouldResemble, []string{
				"f",
			})
			So(config.UserAudit.PwHistorySize, ShouldEqual, 2)
			So(config.UserAudit.PwHistoryDays, ShouldEqual, 3)
			So(config.UserAudit.PwExpiryDays, ShouldEqual, 4)

			os.Setenv("USER_AUDIT_ENABLED", "")
			os.Setenv("USER_AUDIT_TRAIL_HANDLER_URL", "")
			os.Setenv("USER_AUDIT_PW_MIN_LENGTH", "")
			os.Setenv("USER_AUDIT_PW_UPPERCASE_REQUIRED", "")
			os.Setenv("USER_AUDIT_PW_LOWERCASE_REQUIRED", "")
			os.Setenv("USER_AUDIT_PW_DIGIT_REQUIRED", "")
			os.Setenv("USER_AUDIT_PW_SYMBOL_REQUIRED", "")
			os.Setenv("USER_AUDIT_PW_MIN_GUESSABLE_LEVEL", "")
			os.Setenv("USER_AUDIT_PW_EXCLUDED_KEYWORDS", "")
			os.Setenv("USER_AUDIT_PW_EXCLUDED_FIELDS", "")
			os.Setenv("USER_AUDIT_PW_HISTORY_SIZE", "")
			os.Setenv("USER_AUDIT_PW_HISTORY_DAYS", "")
			os.Setenv("USER_AUDIT_PW_EXPIRY_DAYS", "")
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

func TestParseCommaSeparatedString(t *testing.T) {
	Convey("Parse comma separated string correctly", t, func() {
		result := parseCommaSeparatedString(" , , ")
		So(result, ShouldResemble, []string{})

		result = parseCommaSeparatedString(",,")
		So(result, ShouldResemble, []string{})

		result = parseCommaSeparatedString("a,b")
		So(result, ShouldResemble, []string{"a", "b"})

		result = parseCommaSeparatedString("a, b")
		So(result, ShouldResemble, []string{"a", "b"})
	})
}
