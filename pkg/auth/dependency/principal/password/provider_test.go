package password

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/passwordhistory"
	"github.com/skygeario/skygear-server/pkg/core/config"
	coreTime "github.com/skygeario/skygear-server/pkg/core/time"
)

func TestProvider(t *testing.T) {
	Convey("Test PasswordProvider", t, func() {
		logger, _ := test.NewNullLogger()
		loggerEntry := logrus.NewEntry(logger)
		allowedRealms := []string{DefaultRealm}
		one := 1
		on := true
		off := false
		loginIDsKeys := []config.LoginIDKeyConfiguration{
			config.LoginIDKeyConfiguration{Key: "email", Type: "email", Maximum: &one},
			config.LoginIDKeyConfiguration{Key: "username", Type: "username", Maximum: &one},
		}
		loginIDTypes := &config.LoginIDTypesConfiguration{
			Email: &config.LoginIDTypeEmailConfiguration{
				CaseSensitive: &off,
				BlockPlusSign: &off,
				IgnoreDotSign: &on,
			},
			Username: &config.LoginIDTypeUsernameConfiguration{
				BlockReservedUsernames: &on,
				ExcludedKeywords:       []string{"myapp"},
				ASCIIOnly:              &off,
				CaseSensitive:          &off,
			},
		}
		reservedNameChecker, _ := loginid.NewReservedNameChecker("../../../../../reserved_name.txt")
		pwProvider := &providerImpl{
			timeProvider: &coreTime.MockProvider{TimeNowUTC: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)},
			store:        NewMockStore(),
			logger:       loggerEntry,
			loginIDsKeys: loginIDsKeys,
			loginIDChecker: loginid.NewDefaultLoginIDChecker(
				loginIDsKeys,
				loginIDTypes,
				reservedNameChecker,
			),
			realmChecker: defaultRealmChecker{
				allowedRealms: allowedRealms,
			},
			loginIDNormalizerFactory: loginid.NewLoginIDNormalizerFactory(loginIDsKeys, loginIDTypes),
			allowedRealms:            allowedRealms,
			passwordHistoryEnabled:   false,
			passwordHistoryStore:     passwordhistory.NewMockPasswordHistoryStore(),
		}

		Convey("create principal", func() {
			Convey("should reject same email with different cases", func() {
				loginIDs := []loginid.LoginID{
					loginid.LoginID{
						Key:   "email",
						Value: "Faseng@example.com",
					},
				}
				principals, err := pwProvider.CreatePrincipalsByLoginID("user1", "password", loginIDs, DefaultRealm)
				So(principals[0].OriginalLoginID, ShouldEqual, loginIDs[0].Value)
				So(err, ShouldBeNil)

				loginIDs = []loginid.LoginID{
					loginid.LoginID{
						Key:   "email",
						Value: "FASENG@example.com",
					},
				}

				_, err = pwProvider.CreatePrincipalsByLoginID("user2", "password", loginIDs, DefaultRealm)
				So(err, ShouldBeError, "login ID is already used")
			})

			Convey("should reject email with same punycode encoded domain", func() {
				loginIDs := []loginid.LoginID{
					loginid.LoginID{
						Key:   "email",
						Value: "faseng@測試.com",
					},
				}
				principals, err := pwProvider.CreatePrincipalsByLoginID("user1", "password", loginIDs, DefaultRealm)
				So(principals[0].OriginalLoginID, ShouldEqual, loginIDs[0].Value)
				So(err, ShouldBeNil)

				loginIDs = []loginid.LoginID{
					loginid.LoginID{
						Key:   "email",
						Value: "faseng@xn--g6w251d.com",
					},
				}

				_, err = pwProvider.CreatePrincipalsByLoginID("user2", "password", loginIDs, DefaultRealm)
				So(err, ShouldBeError, "login ID is already used")
			})
		})

		Convey("get principals", func() {
			Convey("should be able to get principals with value before normalization", func() {
				loginIDs := []loginid.LoginID{
					loginid.LoginID{
						Key:   "email",
						Value: "Faseng.Cat@example.com",
					},
				}
				principals, err := pwProvider.CreatePrincipalsByLoginID("user1", "password", loginIDs, DefaultRealm)
				So(principals[0].OriginalLoginID, ShouldEqual, loginIDs[0].Value)
				So(err, ShouldBeNil)

				principalID := principals[0].ID

				_, err = pwProvider.CreatePrincipalsByLoginID("user2", "password", []loginid.LoginID{
					loginid.LoginID{
						Key:   "email",
						Value: "chima@example.com",
					},
				}, DefaultRealm)
				So(err, ShouldBeNil)

				principal := Principal{}
				err = pwProvider.GetPrincipalByLoginIDWithRealm("email", "Faseng.Cat@example.com", DefaultRealm, &principal)
				So(err, ShouldBeNil)
				So(principal.ID, ShouldEqual, principalID)

				principal = Principal{}
				err = pwProvider.GetPrincipalByLoginIDWithRealm("email", "FASENGCAT@EXAMPLE.COM", DefaultRealm, &principal)
				So(err, ShouldBeNil)
				So(principal.ID, ShouldEqual, principalID)

				principal = Principal{}
				err = pwProvider.GetPrincipalByLoginIDWithRealm("email", "fasengcat@example.com", DefaultRealm, &principal)
				So(err, ShouldBeNil)
				So(principal.ID, ShouldEqual, principalID)

				principal = Principal{}
				err = pwProvider.GetPrincipalByLoginIDWithRealm("email", "chima@example.com", DefaultRealm, &principal)
				So(err, ShouldBeNil)
				So(principal.ID, ShouldNotEqual, principalID)

				principal = Principal{}
				err = pwProvider.GetPrincipalByLoginIDWithRealm("email", "milktea@example.com", DefaultRealm, &principal)
				So(err, ShouldBeError, "principal not found")
			})

			Convey("should be able to get principals without login id key", func() {
				loginIDs := []loginid.LoginID{
					loginid.LoginID{
						Key:   "email",
						Value: "Faseng.Cat@example.com",
					},
				}
				principals, err := pwProvider.CreatePrincipalsByLoginID("user1", "password", loginIDs, DefaultRealm)
				So(principals[0].OriginalLoginID, ShouldEqual, loginIDs[0].Value)
				So(err, ShouldBeNil)

				principalID := principals[0].ID

				_, err = pwProvider.CreatePrincipalsByLoginID("user2", "password", []loginid.LoginID{
					loginid.LoginID{
						Key:   "email",
						Value: "chima@example.com",
					},
				}, DefaultRealm)
				So(err, ShouldBeNil)

				principal := Principal{}
				err = pwProvider.GetPrincipalByLoginIDWithRealm("", "faseng.cat@example.com", DefaultRealm, &principal)
				So(err, ShouldBeNil)
				So(principal.ID, ShouldEqual, principalID)

				principal = Principal{}
				err = pwProvider.GetPrincipalByLoginIDWithRealm("", "chima@example.com", DefaultRealm, &principal)
				So(err, ShouldBeNil)
				So(principal.ID, ShouldNotEqual, principalID)

				principal = Principal{}
				err = pwProvider.GetPrincipalByLoginIDWithRealm("", "milktea@example.com", DefaultRealm, &principal)
				So(err, ShouldBeError, "principal not found")
			})

			Convey("should return error with ambiguous login id", func() {
				loginIDs := []loginid.LoginID{
					loginid.LoginID{
						Key:   "email",
						Value: "Faseng.Cat@example.com",
					},
				}
				principals, err := pwProvider.CreatePrincipalsByLoginID("user1", "password", loginIDs, DefaultRealm)
				So(principals[0].OriginalLoginID, ShouldEqual, loginIDs[0].Value)
				So(err, ShouldBeNil)

				// ASCIIOnly of username need to be false for this test
				_, err = pwProvider.CreatePrincipalsByLoginID("user2", "password", []loginid.LoginID{
					loginid.LoginID{
						Key:   "username",
						Value: "faseng.cat@example.com",
					},
				}, DefaultRealm)
				So(err, ShouldBeNil)

				principal := Principal{}
				err = pwProvider.GetPrincipalByLoginIDWithRealm("", "faseng.cat@example.com", DefaultRealm, &principal)
				So(err, ShouldBeError, "multiple principals found")

			})
		})
	})
}
