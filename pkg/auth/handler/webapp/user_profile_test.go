package webapp

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
)

func makeFull() *model.User {
	return &model.User{
		StandardAttributes: map[string]interface{}{
			stdattrs.Email:               "johndoe@example.com",
			stdattrs.EmailVerified:       true,
			stdattrs.PhoneNumber:         "+1234567890",
			stdattrs.PhoneNumberVerified: true,
			stdattrs.PreferredUsername:   "johndoe",
			stdattrs.FamilyName:          "Doe",
			stdattrs.GivenName:           "John",
			stdattrs.MiddleName:          "M",
			stdattrs.Name:                "John Doe",
			stdattrs.Nickname:            "JD",
			stdattrs.Picture:             "http://example.com/johndoe.jpg",
			stdattrs.Profile:             "http://example.com/johndoe",
			stdattrs.Website:             "http://johndoe.com",
			stdattrs.Gender:              "male",
			stdattrs.Birthdate:           "1980-01-01",
			stdattrs.Zoneinfo:            "America/Los_Angeles",
			stdattrs.Locale:              "en_US",
		},
	}
}

func makeFalsy() *model.User {
	return &model.User{
		StandardAttributes: map[string]interface{}{
			stdattrs.Email:               "",
			stdattrs.EmailVerified:       false,
			stdattrs.PhoneNumber:         "",
			stdattrs.PhoneNumberVerified: false,
			stdattrs.PreferredUsername:   "",
			stdattrs.FamilyName:          "",
			stdattrs.GivenName:           "",
			stdattrs.MiddleName:          "",
			stdattrs.Name:                "",
			stdattrs.Nickname:            "",
			stdattrs.Picture:             "",
			stdattrs.Profile:             "",
			stdattrs.Website:             "",
			stdattrs.Gender:              "",
			stdattrs.Birthdate:           "",
			stdattrs.Zoneinfo:            "",
			stdattrs.Locale:              "",
		},
	}
}

func makePartial() *model.User {
	return &model.User{
		StandardAttributes: map[string]interface{}{
			stdattrs.Email:             "johndoe@example.com",
			stdattrs.EmailVerified:     true,
			stdattrs.PreferredUsername: "johndoe",
			stdattrs.FamilyName:        "Doe",
			stdattrs.MiddleName:        "M",
			stdattrs.Name:              "John Doe",
			stdattrs.Nickname:          "JD",
			stdattrs.Picture:           "http://example.com/johndoe.jpg",
			stdattrs.Website:           "http://johndoe.com",
			stdattrs.Gender:            "male",
			stdattrs.Birthdate:         "1980-01-01",
			stdattrs.Zoneinfo:          "America/Los_Angeles",
			stdattrs.Locale:            "",
		},
	}
}

