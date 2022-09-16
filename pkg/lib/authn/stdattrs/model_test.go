package stdattrs

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

func TestT(t *testing.T) {
	Convey("T", t, func() {
		Convey("WithNameCopiedToGivenName", func() {
			So(T{}.WithNameCopiedToGivenName(), ShouldResemble, T{})
			So(T{
				"name": "John",
			}.WithNameCopiedToGivenName(), ShouldResemble, T{
				"name":       "John",
				"given_name": "John",
			})
			So(T{
				"name":       "John",
				"given_name": "Jonathan",
			}.WithNameCopiedToGivenName(), ShouldResemble, T{
				"name":       "John",
				"given_name": "Jonathan",
			})
		})

		Convey("NonIdentityAware", func() {
			So(T{
				"a":     "b",
				"name":  "John Doe",
				"email": "louischan@oursky.com",
			}.NonIdentityAware(), ShouldResemble, T{
				"name": "John Doe",
			})
		})

		Convey("MergedWith", func() {
			So(T{
				"a":    "b",
				"keep": "this",
			}.MergedWith(T{
				"a":   "c",
				"new": "key",
			}), ShouldResemble, T{
				"keep": "this",
				"new":  "key",
				"a":    "c",
			})
		})

		Convey("FormattedName", func() {
			So(T{
				"name": "John Doe",
			}.FormattedName(), ShouldEqual, "John Doe")
			So(T{
				"given_name":  "John",
				"family_name": "Doe",
			}.FormattedName(), ShouldEqual, "John Doe")
			So(T{
				"given_name":  "John",
				"middle_name": "William",
				"family_name": "Doe",
			}.FormattedName(), ShouldEqual, "John William Doe")
			So(T{
				"given_name":  "John",
				"family_name": "Doe",
				"nickname":    "Johnny",
			}.FormattedName(), ShouldEqual, "John Doe (Johnny)")
		})

		Convey("FormattedNames", func() {
			So(T{}.FormattedNames(), ShouldEqual, "")
			So(T{
				"nickname": "John",
			}.FormattedNames(), ShouldEqual, "John")
			So(T{
				"given_name":  "John",
				"middle_name": "William",
				"family_name": "Doe",
			}.FormattedNames(), ShouldEqual, "John William Doe")
			So(T{
				"given_name":  "John",
				"family_name": "Doe",
				"nickname":    "Johnny",
			}.FormattedNames(), ShouldEqual, "John Doe (Johnny)")
			So(T{
				"name": "John Doe",
			}.FormattedNames(), ShouldEqual, "John Doe")
			So(T{
				"name":     "John Doe",
				"nickname": "Johnny",
			}.FormattedNames(), ShouldEqual, "John Doe (Johnny)")
			So(T{
				"name":        "John Doe",
				"given_name":  "John",
				"family_name": "Doe",
			}.FormattedNames(), ShouldEqual, "John Doe\nJohn Doe")
			So(T{
				"name":        "John Doe",
				"given_name":  "John",
				"family_name": "Doe",
				"nickname":    "Johnny",
			}.FormattedNames(), ShouldEqual, "John Doe (Johnny)\nJohn Doe")
		})

		Convey("Clone", func() {
			a := T{
				"address": map[string]interface{}{
					"street_address": "a",
				},
			}
			b := a.Clone()
			b["address"].(map[string]interface{})["street_address"] = "b"

			So(b, ShouldResemble, T{
				"address": map[string]interface{}{
					"street_address": "b",
				},
			})
			So(a, ShouldResemble, T{
				"address": map[string]interface{}{
					"street_address": "a",
				},
			})
		})

		Convey("Tidy", func() {
			a := T{
				"name":    "John Doe",
				"address": map[string]interface{}{},
			}
			So(a.Tidy(), ShouldResemble, T{
				"name": "John Doe",
			})
		})

		Convey("MergedWithJSONPointer", func() {
			test := func(original T, ptrs map[string]string, expected T) {
				actual, err := original.MergedWithJSONPointer(ptrs)
				So(err, ShouldBeNil)
				So(actual, ShouldResemble, expected)
			}

			test(T{
				"name":        "John Doe",
				"given_name":  "John",
				"family_name": "Doe",
				"address": map[string]interface{}{
					"street_address": "Some street",
				},
			}, map[string]string{
				"/given_name":             "",
				"/family_name":            "Lee",
				"/middle_name":            "William",
				"/address/street_address": "",
			}, T{
				"name":        "John Doe",
				"middle_name": "William",
				"family_name": "Lee",
			})
		})

		Convey("ReadWithAccessControl", func() {
			accessControl := accesscontrol.T{
				accesscontrol.Subject("/name"): map[accesscontrol.Role]accesscontrol.Level{
					config.RoleEndUser:  config.AccessControlLevelHidden,
					config.RoleBearer:   config.AccessControlLevelReadwrite,
					config.RolePortalUI: config.AccessControlLevelReadwrite,
				},
				accesscontrol.Subject("/given_name"): map[accesscontrol.Role]accesscontrol.Level{
					config.RoleEndUser:  config.AccessControlLevelReadwrite,
					config.RoleBearer:   config.AccessControlLevelReadwrite,
					config.RolePortalUI: config.AccessControlLevelReadwrite,
				},
			}
			stdAttrs := T{
				"name":       "John Doe",
				"given_name": "John Doe",
			}

			So(stdAttrs.ReadWithAccessControl(accessControl, config.RoleEndUser), ShouldResemble, T{
				"given_name": "John Doe",
			})

			So(stdAttrs.ReadWithAccessControl(accessControl, config.RolePortalUI), ShouldResemble, T{
				"name":       "John Doe",
				"given_name": "John Doe",
			})
		})

		Convey("CheckWrite", func() {
			accessControl := accesscontrol.T{
				accesscontrol.Subject("/name"): map[accesscontrol.Role]accesscontrol.Level{
					config.RoleEndUser:  config.AccessControlLevelHidden,
					config.RoleBearer:   config.AccessControlLevelReadwrite,
					config.RolePortalUI: config.AccessControlLevelReadwrite,
				},
				accesscontrol.Subject("/nickname"): map[accesscontrol.Role]accesscontrol.Level{
					config.RoleEndUser:  config.AccessControlLevelHidden,
					config.RoleBearer:   config.AccessControlLevelReadwrite,
					config.RolePortalUI: config.AccessControlLevelReadwrite,
				},
				accesscontrol.Subject("/given_name"): map[accesscontrol.Role]accesscontrol.Level{
					config.RoleEndUser:  config.AccessControlLevelReadwrite,
					config.RoleBearer:   config.AccessControlLevelReadwrite,
					config.RolePortalUI: config.AccessControlLevelReadwrite,
				},
			}

			stdAttrs := T{
				"name":       "John Doe",
				"given_name": "John Doe",
			}

			// Edition
			So(stdAttrs.CheckWrite(accessControl, config.RoleEndUser, T{
				"name":       "42",
				"given_name": "John Doe",
			}), ShouldBeError, "/name being written by end_user with level 1")

			// Deletion
			So(stdAttrs.CheckWrite(accessControl, config.RoleEndUser, T{
				"given_name": "John Doe",
			}), ShouldBeError, "/name being written by end_user with level 1")

			// Addition
			So(stdAttrs.CheckWrite(accessControl, config.RoleEndUser, T{
				"name":       "John Doe",
				"given_name": "John Doe",
				"nickname":   "42",
			}), ShouldBeError, "/nickname being written by end_user with level 1")

			// OK
			So(stdAttrs.CheckWrite(accessControl, config.RoleEndUser, T{
				"name":       "John Doe",
				"given_name": "Jane Doe",
			}), ShouldBeNil)
		})

		Convey("WithDerivedAttributesRemoved", func() {
			So(T{}.WithDerivedAttributesRemoved(), ShouldResemble, T{})

			So(T{
				"name": "John Doe",
			}.WithDerivedAttributesRemoved(), ShouldResemble, T{
				"name": "John Doe",
			})

			So(T{
				"email_verified":        true,
				"phone_number_verified": true,
				"updated_at":            1,
			}.WithDerivedAttributesRemoved(), ShouldResemble, T{})
		})
	})
}