func TestExtract(t *testing.T) {
	Convey("Extract", t, func() {
		Convey("Empty profile should have nil filled", func() {
			actual := GetUserProfile(&model.User{})
			expected := UserProfile{
				stdattrs.Email:               "",
				stdattrs.EmailVerified:       false,
				stdattrs.PhoneNumber:         "",
				stdattrs.PhoneNumberVerified: false,
				stdattrs.PreferredUsername:   "",
				stdattrs.FamilyName:          "",
				stdattrs.GivenName:           "",
				stdattrs.MiddleName:          "",
				stdattrs.Name:                "",
				stdattrs.Nickname:            "",
				stdattrs.Picture:             "",
				stdattrs.Profile:             "",
				stdattrs.Website:             "",
				stdattrs.Gender:              "",
				stdattrs.Birthdate:           "",
				stdattrs.Zoneinfo:            "",
				stdattrs.Locale:              "",

				makeIsPresentAttr(stdattrs.Email):               false,
				makeIsPresentAttr(stdattrs.EmailVerified):       false,
				makeIsPresentAttr(stdattrs.PhoneNumber):         false,
				makeIsPresentAttr(stdattrs.PhoneNumberVerified): false,
				makeIsPresentAttr(stdattrs.PreferredUsername):   false,
				makeIsPresentAttr(stdattrs.FamilyName):          false,
				makeIsPresentAttr(stdattrs.GivenName):           false,
				makeIsPresentAttr(stdattrs.MiddleName):          false,
				makeIsPresentAttr(stdattrs.Name):                false,
				makeIsPresentAttr(stdattrs.Nickname):            false,
				makeIsPresentAttr(stdattrs.Picture):             false,
				makeIsPresentAttr(stdattrs.Profile):             false,
				makeIsPresentAttr(stdattrs.Website):             false,
				makeIsPresentAttr(stdattrs.Gender):              false,
				makeIsPresentAttr(stdattrs.Birthdate):           false,
				makeIsPresentAttr(stdattrs.Zoneinfo):            false,
				makeIsPresentAttr(stdattrs.Locale):              false,
			}
			So(actual, ShouldResemble, expected)
		})

		Convey("Full profile should be extracted correctly with is_present flags", func() {
			actual := GetUserProfile(makeFull())
			expected := UserProfile{
				stdattrs.Email:               "johndoe@example.com",
				stdattrs.EmailVerified:       true,
				stdattrs.PhoneNumber:         "+1234567890",
				stdattrs.PhoneNumberVerified: true,
				stdattrs.PreferredUsername:   "johndoe",
				stdattrs.FamilyName:          "Doe",
				stdattrs.GivenName:           "John",
				stdattrs.MiddleName:          "M",
				stdattrs.Name:                "John Doe",
				stdattrs.Nickname:            "JD",
				stdattrs.Picture:             "http://example.com/johndoe.jpg",
				stdattrs.Profile:             "http://example.com/johndoe",
				stdattrs.Website:             "http://johndoe.com",
				stdattrs.Gender:              "male",
				stdattrs.Birthdate:           "1980-01-01",
				stdattrs.Zoneinfo:            "America/Los_Angeles",
				stdattrs.Locale:              "en_US",

				makeIsPresentAttr(stdattrs.Email):               true,
				makeIsPresentAttr(stdattrs.EmailVerified):       true,
				makeIsPresentAttr(stdattrs.PhoneNumber):         true,
				makeIsPresentAttr(stdattrs.PhoneNumberVerified): true,
				makeIsPresentAttr(stdattrs.PreferredUsername):   true,
				makeIsPresentAttr(stdattrs.FamilyName):          true,
				makeIsPresentAttr(stdattrs.GivenName):           true,
				makeIsPresentAttr(stdattrs.MiddleName):          true,
				makeIsPresentAttr(stdattrs.Name):                true,
				makeIsPresentAttr(stdattrs.Nickname):            true,
				makeIsPresentAttr(stdattrs.Picture):             true,
				makeIsPresentAttr(stdattrs.Profile):             true,
				makeIsPresentAttr(stdattrs.Website):             true,
				makeIsPresentAttr(stdattrs.Gender):              true,
				makeIsPresentAttr(stdattrs.Birthdate):           true,
				makeIsPresentAttr(stdattrs.Zoneinfo):            true,
				makeIsPresentAttr(stdattrs.Locale):              true,
			}
			So(actual, ShouldResemble, expected)
		})

		Convey("Falsy profile should be extracted correctly with is_present flags", func() {
			actual := GetUserProfile(makeFalsy())
			expected := UserProfile{
				stdattrs.Email:               "",
				stdattrs.EmailVerified:       false,
				stdattrs.PhoneNumber:         "",
				stdattrs.PhoneNumberVerified: false,
				stdattrs.PreferredUsername:   "",
				stdattrs.FamilyName:          "",
				stdattrs.GivenName:           "",
				stdattrs.MiddleName:          "",
				stdattrs.Name:                "",
				stdattrs.Nickname:            "",
				stdattrs.Picture:             "",
				stdattrs.Profile:             "",
				stdattrs.Website:             "",
				stdattrs.Gender:              "",
				stdattrs.Birthdate:           "",
				stdattrs.Zoneinfo:            "",
				stdattrs.Locale:              "",

				makeIsPresentAttr(stdattrs.Email):               false,
				makeIsPresentAttr(stdattrs.EmailVerified):       true,
				makeIsPresentAttr(stdattrs.PhoneNumber):         false,
				makeIsPresentAttr(stdattrs.PhoneNumberVerified): true,
				makeIsPresentAttr(stdattrs.PreferredUsername):   false,
				makeIsPresentAttr(stdattrs.FamilyName):          false,
				makeIsPresentAttr(stdattrs.GivenName):           false,
				makeIsPresentAttr(stdattrs.MiddleName):          false,
				makeIsPresentAttr(stdattrs.Name):                false,
				makeIsPresentAttr(stdattrs.Nickname):            false,
				makeIsPresentAttr(stdattrs.Picture):             false,
				makeIsPresentAttr(stdattrs.Profile):             false,
				makeIsPresentAttr(stdattrs.Website):             false,
				makeIsPresentAttr(stdattrs.Gender):              false,
				makeIsPresentAttr(stdattrs.Birthdate):           false,
				makeIsPresentAttr(stdattrs.Zoneinfo):            false,
				makeIsPresentAttr(stdattrs.Locale):              false,
			}
			So(actual, ShouldResemble, expected)
		})

		Convey("Partial profile should be extracted correctly with is_present flags", func() {
			actual := GetUserProfile(makePartial())
			expected := UserProfile{
				stdattrs.Email:               "johndoe@example.com",
				stdattrs.EmailVerified:       true,
				stdattrs.PhoneNumber:         "",
				stdattrs.PhoneNumberVerified: false,
				stdattrs.PreferredUsername:   "johndoe",
				stdattrs.FamilyName:          "Doe",
				stdattrs.GivenName:           "",
				stdattrs.MiddleName:          "M",
				stdattrs.Name:                "John Doe",
				stdattrs.Nickname:            "JD",
				stdattrs.Picture:             "http://example.com/johndoe.jpg",
				stdattrs.Profile:             "",
				stdattrs.Website:             "http://johndoe.com",
				stdattrs.Gender:              "male",
				stdattrs.Birthdate:           "1980-01-01",
				stdattrs.Zoneinfo:            "America/Los_Angeles",
				stdattrs.Locale:              "",

				makeIsPresentAttr(stdattrs.Email):               true,
				makeIsPresentAttr(stdattrs.EmailVerified):       true,
				makeIsPresentAttr(stdattrs.PhoneNumber):         false,
				makeIsPresentAttr(stdattrs.PhoneNumberVerified): false,
				makeIsPresentAttr(stdattrs.PreferredUsername):   true,
				makeIsPresentAttr(stdattrs.FamilyName):          true,
				makeIsPresentAttr(stdattrs.GivenName):           false,
				makeIsPresentAttr(stdattrs.MiddleName):          true,
				makeIsPresentAttr(stdattrs.Name):                true,
				makeIsPresentAttr(stdattrs.Nickname):            true,
				makeIsPresentAttr(stdattrs.Picture):             true,
				makeIsPresentAttr(stdattrs.Profile):             false,
				makeIsPresentAttr(stdattrs.Website):             true,
				makeIsPresentAttr(stdattrs.Gender):              true,
				makeIsPresentAttr(stdattrs.Birthdate):           true,
				makeIsPresentAttr(stdattrs.Zoneinfo):            true,
				makeIsPresentAttr(stdattrs.Locale):              false,
			}

			So(actual, ShouldResemble, expected)
		})
	})
}
